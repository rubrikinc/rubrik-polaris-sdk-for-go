// Copyright 2023 Rubrik, Inc.
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
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// client is the common RSC client used for tests. By reusing the same client
// we reduce the risk of hitting rate limits when access tokens are created.
var client *polaris.Client

func TestMain(m *testing.M) {
	if testsetup.BoolEnvSet("TEST_INTEGRATION") {
		// Load configuration and create client. Usually resolved using the
		// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
		polAccount, err := polaris.DefaultServiceAccount(true)
		if err != nil {
			fmt.Printf("failed to get default service account: %v\n", err)
			os.Exit(1)
		}

		// The integration tests defaults the log level to INFO. Note that
		// RUBRIK_POLARIS_LOGLEVEL can be used to override this.
		logger := log.NewStandardLogger()
		logger.SetLogLevel(log.Info)
		if err := polaris.SetLogLevelFromEnv(logger); err != nil {
			fmt.Printf("failed to get log level from env: %v\n", err)
			os.Exit(1)
		}

		client, err = polaris.NewClientWithLogger(polAccount, logger)
		if err != nil {
			fmt.Printf("failed to create polaris client: %v\n", err)
			os.Exit(1)
		}

		version, err := client.GQL.DeploymentVersion(context.Background())
		if err != nil {
			fmt.Printf("failed to get deployment version: %v\n", err)
			os.Exit(1)
		}
		logger.Printf(log.Info, "Polaris version: %s", version)
	}

	os.Exit(m.Run())
}

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
	id, err := gcpClient.AddProject(ctx, Default(), core.FeatureCloudNativeProtection)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully added. ProjectID is compared
	// in a case-insensitive fashion due to a bug causing the initial project
	// id to be the same as the name.
	account, err := gcpClient.Project(ctx, CloudAccountID(id), core.FeatureCloudNativeProtection)
	if err != nil {
		t.Error(err)
	}
	if account.Name != testProject.ProjectName {
		t.Errorf("invalid name: %v", account.Name)
	}
	if strings.ToLower(account.NativeID) != testProject.ProjectID {
		t.Errorf("invalid project id: %v", account.NativeID)
	}
	if account.ProjectNumber != testProject.ProjectNumber {
		t.Errorf("invalid project number: %v", account.ProjectNumber)
	}
	if n := len(account.Features); n == 1 {
		if feature := account.Features[0].Feature; !feature.Equal(core.FeatureCloudNativeProtection) {
			t.Errorf("invalid feature name: %v", feature)
		}
		if status := account.Features[0].Status; status != core.StatusConnected {
			t.Errorf("invalid feature status: %v", status)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	// Verify that the Project function does not return a project given a prefix
	// of the project id.
	prefix := testProject.ProjectID[:len(testProject.ProjectID)/2]
	account, err = gcpClient.Project(ctx, ProjectID(prefix), core.FeatureCloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatalf("invalid error: %v", err)
	}

	// Remove GCP project from RSC keeping the snapshots.
	err = gcpClient.RemoveProject(ctx, ID(Default()), core.FeatureCloudNativeProtection, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully removed.
	_, err = gcpClient.Project(ctx, ID(Default()), core.FeatureCloudNativeProtection)
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
		core.FeatureCloudNativeProtection, Name(testProject.ProjectName), Organization(testProject.OrganizationName))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully added.
	account, err := gcpClient.Project(ctx, CloudAccountID(id), core.FeatureCloudNativeProtection)
	if err != nil {
		t.Error(err)
	}
	if account.Name != testProject.ProjectName {
		t.Errorf("invalid name: %v", account.Name)
	}
	if strings.ToLower(account.NativeID) != testProject.ProjectID {
		t.Errorf("invalid project id: %v", account.NativeID)
	}
	if account.ProjectNumber != testProject.ProjectNumber {
		t.Errorf("invalid project number: %v", account.ProjectNumber)
	}
	if n := len(account.Features); n == 1 {
		if feature := account.Features[0].Feature; !feature.Equal(core.FeatureCloudNativeProtection) {
			t.Errorf("invalid feature name: %v", feature)
		}
		if status := account.Features[0].Status; status != core.StatusConnected {
			t.Errorf("invalid feature status: %v", status)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	// Remove GCP project from RSC keeping the snapshots.
	err = gcpClient.RemoveProject(ctx, ProjectNumber(testProject.ProjectNumber), core.FeatureCloudNativeProtection, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully removed.
	_, err = gcpClient.Project(ctx, ProjectNumber(testProject.ProjectNumber), core.FeatureCloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}
