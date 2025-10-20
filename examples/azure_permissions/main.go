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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Example showing how to manage permissions for an Azure subscription with the
// Polaris Go SDK.
//
// The Polaris service account key file identifying the Polaris account should
// either be placed at ~/.rubrik/polaris-service-account.json or pointed out by
// the RUBRIK_POLARIS_SERVICEACCOUNT_FILE environment variable.
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

	// List Azure permissions needed for the Cloud Native Protection feature.
	perms, permGroups, err := azureClient.ScopedPermissions(ctx, core.FeatureCloudNativeProtection)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Subscription level permissions required for Cloud Native Protection:")
	for _, perm := range perms[azure.ScopeSubscription].Actions {
		fmt.Println(perm)
	}
	for _, perm := range perms[azure.ScopeSubscription].NotActions {
		fmt.Println(perm)
	}
	for _, perm := range perms[azure.ScopeSubscription].DataActions {
		fmt.Println(perm)
	}
	for _, perm := range perms[azure.ScopeSubscription].NotDataActions {
		fmt.Println(perm)
	}

	fmt.Println("Resource group level permissions required for Cloud Native Protection:")
	for _, perm := range perms[azure.ScopeResourceGroup].Actions {
		fmt.Println(perm)
	}
	for _, perm := range perms[azure.ScopeResourceGroup].NotActions {
		fmt.Println(perm)
	}
	for _, perm := range perms[azure.ScopeResourceGroup].DataActions {
		fmt.Println(perm)
	}
	for _, perm := range perms[azure.ScopeResourceGroup].NotDataActions {
		fmt.Println(perm)
	}

	fmt.Println("Permission groups available for Cloud Native Protection:")
	for _, permGroup := range permGroups {
		fmt.Printf("Permission group %s: %d\n", permGroup.Name, permGroup.Version)
	}
}
