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
	"context"
	"encoding/json"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/cluster"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
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

// SubnetAzConfig represents a subnet and availability zone pair for Multi-AZ
// deployments. Used when IsAzResilient is true.
type SubnetAzConfig struct {
	AvailabilityZone string `json:"availabilityZone"`
	Subnet           string `json:"subnet"`
}

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
	EndpointSuffix        string                   `json:"endpointSuffix,omitempty"`
	ManagedIdentity       AzureManagedIdentityName `json:"managedIdentity"`
	ResourceGroup         string                   `json:"resourceGroup"`
	ShouldCreateContainer bool                     `json:"shouldCreateContainer"`
	StorageAccount        string                   `json:"storageAccount"`
	StorageSecret         secret.String            `json:"storageSecret,omitempty"`
}

// GcpEsConfigInput represents the input for creating a GCP ES config.
type GcpEsConfigInput struct {
	BucketName string `json:"bucketName"`
	// Region should be the GCS bucket's actual location. RSC requires the bucket
	// to be in the same region as the cluster, and the high-level
	// CreateGcpCloudCluster validates this against the cluster region. Note the
	// RSC UI instead sets this to the cluster region and filters buckets by
	// location client-side, so when following UI semantics this is always the
	// cluster region.
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

// CcpJobStatus represents the valid job statuses.
//
// Deprecated: use cluster.CCPJobStatus instead.
type CcpJobStatus = cluster.CCPJobStatus

// Deprecated: use the cluster.CCPJobStatus* constants instead.
const (
	CcpJobStatusInitializing               = cluster.CCPJobStatusInitializing
	CcpJobStatusNodeCreate                 = cluster.CCPJobStatusNodeCreate
	CcpJobStatusNodeConnectionVerification = cluster.CCPJobStatusNodeConnectionVerification
	CcpJobStatusNodeInfoExtraction         = cluster.CCPJobStatusNodeInfoExtraction
	CcpJobStatusBootstrapping              = cluster.CCPJobStatusBootstrapping
	CcpJobStatusRotateToken                = cluster.CCPJobStatusRotateToken
	CcpJobStatusFailed                     = cluster.CCPJobStatusFailed
	CcpJobStatusCompleted                  = cluster.CCPJobStatusCompleted
	CcpJobStatusInvalid                    = cluster.CCPJobStatusInvalid
)

// CcpJobType represents the valid job types.
//
// Deprecated: use cluster.CCPJobType instead.
type CcpJobType = cluster.CCPJobType

// Deprecated: use the cluster.CCPJobType* constants instead.
const (
	CcpJobTypeClusterCreate                   = cluster.CCPJobTypeClusterCreate
	CcpJobTypeClusterDelete                   = cluster.CCPJobTypeClusterDelete
	CcpJobTypeAddNode                         = cluster.CCPJobTypeAddNode
	CcpJobTypeRemoveNode                      = cluster.CCPJobTypeRemoveNode
	CcpJobTypeReplaceNode                     = cluster.CCPJobTypeReplaceNode
	CcpJobTypeClusterRecover                  = cluster.CCPJobTypeClusterRecover
	CcpJobTypeClusterOps                      = cluster.CCPJobTypeClusterOps
	CcpJobTypeMigrateNodes                    = cluster.CCPJobTypeMigrateNodes
	CcpJobTypeMigrateClusterToManagedIdentity = cluster.CCPJobTypeMigrateClusterToManagedIdentity
	CcpJobTypeManualAddNodes                  = cluster.CCPJobTypeManualAddNodes
)

// CloudClusterProvisionInfo represents the cloud cluster provision info.
//
// Deprecated: use cluster.ProvisionInfo instead.
type CloudClusterProvisionInfo = cluster.ProvisionInfo

// CloudClusterStorageConfig represents the cluster's cloud storage
// configuration.
//
// Deprecated: use cluster.StorageConfig instead.
type CloudClusterStorageConfig = cluster.StorageConfig

// CloudClusterCloudInfo represents the cloud placement of a cluster.
//
// Deprecated: use cluster.CloudInfo instead.
type CloudClusterCloudInfo = cluster.CloudInfo

// CloudClusterNode represents a single node within a cluster.
//
// Deprecated: use cluster.Node instead.
type CloudClusterNode = cluster.Node

// CloudClusterNodeConnection is the paginated list of nodes within a cluster.
//
// Deprecated: use cluster.NodeConnection instead.
type CloudClusterNodeConnection = cluster.NodeConnection

// CloudCluster represents a cluster registered with the cluster management
// service, including in-flight cloud clusters being provisioned.
//
// Deprecated: use cluster.Cluster instead.
type CloudCluster = cluster.Cluster

// AllCloudClusters returns all cloud clusters. It is a thin wrapper around
// cluster.AllClusters that returns the clusters of a single page.
//
// Deprecated: use cluster.API.ListClusters or cluster.AllClusters instead.
func (a API) AllCloudClusters(ctx context.Context, first int, after string, filter cluster.SearchFilter, sortBy cluster.SortBy, sortOrder core.SortOrder) ([]CloudCluster, error) {
	a.log.Print(log.Trace)

	page, err := cluster.AllClusters(ctx, a.GQL, first, after, filter, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}
	return page.Clusters, nil
}

// CloudClusterInstanceProperties represents the cloud cluster instance properties.
type CloudClusterInstanceProperties struct {
	InstanceType       string `json:"instanceType"`
	Vendor             string `json:"vendor"`
	VcpuCount          int    `json:"vcpuCount"`
	MemoryGib          int    `json:"memoryGib"`
	CapacityTb         int    `json:"capacityTb"`
	ProcessorType      string `json:"processorType"`
	VmType             string `json:"vmType"`
	InstanceTypeString string `json:"instanceTypeString"`
}

// CloudClusterInstancePropertiesRequest represents the request for cloud cluster instance properties.
type CloudClusterInstancePropertiesRequest struct {
	CloudVendor  string `json:"cloudVendor"`
	InstanceType string `json:"instanceType"`
}

// CloudClusterInstanceProperties returns the cloud cluster instance properties.
func (a API) CloudClusterInstanceProperties(ctx context.Context, input CloudClusterInstancePropertiesRequest) (CloudClusterInstanceProperties, error) {
	a.log.Print(log.Trace)

	query := cloudClusterInstancePropertiesQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input CloudClusterInstancePropertiesRequest `json:"input"`
	}{Input: input})

	if err != nil {
		return CloudClusterInstanceProperties{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				InstanceProperties CloudClusterInstanceProperties `json:"instanceProperties"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return CloudClusterInstanceProperties{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.InstanceProperties, nil
}
