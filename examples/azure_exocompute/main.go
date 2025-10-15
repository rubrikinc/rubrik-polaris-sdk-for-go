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

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/exocompute"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	gqlazure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

func main() {
	ctx := context.Background()

	// Load configuration and create a client.
	polAccount, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		log.Fatal(err)
	}
	client, err := polaris.NewClientWithLogger(polAccount, polarislog.NewStandardLogger())
	if err != nil {
		log.Fatal(err)
	}

	azureClient := azure.Wrap(client)
	exoClient := exocompute.Wrap(client)

	// Add default Azure service principal to RSC. Usually resolved using the
	// environment variable AZURE_SERVICEPRINCIPAL_LOCATION.
	_, err = azureClient.SetServicePrincipal(ctx, azure.Default("my-domain.onmicrosoft.com"))
	if err != nil {
		log.Fatal(err)
	}

	// Add Azure subscription to RSC.
	subscription := azure.Subscription(uuid.MustParse("9318aeec-d357-11eb-9b37-5f4e9f79db5d"),
		"my-domain.onmicrosoft.com")
	accountID, err := azureClient.AddSubscription(ctx, subscription, core.FeatureCloudNativeProtection, azure.Regions("eastus2", "westus2"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Account ID: %v\n", accountID)

	// Enable the exocompute feature for the account.
	exoAccountID, err := azureClient.AddSubscription(ctx, subscription, core.FeatureExocompute, azure.Regions("eastus2"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Exocompute Account ID: %v\n", exoAccountID)

	account, err := azureClient.Subscription(ctx, azure.CloudAccountID(accountID), core.FeatureAll)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %v, NativeID: %v\n", account.Name, account.NativeID)
	for _, feature := range account.Features {
		fmt.Printf("Feature: %v, Regions: %v, Status: %v\n", feature.Name, feature.Regions, feature.Status)
	}

	// Add an exocompute configuration to the cloud account.
	configID, err := exoClient.AddAzureConfiguration(ctx, accountID,
		exocompute.AzureManaged(gqlazure.RegionEastUS2, "/subscriptions/9318aeec-d357-11eb-9b37-5f4e9f79db5d/resourceGroups/terraform-test/providers/Microsoft.Network/virtualNetworks/terraform-test/subnets/default"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Exocompute configuration ID: %v\n", configID)

	// Retrieve the exocompute configuration added.
	exoConfig, err := exoClient.AzureConfigurationByID(ctx, configID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Exocompute Configuration: %v\n", exoConfig)

	// Remove the exocompute configuration.
	err = exoClient.RemoveAzureConfiguration(ctx, configID)
	if err != nil {
		log.Fatal(err)
	}

	// Disable the exocompute feature for the account.
	err = azureClient.RemoveSubscription(ctx, azure.CloudAccountID(accountID), core.FeatureExocompute, false)
	if err != nil {
		log.Fatal(err)
	}

	// Remove subscription.
	err = azureClient.RemoveSubscription(ctx, azure.CloudAccountID(accountID), core.FeatureCloudNativeProtection, false)
	if err != nil {
		log.Fatal(err)
	}
}
