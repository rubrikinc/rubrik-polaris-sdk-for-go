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

// Package azure provides a high-level interface to the Azure part of the RSC
// platform.
package azure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

func TestRemoveManagementGroupSubscriptionFeatureWaitsForDeletion(t *testing.T) {
	for _, test := range []struct {
		name         string
		combined     bool
		rbac         bool
		status       string
		partial      bool
		feature      string
		wantDelete   bool
		wantError    string
		wantDeadline bool
	}{
		{name: "combined", combined: true},
		{name: "combined with task RBAC lag", combined: true, rbac: true},
		{name: "combined with task RBAC lag still has disabled feature", combined: true, rbac: true, status: "DISABLED", wantDeadline: true},
		{name: "combined task left feature disabled", combined: true, status: "DISABLED", wantError: "remains on subscription"},
		{name: "disable then delete", status: "DISABLED", wantDelete: true},
		{name: "disable then delete with task RBAC lag", rbac: true, status: "DISABLED", wantDelete: true},
		{name: "disable then delete with rejected target", status: "DISABLED", partial: true, wantDelete: true, wantError: "rejected"},
		{name: "disable left feature connected", status: "CONNECTED", wantError: "expected DISABLED"},
		{name: "Discovery uses a disable job", feature: core.CloudDiscovery, combined: true},
		{name: "Exocompute uses a disable job", feature: core.Exocompute, status: "DISABLED", wantDelete: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			sessionID, jobID, target, rejected := uuid.New(), uuid.New(), uuid.New(), uuid.New()
			feature := core.FeatureCloudNativeProtection
			if test.feature != "" {
				feature = core.Feature{Name: test.feature}
			}
			var waited, deleted atomic.Bool
			srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					Query     string         `json:"query"`
					Variables map[string]any `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				switch {
				case strings.Contains(request.Query, "result: featureFlag("):
					if request.Variables["flagName"] != string(core.FeatureFlagAzureDisableDeleteCombined) {
						t.Errorf("unexpected flag: %v", request.Variables["flagName"])
					}
					fmt.Fprintf(w, `{"data":{"result":{"name":"CNP_AZURE_DISABLE_DELETE_COMBINED","variant":%q}}}`, fmt.Sprint(test.combined))
				case strings.Contains(request.Query, "azureCloudAccountAddWithCustomerAppInitiate"):
					if got, want := request.Variables["redirectUrl"], "http://"+r.Host+"/setup_azure"; got != want {
						t.Errorf("redirectUrl = %v, want %s", got, want)
					}
					fmt.Fprintf(w, `{"data":{"result":{"success":true,"sessionId":%q}}}`, sessionID)
				case strings.Contains(request.Query, "startDisableAzureCloudAccountJob"):
					if request.Variables["sessionId"] != sessionID.String() || request.Variables["deleteSnapshots"] != false ||
						request.Variables["feature"] != feature.Name {
						t.Errorf("unexpected disable variables: %+v", request.Variables)
					}
					var rejection string
					if test.partial {
						rejection = fmt.Sprintf(`{"error":"rejected","rubrikObjectId":%q}`, rejected)
					}
					fmt.Fprintf(w, `{"data":{"result":{"jobIds":[{"jobId":%q,"rubrikObjectId":%q}],"errors":[%s]}}}`, jobID, target, rejection)
				case strings.Contains(request.Query, "getKorgTaskchainStatus"):
					waited.Store(true)
					if test.rbac {
						fmt.Fprint(w, `{"errors":[{"message":"task not visible yet","extensions":{"code":403}}]}`)
					} else {
						fmt.Fprint(w, `{"data":{"getKorgTaskchainStatus":{"taskchain":{"state":"SUCCEEDED"}}}}`)
					}
				case strings.Contains(request.Query, "deleteAzureCloudAccount(input:"):
					if !waited.Load() || !test.wantDelete || deleted.Swap(true) {
						t.Error("unexpected explicit delete call")
					}
					ids, ok := request.Variables["subscriptionIds"].([]any)
					if !ok || len(ids) != 1 || ids[0] != target.String() || request.Variables["sessionId"] != sessionID.String() {
						t.Error("expected a session delete for only the accepted target")
					}
					fmt.Fprintf(w, `{"data":{"result":{"status":[{"azureSubscriptionNativeId":%q,"isSuccess":true}]}}}`, target)
				case strings.Contains(request.Query, "allAzureCloudAccountTenants"):
					if deleted.Load() || test.status == "" {
						fmt.Fprint(w, `{"data":{"result":[]}}`)
					} else {
						fmt.Fprintf(w, `{"data":{"result":[{"subscriptions":[{"id":%q,"nativeId":%q,"featureDetail":{"feature":%q,"status":%q}}]}]}}`, target, target, feature.Name, test.status)
					}
				default:
					t.Errorf("unexpected request: %s", request.Query)
					w.WriteHeader(http.StatusBadRequest)
				}
			}))
			defer srv.Close()

			ids := []uuid.UUID{target}
			if test.partial {
				ids = append(ids, rejected)
			}
			ctx := t.Context()
			if test.wantDeadline {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, time.Second)
				defer cancel()
			}
			err := WrapGQL(graphql.NewTestClient(srv)).RemoveManagementGroupSubscriptionFeature(ctx, RemoveManagementGroupSubscriptionFeatureParams{
				TenantDomain: "example.onmicrosoft.com", AppID: uuid.New(), CloudAccountIDs: ids, Feature: feature,
			})
			if !waited.Load() || deleted.Load() != test.wantDelete {
				t.Errorf("waited=%t, deleted=%t, want waited=true, deleted=%t", waited.Load(), deleted.Load(), test.wantDelete)
			}
			if test.wantDeadline {
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("got %v, want deadline while combined deletion is incomplete", err)
				}
				return
			}
			if test.wantError == "" && err != nil {
				t.Fatal(err)
			}
			if test.wantError != "" && (err == nil || !strings.Contains(err.Error(), test.wantError)) {
				t.Fatalf("got %v, want error containing %q", err, test.wantError)
			}
		})
	}
}

// Features outside the Azure disable-job workflow go directly to session delete.
func TestRemoveManagementGroupFeatureUsesSessionDelete(t *testing.T) {
	for _, feature := range []core.Feature{
		core.FeatureCloudNativeArchival, core.FeatureCloudNativeArchivalEncryption,
		core.FeatureCloudNativeUEMKeyManagement, {Name: core.ServerAndApps},
	} {
		t.Run(feature.Name, func(t *testing.T) {
			sessionID, target := uuid.New(), uuid.New()
			var deleted atomic.Bool
			srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					Query     string `json:"query"`
					Variables struct {
						SessionID       uuid.UUID   `json:"sessionId"`
						SubscriptionIDs []uuid.UUID `json:"subscriptionIds"`
						Features        []string    `json:"features"`
					} `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				switch {
				case strings.Contains(request.Query, "azureCloudAccountAddWithCustomerAppInitiate"):
					fmt.Fprintf(w, `{"data":{"result":{"success":true,"sessionId":%q}}}`, sessionID)
				case strings.Contains(request.Query, "deleteAzureCloudAccount(input:"):
					deleted.Store(true)
					vars := request.Variables
					if vars.SessionID != sessionID || !slices.Equal(vars.SubscriptionIDs, []uuid.UUID{target}) || !slices.Equal(vars.Features, []string{feature.Name}) {
						t.Errorf("unexpected delete variables: %+v", vars)
					}
					fmt.Fprintf(w, `{"data":{"result":{"status":[{"azureSubscriptionNativeId":%q,"isSuccess":true,"error":""}]}}}`, uuid.New())
				case strings.Contains(request.Query, "allAzureCloudAccountTenants"):
					fmt.Fprint(w, `{"data":{"result":[]}}`)
				default:
					t.Errorf("unexpected request (must not disable or poll a job): %s", request.Query)
					w.WriteHeader(http.StatusBadRequest)
				}
			}))
			defer srv.Close()

			if err := WrapGQL(graphql.NewTestClient(srv)).RemoveManagementGroupSubscriptionFeature(t.Context(), RemoveManagementGroupSubscriptionFeatureParams{
				TenantDomain: "example.onmicrosoft.com", AppID: uuid.New(), CloudAccountIDs: []uuid.UUID{target}, Feature: feature,
			}); err != nil {
				t.Fatal(err)
			}
			if !deleted.Load() {
				t.Fatal("did not delete the feature")
			}
		})
	}
}

// TestManagementGroupHierarchy checks the default eligibility feature and that
// the high-level API splits the flat wire entities without inventing parents.
func TestManagementGroupHierarchy(t *testing.T) {
	sessionID, subID := uuid.New(), uuid.New()
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Query     string `json:"query"`
			Variables struct {
				SessionID     uuid.UUID `json:"sessionId"`
				ShouldRecurse bool      `json:"shouldRecurse"`
				Features      []string  `json:"features"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch {
		case strings.Contains(request.Query, "azureCloudAccountAddWithCustomerAppInitiate"):
			fmt.Fprintf(w, `{"data":{"result":{"success":true,"sessionId":%q}}}`, sessionID)
		case strings.Contains(request.Query, "azureListManagementGroupHierarchy"):
			vars := request.Variables
			if vars.SessionID != sessionID || !vars.ShouldRecurse || !slices.Equal(vars.Features, []string{core.CloudDiscovery}) {
				t.Error("invalid hierarchy request")
			}
			fmt.Fprintf(w, `{"data":{"result":{"entities":[
				{"entity":{"__typename":"AzureManagementGroup","nativeId":"named-child","name":"Child","isAuthorized":true}},
				{"entity":{"__typename":"CloudAccountsAzureSubscription","nativeId":%q,"name":"Nested Subscription"},"isEligible":true}
			]}}}`, subID)
		default:
			t.Errorf("unexpected request: %s", request.Query)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	h, err := WrapGQL(graphql.NewTestClient(srv)).ManagementGroupHierarchy(t.Context(), ManagementGroupHierarchyParams{
		TenantDomain: "example.onmicrosoft.com", AppID: uuid.New(), ManagementGroupID: "root",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(h.ManagementGroups) != 1 || h.ManagementGroups[0].NativeID != "named-child" || !h.ManagementGroups[0].IsAuthorized {
		t.Errorf("unexpected groups: %+v", h.ManagementGroups)
	}
	if len(h.Subscriptions) != 1 || h.Subscriptions[0].NativeID != subID || !h.Subscriptions[0].IsEligible {
		t.Errorf("unexpected subscriptions: %+v", h.Subscriptions)
	}
}

func TestStartManagementGroupOnboardingMissingGroup(t *testing.T) {
	for _, test := range []struct{ name, groups string }{
		{name: "empty list", groups: "[]"},
		{name: "different group", groups: `[{"nativeId":"other","name":"Other"}]`},
	} {
		t.Run(test.name, func(t *testing.T) {
			srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, r *http.Request) {
				var body struct{ Query string }
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				switch {
				case strings.Contains(body.Query, "azureCloudAccountAddWithCustomerAppInitiate"):
					fmt.Fprintf(w, `{"data":{"result":{"success":true,"sessionId":%q}}}`, uuid.New())
				case strings.Contains(body.Query, "azureListManagementGroups"):
					fmt.Fprintf(w, `{"data":{"result":{"managementGroups":%s}}}`, test.groups)
				default:
					t.Errorf("unexpected request: %s", body.Query)
					w.WriteHeader(http.StatusBadRequest)
				}
			}))
			defer srv.Close()

			_, err := WrapGQL(graphql.NewTestClient(srv)).StartManagementGroupOnboarding(t.Context(), StartManagementGroupOnboardingParams{
				TenantDomain: "example.onmicrosoft.com", AppID: uuid.New(), ManagementGroupID: "root",
				Subscriptions: []ManagementGroupSubscription{{NativeID: uuid.New(), Features: []ManagementGroupFeature{{Feature: core.FeatureCloudDiscovery}}}},
			})
			if !errors.Is(err, graphql.ErrNotFound) {
				t.Fatalf("expected ErrNotFound, got %v", err)
			}
		})
	}
}
