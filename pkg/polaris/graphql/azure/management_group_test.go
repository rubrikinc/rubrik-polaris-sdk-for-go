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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// TestListManagementGroupHierarchy verifies the wire order and types of the
// flat union entities, and that the scopes and non-OAuth auth type are sent.
func TestListManagementGroupHierarchy(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	subID := uuid.New()
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
					"isAuthorized": true,
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

	uuidNamedGroup := uuid.NewString()
	entities, err := Wrap(graphql.NewTestClient(srv)).ListManagementGroupHierarchy(ctx, ListManagementGroupHierarchyParams{
		Cloud:                    PublicCloud,
		TenantDomain:             "example.onmicrosoft.com",
		SessionID:                uuid.New(),
		ManagementGroupID:        "root",
		ScopedManagementGroupIDs: []string{"child", uuidNamedGroup},
		ShouldRecurse:            true,
		Operation:                OperationAdd,
		AuthType:                 AuthTypeNonOAuth,
	})
	if err != nil {
		t.Fatalf("ListManagementGroupHierarchy failed: %v", err)
	}

	if vars["authType"] != string(AuthTypeNonOAuth) || vars["operationType"] != string(OperationAdd) {
		t.Errorf("invalid request variables: %v", vars)
	}
	if scopes, ok := vars["scopedManagementGroupIds"].([]any); !ok || len(scopes) != 2 || scopes[0] != "child" || scopes[1] != uuidNamedGroup {
		t.Errorf("invalid scoped IDs: %v", vars["scopedManagementGroupIds"])
	}
	if len(entities) != 2 {
		t.Fatalf("invalid hierarchy entities: %+v", entities)
	}
	if group := entities[0]; group.Entity.Type != ManagementGroupEntityType || group.Entity.NativeID != "child" || !group.Entity.IsAuthorized || !group.IsEligible {
		t.Errorf("invalid group entity: %+v", group)
	}
	if sub := entities[1]; sub.Entity.Type != SubscriptionEntityType || sub.Entity.NativeID != subID.String() ||
		sub.IsEligible || sub.IneligibilityReason != IneligibleAlreadyOnboarded {
		t.Errorf("invalid subscription entity: %+v", sub)
	}
}

func TestInitiateCloudAccountAddWithCustomerAppDoesNotLogCredentials(t *testing.T) {
	appID, sessionID := uuid.New(), uuid.New()
	const appSecret = secret.String("customer-secret-used-only-by-test")
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Variables struct {
				AppID     uuid.UUID     `json:"appId"`
				AppSecret secret.String `json:"appSecretKey"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if request.Variables.AppID != appID || request.Variables.AppSecret != appSecret {
			t.Error("expected the customer app credentials on the wire")
		}
		fmt.Fprintf(w, `{"data":{"result":{"success":true,"sessionId":%q}}}`, sessionID)
	}))
	defer srv.Close()

	var logs bytes.Buffer
	logger := log.NewStandardLogger()
	logger.SetOutput(&logs)
	logger.SetLogLevel(log.Trace)
	client := graphql.NewTestClient(srv)
	client.SetLogger(logger)
	actual, err := Wrap(client).InitiateCloudAccountAddWithCustomerApp(t.Context(), InitiateCloudAccountAddWithCustomerAppParams{
		Cloud: PublicCloud, TenantDomain: "example.onmicrosoft.com", AppID: appID, AppSecret: appSecret,
	})
	if err != nil || actual != sessionID {
		t.Fatal("did not receive the expected customer app session")
	}
	if strings.Contains(logs.String(), sessionID.String()) || strings.Contains(logs.String(), string(appSecret)) {
		t.Error("customer app session or secret leaked into SDK logging")
	}
}
