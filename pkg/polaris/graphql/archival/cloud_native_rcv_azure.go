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

package archival

import (
	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core/secret"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
)

// Rubrik Cloud Vault, RCV, is Rubrik hosted archival storage. Unlike the other
// cloud native archival locations, the cloud account and the storage container
// are owned by Rubrik, so no cloud account ID is passed when creating an RCV
// archival location.
//
// The cloud in the type names below is the destination cloud, the cloud hosting
// the vault, and not the cloud the archived workloads originate from.
//
// Azure is the only RCV cloud whose API accepts a source workload cloud,
// letting an Azure hosted vault archive AWS or GCP workloads. When it is left
// out, RSC derives it from the destination cloud.
//
// Note that the RSC API is inconsistent in its naming: the Azure archival
// location type is RCS_AZURE and its target template is
// RcsAzureTargetTemplate, whereas the AWS and GCP equivalents use RCV.

// RCVImmutabilitySettings represents the immutability settings of an RCV
// archival location.
type RCVImmutabilitySettings struct {
	LockDurationDays int64 `json:"lockDurationDays"`
}

// AzureKeyVaultKey identifies an Azure Key Vault key used for client-side
// encryption of an RCV Azure archival location.
type AzureKeyVaultKey struct {
	KeyVaultID uuid.UUID `json:"kmsKeyVaultId"`
	KeyName    string    `json:"keyName"`
	KeyVersion string    `json:"keyVersion"`
}

// RCVAzureTargetMapping holds the result of an RCV Azure target mapping list
// operation.
type RCVAzureTargetMapping struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	GroupType        string    `json:"groupType"`
	TargetType       string    `json:"targetType"`
	ConnectionStatus struct {
		Status string `json:"status"`
	} `json:"connectionStatus"`
	TargetTemplate struct {
		Region               azure.RCSRegionEnum     `json:"rcvAzureRegion"`
		Tier                 RCVTier                 `json:"rcvAzureTier"`
		Redundancy           RCVRedundancy           `json:"rcvAzureRedundancy"`
		EncryptionType       TargetEncryptionType    `json:"rcvAzureEncryptionType"`
		ImmutabilitySettings RCVImmutabilitySettings `json:"immutabilitySettings"`
		LocTemplate          LocTemplateType         `json:"cloudNativeLocTemplateType"`
		TemplateLocationID   uuid.UUID               `json:"templateLocationId"`
		SourceWorkloadCloud  SourceWorkloadCloud     `json:"sourceWorkloadCloud"`
	}
}

func (RCVAzureTargetMapping) ListQuery(filters []ListTargetMappingFilter) (string, any) {
	extraFilters := []ListTargetMappingFilter{{
		Field: "ARCHIVAL_GROUP_TYPE",
		Text:  "CLOUD_NATIVE_ARCHIVAL_GROUP",
	}, {
		Field: "ARCHIVAL_LOCATION_TYPE",
		Text:  "RCS_AZURE",
	}}

	return allTargetMappingsQuery, struct {
		Filters []ListTargetMappingFilter `json:"filter,omitempty"`
	}{Filters: append(filters, extraFilters...)}
}

// CreateRCVAzureStorageSettingParams holds the parameters for an RCV Azure
// storage setting create operation. The tier, source workload cloud,
// redundancy, RSA key and Key Vault key are optional. When the RSA key is left
// out, RCV creates and manages the encryption key.
//
// The RSA key must be an unencrypted PEM-encoded RSA private key, in either
// PKCS#1 or PKCS#8 form, with a size of 2048, 3072 or 4096 bits. DER and base64
// encoded keys are rejected by RSC.
//
// Note that the RSC API requires both the Azure region and the RCV region. The
// high level API derives both from a single region to keep them from drifting
// apart. RCV Azure storage settings are also referred to as RCV Azure target
// mappings.
type CreateRCVAzureStorageSettingParams struct {
	Name                string              `json:"name"`
	Region              azure.RegionEnum    `json:"region"`
	RCVRegion           azure.RCSRegionEnum `json:"rcvRegion"`
	Tier                RCVTier             `json:"rcvTier,omitempty"`
	LocTemplate         LocTemplateType     `json:"locTemplateType"`
	SourceWorkloadCloud SourceWorkloadCloud `json:"sourceWorkloadCloud,omitempty"`
	Redundancy          RCVRedundancy       `json:"redundancy,omitempty"`
	RSAKey              secret.String       `json:"rsaKey,omitempty"`
	KeyVaultKey         *AzureKeyVaultKey   `json:"azureKeyVaultKey,omitempty"`
}

// CreateRCVAzureStorageSettingResult holds the result of an RCV Azure storage
// setting create operation.
type CreateRCVAzureStorageSettingResult struct {
	TargetMapping struct {
		ID string `json:"id"`
	} `json:"targetMapping"`
}

func (CreateRCVAzureStorageSettingResult) CreateQuery(createParams CreateRCVAzureStorageSettingParams) (string, any) {
	return createCloudNativeRcvAzureStorageSettingQuery, createParams
}

func (r CreateRCVAzureStorageSettingResult) Validate() (uuid.UUID, error) {
	return uuid.Parse(r.TargetMapping.ID)
}

// UpdateRCVAzureStorageSettingParams holds the parameters for an RCV Azure
// storage setting update operation. The RSC API only supports updating the
// name, the tier and the redundancy. Child vaults are only updated when
// UpdateChildVaults is true.
type UpdateRCVAzureStorageSettingParams struct {
	Name              string        `json:"name"`
	Tier              RCVTier       `json:"tier,omitempty"`
	Redundancy        RCVRedundancy `json:"redundancy,omitempty"`
	UpdateChildVaults *bool         `json:"updateChildVaults,omitempty"`
}

// UpdateRCVAzureStorageSettingResult holds the result of an RCV Azure storage
// setting update operation.
type UpdateRCVAzureStorageSettingResult CreateRCVAzureStorageSettingResult

func (r UpdateRCVAzureStorageSettingResult) UpdateQuery(targetMappingID uuid.UUID, updateParams UpdateRCVAzureStorageSettingParams) (string, any) {
	return updateCloudNativeRcvAzureStorageSettingQuery, struct {
		ID uuid.UUID `json:"id"`
		UpdateRCVAzureStorageSettingParams
	}{ID: targetMappingID, UpdateRCVAzureStorageSettingParams: updateParams}
}
