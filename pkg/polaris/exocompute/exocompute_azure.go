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

package exocompute

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/exocompute"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// AzureConfiguration holds a single Azure exocompute configuration.
type AzureConfiguration struct {
	exocompute.AzureConfiguration
	CloudAccountID uuid.UUID
}

// AzureConfigurationByID returns the Azure exocompute configuration with the
// specified ID. If a configuration with the specified ID isn't found,
// graphql.ErrNotFound is returned.
func (a API) AzureConfigurationByID(ctx context.Context, configID uuid.UUID) (AzureConfiguration, error) {
	a.log.Print(log.Trace)

	configs, err := a.AzureConfigurations(ctx)
	if err != nil {
		return AzureConfiguration{}, err
	}
	for _, config := range configs {
		if config.ID == configID {
			return config, nil
		}
	}

	return AzureConfiguration{}, fmt.Errorf("exocompute configuration %s %w", configID, graphql.ErrNotFound)
}

// AzureConfigurationsByCloudAccountID returns all Azure exocompute
// configurations for the cloud account with the specified ID.
func (a API) AzureConfigurationsByCloudAccountID(ctx context.Context, cloudAccountID uuid.UUID) ([]AzureConfiguration, error) {
	a.log.Print(log.Trace)

	var configs []AzureConfiguration
	configsForAccounts, err := exocompute.ListConfigurations(ctx, a.client, exocompute.AzureConfigurationsFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to get exocompute configurations: %s", err)
	}
	i := slices.IndexFunc(configsForAccounts, func(configsForAccount exocompute.AzureConfigurationsForCloudAccount) bool {
		return configsForAccount.CloudAccount.ID == cloudAccountID
	})
	if i != -1 {
		for _, config := range configsForAccounts[i].Configs {
			configs = append(configs, AzureConfiguration{
				CloudAccountID:     cloudAccountID,
				AzureConfiguration: config,
			})
		}
	}

	return configs, nil
}

// AzureConfigurations returns all Azure exocompute configurations.
func (a API) AzureConfigurations(ctx context.Context) ([]AzureConfiguration, error) {
	a.log.Print(log.Trace)

	configsForAccounts, err := exocompute.ListConfigurations(ctx, a.client, exocompute.AzureConfigurationsFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to get exocompute configurations: %s", err)
	}

	var configs []AzureConfiguration
	for _, configsForAccount := range configsForAccounts {
		for _, config := range configsForAccount.Configs {
			configs = append(configs, AzureConfiguration{
				CloudAccountID:     configsForAccount.CloudAccount.ID,
				AzureConfiguration: config,
			})
		}
	}

	return configs, nil
}

// AzureConfigurationFunc returns an CreateAzureConfigurationParams object
// initialized from the values passed to the function.
type AzureConfigurationFunc func(ctx context.Context, cloudAccountID uuid.UUID) (exocompute.CreateAzureConfigurationParams, error)

// AzureManaged returns an AzureConfigurationFunc which initializes a
// CreateAzureConfigurationParams object with the specified values.
func AzureManaged(region azure.Region, subnetID string, triggerHealthCheck bool) AzureConfigurationFunc {
	return func(ctx context.Context, cloudAccountID uuid.UUID) (exocompute.CreateAzureConfigurationParams, error) {
		return exocompute.CreateAzureConfigurationParams{
			CloudAccountID:     cloudAccountID,
			IsManagedByRubrik:  true,
			Region:             region.ToCloudAccountRegionEnum(),
			SubnetID:           subnetID,
			TriggerHealthCheck: triggerHealthCheck,
		}, nil
	}
}

// AzureManagedWithOverlayNetwork returns an AzureConfigurationFunc which
// initializes a CreateAzureConfigurationParams object with the specified
// values.
func AzureManagedWithOverlayNetwork(region azure.Region, subnetID, podOverlayNetworkCIDR string, triggerHealthCheck bool) AzureConfigurationFunc {
	return func(ctx context.Context, cloudAccountID uuid.UUID) (exocompute.CreateAzureConfigurationParams, error) {
		return exocompute.CreateAzureConfigurationParams{
			CloudAccountID:        cloudAccountID,
			IsManagedByRubrik:     true,
			Region:                region.ToCloudAccountRegionEnum(),
			SubnetID:              subnetID,
			PodOverlayNetworkCIDR: podOverlayNetworkCIDR,
			TriggerHealthCheck:    triggerHealthCheck,
		}, nil
	}
}

// AzureBYOKCluster returns an AzureConfigurationFunc which initializes an
// exocompute configuration with a Bring-Your-Own-Kubernetes cluster.
func AzureBYOKCluster(region azure.Region) AzureConfigurationFunc {
	return func(ctx context.Context, cloudAccountID uuid.UUID) (exocompute.CreateAzureConfigurationParams, error) {
		return exocompute.CreateAzureConfigurationParams{
			CloudAccountID:    cloudAccountID,
			IsManagedByRubrik: false,
			Region:            region.ToCloudAccountRegionEnum(),
		}, nil
	}
}

// AddAzureConfiguration adds the exocompute configuration to the cloud account
// with the specified ID. Returns the ID of the added exocompute configuration.
func (a API) AddAzureConfiguration(ctx context.Context, cloudAccountID uuid.UUID, config AzureConfigurationFunc) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	exoConfig, err := config(ctx, cloudAccountID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse exocompute configuration: %s", err)
	}

	configID, err := exocompute.CreateConfiguration(ctx, a.client, exoConfig)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create exocompute configuration for cloud account %s: %s", cloudAccountID, err)
	}

	return configID, nil
}

// RemoveAzureConfiguration removes the Azure exocompute configuration with the
// specified ID.
func (a API) RemoveAzureConfiguration(ctx context.Context, configID uuid.UUID) error {
	a.log.Print(log.Trace)

	params := exocompute.DeleteAzureConfigurationParams{ConfigID: configID}
	err := exocompute.DeleteConfiguration(ctx, a.client, params)
	if err != nil {
		return fmt.Errorf("failed to remove exocompute configuration %s: %s", configID, err)
	}

	return nil
}

// AzureAppCloudAccounts returns the Azure exocompute application cloud account
// IDs for the specified host cloud account.
func (a API) AzureAppCloudAccounts(ctx context.Context, hostCloudAccountID uuid.UUID) ([]uuid.UUID, error) {
	a.log.Print(log.Trace)

	mappings, err := exocompute.ListCloudAccountMappings(ctx, a.client, core.CloudVendorAzure)
	if err != nil {
		return nil, fmt.Errorf("failed to get exocompute cloud account mappings: %s", err)
	}

	var appCloudAccountIDs []uuid.UUID
	for _, mapping := range mappings {
		if mapping.HostCloudAccountID == hostCloudAccountID {
			appCloudAccountIDs = append(appCloudAccountIDs, mapping.AppCloudAccountID)
		}
	}

	return appCloudAccountIDs, nil
}

// AzureHostCloudAccount returns the Azure exocompute host cloud account ID for
// the specified application cloud account.
func (a API) AzureHostCloudAccount(ctx context.Context, appCloudAccountID uuid.UUID) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	mappings, err := exocompute.ListCloudAccountMappings(ctx, a.client, core.CloudVendorAzure)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get exocompute cloud account mappings: %s", err)
	}
	for _, mapping := range mappings {
		if mapping.AppCloudAccountID == appCloudAccountID {
			return mapping.HostCloudAccountID, nil
		}
	}

	return uuid.Nil, fmt.Errorf("host cloud account for application cloud account %s %w", appCloudAccountID, graphql.ErrNotFound)
}

// MapAzureCloudAccount maps the Azure exocompute application cloud account to
// the specified Azure host cloud account.
func (a API) MapAzureCloudAccount(ctx context.Context, appCloudAccountID, hostCloudAccountID uuid.UUID) error {
	a.log.Print(log.Trace)

	params := exocompute.MapAzureCloudAccountsParams{AppCloudAccountIDs: []uuid.UUID{appCloudAccountID}, HostCloudAccountID: hostCloudAccountID}
	if err := exocompute.MapCloudAccounts(ctx, a.client, params); err != nil {
		return fmt.Errorf("failed to map application cloud account %s to host cloud account %s: %s", appCloudAccountID, hostCloudAccountID, err)
	}

	return nil
}

// UnmapAzureCloudAccount unmaps the Azure exocompute application cloud account with
// the specified cloud account ID.
func (a API) UnmapAzureCloudAccount(ctx context.Context, appCloudAccountID uuid.UUID) error {
	a.log.Print(log.Trace)

	params := exocompute.UnmapAzureCloudAccountsParams{AppCloudAccountIDs: []uuid.UUID{appCloudAccountID}}
	if err := exocompute.UnmapCloudAccounts(ctx, a.client, params); err != nil {
		return fmt.Errorf("failed to unmap application cloud account %s: %s", appCloudAccountID, err)
	}

	return nil
}

// AzureClusterConnection returns information about an Azure exocompute cluster
// connection.
func (a API) AzureClusterConnection(ctx context.Context, clusterName string, configID uuid.UUID) (exocompute.AzureClusterConnectionResult, error) {
	a.log.Print(log.Trace)

	params := exocompute.AzureClusterConnectionParams{ClusterName: clusterName, CloudType: "AZURE", ConfigID: configID}
	info, err := exocompute.ClusterConnection(ctx, a.client, params)
	if err != nil {
		return exocompute.AzureClusterConnectionResult{}, fmt.Errorf("failed to get cluster connection info for %q: %s", clusterName, err)
	}

	return info, nil
}

// ConnectAzureCluster connects the named Azure exocompute cluster to the
// specified exocompute configuration. Returns the cluster ID and information
// about the connection.
func (a API) ConnectAzureCluster(ctx context.Context, clusterName string, configID uuid.UUID) (uuid.UUID, exocompute.AzureClusterConnectionResult, error) {
	a.log.Print(log.Trace)

	params := exocompute.ConnectAzureClusterParams{ClusterName: clusterName, CloudType: "AZURE", ConfigID: configID}
	info, err := exocompute.ConnectCluster(ctx, a.client, params)
	if err != nil {
		return uuid.Nil, exocompute.AzureClusterConnectionResult{}, fmt.Errorf("failed to connect exocompute cluster %q: %s", clusterName, err)
	}

	return info.ClusterID, info.AzureClusterConnectionResult, nil
}

// DisconnectAzureCluster disconnects the Azure exocompute cluster with the
// specified cluster ID.
func (a API) DisconnectAzureCluster(ctx context.Context, clusterID uuid.UUID) error {
	a.log.Print(log.Trace)

	params := exocompute.DisconnectAzureClusterParams{ClusterID: clusterID, CloudType: "AZURE"}
	if err := exocompute.DisconnectCluster(ctx, a.client, params); err != nil {
		return fmt.Errorf("failed to disconnect exocompute cluster %s: %s", clusterID, err)
	}

	return nil
}
