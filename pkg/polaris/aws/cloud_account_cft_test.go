// Copyright 2025 Rubrik, Inc.
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
	"cmp"
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// TestAwsAccountAddAndRemoveWithCFT verifies that the SDK can perform the basic
// AWS account operations on a real RSC instance.
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
func TestAwsAccountAddAndRemoveWithCFT(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	testAccount, err := testsetup.AWSAccount()
	if err != nil {
		t.Fatal(err)
	}

	testAccountID, err := uuid.Parse(testAccount.AccountID)
	if err != nil {
		t.Fatal(err)
	}

	awsClient := Wrap(client)

	// Adds the AWS account identified by the specified profile to RSC. Note
	// that the profile needs to have a default region.
	id, err := awsClient.AddAccountWithCFT(ctx, Profile(testAccount.Profile), []core.Feature{core.FeatureCloudNativeProtection},
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
		t.Fatalf("invalid name: %v", account.Name)
	}
	if account.NativeID != testAccount.AccountID {
		t.Fatalf("invalid native id: %v", account.NativeID)
	}
	if n := len(account.Features); n != 1 {
		t.Fatalf("invalid number of features: %v", n)
	}
	if !account.Features[0].Equal(core.FeatureCloudNativeProtection) {
		t.Fatalf("invalid feature name: %v", account.Features[0].Name)
	}
	if regions := account.Features[0].Regions; !reflect.DeepEqual(regions, []string{"us-east-2"}) {
		t.Fatalf("invalid feature regions: %v", regions)
	}
	if account.Features[0].Status != core.StatusConnected {
		t.Fatalf("invalid feature status: %v", account.Features[0].Status)
	}

	// Update and verify regions for AWS account.
	err = awsClient.UpdateAccount(ctx, testAccountID, core.FeatureCloudNativeProtection,
		Regions("us-west-2"))
	if err != nil {
		t.Fatal(err)
	}
	account, err = awsClient.Account(ctx, AccountID(testAccount.AccountID), core.FeatureCloudNativeProtection)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(account.Features); n != 1 {
		t.Fatalf("invalid number of features: %v", n)
	}
	if regions := account.Features[0].Regions; !reflect.DeepEqual(regions, []string{"us-west-2"}) {
		t.Fatalf("invalid feature regions: %v", regions)
	}

	// Remove AWS account from RSC.
	err = awsClient.RemoveAccountWithCFT(ctx, Profile(testAccount.Profile), []core.Feature{core.FeatureCloudNativeProtection}, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	account, err = awsClient.Account(ctx, AccountID(testAccount.AccountID), core.FeatureCloudNativeProtection)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAwsAccountAddAndRemoveUsingPermissionGroupsWithCFT verifies that the SDK
// can add and remove and AWS account using permission groups on a real RSC
// instance.
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
func TestAwsAccountAddAndRemoveUsingPermissionGroupsWithCFT(t *testing.T) {
	ctx := context.Background()

	if !testsetup.BoolEnvSet("TEST_INTEGRATION") {
		t.Skipf("skipping due to env TEST_INTEGRATION not set")
	}

	testAccount, err := testsetup.AWSAccount()
	if err != nil {
		t.Fatal(err)
	}

	awsClient := Wrap(client)

	// RSC features and their permission groups.
	features := []core.Feature{
		core.FeatureCloudNativeProtection.WithPermissionGroups(core.PermissionGroupBasic),
		core.FeatureExocompute.WithPermissionGroups(core.PermissionGroupBasic, core.PermissionGroupRSCManagedCluster),
	}

	// Adds the AWS account identified by the specified profile to RSC. Note
	// that the profile needs to have a default region.
	id, err := awsClient.AddAccountWithCFT(ctx, Profile(testAccount.Profile), features, Name(testAccount.AccountName),
		Regions("us-east-2"))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully added.
	account, err := awsClient.Account(ctx, CloudAccountID(id), core.FeatureAll)
	if err != nil {
		t.Fatal(err)
	}

	if account.Name != testAccount.AccountName {
		t.Fatalf("invalid name: %v", account.Name)
	}
	if account.NativeID != testAccount.AccountID {
		t.Fatalf("invalid native id: %v", account.NativeID)
	}
	if n := len(account.Features); n != 2 {
		t.Fatalf("invalid number of features: %v", n)
	}
	slices.SortFunc(account.Features, func(lhs, rhs Feature) int {
		return cmp.Compare(lhs.Feature.Name, rhs.Feature.Name)
	})
	if !account.Features[0].Equal(core.FeatureCloudNativeProtection) {
		t.Fatalf("invalid feature name: %v", account.Features[0].Name)
	}
	if regions := account.Features[0].Regions; !reflect.DeepEqual(regions, []string{"us-east-2"}) {
		t.Fatalf("invalid feature regions: %v", regions)
	}
	if account.Features[0].Status != core.StatusConnected {
		t.Fatalf("invalid feature status: %v", account.Features[0].Status)
	}
	if groups := account.Features[0].PermissionGroups; !reflect.DeepEqual(groups, []core.PermissionGroup{core.PermissionGroupBasic}) {
		t.Fatalf("invalid permission groups: %v", groups)
	}
	if !account.Features[1].Equal(core.FeatureExocompute) {
		t.Fatalf("invalid feature name: %v", account.Features[1].Name)
	}
	if regions := account.Features[1].Regions; !reflect.DeepEqual(regions, []string{"us-east-2"}) {
		t.Fatalf("invalid feature regions: %v", regions)
	}
	if account.Features[1].Status != core.StatusConnected {
		t.Fatalf("invalid feature status: %v", account.Features[0].Status)
	}
	if groups := account.Features[1].PermissionGroups; !reflect.DeepEqual(groups, []core.PermissionGroup{core.PermissionGroupBasic, core.PermissionGroupRSCManagedCluster}) {
		t.Fatalf("invalid permission groups: %v", groups)
	}
	// Remove AWS account from RSC.
	err = awsClient.RemoveAccountWithCFT(ctx, Profile(testAccount.Profile), features, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	account, err = awsClient.Account(ctx, AccountID(testAccount.AccountID), core.FeatureAll)
	if !errors.Is(err, graphql.ErrNotFound) {
		t.Fatal(err)
	}
}

// TestAwsCrossAccountAddAndRemoveWithCFT verifies that the SDK can perform the
// basic AWS cross account operations on a real RSC instance.
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
func TestAwsCrossAccountAddAndRemoveWithCFT(t *testing.T) {
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
	id, err := awsClient.AddAccountWithCFT(ctx,
		ProfileWithRole(testAccount.Profile, testAccount.CrossAccountRole), []core.Feature{core.FeatureCloudNativeProtection},
		Name(testAccount.CrossAccountName), Regions("us-east-2"))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully added.
	account, err := awsClient.Account(ctx, CloudAccountID(id), core.FeatureCloudNativeProtection)
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != testAccount.CrossAccountName {
		t.Fatalf("invalid name: %v", account.Name)
	}
	if account.NativeID != testAccount.CrossAccountID {
		t.Fatalf("invalid native id: %v", account.NativeID)
	}
	if n := len(account.Features); n != 1 {
		t.Fatalf("invalid number of features: %v", n)
	}
	if !account.Features[0].Equal(core.FeatureCloudNativeProtection) {
		t.Fatalf("invalid feature name: %v", account.Features[0].Name)
	}
	if regions := account.Features[0].Regions; !reflect.DeepEqual(regions, []string{"us-east-2"}) {
		t.Fatalf("invalid feature regions: %v", regions)
	}
	if account.Features[0].Status != core.StatusConnected {
		t.Fatalf("invalid feature status: %v", account.Features[0].Status)
	}

	// Verify that it's possible to search for the account using a role.
	account, err = awsClient.Account(ctx, Role(testAccount.CrossAccountRole), core.FeatureCloudNativeProtection)
	if err != nil {
		t.Fatal(err)
	}
	if account.ID != id {
		t.Fatalf("invalid id: %v", account.ID)
	}

	// Remove AWS account from RSC using a cross account role.
	err = awsClient.RemoveAccountWithCFT(ctx, ProfileWithRole(testAccount.Profile, testAccount.CrossAccountRole),
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
