package aws

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/cloudcluster"
)

// AwsCdmVersion represents the CDM version for AWS Cloud Cluster.
type AwsCdmVersion struct {
	Version                string              `json:"version"`
	IsLatest               bool                `json:"isLatest"`
	ProductCodes           []string            `json:"productCodes"`
	SupportedInstanceTypes []AwsCCInstanceType `json:"supportedInstanceTypes"`
}

type AwsCdmVersionRequest struct {
	CloudAccountID uuid.UUID `json:"cloudAccountId"`
	Region         string    `json:"region"`
}

// AllAwsCdmVersions returns all the available CDM versions for the specified
// cloud account.
func (a API) AllAwsCdmVersions(ctx context.Context, cloudAccountID uuid.UUID, region Region) ([]AwsCdmVersion, error) {

	query := awsCcCdmVersionsQuery
	input := AwsCdmVersionRequest{
		CloudAccountID: cloudAccountID,
		Region:         region.Name(),
	}
	buf, err := a.GQL.Request(ctx, query, struct {
		Input AwsCdmVersionRequest `json:"input"`
	}{Input: input})
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

type AwsCloudAccountVpcResult struct {
	Result []AwsCloudAccountListVpcs `json:"result"`
}

// AwsCloudAccountListVpcs returns all the available VPCs for the specified cloud account.
func (a API) AwsCloudAccountListVpcs(ctx context.Context, cloudAccountID uuid.UUID, region Region) ([]AwsCloudAccountListVpcs, error) {

	query := awsCcVpcQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID uuid.UUID  `json:"cloudAccountId"`
		AwsRegion      RegionEnum `json:"awsRegion"`
	}{CloudAccountID: cloudAccountID, AwsRegion: region.ToRegionEnum()})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result AwsCloudAccountVpcResult `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.Result, nil
}

// AwsCloudAccountRegions returns all the available regions for the specified cloud account.
func (a API) AwsCloudAccountRegions(ctx context.Context, cloudAccountID uuid.UUID) (map[Region]struct{}, error) {
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

	regionMap := make(map[Region]struct{})
	for _, regionStr := range payload.Data.Result {
		region := RegionFromNativeRegionEnum(regionStr)
		regionMap[region] = struct{}{}
	}

	return regionMap, nil
}

// AwsCloudAccountSecurityGroups represents the security groups for AWS Cloud Cluster.
type AwsCloudAccountSecurityGroups struct {
	SecurityGroupID string `json:"securityGroupId"`
	Name            string `json:"name"`
}

type AwsCloudAccountSecurityGroupsResult struct {
	Result []AwsCloudAccountSecurityGroups `json:"result"`
}

// AwsCloudAccountListSecurityGroups returns all the available security groups for the specified cloud account.
func (a API) AwsCloudAccountListSecurityGroups(ctx context.Context, cloudAccountID uuid.UUID, region Region, vpcID string) ([]AwsCloudAccountSecurityGroups, error) {
	query := awsCcSecurityGroupsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID uuid.UUID  `json:"cloudAccountId"`
		AwsRegion      RegionEnum `json:"awsRegion"`
		AwsVpc         string     `json:"awsVpc"`
	}{CloudAccountID: cloudAccountID, AwsRegion: region.ToRegionEnum(), AwsVpc: vpcID})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result AwsCloudAccountSecurityGroupsResult `json:"result"`
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

type AwsCloudAccountSubnetsResult struct {
	Result []AwsCloudAccountSubnets `json:"result"`
}

// AwsCloudAccountListSubnets returns all the available subnets for the specified cloud account.
func (a API) AwsCloudAccountListSubnets(ctx context.Context, cloudAccountID uuid.UUID, region Region, vpcID string) ([]AwsCloudAccountSubnets, error) {
	query := awsCcSubnetQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID uuid.UUID  `json:"cloudAccountId"`
		AwsRegion      RegionEnum `json:"awsRegion"`
		AwsVpc         string     `json:"awsVpc"`
	}{CloudAccountID: cloudAccountID, AwsRegion: region.ToRegionEnum(), AwsVpc: vpcID})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result AwsCloudAccountSubnetsResult `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.Result, nil
}

// AllAwsInstanceProfileNames returns all the available instance profiles for the specified cloud account.
func (a API) AllAwsInstanceProfileNames(ctx context.Context, cloudAccountID uuid.UUID, region Region) ([]string, error) {
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
	CdmProduct          string                    `json:"cdmProduct"`
	CdmVersion          string                    `json:"cdmVersion"`
	InstanceProfileName string                    `json:"instanceProfileName"`
	InstanceType        AwsCCInstanceType         `json:"instanceType"`
	SecurityGroups      []string                  `json:"securityGroups"`
	Subnet              string                    `json:"subnet"`
	VmType              cloudcluster.VmConfigType `json:"vmType"`
	Vpc                 string                    `json:"vpc"`
}

type AwsClusterConfig struct {
	ClusterName      string                        `json:"clusterName"`
	UserEmail        string                        `json:"userEmail"`
	AdminPassword    string                        `json:"adminPassword"`
	DnsNameServers   []string                      `json:"dnsNameServers"`
	DnsSearchDomains []string                      `json:"dnsSearchDomains"`
	NtpServers       []string                      `json:"ntpServers"`
	NumNodes         int                           `json:"numNodes"`
	AwsEsConfig      cloudcluster.AwsEsConfigInput `json:"awsEsConfig"`
}

// CreateAwsClusterInput represents the input for creating an AWS Cloud Cluster.
type CreateAwsClusterInput struct {
	CloudAccountID       uuid.UUID                               `json:"cloudAccountId"`
	ClusterConfig        AwsClusterConfig                        `json:"clusterConfig"`
	IsEsType             bool                                    `json:"isEsType"`
	KeepClusterOnFailure bool                                    `json:"keepClusterOnFailure"`
	Region               string                                  `json:"region"`
	UsePlacementGroups   bool                                    `json:"usePlacementGroups"`
	Validations          []cloudcluster.ClusterCreateValidations `json:"validations"`
	VmConfig             AwsVmConfig                             `json:"vmConfig"`
}

// ValidateCreateAwsClusterInput validates the aws cloud cluster create request
func (a API) ValidateCreateAwsClusterInput(ctx context.Context, input CreateAwsClusterInput) (bool, error) {
	query := validateAwsClusterCreateRequestQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input CreateAwsClusterInput `json:"input"`
	}{Input: input})
	if err != nil {
		return false, graphql.RequestError(query, err)
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
		return false, graphql.UnmarshalError(query, err)
	}

	// validate if the response is success
	if !payload.Data.Result.IsSuccessful {
		return false, graphql.ResponseError(query, errors.New(payload.Data.Result.Message))
	}

	// return jonID and message,
	return payload.Data.Result.IsSuccessful, nil
}

// CreateAwsCloudCluster create AWS Cloud Cluster in RSC
func (a API) CreateAwsCloudCluster(ctx context.Context, input CreateAwsClusterInput) (uuid.UUID, error) {
	query := createAwsCloudClusterQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input CreateAwsClusterInput `json:"input"`
	}{Input: input})
	if err != nil {
		return uuid.UUID{}, graphql.RequestError(query, err)
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
		return uuid.UUID{}, graphql.UnmarshalError(query, err)
	}

	// validate if the response is success
	if !payload.Data.Result.Success {
		return uuid.UUID{}, graphql.ResponseError(query, errors.New(payload.Data.Result.Message))
	}
	// use regex to find the UUID in the message string
	re := regexp.MustCompile(`([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)
	match := re.FindString(payload.Data.Result.Message)
	// return jonID
	jobID, err := uuid.Parse(match)
	if err != nil {
		return uuid.UUID{}, err
	}
	return jobID, nil
}
