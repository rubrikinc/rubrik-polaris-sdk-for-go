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
	Name    core.Feature `json:"feature"`
	Regions []Region     `json:"regions"`
	Status  core.Status  `json:"status"`
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
func (a API) CloudAccountTenant(ctx context.Context, id uuid.UUID, feature core.Feature, filter string) (CloudAccountTenant, error) {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, azureCloudAccountTenantQuery, struct {
		ID      uuid.UUID    `json:"tenantId"`
		Feature core.Feature `json:"feature"`
		Filter  string       `json:"subscriptionSearchText"`
	}{ID: id, Feature: feature, Filter: filter})
	if err != nil {
		return CloudAccountTenant{}, fmt.Errorf("failed to request CloudAccountTenant: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "azureCloudAccountTenant(%q, %q, %q): %s", id, feature, filter, string(buf))

	var payload struct {
		Data struct {
			Result CloudAccountTenant `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return CloudAccountTenant{}, fmt.Errorf("failed to unmarshal CloudAccountTenant: %v", err)
	}

	return payload.Data.Result, nil
}

// CloudAccountTenants return all tenants for the specified feature. If
// includeSubscription is true all cloud accounts for each tenant are also
// returned. Note that this function does not support AllFeatures.
func (a API) CloudAccountTenants(ctx context.Context, feature core.Feature, includeSubscriptions bool) ([]CloudAccountTenant, error) {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, allAzureCloudAccountTenantsQuery, struct {
		Feature              core.Feature `json:"feature"`
		IncludeSubscriptions bool         `json:"includeSubscriptionDetails"`
	}{Feature: feature, IncludeSubscriptions: includeSubscriptions})
	if err != nil {
		return nil, fmt.Errorf("failed to request CloudAccountTenants: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "allAzureCloudAccountTenants(%q, %t): %s", feature, includeSubscriptions, string(buf))

	var payload struct {
		Data struct {
			Result []CloudAccountTenant `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CloudAccountTenants: %v", err)
	}

	return payload.Data.Result, nil
}

type addSubscription struct {
	ID   uuid.UUID `json:"nativeId"`
	Name string    `json:"name"`
}

// AddCloudAccountWithoutOAuth adds the Azure Subscription cloud account
// for given feature without OAuth.
func (a API) AddCloudAccountWithoutOAuth(ctx context.Context, cloud Cloud, id uuid.UUID, feature core.Feature,
	name, tenantDomain string, regions []Region, policyVersion int) (string, error) {

	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, addAzureCloudAccountWithoutOauthQuery, struct {
		Cloud         Cloud             `json:"azureCloudType"`
		Features      []core.Feature    `json:"features"`
		Subscriptions []addSubscription `json:"subscriptions"`
		TenantDomain  string            `json:"tenantDomainName"`
		Regions       []Region          `json:"regions"`
		PolicyVersion int               `json:"policyVersion"`
	}{Cloud: cloud, Features: []core.Feature{feature}, Subscriptions: []addSubscription{{id, name}}, TenantDomain: tenantDomain, Regions: regions, PolicyVersion: policyVersion})
	if err != nil {
		return "", fmt.Errorf("failed to request AddCloudAccountWithoutOAuth: %v", err)
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
		return "", fmt.Errorf("failed to unmarshal AddCloudAccountWithoutOAuth: %v", err)
	}
	if len(payload.Data.Result.Status) != 1 {
		return "", errors.New("expected a single result")
	}
	if payload.Data.Result.Status[0].Error != "" {
		return "", errors.New(payload.Data.Result.Status[0].Error)
	}

	return payload.Data.Result.TenantID, nil
}

// DeleteCloudAccountWithoutOAuth delete the Azure subscription cloud account
// feature with the specified Polaris cloud account id
func (a API) DeleteCloudAccountWithoutOAuth(ctx context.Context, id uuid.UUID, feature core.Feature) error {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, deleteAzureCloudAccountWithoutOauthQuery, struct {
		IDs      []uuid.UUID    `json:"subscriptionIds"`
		Features []core.Feature `json:"features"`
	}{IDs: []uuid.UUID{id}, Features: []core.Feature{feature}})
	if err != nil {
		return fmt.Errorf("failed to request DeleteCloudAccountWithoutOAuth: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "deleteAzureCloudAccountWithoutOauth(%v, %q): %s", id, feature, string(buf))

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
		return fmt.Errorf("failed to unmarshal DeleteCloudAccountWithoutOAuth: %v", err)
	}
	if len(payload.Data.Result.Status) != 1 {
		return errors.New("expected a single result")
	}
	if !payload.Data.Result.Status[0].Success {
		return errors.New(payload.Data.Result.Status[0].Error)
	}

	return nil
}

type updateSubscription struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// UpdateCloudAccount updates the name and the regions for the cloud account
// with the specified Polaris cloud account id.
func (a API) UpdateCloudAccount(ctx context.Context, id uuid.UUID, feature core.Feature, name string, toAdd, toRemove []Region) error {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, updateAzureCloudAccountQuery, struct {
		Features      []core.Feature       `json:"features"`
		ToAdd         []Region             `json:"regionsToAdd,omitempty"`
		ToRemove      []Region             `json:"regionsToRemove,omitempty"`
		Subscriptions []updateSubscription `json:"subscriptions"`
	}{Features: []core.Feature{feature}, ToAdd: toAdd, ToRemove: toRemove, Subscriptions: []updateSubscription{{ID: id, Name: name}}})
	if err != nil {
		return fmt.Errorf("failed to request UpdateCloudAccount: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "updateAzureCloudAccount(%q, %v, %v %v, %v): %s", id, feature, name, toAdd, toRemove, string(buf))

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
		return fmt.Errorf("failed to unmarshal UpdateCloudAccount: %v", err)
	}
	if len(payload.Data.Result.Status) != 1 {
		return errors.New("expected a single result")
	}
	if !payload.Data.Result.Status[0].Success {
		return errors.New("update cloud account failed")
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
func (a API) CloudAccountPermissionConfig(ctx context.Context, feature core.Feature) (PermissionConfig, error) {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, azureCloudAccountPermissionConfigQuery, struct {
		Feature core.Feature `json:"feature"`
	}{Feature: feature})
	if err != nil {
		return PermissionConfig{}, fmt.Errorf("failed to request CloudAccountPermissionConfig: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "azureCloudAccountPermissionConfig(%q): %s", feature, string(buf))

	var payload struct {
		Data struct {
			PermissionConfig PermissionConfig `json:"azureCloudAccountPermissionConfig"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return PermissionConfig{}, fmt.Errorf("failed to unmarshal CloudAccountPermissionConfig: %v", err)
	}

	return payload.Data.PermissionConfig, nil
}

// UpgradeCloudAccountPermissionsWithoutOAuth notifies Polaris that the
// permissions for the Azure service prinicpal has been updated for the
// specified Polaris cloud account id and feature.
func (a API) UpgradeCloudAccountPermissionsWithoutOAuth(ctx context.Context, id uuid.UUID, feature core.Feature) error {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, upgradeAzureCloudAccountPermissionsWithoutOauthQuery, struct {
		ID      uuid.UUID    `json:"cloudAccountId"`
		Feature core.Feature `json:"feature"`
	}{ID: id, Feature: feature})
	if err != nil {
		return fmt.Errorf("failed to request UpgradeCloudAccountPermissionsWithoutOAuth: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "upgradeAzureCloudAccountPermissionsWithoutOauth(%q, %q): %s", id, feature, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				Status bool `json:"status"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal UpgradeCloudAccountPermissionsWithoutOAuth: %v", err)
	}

	if !payload.Data.Result.Status {
		return errors.New("update cloud account permissions failed")
	}

	return nil
}
