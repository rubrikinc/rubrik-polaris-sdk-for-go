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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp"
)

// Rubrik Cloud Vault, RCV, is Rubrik hosted archival storage. Unlike the other
// cloud native archival locations, the cloud account and the storage container
// are owned by Rubrik, so no cloud account ID is passed when creating an RCV
// archival location.
//
// The cloud in the type names below is the destination cloud, the cloud hosting
// the vault, and not the cloud the archived workloads originate from.
//
// The GCP API is the smallest of the three: it takes only a name, region, tier
// and location template, with no redundancy or encryption key options.

// RCVGCPTargetMapping holds the result of an RCV GCP target mapping list
// operation.
type RCVGCPTargetMapping struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	GroupType        string    `json:"groupType"`
	TargetType       string    `json:"targetType"`
	ConnectionStatus struct {
		Status string `json:"status"`
	} `json:"connectionStatus"`
	TargetTemplate struct {
		Region              gcp.RCSRegionEnum `json:"rcvGcpRegion"`
		Tier                string            `json:"rcvGcpTier"`
		EncryptionType      string            `json:"rcvGcpEncryptionType"`
		LocTemplate         string            `json:"cloudNativeLocTemplateType"`
		TemplateLocationID  uuid.UUID         `json:"templateLocationId"`
		SourceWorkloadCloud string            `json:"sourceWorkloadCloud"`
	}
}

func (RCVGCPTargetMapping) ListQuery(filters []ListTargetMappingFilter) (string, any) {
	extraFilters := []ListTargetMappingFilter{{
		Field: "ARCHIVAL_GROUP_TYPE",
		Text:  "CLOUD_NATIVE_ARCHIVAL_GROUP",
	}, {
		Field: "ARCHIVAL_LOCATION_TYPE",
		Text:  "RCV_GCP",
	}}

	return allTargetMappingsQuery, struct {
		Filters []ListTargetMappingFilter `json:"filter,omitempty"`
	}{Filters: append(filters, extraFilters...)}
}

// CreateRCVGCPStorageSettingParams holds the parameters for an RCV GCP storage
// setting create operation. RCV GCP storage settings are also referred to as
// RCV GCP target mappings.
type CreateRCVGCPStorageSettingParams struct {
	Name        string            `json:"name"`
	Region      gcp.RCSRegionEnum `json:"region"`
	Tier        string            `json:"tier"`
	LocTemplate string            `json:"locTemplateType"`
}

// CreateRCVGCPStorageSettingResult holds the result of an RCV GCP storage
// setting create operation.
type CreateRCVGCPStorageSettingResult struct {
	TargetMapping struct {
		ID string `json:"id"`
	} `json:"targetMapping"`
}

func (CreateRCVGCPStorageSettingResult) CreateQuery(createParams CreateRCVGCPStorageSettingParams) (string, any) {
	return createCloudNativeRcvGcpStorageSettingQuery, createParams
}

func (r CreateRCVGCPStorageSettingResult) Validate() (uuid.UUID, error) {
	return uuid.Parse(r.TargetMapping.ID)
}

// UpdateRCVGCPStorageSettingParams holds the parameters for an RCV GCP storage
// setting update operation. The RSC API only supports updating the name.
type UpdateRCVGCPStorageSettingParams struct {
	Name string `json:"name"`
}

// UpdateRCVGCPStorageSettingResult holds the result of an RCV GCP storage
// setting update operation.
type UpdateRCVGCPStorageSettingResult CreateRCVGCPStorageSettingResult

func (r UpdateRCVGCPStorageSettingResult) UpdateQuery(targetMappingID uuid.UUID, updateParams UpdateRCVGCPStorageSettingParams) (string, any) {
	return updateCloudNativeRcvGcpStorageSettingQuery, struct {
		ID uuid.UUID `json:"id"`
		UpdateRCVGCPStorageSettingParams
	}{ID: targetMappingID, UpdateRCVGCPStorageSettingParams: updateParams}
}
