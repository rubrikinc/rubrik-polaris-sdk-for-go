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
	"errors"
	"strings"
	"testing"

	polaris_log "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Between the project has been added and it has been removed we never fail
// fatally to allow the project to be removed in case of an error.
func TestGcpProjectAddAndRemove(t *testing.T) {
	requireEnv(t, "SDK_INTEGRATION")

	ctx := context.Background()

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
	if project.Name != "Trinity-FDSE" {
		t.Errorf("invalid name: %v", project.Name)
	}
	if project.ProjectName != "Trinity-FDSE" {
		t.Errorf("invalid project name: %v", project.ProjectName)
	}
	if strings.ToLower(project.ProjectID) != "trinity-fdse" {
		t.Errorf("invalid project id: %v", project.ProjectID)
	}
	if project.ProjectNumber != 994761414559 {
		t.Errorf("invalid project number: %v", project.ProjectNumber)
	}
	if project.OrganizationName != "" {
		t.Errorf("invalid organization name: %v", project.OrganizationName)
	}
	if n := len(project.Features); n != 1 {
		t.Errorf("invalid number of features: %v", n)
	}
	if project.Features[0].Feature != "CLOUD_NATIVE_PROTECTION" {
		t.Errorf("invalid feature name: %v", project.Features[0].Feature)
	}
	if project.Features[0].Status != "CONNECTED" {
		t.Errorf("invalid feature status: %v", project.Features[0].Status)
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

func TestGcpProjectAddAndRemoveWithServiceAccountSet(t *testing.T) {
	requireEnv(t, "SDK_INTEGRATION")

	ctx := context.Background()

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
		FromGcpProject("trinity-fdse", "Trinity-FDSE", 994761414559, "Trinity Org"))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully added.
	project, err := client.GcpProject(ctx, WithGcpProjectNumber(994761414559))
	if err != nil {
		t.Error(err)
	}
	if project.Name != "Trinity-FDSE" {
		t.Errorf("invalid name: %v", project.Name)
	}
	if project.ProjectName != "Trinity-FDSE" {
		t.Errorf("invalid native id: %v", project.ProjectName)
	}
	if strings.ToLower(project.ProjectID) != "trinity-fdse" {
		t.Errorf("invalid native id: %v", project.ProjectID)
	}
	if project.ProjectNumber != 994761414559 {
		t.Errorf("invalid native id: %v", project.ProjectNumber)
	}
	if project.OrganizationName != "" {
		t.Errorf("invalid native id: %v", project.OrganizationName)
	}
	if n := len(project.Features); n != 1 {
		t.Errorf("invalid number of features: %v", n)
	}
	if project.Features[0].Feature != "CLOUD_NATIVE_PROTECTION" {
		t.Errorf("invalid feature name: %v", project.Features[0].Feature)
	}
	if project.Features[0].Status != "CONNECTED" {
		t.Errorf("invalid feature status: %v", project.Features[0].Status)
	}

	// Remove GCP project from Polaris keeping the snapshots.
	if err := client.GcpProjectRemove(ctx, WithGcpProjectNumber(994761414559), false); err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully removed.
	project, err = client.GcpProject(ctx, FromGcpDefault())
	if !errors.Is(err, ErrNotFound) {
		t.Error(err)
	}
}
