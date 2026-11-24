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
)

// Values of the jobType and status fields of the ccProvisionMetadata query.
const (
	provisionJobTypeClusterCreate = "CLUSTER_CREATE"
	provisionStatusJobNotFound    = "JOB_NOTFOUND"
)

// ProvisionConfig is the result of the ccProvisionMetadata query, which holds
// the configuration a cloud cluster was created with, as recorded by RSC when
// the cluster create job was started.
//
// MarshaledConfig is the configuration as returned by RSC. For convenience, it's
// also unmarshaled into the field matching the cloud vendor: AWS, Azure or GCP.
// The other fields are left nil, as are all of them if the vendor is unknown to
// the SDK. The fields of the vendor specific configurations match the fields of
// the marshaled configuration, enums are therefore numbers. Use the cloudcluster
// package for a configuration with the enums resolved.
//
// The configuration is what was requested, it's not refreshed when the cluster
// changes after it has been created. Credentials, such as the admin password,
// are not part of the configuration.
type ProvisionConfig struct {
	ClusterUUID     string `json:"clusterUuid"`
	JobType         string `json:"jobType"`
	Status          string `json:"status"`
	Vendor          string `json:"vendor"` // The cloud vendor, e.g. AWS or AZURE.
	MarshaledConfig string `json:"marshaledConfig"`

	AWS   *AwsProvisionConfig   `json:"-"`
	Azure *AzureProvisionConfig `json:"-"`
	GCP   *GcpProvisionConfig   `json:"-"`
}

// ProvisionClusterConfig holds the cluster configuration shared by all cloud
// vendors.
type ProvisionClusterConfig struct {
	ClusterName           string   `json:"cluster_name"`
	NumNodes              int      `json:"num_nodes"`
	DNSNameServers        []string `json:"dns_name_servers"`
	DNSSearchDomains      []string `json:"dns_search_domains"`
	NTPServers            []string `json:"ntp_servers"`
	DynamicScalingEnabled bool     `json:"dynamic_scaling_enabled"`
}

// ProvisionSubnetAzConfig is a subnet and availability zone pair.
type ProvisionSubnetAzConfig struct {
	Subnet           string `json:"subnet"`
	AvailabilityZone string `json:"availability_zone"`
}

// AwsProvisionConfig is the AwsClusterCreateReq recorded for an AWS cloud
// cluster.
type AwsProvisionConfig struct {
	CloudAccountID       string                    `json:"cloud_account_id"`
	Region               string                    `json:"region"`
	IsAzResilient        bool                      `json:"is_az_resilient"`
	KeepClusterOnFailure bool                      `json:"keep_cluster_on_failure"`
	UsePlacementGroups   bool                      `json:"use_placement_groups"`
	ClusterConfig        AwsProvisionClusterConfig `json:"cluster_config"`
	VMConfig             AwsProvisionVMConfig      `json:"vm_config"`
}

type AwsProvisionClusterConfig struct {
	ProvisionClusterConfig
	AwsEsConfig AwsProvisionEsConfig `json:"aws_es_config"`
}

type AwsProvisionEsConfig struct {
	BucketName         string `json:"bucket_name"`
	EnableImmutability bool   `json:"enable_immutability"`
	EnableObjectLock   bool   `json:"enable_object_lock"`
}

// AwsProvisionVMConfig holds the VM configuration of an AWS cloud cluster.
// InstanceType and VMType are the numbers of the AwsInstanceType and VmType
// enums of the cluster-mgmt protobuf definition. VMType is nil when RSC left it
// out, which happens for the STANDARD VM type.
type AwsProvisionVMConfig struct {
	CDMVersion          string                    `json:"cdm_version"`
	CDMProduct          string                    `json:"cdm_product"`
	VMType              *int                      `json:"vm_type"`
	InstanceType        int                       `json:"instance_type"`
	InstanceProfileName string                    `json:"instance_profile_name"`
	VPC                 string                    `json:"vpc"`
	Subnet              string                    `json:"subnet"`
	SecurityGroups      []string                  `json:"security_groups"`
	SubnetAzConfigs     []ProvisionSubnetAzConfig `json:"subnet_az_configs"`
}

// AzureProvisionConfig is the AzureClusterCreateReq recorded for an Azure cloud
// cluster.
type AzureProvisionConfig struct {
	CloudAccountID       string                      `json:"cloud_account_id"`
	IsAzResilient        bool                        `json:"is_az_resilient"`
	KeepClusterOnFailure bool                        `json:"keep_cluster_on_failure"`
	ClusterConfig        AzureProvisionClusterConfig `json:"cluster_config"`
	VMConfig             AzureProvisionVMConfig      `json:"vm_config"`
}

type AzureProvisionClusterConfig struct {
	ProvisionClusterConfig
	AzureEsConfig AzureProvisionEsConfig `json:"azure_es_config"`
}

type AzureProvisionEsConfig struct {
	StorageAccount     string                        `json:"storage_account"`
	ResourceGroup      string                        `json:"resource_group"`
	ContainerName      string                        `json:"container_name"`
	EnableImmutability bool                          `json:"enable_immutability"`
	ManagedIdentity    AzureProvisionManagedIdentity `json:"managed_identity"`
}

type AzureProvisionManagedIdentity struct {
	Name          string `json:"name"`
	ResourceGroup string `json:"resource_group"`
}

// AzureProvisionVMConfig holds the VM configuration of an Azure cloud cluster.
// CDMVersion is the version of the marketplace image, not of CDM, and CDMProduct
// is the marketplace SKU. InstanceType and VMType are the numbers of the
// AzureInstanceType and VmType enums of the cluster-mgmt protobuf definition.
// VMType is nil when RSC left it out, which happens for the STANDARD VM type.
type AzureProvisionVMConfig struct {
	CDMVersion                   string                    `json:"cdm_version"`
	CDMProduct                   string                    `json:"cdm_product"`
	VMType                       *int                      `json:"vm_type"`
	Location                     string                    `json:"location"`
	InstanceType                 int                       `json:"instance_type"`
	ResourceGroup                string                    `json:"resource_group"`
	NetworkResourceGroup         string                    `json:"network_resource_group"`
	VnetResourceGroup            string                    `json:"vnet_resource_group"`
	Vnet                         string                    `json:"vnet"`
	Subnet                       string                    `json:"subnet"`
	NetworkSecurityGroup         string                    `json:"network_security_group"`
	NetworkSecurityResourceGroup string                    `json:"network_security_resource_group"`
	AvailabilityZone             string                    `json:"availability_zone"`
	SubnetAzConfigs              []ProvisionSubnetAzConfig `json:"subnet_az_configs"`
}

// GcpProvisionConfig is the GcpClusterCreateReq recorded for a GCP cloud
// cluster.
type GcpProvisionConfig struct {
	CloudAccountID       string                    `json:"cloud_account_id"`
	Region               string                    `json:"region"`
	Zone                 string                    `json:"zone"`
	IsAzResilient        bool                      `json:"is_az_resilient"`
	KeepClusterOnFailure bool                      `json:"keep_cluster_on_failure"`
	ClusterConfig        GcpProvisionClusterConfig `json:"cluster_config"`
	VMConfig             GcpProvisionVMConfig      `json:"vm_config"`
}

type GcpProvisionClusterConfig struct {
	ProvisionClusterConfig
	GcpEsConfig GcpProvisionEsConfig `json:"gcp_es_config"`
}

type GcpProvisionEsConfig struct {
	BucketName string `json:"bucket_name"`
}

// GcpProvisionVMConfig holds the VM configuration of a GCP cloud cluster.
// InstanceType is the number of the GcpInstanceType enum of the cluster-mgmt
// protobuf definition. NetworkConfig holds the subnet assigned to each node; it
// is separate from SubnetAzConfigs, which defines the Multi-AZ layout.
type GcpProvisionVMConfig struct {
	CDMVersion       string                       `json:"cdm_version"`
	CDMProduct       string                       `json:"cdm_product"`
	InstanceType     int                          `json:"instance_type"`
	DeleteProtection bool                         `json:"delete_protection"`
	NetworkConfig    []GcpProvisionNetworkConfig  `json:"network_config"`
	SubnetAzConfigs  []ProvisionSubnetAzConfig    `json:"subnet_az_configs"`
	ServiceAccounts  []GcpProvisionServiceAccount `json:"service_accounts"`
}

type GcpProvisionNetworkConfig struct {
	Network     string `json:"network"`
	HostProject string `json:"host_project"`
	Region      string `json:"region"`
	Name        string `json:"name"`
}

type GcpProvisionServiceAccount struct {
	Email string `json:"email"`
}

// ProvisionConfig returns the configuration the cloud cluster with the
// specified ID was created with. If RSC has no record of the cluster create
// job, e.g. because the cluster wasn't created by RSC, graphql.ErrNotFound is
// returned.
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
			Result ProvisionConfig `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return ProvisionConfig{}, graphql.UnmarshalError(query, err)
	}
	config := payload.Data.Result

	// Without a cluster create job, RSC returns the JOB_NOTFOUND status
	// together with the configuration of the cluster's latest job of any kind,
	// or an empty JSON array when the cluster has no jobs at all.
	if config.Status == provisionStatusJobNotFound || config.JobType != provisionJobTypeClusterCreate {
		return ProvisionConfig{}, fmt.Errorf("create job for cloud cluster %q %w", clusterID, graphql.ErrNotFound)
	}
	if id, err := uuid.Parse(config.ClusterUUID); err != nil || id != clusterID {
		return ProvisionConfig{}, graphql.ResponseError(query, fmt.Errorf("cloud cluster ID mismatch: expected %q, got %q", clusterID, config.ClusterUUID))
	}

	// The marshaled configuration is the cluster create request, encoded with
	// encoding/json from its Go protobuf type. Field names are the protobuf
	// names, enums are encoded as numbers and fields holding the zero value of
	// their type are left out.
	var target any
	switch config.Vendor {
	case "AWS":
		config.AWS = &AwsProvisionConfig{}
		target = config.AWS
	case "AZURE":
		config.Azure = &AzureProvisionConfig{}
		target = config.Azure
	case "GCP":
		config.GCP = &GcpProvisionConfig{}
		target = config.GCP
	default:
		return config, nil
	}
	data := []byte(config.MarshaledConfig)
	if trimmed := bytes.TrimSpace(data); len(trimmed) == 0 || bytes.Equal(trimmed, []byte("[]")) {
		return ProvisionConfig{}, graphql.UnmarshalError(query, errors.New("cluster create job has no configuration"))
	}
	if err := json.Unmarshal(data, target); err != nil {
		return ProvisionConfig{}, graphql.UnmarshalError(query, err)
	}

	return config, nil
}
