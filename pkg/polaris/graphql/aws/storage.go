// Copyright 2023 Rubrik, Inc.
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

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Tag represents a key-value pair used to tag cloud resources.
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
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

// TargetMappingFilter is used to filter target mappings.
//
// Common field values are:
//   - NAME - The name of the target mapping. Can also be used to search for a
//     prefix of the name.
//   - ARCHIVAL_GROUP_TYPE - The type of the archival group, e.g.
//     CLOUD_NATIVE_ARCHIVAL_GROUP.
//   - CLOUD_ACCOUNT_ID - The ID of an RSC cloud account.
//   - ARCHIVAL_GROUP_ID - The ID of an archival group. Also known as target
//     mapping ID.
type TargetMappingFilter struct {
	Field string `json:"field"`
	Text  string `json:"text"`
}

// AllTargetMappings returns all AWS target mappings that match the specified
// filter. In RSC cloud archival locations are also referred to as target
// mappings.
func (a API) AllTargetMappings(ctx context.Context, filter []TargetMappingFilter) ([]TargetMapping, error) {
	a.log.Print(log.Trace)

	// Always filter for only AWS target mappings.
	filter = append(filter, TargetMappingFilter{
		Field: "ARCHIVAL_LOCATION_TYPE",
		Text:  "AWS",
	})
	buf, err := a.GQL.Request(ctx, allTargetMappingsQuery, struct {
		Filter []TargetMappingFilter `json:"filter,omitempty"`
	}{Filter: filter})
	if err != nil {
		return nil, fmt.Errorf("failed to request allTargetMappings: %w", err)
	}
	a.log.Printf(log.Debug, "allTargetMappings(%v): %s", filter, string(buf))

	var payload struct {
		Data struct {
			Result []TargetMapping `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allTargetMappings: %v", err)
	}

	return payload.Data.Result, nil
}

// DeleteTargetMapping deletes the target mapping with the specified ID. In RSC
// cloud archival locations are also referred to as target mappings.
func (a API) DeleteTargetMapping(ctx context.Context, id uuid.UUID) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, deleteTargetMappingQuery, struct {
		ID uuid.UUID `json:"id"`
	}{ID: id})
	if err != nil {
		return fmt.Errorf("failed to request deleteTargetMapping: %w", err)
	}
	a.log.Printf(log.Debug, "deleteTargetMapping(%q): %s", id, string(buf))

	var payload struct {
		Data struct {
			Result string `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal deleteTargetMapping: %v", err)
	}

	return nil
}

// CreateCloudNativeStorageSetting creates a cloud native archival location.
// The KMS master key can be either a key alias or a key ID. Region and bucket
// tags are optional.
//
// Common storage class values are:
//   - STANDARD
//   - STANDARD_IA
//   - ONEZONE_IA
//   - GLACIER_INSTANT_RETRIEVAL
//
// Common location template type values are:
//   - SOURCE_REGION
//   - SPECIFIC_REGION
func (a API) CreateCloudNativeStorageSetting(ctx context.Context, id uuid.UUID, name, bucketPrefix, storageClass string, region Region, kmsMasterKey, locTemplateType string, bucketTags []Tag) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	tags := &struct {
		TagList []Tag `json:"tagList"`
	}{TagList: bucketTags}
	if len(bucketTags) == 0 {
		tags = nil
	}

	buf, err := a.GQL.Request(ctx, createCloudNativeAwsStorageSettingQuery, struct {
		CloudAccountID uuid.UUID `json:"cloudAccountId"`
		Name           string    `json:"name"`
		BucketPrefix   string    `json:"bucketPrefix"`
		StorageClass   string    `json:"storageClass"`
		Region         Region    `json:"region,omitempty"`
		KmsMasterKey   string    `json:"kmsMasterKeyId"`
		LocTemplate    string    `json:"locTemplateType"`
		BucketTags     *struct {
			TagList []Tag `json:"tagList"`
		} `json:"bucketTags,omitempty"`
	}{CloudAccountID: id, Name: name, BucketPrefix: bucketPrefix, StorageClass: storageClass, Region: region, KmsMasterKey: kmsMasterKey, LocTemplate: locTemplateType, BucketTags: tags})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to request createCloudNativeAwsStorageSetting: %w", err)
	}
	a.log.Printf(log.Debug, "createCloudNativeAwsStorageSetting(%q, %q, %q, %q, %q, <REDACTED>, %q, %v): %s",
		id, name, bucketPrefix, storageClass, region, locTemplateType, bucketTags, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				TargetMapping struct {
					ID uuid.UUID `json:"id"`
				} `json:"targetMapping"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal createCloudNativeAwsStorageSetting: %v", err)
	}

	return payload.Data.Result.TargetMapping.ID, nil
}

// UpdateCloudNativeStorageSetting updates the cloud native archival location
// with the specified ID. The KMS master key can be either a key alias or a key
// ID. Note that not all properties can be updated, only the name, storage and
// KMS master key.
func (a API) UpdateCloudNativeStorageSetting(ctx context.Context, id uuid.UUID, name string, storageClass string, kmsMasterKey string) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, updateCloudNativeAwsStorageSettingQuery, struct {
		ID           uuid.UUID `json:"id"`
		Name         string    `json:"name,omitempty"`
		StorageClass string    `json:"storageClass,omitempty"`
		KmsMasterKey string    `json:"kmsMasterKeyId,omitempty"`
	}{ID: id, Name: name, StorageClass: storageClass, KmsMasterKey: kmsMasterKey})
	if err != nil {
		return fmt.Errorf("failed to request updateCloudNativeAwsStorageSetting: %w", err)
	}
	a.log.Printf(log.Debug, "updateCloudNativeAwsStorageSetting(%q, %q, %q <REDACTED>): %s",
		id, name, storageClass, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				TargetMapping struct {
					ID uuid.UUID `json:"id"`
				} `json:"targetMapping"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal createCloudNativeAwsStorageSetting: %v", err)
	}
	if id != payload.Data.Result.TargetMapping.ID {
		return errors.New("wrong id returned from updateCloudNativeAwsStorageSetting")
	}

	return nil
}
