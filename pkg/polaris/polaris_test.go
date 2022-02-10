// Copyright 2021 Rubrik, Inc.
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

package polaris

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	polaris_log "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// client is the common Polaris client used for tests. By reusing the same
// client we reduce the risk of hitting rate limits when access tokens are
// created.
var client *Client

func TestMain(m *testing.M) {
	if testsetup.BoolEnvSet("TEST_INTEGRATION") {
		// Load configuration and create client. Usually resolved using the
		// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
		polAccount, err := DefaultServiceAccount(true)
		if err != nil {
			fmt.Printf("failed to get default service account: %v\n", err)
			os.Exit(1)
		}

		// The integration tests defaults the log level to INFO. Note that
		// RUBRIK_POLARIS_LOGLEVEL can be used to override this.
		logger := polaris_log.NewStandardLogger()
		logger.SetLogLevel(polaris_log.Info)
		client, err = NewClient(context.Background(), polAccount, logger)
		if err != nil {
			fmt.Printf("failed to create polaris client: %v\n", err)
			os.Exit(1)
		}
	}

	os.Exit(m.Run())
}

// TestAwsAccountAddAndRemove verifies that the SDK can perform the basic AWS
// account operations on a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * TEST_INTEGRATION=1
//   * TEST_AWSACCOUNT_FILE=<path-to-test-aws-account-file>
//   * AWS_SHARED_CREDENTIALS_FILE=<path-to-aws-credentials-file>
//   * AWS_CONFIG_FILE=<path-to-aws-config-file>
//
// The file referred to by TEST_AWSACCOUNT_FILE should contain a single
// testAwsAccount JSON object.
func TestAwsAccountAddAndRemove(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	ctx := context.Background()

	testAccount, err := testsetup.AWSAccount()
	if err != nil {
		t.Fatal(err)
	}

	// Adds the AWS account identified by the specified profile to Polaris. Note
	// that the profile needs to have a default region.
	id, err := client.AWS().AddAccount(ctx, aws.Profile(testAccount.Profile), core.FeatureCloudNativeProtection,
		aws.Name(testAccount.AccountName), aws.Regions("us-east-2"))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully added.
	account, err := client.AWS().Account(ctx, aws.CloudAccountID(id), core.FeatureCloudNativeProtection)
	if err != nil {
		t.Error(err)
	}
	if account.Name != testAccount.AccountName {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.NativeID != testAccount.AccountID {
		t.Errorf("invalid native id: %v", account.NativeID)
	}
	if n := len(account.Features); n == 1 {
		if account.Features[0].Name != "CLOUD_NATIVE_PROTECTION" {
			t.Errorf("invalid feature name: %v", account.Features[0].Name)
		}
		if regions := account.Features[0].Regions; !reflect.DeepEqual(regions, []string{"us-east-2"}) {
			t.Errorf("invalid feature regions: %v", regions)
		}
		if account.Features[0].Status != "CONNECTED" {
			t.Errorf("invalid feature status: %v", account.Features[0].Status)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	// Update and verify regions for AWS account.
	err = client.AWS().UpdateAccount(ctx, aws.AccountID(testAccount.AccountID), core.FeatureCloudNativeProtection,
		aws.Regions("us-west-2"))
	if err != nil {
		t.Error(err)
	}
	account, err = client.AWS().Account(ctx, aws.AccountID(testAccount.AccountID), core.FeatureCloudNativeProtection)
	if err != nil {
		t.Error(err)
	}
	if n := len(account.Features); n == 1 {
		if regions := account.Features[0].Regions; !reflect.DeepEqual(regions, []string{"us-west-2"}) {
			t.Errorf("invalid feature regions: %v", regions)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	// Remove AWS account from Polaris.
	err = client.AWS().RemoveAccount(ctx, aws.Profile(testAccount.Profile), core.FeatureCloudNativeProtection, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	account, err = client.AWS().Account(ctx, aws.AccountID(testAccount.AccountID), core.FeatureCloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAwsExocompute verifies that the SDK can perform basic Exocompute
// operations on a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * TEST_INTEGRATION=1
//   * TEST_AWSACCOUNT_FILE=<path-to-test-aws-account-file>
//   * AWS_SHARED_CREDENTIALS_FILE=<path-to-aws-credentials-file>
//   * AWS_CONFIG_FILE=<path-to-aws-config-file>
//
// The file referred to by TEST_AWSACCOUNT_FILE should contain a single
// testAwsAccount JSON object.
func TestAwsExocompute(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	ctx := context.Background()

	testAccount, err := testsetup.AWSAccount()
	if err != nil {
		t.Fatal(err)
	}

	// Adds the AWS account identified by the specified profile to Polaris. Note
	// that the profile needs to have a default region.
	accountID, err := client.AWS().AddAccount(ctx, aws.Profile(testAccount.Profile), core.FeatureCloudNativeProtection,
		aws.Name(testAccount.AccountName), aws.Regions("us-east-2"))
	if err != nil {
		t.Fatal(err)
	}

	// Enable the exocompute feature for the account.
	exoAccountID, err := client.AWS().AddAccount(ctx, aws.Profile(testAccount.Profile), core.FeatureExocompute,
		aws.Regions("us-east-2"))
	if err != nil {
		t.Fatal(err)
	}
	if accountID != exoAccountID {
		t.Fatalf("cloud native protection and exocompute added to different cloud accounts: %q vs %q",
			accountID, exoAccountID)
	}

	// Verify that the account was successfully added.
	account, err := client.AWS().Account(ctx, aws.CloudAccountID(accountID), core.FeatureExocompute)
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != testAccount.AccountName {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.NativeID != testAccount.AccountID {
		t.Errorf("invalid native id: %v", account.NativeID)
	}
	if n := len(account.Features); n == 1 {
		if account.Features[0].Name != "EXOCOMPUTE" {
			t.Errorf("invalid feature name: %v", account.Features[0].Name)
		}
		if regions := account.Features[0].Regions; !reflect.DeepEqual(regions, []string{"us-east-2"}) {
			t.Errorf("invalid feature regions: %v", regions)
		}
		if account.Features[0].Status != "CONNECTED" {
			t.Errorf("invalid feature status: %v", account.Features[0].Status)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	exoID, err := client.AWS().AddExocomputeConfig(ctx, aws.AccountID(testAccount.AccountID),
		aws.Managed("us-east-2", testAccount.Exocompute.VPCID,
			[]string{testAccount.Exocompute.Subnets[0].ID, testAccount.Exocompute.Subnets[1].ID}))
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve the exocompute config added.
	exoConfig, err := client.AWS().ExocomputeConfig(ctx, exoID)
	if err != nil {
		t.Fatal(err)
	}
	if exoConfig.ID != exoID {
		t.Errorf("invalid id: %v", exoConfig.ID)
	}
	if exoConfig.Region != "us-east-2" {
		t.Errorf("invalid region: %v", exoConfig.Region)
	}
	if exoConfig.VPCID != testAccount.Exocompute.VPCID {
		t.Errorf("invalid vpc id: %v", exoConfig.VPCID)
	}
	if exoConfig.Subnets[0].ID != testAccount.Exocompute.Subnets[0].ID && exoConfig.Subnets[0].ID != testAccount.Exocompute.Subnets[1].ID {
		t.Errorf("invalid subnet id: %v", exoConfig.Subnets[0].ID)
	}
	if exoConfig.Subnets[0].AvailabilityZone != testAccount.Exocompute.Subnets[0].AvailabilityZone && exoConfig.Subnets[0].AvailabilityZone != testAccount.Exocompute.Subnets[1].AvailabilityZone {
		t.Errorf("invalid subnet availability zone: %v", exoConfig.Subnets[0].AvailabilityZone)
	}
	if exoConfig.Subnets[1].ID != testAccount.Exocompute.Subnets[0].ID && exoConfig.Subnets[1].ID != testAccount.Exocompute.Subnets[1].ID {
		t.Errorf("invalid subnet id: %v", exoConfig.Subnets[1].ID)
	}
	if exoConfig.Subnets[1].AvailabilityZone != testAccount.Exocompute.Subnets[0].AvailabilityZone && exoConfig.Subnets[1].AvailabilityZone != testAccount.Exocompute.Subnets[1].AvailabilityZone {
		t.Errorf("invalid subnet availability zone: %v", exoConfig.Subnets[1].AvailabilityZone)
	}
	if !exoConfig.PolarisManaged {
		t.Errorf("invalid polaris managed state: %t", exoConfig.PolarisManaged)
	}

	// Remove the exocompute config.
	err = client.AWS().RemoveExocomputeConfig(ctx, exoID)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the exocompute config was successfully removed.
	exoConfig, err = client.AWS().ExocomputeConfig(ctx, exoID)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}

	// Disable the exocompute feature for the account.
	err = client.AWS().RemoveAccount(ctx, aws.Profile(testAccount.Profile), core.FeatureExocompute, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the exocompute feature was successfully disabled.
	account, err = client.AWS().Account(ctx, aws.AccountID(testAccount.AccountID), core.FeatureExocompute)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}

	// Remove the AWS account from Polaris.
	err = client.AWS().RemoveAccount(ctx, aws.Profile(testAccount.Profile), core.FeatureCloudNativeProtection, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	account, err = client.AWS().Account(ctx, aws.AccountID(testAccount.AccountID), core.FeatureCloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAzureSubscriptionAddAndRemove verifies that the SDK can perform the
// basic Azure subscription operations on a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * TEST_INTEGRATION=1
//   * TEST_AZURESUBSCRIPTION_FILE=<path-to-test-azure-subscription-file>
//   * AZURE_AUTH_LOCATION=<path-to-azure-sdk-auth-file>
//
// The file referred to by TEST_AZURESUBSCRIPTION_FILE should contain a single
// testAzureSubscription JSON object.
func TestAzureSubscriptionAddAndRemove(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	ctx := context.Background()

	testSubscription, err := testsetup.AzureSubscription()
	if err != nil {
		t.Fatal(err)
	}

	// Add default Azure service principal to Polaris. Usually resolved using
	// the environment variable AZURE_AUTH_LOCATION.
	_, err = client.Azure().SetServicePrincipal(ctx, azure.Default(testSubscription.TenantDomain))
	if err != nil {
		t.Fatal(err)
	}

	// Add default Azure subscription to Polaris.
	subscription := azure.Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	id, err := client.Azure().AddSubscription(ctx, subscription, core.FeatureCloudNativeProtection,
		azure.Regions("eastus2"), azure.Name(testSubscription.SubscriptionName))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the subscription was successfully added.
	account, err := client.Azure().Subscription(ctx, azure.CloudAccountID(id), core.FeatureCloudNativeProtection)
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
	err = client.Azure().UpdateSubscription(ctx, azure.ID(subscription), core.FeatureCloudNativeProtection,
		azure.Regions("westus2"))
	if err != nil {
		t.Error(err)
	}
	account, err = client.Azure().Subscription(ctx, azure.ID(subscription), core.FeatureCloudNativeProtection)
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

	// Remove the Azure subscription from Polaris keeping the snapshots.
	err = client.Azure().RemoveSubscription(ctx, azure.ID(subscription), core.FeatureCloudNativeProtection, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	_, err = client.Azure().Subscription(ctx, azure.ID(subscription), core.FeatureCloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAzureArchivalSubscriptionAddAndRemove verifies that the SDK can perform the
// adding and removal of Azure subscription for archival feature
// on a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * TEST_INTEGRATION=1
//   * TEST_AZURESUBSCRIPTION_FILE=<path-to-test-azure-subscription-file>
//   * AZURE_AUTH_LOCATION=<path-to-azure-sdk-auth-file>
//
// The file referred to by TEST_AZURESUBSCRIPTION_FILE should contain a single
// testAzureSubscription JSON object.
func TestAzureArchivalSubscriptionAddAndRemove(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	ctx := context.Background()

	testSubscription, err := testsetup.AzureSubscription()
	if err != nil {
		t.Fatal(err)
	}

	// Add default Azure service principal to Polaris. Usually resolved using
	// the environment variable AZURE_AUTH_LOCATION.
	_, err = client.Azure().SetServicePrincipal(ctx, azure.Default(testSubscription.TenantDomain))
	if err != nil {
		t.Fatal(err)
	}

	// Add default Azure subscription to Polaris.
	subscription := azure.Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	id, err := client.Azure().AddSubscription(
		ctx,
		subscription,
		core.FeatureCloudNativeArchival,
		azure.Regions("eastus2"),
		azure.Name(testSubscription.SubscriptionName),
		azure.ResourceGroup(testSubscription.Archival.ResourceGroupName, "eastus2", make(map[string]string)),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the subscription was successfully added.
	account, err := client.Azure().Subscription(
		ctx,
		azure.CloudAccountID(id),
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
	err = client.Azure().UpdateSubscription(ctx, azure.ID(subscription), core.FeatureCloudNativeArchival,
		azure.Regions("westus2"))
	if err != nil {
		t.Error(err)
	}
	account, err = client.Azure().Subscription(ctx, azure.ID(subscription), core.FeatureCloudNativeArchival)
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

	// Remove the Azure subscription from Polaris keeping the snapshots.
	err = client.Azure().RemoveSubscription(ctx, azure.ID(subscription), core.FeatureCloudNativeArchival, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	_, err = client.Azure().Subscription(ctx, azure.ID(subscription), core.FeatureCloudNativeArchival)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAzureArchivalEncryptionSubscriptionAddAndRemove verifies that the
// SDK can perform the adding and removal of Azure subscription
// for archival encryption feature on a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * TEST_INTEGRATION=1
//   * TEST_AZURESUBSCRIPTION_FILE=<path-to-test-azure-subscription-file>
//   * AZURE_AUTH_LOCATION=<path-to-azure-sdk-auth-file>
//
// The file referred to by TEST_AZURESUBSCRIPTION_FILE should contain a single
// testAzureSubscription JSON object.
func TestAzureArchivalEncryptionSubscriptionAddAndRemove(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	ctx := context.Background()

	testSubscription, err := testsetup.AzureSubscription()
	if err != nil {
		t.Fatal(err)
	}

	// Add default Azure service principal to Polaris. Usually resolved using
	// the environment variable AZURE_AUTH_LOCATION.
	_, err = client.Azure().SetServicePrincipal(ctx, azure.Default(testSubscription.TenantDomain))
	if err != nil {
		t.Fatal(err)
	}

	// Add subscription with archival feature
	// as archival encryption is a child feature and cannot be
	// added without that.
	subscription := azure.Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	id, err := client.Azure().AddSubscription(
		ctx,
		subscription,
		core.FeatureCloudNativeArchival,
		azure.Regions("eastus2"),
		azure.Name(testSubscription.SubscriptionName),
		azure.ResourceGroup(testSubscription.Archival.ResourceGroupName, "eastus2", make(map[string]string)),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Add archival encryption feature.
	subscription = azure.Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	id, err = client.Azure().AddSubscription(
		ctx,
		subscription,
		core.FeatureCloudNativeArchivalEncryption,
		azure.Regions("eastus2"),
		azure.Name(testSubscription.SubscriptionName),
		azure.ManagedIdentity(
			testSubscription.Archival.ManagedIdentityName,
			testSubscription.Archival.ResourceGroupName,
			testSubscription.Archival.PrincipalID,
			"eastus2",
		),
		azure.ResourceGroup(
			testSubscription.Archival.ResourceGroupName,
			"eastus2",
			make(map[string]string),
		),
	)

	if err != nil {
		t.Fatal(err)
	}

	// Verify that the subscription was successfully added.
	account, err := client.Azure().Subscription(
		ctx,
		azure.CloudAccountID(id),
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
	err = client.Azure().UpdateSubscription(ctx, azure.ID(subscription), core.FeatureCloudNativeArchivalEncryption,
		azure.Regions("westus2"))
	if err != nil {
		t.Error(err)
	}
	account, err = client.Azure().Subscription(ctx, azure.ID(subscription), core.FeatureCloudNativeArchivalEncryption)
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

	// Remove the Azure subscription from Polaris keeping the snapshots.
	// Removing archival feature as encryption is a child feature of it.
	err = client.Azure().RemoveSubscription(ctx, azure.ID(subscription), core.FeatureCloudNativeArchival, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	_, err = client.Azure().Subscription(ctx, azure.ID(subscription), core.FeatureCloudNativeArchivalEncryption)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAzureExocompute verifies that the SDK can perform basic Exocompute
// operations on a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * TEST_INTEGRATION=1
//   * TEST_AZURESUBSCRIPTION_FILE=<path-to-test-azure-subscription-file>
//   * AZURE_AUTH_LOCATION=<path-to-azure-sdk-auth-file>
//
// The file referred to by TEST_AZURESUBSCRIPTION_FILE should contain a single
// testAzureSubscription JSON object.
func TestAzureExocompute(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	ctx := context.Background()

	testSubscription, err := testsetup.AzureSubscription()
	if err != nil {
		t.Fatal(err)
	}

	// Add default Azure service principal to Polaris. Usually resolved using
	// the environment variable AZURE_SERVICEPRINCIPAL_LOCATION.
	_, err = client.Azure().SetServicePrincipal(ctx, azure.Default(testSubscription.TenantDomain))
	if err != nil {
		t.Fatal(err)
	}

	// Add default Azure subscription to Polaris.
	subscription := azure.Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	accountID, err := client.Azure().AddSubscription(ctx, subscription, core.FeatureCloudNativeProtection,
		azure.Regions("eastus2"), azure.Name(testSubscription.SubscriptionName))
	if err != nil {
		t.Fatal(err)
	}

	// Enable the exocompute feature for the account.
	exoAccountID, err := client.Azure().AddSubscription(ctx, subscription, core.FeatureExocompute, azure.Regions("eastus2"))
	if err != nil {
		t.Fatal(err)
	}
	if accountID != exoAccountID {
		t.Fatal("cloud native protection and exocompute added to different cloud accounts")
	}

	account, err := client.Azure().Subscription(ctx, azure.CloudAccountID(accountID), core.FeatureExocompute)
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
		if name := account.Features[0].Name; name != "EXOCOMPUTE" {
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

	// Add exocompute config for the account.
	exoID, err := client.Azure().AddExocomputeConfig(ctx, azure.CloudAccountID(accountID),
		azure.Managed("eastus2", testSubscription.Exocompute.SubnetID))
	if err != nil {
		t.Error(err)
	}

	// Retrieve the exocompute config added.
	exoConfig, err := client.Azure().ExocomputeConfig(ctx, exoID)
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
	if !exoConfig.PolarisManaged {
		t.Errorf("invalid polaris managed state: %t", exoConfig.PolarisManaged)
	}

	// Remove the exocompute config.
	err = client.Azure().RemoveExocomputeConfig(ctx, exoID)
	if err != nil {
		t.Error(err)
	}

	// Verify that the exocompute config was successfully removed.
	exoConfig, err = client.Azure().ExocomputeConfig(ctx, exoID)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Error(err)
	}

	// Remove exocompute feature.
	err = client.Azure().RemoveSubscription(ctx, azure.CloudAccountID(accountID), core.FeatureExocompute, false)
	if err != nil {
		t.Error(err)
	}

	// Verify that the feature was successfully removed.
	_, err = client.Azure().Subscription(ctx, azure.CloudAccountID(accountID), core.FeatureExocompute)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Error(err)
	}

	// Remove subscription.
	err = client.Azure().RemoveSubscription(ctx, azure.CloudAccountID(accountID), core.FeatureCloudNativeProtection, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	_, err = client.Azure().Subscription(ctx, azure.ID(subscription), core.FeatureCloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAzurePermissions verifies that the SDK can read the required Azure
// permissions from a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * TEST_INTEGRATION=1
func TestAzurePermissions(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	ctx := context.Background()

	perms, err := client.Azure().Permissions(ctx, []core.Feature{core.FeatureCloudNativeProtection})
	if err != nil {
		t.Fatal(err)
	}

	// Note that we don't verify the exact permissions returned since they will
	// change over time.
	if len(perms.Actions) == 0 {
		t.Fatal("invalid number of actions: 0")
	}

	if len(perms.DataActions) == 0 {
		t.Fatal("invalid number of data actions: 0")
	}
}

// TestGcpProjectAddAndRemove verifies that the SDK can perform the basic GCP
// project operations on a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * TEST_INTEGRATION=1
//   * TEST_GCPPROJECT_FILE=<path-to-test-gcp-project-file>
//   * GOOGLE_APPLICATION_CREDENTIALS=<path-to-gcp-service-account-key-file>
//
// The file referred to by TEST_GCPPROJECT_FILE should contain a single
// testGcpProject JSON object.
func TestGcpProjectAddAndRemove(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	ctx := context.Background()

	testProject, err := testsetup.GCPProject()
	if err != nil {
		t.Fatal(err)
	}

	// Add the default GCP project to Polaris. Usually resolved using the
	// environment variable GOOGLE_APPLICATION_CREDENTIALS.
	id, err := client.GCP().AddProject(ctx, gcp.Default(), core.FeatureCloudNativeProtection)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully added. ProjectID is compared
	// in a case-insensitive fashion due to a bug causing the initial project
	// id to be the same as the name.
	account, err := client.GCP().Project(ctx, gcp.CloudAccountID(id), core.FeatureCloudNativeProtection)
	if err != nil {
		t.Error(err)
	}
	if account.Name != testProject.ProjectName {
		t.Errorf("invalid name: %v", account.Name)
	}
	if strings.ToLower(account.NativeID) != testProject.ProjectID {
		t.Errorf("invalid project id: %v", account.NativeID)
	}
	if account.ProjectNumber != testProject.ProjectNumber {
		t.Errorf("invalid project number: %v", account.ProjectNumber)
	}
	if n := len(account.Features); n == 1 {
		if name := account.Features[0].Name; name != "CLOUD_NATIVE_PROTECTION" {
			t.Errorf("invalid feature name: %v", name)
		}
		if status := account.Features[0].Status; status != "CONNECTED" {
			t.Errorf("invalid feature status: %v", status)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	// Remove GCP project from Polaris keeping the snapshots.
	err = client.GCP().RemoveProject(ctx, gcp.ID(gcp.Default()), core.FeatureCloudNativeProtection, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully removed.
	_, err = client.GCP().Project(ctx, gcp.ID(gcp.Default()), core.FeatureCloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestGcpProjectAddAndRemoveWithServiceAccountSet verifies that the SDK can
// perform the basic GCP project operations on a real Polaris instance using a
// Polaris account global GCP service account.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * TEST_INTEGRATION=1
//   * TEST_GCPPROJECT_FILE=<path-to-test-gcp-project-file>
//   * GOOGLE_APPLICATION_CREDENTIALS=<path-to-gcp-service-account-key-file>
//
// The file referred to by TEST_GCPPROJECT_FILE should contain a single
// testGcpProject JSON object.
func TestGcpProjectAddAndRemoveWithServiceAccountSet(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	ctx := context.Background()

	testProject, err := testsetup.GCPProject()
	if err != nil {
		t.Fatal(err)
	}

	// Add the service account to Polaris.
	err = client.GCP().SetServiceAccount(ctx, gcp.Default())
	if err != nil {
		t.Fatal(err)
	}

	// Add the default GCP project to Polaris. Usually resolved using the
	// environment variable GOOGLE_APPLICATION_CREDENTIALS.
	id, err := client.GCP().AddProject(ctx, gcp.Project(testProject.ProjectID, testProject.ProjectNumber),
		core.FeatureCloudNativeProtection, gcp.Name(testProject.ProjectName), gcp.Organization(testProject.OrganizationName))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully added.
	account, err := client.GCP().Project(ctx, gcp.CloudAccountID(id), core.FeatureCloudNativeProtection)
	if err != nil {
		t.Error(err)
	}
	if account.Name != testProject.ProjectName {
		t.Errorf("invalid name: %v", account.Name)
	}
	if strings.ToLower(account.NativeID) != testProject.ProjectID {
		t.Errorf("invalid project id: %v", account.NativeID)
	}
	if account.ProjectNumber != testProject.ProjectNumber {
		t.Errorf("invalid project number: %v", account.ProjectNumber)
	}
	if n := len(account.Features); n == 1 {
		if name := account.Features[0].Name; name != "CLOUD_NATIVE_PROTECTION" {
			t.Errorf("invalid feature name: %v", name)
		}
		if status := account.Features[0].Status; status != "CONNECTED" {
			t.Errorf("invalid feature status: %v", status)
		}
	} else {
		t.Errorf("invalid number of features: %v", n)
	}

	// Remove GCP project from Polaris keeping the snapshots.
	err = client.GCP().RemoveProject(ctx, gcp.ProjectNumber(testProject.ProjectNumber), core.FeatureCloudNativeProtection, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully removed.
	_, err = client.GCP().Project(ctx, gcp.ProjectNumber(testProject.ProjectNumber), core.FeatureCloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestGcpPermissions verifies that the SDK can read the required GCP
// permissions from a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * TEST_INTEGRATION=1
func TestGcpPermissions(t *testing.T) {
	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	ctx := context.Background()

	perms, err := client.GCP().Permissions(ctx, []core.Feature{core.FeatureCloudNativeProtection})
	if err != nil {
		t.Fatal(err)
	}

	// Note that we don't verify the exact permissions returned since they will
	// change over time.
	if len(perms) == 0 {
		t.Fatal("invalid number of permissions: 0")
	}
}

// TestListSLA verifies that the SDK can list the default SLAs - Gold, Silver
// and Bronze.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * TEST_INTEGRATION=1
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
func TestListSLA(t *testing.T) {
	//if !boolEnvSet("TEST_INTEGRATION") {
	//	t.Skipf("skipping due to env TEST_INTEGRATION not set")
	//}

	ctx := context.Background()

	// Load configuration and create client. Usually resolved using the
	// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
	polAccount, err := DefaultServiceAccount(true)
	if err != nil {
		t.Fatal(err)
	}

	client, err := NewClient(ctx, polAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	slaDomains, err := client.Core().ListSLA(
		ctx,
		core.SortByName,
		core.ASC,
		[]core.GlobalSLAFilterInput{
			{
				Field: core.ObjectType,
				Text:  "",
				ObjectTypeList: []core.SLAObjectType{
					core.KuprObjectType,
				},
			},
		},
		core.Default,
		[]core.ContextFilterInputField{{"", ""}},
		false,
		false,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}

	basicCheck := make(map[string]bool)
	for _, slaDomain := range slaDomains {
		if slaDomain.Name == "Gold" || slaDomain.Name == "Silver" || slaDomain.Name == "Bronze" {
			basicCheck[slaDomain.Name] = true
		}
	}
	if !basicCheck["Gold"] || !basicCheck["Silver"] || !basicCheck["Bronze"] {
		t.Errorf("failed to list default SLAs: %v", basicCheck)
	}
}

// TestListSLA verifies that the SDK can list the default SLAs - Gold, Silver
// and Bronze.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * TEST_INTEGRATION=1
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
func TestK8sListSLA(t *testing.T) {
	//if !boolEnvSet("TEST_INTEGRATION") {
	//	t.Skipf("skipping due to env TEST_INTEGRATION not set")
	//}

	ctx := context.Background()
	testServiceAccount := ServiceAccount{
		ClientID:       "client|2YDsG5FIYWQ7OW8AiRqA2tPwQEwhHpfU",
		ClientSecret:   "ZLeoEHA6dvDGciAO1bFIIWJmdjlLLawFa6WC8IdpBfgphecsUQYB4pAOKovQbt4e",
		Name:           "kupatest",
		AccessTokenURI: "https://demo.dev-017.my.rubrik-lab.com/api/client_token",
	}
	// Load configuration and create client. Usually resolved using the
	// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
	//polAccount, err := DefaultServiceAccount(true)
	//if err != nil {
	//	t.Fatal(err)
	//}

	client, err := NewClient(ctx, &testServiceAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	slaDomains, err := client.K8s().ListSLA(ctx)
	if err != nil {
		t.Fatal(err)
	}

	basicCheck := make(map[string]bool)
	for _, slaDomain := range slaDomains {
		if slaDomain.Name == "Gold" || slaDomain.Name == "Silver" || slaDomain.Name == "Bronze" {
			basicCheck[slaDomain.Name] = true
		}
	}
	if !basicCheck["Gold"] || !basicCheck["Silver"] || !basicCheck["Bronze"] {
		t.Errorf("failed to list default SLAs: %v", basicCheck)
	}
}

// TestListSLA verifies that the SDK can list the default SLAs - Gold, Silver
// and Bronze.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * TEST_INTEGRATION=1
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
func TestListK8sNamespace(t *testing.T) {
	//if !boolEnvSet("TEST_INTEGRATION") {
	//	t.Skipf("skipping due to env TEST_INTEGRATION not set")
	//}

	testClusterID := uuid.MustParse("a840e205-ac36-408d-9a12-7690769aaa88")
	ctx := context.Background()
	testServiceAccount := ServiceAccount{
		ClientID:       "client|sIIw3uAxHqFsn3kUR78AUf1zMewyLB7p",
		ClientSecret:   "WnmUX2luK5X_TcrMMzZUrFh-mU7gWWti0VS90onJ_uwygXsYajUwVOlWE1MArIs_",
		Name:           "test",
		AccessTokenURI: "https://manifest.dev-045.my.rubrik-lab.com/api/client_token",
	}
	// Load configuration and create client. Usually resolved using the
	// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
	//polAccount, err := DefaultServiceAccount(true)
	//if err != nil {
	//	t.Fatal(err)
	//}

	client, err := NewClient(ctx, &testServiceAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}
	nss, err := client.K8s().ListK8sNamespace(ctx, testClusterID)
	if err != nil {
		t.Fatal(err)
	}
	for _, ns := range nss {
		fmt.Printf("%v\n", ns)
	}

}

func TestUnassignSLA(t *testing.T) {
	//if !boolEnvSet("TEST_INTEGRATION") {
	//	t.Skipf("skipping due to env TEST_INTEGRATION not set")
	//}

	testNSID := uuid.MustParse("3b3d22b7-385c-5865-bd8f-0ff7534db42b")
	ctx := context.Background()
	testServiceAccount := ServiceAccount{
		ClientID:       "client|YVaYtseZQXRSiwEHV8HaRKfiRfsU0BhX",
		ClientSecret:   "BS0-2olXRhY51R1AW0qj9gdgYBOs6x4uJbUg-7DXIxQImigEiavo819R0ZTPwq8a",
		Name:           "np-test",
		AccessTokenURI: "https://demo.dev-017.my.rubrik-lab.com/api/client_token",
	}
	client, err := NewClient(ctx, &testServiceAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}
	oks, err := client.Core().AssignSlaForSnappableHierarchies(
		ctx,
		nil,
		core.NoAssignment,
		[]uuid.UUID{testNSID},
		[]core.SnappableLevelHierarchyType{},
		true,  // shouldApplyToExistingSnapshots
		false, // shouldApplyToNonPolicySnapshots
		core.RetainSnapshots,
	)
	// Load configuration and create client. Usually resolved using the
	// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
	//polAccount, err := DefaultServiceAccount(true)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v\n", oks)

}

// TestTakeK8NamespaceSnapshot verifies that the SDK can take a
// namespace snapshot
func TestTakeK8NamespaceSnapshot(t *testing.T) {
	testClusterID := uuid.MustParse("a840e205-ac36-408d-9a12-7690769aaa88")
	ctx := context.Background()
	testServiceAccount := ServiceAccount{
		ClientID:       "client|2YDsG5FIYWQ7OW8AiRqA2tPwQEwhHpfU",
		ClientSecret:   "ZLeoEHA6dvDGciAO1bFIIWJmdjlLLawFa6WC8IdpBfgphecsUQYB4pAOKovQbt4e",
		Name:           "kupatest",
		AccessTokenURI: "https://demo.dev-017.my.rubrik-lab.com/api/client_token",
	}

	client, err := NewClient(ctx, &testServiceAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	nss, err := client.K8s().ListK8sNamespace(ctx, testClusterID)
	if err != nil {
		t.Fatal(err)
	}

	if len(nss) == 0 {
		t.Fatal("No namespace found")
	}

	ns := nss[0] // get first namespace
	info, err := client.K8s().TakeK8NamespaceSnapshot(ctx, ns.ID, "00000000-0000-0000-0000-000000000000")
	if err!= nil {
		t.Fatal(err)
	}
	fmt.Printf("%v\n", info)
}

// TestGetTaskchainInfo verifies that the SDK can fetch taskchain states
func TestGetTaskchainInfo(t *testing.T) {
	testClusterID := uuid.MustParse("a840e205-ac36-408d-9a12-7690769aaa88")
	ctx := context.Background()
	testServiceAccount := ServiceAccount{
		ClientID:       "client|2YDsG5FIYWQ7OW8AiRqA2tPwQEwhHpfU",
		ClientSecret:   "ZLeoEHA6dvDGciAO1bFIIWJmdjlLLawFa6WC8IdpBfgphecsUQYB4pAOKovQbt4e",
		Name:           "kupatest",
		AccessTokenURI: "https://demo.dev-017.my.rubrik-lab.com/api/client_token",
	}

	client, err := NewClient(ctx, &testServiceAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	nss, err := client.K8s().ListK8sNamespace(ctx, testClusterID)
	if err != nil {
		t.Fatal(err)
	}

	if len(nss) == 0 {
		t.Fatal("No namespace found")
	}

	ns := nss[1] // get first namespace
	info, err := client.K8s().TakeK8NamespaceSnapshot(ctx, ns.ID, "00000000-0000-0000-0000-000000000000")
	if err!= nil {
		t.Fatal(err)
	}

	for i := 0; i < 6; i++ {
		tk, err := client.K8s().GetTaskchainInfo(ctx, info.TaskchainId, "snapshot-namespace")
		if err == nil {
			fmt.Printf("%v\n", tk)
			return
		}
		if !strings.Contains(err.Error(), "Taskchain does not exist") ||
			!strings.Contains(err.Error(), "State key not found in response") ||
			!strings.Contains(err.Error(), "Taskchain still in wait state") {
			fmt.Printf("%v : %v\n", info, err)
			time.Sleep(15 * time.Second)
		} else {
			t.Fatal(err.Error())
		}
	}

	t.Fatal("Task chain does not start")
}

// TestGetActivitySeriesConnection verifies that the SDK can fetch events
func TestGetActivitySeriesConnection(t *testing.T) {
	ctx := context.Background()
	testServiceAccount := ServiceAccount{
		ClientID:       "client|gYtGLLyhcp3rI02kkbIdZt7CMFzIhg54",
		ClientSecret:   "JQfCRwuiiwzB_7Ibt1UIOcAT0wQTYnRNF2ikPD8aTaTOqbiYhfm8v7Lb6pY6NBis",
		Name:           "manifest-unit-test",
		AccessTokenURI: "https://manifest.dev-045.my.rubrik-lab.com/api/client_token",
	}

	client, err := NewClient(ctx, &testServiceAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}
	ts := time.Now().Add(time.Duration(-2) * time.Hour).UTC()
	as, err := client.K8s().GetActivitySeriesConnection(ctx, []string{"KuprNamespace"}, ts)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%v\n", len(as))
	fmt.Printf("%v\n", as)
}

// TestListNamespace verifies that the SDK can fetch snapshots
func TestListNamespace(t *testing.T) {
	ctx := context.Background()
	testServiceAccount := ServiceAccount{
		ClientID:       "client|sIIw3uAxHqFsn3kUR78AUf1zMewyLB7p",
		ClientSecret:   "WnmUX2luK5X_TcrMMzZUrFh-mU7gWWti0VS90onJ_uwygXsYajUwVOlWE1MArIs_",
		Name:           "test",
		AccessTokenURI: "https://manifest.dev-045.my.rubrik-lab.com/api/client_token",
	}

	client, err := NewClient(ctx, &testServiceAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	st := time.Now().Add(time.Duration(-2) * time.Hour).UTC()
	et := time.Now().UTC()

	sId, err := uuid.Parse("922a32e7-674a-56b7-9750-97ae683cd76f")
	if err != nil {
		t.Fatal(err)
	}

	cursor := ""
	for {
		s, cur, err := client.K8s().GetK8sNamespace(ctx, sId, st, et, cursor)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%v\n", len(s))
		fmt.Printf("%v\n", s)
		if cur == "" {
			fmt.Println("Completed")
			return
		}
		cursor = cur
	}
}

// TestGetAllSnapshotPvcs verifies that the SDK can fetch all PVC's for a
// snapshot
func TestGetAllSnapshotPvcs(t *testing.T) {
	ctx := context.Background()
	testServiceAccount := ServiceAccount{
		ClientID:       "client|sIIw3uAxHqFsn3kUR78AUf1zMewyLB7p",
		ClientSecret:   "WnmUX2luK5X_TcrMMzZUrFh-mU7gWWti0VS90onJ_uwygXsYajUwVOlWE1MArIs_",
		Name:           "test",
		AccessTokenURI: "https://manifest.dev-045.my.rubrik-lab.com/api/client_token",
	}

	client, err := NewClient(ctx, &testServiceAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	snappableId, err := uuid.Parse("0a9758bf-2dca-5b93-9482-6f9a2940012e")
	if err != nil {
		t.Fatal(err)
	}

	s, err := client.K8s().GetAllSnapshotPVCS(ctx, snappableId, "550a8411-0f2d-4d71-acbc-1560970ab001")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%v\n", len(s))
	fmt.Printf("%v\n", s)
}
// TestRestoreK8NamespaceSnapshot verifies that the SDK can restore a
// namespace snapshot
func TestRestoreK8NamespaceSnapshot(t *testing.T) {
	testClusterID := uuid.MustParse("a840e205-ac36-408d-9a12-7690769aaa88")
	ctx := context.Background()
	testServiceAccount := ServiceAccount{
		ClientID:       "client|2YDsG5FIYWQ7OW8AiRqA2tPwQEwhHpfU",
		ClientSecret:   "ZLeoEHA6dvDGciAO1bFIIWJmdjlLLawFa6WC8IdpBfgphecsUQYB4pAOKovQbt4e",
		Name:           "kupatest",
		AccessTokenURI: "https://demo.dev-017.my.rubrik-lab.com/api/client_token",
	}

	client, err := NewClient(ctx, &testServiceAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	nss, err := client.K8s().ListK8sNamespace(ctx, testClusterID)
	if err != nil {
		t.Fatal(err)
	}

	if len(nss) == 0 {
		t.Fatal("No namespace found")
	}

	ns := nss[13] // get first namespace
	info, err := client.K8s().TakeK8NamespaceSnapshot(ctx, ns.ID, "00000000-0000-0000-0000-000000000000")
	if err!= nil {
		t.Fatal(err)
	}

	for i := 0; i < 50; i++ {
		tk, err := client.K8s().GetTaskchainInfo(ctx, info.TaskchainId, "snapshot-namespace")
		if err == nil {
			fmt.Printf("%v\n", tk)
			return
		}
		time.Sleep(30 * time.Second)
	}

	t.Fatal("Task chain does not start")
}

// TestGetActivitySeries verifies that the SDK can fetch events
func TestGetActivitySeries(t *testing.T) {
	ctx := context.Background()
	testServiceAccount := ServiceAccount{
		ClientID:       "client|sIIw3uAxHqFsn3kUR78AUf1zMewyLB7p",
		ClientSecret:   "WnmUX2luK5X_TcrMMzZUrFh-mU7gWWti0VS90onJ_uwygXsYajUwVOlWE1MArIs_",
		Name:           "test",
		AccessTokenURI: "https://manifest.dev-045.my.rubrik-lab.com/api/client_token",
	}

	client, err := NewClient(ctx, &testServiceAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	asid, err := uuid.Parse("82337567-3b65-4114-8b7b-f2df470a9fee")
	if err != nil {
		t.Fatal(err)
	}

	cid, err := uuid.Parse("00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatal(err)
	}

	as, cursor, err := client.K8s().GetActivitySeries(ctx, asid, cid, "")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%v\n", len(as))
	fmt.Printf("%v\n", as)
	fmt.Printf("%v\n", cursor)
}