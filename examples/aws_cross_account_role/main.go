// Copyright 2023 Rubrik, Inc.
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

// Example showing how to manage an AWS account with the Polaris Go SDK using a
// cross account role.
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

	awsClient := aws.Wrap(client)

	// Use the default profile to add an AWS account to Polaris using a cross
	// account role. The default profile can be configured using the environment
	// variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_REGION.
	id, err := awsClient.AddAccount(ctx,
		aws.DefaultWithRole("arn:aws:iam::123456789012:role/MyCrossAccountRole"),
		[]core.Feature{core.FeatureCloudNativeProtection}, aws.Regions("us-east-2"))
	if err != nil {
		log.Fatal(err)
	}

	// List AWS accounts added to Polaris.
	account, err := awsClient.Account(ctx, aws.CloudAccountID(id), core.FeatureCloudNativeProtection)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %v, NativeID: %v\n", account.Name, account.NativeID)
	for _, feature := range account.Features {
		fmt.Printf("Feature: %v, Regions: %v, Status: %v\n", feature.Feature, feature.Regions, feature.Status)
	}

	// Remove the AWS account from Polaris using a cross account role.
	err = awsClient.RemoveAccount(ctx,
		aws.DefaultWithRole("arn:aws:iam::123456789012:role/MyCrossAccountRole"),
		[]core.Feature{core.FeatureCloudNativeProtection}, false)
	if err != nil {
		log.Fatal(err)
	}
}
