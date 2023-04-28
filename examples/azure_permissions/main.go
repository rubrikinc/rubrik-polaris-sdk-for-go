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

// Example showing how to manage permissions for an Azure subscription with the
// Polaris Go SDK.
//
// The Polaris service account key file identifying the Polaris account should
// either be placed at ~/.rubrik/polaris-service-account.json or pointed out by
// the RUBRIK_POLARIS_SERVICEACCOUNT_FILE environment variable.
func main() {
	ctx := context.Background()

	// Load configuration and create client.
	polAccount, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		log.Fatal(err)
	}
	client, err := polaris.NewClientWithLogger(polAccount, polaris_log.NewStandardLogger())
	if err != nil {
		log.Fatal(err)
	}

	azureClient := azure.Wrap(client)

	// List Azure permissions needed for features.
	features := []core.Feature{core.FeatureCloudNativeProtection}
	perms, err := azureClient.Permissions(ctx, features)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Permissions requried for Cloud Native Protection:")
	for _, perm := range perms.Actions {
		fmt.Println(perm)
	}
	for _, perm := range perms.NotActions {
		fmt.Println(perm)
	}
	for _, perm := range perms.DataActions {
		fmt.Println(perm)
	}
	for _, perm := range perms.NotDataActions {
		fmt.Println(perm)
	}

	// Notify Polaris about updated permissions for the Cloud Native Protection
	// feature of the already added subscription.
	account, err := azureClient.Subscription(ctx,
		azure.SubscriptionID(uuid.MustParse("27dce22c-1b84-11ec-9992-a3d4a0eb7b90")), core.FeatureCloudNativeProtection)
	if err != nil {
		log.Fatal(err)
	}
	err = azureClient.PermissionsUpdated(ctx, azure.CloudAccountID(account.ID), features)
	if err != nil {
		log.Fatal(err)
	}
}
