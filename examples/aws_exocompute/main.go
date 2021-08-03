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

func main() {
	ctx := context.Background()

	// Load configuration and create client.
	polAccount, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		log.Fatal(err)
	}
	client, err := polaris.NewClient(polAccount, &polaris_log.StandardLogger{})
	if err != nil {
		log.Fatal(err)
	}

	// Add the AWS default account to Polaris. Usually resolved using the
	// environment variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and
	// AWS_DEFAULT_REGION.
	accountID, err := client.AWS().AddAccount(ctx, aws.Default(), core.FeatureCloudNativeProtection, aws.Regions("us-east-2", "us-west-2"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Account ID: %v\n", accountID)

	// Enable the exocompute feature for the account. Note that the
	// cnpAccountID and exoAccountID should be the same, they refer to the same
	// Polaris cloud account.
	exoAccountID, err := client.AWS().AddAccount(ctx, aws.Default(), core.FeatureExocompute, aws.Regions("us-east-2"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Exocompute Account ID: %v\n", exoAccountID)

	account, err := client.AWS().Account(ctx, aws.CloudAccountID(accountID), core.FeatureAll)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %v, NativeID: %v\n", account.Name, account.NativeID)
	for _, feature := range account.Features {
		fmt.Printf("Feature: %v, Regions: %v, Status: %v\n", feature.Name, feature.Regions, feature.Status)
	}

	// Add exocompute config for the account.
	exoID, err := client.AWS().AddExocomputeConfig(ctx, aws.ID(aws.Default()),
		aws.Managed("us-east-2", "vpc-4859acb9", []string{"subnet-ea67b67b", "subnet-ea43ec78"}))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Exocompute config ID: %v\n", exoID)

	// Retrieve the exocompute config added.
	exoConfig, err := client.AWS().ExocomputeConfig(ctx, exoID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Exocompute Config: %v\n", exoConfig)

	// Remove the exocompute config.
	err = client.AWS().RemoveExocomputeConfig(ctx, exoID)
	if err != nil {
		log.Fatal(err)
	}

	// Disable the exocompute feature for the account.
	err = client.AWS().RemoveAccount(ctx, aws.Default(), core.FeatureExocompute, false)
	if err != nil {
		log.Fatal(err)
	}

	// Remove the AWS account from Polaris.
	err = client.AWS().RemoveAccount(ctx, aws.Default(), core.FeatureCloudNativeProtection, false)
	if err != nil {
		log.Fatal(err)
	}
}
