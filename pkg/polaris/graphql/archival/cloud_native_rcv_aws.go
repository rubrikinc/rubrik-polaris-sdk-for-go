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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
)

// Rubrik Cloud Vault, RCV, is Rubrik hosted archival storage. Unlike the other
// cloud native archival locations, the cloud account and the storage container
// are owned by Rubrik, so no cloud account ID is passed when creating an RCV
// archival location.
//
// The cloud in the type names below is the destination cloud, the cloud hosting
// the vault, and not the cloud the archived workloads originate from.
//
// Note that the RSC API calls the AWS archival location type RCV_AWS and its
// target template RcvAwsTargetTemplate, whereas the Azure equivalents use RCS
// rather than RCV.

// AWSKMSKey identifies an AWS KMS key used for client-side encryption of an RCV
// AWS archival location.
type AWSKMSKey struct {
	KeyManagerID uuid.UUID `json:"keyManagerId"`
	KeyID        string    `json:"keyId"`
}

// RCVAWSTargetMapping holds the result of an RCV AWS target mapping list
// operation.
type RCVAWSTargetMapping struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	GroupType        string    `json:"groupType"`
	TargetType       string    `json:"targetType"`
	ConnectionStatus struct {
		Status string `json:"status"`
	} `json:"connectionStatus"`
	TargetTemplate struct {
		Region              aws.RCSRegionEnum    `json:"rcvAwsRegion"`
		Tier                RCVTier              `json:"rcvAwsTier"`
		Redundancy          RCVRedundancy        `json:"rcvAwsRedundancy"`
		EncryptionType      TargetEncryptionType `json:"rcvAwsEncryptionType"`
		LocTemplate         LocTemplateType      `json:"cloudNativeLocTemplateType"`
		TemplateLocationID  uuid.UUID            `json:"templateLocationId"`
		SourceWorkloadCloud SourceWorkloadCloud  `json:"sourceWorkloadCloud"`
	}
}

func (RCVAWSTargetMapping) ListQuery(filters []ListTargetMappingFilter) (string, any) {
	extraFilters := []ListTargetMappingFilter{{
		Field: "ARCHIVAL_GROUP_TYPE",
		Text:  "CLOUD_NATIVE_ARCHIVAL_GROUP",
	}, {
		Field: "ARCHIVAL_LOCATION_TYPE",
		Text:  "RCV_AWS",
	}}

	return allTargetMappingsQuery, struct {
		Filters []ListTargetMappingFilter `json:"filter,omitempty"`
	}{Filters: append(filters, extraFilters...)}
}

// CreateRCVAWSStorageSettingParams holds the parameters for an RCV AWS storage
// setting create operation. The tier, redundancy, encryption key and KMS key
// are optional. RCV AWS storage settings are also referred to as RCV AWS target
// mappings.
//
// EncryptionKeyInDER is misnamed in the RSC API: it takes a base64 encoded
// 32 byte AES-256 key, not a DER encoded key. RSC rejects anything which does
// not base64 decode to exactly 32 non-zero bytes. It must be set together with
// EnableSSEWithCPK, and neither may be combined with KMSKey.
type CreateRCVAWSStorageSettingParams struct {
	Name               string            `json:"name"`
	Region             aws.RCSRegionEnum `json:"region"`
	Tier               RCVTier           `json:"tier,omitempty"`
	LocTemplate        LocTemplateType   `json:"locTemplateType"`
	Redundancy         RCVRedundancy     `json:"redundancy,omitempty"`
	EnableSSEWithCPK   *bool             `json:"enableSseWithCpk,omitempty"`
	EncryptionKeyInDER secret.String     `json:"encryptionKeyInDer,omitempty"`
	KMSKey             *AWSKMSKey        `json:"awsKmsKey,omitempty"`
}

// CreateRCVAWSStorageSettingResult holds the result of an RCV AWS storage
// setting create operation.
type CreateRCVAWSStorageSettingResult struct {
	TargetMapping struct {
		ID string `json:"id"`
	} `json:"targetMapping"`
}

func (CreateRCVAWSStorageSettingResult) CreateQuery(createParams CreateRCVAWSStorageSettingParams) (string, any) {
	return createCloudNativeRcvAwsStorageSettingQuery, createParams
}

func (r CreateRCVAWSStorageSettingResult) Validate() (uuid.UUID, error) {
	return uuid.Parse(r.TargetMapping.ID)
}

// UpdateRCVAWSStorageSettingParams holds the parameters for an RCV AWS storage
// setting update operation. The RSC API only supports updating the name and the
// tier.
type UpdateRCVAWSStorageSettingParams struct {
	Name string  `json:"name"`
	Tier RCVTier `json:"tier,omitempty"`
}

// UpdateRCVAWSStorageSettingResult holds the result of an RCV AWS storage
// setting update operation.
type UpdateRCVAWSStorageSettingResult CreateRCVAWSStorageSettingResult

func (r UpdateRCVAWSStorageSettingResult) UpdateQuery(targetMappingID uuid.UUID, updateParams UpdateRCVAWSStorageSettingParams) (string, any) {
	return updateCloudNativeRcvAwsStorageSettingQuery, struct {
		ID uuid.UUID `json:"id"`
		UpdateRCVAWSStorageSettingParams
	}{ID: targetMappingID, UpdateRCVAWSStorageSettingParams: updateParams}
}
