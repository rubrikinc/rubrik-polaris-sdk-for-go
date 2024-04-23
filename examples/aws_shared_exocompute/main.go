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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	polaris_log "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Note: This example requires an existing AWS account with exocompute
// configured in RSC.
func main() {
	ctx := context.Background()

	// Load configuration and create a client.
	polAccount, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		log.Fatal(err)
	}
	logger := polaris_log.NewStandardLogger()
	polaris.SetLogLevelFromEnv(logger)
	client, err := polaris.NewClientWithLogger(polAccount, logger)
	if err != nil {
		log.Fatal(err)
	}

	awsClient := aws.Wrap(client)

	// The AWS account ID of the existing AWS account with exocompute
	// configured.
	hostAccountID := aws.AccountID("123456789012")

	// Add the AWS default account to Polaris. Usually resolved using the
	// environment variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and
	// AWS_DEFAULT_REGION.
	accountID, err := awsClient.AddAccount(ctx, aws.Default(),
		[]core.Feature{core.FeatureCloudNativeProtection, core.FeatureExocompute}, aws.Regions("us-east-2"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Account ID: %v\n", accountID)

	// Map the application account to an existing exocompute host account.
	err = awsClient.MapExocompute(ctx, hostAccountID, aws.CloudAccountID(accountID))
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve the exocompute host account for the application account.
	hostID, err := awsClient.ExocomputeHostAccount(ctx, aws.CloudAccountID(accountID))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Exocompute Host Account: %v\n", hostID)

	// Retrieve the exocompute application accounts for the exocompute host
	// account.
	appIDs, err := awsClient.ExocomputeApplicationAccounts(ctx, hostAccountID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Exocompute Application Accounts:")
	for _, appID := range appIDs {
		fmt.Println(appID)
	}

	// Unmap the application account from the shared exocompute host account.
	err = awsClient.UnmapExocompute(ctx, aws.CloudAccountID(accountID))
	if err != nil {
		log.Fatal(err)
	}

	// Remove the AWS account from Polaris.
	err = awsClient.RemoveAccount(ctx, aws.Default(),
		[]core.Feature{core.FeatureCloudNativeProtection, core.FeatureExocompute}, false)
	if err != nil {
		log.Fatal(err)
	}
}
