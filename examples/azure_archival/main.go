// Copyright 2024 Rubrik, Inc.
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

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/archival"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/azure"
	gqlarchival "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/archival"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Example showing how to manage an Azure archival location with the SDK.
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

	azureClient := azure.Wrap(client)

	// Add default Azure service principal to RSC. Usually resolved using
	// the environment variable AZURE_SERVICEPRINCIPAL_LOCATION.
	if _, err := azureClient.SetServicePrincipal(ctx, azure.Default("my-domain.onmicrosoft.com")); err != nil {
		log.Fatal(err)
	}

	// Add Azure subscription to RSC.
	subscription := azure.Subscription(uuid.MustParse("9318aeec-d357-11eb-9b37-5f4e9f79db5d"),
		"my-domain.onmicrosoft.com")
	id, err := azureClient.AddSubscription(ctx, subscription, core.FeatureCloudNativeArchival, azure.Regions("eastus2"),
		azure.ResourceGroup("terraform-test", "eastus2", nil))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("RSC cloud account ID: %v\n", id)

	archivalClient := archival.Wrap(client)

	// Create an Azure archival location.
	targetMappingID, err := archivalClient.CreateAzureStorageSetting(ctx, gqlarchival.CreateAzureStorageSettingParams{
		CloudAccountID:     id,
		ContainerName:      "my-container-name",
		Name:               "Test",
		Redundancy:         "LRS",
		StorageAccountName: "mystorageaccount",
		StorageTier:        "COOL",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Target mapping ID: %v\n", targetMappingID)

	// Get the Azure archival location by ID.
	targetMapping, err := archivalClient.AzureTargetMappingByID(ctx, targetMappingID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID: %v, Name: %s\n", targetMapping.ID, targetMapping.Name)

	// Update the Azure archival location.
	if err := archivalClient.UpdateAzureStorageSetting(ctx, targetMappingID, gqlarchival.UpdateAzureStorageSettingParams{
		Name: "TestUpdated",
	}); err != nil {
		log.Fatal(err)
	}

	// Search for an Azure archival location by a name prefix.
	targetMappings, err := archivalClient.AzureTargetMappings(ctx, "Test")
	if err != nil {
		log.Fatal(err)
	}
	for _, targetMapping := range targetMappings {
		fmt.Printf("ID: %v, Name: %s\n", targetMapping.ID, targetMapping.Name)
	}

	// Delete the Azure archival location.
	if err := archivalClient.DeleteTargetMapping(ctx, targetMappingID); err != nil {
		log.Fatal(err)
	}

	// Remove the Azure subscription from RSC.
	if err := azureClient.RemoveSubscription(ctx, id, core.FeatureCloudNativeArchival, false); err != nil {
		log.Fatal(err)
	}
}
