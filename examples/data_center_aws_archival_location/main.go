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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/archival"
	gqlarchival "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/archival"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
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

	id, err := archival.Wrap(client).CreateAWSCloudAccount(ctx, gqlarchival.CreateAWSCloudAccountParams{
		Name:      "my-aws-account",
		AccessKey: "AKIAIOSFODNN7EXAMPLE",
		SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Cloud Account ID: %v\n", id)

	cloudAccounts, err := archival.Wrap(client).AWSCloudAccounts(ctx, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Cloud Accounts:")
	for _, cloudAccount := range cloudAccounts {
		fmt.Printf("ID: %v, Name: %v\n", cloudAccount.ID, cloudAccount.Name)
	}

	targetID, err := archival.Wrap(client).CreateAWSTarget(ctx, gqlarchival.CreateAWSTargetParams{
		Name:           "my-aws-target",
		ClusterID:      uuid.MustParse("b83cf0de-0c26-4bd1-82a6-f2f8a099a53e"),
		CloudAccountID: id,
		BucketName:     "my-bucket",
		KMSMasterKeyID: "aws/s3",
		Region:         aws.RegionFromName("us-east-2").ToRegionEnum(),
		StorageClass:   "STANDARD",
		RetrievalTier:  "STANDARD_TIER",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Target ID: %v\n", targetID)

	targets, err := archival.Wrap(client).AWSTargets(ctx, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Targets:")
	for _, target := range targets {
		fmt.Printf("ID: %v, Name: %v\n", target.ID, target.Name)
	}

	if err := archival.Wrap(client).DeleteTarget(ctx, targetID); err != nil {
		log.Fatal(err)
	}

	if err := archival.Wrap(client).DeleteAWSCloudAccount(ctx, id); err != nil {
		log.Fatal(err)
	}
}
