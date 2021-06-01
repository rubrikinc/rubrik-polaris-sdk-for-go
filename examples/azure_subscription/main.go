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

	"github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris"
	polaris_log "github.com/trinity-team/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Example showing how to manage an AWS account with the Polaris Go SDK.
//
// The Polaris service account key file identifying the Polaris account should
// either be placed at ~/.rubrik/polaris-service-account.json or pointed out by
// the RUBRIK_POLARIS_SERVICEACCOUNT_FILE environment variable.
func main() {
	ctx := context.Background()

	// Load configuration and create client.
	account, err := polaris.DefaultServiceAccount()
	if err != nil {
		log.Fatal(err)
	}
	client, err := polaris.NewClientFromServiceAccount(account, &polaris_log.StandardLogger{})
	if err != nil {
		log.Fatal(err)
	}

	// Add default Azure service principal to Polaris. Usually resolved using
	// the environment variable AZURE_SERVICEPRINCIPAL_LOCATION.
	principal, err := polaris.AzureDefaultServicePrincipal()
	if err != nil {
		log.Fatal(err)
	}

	err = client.AzureServicePrincipalSet(ctx, principal)
	if err != nil {
		log.Fatal(err)
	}

	// Add default Azure subscription to Polaris. Usually resolved using the
	// environment variable AZURE_SUBSCRIPTION_LOCATION
	subscriptionIn, err := polaris.AzureDefaultSubscription()
	if err != nil {
		log.Fatal(err)
	}

	err = client.AzureSubscriptionAdd(ctx, subscriptionIn)
	if err != nil {
		log.Fatal(err)
	}

	// Lookup the newly added subscription.
	subscriptions, err := client.AzureSubscriptions(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, subscription := range subscriptions {
		fmt.Printf("Name: %v, NativeID: %v\n", subscription.Name, subscription.NativeID)
		fmt.Printf("Feature: %v, Regions: %v, Status: %v\n", subscription.Feature.Name,
			subscription.Feature.Regions, subscription.Feature.Status)
	}

	// Remove subscription.
	err = client.AzureSubscriptionRemove(ctx,
		polaris.WithAzureSubscriptionID(subscriptionIn.ID.String()), false)
	if err != nil {
		log.Fatal(err)
	}
}
