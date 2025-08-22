//go:generate go run ../queries_gen.go cloudcluster

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

package cloudcluster

import (
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// API wraps around GraphQL clients to give them the RSC CloudCluster API.
type API struct {
	GQL *graphql.Client
	log log.Logger
}

// Wrap the GraphQL client in the Azure API.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql, log: gql.Log()}
}

// ClusterCreateValidations represents the valid cluster create validations.
type ClusterCreateValidations string

const (
	NoChecks                   ClusterCreateValidations = "NO_CHECKS"
	AllChecks                  ClusterCreateValidations = "ALL_CHECKS"
	ClusterNameCheck           ClusterCreateValidations = "CLUSTER_NAME_CHECK"
	DNSServersCheck            ClusterCreateValidations = "DNS_SERVERS_CHECK"
	NTPServersCheck            ClusterCreateValidations = "NTP_SERVERS_CHECK"
	NodeCountCheck             ClusterCreateValidations = "NODE_COUNT_CHECK"
	ObjectStoreCheck           ClusterCreateValidations = "OBJECT_STORE_CHECK"
	ImmutabilityCheck          ClusterCreateValidations = "IMMUTABILITY_CHECK"
	AWSInstanceProfileCheck    ClusterCreateValidations = "AWS_INSTANCE_PROFILE_CHECK"
	AWSNetworkConfigCheck      ClusterCreateValidations = "AWS_NETWORK_CONFIG_CHECK"
	AzureVMImageCheck          ClusterCreateValidations = "AZURE_VM_IMAGE_CHECK"
	AzureAvailabilityZoneCheck ClusterCreateValidations = "AZURE_AVAILABILITY_ZONE_CHECK"
	AzureQuotaCheck            ClusterCreateValidations = "AZURE_QUOTA_CHECK"
	AzureMICheck               ClusterCreateValidations = "AZURE_MI_CHECK"
	CloudAccountCheck          ClusterCreateValidations = "CLOUD_ACCOUNT_CHECK"
	GCPNetworkConfigCheck      ClusterCreateValidations = "GCP_NETWORK_CONFIG_CHECK"
	GCPSerivceAccountCheck     ClusterCreateValidations = "GCP_SERVICE_ACCOUNT_CHECK"
	GCPInstanceLabelKeyCheck   ClusterCreateValidations = "GCP_INSTANCE_LABEL_KEY_CHECK"
	GCPClusterNameLengthCheck  ClusterCreateValidations = "GCP_CLUSTER_NAME_LENGTH_CHECK"
)

// VmConfigType represents the valid VM config types.
type VmConfigType string

// VmConfigType values.
const (
	CCVmConfigStandard   VmConfigType = "STANDARD"
	CCVmConfigDense      VmConfigType = "DENSE"
	CCVmConfigExtraDense VmConfigType = "EXTRA_DENSE"
)

// AwsEsConfigInput represents the input for creating an AWS ES config.
type AwsEsConfigInput struct {
	BucketName         string `json:"bucketName"`
	ShouldCreateBucket bool   `json:"shouldCreateBucket"`
	EnableImmutability bool   `json:"enableImmutability"`
	EnableObjectLock   bool   `json:"enableObjectLock"`
}

// AzureManagedIdentityName represents the input for creating an Azure managed identity.
type AzureManagedIdentityName struct {
	ClientID      string `json:"clientId"`
	Name          string `json:"name"`
	ResourceGroup string `json:"resourceGroup"`
}

// AzureEsConfigInput represents the input for creating an Azure ES config.
type AzureEsConfigInput struct {
	ContainerName         string                   `json:"containerName"`
	EnableImmutability    bool                     `json:"enableImmutability"`
	ManagedIdentity       AzureManagedIdentityName `json:"managedIdentity"`
	ResourceGroup         string                   `json:"resourceGroup"`
	ShouldCreateContainer bool                     `json:"shouldCreateContainer"`
	StorageAccount        string                   `json:"storageAccount"`
	StorageSecret         string                   `json:"storageSecret"`
}

// GcpEsConfigInput represents the input for creating a GCP ES config.
type GcpEsConfigInput struct {
	BucketName         string `json:"bucketName"`
	Region             string `json:"region"`
	ShouldCreateBucket bool   `json:"shouldCreateBucket"`
}

// OciEsConfigInput represents the input for creating an OCI ES config.
type OciEsConfigInput struct {
	AccessKey    string `json:"accessKey"`
	BucketName   string `json:"bucketName"`
	OciNamespace string `json:"ociNamespace"`
	SecretKey    string `json:"secretKey"`
}
