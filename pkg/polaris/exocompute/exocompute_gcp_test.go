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
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	gqlgcp "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp"
)

// TestGCPExocompute verifies that the SDK can perform basic Exocompute
// operations on a real RSC instance.
//
// To run this test against an RSC instance the following environment variables
// needs to be set:
//   - RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   - TEST_INTEGRATION=1
//   - TEST_GCPPROJECT_FILE=<path-to-test-gcp-project-file>
//   - GOOGLE_APPLICATION_CREDENTIALS=<path-to-gcp-service-account-key-file>
//
// The file referred to by TEST_GCPPROJECT_FILE should contain a single
// testGcpProject JSON object.
func TestGCPExocompute(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	testProject, err := testsetup.GCPProject()
	if err != nil {
		t.Fatal(err)
	}

	// Add GCP project to RSC.
	features := []core.Feature{core.FeatureCloudNativeProtection.WithPermissionGroups(core.PermissionGroupBasic),
		core.FeatureExocompute.WithPermissionGroups(core.PermissionGroupBasic)}
	cloudAccountID, err := gcp.Wrap(client).AddProject(ctx,
		gcp.Project(testProject.ProjectID, testProject.ProjectNumber), features, gcp.Name(testProject.ProjectName))
	if err != nil {
		t.Fatal(err)
	}

	// Lookup the GCP project using the cloud account ID.
	account, err := gcp.Wrap(client).ProjectByID(ctx, cloudAccountID)
	if err != nil {
		t.Fatal(err)
	}
	if id := account.NativeID; id != testProject.ProjectID {
		t.Fatalf("invalid native id: %s", id)
	}
	if name := account.Name; name != testProject.ProjectName {
		t.Fatalf("invalid name: %s", name)
	}
	if number := account.ProjectNumber; number != testProject.ProjectNumber {
		t.Fatalf("invalid project number: %d", number)
	}
	if n := len(account.Features); n != 2 {
		t.Fatalf("invalid number of features: %d", n)
	}
	cnpFeature, ok := account.Feature(core.FeatureCloudNativeProtection)
	if !ok {
		t.Fatalf("%s feature not found", core.FeatureCloudNativeProtection)
	}
	if name := cnpFeature.Name; name != core.FeatureCloudNativeProtection.Name {
		t.Fatalf("invalid feature name: %s", name)
	}
	if status := cnpFeature.Status; status != "CONNECTED" {
		t.Fatalf("invalid feature status: %s", status)
	}
	exoFeature, ok := account.Feature(core.FeatureExocompute)
	if !ok {
		t.Fatalf("%s feature not found", core.FeatureExocompute)
	}
	if name := exoFeature.Name; name != core.FeatureExocompute.Name {
		t.Fatalf("invalid feature name: %s", name)
	}
	if status := exoFeature.Status; status != "CONNECTED" {
		t.Fatalf("invalid feature status: %s", status)
	}

	// Add an exocompute configuration to the cloud account.
	err = Wrap(client).UpdateGCPConfiguration(ctx, cloudAccountID, []RegionalConfig{{
		Region:         gqlgcp.RegionFromName(testProject.Exocompute.Region),
		SubnetName:     testProject.Exocompute.SubnetName,
		VPCNetworkName: testProject.Exocompute.VPCNetworkName,
	}}, false)
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve the exocompute configuration just added.
	exoConfigs, err := Wrap(client).GCPConfigurationsByCloudAccountID(ctx, cloudAccountID, false)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(exoConfigs); n != 1 {
		t.Fatalf("invalid number of exocompute configurations: %d", n)
	}
	if id := exoConfigs[0].ID; id == uuid.Nil {
		t.Fatalf("invalid exocompute config id: %s", id)
	}
	if region := exoConfigs[0].Config.Region; region.Region != gqlgcp.RegionFromName(testProject.Exocompute.Region) {
		t.Fatalf("invalid exocompute config subnet name: %s", region)
	}
	if name := exoConfigs[0].Config.SubnetName; name != testProject.Exocompute.SubnetName {
		t.Fatalf("invalid exocompute config subnet name: %s", name)
	}
	if name := exoConfigs[0].Config.VPCNetworkName; name != testProject.Exocompute.VPCNetworkName {
		t.Fatalf("invalid exocompute config vpc network name: %s", name)
	}
	if status := exoConfigs[0].HealthCheckStatus; status != nil {
		t.Fatalf("invalid exocompute config health check status: %v", status)
	}

	// Remove the exocompute configuration.
	if err := Wrap(client).RemoveGCPConfiguration(ctx, cloudAccountID); err != nil {
		t.Fatal(err)
	}

	// Verify that the exocompute configuration was successfully removed.
	if _, err = Wrap(client).GCPConfigurationsByCloudAccountID(ctx, cloudAccountID, false); !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}

	// Remove cloud account from RSC.
	if err := gcp.Wrap(client).RemoveProject(ctx, cloudAccountID, features, false); err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	if _, err := gcp.Wrap(client).ProjectByID(ctx, cloudAccountID); !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}
