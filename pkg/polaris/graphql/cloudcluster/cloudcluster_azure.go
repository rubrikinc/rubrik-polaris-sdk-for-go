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

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
	azure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
)

// AzureCCESSupportedInstanceType represents the instance types for Azure Cloud Cluster.
type AzureCCESSupportedInstanceType string

const (
	AzureInstanceTypeUnspecified     AzureCCESSupportedInstanceType = "TYPE_UNSPECIFIED"
	AzureInstanceTypeStandardDS5V2   AzureCCESSupportedInstanceType = "STANDARD_DS5_V2"
	AzureInstanceTypeStandardD16SV5  AzureCCESSupportedInstanceType = "STANDARD_D16S_V5"
	AzureInstanceTypeStandardD8SV5   AzureCCESSupportedInstanceType = "STANDARD_D8S_V5"
	AzureInstanceTypeStandardD32SV5  AzureCCESSupportedInstanceType = "STANDARD_D32S_V5"
	AzureInstanceTypeStandardE16SV5  AzureCCESSupportedInstanceType = "STANDARD_E16S_V5"
	AzureInstanceTypeStandardD8ASV5  AzureCCESSupportedInstanceType = "STANDARD_D8AS_V5"
	AzureInstanceTypeStandardD16ASV5 AzureCCESSupportedInstanceType = "STANDARD_D16AS_V5"
	AzureInstanceTypeStandardD32ASV5 AzureCCESSupportedInstanceType = "STANDARD_D32AS_V5"
	AzureInstanceTypeStandardE16ASV5 AzureCCESSupportedInstanceType = "STANDARD_E16AS_V5"
)

// AzureCDMVersionTag represents a tag for a CDM version.
type AzureCDMVersionTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// AzureCDMVersions represents the CDM versions for Azure Cloud Cluster.
type AzureCDMVersions struct {
	CDMVersion             string                           `json:"cdmVersion"`
	SKU                    string                           `json:"sku"`
	SupportedInstanceTypes []AzureCCESSupportedInstanceType `json:"supportedInstanceTypes"`
	Tags                   []AzureCDMVersionTag             `json:"tags"`
	Version                string                           `json:"version"`
}

// AllAzureCdmVersions returns all the available CDM versions for the specified cloud account.
func (a API) AllAzureCdmVersions(ctx context.Context, cloudAccountID uuid.UUID, region azure.Region) ([]AzureCDMVersions, error) {
	query := azureCcCdmVersionsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID uuid.UUID                    `json:"cloudAccountId"`
		Location       azure.CloudAccountRegionEnum `json:"location"`
	}{CloudAccountID: cloudAccountID, Location: region.ToCloudAccountRegionEnum()})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []AzureCDMVersions `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// AzureCCRegionDetails represents the region details available for the Azure Cloud Cluster.
type AzureCCRegionDetails struct {
	Location                 azure.CloudAccountRegionEnum `json:"location"`
	LogicalAvailabilityZones []azure.NativeRegionEnum     `json:"logicalAvailabilityZones"`
}

// AzureCCRegionDetails returns all the available regions for the specified cloud account.
func (a API) AzureCCRegionDetails(ctx context.Context, cloudAccountID uuid.UUID) ([]AzureCCRegionDetails, error) {
	query := azureCcRegionQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID uuid.UUID `json:"cloudAccountId"`
	}{CloudAccountID: cloudAccountID})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []AzureCCRegionDetails `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// AzureMarketplaceTerms represents the marketplace terms for the Azure Cloud Cluster CDM Version requested.
type AzureMarketplaceTerms struct {
	MarketplaceSKU       string `json:"marketplaceSku"`
	MarketplaceTermsLink string `json:"marketplaceTermsLink"`
	Message              string `json:"message"`
	Offer                string `json:"offer"`
	Publisher            string `json:"publisher"`
	TermsAccepted        bool   `json:"termsAccepted"`
}

// AzureMarketplaceTerms returns the marketplace terms for the specified cloud account and CDM version.
func (a API) AzureMarketplaceTerms(ctx context.Context, cloudAccountID uuid.UUID, cdmVersion string) (AzureMarketplaceTerms, error) {
	query := azureCcMarketplaceTermsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID uuid.UUID `json:"cloudAccountId"`
		CdmVersion     string    `json:"cdmVersion"`
	}{CloudAccountID: cloudAccountID, CdmVersion: cdmVersion})
	if err != nil {
		return AzureMarketplaceTerms{}, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result AzureMarketplaceTerms `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return AzureMarketplaceTerms{}, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// AzureCCResourceGroup represents the resource group and region for the specific cloud account available to the cloud cluster.
type AzureCCResourceGroup struct {
	Name   string                 `json:"name"`
	Region azure.NativeRegionEnum `json:"region"`
}

// AzureCCResourceGroups returns all the available resource groups for the specified cloud account.
func (a API) AzureCCResourceGroups(ctx context.Context, cloudAccountID uuid.UUID, azureSubscriptionID uuid.UUID) ([]AzureCCResourceGroup, error) {
	query := azureCcResourceGroupQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID      uuid.UUID `json:"cloudAccountId"`
		AzureSubscriptionID uuid.UUID `json:"azureSubscriptionNativeId"`
	}{CloudAccountID: cloudAccountID, AzureSubscriptionID: azureSubscriptionID})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []AzureCCResourceGroup `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// AzureCCManagedIdentity represents the managed identity for the specific cloud account available to the cloud cluster.
type AzureCCManagedIdentity struct {
	Name          string `json:"name"`
	ClientID      string `json:"clientId"`
	ResourceGroup string `json:"resourceGroup"`
}

// AzureCCManagedIdentities returns all the available managed identities for the specified cloud account.
func (a API) AzureCCManagedIdentities(ctx context.Context, cloudAccountID uuid.UUID) ([]AzureCCManagedIdentity, error) {
	query := azureCcManagedIdentitiesQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID uuid.UUID `json:"cloudAccountId"`
	}{CloudAccountID: cloudAccountID})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []AzureCCManagedIdentity `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// AzureCCVnet represents the vnet for the specific cloud account available to the cloud cluster.
type AzureCCVnet struct {
	Name            string `json:"name"`
	ResourceGroup   string `json:"resourceGroupName"`
	ResourceGroupID string `json:"resourceGroupId"`
}

// AzureCCSubnet represents the subnet for the specific cloud account available to the cloud cluster.
type AzureCCSubnet struct {
	Name string      `json:"name"`
	ID   string      `json:"nativeId"`
	Vnet AzureCCVnet `json:"vnet"`
}

// AzureCCSubnets returns all the available subnets for the specified cloud account.
func (a API) AzureCCSubnets(ctx context.Context, cloudAccountID uuid.UUID, region azure.Region) ([]AzureCCSubnet, error) {
	query := azureCcSubnetQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID uuid.UUID                    `json:"cloudAccountId"`
		Region         azure.CloudAccountRegionEnum `json:"region"`
	}{CloudAccountID: cloudAccountID, Region: region.ToCloudAccountRegionEnum()})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []AzureCCSubnet `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// AzureCCStorageAccount represents the storage account for the specific cloud account available to the cloud cluster.
type AzureCCStorageAccount struct {
	Name          string `json:"name"`
	ResourceGroup string `json:"resourceGroup"`
}

// AzureCCStorageAccounts returns all the available storage accounts for the specified cloud account.
func (a API) AzureCCStorageAccounts(ctx context.Context, cloudAccountID uuid.UUID, region azure.Region) ([]AzureCCStorageAccount, error) {
	query := azureCcStorageAccountsQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		CloudAccountID uuid.UUID    `json:"cloudAccountId"`
		Region         azure.Region `json:"region"`
	}{CloudAccountID: cloudAccountID, Region: region})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []AzureCCStorageAccount `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// AzureClusterConfig represents the cluster configuration for the Azure Cloud Cluster.
type AzureClusterConfig struct {
	ClusterName           string             `json:"clusterName"`
	UserEmail             string             `json:"userEmail"`
	AdminPassword         secret.String      `json:"adminPassword"`
	DNSNameServers        []string           `json:"dnsNameServers"`
	DNSSearchDomains      []string           `json:"dnsSearchDomains"`
	NTPServers            []string           `json:"ntpServers"`
	NumNodes              int                `json:"numNodes"`
	DynamicScalingEnabled bool               `json:"dynamicScalingEnabled"`
	AzureESConfig         AzureEsConfigInput `json:"azureEsConfig"`
}

// AzureVMConfig represents the VM configuration for the Azure Cloud Cluster.
type AzureVMConfig struct {
	CDMVersion                   string                         `json:"cdmVersion"`
	Subnet                       string                         `json:"subnet"`
	VMType                       VmConfigType                   `json:"vmType"`
	CDMProduct                   string                         `json:"cdmProduct"`
	Location                     azure.Region                   `json:"location"`
	AvailabilityZone             string                         `json:"availabilityZone"`
	Vnet                         string                         `json:"vnet"`
	ResourceGroup                string                         `json:"resourceGroup"`
	NetworkResourceGroup         string                         `json:"networkResourceGroup"`
	VnetResourceGroup            string                         `json:"vnetResourceGroup"`
	InstanceType                 AzureCCESSupportedInstanceType `json:"instanceType"`
	NetworkSecurityGroup         string                         `json:"networkSecurityGroup"`
	NetworkSecurityResourceGroup string                         `json:"networkSecurityResourceGroup"`
}

// CreateAzureClusterInput represents the input for creating an Azure Cloud Cluster.
type CreateAzureClusterInput struct {
	CloudAccountID       uuid.UUID                  `json:"cloudAccountId"`
	ClusterConfig        AzureClusterConfig         `json:"clusterConfig"`
	IsESType             bool                       `json:"isEsType"`
	KeepClusterOnFailure bool                       `json:"keepClusterOnFailure"`
	Validations          []ClusterCreateValidations `json:"validations"`
	VMConfig             AzureVMConfig              `json:"vmConfig"`
}

// ValidateCreateAzureClusterInput validates the create Azure cluster input.
func (a API) ValidateCreateAzureClusterInput(ctx context.Context, input CreateAzureClusterInput) error {
	query := validateAzureClusterCreateRequestQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input CreateAzureClusterInput `json:"input"`
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

// CreateAzureCloudCluster creates an Azure Cloud Cluster.
func (a API) CreateAzureCloudCluster(ctx context.Context, input CreateAzureClusterInput) (uuid.UUID, error) {
	query := createAzureCcClusterQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		Input CreateAzureClusterInput `json:"input"`
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

	return uuid.Nil, nil
}
