package aws

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CloudAccount for Amazon Web Services accounts.
type CloudAccount struct {
	Cloud    string
	ID       uuid.UUID
	NativeID string
	Name     string
	Features []Feature
}

// Feature returns the specified feature from the CloudAccount's features.
func (c CloudAccount) Feature(feature core.Feature) (Feature, bool) {
	for _, f := range c.Features {
		if f.Equal(feature) {
			return f, true
		}
	}

	return Feature{}, false
}

type MappedAccount struct {
	Account struct {
		ID   uuid.UUID
		Name string
	}
}

type RoleChainingDetails struct {
	RoleArn string
	RoleUrl string
}

// Feature for Amazon Web Services accounts.
type Feature struct {
	core.Feature
	Regions             []string
	RoleArn             string
	StackArn            string
	Status              core.Status
	MappedAccounts      []MappedAccount
	RoleChainingDetails []RoleChainingDetails
}

// HasRegion returns true if the feature is enabled for the specified region.
func (f Feature) HasRegion(region string) bool {
	for _, r := range f.Regions {
		if r == region {
			return true
		}
	}

	return false
}

// AccountByID returns the account with the specified RSC cloud account ID.
func (a API) AccountByID(ctx context.Context, cloudAccountID uuid.UUID) (CloudAccount, error) {
	a.log.Print(log.Trace)

	// We need to list all accounts and filter on the cloud account id since
	// the API that looks up cloud accounts returns archived accounts too.
	accounts, err := a.Accounts(ctx, "")
	if err != nil {
		return CloudAccount{}, err
	}

	for _, account := range accounts {
		if account.ID == cloudAccountID {
			return account, nil
		}
	}

	return CloudAccount{}, fmt.Errorf("account %q %w", cloudAccountID, graphql.ErrNotFound)
}

// AccountByNativeID returns the account with the specified native ID.
// For AWS cloud accounts the native ID is the 12-digit account ID.
func (a API) AccountByNativeID(ctx context.Context, nativeID string) (CloudAccount, error) {
	a.log.Print(log.Trace)

	if !verifyAccountID(nativeID) {
		return CloudAccount{}, fmt.Errorf("invalid AWS account id: %q", nativeID)
	}

	// We need to list accounts and filter on the native id since there is
	// no API to look up an account by native id.
	accounts, err := a.Accounts(ctx, nativeID)
	if err != nil {
		return CloudAccount{}, err
	}

	for _, account := range accounts {
		if account.NativeID == nativeID {
			return account, nil
		}
	}

	return CloudAccount{}, fmt.Errorf("account %q %w", nativeID, graphql.ErrNotFound)
}

// AccountByName returns the account with the specified name.
func (a API) AccountByName(ctx context.Context, name string) (CloudAccount, error) {
	a.log.Print(log.Trace)

	accounts, err := a.Accounts(ctx, name)
	if err != nil {
		return CloudAccount{}, err
	}

	for _, account := range accounts {
		if account.Name == name {
			return account, nil
		}
	}

	return CloudAccount{}, fmt.Errorf("account %q %w", name, graphql.ErrNotFound)
}

// Accounts return all accounts matching the specified filter. The filter can
// be used to search for account native ID, account name and role ARN.
func (a API) Accounts(ctx context.Context, filter string) ([]CloudAccount, error) {
	a.log.Print(log.Trace)

	accountsWithFeatures, err := aws.Wrap(a.client).CloudAccountsWithFeatures(ctx, core.FeatureAll, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %s", err)
	}

	accounts := make([]CloudAccount, 0, len(accountsWithFeatures))
	for _, accountWithFeatures := range accountsWithFeatures {
		accounts = append(accounts, toCloudAccount(accountWithFeatures))
	}

	return accounts, nil
}

// AccountsByFeatureStatus return all accounts with the specified feature
// matching the filter. The filter can be used to search for account ID, account
// name and role arn.
func (a API) AccountsByFeatureStatus(ctx context.Context, feature core.Feature, filter string, statusFilters []core.Status) ([]CloudAccount, error) {
	a.log.Print(log.Trace)

	accountsWithFeatures, err := aws.Wrap(a.client).CloudAccountsWithFeatures(ctx, feature, filter, statusFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %s", err)
	}

	accounts := make([]CloudAccount, 0, len(accountsWithFeatures))
	for _, accountWithFeatures := range accountsWithFeatures {
		accounts = append(accounts, toCloudAccount(accountWithFeatures))
	}

	return accounts, nil
}

// UpdateAccount updates the specified feature of the account with the given ID.
// Note that the account name is not tied to a specific feature.
func (a API) UpdateAccount(ctx context.Context, cloudAccountID uuid.UUID, feature core.Feature, opts ...OptionFunc) error {
	a.log.Print(log.Trace)

	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return fmt.Errorf("failed to lookup option: %s", err)
		}
	}

	if options.name != "" {
		if err := aws.Wrap(a.client).UpdateCloudAccount(ctx, cloudAccountID, options.name); err != nil {
			return fmt.Errorf("failed to update account: %s", err)
		}
	}

	if len(options.regions) > 0 {
		if err := aws.Wrap(a.client).UpdateCloudAccountFeature(ctx, core.UpdateRegions, cloudAccountID, feature, options.regions); err != nil {
			return fmt.Errorf("failed to update account: %s", err)
		}
	}

	return nil
}

// disableFeature disables the specified account feature.
func (a API) disableFeature(ctx context.Context, account CloudAccount, feature core.Feature, deleteSnapshots bool) error {
	a.log.Print(log.Trace)

	// If the feature has not been onboarded or the feature is in the disabled
	// or connecting state, there is no need to disable the feature.
	if feature, ok := account.Feature(feature); ok {
		if feature.Status == core.StatusDisabled || feature.Status == core.StatusConnecting {
			return nil
		}
	} else {
		return nil
	}

	// Only the following features need to be disabled. Note that the Exocompute
	// feature only needs to be disabled in the CFT workflow.
	features := []core.Feature{
		core.FeatureCloudNativeProtection,
		core.FeatureRDSProtection,
		core.FeatureCloudNativeS3Protection,
		core.FeatureCloudNativeDynamoDBProtection,
		core.FeatureExocompute,
	}
	if _, ok := core.LookupFeature(features, feature); !ok {
		return nil
	}

	var jobID uuid.UUID
	if feature.Equal(core.FeatureExocompute) {
		var err error
		jobID, err = aws.Wrap(a.client).StartExocomputeDisableJob(ctx, account.ID)
		if err != nil {
			return fmt.Errorf("failed to disable exocompute feature: %s", err)
		}
	} else {
		var err error
		jobID, err = a.disableProtectionFeature(ctx, account.ID, feature, deleteSnapshots)
		if err != nil {
			return fmt.Errorf("failed to disable protection %s feature: %s", feature.Name, err)
		}
	}

	if err := core.Wrap(a.client).WaitForFeatureDisableTaskChain(ctx, jobID, func(ctx context.Context) (bool, error) {
		account, err := a.AccountByID(ctx, account.ID)
		if err != nil {
			return false, fmt.Errorf("failed to retrieve status for feature %s: %s", feature, err)
		}

		feature, ok := account.Feature(feature)
		if !ok {
			return false, fmt.Errorf("failed to retrieve status for feature %s: not found", feature)
		}
		return feature.Status == core.StatusDisabled, nil
	}); err != nil {
		return fmt.Errorf("failed to wait for task chain %s: %s", jobID, err)
	}

	return nil
}

// disableProtectionFeature disables the specific Protection Feature of the
// cloud native protection feature.
func (a API) disableProtectionFeature(ctx context.Context, cloudAccountID uuid.UUID, feature core.Feature, deleteSnapshots bool) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	var protectionFeature aws.ProtectionFeature
	switch {
	case feature.Equal(core.FeatureCloudNativeProtection):
		protectionFeature = aws.EC2
	case feature.Equal(core.FeatureRDSProtection):
		protectionFeature = aws.RDS
	case feature.Equal(core.FeatureCloudNativeS3Protection):
		protectionFeature = aws.S3
	case feature.Equal(core.FeatureCloudNativeDynamoDBProtection):
		protectionFeature = aws.DynamoDB
	case feature.Equal(core.FeatureKubernetesProtection):
		protectionFeature = aws.EKS
	default:
		return uuid.Nil, fmt.Errorf("feature %s is not a protection feature", feature.Name)
	}

	return aws.Wrap(a.client).StartNativeAccountDisableJob(ctx, cloudAccountID, protectionFeature, deleteSnapshots)
}

// SupportedFeatures returns the features supported by AWS cloud accounts.
func SupportedFeatures() []core.Feature {
	return []core.Feature{
		core.FeatureArchival,
		core.FeatureCloudNativeArchival,
		core.FeatureCloudNativeDynamoDBProtection,
		core.FeatureCloudNativeProtection,
		core.FeatureCloudNativeS3Protection,
		core.FeatureDSPMData,
		core.FeatureDSPMMetadata,
		core.FeatureExocompute,
		core.FeatureKubernetesProtection,
		core.FeatureLaminarCrossAccount,
		core.FeatureLaminarInternal,
		core.FeatureOutpost,
		core.FeatureRDSProtection,
		core.FeatureServerAndApps,
	}
}

var supportedFeatures = func() map[string]struct{} {
	m := make(map[string]struct{}, len(SupportedFeatures()))
	for _, f := range SupportedFeatures() {
		m[f.Name] = struct{}{}
	}
	return m
}()

// toCloudAccount converts a polaris/graphql/aws CloudAccountWithFeatures to a
// polaris/aws CloudAccount.
func toCloudAccount(accountWithFeatures aws.CloudAccountWithFeatures) CloudAccount {
	features := make([]Feature, 0, len(accountWithFeatures.Features))
	for _, feature := range accountWithFeatures.Features {
		if _, ok := supportedFeatures[feature.Feature]; !ok {
			continue
		}

		regions := make([]string, 0, len(feature.Regions))
		for _, region := range feature.Regions {
			regions = append(regions, region.Name())
		}
		mappedAccounts := make([]MappedAccount, 0, len(feature.MappedAccounts))
		for _, mappedAccount := range feature.MappedAccounts {
			mappedAccounts = append(mappedAccounts, MappedAccount{
				Account: struct {
					ID   uuid.UUID
					Name string
				}{
					ID:   mappedAccount.Account.ID,
					Name: mappedAccount.Account.Name,
				},
			})
		}
		roleChainingDetails := make([]RoleChainingDetails, 0, len(feature.RoleChainingDetails))
		for _, roleChainingDetail := range feature.RoleChainingDetails {
			roleChainingDetails = append(roleChainingDetails, RoleChainingDetails{
				RoleArn: roleChainingDetail.RoleArn,
				RoleUrl: roleChainingDetail.RoleUrl,
			})
		}
		features = append(features, Feature{
			Feature:             core.Feature{Name: feature.Feature, PermissionGroups: feature.PermissionGroups},
			Regions:             regions,
			RoleArn:             feature.RoleArn,
			StackArn:            feature.StackArn,
			Status:              feature.Status,
			MappedAccounts:      mappedAccounts,
			RoleChainingDetails: roleChainingDetails,
		})
	}

	return CloudAccount{
		Cloud:    string(accountWithFeatures.Account.Cloud),
		ID:       accountWithFeatures.Account.ID,
		NativeID: accountWithFeatures.Account.NativeID,
		Name:     accountWithFeatures.Account.Name,
		Features: features,
	}
}
