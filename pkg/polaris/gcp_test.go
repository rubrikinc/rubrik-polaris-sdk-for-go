// Copyright 2021 Rubrik, Inc.
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

package polaris

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	polaris_log "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// testGcpProject holds information about the GCP project used in the
// integration tests. Normally used to assert that the project information read
// from Polaris is correct.
type testGcpProject struct {
	Name             string `json:"name"`
	ProjectName      string `json:"projectName"`
	ProjectID        string `json:"projectId"`
	ProjectNumber    int64  `json:"projectNumber"`
	OrganizationName string `json:"organizationName"`
}

// TestGcpProjectAddAndRemove verifies that the SDK can perform the basic GCP
// project operations on a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * SDK_INTEGRATION=1
//   * SDK_GCPPROJECT_FILE=<path-to-test-gcp-project-file>
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * GOOGLE_APPLICATION_CREDENTIALS=<path-to-gcp-service-account-key-file>
//
// The file referred to by SDK_GCPPROJECT_FILE should contain a single
// testGcpProject JSON object.
//
// Note that between the project has been added and it has been removed we
// never fail fatally to allow the project to be removed in case of an error.
func TestGcpProjectAddAndRemove(t *testing.T) {
	requireEnv(t, "SDK_INTEGRATION")

	ctx := context.Background()

	// Load test project information from the file pointed to by the
	// SDK_GCPPROJECT_FILE environment variable.
	buf, err := os.ReadFile(os.Getenv("SDK_GCPPROJECT_FILE"))
	if err != nil {
		t.Fatalf("failed to read file pointed to by SDK_GCPPROJECT_FILE: %v", err)
	}
	testProject := testGcpProject{}
	if err := json.Unmarshal(buf, &testProject); err != nil {
		t.Fatal(err)
	}

	// Load configuration and create client. Usually resolved using the
	// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
	polAccount, err := DefaultServiceAccount()
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClientFromServiceAccount(polAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	// Add the default GCP project to Polaris. Usually resolved using the
	// environment variable GOOGLE_APPLICATION_CREDENTIALS.
	err = client.GcpProjectAdd(ctx, FromGcpDefault())
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully added. ProjectID is compared
	// in a case-insensitive fashion due to a bug causing the initial project
	// id to be the same as the name.
	project, err := client.GcpProject(ctx, FromGcpDefault())
	if err != nil {
		t.Error(err)
	}
	if project.Name != testProject.Name {
		t.Errorf("invalid name: %v", project.Name)
	}
	if project.ProjectName != testProject.ProjectName {
		t.Errorf("invalid project name: %v", project.ProjectName)
	}
	if strings.ToLower(project.ProjectID) != testProject.ProjectID {
		t.Errorf("invalid project id: %v", project.ProjectID)
	}
	if project.ProjectNumber != testProject.ProjectNumber {
		t.Errorf("invalid project number: %v", project.ProjectNumber)
	}
	if project.OrganizationName != testProject.OrganizationName {
		t.Errorf("invalid organization name: %v", project.OrganizationName)
	}
	if n := len(project.Features); n == 1 {
		if project.Features[0].Feature != "CLOUD_NATIVE_PROTECTION" {
			t.Errorf("invalid feature name: %v", project.Features[0].Feature)
		}
		if project.Features[0].Status != "CONNECTED" {
			t.Errorf("invalid feature status: %v", project.Features[0].Status)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	// Remove GCP project from Polaris keeping the snapshots.
	if err := client.GcpProjectRemove(ctx, FromGcpDefault(), false); err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully removed.
	project, err = client.GcpProject(ctx, FromGcpDefault())
	if !errors.Is(err, ErrNotFound) {
		t.Error(err)
	}
}

// TestGcpProjectAddAndRemoveWithServiceAccountSet verifies that the SDK can
// perform the basic GCP project operations on a real Polaris instance using a
// Polaris account global GCP service account.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * SDK_INTEGRATION=1
//   * SDK_GCPPROJECT_FILE=<path-to-test-gcp-project-file>
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * GOOGLE_APPLICATION_CREDENTIALS=<path-to-gcp-service-account-key-file>
//
// The file referred to by SDK_GCPPROJECT_FILE should contain a single
// testGcpProject JSON object.
//
// Note that between the project has been added and it has been removed we
// never fail fatally to allow the project to be removed in case of an error.
func TestGcpProjectAddAndRemoveWithServiceAccountSet(t *testing.T) {
	requireEnv(t, "SDK_INTEGRATION")

	ctx := context.Background()

	// Load test project information from the file pointed to by the
	// SDK_GCPPROJECT_FILE environment variable.
	buf, err := os.ReadFile(os.Getenv("SDK_GCPPROJECT_FILE"))
	if err != nil {
		t.Fatalf("failed to read file pointed to by SDK_GCPPROJECT_FILE: %v", err)
	}

	testProject := testGcpProject{}
	if err := json.Unmarshal(buf, &testProject); err != nil {
		t.Fatal(err)
	}

	// Load configuration and create client. Usually resolved using the
	// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
	polAccount, err := DefaultServiceAccount()
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClientFromServiceAccount(polAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	// Add the service account to Polaris.
	if err = client.GcpServiceAccountSet(ctx, FromGcpDefault()); err != nil {
		t.Fatal(err)
	}

	// Add the default GCP project to Polaris. Usually resolved using the
	// environment variable GOOGLE_APPLICATION_CREDENTIALS.
	err = client.GcpProjectAdd(ctx,
		FromGcpProject(testProject.ProjectID, testProject.ProjectName, testProject.ProjectNumber,
			testProject.OrganizationName))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully added.
	project, err := client.GcpProject(ctx, WithGcpProjectNumber(testProject.ProjectNumber))
	if err != nil {
		t.Error(err)
	}
	if project.Name != testProject.Name {
		t.Errorf("invalid name: %v", project.Name)
	}
	if project.ProjectName != testProject.ProjectName {
		t.Errorf("invalid project name: %v", project.ProjectName)
	}
	if strings.ToLower(project.ProjectID) != testProject.ProjectID {
		t.Errorf("invalid project id: %v", project.ProjectID)
	}
	if project.ProjectNumber != testProject.ProjectNumber {
		t.Errorf("invalid project number: %v", project.ProjectNumber)
	}
	if project.OrganizationName != testProject.OrganizationName {
		t.Errorf("invalid organization name: %v", project.OrganizationName)
	}
	if n := len(project.Features); n == 1 {
		if project.Features[0].Feature != "CLOUD_NATIVE_PROTECTION" {
			t.Errorf("invalid feature name: %v", project.Features[0].Feature)
		}
		if project.Features[0].Status != "CONNECTED" {
			t.Errorf("invalid feature status: %v", project.Features[0].Status)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	// Remove GCP project from Polaris keeping the snapshots.
	err = client.GcpProjectRemove(ctx, WithGcpProjectNumber(testProject.ProjectNumber), false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully removed.
	project, err = client.GcpProject(ctx, WithGcpProjectNumber(testProject.ProjectNumber))
	if !errors.Is(err, ErrNotFound) {
		t.Error(err)
	}
}
