// Copyright 2025 Rubrik, Inc.
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
	"log"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/cloudcluster"
	gqlaws "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	gqlcloudcluster "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/cloudcluster"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Example showing how to create AWS Cloud Clusters via the RSC
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
	client, err := polaris.NewClientWithLogger(polAccount, polarislog.NewStandardLogger())
	if err != nil {
		log.Fatal(err)
	}

	awsClient := aws.Wrap(client)
	cloudClusterClient := cloudcluster.Wrap(client)

	// RSC features and their permission groups.
	features := []core.Feature{
		core.FeatureServerAndApps.WithPermissionGroups(core.PermissionGroupCCES),
	}

	// Add the AWS default account to Polaris. Usually resolved using the
	// environment variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and
	// AWS_DEFAULT_REGION.
	id, err := awsClient.AddAccountWithCFT(ctx, aws.Default(), features, aws.Regions("us-east-2"))
	if err != nil {
		log.Fatal(err)
	}

	// List the AWS accounts added to Polaris.
	account, err := awsClient.Account(ctx, aws.CloudAccountID(id), core.FeatureAll)
	if err != nil {
		log.Fatal(err)
	}

	// Create the Cloud Cluster
	clusterID, err := cloudClusterClient.CreateCloudCluster(ctx, gqlcloudcluster.CreateAwsClusterInput{
		CloudAccountID:       account.ID,
		Region:               gqlaws.RegionUsWest2.Name(),
		IsEsType:             true,
		UsePlacementGroups:   true,
		KeepClusterOnFailure: false,
		ClusterConfig: gqlcloudcluster.AwsClusterConfig{
			ClusterName:      "cces-cluster",
			UserEmail:        "hello@domain.com",
			AdminPassword:    "RubrikGoForward!",
			DnsNameServers:   []string{"169.254.169.253"},
			DnsSearchDomains: []string{},
			NtpServers:       []string{"169.254.169.123"},
			NumNodes:         3,
			AwsEsConfig: gqlcloudcluster.AwsEsConfigInput{
				BucketName:         "rbrk-cces.do-not-delete",
				ShouldCreateBucket: true,
				EnableImmutability: false,
				EnableObjectLock:   false,
			},
		},
		Validations: []gqlcloudcluster.ClusterCreateValidations{
			gqlcloudcluster.AllChecks,
		},
		VmConfig: gqlcloudcluster.AwsVmConfig{
			InstanceProfileName: "rubrik-cces-profile",
			InstanceType:        gqlcloudcluster.AwsInstanceTypeM6I_2XLarge,
			SecurityGroups:      []string{"sg-1234567890"},
			Subnet:              "subnet-1234567890",
			VmType:              gqlcloudcluster.CCVmConfigDense,
			Vpc:                 "vpc-1234567890",
		},
	}, true)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Cloud Cluster ID: %v\n", clusterID)
}
