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
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// AddAccountWithIAM adds the AWS account, with specified RSC features, to RSC
// using an AWS IAM roles workflow. Returns the RSC cloud account ID of the
// added account.
// If name isn't given as an option it's derived from information in the cloud.
// The result can vary slightly depending on AWS permissions.
func (a API) AddAccountWithIAM(ctx context.Context, account AccountFunc, features []core.Feature, opts ...OptionFunc) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	if account == nil {
		return uuid.Nil, errors.New("account is not allowed to be nil")
	}
	config, err := account(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to lookup account: %s", err)
	}
	if config.config != nil {
		return uuid.Nil, errors.New("account config is not used by the IAM roles workflow")
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
	cloudAccount, err := a.Account(ctx, AccountID(config.id), core.FeatureAll)
	if err != nil && !errors.Is(err, graphql.ErrNotFound) {
		return uuid.Nil, fmt.Errorf("failed to get account: %s", err)
	}
	if err == nil {
		config.name = cloudAccount.Name
	}

	accountInit := aws.CloudAccountInitiate{
		FeatureVersions: []aws.FeatureVersion{},
	}
	if err := aws.Wrap(a.client).FinalizeCloudAccountProtection(ctx, config.cloud, config.id, config.name, features, options.regions, accountInit); err != nil {
		return uuid.Nil, fmt.Errorf("failed to add account: %s", err)
	}

	// If the RSC cloud account did not exist prior, we retrieve the RSC cloud
	// account ID.
	if cloudAccount.ID == uuid.Nil {
		cloudAccount, err = a.Account(ctx, AccountID(config.id), core.FeatureAll)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to get account: %s", err)
		}
	}

	return cloudAccount.ID, nil
}

// RemoveAccountWithIAM removes the RSC features from the account with the
// specified ID, onboarded with the onboarded with the AWS IAM roles workflow.
// If a Cloud Native Protection feature is being removed and deleteSnapshots is
// true, the snapshots are deleted otherwise they are kept.
func (a API) RemoveAccountWithIAM(ctx context.Context, account AccountFunc, features []core.Feature, deleteSnapshots bool) error {
	a.log.Print(log.Trace)

	if account == nil {
		return errors.New("account is not allowed to be nil")
	}
	config, err := account(ctx)
	if err != nil {
		return fmt.Errorf("failed to lookup account: %s", err)
	}

	cloudAccount, err := a.Account(ctx, AccountID(config.id), core.FeatureAll)
	if err != nil {
		return fmt.Errorf("failed to get account: %s", err)
	}

	// Check that the account has all the features that are going to be removed.
	for _, feature := range features {
		if _, ok := cloudAccount.Feature(feature); !ok {
			return fmt.Errorf("feature %s %w", feature, graphql.ErrNotFound)
		}
	}

	return a.removeAccountWithIAM(ctx, cloudAccount, features, deleteSnapshots)
}

func (a API) removeAccountWithIAM(ctx context.Context, account CloudAccount, features []core.Feature, deleteSnapshots bool) error {
	a.log.Print(log.Trace)

	for _, feature := range features {
		// Exocompute does not need to be disabled with the IAM roles workflow.
		if feature.Equal(core.FeatureExocompute) {
			continue
		}
		if err := a.disableFeature(ctx, account, feature, deleteSnapshots); err != nil {
			return fmt.Errorf("failed to disable feature %s: %s", feature, err)
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

// TrustPolicyMap maps the role key to the trust policy.
type TrustPolicyMap map[string]string

// TrustPolicies returns the trust policies required by RSC for the specified
// features. If the external ID is empty, RSC will generate an external ID.
// The same endpoint is used both for reading and writing the trust policy,
// a side effect of this is that the first call always set the trust policy.
// Once the trust policy has been set, it cannot be changed.
// If the account cannot be found, graphql.ErrNotFound is returned.
func (a API) TrustPolicies(ctx context.Context, cloud aws.Cloud, accountID uuid.UUID, features []core.Feature, externalID string) (TrustPolicyMap, error) {
	a.log.Print(log.Trace)

	// We need to look up the account to obtain the AWS native account ID.
	// The call returns graphql.NotFound if the cloud account isn't found.
	account, err := a.AccountByID(ctx, core.FeatureAll, accountID)
	if err != nil {
		return nil, err
	}

	policies, err := aws.Wrap(a.client).TrustPolicy(ctx, cloud, features, []aws.TrustPolicyAccount{{
		ID:         account.NativeID,
		ExternalID: externalID,
	}})
	if err != nil {
		return nil, fmt.Errorf("failed to get trust policies: %s", err)
	}
	if len(policies) != 1 {
		return nil, errors.New("expected trust policies for a single account")
	}

	trustPolicies := make(TrustPolicyMap)
	for _, artifact := range policies[0].Artifacts {
		if msg := artifact.ErrorMessage; msg != "" {
			return nil, fmt.Errorf("failed to get trust policies: %s", msg)
		}
		artifact.ExternalArtifactKey = strings.TrimSuffix(artifact.ExternalArtifactKey, roleArnSuffix)
		trustPolicies[artifact.ExternalArtifactKey] = artifact.TrustPolicyDoc
	}

	return trustPolicies, nil
}
