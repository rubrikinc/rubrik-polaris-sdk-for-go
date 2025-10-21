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

package archival

import (
	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
)

// AzureTargetMapping holds the result of an Azure target mapping list
// operation. Note, the ContainerNamePrefix field is not the prefix but the
// full container name.
type AzureTargetMapping struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	GroupType        string    `json:"groupType"`
	TargetType       string    `json:"targetType"`
	ConnectionStatus struct {
		Status string `json:"status"`
	} `json:"connectionStatus"`
	TargetTemplate struct {
		CloudAccount struct {
			ID uuid.UUID `json:"cloudAccountId"`
		} `json:"cloudAccount"`
		ContainerNamePrefix  string `json:"containerNamePrefix"`
		StorageAccountName   string `json:"storageAccountName"`
		CloudNativeCompanion struct {
			LocTemplate          string             `json:"cloudNativeLocTemplateType"`
			Redundancy           string             `json:"redundancy"`
			StorageTier          string             `json:"storageTier"`
			NativeID             uuid.UUID          `json:"subscriptionNativeId"`
			StorageAccountRegion azure.RegionEnum   `json:"storageAccountRegion"`
			StorageAccountTags   []core.Tag         `json:"storageAccountTags"`
			CMKInfo              []AzureCustomerKey `json:"cmkInfo"`
		} `json:"cloudNativeCompanion"`
	}
}

func (AzureTargetMapping) ListQuery(filters []ListTargetMappingFilter) (string, any) {
	extraFilters := []ListTargetMappingFilter{{
		Field: "ARCHIVAL_GROUP_TYPE",
		Text:  "CLOUD_NATIVE_ARCHIVAL_GROUP",
	}, {
		Field: "ARCHIVAL_LOCATION_TYPE",
		Text:  "AZURE",
	}}

	return allTargetMappingsQuery, struct {
		Filters []ListTargetMappingFilter `json:"filters,omitempty"`
	}{Filters: append(filters, extraFilters...)}
}

// CreateAzureStorageSettingParams holds the parameters for an Azure storage
// setting create operation.
// The native ID, storage account region, the storage account tags, and the
// customer managed keys are optional. Note, the API ignores the ContainerName
// field and generates its own name.
// Azure storage settings are also referred to as Azure target mappings.
type CreateAzureStorageSettingParams struct {
	CloudAccountID       uuid.UUID          `json:"cloudAccountId"`
	LocTemplate          string             `json:"cloudNativeLocTemplateType"`
	ContainerName        string             `json:"containerName"`
	Name                 string             `json:"name"`
	Redundancy           string             `json:"redundancy"`
	StorageTier          string             `json:"storageTier"`
	NativeID             uuid.UUID          `json:"subscriptionNativeId"`
	StorageAccountName   string             `json:"storageAccountName"`
	StorageAccountRegion azure.RegionEnum   `json:"storageAccountRegion"`
	StorageAccountTags   *core.Tags         `json:"storageAccountTags,omitempty"`
	CMKInfo              []AzureCustomerKey `json:"cmkInfo,omitempty"`
}

// CreateAzureStorageSettingResult holds the result of an Azure storage setting
// create operation. Azure storage settings are also referred to as Azure target
// mappings.
type CreateAzureStorageSettingResult struct {
	TargetMapping struct {
		ID string `json:"id"`
	} `json:"targetMapping"`
}

func (CreateAzureStorageSettingResult) CreateQuery(createParams CreateAzureStorageSettingParams) (string, any) {
	return createCloudNativeAzureStorageSettingQuery, createParams
}

func (r CreateAzureStorageSettingResult) Validate() (uuid.UUID, error) {
	return uuid.Parse(r.TargetMapping.ID)
}

// UpdateAzureStorageSettingParams holds the parameters for an Azure storage
// setting update operation. Azure storage settings are also referred to as
// Azure target mappings.
type UpdateAzureStorageSettingParams struct {
	Name               string             `json:"name"`
	StorageTier        string             `json:"storageTier"`
	StorageAccountTags core.Tags          `json:"storageAccountTags"`
	CMKInfo            []AzureCustomerKey `json:"cmkInfo,omitempty"`
}

// UpdateAzureStorageSettingResult holds the result of an Azure storage setting
// update operation. Azure storage settings are also referred to as Azure target
// mappings.
type UpdateAzureStorageSettingResult CreateAzureStorageSettingResult

func (r UpdateAzureStorageSettingResult) UpdateQuery(targetMappingID uuid.UUID, updateParams UpdateAzureStorageSettingParams) (string, any) {
	return updateCloudNativeAzureStorageSettingQuery, struct {
		ID uuid.UUID `json:"id"`
		UpdateAzureStorageSettingParams
	}{ID: targetMappingID, UpdateAzureStorageSettingParams: updateParams}
}

// AzureCustomerKey represents a customer managed key required for encryption
// of Azure storage.
type AzureCustomerKey struct {
	Name      string           `json:"keyName"`
	VaultName string           `json:"keyVaultName"`
	Region    azure.RegionEnum `json:"region"`
}
