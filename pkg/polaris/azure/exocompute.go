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
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// ExocomputeConfig represents a single exocompute config.
type ExocomputeConfig struct {
	ID       uuid.UUID `json:"configUuid"`
	Region   string    `json:"region"`
	SubnetID string    `json:"subnets"`

	// When true Polaris will manage the security groups.
	PolarisManaged bool `json:"isPolarisManaged"`
}

// ExoConfigFunc returns an exocompute config initialized from the values
// passed to the function creating the ExoConfigFunc.
type ExoConfigFunc func(ctx context.Context) (azure.ExocomputeConfigCreate, error)

// Managed returns an ExoConfigFunc that initializes an exocompute config with
// security groups managed by Polaris using the specified values.
func Managed(region, subnetID string) ExoConfigFunc {
	return func(ctx context.Context) (azure.ExocomputeConfigCreate, error) {
		r, err := azure.ParseRegion(region)
		if err != nil {
			return azure.ExocomputeConfigCreate{}, err
		}

		return azure.ExocomputeConfigCreate{
			Region:           r,
			SubnetID:         subnetID,
			IsPolarisManaged: true,
		}, nil
	}
}

// Unmanaged returns an ExoConfigFunc that initializes an exocompute config
// with security groups managed by the user using the specified values.
func Unmanaged(region, subnetID string) ExoConfigFunc {
	return func(ctx context.Context) (azure.ExocomputeConfigCreate, error) {
		r, err := azure.ParseRegion(region)
		if err != nil {
			return azure.ExocomputeConfigCreate{}, err
		}

		return azure.ExocomputeConfigCreate{
			Region:           r,
			SubnetID:         subnetID,
			IsPolarisManaged: false,
		}, nil
	}
}

// EnableExocompute enables the exocompute feature for the account with the
// specified id for the given regions. The account must already be added to
// Polaris. Note that to disable the feature the account must be removed.
// The returned error will be graphql.ErrAlreadyEnabeled if the feature has
// already been added for the specified account.
func (a API) EnableExocompute(ctx context.Context, id IdentityFunc, regions ...string) error {
	a.gql.Log().Print(log.Trace, "polaris/azure.EnableExocompute")

	regs, err := azure.ParseRegions(regions)
	if err != nil {
		return err
	}

	account, err := a.Subscription(ctx, id, core.Exocompute)
	if err == nil {
		return fmt.Errorf("polaris: feature %w", graphql.ErrAlreadyEnabled)
	}
	if !errors.Is(err, graphql.ErrNotFound) {
		return err
	}

	perms, err := azure.Wrap(a.gql).CloudAccountPermissionConfig(ctx, core.Exocompute)
	if err != nil {
		return err
	}

	account, err = a.Subscription(ctx, id, core.CloudNativeProtection)
	if err != nil {
		return err
	}

	_, err = azure.Wrap(a.gql).CloudAccountAddWithoutOAuth(ctx, azure.PublicCloud, account.NativeID,
		core.Exocompute, account.Name, account.TenantDomain, regs, perms.PermissionVersion)
	if err != nil {
		return err
	}

	return nil
}

// toExocomputeConfig converts an polaris/graphql/azure exocompute config to an
// polaris/azure exocompute config.
func toExocomputeConfig(config azure.ExocomputeConfig) ExocomputeConfig {
	return ExocomputeConfig{
		ID:             config.ID,
		Region:         azure.FormatRegion(config.Region),
		SubnetID:       config.SubnetID,
		PolarisManaged: config.IsPolarisManaged,
	}
}

// ExocomputeConfig returns the exocompute config with the specified exocompute
// config id.
func (a API) ExocomputeConfig(ctx context.Context, id uuid.UUID) (ExocomputeConfig, error) {
	a.gql.Log().Print(log.Trace, "polaris/azure.ExocomputeConfig")

	selectors, err := azure.Wrap(a.gql).ExocomputeConfigs(ctx, "")
	if err != nil {
		return ExocomputeConfig{}, err
	}

	for _, selector := range selectors {
		for _, config := range selector.Configs {
			if config.ID == id {
				return toExocomputeConfig(config), nil
			}
		}
	}

	return ExocomputeConfig{}, nil
}

// ExocomputeConfigs returns all exocompute configs for the account with the
// specified id.
func (a API) ExocomputeConfigs(ctx context.Context, id IdentityFunc) ([]ExocomputeConfig, error) {
	a.gql.Log().Print(log.Trace, "polaris/azure.ExocomputeConfigs")

	nativeID, err := a.toNativeID(ctx, id)
	if err != nil {
		return nil, err
	}

	selectors, err := azure.Wrap(a.gql).ExocomputeConfigs(ctx, nativeID.String())
	if err != nil {
		return nil, err
	}

	var exoConfigs []ExocomputeConfig
	for _, selector := range selectors {
		for _, config := range selector.Configs {
			exoConfigs = append(exoConfigs, toExocomputeConfig(config))
		}
	}

	return exoConfigs, nil
}

// AddExocomputeConfig adds the exocompute config to the account with the
// specified id. Returns the id of the added exocompute config.
func (a API) AddExocomputeConfig(ctx context.Context, id IdentityFunc, config ExoConfigFunc) (uuid.UUID, error) {
	a.gql.Log().Print(log.Trace, "polaris/azure.AddExocomputeConfig")

	exoConfig, err := config(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	accountID, err := a.toCloudAccountID(ctx, id)
	if err != nil {
		return uuid.Nil, err
	}

	exo, err := azure.Wrap(a.gql).ExocomputeAdd(ctx, accountID, exoConfig)
	if err != nil {
		return uuid.Nil, err
	}

	return exo.ID, nil
}

// RemoveExocomputeConfig removes the exocompute config with the specified
// exocompute config id.
func (a API) RemoveExocomputeConfig(ctx context.Context, id uuid.UUID) error {
	a.gql.Log().Print(log.Trace, "polaris/azure.RemoveExocomputeConfig")

	err := azure.Wrap(a.gql).ExocomputeConfigsDelete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
