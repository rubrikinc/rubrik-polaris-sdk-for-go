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

package graphql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// AzureCloud
type AzureCloud string

const (
	AzureChina  AzureCloud = "AZURECHINACLOUD"
	AzurePublic AzureCloud = "AZUREPUBLICCLOUD"
)

// AzureSubscriptionIn
type AzureSubscriptionIn struct {
	ID   string `json:"nativeId"`
	Name string `json:"name"`
}

type AzureSubscriptionIn2 struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AzureSubscriptionStatus
type AzureSubscriptionStatus struct {
	SubscriptionID       string `json:"azureSubscriptionRubrikId"`
	SubscriptionNativeID string `json:"azureSubscriptionNativeId"`
	Error                string `json:"error"`
}

// AzureAddCloudAccountWithoutOAuth adds the Azure Subscription cloud account
// for given feature without OAuth.
func (c *Client) AzureAddCloudAccountWithoutOAuth(ctx context.Context, cloud AzureCloud, tenantDomain string,
	regions []AzureRegion, feature CloudAccountFeature, subscriptions []AzureSubscriptionIn, policyVersion int) (string, []AzureSubscriptionStatus, error) {
	c.log.Print(log.Trace, "graphql.Client.AzureAddCloudAccountWithoutOAuth")

	query := azureAddCloudAccountWithoutOauthQuery
	if VersionOlderThan(c.Version, "master-40839", "v20210810") {
		query = azureAddCloudAccountWithoutOauthV0Query
	}
	buf, err := c.Request(ctx, query, struct {
		Cloud         AzureCloud            `json:"azure_cloud_type"`
		TenantDomain  string                `json:"azure_tenant_domain_name"`
		Regions       []AzureRegion         `json:"azure_regions"`
		Feature       CloudAccountFeature   `json:"feature"`
		Subscriptions []AzureSubscriptionIn `json:"azure_subscriptions"`
		PolicyVersion int                   `json:"azure_policy_version"`
	}{Cloud: cloud, TenantDomain: tenantDomain, Regions: regions, Feature: feature, Subscriptions: subscriptions, PolicyVersion: policyVersion})
	if err != nil {
		return "", nil, err
	}

	c.log.Printf(log.Debug, "%s(%q, %q, %q, %q, %q, %d): %s", queryName(query), cloud, tenantDomain, regions, feature, subscriptions,
		policyVersion, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				TenantID string                    `json:"tenantId"`
				Status   []AzureSubscriptionStatus `json:"status"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", nil, err
	}
	if len(payload.Data.Result.Status) == 0 {
		return "", nil, errors.New("polaris: add returned no status")
	}

	for _, status := range payload.Data.Result.Status {
		if status.Error != "" {
			return payload.Data.Result.TenantID, payload.Data.Result.Status, fmt.Errorf("polaris: add failed: %s", status.Error)
		}
	}

	return payload.Data.Result.TenantID, payload.Data.Result.Status, nil
}

// AzureDeleteStatus
type AzureDeleteStatus struct {
	SubscriptionID string `json:"azureSubscriptionNativeId"`
	Success        bool   `json:"isSuccess"`
	Error          string `json:"error"`
}

// AzureDeleteCloudAccountWithoutOAuth delete the Azure subscriptions cloud
// account for given feature without OAuth.
func (c *Client) AzureDeleteCloudAccountWithoutOAuth(ctx context.Context, subscriptionIDs []string, feature CloudAccountFeature) ([]AzureDeleteStatus, error) {
	c.log.Print(log.Trace, "graphql.Client.AzureDeleteCloudAccountWithoutOAuth")

	query := azureDeleteCloudAccountWithoutOauthQuery
	if VersionOlderThan(c.Version, "master-40839", "v20210810") {
		query = azureDeleteCloudAccountWithoutOauthV0Query
	}
	buf, err := c.Request(ctx, query, struct {
		SubscriptionIDs []string            `json:"azure_subscription_ids"`
		Feature         CloudAccountFeature `json:"feature"`
	}{SubscriptionIDs: subscriptionIDs, Feature: feature})
	if err != nil {
		return nil, err
	}

	c.log.Printf(log.Debug, "%s(%v, %q): %s", queryName(query), subscriptionIDs, feature, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				Status []AzureDeleteStatus `json:"status"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, err
	}
	if len(payload.Data.Result.Status) == 0 {
		return nil, errors.New("polaris: delete returned no status")
	}

	for _, status := range payload.Data.Result.Status {
		if !status.Success {
			return payload.Data.Result.Status, fmt.Errorf("polaris: delete failed: %s", status.Error)
		}
	}

	return payload.Data.Result.Status, nil
}

// AzureUpdate
type AzureUpdateStatus struct {
	SubscriptionID string `json:"azureSubscriptionNativeId"`
	Success        bool   `json:"isSuccess"`
}

// AzureUpdateCloudAccount update names of the Azure subscriptions and regions
// for the given feature.
func (c *Client) AzureUpdateCloudAccount(ctx context.Context, feature CloudAccountFeature, regionsToAdd, regionsToRemove []AzureRegion, subscriptions []AzureSubscriptionIn2) ([]AzureUpdateStatus, error) {
	c.log.Print(log.Trace, "graphql.Client.AzureUpdateCloudAccount")

	query := azureUpdateCloudAccountQuery
	if VersionOlderThan(c.Version, "master-40839", "v20210810") {
		query = azureUpdateCloudAccountV0Query
	}
	buf, err := c.Request(ctx, query, struct {
		Feature         CloudAccountFeature    `json:"feature"`
		RegionsToAdd    []AzureRegion          `json:"regions_to_add,omitempty"`
		RegionsToRemove []AzureRegion          `json:"regions_to_remove,omitempty"`
		Subscriptions   []AzureSubscriptionIn2 `json:"subscriptions"`
	}{Feature: feature, RegionsToAdd: regionsToAdd, RegionsToRemove: regionsToRemove, Subscriptions: subscriptions})
	if err != nil {
		return nil, err
	}

	c.log.Printf(log.Debug, "%s(%q, %v, %v %v): %s", queryName(query), feature, regionsToAdd, regionsToRemove,
		subscriptions, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				Status []AzureUpdateStatus `json:"status"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, err
	}
	if len(payload.Data.Result.Status) == 0 {
		return nil, errors.New("polaris: update returned no status")
	}

	for _, status := range payload.Data.Result.Status {
		if !status.Success {
			return payload.Data.Result.Status, errors.New("polaris: update failed")
		}
	}

	return payload.Data.Result.Status, nil
}

// AzureSubscription
type AzureSubscription struct {
	ID            string `json:"id"`
	NativeID      string `json:"nativeId"`
	Name          string `json:"name"`
	FeatureDetail struct {
		Feature CloudAccountFeature `json:"feature"`
		Status  CloudAccountStatus  `json:"status"`
		Regions []AzureRegion       `json:"regions"`
	} `json:"featureDetail"`
}

// AzureTenant
type AzureTenant struct {
	Cloud             AzureCloud          `json:"cloutType"`
	ID                string              `json:"azureCloudAccountTenantRubrikId"`
	DomainName        string              `json:"domainName"`
	SubscriptionCount int                 `json:"subscriptionCount"`
	Subscriptions     []AzureSubscription `json:"subscriptions"`
}

// AzureCloudAccountTenants return all tenants for the specified feature. If
// includeSubscription is true all subscriptions for each tenant is included.
func (c *Client) AzureCloudAccountTenants(ctx context.Context, feature CloudAccountFeature, includeSubscriptions bool) ([]AzureTenant, error) {
	c.log.Print(log.Trace, "graphql.Client.AzureCloudAccountTenants")

	query := azureAllCloudAccountTenantsQuery
	switch {
	case VersionOlderThan(c.Version, "master-40839", "v20210810"):
		query = azureAllCloudAccountTenantsV0Query
	case VersionOlderThan(c.Version, "master-41103", "v20210817"):
		query = azureAllCloudAccountTenantsV1Query
	}
	buf, err := c.Request(ctx, query, struct {
		Feature              CloudAccountFeature `json:"feature"`
		IncludeSubscriptions bool                `json:"include_subscriptions"`
	}{Feature: feature, IncludeSubscriptions: includeSubscriptions})
	if err != nil {
		return nil, err
	}

	c.log.Printf(log.Debug, "%s(%q, %t): %s", queryName(query), feature, includeSubscriptions, string(buf))

	var payload struct {
		Data struct {
			Result []AzureTenant `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, err
	}

	return payload.Data.Result, nil
}

// AzureSetCloudAccountCustomerAppCredentials sets the credentials for the
// customer application for the specified tenant domain. If the tenant domain
// is empty, set it for all the tenants of the customer.
func (c *Client) AzureSetCloudAccountCustomerAppCredentials(ctx context.Context, cloud AzureCloud, appID, appName, appTenantID, appTenantDomain, appSecretKey string) error {
	c.log.Print(log.Trace, "graphql.Client.AzureSetCloudAccountCustomerAppCredentials")

	query := azureSetCloudAccountCustomerAppCredentialsQuery
	if VersionOlderThan(c.Version, "master-40839", "v20210810") {
		query = azureSetCloudAccountCustomerAppCredentialsV0Query
	}
	buf, err := c.Request(ctx, query, struct {
		Cloud        AzureCloud `json:"azure_cloud_type"`
		ID           string     `json:"azure_app_id"`
		Name         string     `json:"azure_app_name"`
		TenantID     string     `json:"azure_app_tenant_id"`
		TenantDomain string     `json:"azure_tenant_domain_name"`
		SecretKey    string     `json:"azure_app_secret_key"`
	}{Cloud: cloud, ID: appID, Name: appName, TenantID: appTenantID, TenantDomain: appTenantDomain, SecretKey: appSecretKey})
	if err != nil {
		return err
	}

	c.log.Printf(log.Debug, "%s(%q, %q, %q, %q, %q, %q): %s", queryName(query), cloud, appID, appName,
		appTenantID, appTenantDomain, appSecretKey, string(buf))

	var payload struct {
		Data struct {
			Success bool `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return err
	}

	if !payload.Data.Success {
		return errors.New("polaris: failed to set customer credentials")
	}

	return nil
}

// AzureStartDisableNativeSubscriptionProtectionJob deletes an azure
// subscription.
func (c *Client) AzureStartDisableNativeSubscriptionProtectionJob(ctx context.Context, subscriptionID string, deleteSnapshots bool) (TaskChainUUID, error) {
	c.log.Print(log.Trace, "graphql.Client.AzureStartDisableNativeSubscriptionProtectionJob")

	query := azureStartDisableNativeSubscriptionProtectionJobQuery
	if VersionOlderThan(c.Version, "master-40766", "v20210803") {
		query = azureStartDisableNativeSubscriptionProtectionJobV0Query
	}
	buf, err := c.Request(ctx, query, struct {
		SubscriptionID  string `json:"subscription_id"`
		DeleteSnapshots bool   `json:"delete_snapshots"`
	}{SubscriptionID: subscriptionID, DeleteSnapshots: deleteSnapshots})
	if err != nil {
		return "", err
	}

	c.log.Printf(log.Debug, "%s(%q, %t): %s", queryName(query), subscriptionID, deleteSnapshots, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				JobID TaskChainUUID `json:"jobId"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", err
	}

	return payload.Data.Result.JobID, nil
}

// AzureNativeSubscription holds a Polaris native Azure subscription.
type AzureNativeSubscription struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	NativeID      string        `json:"azureSubscriptionNativeId"`
	Status        string        `json:"azureSubscriptionStatus"`
	SLAAssignment SLAAssignment `json:"slaAssignment"`

	ConfiguredSLADomain struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"configuredSlaDomain"`

	EffectiveSLADomain struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"effectiveSlaDomain"`
}

// AzureNativeSubscriptions
func (c *Client) AzureNativeSubscriptions(ctx context.Context, nameFilter string) ([]AzureNativeSubscription, error) {
	c.log.Print(log.Trace, "graphql.Client.AzureNativeSubscriptions")

	query := azureNativeSubscriptionsQuery
	if VersionOlderThan(c.Version, "master-40644", "v20210803") {
		query = azureNativeSubscriptionsV0Query
	}

	var subscriptions []AzureNativeSubscription
	var endCursor string
	for {
		buf, err := c.Request(ctx, query, struct {
			After      string `json:"after,omitempty"`
			NameFilter string `json:"filter,omitempty"`
		}{After: endCursor, NameFilter: nameFilter})
		if err != nil {
			return nil, err
		}

		c.log.Printf(log.Debug, "%s(%q): %s", queryName(query), nameFilter, string(buf))

		var payload struct {
			Data struct {
				Result struct {
					Count int `json:"count"`
					Edges []struct {
						Node AzureNativeSubscription `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"result"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, err
		}
		for _, subscription := range payload.Data.Result.Edges {
			subscriptions = append(subscriptions, subscription.Node)
		}

		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		endCursor = payload.Data.Result.PageInfo.EndCursor
	}

	return subscriptions, nil
}

// AzurePermissionConfig holds the permissions and version required to enable a
// given feature. IncludedActions refers to actions which should be allowed on
// the Azure role for the subscription. ExcludedActions refers to actions which
// should be explicitly disallowed on the Azure role for the subscription.
type AzurePermissionConfig struct {
	PermissionVersion int `json:"permissionVersion"`
	RolePermissions   []struct {
		ExcludedActions     []string `json:"excludedActions"`
		ExcludedDataActions []string `json:"excludedDataActions"`
		IncludedActions     []string `json:"includedActions"`
		IncludedDataActions []string `json:"includedDataActions"`
	} `json:"rolePermissions"`
}

// AzureCloudAccountPermissionConfig returns the permissions and version
// required to enable the given feature for Azure subscription.
func (c *Client) AzureCloudAccountPermissionConfig(ctx context.Context) (AzurePermissionConfig, error) {
	c.log.Print(log.Trace, "graphql.Client.AzureCloudAccountPermissionConfig")

	buf, err := c.Request(ctx, azureCloudAccountPermissionConfigQuery, struct{}{})
	if err != nil {
		return AzurePermissionConfig{}, err
	}

	c.log.Printf(log.Debug, "azureCloudAccountPermissionConfig(): %s", string(buf))

	var payload struct {
		Data struct {
			PermissionConfig AzurePermissionConfig `json:"azureCloudAccountPermissionConfig"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return AzurePermissionConfig{}, err
	}

	return payload.Data.PermissionConfig, nil
}
