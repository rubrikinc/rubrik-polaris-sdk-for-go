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

// Package azure provides a high-level interface to the Azure part of the RSC
// platform.
package azure

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

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
	log    log.Logger
}

// Deprecated: use Wrap instead.
func NewAPI(gql *graphql.Client) API {
	return API{client: gql, log: gql.Log()}
}

// Wrap the RSC client in the azure API.
func Wrap(client *polaris.Client) API {
	return API{client: client.GQL, log: client.GQL.Log()}
}

// CloudAccountTenant represents an Azure tenant in RSC.
type CloudAccountTenant struct {
	Cloud             string
	ID                uuid.UUID // Rubrik tenant ID.
	ClientID          uuid.UUID // Azure app registration application id.
	AppName           string
	DomainName        string // Azure tenant domain.
	SubscriptionCount int
}

// CloudAccount for Azure subscriptions.
type CloudAccount struct {
	ID           uuid.UUID // Rubrik cloud account ID.
	NativeID     uuid.UUID // Azure subscription ID.
	Name         string
	TenantID     uuid.UUID // Rubrik tenant ID.
	TenantDomain string    // Azure tenant domain.
	Features     []Feature
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

// Feature for Azure cloud account.
type Feature struct {
	core.Feature
	ResourceGroup               FeatureResourceGroup
	Regions                     []string
	Status                      core.Status
	UserAssignedManagedIdentity FeatureUserAssignedManagedIdentity
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

// SupportResourceGroup returns true if the feature supports being onboarded
// with a resource group.
func (f Feature) SupportResourceGroup() bool {
	return !f.Equal(core.FeatureAzureSQLDBProtection) && !f.Equal(core.FeatureAzureSQLMIProtection) && !f.Equal(core.FeatureCloudNativeBlobProtection)
}

// SupportUserAssignedManagedIdentity returns true if the feature supports
// being onboarded with user-assigned managed identity.
func (f Feature) SupportUserAssignedManagedIdentity() bool {
	return f.Equal(core.FeatureCloudNativeArchivalEncryption)
}

type FeatureResourceGroup struct {
	Name     string
	NativeID string
	Tags     map[string]string
	Region   string
}

type FeatureUserAssignedManagedIdentity struct {
	Name        string
	NativeID    string
	PrincipalID string
}

// Tenant returns the tenant with the specified ID.
func (a API) Tenant(ctx context.Context, tenantID uuid.UUID) (CloudAccountTenant, error) {
	a.log.Print(log.Trace)

	rawTenants, err := azure.Wrap(a.client).CloudAccountTenants(ctx, core.FeatureAll, false)
	if err != nil {
		return CloudAccountTenant{}, fmt.Errorf("failed to get tenants: %s", err)
	}

	for _, tenant := range toTenants(rawTenants) {
		if tenant.ID == tenantID {
			return tenant, nil
		}
	}

	return CloudAccountTenant{}, fmt.Errorf("tenant %w", graphql.ErrNotFound)
}

// TenantFromAppID returns the tenant with the specified app registration
// application ID.
func (a API) TenantFromAppID(ctx context.Context, appID uuid.UUID) (CloudAccountTenant, error) {
	a.log.Print(log.Trace)

	rawTenants, err := azure.Wrap(a.client).CloudAccountTenants(ctx, core.FeatureAll, false)
	if err != nil {
		return CloudAccountTenant{}, fmt.Errorf("failed to get tenants: %s", err)
	}

	for _, tenant := range toTenants(rawTenants) {
		if tenant.ClientID == appID {
			return tenant, nil
		}
	}

	return CloudAccountTenant{}, fmt.Errorf("tenant %w", graphql.ErrNotFound)
}

// Tenants returns all tenants with the specified feature. This function accepts
// the FeatureAll feature. The filter can be used to search for application ID
// and tenant domain.
func (a API) Tenants(ctx context.Context, filter string) ([]CloudAccountTenant, error) {
	a.log.Print(log.Trace)

	rawTenants, err := azure.Wrap(a.client).CloudAccountTenants(ctx, core.FeatureAll, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenants: %s", err)
	}

	// Filter tenants.
	tenants := make([]CloudAccountTenant, 0, len(rawTenants))
	for _, tenant := range toTenants(rawTenants) {
		if filter == "" || strings.HasPrefix(tenant.DomainName, filter) || strings.HasPrefix(tenant.ClientID.String(), filter) {
			tenants = append(tenants, tenant)
		}
	}

	return tenants, nil
}

// Subscription returns the subscription with specified ID and feature.
func (a API) Subscription(ctx context.Context, id IdentityFunc, feature core.Feature) (CloudAccount, error) {
	a.log.Print(log.Trace)

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

	rawTenants, err := azure.Wrap(a.client).CloudAccountTenants(ctx, feature, true)
	if err != nil {
		return CloudAccount{}, fmt.Errorf("failed to get tenants: %s", err)
	}

	// Find the exact match.
	for _, subscription := range toSubscriptions(rawTenants) {
		if identity.internal {
			if subscription.ID == uid {
				return subscription, nil
			}
		} else {
			if subscription.NativeID == uid {
				return subscription, nil
			}
		}
	}

	return CloudAccount{}, fmt.Errorf("subscription %w", graphql.ErrNotFound)
}

// SubscriptionByNativeID returns the subscription with the specified feature
// and native ID.
func (a API) SubscriptionByNativeID(ctx context.Context, feature core.Feature, nativeID uuid.UUID) (CloudAccount, error) {
	a.log.Print(log.Trace)

	subscriptions, err := a.Subscriptions(ctx, feature, nativeID.String())
	if err != nil {
		return CloudAccount{}, err
	}

	for _, subscription := range subscriptions {
		if subscription.NativeID == nativeID {
			return subscription, nil
		}
	}

	return CloudAccount{}, fmt.Errorf("subscription %q %w", nativeID, graphql.ErrNotFound)
}

// SubscriptionByName returns the subscription with the specified feature and
// name. Tenant domain is optional and ignored if an empty string is passed in.
func (a API) SubscriptionByName(ctx context.Context, feature core.Feature, name, tenantDomain string) (CloudAccount, error) {
	a.log.Print(log.Trace)

	subscriptions, err := a.Subscriptions(ctx, feature, name)
	if err != nil {
		return CloudAccount{}, err
	}

	// Sort the subscriptions ascending on tenant domain and name.
	slices.SortFunc(subscriptions, func(i, j CloudAccount) int {
		return cmp.Compare(i.TenantDomain+i.Name, j.TenantDomain+i.Name)
	})
	for _, subscription := range subscriptions {
		if subscription.Name == name && (tenantDomain == "" || subscription.TenantDomain == tenantDomain) {
			return subscription, nil
		}
	}

	if tenantDomain != "" {
		name = tenantDomain + "/" + name
	}
	return CloudAccount{}, fmt.Errorf("subscription %q %w", name, graphql.ErrNotFound)
}

// Subscriptions return all subscriptions with the specified feature matching
// the filter. The filter can be used to search for subscription name and native
// subscription ID.
func (a API) Subscriptions(ctx context.Context, feature core.Feature, filter string) ([]CloudAccount, error) {
	a.log.Print(log.Trace)

	rawTenants, err := azure.Wrap(a.client).CloudAccountTenants(ctx, feature, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenants: %s", err)
	}

	// Filter subscriptions.
	accounts := make([]CloudAccount, 0, len(rawTenants))
	for _, subscription := range toSubscriptions(rawTenants) {
		if filter == "" || strings.HasPrefix(subscription.Name, filter) || strings.HasPrefix(subscription.NativeID.String(), filter) {
			accounts = append(accounts, subscription)
		}
	}

	return accounts, nil
}

// AddSubscription adds the specified subscription to RSC. If a name isn't given
// as an option, it's derived from the tenant name. Returns the RSC cloud
// account ID of the added subscription.
func (a API) AddSubscription(ctx context.Context, subscription SubscriptionFunc, feature core.Feature, opts ...OptionFunc) (uuid.UUID, error) {
	a.log.Print(log.Trace)

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
	if options.name != "" {
		config.name = options.name
	}

	// If there already is an RSC cloud account for the given Azure
	// subscription, we use the same name when adding the new feature.
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
		PermissionGroups:    perms.PermissionGroupVersions,
		FeatureType:         feature.Name,
		ResourceGroup:       options.resourceGroup,
		FeatureSpecificInfo: options.featureSpecificInfo,
	}

	_, err = azure.Wrap(a.client).AddCloudAccountWithoutOAuth(ctx, azure.PublicCloud, config.id, cloudAccountFeature,
		config.name, config.tenantDomain, options.regions)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to add subscription: %v", err)
	}

	// If the RSC cloud account did not exist prior, we retrieve the RSC cloud
	// account id.
	if account.ID == uuid.Nil {
		account, err = a.Subscription(ctx, SubscriptionID(config.id), feature)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to get subscription: %w", err)
		}
	}

	return account.ID, nil
}

// RemoveSubscription removes the RSC feature from the subscription with the
// specified id.
//
// If a cloud native protection feature is being removed and deleteSnapshots is
// true, the snapshots are deleted otherwise they are kept.
func (a API) RemoveSubscription(ctx context.Context, id IdentityFunc, feature core.Feature, deleteSnapshots bool) error {
	a.log.Print(log.Trace)

	account, err := a.Subscription(ctx, id, feature)
	if err != nil {
		return fmt.Errorf("failed to retrieve subscription: %w", err)
	}

	if err := a.disableFeature(ctx, account, feature, deleteSnapshots); err != nil {
		return fmt.Errorf("failed to disable subscripition feature %s: %s", feature, err)
	}

	err = azure.Wrap(a.client).DeleteCloudAccountWithoutOAuth(ctx, account.ID, feature)
	if err != nil {
		return fmt.Errorf("failed to delete subscription feature %s: %s", feature, err)
	}

	return nil
}

// disableFeature disables the specified subscription feature.
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

	// The Cloud Native Archival and Cloud Native Archival Encryption features
	// should not be disabled.
	if feature.Equal(core.FeatureCloudNativeArchival) || feature.Equal(core.FeatureCloudNativeArchivalEncryption) {
		return nil
	}

	jobID, err := azure.Wrap(a.client).StartDisableCloudAccountJob(ctx, account.ID, feature)
	if err != nil {
		return fmt.Errorf("failed to disable feature %s: %s", feature, err)
	}

	if err := core.Wrap(a.client).WaitForFeatureDisableTaskChain(ctx, jobID, func(ctx context.Context) (bool, error) {
		account, err := a.Subscription(ctx, CloudAccountID(account.ID), feature)
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

// UpdateSubscription updates the subscription with the specified ID and feature.
func (a API) UpdateSubscription(ctx context.Context, id IdentityFunc, feature core.Feature, opts ...OptionFunc) error {
	a.log.Print(log.Trace)

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
		return fmt.Errorf("failed to get subscription: %w", err)
	}
	if options.name == "" {
		options.name = account.Name
	}

	// Update only name.
	if len(options.regions) == 0 {
		if len(account.Features) == 0 {
			return errors.New("invalid cloud account: no features")
		}

		err := azure.Wrap(a.client).UpdateCloudAccount(ctx, account.ID, account.Features[0].Feature, options.name,
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
			reg := azure.RegionFromName(region)
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

		err = azure.Wrap(a.client).UpdateCloudAccount(ctx, account.ID, accountFeature.Feature, options.name, add, remove)
		if err != nil {
			return fmt.Errorf("failed to update subscription: %v", err)
		}
	}

	return nil
}

// AddServicePrincipal adds the service principal for the app. If shouldReplace
// is true and the app already has a service principal, it will be replaced.
// Note that it's not possible to remove a service principal once it has been
// set. Returns the application ID of the service principal set.
func (a API) AddServicePrincipal(ctx context.Context, principal ServicePrincipalFunc, shouldReplace bool) (uuid.UUID, error) {
	a.log.Print(log.Trace)

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
// application ID of the service principal set.
func (a API) SetServicePrincipal(ctx context.Context, principal ServicePrincipalFunc) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	return a.AddServicePrincipal(ctx, principal, true)
}

// toSubscriptions returns the unique subscriptions found in the rawTenants
// slice. This function requires that the tenants include subscription details.
func toSubscriptions(rawTenants []azure.CloudAccountTenant) []CloudAccount {
	type tenantAccounts struct {
		tenant   CloudAccountTenant
		accounts map[uuid.UUID]*CloudAccount
	}

	tenantSet := make(map[uuid.UUID]*tenantAccounts)
	for _, rawTenant := range rawTenants {
		tenant, ok := tenantSet[rawTenant.ID]
		if !ok {
			tenantSet[rawTenant.ID] = &tenantAccounts{
				tenant: CloudAccountTenant{
					Cloud:             string(rawTenant.Cloud),
					ID:                rawTenant.ID,
					ClientID:          rawTenant.ClientID,
					AppName:           rawTenant.AppName,
					DomainName:        rawTenant.DomainName,
					SubscriptionCount: rawTenant.SubscriptionCount,
				},
				accounts: make(map[uuid.UUID]*CloudAccount),
			}
			tenant = tenantSet[rawTenant.ID]
		}

		for _, rawAccount := range rawTenant.Accounts {
			account, ok := tenant.accounts[rawAccount.ID]
			if !ok {
				tenant.accounts[rawAccount.ID] = &CloudAccount{
					ID:           rawAccount.ID,
					NativeID:     rawAccount.NativeID,
					Name:         rawAccount.Name,
					TenantID:     rawTenant.ID,
					TenantDomain: rawTenant.DomainName,
				}
				account = tenant.accounts[rawAccount.ID]
			}

			feature := core.Feature{Name: rawAccount.Feature.Feature}
			if _, ok := account.Feature(feature); !ok {
				tags := make(map[string]string, len(rawAccount.Feature.ResourceGroup.Tags))
				for _, tag := range rawAccount.Feature.ResourceGroup.Tags {
					tags[tag.Key] = tag.Value
				}
				regions := make([]string, 0, len(rawAccount.Feature.Regions))
				for _, region := range rawAccount.Feature.Regions {
					regions = append(regions, region.Name())
				}
				account.Features = append(account.Features, Feature{
					Feature: feature,
					ResourceGroup: FeatureResourceGroup{
						Name:     rawAccount.Feature.ResourceGroup.Name,
						NativeID: rawAccount.Feature.ResourceGroup.NativeID,
						Tags:     tags,
						Region:   rawAccount.Feature.ResourceGroup.Region.Name(),
					},
					Regions: regions,
					Status:  rawAccount.Feature.Status,
					UserAssignedManagedIdentity: FeatureUserAssignedManagedIdentity{
						Name:        rawAccount.Feature.UserAssignedManagedIdentity.Name,
						NativeID:    rawAccount.Feature.UserAssignedManagedIdentity.NativeId,
						PrincipalID: rawAccount.Feature.UserAssignedManagedIdentity.PrincipalID,
					},
				})
			}
		}
	}

	var accounts []CloudAccount
	for _, tenant := range tenantSet {
		for _, account := range tenant.accounts {
			accounts = append(accounts, *account)
		}
	}
	slices.SortFunc(accounts, func(i, j CloudAccount) int {
		return cmp.Compare(i.Name, j.Name)
	})

	return accounts
}

// toTenants returns the unique tenants found in the rawTenants slice.
func toTenants(rawTenants []azure.CloudAccountTenant) []CloudAccountTenant {
	tenantSet := make(map[uuid.UUID]CloudAccountTenant)
	for _, rawTenant := range rawTenants {
		if _, ok := tenantSet[rawTenant.ID]; !ok {
			tenantSet[rawTenant.ID] = CloudAccountTenant{
				Cloud:             string(rawTenant.Cloud),
				ID:                rawTenant.ID,
				ClientID:          rawTenant.ClientID,
				AppName:           rawTenant.AppName,
				DomainName:        rawTenant.DomainName,
				SubscriptionCount: rawTenant.SubscriptionCount,
			}
		}
	}

	tenants := make([]CloudAccountTenant, 0, len(tenantSet))
	for _, tenant := range tenantSet {
		tenants = append(tenants, tenant)
	}
	slices.SortFunc(tenants, func(i, j CloudAccountTenant) int {
		return cmp.Compare(i.DomainName, j.DomainName)
	})

	return tenants
}
