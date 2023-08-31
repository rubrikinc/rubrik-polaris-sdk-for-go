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

package aws

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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
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

// TestAwsAccountAddAndRemove verifies that the SDK can perform the basic AWS
// account operations on a real RSC instance.
//
// To run this test against an RSC instance the following environment variables
// needs to be set:
//   - RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   - TEST_INTEGRATION=1
//   - TEST_AWSACCOUNT_FILE=<path-to-test-aws-account-file>
//   - AWS_SHARED_CREDENTIALS_FILE=<path-to-aws-credentials-file>
//   - AWS_CONFIG_FILE=<path-to-aws-config-file>
//
// The file referred to by TEST_AWSACCOUNT_FILE should contain a single
// testAwsAccount JSON object.
func TestAwsAccountAddAndRemove(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	testAccount, err := testsetup.AWSAccount()
	if err != nil {
		t.Fatal(err)
	}

	awsClient := Wrap(client)

	// Adds the AWS account identified by the specified profile to RSC. Note
	// that the profile needs to have a default region.
	id, err := awsClient.AddAccount(ctx, Profile(testAccount.Profile), []core.Feature{core.FeatureCloudNativeProtection},
		Name(testAccount.AccountName), Regions("us-east-2"))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully added.
	account, err := awsClient.Account(ctx, CloudAccountID(id), core.FeatureCloudNativeProtection)
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
	err = awsClient.UpdateAccount(ctx, AccountID(testAccount.AccountID), core.FeatureCloudNativeProtection,
		Regions("us-west-2"))
	if err != nil {
		t.Error(err)
	}
	account, err = awsClient.Account(ctx, AccountID(testAccount.AccountID), core.FeatureCloudNativeProtection)
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

	// Remove AWS account from RSC.
	err = awsClient.RemoveAccount(ctx, Profile(testAccount.Profile), []core.Feature{core.FeatureCloudNativeProtection}, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	account, err = awsClient.Account(ctx, AccountID(testAccount.AccountID), core.FeatureCloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAwsCrossAccountAddAndRemove verifies that the SDK can perform the basic
// AWS cross account operations on a real RSC instance.
//
// To run this test against an RSC instance the following environment variables
// needs to be set:
//   - RUBRIK_POLARIS_SERVICEACCOUNT_FILE=<path-to-polaris-service-account-file>
//   - TEST_INTEGRATION=1
//   - TEST_AWSACCOUNT_FILE=<path-to-test-aws-account-file>
//   - AWS_SHARED_CREDENTIALS_FILE=<path-to-aws-credentials-file>
//   - AWS_CONFIG_FILE=<path-to-aws-config-file>
//
// The file referred to by TEST_AWSACCOUNT_FILE should contain a single
// testAwsAccount JSON object.
func TestAwsCrossAccountAddAndRemove(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	testAccount, err := testsetup.AWSAccount()
	if err != nil {
		t.Fatal(err)
	}

	awsClient := Wrap(client)

	// Use the default profile to add an AWS account to RSC using a cross
	// account role. Note that the profile needs to have a region.
	id, err := awsClient.AddAccount(ctx,
		ProfileWithRole(testAccount.Profile, testAccount.CrossAccountRole), []core.Feature{core.FeatureCloudNativeProtection},
		Name(testAccount.CrossAccountName), Regions("us-east-2"))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully added.
	account, err := awsClient.Account(ctx, CloudAccountID(id), core.FeatureCloudNativeProtection)
	if err != nil {
		t.Error(err)
	}
	if account.Name != testAccount.CrossAccountName {
		t.Errorf("invalid name: %v", account.Name)
	}
	if account.NativeID != testAccount.CrossAccountID {
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

	// Verify that it's possible to search for the account using a role.
	account, err = awsClient.Account(ctx, Role(testAccount.CrossAccountRole), core.FeatureCloudNativeProtection)
	if err != nil {
		t.Error(err)
	}
	if account.ID != id {
		t.Errorf("invalid id: %v", account.ID)
	}

	// Remove AWS account from RSC using a cross account role.
	err = awsClient.RemoveAccount(ctx, ProfileWithRole(testAccount.Profile, testAccount.CrossAccountRole),
		[]core.Feature{core.FeatureCloudNativeProtection}, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	account, err = awsClient.Account(ctx, AccountID(testAccount.AccountID), core.FeatureCloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}
