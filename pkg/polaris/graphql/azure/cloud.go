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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CloudAccount represents a Polaris Cloud Account for Azure.
type CloudAccount struct {
	ID       uuid.UUID `json:"id"`
	NativeID uuid.UUID `json:"nativeId"`
	Name     string    `json:"name"`
	Feature  Feature   `json:"featureDetail"`
}

// Feature represents a Polaris Cloud Account feature for Azure, e.g Cloud
// Native Protection.
type Feature struct {
	Name    core.CloudAccountFeature `json:"feature"`
	Regions []Region                 `json:"regions"`
	Status  core.CloudAccountStatus  `json:"status"`
}

// CloudAccountTenant hold details about an Azure tenant and the cloud
// account associated with that tenant.
type CloudAccountTenant struct {
	Cloud      Cloud          `json:"cloudType"`
	ID         uuid.UUID      `json:"azureCloudAccountTenantRubrikId"`
	ClientID   uuid.UUID      `json:"clientId"`
	DomainName string         `json:"domainName"`
	Accounts   []CloudAccount `json:"subscriptions"`
}

// CloudAccountTenant returns the tenant and cloud accounts for the specified
// feature and Polaris tenant id. The filter can be used to search for
// subscription name and subscription id.
func (a API) CloudAccountTenant(ctx context.Context, id uuid.UUID, feature core.CloudAccountFeature, filter string) (CloudAccountTenant, error) {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/azure.CloudAccountTenant")

	buf, err := a.GQL.Request(ctx, azureCloudAccountTenantQuery, struct {
		ID      uuid.UUID                `json:"tenantId"`
		Feature core.CloudAccountFeature `json:"feature"`
		Filter  string                   `json:"subscriptionSearchText"`
	}{ID: id, Feature: feature, Filter: filter})
	if err != nil {
		return CloudAccountTenant{}, err
	}

	a.GQL.Log().Printf(log.Debug, "azureCloudAccountTenant(%q, %q, %q): %s", id, feature, filter, string(buf))

	var payload struct {
		Data struct {
			Result CloudAccountTenant `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return CloudAccountTenant{}, err
	}

	return payload.Data.Result, nil
}

// CloudAccountTenants return all tenants for the specified feature. If
// includeSubscription is true all cloud accounts for each tenant are also
// returned. Note that this function does not support AllFeatures.
func (a API) CloudAccountTenants(ctx context.Context, feature core.CloudAccountFeature, includeSubscriptions bool) ([]CloudAccountTenant, error) {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/azure.CloudAccountTenants")

	buf, err := a.GQL.Request(ctx, allAzureCloudAccountTenantsQuery, struct {
		Feature              core.CloudAccountFeature `json:"feature"`
		IncludeSubscriptions bool                     `json:"includeSubscriptionDetails"`
	}{Feature: feature, IncludeSubscriptions: includeSubscriptions})
	if err != nil {
		return nil, err
	}

	a.GQL.Log().Printf(log.Debug, "allAzureCloudAccountTenants(%q, %t): %s", feature, includeSubscriptions, string(buf))

	var payload struct {
		Data struct {
			Result []CloudAccountTenant `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, err
	}

	return payload.Data.Result, nil
}

type addSubscription struct {
	ID   uuid.UUID `json:"nativeId"`
	Name string    `json:"name"`
}

// AddCloudAccountWithoutOAuth adds the Azure Subscription cloud account
// for given feature without OAuth.
func (a API) AddCloudAccountWithoutOAuth(ctx context.Context, cloud Cloud, id uuid.UUID, feature core.CloudAccountFeature,
	name, tenantDomain string, regions []Region, policyVersion int) (string, error) {

	a.GQL.Log().Print(log.Trace, "polaris/graphql/azure.AddCloudAccountWithoutOAuth")

	buf, err := a.GQL.Request(ctx, addAzureCloudAccountWithoutOauthQuery, struct {
		Cloud         Cloud                      `json:"azureCloudType"`
		Features      []core.CloudAccountFeature `json:"features"`
		Subscriptions []addSubscription          `json:"subscriptions"`
		TenantDomain  string                     `json:"tenantDomainName"`
		Regions       []Region                   `json:"regions"`
		PolicyVersion int                        `json:"policyVersion"`
	}{Cloud: cloud, Features: []core.CloudAccountFeature{feature}, Subscriptions: []addSubscription{{id, name}},
		TenantDomain: tenantDomain, Regions: regions, PolicyVersion: policyVersion})
	if err != nil {
		return "", err
	}

	a.GQL.Log().Printf(log.Debug, "addAzureCloudAccountWithoutOauth(%q, %q, %q, %q, %q, %q, %d): %s", cloud, id, feature, name,
		tenantDomain, regions, policyVersion, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				TenantID string `json:"tenantId"`
				Status   []struct {
					SubscriptionID       string `json:"azureSubscriptionRubrikId"`
					SubscriptionNativeID string `json:"azureSubscriptionNativeId"`
					Error                string `json:"error"`
				} `json:"status"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", err
	}
	if len(payload.Data.Result.Status) != 1 {
		return "", errors.New("polaris: add returned no status")
	}
	if err := payload.Data.Result.Status[0].Error; err != "" {
		return "", fmt.Errorf("polaris: %s", err)
	}

	return payload.Data.Result.TenantID, nil
}

// DeleteCloudAccountWithoutOAuth delete the Azure subscription cloud account
// feature with the specified Polaris cloud account id
func (a API) DeleteCloudAccountWithoutOAuth(ctx context.Context, id uuid.UUID, feature core.CloudAccountFeature) error {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/azure.DeleteCloudAccountWithoutOAuth")

	query := deleteAzureCloudAccountWithoutOauthQuery
	if graphql.VersionOlderThan(a.Version, "master-41845", "v20210921") {
		query = deleteAzureCloudAccountWithoutOauthV0Query
	}
	buf, err := a.GQL.Request(ctx, query, struct {
		IDs      []uuid.UUID                `json:"subscriptionIds"`
		Feature  core.CloudAccountFeature   `json:"feature"`
		Features []core.CloudAccountFeature `json:"features"`
	}{IDs: []uuid.UUID{id}, Feature: feature, Features: []core.CloudAccountFeature{feature}})
	if err != nil {
		return err
	}

	a.GQL.Log().Printf(log.Debug, "%s(%v, %q): %s", graphql.QueryName(query), id, feature, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				Status []struct {
					SubscriptionID string `json:"azureSubscriptionNativeId"`
					Success        bool   `json:"isSuccess"`
					Error          string `json:"error"`
				} `json:"status"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return err
	}
	if len(payload.Data.Result.Status) != 1 {
		return errors.New("polaris: delete returned no status")
	}
	if !payload.Data.Result.Status[0].Success {
		return fmt.Errorf("polaris: %s", payload.Data.Result.Status[0].Error)
	}

	return nil
}

type updateSubscription struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// UpdateCloudAccount updates the name and the regions for the cloud account
// with the specified Polaris cloud account id.
func (a API) UpdateCloudAccount(ctx context.Context, id uuid.UUID, feature core.CloudAccountFeature, name string, toAdd, toRemove []Region) error {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/azure.UpdateCloudAccount")

	query := updateAzureCloudAccountQuery
	if graphql.VersionOlderThan(a.Version, "master-41845", " v20210921") {
		query = updateAzureCloudAccountV0Query
	}

	a.GQL.Log().Print(log.Debug, query)

	buf, err := a.GQL.Request(ctx, query, struct {
		Feature       core.CloudAccountFeature   `json:"feature"`
		Features      []core.CloudAccountFeature `json:"features"`
		ToAdd         []Region                   `json:"regionsToAdd,omitempty"`
		ToRemove      []Region                   `json:"regionsToRemove,omitempty"`
		Subscriptions []updateSubscription       `json:"subscriptions"`
	}{Feature: feature, Features: []core.CloudAccountFeature{feature}, ToAdd: toAdd, ToRemove: toRemove,
		Subscriptions: []updateSubscription{{ID: id, Name: name}}})
	if err != nil {
		return err
	}

	a.GQL.Log().Printf(log.Debug, "%s(%q, %v, %v %v): %s", graphql.QueryName(query), id, feature, name, toAdd,
		toRemove, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				Status []struct {
					SubscriptionID string `json:"azureSubscriptionNativeId"`
					Success        bool   `json:"isSuccess"`
				} `json:"status"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return err
	}
	if len(payload.Data.Result.Status) != 1 {
		return errors.New("polaris: update returned no status")
	}
	if !payload.Data.Result.Status[0].Success {
		return errors.New("polaris: update failed")
	}

	return nil
}

// PermissionConfig holds the permissions and version required to enable a
// given feature. IncludedActions refers to actions which should be allowed on
// the Azure role for the subscription. ExcludedActions refers to actions which
// should be explicitly disallowed on the Azure role for the subscription.
type PermissionConfig struct {
	PermissionVersion int `json:"permissionVersion"`
	RolePermissions   []struct {
		ExcludedActions     []string `json:"excludedActions"`
		ExcludedDataActions []string `json:"excludedDataActions"`
		IncludedActions     []string `json:"includedActions"`
		IncludedDataActions []string `json:"includedDataActions"`
	} `json:"rolePermissions"`
}

// CloudAccountPermissionConfig returns the permissions and version required to
// enable the given feature for the Azure subscription.
func (a API) CloudAccountPermissionConfig(ctx context.Context, feature core.CloudAccountFeature) (PermissionConfig, error) {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/azure.CloudAccountPermissionConfig")

	buf, err := a.GQL.Request(ctx, azureCloudAccountPermissionConfigQuery, struct {
		Feature core.CloudAccountFeature `json:"feature"`
	}{Feature: feature})
	if err != nil {
		return PermissionConfig{}, err
	}

	a.GQL.Log().Printf(log.Debug, "azureCloudAccountPermissionConfig(%q): %s", feature, string(buf))

	var payload struct {
		Data struct {
			PermissionConfig PermissionConfig `json:"azureCloudAccountPermissionConfig"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return PermissionConfig{}, err
	}

	return payload.Data.PermissionConfig, nil
}
