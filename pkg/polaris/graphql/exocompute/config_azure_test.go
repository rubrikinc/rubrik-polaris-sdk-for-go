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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/assert"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
)

func TestValidateAzureConfiguration(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer assert.Context(t, ctx, cancel)

	tmpl, err := template.ParseFiles("testdata/validate_azure_exocompute_configurations_response.tmpl")
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name                      string
		expectedError             string
		ErrorMessage              string
		BlockedSecurityRules      bool
		ClusterSubnetSizeTooSmall bool
		PodAndClusterSubnetSame   bool
		PodCidrRangeTooSmall      bool
		PodSubnetSizeTooSmall     bool
		SubnetDelegated           bool
	}{{
		name: "NoError",
	}, {
		name:          "ErrorMessage",
		expectedError: "error message",
		ErrorMessage:  "error message",
	}, {
		name:            "ErrorMessageAndValidationError",
		expectedError:   "multiple configuration errors:  (1) error message; (2) subnet is delegated",
		ErrorMessage:    "error message",
		SubnetDelegated: true,
	}, {
		name: "ErrorMessageAndMultipleValidationErrors",
		expectedError: "multiple configuration errors:  (1) subnet is delegated; (2) network security group has " +
			"blocking security rules; (3) cluster subnet size is too small; (4) pod and cluster subnets must be " +
			"different; (5) pod cidr range is too small; (6) pod subnet size is too small",
		BlockedSecurityRules:      true,
		ClusterSubnetSizeTooSmall: true,
		PodAndClusterSubnetSame:   true,
		PodCidrRangeTooSmall:      true,
		PodSubnetSizeTooSmall:     true,
		SubnetDelegated:           true,
	}, {
		name:            "ValidationError",
		expectedError:   "subnet is delegated",
		SubnetDelegated: true,
	}, {
		name: "MultipleValidationErrors",
		expectedError: "multiple configuration errors:  (1) subnet is delegated; (2) network security group has " +
			"blocking security rules; (3) cluster subnet size is too small; (4) pod and cluster subnets must be " +
			"different; (5) pod cidr range is too small; (6) pod subnet size is too small",
		BlockedSecurityRules:      true,
		ClusterSubnetSizeTooSmall: true,
		PodAndClusterSubnetSame:   true,
		PodCidrRangeTooSmall:      true,
		PodSubnetSizeTooSmall:     true,
		SubnetDelegated:           true,
	}}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewTLSServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
				if err := tmpl.Execute(w, tc); err != nil {
					cancel(err)
				}
			}))
			defer srv.Close()

			err = ValidateConfiguration(ctx, graphql.NewTestClient(srv), CreateAzureConfigurationParams{})
			if err == nil {
				if tc.expectedError != "" {
					t.Fatal("expected validate configuration call to fail")
				}
				return
			}
			if msg := err.Error(); !strings.HasSuffix(msg, tc.expectedError) {
				t.Errorf("invalid error: %s", err)
			}
		})
	}
}
