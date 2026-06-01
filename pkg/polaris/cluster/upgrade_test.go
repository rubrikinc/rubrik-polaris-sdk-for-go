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

package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlcluster "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/cluster"
)

// pageResponse returns a JSON response with the given details, cursor, and
// hasNextPage flag.
func pageResponse(cursor string, hasNextPage bool, nodes ...string) string {
	edges := ""
	for i, node := range nodes {
		if i > 0 {
			edges += ","
		}
		edges += fmt.Sprintf(`{"cursor":"c","node":%s}`, node)
	}
	return fmt.Sprintf(`{
        "data": {
            "result": {
                "edges": [%s],
                "pageInfo": {"startCursor":"","endCursor":%q,"hasPreviousPage":false,"hasNextPage":%t},
                "count": %d
            }
        }
    }`, edges, cursor, hasNextPage, len(nodes))
}

func TestListClusterUpgradesPaginates(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	id1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	id2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	id3 := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	node := func(id uuid.UUID, name string) string {
		return fmt.Sprintf(`{"id":%q,"name":%q,"cdmUpgradeInfo":{"clusterUuid":%q,"version":"9.2.1"}}`, id, name, id)
	}

	calls := 0
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Variables map[string]json.RawMessage `json:"variables"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		calls++
		switch calls {
		case 1:
			if _, ok := body.Variables["after"]; ok {
				http.Error(w, "first call should have no after cursor", http.StatusBadRequest)
				return
			}
			fmt.Fprint(w, pageResponse("cursor-1", true, node(id1, "a"), node(id2, "b")))
		case 2:
			var after string
			if err := json.Unmarshal(body.Variables["after"], &after); err != nil || after != "cursor-1" {
				http.Error(w, "second call should send cursor-1", http.StatusBadRequest)
				return
			}
			fmt.Fprint(w, pageResponse("cursor-2", false, node(id3, "c")))
		default:
			http.Error(w, "too many calls", http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	api := Wrap(newTestClient(srv))
	details, err := api.ListClusterUpgrades(ctx, nil, "", "")
	if err != nil {
		t.Fatalf("ListClusterUpgrades: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
	if len(details) != 3 {
		t.Fatalf("expected 3 details, got %d", len(details))
	}
	wantIDs := []uuid.UUID{id1, id2, id3}
	for i, d := range details {
		if d.ID != wantIDs[i] {
			t.Errorf("details[%d].ID: got %v, want %v", i, d.ID, wantIDs[i])
		}
	}
}

// TestListClusterUpgradesNonAdvancingCursor verifies that a server returning
// HasNextPage=true with the same EndCursor does not cause an infinite loop.
func TestListClusterUpgradesNonAdvancingCursor(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	id1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	node := fmt.Sprintf(`{"id":%q,"name":"a","cdmUpgradeInfo":{"clusterUuid":%q,"version":"9.2.1"}}`, id1, id1)

	calls := 0
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		calls++
		// Always return the same cursor with HasNextPage=true to simulate a
		// buggy server. The client must give up rather than loop.
		fmt.Fprint(w, pageResponse("stuck-cursor", true, node))
	}))
	defer srv.Close()

	api := Wrap(newTestClient(srv))
	details, err := api.ListClusterUpgrades(ctx, nil, "", "")
	if err != nil {
		t.Fatalf("ListClusterUpgrades: %v", err)
	}
	// Expect two calls: first page (no After), then a second call with
	// After=stuck-cursor that returns the same cursor — at that point we bail.
	if calls != 2 {
		t.Fatalf("expected 2 calls before bailing, got %d", calls)
	}
	if len(details) != 2 {
		t.Fatalf("expected 2 details (one per call before bail), got %d", len(details))
	}
}

func TestClusterUpgradeNotFound(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, `{"data":{"result":{"edges":[],"pageInfo":{"hasNextPage":false},"count":0}}}`)
	}))
	defer srv.Close()

	api := Wrap(newTestClient(srv))
	_, err := api.ClusterUpgrade(ctx, uuid.MustParse("11111111-1111-1111-1111-111111111111"))
	if err == nil {
		t.Fatal("expected ErrNotFound")
	}
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Errorf("error should wrap ErrNotFound: %v", err)
	}
}

func TestClusterUpgradeWrongIDIsNotFound(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	wantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	otherID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		// Server returns a different cluster than the one we asked for.
		node := fmt.Sprintf(`{"id":%q,"name":"other","cdmUpgradeInfo":{"clusterUuid":%q,"version":"9.2.1"}}`, otherID, otherID)
		fmt.Fprint(w, pageResponse("c1", false, node))
	}))
	defer srv.Close()

	api := Wrap(newTestClient(srv))
	_, err := api.ClusterUpgrade(ctx, wantID)
	if err == nil {
		t.Fatal("expected ErrNotFound")
	}
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Errorf("error should wrap ErrNotFound: %v", err)
	}
}

func TestClusterUpgradeFound(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	wantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		node := fmt.Sprintf(`{"id":%q,"name":"cluster-a","cdmUpgradeInfo":{"clusterUuid":%q,"version":"9.2.1"}}`, wantID, wantID)
		fmt.Fprint(w, pageResponse("c1", false, node))
	}))
	defer srv.Close()

	api := Wrap(newTestClient(srv))
	got, err := api.ClusterUpgrade(ctx, wantID)
	if err != nil {
		t.Fatalf("ClusterUpgrade: %v", err)
	}
	if got.ID != wantID {
		t.Errorf("ID: got %v, want %v", got.ID, wantID)
	}
	if got.CDMInfo == nil || got.CDMInfo.Version != "9.2.1" {
		t.Errorf("CDMInfo: got %+v", got.CDMInfo)
	}
}

// shortenPollDelays sets the WaitForDownload poll intervals to a few
// milliseconds so polling-loop tests finish quickly.
func shortenPollDelays(t *testing.T) {
	t.Helper()
	origInitial, origMax := upgradePollInitialDelay, upgradePollMaxDelay
	upgradePollInitialDelay = 5 * time.Millisecond
	upgradePollMaxDelay = 10 * time.Millisecond
	t.Cleanup(func() {
		upgradePollInitialDelay = origInitial
		upgradePollMaxDelay = origMax
	})
}

// upgradeNode renders a clusterWithUpgradesInfo node with the given V2 status
// and (optional) target version in UIStatusAttributes.
func upgradeNode(id uuid.UUID, v2 gqlcluster.RSCUpgradeStatusType, targetVersion string) string {
	return fmt.Sprintf(`{"id":%q,"name":"c","cdmUpgradeInfo":{"clusterUuid":%q,"version":"9.2.1","upgradeStatusV2":{"rscClusterUpgradeStatus":%q,"uiStatus":"","uiStatusAttributes":{"sourceVersion":"","targetVersion":%q,"progress":0,"errorMsg":"","upgradeMode":""}}}}`, id, id, v2, targetVersion)
}

// scriptedGraphQL serves canned responses for each operation name. Each entry
// is consumed in order; a missing entry / unknown op returns 500. Used to
// script a sequence of polling responses.
type scriptedGraphQL struct {
	t         *testing.T
	responses map[string][]string
	calls     map[string]int
}

func newScriptedGraphQL(t *testing.T) *scriptedGraphQL {
	return &scriptedGraphQL{t: t, responses: map[string][]string{}, calls: map[string]int{}}
}

func (s *scriptedGraphQL) script(op string, responses ...string) {
	s.responses[op] = append(s.responses[op], responses...)
}

func (s *scriptedGraphQL) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			OperationName string `json:"operationName"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.calls[body.OperationName]++
		queue := s.responses[body.OperationName]
		if len(queue) == 0 {
			s.t.Errorf("no scripted response for op %q (call #%d)", body.OperationName, s.calls[body.OperationName])
			http.Error(w, "no scripted response", http.StatusInternalServerError)
			return
		}
		// Repeat the last scripted response for subsequent calls.
		next := queue[0]
		if len(queue) > 1 {
			s.responses[body.OperationName] = queue[1:]
		}
		fmt.Fprint(w, next)
	}
}

func clusterUpgradeResp(node string) string {
	return fmt.Sprintf(`{"data":{"result":{"edges":[{"cursor":"c","node":%s}],"pageInfo":{"startCursor":"","endCursor":"c","hasPreviousPage":false,"hasNextPage":false},"count":1}}}`, node)
}

func TestWaitForDownloadStagesSuccessfully(t *testing.T) {
	shortenPollDelays(t)
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	target := "9.3.3-p8-29908"

	s := newScriptedGraphQL(t)
	s.script("SdkGolangClusterWithUpgradesInfo",
		clusterUpgradeResp(upgradeNode(id, gqlcluster.RSCUpgradeStatusDownloading, target)),
		clusterUpgradeResp(upgradeNode(id, gqlcluster.RSCUpgradeStatusDownloading, target)),
		clusterUpgradeResp(upgradeNode(id, gqlcluster.RSCUpgradeStatusReadyForUpgrade, target)),
	)

	srv := httptest.NewServer(handler.GraphQL(s.handler()))
	defer srv.Close()

	info, err := Wrap(newTestClient(srv)).WaitForDownload(ctx, id, target)
	if err != nil {
		t.Fatalf("WaitForDownload: %v", err)
	}
	if info.UpgradeStatusV2 == nil || info.UpgradeStatusV2.RSCClusterUpgradeStatus != gqlcluster.RSCUpgradeStatusReadyForUpgrade {
		t.Errorf("expected final state ReadyForUpgrade, got %+v", info.UpgradeStatusV2)
	}
	if s.calls["SdkGolangClusterWithUpgradesInfo"] != 3 {
		t.Errorf("expected 3 polls, got %d", s.calls["SdkGolangClusterWithUpgradesInfo"])
	}
}

func TestWaitForDownloadTerminalFailure(t *testing.T) {
	shortenPollDelays(t)
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	target := "9.3.3-p8-29908"

	s := newScriptedGraphQL(t)
	s.script("SdkGolangClusterWithUpgradesInfo",
		clusterUpgradeResp(upgradeNode(id, gqlcluster.RSCUpgradeStatusDownloading, target)),
		clusterUpgradeResp(upgradeNode(id, gqlcluster.RSCUpgradeStatusDownloadFailed, target)),
	)

	srv := httptest.NewServer(handler.GraphQL(s.handler()))
	defer srv.Close()

	_, err := Wrap(newTestClient(srv)).WaitForDownload(ctx, id, target)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "DOWNLOAD_FAILED") {
		t.Errorf("error should mention DOWNLOAD_FAILED, got: %v", err)
	}
}

func TestWaitForDownloadReturnsOnPostDownload(t *testing.T) {
	shortenPollDelays(t)
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	target := "9.3.3-p8-29908"

	s := newScriptedGraphQL(t)
	// Server runs prechecks at the tail of download. PRECHECKING means package
	// is on the cluster; WaitForDownload should return without waiting for
	// READY_FOR_UPGRADE.
	s.script("SdkGolangClusterWithUpgradesInfo",
		clusterUpgradeResp(upgradeNode(id, gqlcluster.RSCUpgradeStatusDownloading, target)),
		clusterUpgradeResp(upgradeNode(id, gqlcluster.RSCUpgradeStatusPrechecking, target)),
	)

	srv := httptest.NewServer(handler.GraphQL(s.handler()))
	defer srv.Close()

	info, err := Wrap(newTestClient(srv)).WaitForDownload(ctx, id, target)
	if err != nil {
		t.Fatalf("WaitForDownload: %v", err)
	}
	if info.UpgradeStatusV2.RSCClusterUpgradeStatus != gqlcluster.RSCUpgradeStatusPrechecking {
		t.Errorf("expected PRECHECKING, got %s", info.UpgradeStatusV2.RSCClusterUpgradeStatus)
	}
}

// TestWaitForDownloadIgnoresWaitingForOperationStart is a regression test for
// the bug where WAITING_FOR_OPERATION_TO_START — the pre-progress state right
// after any operation is triggered — was treated as post-download and caused
// WaitForDownload to return immediately.
func TestWaitForDownloadIgnoresWaitingForOperationStart(t *testing.T) {
	shortenPollDelays(t)
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	target := "9.3.3-p8-29908"

	s := newScriptedGraphQL(t)
	s.script("SdkGolangClusterWithUpgradesInfo",
		clusterUpgradeResp(upgradeNode(id, gqlcluster.RSCUpgradeStatusWaitingForOperationStart, "")),
		clusterUpgradeResp(upgradeNode(id, gqlcluster.RSCUpgradeStatusDownloading, target)),
		clusterUpgradeResp(upgradeNode(id, gqlcluster.RSCUpgradeStatusReadyForUpgrade, target)),
	)

	srv := httptest.NewServer(handler.GraphQL(s.handler()))
	defer srv.Close()

	info, err := Wrap(newTestClient(srv)).WaitForDownload(ctx, id, target)
	if err != nil {
		t.Fatalf("WaitForDownload: %v", err)
	}
	if info.UpgradeStatusV2.RSCClusterUpgradeStatus != gqlcluster.RSCUpgradeStatusReadyForUpgrade {
		t.Errorf("expected final state ReadyForUpgrade, got %s", info.UpgradeStatusV2.RSCClusterUpgradeStatus)
	}
	if s.calls["SdkGolangClusterWithUpgradesInfo"] != 3 {
		t.Errorf("expected 3 polls (WAITING_FOR_OPERATION_TO_START must not short-circuit), got %d", s.calls["SdkGolangClusterWithUpgradesInfo"])
	}
}

// TestWaitForDownloadTargetVer is a regression test ensuring a cluster
// sitting in READY_FOR_UPGRADE for a version other than the requested target
// does not short-circuit the wait. Only once the target version is staged
// should WaitForDownload return success.
func TestWaitForDownloadTargetVer(t *testing.T) {
	shortenPollDelays(t)
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	target := "9.3.3-p8-29908"
	stale := "9.2.1-p1-12345"

	s := newScriptedGraphQL(t)
	s.script("SdkGolangClusterWithUpgradesInfo",
		clusterUpgradeResp(upgradeNode(id, gqlcluster.RSCUpgradeStatusReadyForUpgrade, stale)),
		clusterUpgradeResp(upgradeNode(id, gqlcluster.RSCUpgradeStatusDownloading, target)),
		clusterUpgradeResp(upgradeNode(id, gqlcluster.RSCUpgradeStatusReadyForUpgrade, target)),
	)

	srv := httptest.NewServer(handler.GraphQL(s.handler()))
	defer srv.Close()

	info, err := Wrap(newTestClient(srv)).WaitForDownload(ctx, id, target)
	if err != nil {
		t.Fatalf("WaitForDownload: %v", err)
	}
	if info.UpgradeStatusV2.UIStatusAttributes.TargetVersion != target {
		t.Errorf("expected staged target %q, got %q", target, info.UpgradeStatusV2.UIStatusAttributes.TargetVersion)
	}
	if s.calls["SdkGolangClusterWithUpgradesInfo"] != 3 {
		t.Errorf("expected 3 polls (stale READY_FOR_UPGRADE must not short-circuit), got %d", s.calls["SdkGolangClusterWithUpgradesInfo"])
	}
}
