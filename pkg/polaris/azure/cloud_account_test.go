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
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// client is the common RSC client used for tests. Reusing the same client
// reduces the risk of hitting rate limits when access tokens are created.
var client *polaris.Client

func TestMain(m *testing.M) {
	if testsetup.BoolEnvSet("TEST_INTEGRATION") {
		// Load configuration and create the client. Usually resolved using the
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

		version, err := client.GQL.DeploymentVersion(context.Background())
		if err != nil {
			fmt.Printf("failed to get deployment version: %v\n", err)
			os.Exit(1)
		}
		logger.Printf(log.Info, "Polaris version: %s", version)
	}

	os.Exit(m.Run())
}

// TestAzureSubscriptionAddAndRemove verifies that the SDK can perform the
// basic Azure subscription operations on a real RSC instance.
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

	// Add Azure subscription to RSC.
	subscription := Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	cnpRegions := Regions(testSubscription.CloudNativeProtection.Regions...)
	cnpResourceGroup := ResourceGroup(testSubscription.CloudNativeProtection.ResourceGroupName,
		testSubscription.CloudNativeProtection.ResourceGroupRegion, nil)
	id, err := azureClient.AddSubscription(ctx, subscription, core.FeatureCloudNativeProtection,
		Name(testSubscription.SubscriptionName), cnpRegions, cnpResourceGroup)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the subscription was successfully added.
	account, err := azureClient.SubscriptionByID(ctx, id)
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
	feature, ok := account.Feature(core.FeatureCloudNativeProtection)
	if !ok {
		t.Fatalf("%s feature not found", core.FeatureCloudNativeProtection)
	}
	if name := feature.Name; name != core.FeatureCloudNativeProtection.Name {
		t.Fatalf("invalid feature name: %v", name)
	}
	slices.Sort(feature.Regions)
	slices.Sort(testSubscription.CloudNativeProtection.Regions)
	if regions := feature.Regions; !slices.Equal(regions, testSubscription.CloudNativeProtection.Regions) {
		t.Fatalf("invalid feature regions: %v", regions)
	}
	if status := feature.Status; status != "CONNECTED" {
		t.Fatalf("invalid feature status: %v", status)
	}
	if name := feature.ResourceGroup.Name; name != testSubscription.CloudNativeProtection.ResourceGroupName {
		t.Fatalf("invalid feature resource group name: %v", name)
	}
	if region := feature.ResourceGroup.Region; region != testSubscription.CloudNativeProtection.ResourceGroupRegion {
		t.Fatalf("invalid feature resource group region: %v", region)
	}
	if tags := feature.ResourceGroup.Tags; !maps.Equal(tags, map[string]string{}) {
		t.Fatalf("invalid feature resource group tags: %v", tags)
	}

	// Update and verify regions for the Azure subscription.
	err = azureClient.UpdateSubscription(ctx, id, core.FeatureCloudNativeProtection, Regions("westus2"))
	if err != nil {
		t.Fatal(err)
	}
	account, err = azureClient.SubscriptionByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(account.Features); n != 1 {
		t.Fatalf("invalid number of features: %v", n)
	}
	feature, ok = account.Feature(core.FeatureCloudNativeProtection)
	if !ok {
		t.Fatalf("%s feature not found", core.FeatureCloudNativeProtection)
	}
	slices.Sort(feature.Regions)
	if regions := feature.Regions; !slices.Equal(regions, []string{"westus2"}) {
		t.Fatalf("invalid feature regions: %v", regions)
	}

	// Remove the Azure subscription from RSC keeping the snapshots.
	err = azureClient.RemoveSubscription(ctx, id, core.FeatureCloudNativeProtection, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	_, err = azureClient.SubscriptionByID(ctx, id)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAzureSubscriptionAddAndRemoveWithEntraID verifies that the SDK can add
// and remove an Azure subscription with an Entra ID group ID on a real RSC
// instance, and that the Entra ID group ID is correctly persisted and read
// back at the tenant level.
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
func TestAzureSubscriptionAddAndRemoveWithEntraID(t *testing.T) {
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

	// Dummy Entra ID group object ID used for testing the field plumbing.
	const entraGroupID = "00000000-0000-0000-0000-000000000001"

	// Add Azure subscription to RSC with an Entra ID group ID.
	subscription := Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	cnpRegions := Regions(testSubscription.CloudNativeProtection.Regions...)
	cnpResourceGroup := ResourceGroup(testSubscription.CloudNativeProtection.ResourceGroupName,
		testSubscription.CloudNativeProtection.ResourceGroupRegion, nil)
	id, err := azureClient.AddSubscription(ctx, subscription, core.FeatureCloudNativeProtection,
		Name(testSubscription.SubscriptionName), cnpRegions, cnpResourceGroup,
		EntraGroupID(entraGroupID))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the Entra ID group ID was persisted at the tenant level and
	// is readable via the subscription.
	account, err := azureClient.SubscriptionByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got := account.EntraGroupID; got != entraGroupID {
		t.Fatalf("invalid Entra ID group ID: %v", got)
	}

	// Enable the exocompute feature for the account.
	exoRegions := Regions(testSubscription.Exocompute.Regions...)
	exoResourceGroup := ResourceGroup(testSubscription.Exocompute.ResourceGroupName,
		testSubscription.Exocompute.ResourceGroupRegion, nil)
	exoAccountID, err := azureClient.AddSubscription(ctx, subscription, core.FeatureExocompute,
		exoRegions, exoResourceGroup)
	if err != nil {
		t.Fatal(err)
	}
	if exoAccountID != id {
		t.Fatal("cloud native protection and exocompute added to different cloud accounts")
	}

	// Verify that the account now has both features and the Entra ID group ID
	// is still correctly set.
	account, err = azureClient.SubscriptionByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(account.Features); n != 2 {
		t.Fatalf("invalid number of features: %v", n)
	}
	if _, ok := account.Feature(core.FeatureExocompute); !ok {
		t.Fatalf("%s feature not found", core.FeatureExocompute)
	}
	if got := account.EntraGroupID; got != entraGroupID {
		t.Fatalf("invalid Entra ID group ID after adding exocompute: %v", got)
	}

	// Remove exocompute feature first, then CNP.
	if err := azureClient.RemoveSubscription(ctx, id, core.FeatureExocompute, false); err != nil {
		t.Fatal(err)
	}

	// Remove the Azure subscription from RSC keeping the snapshots.
	if err := azureClient.RemoveSubscription(ctx, id, core.FeatureCloudNativeProtection, false); err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	_, err = azureClient.SubscriptionByID(ctx, id)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAzureUpdateEntraGroupID verifies that the Entra ID group ID can be
// changed on an existing subscription by calling
// UpgradeCloudAccountPermissionsWithoutOAuth with a new non-empty value.
//
// To run this test against an RSC instance, the following environment variables
// need to be set:
//   - RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   - TEST_INTEGRATION=1
//   - TEST_AZURESUBSCRIPTION_FILE=<path-to-test-azure-subscription-file>
//   - AZURE_AUTH_LOCATION=<path-to-azure-sdk-auth-file>
func TestAzureUpdateEntraGroupID(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	testSubscription, err := testsetup.AzureSubscription()
	if err != nil {
		t.Fatal(err)
	}

	azureClient := Wrap(client)

	_, err = azureClient.SetServicePrincipal(ctx, Default(testSubscription.TenantDomain))
	if err != nil {
		t.Fatal(err)
	}

	const (
		entraGroupID1 = "00000000-0000-0000-0000-000000000001"
		entraGroupID2 = "00000000-0000-0000-0000-000000000002"
	)

	// Onboard with the first Entra ID group ID.
	subscription := Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	cnpRegions := Regions(testSubscription.CloudNativeProtection.Regions...)
	cnpResourceGroup := ResourceGroup(testSubscription.CloudNativeProtection.ResourceGroupName,
		testSubscription.CloudNativeProtection.ResourceGroupRegion, nil)
	id, err := azureClient.AddSubscription(ctx, subscription, core.FeatureCloudNativeProtection,
		Name(testSubscription.SubscriptionName), cnpRegions, cnpResourceGroup,
		EntraGroupID(entraGroupID1))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		azureClient.RemoveSubscription(ctx, id, core.FeatureCloudNativeProtection, false)
	})

	// Change the Entra ID group ID by calling the upgrade mutation with the
	// new value.
	gqlClient := azure.Wrap(client.GQL)
	if err := gqlClient.UpgradeCloudAccountPermissionsWithoutOAuth(ctx, azure.PermissionUpgrade{
		CloudAccountID: id,
		Feature:        core.FeatureCloudNativeProtection,
		EntraGroupID:   entraGroupID2,
	}); err != nil {
		t.Fatal(err)
	}

	// Verify the new value was accepted and persisted.
	account, err := azureClient.SubscriptionByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got := account.EntraGroupID; got != entraGroupID2 {
		t.Fatalf("Entra ID group ID was not updated: got %q, want %q", got, entraGroupID2)
	}
}

// TestAzureUpdateSubscriptionEntraID verifies that UpdateSubscription
// correctly updates the Entra ID group ID on an existing subscription, both
// with and without explicit permission groups on the feature.
//
// The first update (without permission groups) is a regression test for a bug
// where the Entra ID group ID update was silently skipped when
// len(feature.PermissionGroups) == 0.
//
// To run this test against an RSC instance, the following environment variables
// need to be set:
//   - RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   - TEST_INTEGRATION=1
//   - TEST_AZURESUBSCRIPTION_FILE=<path-to-test-azure-subscription-file>
//   - AZURE_AUTH_LOCATION=<path-to-azure-sdk-auth-file>
func TestAzureUpdateSubscriptionEntraID(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	testSubscription, err := testsetup.AzureSubscription()
	if err != nil {
		t.Fatal(err)
	}

	azureClient := Wrap(client)

	_, err = azureClient.SetServicePrincipal(ctx, Default(testSubscription.TenantDomain))
	if err != nil {
		t.Fatal(err)
	}

	const (
		entraGroupID1 = "00000000-0000-0000-0000-000000000001"
		entraGroupID2 = "00000000-0000-0000-0000-000000000002"
		entraGroupID3 = "00000000-0000-0000-0000-000000000003"
	)

	// Onboard the subscription without an Entra ID group ID.
	subscription := Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	cnpRegions := Regions(testSubscription.CloudNativeProtection.Regions...)
	cnpResourceGroup := ResourceGroup(testSubscription.CloudNativeProtection.ResourceGroupName,
		testSubscription.CloudNativeProtection.ResourceGroupRegion, nil)
	id, err := azureClient.AddSubscription(ctx, subscription, core.FeatureCloudNativeProtection,
		Name(testSubscription.SubscriptionName), cnpRegions, cnpResourceGroup)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		azureClient.RemoveSubscription(ctx, id, core.FeatureCloudNativeProtection, false)
	})

	// Set the Entra ID group ID via UpdateSubscription using a bare feature
	// (no permission groups). This exercises the bug-fix path where
	// entraGroupID is checked independently of len(feature.PermissionGroups).
	if err := azureClient.UpdateSubscription(ctx, id, core.FeatureCloudNativeProtection,
		EntraGroupID(entraGroupID1)); err != nil {
		t.Fatal(err)
	}

	account, err := azureClient.SubscriptionByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got := account.EntraGroupID; got != entraGroupID1 {
		t.Fatalf("Entra ID group ID was not set: got %q, want %q", got, entraGroupID1)
	}

	// Update the Entra ID group ID again, still without permission groups.
	if err := azureClient.UpdateSubscription(ctx, id, core.FeatureCloudNativeProtection,
		EntraGroupID(entraGroupID2)); err != nil {
		t.Fatal(err)
	}

	account, err = azureClient.SubscriptionByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got := account.EntraGroupID; got != entraGroupID2 {
		t.Fatalf("Entra ID group ID was not updated: got %q, want %q", got, entraGroupID2)
	}

	// Update the Entra ID group ID using a feature with explicit permission
	// groups. The permission groups themselves have not changed, so pgChanged
	// will be false, but the Entra ID group ID update should still be applied.
	if err := azureClient.UpdateSubscription(ctx, id,
		core.FeatureCloudNativeProtection.WithPermissionGroups(core.PermissionGroupBasic),
		EntraGroupID(entraGroupID3)); err != nil {
		t.Fatal(err)
	}

	account, err = azureClient.SubscriptionByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got := account.EntraGroupID; got != entraGroupID3 {
		t.Fatalf("Entra ID group ID was not updated with permission groups: got %q, want %q", got, entraGroupID3)
	}
}

// TestAzureArchivalSubscriptionAddAndRemove verifies that the SDK can perform
// the adding and removal of Azure subscription for archival feature on a real
// RSC instance.
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

	// Add Azure subscription to RSC.
	subscription := Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	arcRegions := Regions(testSubscription.Archival.Regions...)
	arcResourceGroup := ResourceGroup(testSubscription.Archival.ResourceGroupName,
		testSubscription.Archival.ResourceGroupRegion, nil)
	id, err := azureClient.AddSubscription(ctx, subscription, core.FeatureCloudNativeArchival,
		Name(testSubscription.SubscriptionName), arcRegions, arcResourceGroup)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the subscription was successfully added.
	account, err := azureClient.SubscriptionByID(ctx, id)
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
	feature, ok := account.Feature(core.FeatureCloudNativeArchival)
	if !ok {
		t.Fatalf("%s feature not found", core.FeatureCloudNativeArchival)
	}
	if name := feature.Name; name != core.FeatureCloudNativeArchival.Name {
		t.Fatalf("invalid feature name: %v", name)
	}
	slices.Sort(feature.Regions)
	slices.Sort(testSubscription.Archival.Regions)
	if regions := feature.Regions; !slices.Equal(regions, testSubscription.Archival.Regions) {
		t.Fatalf("invalid feature regions: %v", regions)
	}
	if status := feature.Status; status != "CONNECTED" {
		t.Fatalf("invalid feature status: %v", status)
	}
	if name := feature.ResourceGroup.Name; name != testSubscription.Archival.ResourceGroupName {
		t.Fatalf("invalid feature resource group name: %v", name)
	}
	if region := feature.ResourceGroup.Region; region != testSubscription.Archival.ResourceGroupRegion {
		t.Fatalf("invalid feature resource group region: %v", region)
	}
	if tags := feature.ResourceGroup.Tags; !maps.Equal(tags, map[string]string{}) {
		t.Fatalf("invalid feature resource group tags: %v", tags)
	}

	// Update and verify regions for Azure subscription.
	err = azureClient.UpdateSubscription(ctx, id, core.FeatureCloudNativeArchival, Regions("westus2"))
	if err != nil {
		t.Fatal(err)
	}
	account, err = azureClient.SubscriptionByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(account.Features); n != 1 {
		t.Fatalf("invalid number of features: %v", n)
	}
	feature, ok = account.Feature(core.FeatureCloudNativeArchival)
	if !ok {
		t.Fatalf("%s feature not found", core.FeatureCloudNativeArchival)
	}
	slices.Sort(feature.Regions)
	if regions := feature.Regions; !slices.Equal(regions, []string{"westus2"}) {
		t.Fatalf("invalid feature regions: %v", regions)
	}

	// Remove the Azure subscription from RSC keeping the snapshots.
	err = azureClient.RemoveSubscription(ctx, id, core.FeatureCloudNativeArchival, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	_, err = azureClient.SubscriptionByID(ctx, id)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAzureArchivalEncryptionSubscriptionAddAndRemove verifies that the SDK
// can perform the adding and removal of Azure subscription for archival
// encryption feature on a real RSC instance.
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
	arcRegions := Regions(testSubscription.Archival.Regions...)
	arcResourceGroup := ResourceGroup(testSubscription.Archival.ResourceGroupName,
		testSubscription.Archival.ResourceGroupRegion, nil)
	_, err = azureClient.AddSubscription(ctx, subscription, core.FeatureCloudNativeArchival,
		Name(testSubscription.SubscriptionName), arcRegions, arcResourceGroup)
	if err != nil {
		t.Fatal(err)
	}

	// Add archival encryption feature.
	encManagedIdentity := ManagedIdentity(testSubscription.Archival.ManagedIdentityName,
		testSubscription.Archival.ResourceGroupName, testSubscription.Archival.PrincipalID,
		testSubscription.Archival.ResourceGroupRegion)
	id, err := azureClient.AddSubscription(ctx, subscription, core.FeatureCloudNativeArchivalEncryption,
		Name(testSubscription.SubscriptionName), arcRegions, arcResourceGroup, encManagedIdentity)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the subscription was successfully added.
	account, err := azureClient.SubscriptionByID(ctx, id)
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
	if n := len(account.Features); n != 2 {
		t.Fatalf("invalid number of features: %v", n)
	}
	feature, ok := account.Feature(core.FeatureCloudNativeArchivalEncryption)
	if !ok {
		t.Fatalf("%s feature not found", core.FeatureCloudNativeArchivalEncryption)
	}
	if name := feature.Name; name != core.FeatureCloudNativeArchivalEncryption.Name {
		t.Fatalf("invalid feature name: %v", name)
	}
	slices.Sort(feature.Regions)
	slices.Sort(testSubscription.Archival.Regions)
	if regions := feature.Regions; !slices.Equal(regions, testSubscription.Archival.Regions) {
		t.Fatalf("invalid feature regions: %v", regions)
	}
	if status := feature.Status; status != "CONNECTED" {
		t.Fatalf("invalid feature status: %v", status)
	}
	if name := feature.ResourceGroup.Name; name != testSubscription.Archival.ResourceGroupName {
		t.Fatalf("invalid feature resource group name: %v", name)
	}
	if region := feature.ResourceGroup.Region; region != testSubscription.Archival.ResourceGroupRegion {
		t.Fatalf("invalid feature resource group region: %v", region)
	}
	if tags := feature.ResourceGroup.Tags; !maps.Equal(tags, map[string]string{}) {
		t.Fatalf("invalid feature resource group tags: %v", tags)
	}

	// Update and verify regions for the Azure account.
	err = azureClient.UpdateSubscription(ctx, id, core.FeatureCloudNativeArchivalEncryption,
		Regions("westus2"))
	if err != nil {
		t.Fatal(err)
	}
	account, err = azureClient.SubscriptionByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(account.Features); n != 2 {
		t.Fatalf("invalid number of features: %v", n)
	}
	feature, ok = account.Feature(core.FeatureCloudNativeArchivalEncryption)
	if !ok {
		t.Fatalf("%s feature not found", core.FeatureCloudNativeArchivalEncryption)
	}
	slices.Sort(feature.Regions)
	if regions := feature.Regions; !slices.Equal(regions, []string{"westus2"}) {
		t.Fatalf("invalid feature regions: %v", regions)
	}

	// Remove the Azure subscription from RSC keeping the snapshots. Removing
	// archival feature as encryption is a child feature of it.
	err = azureClient.RemoveSubscription(ctx, id, core.FeatureCloudNativeArchival, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	_, err = azureClient.SubscriptionByID(ctx, id)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAzureUpdateSubscriptionRegionIsolation verifies that updating the
// regions for one feature does not affect the regions of other features on the
// same subscription.
//
// This is a regression test for a bug where UpdateSubscription iterated over
// all account features and applied the desired regions to each one, causing
// unrelated features to have their regions overwritten.
//
// To run this test against an RSC instance, the following environment variables
// need to be set:
//   - RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   - TEST_INTEGRATION=1
//   - TEST_AZURESUBSCRIPTION_FILE=<path-to-test-azure-subscription-file>
//   - AZURE_AUTH_LOCATION=<path-to-azure-sdk-auth-file>
func TestAzureUpdateSubscriptionRegionIsolation(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	testSubscription, err := testsetup.AzureSubscription()
	if err != nil {
		t.Fatal(err)
	}

	azureClient := Wrap(client)

	_, err = azureClient.SetServicePrincipal(ctx, Default(testSubscription.TenantDomain))
	if err != nil {
		t.Fatal(err)
	}

	// Add the subscription with CNP.
	subscription := Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	cnpResourceGroup := ResourceGroup(testSubscription.CloudNativeProtection.ResourceGroupName,
		testSubscription.CloudNativeProtection.ResourceGroupRegion, nil)
	id, err := azureClient.AddSubscription(ctx, subscription, core.FeatureCloudNativeProtection,
		Name(testSubscription.SubscriptionName),
		Regions(testSubscription.CloudNativeProtection.Regions...),
		cnpResourceGroup)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		azureClient.RemoveSubscription(ctx, id, core.FeatureExocompute, false)
		azureClient.RemoveSubscription(ctx, id, core.FeatureCloudNativeProtection, false)
	})

	// Add Exocompute to the same subscription with its own distinct regions.
	exoResourceGroup := ResourceGroup(testSubscription.Exocompute.ResourceGroupName,
		testSubscription.Exocompute.ResourceGroupRegion, nil)
	if _, err = azureClient.AddSubscription(ctx, subscription, core.FeatureExocompute,
		Regions(testSubscription.Exocompute.Regions...),
		exoResourceGroup); err != nil {
		t.Fatal(err)
	}

	// Record the initial Exocompute regions before we touch CNP.
	account, err := azureClient.SubscriptionByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	exoFeature, ok := account.Feature(core.FeatureExocompute)
	if !ok {
		t.Fatalf("%s feature not found", core.FeatureExocompute)
	}
	initialExoRegions := slices.Clone(exoFeature.Regions)
	slices.Sort(initialExoRegions)

	// Update only CNP regions to "westus2".
	if err := azureClient.UpdateSubscription(ctx, id, core.FeatureCloudNativeProtection,
		Regions("westus2")); err != nil {
		t.Fatal(err)
	}

	// Re-read the account and verify that:
	//   1. CNP now has the new regions.
	//   2. Exocompute regions are unchanged.
	account, err = azureClient.SubscriptionByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	cnpFeature, ok := account.Feature(core.FeatureCloudNativeProtection)
	if !ok {
		t.Fatalf("%s feature not found after update", core.FeatureCloudNativeProtection)
	}
	slices.Sort(cnpFeature.Regions)
	if !slices.Equal(cnpFeature.Regions, []string{"westus2"}) {
		t.Fatalf("CNP regions after update: got %v, want [westus2]", cnpFeature.Regions)
	}

	exoFeature, ok = account.Feature(core.FeatureExocompute)
	if !ok {
		t.Fatalf("%s feature not found after CNP update", core.FeatureExocompute)
	}
	slices.Sort(exoFeature.Regions)
	if !slices.Equal(exoFeature.Regions, initialExoRegions) {
		t.Fatalf("Exocompute regions changed after CNP update: got %v, want %v",
			exoFeature.Regions, initialExoRegions)
	}
}

func TestToSubscription(t *testing.T) {
	rawTenants, err := allAzureCloudAccountTenantsResponse()
	if err != nil {
		t.Fatal(err)
	}

	subs := toSubscriptions(rawTenants)
	if n := len(subs); n != 3 {
		t.Fatalf("invalid number of subscriptions: %v", n)
	}

	// Subscription 1.
	if subs[0].ID != uuid.MustParse("f4b69681-2ab8-4edc-81c2-8852e46c1ba3") {
		t.Errorf("invalid id: %v", subs[0].ID)
	}
	if subs[0].NativeID != uuid.MustParse("c263212c-3f26-4b9a-8601-9efb466c8837") {
		t.Errorf("invalid native id: %v", subs[0].NativeID)
	}
	if subs[0].Name != "subscription1" {
		t.Errorf("invalid name: %v", subs[0].Name)
	}
	if subs[0].TenantID != uuid.MustParse("ca997d29-1811-4aab-a5dc-649082debe89") {
		t.Errorf("invalid tenant id: %v", subs[0].TenantID)
	}
	if subs[0].TenantDomain != "domain1.onmicrosoft.com" {
		t.Errorf("invalid tenant domain: %v", subs[0].TenantDomain)
	}
	if subs[0].EntraGroupID != "" {
		t.Errorf("invalid entra group id: %v", subs[0].EntraGroupID)
	}
	if !reflect.DeepEqual(subs[0].Features, []Feature{{
		Feature: core.Feature{Name: "CLOUD_NATIVE_PROTECTION"},
		ResourceGroup: FeatureResourceGroup{
			Name:     "rg1",
			NativeID: "/subscriptions/f4b69681-2ab8-4edc-81c2-8852e46c1ba3/resourceGroups/rg1",
			Region:   "westus",
			Tags:     map[string]string{},
		},
		Regions: []string{"eastus", "westus"},
		Status:  "MISSING_PERMISSIONS",
	}, {
		Feature: core.Feature{Name: "EXOCOMPUTE"},
		ResourceGroup: FeatureResourceGroup{
			Name:     "rg2",
			NativeID: "/subscriptions/f4b69681-2ab8-4edc-81c2-8852e46c1ba3/resourceGroups/rg2",
			Region:   "westus",
			Tags:     map[string]string{},
		},
		Regions: []string{"westus"},
		Status:  "CONNECTED",
	}}) {
		t.Errorf("invalid features: %v", subs[0].Features)
	}

	// Subscription 2.
	if subs[1].ID != uuid.MustParse("e2e3fb63-2230-4154-9b1b-923f018dbc4f") {
		t.Errorf("invalid id: %v", subs[1].ID)
	}
	if subs[1].NativeID != uuid.MustParse("1ee74f16-10d3-45fe-adfb-7f70ee77f5ee") {
		t.Errorf("invalid native id: %v", subs[1].NativeID)
	}
	if subs[1].Name != "subscription2" {
		t.Errorf("invalid name: %v", subs[1].Name)
	}
	if subs[1].TenantID != uuid.MustParse("88af4472-ea52-4c8e-bf05-e4ca581370a7") {
		t.Errorf("invalid tenant id: %v", subs[1].TenantID)
	}
	if subs[1].TenantDomain != "domain2.onmicrosoft.com" {
		t.Errorf("invalid tenant domain: %v", subs[1].TenantDomain)
	}
	if subs[1].EntraGroupID != "00000000-0000-0000-0000-000000000042" {
		t.Errorf("invalid entra group id: %v", subs[1].EntraGroupID)
	}
	if !reflect.DeepEqual(subs[1].Features, []Feature{{
		Feature: core.Feature{Name: "CLOUD_NATIVE_PROTECTION"},
		ResourceGroup: FeatureResourceGroup{
			Name:     "rg3",
			NativeID: "/subscriptions/e2e3fb63-2230-4154-9b1b-923f018dbc4f/resourceGroups/rg3",
			Region:   "westus2",
			Tags:     map[string]string{},
		},
		Regions: []string{"westus2"},
		Status:  "MISSING_PERMISSIONS",
	}, {
		Feature: core.Feature{Name: "EXOCOMPUTE"},
		ResourceGroup: FeatureResourceGroup{
			Name:     "rg4",
			NativeID: "/subscriptions/e2e3fb63-2230-4154-9b1b-923f018dbc4f/resourceGroups/rg4",
			Region:   "westus2",
			Tags:     map[string]string{},
		},
		Regions: []string{},
		Status:  "CONNECTED",
	}}) {
		t.Errorf("invalid features: %v", subs[1].Features)
	}

	// Subscription 3.
	if subs[2].ID != uuid.MustParse("31116cf6-6259-4cfc-b8a6-307cb0744ba1") {
		t.Errorf("invalid id: %v", subs[2].ID)
	}
	if subs[2].NativeID != uuid.MustParse("973bfa00-0bfd-4850-aab1-ebd3f9d9b6b7") {
		t.Errorf("invalid native id: %v", subs[2].NativeID)
	}
	if subs[2].Name != "subscription3" {
		t.Errorf("invalid name: %v", subs[2].Name)
	}
	if subs[2].TenantID != uuid.MustParse("88af4472-ea52-4c8e-bf05-e4ca581370a7") {
		t.Errorf("invalid tenant id: %v", subs[2].TenantID)
	}
	if subs[2].TenantDomain != "domain2.onmicrosoft.com" {
		t.Errorf("invalid tenant domain: %v", subs[2].TenantDomain)
	}
	if subs[2].EntraGroupID != "00000000-0000-0000-0000-000000000042" {
		t.Errorf("invalid entra group id: %v", subs[2].EntraGroupID)
	}
	if !reflect.DeepEqual(subs[2].Features, []Feature{{
		Feature: core.Feature{Name: "CLOUD_NATIVE_PROTECTION"},
		ResourceGroup: FeatureResourceGroup{
			Name:     "rg5",
			NativeID: "/subscriptions/31116cf6-6259-4cfc-b8a6-307cb0744ba1/resourceGroups/rg5",
			Region:   "westus",
			Tags: map[string]string{
				"key1": "value1",
			},
		},
		Regions: []string{"westus"},
		Status:  "MISSING_PERMISSIONS",
	}, {
		Feature: core.Feature{Name: "EXOCOMPUTE"},
		ResourceGroup: FeatureResourceGroup{
			Name:     "rg5",
			NativeID: "/subscriptions/31116cf6-6259-4cfc-b8a6-307cb0744ba1/resourceGroups/rg5",
			Region:   "westus",
			Tags:     map[string]string{},
		},
		Regions: []string{"westus"},
		Status:  "MISSING_PERMISSIONS",
	}}) {
		t.Errorf("invalid features: %v", subs[2].Features)
	}
}

func TestToTenant(t *testing.T) {
	rawTenants, err := allAzureCloudAccountTenantsResponse()
	if err != nil {
		t.Fatal(err)
	}

	tenants := toTenants(rawTenants)
	if n := len(tenants); n != 2 {
		t.Errorf("invalid number of tenants: %v", n)
	}

	// Tenant 1.
	if tenants[0].Cloud != "AZUREPUBLICCLOUD" {
		t.Errorf("invalid cloud: %v", tenants[0].Cloud)
	}
	if tenants[0].ID != uuid.MustParse("ca997d29-1811-4aab-a5dc-649082debe89") {
		t.Errorf("invalid id: %v", tenants[0].ID)
	}
	if tenants[0].ClientID != uuid.MustParse("b6a26799-b722-4df6-b2df-9c70433ee55f") {
		t.Errorf("invalid client id: %v", tenants[0].ClientID)
	}
	if tenants[0].AppName != "app1" {
		t.Errorf("invalid app name: %v", tenants[0].AppName)
	}
	if tenants[0].DomainName != "domain1.onmicrosoft.com" {
		t.Errorf("invalid domain: %v", tenants[0].DomainName)
	}
	if tenants[0].EntraGroupID != "" {
		t.Errorf("invalid entra group id: %v", tenants[0].EntraGroupID)
	}
	if tenants[0].SubscriptionCount != 1 {
		t.Errorf("invalid subscription count: %v", tenants[0].SubscriptionCount)
	}

	// Tenant 2.
	if tenants[1].Cloud != "AZUREPUBLICCLOUD" {
		t.Errorf("invalid cloud: %v", tenants[1].Cloud)
	}
	if tenants[1].ID != uuid.MustParse("88af4472-ea52-4c8e-bf05-e4ca581370a7") {
		t.Errorf("invalid id: %v", tenants[1].ID)
	}
	if tenants[1].ClientID != uuid.MustParse("6688f45e-b1dc-41d8-b926-3acef4a4beaf") {
		t.Errorf("invalid client id: %v", tenants[1].ClientID)
	}
	if tenants[1].AppName != "app2" {
		t.Errorf("invalid app name: %v", tenants[1].AppName)
	}
	if tenants[1].DomainName != "domain2.onmicrosoft.com" {
		t.Errorf("invalid domain: %v", tenants[1].DomainName)
	}
	if tenants[1].EntraGroupID != "00000000-0000-0000-0000-000000000042" {
		t.Errorf("invalid entra group id: %v", tenants[1].EntraGroupID)
	}
	if tenants[1].SubscriptionCount != 2 {
		t.Errorf("invalid subscription count: %v", tenants[1].SubscriptionCount)
	}
}

func allAzureCloudAccountTenantsResponse() ([]azure.CloudAccountTenant, error) {
	buf, err := os.ReadFile("testdata/all_azure_cloud_account_tenants_response.json")
	if err != nil {
		return nil, err
	}

	var payload struct {
		Data struct {
			Result []azure.CloudAccountTenant `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, err
	}

	return payload.Data.Result, nil
}
