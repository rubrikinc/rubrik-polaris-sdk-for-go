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
	"strconv"
	"strings"
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	polaris_log "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// client is the common Polaris client used for tests. By reusing the same client we
// reduce the risk of hitting rate limits when access tokens are creted.
var client *Client

func TestMain(m *testing.M) {
	if boolEnvSet("TEST_INTEGRATION") {
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

// boolEnvSet returns true if the provided argument is defined and true
// according to the definition given by strconv.ParseBool.
func boolEnvSet(env string) bool {
	val := os.Getenv(env)

	b, err := strconv.ParseBool(val)
	return err == nil && b
}

// TestAwsAccountAddAndRemove verifies that the SDK can perform the basic AWS
// account operations on a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * TEST_INTEGRATION=1
//   * TEST_AWSACCOUNT_FILE=<path-to-test-aws-account-file>
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//
// In addition to the above environment variables a default AWS profile must be
// defined. As an alternative to the credentials and config files the
// environment variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and
// AWS_DEFAULT_REGION can be used. The file referred to by TEST_AWSACCOUNT_FILE
// should contain a single testAwsAccount JSON object.
func TestAwsAccountAddAndRemove(t *testing.T) {
	if !boolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	ctx := context.Background()

	testAccount, err := testsetup.AWSAccount()
	if err != nil {
		t.Fatal(err)
	}

	// Add the default AWS account to Polaris. Usually resolved using the
	// environment variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and
	// AWS_DEFAULT_REGION.
	id, err := client.AWS().AddAccount(ctx, aws.Default(), core.FeatureCloudNativeProtection,
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
	err = client.AWS().UpdateAccount(ctx, aws.ID(aws.Default()), core.FeatureCloudNativeProtection,
		aws.Regions("us-west-2"))
	if err != nil {
		t.Error(err)
	}
	account, err = client.AWS().Account(ctx, aws.ID(aws.Default()), core.FeatureCloudNativeProtection)
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
	err = client.AWS().RemoveAccount(ctx, aws.Default(), core.FeatureCloudNativeProtection, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	account, err = client.AWS().Account(ctx, aws.ID(aws.Default()), core.FeatureCloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAwsExocompute verifies that the SDK can perform basic Exocompute
// operations on a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * TEST_INTEGRATION=1
//   * TEST_AWSACCOUNT_FILE=<path-to-test-aws-account-file>
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//
// In addition to the above environment variables a default AWS profile must be
// defined. As an alternative to the credentials and config files the
// environment variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and
// AWS_DEFAULT_REGION can be used. The file referred to by TEST_AWSACCOUNT_FILE
// should contain a single testAwsAccount JSON object.
func TestAwsExocompute(t *testing.T) {
	if !boolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	ctx := context.Background()

	testAccount, err := testsetup.AWSAccount()
	if err != nil {
		t.Fatal(err)
	}

	// Add the default AWS account to Polaris. Usually resolved using the
	// environment variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and
	// AWS_DEFAULT_REGION.
	accountID, err := client.AWS().AddAccount(ctx, aws.Default(), core.FeatureCloudNativeProtection,
		aws.Name(testAccount.AccountName), aws.Regions("us-east-2"))
	if err != nil {
		t.Fatal(err)
	}

	// Enable the exocompute feature for the account.
	exoAccountID, err := client.AWS().AddAccount(ctx, aws.Default(), core.FeatureExocompute, aws.Regions("us-east-2"))
	if err != nil {
		t.Error(err)
	}
	if accountID != exoAccountID {
		t.Error("cloud native protection and exocompute added to different cloud accounts")
	}

	// Verify that the account was successfully added.
	account, err := client.AWS().Account(ctx, aws.CloudAccountID(accountID), core.FeatureExocompute)
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

	exoID, err := client.AWS().AddExocomputeConfig(ctx, aws.ID(aws.Default()),
		aws.Managed("us-east-2", testAccount.Exocompute.VPCID,
			[]string{testAccount.Exocompute.Subnets[0].ID, testAccount.Exocompute.Subnets[1].ID}))
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve the exocompute config added.
	exoConfig, err := client.AWS().ExocomputeConfig(ctx, exoID)
	if err != nil {
		t.Error(err)
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
		t.Error(err)
	}

	// Verify that the exocompute config was successfully removed.
	exoConfig, err = client.AWS().ExocomputeConfig(ctx, exoID)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Error(err)
	}

	// Disable the exocompute feature for the account.
	err = client.AWS().RemoveAccount(ctx, aws.Default(), core.FeatureExocompute, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the exocompute feature was successfully disabled.
	account, err = client.AWS().Account(ctx, aws.ID(aws.Default()), core.FeatureExocompute)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}

	// Remove the AWS account from Polaris.
	err = client.AWS().RemoveAccount(ctx, aws.Default(), core.FeatureCloudNativeProtection, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	account, err = client.AWS().Account(ctx, aws.ID(aws.Default()), core.FeatureCloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAzureSubscriptionAddAndRemove verifies that the SDK can perform the
// basic Azure subscription operations on a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * TEST_INTEGRATION=1
//   * TEST_AZURESUBSCRIPTION_FILE=<path-to-test-azure-subscription-file>
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * AZURE_AUTH_LOCATION=<path-to-azure-sdk-auth-file>
//
// The file referred to by TEST_AZURESUBSCRIPTION_FILE should contain a single
// testAzureSubscription JSON object.
func TestAzureSubscriptionAddAndRemove(t *testing.T) {
	if !boolEnvSet("TEST_INTEGRATION") {
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

// TestAzureExocompute verifies that the SDK can perform basic Exocompute
// operations on a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * TEST_INTEGRATION=1
//   * TEST_AZURESUBSCRIPTION_FILE=<path-to-test-azure-subscription-file>
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * AZURE_AUTH_LOCATION=<path-to-azure-sdk-auth-file>
//
// The file referred to by TEST_AZURESUBSCRIPTION_FILE should contain a single
// testAzureSubscription JSON object.
func TestAzureExocompute(t *testing.T) {
	if !boolEnvSet("TEST_INTEGRATION") {
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
//   * TEST_INTEGRATION=1
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
func TestAzurePermissions(t *testing.T) {
	if !boolEnvSet("TEST_INTEGRATION") {
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
//   * TEST_INTEGRATION=1
//   * TEST_GCPPROJECT_FILE=<path-to-test-gcp-project-file>
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * GOOGLE_APPLICATION_CREDENTIALS=<path-to-gcp-service-account-key-file>
//
// The file referred to by TEST_GCPPROJECT_FILE should contain a single
// testGcpProject JSON object.
func TestGcpProjectAddAndRemove(t *testing.T) {
	if !boolEnvSet("TEST_INTEGRATION") {
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
//   * TEST_INTEGRATION=1
//   * TEST_GCPPROJECT_FILE=<path-to-test-gcp-project-file>
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * GOOGLE_APPLICATION_CREDENTIALS=<path-to-gcp-service-account-key-file>
//
// The file referred to by TEST_GCPPROJECT_FILE should contain a single
// testGcpProject JSON object.
func TestGcpProjectAddAndRemoveWithServiceAccountSet(t *testing.T) {
	if !boolEnvSet("TEST_INTEGRATION") {
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
//   * TEST_INTEGRATION=1
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
func TestGcpPermissions(t *testing.T) {
	if !boolEnvSet("TEST_INTEGRATION") {
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
