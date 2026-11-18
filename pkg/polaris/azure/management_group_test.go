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
	"strings"
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
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch {
		case strings.Contains(request.Query, "azureCloudAccountAddWithCustomerAppInitiate"):
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
