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
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	polaris_log "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// testAwsAccount hold AWS account information used in the integration tests.
// Normally used to assert that the information read from Polaris is correct.
type testAwsAccount struct {
	Name      string `json:"name"`
	AccountID string `json:"accountId"`
}

// TestAwsAccountAddAndRemove verifies that the SDK can perform the basic AWS
// account operations on a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * SDK_INTEGRATION=1
//   * SDK_AWSACCOUNT_FILE=<path-to-test-aws-account-file>
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * AWS_ACCESS_KEY_ID=<aws-access-key>
//   * AWS_SECRET_ACCESS_KEY=<aws-secret-key>
//   * AWS_DEFAULT_REGION=<aws-default-region>
//
// The file referred to by SDK_AWSACCOUNT_FILE should contain a single
// testAwsAccount JSON object.
//
// Note that between the project has been added and it has been removed we
// never fail fatally to allow the project to be removed in case of an error.
func TestAwsAccountAddAndRemove(t *testing.T) {
	requireEnv(t, "SDK_INTEGRATION")

	ctx := context.Background()

	// Load test account information from the file pointed to by the
	// SDK_AWSACCOUNT_FILE environment variable.
	buf, err := os.ReadFile(os.Getenv("SDK_AWSACCOUNT_FILE"))
	if err != nil {
		t.Fatalf("failed to read file pointed to by SDK_AWSACCOUNT_FILE: %v", err)
	}
	testAccount := testAwsAccount{}
	if err := json.Unmarshal(buf, &testAccount); err != nil {
		t.Fatal(err)
	}

	// Load configuration and create client. Usually resolved using the
	// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
	polAccount, err := DefaultServiceAccount()
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClientFromServiceAccount(polAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	// Add the default AWS account to Polaris. Usually resolved using the
	// environment variables AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and
	// AWS_DEFAULT_REGION.
	id, err := client.AWS().AddAccount(ctx, aws.Default(), aws.Name(testAccount.Name),
		aws.Regions("us-east-2"))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully added.
	account, err := client.AWS().Account(ctx, aws.CloudAccountID(id), core.CloudNativeProtection)
	if err != nil {
		t.Error(err)
	}
	if n := len(account.Features); n != 1 {
		t.Errorf("invalid number of features: %v", n)
	}
	if account.Name != testAccount.Name {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.NativeID != testAccount.AccountID {
		t.Errorf("invalid native id: %v", account.NativeID)
	}
	if account.Features[0].Name != "CLOUD_NATIVE_PROTECTION" {
		t.Errorf("invalid feature name: %v", account.Features[0].Name)
	}
	if regions := account.Features[0].Regions; !reflect.DeepEqual(regions, []string{"us-east-2"}) {
		t.Errorf("invalid feature regions: %v", regions)
	}
	if account.Features[0].Status != "CONNECTED" {
		t.Errorf("invalid feature status: %v", account.Features[0].Status)
	}

	// Update and verify regions for AWS account.
	err = client.AWS().UpdateAccount(ctx, aws.ID(aws.Default()), core.CloudNativeProtection,
		aws.Regions("us-west-2"))
	if err != nil {
		t.Error(err)
	}
	account, err = client.AWS().Account(ctx, aws.ID(aws.Default()), core.CloudNativeProtection)
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
	err = client.AWS().RemoveAccount(ctx, aws.Default(), false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	account, err = client.AWS().Account(ctx, aws.ID(aws.Default()), core.CloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Error(err)
	}
}

// testAzureSubscription hold Azure subscription information used in the
// integration tests. Normally used to assert that the information read from
// Polaris is correct.
type testAzureSubscription struct {
	Name           string    `json:"name"`
	SubscriptionID uuid.UUID `json:"subscriptionId"`
	TenantDomain   string    `json:"tenantDomain"`
}

// TestAzureSubscriptionAddAndRemove verifies that the SDK can perform the
// basic AWS account operations on a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * SDK_INTEGRATION=1
//   * SDK_AZUREACCOUNT_FILE=<path-to-test-azure-subscription-file>
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * AZURE_SERVICEPRINCIPAL_LOCATION=<path-to-azure-service-principal-file>
//
// The file referred to by SDK_AWSACCOUNT_FILE should contain a single
// testAwsAccount JSON object.
//
// Between the account has been added and it has been removed we never fail
// fatally to allow the account to be removed in case of an error.
func TestAzureSubscriptionAddAndRemove(t *testing.T) {
	requireEnv(t, "SDK_INTEGRATION")

	ctx := context.Background()

	// Load test project information from the file pointed to by the
	// SDK_AZURESUBSCRIPTION_FILE environment variable.
	buf, err := os.ReadFile(os.Getenv("SDK_AZURESUBSCRIPTION_FILE"))
	if err != nil {
		t.Fatalf("failed to read file pointed to by SDK_AZURESUBSCRIPTION_FILE: %v", err)
	}
	testSubscription := testAzureSubscription{}
	if err := json.Unmarshal(buf, &testSubscription); err != nil {
		t.Fatal(err)
	}

	// Load configuration and create client. Usually resolved using the
	// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
	polAccount, err := DefaultServiceAccount()
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClientFromServiceAccount(polAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	// Add default Azure service principal to Polaris. Usually resolved using
	// the environment variable AZURE_SERVICEPRINCIPAL_LOCATION.
	_, err = client.Azure().SetServicePrincipal(ctx, azure.Default())
	if err != nil {
		t.Fatal(err)
	}

	// Add default Azure subscription to Polaris.
	subscription := azure.Subscription(testSubscription.SubscriptionID, testSubscription.TenantDomain)
	id, err := client.Azure().AddSubscription(ctx, subscription, azure.Regions("eastus2"),
		azure.Name(testSubscription.Name))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the subscription was successfully added.
	account, err := client.Azure().Subscription(ctx, azure.CloudAccountID(id), core.CloudNativeProtection)
	if err != nil {
		t.Error(err)
	}
	if account.Name != testSubscription.Name {
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
	err = client.Azure().UpdateSubscription(ctx, azure.ID(subscription), core.CloudNativeProtection,
		azure.Regions("westus2"))
	if err != nil {
		t.Error(err)
	}
	account, err = client.Azure().Subscription(ctx, azure.ID(subscription), core.CloudNativeProtection)
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
	err = client.Azure().RemoveSubscription(ctx, azure.ID(subscription), false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	_, err = client.Azure().Subscription(ctx, azure.ID(subscription), core.CloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Error(err)
	}
}

// testGcpProject hold GCP project information used in the integration tests.
// Normally used to assert that the information read from Polaris is correct.
type testGcpProject struct {
	Name             string `json:"name"`
	ProjectID        string `json:"projectId"`
	ProjectNumber    int64  `json:"projectNumber"`
	OrganizationName string `json:"organizationName"`
}

// TestGcpProjectAddAndRemove verifies that the SDK can perform the basic GCP
// project operations on a real Polaris instance.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * SDK_INTEGRATION=1
//   * SDK_GCPPROJECT_FILE=<path-to-test-gcp-project-file>
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * GOOGLE_APPLICATION_CREDENTIALS=<path-to-gcp-service-account-key-file>
//
// The file referred to by SDK_GCPPROJECT_FILE should contain a single
// testGcpProject JSON object.
//
// Note that between the project has been added and it has been removed we
// never fail fatally to allow the project to be removed in case of an error.
func TestGcpProjectAddAndRemove(t *testing.T) {
	requireEnv(t, "SDK_INTEGRATION")

	ctx := context.Background()

	// Load test project information from the file pointed to by the
	// SDK_GCPPROJECT_FILE environment variable.
	buf, err := os.ReadFile(os.Getenv("SDK_GCPPROJECT_FILE"))
	if err != nil {
		t.Fatalf("failed to read file pointed to by SDK_GCPPROJECT_FILE: %v", err)
	}
	testProject := testGcpProject{}
	if err := json.Unmarshal(buf, &testProject); err != nil {
		t.Fatal(err)
	}

	// Load configuration and create client. Usually resolved using the
	// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
	polAccount, err := DefaultServiceAccount()
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClientFromServiceAccount(polAccount, &polaris_log.DiscardLogger{})
	if err != nil {
		t.Fatal(err)
	}

	// Add the default GCP project to Polaris. Usually resolved using the
	// environment variable GOOGLE_APPLICATION_CREDENTIALS.
	id, err := client.GCP().AddProject(ctx, gcp.Default())
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully added. ProjectID is compared
	// in a case-insensitive fashion due to a bug causing the initial project
	// id to be the same as the name.
	account, err := client.GCP().Project(ctx, gcp.CloudAccountID(id), core.CloudNativeProtection)
	if err != nil {
		t.Error(err)
	}
	if account.Name != testProject.Name {
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
	if err := client.GCP().RemoveProject(ctx, gcp.ID(gcp.Default()), false); err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully removed.
	_, err = client.gcp.Project(ctx, gcp.ID(gcp.Default()), core.CloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Error(err)
	}
}

// TestGcpProjectAddAndRemoveWithServiceAccountSet verifies that the SDK can
// perform the basic GCP project operations on a real Polaris instance using a
// Polaris account global GCP service account.
//
// To run this test against a Polaris instance the following environment
// variables needs to be set:
//   * SDK_INTEGRATION=1
//   * SDK_GCPPROJECT_FILE=<path-to-test-gcp-project-file>
//   * RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   * GOOGLE_APPLICATION_CREDENTIALS=<path-to-gcp-service-account-key-file>
//
// The file referred to by SDK_GCPPROJECT_FILE should contain a single
// testGcpProject JSON object.
//
// Note that between the project has been added and it has been removed we
// never fail fatally to allow the project to be removed in case of an error.
func TestGcpProjectAddAndRemoveWithServiceAccountSet(t *testing.T) {
	requireEnv(t, "SDK_INTEGRATION")

	ctx := context.Background()

	// Load test project information from the file pointed to by the
	// SDK_GCPPROJECT_FILE environment variable.
	buf, err := os.ReadFile(os.Getenv("SDK_GCPPROJECT_FILE"))
	if err != nil {
		t.Fatalf("failed to read file pointed to by SDK_GCPPROJECT_FILE: %v", err)
	}

	testProject := testGcpProject{}
	if err := json.Unmarshal(buf, &testProject); err != nil {
		t.Fatal(err)
	}

	// Load configuration and create client. Usually resolved using the
	// environment variable RUBRIK_POLARIS_SERVICEACCOUNT_FILE.
	polAccount, err := DefaultServiceAccount()
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClientFromServiceAccount(polAccount, &polaris_log.DiscardLogger{})
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
	id, err := client.GCP().AddProject(ctx, gcp.Project(testProject.ProjectID, testProject.Name, testProject.ProjectNumber,
		testProject.OrganizationName))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully added.
	account, err := client.GCP().Project(ctx, gcp.CloudAccountID(id), core.CloudNativeProtection)
	if err != nil {
		t.Error(err)
	}
	if account.Name != testProject.Name {
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
	err = client.GCP().RemoveProject(ctx, gcp.ProjectNumber(testProject.ProjectNumber), false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the project was successfully removed.
	_, err = client.GCP().Project(ctx, gcp.ProjectNumber(testProject.ProjectNumber), core.CloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Error(err)
	}
}
