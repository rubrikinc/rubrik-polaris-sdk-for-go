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
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
)

// GcpCCInstanceType represents the instance types for GCP Cloud Cluster.
type GcpCCInstanceType string

const (
	GcpInstanceTypeUnspecified   GcpCCInstanceType = "GCP_TYPE_UNSPECIFIED"
	GcpInstanceTypeN2Standard8   GcpCCInstanceType = "N2_STANDARD_8"
	GcpInstanceTypeN2Standard16  GcpCCInstanceType = "N2_STANDARD_16"
	GcpInstanceTypeN2Highmem16   GcpCCInstanceType = "N2_HIGHMEM_16"
	GcpInstanceTypeN2DStandard8  GcpCCInstanceType = "N2D_STANDARD_8"
	GcpInstanceTypeN2DStandard16 GcpCCInstanceType = "N2D_STANDARD_16"
	GcpInstanceTypeN2DHighmem16  GcpCCInstanceType = "N2D_HIGHMEM_16"
)

// GcpCdmVersion represents the CDM version for GCP Cloud Cluster.
type GcpCdmVersion struct {
	CdmVersion             string              `json:"cdmVersion"`
	CdmProduct             string              `json:"cdmProduct"`
	ImageID                string              `json:"imageId"`
	IsLatest               bool                `json:"isLatest"`
	SupportedInstanceTypes []GcpCCInstanceType `json:"supportedInstanceTypes"`
}

// GcpRegionInfo represents a GCP region with its availability zones.
type GcpRegionInfo struct {
	Name  string   `json:"name"`
	Zones []string `json:"zones"`
}

// GcpSubnet represents a GCP subnet.
type GcpSubnet struct {
	HostProject string `json:"hostProject"`
	Name        string `json:"name"`
	Network     string `json:"network"`
	Region      string `json:"region"`
}

// GcpServiceAccount represents a GCP service account.
type GcpServiceAccount struct {
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

// GcpBucket represents a GCP storage bucket.
type GcpBucket struct {
	BucketName         string `json:"bucketName"`
	Region             string `json:"region"`
	ShouldCreateBucket bool   `json:"shouldCreateBucket"`
}

// GcpSubnetInput represents the subnet input for GCP Cloud Cluster VM config.
type GcpSubnetInput struct {
	HostProject string `json:"hostProject,omitempty"`
	Name        string `json:"name,omitempty"`
	Network     string `json:"network,omitempty"`
	Region      string `json:"region,omitempty"`
}

// GcpServiceAccountInput represents the service account input for GCP Cloud Cluster VM config.
type GcpServiceAccountInput struct {
	Email  string   `json:"email,omitempty"`
	Name   string   `json:"name,omitempty"`
	Scopes []string `json:"scopes,omitempty"`
}

// SubnetAzConfig represents the subnet availability zone configuration for GCP Cloud Cluster.
type SubnetAzConfig struct {
	AvailabilityZone string `json:"availabilityZone,omitempty"`
	Subnet           string `json:"subnet,omitempty"`
}

// GcpVmConfig represents the VM configuration for the GCP Cloud Cluster.
type GcpVmConfig struct {
	CDMProduct       string                   `json:"cdmProduct,omitempty"`
	CDMVersion       string                   `json:"cdmVersion,omitempty"`
	DeleteProtection bool                     `json:"deleteProtection"`
	InstanceType     GcpCCInstanceType        `json:"instanceType"`
	NetworkConfig    []GcpSubnetInput         `json:"networkConfig"`
	ServiceAccounts  []GcpServiceAccountInput `json:"serviceAccounts"`
	SubnetAzConfigs  []SubnetAzConfig         `json:"subnetAzConfigs,omitempty"`
	VMType           VmConfigType             `json:"vmType"`
}

// GcpClusterConfig represents the cluster configuration for the GCP Cloud Cluster.
type GcpClusterConfig struct {
	ClusterName           string           `json:"clusterName"`
	UserEmail             string           `json:"userEmail"`
	AdminPassword         secret.String    `json:"adminPassword"`
	DNSNameServers        []string         `json:"dnsNameServers"`
	DNSSearchDomains      []string         `json:"dnsSearchDomains"`
	NTPServers            []string         `json:"ntpServers"`
	NumNodes              int              `json:"numNodes"`
	GcpEsConfig           GcpEsConfigInput `json:"gcpEsConfig"`
	DynamicScalingEnabled bool             `json:"dynamicScalingEnabled,omitempty"`
}

// CreateGcpClusterInput represents the input for creating a GCP Cloud Cluster.
type CreateGcpClusterInput struct {
	CloudAccountID       uuid.UUID                  `json:"cloudAccountId"`
	ClusterConfig        GcpClusterConfig           `json:"clusterConfig"`
	IsAzResilient        *bool                      `json:"isAzResilient,omitempty"`
	IsEsType             bool                       `json:"isEsType"`
	KeepClusterOnFailure bool                       `json:"keepClusterOnFailure"`
	Region               string                     `json:"region"`
	Validations          []ClusterCreateValidations `json:"validations"`
	VMConfig             GcpVmConfig                `json:"vmConfig"`
	Zone                 string                     `json:"zone"`
}

// DeleteGcpClusterInput represents the input for deleting a GCP Cloud Cluster.
type DeleteGcpClusterInput struct {
	CloudAccountID uuid.UUID `json:"cloudAccountId"`
	ClusterUUID    uuid.UUID `json:"clusterUuid"`
	IsEsType       bool      `json:"isEsType"`
	NumNodes       int       `json:"numNodes,omitempty"`
}

// AllGcpCdmVersions returns all the available CDM versions for the specified
// GCP cloud account.
func (a API) AllGcpCdmVersions(ctx context.Context, cloudAccountID uuid.UUID) ([]GcpCdmVersion, error) {
	query := gcpCcCdmVersionsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input struct {
			CloudAccountID string `json:"cloudAccountId"`
		} `json:"input"`
	}{Input: struct {
		CloudAccountID string `json:"cloudAccountId"`
	}{CloudAccountID: cloudAccountID.String()}})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				CdmVersions []GcpCdmVersion `json:"cdmVersions"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.CdmVersions, nil
}

// GcpRegions returns all the available regions for the specified GCP cloud account.
func (a API) GcpRegions(ctx context.Context, cloudAccountID uuid.UUID) ([]GcpRegionInfo, error) {
	query := gcpCcRegionsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input struct {
			CloudAccountID string `json:"cloudAccountId"`
		} `json:"input"`
	}{Input: struct {
		CloudAccountID string `json:"cloudAccountId"`
	}{CloudAccountID: cloudAccountID.String()}})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				GcpRegions []GcpRegionInfo `json:"gcpRegions"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.GcpRegions, nil
}

// GcpSubnets returns all the available subnets for the specified GCP cloud account and region.
func (a API) GcpSubnets(ctx context.Context, cloudAccountID uuid.UUID, region string) ([]GcpSubnet, error) {
	query := gcpCcSubnetsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input struct {
			CloudAccountID string `json:"cloudAccountId"`
			Region         string `json:"region"`
		} `json:"input"`
	}{Input: struct {
		CloudAccountID string `json:"cloudAccountId"`
		Region         string `json:"region"`
	}{CloudAccountID: cloudAccountID.String(), Region: region}})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Subnets []GcpSubnet `json:"subnets"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.Subnets, nil
}

// GcpServiceAccounts returns all the available service accounts for the specified GCP cloud account.
func (a API) GcpServiceAccounts(ctx context.Context, cloudAccountID uuid.UUID) ([]GcpServiceAccount, error) {
	query := gcpCcServiceAccountsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input struct {
			CloudAccountID string `json:"cloudAccountId"`
		} `json:"input"`
	}{Input: struct {
		CloudAccountID string `json:"cloudAccountId"`
	}{CloudAccountID: cloudAccountID.String()}})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				ServiceAccounts []GcpServiceAccount `json:"serviceAccounts"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.ServiceAccounts, nil
}

// GcpBuckets returns all the available buckets for the specified GCP cloud account.
func (a API) GcpBuckets(ctx context.Context, cloudAccountID uuid.UUID) ([]GcpBucket, error) {
	query := gcpCcBucketsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input struct {
			CloudAccountID string `json:"cloudAccountId"`
		} `json:"input"`
	}{Input: struct {
		CloudAccountID string `json:"cloudAccountId"`
	}{CloudAccountID: cloudAccountID.String()}})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				Buckets []GcpBucket `json:"buckets"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.Buckets, nil
}

// ValidateCreateGcpClusterInput validates the GCP cloud cluster create request.
func (a API) ValidateCreateGcpClusterInput(ctx context.Context, input CreateGcpClusterInput) error {
	query := validateGcpClusterCreateRequestQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input CreateGcpClusterInput `json:"input"`
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

	if !payload.Data.Result.IsSuccessful {
		return graphql.ResponseError(query, errors.New(payload.Data.Result.Message))
	}

	return nil
}

// CreateGcpCloudCluster creates a GCP Cloud Cluster in RSC.
func (a API) CreateGcpCloudCluster(ctx context.Context, input CreateGcpClusterInput) (uuid.UUID, error) {
	query := createGcpCloudClusterQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input CreateGcpClusterInput `json:"input"`
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

	if !payload.Data.Result.Success {
		return uuid.Nil, graphql.ResponseError(query, errors.New(payload.Data.Result.Message))
	}

	// Use regex to find the UUID in the message string.
	match := uuidRegex.FindString(payload.Data.Result.Message)
	jobID, err := uuid.Parse(match)
	if err != nil {
		return uuid.Nil, err
	}

	return jobID, nil
}

// DeleteGcpCloudCluster deletes a GCP Cloud Cluster in RSC.
func (a API) DeleteGcpCloudCluster(ctx context.Context, input DeleteGcpClusterInput) (uuid.UUID, error) {
	query := deleteGcpCcClusterQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input DeleteGcpClusterInput `json:"input"`
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

	if !payload.Data.Result.Success {
		return uuid.Nil, graphql.ResponseError(query, errors.New(payload.Data.Result.Message))
	}

	// Use regex to find the UUID in the message string.
	match := uuidRegex.FindString(payload.Data.Result.Message)
	jobID, err := uuid.Parse(match)
	if err != nil {
		return uuid.Nil, err
	}

	return jobID, nil
}
