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
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// TestAllPermissionPoliciesSendsBothFeatureLists verifies that a request using
// permission groups names the features in the plain features list too. Part of
// a workaround for RSC deciding CLOUD_COST_REPORT applicability from the plain
// list only.
func TestAllPermissionPoliciesSendsBothFeatureLists(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	type wire struct {
		Features       []string       `json:"features"`
		FeaturesWithPG []core.Feature `json:"featuresWithPG"`
	}
	var got wire
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			cancel(err)
			return
		}
		var payload struct {
			Variables wire `json:"variables"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			cancel(err)
			return
		}
		got = payload.Variables
		if _, err := w.Write([]byte(`{"data":{"result":[]}}`)); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	cnp := core.FeatureCloudNativeProtection.WithPermissionGroups(core.PermissionGroupBasic)
	ccr := core.FeatureCloudCostReport.WithPermissionGroups(core.PermissionGroupBasic)

	tests := []struct {
		name           string
		features       []core.Feature
		wantFeatures   []string
		wantFeaturesPG []core.Feature
	}{{
		name:           "permission groups are also sent as plain features",
		features:       []core.Feature{cnp, ccr},
		wantFeatures:   []string{cnp.Name, ccr.Name},
		wantFeaturesPG: []core.Feature{cnp, ccr},
	}, {
		name:         "plain features are sent unchanged",
		features:     []core.Feature{core.FeatureCloudNativeProtection, core.FeatureCloudCostReport},
		wantFeatures: []string{core.FeatureCloudNativeProtection.Name, core.FeatureCloudCostReport.Name},
	}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got = wire{}
			if _, err := Wrap(graphql.NewTestClient(srv)).AllPermissionPolicies(ctx, CloudStandard, tc.features, ""); err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got.Features, tc.wantFeatures) {
				t.Errorf("expected features %v, got %v", tc.wantFeatures, got.Features)
			}
			if len(got.FeaturesWithPG) != len(tc.wantFeaturesPG) {
				t.Fatalf("expected %d features with permission groups, got %d", len(tc.wantFeaturesPG), len(got.FeaturesWithPG))
			}
			for i, feature := range tc.wantFeaturesPG {
				if !got.FeaturesWithPG[i].DeepEqual(feature) {
					t.Errorf("expected feature with permission groups %v, got %v", feature, got.FeaturesWithPG[i])
				}
			}
		})
	}
}

// TestAllPermissionPoliciesSendsNativeID verifies that the native ID is sent
// as awsNativeId when given, and left out of the request when empty, so that
// RSC renders the policies for a single-role account.
func TestAllPermissionPoliciesSendsNativeID(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	var got map[string]json.RawMessage
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			cancel(err)
			return
		}
		var payload struct {
			Variables map[string]json.RawMessage `json:"variables"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			cancel(err)
			return
		}
		got = payload.Variables
		if _, err := w.Write([]byte(`{"data":{"result":[]}}`)); err != nil {
			cancel(err)
		}
	}))
	defer srv.Close()

	tests := []struct {
		name     string
		nativeID string
		want     string
	}{{
		name:     "native ID is sent",
		nativeID: "123456789012",
		want:     `"123456789012"`,
	}, {
		name: "empty native ID is not sent",
	}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got = nil
			_, err := Wrap(graphql.NewTestClient(srv)).AllPermissionPoliciesWithParams(ctx, AllPermissionPoliciesParams{
				Cloud:    CloudStandard,
				Features: []core.Feature{core.FeatureCloudNativeProtection},
				NativeID: tc.nativeID,
			})
			if err != nil {
				t.Fatal(err)
			}
			nativeID, ok := got["awsNativeId"]
			if tc.want == "" {
				if ok {
					t.Errorf("expected no awsNativeId, got %s", nativeID)
				}
				return
			}
			if string(nativeID) != tc.want {
				t.Errorf("expected awsNativeId %s, got %s", tc.want, nativeID)
			}
		})
	}
}
