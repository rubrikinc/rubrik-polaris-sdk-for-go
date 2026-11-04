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

package exocompute

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
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/exocompute"
)

// TestTriggerHealthCheck verifies that TriggerHealthCheck sends the cloud
// vendor and configuration ID to RSC and returns the health check job ID, and
// that a reply without a valid job ID is reported as an error.
func TestTriggerHealthCheck(t *testing.T) {
	configID := uuid.MustParse("5f1b8c0e-4a1d-4c53-9d57-1d1f4a3f2b10")
	awsJobID := uuid.MustParse("01a1a6a9-c79b-72ae-af78-4b3fbf8f06ad")
	azureJobID := uuid.MustParse("01a0621d-343f-728d-a013-c9c17782aa38")

	testCases := []struct {
		name        string
		cloudVendor core.CloudVendor
		reply       string
		expectJobID uuid.UUID
		expectErr   bool
	}{{
		name:        "AWS",
		cloudVendor: core.CloudVendorAWS,
		reply:       fmt.Sprintf(`{"data": {"result": {"healthCheckJobId": %q}}}`, awsJobID),
		expectJobID: awsJobID,
	}, {
		name:        "Azure",
		cloudVendor: core.CloudVendorAzure,
		reply:       fmt.Sprintf(`{"data": {"result": {"healthCheckJobId": %q}}}`, azureJobID),
		expectJobID: azureJobID,
	}, {
		name:        "NoJobID",
		cloudVendor: core.CloudVendorAzure,
		reply:       `{"data": {"result": {"healthCheckJobId": ""}}}`,
		expectErr:   true,
	}, {
		name:        "InvalidJobID",
		cloudVendor: core.CloudVendorAWS,
		reply:       `{"data": {"result": {"healthCheckJobId": "job-1"}}}`,
		expectErr:   true,
	}}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewTLSServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
				var body struct {
					Query     string `json:"query"`
					Variables struct {
						CloudVendor core.CloudVendor `json:"cloudVendor"`
						ConfigID    uuid.UUID        `json:"exocomputeConfigId"`
					} `json:"variables"`
				}
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
					t.Error(err)
					return
				}
				if !strings.Contains(body.Query, "triggerExocomputeHealthCheck") {
					t.Errorf("unexpected request: %s", body.Query)
				}
				if body.Variables.CloudVendor != tc.cloudVendor {
					t.Errorf("invalid cloud vendor: %s, expected: %s", body.Variables.CloudVendor, tc.cloudVendor)
				}
				if body.Variables.ConfigID != configID {
					t.Errorf("invalid configuration id: %s, expected: %s", body.Variables.ConfigID, configID)
				}
				if _, err := io.WriteString(w, tc.reply); err != nil {
					t.Error(err)
				}
			}))
			defer srv.Close()

			jobID, err := Wrap(&polaris.Client{GQL: graphql.NewTestClient(srv)}).TriggerHealthCheck(
				t.Context(), tc.cloudVendor, configID)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected trigger health check call to fail")
				}
				if jobID != uuid.Nil {
					t.Errorf("invalid job id: %s, expected: %s", jobID, uuid.Nil)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if jobID != tc.expectJobID {
				t.Errorf("invalid job id: %s, expected: %s", jobID, tc.expectJobID)
			}
		})
	}
}

// TestWaitForHealthCheck verifies that WaitForHealthCheck ignores health check
// statuses produced by other jobs, keeps waiting while the job is verifying,
// and returns the final status of the job. It also verifies that a status
// which never matches the job, or stays verifying, waits until the context is
// done, that the wait ends when the job's task chain fails, and that the wait
// continues when the task chain state isn't available.
func TestWaitForHealthCheck(t *testing.T) {
	origInterval := healthCheckPollInterval
	healthCheckPollInterval = time.Millisecond
	t.Cleanup(func() { healthCheckPollInterval = origInterval })

	configID := uuid.MustParse("5f1b8c0e-4a1d-4c53-9d57-1d1f4a3f2b10")
	jobID := uuid.MustParse("01a1a6a9-c79b-72ae-af78-4b3fbf8f06ad")
	otherJobID := uuid.MustParse("01a0621d-343f-728d-a013-c9c17782aa38")

	status := func(status string, taskchainID any) string {
		return fmt.Sprintf(`{"status": %q, "failureReason": "reason", "lastUpdatedAt": "", "taskchainId": "%v"}`,
			status, taskchainID)
	}
	taskChain := func(state string) string {
		return fmt.Sprintf(`{"data": {"getKorgTaskchainStatus": {"taskchain": {"id": 1, "state": %q, "taskchainUuid": %q}}}}`,
			state, jobID)
	}
	const notAuthorized = `{"data": null, "errors": [{"message": "Objects are not authorized", "extensions": {"code": 403}}]}`

	testCases := []struct {
		name          string
		cloudVendor   core.CloudVendor
		statuses      []string
		taskChain     string
		expectStatus  exocompute.HealthCheckStatusValue
		expectErr     error
		expectErrText string
	}{{
		name:        "AWSHealthy",
		cloudVendor: core.CloudVendorAWS,
		statuses: []string{
			status("UNHEALTHY", otherJobID),
			status("VERIFYING", jobID),
			status("HEALTHY", jobID),
		},
		taskChain:    taskChain("RUNNING"),
		expectStatus: exocompute.HealthCheckStatusHealthy,
	}, {
		name:        "AzureUnhealthy",
		cloudVendor: core.CloudVendorAzure,
		statuses: []string{
			status("HEALTHY", otherJobID),
			status("VERIFYING", jobID),
			status("UNHEALTHY", jobID),
		},
		taskChain:    taskChain("RUNNING"),
		expectStatus: exocompute.HealthCheckStatusUnhealthy,
	}, {
		name:        "AzureNoStatusReported",
		cloudVendor: core.CloudVendorAzure,
		statuses: []string{
			status("", ""),
		},
		taskChain: taskChain("RUNNING"),
		expectErr: context.DeadlineExceeded,
	}, {
		name:        "AWSStillVerifying",
		cloudVendor: core.CloudVendorAWS,
		statuses: []string{
			status("VERIFYING", jobID),
		},
		taskChain:    taskChain("RUNNING"),
		expectStatus: exocompute.HealthCheckStatusVerifying,
		expectErr:    context.DeadlineExceeded,
	}, {
		name:        "AzureTaskChainFailed",
		cloudVendor: core.CloudVendorAzure,
		statuses: []string{
			status("VERIFYING", jobID),
		},
		taskChain:     taskChain("FAILED"),
		expectStatus:  exocompute.HealthCheckStatusVerifying,
		expectErrText: "ended with state FAILED",
	}, {
		name:        "AWSTaskChainNotAuthorized",
		cloudVendor: core.CloudVendorAWS,
		statuses: []string{
			status("VERIFYING", jobID),
			status("VERIFYING", jobID),
			status("HEALTHY", jobID),
		},
		taskChain:    notAuthorized,
		expectStatus: exocompute.HealthCheckStatusHealthy,
	}}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var polls int
			srv := httptest.NewTLSServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Error(err)
					return
				}

				switch {
				case strings.Contains(string(body), "getKorgTaskchainStatus"):
					_, err = io.WriteString(w, tc.taskChain)
				case strings.Contains(string(body), "allAwsExocomputeConfigs"):
					// Repeat the last status once all statuses have been
					// returned.
					healthCheckStatus := tc.statuses[min(polls, len(tc.statuses)-1)]
					polls++
					_, err = fmt.Fprintf(w, `{"data": {"result": [{"awsCloudAccount": {"id": %q}, "exocomputeConfigs": [{"configUuid": %q, "healthCheckStatus": %s}]}]}}`,
						uuid.New(), configID, healthCheckStatus)
				case strings.Contains(string(body), "allAzureExocomputeConfigsInAccount"):
					healthCheckStatus := tc.statuses[min(polls, len(tc.statuses)-1)]
					polls++
					_, err = fmt.Fprintf(w, `{"data": {"result": [{"azureCloudAccount": {"id": %q}, "configs": [{"configUuid": %q, "healthCheckStatus": %s}]}]}}`,
						uuid.New(), configID, healthCheckStatus)
				default:
					t.Errorf("unexpected request: %s", body)
				}
				if err != nil {
					t.Error(err)
				}
			}))
			defer srv.Close()

			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			defer cancel()
			result, err := Wrap(&polaris.Client{GQL: graphql.NewTestClient(srv)}).WaitForHealthCheck(
				ctx, tc.cloudVendor, configID, jobID)
			switch {
			case tc.expectErr != nil:
				if !errors.Is(err, tc.expectErr) {
					t.Fatalf("invalid error: %v, expected: %v", err, tc.expectErr)
				}
			case tc.expectErrText != "":
				if err == nil || !strings.Contains(err.Error(), tc.expectErrText) {
					t.Fatalf("invalid error: %v, expected error containing: %q", err, tc.expectErrText)
				}
				if errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("invalid error: %v, expected the wait to end before the context", err)
				}
			case err != nil:
				t.Fatal(err)
			}
			if result.Status != tc.expectStatus {
				t.Errorf("invalid status: %q, expected: %q", result.Status, tc.expectStatus)
			}
			if tc.expectStatus != "" && result.TaskchainID != jobID.String() {
				t.Errorf("invalid taskchain id: %q, expected: %q", result.TaskchainID, jobID)
			}
			if tc.expectErr == nil && tc.expectErrText == "" && polls != len(tc.statuses) {
				t.Errorf("invalid number of polls: %d, expected: %d", polls, len(tc.statuses))
			}
		})
	}
}
