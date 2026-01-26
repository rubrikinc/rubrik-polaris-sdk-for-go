// Copyright 2026 Rubrik, Inc.
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

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/exocompute"
	gqlgcp "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// GCPConfiguration holds a single GCP exocompute configuration.
type GCPConfiguration struct {
	exocompute.GCPConfiguration
	CloudAccountID uuid.UUID
}

// RegionalConfig holds the configuration for a GCP region. A GCP Exocompute
// configuration consist of a set of regional configurations.
type RegionalConfig struct {
	Region         gqlgcp.Region
	SubnetName     string
	VPCNetworkName string
}

// GCPConfigurationsByCloudAccountID returns all GCP exocompute configurations
// for the cloud account with the specified ID.
func (a API) GCPConfigurationsByCloudAccountID(ctx context.Context, cloudAccountID uuid.UUID) ([]GCPConfiguration, error) {
	a.log.Print(log.Trace)

	accountConfigs, err := exocompute.ListConfigurations(ctx, a.client, exocompute.GCPConfigurationsFilter{
		CloudAccountID: cloudAccountID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get exocompute configurations for cloud account %s: %s", cloudAccountID, err)
	}

	var configs []GCPConfiguration
	for _, config := range accountConfigs.Configs {
		configs = append(configs, GCPConfiguration{
			CloudAccountID:   cloudAccountID,
			GCPConfiguration: config,
		})
	}

	return configs, nil
}

// GCPConfigurations returns all GCP exocompute configurations.
func (a API) GCPConfigurations(ctx context.Context) ([]GCPConfiguration, error) {
	a.log.Print(log.Trace)

	cloudAccounts, err := gcp.WrapGQL(a.client).Projects(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get cloud accounts: %s", err)
	}

	var configs []GCPConfiguration
	for _, cloudAccount := range cloudAccounts {
		accountConfigs, err := a.GCPConfigurationsByCloudAccountID(ctx, cloudAccount.ID)
		if err != nil {
			return nil, err
		}
		configs = append(configs, accountConfigs...)
	}

	return configs, nil
}

// UpdateGCPConfiguration updates the GCP exocompute configuration for the cloud
// account with the specified ID.
// Note, it's not possible to add or remove Exocompute configurations for a
// cloud account. Instead, it's only possible to update the single configuration
// automatically created for the cloud account.
func (a API) UpdateGCPConfiguration(ctx context.Context, cloudAccountID uuid.UUID, regionalConfigs []RegionalConfig, triggerHealthCheck bool) error {
	a.log.Print(log.Trace)

	configs := make([]exocompute.GCPRegionalConfig, 0, len(regionalConfigs))
	for _, config := range regionalConfigs {
		configs = append(configs, exocompute.GCPRegionalConfig{
			Region:         config.Region.ToCloudAccountRegionEnum(),
			SubnetName:     config.SubnetName,
			VPCNetworkName: config.VPCNetworkName,
		})
	}

	// Note, the Exocompute GraphQL API for GCP does not return the config ID.
	_, err := exocompute.UpdateConfiguration(ctx, a.client, exocompute.UpdateGCPConfigurationParams{
		CloudAccountID:     cloudAccountID,
		RegionalConfigs:    configs,
		TriggerHealthCheck: triggerHealthCheck,
	})
	if err != nil {
		return fmt.Errorf("failed to create exocompute configuration for cloud account %s: %s", cloudAccountID, err)
	}

	return nil
}

// RemoveGCPConfiguration removes the GCP Exocompute configuration for the cloud
// account with the specified ID.
// Note, the Exocompute configuration is not actually removed, only the regional
// configurations are cleared.
func (a API) RemoveGCPConfiguration(ctx context.Context, cloudAccountID uuid.UUID) error {
	a.log.Print(log.Trace)

	err := exocompute.DeleteConfiguration(ctx, a.client, exocompute.DeleteGCPConfigurationParams{
		CloudAccountID: cloudAccountID,
	})
	if err != nil {
		return fmt.Errorf("failed to remove exocompute configuration for cloud account %s: %s", cloudAccountID, err)
	}

	return nil
}
