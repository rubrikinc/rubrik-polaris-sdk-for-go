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
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

type Scope int

const (
	// ScopeLegacy provides backwards compatibility with how permissions worked
	// before scoped permissions were introduced.
	ScopeLegacy Scope = iota

	// ScopeSubscription represents the subscription level permissions.
	ScopeSubscription

	// ScopeResourceGroup represents the resource group level permissions.
	ScopeResourceGroup
)

// Permissions for Azure.
type Permissions struct {
	Actions        []string
	DataActions    []string
	NotActions     []string
	NotDataActions []string
}

// PermissionGroupWithVersion represents a permission group with a specific
// version.
type PermissionGroupWithVersion struct {
	Name    string
	Version int
}

// Note, permissions must be sorted in alphabetical order.
func (p *Permissions) addPermissions(perm Permissions) {
	p.Actions = append(p.Actions, perm.Actions...)
	slices.Sort(p.Actions)
	p.Actions = slices.Compact(p.Actions)

	p.DataActions = append(p.DataActions, perm.DataActions...)
	slices.Sort(p.DataActions)
	p.DataActions = slices.Compact(p.DataActions)

	p.NotActions = append(p.NotActions, perm.NotActions...)
	slices.Sort(p.NotActions)
	p.NotActions = slices.Compact(p.NotActions)

	p.NotDataActions = append(p.NotDataActions, perm.NotDataActions...)
	slices.Sort(p.NotDataActions)
	p.NotDataActions = slices.Compact(p.NotDataActions)
}

// Deprecated: Use ScopedPermissions with ScopeLegacy instead.
func (a API) Permissions(ctx context.Context, features []core.Feature) (Permissions, error) {
	a.client.Log().Print(log.Trace)

	scopedPerms, err := a.ScopedPermissionsForFeatures(ctx, features)
	if err != nil {
		return Permissions{}, err
	}

	return scopedPerms[ScopeLegacy], nil
}

// ScopedPermissions returns the permissions and permission groups for the
// specified RSC feature. The Permissions return value always contains three
// items representing the different permission scopes: legacy, subscription and
// resource group.
func (a API) ScopedPermissions(ctx context.Context, feature core.Feature) ([]Permissions, []PermissionGroupWithVersion, error) {
	a.client.Log().Print(log.Trace)

	permConfig, err := azure.Wrap(a.client).CloudAccountPermissionConfig(ctx, feature)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get permissions: %s", err)
	}

	scopedPerms := make([]Permissions, 3)

	// Subscription scope.
	for _, perm := range permConfig.RolePermissions {
		scopedPerms[ScopeSubscription].addPermissions(Permissions{
			Actions:        perm.IncludedActions,
			DataActions:    perm.IncludedDataActions,
			NotActions:     perm.ExcludedActions,
			NotDataActions: perm.ExcludedDataActions,
		})
	}

	// Resource group scope.
	for _, perm := range permConfig.ResourceGroupRolePermissions {
		scopedPerms[ScopeResourceGroup].addPermissions(Permissions{
			Actions:        perm.IncludedActions,
			DataActions:    perm.IncludedDataActions,
			NotActions:     perm.ExcludedActions,
			NotDataActions: perm.ExcludedDataActions,
		})
	}

	// Legacy scope, provides backwards compatibility with how permissions
	// worked before scoped permissions were introduced.
	scopedPerms[ScopeLegacy].addPermissions(scopedPerms[ScopeSubscription])
	scopedPerms[ScopeLegacy].addPermissions(scopedPerms[ScopeResourceGroup])

	// Permission groups. Note, permissions groups must be sorted in
	// alphabetical order.
	permGroups := make([]PermissionGroupWithVersion, 0, len(permConfig.PermissionGroupVersions))
	for _, permissionGroup := range permConfig.PermissionGroupVersions {
		permGroups = append(permGroups, PermissionGroupWithVersion{
			Name:    permissionGroup.PermissionGroup,
			Version: permissionGroup.Version,
		})
	}
	slices.SortFunc(permGroups, func(i, j PermissionGroupWithVersion) int {
		return cmp.Compare(i.Name, j.Name)
	})

	return scopedPerms, permGroups, nil
}

// ScopedPermissionsForFeatures returns the scoped permissions for a feature
// set. This function violates the RSC Azure permission model and will be
// removed in a future release, use ScopedPermissions instead.
func (a API) ScopedPermissionsForFeatures(ctx context.Context, features []core.Feature) ([]Permissions, error) {
	a.client.Log().Print(log.Trace)

	scopedPerms := make([]Permissions, 3)
	for _, feature := range features {
		scopedPermsForFeature, _, err := a.ScopedPermissions(ctx, feature)
		if err != nil {
			return nil, err
		}
		for i := range scopedPerms {
			scopedPerms[i].addPermissions(scopedPermsForFeature[i])
		}
	}

	return scopedPerms, nil
}

// PermissionsUpdated notifies RSC that the permissions for the Azure service
// principal for the RSC cloud account with the specified id has been updated.
// The permissions should be updated when a feature has the status
// StatusMissingPermissions. Updating the permissions is done outside this SDK.
// The feature parameter is allowed to be nil. When features are nil, all
// features are updated. Note that RSC is only notified about features with
// status StatusMissingPermissions.
func (a API) PermissionsUpdated(ctx context.Context, id IdentityFunc, features []core.Feature) error {
	a.client.Log().Print(log.Trace)

	featureSet := make(map[string]struct{})
	for _, feature := range features {
		featureSet[feature.Name] = struct{}{}
	}

	account, err := a.Subscription(ctx, id, core.FeatureAll)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %s", err)
	}

	for _, feature := range account.Features {
		if feature.Status != core.StatusMissingPermissions {
			continue
		}

		// Check that the feature is in the feature set unless the set is
		//  empty, which is when all features should be updated.
		if _, ok := featureSet[feature.Name]; len(featureSet) > 0 && !ok {
			continue
		}

		err := azure.Wrap(a.client).UpgradeCloudAccountPermissionsWithoutOAuth(ctx, account.ID, feature.Feature, nil)
		if err != nil {
			return fmt.Errorf("failed to update permissions: %s", err)
		}
	}

	return nil
}

// PermissionsUpdatedForTenantDomain notifies RSC that the permissions for the
// Azure service principal in a tenant domain have been updated. The permissions
// should be updated when a feature has the status StatusMissingPermissions.
// Updating the permissions is done outside the SDK. The feature parameter is
// allowed to be nil. When features are nil, all features are updated. Note that
// RSC is only notified about features with status StatusMissingPermissions.
func (a API) PermissionsUpdatedForTenantDomain(ctx context.Context, tenantDomain string, features []core.Feature) error {
	a.client.Log().Print(log.Trace)

	featureSet := make(map[string]struct{})
	for _, feature := range features {
		featureSet[feature.Name] = struct{}{}
	}

	accounts, err := a.Subscriptions(ctx, core.FeatureAll, "")
	if err != nil {
		return fmt.Errorf("failed to get subscriptions: %s", err)
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
			// empty, which is when all features should be updated.
			if _, ok := featureSet[feature.Name]; len(featureSet) > 0 && !ok {
				continue
			}

			err := azure.Wrap(a.client).UpgradeCloudAccountPermissionsWithoutOAuth(ctx, account.ID, feature.Feature, nil)
			if err != nil {
				return fmt.Errorf("failed to update permissions: %s", err)
			}
		}
	}

	return nil
}
