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

package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// deleteCall is one gcpCloudAccountDeleteProjectsV2 mutation as received by
// the stubbed RSC.
type deleteCall struct {
	features        []string
	deleteSnapshots []bool
}

// removeProjectHandler stubs the RSC calls made by RemoveProject. The project
// has the specified features and statuses. Each delete mutation is recorded
// and every task chain reports success. The returned function returns the
// recorded delete mutations.
func removeProjectHandler(cancel context.CancelCauseFunc, cloudAccountID uuid.UUID, projectFeatures map[string]core.Status) (http.Handler, func() []deleteCall) {
	var mu sync.Mutex
	var calls []deleteCall

	return handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				cancel(err)
				return
			}

			var payload struct {
				Query     string `json:"query"`
				Variables struct {
					Features []struct {
						Feature         string `json:"feature"`
						DeleteSnapshots bool   `json:"deleteSnapshots"`
					} `json:"features"`
				} `json:"variables"`
			}
			if err := json.Unmarshal(body, &payload); err != nil {
				cancel(err)
				return
			}

			var response any
			switch {
			case strings.Contains(payload.Query, "allGcpCloudAccountProjectsByFeature"):
				type entry struct {
					Project map[string]any `json:"project"`
					Detail  map[string]any `json:"featureDetail"`
				}
				result := []entry{}
				for name, status := range projectFeatures {
					result = append(result, entry{
						Project: map[string]any{
							"id":            cloudAccountID,
							"name":          "project",
							"projectId":     "project-id",
							"projectNumber": 1234,
						},
						Detail: map[string]any{"feature": name, "status": status},
					})
				}
				response = map[string]any{"data": map[string]any{"result": result}}
			case strings.Contains(payload.Query, "allLatestFeaturePermissionsForCloudAccounts"):
				response = map[string]any{"data": map[string]any{"result": []any{}}}
			case strings.Contains(payload.Query, "gcpNativeProjects"):
				response = map[string]any{"data": map[string]any{"result": map[string]any{
					"count":    0,
					"edges":    []any{},
					"pageInfo": map[string]any{"endCursor": "", "hasNextPage": false},
				}}}
			case strings.Contains(payload.Query, "gcpCloudAccountDeleteProjectsV2"):
				var call deleteCall
				type jobID struct {
					JobID          uuid.UUID `json:"jobId"`
					RubrikObjectID string    `json:"rubrikObjectId"`
				}
				jobs := []jobID{}
				for _, f := range payload.Variables.Features {
					call.features = append(call.features, f.Feature)
					call.deleteSnapshots = append(call.deleteSnapshots, f.DeleteSnapshots)
					jobs = append(jobs, jobID{JobID: uuid.New(), RubrikObjectID: fmt.Sprintf("%s:%s", cloudAccountID, f.Feature)})
				}
				mu.Lock()
				calls = append(calls, call)
				mu.Unlock()
				response = map[string]any{"data": map[string]any{"result": map[string]any{"jobIds": jobs, "errors": []any{}}}}
			case strings.Contains(payload.Query, "getKorgTaskchainStatus"):
				response = map[string]any{"data": map[string]any{"getKorgTaskchainStatus": map[string]any{
					"taskchain": map[string]any{"id": 1, "state": core.TaskChainSucceeded},
				}}}
			default:
				cancel(errors.New("unexpected query: " + payload.Query))
				return
			}

			if err := json.NewEncoder(w).Encode(response); err != nil {
				cancel(err)
			}
		}), func() []deleteCall {
			mu.Lock()
			defer mu.Unlock()
			return calls
		}
}

// TestRemoveProjectCloudDiscovery verifies that RemoveProject removes Cloud
// Discovery last and without deleting snapshots, and that it refuses to remove
// Cloud Discovery while a protection feature stays on the project.
func TestRemoveProjectCloudDiscovery(t *testing.T) {
	tests := []struct {
		name            string
		projectFeatures map[string]core.Status
		remove          []core.Feature
		wantCalls       []deleteCall
		wantErr         string
	}{{
		name: "cloud discovery removed after the protection features it was requested with",
		projectFeatures: map[string]core.Status{
			core.FeatureCloudDiscovery.Name:        core.StatusConnected,
			core.FeatureCloudNativeProtection.Name: core.StatusConnected,
			core.FeatureCloudSQLProtection.Name:    core.StatusConnected,
		},
		// Cloud Discovery is deliberately listed first.
		remove: []core.Feature{core.FeatureCloudDiscovery, core.FeatureCloudNativeProtection, core.FeatureCloudSQLProtection},
		wantCalls: []deleteCall{
			{features: []string{"CLOUD_NATIVE_PROTECTION", "CLOUD_SQL_PROTECTION"}, deleteSnapshots: []bool{true, true}},
			{features: []string{"CLOUD_DISCOVERY"}, deleteSnapshots: []bool{false}},
		},
	}, {
		name: "cloud discovery alone never deletes snapshots",
		projectFeatures: map[string]core.Status{
			core.FeatureCloudDiscovery.Name: core.StatusConnected,
		},
		remove:    []core.Feature{core.FeatureCloudDiscovery},
		wantCalls: []deleteCall{{features: []string{"CLOUD_DISCOVERY"}, deleteSnapshots: []bool{false}}},
	}, {
		name: "features other than cloud discovery are removed in one call",
		projectFeatures: map[string]core.Status{
			core.FeatureCloudDiscovery.Name:        core.StatusConnected,
			core.FeatureCloudNativeProtection.Name: core.StatusConnected,
			core.FeatureCloudSQLProtection.Name:    core.StatusConnected,
		},
		remove:    []core.Feature{core.FeatureCloudNativeProtection, core.FeatureCloudSQLProtection},
		wantCalls: []deleteCall{{features: []string{"CLOUD_NATIVE_PROTECTION", "CLOUD_SQL_PROTECTION"}, deleteSnapshots: []bool{true, true}}},
	}, {
		name: "cloud discovery refused while a protection feature is connected",
		projectFeatures: map[string]core.Status{
			core.FeatureCloudDiscovery.Name:        core.StatusConnected,
			core.FeatureCloudNativeProtection.Name: core.StatusConnected,
		},
		remove:  []core.Feature{core.FeatureCloudDiscovery},
		wantErr: "cloud discovery must be removed after all other protection features, CLOUD_NATIVE_PROTECTION is enabled",
	}, {
		name: "cloud discovery refused while a protection feature is missing permissions",
		projectFeatures: map[string]core.Status{
			core.FeatureCloudDiscovery.Name:     core.StatusConnected,
			core.FeatureCloudSQLProtection.Name: core.StatusMissingPermissions,
		},
		remove:  []core.Feature{core.FeatureCloudDiscovery},
		wantErr: "CLOUD_SQL_PROTECTION is enabled",
	}, {
		name: "cloud discovery allowed when the remaining protection feature is disconnected",
		projectFeatures: map[string]core.Status{
			core.FeatureCloudDiscovery.Name:        core.StatusConnected,
			core.FeatureCloudNativeProtection.Name: core.StatusDisconnected,
		},
		remove:    []core.Feature{core.FeatureCloudDiscovery},
		wantCalls: []deleteCall{{features: []string{"CLOUD_DISCOVERY"}, deleteSnapshots: []bool{false}}},
	}, {
		name: "cloud discovery allowed when only non-protection features remain",
		projectFeatures: map[string]core.Status{
			core.FeatureCloudDiscovery.Name: core.StatusConnected,
			core.FeatureExocompute.Name:     core.StatusConnected,
		},
		remove:    []core.Feature{core.FeatureCloudDiscovery},
		wantCalls: []deleteCall{{features: []string{"CLOUD_DISCOVERY"}, deleteSnapshots: []bool{false}}},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancelCause(context.Background())
			defer assert.Context(t, ctx, cancel)

			cloudAccountID := uuid.New()
			h, recorded := removeProjectHandler(cancel, cloudAccountID, tt.projectFeatures)
			srv := httptest.NewServer(h)
			defer srv.Close()

			err := WrapGQL(graphql.NewTestClient(srv)).RemoveProject(ctx, cloudAccountID, tt.remove, true)
			switch {
			case tt.wantErr != "" && err == nil:
				t.Fatalf("expected error containing %q", tt.wantErr)
			case tt.wantErr != "" && !strings.Contains(err.Error(), tt.wantErr):
				t.Fatalf("error %q does not contain %q", err, tt.wantErr)
			case tt.wantErr == "" && err != nil:
				t.Fatalf("RemoveProject failed: %v", err)
			}

			if got := recorded(); !reflect.DeepEqual(got, tt.wantCalls) {
				t.Errorf("delete mutations: got %+v, want %+v", got, tt.wantCalls)
			}
		})
	}
}
