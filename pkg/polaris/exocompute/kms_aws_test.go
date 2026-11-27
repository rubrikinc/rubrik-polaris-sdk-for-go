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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/exocompute"
)

func TestAWSKMSMapping(t *testing.T) {
	const appID = "aec60ea2-dd85-483b-88ea-294096493cac"
	const hostID = "afc24b01-6586-47cc-896a-8a5aed160b82"
	app, host := uuid.MustParse(appID), uuid.MustParse(hostID)

	mapKMS := func(ctx context.Context, api API) (any, error) {
		return api.MapAWSCloudAccountWithKMSSharing(ctx, app, host)
	}
	unmapKMS := func(ctx context.Context, api API) (any, error) {
		return api.UnmapAWSCloudAccountWithKMSSharing(ctx, app)
	}
	getStatus := func(ctx context.Context, api API) (any, error) {
		return api.AWSKMSMappingStatus(ctx, app)
	}
	confirm := func(ctx context.Context, api API) (any, error) {
		return nil, api.ConfirmAWSManualKMSMapping(ctx, app)
	}

	mapVars := map[string]any{"applicationCloudAccountId": appID, "exocomputeCloudAccountId": hostID}
	appVars := map[string]any{"applicationCloudAccountId": appID}
	var tests = []struct {
		name      string
		call      func(context.Context, API) (any, error)
		query     []string
		omit      []string
		vars      map[string]any
		reply     string
		want      any
		wantError string
	}{
		{
			name: "LegacyMap",
			call: func(ctx context.Context, api API) (any, error) {
				return nil, api.MapAWSCloudAccount(ctx, app, host)
			},
			query: []string{"result: mapCloudAccountExocomputeAccount", "isSuccess"},
			omit:  []string{"cloudSpecificParams", "results"},
			vars:  map[string]any{"cloudVendor": "AWS", "cloudAccountIds": []any{appID}, "exocomputeCloudAccountId": hostID},
			reply: `{"isSuccess":true}`,
		},
		{
			name: "LegacyUnmap",
			call: func(ctx context.Context, api API) (any, error) {
				return nil, api.UnmapAWSCloudAccount(ctx, app)
			},
			query: []string{"result: unmapCloudAccountExocomputeAccount", "isSuccess"},
			omit:  []string{"results"},
			vars:  map[string]any{"cloudVendor": "AWS", "cloudAccountIds": []any{appID}},
			reply: `{"isSuccess":true}`,
		},
		{
			name:  "MapManual",
			call:  mapKMS,
			query: []string{"result: mapCloudAccountExocomputeAccount", "shouldEnableAutomatedKeySharing: true", "manualRolePolicyJson"},
			vars:  mapVars,
			reply: `{"isSuccess":true,"results":[{"applicationCloudAccountId":"APP_ID","manualRolePolicyJson":"{\"Version\":\"2012-10-17\"}"}]}`,
			want:  exocompute.AWSKMSMappingAction{AppCloudAccountID: app, ManualRolePolicy: `{"Version":"2012-10-17"}`},
		},
		{
			name:  "MapCFT",
			call:  mapKMS,
			query: []string{"result: mapCloudAccountExocomputeAccount", "shouldEnableAutomatedKeySharing: true", "cloudFormationUrl", "templateUrl"},
			vars:  mapVars,
			reply: `{"isSuccess":true,"results":[{"applicationCloudAccountId":"APP_ID","cloudFormationUrl":"https://example.com/update","templateUrl":"https://example.com/template","isPendingApproval":true}]}`,
			want: exocompute.AWSKMSMappingAction{
				AppCloudAccountID: app, PendingApproval: true, CloudFormationURL: "https://example.com/update",
				TemplateURL: "https://example.com/template",
			},
		},
		{
			name:      "MapAccountError",
			call:      mapKMS,
			query:     []string{"result: mapCloudAccountExocomputeAccount", "errorMessage"},
			vars:      mapVars,
			reply:     `{"isSuccess":false,"results":[{"applicationCloudAccountId":"APP_ID","errorMessage":"unsupported onboarding type"}]}`,
			wantError: "unsupported onboarding type",
		},
		{
			name:      "MapSkippedAccount",
			call:      mapKMS,
			query:     []string{"result: mapCloudAccountExocomputeAccount", "results"},
			vars:      mapVars,
			reply:     `{"isSuccess":true,"results":[]}`,
			wantError: "verify that Cloud Native Protection or RDS Protection is enabled",
		},
		{
			name:  "UnmapNeedsManualRevoke",
			call:  unmapKMS,
			query: []string{"result: unmapCloudAccountExocomputeAccount", "isPendingApproval"},
			vars:  appVars,
			reply: `{"isSuccess":true,"results":[{"applicationCloudAccountId":"APP_ID","isPendingApproval":true}]}`,
			want:  exocompute.AWSKMSMappingAction{AppCloudAccountID: app, PendingApproval: true},
		},
		{
			name:  "UnmapWithoutRevoke",
			call:  unmapKMS,
			query: []string{"result: unmapCloudAccountExocomputeAccount", "results"},
			vars:  appVars,
			reply: `{"isSuccess":true,"results":[]}`,
			want:  exocompute.AWSKMSMappingAction{AppCloudAccountID: app},
		},
		{
			name:      "UnmapAccountError",
			call:      unmapKMS,
			query:     []string{"result: unmapCloudAccountExocomputeAccount", "errorMessage"},
			vars:      appVars,
			reply:     `{"isSuccess":false,"results":[{"applicationCloudAccountId":"APP_ID","errorMessage":"revoke failed"}]}`,
			wantError: "revoke failed",
		},
		{
			name:  "StatusConnecting",
			call:  getStatus,
			query: []string{"result: awsExocomputeKmsMappingStatus", "kmsSharingMode", "provisioningStatus"},
			vars:  appVars,
			reply: `{"results":[{"applicationCloudAccountId":"APP_ID","kmsSharingMode":"KMS_SHARING_MODE_SHARING_ONLY","provisioningStatus":"KMS_MAPPING_PROVISIONING_STATUS_CONNECTING"}]}`,
			want: exocompute.AWSKMSMappingStatus{
				AppCloudAccountID: app, SharingMode: exocompute.KMSSharingModeSharingOnly,
				ProvisioningStatus: exocompute.KMSMappingProvisioningStatusConnecting,
			},
		},
		{
			name:  "Confirm",
			call:  confirm,
			query: []string{"result: updateManualKmsMapping", "errorMessage"},
			omit:  []string{"gatewayKmsKeyArn"},
			vars:  appVars,
			reply: `{"results":[{"applicationCloudAccountId":"APP_ID","errorMessage":""}]}`,
		},
		{
			name:      "ConfirmAccountError",
			call:      confirm,
			query:     []string{"result: updateManualKmsMapping", "errorMessage"},
			vars:      appVars,
			reply:     `{"results":[{"applicationCloudAccountId":"APP_ID","errorMessage":"confirmation failed"}]}`,
			wantError: "confirmation failed",
		},
		{
			name:      "ConfirmUnexpectedAccount",
			call:      confirm,
			query:     []string{"result: updateManualKmsMapping", "errorMessage"},
			vars:      appVars,
			reply:     `{"results":[{"applicationCloudAccountId":"afc24b01-6586-47cc-896a-8a5aed160b82","errorMessage":""}]}`,
			wantError: "missing KMS confirmation result",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancelCause(t.Context())
			defer assert.Context(t, ctx, cancel)

			srv := httptest.NewTLSServer(handler.GraphQL(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					Query     string         `json:"query"`
					Variables map[string]any `json:"variables"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					cancel(err)
					return
				}
				for _, part := range tc.query {
					if !strings.Contains(request.Query, part) {
						cancel(fmt.Errorf("GraphQL query missing %q", part))
						return
					}
				}
				for _, part := range tc.omit {
					if strings.Contains(request.Query, part) {
						cancel(fmt.Errorf("GraphQL query unexpectedly contains %q", part))
						return
					}
				}
				if diff := cmp.Diff(tc.vars, request.Variables); diff != "" {
					cancel(fmt.Errorf("unexpected GraphQL variables (-want +got):\n%s", diff))
					return
				}
				_, err := io.WriteString(w, `{"data":{"result":`+strings.ReplaceAll(tc.reply, "APP_ID", appID)+`}}`)
				if err != nil {
					cancel(err)
				}
			}))
			defer srv.Close()

			gql := graphql.NewTestClient(srv)
			got, err := tc.call(ctx, API{client: gql, log: gql.Log()})
			switch {
			case tc.wantError != "":
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("expected error containing %q, got %v", tc.wantError, err)
				}
			case err != nil:
				t.Fatal(err)
			default:
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Fatalf("unexpected result (-want +got):\n%s", diff)
				}
			}
		})
	}
}
