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

// Package aws provides a high level interface to the AWS part of the RSC
// platform.
package aws

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API for AWS account management.
type API struct {
	client *graphql.Client
	log    log.Logger
}

// Deprecated: use Wrap instead.
func NewAPI(gql *graphql.Client) API {
	return API{client: gql, log: gql.Log()}
}

// Wrap the RSC client in the aws API.
func Wrap(client *polaris.Client) API {
	return API{client: client.GQL, log: client.GQL.Log()}
}

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

// Feature for Amazon Web Services accounts.
type Feature struct {
	core.Feature
	Regions  []string
	RoleArn  string
	StackArn string
	Status   core.Status
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

// toCloudAccountID returns the RSC cloud account id for the specified identity.
// If the identity is an RSC cloud account id no remote endpoint is called.
func (a API) toCloudAccountID(ctx context.Context, id IdentityFunc) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	if id == nil {
		return uuid.Nil, errors.New("id is not allowed to be nil")
	}
	identity, err := id(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to lookup identity: %s", err)
	}

	if identity.internal {
		id, err := uuid.Parse(identity.id)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to parse identity: %s", err)
		}

		return id, nil
	}

	accountsWithFeatures, err := aws.Wrap(a.client).CloudAccountsWithFeatures(ctx, core.FeatureCloudNativeProtection, identity.id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get account: %s", err)
	}

	// Find the exact match.
	for _, accountWithFeatures := range accountsWithFeatures {
		if accountWithFeatures.Account.NativeID == identity.id {
			return accountWithFeatures.Account.ID, nil
		}
	}

	return uuid.Nil, fmt.Errorf("account %w", graphql.ErrNotFound)
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

// toCloudAccount converts a polaris/graphql/aws CloudAccountWithFeatures to a
// polaris/aws CloudAccount.
func toCloudAccount(accountWithFeatures aws.CloudAccountWithFeatures) CloudAccount {
	features := make([]Feature, 0, len(accountWithFeatures.Features))
	for _, feature := range accountWithFeatures.Features {
		features = append(features, Feature{
			Feature:  core.Feature{Name: feature.Feature, PermissionGroups: feature.PermissionGroups},
			Regions:  aws.FormatRegions(feature.Regions),
			RoleArn:  feature.RoleArn,
			StackArn: feature.StackArn,
			Status:   feature.Status,
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

		// We need to list all accounts and filter on the cloud account id since
		// the API that looks up cloud accounts returns archived accounts too.
		accountsWithFeatures, err := aws.Wrap(a.client).CloudAccountsWithFeatures(ctx, feature, "")
		if err != nil {
			return CloudAccount{}, fmt.Errorf("failed to get account: %s", err)
		}

		// Find the exact match.
		for _, accountWithFeatures := range accountsWithFeatures {
			if accountWithFeatures.Account.ID == cloudAccountID {
				return toCloudAccount(accountWithFeatures), nil
			}
		}
	} else {
		// We need to list accounts and filter on the native id since there is
		// no API to look up an account by native id.
		accountsWithFeatures, err := aws.Wrap(a.client).CloudAccountsWithFeatures(ctx, feature, identity.id)
		if err != nil {
			return CloudAccount{}, fmt.Errorf("failed to get account: %s", err)
		}

		// Find the exact match.
		for _, accountWithFeatures := range accountsWithFeatures {
			if accountWithFeatures.Account.NativeID == identity.id {
				return toCloudAccount(accountWithFeatures), nil
			}
		}
	}

	return CloudAccount{}, fmt.Errorf("account %w", graphql.ErrNotFound)
}

// Accounts return all accounts with the specified feature matching the filter.
// The filter can be used to search for account id, account name and role arn.
func (a API) Accounts(ctx context.Context, feature core.Feature, filter string) ([]CloudAccount, error) {
	a.log.Print(log.Trace)

	accountsWithFeatures, err := aws.Wrap(a.client).CloudAccountsWithFeatures(ctx, feature, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %s", err)
	}

	accounts := make([]CloudAccount, 0, len(accountsWithFeatures))
	for _, accountWithFeatures := range accountsWithFeatures {
		accounts = append(accounts, toCloudAccount(accountWithFeatures))
	}

	return accounts, nil
}

// AddAccount adds the AWS account to RSC for the given features. Returns the
// RSC cloud account id of the added account. If name isn't given as an option
// it's derived from information in the cloud. The result can vary slightly
// depending on AWS permissions.
//
// If adding the account fails due to permission problems when creating the
// CloudFormation stack, it's safe to call AddAccount again with the same
// parameters after the permission problems have been resolved.
func (a API) AddAccount(ctx context.Context, account AccountFunc, features []core.Feature, opts ...OptionFunc) (uuid.UUID, error) {
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

	// If there already is an RSC cloud account for the given AWS account we use
	// the same account name when adding the feature. RSC does not allow the
	// name to change between features.
	akkount, err := a.Account(ctx, AccountID(config.id), core.FeatureAll)
	if err == nil {
		config.name = akkount.Name
	}
	if err != nil && !errors.Is(err, graphql.ErrNotFound) {
		return uuid.Nil, fmt.Errorf("failed to get account: %s", err)
	}

	if config.config != nil {
		err = a.addAccountWithCFT(ctx, features, config, options)
	} else {
		err = a.addAccount(ctx, features, config, options)
	}
	if err != nil {
		return uuid.Nil, err
	}

	// If the RSC cloud account did not exist prior we retrieve the RSC cloud
	// account id.
	if akkount.ID == uuid.Nil {
		akkount, err = a.Account(ctx, AccountID(config.id), core.FeatureAll)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to get account: %s", err)
		}
	}

	return akkount.ID, nil
}

func (a API) addAccount(ctx context.Context, features []core.Feature, config account, options options) error {
	a.log.Print(log.Trace)

	accountInit := aws.CloudAccountInitiate{
		CloudFormationURL: "",
		ExternalID:        "",
		FeatureVersions:   []aws.FeatureVersion{},
		StackName:         "",
		TemplateURL:       "",
	}

	err := aws.Wrap(a.client).FinalizeCloudAccountProtection(ctx, config.cloud, config.id, config.name, features, options.regions, accountInit)
	if err != nil {
		return fmt.Errorf("failed to add account: %s", err)
	}

	return nil
}

func (a API) addAccountWithCFT(ctx context.Context, features []core.Feature, config account, options options) error {
	a.log.Print(log.Trace)

	accountInit, err := aws.Wrap(a.client).ValidateAndCreateCloudAccount(ctx, config.id, config.name, features)
	if err != nil {
		return fmt.Errorf("failed to validate account: %s", err)
	}

	err = aws.Wrap(a.client).FinalizeCloudAccountProtection(ctx, config.cloud, config.id, config.name, features, options.regions, accountInit)
	if err != nil {
		return fmt.Errorf("failed to add account: %s", err)
	}

	err = awsUpdateStack(ctx, a.client.Log(), *config.config, accountInit.StackName, accountInit.TemplateURL)
	if err != nil {
		return fmt.Errorf("failed to update CloudFormation stack: %s", err)
	}

	return nil
}

// RemoveAccount removes the account with the specified id from RSC for the
// given feature. If the Cloud Native Protection feature is being removed and
// deleteSnapshots is true the snapshots are deleted otherwise they are kept.
// Note that removing the Cloud Native Protection feature will also remove the
// Exocompute feature.
func (a API) RemoveAccount(ctx context.Context, account AccountFunc, features []core.Feature, deleteSnapshots bool) error {
	a.log.Print(log.Trace)

	if account == nil {
		return errors.New("account is not allowed to be nil")
	}
	config, err := account(ctx)
	if err != nil {
		return fmt.Errorf("failed to lookup account: %s", err)
	}

	akkount, err := a.Account(ctx, AccountID(config.id), core.FeatureAll)
	if err != nil {
		return fmt.Errorf("failed to get account: %s", err)
	}

	// Check that the account has all the features that are going to be removed.
	for _, feature := range features {
		if _, ok := akkount.Feature(feature); !ok {
			return fmt.Errorf("feature %s %w", feature, graphql.ErrNotFound)
		}
	}

	if config.config != nil {
		for _, feature := range features {
			if err := a.removeAccountWithCFT(ctx, config, akkount, feature, deleteSnapshots); err != nil {
				return err
			}
		}
		return nil
	}

	return a.removeAccount(ctx, akkount, features, deleteSnapshots)
}

func (a API) removeAccount(ctx context.Context, account CloudAccount, features []core.Feature, deleteSnapshots bool) error {
	a.log.Print(log.Trace)

	for _, feature := range features {
		if err := a.disableFeature(ctx, account, feature, deleteSnapshots); err != nil {
			return fmt.Errorf("failed to disable native account: %s", err)
		}
	}

	results, err := aws.Wrap(a.client).DeleteCloudAccountWithoutCft(ctx, account.NativeID, features)
	if err != nil {
		return fmt.Errorf("failed to delete account: %s", err)
	}
	var sb strings.Builder
	for _, result := range results {
		if !result.Success {
			sb.WriteString(", ")
			sb.WriteString(result.Feature)
		}
	}
	if sb.Len() > 0 {
		return fmt.Errorf("failed to delete features: %s", sb.String()[2:])
	}

	return nil
}

func (a API) removeAccountWithCFT(ctx context.Context, config account, account CloudAccount, feature core.Feature, deleteSnapshots bool) error {
	a.log.Print(log.Trace)

	if err := a.disableFeature(ctx, account, feature, deleteSnapshots); err != nil {
		return fmt.Errorf("failed to disable native account: %s", err)
	}

	cfmURL, err := aws.Wrap(a.client).PrepareCloudAccountDeletion(ctx, account.ID, feature)
	if err != nil {
		return fmt.Errorf("failed to prepare to delete account: %s", err)
	}

	if cfmURL != "" {
		if strings.Contains(cfmURL, "#/stack/update") {
			i := strings.LastIndex(cfmURL, "#/stack/update") + 1
			if i == 0 {
				return errors.New("CloudFormation url does not contain #/stack/update")
			}

			u, err := url.Parse(cfmURL[i:])
			if err != nil {
				return fmt.Errorf("failed to parse CloudFormation url: %s", err)
			}
			stackID := u.Query().Get("stackId")
			tmplURL := u.Query().Get("templateURL")

			err = awsUpdateStack(ctx, a.client.Log(), *config.config, stackID, tmplURL)
			if err != nil {
				return fmt.Errorf("failed to update CloudFormation stack: %s", err)
			}
		} else {
			i := strings.LastIndex(cfmURL, "#/stack/detail") + 1
			if i == 0 {
				return errors.New("CloudFormation url does not contain #/stack/detail")
			}

			u, err := url.Parse(cfmURL[i:])
			if err != nil {
				return fmt.Errorf("failed to parse CloudFormation url: %s", err)
			}
			stackID := u.Query().Get("stackId")

			err = awsDeleteStack(ctx, a.client.Log(), *config.config, stackID)
			if err != nil {
				return fmt.Errorf("failed to delete CloudFormation stack: %s", err)
			}
		}
	}

	err = aws.Wrap(a.client).FinalizeCloudAccountDeletion(ctx, account.ID, feature)
	if err != nil {
		return fmt.Errorf("failed to delete account: %s", err)
	}

	return nil
}

func (a API) disableFeature(ctx context.Context, account CloudAccount, feature core.Feature, deleteSnapshots bool) error {
	a.log.Print(log.Trace)

	rmFeature, _ := account.Feature(feature)
	if !featureNeedsToBeDisable(rmFeature) {
		return nil
	}

	switch {
	case rmFeature.Equal(core.FeatureCloudNativeProtection):
		return a.disableNativeAccount(ctx, account.ID, aws.EC2, deleteSnapshots)

	case rmFeature.Equal(core.FeatureRDSProtection):
		return a.disableNativeAccount(ctx, account.ID, aws.RDS, deleteSnapshots)

	case rmFeature.Equal(core.FeatureExocompute):
		jobID, err := aws.Wrap(a.client).StartExocomputeDisableJob(ctx, account.ID)
		if err != nil {
			return fmt.Errorf("failed to disable native account: %s", err)
		}

		state, err := core.Wrap(a.client).WaitForTaskChain(ctx, jobID, 10*time.Second)
		if err != nil {
			return fmt.Errorf("failed to wait for taskchain: %s", err)
		}
		if state != core.TaskChainSucceeded {
			return fmt.Errorf("taskchain failed: jobID=%v, state=%v", jobID, state)
		}
	}

	return nil
}

// featureNeedsToBeDisable returns true if the specified feature needs to be
// disabled before being removed. Note, a feature in the connecting state can be
// removed without being disabling first.
func featureNeedsToBeDisable(feature Feature) bool {
	return feature.Status != core.StatusDisabled && feature.Status != core.StatusConnecting
}

func (a API) disableNativeAccount(ctx context.Context, id uuid.UUID, protectionFeature aws.ProtectionFeature, deleteSnapshots bool) error {
	jobID, err := aws.Wrap(a.client).StartNativeAccountDisableJob(ctx, id, protectionFeature, deleteSnapshots)
	if err != nil {
		return fmt.Errorf("failed to disable native account: %s", err)
	}

	state, err := core.Wrap(a.client).WaitForTaskChain(ctx, jobID, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to wait for task chain: %s", err)
	}
	if state != core.TaskChainSucceeded {
		return fmt.Errorf("taskchain failed: jobID=%v, state=%v", jobID, state)
	}

	return nil
}

// UpdateAccount updates the account with the specified id and feature. Note
// that account name is not tied to a specific feature.
func (a API) UpdateAccount(ctx context.Context, id IdentityFunc, feature core.Feature, opts ...OptionFunc) error {
	a.log.Print(log.Trace)

	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return fmt.Errorf("failed to lookup option: %s", err)
		}
	}

	accountID, err := a.toCloudAccountID(ctx, id)
	if err != nil {
		return err
	}

	if options.name != "" {
		if err := aws.Wrap(a.client).UpdateCloudAccount(ctx, accountID, options.name); err != nil {
			return fmt.Errorf("failed to update account: %s", err)
		}
	}

	if len(options.regions) > 0 {
		if err := aws.Wrap(a.client).UpdateCloudAccountFeature(ctx, core.UpdateRegions, accountID, feature, options.regions); err != nil {
			return fmt.Errorf("failed to update account: %s", err)
		}
	}

	return nil
}

const (
	roleArnSuffix         = "_ROLE_ARN"
	instanceProfileSuffix = "_INSTANCE_PROFILE"
)

// Artifacts returns the artifacts, instance profiles and roles, required by RSC
// for the specified features.
func (a API) Artifacts(ctx context.Context, cloud string, features []core.Feature) ([]string, []string, error) {
	a.log.Print(log.Trace)

	c, err := aws.ParseCloud(cloud)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse cloud: %s", err)
	}

	artifacts, err := aws.Wrap(a.client).AllPermissionPolicies(ctx, c, features, "")
	if err != nil {
		return nil, nil, err
	}
	var profiles, roles []string
	for _, artifact := range artifacts {
		key := artifact.ArtifactKey
		switch {
		case strings.HasSuffix(key, instanceProfileSuffix):
			profiles = append(profiles, strings.TrimSuffix(key, instanceProfileSuffix))
		case strings.HasSuffix(key, roleArnSuffix):
			roles = append(roles, strings.TrimSuffix(key, roleArnSuffix))
		default:
			a.log.Printf(log.Info, "Ignoring artifact: %s", key)
		}
	}
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i] < profiles[j]
	})
	sort.Slice(roles, func(i, j int) bool {
		return roles[i] < roles[j]
	})

	return profiles, roles, nil
}

// AccountArtifacts returns the artifacts added to the cloud account.
func (a API) AccountArtifacts(ctx context.Context, id IdentityFunc) (map[string]string, map[string]string, error) {
	a.log.Print(log.Trace)

	nativeID, err := a.toNativeID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	artifacts, err := aws.Wrap(a.client).ArtifactsToDelete(ctx, nativeID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get artifacts registered with account: %s", err)
	}

	instanceProfiles := make(map[string]string)
	roles := make(map[string]string)
	var skipped []string
	for _, artifact := range artifacts {
		for _, artifact := range artifact.ArtifactsToDelete {
			switch {
			case strings.HasSuffix(artifact.ExternalArtifactKey, instanceProfileSuffix):
				key := strings.TrimSuffix(artifact.ExternalArtifactKey, instanceProfileSuffix)
				instanceProfiles[key] = artifact.ExternalArtifactValue
			case strings.HasSuffix(artifact.ExternalArtifactKey, roleArnSuffix):
				key := strings.TrimSuffix(artifact.ExternalArtifactKey, roleArnSuffix)
				roles[key] = artifact.ExternalArtifactValue
			default:
				skipped = append(skipped, artifact.ExternalArtifactKey)
			}
		}
	}
	a.log.Printf(log.Debug, "Skipped the following artifacts: %v", skipped)

	return instanceProfiles, roles, nil
}

// AddAccountArtifacts adds the specified artifacts, instance profiles and
// roles, to the cloud account.
func (a API) AddAccountArtifacts(ctx context.Context, id IdentityFunc, features []core.Feature, instanceProfiles map[string]string, roles map[string]string) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	account, err := a.Account(ctx, id, core.FeatureAll)
	if err != nil {
		return uuid.Nil, err
	}

	externalArtifacts := make([]aws.ExternalArtifact, 0, len(instanceProfiles)+len(roles))
	for key, value := range instanceProfiles {
		if !strings.HasSuffix(key, instanceProfileSuffix) {
			key = key + instanceProfileSuffix
		}
		externalArtifacts = append(externalArtifacts, aws.ExternalArtifact{
			ExternalArtifactKey:   key,
			ExternalArtifactValue: value,
		})
	}
	for key, value := range roles {
		if !strings.HasSuffix(key, roleArnSuffix) {
			key = key + roleArnSuffix
		}
		externalArtifacts = append(externalArtifacts, aws.ExternalArtifact{
			ExternalArtifactKey:   key,
			ExternalArtifactValue: value,
		})
	}

	// RegisterFeatureArtifacts fails with an error referring to RBK30300003
	// if an instance profile or a role passed in as an artifact is not yet
	// available. This can happen if the call to register the artifacts is
	// performed right after the call to create the instance profile or the role
	// returns. When this happens we wait 5 seconds before trying again. After
	// 30 seconds we abort.
	now := time.Now()
	var mappings []aws.NativeIDToRSCIDMapping
	for {
		mappings, err = aws.Wrap(a.client).RegisterFeatureArtifacts(ctx, aws.Cloud(account.Cloud), []aws.AccountFeatureArtifact{{
			NativeID:  account.NativeID,
			Features:  core.FeatureNames(features),
			Artifacts: externalArtifacts,
		}})
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to register feature artifacts: %s", err)
		}
		if len(mappings) != 1 {
			return uuid.Nil, errors.New("expected account mappings for a single account")
		}
		if msg := mappings[0].Message; msg == "" || !strings.Contains(msg, "RBK30300003") {
			break
		}
		if time.Since(now) > 30*time.Second {
			break
		}
		time.Sleep(5 * time.Second)
	}
	if msg := mappings[0].Message; msg != "" {
		return uuid.Nil, fmt.Errorf("failed to register feature artifacts: %s", msg)
	}

	accountID, err := uuid.Parse(mappings[0].CloudAccountID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse cloud account id: %s", err)
	}

	return accountID, nil
}

// TrustPolicies returns the trust policies required by RSC for the specified
// features. If the external ID is empty, RSC will generate an external ID.
func (a API) TrustPolicies(ctx context.Context, id IdentityFunc, features []core.Feature, externalID string) (map[string]string, error) {
	a.log.Print(log.Trace)

	account, err := a.Account(ctx, id, core.FeatureAll)
	if err != nil {
		return nil, err
	}

	policies, err := aws.Wrap(a.client).TrustPolicy(ctx, aws.Cloud(account.Cloud), features, []aws.TrustPolicyAccount{{
		ID:         account.NativeID,
		ExternalID: externalID,
	}})
	if err != nil {
		return nil, fmt.Errorf("failed to get trust policies: %s", err)
	}
	if len(policies) != 1 {
		return nil, fmt.Errorf("expected trust policies for a single account")
	}

	trustPolicies := make(map[string]string)
	for _, artifact := range policies[0].Artifacts {
		if msg := artifact.ErrorMessage; msg != "" {
			return nil, fmt.Errorf("failed to get trust policies: %s", msg)
		}
		artifact.ExternalArtifactKey = strings.TrimSuffix(artifact.ExternalArtifactKey, roleArnSuffix)
		trustPolicies[artifact.ExternalArtifactKey] = artifact.TrustPolicyDoc
	}

	return trustPolicies, nil
}
