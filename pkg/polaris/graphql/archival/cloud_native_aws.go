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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
)

// AWSTargetMapping holds the result of an AWS target mapping list operation.
type AWSTargetMapping struct {
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
		BucketPrefix string         `json:"bucketPrefix"`
		StorageClass string         `json:"storageClass"`
		Region       aws.RegionEnum `json:"region"`
		KMSMasterKey string         `json:"kmsMasterKeyId"`
		LocTemplate  string         `json:"cloudNativeLocTemplateType"`
		BucketTags   []core.Tag     `json:"bucketTags"`
	}
}

func (AWSTargetMapping) ListQuery(filters []ListTargetMappingFilter) (string, any) {
	extraFilters := []ListTargetMappingFilter{{
		Field: "ARCHIVAL_GROUP_TYPE",
		Text:  "CLOUD_NATIVE_ARCHIVAL_GROUP",
	}, {
		Field: "ARCHIVAL_LOCATION_TYPE",
		Text:  "AWS",
	}}

	return allTargetMappingsQuery, struct {
		Filters []ListTargetMappingFilter `json:"filters,omitempty"`
	}{Filters: append(filters, extraFilters...)}
}

// CreateAWSStorageSettingParams holds the parameters for an AWS storage setting
// create operation.
// The KMS master key can be either a key alias or a key ID. Region, KMS master
// key and bucket tags are optional.
// AWS storage settings are also referred to as AWS target mappings.
type CreateAWSStorageSettingParams struct {
	CloudAccountID uuid.UUID      `json:"cloudAccountId"`
	Name           string         `json:"name"`
	BucketPrefix   string         `json:"bucketPrefix"`
	StorageClass   string         `json:"storageClass"`
	Region         aws.RegionEnum `json:"region"`
	KmsMasterKey   string         `json:"kmsMasterKeyId"`
	LocTemplate    string         `json:"locTemplateType"`
	BucketTags     *AWSTags       `json:"bucketTags,omitempty"`
}

// CreateAWSStorageSettingResult holds the result of an AWS storage setting
// create operation. AWS storage settings are also referred to as AWS target
// mappings.
type CreateAWSStorageSettingResult struct {
	TargetMapping struct {
		ID string `json:"id"`
	} `json:"targetMapping"`
}

func (CreateAWSStorageSettingResult) CreateQuery(createParams CreateAWSStorageSettingParams) (string, any) {
	return createCloudNativeAwsStorageSettingQuery, createParams
}

func (r CreateAWSStorageSettingResult) Validate() (uuid.UUID, error) {
	return uuid.Parse(r.TargetMapping.ID)
}

// UpdateAWSStorageSettingParams holds the parameters for an AWS storage setting
// update operation.
// The KMS master key can be either a key alias or a key ID. The bucket tags
// replace all existing tags, unless DeleteAllBucketTags is true, in which case
// all bucket tags are removed.
// AWS storage settings are also referred to as AWS target mappings.
type UpdateAWSStorageSettingParams struct {
	Name                string   `json:"name,omitempty"`
	StorageClass        string   `json:"storageClass,omitempty"`
	KmsMasterKey        string   `json:"kmsMasterKeyId,omitempty"`
	DeleteAllBucketTags bool     `json:"deleteAllBucketTags,omitempty"`
	BucketTags          *AWSTags `json:"bucketTags,omitempty"`
}

// UpdateAWSStorageSettingResult holds the result of an AWS storage setting
// update operation. AWS storage settings are also referred to as AWS target
// mappings.
type UpdateAWSStorageSettingResult CreateAWSStorageSettingResult

func (r UpdateAWSStorageSettingResult) UpdateQuery(targetMappingID uuid.UUID, updateParams UpdateAWSStorageSettingParams) (string, any) {
	return updateCloudNativeAwsStorageSettingQuery, struct {
		ID uuid.UUID `json:"id"`
		UpdateAWSStorageSettingParams
	}{ID: targetMappingID, UpdateAWSStorageSettingParams: updateParams}
}

// AWSTags represents a collection of AWS tags.
type AWSTags struct {
	TagList []core.Tag `json:"tagList"`
}
