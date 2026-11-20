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
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// mockClient returns a polaris.Client whose GraphQL requests are served by the
// given test server.
func mockClient(srv *httptest.Server) *polaris.Client {
	return &polaris.Client{GQL: graphql.NewTestClient(srv)}
}

// TestArtifactsFiltersForRoleChaining verifies that the CROSSACCOUNT role
// artifact is dropped for the role-chaining-only feature set but kept for other
// feature sets. Part of a workaround due to issues with the RSC GraphQL API.
func TestArtifactsFiltersForRoleChaining(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		if _, err := fmt.Fprint(w, `{"data":{"result":[
			{"externalArtifactKey":"CROSSACCOUNT_ROLE_ARN","awsManagedPolicies":[],"customerManagedPolicies":[]},
			{"externalArtifactKey":"ROLE_CHAINING_ROLE_ARN","awsManagedPolicies":[],"customerManagedPolicies":[]}
		]}}`); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	tests := []struct {
		name     string
		features []core.Feature
		expected []string
	}{{
		name:     "role chaining only drops CROSSACCOUNT",
		features: []core.Feature{core.FeatureRoleChaining},
		expected: []string{"ROLE_CHAINING"},
	}, {
		name:     "non role chaining keeps CROSSACCOUNT",
		features: []core.Feature{core.FeatureCloudNativeProtection},
		expected: []string{"CROSSACCOUNT", "ROLE_CHAINING"},
	}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, roles, err := Wrap(mockClient(srv)).Artifacts(ctx, string(aws.CloudStandard), tc.features)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(roles, tc.expected) {
				t.Errorf("expected roles %v, got %v", tc.expected, roles)
			}
		})
	}
}

// TestAccountArtifactsFiltersForRoleChaining verifies that the registered
// CROSSACCOUNT role artifact is dropped for role-chaining-only accounts but
// kept for other accounts. Part of a workaround due to issues with the RSC
// GraphQL API.
func TestAccountArtifactsFiltersForRoleChaining(t *testing.T) {
	tests := []struct {
		name             string
		feature          string
		wantCrossAccount bool
	}{{
		name:             "role chaining only drops CROSSACCOUNT",
		feature:          core.FeatureRoleChaining.Name,
		wantCrossAccount: false,
	}, {
		name:             "non role chaining keeps CROSSACCOUNT",
		feature:          core.FeatureCloudNativeProtection.Name,
		wantCrossAccount: true,
	}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancelCause(context.Background())
			defer assert.Context(t, ctx, cancel)

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
				switch {
				case strings.Contains(payload.Query, "allAwsCloudAccountsWithFeatures"):
					if _, err := fmt.Fprintf(w, `{"data":{"result":[
						{"awsCloudAccount":{"cloudType":"STANDARD","id":"b48e7ad0-7b86-4c96-b6ba-97eb6a82f765","nativeId":"123456789012","accountName":"test","message":""},
						 "featureDetails":[{"feature":%q,"permissionsGroups":[],"awsRegions":[],"roleArn":"","stackArn":"","status":"CONNECTED","mappedAccounts":[],"roleChainingDetails":null}],
						 "roleChainingAccount":null}
					]}}`, tc.feature); err != nil {
						cancel(err)
					}
				case strings.Contains(payload.Query, "awsArtifactsToDelete"):
					if _, err := fmt.Fprint(w, `{"data":{"result":{"artifactsToDelete":[
						{"feature":"ROLE_CHAINING","artifactsToDelete":[
							{"externalArtifactKey":"CROSSACCOUNT_ROLE_ARN","externalArtifactValue":"arn:aws:iam::123456789012:role/chain"},
							{"externalArtifactKey":"ROLE_CHAINING_ROLE_ARN","externalArtifactValue":"arn:aws:iam::123456789012:role/chain"}
						]}
					]}}}`); err != nil {
						cancel(err)
					}
				default:
					cancel(fmt.Errorf("unexpected query: %s", payload.Query))
				}
			}))
			defer srv.Close()

			_, roles, err := Wrap(mockClient(srv)).AccountArtifacts(ctx, uuid.MustParse("b48e7ad0-7b86-4c96-b6ba-97eb6a82f765"))
			if err != nil {
				t.Fatal(err)
			}
			if _, ok := roles[crossAccountArtifact]; ok != tc.wantCrossAccount {
				t.Errorf("CROSSACCOUNT present=%v, want %v (roles=%v)", ok, tc.wantCrossAccount, roles)
			}
			if _, ok := roles["ROLE_CHAINING"]; !ok {
				t.Errorf("expected ROLE_CHAINING to be present (roles=%v)", roles)
			}
		})
	}
}
