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

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/archival"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/gcp"
	gqlarchival "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/archival"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Example showing how to manage a GCP archival location with the SDK.
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

	// Add the GCP default project to RSC. Usually resolved using the
	// environment variable GOOGLE_APPLICATION_CREDENTIALS.
	id, err := gcpClient.AddProject(ctx, gcp.Default(), []core.Feature{core.FeatureCloudNativeArchival})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("RSC cloud account ID: %s\n", id)

	archivalClient := archival.Wrap(client)

	// Create a GCP archival location.
	targetMappingID, err := archivalClient.CreateGCPStorageSetting(ctx, gqlarchival.CreateGCPStorageSettingParams{
		CloudAccountID: id,
		Name:           "Test",
		BucketPrefix:   "my-prefix",
		StorageClass:   "STANDARD_GCP",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Target mapping ID: %v\n", targetMappingID)

	// Get the GCP archival location by ID.
	targetMapping, err := archivalClient.GCPTargetMappingByID(ctx, targetMappingID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID: %s, Name: %s\n", targetMapping.ID, targetMapping.Name)

	// Update the GCP archival location.
	err = archivalClient.UpdateGCPStorageSetting(ctx, targetMappingID, gqlarchival.UpdateGCPStorageSettingParams{Name: "Test-Updated"})
	if err != nil {
		log.Fatal(err)
	}

	// Search for a GCP archival location by a name prefix.
	targetMappings, err := archivalClient.GCPTargetMappings(ctx, "Test")
	if err != nil {
		log.Fatal(err)
	}
	for _, targetMapping := range targetMappings {
		fmt.Printf("ID: %s, Name: %s\n", targetMapping.ID, targetMapping.Name)
	}

	// Delete the GCP archival location.
	err = archivalClient.DeleteTargetMapping(ctx, targetMappingID)
	if err != nil {
		log.Fatal(err)
	}

	// Remove the GCP project from RSC.
	err = gcpClient.RemoveProject(ctx, id, []core.Feature{core.FeatureCloudNativeArchival}, false)
	if err != nil {
		log.Fatal(err)
	}
}
