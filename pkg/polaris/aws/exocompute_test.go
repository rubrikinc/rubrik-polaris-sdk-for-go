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
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// TestAwsExocompute verifies that the SDK can perform basic Exocompute
// operations on a real RSC instance.
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
func TestAwsExocompute(t *testing.T) {
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
	accountID, err := awsClient.AddAccount(ctx, Profile(testAccount.Profile), []core.Feature{core.FeatureCloudNativeProtection},
		Name(testAccount.AccountName), Regions("us-east-2"))
	if err != nil {
		t.Fatal(err)
	}

	// Enable the exocompute feature for the account.
	exoAccountID, err := awsClient.AddAccount(ctx, Profile(testAccount.Profile),
		[]core.Feature{core.FeatureExocompute.WithPermissionGroups(core.PermissionGroupBasic, core.PermissionGroupRSCManagedCluster)},
		Regions("us-east-2"))
	if err != nil {
		t.Fatal(err)
	}
	if accountID != exoAccountID {
		t.Fatalf("cloud native protection and exocompute added to different cloud accounts: %q vs %q",
			accountID, exoAccountID)
	}

	// Verify that the account was successfully added.
	account, err := awsClient.Account(ctx, CloudAccountID(accountID), core.FeatureExocompute)
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != testAccount.AccountName {
		t.Fatalf("invalid name: %v", account.Name)
	}
	if account.NativeID != testAccount.AccountID {
		t.Fatalf("invalid native id: %v", account.NativeID)
	}
	if n := len(account.Features); n != 1 {
		t.Fatalf("invalid number of features: %v", n)
	}
	if !account.Features[0].Equal(core.FeatureExocompute) {
		t.Fatalf("invalid feature name: %v", account.Features[0].Name)
	}
	if regions := account.Features[0].Regions; !reflect.DeepEqual(regions, []string{"us-east-2"}) {
		t.Fatalf("invalid feature regions: %v", regions)
	}
	if account.Features[0].Status != core.StatusConnected {
		t.Fatalf("invalid feature status: %v", account.Features[0].Status)
	}

	exoID, err := awsClient.AddExocomputeConfig(ctx, AccountID(testAccount.AccountID),
		Managed("us-east-2", testAccount.Exocompute.VPCID,
			[]string{testAccount.Exocompute.Subnets[0].ID, testAccount.Exocompute.Subnets[1].ID}))
	if err != nil {
		t.Fatal(err)
	}

	validateConfig := func(exoID uuid.UUID, sn1inx, sn2inx int) {
		// Retrieve the exocompute config added.
		exoConfig, err := awsClient.ExocomputeConfig(ctx, exoID)
		if err != nil {
			t.Fatal(err)
		}
		if exoConfig.ID != exoID {
			t.Fatalf("invalid id: %v", exoConfig.ID)
		}
		if exoConfig.Region != "us-east-2" {
			t.Fatalf("invalid region: %v", exoConfig.Region)
		}
		if exoConfig.VPCID != testAccount.Exocompute.VPCID {
			t.Fatalf("invalid vpc id: %v", exoConfig.VPCID)
		}
		sn1 := testAccount.Exocompute.Subnets[sn1inx]
		sn2 := testAccount.Exocompute.Subnets[sn2inx]
		if sn := exoConfig.Subnets[0]; sn.ID != sn1.ID && sn.ID != sn2.ID {
			t.Fatalf("invalid subnet id: %v", sn.ID)
		}
		if sn := exoConfig.Subnets[0]; sn.AvailabilityZone != sn1.AvailabilityZone && sn.AvailabilityZone != sn2.AvailabilityZone {
			t.Fatalf("invalid subnet availability zone: %v", sn.AvailabilityZone)
		}
		if sn := exoConfig.Subnets[1]; sn.ID != sn1.ID && sn.ID != sn2.ID {
			t.Fatalf("invalid subnet id: %v", sn.ID)
		}
		if sn := exoConfig.Subnets[1]; sn.AvailabilityZone != sn1.AvailabilityZone && sn.AvailabilityZone != sn2.AvailabilityZone {
			t.Fatalf("invalid subnet availability zone: %v", sn.AvailabilityZone)
		}
		if !exoConfig.ManagedByRubrik {
			t.Fatalf("invalid polaris managed state: %t", exoConfig.ManagedByRubrik)
		}
	}

	// Verify that the exocompute config uses subnet 0 & 1 from the test account.
	validateConfig(exoID, 0, 1)

	// Update the exocompute config to use subnet 1 & 2.
	updatedExoID, err := awsClient.UpdateExocomputeConfig(ctx, AccountID(testAccount.AccountID),
		Managed("us-east-2", testAccount.Exocompute.VPCID,
			[]string{testAccount.Exocompute.Subnets[1].ID, testAccount.Exocompute.Subnets[2].ID}))
	if err != nil {
		t.Fatal(err)
	}

	if updatedExoID != exoID {
		t.Fatalf("invalid exo id post update, expected: %v, got: %v", exoID, updatedExoID)
	}

	// Verify that the exocompute config has been updated to use subnet 1 & 2 from the test account.
	validateConfig(exoID, 1, 2)

	// Remove the exocompute config.
	err = awsClient.RemoveExocomputeConfig(ctx, exoID)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the exocompute config was successfully removed.
	_, err = awsClient.ExocomputeConfig(ctx, exoID)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}

	// Disable the exocompute feature for the account.
	err = awsClient.RemoveAccount(ctx, Profile(testAccount.Profile), []core.Feature{core.FeatureExocompute}, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the exocompute feature was successfully disabled.
	account, err = awsClient.Account(ctx, AccountID(testAccount.AccountID), core.FeatureExocompute)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}

	// Remove the AWS account from RSC.
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
