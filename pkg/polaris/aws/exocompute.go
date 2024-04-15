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
	"fmt"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Subnet represents an AWS subnet.
type Subnet struct {
	ID               string
	AvailabilityZone string
}

type HealthCheckStatus struct {
	Status        string
	FailureReason string
	LastUpdatedAt string
	TaskchainID   string
}

// ExocomputeConfig represents a single exocompute config.
type ExocomputeConfig struct {
	ID      uuid.UUID
	Region  string
	VPCID   string
	Subnets []Subnet

	// When true, Rubrik manages the security groups.
	ManagedByRubrik bool

	// Security group ids of cluster control plane and worker node.
	ClusterSecurityGroupID string
	NodeSecurityGroupID    string

	// Health status of the exocompute cluster.
	HealthCheckStatus HealthCheckStatus
}

// ExoConfigFunc returns an exocompute config initialized from the values
// passed to the function creating the ExoConfigFunc.
type ExoConfigFunc func(ctx context.Context, gql *graphql.Client, id uuid.UUID) (aws.ExocomputeConfigCreate, error)

// hasSecurityGroup returns true if a security group with the specified id
// exists.
func hasSecurityGroup(vpc aws.VPC, groupID string) bool {
	for _, group := range vpc.SecurityGroups {
		if group.ID == groupID {
			return true
		}
	}

	return false
}

// findVPC returns the VPC with the specified VPC id.
func findVPC(vpcs []aws.VPC, vpcID string) (aws.VPC, error) {
	for _, vpc := range vpcs {
		if vpc.ID == vpcID {
			return vpc, nil
		}
	}

	return aws.VPC{}, fmt.Errorf("invalid vpc id: %v", vpcID)
}

// findSubnet returns the subnet with the specified subnet id.
func findSubnet(vpc aws.VPC, subnetID string) (aws.Subnet, error) {
	for _, subnet := range vpc.Subnets {
		if subnet.ID == subnetID {
			return aws.Subnet{
				ID:               subnet.ID,
				AvailabilityZone: subnet.AvailabilityZone,
			}, nil
		}
	}

	return aws.Subnet{}, fmt.Errorf("invalid subnet id: %v", subnetID)
}

// Managed returns an ExoConfigFunc that initializes an exocompute config with
// security groups managed by RSC using the specified values.
func Managed(region, vpcID string, subnetIDs []string) ExoConfigFunc {
	return func(ctx context.Context, gql *graphql.Client, id uuid.UUID) (aws.ExocomputeConfigCreate, error) {
		reg, err := aws.ParseRegion(region)
		if err != nil {
			return aws.ExocomputeConfigCreate{}, fmt.Errorf("failed to parse region: %v", err)
		}

		// Validate VPC.
		vpcs, err := aws.Wrap(gql).AllVpcsByRegion(ctx, id, reg)
		if err != nil {
			return aws.ExocomputeConfigCreate{}, fmt.Errorf("failed to get vpcs: %v", err)
		}
		vpc, err := findVPC(vpcs, vpcID)
		if err != nil {
			return aws.ExocomputeConfigCreate{}, fmt.Errorf("failed to find vpc: %v", err)
		}

		// Validate subnets.
		if len(subnetIDs) != 2 {
			return aws.ExocomputeConfigCreate{}, errors.New("there should be exactly 2 subnet ids")
		}
		subnet1, err := findSubnet(vpc, subnetIDs[0])
		if err != nil {
			return aws.ExocomputeConfigCreate{}, fmt.Errorf("failed to find subnet: %v", err)
		}
		subnet2, err := findSubnet(vpc, subnetIDs[1])
		if err != nil {
			return aws.ExocomputeConfigCreate{}, fmt.Errorf("failed to find subnet: %v", err)
		}

		return aws.ExocomputeConfigCreate{
			Region:            reg,
			VPCID:             vpcID,
			Subnets:           []aws.Subnet{subnet1, subnet2},
			IsManagedByRubrik: true,
		}, nil
	}
}

// Unmanaged returns an ExoConfigFunc that initializes an exocompute config
// with security groups managed by the user using the specified values.
func Unmanaged(region, vpcID string, subnetIDs []string, clusterSecurityGroupID, nodeSecurityGroupID string) ExoConfigFunc {
	return func(ctx context.Context, gql *graphql.Client, id uuid.UUID) (aws.ExocomputeConfigCreate, error) {
		reg, err := aws.ParseRegion(region)
		if err != nil {
			return aws.ExocomputeConfigCreate{}, fmt.Errorf("failed to parse region: %v", err)
		}

		// Validate VPC.
		vpcs, err := aws.Wrap(gql).AllVpcsByRegion(ctx, id, reg)
		if err != nil {
			return aws.ExocomputeConfigCreate{}, fmt.Errorf("failed to get vpcs: %v", err)
		}
		vpc, err := findVPC(vpcs, vpcID)
		if err != nil {
			return aws.ExocomputeConfigCreate{}, fmt.Errorf("failed to find vpc: %v", err)
		}

		// Validate subnets.
		if len(subnetIDs) != 2 {
			return aws.ExocomputeConfigCreate{}, errors.New("there should be exactly 2 subnet ids")
		}
		subnet1, err := findSubnet(vpc, subnetIDs[0])
		if err != nil {
			return aws.ExocomputeConfigCreate{}, fmt.Errorf("failed to find subnet: %v", err)
		}
		subnet2, err := findSubnet(vpc, subnetIDs[1])
		if err != nil {
			return aws.ExocomputeConfigCreate{}, fmt.Errorf("failed to find subnet: %v", err)
		}

		// Validate security groups.
		if hasSecurityGroup(vpc, clusterSecurityGroupID) {
			return aws.ExocomputeConfigCreate{},
				fmt.Errorf("invalid cluster security group id: %v", clusterSecurityGroupID)
		}
		if hasSecurityGroup(vpc, nodeSecurityGroupID) {
			return aws.ExocomputeConfigCreate{},
				fmt.Errorf("invalid node security group id: %v", nodeSecurityGroupID)
		}

		return aws.ExocomputeConfigCreate{
			Region:                 reg,
			VPCID:                  vpcID,
			Subnets:                []aws.Subnet{subnet1, subnet2},
			IsManagedByRubrik:      false,
			ClusterSecurityGroupId: clusterSecurityGroupID,
			NodeSecurityGroupId:    nodeSecurityGroupID,
		}, nil
	}
}

// BYOKCluster returns an ExoConfigFunc that initializes an exocompute config
// with a Bring-Your-Own-Kubernetes cluster.
func BYOKCluster(region string) ExoConfigFunc {
	return func(ctx context.Context, gql *graphql.Client, id uuid.UUID) (aws.ExocomputeConfigCreate, error) {
		reg, err := aws.ParseRegion(region)
		if err != nil {
			return aws.ExocomputeConfigCreate{}, fmt.Errorf("failed to parse region: %v", err)
		}

		return aws.ExocomputeConfigCreate{Region: reg}, nil
	}
}

// toExocomputeConfig converts a polaris/graphql/aws exocompute config to a
// polaris/aws exocompute config.
func toExocomputeConfig(config aws.ExocomputeConfig) (ExocomputeConfig, error) {
	id, err := uuid.Parse(config.ID)
	if err != nil {
		return ExocomputeConfig{}, fmt.Errorf("invalid exocompute configuration id: %s", err)
	}

	var subnets []Subnet
	for _, s := range []aws.Subnet{config.Subnet1, config.Subnet2} {
		if s.ID == "" || s.AvailabilityZone == "" {
			break
		}
		subnets = append(subnets, Subnet{ID: s.ID, AvailabilityZone: s.AvailabilityZone})
	}
	return ExocomputeConfig{
		ID:                     id,
		Region:                 aws.FormatRegion(config.Region),
		VPCID:                  config.VPCID,
		Subnets:                subnets,
		ManagedByRubrik:        config.IsManagedByRubrik,
		ClusterSecurityGroupID: config.ClusterSecurityGroupID,
		NodeSecurityGroupID:    config.NodeSecurityGroupID,
		HealthCheckStatus: HealthCheckStatus{
			Status:        config.HealthCheckStatus.Status,
			FailureReason: config.HealthCheckStatus.FailureReason,
			LastUpdatedAt: config.HealthCheckStatus.LastUpdatedAt,
			TaskchainID:   config.HealthCheckStatus.TaskchainID,
		},
	}, nil
}

// ExocomputeConfig returns the exocompute config with the specified exocompute
// config id.
func (a API) ExocomputeConfig(ctx context.Context, configID uuid.UUID) (ExocomputeConfig, error) {
	a.log.Print(log.Trace)

	configsForAccounts, err := aws.Wrap(a.client).ExocomputeConfigs(ctx, "")
	if err != nil {
		return ExocomputeConfig{}, fmt.Errorf("failed to get exocompute configs for account: %s", err)
	}

	exoID := configID.String()
	for _, configsForAccount := range configsForAccounts {
		for _, config := range configsForAccount.Configs {
			if config.ID == exoID {
				return toExocomputeConfig(config)
			}
		}
	}

	return ExocomputeConfig{}, fmt.Errorf("exocompute config %w", graphql.ErrNotFound)
}

// ExocomputeConfigs returns all exocompute configs for the account with the
// specified id.
func (a API) ExocomputeConfigs(ctx context.Context, id IdentityFunc) ([]ExocomputeConfig, error) {
	a.log.Print(log.Trace)

	nativeID, err := a.toNativeID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get native id: %s", err)
	}

	configsForAccounts, err := aws.Wrap(a.client).ExocomputeConfigs(ctx, nativeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exocompute configs for account: %s", err)
	}

	var exoConfigs []ExocomputeConfig
	for _, configsForAccount := range configsForAccounts {
		for _, config := range configsForAccount.Configs {
			exoConfig, err := toExocomputeConfig(config)
			if err != nil {
				return nil, err
			}
			exoConfigs = append(exoConfigs, exoConfig)
		}
	}

	return exoConfigs, nil
}

// AddExocomputeConfig adds the exocompute config to the account with the
// specified id. Returns the id of the added exocompute config.
func (a API) AddExocomputeConfig(ctx context.Context, id IdentityFunc, config ExoConfigFunc) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	accountID, err := a.toCloudAccountID(ctx, id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get cloud account id: %s", err)
	}

	exoConfig, err := config(ctx, a.client, accountID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to lookup exocompute config: %s", err)
	}

	exo, err := aws.Wrap(a.client).CreateExocomputeConfig(ctx, accountID, exoConfig)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create exocompute config: %s", err)
	}

	exoID, err := uuid.Parse(exo.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid exocompute configuration id: %s", err)
	}

	return exoID, nil
}

func (a API) UpdateExocomputeConfig(ctx context.Context, id IdentityFunc, config ExoConfigFunc) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	accountID, err := a.toCloudAccountID(ctx, id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get cloud account id: %s", err)
	}

	exoConfig, err := config(ctx, a.client, accountID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to lookup exocompute config: %s", err)
	}

	exo, err := aws.Wrap(a.client).UpdateExocomputeConfig(ctx, accountID, exoConfig)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create exocompute config: %s", err)
	}

	exoID, err := uuid.Parse(exo.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid exocompute configuration id: %s", err)
	}

	return exoID, nil
}

// RemoveExocomputeConfig removes the exocompute config with the specified
// exocompute config id.
func (a API) RemoveExocomputeConfig(ctx context.Context, configID uuid.UUID) error {
	a.log.Print(log.Trace)

	err := aws.Wrap(a.client).DeleteExocomputeConfig(ctx, configID)
	if err != nil {
		return fmt.Errorf("failed to remove exocompute config: %v", err)
	}

	return nil
}

// ExocomputeHostAccount returns the exocompute host account for the exocompute
// application account with the specified id.
func (a API) ExocomputeHostAccount(ctx context.Context, appID IdentityFunc) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	cloudAccountID, err := a.toCloudAccountID(ctx, appID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get cloud account id: %s", err)
	}

	configsForAccounts, err := aws.Wrap(a.client).ExocomputeConfigs(ctx, "")
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get exocompute configs for account: %s", err)
	}

	for _, configsForAccount := range configsForAccounts {
		for _, mappedAccount := range configsForAccount.MappedAccounts {
			if mappedAccount.ID == cloudAccountID {
				return configsForAccount.Account.ID, nil
			}
		}
	}

	return uuid.Nil, fmt.Errorf("exocompute account %w", graphql.ErrNotFound)
}

// ExocomputeApplicationAccounts returns the exocompute application accounts for
// the exocompute host account with the specified id.
func (a API) ExocomputeApplicationAccounts(ctx context.Context, hostID IdentityFunc) ([]uuid.UUID, error) {
	a.log.Print(log.Trace)

	nativeID, err := a.toNativeID(ctx, hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get native id: %s", err)
	}

	configsForAccounts, err := aws.Wrap(a.client).ExocomputeConfigs(ctx, nativeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exocompute configs for account: %s", err)
	}

	var mappedAccounts []uuid.UUID
	for _, configsForAccount := range configsForAccounts {
		if configsForAccount.Account.NativeID == nativeID {
			for _, mappedAccount := range configsForAccount.MappedAccounts {
				mappedAccounts = append(mappedAccounts, mappedAccount.ID)
			}
		}
	}
	if len(mappedAccounts) == 0 {
		return nil, fmt.Errorf("exocompute mapped accounts %w", graphql.ErrNotFound)
	}

	return mappedAccounts, nil
}

// MapExocompute maps the exocompute application account to the specified
// exocompute host account.
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

	if err = aws.Wrap(a.client).MapCloudAccountExocomputeAccount(ctx, hostCloudAccountID, []uuid.UUID{appCloudAccountID}); err != nil {
		return fmt.Errorf("failed to map exocompute config: %v", err)
	}

	return nil
}

// UnmapExocompute unmaps the exocompute application account with the specified
// id.
func (a API) UnmapExocompute(ctx context.Context, appID IdentityFunc) error {
	a.log.Print(log.Trace)

	cloudAccountID, err := a.toCloudAccountID(ctx, appID)
	if err != nil {
		return fmt.Errorf("failed to get cloud account id: %s", err)
	}

	if err := aws.Wrap(a.client).UnmapCloudAccountExocomputeAccount(ctx, []uuid.UUID{cloudAccountID}); err != nil {
		return fmt.Errorf("failed to unmap exocompute config: %v", err)
	}

	return nil
}

// AddClusterToExocomputeConfig adds the named cluster to specified exocompute
// configration. The cluster ID and connection command are returned.
func (a API) AddClusterToExocomputeConfig(ctx context.Context, configID uuid.UUID, clusterName string) (uuid.UUID, string, error) {
	a.log.Print(log.Trace)

	clusterID, cmd, err := aws.Wrap(a.client).ConnectExocomputeCluster(ctx, configID, clusterName)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to connect exocompute cluster: %v", err)
	}

	return clusterID, cmd, nil
}

// RemoveExocomputeCluster removes the exocompute cluster with the specified ID.
func (a API) RemoveExocomputeCluster(ctx context.Context, clusterID uuid.UUID) error {
	a.log.Print(log.Trace)

	if err := aws.Wrap(a.client).DisconnectExocomputeCluster(ctx, clusterID); err != nil {
		return fmt.Errorf("failed to disconnect exocompute cluster: %v", err)
	}

	return nil
}
