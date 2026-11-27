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
	"context"
	"fmt"

	"github.com/google/uuid"

	gqlcc "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/cloudcluster"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp"
)

// ProvisionConfig holds the configuration a cloud cluster was created with, as
// recorded by RSC when the cluster create job was started. Exactly one of AWS,
// Azure and GCP is set, depending on the cloud vendor of the cluster.
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
	GCP       *GcpProvisionConfig
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
	InstanceType        gqlcc.AwsCCInstanceType
	InstanceProfileName string
	VPC                 string
	Subnet              string
	SecurityGroups      []string
	SubnetAzConfigs     []gqlcc.SubnetAzConfig

	// VMType is the VM type the cluster was created with. VMTypeKnown is
	// false when RSC didn't record the VM type, which also happens when the
	// cluster was created with the STANDARD VM type.
	VMType      gqlcc.VmConfigType
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
	InstanceType                 gqlcc.AzureCCESSupportedInstanceType
	ResourceGroup                string
	NetworkResourceGroup         string
	VnetResourceGroup            string
	Vnet                         string
	Subnet                       string
	NetworkSecurityGroup         string
	NetworkSecurityResourceGroup string
	AvailabilityZone             string
	SubnetAzConfigs              []gqlcc.SubnetAzConfig
	StorageAccount               string
	StorageResourceGroup         string
	ContainerName                string
	EnableImmutability           bool
	ManagedIdentityName          string
	ManagedIdentityResourceGroup string

	// VMType is the VM type the cluster was created with. VMTypeKnown is
	// false when RSC didn't record the VM type, which also happens when the
	// cluster was created with the STANDARD VM type.
	VMType      gqlcc.VmConfigType
	VMTypeKnown bool
}

// GcpProvisionConfig holds the configuration a GCP cloud cluster was created
// with. NetworkConfig contains the subnet assigned to each node; it is
// separate from SubnetAzConfigs, which defines the Multi-AZ layout.
type GcpProvisionConfig struct {
	CloudAccountID       uuid.UUID
	Region               gcp.Region
	Zone                 string
	IsAzResilient        bool
	KeepClusterOnFailure bool
	ClusterConfig        ProvisionClusterConfig
	BucketName           string

	CDMVersion       string
	CDMProduct       string
	InstanceType     gqlcc.GcpCCInstanceType
	DeleteProtection bool
	NetworkConfig    []gqlcc.GcpSubnetInput
	SubnetAzConfigs  []gqlcc.SubnetAzConfig
	ServiceAccounts  []string // Service account email addresses.
}

// ProvisionConfig returns the configuration the cloud cluster with the
// specified ID was created with. If RSC has no record of the cluster create
// job, e.g. because the cluster wasn't created by RSC, graphql.ErrNotFound is
// returned. AWS, Azure and GCP cloud clusters are supported.
func (a API) ProvisionConfig(ctx context.Context, clusterID uuid.UUID) (ProvisionConfig, error) {
	raw, err := gqlcc.Wrap(a.client).ProvisionConfig(ctx, clusterID)
	if err != nil {
		return ProvisionConfig{}, err
	}

	config := ProvisionConfig{ClusterID: clusterID, Vendor: raw.Vendor, Status: raw.Status}
	switch {
	case raw.AWS != nil:
		awsConfig, err := fromAwsProvisionConfig(*raw.AWS)
		if err != nil {
			return ProvisionConfig{}, fmt.Errorf("invalid AWS provision config for cloud cluster %q: %w", clusterID, err)
		}
		config.AWS = &awsConfig
	case raw.Azure != nil:
		azureConfig, err := fromAzureProvisionConfig(*raw.Azure)
		if err != nil {
			return ProvisionConfig{}, fmt.Errorf("invalid Azure provision config for cloud cluster %q: %w", clusterID, err)
		}
		config.Azure = &azureConfig
	case raw.GCP != nil:
		gcpConfig, err := fromGcpProvisionConfig(*raw.GCP)
		if err != nil {
			return ProvisionConfig{}, fmt.Errorf("invalid GCP provision config for cloud cluster %q: %w", clusterID, err)
		}
		config.GCP = &gcpConfig
	default:
		return ProvisionConfig{}, fmt.Errorf("unsupported cloud vendor %q for cloud cluster %q", raw.Vendor, clusterID)
	}

	return config, nil
}

// provisionVMTypes maps the protobuf VmType enum to the VM type. A missing
// value is the zero value, STANDARD, but can't be told apart from a VM type
// that was never recorded.
var provisionVMTypes = map[int]gqlcc.VmConfigType{
	0: gqlcc.CCVmConfigStandard,
	1: gqlcc.CCVmConfigDense,
	2: gqlcc.CCVmConfigExtraDense,
}

// provisionAwsInstanceTypes maps the protobuf AwsInstanceType enum to the AWS
// instance type.
var provisionAwsInstanceTypes = map[int]gqlcc.AwsCCInstanceType{
	0:  gqlcc.AwsInstanceTypeUnspecified,
	1:  gqlcc.AwsInstanceTypeM5_4XLarge,
	2:  gqlcc.AwsInstanceTypeM6I_2XLarge,
	3:  gqlcc.AwsInstanceTypeM6I_4XLarge,
	4:  gqlcc.AwsInstanceTypeM6I_8XLarge,
	5:  gqlcc.AwsInstanceTypeR6I_4XLarge,
	6:  gqlcc.AwsInstanceTypeM6A_2XLarge,
	7:  gqlcc.AwsInstanceTypeM6A_4XLarge,
	8:  gqlcc.AwsInstanceTypeM6A_8XLarge,
	9:  gqlcc.AwsInstanceTypeR6A_4XLarge,
	10: gqlcc.AwsInstanceTypeM7I_2XLarge,
	11: gqlcc.AwsInstanceTypeM7I_4XLarge,
	12: gqlcc.AwsInstanceTypeM7I_8XLarge,
	13: gqlcc.AwsInstanceTypeM7A_2XLarge,
	14: gqlcc.AwsInstanceTypeM7A_4XLarge,
	15: gqlcc.AwsInstanceTypeM7A_8XLarge,
	16: gqlcc.AwsInstanceTypeR7I_4XLarge,
	17: gqlcc.AwsInstanceTypeR7A_4XLarge,
}

// provisionAzureInstanceTypes maps the protobuf AzureInstanceType enum to the
// Azure instance type.
var provisionAzureInstanceTypes = map[int]gqlcc.AzureCCESSupportedInstanceType{
	0:  gqlcc.AzureInstanceTypeUnspecified,
	1:  gqlcc.AzureInstanceTypeStandardDS5V2,
	2:  gqlcc.AzureInstanceTypeStandardD16SV5,
	3:  gqlcc.AzureInstanceTypeStandardD8SV5,
	4:  gqlcc.AzureInstanceTypeStandardD32SV5,
	5:  gqlcc.AzureInstanceTypeStandardE16SV5,
	6:  gqlcc.AzureInstanceTypeStandardD8ASV5,
	7:  gqlcc.AzureInstanceTypeStandardD16ASV5,
	8:  gqlcc.AzureInstanceTypeStandardD32ASV5,
	9:  gqlcc.AzureInstanceTypeStandardE16ASV5,
	10: gqlcc.AzureInstanceTypeStandardD8SV6,
	11: gqlcc.AzureInstanceTypeStandardD16SV6,
	12: gqlcc.AzureInstanceTypeStandardE16SV6,
	13: gqlcc.AzureInstanceTypeStandardD8ASV6,
	14: gqlcc.AzureInstanceTypeStandardD16ASV6,
	15: gqlcc.AzureInstanceTypeStandardE16ASV6,
}

// provisionGcpInstanceTypes maps the protobuf GcpInstanceType enum to the GCP
// instance type.
var provisionGcpInstanceTypes = map[int]gqlcc.GcpCCInstanceType{
	0:  gqlcc.GcpInstanceTypeUnspecified,
	1:  gqlcc.GcpInstanceTypeN2Standard8,
	2:  gqlcc.GcpInstanceTypeN2Standard16,
	3:  gqlcc.GcpInstanceTypeN2Highmem16,
	4:  gqlcc.GcpInstanceTypeN2DStandard8,
	5:  gqlcc.GcpInstanceTypeN2DStandard16,
	6:  gqlcc.GcpInstanceTypeN2DHighmem16,
	7:  gqlcc.GcpInstanceTypeN4Standard8,
	8:  gqlcc.GcpInstanceTypeN4Standard16,
	9:  gqlcc.GcpInstanceTypeN4Highmem16,
	10: gqlcc.GcpInstanceTypeN4DStandard8,
	11: gqlcc.GcpInstanceTypeN4DStandard16,
	12: gqlcc.GcpInstanceTypeN4DHighmem16,
}

// provisionVMType returns the VM type for the protobuf VmType enum value and
// whether the VM type is known. A nil value means the value was left out.
func provisionVMType(value *int) (gqlcc.VmConfigType, bool, error) {
	if value == nil {
		return gqlcc.CCVmConfigStandard, false, nil
	}
	vmType, ok := provisionVMTypes[*value]
	if !ok {
		return "", false, fmt.Errorf("unknown VM type %d", *value)
	}
	return vmType, true, nil
}

func fromClusterConfig(config gqlcc.ProvisionClusterConfig) ProvisionClusterConfig {
	return ProvisionClusterConfig(config)
}

func fromSubnetAzConfigs(configs []gqlcc.ProvisionSubnetAzConfig) []gqlcc.SubnetAzConfig {
	if len(configs) == 0 {
		return nil
	}
	subnetAzConfigs := make([]gqlcc.SubnetAzConfig, 0, len(configs))
	for _, config := range configs {
		subnetAzConfigs = append(subnetAzConfigs, gqlcc.SubnetAzConfig{Subnet: config.Subnet, AvailabilityZone: config.AvailabilityZone})
	}
	return subnetAzConfigs
}

func fromAwsProvisionConfig(config gqlcc.AwsProvisionConfig) (AwsProvisionConfig, error) {
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
		ClusterConfig:        fromClusterConfig(config.ClusterConfig.ProvisionClusterConfig),
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
		SubnetAzConfigs:      fromSubnetAzConfigs(config.VMConfig.SubnetAzConfigs),
		VMType:               vmType,
		VMTypeKnown:          vmTypeKnown,
	}, nil
}

func fromAzureProvisionConfig(config gqlcc.AzureProvisionConfig) (AzureProvisionConfig, error) {
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
		ClusterConfig:                fromClusterConfig(config.ClusterConfig.ProvisionClusterConfig),
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
		SubnetAzConfigs:              fromSubnetAzConfigs(config.VMConfig.SubnetAzConfigs),
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

func fromGcpProvisionConfig(config gqlcc.GcpProvisionConfig) (GcpProvisionConfig, error) {
	accountID, err := uuid.Parse(config.CloudAccountID)
	if err != nil {
		return GcpProvisionConfig{}, fmt.Errorf("invalid cloud account ID %q: %s", config.CloudAccountID, err)
	}
	region := gcp.RegionFromName(config.Region)
	if region == gcp.RegionUnknown {
		return GcpProvisionConfig{}, fmt.Errorf("unknown GCP region %q", config.Region)
	}
	instanceType, ok := provisionGcpInstanceTypes[config.VMConfig.InstanceType]
	if !ok {
		return GcpProvisionConfig{}, fmt.Errorf("unknown GCP instance type %d", config.VMConfig.InstanceType)
	}

	var networks []gqlcc.GcpSubnetInput
	for _, subnet := range config.VMConfig.NetworkConfig {
		networks = append(networks, gqlcc.GcpSubnetInput{
			Network: subnet.Network, HostProject: subnet.HostProject, Region: subnet.Region, Name: subnet.Name,
		})
	}
	var serviceAccounts []string
	for _, account := range config.VMConfig.ServiceAccounts {
		serviceAccounts = append(serviceAccounts, account.Email)
	}

	return GcpProvisionConfig{
		CloudAccountID:       accountID,
		Region:               region,
		Zone:                 config.Zone,
		IsAzResilient:        config.IsAzResilient,
		KeepClusterOnFailure: config.KeepClusterOnFailure,
		ClusterConfig:        fromClusterConfig(config.ClusterConfig.ProvisionClusterConfig),
		BucketName:           config.ClusterConfig.GcpEsConfig.BucketName,
		CDMVersion:           config.VMConfig.CDMVersion,
		CDMProduct:           config.VMConfig.CDMProduct,
		InstanceType:         instanceType,
		DeleteProtection:     config.VMConfig.DeleteProtection,
		NetworkConfig:        networks,
		SubnetAzConfigs:      fromSubnetAzConfigs(config.VMConfig.SubnetAzConfigs),
		ServiceAccounts:      serviceAccounts,
	}, nil
}
