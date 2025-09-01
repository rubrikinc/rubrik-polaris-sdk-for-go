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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
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

type ClusterProductEnum string

const (
	CDM          ClusterProductEnum = "CDM"
	CLOUD_DIRECT ClusterProductEnum = "CLOUD_DIRECT"
	DATOS        ClusterProductEnum = "DATOS"
	POLARIS      ClusterProductEnum = "POLARIS"
)

type ClusterProductTypeEnum string

const (
	CLOUD      ClusterProductTypeEnum = "Cloud"
	RSC        ClusterProductTypeEnum = "Polaris"
	EXOCOMPUTE ClusterProductTypeEnum = "ExoCompute"
	ONPREM     ClusterProductTypeEnum = "OnPrem"
	ROBO       ClusterProductTypeEnum = "Robo"
	UNKNOWN    ClusterProductTypeEnum = "Unknown"
)

type ClusterStatusEnum string

const (
	ClusterConnected    ClusterStatusEnum = "Connected"
	ClusterDisconnected ClusterStatusEnum = "Disconnected"
	ClusterInitializing ClusterStatusEnum = "Initializing"
)

type ClusterSystemStatusEnum string

const (
	ClusterSystemStatusOK      ClusterSystemStatusEnum = "OK"
	ClusterSystemStatusWARNING ClusterSystemStatusEnum = "WARNING"
	ClusterSystemStatusFATAL   ClusterSystemStatusEnum = "FATAL"
)

type ClusterFilterInput struct {
	ID              []string                  `json:"id"`
	Name            []string                  `json:"name"`
	Type            []ClusterProductTypeEnum  `json:"type"`
	ConnectionState []ClusterStatusEnum       `json:"connectionState"`
	SystemStatus    []ClusterSystemStatusEnum `json:"systemStatus"`
	ProductType     []ClusterProductEnum      `json:"productType"`
}

// ClusterSortByEnum represents the valid sort by values.
type ClusterSortByEnum string

const (
	SortByEstimatedRunway  ClusterSortByEnum = "ESTIMATED_RUNWAY"
	SortByInstalledVersion ClusterSortByEnum = "INSTALLED_VERSION"
	SortByClusterName      ClusterSortByEnum = "ClusterName"
	SortByClusterType      ClusterSortByEnum = "ClusterType"
	SortByClusterLocation  ClusterSortByEnum = "CLUSTER_LOCATION"
	SortByRegisteredAt     ClusterSortByEnum = "RegisteredAt"
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
}

// CloudCluster represents the cloud cluster.
type CloudCluster struct {
	ID            uuid.UUID                 `json:"id"`
	Name          string                    `json:"name"`
	ProvisionInfo CloudClusterProvisionInfo `json:"ccprovisionInfo"`
	Vendor        core.CloudVendor          `json:"vendor"`
}

// AllCloudClusters returns all cloud clusters.
func (a API) AllCloudClusters(ctx context.Context, first int, after string, filter ClusterFilterInput, sortBy ClusterSortByEnum, sortOrder core.SortOrderEnum) ([]CloudCluster, error) {
	a.log.Print(log.Trace)

	query := allClustersConnectionQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		First     int                `json:"first"`
		After     string             `json:"after,omitempty"`
		Filter    ClusterFilterInput `json:"filter"`
		SortBy    ClusterSortByEnum  `json:"sortBy"`
		SortOrder core.SortOrderEnum `json:"sortOrder"`
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
			} `json:"allClusterConnection"`
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
