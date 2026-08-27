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
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// serviceTypeCase pairs an input service type with the value expected on the
// wire. The unspecified value must be substituted with the non-BaaS value since
// the backend does not accept an unspecified service type.
type serviceTypeCase struct {
	name     string
	input    CloudAccountServiceType
	wantWire string
}

func serviceTypeCases() []serviceTypeCase {
	return []serviceTypeCase{
		{"unspecified defaults to non-baas", ServiceTypeUnspecified, "AWS_CLOUD_ACCOUNT_SERVICE_TYPE_NON_BAAS"},
		{"non-baas", ServiceTypeNonBaaS, "AWS_CLOUD_ACCOUNT_SERVICE_TYPE_NON_BAAS"},
		{"baas", ServiceTypeBaaS, "AWS_CLOUD_ACCOUNT_SERVICE_TYPE_BAAS"},
	}
}

// captureServiceType returns a test server that records the serviceType sent in
// the request variables and replies with the given response body.
func captureServiceType(t *testing.T, cancel context.CancelCauseFunc, gotWire *string, response string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			cancel(err)
			return
		}
		var payload struct {
			Variables struct {
				ServiceType string `json:"serviceType"`
			} `json:"variables"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			cancel(err)
			return
		}
		*gotWire = payload.Variables.ServiceType
		if _, err := io.WriteString(w, response); err != nil {
			cancel(err)
		}
	}))
}

func TestValidateAndCreateCloudAccountServiceType(t *testing.T) {
	const response = `{"data":{"result":{"initiateResponse":{"stackName":"stack"},"validateResponse":{"invalidAwsAccounts":[]}}}}`

	for _, tt := range serviceTypeCases() {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancelCause(context.Background())
			defer assert.Context(t, ctx, cancel)

			var gotWire string
			srv := captureServiceType(t, cancel, &gotWire, response)
			defer srv.Close()

			_, err := Wrap(graphql.NewTestClient(srv)).ValidateAndCreateCloudAccount(ctx, CloudStandard,
				"123456789012", "test-account", []core.Feature{core.FeatureCloudNativeProtection}, "", tt.input)
			if err != nil {
				t.Fatalf("ValidateAndCreateCloudAccount returned error: %v", err)
			}
			if gotWire != tt.wantWire {
				t.Errorf("serviceType sent = %q, want %q", gotWire, tt.wantWire)
			}
		})
	}
}

func TestFinalizeCloudAccountProtectionServiceType(t *testing.T) {
	// An empty top-level message with a single child account that has no error
	// message is treated as success by FinalizeCloudAccountProtection.
	const response = `{"data":{"finalizeAwsCloudAccountProtection":{"awsChildAccounts":[{"nativeId":"123456789012","message":""}],"message":""}}}`

	for _, tt := range serviceTypeCases() {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancelCause(context.Background())
			defer assert.Context(t, ctx, cancel)

			var gotWire string
			srv := captureServiceType(t, cancel, &gotWire, response)
			defer srv.Close()

			err := Wrap(graphql.NewTestClient(srv)).FinalizeCloudAccountProtection(ctx, FinalizeCloudAccountProtectionParams{
				Cloud:       CloudStandard,
				NativeID:    "123456789012",
				Name:        "test-account",
				Features:    []core.Feature{core.FeatureCloudNativeProtection},
				ServiceType: tt.input,
			})
			if err != nil {
				t.Fatalf("FinalizeCloudAccountProtection returned error: %v", err)
			}
			if gotWire != tt.wantWire {
				t.Errorf("serviceType sent = %q, want %q", gotWire, tt.wantWire)
			}
		})
	}
}
