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

package azure

import (
	"context"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Azure permissions.
type Permissions struct {
	Actions        []string
	DataActions    []string
	NotActions     []string
	NotDataActions []string
}

// stringsDiff returns the difference between lhs and rhs, i.e. rhs subtracted
// from lhs.
func stringsDiff(lhs, rhs []string) []string {
	set := make(map[string]struct{})
	for _, s := range lhs {
		set[s] = struct{}{}
	}

	for _, s := range rhs {
		delete(set, s)
	}

	diff := make([]string, 0, len(set))
	for s := range set {
		diff = append(diff, s)
	}

	return diff
}

// Permissions returns all Azure permissions required to use the specified
// Polaris features.
func (a API) Permissions(ctx context.Context, features []core.Feature) (Permissions, error) {
	a.gql.Log().Print(log.Trace, "polaris/azure.Permissions")

	perms := Permissions{}
	for _, feature := range features {
		permConfig, err := azure.Wrap(a.gql).CloudAccountPermissionConfig(ctx, feature)
		if err != nil {
			return Permissions{}, nil
		}

		for _, perm := range permConfig.RolePermissions {
			perms.Actions = append(perms.Actions, stringsDiff(perm.IncludedActions, perms.Actions)...)
			perms.DataActions = append(perms.DataActions, stringsDiff(perm.IncludedDataActions, perms.DataActions)...)
			perms.NotActions = append(perms.NotActions, stringsDiff(perm.ExcludedActions, perms.NotActions)...)
			perms.NotDataActions = append(perms.NotDataActions, stringsDiff(perm.ExcludedDataActions, perms.NotDataActions)...)
		}
	}

	return perms, nil
}

// PermissionsUpdated notifies Polaris that the permissions for the Azure
// service principal for the Polaris cloud account with the specified id has
// been updated. The permissions should be updated when a feature has the
// status StatusMissingPermissions. Updating the permissions is done outside
// of this SDK. The features parameter is allowed to be nil. When features is
// nil all features are updated. Note that Polaris is only notified about
// features with status StatusMissingPermissions.
func (a API) PermissionsUpdated(ctx context.Context, id IdentityFunc, features []core.Feature) error {
	a.gql.Log().Print(log.Trace, "polaris/azure.PermissionsUpdated")

	featureSet := make(map[core.Feature]struct{})
	for _, feature := range features {
		featureSet[feature] = struct{}{}
	}

	account, err := a.Subscription(ctx, id, core.FeatureAll)
	if err != nil {
		return err
	}

	for _, feature := range account.Features {
		if feature.Status != core.StatusMissingPermissions {
			continue
		}

		// Check that the feature is in the feature set unless the set is
		// empty which is when all features should be updated.
		if _, ok := featureSet[feature.Name]; len(featureSet) > 0 && !ok {
			continue
		}

		err := azure.Wrap(a.gql).UpgradeCloudAccountPermissionsWithoutOAuth(ctx, account.ID, feature.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

// PermissionsUpdatedForTenantDomain notifies Polaris that the permissions
// for the Azure service principal in a tenant domain has been updated. The
// permissions should be updated when a feature has the status
// StatusMissingPermissions. Updating the permissions is done outside of the
// SDK. The features parameter is allowed to be nil. When features is nil all
// features are updated. Note that Polaris is only notified about features
// with status StatusMissingPermissions.
func (a API) PermissionsUpdatedForTenantDomain(ctx context.Context, tenantDomain string, features []core.Feature) error {
	a.gql.Log().Print(log.Trace, "polaris/azure.PermissionsUpdatedForTenantDomain")

	featureSet := make(map[core.Feature]struct{})
	for _, feature := range features {
		featureSet[feature] = struct{}{}
	}

	accounts, err := a.Subscriptions(ctx, core.FeatureAll, "")
	if err != nil {
		return err
	}

	for _, account := range accounts {
		if account.TenantDomain != tenantDomain {
			continue
		}

		for _, feature := range account.Features {
			if feature.Status != core.StatusMissingPermissions {
				continue
			}

			// Check that the feature is in the feature set unless the set is
			// empty which is when all features should be updated.
			if _, ok := featureSet[feature.Name]; len(featureSet) > 0 && !ok {
				continue
			}

			err := azure.Wrap(a.gql).UpgradeCloudAccountPermissionsWithoutOAuth(ctx, account.ID, feature.Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
