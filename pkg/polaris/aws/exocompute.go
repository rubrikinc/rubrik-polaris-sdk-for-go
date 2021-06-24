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

package aws

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Subnet represents an AWS subnet.
type Subnet struct {
	ID               string
	AvailabilityZone string
}

// ExocomputeConfig represents a single exocompute config.
type ExocomputeConfig struct {
	ID      uuid.UUID
	Region  string
	VPCID   string
	Subnets []Subnet

	// When true Polaris manages the security groups.
	PolarisManaged bool

	// Security group ids of cluster control plane and worker node.
	ClusterSecurityGroupID string
	NodeSecurityGroupID    string
}

// ExoConfigFunc returns an exocompute config initialized from the values
// passed to the function creating the ExoConfigFunc.
type ExoConfigFunc func(ctx context.Context) (aws.ExocomputeConfigCreate, error)

// Unmanaged returns an ExoConfigFunc that initializes an exocompute config
// with security groups managed by the user using the specified values.
func Unmanaged(region, vpcID string, subnets []Subnet, clusterSecurityGroupID, nodeSecurityGroupID string) ExoConfigFunc {
	return func(ctx context.Context) (aws.ExocomputeConfigCreate, error) {
		if len(subnets) != 2 {
			return aws.ExocomputeConfigCreate{}, errors.New("polaris: there should be exactly 2 subnets")
		}

		r, err := aws.ParseRegion(region)
		if err != nil {
			return aws.ExocomputeConfigCreate{}, err
		}

		return aws.ExocomputeConfigCreate{
			Region: r,
			VPCID:  vpcID,
			Subnets: []aws.Subnet{
				{ID: subnets[0].ID, AvailabilityZone: subnets[0].AvailabilityZone},
				{ID: subnets[1].ID, AvailabilityZone: subnets[1].AvailabilityZone},
			},
			IsPolarisManaged:       false,
			ClusterSecurityGroupId: clusterSecurityGroupID,
			NodeSecurityGroupId:    nodeSecurityGroupID,
		}, nil
	}
}

// Managed returns an ExoConfigFunc that initializes an exocompute config with
// security groups managed by Polaris using the specified values.
func Managed(region, vpcID string, subnets []Subnet) ExoConfigFunc {
	return func(ctx context.Context) (aws.ExocomputeConfigCreate, error) {
		if len(subnets) != 2 {
			return aws.ExocomputeConfigCreate{}, errors.New("polaris: there should be exactly 2 subnets")
		}

		r, err := aws.ParseRegion(region)
		if err != nil {
			return aws.ExocomputeConfigCreate{}, err
		}

		return aws.ExocomputeConfigCreate{
			Region: r,
			VPCID:  vpcID,
			Subnets: []aws.Subnet{
				{ID: subnets[0].ID, AvailabilityZone: subnets[0].AvailabilityZone},
				{ID: subnets[1].ID, AvailabilityZone: subnets[1].AvailabilityZone},
			},
			IsPolarisManaged: true,
		}, nil
	}
}

// EnableExocompute enables the exocompute feature for the account with the
// specified id for the given regions. The account must already be added to
// Polaris. Note that to disable the feature the account must be removed.
func (a API) EnableExocompute(ctx context.Context, account AccountFunc, regions ...string) error {
	a.gql.Log().Print(log.Trace, "polaris/aws.AddExocompute")

	if account == nil {
		return errors.New("polaris: account is not allowed to be nil")
	}
	config, err := account(ctx)
	if err != nil {
		return err
	}

	regs, err := aws.ParseRegions(regions)
	if err != nil {
		return err
	}

	akkount, err := a.Account(ctx, AccountID(config.id), core.CloudNativeProtection)
	if err != nil {
		return err
	}

	accountInit, err := aws.Wrap(a.gql).ValidateAndCreateCloudAccount(ctx, akkount.NativeID,
		akkount.Name, core.Exocompute)
	if err != nil {
		return err
	}

	err = aws.Wrap(a.gql).FinalizeCloudAccountProtection(ctx, akkount.NativeID, akkount.Name,
		core.Exocompute, regs, accountInit)
	if err != nil {
		return err
	}

	a.gql.Log().Printf(log.Debug, "updating CloudFormation stack: %v", accountInit.StackName)

	err = awsUpdateStack(ctx, config.config, accountInit.StackName, accountInit.TemplateURL)
	if err != nil {
		return err
	}

	return nil
}

// toExocomputeConfig converts an polaris/graphql/aws exocompute config to an
// polaris/aws exocompute config.
func toExocomputeConfig(config aws.ExocomputeConfig) ExocomputeConfig {
	return ExocomputeConfig{
		ID:     config.ID,
		Region: aws.FormatRegion(config.Region),
		VPCID:  config.VPCID,
		Subnets: []Subnet{
			{ID: config.Subnet1.ID, AvailabilityZone: config.Subnet1.AvailabilityZone},
			{ID: config.Subnet2.ID, AvailabilityZone: config.Subnet2.AvailabilityZone},
		},
		PolarisManaged:         config.IsPolarisManaged,
		ClusterSecurityGroupID: config.ClusterSecurityGroupID,
		NodeSecurityGroupID:    config.NodeSecurityGroupID,
	}
}

// ExocomputeConfig returns the exocompute config with the specified exocompute
// config id.
func (a API) ExocomputeConfig(ctx context.Context, id uuid.UUID) (ExocomputeConfig, error) {
	a.gql.Log().Print(log.Trace, "polaris/aws.ExocomputeConfig")

	selectors, err := aws.Wrap(a.gql).ExocomputeConfigs(ctx, "")
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

	return ExocomputeConfig{}, graphql.ErrNotFound
}

// ExocomputeConfigs returns all exocompute configs for the account with the
// specified id.
func (a API) ExocomputeConfigs(ctx context.Context, id IdentityFunc) ([]ExocomputeConfig, error) {
	a.gql.Log().Print(log.Trace, "polaris/aws.ExocomputeConfigs")

	identity, err := id(ctx)
	if err != nil {
		return nil, err
	}

	var nativeID string
	if identity.internal {
		u, err := uuid.Parse(identity.id)
		if err != nil {
			return nil, err
		}

		account, err := aws.Wrap(a.gql).CloudAccount(ctx, u, core.Exocompute)
		if err != nil {
			return nil, err
		}

		nativeID = account.Account.NativeID
	} else {
		nativeID = identity.id
	}

	selectors, err := aws.Wrap(a.gql).ExocomputeConfigs(ctx, nativeID)
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
	a.gql.Log().Print(log.Trace, "polaris/aws.AddExocomputeConfig")

	exoConfig, err := config(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	account, err := a.Account(ctx, id, core.Exocompute)
	if err != nil {
		return uuid.Nil, err
	}

	exo, err := aws.Wrap(a.gql).CreateExocomputeConfig(ctx, account.ID, exoConfig)
	if err != nil {
		return uuid.Nil, err
	}

	return exo.ID, nil
}

// RemoveExocomputeConfig removes the exocompute config with the specified
// exocompute config id.
func (a API) RemoveExocomputeConfig(ctx context.Context, id uuid.UUID) error {
	a.gql.Log().Print(log.Trace, "polaris/aws.RemoveExocomputeConfig")

	err := aws.Wrap(a.gql).DeleteExocomputeConfig(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
