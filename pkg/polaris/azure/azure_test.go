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
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	polaris_log "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// client is the common RSC client used for tests. By reusing the same client
// we reduce the risk of hitting rate limits when access tokens are created.
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
		logger := polaris_log.NewStandardLogger()
		logger.SetLogLevel(polaris_log.Info)
		client, err = polaris.NewClient(context.Background(), polAccount, logger)
		if err != nil {
			fmt.Printf("failed to create polaris client: %v\n", err)
			os.Exit(1)
		}
	}

	os.Exit(m.Run())
}

// TestAzureSubscriptionAddAndRemove verifies that the SDK can perform the
// basic Azure subscription operations on a real RSC instance.
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
func TestAzureSubscriptionAddAndRemove(t *testing.T) {
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
	// environment variable AZURE_AUTH_LOCATION.
	_, err = azureClient.SetServicePrincipal(ctx, Default(testSubscription.TenantDomain))
	if err != nil {
		t.Fatal(err)
	}

	// Add default Azure subscription to RSC.
	subscription := Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	id, err := azureClient.AddSubscription(ctx, subscription, core.FeatureCloudNativeProtection,
		Regions("eastus2"), Name(testSubscription.SubscriptionName))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the subscription was successfully added.
	account, err := azureClient.Subscription(ctx, CloudAccountID(id), core.FeatureCloudNativeProtection)
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
		if name := account.Features[0].Name; name != "CLOUD_NATIVE_PROTECTION" {
			t.Errorf("invalid feature name: %v", name)
		}
		if regions := account.Features[0].Regions; !reflect.DeepEqual(regions, []string{"eastus2"}) {
			t.Errorf("invalid feature regions: %v", regions)
		}
		if account.Features[0].Status != "CONNECTED" {
			t.Errorf("invalid feature status: %v", account.Features[0].Status)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	// Update and verify regions for Azure account.
	err = azureClient.UpdateSubscription(ctx, ID(subscription), core.FeatureCloudNativeProtection,
		Regions("westus2"))
	if err != nil {
		t.Error(err)
	}
	account, err = azureClient.Subscription(ctx, ID(subscription), core.FeatureCloudNativeProtection)
	if err != nil {
		t.Error(err)
	}
	if n := len(account.Features); n == 1 {
		if regions := account.Features[0].Regions; !reflect.DeepEqual(regions, []string{"westus2"}) {
			t.Errorf("invalid feature regions: %v", regions)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	// Remove the Azure subscription from RSC keeping the snapshots.
	err = azureClient.RemoveSubscription(ctx, ID(subscription), core.FeatureCloudNativeProtection, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	_, err = azureClient.Subscription(ctx, ID(subscription), core.FeatureCloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAzureArchivalSubscriptionAddAndRemove verifies that the SDK can perform
// the adding and removal of Azure subscription for archival feature on a real
// RSC instance.
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
func TestAzureArchivalSubscriptionAddAndRemove(t *testing.T) {
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
	// environment variable AZURE_AUTH_LOCATION.
	_, err = azureClient.SetServicePrincipal(ctx, Default(testSubscription.TenantDomain))
	if err != nil {
		t.Fatal(err)
	}

	// Add default Azure subscription to RSC.
	subscription := Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	id, err := azureClient.AddSubscription(
		ctx,
		subscription,
		core.FeatureCloudNativeArchival,
		Regions("eastus2"),
		Name(testSubscription.SubscriptionName),
		ResourceGroup(testSubscription.Archival.ResourceGroupName, "eastus2", make(map[string]string)),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the subscription was successfully added.
	account, err := azureClient.Subscription(
		ctx,
		CloudAccountID(id),
		core.FeatureCloudNativeArchival,
	)
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
		if name := account.Features[0].Name; name != "CLOUD_NATIVE_ARCHIVAL" {
			t.Errorf("invalid feature name: %v", name)
		}
		if regions := account.Features[0].Regions; !reflect.DeepEqual(regions, []string{"eastus2"}) {
			t.Errorf("invalid feature regions: %v", regions)
		}
		if account.Features[0].Status != "CONNECTED" {
			t.Errorf("invalid feature status: %v", account.Features[0].Status)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	// Update and verify regions for Azure account.
	err = azureClient.UpdateSubscription(ctx, ID(subscription), core.FeatureCloudNativeArchival,
		Regions("westus2"))
	if err != nil {
		t.Error(err)
	}
	account, err = azureClient.Subscription(ctx, ID(subscription), core.FeatureCloudNativeArchival)
	if err != nil {
		t.Error(err)
	}
	if n := len(account.Features); n == 1 {
		if regions := account.Features[0].Regions; !reflect.DeepEqual(regions, []string{"westus2"}) {
			t.Errorf("invalid feature regions: %v", regions)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	// Remove the Azure subscription from RSC keeping the snapshots.
	err = azureClient.RemoveSubscription(ctx, ID(subscription), core.FeatureCloudNativeArchival, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	_, err = azureClient.Subscription(ctx, ID(subscription), core.FeatureCloudNativeArchival)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAzureArchivalEncryptionSubscriptionAddAndRemove verifies that the SDK
// can perform the adding and removal of Azure subscription for archival
// encryption feature on a real RSC instance.
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
func TestAzureArchivalEncryptionSubscriptionAddAndRemove(t *testing.T) {
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
	// environment variable AZURE_AUTH_LOCATION.
	_, err = azureClient.SetServicePrincipal(ctx, Default(testSubscription.TenantDomain))
	if err != nil {
		t.Fatal(err)
	}

	// Add subscription with archival feature as archival encryption is a child
	// feature and cannot be added without that.
	subscription := Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	_, err = azureClient.AddSubscription(
		ctx,
		subscription,
		core.FeatureCloudNativeArchival,
		Regions("eastus2"),
		Name(testSubscription.SubscriptionName),
		ResourceGroup(testSubscription.Archival.ResourceGroupName, "eastus2", make(map[string]string)),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Add archival encryption feature.
	subscription = Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	id, err := azureClient.AddSubscription(
		ctx,
		subscription,
		core.FeatureCloudNativeArchivalEncryption,
		Regions("eastus2"),
		Name(testSubscription.SubscriptionName),
		ManagedIdentity(
			testSubscription.Archival.ManagedIdentityName,
			testSubscription.Archival.ResourceGroupName,
			testSubscription.Archival.PrincipalID,
			"eastus2",
		),
		ResourceGroup(
			testSubscription.Archival.ResourceGroupName,
			"eastus2",
			make(map[string]string),
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the subscription was successfully added.
	account, err := azureClient.Subscription(
		ctx,
		CloudAccountID(id),
		core.FeatureCloudNativeArchivalEncryption,
	)
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
		if name := account.Features[0].Name; name != "CLOUD_NATIVE_ARCHIVAL_ENCRYPTION" {
			t.Errorf("invalid feature name: %v", name)
		}
		if regions := account.Features[0].Regions; !reflect.DeepEqual(regions, []string{"eastus2"}) {
			t.Errorf("invalid feature regions: %v", regions)
		}
		if account.Features[0].Status != "CONNECTED" {
			t.Errorf("invalid feature status: %v", account.Features[0].Status)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	// Update and verify regions for Azure account.
	err = azureClient.UpdateSubscription(ctx, ID(subscription), core.FeatureCloudNativeArchivalEncryption,
		Regions("westus2"))
	if err != nil {
		t.Error(err)
	}
	account, err = azureClient.Subscription(ctx, ID(subscription), core.FeatureCloudNativeArchivalEncryption)
	if err != nil {
		t.Error(err)
	}
	if n := len(account.Features); n == 1 {
		if regions := account.Features[0].Regions; !reflect.DeepEqual(regions, []string{"westus2"}) {
			t.Errorf("invalid feature regions: %v", regions)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	// Remove the Azure subscription from RSC keeping the snapshots. Removing
	// archival feature as encryption is a child feature of it.
	err = azureClient.RemoveSubscription(ctx, ID(subscription), core.FeatureCloudNativeArchival, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	_, err = azureClient.Subscription(ctx, ID(subscription), core.FeatureCloudNativeArchivalEncryption)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}
