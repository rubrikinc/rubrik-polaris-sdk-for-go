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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

func TestRemoveManagementGroupSubscriptionFeatureWaitsForDeletion(t *testing.T) {
	sessionID, jobID, target, rejected := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	var waited bool
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
		case strings.Contains(request.Query, "azureCloudAccountAddWithCustomerAppInitiate"):
			if got, want := request.Variables["redirectUrl"], "http://"+r.Host+"/setup_azure"; got != want {
				t.Errorf("redirectUrl = %v, want %s", got, want)
			}
			fmt.Fprintf(w, `{"data":{"result":{"success":true,"sessionId":%q}}}`, sessionID)
		case strings.Contains(request.Query, "startDisableAzureCloudAccountJob"):
			fmt.Fprintf(w, `{"data":{"result":{"jobIds":[{"jobId":%q,"rubrikObjectId":%q}],"errors":[{"error":"rejected","rubrikObjectId":%q}]}}}`, jobID, target, rejected)
		case strings.Contains(request.Query, "getKorgTaskchainStatus"):
			waited = true
			fmt.Fprint(w, `{"data":{"getKorgTaskchainStatus":{"taskchain":{"state":"SUCCEEDED"}}}}`)
		case strings.Contains(request.Query, "allAzureCloudAccountTenants"):
			fmt.Fprintf(w, `{"data":{"result":[{"subscriptions":[{"id":%q,"nativeId":%q,"featureDetail":{"feature":"CLOUD_NATIVE_PROTECTION","status":"DISABLED"}}]}]}}`, target, target)
		default:
			t.Errorf("unexpected request: %s", request.Query)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	err := WrapGQL(graphql.NewTestClient(srv)).RemoveManagementGroupSubscriptionFeature(t.Context(), "example.onmicrosoft.com", uuid.New(),
		[]uuid.UUID{target, rejected}, core.FeatureCloudNativeProtection, false)
	if !waited {
		t.Error("did not wait for accepted job when another subscription was rejected")
	}
	if err == nil || !strings.Contains(err.Error(), "rejected") || !strings.Contains(err.Error(), "remains on subscription") {
		t.Fatalf("expected both the submission error and the incomplete deletion, got %v", err)
	}
}

// Archival features must use the synchronous, session-based delete mutation
// because they do not support the disable step or produce its job events.
func TestRemoveManagementGroupArchivalFeatureUsesSessionDelete(t *testing.T) {
	for _, feature := range []core.Feature{core.FeatureCloudNativeArchival, core.FeatureCloudNativeArchivalEncryption} {
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

			if err := WrapGQL(graphql.NewTestClient(srv)).RemoveManagementGroupSubscriptionFeature(t.Context(), "example.onmicrosoft.com", uuid.New(),
				[]uuid.UUID{target}, feature, false); err != nil {
				t.Fatal(err)
			}
			if !deleted.Load() {
				t.Fatal("did not delete the archival feature")
			}
		})
	}
}
