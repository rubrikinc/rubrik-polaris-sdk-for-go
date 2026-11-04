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

package cloudcluster

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlcloudcluster "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/cloudcluster"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	gqlazure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// client is the RSC client shared across all integration tests in this package.
// Reusing a single client avoids repeated token creation, reducing the risk of
// hitting rate limits.
var client *polaris.Client

func TestMain(m *testing.M) {
	if testsetup.BoolEnvSet("TEST_INTEGRATION") {
		// Load configuration and create client. Usually resolved using the
		// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
		polAccount, err := polaris.DefaultServiceAccount(true)
		if err != nil {
			fmt.Printf("failed to get default service account: %v\n", err)
			os.Exit(1)
		}

		// The integration tests defaults the log level to INFO. Note that
		// RUBRIK_POLARIS_LOGLEVEL can be used to override this.
		logger := log.NewStandardLogger()
		logger.SetLogLevel(log.Info)
		if err := polaris.SetLogLevelFromEnv(logger); err != nil {
			fmt.Printf("failed to get log level from env: %v\n", err)
			os.Exit(1)
		}

		client, err = polaris.NewClientWithLogger(polAccount, logger)
		if err != nil {
			fmt.Printf("failed to create polaris client: %v\n", err)
			os.Exit(1)
		}
	}

	os.Exit(m.Run())
}

// TestAzureCdmVersionsRegionCanary acts as a canary test to catch region
// handling issues for Azure Cloud Cluster early. It onboards the test
// subscription with the Servers and Apps feature in all public cloud regions,
// since RSC only reports the regions the feature is onboarded in. For every
// region reported, it verifies that the region is known to the SDK and that the
// CDM versions can be looked up for it.
//
// The test is skipped if the RSC account doesn't have the
// CLOUD_CLUSTER_ES_ENABLED feature flag enabled, since all Azure Cloud Cluster
// APIs are gated by it.
//
// To run this test against an RSC instance, the following environment variables
// need to be set:
//   - RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   - TEST_INTEGRATION=1
//   - TEST_AZURESUBSCRIPTION_FILE=<path-to-test-azure-subscription-file>
//   - AZURE_SERVICEPRINCIPAL_LOCATION=<path-to-azure-service-principal-file>
func TestAzureCdmVersionsRegionCanary(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	ctx := context.Background()

	flag, err := core.Wrap(client.GQL).FeatureFlag(ctx, core.FeatureFlagCloudClusterES)
	if err != nil {
		t.Fatal(err)
	}
	if !flag.Enabled {
		t.Skipf("skipping due to feature flag %s not enabled", flag.Name)
	}

	testSubscription, err := testsetup.AzureSubscription()
	if err != nil {
		t.Fatal(err)
	}

	azureClient := azure.Wrap(client)

	// Add default Azure service principal to RSC. Usually resolved using the
	// environment variable AZURE_SERVICEPRINCIPAL_LOCATION.
	if _, err := azureClient.SetServicePrincipal(ctx, azure.Default(testSubscription.TenantDomain)); err != nil {
		t.Fatal(err)
	}

	// Add Azure subscription to RSC with the Servers and Apps feature in all
	// public cloud regions.
	regions := publicCloudRegionNames()
	subscription := azure.Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	accountID, err := azureClient.AddSubscription(ctx, subscription, core.FeatureServerAndApps,
		azure.Name(testSubscription.SubscriptionName), azure.Regions(regions...))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := azureClient.RemoveSubscription(ctx, accountID, core.FeatureServerAndApps, false); err != nil {
			t.Error(err)
		}

		// Verify that the subscription was successfully removed.
		if _, err := azureClient.SubscriptionByID(ctx, accountID); !errors.Is(err, graphql.ErrNotFound) {
			t.Error(err)
		}
	})

	gqlClient := gqlcloudcluster.Wrap(client.GQL)
	regionDetails, err := gqlClient.AzureCCRegionDetails(ctx, accountID)
	if err != nil {
		t.Fatal(err)
	}
	if len(regionDetails) != len(regions) {
		t.Fatalf("got %d Azure Cloud Cluster regions, want %d", len(regionDetails), len(regions))
	}

	for _, details := range regionDetails {
		region := details.Location.Region
		if region == gqlazure.RegionUnknown {
			t.Errorf("RSC returned an Azure Cloud Cluster region unknown to the SDK: %q", details.Location)
			continue
		}
		if _, ok := restrictedRegions[region]; ok {
			t.Logf("skipping CDM versions lookup for access-restricted region %s", region)
			continue
		}
		if _, err := gqlClient.AllAzureCdmVersions(ctx, accountID, region); err != nil {
			t.Errorf("failed to get CDM versions for region %s: %v", region, err)
		}
	}
}

// restrictedRegions holds the Azure regions which require access to be
// requested before resources can be deployed in them. Azure fails to list the
// CDM images in these regions for the test subscription, with the error
// NoRegisteredProviderFound, which RSC returns as an internal error.
var restrictedRegions = map[gqlazure.Region]struct{}{
	gqlazure.RegionAustraliaCentral2: {},
	gqlazure.RegionBrazilSoutheast:   {},
	gqlazure.RegionFranceSouth:       {},
	gqlazure.RegionGermanyNorth:      {},
	gqlazure.RegionNorwayWest:        {},
	gqlazure.RegionSouthAfricaWest:   {},
	gqlazure.RegionSwedenSouth:       {},
	gqlazure.RegionSwitzerlandWest:   {},
	gqlazure.RegionUAECentral:        {},
}

// publicCloudRegionNames returns the names of all regions a subscription in the
// Azure public cloud can be onboarded in, i.e. all regions with an
// AzureCloudAccountRegion enum value except the China and US Government cloud
// regions.
func publicCloudRegionNames() []string {
	var names []string
	for _, name := range gqlazure.CloudAccountRegionEnumRegionNames() {
		if !strings.HasPrefix(name, "china") && !strings.HasPrefix(name, "usgov") {
			names = append(names, name)
		}
	}
	return names
}
