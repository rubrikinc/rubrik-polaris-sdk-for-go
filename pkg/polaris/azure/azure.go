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

// Package azure provides a high level interface to the Azure part of the
// Polaris platform.
package azure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API for Microsoft Azure.
type API struct {
	Version string
	gql     *graphql.Client
}

// NewAPI returns a new API instance. Note that this is a very cheap call to
// make.
func NewAPI(gql *graphql.Client) API {
	return API{Version: gql.Version, gql: gql}
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

// toCloudAccountID returns the Polaris cloud account id for the specified
// identity. If the identity is a Polaris cloud account id no remote endpoint
// is called.
func (a API) toCloudAccountID(ctx context.Context, id IdentityFunc) (uuid.UUID, error) {
	a.gql.Log().Print(log.Trace, "polaris/azure.toCloudAccountID")

	if id == nil {
		return uuid.Nil, errors.New("polaris: id is not allowed to be nil")
	}
	identity, err := id(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	if identity.internal {
		id, err := uuid.Parse(identity.id)
		if err != nil {
			return uuid.Nil, err
		}

		return id, nil
	}

	tenants, err := azure.Wrap(a.gql).CloudAccountTenants(ctx, core.FeatureCloudNativeProtection, false)
	if err != nil {
		return uuid.Nil, err
	}

	for _, tenant := range tenants {
		tenantWithAccounts, err := azure.Wrap(a.gql).CloudAccountTenant(ctx, tenant.ID, core.FeatureCloudNativeProtection, identity.id)
		if err != nil {
			return uuid.Nil, err
		}
		if len(tenantWithAccounts.Accounts) == 0 {
			continue
		}
		if len(tenantWithAccounts.Accounts) > 1 {
			return uuid.Nil, fmt.Errorf("polaris: account %w", graphql.ErrNotUnique)
		}

		return tenantWithAccounts.Accounts[0].ID, nil
	}

	return uuid.Nil, fmt.Errorf("polaris: account %w", graphql.ErrNotFound)
}

// toNativeID returns the Azure subscription id for the specified identity.
// If the identity is an Azure subscription id no remote endpoint is called.
func (a API) toNativeID(ctx context.Context, id IdentityFunc) (uuid.UUID, error) {
	a.gql.Log().Print(log.Trace, "polaris/azure.toNativeID")

	if id == nil {
		return uuid.Nil, errors.New("polaris: id is not allowed to be nil")
	}
	identity, err := id(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	uid, err := uuid.Parse(identity.id)
	if err != nil {
		return uuid.Nil, err
	}

	if !identity.internal {
		return uid, nil
	}

	tenants, err := azure.Wrap(a.gql).CloudAccountTenants(ctx, core.FeatureCloudNativeProtection, false)
	if err != nil {
		return uuid.Nil, err
	}

	for _, tenant := range tenants {
		tenantWithAccounts, err := azure.Wrap(a.gql).CloudAccountTenant(ctx, tenant.ID, core.FeatureCloudNativeProtection, "")
		if err != nil {
			return uuid.Nil, err
		}
		for _, account := range tenantWithAccounts.Accounts {
			if account.ID == uid {
				return account.ID, nil
			}
		}
	}

	return uuid.Nil, fmt.Errorf("polaris: account %w", graphql.ErrNotFound)
}

// Polaris does not support the AllFeatures for Azure cloud accounts. We work
// around this by translating FeatureAll to the following list of features.
var allFeatures = []core.Feature{
	core.FeatureCloudNativeArchival,
	core.FeatureCloudNativeProtection,
	core.FeatureExocompute,
}

// subscriptions return all subscriptions for the given feature and filter.
func (a API) subscriptions(ctx context.Context, feature core.Feature, filter string) ([]CloudAccount, error) {
	a.gql.Log().Print(log.Trace, "polaris/azure.subscriptions")

	tenants, err := azure.Wrap(a.gql).CloudAccountTenants(ctx, feature, false)
	if err != nil {
		return nil, err
	}

	var accounts []CloudAccount
	for _, tenant := range tenants {
		tenantWithAccounts, err := azure.Wrap(a.gql).CloudAccountTenant(ctx, tenant.ID, feature, filter)
		if err != nil {
			return nil, err
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
	a.gql.Log().Print(log.Trace, "polaris/azure.subscriptionsAllFeatures")

	accountMap := make(map[uuid.UUID]*CloudAccount)
	for _, feature := range allFeatures {
		accounts, err := a.subscriptions(ctx, feature, filter)
		if err != nil {
			return nil, err
		}

		for _, account := range accounts {
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
	a.gql.Log().Print(log.Trace, "polaris/azure.Subscription")

	if id == nil {
		return CloudAccount{}, errors.New("polaris: id is not allowed to be nil")
	}
	identity, err := id(ctx)
	if err != nil {
		return CloudAccount{}, err
	}

	if identity.internal {
		id, err := uuid.Parse(identity.id)
		if err != nil {
			return CloudAccount{}, err
		}

		accounts, err := a.Subscriptions(ctx, feature, "")
		if err != nil {
			return CloudAccount{}, err
		}

		for _, account := range accounts {
			if account.ID == id {
				return account, nil
			}
		}
	} else {
		accounts, err := a.Subscriptions(ctx, feature, identity.id)
		if err != nil {
			return CloudAccount{}, err
		}
		if len(accounts) > 1 {
			return CloudAccount{}, fmt.Errorf("polaris: account %w", graphql.ErrNotUnique)
		}
		if len(accounts) == 1 {
			return accounts[0], nil
		}
	}

	return CloudAccount{}, fmt.Errorf("polaris: account %w", graphql.ErrNotFound)
}

// Subscriptions return all subscriptions with the specified feature matching
// the filter. The filter can be used to search for subscription name and
// subscription id.
func (a API) Subscriptions(ctx context.Context, feature core.Feature, filter string) ([]CloudAccount, error) {
	a.gql.Log().Print(log.Trace, "polaris/azure.Subscriptions")

	var accounts []CloudAccount
	var err error
	if feature == core.FeatureAll {
		accounts, err = a.subscriptionsAllFeatures(ctx, filter)
	} else {
		accounts, err = a.subscriptions(ctx, feature, filter)
	}
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

// AddSubscription adds the specified subscription to Polaris. If name isn't
// given as an option it's derived from the tenant name. Returns the Polaris
// cloud account id of the added subscription.
func (a API) AddSubscription(ctx context.Context, subscription SubscriptionFunc, feature core.Feature, opts ...OptionFunc) (uuid.UUID, error) {
	a.gql.Log().Print(log.Trace, "polaris/azure.AddSubscription")

	if subscription == nil {
		return uuid.Nil, errors.New("polaris: subscription is not allowed to be nil")
	}
	config, err := subscription(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return uuid.Nil, err
		}
	}
	if options.name != "" {
		config.name = options.name
	}

	// If there already is a Polaris cloud account for the given Azure
	// subscription we use the same name when adding the new feature.
	account, err := a.Subscription(ctx, SubscriptionID(config.id), core.FeatureAll)
	if err == nil {
		config.name = account.Name
	}
	if err != nil && !errors.Is(err, graphql.ErrNotFound) {
		return uuid.Nil, err
	}

	perms, err := azure.Wrap(a.gql).CloudAccountPermissionConfig(ctx, feature)
	if err != nil {
		return uuid.Nil, err
	}

	_, err = azure.Wrap(a.gql).AddCloudAccountWithoutOAuth(ctx, azure.PublicCloud, config.id, feature,
		config.name, config.tenantDomain, options.regions, perms.PermissionVersion)
	if err != nil {
		return uuid.Nil, err
	}

	// If the Polaris cloud account did not exist prior we retrieve the Polaris
	// cloud account id.
	if account.ID == uuid.Nil {
		account, err = a.Subscription(ctx, SubscriptionID(config.id), feature)
		if err != nil {
			return uuid.Nil, err
		}
	}

	return account.ID, nil
}

// RemoveSubscription removes the subscription with the specified id from
// Polaris. If deleteSnapshots is true the snapshots are deleted otherwise they
// are kept.
func (a API) RemoveSubscription(ctx context.Context, id IdentityFunc, feature core.Feature, deleteSnapshots bool) error {
	a.gql.Log().Print(log.Trace, "polaris/azure.RemoveSubscription")

	account, err := a.Subscription(ctx, id, feature)
	if err != nil {
		return err
	}

	if account.Features[0].Name == core.FeatureCloudNativeProtection && account.Features[0].Status != core.StatusDisabled {
		// Lookup the Polaris native account id from the Polaris subscription name
		// and the Azure subscription id. The Polaris native account id is needed
		// to delete the Polaris native account subscription.
		natives, err := azure.Wrap(a.gql).NativeSubscriptions(ctx, account.Name)
		if err != nil {
			return err
		}

		var nativeID uuid.UUID
		for _, native := range natives {
			if native.NativeID == account.NativeID {
				nativeID = native.ID
				break
			}
		}
		if nativeID == uuid.Nil {
			return fmt.Errorf("polaris: account: %w", graphql.ErrNotFound)
		}

		jobID, err := azure.Wrap(a.gql).StartDisableNativeSubscriptionProtectionJob(ctx, nativeID, deleteSnapshots)
		if err != nil {
			return err
		}

		state, err := core.Wrap(a.gql).WaitForTaskChain(ctx, jobID, 10*time.Second)
		if err != nil {
			return err
		}
		if state != core.TaskChainSucceeded {
			return fmt.Errorf("polaris: taskchain failed: jobID=%v, state=%v", jobID, state)
		}
	}

	err = azure.Wrap(a.gql).DeleteCloudAccountWithoutOAuth(ctx, account.ID, feature)
	if err != nil {
		return err
	}

	return nil
}

// UpdateSubscription updates the subscription with the specied id and feature.
func (a API) UpdateSubscription(ctx context.Context, id IdentityFunc, feature core.Feature, opts ...OptionFunc) error {
	a.gql.Log().Print(log.Trace, "polaris/azure.UpdateSubscription")

	var options options
	for _, option := range opts {
		if err := option(ctx, &options); err != nil {
			return err
		}
	}
	if options.name == "" && len(options.regions) == 0 {
		return errors.New("polaris: nothing to update")
	}

	account, err := a.Subscription(ctx, id, feature)
	if err != nil {
		return err
	}
	if options.name == "" {
		options.name = account.Name
	}

	// Update only name.
	if len(options.regions) == 0 {
		if len(account.Features) == 0 {
			return errors.New("polaris: invalid cloud account: no features")
		}

		err := azure.Wrap(a.gql).UpdateCloudAccount(ctx, account.ID, account.Features[0].Name, options.name, []azure.Region{}, []azure.Region{})
		if err != nil {
			return err
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
				return err
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

		err = azure.Wrap(a.gql).UpdateCloudAccount(ctx, account.ID, accountFeature.Name, options.name, add, remove)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetServicePrincipal sets the default service principal. Note that it's not
// possible to remove a service account once it has been set. Returns the
// application id of the service principal set.
func (a API) SetServicePrincipal(ctx context.Context, principal ServicePrincipalFunc) (uuid.UUID, error) {
	a.gql.Log().Print(log.Trace, "polaris/azure.SetServicePrincipal")

	config, err := principal(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	err = azure.Wrap(a.gql).SetCloudAccountCustomerAppCredentials(ctx, azure.PublicCloud, config.appID,
		config.tenantID, config.appName, config.tenantDomain, config.appSecret)
	if err != nil {
		return uuid.Nil, err
	}

	return config.appID, nil
}
