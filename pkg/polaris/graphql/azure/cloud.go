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

// CloudAccount represents an RSC Cloud Account for Azure.
type CloudAccount struct {
	ID       uuid.UUID `json:"id"`
	NativeID uuid.UUID `json:"nativeId"`
	Name     string    `json:"name"`
	Feature  Feature   `json:"featureDetail"`
}

// Feature represents an RSC Cloud Account feature for Azure, e.g. Cloud Native
// Protection.
type Feature struct {
	Feature string      `json:"feature"`
	Regions []Region    `json:"regions"`
	Status  core.Status `json:"status"`
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

// Tag represents tags to be applied to Azure resource.
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TagList represents a list of Tags.
type TagList struct {
	Tags []Tag `json:"tagList"`
}

// ResourceGroup contains the information of resource group
// created for a particular feature.
type ResourceGroup struct {
	Name    string   `json:"name"`
	TagList *TagList `json:"tags"`
	Region  `json:"region"`
}

// FeatureSpecificInfo represents feature specific information.
// Supports:
// 1. Managed Identity for Archival Encryption feature.
type FeatureSpecificInfo struct {
	UserAssignedManagedIdentity *UserAssignedManagedIdentity `json:"userAssignedManagedIdentityInput"`
}

// UserAssignedManagedIdentity represents the managed identity
// information for archival.
type UserAssignedManagedIdentity struct {
	Name              string `json:"name"`
	ResourceGroupName string `json:"resourceGroupName"`
	PrincipalID       string `json:"principalId"`
	Region            `json:"region"`
}

// CloudAccountFeature represents feature information for
// specific cloud native azure features.
type CloudAccountFeature struct {
	PolicyVersion       int                  `json:"policyVersion"`
	ResourceGroup       *ResourceGroup       `json:"resourceGroup"`
	FeatureType         string               `json:"featureType"`
	FeatureSpecificInfo *FeatureSpecificInfo `json:"specificFeatureInput"`
}

// CloudAccountTenant returns the tenant and cloud accounts for the specified
// feature and Polaris tenant id. The filter can be used to search for
// subscription name and subscription id.
func (a API) CloudAccountTenant(ctx context.Context, id uuid.UUID, feature core.Feature, filter string) (CloudAccountTenant, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, azureCloudAccountTenantQuery, struct {
		ID      uuid.UUID `json:"tenantId"`
		Feature string    `json:"feature"`
		Filter  string    `json:"subscriptionSearchText"`
	}{ID: id, Feature: feature.Name, Filter: filter})
	if err != nil {
		return CloudAccountTenant{}, fmt.Errorf("failed to request azureCloudAccountTenant: %w", err)
	}
	a.log.Printf(log.Debug, "azureCloudAccountTenantQuery(%q, %q, %q): %s", id, feature, filter,
		string(buf))

	var payload struct {
		Data struct {
			Result CloudAccountTenant `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return CloudAccountTenant{}, fmt.Errorf("failed to unmarshal azureCloudAccountTenantQuery: %v", err)
	}

	return payload.Data.Result, nil
}

// CloudAccountTenants return all tenants for the specified feature. If
// includeSubscription is true all cloud accounts for each tenant are also
// returned. Note that this function does not support AllFeatures.
func (a API) CloudAccountTenants(ctx context.Context, feature core.Feature, includeSubscriptions bool) ([]CloudAccountTenant, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, allAzureCloudAccountTenantsQuery, struct {
		Feature              string `json:"feature"`
		IncludeSubscriptions bool   `json:"includeSubscriptionDetails"`
	}{Feature: feature.Name, IncludeSubscriptions: includeSubscriptions})
	if err != nil {
		return nil, fmt.Errorf("failed to request allAzureCloudAccountTenants: %w", err)
	}
	a.log.Printf(log.Debug, "allAzureCloudAccountTenants(%q, %t): %s", feature.Name, includeSubscriptions, string(buf))

	var payload struct {
		Data struct {
			Result []CloudAccountTenant `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allAzureCloudAccountTenants: %v", err)
	}

	return payload.Data.Result, nil
}

// AddCloudAccountWithoutOAuth adds the Azure Subscription cloud account
// for given feature without OAuth.
func (a API) AddCloudAccountWithoutOAuth(ctx context.Context, cloud Cloud, id uuid.UUID, feature CloudAccountFeature,
	name, tenantDomain string, regions []Region) (string, error) {

	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, addAzureCloudAccountWithoutOauthQuery, struct {
		Cloud            Cloud               `json:"azureCloudType"`
		Feature          CloudAccountFeature `json:"feature"`
		SubscriptionName string              `json:"subscriptionName"`
		SubscriptionID   uuid.UUID           `json:"subscriptionId"`
		TenantDomain     string              `json:"tenantDomainName"`
		Regions          []Region            `json:"regions"`
	}{Cloud: cloud, Feature: feature, SubscriptionName: name, SubscriptionID: id, TenantDomain: tenantDomain, Regions: regions})
	if err != nil {
		return "", fmt.Errorf("failed to request addAzureCloudAccountWithoutOauth: %w", err)
	}
	a.log.Printf(log.Debug, "addAzureCloudAccountWithoutOauth(%q, %q, %q, %q, %q, %q, %d): %s", cloud, id,
		feature.FeatureType, name, tenantDomain, regions, feature.PolicyVersion, string(buf))

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
		return "", fmt.Errorf("failed to unmarshal addAzureCloudAccountWithoutOauth: %v", err)
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
// feature with the specified RSC cloud account id
func (a API) DeleteCloudAccountWithoutOAuth(ctx context.Context, id uuid.UUID, feature core.Feature) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, deleteAzureCloudAccountWithoutOauthQuery, struct {
		IDs      []uuid.UUID `json:"subscriptionIds"`
		Features []string    `json:"features"`
	}{IDs: []uuid.UUID{id}, Features: []string{feature.Name}})
	if err != nil {
		return fmt.Errorf("failed to request deleteAzureCloudAccountWithoutOauth: %w", err)
	}
	a.log.Printf(log.Debug, "deleteAzureCloudAccountWithoutOauth(%v, %q): %s", id, feature.Name, string(buf))

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
		return fmt.Errorf("failed to unmarshal deleteAzureCloudAccountWithoutOauth: %v", err)
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
// with the specified RSC cloud account id.
func (a API) UpdateCloudAccount(ctx context.Context, id uuid.UUID, feature core.Feature, name string, toAdd, toRemove []Region) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, updateAzureCloudAccountQuery, struct {
		Features      []string             `json:"features"`
		ToAdd         []Region             `json:"regionsToAdd,omitempty"`
		ToRemove      []Region             `json:"regionsToRemove,omitempty"`
		Subscriptions []updateSubscription `json:"subscriptions"`
	}{Features: []string{feature.Name}, ToAdd: toAdd, ToRemove: toRemove, Subscriptions: []updateSubscription{{ID: id, Name: name}}})
	if err != nil {
		return fmt.Errorf("failed to request updateAzureCloudAccount: %w", err)
	}
	a.log.Printf(log.Debug, "updateAzureCloudAccount(%q, %v, %v %v, %v): %s", id, feature.Name, name, toAdd,
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
		return fmt.Errorf("failed to unmarshal updateAzureCloudAccount: %v", err)
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
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, azureCloudAccountPermissionConfigQuery, struct {
		Feature string `json:"feature"`
	}{Feature: feature.Name})
	if err != nil {
		return PermissionConfig{}, fmt.Errorf("failed to request azureCloudAccountPermissionConfig: %w", err)
	}

	a.log.Printf(log.Debug, "azureCloudAccountPermissionConfig(%q): %s", feature, string(buf))

	var payload struct {
		Data struct {
			PermissionConfig PermissionConfig `json:"azureCloudAccountPermissionConfig"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return PermissionConfig{}, fmt.Errorf("failed to unmarshal azureCloudAccountPermissionConfig: %v", err)
	}

	return payload.Data.PermissionConfig, nil
}

// UpgradeCloudAccountPermissionsWithoutOAuth notifies RSC that the permissions
// for the Azure service principal has been updated for the specified RSC cloud
// account id and feature.
func (a API) UpgradeCloudAccountPermissionsWithoutOAuth(ctx context.Context, id uuid.UUID, feature core.Feature) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, upgradeAzureCloudAccountPermissionsWithoutOauthQuery, struct {
		ID      uuid.UUID `json:"cloudAccountId"`
		Feature string    `json:"feature"`
	}{ID: id, Feature: feature.Name})
	if err != nil {
		return fmt.Errorf("failed to request upgradeAzureCloudAccountPermissionsWithoutOauth: %w", err)
	}
	a.log.Printf(log.Debug, "upgradeAzureCloudAccountPermissionsWithoutOauth(%q, %q): %s", id, feature.Name, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				Status bool `json:"status"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal upgradeAzureCloudAccountPermissionsWithoutOauth: %v", err)
	}
	if !payload.Data.Result.Status {
		return errors.New("update cloud account permissions failed")
	}

	return nil
}

// StartDisableCloudAccountJob starts a task chain to disable the feature in the
// cloud account with the specified cloud account id. Returns the RSC task chain
// id.
func (a API) StartDisableCloudAccountJob(ctx context.Context, id uuid.UUID, feature core.Feature) (uuid.UUID, error) {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, startDisableAzureCloudAccountJobQuery, struct {
		ID      uuid.UUID `json:"cloudAccountId"`
		Feature string    `json:"feature"`
	}{ID: id, Feature: feature.Name})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to request StartDisableCloudAccountJob: %w", err)
	}
	a.GQL.Log().Printf(log.Debug, "startDisableAzureCloudAccountJobQuery(%q, %q): %s", id, feature.Name, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				JobIDs []struct {
					JobID uuid.UUID `json:"jobId"`
				} `json:"jobIds"`
				Errors []struct {
					Error string `json:"error"`
				} `json:"errors"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal StartDisableCloudAccountJob: %v", err)
	}
	if len(payload.Data.Result.Errors) != 0 {
		return uuid.Nil, errors.New(payload.Data.Result.Errors[0].Error)
	}
	if len(payload.Data.Result.JobIDs) != 1 {
		return uuid.Nil, fmt.Errorf("expected a single result")
	}

	return payload.Data.Result.JobIDs[0].JobID, nil
}
