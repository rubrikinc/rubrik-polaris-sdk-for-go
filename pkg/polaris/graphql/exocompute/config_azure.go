// Copyright 2024 Rubrik, Inc.
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

package exocompute

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	gqlazure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
)

// AzureConfigurationsFilter holds the filter for an Azure exocompute
// configuration list operation.
type AzureConfigurationsFilter struct {
	SearchQuery string `json:"azureExocomputeSearchQuery"`
}

func (p AzureConfigurationsFilter) ListQuery() (string, any, AzureConfigurationsForCloudAccount) {
	return allAzureExocomputeConfigsInAccountQuery, p, AzureConfigurationsForCloudAccount{}
}

// AzureConfigurationsForCloudAccount holds the result of an Azure exocompute
// configuration list operation.
type AzureConfigurationsForCloudAccount struct {
	CloudAccount gqlazure.CloudAccount `json:"azureCloudAccount"`
	Configs      []AzureConfiguration  `json:"configs"`
}

// AzureConfiguration holds a single Azure exocompute configuration.
type AzureConfiguration struct {
	ID                    uuid.UUID                    `json:"configUuid"`
	Region                azure.CloudAccountRegionEnum `json:"region"`
	SubnetID              string                       `json:"subnetNativeId"`
	ManagedByRubrik       bool                         `json:"isRscManaged"`
	PodOverlayNetworkCIDR string                       `json:"podOverlayNetworkCidr"`
	PodSubnetID           string                       `json:"podSubnetNativeId"`
	OptionalConfig        *AzureOptionalConfig         `json:"optionalConfig"`

	// HealthCheckStatus represents the health status of an exocompute cluster.
	HealthCheckStatus struct {
		Status        string `json:"status"`
		FailureReason string `json:"failureReason"`
		LastUpdatedAt string `json:"lastUpdatedAt"`
		TaskchainID   string `json:"taskchainId"`
	} `json:"healthCheckStatus"`
}

// CreateAzureConfigurationParams holds the parameters for an Azure exocompute
// configuration create operation.
type CreateAzureConfigurationParams struct {
	CloudAccountID uuid.UUID                    `json:"-"`
	Region         azure.CloudAccountRegionEnum `json:"region"`

	// When true, Rubrik will manage the security groups.
	IsManagedByRubrik bool `json:"isRscManaged"`

	SubnetID              string               `json:"subnetNativeId,omitempty"`
	PodOverlayNetworkCIDR string               `json:"podOverlayNetworkCidr,omitempty"`
	PodSubnetID           string               `json:"podSubnetNativeId,omitempty"`
	OptionalConfig        *AzureOptionalConfig `json:"optionalConfig,omitempty"`

	// When true, a health check will be triggered after the configuration is
	// created.
	TriggerHealthCheck bool `json:"-"`
}

func (p CreateAzureConfigurationParams) CreateQuery() (string, any, CreateAzureConfigurationResult) {
	params := struct {
		CloudAccountID     uuid.UUID                      `json:"cloudAccountId"`
		Config             CreateAzureConfigurationParams `json:"azureExocomputeRegionConfig"`
		TriggerHealthCheck bool                           `json:"triggerHealthCheck"`
	}{CloudAccountID: p.CloudAccountID, Config: p, TriggerHealthCheck: p.TriggerHealthCheck}
	return addAzureCloudAccountExocomputeConfigurationsQuery, params, CreateAzureConfigurationResult{}
}

// CreateAzureConfigurationResult holds the result of an Azure exocompute
// configuration create operation.
type CreateAzureConfigurationResult struct {
	Configs []struct {
		ID      string `json:"configUuid"`
		Message string `json:"message"`
	} `json:"configs"`
}

func (r CreateAzureConfigurationResult) Validate() (uuid.UUID, error) {
	if len(r.Configs) != 1 {
		return uuid.Nil, errors.New("expected a single configuration to be created")
	}
	if msg := r.Configs[0].Message; msg != "" {
		return uuid.Nil, errors.New(msg)
	}
	id, err := uuid.Parse(r.Configs[0].ID)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (p CreateAzureConfigurationParams) ValidateQuery() (string, any, ValidateAzureConfigurationResult) {
	params := struct {
		CloudAccountID                 uuid.UUID `json:"cloudAccountId"`
		CreateAzureConfigurationParams `json:"azureExocomputeRegionConfig"`
	}{CloudAccountID: p.CloudAccountID, CreateAzureConfigurationParams: p}
	return validateAzureCloudAccountExocomputeConfigurationsQuery, params, ValidateAzureConfigurationResult{}
}

// ValidateAzureConfigurationResult holds the result of an Azure exocompute
// configuration validate operation.
type ValidateAzureConfigurationResult struct {
	ValidationInfo []struct {
		ErrorMessage                                           string `json:"errorMessage"`
		BlockedSecurityRules                                   bool   `json:"hasBlockedSecurityRules"`
		RestrictedAddressRangeOverlap                          bool   `json:"hasRestrictedAddressRangeOverlap"`
		ExocomputePrivateDnsZoneDoesNotExist                   bool   `json:"isAksCustomPrivateDnsZoneDoesNotExist"`
		ExocomputePrivateDnsZoneInDifferentSubscription        bool   `json:"isAksCustomPrivateDnsZoneInDifferentSubscription"`
		ExocomputePrivateDnsZoneInvalid                        bool   `json:"isAksCustomPrivateDnsZoneInvalid"`
		ExocomputePrivateDnsZoneNotLinkedToVnet                bool   `json:"isAksCustomPrivateDnsZoneNotLinkedToVnet"`
		ExocomputePrivateDnsZonePermissionsGroupNotEnabled     bool   `json:"isAksCustomPrivateDnsZonePermissionsGroupNotEnabled"`
		ClusterSubnetSizeTooSmall                              bool   `json:"isClusterSubnetSizeTooSmall"`
		PodAndClusterSubnetSame                                bool   `json:"isPodAndClusterSubnetSame"`
		PodAndClusterVnetDifferent                             bool   `json:"isPodAndClusterVnetDifferent" `
		PodCidrAndSubnetCidrOverlap                            bool   `json:"isPodCidrAndSubnetCidrOverlap"`
		PodCidrRangeTooSmall                                   bool   `json:"isPodCidrRangeTooSmall"`
		PodSubnetSizeTooSmall                                  bool   `json:"isPodSubnetSizeTooSmall"`
		SnapshotPrivateDnsZoneDoesNotExist                     bool   `json:"isPrivateDnsZoneDoesNotExist"`
		SnapshotPrivateDnsZoneInDifferentSubscription          bool   `json:"isPrivateDnsZoneInDifferentSubscription"`
		SnapshotPrivateDnsZoneInvalid                          bool   `json:"isPrivateDnsZoneInvalid"`
		SnapshotPrivateDnsZoneNotLinkedToVnet                  bool   `json:"isPrivateDnsZoneNotLinkedToVnet"`
		SubnetDelegated                                        bool   `json:"isSubnetDelegated"`
		UnsupportedCustomerManagedExocomputeConfigFieldPresent bool   `json:"isUnsupportedCustomerManagedExocomputeConfigFieldPresent"`
	} `json:"validationInfo"`
}

func (r ValidateAzureConfigurationResult) Validate() error {
	if len(r.ValidationInfo) != 1 {
		return errors.New("expected a single validation result")
	}

	var strs []string
	if r.ValidationInfo[0].ErrorMessage != "" {
		strs = append(strs, r.ValidationInfo[0].ErrorMessage)
	}
	if r.ValidationInfo[0].RestrictedAddressRangeOverlap {
		strs = append(strs, "subnet address range overlaps with restricted address ranges")
	}
	if r.ValidationInfo[0].ExocomputePrivateDnsZoneDoesNotExist {
		strs = append(strs, "exocompute private dns zone does not exist")
	}
	if r.ValidationInfo[0].ExocomputePrivateDnsZoneInDifferentSubscription {
		strs = append(strs, "exocompute private dns zone is in a different subscription than the exocompute vnet")
	}
	if r.ValidationInfo[0].ExocomputePrivateDnsZoneInvalid {
		strs = append(strs, "exocompute private dns zone is invalid: name must be [subzone.]privatelink.<region>.azmk8s.io")
	}
	if r.ValidationInfo[0].ExocomputePrivateDnsZoneNotLinkedToVnet {
		strs = append(strs, "exocompute private dns zone is not linked to the exocompute vnet")
	}
	if r.ValidationInfo[0].ExocomputePrivateDnsZonePermissionsGroupNotEnabled {
		strs = append(strs, "exocompute private dns zone permissions group is not enabled")
	}
	if r.ValidationInfo[0].SnapshotPrivateDnsZoneDoesNotExist {
		strs = append(strs, "snapshot access private dns zone does not exist")
	}
	if r.ValidationInfo[0].SnapshotPrivateDnsZoneInDifferentSubscription {
		strs = append(strs, "snapshot access private dns zone is in a different subscription than the exocompute vnet")
	}
	if r.ValidationInfo[0].SnapshotPrivateDnsZoneInvalid {
		strs = append(strs, "snapshot access private dns zone is invalid: name must be privatelink.blob.core.windows.net")
	}
	if r.ValidationInfo[0].SnapshotPrivateDnsZoneNotLinkedToVnet {
		strs = append(strs, "snapshot access private dns zone is not linked to the exocompute vnet")
	}
	if r.ValidationInfo[0].SubnetDelegated {
		strs = append(strs, "subnet is delegated")
	}
	if r.ValidationInfo[0].UnsupportedCustomerManagedExocomputeConfigFieldPresent {
		strs = append(strs, "unsupported fields for customer managed exocompute")
	}
	if r.ValidationInfo[0].BlockedSecurityRules {
		strs = append(strs, "network security group has blocking security rules")
	}
	if r.ValidationInfo[0].ClusterSubnetSizeTooSmall {
		strs = append(strs, "cluster subnet size is too small")
	}
	if r.ValidationInfo[0].PodAndClusterSubnetSame {
		strs = append(strs, "pod and cluster subnets must be different")
	}
	if r.ValidationInfo[0].PodAndClusterVnetDifferent {
		strs = append(strs, "pod and cluster vnet must be the same")
	}
	if r.ValidationInfo[0].PodCidrAndSubnetCidrOverlap {
		strs = append(strs, "pod cidr range overlaps with cluster subnet cidr range")
	}
	if r.ValidationInfo[0].PodCidrRangeTooSmall {
		strs = append(strs, "pod cidr range is too small")
	}
	if r.ValidationInfo[0].PodSubnetSizeTooSmall {
		strs = append(strs, "pod subnet size is too small")
	}

	switch len(strs) {
	case 0:
		return nil
	case 1:
		return errors.New(strs[0])
	default:
		var str strings.Builder
		for i, s := range strs {
			str.WriteString(fmt.Sprintf("; (%d) %s", i+1, s))
		}

		return fmt.Errorf("multiple configuration errors: %s", str.String()[1:])
	}
}

// DeleteAzureConfigurationParams holds the parameters for an Azure exocompute
// configuration delete operation.
type DeleteAzureConfigurationParams struct {
	ConfigID uuid.UUID `json:"cloudAccountId"`
}

func (p DeleteAzureConfigurationParams) DeleteQuery() (string, any, DeleteAzureConfigurationResult) {
	return deleteAzureCloudAccountExocomputeConfigurationsQuery, p, DeleteAzureConfigurationResult{}
}

// DeleteAzureConfigurationResult holds the result of an Azure exocompute
// configuration delete operation.
type DeleteAzureConfigurationResult struct {
	FailIDs    []uuid.UUID `json:"deletionFailedIds"`
	SuccessIDs []uuid.UUID `json:"deletionSuccessIds"`
}

func (r DeleteAzureConfigurationResult) Validate() error {
	if n := len(r.FailIDs); n > 0 {
		return fmt.Errorf("expected no delete failures: %d", n)
	}
	if len(r.SuccessIDs) != 1 {
		return errors.New("expected a single configuration to be deleted")
	}
	return nil
}

// AzureClusterAccess represents the different Azure exocompute cluster access
// types.
type AzureClusterAccess string

const (
	AzureClusterAccessUnspecified AzureClusterAccess = "AKS_CLUSTER_ACCESS_TYPE_UNSPECIFIED"
	AzureClusterAccessPrivate     AzureClusterAccess = "AKS_CLUSTER_ACCESS_TYPE_PRIVATE"
	AzureClusterAccessPublic      AzureClusterAccess = "AKS_CLUSTER_ACCESS_TYPE_PUBLIC"
)

// AzureClusterTier represents the different Azure exocompute cluster tiers.
type AzureClusterTier string

const (
	AzureClusterTierUnspecified AzureClusterTier = "AKS_CLUSTER_TIER_UNSPECIFIED"
	AzureClusterTierFree        AzureClusterTier = "AKS_CLUSTER_TIER_FREE"
	AzureClusterTierStandard    AzureClusterTier = "AKS_CLUSTER_TIER_STANDARD"
)

// AzureClusterNodeCount represents the different Azure exocompute cluster
// sizes.
type AzureClusterNodeCount string

const (
	AzureClusterNodeCountUnspecified AzureClusterNodeCount = "AKS_NODE_COUNT_BUCKET_UNSPECIFIED"
	AzureClusterNodeCountSmall       AzureClusterNodeCount = "AKS_NODE_COUNT_BUCKET_SMALL"
	AzureClusterNodeCountMedium      AzureClusterNodeCount = "AKS_NODE_COUNT_BUCKET_MEDIUM"
	AzureClusterNodeCountLarge       AzureClusterNodeCount = "AKS_NODE_COUNT_BUCKET_LARGE"
	AzureClusterNodeCountXLarge      AzureClusterNodeCount = "AKS_NODE_COUNT_BUCKET_XLARGE"
)

// AzureOptionalConfig holds an Azure exocompute optional configuration. Note,
// AdditionalWhitelistIPs requires that the WhitelistRubrikIPs field is set to
// true.
type AzureOptionalConfig struct {
	AdditionalWhitelistIPs     []string              `json:"additionalWhitelistIps,omitempty"`
	ClusterAccess              AzureClusterAccess    `json:"aksClusterAccessType"`
	ClusterTier                AzureClusterTier      `json:"aksClusterTier"`
	DiskEncryptionAtHost       bool                  `json:"diskEncryptionAtHost,omitempty"`
	ExocomputePrivateDnsZoneID string                `json:"aksCustomPrivateDnsZoneId,omitempty"`
	NodeCount                  AzureClusterNodeCount `json:"aksNodeCountBucket"`
	NodeRGPrefix               string                `json:"aksNodeRgPrefix,omitempty"`
	SnapshotPrivateDnsZoneId   string                `json:"privateDnsZoneId,omitempty"`
	UserDefinedRouting         bool                  `json:"enableUserDefinedRouting,omitempty"`
	WhitelistRubrikIPs         bool                  `json:"shouldWhitelistRubrikIps,omitempty"`
}
