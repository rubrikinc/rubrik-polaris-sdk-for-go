//go:generate go run ../queries_gen.go archival

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
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// ListFilter filters target mappings.
type ListFilter interface {
	aws.TargetMappingFilter | azure.TargetMappingFilter
}

// ListResult represents the result of a list operation.
type ListResult[F ListFilter] interface {
	ListQuery(filters []F) (string, any)
}

// ListTargetMappings return all target mappings matching the specified filters.
// In RSC, cloud archival locations are also referred to as target mappings.
func ListTargetMappings[R ListResult[F], F ListFilter](ctx context.Context, gql *graphql.Client, filters []F) ([]R, error) {
	gql.Log().Print(log.Trace)

	var result R
	query, params := result.ListQuery(filters)
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

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

// CreateParams represents the valid type parameters for a create operation.
type CreateParams interface {
	aws.StorageSettingCreateParams | azure.StorageSettingCreateParams
}

// CreateResult represents the result of a create operation.
type CreateResult[P CreateParams] interface {
	CreateQuery(cloudAccountID uuid.UUID, createParams P) (string, any)
	Validate() (uuid.UUID, error)
}

// CreateCloudNativeStorageSetting creates a cloud native archival location for
// the specified cloud account.
func CreateCloudNativeStorageSetting[R CreateResult[P], P CreateParams](ctx context.Context, gql *graphql.Client, cloudAccountID uuid.UUID, createParams P) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	var result R
	query, queryParams := result.CreateQuery(cloudAccountID, createParams)
	buf, err := gql.Request(ctx, query, queryParams)
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

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

// UpdateParams represents the valid type parameters for an update operation.
type UpdateParams interface {
	aws.StorageSettingUpdateParams | azure.StorageSettingUpdateParams
}

// UpdateResult represents the result of an update operation.
type UpdateResult[P UpdateParams] interface {
	UpdateQuery(targetMappingID uuid.UUID, updateParams P) (string, any)
	Validate() (uuid.UUID, error)
}

// UpdateCloudNativeStorageSetting updates the cloud native archival location
// with the specified ID.
func UpdateCloudNativeStorageSetting[R UpdateResult[P], P UpdateParams](ctx context.Context, gql *graphql.Client, targetMappingID uuid.UUID, updateParams P) error {
	gql.Log().Print(log.Trace)

	var result R
	query, queryParams := result.UpdateQuery(targetMappingID, updateParams)
	buf, err := gql.Request(ctx, query, queryParams)
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result R `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	id, err := payload.Data.Result.Validate()
	if err != nil {
		return graphql.ResponseError(query, err)
	}
	if id != targetMappingID {
		return graphql.ResponseError(query, fmt.Errorf("response ID does not match request ID: %s != %s", id, targetMappingID))
	}

	return nil
}

// DeleteTargetMapping deletes the target mapping with the specified ID.
// In RSC, cloud archival locations are also referred to as target mappings.
func DeleteTargetMapping(ctx context.Context, gql *graphql.Client, targetMappingID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	query := deleteTargetMappingQuery
	buf, err := gql.Request(ctx, query, struct {
		ID uuid.UUID `json:"id"`
	}{ID: targetMappingID})
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

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
