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

package archival

import (
	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp"
)

// GCPTargetMapping holds the result of a GCP target mapping list operation.
type GCPTargetMapping struct {
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
		BucketPrefix string           `json:"bucketPrefix"`
		StorageClass string           `json:"gcpStorageClass"`
		Region       gcp.RegionEnum   `json:"gcpRegion"`
		CMKInfo      []GCPCustomerKey `json:"cmkInfo"`
		LocTemplate  string           `json:"cloudNativeLocTemplateType"`
		BucketLabels []core.Tag       `json:"labels"`
	}
}

func (GCPTargetMapping) ListQuery(filters []ListTargetMappingFilter) (string, any) {
	extraFilters := []ListTargetMappingFilter{{
		Field: "ARCHIVAL_GROUP_TYPE",
		Text:  "CLOUD_NATIVE_ARCHIVAL_GROUP",
	}, {
		Field: "ARCHIVAL_LOCATION_TYPE",
		Text:  "GCP",
	}}

	return allTargetMappingsQuery, struct {
		Filters []ListTargetMappingFilter `json:"filters,omitempty"`
	}{Filters: append(filters, extraFilters...)}
}

// CreateGCPStorageSettingParams holds the parameters for a GCP storage setting
// create operation.
// Region, customer managed keys and bucket labels are optional.
// GCP storage settings are also referred to as GCP target mappings.
type CreateGCPStorageSettingParams struct {
	CloudAccountID uuid.UUID        `json:"cloudAccountId"`
	Name           string           `json:"name"`
	BucketPrefix   string           `json:"bucketPrefix"`
	StorageClass   string           `json:"storageClass"`
	Region         gcp.RegionEnum   `json:"region,omitempty"`
	LocTemplate    string           `json:"locTemplateType"`
	BucketLabels   *core.Tags       `json:"bucketLabels,omitempty"`
	CMKInfo        []GCPCustomerKey `json:"cmkInfo,omitempty"`
}

// CreateGCPStorageSettingResult holds the result of a GCP storage setting
// create operation. GCP storage settings are also referred to as GCP target
// mappings.
type CreateGCPStorageSettingResult struct {
	TargetMapping struct {
		ID string `json:"id"`
	} `json:"targetMapping"`
}

func (CreateGCPStorageSettingResult) CreateQuery(createParams CreateGCPStorageSettingParams) (string, any) {
	return createCloudNativeGcpStorageSettingQuery, createParams
}

func (r CreateGCPStorageSettingResult) Validate() (uuid.UUID, error) {
	return uuid.Parse(r.TargetMapping.ID)
}

// UpdateGCPStorageSettingParams holds the parameters for a GCP storage setting
// update operation. GCP storage settings are also referred to a GCP target
// mappings.
type UpdateGCPStorageSettingParams struct {
	Name         string           `json:"name"`
	StorageClass string           `json:"storageClass"`
	BucketLabels *core.Tags       `json:"bucketLabels,omitempty"`
	CMKInfo      []GCPCustomerKey `json:"cmkInfo,omitempty"`
}

// UpdateGCPStorageSettingResult holds the result of a GCP storage setting
// update operation. GCP storage settings are also referred to as GCP target
// mappings.
type UpdateGCPStorageSettingResult CreateGCPStorageSettingResult

func (r UpdateGCPStorageSettingResult) UpdateQuery(targetMappingID uuid.UUID, updateParams UpdateGCPStorageSettingParams) (string, any) {
	return updateCloudNativeGcpStorageSettingQuery, struct {
		ID uuid.UUID `json:"id"`
		UpdateGCPStorageSettingParams
	}{ID: targetMappingID, UpdateGCPStorageSettingParams: updateParams}
}

// GCPCustomerKey represents a customer managed key required for encryption
// of GCP storage.
type GCPCustomerKey struct {
	Name     string         `json:"keyName"`
	RingName string         `json:"keyRingName"`
	Region   gcp.RegionEnum `json:"region"`
}
