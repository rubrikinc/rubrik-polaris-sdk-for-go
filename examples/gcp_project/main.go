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

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Example showing how to manage a GCP project with the SDK.
//
// The RSC service account key file, identifying the RSC account, should be
// pointed out by the RUBRIK_POLARIS_SERVICEACCOUNT_FILE environment variable.
func main() {
	ctx := context.Background()

	// Load configuration and create client.
	polAccount, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		log.Fatal(err)
	}
	logger := polarislog.NewStandardLogger()
	if err := polaris.SetLogLevelFromEnv(logger); err != nil {
		log.Fatal(err)
	}
	client, err := polaris.NewClientWithLogger(polAccount, logger)
	if err != nil {
		log.Fatal(err)
	}

	gcpClient := gcp.Wrap(client)

	features := []core.Feature{
		core.FeatureCloudNativeProtection.WithPermissionGroups("BASIC", "EXPORT_AND_RESTORE"),
		core.FeatureGCPSharedVPCHost.WithPermissionGroups("BASIC"),
	}

	// Add the GCP default project to Polaris. Usually resolved using the
	// environment variable GOOGLE_APPLICATION_CREDENTIALS.
	id, err := gcpClient.AddProject(ctx, gcp.Default(), features)
	if err != nil {
		log.Fatal(err)
	}

	// List the GCP projects added to Polaris.
	account, err := gcpClient.ProjectByID(ctx, id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Name: %s, ProjectID: %s\n", account.Name, account.ID)
	for _, feature := range account.Features {
		fmt.Printf("Feature: %s, Status: %s\n", feature, feature.Status)
	}

	// Remove the GCP project from RSC.
	err = gcpClient.RemoveProject(ctx, id, features, false)
	if err != nil {
		log.Fatal(err)
	}
}
