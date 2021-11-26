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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	polaris_log "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

func main() {
	ctx := context.Background()

	// Load configuration and create client.
	polAccount, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		log.Fatal(err)
	}
	client, err := polaris.NewClient(ctx, polAccount, polaris_log.NewStandardLogger())
	if err != nil {
		log.Fatal(err)
	}

	// Add default Azure service principal to Polaris. Usually resolved using
	// the environment variable AZURE_SERVICEPRINCIPAL_LOCATION.
	_, err = client.Azure().SetServicePrincipal(ctx, azure.Default("my-domain.onmicrosoft.com"))
	if err != nil {
		log.Fatal(err)
	}

	// Add Azure subscription to Polaris.
	subscription := azure.Subscription(uuid.MustParse("9318aeec-d357-11eb-9b37-5f4e9f79db5d"),
		"my-domain.onmicrosoft.com")
	accountID, err := client.Azure().AddSubscription(ctx, subscription, core.FeatureCloudNativeProtection, azure.Regions("eastus2", "westus2"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Account ID: %v\n", accountID)

	// Enable the exocompute feature for the account.
	exoAccountID, err := client.Azure().AddSubscription(ctx, subscription, core.FeatureExocompute, azure.Regions("eastus2"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Exocompute Account ID: %v\n", exoAccountID)

	account, err := client.Azure().Subscription(ctx, azure.CloudAccountID(accountID), core.FeatureAll)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %v, NativeID: %v\n", account.Name, account.NativeID)
	for _, feature := range account.Features {
		fmt.Printf("Feature: %v, Regions: %v, Status: %v\n", feature.Name, feature.Regions, feature.Status)
	}

	// Add exocompute config for the account.
	exoID, err := client.Azure().AddExocomputeConfig(ctx, azure.CloudAccountID(accountID),
		azure.Managed("eastus2", "/subscriptions/9318aeec-d357-11eb-9b37-5f4e9f79db5d/resourceGroups/terraform-test/providers/Microsoft.Network/virtualNetworks/terraform-test/subnets/default"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Exocompute config ID: %v\n", exoID)

	// Retrieve the exocompute config added.
	exoConfig, err := client.Azure().ExocomputeConfig(ctx, exoID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Exocompute Config: %v\n", exoConfig)

	// Remove the exocompute config.
	err = client.Azure().RemoveExocomputeConfig(ctx, exoID)
	if err != nil {
		log.Fatal(err)
	}

	// Disable the exocompute feature for the account.
	err = client.Azure().RemoveSubscription(ctx, azure.CloudAccountID(accountID), core.FeatureExocompute, false)
	if err != nil {
		log.Fatal(err)
	}

	// Remove subscription.
	err = client.Azure().RemoveSubscription(ctx, azure.CloudAccountID(accountID), core.FeatureCloudNativeProtection, false)
	if err != nil {
		log.Fatal(err)
	}
}
