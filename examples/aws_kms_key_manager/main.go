// Copyright 2026 Rubrik, Inc.
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

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	gqlkms "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/kms"
	gqlaws "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/kms"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Example showing how to manage an AWS KMS key manager with the SDK.
//
// A key manager gives RSC access to the KMS keys of an AWS account. Archival
// locations encrypted with a customer managed key refer to the key by the ID of
// the key manager together with the ID of the key itself.
//
// The example onboards the default AWS account with the
// CLOUD_NATIVE_UEM_KEY_MANAGEMENT feature, and removes it again at the end.
// The AWS account is usually resolved using the environment variables
// AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_DEFAULT_REGION.
//
// Note that the RSC account needs the AWS_KMS_FOR_UEM_ENABLED and
// REL_ENABLE_AWS_EC2_UEM feature flags.
//
// The RSC service account key file, identifying the RSC account, should either
// be placed at ~/.rubrik/polaris-service-account.json or pointed out by the
// RUBRIK_POLARIS_SERVICEACCOUNT_FILE environment variable.
func main() {
	ctx := context.Background()

	// Load configuration and create client.
	polAccount, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		log.Fatal(err)
	}
	logger := polarislog.NewStandardLogger()
	if err := polaris.SetLogLevelFromEnv(logger); err != nil {
		log.Fatal(err)
	}
	client, err := polaris.NewClientWithLogger(polAccount, logger)
	if err != nil {
		log.Fatal(err)
	}

	// RSC features and their permission groups.
	features := []core.Feature{
		core.FeatureCloudNativeUEMKeyManagement.WithPermissionGroups(core.PermissionGroupBasic),
	}

	// Add the AWS default account to RSC with the key management feature.
	cloudAccountID, err := aws.Wrap(client).AddAccountWithCFT(ctx, aws.Default(), features, aws.Regions("us-east-2"))
	if err != nil {
		log.Fatal(err)
	}

	// Register the AWS account's KMS as a key manager.
	keyManagerID, err := kms.Wrap(client).CreateAWSKeyManager(ctx, gqlkms.CreateAWSKeyManagerParams{
		CloudAccountID: cloudAccountID.String(),
		Name:           "Example Key Manager",
		Region:         gqlaws.RegionUsEast2,
		Description:    "Managed by the Rubrik Polaris SDK example",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Key manager ID: %s\n", keyManagerID)

	// List the KMS keys reachable through the key manager. Together with the
	// key manager ID, one of these key IDs identifies the encryption key of an
	// archival location.
	keyIDs, err := kms.Wrap(client).AWSKeyIDs(ctx, keyManagerID)
	if err != nil {
		log.Fatal(err)
	}
	for _, keyID := range keyIDs {
		fmt.Printf("KMS key ID: %s\n", keyID)
	}

	// List the AWS key managers.
	keyManagers, err := kms.Wrap(client).AWSKeyManagers(ctx, gqlkms.ListKeyManagerFilter{})
	if err != nil {
		log.Fatal(err)
	}
	for _, keyManager := range keyManagers {
		fmt.Printf("Key manager: %s, %s, %s\n", keyManager.KeyVaultInfo.ID, keyManager.Name, keyManager.Type)
	}

	// Remove the key manager. RSC refuses to remove a key manager which still
	// encrypts a data location.
	if err := kms.Wrap(client).DeleteKeyManager(ctx, keyManagerID); err != nil {
		log.Fatal(err)
	}

	// Remove the AWS account from RSC.
	if err := aws.Wrap(client).RemoveAccountWithCFT(ctx, aws.Default(), features, false); err != nil {
		log.Fatal(err)
	}
}
