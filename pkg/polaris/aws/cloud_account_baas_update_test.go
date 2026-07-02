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

package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// TestManagedAccountPermissionUpdate proves the permission-upgrade loop end to
// end against a stateful mock RSC:
//
//  1. A feature is MISSING_PERMISSIONS (as when RSC raises a permission
//     version) -> PrepareManagedAccountUpdate detects it and returns the update
//     CloudFormation template URL (the aws_cloudformation_stack would redeploy
//     with it).
//  2. After the stack is redeployed, CompleteManagedAccountUpdate notifies RSC
//     (upgradeAwsCloudAccountFeaturesWithoutCft) and waits for the feature to
//     reconnect.
//  3. Once reconnected, PrepareManagedAccountUpdate reports no further update is
//     needed (empty URL) - i.e. no spurious drift once healed.
func TestManagedAccountPermissionUpdate(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	accountID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	const (
		nativeID          = "123456789012"
		updateTemplateURL = "https://example.com/update-template.json?X-Amz-Signature=abc123"
	)

	// Mock RSC state: the feature starts MISSING_PERMISSIONS and flips to
	// CONNECTED once RSC is notified that the roles were updated.
	var (
		connected     bool
		prepareCalled bool
		upgradeCalled bool
	)

	accountResponse := func(status core.Status) string {
		return fmt.Sprintf(`{"data":{"result":[{`+
			`"awsCloudAccount":{"cloudType":"STANDARD","id":%q,"nativeId":%q,"accountName":"example-account"},`+
			`"featureDetails":[{"feature":"CLOUD_NATIVE_PROTECTION","status":%q,"awsRegions":["US_EAST_1"],`+
			`"stackArn":"arn:aws:cloudformation:us-east-1:123456789012:stack/rubrik-polaris/abc"}]`+
			`}]}}`, accountID.String(), nativeID, status)
	}

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			cancel(err)
			return
		}
		var payload struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			cancel(err)
			return
		}

		var resp string
		switch {
		case strings.Contains(payload.Query, "allAwsCloudAccountsWithFeatures"):
			status := core.StatusMissingPermissions
			if connected {
				status = core.StatusConnected
			}
			resp = accountResponse(status)
		case strings.Contains(payload.Query, "prepareFeatureUpdateForAwsCloudAccount"):
			prepareCalled = true
			resp = fmt.Sprintf(`{"data":{"result":{"cloudFormationUrl":"https://console.aws.amazon.com/x","templateUrl":%q}}}`, updateTemplateURL)
		case strings.Contains(payload.Query, "upgradeAwsCloudAccountFeaturesWithoutCft"):
			upgradeCalled = true
			connected = true
			resp = `{"data":{"result":true}}`
		default:
			cancel(fmt.Errorf("unexpected operation in query: %s", payload.Query))
			return
		}
		if _, err := io.WriteString(w, resp); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	api := WrapGQL(graphql.NewTestClient(srv))

	// 1. Drift is detected and the update template URL is returned.
	templateURL, err := api.PrepareManagedAccountUpdate(ctx, accountID)
	if err != nil {
		t.Fatalf("PrepareManagedAccountUpdate (drift): %v", err)
	}
	if templateURL != updateTemplateURL {
		t.Errorf("update template URL = %q, want %q", templateURL, updateTemplateURL)
	}
	if !prepareCalled {
		t.Error("expected prepareFeatureUpdateForAwsCloudAccount to be called")
	}

	// 2. The stack has been redeployed; complete the update (notify + reconnect).
	if err := api.CompleteManagedAccountUpdate(ctx, accountID); err != nil {
		t.Fatalf("CompleteManagedAccountUpdate: %v", err)
	}
	if !upgradeCalled {
		t.Error("expected the permissions-updated notification to be sent")
	}

	// 3. Once healed, no further update is reported.
	templateURL, err = api.PrepareManagedAccountUpdate(ctx, accountID)
	if err != nil {
		t.Fatalf("PrepareManagedAccountUpdate (healed): %v", err)
	}
	if templateURL != "" {
		t.Errorf("expected no update after reconnect, got %q", templateURL)
	}
}

// TestManagedAccountPermissionsVersion proves the permissions-version signal is
// deterministic for a given set of versions and changes when RSC raises a
// permission version. This is the stable trigger the phase-2 resource keys off
// to re-complete onboarding on an upgrade.
func TestManagedAccountPermissionsVersion(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	accountID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	basicVersion := 9 // the current BASIC permission-group version RSC reports

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			cancel(err)
			return
		}
		var payload struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			cancel(err)
			return
		}

		var resp string
		switch {
		case strings.Contains(payload.Query, "allAwsCloudAccountsWithFeatures"):
			resp = fmt.Sprintf(`{"data":{"result":[{`+
				`"awsCloudAccount":{"cloudType":"STANDARD","id":%q,"nativeId":"123456789012","accountName":"example-account"},`+
				`"featureDetails":[{"feature":"CLOUD_NATIVE_PROTECTION","status":"CONNECTED","awsRegions":["US_EAST_1"]}]`+
				`}]}}`, accountID.String())
		case strings.Contains(payload.Query, "allAWSLatestPermissionsByPermissionsGroup"):
			resp = fmt.Sprintf(`{"data":{"result":[{"feature":"CLOUD_NATIVE_PROTECTION",`+
				`"permissionsGroupPermissions":[{"permissionsGroup":"BASIC","version":%d}]}]}}`, basicVersion)
		default:
			cancel(fmt.Errorf("unexpected operation in query: %s", payload.Query))
			return
		}
		if _, err := io.WriteString(w, resp); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	api := WrapGQL(graphql.NewTestClient(srv))

	v1, err := api.ManagedAccountPermissionsVersion(ctx, accountID)
	if err != nil {
		t.Fatalf("ManagedAccountPermissionsVersion: %v", err)
	}
	v1again, err := api.ManagedAccountPermissionsVersion(ctx, accountID)
	if err != nil {
		t.Fatalf("ManagedAccountPermissionsVersion: %v", err)
	}
	if v1 != v1again {
		t.Errorf("version not deterministic: %q != %q", v1, v1again)
	}

	basicVersion = 10 // RSC raises the permission version
	v2, err := api.ManagedAccountPermissionsVersion(ctx, accountID)
	if err != nil {
		t.Fatalf("ManagedAccountPermissionsVersion: %v", err)
	}
	if v2 == v1 {
		t.Errorf("expected the version to change after a permission bump, still %q", v2)
	}
}
