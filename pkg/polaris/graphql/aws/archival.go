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

package aws

import "github.com/google/uuid"

// TargetMappingFilter is used to filter AWS target mappings. Common field
// values are:
//
//   - NAME - The name of the target mapping. It can also be used to search for
//     a prefix of the name.
//
//   - ARCHIVAL_GROUP_TYPE - The type of the archival group, e.g.,
//     CLOUD_NATIVE_ARCHIVAL_GROUP.
//
//   - CLOUD_ACCOUNT_ID - The ID of an RSC cloud account.
//
//   - ARCHIVAL_GROUP_ID - The ID of an archival group. Also known as target
//     mapping ID.
type TargetMappingFilter struct {
	Field    string   `json:"field"`
	Text     string   `json:"text,omitempty"`
	TestList []string `json:"testList,omitempty"`
}

// TargetMapping represents an AWS cloud archival location.
type TargetMapping struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	GroupType        string    `json:"groupType"`
	TargetType       string    `json:"targetType"`
	ConnectionStatus struct {
		Status string `json:"status"`
	} `json:"connectionStatus"`
	TargetTemplate struct {
		CloudAccount struct {
			ID uuid.UUID `json:"id"`
		} `json:"cloudAccount"`
		BucketPrefix string `json:"bucketPrefix"`
		StorageClass string `json:"storageClass"`
		Region       Region `json:"region"`
		KMSMasterKey string `json:"kmsMasterKeyId"`
		LocTemplate  string `json:"cloudNativeLocTemplateType"`
		BucketTags   []Tag  `json:"bucketTags"`
	}
}

func (TargetMapping) ListQuery(filters []TargetMappingFilter) (string, any) {
	return allTargetMappingsQuery, append(filters, TargetMappingFilter{
		Field: "ARCHIVAL_LOCATION_TYPE",
		Text:  "AWS",
	})
}

// StorageSettingCreateParams represents the parameters required to create an
// AWS storage setting.
type StorageSettingCreateParams struct {
	Name         string `json:"name"`
	BucketPrefix string `json:"bucketPrefix"`
	StorageClass string `json:"storageClass"`
	Region       Region `json:"region,omitempty"`
	KmsMasterKey string `json:"kmsMasterKeyId"`
	LocTemplate  string `json:"locTemplateType"`
	BucketTags   *struct {
		TagList []Tag `json:"tagList"`
	} `json:"bucketTags,omitempty"`
}

// StorageSettingCreateResult represents the result of creating an AWS storage
// setting.
type StorageSettingCreateResult struct {
	TargetMapping struct {
		ID uuid.UUID `json:"id"`
	} `json:"targetMapping"`
}

func (StorageSettingCreateResult) CreateQuery(cloudAccountID uuid.UUID, createParams StorageSettingCreateParams) (string, any) {
	return createCloudNativeAwsStorageSettingQuery, struct {
		CloudAccountID uuid.UUID `json:"cloudAccountId"`
		StorageSettingCreateParams
	}{CloudAccountID: cloudAccountID, StorageSettingCreateParams: createParams}
}

func (r StorageSettingCreateResult) Validate() (uuid.UUID, error) {
	return r.TargetMapping.ID, nil
}

// StorageSettingUpdateParams represents the parameters required to update an
// AWS storage setting.
type StorageSettingUpdateParams struct {
	Name         string `json:"name,omitempty"`
	StorageClass string `json:"storageClass,omitempty"`
	KmsMasterKey string `json:"kmsMasterKeyId,omitempty"`
}

// StorageSettingUpdateResult represents the result of updating an AWS storage
// setting.
type StorageSettingUpdateResult StorageSettingCreateResult

func (r StorageSettingUpdateResult) UpdateQuery(targetMappingID uuid.UUID, updateParams StorageSettingUpdateParams) (string, any) {
	return updateCloudNativeAwsStorageSettingQuery, struct {
		ID uuid.UUID `json:"id"`
		StorageSettingUpdateParams
	}{ID: targetMappingID, StorageSettingUpdateParams: updateParams}
}

func (r StorageSettingUpdateResult) Validate() (uuid.UUID, error) {
	return r.TargetMapping.ID, nil
}
