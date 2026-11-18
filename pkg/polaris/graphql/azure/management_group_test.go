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

package azure

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
)

// TestListManagementGroupHierarchy verifies that the management groups and
// subscriptions of the EntityType union are separated, and that the request
// is sent with the non-OAuth auth type.
func TestListManagementGroupHierarchy(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	subID := uuid.New()
	mgID := uuid.New()
	var vars map[string]any
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			cancel(err)
			return
		}
		vars = body.Variables

		response := map[string]any{"data": map[string]any{"result": map[string]any{"entities": []any{
			map[string]any{
				"isEligible":          true,
				"ineligibilityReason": string(IneligibleUnspecified),
				"entity": map[string]any{
					"__typename": "AzureManagementGroup", "nativeId": "child", "name": "Child",
					"customerManagementGroupId": mgID, "isAuthorized": true,
				},
			},
			map[string]any{
				"isEligible":          false,
				"ineligibilityReason": string(IneligibleAlreadyOnboarded),
				"entity": map[string]any{
					"__typename": "CloudAccountsAzureSubscription", "nativeId": subID, "name": "Sub", "isAuthorized": true,
				},
			},
		}}}}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	hierarchy, err := Wrap(graphql.NewTestClient(srv)).ListManagementGroupHierarchy(ctx, ManagementGroupHierarchyRequest{
		Cloud:             PublicCloud,
		TenantDomain:      "example.onmicrosoft.com",
		SessionID:         uuid.New(),
		ManagementGroupID: "root",
		ShouldRecurse:     true,
		Operation:         OperationAdd,
		AuthType:          AuthTypeNonOAuth,
	})
	if err != nil {
		t.Fatalf("ListManagementGroupHierarchy failed: %v", err)
	}

	if vars["authType"] != string(AuthTypeNonOAuth) || vars["operationType"] != string(OperationAdd) {
		t.Errorf("invalid request variables: %v", vars)
	}
	if len(hierarchy.ManagementGroups) != 1 || hierarchy.ManagementGroups[0].NativeID != "child" ||
		hierarchy.ManagementGroups[0].CustomerManagementGroupID != mgID {
		t.Errorf("invalid management groups: %+v", hierarchy.ManagementGroups)
	}
	if len(hierarchy.Subscriptions) != 1 || hierarchy.Subscriptions[0].NativeID != subID ||
		hierarchy.Subscriptions[0].IsEligible ||
		hierarchy.Subscriptions[0].IneligibilityReason != IneligibleAlreadyOnboarded {
		t.Errorf("invalid subscriptions: %+v", hierarchy.Subscriptions)
	}
}

// TestAddCloudAccountThroughManagementGroup verifies that the management group
// is sent to RSC and that only the task chain id is needed in the response,
// since RSC doesn't return a status for management group onboarding.
func TestAddCloudAccountThroughManagementGroup(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	taskChainID := uuid.New()
	var vars struct {
		ManagementGroup ManagementGroupInput          `json:"managementGroup"`
		Subscriptions   []ManagementGroupSubscription `json:"subscriptions"`
		Regions         []string                      `json:"regions"`
	}
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Variables json.RawMessage `json:"variables"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			cancel(err)
			return
		}
		if err := json.Unmarshal(body.Variables, &vars); err != nil {
			cancel(err)
			return
		}

		response := map[string]any{"data": map[string]any{"result": map[string]any{
			"taskchainUuid": taskChainID, "tenantId": "", "status": []any{},
		}}}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	subID := uuid.New()
	got, err := Wrap(graphql.NewTestClient(srv)).AddCloudAccountThroughManagementGroup(ctx, uuid.New(),
		"example.onmicrosoft.com", []azure.Region{azure.RegionFromName("eastus2")},
		ManagementGroupInput{NativeID: "mg", Name: "MG"}, []ManagementGroupSubscription{{
			Subscription: SubscriptionInput{NativeID: subID, Name: "Sub"},
			Features:     []ManagementGroupFeature{{FeatureType: "CLOUD_NATIVE_PROTECTION", PermissionGroups: []string{"BASIC"}}},
		}})
	if err != nil {
		t.Fatalf("AddCloudAccountThroughManagementGroup failed: %v", err)
	}

	if got != taskChainID {
		t.Errorf("got task chain %s, want %s", got, taskChainID)
	}
	if vars.ManagementGroup.NativeID != "mg" || len(vars.Subscriptions) != 1 ||
		vars.Subscriptions[0].Subscription.NativeID != subID ||
		len(vars.Regions) != 1 || vars.Regions[0] != "EASTUS2" {
		t.Errorf("invalid request variables: %+v", vars)
	}
}
