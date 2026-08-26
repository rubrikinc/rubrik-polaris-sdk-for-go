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
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
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

	awsClient := Wrap(client)

	// Adds the AWS account identified by the specified profile to RSC. Note
	// that the profile needs to have a default region.
	cnpFeature := []core.Feature{core.FeatureCloudNativeProtection.WithPermissionGroups(core.PermissionGroupBasic)}
	id, err := awsClient.AddAccountWithCFT(ctx, Profile(testAccount.Profile), cnpFeature, Name(testAccount.AccountName),
		Regions("us-east-2"))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully added.
	account, err := awsClient.AccountByID(ctx, id)
	if err != nil {
		t.Error(err)
	}
	if account.Name != testAccount.AccountName {
		t.Fatalf("invalid name: %v", account.Name)
	}
	if account.NativeID != testAccount.AccountID {
		t.Fatalf("invalid native id: %v", account.NativeID)
	}
	cnp, ok := account.Feature(core.FeatureCloudNativeProtection)
	if !ok {
		t.Fatal("cloud native protection feature not found")
	}
	if regions := cnp.Regions; !reflect.DeepEqual(regions, []string{"us-east-2"}) {
		t.Fatalf("invalid feature regions: %v", regions)
	}
	if cnp.Status != core.StatusConnected {
		t.Fatalf("invalid feature status: %v", cnp.Status)
	}
	if mode := account.OnboardingMode(); mode != OnboardingModeCFT {
		t.Fatalf("invalid onboarding mode: %v", mode)
	}
	if mode := cnp.OnboardingMode(); mode != OnboardingModeCFT {
		t.Fatalf("invalid feature onboarding mode: %v", mode)
	}

	// Update and verify regions for AWS account.
	err = awsClient.UpdateAccount(ctx, account.ID, core.FeatureCloudNativeProtection, Regions("us-west-2"))
	if err != nil {
		t.Fatal(err)
	}
	account, err = awsClient.AccountByNativeID(ctx, testAccount.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	cnp, ok = account.Feature(core.FeatureCloudNativeProtection)
	if !ok {
		t.Fatal("cloud native protection feature not found")
	}
	if regions := cnp.Regions; !reflect.DeepEqual(regions, []string{"us-west-2"}) {
		t.Fatalf("invalid feature regions: %v", regions)
	}

	// Remove AWS account from RSC. All the features of the account are removed
	// and not just the features added by the test, since RSC may add the cloud
	// cost report feature to the account when it is onboarded.
	// RemoveAccountWithCFT orders the features for removal.
	removeFeatures := make([]core.Feature, 0, len(account.Features))
	for _, feature := range account.Features {
		removeFeatures = append(removeFeatures, feature.Feature)
	}
	if err := awsClient.RemoveAccountWithCFT(ctx, Profile(testAccount.Profile), removeFeatures, false); err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	if _, err := awsClient.AccountByNativeID(ctx, testAccount.AccountID); !errors.Is(err, graphql.ErrNotFound) {
		t.Fatalf("expected the account to be removed, got error: %v", err)
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

	// Adds the AWS account identified by the specified profile to RSC. Note
	// that the profile needs to have a default region.
	features := []core.Feature{
		core.FeatureCloudNativeProtection.WithPermissionGroups(core.PermissionGroupBasic),
		core.FeatureExocompute.WithPermissionGroups(core.PermissionGroupBasic, core.PermissionGroupRSCManagedCluster),
	}
	id, err := awsClient.AddAccountWithCFT(ctx, Profile(testAccount.Profile), features, Name(testAccount.AccountName),
		Regions("us-east-2"))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully added.
	account, err := awsClient.AccountByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	if account.Name != testAccount.AccountName {
		t.Fatalf("invalid name: %v", account.Name)
	}
	if account.NativeID != testAccount.AccountID {
		t.Fatalf("invalid native id: %v", account.NativeID)
	}
	cnp, ok := account.Feature(core.FeatureCloudNativeProtection)
	if !ok {
		t.Fatal("cloud native protection feature not found")
	}
	if regions := cnp.Regions; !reflect.DeepEqual(regions, []string{"us-east-2"}) {
		t.Fatalf("invalid feature regions: %v", regions)
	}
	if cnp.Status != core.StatusConnected {
		t.Fatalf("invalid feature status: %v", cnp.Status)
	}
	if groups := cnp.PermissionGroups; !reflect.DeepEqual(groups, []core.PermissionGroup{core.PermissionGroupBasic}) {
		t.Fatalf("invalid permission groups: %v", groups)
	}
	exo, ok := account.Feature(core.FeatureExocompute)
	if !ok {
		t.Fatal("exocompute feature not found")
	}
	if regions := exo.Regions; !reflect.DeepEqual(regions, []string{"us-east-2"}) {
		t.Fatalf("invalid feature regions: %v", regions)
	}
	if exo.Status != core.StatusConnected {
		t.Fatalf("invalid feature status: %v", exo.Status)
	}
	if groups := exo.PermissionGroups; !reflect.DeepEqual(groups, []core.PermissionGroup{core.PermissionGroupBasic, core.PermissionGroupRSCManagedCluster}) {
		t.Fatalf("invalid permission groups: %v", groups)
	}
	if mode := account.OnboardingMode(); mode != OnboardingModeCFT {
		t.Fatalf("invalid onboarding mode: %v", mode)
	}
	if mode := cnp.OnboardingMode(); mode != OnboardingModeCFT {
		t.Fatalf("invalid cloud native protection onboarding mode: %v", mode)
	}
	if mode := exo.OnboardingMode(); mode != OnboardingModeCFT {
		t.Fatalf("invalid exocompute onboarding mode: %v", mode)
	}

	// Remove AWS account from RSC. All the features of the account are removed
	// and not just the features added by the test, since RSC may add the cloud
	// cost report feature to the account when it is onboarded.
	// RemoveAccountWithCFT orders the features for removal.
	removeFeatures := make([]core.Feature, 0, len(account.Features))
	for _, feature := range account.Features {
		removeFeatures = append(removeFeatures, feature.Feature)
	}
	if err := awsClient.RemoveAccountWithCFT(ctx, Profile(testAccount.Profile), removeFeatures, false); err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	if _, err := awsClient.AccountByNativeID(ctx, testAccount.AccountID); !errors.Is(err, graphql.ErrNotFound) {
		t.Fatalf("expected the account to be removed, got error: %v", err)
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
	cnpFeature := []core.Feature{core.FeatureCloudNativeProtection.WithPermissionGroups(core.PermissionGroupBasic)}
	id, err := awsClient.AddAccountWithCFT(ctx, ProfileWithRole(testAccount.Profile, testAccount.CrossAccountRole),
		cnpFeature, Name(testAccount.CrossAccountName), Regions("us-east-2"))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully added.
	account, err := awsClient.AccountByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != testAccount.CrossAccountName {
		t.Fatalf("invalid name: %v", account.Name)
	}
	if account.NativeID != testAccount.CrossAccountID {
		t.Fatalf("invalid native id: %v", account.NativeID)
	}
	cnp, ok := account.Feature(core.FeatureCloudNativeProtection)
	if !ok {
		t.Fatal("cloud native protection feature not found")
	}
	if regions := cnp.Regions; !reflect.DeepEqual(regions, []string{"us-east-2"}) {
		t.Fatalf("invalid feature regions: %v", regions)
	}
	if cnp.Status != core.StatusConnected {
		t.Fatalf("invalid feature status: %v", cnp.Status)
	}
	if mode := account.OnboardingMode(); mode != OnboardingModeCFT {
		t.Fatalf("invalid onboarding mode: %v", mode)
	}
	if mode := cnp.OnboardingMode(); mode != OnboardingModeCFT {
		t.Fatalf("invalid feature onboarding mode: %v", mode)
	}

	// Verify that it's possible to find the account using the account ID
	// of the cross account role.
	roleARN, err := arn.Parse(testAccount.CrossAccountRole)
	if err != nil {
		t.Fatal(err)
	}
	account, err = awsClient.AccountByNativeID(ctx, roleARN.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	if account.ID != id {
		t.Fatalf("invalid id: %v", account.ID)
	}

	// Remove AWS account from RSC using a cross account role. All the features
	// of the account are removed and not just the features added by the test,
	// since RSC may add the cloud cost report feature to the account when it is
	// onboarded. RemoveAccountWithCFT orders the features for removal.
	removeFeatures := make([]core.Feature, 0, len(account.Features))
	for _, feature := range account.Features {
		removeFeatures = append(removeFeatures, feature.Feature)
	}
	err = awsClient.RemoveAccountWithCFT(ctx, ProfileWithRole(testAccount.Profile, testAccount.CrossAccountRole),
		removeFeatures, false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the account was successfully removed.
	if _, err := awsClient.AccountByNativeID(ctx, testAccount.CrossAccountID); !errors.Is(err, graphql.ErrNotFound) {
		t.Fatalf("expected the account to be removed, got error: %v", err)
	}
}
