// Copyright 2025 Rubrik, Inc.
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

package gcp

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// TestGcpProjectAddAndRemove verifies that the SDK can perform the basic GCP
// project operations on a real RSC instance.
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
func TestGcpProjectAddAndRemove(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	gcpClient := Wrap(client)

	testProject, err := testsetup.GCPProject()
	if err != nil {
		t.Fatal(err)
	}

	// Add the default GCP project to RSC. Usually resolved using the
	// environment variable GOOGLE_APPLICATION_CREDENTIALS.
	id, err := gcpClient.AddProject(ctx, Default(), []core.Feature{core.FeatureCloudNativeProtection, core.FeatureCloudDiscovery})
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully added. ProjectID is compared
	// in a case-insensitive fashion due to a bug causing the initial project
	// id to be the same as the name.
	account, err := gcpClient.ProjectByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != testProject.ProjectName {
		t.Fatalf("invalid name: %v", account.Name)
	}
	if strings.ToLower(account.NativeID) != testProject.ProjectID {
		t.Fatalf("invalid project id: %v", account.NativeID)
	}
	if account.ProjectNumber != testProject.ProjectNumber {
		t.Fatalf("invalid project number: %v", account.ProjectNumber)
	}
	if n := len(account.Features); n != 2 {
		t.Fatalf("invalid number of features: %v", n)
	}
	for _, feature := range []core.Feature{core.FeatureCloudNativeProtection, core.FeatureCloudDiscovery} {
		f, ok := account.Feature(feature)
		if !ok {
			t.Fatalf("missing feature: %v", feature)
		}
		if f.Status != core.StatusConnected {
			t.Fatalf("invalid status for feature %v: %v", feature, f.Status)
		}
	}

	// Verify that the Project function does not return a project given a prefix
	// of the project id.
	prefix := testProject.ProjectID[:len(testProject.ProjectID)/2]
	account, err = gcpClient.ProjectByNativeID(ctx, prefix)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatalf("invalid error: %v", err)
	}

	// Verify that Cloud Discovery cannot be removed on its own while Cloud
	// Native Protection is enabled, and that nothing was removed.
	err = gcpClient.RemoveProject(ctx, id, []core.Feature{core.FeatureCloudDiscovery}, false)
	if err == nil {
		t.Fatal("expected removing Cloud Discovery while Cloud Native Protection is enabled to fail")
	}
	account, err = gcpClient.ProjectByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(account.Features); n != 2 {
		t.Fatalf("invalid number of features after refused removal: %v", n)
	}

	// Remove GCP project from RSC keeping the snapshots. Cloud Discovery is
	// listed first to verify that it is still removed last.
	err = gcpClient.RemoveProject(ctx, id, []core.Feature{core.FeatureCloudDiscovery, core.FeatureCloudNativeProtection}, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully removed.
	_, err = gcpClient.ProjectByID(ctx, id)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestGcpProjectAddAndRemoveWithServiceAccountSet verifies that the SDK can
// perform the basic GCP project operations on a real RSC instance using an RSC
// account global GCP service account.
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
func TestGcpProjectAddAndRemoveWithServiceAccountSet(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	testProject, err := testsetup.GCPProject()
	if err != nil {
		t.Fatal(err)
	}

	gcpClient := Wrap(client)

	// Add the service account to RSC.
	err = gcpClient.SetServiceAccount(ctx, Default())
	if err != nil {
		t.Fatal(err)
	}

	// Add the default GCP project to RSC. Usually resolved using the
	// environment variable GOOGLE_APPLICATION_CREDENTIALS.
	id, err := gcpClient.AddProject(ctx, Project(testProject.ProjectID, testProject.ProjectNumber),
		[]core.Feature{core.FeatureCloudNativeProtection, core.FeatureCloudDiscovery}, Name(testProject.ProjectName), Organization(testProject.OrganizationName))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully added.
	account, err := gcpClient.ProjectByID(ctx, id)
	if err != nil {
		t.Error(err)
	}
	if account.Name != testProject.ProjectName {
		t.Fatalf("invalid name: %v", account.Name)
	}
	if strings.ToLower(account.NativeID) != testProject.ProjectID {
		t.Fatalf("invalid project id: %v", account.NativeID)
	}
	if account.ProjectNumber != testProject.ProjectNumber {
		t.Fatalf("invalid project number: %v", account.ProjectNumber)
	}
	if n := len(account.Features); n != 2 {
		t.Fatalf("invalid number of features: %v", n)
	}
	for _, feature := range []core.Feature{core.FeatureCloudNativeProtection, core.FeatureCloudDiscovery} {
		f, ok := account.Feature(feature)
		if !ok {
			t.Fatalf("missing feature: %v", feature)
		}
		if f.Status != core.StatusConnected {
			t.Fatalf("invalid status for feature %v: %v", feature, f.Status)
		}
	}

	// Remove GCP project from RSC keeping the snapshots.
	err = gcpClient.RemoveProject(ctx, id, []core.Feature{core.FeatureCloudNativeProtection, core.FeatureCloudDiscovery}, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully removed.
	_, err = gcpClient.ProjectByID(ctx, id)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestSupportedFeaturesIncludesCloudSQL guards against the Cloud SQL
// protection feature being dropped from SupportedFeatures. The list is the
// SDK's answer to which features a GCP project can be onboarded with, so a
// missing entry misleads callers that enumerate it rather than failing loudly.
func TestSupportedFeaturesIncludesCloudSQL(t *testing.T) {
	if _, ok := core.LookupFeature(core.GCPProjectFeatures(), core.FeatureCloudSQLProtection); !ok {
		t.Error("CLOUD_SQL_PROTECTION missing from SupportedFeatures")
	}
}

// TestSupportedFeaturesIncludesUEMKeyManagement guards against the unified
// encryption key management feature being dropped from SupportedFeatures.
// Features missing from the list are silently discarded when a cloud account
// is read, hiding a feature the account actually has enabled.
func TestSupportedFeaturesIncludesUEMKeyManagement(t *testing.T) {
	if _, ok := core.LookupFeature(SupportedFeatures(), core.FeatureCloudNativeUEMKeyManagement); !ok {
		t.Error("CLOUD_NATIVE_UEM_KEY_MANAGEMENT missing from SupportedFeatures")
	}
}
