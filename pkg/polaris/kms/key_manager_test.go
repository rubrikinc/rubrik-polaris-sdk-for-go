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

package kms

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	kms "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/kms"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
)

// TestCreateKeyManagerInvalidRegion verifies that the unknown region and the
// source region are rejected without a request being sent to RSC.
func TestCreateKeyManagerInvalidRegion(t *testing.T) {
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		t.Error("unexpected request to RSC")
	}))
	defer srv.Close()

	for _, region := range []aws.Region{aws.RegionUnknown, aws.RegionSource} {
		_, err := WrapGQL(graphql.NewTestClient(srv)).CreateAWSKeyManager(t.Context(), kms.CreateAWSKeyManagerParams{
			CloudAccountID: "aa11bb22-cc33-dd44-ee55-ff6677889900",
			Name:           "aws-key-manager",
			Region:         region,
		})
		if err == nil {
			t.Errorf("expected CreateAWSKeyManager to fail for region %q", region.DisplayName())
		}
	}

	for _, region := range []azure.Region{azure.RegionUnknown, azure.RegionSource} {
		_, err := WrapGQL(graphql.NewTestClient(srv)).CreateAzureKeyManager(t.Context(), kms.CreateAzureKeyManagerParams{
			CloudAccountID: "aa11bb22-cc33-dd44-ee55-ff6677889900",
			Name:           "azure-key-manager",
			Region:         region,
			VaultName:      "example-vault",
		})
		if err == nil {
			t.Errorf("expected CreateAzureKeyManager to fail for region %q", region.DisplayName())
		}
	}
}
