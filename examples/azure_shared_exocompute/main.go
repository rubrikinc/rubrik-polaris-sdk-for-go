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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/exocompute"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Note: This example requires an existing Azure account with exocompute
// configured in RSC.
func main() {
	ctx := context.Background()

	// Load configuration and create a client.
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
	exoClient := exocompute.Wrap(client)

	// The Azure cloud account ID of the existing Azure cloud account with
	// exocompute configured.
	hostAccountID := uuid.MustParse("3cad3091-a1b3-4e0e-823d-84589568983e")

	// Add Azure subscription to RSC.
	subscription := azure.Subscription(uuid.MustParse("e4b247e7-66c5-4f10-9042-1eeac424c7a4"),
		"my-domain.onmicrosoft.com")
	accountID, err := azureClient.AddSubscription(ctx, subscription, core.FeatureCloudNativeProtection, azure.Regions("us-east-2"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Account ID: %v\n", accountID)

	// Map the application cloud account to the existing exocompute host cloud
	// account.
	err = exoClient.MapAzureCloudAccount(ctx, hostAccountID, accountID)
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve the exocompute host cloud account for the application cloud
	// account.
	hostID, err := exoClient.AzureHostCloudAccount(ctx, accountID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Exocompute Host Account: %v\n", hostID)

	// Unmap the application cloud account from the shared exocompute host cloud
	// account.
	err = exoClient.UnmapAzureCloudAccount(ctx, accountID)
	if err != nil {
		log.Fatal(err)
	}

	// Remove the Azure account from RSC.
	err = azureClient.RemoveSubscription(ctx, accountID, core.FeatureCloudNativeProtection, false)
	if err != nil {
		log.Fatal(err)
	}
}
