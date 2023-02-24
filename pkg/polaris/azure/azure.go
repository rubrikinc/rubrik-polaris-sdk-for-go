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

// Package azure provides a high level interface to the Azure part of the RSC
// platform.
package azure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API for Azure subscription management.
type API struct {
	client *graphql.Client
}

// Deprecated: use Wrap instead.
func NewAPI(gql *graphql.Client) API {
	return API{client: gql}
}

// Wrap the RSC client in the azure API.
func Wrap(client *polaris.Client) API {
	return API{client: client.GQL}
}

// CloudAccount for Microsoft Azure subscriptions.
type CloudAccount struct {
	ID           uuid.UUID
	NativeID     uuid.UUID
	Name         string
	TenantDomain string
	Features     []Feature
}

// Feature returns the specified feature from the CloudAccount's features.
func (c CloudAccount) Feature(feature core.Feature) (Feature, bool) {
	for _, f := range c.Features {
		if f.Name == feature {
			return f, true
		}
	}

	return Feature{}, false
}

// Feature for Microsoft Azure subscriptions.
type Feature struct {
	Name    core.Feature
	Regions []string
	Status  core.Status
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

// Polaris does not support the AllFeatures for Azure cloud accounts. We work
// around this by translating FeatureAll to the following list of features.
var allFeatures = []core.Feature{
	core.FeatureCloudNativeArchival,
	core.FeatureCloudNativeArchivalEncryption,
	core.FeatureCloudNativeProtection,
	core.FeatureExocompute,
}

// toCloudAccountID returns the RSC cloud account id for the specified identity.
// If the identity is a RSC cloud account id no remote endpoint is called.
func (a API) toCloudAccountID(ctx context.Context, id IdentityFunc) (uuid.UUID, error) {
	a.client.Log().Print(log.Trace)

	if id == nil {
		return uuid.Nil, errors.New("id is not allowed to be nil")
	}
	identity, err := id(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to lookup identity: %v", err)
	}

	uid, err := uuid.Parse(identity.id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse identity: %v", err)
	}
	if identity.internal {
		return uid, nil
	}

	// Note that the same tenant can show up for multiple features.
	tenantIDs := make(map[uuid.UUID]struct{})
	for _, feature := range allFeatures {
		tenants, err := azure.Wrap(a.client).CloudAccountTenants(ctx, feature, false)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to get tenants: %v", err)
		}
		for _, tenant := range tenants {
			tenantIDs[tenant.ID] = struct{}{}
		}
	}

	for tenantID := range tenantIDs {
		for _, feature := range allFeatures {
			tenantWithAccounts, err := azure.Wrap(a.client).CloudAccountTenant(ctx, tenantID, feature, identity.id)
			if err != nil {
				return uuid.Nil, fmt.Errorf("failed to get tenant: %v", err)
			}

			// Find the exact match.
			for _, account := range tenantWithAccounts.Accounts {
				if account.NativeID == uid {
					return account.ID, nil
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("subscription %w", graphql.ErrNotFound)
}

// toNativeID returns the Azure subscription id for the specified identity.
// If the identity is an Azure subscription id no remote endpoint is called.
func (a API) toNativeID(ctx context.Context, id IdentityFunc) (uuid.UUID, error) {
	a.client.Log().Print(log.Trace)

	if id == nil {
		return uuid.Nil, errors.New("id is not allowed to be nil")
	}
	identity, err := id(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to lookup identity: %v", err)
	}

	uid, err := uuid.Parse(identity.id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse identity: %v", err)
	}

	if !identity.internal {
		return uid, nil
	}

	// Note that the same tenant can show up for multiple features.
	tenantIDs := make(map[uuid.UUID]struct{})
	for _, feature := range allFeatures {
		tenants, err := azure.Wrap(a.client).CloudAccountTenants(ctx, feature, false)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to get tenants: %v", err)
		}
		for _, tenant := range tenants {
			tenantIDs[tenant.ID] = struct{}{}
		}
	}

	for tenantID := range tenantIDs {
		for _, feature := range allFeatures {
			tenantWithAccounts, err := azure.Wrap(a.client).CloudAccountTenant(ctx, tenantID, feature, "")
			if err != nil {
				return uuid.Nil, fmt.Errorf("failed to get tenant: %v", err)
			}
			for _, account := range tenantWithAccounts.Accounts {
				if account.ID == uid {
					return account.ID, nil
				}
			}
		}
	}

	return uuid.Nil, fmt.Errorf("subscription %w", graphql.ErrNotFound)
}

// subscriptions return all subscriptions for the given feature and filter.
func (a API) subscriptions(ctx context.Context, feature core.Feature, filter string) ([]CloudAccount, error) {
	a.client.Log().Print(log.Trace)

	tenants, err := azure.Wrap(a.client).CloudAccountTenants(ctx, feature, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenants: %v", err)
	}

	var accounts []CloudAccount
	for _, tenant := range tenants {
		tenantWithAccounts, err := azure.Wrap(a.client).CloudAccountTenant(ctx, tenant.ID, feature, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to get tenant: %v", err)
		}

		for _, account := range tenantWithAccounts.Accounts {
			accounts = append(accounts, CloudAccount{
				ID:           account.ID,
				NativeID:     account.NativeID,
				Name:         account.Name,
				TenantDomain: tenantWithAccounts.DomainName,
				Features: []Feature{{
					Name:    account.Feature.Name,
					Regions: azure.FormatRegions(account.Feature.Regions),
					Status:  account.Feature.Status,
				}},
			})
		}
	}

	return accounts, nil
}

// subscriptionsAllFeatures return all subscriptions with all features for
// the given filter. Note that the organization name of the cloud account is
// not set.
func (a API) subscriptionsAllFeatures(ctx context.Context, filter string) ([]CloudAccount, error) {
	a.client.Log().Print(log.Trace)

	accountMap := make(map[uuid.UUID]*CloudAccount)
	for _, feature := range allFeatures {
		accounts, err := a.subscriptions(ctx, feature, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to get subscriptions: %v", err)
		}

		for i := range accounts {
			// We need to create a copy of the account here since we use it as a
			// pointer further down.
			account := accounts[i]

			if mapped, ok := accountMap[account.ID]; ok {
				mapped.Features = append(mapped.Features, account.Features...)
			} else {
				accountMap[account.ID] = &account
			}
		}
	}

	accounts := make([]CloudAccount, 0, len(accountMap))
	for _, account := range accountMap {
		accounts = append(accounts, *account)
	}

	return accounts, nil
}

// Subscription returns the subscription with specified id and feature.
func (a API) Subscription(ctx context.Context, id IdentityFunc, feature core.Feature) (CloudAccount, error) {
	a.client.Log().Print(log.Trace)

	if id == nil {
		return CloudAccount{}, errors.New("id is not allowed to be nil")
	}
	identity, err := id(ctx)
	if err != nil {
		return CloudAccount{}, fmt.Errorf("failed to lookup identity: %v", err)
	}

	uid, err := uuid.Parse(identity.id)
	if err != nil {
		return CloudAccount{}, fmt.Errorf("failed to parse identity: %v", err)
	}

	if identity.internal {
		accounts, err := a.Subscriptions(ctx, feature, "")
		if err != nil {
			return CloudAccount{}, fmt.Errorf("failed to get subscriptions: %v", err)
		}

		// Find the exact match.
		for _, account := range accounts {
			if account.ID == uid {
				return account, nil
			}
		}
	} else {
		accounts, err := a.Subscriptions(ctx, feature, identity.id)
		if err != nil {
			return CloudAccount{}, fmt.Errorf("failed to get subscriptions: %v", err)
		}

		// Find the exact match.
		for _, account := range accounts {
			if account.NativeID == uid {
				return account, nil
			}
		}
	}

	return CloudAccount{}, fmt.Errorf("subscription %w", graphql.ErrNotFound)
}

// Subscriptions return all subscriptions with the specified feature matching
// the filter. The filter can be used to search for subscription name and
// subscription id.
func (a API) Subscriptions(ctx context.Context, feature core.Feature, filter string) ([]CloudAccount, error) {
	a.client.Log().Print(log.Trace)

	var accounts []CloudAccount
	var err error
	if feature == core.FeatureAll {
		accounts, err = a.subscriptionsAllFeatures(ctx, filter)
	} else {
		accounts, err = a.subscriptions(ctx, feature, filter)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %v", err)
	}

	return accounts, nil
}

// AddSubscription adds the specified subscription to RSC. If name isn't given
// as an option it's derived from the tenant name. Returns the Polaris cloud
// account id of the added subscription.
func (a API) AddSubscription(ctx context.Context, subscription SubscriptionFunc, feature core.Feature, opts ...OptionFunc) (uuid.UUID, error) {
	a.client.Log().Print(log.Trace)

	if subscription == nil {
		return uuid.Nil, errors.New("subscription is not allowed to be nil")
	}
	config, err := subscription(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to lookup subscription: %v", err)
	}

	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return uuid.Nil, fmt.Errorf("failed to lookup option: %v", err)
		}
	}
	err = verifyOptionsForFeature(options, feature)
	if err != nil {
		return uuid.Nil, err
	}
	if options.name != "" {
		config.name = options.name
	}

	// If there already is a RSC cloud account for the given Azure subscription
	// we use the same name when adding the new feature.
	account, err := a.Subscription(ctx, SubscriptionID(config.id), core.FeatureAll)
	if err == nil {
		config.name = account.Name
	}
	if err != nil && !errors.Is(err, graphql.ErrNotFound) {
		return uuid.Nil, fmt.Errorf("failed to get subscription: %v", err)
	}

	perms, err := azure.Wrap(a.client).CloudAccountPermissionConfig(ctx, feature)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get permissions: %v", err)
	}

	cloudAccountFeature := azure.CloudAccountFeature{
		PolicyVersion:       perms.PermissionVersion,
		FeatureType:         feature,
		ResourceGroup:       options.resourceGroup,
		FeatureSpecificInfo: options.featureSpecificInfo,
	}

	_, err = azure.Wrap(a.client).AddCloudAccountWithoutOAuth(ctx, azure.PublicCloud, config.id, cloudAccountFeature,
		config.name, config.tenantDomain, options.regions)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to add subscription: %v", err)
	}

	// If the RSC cloud account did not exist prior we retrieve the RSC cloud
	// account id.
	if account.ID == uuid.Nil {
		account, err = a.Subscription(ctx, SubscriptionID(config.id), feature)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to get subscription: %v", err)
		}
	}

	return account.ID, nil
}

// RemoveSubscription removes the subscription with the specified id from RSC.
// If deleteSnapshots is true the snapshots are deleted otherwise they are kept.
func (a API) RemoveSubscription(ctx context.Context, id IdentityFunc, feature core.Feature, deleteSnapshots bool) error {
	a.client.Log().Print(log.Trace)

	account, err := a.Subscription(ctx, id, feature)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %v", err)
	}

	if account.Features[0].Name == core.FeatureCloudNativeProtection && account.Features[0].Status != core.StatusDisabled {
		// Lookup the RSC native account id from the RSC subscription name and
		// the Azure subscription id. The RSC native account id is needed to
		// delete the RSC native account subscription.
		natives, err := azure.Wrap(a.client).NativeSubscriptions(ctx, account.Name)
		if err != nil {
			return fmt.Errorf("failed to get native subscriptions: %v", err)
		}

		var nativeID uuid.UUID
		for _, native := range natives {
			if native.NativeID == account.NativeID {
				nativeID = native.ID
				break
			}
		}
		if nativeID == uuid.Nil {
			return fmt.Errorf("subscription %w", graphql.ErrNotFound)
		}

		jobID, err := azure.Wrap(a.client).StartDisableNativeSubscriptionProtectionJob(ctx, nativeID, azure.VM, deleteSnapshots)
		if err != nil {
			return fmt.Errorf("failed to disable native subscription: %v", err)
		}

		state, err := core.Wrap(a.client).WaitForTaskChain(ctx, jobID, 10*time.Second)
		if err != nil {
			return fmt.Errorf("failed to wait for task chain: %v", err)
		}
		if state != core.TaskChainSucceeded {
			return fmt.Errorf("taskchain failed: jobID=%v, state=%v", jobID, state)
		}
	}

	err = azure.Wrap(a.client).DeleteCloudAccountWithoutOAuth(ctx, account.ID, feature)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %v", err)
	}

	return nil
}

// UpdateSubscription updates the subscription with the specified id and
// feature.
func (a API) UpdateSubscription(ctx context.Context, id IdentityFunc, feature core.Feature, opts ...OptionFunc) error {
	a.client.Log().Print(log.Trace)

	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return fmt.Errorf("failed to lookup option: %v", err)
		}
	}
	if options.name == "" && len(options.regions) == 0 {
		return errors.New("nothing to update")
	}

	account, err := a.Subscription(ctx, id, feature)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %v", err)
	}
	if options.name == "" {
		options.name = account.Name
	}

	// Update only name.
	if len(options.regions) == 0 {
		if len(account.Features) == 0 {
			return errors.New("invalid cloud account: no features")
		}

		err := azure.Wrap(a.client).UpdateCloudAccount(ctx, account.ID, account.Features[0].Name, options.name,
			[]azure.Region{}, []azure.Region{})
		if err != nil {
			return fmt.Errorf("failed to update subscription: %v", err)
		}

		return nil
	}

	for _, accountFeature := range account.Features {
		regions := make(map[azure.Region]struct{})
		for _, region := range options.regions {
			regions[region] = struct{}{}
		}

		var remove []azure.Region
		for _, region := range accountFeature.Regions {
			reg, err := azure.ParseRegion(region)
			if err != nil {
				return fmt.Errorf("failed to parse region: %v", err)
			}
			if _, ok := regions[reg]; ok {
				delete(regions, reg)
			} else {
				remove = append(remove, reg)
			}
		}

		var add []azure.Region
		for region := range regions {
			add = append(add, region)
		}

		err = azure.Wrap(a.client).UpdateCloudAccount(ctx, account.ID, accountFeature.Name, options.name, add, remove)
		if err != nil {
			return fmt.Errorf("failed to update subscription: %v", err)
		}
	}

	return nil
}

// AddServicePrincipal adds the service principal for the app. If shouldReplace
// is true and the app already has a service principal, it will be replaced.
// Note that it's not possible to remove a service principal once it has been
// set. Returns the application id of the service principal set.
func (a API) AddServicePrincipal(ctx context.Context, principal ServicePrincipalFunc, shouldReplace bool) (uuid.UUID, error) {
	a.client.Log().Print(log.Trace)

	config, err := principal(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to lookup principal: %v", err)
	}

	err = azure.Wrap(a.client).SetCloudAccountCustomerAppCredentials(ctx, azure.PublicCloud, config.appID,
		config.tenantID, config.appName, config.tenantDomain, config.appSecret, shouldReplace)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to set customer app credentials: %v", err)
	}

	return config.appID, nil
}

// SetServicePrincipal sets the service principal for the app. If the app
// already has a service principal, it will be replaced. Note that it's not
// possible to remove a service principal once it has been set. Returns the
// application id of the service principal set.
func (a API) SetServicePrincipal(ctx context.Context, principal ServicePrincipalFunc) (uuid.UUID, error) {
	a.client.Log().Print(log.Trace)

	return a.AddServicePrincipal(ctx, principal, true)
}
