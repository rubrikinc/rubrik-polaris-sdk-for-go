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

package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
)

func TestKorgTaskChainStatus(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/korg_taskchain_status_response.json")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	// Respond with status code 200 and a valid body.
	srv := httptest.NewTLSServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			cancel(err)
			return
		}

		var payload struct {
			Query     string `json:"query"`
			Variables struct {
				TaskChainID string `json:"taskchainId,omitempty"`
			} `json:"variables,omitempty"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			cancel(err)
			return
		}

		if err := tmpl.Execute(w, struct {
			ChainState string
			ChainUUID  string
		}{ChainState: "RUNNING", ChainUUID: payload.Variables.TaskChainID}); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	id := uuid.MustParse("b48e7ad0-7b86-4c96-b6ba-97eb6a82f765")
	taskChain, err := Wrap(graphql.NewTestClient(srv)).KorgTaskChainStatus(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if taskChain.ID != 12761540 {
		t.Errorf("invalid task chain id: %v", taskChain.ID)
	}
	if taskChain.TaskChainID != id {
		t.Errorf("invalid task chain uuid: %v", taskChain.TaskChainID)
	}
	if taskChain.State != TaskChainRunning {
		t.Errorf("invalid task chain state: %v", taskChain.State)
	}
}

func TestWaitForTaskChain(t *testing.T) {
	tmpl, err := template.ParseFiles("testdata/korg_taskchain_status_response.json")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	// Respond with status code 200 and a valid body. The First 2 responses have
	// state RUNNING. The Third response is SUCCEEDED.
	reqCount := 3
	srv := httptest.NewTLSServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			cancel(err)
			return
		}

		var payload struct {
			Query     string `json:"query"`
			Variables struct {
				TaskChainID string `json:"taskchainId,omitempty"`
			} `json:"variables,omitempty"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			cancel(err)
			return
		}

		reqCount--
		chainState := "RUNNING"
		if reqCount == 0 {
			chainState = "SUCCEEDED"
		}
		if err := tmpl.Execute(w, struct {
			ChainState string
			ChainUUID  string
		}{ChainState: chainState, ChainUUID: payload.Variables.TaskChainID}); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	id := uuid.MustParse("b48e7ad0-7b86-4c96-b6ba-97eb6a82f765")
	state, err := Wrap(graphql.NewTestClient(srv)).WaitForTaskChain(ctx, id, 5*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if state != "SUCCEEDED" {
		t.Errorf("invalid task chain state: %v", state)
	}
}

// TestWaitForTaskChainWithStatus verifies that WaitForTaskChainWithStatus
// returns when the status function is done, also when it's done just as the
// task chain ends, and that it returns an error when the task chain ends
// before the status function is done. It also verifies that 403 errors from
// the task chain status are tolerated, that other errors are returned and that
// the wait ends when the context is done.
func TestWaitForTaskChainWithStatus(t *testing.T) {
	id := uuid.MustParse("b48e7ad0-7b86-4c96-b6ba-97eb6a82f765")
	errStatus := errors.New("status failed")

	taskChain := func(state string) string {
		return fmt.Sprintf(`{"data": {"getKorgTaskchainStatus": {"taskchain": {"id": 1, "state": %q, "taskchainUuid": %q}}}}`,
			state, id)
	}
	gqlError := func(code int) string {
		return fmt.Sprintf(`{"data": null, "errors": [{"message": "error", "extensions": {"code": %d}}]}`, code)
	}

	testCases := []struct {
		name string
		// Replies to the task chain status requests. The last reply is
		// repeated.
		taskChains []string
		// The status function returns done on this call, counting from 1.
		// Zero means never.
		doneAt int
		// The status function fails on this call, counting from 1. Zero
		// means never.
		failAt        int
		expectErr     error
		expectErrText string
		expectCalls   int
	}{{
		name:        "DoneWhileRunning",
		taskChains:  []string{taskChain("RUNNING")},
		doneAt:      3,
		expectCalls: 3,
	}, {
		name:        "DoneWhenTaskChainEnds",
		taskChains:  []string{taskChain("RUNNING"), taskChain("SUCCEEDED")},
		doneAt:      3,
		expectCalls: 3,
	}, {
		name:          "TaskChainSucceededNotDone",
		taskChains:    []string{taskChain("SUCCEEDED")},
		expectErrText: "ended with state SUCCEEDED",
		expectCalls:   2,
	}, {
		name:          "TaskChainCanceledNotDone",
		taskChains:    []string{taskChain("RUNNING"), taskChain("CANCELED")},
		expectErrText: "ended with state CANCELED",
		expectCalls:   3,
	}, {
		name:        "TaskChainNotReady",
		taskChains:  []string{taskChain(""), taskChain("READY"), taskChain("RUNNING")},
		doneAt:      4,
		expectCalls: 4,
	}, {
		name:        "TaskChainNotAuthorized",
		taskChains:  []string{gqlError(403)},
		doneAt:      3,
		expectCalls: 3,
	}, {
		name:          "TaskChainError",
		taskChains:    []string{gqlError(400)},
		expectErrText: "failed to get task chain status",
		expectCalls:   1,
	}, {
		name:        "StatusError",
		taskChains:  []string{taskChain("RUNNING")},
		failAt:      2,
		expectErr:   errStatus,
		expectCalls: 2,
	}, {
		name:       "ContextDone",
		taskChains: []string{taskChain("RUNNING")},
		expectErr:  context.DeadlineExceeded,
	}}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var requests int
			srv := httptest.NewTLSServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
				buf, err := io.ReadAll(req.Body)
				if err != nil {
					t.Error(err)
					return
				}
				if !strings.Contains(string(buf), "getKorgTaskchainStatus") {
					t.Errorf("unexpected request: %s", buf)
				}
				reply := tc.taskChains[min(requests, len(tc.taskChains)-1)]
				requests++
				if _, err := io.WriteString(w, reply); err != nil {
					t.Error(err)
				}
			}))
			defer srv.Close()

			ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
			defer cancel()
			var calls int
			err := Wrap(graphql.NewTestClient(srv)).WaitForTaskChainWithStatus(ctx, id, time.Millisecond, func(ctx context.Context) (bool, error) {
				calls++
				if calls == tc.failAt {
					return false, errStatus
				}
				return calls == tc.doneAt, nil
			})
			switch {
			case tc.expectErr != nil:
				if !errors.Is(err, tc.expectErr) {
					t.Fatalf("invalid error: %v, expected: %v", err, tc.expectErr)
				}
			case tc.expectErrText != "":
				if err == nil || !strings.Contains(err.Error(), tc.expectErrText) {
					t.Fatalf("invalid error: %v, expected error containing: %q", err, tc.expectErrText)
				}
			case err != nil:
				t.Fatal(err)
			}
			if tc.expectCalls != 0 && calls != tc.expectCalls {
				t.Errorf("invalid number of status calls: %d, expected: %d", calls, tc.expectCalls)
			}
		})
	}
}
