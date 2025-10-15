// Copyright 2024 Rubrik, Inc.
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
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	gqlaws "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// AWSConfigurationsFilter holds the filter for an AWS exocompute configuration
// list operation.
type AWSConfigurationsFilter struct {
	NativeIDOrNamePrefix string `json:"awsNativeAccountIdOrNamePrefix"`
}

func (p AWSConfigurationsFilter) ListQuery() (string, any, AWSConfigurationsForCloudAccount) {
	return allAwsExocomputeConfigsQuery, p, AWSConfigurationsForCloudAccount{}
}

// AWSConfigurationsForCloudAccount holds the result of an AWS exocompute
// configuration list operation.
type AWSConfigurationsForCloudAccount struct {
	CloudAccount        gqlaws.CloudAccount `json:"awsCloudAccount"`
	Configs             []AWSConfiguration  `json:"exocomputeConfigs"`
	MappedCloudAccounts []struct {
		ID       uuid.UUID `json:"id"`
		NativeID string    `json:"nativeId"`
		Name     string    `json:"name"`
	} `json:"mappedCloudAccounts"`
}

// AWSConfiguration holds a single AWS exocompute configuration.
type AWSConfiguration struct {
	ID      uuid.UUID      `json:"configUuid"`
	Region  aws.RegionEnum `json:"region"`
	VPCID   string         `json:"vpcId"`
	Subnet1 AWSSubnet      `json:"subnet1"`
	Subnet2 AWSSubnet      `json:"subnet2"`

	// When true, Polaris manages the security groups.
	IsManagedByRubrik bool `json:"areSecurityGroupsRscManaged"`

	// The security group IDs of the cluster control plane and worker nodes.
	// Only needs to be specified if IsPolarisManaged is false.
	ClusterSecurityGroupID string `json:"clusterSecurityGroupId"`
	NodeSecurityGroupID    string `json:"nodeSecurityGroupId"`

	HealthCheckStatus struct {
		Status        string `json:"status"`
		FailureReason string `json:"failureReason"`
		LastUpdatedAt string `json:"lastUpdatedAt"`
		TaskchainID   string `json:"taskchainId"`
	}
}

// CreateAWSConfigurationParams holds the parameters for an AWS exocompute
// configuration create operation.
type CreateAWSConfigurationParams struct {
	CloudAccountID uuid.UUID      `json:"-"`
	Region         aws.RegionEnum `json:"region"`

	// When true, Rubrik will manage the security groups.
	IsManagedByRubrik bool `json:"isRscManaged"`

	// Only required for RSC managed clusters
	VPCID   string      `json:"vpcId,omitempty"`
	Subnets []AWSSubnet `json:"subnets,omitempty"`

	// The security group IDs of the cluster control plane and worker nodes.
	// Only needs to be specified if IsPolarisManaged is false.
	ClusterSecurityGroupId string `json:"clusterSecurityGroupId,omitempty"`
	NodeSecurityGroupId    string `json:"nodeSecurityGroupId,omitempty"`
}

func (p CreateAWSConfigurationParams) CreateQuery() (string, any, CreateAWSConfigurationResult) {
	params := struct {
		CloudAccountID uuid.UUID                    `json:"cloudAccountId"`
		Config         CreateAWSConfigurationParams `json:"config"`
	}{CloudAccountID: p.CloudAccountID, Config: p}
	return createAwsExocomputeConfigsQuery, params, CreateAWSConfigurationResult{}
}

// CreateAWSConfigurationResult holds the result of an AWS exocompute
// configuration create operation.
type CreateAWSConfigurationResult struct {
	Configs []struct {
		ID      string `json:"configUuid"`
		Message string `json:"message"`
	} `json:"exocomputeConfigs"`
}

func (r CreateAWSConfigurationResult) Validate() (uuid.UUID, error) {
	if len(r.Configs) != 1 {
		return uuid.Nil, errors.New("expected a single configuration to be created")
	}
	if msg := r.Configs[0].Message; msg != "" {
		return uuid.Nil, errors.New(msg)
	}
	id, err := uuid.Parse(r.Configs[0].ID)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// UpdateAWSConfigurationParams holds the parameters for an AWS exocompute
// configuration update operation.
type UpdateAWSConfigurationParams CreateAWSConfigurationParams

func (p UpdateAWSConfigurationParams) UpdateQuery() (string, any, UpdateAWSConfigurationResult) {
	params := struct {
		CloudAccountID uuid.UUID                    `json:"cloudAccountId"`
		Config         UpdateAWSConfigurationParams `json:"config"`
	}{CloudAccountID: p.CloudAccountID, Config: p}
	return updateAwsExocomputeConfigsQuery, params, UpdateAWSConfigurationResult{}
}

// UpdateAWSConfigurationResult holds the result of an AWS exocompute
// configuration update operation.
type UpdateAWSConfigurationResult struct {
	Configs []struct {
		ID      string `json:"configUuid"`
		Message string `json:"message"`
	} `json:"exocomputeConfigs"`
}

func (r UpdateAWSConfigurationResult) Validate() (uuid.UUID, error) {
	if len(r.Configs) != 1 {
		return uuid.Nil, errors.New("expected a single update result")
	}
	if msg := r.Configs[0].Message; msg != "" {
		return uuid.Nil, errors.New(msg)
	}
	configID, err := uuid.Parse(r.Configs[0].ID)
	if err != nil {
		return uuid.Nil, err
	}

	return configID, nil
}

// DeleteAWSConfigurationParams holds the parameters for an AWS exocompute
// configuration delete operation.
type DeleteAWSConfigurationParams struct {
	ConfigID uuid.UUID `json:"configId"`
}

func (p DeleteAWSConfigurationParams) DeleteQuery() (string, any, DeleteAWSConfigurationResult) {
	return deleteAwsExocomputeConfigsQuery, p, DeleteAWSConfigurationResult{}
}

// DeleteAWSConfigurationResult holds the result of an AWS exocompute
// configuration delete operation.
type DeleteAWSConfigurationResult struct {
	Status []struct {
		Success bool `json:"success"`
	} `json:"deletionStatus"`
}

func (r DeleteAWSConfigurationResult) Validate() error {
	if len(r.Status) != 1 {
		return errors.New("expected a single configuration to be deleted")
	}
	if !r.Status[0].Success {
		return errors.New("failed to delete configuration")
	}
	return nil
}

// AWSSubnet represents an AWS VPC subnet.
type AWSSubnet struct {
	ID               string `json:"subnetId"`
	AvailabilityZone string `json:"availabilityZone"`
}

// AWSVPC represents an AWS VPC together with its subnets and security groups.
type AWSVPC struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Subnets []struct {
		ID               string `json:"id"`
		Name             string `json:"name"`
		AvailabilityZone string `json:"availabilityZone"`
	} `json:"subnets"`
	SecurityGroups []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"securityGroups"`
}

// AWSVPCsByRegion returns all AWS VPCs for the specified RSC cloud account ID.
func AWSVPCsByRegion(ctx context.Context, gql *graphql.Client, cloudAccountID uuid.UUID, region aws.Region) ([]AWSVPC, error) {
	gql.Log().Print(log.Trace)

	query := allVpcsByRegionFromAwsQuery
	buf, err := gql.Request(ctx, query, struct {
		CloudAccountID uuid.UUID      `json:"awsAccountRubrikId"`
		Region         aws.RegionEnum `json:"region"`
	}{CloudAccountID: cloudAccountID, Region: region.ToRegionEnum()})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			VPCs []AWSVPC `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.VPCs, nil
}
