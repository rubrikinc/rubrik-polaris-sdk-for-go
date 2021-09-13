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
	"net/url"
	"strings"
	"time"

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

	return aws.VPC{}, fmt.Errorf("polaris: invalid vpc id: %s", vpcID)
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

	return aws.Subnet{}, fmt.Errorf("polaris: invalid subnet id: %s", subnetID)
}

// Managed returns an ExoConfigFunc that initializes an exocompute config with
// security groups managed by Polaris using the specified values.
func Managed(region, vpcID string, subnetIDs []string) ExoConfigFunc {
	return func(ctx context.Context, gql *graphql.Client, id uuid.UUID) (aws.ExocomputeConfigCreate, error) {
		reg, err := aws.ParseRegion(region)
		if err != nil {
			return aws.ExocomputeConfigCreate{}, err
		}

		// Validate VPC.
		vpcs, err := aws.Wrap(gql).AllVpcsByRegion(ctx, id, reg)
		if err != nil {
			return aws.ExocomputeConfigCreate{}, err
		}
		vpc, err := findVPC(vpcs, vpcID)
		if err != nil {
			return aws.ExocomputeConfigCreate{}, err
		}

		// Validate subnets.
		if len(subnetIDs) != 2 {
			return aws.ExocomputeConfigCreate{}, errors.New("polaris: there should be exactly 2 subnet ids")
		}
		subnet1, err := findSubnet(vpc, subnetIDs[0])
		if err != nil {
			return aws.ExocomputeConfigCreate{}, err
		}
		subnet2, err := findSubnet(vpc, subnetIDs[1])
		if err != nil {
			return aws.ExocomputeConfigCreate{}, err
		}

		return aws.ExocomputeConfigCreate{
			Region:           reg,
			VPCID:            vpcID,
			Subnets:          []aws.Subnet{subnet1, subnet2},
			IsPolarisManaged: true,
		}, nil
	}
}

// Unmanaged returns an ExoConfigFunc that initializes an exocompute config
// with security groups managed by the user using the specified values.
func Unmanaged(region, vpcID string, subnetIDs []string, clusterSecurityGroupID, nodeSecurityGroupID string) ExoConfigFunc {
	return func(ctx context.Context, gql *graphql.Client, id uuid.UUID) (aws.ExocomputeConfigCreate, error) {
		reg, err := aws.ParseRegion(region)
		if err != nil {
			return aws.ExocomputeConfigCreate{}, err
		}

		// Validate VPC.
		vpcs, err := aws.Wrap(gql).AllVpcsByRegion(ctx, id, reg)
		if err != nil {
			return aws.ExocomputeConfigCreate{}, err
		}
		vpc, err := findVPC(vpcs, vpcID)
		if err != nil {
			return aws.ExocomputeConfigCreate{}, err
		}

		// Validate subnets.
		if len(subnetIDs) != 2 {
			return aws.ExocomputeConfigCreate{}, errors.New("polaris: there should be exactly 2 subnet ids")
		}
		subnet1, err := findSubnet(vpc, subnetIDs[0])
		if err != nil {
			return aws.ExocomputeConfigCreate{}, err
		}
		subnet2, err := findSubnet(vpc, subnetIDs[1])
		if err != nil {
			return aws.ExocomputeConfigCreate{}, err
		}

		// Validate security groups.
		if hasSecurityGroup(vpc, clusterSecurityGroupID) {
			return aws.ExocomputeConfigCreate{},
				fmt.Errorf("polaris: invalid cluster security group id: %s", clusterSecurityGroupID)
		}
		if hasSecurityGroup(vpc, nodeSecurityGroupID) {
			return aws.ExocomputeConfigCreate{},
				fmt.Errorf("polaris: invalid cluster security group id: %s", clusterSecurityGroupID)
		}

		return aws.ExocomputeConfigCreate{
			Region:                 reg,
			VPCID:                  vpcID,
			Subnets:                []aws.Subnet{subnet1, subnet2},
			IsPolarisManaged:       false,
			ClusterSecurityGroupId: clusterSecurityGroupID,
			NodeSecurityGroupId:    nodeSecurityGroupID,
		}, nil
	}
}

// EnableExocompute enables the exocompute feature for the account with the
// specified id for the given regions. The account must already be added to
// Polaris. Note that to disable the feature the account must be removed.
// The returned error will be graphql.ErrAlreadyEnabled if the feature has
// already been enabled for the specified account.
func (a API) EnableExocompute(ctx context.Context, account AccountFunc, regions ...string) error {
	a.gql.Log().Print(log.Trace, "polaris/aws.EnableExocompute")

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

	akkount, err := a.Account(ctx, AccountID(config.id), core.AllFeatures)
	if err != nil {
		return err
	}

	if _, ok := akkount.Feature(core.Exocompute); ok {
		return fmt.Errorf("polaris: feature %w", graphql.ErrAlreadyEnabled)
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

// DisableExocompute disables the exocompute feature for the account with the
// specified polaris cloud account id.
func (a API) DisableExocompute(ctx context.Context, account AccountFunc) error {
	a.gql.Log().Print(log.Trace, "polaris/aws.DisableExocompute")

	if account == nil {
		return errors.New("polaris: account is not allowed to be nil")
	}
	config, err := account(ctx)
	if err != nil {
		return err
	}

	accountID, err := a.toCloudAccountID(ctx, AccountID(config.id))
	if err != nil {
		return err
	}

	jobID, err := aws.Wrap(a.gql).StartExocomputeDisableJob(ctx, accountID)
	if err != nil {
		return err
	}

	state, err := core.Wrap(a.gql).WaitForTaskChain(ctx, jobID, 10*time.Second)
	if err != nil {
		return err
	}
	if state != core.TaskChainSucceeded {
		return fmt.Errorf("polaris: taskchain failed: jobID=%v, state=%v", jobID, state)
	}

	cfmURL, err := aws.Wrap(a.gql).PrepareCloudAccountDeletion(ctx, accountID, core.Exocompute)
	if err != nil {
		return err
	}

	i := strings.LastIndex(cfmURL, "#/stack/update") + 1
	if i == 0 {
		return errors.New("polaris: CloudFormation url does not contain #/stack/update")
	}

	u, err := url.Parse(cfmURL[i:])
	if err != nil {
		return err
	}
	stackID := u.Query().Get("stackId")
	tmplURL := u.Query().Get("templateURL")

	a.gql.Log().Printf(log.Debug, "updating CloudFormation stack: %s", stackID)
	err = awsUpdateStack(ctx, config.config, stackID, tmplURL)
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

	configsForAccounts, err := aws.Wrap(a.gql).ExocomputeConfigs(ctx, "")
	if err != nil {
		return ExocomputeConfig{}, err
	}

	for _, configsForAccount := range configsForAccounts {
		for _, config := range configsForAccount.Configs {
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

	nativeID, err := a.toNativeID(ctx, id)
	if err != nil {
		return nil, err
	}

	configsForAccounts, err := aws.Wrap(a.gql).ExocomputeConfigs(ctx, nativeID)
	if err != nil {
		return nil, err
	}

	var exoConfigs []ExocomputeConfig
	for _, configsForAccount := range configsForAccounts {
		for _, config := range configsForAccount.Configs {
			exoConfigs = append(exoConfigs, toExocomputeConfig(config))
		}
	}

	return exoConfigs, nil
}

// AddExocomputeConfig adds the exocompute config to the account with the
// specified id. Returns the id of the added exocompute config.
func (a API) AddExocomputeConfig(ctx context.Context, id IdentityFunc, config ExoConfigFunc) (uuid.UUID, error) {
	a.gql.Log().Print(log.Trace, "polaris/aws.AddExocomputeConfig")

	accountID, err := a.toCloudAccountID(ctx, id)
	if err != nil {
		return uuid.Nil, err
	}

	exoConfig, err := config(ctx, a.gql, accountID)
	if err != nil {
		return uuid.Nil, err
	}

	exo, err := aws.Wrap(a.gql).CreateExocomputeConfig(ctx, accountID, exoConfig)
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
