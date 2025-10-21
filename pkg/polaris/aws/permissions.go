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
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CustomerManagedPolicy represents a policy that is managed by the customer.
type CustomerManagedPolicy struct {
	Artifact string
	Feature  core.Feature
	Name     string
	Policy   string
}

func (policy CustomerManagedPolicy) lessThan(other CustomerManagedPolicy) bool {
	return policy.Artifact < other.Artifact || policy.Feature.Name < other.Feature.Name || policy.Name < other.Name
}

// ManagedPolicy represents a policy that is managed by AWS.
type ManagedPolicy struct {
	Artifact string
	Name     string
}

func (policy ManagedPolicy) lessThan(other ManagedPolicy) bool {
	return policy.Artifact < other.Artifact || policy.Name < other.Name
}

// Permissions returns the policies required by RSC for the specified features.
func (a API) Permissions(ctx context.Context, cloud string, features []core.Feature, ec2RecoveryRolePath string) ([]CustomerManagedPolicy, []ManagedPolicy, error) {
	a.log.Print(log.Trace)

	c, err := aws.ParseCloud(cloud)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse cloud: %s", err)
	}

	artifacts, err := aws.Wrap(a.client).AllPermissionPolicies(ctx, c, features, ec2RecoveryRolePath)
	if err != nil {
		return nil, nil, err
	}

	var customerPolicies []CustomerManagedPolicy
	var managedPolicies []ManagedPolicy
	for _, artifact := range artifacts {
		for _, policy := range artifact.CustomerManagedPolicies {
			customerPolicies = append(customerPolicies, CustomerManagedPolicy{
				Artifact: strings.TrimSuffix(artifact.ArtifactKey, roleArnSuffix),
				Feature:  core.Feature{Name: policy.Feature},
				Name:     policy.PolicyName,
				Policy:   policy.PolicyDocument,
			})
		}
		for _, policy := range artifact.ManagedPolicies {
			managedPolicies = append(managedPolicies, ManagedPolicy{
				Artifact: strings.TrimSuffix(artifact.ArtifactKey, roleArnSuffix),
				Name:     policy,
			})
		}
	}
	sort.Slice(customerPolicies, func(i, j int) bool {
		return customerPolicies[i].lessThan(customerPolicies[j])
	})
	sort.Slice(managedPolicies, func(i, j int) bool {
		return managedPolicies[i].lessThan(managedPolicies[j])
	})

	return customerPolicies, managedPolicies, nil
}

// UpdatePermissions updates the permissions of the CloudFormation stack in
// AWS.
func (a API) UpdatePermissions(ctx context.Context, account AccountFunc, features []core.Feature) error {
	a.log.Print(log.Trace)

	if account == nil {
		return errors.New("account is not allowed to be nil")
	}
	config, err := account(ctx)
	if err != nil {
		return fmt.Errorf("failed to lookup account: %v", err)
	}

	if config.config == nil {
		return fmt.Errorf("only applicable to cloud accounts using cft")
	}

	akkount, err := a.AccountByNativeID(ctx, config.id)
	if err != nil {
		return fmt.Errorf("failed to get account: %v", err)
	}

	cfmURL, tmplURL, err := aws.Wrap(a.client).PrepareFeatureUpdateForAwsCloudAccount(ctx, akkount.ID, features)
	if err != nil {
		return fmt.Errorf("failed to update account: %v", err)
	}

	// Extract stack id/name from returned CloudFormationURL.
	i := strings.LastIndex(cfmURL, "#/stack/update") + 1
	if i == 0 {
		return errors.New("CloudFormation url does not contain #/stack/update")
	}

	u, err := url.Parse(cfmURL[i:])
	if err != nil {
		return fmt.Errorf("failed to parse CloudFormation url: %v", err)
	}
	stackID := u.Query().Get("stackId")

	err = awsUpdateStack(ctx, a.client.Log(), *config.config, stackID, tmplURL)
	if err != nil {
		return fmt.Errorf("failed to update CloudFormation stack: %v", err)
	}

	return nil
}

// PermissionsUpdated notifies RSC that the AWS roles for the RSC cloud account
// with the specified ID has been updated.
//
// The permissions should be updated when a feature has the status
// StatusMissingPermissions. Updating the permissions is done outside of this
// SDK. The feature parameter is allowed to be nil. When features are nil, all
// features are updated. Note that RSC is only notified about features with
// status StatusMissingPermissions.
func (a API) PermissionsUpdated(ctx context.Context, cloudAccountID uuid.UUID, features []core.Feature) error {
	a.log.Print(log.Trace)

	featureSet := make(map[string]struct{})
	for _, feature := range features {
		featureSet[feature.Name] = struct{}{}
	}

	account, err := a.AccountByID(ctx, cloudAccountID)
	if err != nil {
		return fmt.Errorf("failed to read cloud account %s: %s", cloudAccountID, err)
	}

	for _, feature := range account.Features {
		if feature.Status != core.StatusMissingPermissions {
			continue
		}

		// Check that the feature is in the feature set unless the set is empty
		// which is when all features should be updated.
		if _, ok := featureSet[feature.Name]; len(featureSet) > 0 && !ok {
			continue
		}

		if err := aws.Wrap(a.client).UpgradeCloudAccountFeaturesWithoutCft(ctx, cloudAccountID, []core.Feature{feature.Feature}); err != nil {
			return fmt.Errorf("failed to notify RSC about upgraded permissions for cloud account %s: %s", cloudAccountID, err)
		}
	}

	return nil
}
