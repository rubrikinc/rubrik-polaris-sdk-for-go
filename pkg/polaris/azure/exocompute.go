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
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/exocompute"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// ExocomputeConfig represents a single exocompute configuration.
type ExocomputeConfig struct {
	ID                    uuid.UUID // Rubrik exocompute configuration ID.
	Region                string
	SubnetID              string // Azure subnet ID.
	ManagedByRubrik       bool   // When true, Rubrik will manage the security groups.
	PodOverlayNetworkCIDR string
	PodSubnetID           string            // Azure subnet ID.
	HealthCheckStatus     HealthCheckStatus // Health status of the exocompute cluster.
}

// HealthCheckStatus represents the health status of an exocompute cluster.
type HealthCheckStatus struct {
	Status        string
	FailureReason string
	LastUpdatedAt string
	TaskchainID   string
}

// ExoConfigFunc returns an ExoCreateParams object initialized from the values
// passed to the function when creating the ExoConfigFunc.
type ExoConfigFunc func(ctx context.Context) (azure.ExoCreateParams, error)

// Managed returns an ExoConfigFunc that initializes an ExoCreateParams object
// with the specified values.
func Managed(region, subnetID string) ExoConfigFunc {
	return func(ctx context.Context) (azure.ExoCreateParams, error) {
		return azure.ExoCreateParams{
			IsManagedByRubrik: true,
			Region:            azure.RegionFromName(region).ToCloudAccountRegionEnum(),
			SubnetID:          subnetID,
		}, nil
	}
}

// ManagedWithOverlayNetwork returns an ExoConfigFunc that initializes an
// ExoCreateParams object with the specified values.
func ManagedWithOverlayNetwork(region, subnetID, podOverlayNetworkCIDR string) ExoConfigFunc {
	return func(ctx context.Context) (azure.ExoCreateParams, error) {
		return azure.ExoCreateParams{
			IsManagedByRubrik:     true,
			Region:                azure.RegionFromName(region).ToCloudAccountRegionEnum(),
			SubnetID:              subnetID,
			PodOverlayNetworkCIDR: podOverlayNetworkCIDR,
		}, nil
	}
}

// toExocomputeConfig converts an polaris/graphql/azure exocompute config to an
// polaris/azure exocompute config.
func toExocomputeConfig(configID uuid.UUID, config azure.ExoConfig) ExocomputeConfig {
	return ExocomputeConfig{
		ID:                    configID,
		Region:                config.Region.Name(),
		SubnetID:              config.SubnetID,
		ManagedByRubrik:       config.ManagedByRubrik,
		PodOverlayNetworkCIDR: config.PodOverlayNetworkCIDR,
		PodSubnetID:           config.PodSubnetID,
		HealthCheckStatus: HealthCheckStatus{
			Status:        config.HealthCheckStatus.Status,
			FailureReason: config.HealthCheckStatus.FailureReason,
			LastUpdatedAt: config.HealthCheckStatus.LastUpdatedAt,
			TaskchainID:   config.HealthCheckStatus.TaskchainID,
		},
	}
}

// ExocomputeConfig returns the exocompute config with the specified exocompute
// config ID.
func (a API) ExocomputeConfig(ctx context.Context, configID uuid.UUID) (ExocomputeConfig, error) {
	a.log.Print(log.Trace)

	configsForAccounts, err := exocompute.ListConfigurations[azure.ExoConfigsForAccount](ctx, a.client, "")
	if err != nil {
		return ExocomputeConfig{}, fmt.Errorf("failed to get exocompute configs: %s", err)
	}
	for _, configsForAccount := range configsForAccounts {
		for _, config := range configsForAccount.Configs {
			id, err := uuid.Parse(config.ID)
			if err != nil {
				return ExocomputeConfig{}, fmt.Errorf("failed to parse exocompute config id: %s", err)
			}
			if id == configID {
				return toExocomputeConfig(id, config), nil
			}
		}
	}

	return ExocomputeConfig{}, fmt.Errorf("exocompute config %w", graphql.ErrNotFound)
}

// ExocomputeConfigs returns all exocompute configs for the account with the
// specified ID.
func (a API) ExocomputeConfigs(ctx context.Context, id IdentityFunc) ([]ExocomputeConfig, error) {
	a.log.Print(log.Trace)

	nativeID, err := a.toNativeID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get native id: %s", err)
	}

	configsForAccounts, err := exocompute.ListConfigurations[azure.ExoConfigsForAccount](ctx, a.client, nativeID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get exocompute configs: %s", err)
	}
	var exoConfigs []ExocomputeConfig
	for _, configsForAccount := range configsForAccounts {
		for _, config := range configsForAccount.Configs {
			id, err := uuid.Parse(config.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to parse exocompute configuration id: %s", err)
			}
			exoConfigs = append(exoConfigs, toExocomputeConfig(id, config))
		}
	}

	return exoConfigs, nil
}

// AddExocomputeConfig adds the exocompute config to the account with the
// specified id. Returns the id of the added exocompute config.
func (a API) AddExocomputeConfig(ctx context.Context, id IdentityFunc, config ExoConfigFunc) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	exoConfig, err := config(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to lookup exocompute config: %s", err)
	}

	accountID, err := a.toCloudAccountID(ctx, id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get cloud account id: %s", err)
	}

	exoID, err := exocompute.CreateConfiguration[azure.ExoCreateResult](ctx, a.client, accountID, exoConfig)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create exocompute config: %s", err)
	}

	return exoID, nil
}

// RemoveExocomputeConfig removes the exocompute config with the specified
// exocompute config id.
func (a API) RemoveExocomputeConfig(ctx context.Context, configID uuid.UUID) error {
	a.log.Print(log.Trace)

	err := exocompute.DeleteConfiguration[azure.ExoDeleteResult](ctx, a.client, configID)
	if err != nil {
		return fmt.Errorf("failed to remove exocompute config: %s", err)
	}

	return nil
}

// ExocomputeHostAccount returns the exocompute host cloud account ID for the
// specified application cloud account.
func (a API) ExocomputeHostAccount(ctx context.Context, appID IdentityFunc) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	appCloudAccountID, err := a.toCloudAccountID(ctx, appID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get cloud account id: %s", err)
	}

	mappings, err := exocompute.AllCloudAccountMappings(ctx, a.client, core.CloudVendorAzure)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get cloud account exocompute mappings: %s", err)
	}
	for _, mapping := range mappings {
		if mapping.AppCloudAccountID == appCloudAccountID {
			return mapping.HostCloudAccountID, nil
		}
	}

	return uuid.Nil, fmt.Errorf("exocompute account %w", graphql.ErrNotFound)
}

// MapExocompute maps the exocompute application cloud account to the specified
// host cloud account.
func (a API) MapExocompute(ctx context.Context, hostID IdentityFunc, appID IdentityFunc) error {
	a.log.Print(log.Trace)

	hostCloudAccountID, err := a.toCloudAccountID(ctx, hostID)
	if err != nil {
		return fmt.Errorf("failed to get cloud account id: %s", err)
	}

	appCloudAccountID, err := a.toCloudAccountID(ctx, appID)
	if err != nil {
		return fmt.Errorf("failed to get cloud account id: %s", err)
	}

	if err := exocompute.MapCloudAccount[azure.ExoMapResult](ctx, a.client, hostCloudAccountID, appCloudAccountID); err != nil {
		return fmt.Errorf("failed to map exocompute config: %s", err)
	}

	return nil
}

// UnmapExocompute unmaps the exocompute application cloud account with the
// specified cloud account ID.
func (a API) UnmapExocompute(ctx context.Context, appID IdentityFunc) error {
	a.log.Print(log.Trace)

	appCloudAccountID, err := a.toCloudAccountID(ctx, appID)
	if err != nil {
		return fmt.Errorf("failed to get cloud account id: %s", err)
	}

	if err := exocompute.UnmapCloudAccount[azure.ExoUnmapResult](ctx, a.client, appCloudAccountID); err != nil {
		return fmt.Errorf("failed to unmap exocompute config: %s", err)
	}

	return nil
}
