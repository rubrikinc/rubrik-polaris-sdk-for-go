// Copyright 2025 Rubrik, Inc.
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

package cloudcluster

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
)

// AwsInstanceType represents the instance types for AWS Cloud Cluster.
type AwsCCInstanceType string

const (
	AwsInstanceTypeUnspecified AwsCCInstanceType = "AWS_TYPE_UNSPECIFIED"
	AwsInstanceTypeM5_4XLarge  AwsCCInstanceType = "M5_4XLARGE"
	AwsInstanceTypeM6I_2XLarge AwsCCInstanceType = "M6I_2XLARGE"
	AwsInstanceTypeM6I_4XLarge AwsCCInstanceType = "M6I_4XLARGE"
	AwsInstanceTypeM6I_8XLarge AwsCCInstanceType = "M6I_8XLARGE"
	AwsInstanceTypeR6I_4XLarge AwsCCInstanceType = "R6I_4XLARGE"
	AwsInstanceTypeM6A_2XLarge AwsCCInstanceType = "M6A_2XLARGE"
	AwsInstanceTypeM6A_4XLarge AwsCCInstanceType = "M6A_4XLARGE"
	AwsInstanceTypeM6A_8XLarge AwsCCInstanceType = "M6A_8XLARGE"
	AwsInstanceTypeR6A_4XLarge AwsCCInstanceType = "R6A_4XLARGE"
)

// AwsCdmVersion represents the CDM version for AWS Cloud Cluster.
type AwsCdmVersion struct {
	Version                string              `json:"version"`
	IsLatest               bool                `json:"isLatest"`
	ProductCodes           []string            `json:"productCodes"`
	SupportedInstanceTypes []AwsCCInstanceType `json:"supportedInstanceTypes"`
}

var uuidRegex = regexp.MustCompile(`([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)

// AllAwsCdmVersions returns all the available CDM versions for the specified
// cloud account.
func (a API) AllAwsCdmVersions(ctx context.Context, cloudAccountID uuid.UUID, region aws.Region) ([]AwsCdmVersion, error) {
	query := awsCcCdmVersionsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID string `json:"cloudAccountId"`
		Region         string `json:"region"`
	}{CloudAccountID: cloudAccountID.String(), Region: region.Name()})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []AwsCdmVersion `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// AwsCloudAccountListVpcs represents the VPCs for AWS Cloud Cluster.
type AwsCloudAccountListVpcs struct {
	VpcID string `json:"vpcId"`
	Name  string `json:"name"`
}

// AwsCloudAccountListVpcs returns all the available VPCs for the specified cloud account.
func (a API) AwsCloudAccountListVpcs(ctx context.Context, cloudAccountID uuid.UUID, region aws.Region) ([]AwsCloudAccountListVpcs, error) {
	query := awsCcVpcQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID uuid.UUID      `json:"cloudAccountId"`
		AwsRegion      aws.RegionEnum `json:"awsRegion"`
	}{CloudAccountID: cloudAccountID, AwsRegion: region.ToRegionEnum()})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Result []AwsCloudAccountListVpcs `json:"result"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.Result, nil
}

// AwsCloudAccountRegions returns all the available regions for the specified cloud account.
func (a API) AwsCloudAccountRegions(ctx context.Context, cloudAccountID uuid.UUID) ([]aws.Region, error) {
	query := awsCcRegionQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID uuid.UUID `json:"cloudAccountId"`
	}{CloudAccountID: cloudAccountID})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []string `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	regions := make([]aws.Region, 0, len(payload.Data.Result))
	for _, regionStr := range payload.Data.Result {
		region := aws.RegionFromNativeRegionEnum(regionStr)
		regions = append(regions, region)
	}

	return regions, nil
}

// AwsCloudAccountSecurityGroup represents the security groups for AWS Cloud Cluster.
type AwsCloudAccountSecurityGroup struct {
	SecurityGroupID string `json:"securityGroupId"`
	Name            string `json:"name"`
}

// AwsCloudAccountListSecurityGroups returns all the available security groups for the specified cloud account.
func (a API) AwsCloudAccountListSecurityGroups(ctx context.Context, cloudAccountID uuid.UUID, region aws.Region, vpcID string) ([]AwsCloudAccountSecurityGroup, error) {
	query := awsCcSecurityGroupsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID uuid.UUID      `json:"cloudAccountId"`
		AwsRegion      aws.RegionEnum `json:"awsRegion"`
		AwsVpc         string         `json:"awsVpc"`
	}{CloudAccountID: cloudAccountID, AwsRegion: region.ToRegionEnum(), AwsVpc: vpcID})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Result []AwsCloudAccountSecurityGroup `json:"result"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.Result, nil
}

// AwsCloudAccountSubnets represents the subnets for AWS Cloud Cluster.
type AwsCloudAccountSubnets struct {
	SubnetID string `json:"subnetId"`
	Name     string `json:"name"`
}

// AwsCloudAccountListSubnets returns all the available subnets for the specified cloud account.
func (a API) AwsCloudAccountListSubnets(ctx context.Context, cloudAccountID uuid.UUID, region aws.Region, vpcID string) ([]AwsCloudAccountSubnets, error) {
	query := awsCcSubnetQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID uuid.UUID      `json:"cloudAccountId"`
		AwsRegion      aws.RegionEnum `json:"awsRegion"`
		AwsVpc         string         `json:"awsVpc"`
	}{CloudAccountID: cloudAccountID, AwsRegion: region.ToRegionEnum(), AwsVpc: vpcID})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Result []AwsCloudAccountSubnets `json:"result"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.Result, nil
}

// AllAwsInstanceProfileNames returns all the available instance profiles for the specified cloud account.
func (a API) AllAwsInstanceProfileNames(ctx context.Context, cloudAccountID uuid.UUID, region aws.Region) ([]string, error) {
	query := awsCcInstanceProfileQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID uuid.UUID `json:"cloudAccountId"`
		AwsRegion      string    `json:"awsRegion"`
	}{CloudAccountID: cloudAccountID, AwsRegion: region.Name()})

	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []string `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

type AwsVmConfig struct {
	CDMProduct          string            `json:"cdmProduct"`
	CDMVersion          string            `json:"cdmVersion"`
	InstanceProfileName string            `json:"instanceProfileName"`
	InstanceType        AwsCCInstanceType `json:"instanceType"`
	SecurityGroups      []string          `json:"securityGroups"`
	Subnet              string            `json:"subnet"`
	VMType              VmConfigType      `json:"vmType"`
	VPC                 string            `json:"vpc"`
}

type AwsClusterConfig struct {
	ClusterName      string           `json:"clusterName"`
	UserEmail        string           `json:"userEmail"`
	AdminPassword    secret.String    `json:"adminPassword"`
	DNSNameServers   []string         `json:"dnsNameServers"`
	DNSSearchDomains []string         `json:"dnsSearchDomains"`
	NTPServers       []string         `json:"ntpServers"`
	NumNodes         int              `json:"numNodes"`
	AwsEsConfig      AwsEsConfigInput `json:"awsEsConfig"`
}

// CreateAwsClusterInput represents the input for creating an AWS Cloud Cluster.
type CreateAwsClusterInput struct {
	CloudAccountID       uuid.UUID                  `json:"cloudAccountId"`
	ClusterConfig        AwsClusterConfig           `json:"clusterConfig"`
	IsEsType             bool                       `json:"isEsType"`
	KeepClusterOnFailure bool                       `json:"keepClusterOnFailure"`
	Region               string                     `json:"region"`
	UsePlacementGroups   bool                       `json:"usePlacementGroups"`
	Validations          []ClusterCreateValidations `json:"validations"`
	VMConfig             AwsVmConfig                `json:"vmConfig"`
}

// ValidateCreateAwsClusterInput validates the aws cloud cluster create request
func (a API) ValidateCreateAwsClusterInput(ctx context.Context, input CreateAwsClusterInput) error {
	query := validateAwsClusterCreateRequestQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input CreateAwsClusterInput `json:"input"`
	}{Input: input})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				IsSuccessful bool   `json:"isSuccessful"`
				Message      string `json:"message"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	// validate if the response is success
	if !payload.Data.Result.IsSuccessful {
		return graphql.ResponseError(query, errors.New(payload.Data.Result.Message))
	}

	return nil
}

// CreateAwsCloudCluster create AWS Cloud Cluster in RSC.
// The job ID returned is the taskchain ID and not the event ID needed to check the taskchain status.
func (a API) CreateAwsCloudCluster(ctx context.Context, input CreateAwsClusterInput) (uuid.UUID, error) {
	query := createAwsCloudClusterQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input CreateAwsClusterInput `json:"input"`
	}{Input: input})
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				JobID   int    `json:"jobId"`
				Message string `json:"message"`
				Success bool   `json:"success"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}

	// validate if the response is success
	if !payload.Data.Result.Success {
		return uuid.Nil, graphql.ResponseError(query, errors.New(payload.Data.Result.Message))
	}
	// use regex to find the UUID in the message string
	match := uuidRegex.FindString(payload.Data.Result.Message)

	// return job ID
	jobID, err := uuid.Parse(match)
	if err != nil {
		return uuid.Nil, err
	}
	return jobID, nil
}

// RemoveAwsCloudCluster deletes the specified AWS Cloud Cluster.
func (a API) RemoveAwsCloudCluster(ctx context.Context, clusterID uuid.UUID, expireInDays int, isForce bool) (bool, error) {
	query := removeAwsCcClusterQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ClusterID    uuid.UUID `json:"clusterUuid"`
		ExpireInDays int       `json:"expireInDays"`
		IsForce      bool      `json:"isForce"`
	}{ClusterID: clusterID, ExpireInDays: expireInDays, IsForce: isForce})

	if err != nil {
		return false, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result bool `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return false, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}
