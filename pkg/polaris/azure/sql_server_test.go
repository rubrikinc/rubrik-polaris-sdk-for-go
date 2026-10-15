// Copyright 2026 Rubrik, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package azure

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlazure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// These tests cover the orchestration in SetupSQLManagedInstanceBackup: which
// servers are reported as failed, which task chains are waited for, and how the
// failures are aggregated. The two RSC calls involved, the setup mutation and
// the task chain polling, are stubbed by a single handler which dispatches on
// the query body.
//
// Unlike the graphql-layer tests, the responses here vary per test case rather
// than being one recorded reply, so they are marshalled from Go values instead
// of being rendered from a testdata fixture.

var testCredentials = gqlazure.LoginCredentials{Login: "draper", Password: "s3cret"}

// setupReply is the stubbed reply to the backup setup mutation. Job IDs are
// generated per server, so the tests only name the servers.
type setupReply struct {
	// started lists the servers RSC accepts and starts a task chain for.
	started []uuid.UUID
	// preValidationFailed maps a server to the reason RSC rejected it before
	// starting a task chain.
	preValidationFailed map[uuid.UUID]string
	// taskChainState is the state every started task chain reports.
	taskChainState core.TaskChainState
}

// sqlServerHandler stubs the setup mutation and the task chain polling. It
// returns the handler and a counter of task chain status polls, which the tests
// use to tell "waited for" from "skipped".
func sqlServerHandler(cancel context.CancelCauseFunc, reply setupReply, onPoll func()) (http.Handler, *atomic.Int64) {
	var polls atomic.Int64

	jobIDs := make(map[uuid.UUID]uuid.UUID, len(reply.started))
	for _, serverID := range reply.started {
		jobIDs[serverID] = uuid.New()
	}

	return handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			cancel(err)
			return
		}

		var payload struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			cancel(err)
			return
		}

		var response any
		switch {
		case strings.Contains(payload.Query, "setupCloudNativeSqlServerBackup"):
			type jobID struct {
				JobID    uuid.UUID `json:"jobId"`
				ObjectID uuid.UUID `json:"rubrikObjectId"`
			}
			type jobError struct {
				Error    string    `json:"error"`
				ObjectID uuid.UUID `json:"rubrikObjectId"`
			}

			jobs := []jobID{}
			for _, serverID := range reply.started {
				jobs = append(jobs, jobID{JobID: jobIDs[serverID], ObjectID: serverID})
			}
			jobErrors := []jobError{}
			for serverID, msg := range reply.preValidationFailed {
				jobErrors = append(jobErrors, jobError{Error: msg, ObjectID: serverID})
			}

			response = map[string]any{"data": map[string]any{"result": map[string]any{
				"jobIds": jobs,
				"errors": jobErrors,
			}}}
		case strings.Contains(payload.Query, "getKorgTaskchainStatus"):
			polls.Add(1)
			if onPoll != nil {
				onPoll()
			}
			response = map[string]any{"data": map[string]any{"getKorgTaskchainStatus": map[string]any{
				"taskchain": map[string]any{"id": 1, "state": reply.taskChainState},
			}}}
		default:
			cancel(errors.New("unexpected query: " + payload.Query))
			return
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			cancel(err)
		}
	}), &polls
}

// TestSetupSQLManagedInstanceBackup verifies that the happy path waits for every
// task chain and reports success.
func TestSetupSQLManagedInstanceBackup(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	serverIDs := []uuid.UUID{uuid.New(), uuid.New()}
	h, polls := sqlServerHandler(cancel, setupReply{
		started:        serverIDs,
		taskChainState: core.TaskChainSucceeded,
	}, nil)
	srv := httptest.NewServer(h)
	defer srv.Close()

	err := WrapGQL(graphql.NewTestClient(srv)).SetupSQLManagedInstanceBackup(ctx, serverIDs, testCredentials,
		time.Millisecond)
	if err != nil {
		t.Errorf("SetupSQLManagedInstanceBackup failed: %v", err)
	}
	if got := polls.Load(); got != 2 {
		t.Errorf("task chain polls: got %d, want 2, one per server", got)
	}
}

// TestSetupSQLManagedInstanceBackupUnacknowledgedServer verifies that a server
// RSC returns in neither jobIds nor errors is reported rather than being
// silently treated as successful.
func TestSetupSQLManagedInstanceBackupUnacknowledgedServer(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	acknowledged, dropped1, dropped2 := uuid.New(), uuid.New(), uuid.New()
	serverIDs := []uuid.UUID{acknowledged, dropped1, dropped2}

	// RSC acknowledges one of the three servers, and its task chain succeeds.
	h, _ := sqlServerHandler(cancel, setupReply{
		started:        []uuid.UUID{acknowledged},
		taskChainState: core.TaskChainSucceeded,
	}, nil)
	srv := httptest.NewServer(h)
	defer srv.Close()

	err := WrapGQL(graphql.NewTestClient(srv)).SetupSQLManagedInstanceBackup(ctx, serverIDs, testCredentials,
		time.Millisecond)
	if err == nil {
		t.Fatal("expected SetupSQLManagedInstanceBackup to fail for the servers RSC did not acknowledge")
	}
	for _, serverID := range []uuid.UUID{dropped1, dropped2} {
		if !strings.Contains(err.Error(), serverID.String()) {
			t.Errorf("error does not mention dropped server %s: %v", serverID, err)
		}
	}
	if strings.Contains(err.Error(), acknowledged.String()) {
		t.Errorf("error mentions the server that succeeded %s: %v", acknowledged, err)
	}
}

// TestSetupSQLManagedInstanceBackupNoServersAcknowledged verifies that an empty
// reply reports every requested server rather than a single generic error.
func TestSetupSQLManagedInstanceBackupNoServersAcknowledged(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	serverIDs := []uuid.UUID{uuid.New(), uuid.New()}
	h, polls := sqlServerHandler(cancel, setupReply{}, nil)
	srv := httptest.NewServer(h)
	defer srv.Close()

	err := WrapGQL(graphql.NewTestClient(srv)).SetupSQLManagedInstanceBackup(ctx, serverIDs, testCredentials,
		time.Millisecond)
	if err == nil {
		t.Fatal("expected SetupSQLManagedInstanceBackup to fail when no jobs were started")
	}
	for _, serverID := range serverIDs {
		if !strings.Contains(err.Error(), serverID.String()) {
			t.Errorf("error does not mention server %s: %v", serverID, err)
		}
	}
	if got := polls.Load(); got != 0 {
		t.Errorf("task chain polls: got %d, want 0", got)
	}
}

// TestSetupSQLManagedInstanceBackupPreValidationAndTaskChain verifies that a
// pre-validation failure does not abandon the task chains already started for
// the other servers, and that both failures are reported together.
func TestSetupSQLManagedInstanceBackupPreValidationAndTaskChain(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	rejected, started := uuid.New(), uuid.New()
	serverIDs := []uuid.UUID{rejected, started}

	h, polls := sqlServerHandler(cancel, setupReply{
		started:             []uuid.UUID{started},
		preValidationFailed: map[uuid.UUID]string{rejected: "invalid credentials"},
		taskChainState:      core.TaskChainFailed,
	}, nil)
	srv := httptest.NewServer(h)
	defer srv.Close()

	err := WrapGQL(graphql.NewTestClient(srv)).SetupSQLManagedInstanceBackup(ctx, serverIDs, testCredentials,
		time.Millisecond)
	if err == nil {
		t.Fatal("expected SetupSQLManagedInstanceBackup to fail")
	}

	// The task chain must have been waited for despite the pre-validation
	// failure, otherwise it would have been left running unwaited.
	if got := polls.Load(); got == 0 {
		t.Error("the started task chain was not waited for")
	}
	if !strings.Contains(err.Error(), "invalid credentials") {
		t.Errorf("error does not report the pre-validation failure: %v", err)
	}
	if !strings.Contains(err.Error(), string(core.TaskChainFailed)) {
		t.Errorf("error does not report the failed task chain: %v", err)
	}
	for _, serverID := range serverIDs {
		if !strings.Contains(err.Error(), serverID.String()) {
			t.Errorf("error does not mention server %s: %v", serverID, err)
		}
	}
}

// TestSetupSQLManagedInstanceBackupCancelled verifies that cancelling the
// context stops the wait instead of producing one failure per remaining server,
// and that the cause remains detectable with errors.Is.
func TestSetupSQLManagedInstanceBackupCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// The context is cancelled from the first task chain poll, leaving the
	// remaining servers' task chains outstanding.
	serverIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	h, polls := sqlServerHandler(func(error) {}, setupReply{
		started:        serverIDs,
		taskChainState: core.TaskChainRunning,
	}, cancel)
	srv := httptest.NewServer(h)
	defer srv.Close()

	err := WrapGQL(graphql.NewTestClient(srv)).SetupSQLManagedInstanceBackup(ctx, serverIDs, testCredentials,
		time.Millisecond)
	if err == nil {
		t.Fatal("expected SetupSQLManagedInstanceBackup to fail after cancellation")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("cancellation is not detectable with errors.Is: %v", err)
	}

	// One server's wait fails, then the loop stops. Without the check the
	// remaining servers would each add a near-identical failure.
	if got := len(strings.Split(err.Error(), "\n")); got >= len(serverIDs) {
		t.Errorf("got %d errors for %d servers, want the wait to stop on cancellation: %v",
			got, len(serverIDs), err)
	}
	if got := polls.Load(); got > int64(len(serverIDs)) {
		t.Errorf("task chain polls: got %d, want the polling to stop on cancellation", got)
	}
}

// addCredentialsHandler stubs the add credentials mutation, reporting the given
// servers as succeeded and failed, and captures the variables sent.
func addCredentialsHandler(cancel context.CancelCauseFunc, succeeded, failed []uuid.UUID, gotVars *map[string]any) http.Handler {
	return handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			cancel(err)
			return
		}
		var payload struct {
			Variables map[string]any `json:"variables"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			cancel(err)
			return
		}
		*gotVars = payload.Variables

		response := map[string]any{"data": map[string]any{"result": map[string]any{
			"successObjectIds": succeeded,
			"failedObjectIds":  failed,
		}}}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			cancel(err)
		}
	})
}

// TestAddSQLManagedInstanceBackupCredentials verifies the happy path sends the
// SQL Server credentials against the managed instance database workload.
func TestAddSQLManagedInstanceBackupCredentials(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	serverID := uuid.New()

	var gotVars map[string]any
	srv := httptest.NewServer(addCredentialsHandler(cancel, []uuid.UUID{serverID}, nil, &gotVars))
	defer srv.Close()

	if err := WrapGQL(graphql.NewTestClient(srv)).AddSQLManagedInstanceBackupCredentials(ctx,
		[]uuid.UUID{serverID}, testCredentials); err != nil {
		t.Fatalf("AddSQLManagedInstanceBackupCredentials failed: %v", err)
	}

	if got := gotVars["workloadType"]; got != "AzureSqlManagedInstanceDb" {
		t.Errorf("workloadType: got %v, want AzureSqlManagedInstanceDb", got)
	}
	if shouldUseAad, _ := gotVars["shouldUseAad"].(bool); shouldUseAad {
		t.Error("shouldUseAad: got true, want false when SQL Server credentials are given")
	}
	if _, ok := gotVars["backupCredentials"]; !ok {
		t.Error("backupCredentials: not sent, want the given credentials")
	}
}

// TestAddSQLManagedInstanceBackupCredentialsUsingEntraID verifies the Entra ID
// variant sends no credentials at all.
func TestAddSQLManagedInstanceBackupCredentialsUsingEntraID(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	serverID := uuid.New()

	var gotVars map[string]any
	srv := httptest.NewServer(addCredentialsHandler(cancel, []uuid.UUID{serverID}, nil, &gotVars))
	defer srv.Close()

	if err := WrapGQL(graphql.NewTestClient(srv)).AddSQLManagedInstanceBackupCredentialsUsingEntraID(ctx,
		[]uuid.UUID{serverID}); err != nil {
		t.Fatalf("AddSQLManagedInstanceBackupCredentialsUsingEntraID failed: %v", err)
	}

	if shouldUseAad, _ := gotVars["shouldUseAad"].(bool); !shouldUseAad {
		t.Error("shouldUseAad: got false, want true when no credentials are given")
	}
	if credentials, ok := gotVars["backupCredentials"]; ok {
		t.Errorf("backupCredentials: got %v, want the field to be absent", credentials)
	}
}

// TestAddSQLManagedInstanceBackupCredentialsFailedServer verifies that a server
// RSC reports in failedObjectIds is an error, so a rejected server cannot pass
// as success.
func TestAddSQLManagedInstanceBackupCredentialsFailedServer(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	succeeded, rejected := uuid.New(), uuid.New()

	var gotVars map[string]any
	srv := httptest.NewServer(addCredentialsHandler(cancel, []uuid.UUID{succeeded}, []uuid.UUID{rejected}, &gotVars))
	defer srv.Close()

	err := WrapGQL(graphql.NewTestClient(srv)).AddSQLManagedInstanceBackupCredentials(ctx,
		[]uuid.UUID{succeeded, rejected}, testCredentials)
	if err == nil {
		t.Fatal("expected AddSQLManagedInstanceBackupCredentials to fail for the rejected server")
	}
	if !strings.Contains(err.Error(), rejected.String()) {
		t.Errorf("error does not name the rejected server %s: %v", rejected, err)
	}
}

// TestAddSQLManagedInstanceBackupCredentialsUnacknowledgedServer verifies that a
// server RSC lists in neither reply field is reported, rather than passing as
// success for a server whose credentials were never registered.
func TestAddSQLManagedInstanceBackupCredentialsUnacknowledgedServer(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	acknowledged, dropped := uuid.New(), uuid.New()

	var gotVars map[string]any
	srv := httptest.NewServer(addCredentialsHandler(cancel, []uuid.UUID{acknowledged}, nil, &gotVars))
	defer srv.Close()

	err := WrapGQL(graphql.NewTestClient(srv)).AddSQLManagedInstanceBackupCredentials(ctx,
		[]uuid.UUID{acknowledged, dropped}, testCredentials)
	if err == nil {
		t.Fatal("expected AddSQLManagedInstanceBackupCredentials to fail for the server RSC did not acknowledge")
	}
	if !strings.Contains(err.Error(), dropped.String()) {
		t.Errorf("error does not name the dropped server %s: %v", dropped, err)
	}
	if strings.Contains(err.Error(), acknowledged.String()) {
		t.Errorf("error names the server that succeeded %s: %v", acknowledged, err)
	}
}
