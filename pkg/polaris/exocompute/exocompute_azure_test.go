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

package exocompute

import (
	"context"
	"errors"
	"maps"
	"slices"
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlazure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// TestAzureExocompute verifies that the SDK can perform basic Exocompute
// operations on a real RSC instance.
//
// To run this test against an RSC instance, the following environment variables
// need to be set:
//   - RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   - TEST_INTEGRATION=1
//   - TEST_AZURESUBSCRIPTION_FILE=<path-to-test-azure-subscription-file>
//   - AZURE_AUTH_LOCATION=<path-to-azure-sdk-auth-file>
//
// The file referred to by TEST_AZURESUBSCRIPTION_FILE should contain a single
// testAzureSubscription JSON object.
func TestAzureExocompute(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	testSubscription, err := testsetup.AzureSubscription()
	if err != nil {
		t.Fatal(err)
	}

	azureClient := azure.Wrap(client)
	exoClient := Wrap(client)

	// Add default Azure service principal to RSC. Usually resolved using the
	// environment variable AZURE_SERVICEPRINCIPAL_LOCATION.
	_, err = azureClient.SetServicePrincipal(ctx, azure.Default(testSubscription.TenantDomain))
	if err != nil {
		t.Fatal(err)
	}

	// Add Azure subscription to RSC.
	subscription := azure.Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	cnpRegions := azure.Regions(testSubscription.CloudNativeProtection.Regions...)
	cnpResourceGroup := azure.ResourceGroup(testSubscription.CloudNativeProtection.ResourceGroupName,
		testSubscription.CloudNativeProtection.ResourceGroupRegion, nil)
	accountID, err := azureClient.AddSubscription(ctx, subscription, core.FeatureCloudNativeProtection,
		azure.Name(testSubscription.SubscriptionName), cnpRegions, cnpResourceGroup)
	if err != nil {
		t.Fatal(err)
	}

	// Enable the exocompute feature for the account.
	exoRegions := azure.Regions(testSubscription.Exocompute.Regions...)
	exoResourceGroup := azure.ResourceGroup(testSubscription.Exocompute.ResourceGroupName,
		testSubscription.Exocompute.ResourceGroupRegion, nil)
	exoAccountID, err := azureClient.AddSubscription(ctx, subscription, core.FeatureExocompute, exoRegions,
		exoResourceGroup)
	if err != nil {
		t.Fatal(err)
	}
	if accountID != exoAccountID {
		t.Fatal("cloud native protection and exocompute added to different cloud accounts")
	}

	account, err := azureClient.Subscription(ctx, azure.CloudAccountID(accountID), core.FeatureExocompute)
	if err != nil {
		t.Fatal(err)
	}
	if name := account.Name; name != testSubscription.SubscriptionName {
		t.Fatalf("invalid name: %v", name)
	}
	if id := account.NativeID; id != testSubscription.SubscriptionID {
		t.Fatalf("invalid native id: %v", id)
	}
	if domain := account.TenantDomain; domain != testSubscription.TenantDomain {
		t.Fatalf("invalid tenant domain: %v", domain)
	}
	if n := len(account.Features); n != 1 {
		t.Fatalf("invalid number of features: %v", n)
	}
	feature, ok := account.Feature(core.FeatureExocompute)
	if !ok {
		t.Fatalf("%s feature not found", core.FeatureExocompute)
	}
	if name := feature.Name; name != core.FeatureExocompute.Name {
		t.Fatalf("invalid feature name: %v", name)
	}
	slices.Sort(feature.Regions)
	slices.Sort(testSubscription.Exocompute.Regions)
	if regions := feature.Regions; !slices.Equal(regions, testSubscription.Exocompute.Regions) {
		t.Fatalf("invalid feature regions: %v", regions)
	}
	if status := feature.Status; status != "CONNECTED" {
		t.Fatalf("invalid feature status: %v", status)
	}
	if name := feature.ResourceGroup.Name; name != testSubscription.Exocompute.ResourceGroupName {
		t.Fatalf("invalid feature resource group name: %v", name)
	}
	if region := feature.ResourceGroup.Region; region != testSubscription.Exocompute.ResourceGroupRegion {
		t.Fatalf("invalid feature resource group region: %v", region)
	}
	if tags := feature.ResourceGroup.Tags; !maps.Equal(tags, map[string]string{}) {
		t.Fatalf("invalid feature resource group tags: %v", tags)
	}

	// Add an exocompute configuration to the account.
	if len(testSubscription.Exocompute.Regions) == 0 {
		t.Fatalf("exocompute test data must contain at least one region")
	}
	exoConfigRegion := gqlazure.RegionFromName(testSubscription.Exocompute.Regions[0])
	exoID, err := exoClient.AddAzureConfiguration(ctx, accountID, AzureManaged(exoConfigRegion, testSubscription.Exocompute.SubnetID))
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve the exocompute configuration added.
	exoConfig, err := exoClient.AzureConfigurationByID(ctx, exoID)
	if err != nil {
		t.Fatal(err)
	}
	if id := exoConfig.ID; id != exoID {
		t.Fatalf("invalid exocompute config id: %v", id)
	}
	if region := exoConfig.Region; region.Region != exoConfigRegion {
		t.Fatalf("invalid exocompute config region: %v", region)
	}
	if id := exoConfig.SubnetID; id != testSubscription.Exocompute.SubnetID {
		t.Fatalf("invalid exocompute config subnet id: %v", id)
	}
	if !exoConfig.ManagedByRubrik {
		t.Fatalf("invalid exocompute config managed state: %t", exoConfig.ManagedByRubrik)
	}

	// Remove the exocompute configuration.
	err = exoClient.RemoveAzureConfiguration(ctx, exoID)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the exocompute configuration was successfully removed.
	exoConfig, err = exoClient.AzureConfigurationByID(ctx, exoID)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}

	// Remove exocompute feature.
	err = azureClient.RemoveSubscription(ctx, azure.CloudAccountID(accountID), core.FeatureExocompute, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the feature was successfully removed.
	_, err = azureClient.Subscription(ctx, azure.CloudAccountID(accountID), core.FeatureExocompute)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}

	// Remove subscription.
	err = azureClient.RemoveSubscription(ctx, azure.CloudAccountID(accountID), core.FeatureCloudNativeProtection, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	_, err = azureClient.Subscription(ctx, azure.ID(subscription), core.FeatureCloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}
