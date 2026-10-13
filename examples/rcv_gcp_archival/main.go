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

package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/archival"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlarchival "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/archival"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Example showing how to manage a Rubrik Cloud Vault GCP archival location
// with the SDK.
//
// Rubrik Cloud Vault is Rubrik hosted archival storage, so unlike the other
// cloud native archival locations no cloud account needs to be onboarded
// first. The GCP in the name is the cloud hosting the vault, not the cloud the
// archived workloads originate from.
//
// The RSC service account key file, identifying the RSC account, should be
// pointed out by the RUBRIK_POLARIS_SERVICEACCOUNT_FILE environment variable.
func main() {
	ctx := context.Background()

	// Load configuration and create the client.
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

	archivalClient := archival.Wrap(client)

	// Create an RCV GCP archival location in a specific region. Leaving the
	// region out stores the snapshots in the same region as the workload.
	targetMappingID, err := archivalClient.CreateRCVGCPStorageSetting(ctx, gqlarchival.CreateRCVGCPStorageSettingParams{
		Name:   "Test",
		Region: gcp.RegionFromName("us-west1").ToRCSRegionEnum(),
		Tier:   gqlarchival.RCVTierBackup,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Target mapping ID: %s\n", targetMappingID)

	// Get the RCV GCP archival location by ID.
	targetMapping, err := archivalClient.RCVGCPTargetMappingByID(ctx, targetMappingID)
	if err != nil {
		log.Fatal(err)
	}
	targetTemplate := targetMapping.TargetTemplate
	fmt.Printf("ID: %s, Name: %s, Region: %s, Tier: %s, Template: %s\n",
		targetMapping.ID, targetMapping.Name, targetTemplate.Region.Region.Name(), targetTemplate.Tier,
		targetTemplate.LocTemplate)

	// Get the RCV GCP archival location by name.
	if _, err := archivalClient.RCVGCPTargetMappingByName(ctx, "Test"); err != nil {
		log.Fatal(err)
	}

	// Update the RCV GCP archival location. Only the name can be updated.
	if err := archivalClient.UpdateRCVGCPStorageSetting(ctx, targetMappingID, gqlarchival.UpdateRCVGCPStorageSettingParams{
		Name: "TestUpdated",
	}); err != nil {
		log.Fatal(err)
	}

	// Search for an RCV GCP archival location by a name prefix.
	targetMappings, err := archivalClient.RCVGCPTargetMappings(ctx, "Test")
	if err != nil {
		log.Fatal(err)
	}
	for _, targetMapping := range targetMappings {
		fmt.Printf("ID: %s, Name: %s\n", targetMapping.ID, targetMapping.Name)
	}

	// Delete the RCV GCP archival location.
	if err := archivalClient.DeleteTargetMapping(ctx, targetMappingID); err != nil {
		log.Fatal(err)
	}

	// Reading it back now fails with graphql.ErrNotFound.
	if _, err := archivalClient.RCVGCPTargetMappingByID(ctx, targetMappingID); !errors.Is(err, graphql.ErrNotFound) {
		log.Fatalf("expected the archival location to be gone: %s", err)
	}
}
