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

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/handler"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	kms "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/kms"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp"
)

// TestCreateKeyManagerInvalidRegion verifies that the unknown region, the
// source region and a GCP region which isn't a GCP KMS location are rejected
// without a request being sent to RSC.
func TestCreateKeyManagerInvalidRegion(t *testing.T) {
	srv := httptest.NewServer(handler.GraphQL(func(w http.ResponseWriter, req *http.Request) {
		t.Error("unexpected request to RSC")
	}))
	defer srv.Close()

	for _, region := range []aws.Region{aws.RegionUnknown, aws.RegionSource} {
		_, err := WrapGQL(graphql.NewTestClient(srv)).CreateAWSKeyManager(t.Context(), kms.CreateAWSKeyManagerParams{
			CloudAccountID: "b78f366d-d79e-48d0-9a7a-a98ef63755bf",
			Name:           "aws-key-manager",
			Region:         region,
		})
		if err == nil {
			t.Errorf("expected CreateAWSKeyManager to fail for region %q", region.DisplayName())
		}
	}

	for _, region := range []azure.Region{azure.RegionUnknown, azure.RegionSource} {
		_, err := WrapGQL(graphql.NewTestClient(srv)).CreateAzureKeyManager(t.Context(), kms.CreateAzureKeyManagerParams{
			CloudAccountID: "b78f366d-d79e-48d0-9a7a-a98ef63755bf",
			Name:           "azure-key-manager",
			Region:         region,
			VaultName:      "example-vault",
		})
		if err == nil {
			t.Errorf("expected CreateAzureKeyManager to fail for region %q", region.DisplayName())
		}
	}

	// RegionEU is a Cloud Storage multi-region, not a GCP KMS location.
	cloudAccountID := uuid.MustParse("167d3922-2630-470a-9285-3d6246dc194a")
	for _, region := range []gcp.Region{gcp.RegionUnknown, gcp.RegionEU} {
		_, err := WrapGQL(graphql.NewTestClient(srv)).CreateGCPKeyManager(t.Context(), kms.CreateGCPKeyManagerParams{
			CloudAccountID: "b78f366d-d79e-48d0-9a7a-a98ef63755bf",
			Name:           "gcp-key-manager",
			KeyRing:        "example-key-ring",
			Location:       gcp.KMSLocationEnum{Region: region},
		})
		if err == nil {
			t.Errorf("expected CreateGCPKeyManager to fail for location %q", region.DisplayName())
		}

		if _, err := WrapGQL(graphql.NewTestClient(srv)).GCPKeyRings(t.Context(), cloudAccountID, region); err == nil {
			t.Errorf("expected GCPKeyRings to fail for location %q", region.DisplayName())
		}
	}
}
