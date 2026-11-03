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

	"github.com/google/uuid"
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
// handling issues for Azure Cloud Cluster early. For every region RSC reports
// as available for Azure Cloud Cluster in the test subscription, it verifies
// that the region is known to the SDK and that the CDM versions can be looked
// up for it.
//
// The Servers and Apps feature is required to access the Azure Cloud Cluster
// APIs, and RSC only reports the regions the feature is onboarded in. If the
// test subscription doesn't have the feature, it's onboarded in all public
// cloud regions for the duration of the test and removed afterwards. Other
// features of the subscription are left untouched. The test is skipped if the
// RSC account doesn't have the CLOUD_CLUSTER_ES_ENABLED feature flag enabled.
func TestAzureCdmVersionsRegionCanary(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	// All Azure Cloud Cluster APIs are gated by the feature flag.
	flag, err := core.Wrap(client.GQL).FeatureFlag(ctx, "CLOUD_CLUSTER_ES_ENABLED")
	if err != nil {
		t.Fatal(err)
	}
	if !flag.Enabled {
		t.Skipf("skipping due to feature flag %s not enabled for the RSC account", flag.Name)
	}

	testSubscription, err := testsetup.AzureSubscription()
	if err != nil {
		t.Fatal(err)
	}

	azureClient := azure.Wrap(client)
	accountID, err := onboardedFeature(ctx, azureClient, testSubscription.SubscriptionID, core.FeatureServerAndApps)
	if err != nil {
		t.Fatal(err)
	}
	onboardedByTest := accountID == uuid.Nil
	if onboardedByTest {
		// Add default Azure service principal to RSC. Usually resolved using
		// the environment variable AZURE_SERVICEPRINCIPAL_LOCATION.
		if _, err := azureClient.SetServicePrincipal(ctx, azure.Default(testSubscription.TenantDomain)); err != nil {
			t.Fatal(err)
		}

		subscription := azure.Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
		accountID, err = azureClient.AddSubscription(ctx, subscription, core.FeatureServerAndApps,
			azure.Name(testSubscription.SubscriptionName), azure.Regions(publicCloudRegionNames()...))
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := azureClient.RemoveSubscription(ctx, accountID, core.FeatureServerAndApps, false); err != nil {
				t.Error(err)
			}
		}()
	}

	gqlClient := gqlcloudcluster.Wrap(client.GQL)
	regionDetails, err := gqlClient.AzureCCRegionDetails(ctx, accountID)
	if err != nil {
		t.Fatal(err)
	}
	if len(regionDetails) == 0 {
		t.Fatalf("no Azure Cloud Cluster regions returned for cloud account %s", accountID)
	}
	if want := len(publicCloudRegionNames()); onboardedByTest && len(regionDetails) != want {
		t.Errorf("got %d Azure Cloud Cluster regions, want %d", len(regionDetails), want)
	}

	// The subtests run in parallel, wrap them in a group so that the feature
	// is removed after all subtests have completed.
	t.Run("Regions", func(t *testing.T) {
		for _, details := range regionDetails {
			region := details.Location.Region
			t.Run(region.Name(), func(t *testing.T) {
				if region == gqlazure.RegionUnknown {
					t.Fatalf("RSC returned an Azure Cloud Cluster region unknown to the SDK: %q", details.Location)
				}
				if _, ok := restrictedRegions[region]; ok {
					t.Skipf("skipping CDM versions lookup for access-restricted region %s", region.Name())
				}
				t.Parallel()

				if _, err := gqlClient.AllAzureCdmVersions(ctx, accountID, region); err != nil {
					t.Errorf("failed to get CDM versions for region %s: %v", region.Name(), err)
				}
			})
		}
	})
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

// onboardedFeature returns the RSC cloud account ID of the subscription if it
// has the specified feature onboarded, otherwise uuid.Nil is returned.
func onboardedFeature(ctx context.Context, azureClient azure.API, nativeID uuid.UUID, feature core.Feature) (uuid.UUID, error) {
	account, err := azureClient.SubscriptionByNativeID(ctx, nativeID)
	if errors.Is(err, graphql.ErrNotFound) {
		return uuid.Nil, nil
	}
	if err != nil {
		return uuid.Nil, err
	}
	if _, ok := account.Feature(feature); !ok {
		return uuid.Nil, nil
	}
	return account.ID, nil
}
