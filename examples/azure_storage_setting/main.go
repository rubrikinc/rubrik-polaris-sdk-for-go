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

	// Add default Azure service principal to Polaris. Usually resolved using
	// the environment variable AZURE_SERVICEPRINCIPAL_LOCATION.
	_, err = azure.Wrap(client).SetServicePrincipal(ctx, azure.Default("my-domain.onmicrosoft.com"))
	if err != nil {
		log.Fatal(err)
	}

	// Add Azure subscription to Polaris.
	subscription := azure.Subscription(uuid.MustParse("9318aeec-d357-11eb-9b37-5f4e9f79db5d"),
		"my-domain.onmicrosoft.com")
	id, err := azure.Wrap(client).AddSubscription(ctx, subscription, core.FeatureCloudNativeArchival, azure.Regions("eastus2"),
		azure.ResourceGroup("terraform-test", "eastus2", nil))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("RSC cloud account ID: %v\n", id)

	// Create an Azure archival location.
	targetMappingID, err := archival.Wrap(client).CreateAzureStorageSetting(ctx, gqlarchival.CreateAzureStorageSettingParams{
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

	// Get the AWS archival location by ID.
	targetMapping, err := archival.Wrap(client).AzureTargetMappingByID(ctx, targetMappingID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID: %v, Name: %s\n", targetMapping.ID, targetMapping.Name)

	// Update the AWS archival location.
	err = archival.Wrap(client).UpdateAzureStorageSetting(ctx, targetMappingID, gqlarchival.UpdateAzureStorageSettingParams{
		Name: "TestUpdated",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Search for an AWS archival location by a name prefix.
	targetMappings, err := archival.Wrap(client).AzureTargetMappings(ctx, "Test")
	if err != nil {
		log.Fatal(err)
	}
	for _, targetMapping := range targetMappings {
		fmt.Printf("ID: %v, Name: %s\n", targetMapping.ID, targetMapping.Name)
	}

	// Delete the Azure archival location.
	err = archival.Wrap(client).DeleteTargetMapping(ctx, targetMappingID)
	if err != nil {
		log.Fatal(err)
	}

	// Remove the Azure subscription from RSC.
	err = azure.Wrap(client).RemoveSubscription(ctx, azure.CloudAccountID(id), core.FeatureCloudNativeArchival, false)
	if err != nil {
		log.Fatal(err)
	}
}
