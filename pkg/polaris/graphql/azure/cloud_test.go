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
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// tenantsVars holds the variables of an allAzureCloudAccountTenants request.
type tenantsVars struct {
	Feature              string   `json:"feature"`
	Features             []string `json:"features"`
	IncludeSubscriptions bool     `json:"includeSubscriptionDetails"`
	IncludeResourceGroup bool     `json:"includeResourceGroup"`
}

// tenantsHandler records the variables of each request and responds with one
// tenant holding one subscription with the onboarded features matching the
// request. Like RSC, a resource group is only included when requested, and a
// request for the resource group of the UEM key management feature fails.
func tenantsHandler(cancel context.CancelCauseFunc, accountID uuid.UUID, onboarded []string, requests *[]tenantsVars) http.Handler {
	return handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		var vars tenantsVars
		if !decodeVars(cancel, req, &vars) {
			return
		}
		*requests = append(*requests, vars)

		var accounts []any
		for _, feature := range onboarded {
			requested := vars.Feature == core.FeatureAll.Name || vars.Feature == feature || slices.Contains(vars.Features, feature)
			if !vars.IncludeSubscriptions || !requested {
				continue
			}
			featureDetail := map[string]any{
				"feature":           feature,
				"permissionsGroups": []string{"BASIC"},
				"status":            "CONNECTED",
				"regions":           []string{"EASTUS2"},
			}
			if vars.IncludeResourceGroup {
				if feature == core.FeatureCloudNativeUEMKeyManagement.Name {
					cancel(errors.New("resource group requested for the UEM key management feature"))
					return
				}
				featureDetail["resourceGroup"] = map[string]any{"name": "terraform-test", "region": "EASTUS2"}
			}
			accounts = append(accounts, map[string]any{
				"id":            accountID,
				"name":          "subscription",
				"nativeId":      uuid.Nil,
				"featureDetail": featureDetail,
			})
		}
		response := map[string]any{"data": map[string]any{"result": []any{
			map[string]any{
				"cloudType":                       "AZUREPUBLICCLOUD",
				"azureCloudAccountTenantRubrikId": uuid.Nil,
				"domainName":                      "example.onmicrosoft.com",
				"subscriptionCount":               1,
				"subscriptions":                   accounts,
			},
		}}}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			cancel(err)
		}
	})
}

// TestCloudAccountTenantsAllFeatures verifies that reading the cloud accounts
// for all features first reads them without the resource groups, then reads
// the features found, except the UEM key management feature, with the resource
// groups. RSC fails the request when the resource group of the UEM key
// management feature is requested, see SPARK-1027147.
func TestCloudAccountTenantsAllFeatures(t *testing.T) {
	archival := core.FeatureCloudNativeArchival.Name
	protection := core.FeatureCloudNativeProtection.Name
	uem := core.FeatureCloudNativeUEMKeyManagement.Name

	tests := []struct {
		name       string
		onboarded  []string
		wantSecond []string
	}{
		{name: "WithUEM", onboarded: []string{archival, uem, protection}, wantSecond: []string{archival, protection}},
		{name: "WithoutUEM", onboarded: []string{archival, protection}, wantSecond: []string{archival, protection}},
		{name: "OnlyUEM", onboarded: []string{uem}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancelCause(context.Background())
			defer assert.Context(t, ctx, cancel)

			var requests []tenantsVars
			srv := httptest.NewServer(tenantsHandler(cancel, uuid.New(), test.onboarded, &requests))
			defer srv.Close()

			tenants, err := Wrap(graphql.NewTestClient(srv)).CloudAccountTenants(ctx, core.FeatureAll, true)
			if err != nil {
				t.Fatalf("CloudAccountTenants failed: %v", err)
			}

			wantRequests := 2
			if test.wantSecond == nil {
				wantRequests = 1
			}
			if len(requests) != wantRequests {
				t.Fatalf("got %d requests, want %d", len(requests), wantRequests)
			}
			if req := requests[0]; req.Feature != core.FeatureAll.Name || req.Features != nil ||
				!req.IncludeSubscriptions || req.IncludeResourceGroup {
				t.Errorf("request 0: got %+v, want all features, without the resource group", req)
			}
			if wantRequests == 2 {
				if req := requests[1]; req.Feature != test.wantSecond[0] || !slices.Equal(req.Features, test.wantSecond[1:]) ||
					!req.IncludeSubscriptions || !req.IncludeResourceGroup {
					t.Errorf("request 1: got %+v, want features %v, with the resource group", req, test.wantSecond)
				}
			}

			var got []string
			for _, tenant := range tenants {
				for _, account := range tenant.Accounts {
					got = append(got, account.Feature.Feature)
					wantRG := "terraform-test"
					if account.Feature.Feature == uem {
						wantRG = ""
					}
					if account.Feature.ResourceGroup.Name != wantRG {
						t.Errorf("feature %s: resource group %q, want %q", account.Feature.Feature,
							account.Feature.ResourceGroup.Name, wantRG)
					}
				}
			}
			slices.Sort(got)
			want := slices.Sorted(slices.Values(test.onboarded))
			if !slices.Equal(got, want) {
				t.Errorf("features: got %v, want %v", got, want)
			}
		})
	}
}

// TestCloudAccountTenantsSingleRequest verifies that a single request is made
// when reading a specific feature or when the cloud accounts are not included.
// The resource group is included, except for the UEM key management feature.
func TestCloudAccountTenantsSingleRequest(t *testing.T) {
	tests := []struct {
		name                 string
		feature              core.Feature
		includeSubscriptions bool
		wantResourceGroup    bool
	}{
		{name: "AllFeaturesWithoutSubscriptions", feature: core.FeatureAll, includeSubscriptions: false, wantResourceGroup: true},
		{name: "SpecificFeature", feature: core.FeatureCloudNativeArchival, includeSubscriptions: true, wantResourceGroup: true},
		{name: "UEMFeature", feature: core.FeatureCloudNativeUEMKeyManagement, includeSubscriptions: true, wantResourceGroup: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancelCause(context.Background())
			defer assert.Context(t, ctx, cancel)

			var requests []tenantsVars
			onboarded := []string{core.FeatureCloudNativeArchival.Name, core.FeatureCloudNativeUEMKeyManagement.Name}
			srv := httptest.NewServer(tenantsHandler(cancel, uuid.New(), onboarded, &requests))
			defer srv.Close()

			if _, err := Wrap(graphql.NewTestClient(srv)).CloudAccountTenants(ctx, test.feature,
				test.includeSubscriptions); err != nil {
				t.Fatalf("CloudAccountTenants failed: %v", err)
			}

			if len(requests) != 1 {
				t.Fatalf("got %d requests, want 1", len(requests))
			}
			req := requests[0]
			if req.Feature != test.feature.Name || req.Features != nil ||
				req.IncludeSubscriptions != test.includeSubscriptions || req.IncludeResourceGroup != test.wantResourceGroup {
				t.Errorf("got %+v, want feature %s without features, include resource group %t", req, test.feature.Name,
					test.wantResourceGroup)
			}
		})
	}
}
