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
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

func main() {
	ctx := context.Background()

	// Load configuration and create client.
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

	awsClient := aws.Wrap(client)

	// Add the AWS account to RSC.
	id, err := awsClient.AddAccount(ctx, aws.Default(), []core.Feature{core.FeatureCloudNativeProtection, core.FeatureCloudNativeArchival}, aws.Regions("us-east-2"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("RSC cloud account ID: %v\n", id)

	// Create an AWS archival location.
	targetMappingID, err := awsClient.CreateStorageSetting(ctx, aws.CloudAccountID(id), "Test", "my-prefix", "STANDARD", "", "aws/s3", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Target mapping ID: %v\n", targetMappingID)

	// Get the AWS archival location by ID.
	targetMapping, err := awsClient.TargetMappingByID(ctx, targetMappingID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID: %v, Name: %s\n", targetMapping.ID, targetMapping.Name)

	// Update the AWS archival location.
	err = awsClient.UpdateStorageSetting(ctx, targetMappingID, "TestUpdated", "", "")
	if err != nil {
		log.Fatal(err)
	}

	// Search for an AWS archival location by a name prefix.
	targetMappings, err := awsClient.TargetMappings(ctx, "Test")
	if err != nil {
		log.Fatal(err)
	}
	for _, targetMapping := range targetMappings {
		fmt.Printf("ID: %v, Name: %s\n", targetMapping.ID, targetMapping.Name)
	}

	// Delete the AWS archival location.
	err = awsClient.DeleteTargetMapping(ctx, targetMappingID)
	if err != nil {
		log.Fatal(err)
	}

	// Remove the AWS account from RSC.
	err = awsClient.RemoveAccount(ctx, aws.Default(), []core.Feature{core.FeatureCloudNativeProtection, core.FeatureCloudNativeArchival}, false)
	if err != nil {
		log.Fatal(err)
	}
}
