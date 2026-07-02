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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
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

// ManagedAccountArtifacts holds the output of RegisterManagedAccount: the RSC
// cloud account ID plus the CloudFormation template information needed to
// deploy the cross-account stack.
type ManagedAccountArtifacts struct {
	AccountID         uuid.UUID
	CloudFormationURL string
	TemplateURL       string
	StackName         string
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

// baasFeatures expands the given feature names into features with all of their
// BaaS permission groups populated, for use in the validate and finalize
// inputs. When featureNames is empty the BaaS default feature set is used.
func (a API) baasFeatures(ctx context.Context, featureNames []string) ([]core.Feature, error) {
	if len(featureNames) == 0 {
		featureNames = BaaSDefaultFeatureNames()
	}

	lookup := make([]core.Feature, 0, len(featureNames))
	for _, name := range featureNames {
		lookup = append(lookup, core.Feature{Name: name})
	}

	groups, err := gqlaws.Wrap(a.client).AllPermissionsGroupsByFeature(ctx, lookup, gqlaws.ServiceTypeBaaS)
	if err != nil {
		return nil, fmt.Errorf("failed to look up BaaS permission groups: %s", err)
	}

	features := make([]core.Feature, 0, len(groups))
	for _, group := range groups {
		feature := core.Feature{Name: group.Feature}
		for _, pg := range group.PermissionGroups {
			feature = feature.WithPermissionGroups(pg.PermissionGroup)
		}
		features = append(features, feature)
	}

	return features, nil
}

// RegisterManagedAccount runs the first phase of the RSC-managed AWS (BaaS)
// onboarding flow: it validates the account, registers it with RSC (finalize)
// and returns the RSC cloud account ID together with the CloudFormation
// template information. The caller must then deploy the CloudFormation stack
// (e.g. with the Terraform AWS provider) and call CompleteManagedAccountOnboarding.
//
// When featureNames is empty the BaaS default feature set is used and when
// regions is empty the BaaS-supported region set is used.
func (a API) RegisterManagedAccount(ctx context.Context, account AccountFunc, featureNames []string, regions []awsregions.Region, opts ...OptionFunc) (ManagedAccountArtifacts, error) {
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

	// Reconcile the account name. RSC does not allow the name to change between
	// features, so an existing account keeps its name. A caller-supplied name is
	// honored (never silently overwritten); a conflict with an existing account
	// is reported as an error rather than a name change.
	cloudAccount, err := a.AccountByNativeID(ctx, config.NativeID)
	switch {
	case err == nil:
		if config.name != "" && config.name != cloudAccount.Name {
			return ManagedAccountArtifacts{}, fmt.Errorf(
				"AWS account %s is already registered in RSC as %q; use that name or omit it",
				config.NativeID, cloudAccount.Name)
		}
		config.name = cloudAccount.Name
	case errors.Is(err, graphql.ErrNotFound):
		if config.name == "" {
			config.name = config.NativeID
		}
	default:
		return ManagedAccountArtifacts{}, fmt.Errorf("failed to get account: %s", err)
	}

	if len(regions) == 0 {
		regions = BaaSSupportedRegions()
	}

	features, err := a.baasFeatures(ctx, featureNames)
	if err != nil {
		return ManagedAccountArtifacts{}, err
	}

	// 1. Validate and create - generates the CloudFormation template and returns
	// the feature versions.
	init, err := gqlaws.Wrap(a.client).ValidateAndCreateCloudAccount(ctx, config.cloud, config.NativeID, config.name, features, "", gqlaws.ServiceTypeBaaS)
	if err != nil {
		return ManagedAccountArtifacts{}, fmt.Errorf("failed to validate and create account: %s", err)
	}

	// 2. Finalize - registers the account so RSC can process the stack's
	// notifier callback. This must happen before the CloudFormation stack is
	// deployed.
	if err := gqlaws.Wrap(a.client).FinalizeCloudAccountProtection(ctx, gqlaws.FinalizeCloudAccountProtectionParams{
		Cloud:       config.cloud,
		NativeID:    config.NativeID,
		Name:        config.name,
		Features:    features,
		Regions:     regions,
		ServiceType: gqlaws.ServiceTypeBaaS,
		Initiate:    init,
	}); err != nil {
		return ManagedAccountArtifacts{}, fmt.Errorf("failed to finalize account protection: %s", err)
	}

	// 3. Resolve the RSC cloud account ID created by the finalize step.
	registered, err := a.AccountByNativeID(ctx, config.NativeID)
	if err != nil {
		return ManagedAccountArtifacts{}, fmt.Errorf("failed to get account after finalize: %s", err)
	}

	return ManagedAccountArtifacts{
		AccountID:         registered.ID,
		CloudFormationURL: init.CloudFormationURL,
		TemplateURL:       init.TemplateURL,
		StackName:         init.StackName,
		Name:              config.name,
	}, nil
}

// CompleteManagedAccountOnboarding runs the final phase of the RSC-managed AWS
// (BaaS) onboarding flow for the account with the specified RSC cloud account
// ID. It must be called after the CloudFormation stack has been deployed. The
// account's features and regions are read from RSC (they were set by
// RegisterManagedAccount), so they are not passed in again. It triggers
// CloudFormation status polling, waits for all features to connect and then
// completes the BaaS onboarding.
func (a API) CompleteManagedAccountOnboarding(ctx context.Context, accountID uuid.UUID) error {
	a.log.Print(log.Trace)

	account, err := a.AccountByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %s", err)
	}

	features := make([]core.Feature, 0, len(account.Features))
	regionSet := make(map[awsregions.Region]struct{})
	for _, feature := range account.Features {
		features = append(features, core.Feature{Name: feature.Name})
		for _, name := range feature.Regions {
			if region := awsregions.RegionFromAny(name); region != awsregions.RegionUnknown {
				regionSet[region] = struct{}{}
			}
		}
	}
	regions := make([]awsregions.Region, 0, len(regionSet))
	for region := range regionSet {
		regions = append(regions, region)
	}
	if len(regions) == 0 {
		regions = BaaSSupportedRegions()
	}

	// 1. Trigger CloudFormation status polling. Best-effort - RSC also reconciles
	// the status on its own, so a failure here is logged and onboarding continues.
	if err := gqlaws.Wrap(a.client).TriggerCftStatusPolling(ctx, account.NativeID, features); err != nil {
		a.log.Printf(log.Warn, "failed to trigger CloudFormation status polling, continuing: %s", err)
	}

	// 2. Wait for all features to connect.
	if err := a.waitForFeaturesConnected(ctx, accountID, features); err != nil {
		return err
	}

	// 3. Complete the BaaS onboarding.
	if err := gqlaws.Wrap(a.client).CompleteBaasOnboarding(ctx, accountID, account.NativeID, account.Name, features, regions); err != nil {
		return fmt.Errorf("failed to complete BaaS onboarding: %s", err)
	}

	return nil
}

// PrepareManagedAccountUpdate reports whether the account's deployed permissions
// are out of date - i.e. a feature has the MISSING_PERMISSIONS status, which
// happens when RSC raises a permission version. When an update is needed it
// returns the CloudFormation template URL to redeploy the stack with the updated
// permissions; otherwise it returns an empty string. Because it returns an empty
// string when nothing changed, callers can use it to detect permission drift
// without churning on the signed (ever-changing) template URL.
func (a API) PrepareManagedAccountUpdate(ctx context.Context, accountID uuid.UUID) (string, error) {
	a.log.Print(log.Trace)

	account, err := a.AccountByID(ctx, accountID)
	if err != nil {
		return "", fmt.Errorf("failed to get account: %s", err)
	}

	var missing []core.Feature
	for _, feature := range account.Features {
		if feature.Status == core.StatusMissingPermissions {
			missing = append(missing, core.Feature{Name: feature.Name})
		}
	}
	if len(missing) == 0 {
		return "", nil
	}

	_, templateURL, err := gqlaws.Wrap(a.client).PrepareFeatureUpdateForAwsCloudAccount(ctx, accountID, missing)
	if err != nil {
		return "", fmt.Errorf("failed to prepare feature update: %s", err)
	}

	return templateURL, nil
}

// ManagedAccountPermissionsVersion returns a deterministic identifier for the
// latest permission-group versions of the account's features. It changes only
// when RSC raises a permission version, so it can be used to detect a required
// permissions upgrade at plan time without churning on the signed (ever-
// changing) CloudFormation template URL.
func (a API) ManagedAccountPermissionsVersion(ctx context.Context, accountID uuid.UUID) (string, error) {
	a.log.Print(log.Trace)

	account, err := a.AccountByID(ctx, accountID)
	if err != nil {
		return "", fmt.Errorf("failed to get account: %s", err)
	}

	lookup := make([]core.Feature, 0, len(account.Features))
	for _, feature := range account.Features {
		lookup = append(lookup, core.Feature{Name: feature.Name})
	}

	groups, err := gqlaws.Wrap(a.client).AllPermissionsGroupsByFeature(ctx, lookup, gqlaws.ServiceTypeBaaS)
	if err != nil {
		return "", fmt.Errorf("failed to look up permission versions: %s", err)
	}

	lines := make([]string, 0)
	for _, group := range groups {
		for _, pg := range group.PermissionGroups {
			lines = append(lines, fmt.Sprintf("%s/%s/%d", group.Feature, pg.PermissionGroup, pg.Version))
		}
	}
	sort.Strings(lines)

	sum := sha256.Sum256([]byte(strings.Join(lines, ";")))
	return hex.EncodeToString(sum[:]), nil
}

// CompleteManagedAccountUpdate notifies RSC that the CloudFormation stack has
// been redeployed with updated permissions for the account's features that were
// missing permissions, then waits for those features to reconnect. It is the
// permission-update counterpart of CompleteManagedAccountOnboarding and is run
// after the CloudFormation stack has been updated.
func (a API) CompleteManagedAccountUpdate(ctx context.Context, accountID uuid.UUID) error {
	a.log.Print(log.Trace)

	// Notify RSC about every feature that is missing permissions (nil = all).
	if err := a.PermissionsUpdated(ctx, accountID, nil); err != nil {
		return fmt.Errorf("failed to notify RSC of updated permissions: %s", err)
	}

	account, err := a.AccountByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %s", err)
	}
	features := make([]core.Feature, 0, len(account.Features))
	for _, feature := range account.Features {
		features = append(features, core.Feature{Name: feature.Name})
	}

	return a.waitForFeaturesConnected(ctx, accountID, features)
}

// featuresForRemoval orders an account's features for removal: protection
// features first and CLOUD_DISCOVERY last, because RSC rejects removing Cloud
// Discovery while protection features remain.
func featuresForRemoval(features []Feature) []core.Feature {
	protection := make([]core.Feature, 0, len(features))
	var discovery []core.Feature
	for _, feature := range features {
		if feature.Equal(core.FeatureCloudDiscovery) {
			discovery = append(discovery, feature.Feature)
		} else {
			protection = append(protection, feature.Feature)
		}
	}
	return append(protection, discovery...)
}

// DisableManagedAccount disables all of the account's features - protection
// features first, Cloud Discovery last - so the account can be removed. When
// deleteSnapshots is true the features' snapshots are deleted. It waits for the
// asynchronous disable jobs to finish so the CloudFormation stack can be safely
// deleted afterwards. It is a no-op if the account no longer exists.
//
// This is the first step of removal (before the CloudFormation stack is
// deleted): disable everything, then let the stack be deleted, then finalize
// with FinalizeManagedAccountDeletion.
func (a API) DisableManagedAccount(ctx context.Context, accountID uuid.UUID, deleteSnapshots bool) error {
	a.log.Print(log.Trace)

	account, err := a.AccountByID(ctx, accountID)
	if errors.Is(err, graphql.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get account: %s", err)
	}

	for _, feature := range featuresForRemoval(account.Features) {
		if err := a.disableFeature(ctx, account, feature, deleteSnapshots); err != nil {
			return fmt.Errorf("failed to disable feature %s: %s", feature.Name, err)
		}
	}

	return nil
}

// FinalizeManagedAccountDeletion removes the account from RSC after its
// CloudFormation stack has been deleted. For each feature (Cloud Discovery
// last) it prepares and finalizes the deletion. It is a no-op if the account
// has already been removed - e.g. the deleted stack's notifier already removed
// it - so it is safe to run unconditionally.
func (a API) FinalizeManagedAccountDeletion(ctx context.Context, accountID uuid.UUID) error {
	a.log.Print(log.Trace)

	account, err := a.AccountByID(ctx, accountID)
	if errors.Is(err, graphql.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get account: %s", err)
	}

	for _, feature := range featuresForRemoval(account.Features) {
		if _, err := gqlaws.Wrap(a.client).PrepareCloudAccountDeletion(ctx, accountID, feature); err != nil {
			return fmt.Errorf("failed to prepare deletion of feature %s: %s", feature.Name, err)
		}
		if err := gqlaws.Wrap(a.client).FinalizeCloudAccountDeletion(ctx, accountID, feature); err != nil {
			return fmt.Errorf("failed to finalize deletion of feature %s: %s", feature.Name, err)
		}
	}

	return nil
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
