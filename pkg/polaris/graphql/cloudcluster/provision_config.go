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

package cloudcluster

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
)

// Values of the jobType and status fields of the ccProvisionMetadata query.
const (
	provisionJobTypeClusterCreate = "CLUSTER_CREATE"
	provisionStatusJobNotFound    = "JOB_NOTFOUND"
)

// ProvisionConfig holds the configuration a cloud cluster was created with, as
// recorded by RSC when the cluster create job was started. Exactly one of AWS
// and Azure is set, depending on the cloud vendor of the cluster.
//
// The configuration is what was requested, it's not refreshed when the cluster
// changes after it has been created. Credentials, such as the admin password,
// are not part of the configuration.
type ProvisionConfig struct {
	ClusterID uuid.UUID
	Vendor    string // The cloud vendor, e.g. AWS or AZURE.
	Status    string // The status of the cluster create job, e.g. COMPLETED.
	AWS       *AwsProvisionConfig
	Azure     *AzureProvisionConfig
}

// ProvisionClusterConfig holds the cluster configuration shared by all cloud
// vendors.
type ProvisionClusterConfig struct {
	ClusterName           string
	NumNodes              int
	DNSNameServers        []string
	DNSSearchDomains      []string
	NTPServers            []string
	DynamicScalingEnabled bool
}

// AwsProvisionConfig holds the configuration an AWS cloud cluster was created
// with.
type AwsProvisionConfig struct {
	CloudAccountID       uuid.UUID
	Region               aws.Region
	IsAzResilient        bool
	KeepClusterOnFailure bool
	UsePlacementGroups   bool
	ClusterConfig        ProvisionClusterConfig
	BucketName           string
	EnableImmutability   bool
	EnableObjectLock     bool

	// CDMVersion is the CDM version the cluster was created with, e.g.
	// 9.4.0-p2-30507.
	CDMVersion          string
	CDMProduct          string
	InstanceType        AwsCCInstanceType
	InstanceProfileName string
	VPC                 string
	Subnet              string
	SecurityGroups      []string
	SubnetAzConfigs     []SubnetAzConfig

	// VMType is the VM type the cluster was created with. VMTypeKnown is
	// false when RSC didn't record the VM type, which also happens when the
	// cluster was created with the STANDARD VM type.
	VMType      VmConfigType
	VMTypeKnown bool
}

// AzureProvisionConfig holds the configuration an Azure cloud cluster was
// created with.
type AzureProvisionConfig struct {
	CloudAccountID       uuid.UUID
	IsAzResilient        bool
	KeepClusterOnFailure bool
	ClusterConfig        ProvisionClusterConfig

	// ImageVersion is the version of the CDM marketplace image the cluster was
	// created from, e.g. 3.7.31669. It's not the CDM version, use the version
	// of the cluster for that.
	ImageVersion string
	// SKU is the CDM marketplace SKU the cluster was created from, e.g.
	// rubrik-cdm-94.
	SKU                          string
	Region                       azure.Region
	InstanceType                 AzureCCESSupportedInstanceType
	ResourceGroup                string
	NetworkResourceGroup         string
	VnetResourceGroup            string
	Vnet                         string
	Subnet                       string
	NetworkSecurityGroup         string
	NetworkSecurityResourceGroup string
	AvailabilityZone             string
	SubnetAzConfigs              []SubnetAzConfig
	StorageAccount               string
	StorageResourceGroup         string
	ContainerName                string
	EnableImmutability           bool
	ManagedIdentityName          string
	ManagedIdentityResourceGroup string

	// VMType is the VM type the cluster was created with. VMTypeKnown is
	// false when RSC didn't record the VM type, which also happens when the
	// cluster was created with the STANDARD VM type.
	VMType      VmConfigType
	VMTypeKnown bool
}

// ProvisionConfig returns the configuration the cloud cluster with the
// specified ID was created with. If RSC has no record of the cluster create
// job, e.g. because the cluster wasn't created by RSC, graphql.ErrNotFound is
// returned. Only AWS and Azure cloud clusters are supported.
func (a API) ProvisionConfig(ctx context.Context, clusterID uuid.UUID) (ProvisionConfig, error) {
	query := ccProvisionMetadataQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ClusterUUID string `json:"clusterUuid"`
		JobType     string `json:"jobType"`
	}{ClusterUUID: clusterID.String(), JobType: provisionJobTypeClusterCreate})
	if err != nil {
		return ProvisionConfig{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				ClusterUUID     string `json:"clusterUuid"`
				JobType         string `json:"jobType"`
				Status          string `json:"status"`
				Vendor          string `json:"vendor"`
				MarshaledConfig string `json:"marshaledConfig"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return ProvisionConfig{}, graphql.UnmarshalError(query, err)
	}
	result := payload.Data.Result

	// Without a cluster create job, RSC returns the JOB_NOTFOUND status
	// together with the configuration of the cluster's latest job of any kind,
	// or an empty JSON array when the cluster has no jobs at all.
	if result.Status == provisionStatusJobNotFound || result.JobType != provisionJobTypeClusterCreate {
		return ProvisionConfig{}, fmt.Errorf("create job for cloud cluster %q %w", clusterID, graphql.ErrNotFound)
	}
	if id, err := uuid.Parse(result.ClusterUUID); err != nil || id != clusterID {
		return ProvisionConfig{}, graphql.ResponseError(query, fmt.Errorf("cloud cluster ID mismatch: expected %q, got %q", clusterID, result.ClusterUUID))
	}

	config := ProvisionConfig{ClusterID: clusterID, Vendor: result.Vendor, Status: result.Status}
	switch result.Vendor {
	case "AWS":
		awsConfig, err := parseAwsProvisionConfig([]byte(result.MarshaledConfig))
		if err != nil {
			return ProvisionConfig{}, graphql.UnmarshalError(query, err)
		}
		config.AWS = &awsConfig
	case "AZURE":
		azureConfig, err := parseAzureProvisionConfig([]byte(result.MarshaledConfig))
		if err != nil {
			return ProvisionConfig{}, graphql.UnmarshalError(query, err)
		}
		config.Azure = &azureConfig
	default:
		return ProvisionConfig{}, graphql.ResponseError(query, fmt.Errorf("unsupported cloud vendor %q for cloud cluster %q", result.Vendor, clusterID))
	}

	return config, nil
}

// The marshaled configuration is the cluster create request, AwsClusterCreateReq
// or AzureClusterCreateReq, encoded with encoding/json from its Go protobuf
// type. Field names are the protobuf names, enums are encoded as numbers and
// fields holding the zero value of their type are left out.

// provisionClusterConfigJSON has the same fields as ProvisionClusterConfig, so
// it converts to it directly.
type provisionClusterConfigJSON struct {
	ClusterName           string   `json:"cluster_name"`
	NumNodes              int      `json:"num_nodes"`
	DNSNameServers        []string `json:"dns_name_servers"`
	DNSSearchDomains      []string `json:"dns_search_domains"`
	NTPServers            []string `json:"ntp_servers"`
	DynamicScalingEnabled bool     `json:"dynamic_scaling_enabled"`
}

type provisionSubnetAzConfigJSON struct {
	Subnet           string `json:"subnet"`
	AvailabilityZone string `json:"availability_zone"`
}

func toSubnetAzConfigs(configs []provisionSubnetAzConfigJSON) []SubnetAzConfig {
	if len(configs) == 0 {
		return nil
	}
	subnetAzConfigs := make([]SubnetAzConfig, 0, len(configs))
	for _, config := range configs {
		subnetAzConfigs = append(subnetAzConfigs, SubnetAzConfig{Subnet: config.Subnet, AvailabilityZone: config.AvailabilityZone})
	}
	return subnetAzConfigs
}

// provisionVMTypes maps the protobuf VmType enum to the VM type. A missing
// value is the zero value, STANDARD, but can't be told apart from a VM type
// that was never recorded.
var provisionVMTypes = map[int]VmConfigType{
	0: CCVmConfigStandard,
	1: CCVmConfigDense,
	2: CCVmConfigExtraDense,
}

// provisionAwsInstanceTypes maps the protobuf AwsInstanceType enum to the AWS
// instance type.
var provisionAwsInstanceTypes = map[int]AwsCCInstanceType{
	0:  AwsInstanceTypeUnspecified,
	1:  AwsInstanceTypeM5_4XLarge,
	2:  AwsInstanceTypeM6I_2XLarge,
	3:  AwsInstanceTypeM6I_4XLarge,
	4:  AwsInstanceTypeM6I_8XLarge,
	5:  AwsInstanceTypeR6I_4XLarge,
	6:  AwsInstanceTypeM6A_2XLarge,
	7:  AwsInstanceTypeM6A_4XLarge,
	8:  AwsInstanceTypeM6A_8XLarge,
	9:  AwsInstanceTypeR6A_4XLarge,
	10: "M7I_2XLARGE",
	11: "M7I_4XLARGE",
	12: "M7I_8XLARGE",
	13: "M7A_2XLARGE",
	14: "M7A_4XLARGE",
	15: "M7A_8XLARGE",
	16: "R7I_4XLARGE",
	17: "R7A_4XLARGE",
}

// provisionAzureInstanceTypes maps the protobuf AzureInstanceType enum to the
// Azure instance type.
var provisionAzureInstanceTypes = map[int]AzureCCESSupportedInstanceType{
	0:  AzureInstanceTypeUnspecified,
	1:  AzureInstanceTypeStandardDS5V2,
	2:  AzureInstanceTypeStandardD16SV5,
	3:  AzureInstanceTypeStandardD8SV5,
	4:  AzureInstanceTypeStandardD32SV5,
	5:  AzureInstanceTypeStandardE16SV5,
	6:  AzureInstanceTypeStandardD8ASV5,
	7:  AzureInstanceTypeStandardD16ASV5,
	8:  AzureInstanceTypeStandardD32ASV5,
	9:  AzureInstanceTypeStandardE16ASV5,
	10: "STANDARD_D8S_V6",
	11: "STANDARD_D16S_V6",
	12: "STANDARD_E16S_V6",
	13: "STANDARD_D8AS_V6",
	14: "STANDARD_D16AS_V6",
	15: "STANDARD_E16AS_V6",
}

// decodeProvisionConfig decodes the marshaled configuration into config. RSC
// encodes a missing configuration as an empty JSON array.
func decodeProvisionConfig(data []byte, config any) error {
	if trimmed := bytes.TrimSpace(data); len(trimmed) == 0 || bytes.Equal(trimmed, []byte("[]")) {
		return errors.New("cluster create job has no configuration")
	}
	return json.Unmarshal(data, config)
}

// provisionVMType returns the VM type for the protobuf VmType enum value and
// whether the VM type is known. A nil value means the value was left out.
func provisionVMType(value *int) (VmConfigType, bool, error) {
	if value == nil {
		return CCVmConfigStandard, false, nil
	}
	vmType, ok := provisionVMTypes[*value]
	if !ok {
		return "", false, fmt.Errorf("unknown VM type %d", *value)
	}
	return vmType, true, nil
}

func parseAwsProvisionConfig(data []byte) (AwsProvisionConfig, error) {
	var config struct {
		CloudAccountID       string `json:"cloud_account_id"`
		Region               string `json:"region"`
		IsAzResilient        bool   `json:"is_az_resilient"`
		KeepClusterOnFailure bool   `json:"keep_cluster_on_failure"`
		UsePlacementGroups   bool   `json:"use_placement_groups"`
		ClusterConfig        struct {
			provisionClusterConfigJSON
			AwsEsConfig struct {
				BucketName         string `json:"bucket_name"`
				EnableImmutability bool   `json:"enable_immutability"`
				EnableObjectLock   bool   `json:"enable_object_lock"`
			} `json:"aws_es_config"`
		} `json:"cluster_config"`
		VMConfig struct {
			CDMVersion          string                        `json:"cdm_version"`
			CDMProduct          string                        `json:"cdm_product"`
			VMType              *int                          `json:"vm_type"`
			InstanceType        int                           `json:"instance_type"`
			InstanceProfileName string                        `json:"instance_profile_name"`
			VPC                 string                        `json:"vpc"`
			Subnet              string                        `json:"subnet"`
			SecurityGroups      []string                      `json:"security_groups"`
			SubnetAzConfigs     []provisionSubnetAzConfigJSON `json:"subnet_az_configs"`
		} `json:"vm_config"`
	}
	if err := decodeProvisionConfig(data, &config); err != nil {
		return AwsProvisionConfig{}, err
	}

	cloudAccountID, err := uuid.Parse(config.CloudAccountID)
	if err != nil {
		return AwsProvisionConfig{}, fmt.Errorf("invalid cloud account ID %q: %s", config.CloudAccountID, err)
	}
	region := aws.RegionFromAny(config.Region)
	if region == aws.RegionUnknown {
		return AwsProvisionConfig{}, fmt.Errorf("unknown AWS region %q", config.Region)
	}
	instanceType, ok := provisionAwsInstanceTypes[config.VMConfig.InstanceType]
	if !ok {
		return AwsProvisionConfig{}, fmt.Errorf("unknown AWS instance type %d", config.VMConfig.InstanceType)
	}
	vmType, vmTypeKnown, err := provisionVMType(config.VMConfig.VMType)
	if err != nil {
		return AwsProvisionConfig{}, err
	}

	es := config.ClusterConfig.AwsEsConfig
	return AwsProvisionConfig{
		CloudAccountID:       cloudAccountID,
		Region:               region,
		IsAzResilient:        config.IsAzResilient,
		KeepClusterOnFailure: config.KeepClusterOnFailure,
		UsePlacementGroups:   config.UsePlacementGroups,
		ClusterConfig:        ProvisionClusterConfig(config.ClusterConfig.provisionClusterConfigJSON),
		BucketName:           es.BucketName,
		EnableImmutability:   es.EnableImmutability,
		EnableObjectLock:     es.EnableObjectLock,
		CDMVersion:           config.VMConfig.CDMVersion,
		CDMProduct:           config.VMConfig.CDMProduct,
		InstanceType:         instanceType,
		InstanceProfileName:  config.VMConfig.InstanceProfileName,
		VPC:                  config.VMConfig.VPC,
		Subnet:               config.VMConfig.Subnet,
		SecurityGroups:       config.VMConfig.SecurityGroups,
		SubnetAzConfigs:      toSubnetAzConfigs(config.VMConfig.SubnetAzConfigs),
		VMType:               vmType,
		VMTypeKnown:          vmTypeKnown,
	}, nil
}

func parseAzureProvisionConfig(data []byte) (AzureProvisionConfig, error) {
	var config struct {
		CloudAccountID       string `json:"cloud_account_id"`
		IsAzResilient        bool   `json:"is_az_resilient"`
		KeepClusterOnFailure bool   `json:"keep_cluster_on_failure"`
		ClusterConfig        struct {
			provisionClusterConfigJSON
			AzureEsConfig struct {
				StorageAccount     string `json:"storage_account"`
				ResourceGroup      string `json:"resource_group"`
				ContainerName      string `json:"container_name"`
				EnableImmutability bool   `json:"enable_immutability"`
				ManagedIdentity    struct {
					Name          string `json:"name"`
					ResourceGroup string `json:"resource_group"`
				} `json:"managed_identity"`
			} `json:"azure_es_config"`
		} `json:"cluster_config"`
		VMConfig struct {
			CDMVersion                   string                        `json:"cdm_version"`
			CDMProduct                   string                        `json:"cdm_product"`
			VMType                       *int                          `json:"vm_type"`
			Location                     string                        `json:"location"`
			InstanceType                 int                           `json:"instance_type"`
			ResourceGroup                string                        `json:"resource_group"`
			NetworkResourceGroup         string                        `json:"network_resource_group"`
			VnetResourceGroup            string                        `json:"vnet_resource_group"`
			Vnet                         string                        `json:"vnet"`
			Subnet                       string                        `json:"subnet"`
			NetworkSecurityGroup         string                        `json:"network_security_group"`
			NetworkSecurityResourceGroup string                        `json:"network_security_resource_group"`
			AvailabilityZone             string                        `json:"availability_zone"`
			SubnetAzConfigs              []provisionSubnetAzConfigJSON `json:"subnet_az_configs"`
		} `json:"vm_config"`
	}
	if err := decodeProvisionConfig(data, &config); err != nil {
		return AzureProvisionConfig{}, err
	}

	cloudAccountID, err := uuid.Parse(config.CloudAccountID)
	if err != nil {
		return AzureProvisionConfig{}, fmt.Errorf("invalid cloud account ID %q: %s", config.CloudAccountID, err)
	}
	region := azure.RegionFromAny(config.VMConfig.Location)
	if region == azure.RegionUnknown {
		return AzureProvisionConfig{}, fmt.Errorf("unknown Azure region %q", config.VMConfig.Location)
	}
	instanceType, ok := provisionAzureInstanceTypes[config.VMConfig.InstanceType]
	if !ok {
		return AzureProvisionConfig{}, fmt.Errorf("unknown Azure instance type %d", config.VMConfig.InstanceType)
	}
	vmType, vmTypeKnown, err := provisionVMType(config.VMConfig.VMType)
	if err != nil {
		return AzureProvisionConfig{}, err
	}

	es := config.ClusterConfig.AzureEsConfig
	return AzureProvisionConfig{
		CloudAccountID:               cloudAccountID,
		IsAzResilient:                config.IsAzResilient,
		KeepClusterOnFailure:         config.KeepClusterOnFailure,
		ClusterConfig:                ProvisionClusterConfig(config.ClusterConfig.provisionClusterConfigJSON),
		ImageVersion:                 config.VMConfig.CDMVersion,
		SKU:                          config.VMConfig.CDMProduct,
		Region:                       region,
		InstanceType:                 instanceType,
		ResourceGroup:                config.VMConfig.ResourceGroup,
		NetworkResourceGroup:         config.VMConfig.NetworkResourceGroup,
		VnetResourceGroup:            config.VMConfig.VnetResourceGroup,
		Vnet:                         config.VMConfig.Vnet,
		Subnet:                       config.VMConfig.Subnet,
		NetworkSecurityGroup:         config.VMConfig.NetworkSecurityGroup,
		NetworkSecurityResourceGroup: config.VMConfig.NetworkSecurityResourceGroup,
		AvailabilityZone:             config.VMConfig.AvailabilityZone,
		SubnetAzConfigs:              toSubnetAzConfigs(config.VMConfig.SubnetAzConfigs),
		StorageAccount:               es.StorageAccount,
		StorageResourceGroup:         es.ResourceGroup,
		ContainerName:                es.ContainerName,
		EnableImmutability:           es.EnableImmutability,
		ManagedIdentityName:          es.ManagedIdentity.Name,
		ManagedIdentityResourceGroup: es.ManagedIdentity.ResourceGroup,
		VMType:                       vmType,
		VMTypeKnown:                  vmTypeKnown,
	}, nil
}
