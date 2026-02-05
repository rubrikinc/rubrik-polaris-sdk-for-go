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

	"github.com/google/uuid"
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

// CcpJobStatus represents the valid job statuses.
type CcpJobStatus string

const (
	CcpJobStatusInitializing               CcpJobStatus = "INITIALIZING"
	CcpJobStatusNodeCreate                 CcpJobStatus = "NODE_CREATE"
	CcpJobStatusNodeConnectionVerification CcpJobStatus = "NODE_CONNECTION_VERIFICATION"
	CcpJobStatusNodeInfoExtraction         CcpJobStatus = "NODE_INFO_EXTRACTION"
	CcpJobStatusBootstrapping              CcpJobStatus = "BOOTSTRAPPING"
	CcpJobStatusRotateToken                CcpJobStatus = "ROTATE_TOKEN"
	CcpJobStatusFailed                     CcpJobStatus = "FAILED"
	CcpJobStatusCompleted                  CcpJobStatus = "COMPLETED"
	CcpJobStatusInvalid                    CcpJobStatus = "INVALID"
)

// CcpJobType represents the valid job types.
type CcpJobType string

const (
	CcpJobTypeClusterCreate                   CcpJobType = "CLUSTER_CREATE"
	CcpJobTypeClusterDelete                   CcpJobType = "CLUSTER_DELETE"
	CcpJobTypeAddNode                         CcpJobType = "ADD_NODE"
	CcpJobTypeRemoveNode                      CcpJobType = "REMOVE_NODE"
	CcpJobTypeReplaceNode                     CcpJobType = "REPLACE_NODE"
	CcpJobTypeClusterRecover                  CcpJobType = "CLUSTER_RECOVER"
	CcpJobTypeClusterOps                      CcpJobType = "CLUSTER_OPS"
	CcpJobTypeMigrateNodes                    CcpJobType = "MIGRATE_NODES"
	CcpJobTypeMigrateClusterToManagedIdentity CcpJobType = "MIGRATE_CLUSTER_TO_MANAGED_IDENTITY"
	CcpJobTypeManualAddNodes                  CcpJobType = "MANUAL_ADD_NODES"
)

// CloudClusterProvisionInfo represents the cloud cluster provision info.
type CloudClusterProvisionInfo struct {
	Progress  int          `json:"progress"`
	JobStatus CcpJobStatus `json:"jobStatus"`
	JobType   CcpJobType   `json:"jobType"`
	Vendor    string       `json:"vendor"`
}

type CloudClusterStorageConfig struct {
	LocationName           string `json:"locationName"`
	LocationID             string `json:"locationId"`
	IsImmutable            bool   `json:"isImmutable"`
	IsUsingManagedIdentity bool   `json:"isUsingManagedIdentity"`
}

type CloudClusterCloudInfo struct {
	Name                   string                    `json:"name"`
	Region                 string                    `json:"region"`
	RegionID               string                    `json:"regionId"`
	NetworkName            string                    `json:"networkName"`
	NativeCloudAccountName string                    `json:"nativeCloudAccountName"`
	Vendor                 string                    `json:"vendor"`
	NativeCloudAccountID   string                    `json:"nativeCloudAccountId"`
	CloudAccount           string                    `json:"cloudAccount"`
	StorageConfig          CloudClusterStorageConfig `json:"storageConfig"`
}

type CloudClusterNode struct {
	BrikID          string `json:"brikId"`
	IpAddress       string `json:"ipAddress"`
	NeedsInspection bool   `json:"needsInspection"`
	CpuCores        int    `json:"cpuCores,omitempty"`
	Ram             int64  `json:"ram,omitempty"`
	ClusterID       string `json:"clusterId"`
	NetworkSpeed    string `json:"networkSpeed,omitempty"`
	Hostname        string `json:"hostname,omitempty"`
	ID              string `json:"id"`
}

type CloudClusterNodeConnection struct {
	Edges []struct {
		Node CloudClusterNode `json:"node"`
	} `json:"edges"`
}

// CloudCluster represents the cloud cluster.
type CloudCluster struct {
	ID            uuid.UUID                  `json:"id"`
	Name          string                     `json:"name"`
	ProvisionInfo CloudClusterProvisionInfo  `json:"ccprovisionInfo"`
	CloudInfo     CloudClusterCloudInfo      `json:"cloudInfo,omitempty"`
	ClusterNodes  CloudClusterNodeConnection `json:"clusterNodeConnection"`
	ProductType   cluster.Product            `json:"productType"`
	Timezone      cluster.Timezone           `json:"timezone"`
	Version       string                     `json:"version"`
}

// AllCloudClusters returns all cloud clusters.
func (a API) AllCloudClusters(ctx context.Context, first int, after string, filter cluster.SearchFilter, sortBy cluster.SortBy, sortOrder core.SortOrder) ([]CloudCluster, error) {
	a.log.Print(log.Trace)

	query := allClustersConnectionQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		First     int                  `json:"first"`
		After     string               `json:"after,omitempty"`
		Filter    cluster.SearchFilter `json:"filter"`
		SortBy    cluster.SortBy       `json:"sortBy"`
		SortOrder core.SortOrder       `json:"sortOrder"`
	}{First: first, After: after, Filter: filter, SortBy: sortBy, SortOrder: sortOrder})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Edges []struct {
					Node CloudCluster `json:"node"`
				} `json:"edges"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	var clusters []CloudCluster
	for _, edge := range payload.Data.Result.Edges {
		clusters = append(clusters, edge.Node)
	}

	return clusters, nil
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
