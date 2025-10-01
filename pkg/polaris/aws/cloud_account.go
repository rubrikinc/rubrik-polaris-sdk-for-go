package aws

import (
	"context"
	"errors"
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

// Account returns the account with specified id and feature.
func (a API) Account(ctx context.Context, id IdentityFunc, feature core.Feature) (CloudAccount, error) {
	a.log.Print(log.Trace)

	if id == nil {
		return CloudAccount{}, errors.New("id is not allowed to be nil")
	}
	identity, err := id(ctx)
	if err != nil {
		return CloudAccount{}, fmt.Errorf("failed to lookup identity: %s", err)
	}

	if identity.internal {
		cloudAccountID, err := uuid.Parse(identity.id)
		if err != nil {
			return CloudAccount{}, fmt.Errorf("failed to parse identity: %s", err)
		}

		return a.AccountByID(ctx, feature, cloudAccountID)
	}

	return a.AccountByNativeID(ctx, feature, identity.id)
}

// AccountByID returns the account with the specified feature and RSC cloud
// account ID.
func (a API) AccountByID(ctx context.Context, feature core.Feature, id uuid.UUID) (CloudAccount, error) {
	a.log.Print(log.Trace)

	// We need to list all accounts and filter on the cloud account id since
	// the API that looks up cloud accounts returns archived accounts too.
	accounts, err := a.Accounts(ctx, feature, "")
	if err != nil {
		return CloudAccount{}, fmt.Errorf("failed to get account by cloud account id: %s", err)
	}

	for _, account := range accounts {
		if account.ID == id {
			return account, nil
		}
	}

	return CloudAccount{}, fmt.Errorf("account %q %w", id, graphql.ErrNotFound)
}

// AccountByNativeID returns the account with the specified feature and native
// ID.
func (a API) AccountByNativeID(ctx context.Context, feature core.Feature, nativeID string) (CloudAccount, error) {
	a.log.Print(log.Trace)

	// We need to list accounts and filter on the native id since there is
	// no API to look up an account by native id.
	accounts, err := a.Accounts(ctx, feature, nativeID)
	if err != nil {
		return CloudAccount{}, fmt.Errorf("failed to get account by native id: %s", err)
	}

	for _, account := range accounts {
		if account.NativeID == nativeID {
			return account, nil
		}
	}

	return CloudAccount{}, fmt.Errorf("account %q %w", nativeID, graphql.ErrNotFound)
}

// AccountByName returns the account with the specified feature and name.
func (a API) AccountByName(ctx context.Context, feature core.Feature, name string) (CloudAccount, error) {
	a.log.Print(log.Trace)

	accounts, err := a.Accounts(ctx, feature, name)
	if err != nil {
		return CloudAccount{}, fmt.Errorf("failed to get account by name: %s", err)
	}

	for _, account := range accounts {
		if account.Name == name {
			return account, nil
		}
	}

	return CloudAccount{}, fmt.Errorf("account %q %w", name, graphql.ErrNotFound)
}

// Accounts return all accounts with the specified feature matching the filter.
// The filter can be used to search for account id, account name and role arn.
func (a API) Accounts(ctx context.Context, feature core.Feature, filter string) ([]CloudAccount, error) {
	a.log.Print(log.Trace)

	accountsWithFeatures, err := aws.Wrap(a.client).CloudAccountsWithFeatures(ctx, feature, filter, nil)
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
// matching the filter. The filter can be used to search for account id, account
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
func (a API) UpdateAccount(ctx context.Context, id uuid.UUID, feature core.Feature, opts ...OptionFunc) error {
	a.log.Print(log.Trace)

	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return fmt.Errorf("failed to lookup option: %s", err)
		}
	}

	if options.name != "" {
		if err := aws.Wrap(a.client).UpdateCloudAccount(ctx, id, options.name); err != nil {
			return fmt.Errorf("failed to update account: %s", err)
		}
	}

	if len(options.regions) > 0 {
		if err := aws.Wrap(a.client).UpdateCloudAccountFeature(ctx, core.UpdateRegions, id, feature, options.regions); err != nil {
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
	if !core.ContainsFeature(features, feature) {
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
		account, err := a.Account(ctx, CloudAccountID(account.ID), feature)
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
		protectionFeature = aws.DYNAMODB
	default:
		return uuid.Nil, fmt.Errorf("feature %s is not a protection feature", feature.Name)
	}

	return aws.Wrap(a.client).StartNativeAccountDisableJob(ctx, cloudAccountID, protectionFeature, deleteSnapshots)
}

var supportedFeatures = map[string]struct{}{
	core.FeatureArchival.Name:                      {},
	core.FeatureCloudNativeArchival.Name:           {},
	core.FeatureCloudNativeDynamoDBProtection.Name: {},
	core.FeatureCloudNativeProtection.Name:         {},
	core.FeatureCloudNativeS3Protection.Name:       {},
	core.FeatureDSPMData.Name:                      {},
	core.FeatureDSPMMetadata.Name:                  {},
	core.FeatureExocompute.Name:                    {},
	core.FeatureLaminarCrossAccount.Name:           {},
	core.FeatureLaminarInternal.Name:               {},
	core.FeatureOutpost.Name:                       {},
	core.FeatureRDSProtection.Name:                 {},
	core.FeatureServerAndApps.Name:                 {},
}

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

// toNativeID returns the AWS account id for the specified identity. If the
// identity is an AWS account id no remote endpoint is called.
func (a API) toNativeID(ctx context.Context, id IdentityFunc) (string, error) {
	a.log.Print(log.Trace)

	if id == nil {
		return "", errors.New("id is not allowed to be nil")
	}
	identity, err := id(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to lookup identity: %s", err)
	}

	if !identity.internal {
		return identity.id, nil
	}

	uid, err := uuid.Parse(identity.id)
	if err != nil {
		return "", fmt.Errorf("failed to parse identity: %s", err)
	}

	accountWithFeatures, err := aws.Wrap(a.client).CloudAccountWithFeatures(ctx, uid, core.FeatureAll)
	if err != nil {
		return "", fmt.Errorf("failed to get account: %s", err)
	}

	return accountWithFeatures.Account.NativeID, nil
}
