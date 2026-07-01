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

package aws

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlaws "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	awsregions "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// featureConnectPollInterval is how often the onboarding flow polls RSC for the
// account's feature connection status while waiting for the CloudFormation
// stack to finish wiring up the features.
const featureConnectPollInterval = 15 * time.Second

// ManagedAccountArtifacts holds the output of the validate-and-create step
// (phase 1) of the RSC-managed AWS (BaaS) onboarding flow. It carries the
// CloudFormation template information needed to deploy the stack, plus the
// values that the finalize step (phase 2) must pass back to RSC.
type ManagedAccountArtifacts struct {
	CloudFormationURL string
	TemplateURL       string
	StackName         string
	ExternalID        string
	AWSIamPairID      string
	Name              string
}

// BaaSDefaultFeatureNames returns the features onboarded by default for the
// RSC-managed AWS (BaaS) flow: EC2 (CLOUD_NATIVE_PROTECTION), RDS
// (RDS_PROTECTION), S3 (CLOUD_NATIVE_S3_PROTECTION) and Cloud Discovery
// (CLOUD_DISCOVERY).
func BaaSDefaultFeatureNames() []string {
	return []string{
		core.FeatureCloudNativeProtection.Name,
		core.FeatureRDSProtection.Name,
		core.FeatureCloudNativeS3Protection.Name,
		core.FeatureCloudDiscovery.Name,
	}
}

// BaaSSupportedRegions returns the AWS commercial regions supported by the
// RSC-managed AWS (BaaS) onboarding flow. It is used as the default region set
// when the caller does not specify regions explicitly. This is a curated
// allow-list (RSC exposes no API for it) and is not simply the commercial
// partition - GovCloud, China, ISO and the Middle East regions are excluded.
func BaaSSupportedRegions() []awsregions.Region {
	return []awsregions.Region{
		awsregions.RegionAfSouth1,
		awsregions.RegionApEast1,
		awsregions.RegionApNorthEast1,
		awsregions.RegionApNorthEast2,
		awsregions.RegionApNorthEast3,
		awsregions.RegionApSouthEast1,
		awsregions.RegionApSouthEast2,
		awsregions.RegionApSouthEast3,
		awsregions.RegionApSouthEast4,
		awsregions.RegionApSouthEast5,
		awsregions.RegionApSouthEast7,
		awsregions.RegionApSouth1,
		awsregions.RegionApSouth2,
		awsregions.RegionCaCentral1,
		awsregions.RegionCaWest1,
		awsregions.RegionEuCentral1,
		awsregions.RegionEuCentral2,
		awsregions.RegionEuNorth1,
		awsregions.RegionEuSouth1,
		awsregions.RegionEuSouth2,
		awsregions.RegionEuWest1,
		awsregions.RegionEuWest2,
		awsregions.RegionEuWest3,
		awsregions.RegionIlCentral1,
		awsregions.RegionMxCentral1,
		awsregions.RegionSaEast1,
		awsregions.RegionUsEast1,
		awsregions.RegionUsEast2,
		awsregions.RegionUsWest1,
		awsregions.RegionUsWest2,
	}
}

// resolveBaasFeatures expands the given feature names into the full set of
// permission groups (and their versions) supported by the BaaS service type.
// It returns the features with all permission groups populated (for the
// *WithPermissionsGroups inputs) and the matching feature versions (for the
// finalize step). When featureNames is empty the BaaS default feature set is
// used.
func (a API) resolveBaasFeatures(ctx context.Context, featureNames []string) ([]core.Feature, []gqlaws.FeatureVersion, error) {
	if len(featureNames) == 0 {
		featureNames = BaaSDefaultFeatureNames()
	}

	lookup := make([]core.Feature, 0, len(featureNames))
	for _, name := range featureNames {
		lookup = append(lookup, core.Feature{Name: name})
	}

	groups, err := gqlaws.Wrap(a.client).AllPermissionsGroupsByFeature(ctx, lookup, gqlaws.ServiceTypeBaaS)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to look up BaaS permission groups: %s", err)
	}

	features := make([]core.Feature, 0, len(groups))
	versions := make([]gqlaws.FeatureVersion, 0, len(groups))
	for _, group := range groups {
		feature := core.Feature{Name: group.Feature}
		version := gqlaws.FeatureVersion{Name: group.Feature}
		for _, pg := range group.PermissionGroups {
			feature = feature.WithPermissionGroups(pg.PermissionGroup)
			version.PermissionGroupsVersion = append(version.PermissionGroupsVersion, struct {
				PermissionGroups string `json:"permissionsGroup"`
				Version          int    `json:"version"`
			}{PermissionGroups: string(pg.PermissionGroup), Version: pg.Version})
		}
		features = append(features, feature)
		versions = append(versions, version)
	}

	return features, versions, nil
}

// PrepareManagedAccount runs the validate-and-create step of the RSC-managed
// AWS (BaaS) onboarding flow and returns the CloudFormation artifacts. It does
// NOT create the CloudFormation stack - that is the caller's responsibility
// (e.g. the Terraform AWS provider). The returned artifacts must be passed to
// OnboardManagedAccount once the stack has been deployed. When featureNames is
// empty the BaaS default feature set is used.
func (a API) PrepareManagedAccount(ctx context.Context, account AccountFunc, featureNames []string, opts ...OptionFunc) (ManagedAccountArtifacts, error) {
	a.log.Print(log.Trace)

	if account == nil {
		return ManagedAccountArtifacts{}, errors.New("account is not allowed to be nil")
	}

	config, err := account(ctx)
	if err != nil {
		return ManagedAccountArtifacts{}, fmt.Errorf("failed to lookup account: %s", err)
	}
	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return ManagedAccountArtifacts{}, fmt.Errorf("failed to lookup option: %s", err)
		}
	}
	if options.name != "" {
		config.name = options.name
	}

	// If there already is an RSC cloud account for the given AWS account we use
	// the same account name. RSC does not allow the name to change between
	// features.
	cloudAccount, err := a.AccountByNativeID(ctx, config.NativeID)
	if err != nil && !errors.Is(err, graphql.ErrNotFound) {
		return ManagedAccountArtifacts{}, fmt.Errorf("failed to get account: %s", err)
	}
	if err == nil {
		config.name = cloudAccount.Name
	}

	features, _, err := a.resolveBaasFeatures(ctx, featureNames)
	if err != nil {
		return ManagedAccountArtifacts{}, err
	}

	init, err := gqlaws.Wrap(a.client).ValidateAndCreateCloudAccount(ctx, config.cloud, config.NativeID, config.name, features, "", gqlaws.ServiceTypeBaaS)
	if err != nil {
		return ManagedAccountArtifacts{}, fmt.Errorf("failed to validate and create account: %s", err)
	}

	return ManagedAccountArtifacts{
		CloudFormationURL: init.CloudFormationURL,
		TemplateURL:       init.TemplateURL,
		StackName:         init.StackName,
		ExternalID:        init.ExternalID,
		AWSIamPairID:      init.AWSIamPairID,
		Name:              config.name,
	}, nil
}

// OnboardManagedAccount runs the post-CloudFormation steps of the RSC-managed
// AWS (BaaS) onboarding flow: it finalizes the account protection, triggers
// CloudFormation status polling, waits for all features to connect and then
// completes the BaaS onboarding. It returns the RSC cloud account ID. When
// regions is empty the BaaS-supported region set is used, and when featureNames
// is empty the BaaS default feature set is used.
func (a API) OnboardManagedAccount(ctx context.Context, account AccountFunc, featureNames []string, regions []awsregions.Region, artifacts ManagedAccountArtifacts, opts ...OptionFunc) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	if account == nil {
		return uuid.Nil, errors.New("account is not allowed to be nil")
	}

	config, err := account(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to lookup account: %s", err)
	}
	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return uuid.Nil, fmt.Errorf("failed to lookup option: %s", err)
		}
	}
	if options.name != "" {
		config.name = options.name
	}
	if config.name == "" {
		config.name = artifacts.Name
	}
	if len(regions) == 0 {
		regions = BaaSSupportedRegions()
	}

	features, versions, err := a.resolveBaasFeatures(ctx, featureNames)
	if err != nil {
		return uuid.Nil, err
	}

	// 1. Finalize the account protection. This creates the RSC cloud account.
	init := gqlaws.CloudAccountInitiate{
		ExternalID:      artifacts.ExternalID,
		AWSIamPairID:    artifacts.AWSIamPairID,
		StackName:       artifacts.StackName,
		FeatureVersions: versions,
	}
	if err := gqlaws.Wrap(a.client).FinalizeCloudAccountProtection(ctx, config.cloud, config.NativeID, config.name, features, regions, artifacts.AWSIamPairID, gqlaws.ServiceTypeBaaS, init); err != nil {
		return uuid.Nil, fmt.Errorf("failed to finalize account protection: %s", err)
	}

	// 2. Resolve the RSC cloud account ID created by the finalize step.
	cloudAccount, err := a.AccountByNativeID(ctx, config.NativeID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get account after finalize: %s", err)
	}
	accountID := cloudAccount.ID

	// 3. Trigger CloudFormation status polling. This is best-effort - RSC will
	// also reconcile the status on its own, so a failure here is logged and the
	// onboarding continues.
	if err := gqlaws.Wrap(a.client).TriggerCftStatusPolling(ctx, config.NativeID, features); err != nil {
		a.log.Printf(log.Warn, "failed to trigger CloudFormation status polling, continuing: %s", err)
	}

	// 4. Wait for all features to connect.
	if err := a.waitForFeaturesConnected(ctx, accountID, features); err != nil {
		return uuid.Nil, err
	}

	// 5. Complete the BaaS onboarding.
	if err := gqlaws.Wrap(a.client).CompleteBaasOnboarding(ctx, accountID, config.NativeID, config.name, features, regions); err != nil {
		return uuid.Nil, fmt.Errorf("failed to complete BaaS onboarding: %s", err)
	}

	return accountID, nil
}

// waitForFeaturesConnected polls RSC until every specified feature of the
// account reaches the CONNECTED status, or until the context is cancelled.
func (a API) waitForFeaturesConnected(ctx context.Context, accountID uuid.UUID, features []core.Feature) error {
	ticker := time.NewTicker(featureConnectPollInterval)
	defer ticker.Stop()

	for {
		connected, err := a.featuresConnected(ctx, accountID, features)
		if err != nil {
			return err
		}
		if connected {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for account %q features to connect: %w", accountID, ctx.Err())
		case <-ticker.C:
		}
	}
}

// featuresConnected returns true if every specified feature of the account has
// reached the CONNECTED status.
func (a API) featuresConnected(ctx context.Context, accountID uuid.UUID, features []core.Feature) (bool, error) {
	cloudAccount, err := a.AccountByID(ctx, accountID)
	if err != nil {
		return false, fmt.Errorf("failed to get account status: %s", err)
	}

	for _, feature := range features {
		f, ok := cloudAccount.Feature(feature)
		if !ok || f.Status != core.StatusConnected {
			return false, nil
		}
	}

	return true, nil
}
