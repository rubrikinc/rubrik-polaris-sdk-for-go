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

package azure

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// TestAzureExocompute verifies that the SDK can perform basic Exocompute
// operations on a real RSC instance.
//
// To run this test against an RSC instance the following environment variables
// needs to be set:
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

	azureClient := Wrap(client)

	// Add default Azure service principal to RSC. Usually resolved using the
	// environment variable AZURE_SERVICEPRINCIPAL_LOCATION.
	_, err = azureClient.SetServicePrincipal(ctx, Default(testSubscription.TenantDomain))
	if err != nil {
		t.Fatal(err)
	}

	// Add default Azure subscription to RSC.
	subscription := Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	accountID, err := azureClient.AddSubscription(ctx, subscription, core.FeatureCloudNativeProtection,
		Regions("eastus2"), Name(testSubscription.SubscriptionName))
	if err != nil {
		t.Fatal(err)
	}

	// Enable the exocompute feature for the account.
	exoAccountID, err := azureClient.AddSubscription(ctx, subscription, core.FeatureExocompute, Regions("eastus2"))
	if err != nil {
		t.Fatal(err)
	}
	if accountID != exoAccountID {
		t.Fatal("cloud native protection and exocompute added to different cloud accounts")
	}

	account, err := azureClient.Subscription(ctx, CloudAccountID(accountID), core.FeatureExocompute)
	if err != nil {
		t.Error(err)
	}
	if account.Name != testSubscription.SubscriptionName {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.NativeID != testSubscription.SubscriptionID {
		t.Errorf("invalid native id: %v", account.NativeID)
	}
	if account.TenantDomain != testSubscription.TenantDomain {
		t.Errorf("invalid tenant domain: %v", account.TenantDomain)
	}
	if n := len(account.Features); n == 1 {
		if feature := account.Features[0].Feature; !feature.Equal(core.FeatureExocompute) {
			t.Errorf("invalid feature name: %v", feature)
		}
		if regions := account.Features[0].Regions; !reflect.DeepEqual(regions, []string{"eastus2"}) {
			t.Errorf("invalid feature regions: %v", regions)
		}
		if account.Features[0].Status != core.StatusConnected {
			t.Errorf("invalid feature status: %v", account.Features[0].Status)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	// Add exocompute config for the account.
	exoID, err := azureClient.AddExocomputeConfig(ctx, CloudAccountID(accountID),
		Managed("eastus2", testSubscription.Exocompute.SubnetID))
	if err != nil {
		t.Error(err)
	}

	// Retrieve the exocompute config added.
	exoConfig, err := azureClient.ExocomputeConfig(ctx, exoID)
	if err != nil {
		t.Error(err)
	}
	if exoConfig.ID != exoID {
		t.Errorf("invalid id: %v", exoConfig.ID)
	}
	if exoConfig.Region != "eastus2" {
		t.Errorf("invalid region: %v", exoConfig.Region)
	}
	if exoConfig.SubnetID != testSubscription.Exocompute.SubnetID {
		t.Errorf("invalid subnet id: %v", exoConfig.SubnetID)
	}
	if !exoConfig.ManagedByRubrik {
		t.Errorf("invalid polaris managed state: %t", exoConfig.ManagedByRubrik)
	}

	// Remove the exocompute config.
	err = azureClient.RemoveExocomputeConfig(ctx, exoID)
	if err != nil {
		t.Error(err)
	}

	// Verify that the exocompute config was successfully removed.
	exoConfig, err = azureClient.ExocomputeConfig(ctx, exoID)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Error(err)
	}

	// Remove exocompute feature.
	err = azureClient.RemoveSubscription(ctx, CloudAccountID(accountID), core.FeatureExocompute, false)
	if err != nil {
		t.Error(err)
	}

	// Verify that the feature was successfully removed.
	_, err = azureClient.Subscription(ctx, CloudAccountID(accountID), core.FeatureExocompute)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Error(err)
	}

	// Remove subscription.
	err = azureClient.RemoveSubscription(ctx, CloudAccountID(accountID), core.FeatureCloudNativeProtection, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	_, err = azureClient.Subscription(ctx, ID(subscription), core.FeatureCloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}
