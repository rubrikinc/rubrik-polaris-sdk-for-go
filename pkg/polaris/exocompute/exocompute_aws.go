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
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/exocompute"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// AWSConfiguration holds a single AWS exocompute configuration.
type AWSConfiguration struct {
	exocompute.AWSConfiguration
	CloudAccountID uuid.UUID
}

// AWSConfigurationByID returns the AWS exocompute configuration with the
// specified ID. If a configuration with the specified ID isn't found,
// graphql.ErrNotFound is returned.
func (a API) AWSConfigurationByID(ctx context.Context, configID uuid.UUID) (AWSConfiguration, error) {
	a.log.Print(log.Trace)

	configs, err := a.AWSConfigurations(ctx)
	if err != nil {
		return AWSConfiguration{}, err
	}
	for _, config := range configs {
		if config.ID == configID {
			return config, nil
		}
	}

	return AWSConfiguration{}, fmt.Errorf("exocompute configuration %s %w", configID, graphql.ErrNotFound)
}

// AWSConfigurationsByCloudAccountID returns all AWS exocompute configurations
// for the cloud account with the specified ID.
func (a API) AWSConfigurationsByCloudAccountID(ctx context.Context, cloudAccountID uuid.UUID) ([]AWSConfiguration, error) {
	a.log.Print(log.Trace)

	var configs []AWSConfiguration
	configsForAccounts, err := exocompute.ListConfigurations(ctx, a.client, exocompute.AWSConfigurationsFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to get exocompute configurations: %s", err)
	}
	i := slices.IndexFunc(configsForAccounts, func(configsForAccount exocompute.AWSConfigurationsForCloudAccount) bool {
		return configsForAccount.CloudAccount.ID == cloudAccountID
	})
	if i != -1 {
		for _, config := range configsForAccounts[i].Configs {
			configs = append(configs, AWSConfiguration{
				CloudAccountID:   cloudAccountID,
				AWSConfiguration: config,
			})
		}
	}

	return configs, nil
}

// AWSConfigurations returns all AWS exocompute configurations.
func (a API) AWSConfigurations(ctx context.Context) ([]AWSConfiguration, error) {
	a.log.Print(log.Trace)

	configsForAccounts, err := exocompute.ListConfigurations(ctx, a.client, exocompute.AWSConfigurationsFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to get exocompute configurations: %s", err)
	}

	var configs []AWSConfiguration
	for _, configsForAccount := range configsForAccounts {
		for _, config := range configsForAccount.Configs {
			configs = append(configs, AWSConfiguration{
				CloudAccountID:   configsForAccount.CloudAccount.ID,
				AWSConfiguration: config,
			})
		}
	}

	return configs, nil
}

// AWSConfig is implemented by types that can be used to create an AWS
// exocompute configuration. Both AWSConfigurationFunc and AWSConfigParams
// implement this interface.
type AWSConfig interface {
	isAWSConfig()
}

// AWSConfigParams holds the parameters for creating an AWS exocompute
// configuration. It replaces the deprecated AWSManaged, AWSUnmanaged, and
// AWSBYOKCluster functions:
//   - RSC-managed security groups: leave ClusterSecurityGroupID and
//     NodeSecurityGroupID empty.
//   - Customer-managed security groups: set ClusterSecurityGroupID and
//     NodeSecurityGroupID to the desired security group IDs.
//   - BYOK (Bring-Your-Own-Kubernetes) cluster: only set the Region field,
//     leave all other fields at their zero values.
type AWSConfigParams struct {
	Region aws.Region
	VPCID  string

	// Only ID, and optionally PodSubnetID, need to be set; AvailabilityZone is
	// looked up automatically.
	Subnets []exocompute.AWSSubnet

	ClusterSecurityGroupID string                        // Omit for RSC-managed security groups
	NodeSecurityGroupID    string                        // Omit for RSC-managed security groups
	OptionalConfig         *exocompute.AWSOptionalConfig // Omit for no optional config
	TriggerHealthCheck     bool
}

func (AWSConfigParams) isAWSConfig() {}

// Deprecated: use AWSConfigParams instead.
//
// AWSConfigurationFunc returns a CreateAWSConfigurationParams object
// initialized from the values passed to the function.
type AWSConfigurationFunc func(ctx context.Context, gql *graphql.Client, id uuid.UUID) (exocompute.CreateAWSConfigurationParams, error)

func (AWSConfigurationFunc) isAWSConfig() {}

// Deprecated: use AWSConfigParams instead.
//
// AWSManaged returns an AWSConfigurationFunc which initializes the
// CreateAWSConfigurationParams object with security groups managed by RSC.
func AWSManaged(region aws.Region, vpcID string, subnetIDs []string, triggerHealthCheck bool) AWSConfigurationFunc {
	return func(ctx context.Context, gql *graphql.Client, cloudAccountID uuid.UUID) (exocompute.CreateAWSConfigurationParams, error) {
		if len(subnetIDs) != 2 {
			return exocompute.CreateAWSConfigurationParams{}, errors.New("there should be exactly 2 subnets")
		}

		return createAWSConfigParams(ctx, gql, cloudAccountID, AWSConfigParams{
			Region:                 region,
			VPCID:                  vpcID,
			Subnets:                []exocompute.AWSSubnet{{ID: subnetIDs[0]}, {ID: subnetIDs[1]}},
			ClusterSecurityGroupID: "",
			NodeSecurityGroupID:    "",
			TriggerHealthCheck:     triggerHealthCheck,
		})
	}
}

// Deprecated: use AWSConfigParams instead.
//
// AWSUnmanaged returns an AWSConfigurationFunc which initializes the
// CreateAWSConfigurationParams object with security groups managed by the user.
func AWSUnmanaged(region aws.Region, vpcID string, subnetIDs []string, clusterSecurityGroupID, nodeSecurityGroupID string, triggerHealthCheck bool) AWSConfigurationFunc {
	return func(ctx context.Context, gql *graphql.Client, cloudAccountID uuid.UUID) (exocompute.CreateAWSConfigurationParams, error) {
		if len(subnetIDs) != 2 {
			return exocompute.CreateAWSConfigurationParams{}, errors.New("there should be exactly 2 subnets")
		}

		return createAWSConfigParams(ctx, gql, cloudAccountID, AWSConfigParams{
			Region:                 region,
			VPCID:                  vpcID,
			Subnets:                []exocompute.AWSSubnet{{ID: subnetIDs[0]}, {ID: subnetIDs[1]}},
			ClusterSecurityGroupID: clusterSecurityGroupID,
			NodeSecurityGroupID:    nodeSecurityGroupID,
			TriggerHealthCheck:     triggerHealthCheck,
		})
	}
}

// Deprecated: use AWSConfigParams instead.
//
// AWSBYOKCluster returns an AWSConfigurationFunc which initializes an
// exocompute configuration with a Bring-Your-Own-Kubernetes cluster.
func AWSBYOKCluster(region aws.Region) AWSConfigurationFunc {
	return func(ctx context.Context, gql *graphql.Client, cloudAccountID uuid.UUID) (exocompute.CreateAWSConfigurationParams, error) {
		return exocompute.CreateAWSConfigurationParams{
			CloudAccountID: cloudAccountID,
			Region:         region.ToRegionEnum(),
		}, nil
	}
}

func createAWSConfigParams(ctx context.Context, gql *graphql.Client, cloudAccountID uuid.UUID, params AWSConfigParams) (exocompute.CreateAWSConfigurationParams, error) {
	// BYOK: only Region is set, no VPC/subnet validation needed.
	if params.VPCID == "" {
		return exocompute.CreateAWSConfigurationParams{
			CloudAccountID: cloudAccountID,
			Region:         params.Region.ToRegionEnum(),
		}, nil
	}

	// Validate VPC.
	vpcs, err := exocompute.AWSVPCsByRegion(ctx, gql, cloudAccountID, params.Region)
	if err != nil {
		return exocompute.CreateAWSConfigurationParams{}, fmt.Errorf("failed to get vpcs: %s", err)
	}
	vpc, err := awsFindVPC(vpcs, params.VPCID)
	if err != nil {
		return exocompute.CreateAWSConfigurationParams{}, fmt.Errorf("failed to find vpc: %s", err)
	}

	// Validate subnets.
	if len(params.Subnets) != 2 {
		return exocompute.CreateAWSConfigurationParams{}, errors.New("there should be exactly 2 subnets")
	}
	subnet1, err := awsAvailabilityZone(vpc, params.Subnets[0])
	if err != nil {
		return exocompute.CreateAWSConfigurationParams{}, fmt.Errorf("failed to find subnet: %s", err)
	}
	subnet2, err := awsAvailabilityZone(vpc, params.Subnets[1])
	if err != nil {
		return exocompute.CreateAWSConfigurationParams{}, fmt.Errorf("failed to find subnet: %s", err)
	}

	// Validate security groups when customer-managed.
	isManagedByRubrik := params.ClusterSecurityGroupID == "" && params.NodeSecurityGroupID == ""
	if !isManagedByRubrik {
		if !awsHasSecurityGroup(vpc, params.ClusterSecurityGroupID) {
			return exocompute.CreateAWSConfigurationParams{},
				fmt.Errorf("invalid cluster security group id: %s", params.ClusterSecurityGroupID)
		}
		if !awsHasSecurityGroup(vpc, params.NodeSecurityGroupID) {
			return exocompute.CreateAWSConfigurationParams{},
				fmt.Errorf("invalid node security group id: %s", params.NodeSecurityGroupID)
		}
	}

	return exocompute.CreateAWSConfigurationParams{
		CloudAccountID:         cloudAccountID,
		Region:                 params.Region.ToRegionEnum(),
		VPCID:                  params.VPCID,
		Subnets:                []exocompute.AWSSubnet{subnet1, subnet2},
		IsManagedByRubrik:      isManagedByRubrik,
		ClusterSecurityGroupId: params.ClusterSecurityGroupID,
		NodeSecurityGroupId:    params.NodeSecurityGroupID,
		OptionalConfig:         params.OptionalConfig,
		TriggerHealthCheck:     params.TriggerHealthCheck,
	}, nil
}

// AddAWSConfiguration adds the exocompute configuration to the cloud account
// with the specified ID. When using AWSConfigParams, the VPC, subnets, and
// security groups are validated against RSC before the configuration is
// created. Returns the ID of the added exocompute configuration.
func (a API) AddAWSConfiguration(ctx context.Context, cloudAccountID uuid.UUID, config AWSConfig) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	var exoConfig exocompute.CreateAWSConfigurationParams
	var err error
	switch cfg := config.(type) {
	case AWSConfigurationFunc:
		exoConfig, err = cfg(ctx, a.client, cloudAccountID)
	case AWSConfigParams:
		exoConfig, err = createAWSConfigParams(ctx, a.client, cloudAccountID, cfg)
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse exocompute configuration: %s", err)
	}

	configID, err := exocompute.CreateConfiguration(ctx, a.client, exoConfig)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create exocompute configuration for cloud account %s: %s", cloudAccountID, err)
	}

	return configID, nil
}

// RemoveAWSConfiguration removes the AWS exocompute configuration with the
// specified ID.
func (a API) RemoveAWSConfiguration(ctx context.Context, configID uuid.UUID) error {
	a.log.Print(log.Trace)

	params := exocompute.DeleteAWSConfigurationParams{ConfigID: configID}
	if err := exocompute.DeleteConfiguration(ctx, a.client, params); err != nil {
		return fmt.Errorf("failed to remove exocompute configuration %s: %s", configID, err)
	}

	return nil
}

// AWSAppCloudAccounts returns the AWS exocompute application cloud account IDs
// for the specified host cloud account.
func (a API) AWSAppCloudAccounts(ctx context.Context, hostCloudAccountID uuid.UUID) ([]uuid.UUID, error) {
	a.log.Print(log.Trace)

	configsForAccounts, err := exocompute.ListConfigurations(ctx, a.client, exocompute.AWSConfigurationsFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to get exocompute configurations: %s", err)
	}

	var appCloudAccounts []uuid.UUID
	for _, configsForAccount := range configsForAccounts {
		if configsForAccount.CloudAccount.ID != hostCloudAccountID {
			continue
		}
		for _, cloudAccount := range configsForAccount.MappedCloudAccounts {
			appCloudAccounts = append(appCloudAccounts, cloudAccount.ID)
		}
		break
	}

	return appCloudAccounts, nil
}

// AWSHostCloudAccount returns the AWS exocompute host cloud account ID for the
// specified application cloud account. If no host cloud account is found,
// graphql.ErrNotFound is returned.
func (a API) AWSHostCloudAccount(ctx context.Context, appCloudAccountID uuid.UUID) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	configsForAccounts, err := exocompute.ListConfigurations(ctx, a.client, exocompute.AWSConfigurationsFilter{})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get exocompute configurations: %s", err)
	}

	for _, configsForAccount := range configsForAccounts {
		for _, mappedAccount := range configsForAccount.MappedCloudAccounts {
			if mappedAccount.ID == appCloudAccountID {
				return configsForAccount.CloudAccount.ID, nil
			}
		}
	}

	return uuid.Nil, fmt.Errorf("host cloud account for application cloud account %s %w", appCloudAccountID, graphql.ErrNotFound)
}

// MapAWSCloudAccount maps the AWS exocompute application cloud account to the
// specified AWS host cloud account.
func (a API) MapAWSCloudAccount(ctx context.Context, appCloudAccountID, hostCloudAccountID uuid.UUID) error {
	a.log.Print(log.Trace)

	params := exocompute.MapAWSCloudAccountsParams{AppCloudAccountIDs: []uuid.UUID{appCloudAccountID}, HostCloudAccountID: hostCloudAccountID}
	if err := exocompute.MapCloudAccounts(ctx, a.client, params); err != nil {
		return fmt.Errorf("failed to map application cloud account %s to host cloud account %s: %s", appCloudAccountID, hostCloudAccountID, err)
	}

	return nil
}

// UnmapAWSCloudAccount unmaps the AWS exocompute application cloud account with
// the specified cloud account ID.
func (a API) UnmapAWSCloudAccount(ctx context.Context, appCloudAccountID uuid.UUID) error {
	a.log.Print(log.Trace)

	params := exocompute.UnmapAWSCloudAccountsParams{AppCloudAccountIDs: []uuid.UUID{appCloudAccountID}}
	if err := exocompute.UnmapCloudAccounts(ctx, a.client, params); err != nil {
		return fmt.Errorf("failed to unmap application cloud account %s: %s", appCloudAccountID, err)
	}

	return nil
}

// AWSClusterConnection returns the connection command and cluster setup YAML
// manifest for an AWS exocompute cluster connection.
func (a API) AWSClusterConnection(ctx context.Context, clusterName string, configID uuid.UUID) (exocompute.AWSClusterConnectionResult, error) {
	a.log.Print(log.Trace)

	params := exocompute.AWSClusterConnectionParams{ClusterName: clusterName, ConfigID: configID}
	info, err := exocompute.ClusterConnection(ctx, a.client, params)
	if err != nil {
		return exocompute.AWSClusterConnectionResult{}, fmt.Errorf("failed to get cluster connection info for %q: %s", clusterName, err)
	}

	return info, nil
}

// ConnectAWSCluster connects the named AWS exocompute cluster to the specified
// exocompute configuration. Returns the cluster ID and information about the
// connection.
func (a API) ConnectAWSCluster(ctx context.Context, clusterName string, configID uuid.UUID) (uuid.UUID, exocompute.AWSClusterConnectionResult, error) {
	a.log.Print(log.Trace)

	params := exocompute.ConnectAWSClusterParams{ClusterName: clusterName, ConfigID: configID}
	info, err := exocompute.ConnectCluster(ctx, a.client, params)
	if err != nil {
		return uuid.Nil, exocompute.AWSClusterConnectionResult{}, fmt.Errorf("failed to connect exocompute cluster %q: %s", clusterName, err)
	}

	return info.ClusterID, info.AWSClusterConnectionResult, nil
}

// DisconnectAWSCluster disconnects the AWS exocompute cluster with the
// specified cluster ID.
func (a API) DisconnectAWSCluster(ctx context.Context, clusterID uuid.UUID) error {
	a.log.Print(log.Trace)

	params := exocompute.DisconnectAWSClusterParams{ClusterID: clusterID}
	if err := exocompute.DisconnectCluster(ctx, a.client, params); err != nil {
		return fmt.Errorf("failed to disconnect exocompute cluster %s: %s", clusterID, err)
	}

	return nil
}

// awsHasSecurityGroup returns true if an AWS security group with the specified
// ID exists in the VPC.
func awsHasSecurityGroup(vpc exocompute.AWSVPC, groupID string) bool {
	for _, group := range vpc.SecurityGroups {
		if group.ID == groupID {
			return true
		}
	}
	return false
}

// awsFindVPC returns the AWS VPC with the specified VPC ID.
func awsFindVPC(vpcs []exocompute.AWSVPC, vpcID string) (exocompute.AWSVPC, error) {
	for _, vpc := range vpcs {
		if vpc.ID == vpcID {
			return vpc, nil
		}
	}
	return exocompute.AWSVPC{}, fmt.Errorf("invalid vpc id: %s", vpcID)
}

// awsAvailabilityZone looks up the given subnet in the VPC and returns a copy
// with the AvailabilityZone populated from the VPC data. All other fields from
// the input subnet are preserved.
func awsAvailabilityZone(vpc exocompute.AWSVPC, subnet exocompute.AWSSubnet) (exocompute.AWSSubnet, error) {
	for _, s := range vpc.Subnets {
		if s.ID == subnet.ID {
			subnet.AvailabilityZone = s.AvailabilityZone
			return subnet, nil
		}
	}
	return exocompute.AWSSubnet{}, fmt.Errorf("invalid subnet ID: %s", subnet.ID)
}
