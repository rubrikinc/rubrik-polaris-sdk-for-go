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

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CloudAccountTenant hold details about an Azure tenant and the cloud accounts
// associated with the tenant.
type CloudAccountTenant struct {
	Cloud             Cloud          `json:"cloudType"`
	ID                uuid.UUID      `json:"azureCloudAccountTenantRubrikId"`
	ClientID          uuid.UUID      `json:"clientId"`
	AppName           string         `json:"appName"`
	DomainName        string         `json:"domainName"`
	SubscriptionCount int            `json:"subscriptionCount"`
	Accounts          []CloudAccount `json:"subscriptions"`
}

// CloudAccount represents an RSC Azure cloud account.
type CloudAccount struct {
	ID       uuid.UUID `json:"id"`
	NativeID uuid.UUID `json:"nativeId"`
	Name     string    `json:"name"`
	Feature  Feature   `json:"featureDetail"`
}

// Feature represents an RSC Cloud Account feature for Azure, e.g., Cloud Native
// Protection.
type Feature struct {
	Feature                     string                             `json:"feature"`
	PermissionGroups            []core.PermissionGroup             `json:"permissionsGroups"`
	ResourceGroup               FeatureResourceGroup               `json:"resourceGroup"`
	Regions                     []azure.CloudAccountRegionEnum     `json:"regions"`
	Status                      core.Status                        `json:"status"`
	UserAssignedManagedIdentity FeatureUserAssignedManagedIdentity `json:"userAssignedManagedIdentity"`
}

// FeatureResourceGroup represents a resource group for a particular feature.
type FeatureResourceGroup struct {
	Name     string                 `json:"name"`
	NativeID string                 `json:"nativeId"`
	Tags     []core.Tag             `json:"tags"`
	Region   azure.NativeRegionEnum `json:"region"`
}

// ResourceGroup holds the information for a resource group when a particular
// feature is onboarded. Note, when using the ResourceGroup type as input to a
// GraphQL mutation, the TagList.Tags field cannot be nil. An empty slice is
// fine.
type ResourceGroup struct {
	Name    string                       `json:"name"`
	TagList TagList                      `json:"tags"`
	Region  azure.CloudAccountRegionEnum `json:"region"`
}

// TagList represents a list of Tags.
type TagList struct {
	Tags []core.Tag `json:"tagList"`
}

// FeatureUserAssignedManagedIdentity represents a user-assigned managed
// identity for a particular feature.
type FeatureUserAssignedManagedIdentity struct {
	Name        string `json:"name"`
	NativeId    string `json:"nativeId"`
	PrincipalID string `json:"principalId"`
}

// UserAssignedManagedIdentity holds the information for a user-assigned managed
// identity when a particular feature is onboarded.
type UserAssignedManagedIdentity struct {
	Name              string                       `json:"name"`
	ResourceGroupName string                       `json:"resourceGroupName"`
	PrincipalID       string                       `json:"principalId"`
	Region            azure.CloudAccountRegionEnum `json:"region"`
}

// CloudAccountFeature holds the information for a particular feature when it's
// onboarded.
type CloudAccountFeature struct {
	PolicyVersion       int                          `json:"policyVersion"`
	PermissionGroups    []PermissionGroupWithVersion `json:"permissionsGroups,omitempty"`
	ResourceGroup       *ResourceGroup               `json:"resourceGroup,omitempty"`
	FeatureType         string                       `json:"featureType"`
	FeatureSpecificInfo *FeatureSpecificInfo         `json:"specificFeatureInput,omitempty"`
}

// PermissionGroupWithVersion represents a permission group, and its version
// for a particular feature.
type PermissionGroupWithVersion struct {
	PermissionGroup string `json:"permissionsGroup"`
	Version         int    `json:"version"`
}

// FeatureSpecificInfo represents feature specific information.
// Supports:
//
//  1. User-assigned managed identity for the Cloud Native Archival Encryption
//     feature.
type FeatureSpecificInfo struct {
	UserAssignedManagedIdentity *UserAssignedManagedIdentity `json:"userAssignedManagedIdentityInput,omitempty"`
}

// CloudAccountTenants return all tenants for the specified feature. If
// includeSubscription is true, all cloud accounts for each tenant are also
// returned.
func (a API) CloudAccountTenants(ctx context.Context, feature core.Feature, includeSubscriptions bool) ([]CloudAccountTenant, error) {
	a.log.Print(log.Trace)

	query := allAzureCloudAccountTenantsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Feature              string `json:"feature"`
		IncludeSubscriptions bool   `json:"includeSubscriptionDetails"`
	}{Feature: feature.Name, IncludeSubscriptions: includeSubscriptions})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []CloudAccountTenant `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// AddCloudAccountWithoutOAuth adds the Azure Subscription cloud account
// for given feature without OAuth.
func (a API) AddCloudAccountWithoutOAuth(ctx context.Context, cloud Cloud, id uuid.UUID, feature CloudAccountFeature,
	name, tenantDomain string, regions []azure.Region) (string, error) {
	a.log.Print(log.Trace)

	query := addAzureCloudAccountWithoutOauthQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Cloud            Cloud                          `json:"azureCloudType"`
		Feature          CloudAccountFeature            `json:"feature"`
		SubscriptionName string                         `json:"subscriptionName"`
		SubscriptionID   uuid.UUID                      `json:"subscriptionId"`
		TenantDomain     string                         `json:"tenantDomainName"`
		Regions          []azure.CloudAccountRegionEnum `json:"regions"`
	}{
		Cloud:            cloud,
		Feature:          feature,
		SubscriptionName: name,
		SubscriptionID:   id,
		TenantDomain:     tenantDomain,
		Regions:          RegionsToCloudAccountRegionEnum(regions),
	})
	if err != nil {
		return "", graphql.RequestError(query, err)
	}

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
		return "", graphql.UnmarshalError(query, err)
	}
	if len(payload.Data.Result.Status) != 1 {
		return "", graphql.ResponseError(query, errors.New("expected a single result"))
	}
	if payload.Data.Result.Status[0].Error != "" {
		return "", graphql.ResponseError(query, errors.New(payload.Data.Result.Status[0].Error))
	}

	return payload.Data.Result.TenantID, nil
}

// DeleteCloudAccountWithoutOAuth delete the Azure subscription cloud account
// feature with the specified RSC cloud account id
func (a API) DeleteCloudAccountWithoutOAuth(ctx context.Context, id uuid.UUID, feature core.Feature) error {
	a.log.Print(log.Trace)

	query := deleteAzureCloudAccountWithoutOauthQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		IDs      []uuid.UUID `json:"subscriptionIds"`
		Features []string    `json:"features"`
	}{IDs: []uuid.UUID{id}, Features: []string{feature.Name}})
	if err != nil {
		return graphql.RequestError(query, err)
	}

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
		return graphql.UnmarshalError(query, err)
	}
	if len(payload.Data.Result.Status) != 1 {
		return graphql.ResponseError(query, errors.New("expected a single result"))
	}
	if !payload.Data.Result.Status[0].Success {
		return graphql.ResponseError(query, errors.New(payload.Data.Result.Status[0].Error))
	}

	return nil
}

// UpdateCloudAccount updates the name and the regions for the cloud account
// with the specified RSC cloud account id.
func (a API) UpdateCloudAccount(ctx context.Context, id uuid.UUID, feature core.Feature, name string, toAdd, toRemove []azure.Region) error {
	a.log.Print(log.Trace)

	type updateSubscription struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}
	query := updateAzureCloudAccountQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Features      []string                       `json:"features"`
		ToAdd         []azure.CloudAccountRegionEnum `json:"regionsToAdd,omitempty"`
		ToRemove      []azure.CloudAccountRegionEnum `json:"regionsToRemove,omitempty"`
		Subscriptions []updateSubscription           `json:"subscriptions"`
	}{
		Features:      []string{feature.Name},
		ToAdd:         RegionsToCloudAccountRegionEnum(toAdd),
		ToRemove:      RegionsToCloudAccountRegionEnum(toRemove),
		Subscriptions: []updateSubscription{{ID: id, Name: name}},
	})
	if err != nil {
		return graphql.RequestError(query, err)
	}

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
		return graphql.UnmarshalError(query, err)
	}
	if len(payload.Data.Result.Status) != 1 {
		return graphql.ResponseError(query, errors.New("expected a single result"))
	}
	if !payload.Data.Result.Status[0].Success {
		return graphql.ResponseError(query, errors.New("update cloud account failed"))
	}

	return nil
}

// PermissionConfig holds the permissions and version required to enable a
// given feature. IncludedActions refers to actions which should be allowed on
// the Azure role for the subscription. ExcludedActions refers to actions which
// should be explicitly disallowed on the Azure role for the subscription.
type PermissionConfig struct {
	PermissionVersion            int                          `json:"permissionVersion"`
	PermissionGroupVersions      []PermissionGroupWithVersion `json:"permissionsGroupVersions"`
	ResourceGroupRolePermissions []struct {
		ExcludedActions     []string `json:"excludedActions"`
		ExcludedDataActions []string `json:"excludedDataActions"`
		IncludedActions     []string `json:"includedActions"`
		IncludedDataActions []string `json:"includedDataActions"`
	} `json:"resourceGroupRolePermissions"`
	RolePermissions []struct {
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

	// The GraphQL query does not accept a null value for permission groups,
	// but an empty array is accepted.
	if feature.PermissionGroups == nil {
		feature.PermissionGroups = []core.PermissionGroup{}
	}

	query := azureCloudAccountPermissionConfigQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Feature          string                 `json:"feature"`
		PermissionGroups []core.PermissionGroup `json:"permissionsGroups"`
	}{Feature: feature.Name, PermissionGroups: feature.PermissionGroups})
	if err != nil {
		return PermissionConfig{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result PermissionConfig `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return PermissionConfig{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// PermissionUpgrade holds the input for the
// UpgradeCloudAccountPermissionsWithoutOAuth function.
type PermissionUpgrade struct {
	CloudAccountID uuid.UUID // RSC cloud account ID.
	Feature        core.Feature
	ResourceGroup  *ResourceGroup // Optional, only for Azure SQL DB resource group upgrades.
}

// UpgradeCloudAccountPermissionsWithoutOAuth notifies RSC that the permissions
// for the Azure service principal have been updated for the specified RSC cloud
// account id and feature.
// The ResourceGroup field is optional and only required when the Azure SQL DB
// feature is upgraded to support resource groups.
func (a API) UpgradeCloudAccountPermissionsWithoutOAuth(ctx context.Context, in PermissionUpgrade) error {
	a.log.Print(log.Trace)

	query := upgradeAzureCloudAccountPermissionsWithoutOauthQuery
	var queryFeature any = in.Feature.Name
	if len(in.Feature.PermissionGroups) > 0 {
		query = upgradeAzureCloudAccountPermissionsWithoutOauthWithPermissionGroupsQuery
		queryFeature = struct {
			core.Feature
			ResourceGroup *ResourceGroup `json:"resourceGroup,omitempty"`
		}{
			Feature:       in.Feature,
			ResourceGroup: in.ResourceGroup,
		}
	}

	buf, err := a.GQL.Request(ctx, query, struct {
		ID      uuid.UUID `json:"cloudAccountId"`
		Feature any       `json:"feature"`
	}{ID: in.CloudAccountID, Feature: queryFeature})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Status bool `json:"status"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	if !payload.Data.Result.Status {
		return graphql.ResponseError(query, errors.New("update cloud account permissions failed"))
	}

	return nil
}

// StartDisableCloudAccountJob starts a task chain to disable the feature in the
// cloud account with the specified cloud account id. Returns the RSC task chain
// id.
func (a API) StartDisableCloudAccountJob(ctx context.Context, id uuid.UUID, feature core.Feature) (uuid.UUID, error) {
	a.GQL.Log().Print(log.Trace)

	query := startDisableAzureCloudAccountJobQuery
	buf, err := a.GQL.Request(ctx, startDisableAzureCloudAccountJobQuery, struct {
		ID      uuid.UUID `json:"cloudAccountId"`
		Feature string    `json:"feature"`
	}{ID: id, Feature: feature.Name})
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}

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
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}
	if len(payload.Data.Result.Errors) != 0 {
		return uuid.Nil, graphql.ResponseError(query, errors.New(payload.Data.Result.Errors[0].Error))
	}
	if len(payload.Data.Result.JobIDs) != 1 {
		return uuid.Nil, graphql.ResponseError(query, errors.New("expected a single result"))
	}

	return payload.Data.Result.JobIDs[0].JobID, nil
}

// RegionsToCloudAccountRegionEnum converts a slice of Regions to a slice of
// CloudAccountRegionEnums.
func RegionsToCloudAccountRegionEnum(regions []azure.Region) []azure.CloudAccountRegionEnum {
	enums := make([]azure.CloudAccountRegionEnum, 0, len(regions))
	for _, region := range regions {
		enums = append(enums, region.ToCloudAccountRegionEnum())
	}

	return enums
}
