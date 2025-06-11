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
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// ListTargetMappingFilter is used to filter cloud native archival target
// mappings. Common field values are:
//
//   - NAME - The prefix or name of the target mapping.
//
//   - ARCHIVAL_GROUP_TYPE - The type of the archival group, e.g.,
//     CLOUD_NATIVE_ARCHIVAL_GROUP.
//
//   - CLOUD_ACCOUNT_ID - The ID of a cloud native cloud account.
//
//   - ARCHIVAL_GROUP_ID - The ID of an archival group. Also known as target
//     mapping ID.
type ListTargetMappingFilter struct {
	Field    string   `json:"field"`
	Text     string   `json:"text"`
	TestList []string `json:"testList,omitempty"`
}

// ListTargetMappingResult holds the result of a target mapping list operation.
type ListTargetMappingResult interface {
	ListQuery(filters []ListTargetMappingFilter) (string, any)
}

// ListTargetMappings return all target mappings matching the specified filters.
// Cloud archival locations are also referred to as target mappings.
func ListTargetMappings[R ListTargetMappingResult](ctx context.Context, gql *graphql.Client, filters []ListTargetMappingFilter) ([]R, error) {
	gql.Log().Print(log.Trace)

	var result R
	query, params := result.ListQuery(filters)
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []R `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// CreateStorageSettingParams represents the valid type parameters for a storage
// setting create operation.
type CreateStorageSettingParams interface {
	CreateAWSStorageSettingParams | CreateAzureStorageSettingParams
}

// CreateStorageSettingResult represents the result of a storge setting create
// operation.
type CreateStorageSettingResult[P CreateStorageSettingParams] interface {
	CreateQuery(createParams P) (string, any)
	Validate() (uuid.UUID, error)
}

// CreateCloudNativeStorageSetting creates a cloud native archival location.
func CreateCloudNativeStorageSetting[R CreateStorageSettingResult[P], P CreateStorageSettingParams](ctx context.Context, gql *graphql.Client, createParams P) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	var result R
	query, queryParams := result.CreateQuery(createParams)
	buf, err := gql.Request(ctx, query, queryParams)
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result R `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}
	id, err := payload.Data.Result.Validate()
	if err != nil {
		return uuid.Nil, graphql.ResponseError(query, err)
	}

	return id, nil
}

// UpdateStorageSettingParams represents the valid type parameters for a storage
// setting update operation.
type UpdateStorageSettingParams interface {
	UpdateAWSStorageSettingParams | UpdateAzureStorageSettingParams
}

// UpdateStorageSettingResult represents the result of a storage setting update
// operation.
type UpdateStorageSettingResult[P UpdateStorageSettingParams] interface {
	UpdateQuery(targetMappingID uuid.UUID, updateParams P) (string, any)
}

// UpdateCloudNativeStorageSetting updates the cloud native archival location
// with the specified target mapping ID.
func UpdateCloudNativeStorageSetting[R UpdateStorageSettingResult[P], P UpdateStorageSettingParams](ctx context.Context, gql *graphql.Client, targetMappingID uuid.UUID, updateParams P) error {
	gql.Log().Print(log.Trace)

	var result R
	query, queryParams := result.UpdateQuery(targetMappingID, updateParams)
	buf, err := gql.Request(ctx, query, queryParams)
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result R `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	return nil
}

// DeleteTargetMapping deletes the target mapping with the specified target
// mapping ID. Cloud archival locations are also referred to as target mappings.
func DeleteTargetMapping(ctx context.Context, gql *graphql.Client, targetMappingID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	query := deleteTargetMappingQuery
	buf, err := gql.Request(ctx, query, struct {
		ID uuid.UUID `json:"id"`
	}{ID: targetMappingID})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result string `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	return nil
}
