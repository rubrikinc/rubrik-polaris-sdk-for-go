// Copyright 2026 Rubrik, Inc.
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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/exocompute"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	gqlgcp "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Example showing how to manage a GCP Exocompute configuration with the SDK.
//
// The RSC service account key file, identifying the RSC account, should be
// pointed out by the RUBRIK_POLARIS_SERVICEACCOUNT_FILE environment variable.
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

	gcpClient := gcp.Wrap(client)

	// Add the GCP default project to RSC. Usually resolved using the
	// environment variable GOOGLE_APPLICATION_CREDENTIALS.
	cloudAccountID, err := gcpClient.AddProject(ctx, gcp.Default(), []core.Feature{core.FeatureExocompute})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Cloud account ID: %s\n", cloudAccountID)

	exoClient := exocompute.Wrap(client)

	// Add an Exocompute configuration to the cloud account.
	err = exoClient.UpdateGCPConfiguration(ctx, cloudAccountID, []exocompute.RegionalConfig{{
		Region:         gqlgcp.RegionUSWest1,
		SubnetName:     "subnet-1",
		VPCNetworkName: "vpc-1",
	}}, false)
	if err != nil {
		log.Fatal(err)
	}

	// List the Exocompute configurations in the cloud account.
	configs, err := exoClient.GCPConfigurationsByCloudAccountID(ctx, cloudAccountID)
	if err != nil {
		log.Fatal(err)
	}
	for _, config := range configs {
		fmt.Printf("Exocompute Configuration: %v\n", config)
	}

	// Remove the Exocompute configurations from the cloud account.
	err = exoClient.RemoveGCPConfiguration(ctx, cloudAccountID)
	if err != nil {
		log.Fatal(err)
	}

	// Remove the GCP project from RSC.
	if err := gcpClient.RemoveProject(ctx, cloudAccountID, []core.Feature{core.FeatureExocompute}, false); err != nil {
		log.Fatal(err)
	}
}
