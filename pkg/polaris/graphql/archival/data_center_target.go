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

// GetTargetResult holds the result of a target get operation.
type GetTargetResult interface {
	GetQuery(targetID uuid.UUID) (string, any)
	Validate() error
}

// GetTarget return the target with the specified target ID.
func GetTarget[R GetTargetResult](ctx context.Context, gql *graphql.Client, targetID uuid.UUID) (R, error) {
	gql.Log().Print(log.Trace)

	var result R
	query, params := result.GetQuery(targetID)
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return result, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result R `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return result, graphql.UnmarshalError(query, err)
	}
	if err := payload.Data.Result.Validate(); err != nil {
		return result, graphql.ResponseError(query, err)
	}

	return payload.Data.Result, nil
}

// ListTargetFilter is used to filter data center archival targets. Common field
// values are:
//
//   - NAME - The name of the target mapping. It can also be used to search for
//     a prefix of the name.
//
//   - CLUSTER_ID - The ID of a data center cluster.
//
//   - LOCATION_ID - The ID of a data center archival location.
type ListTargetFilter struct {
	Field    string   `json:"field"`
	Text     string   `json:"text,omitempty"`
	TextList []string `json:"textList,omitempty"`
}

// ListTargetResult holds the result of a target list operation.
type ListTargetResult interface {
	ListQuery(cursor string, filters []ListTargetFilter) (string, any)
	Validate() error
}

// ListTargets return all targets matching the specified filters.
func ListTargets[R ListTargetResult](ctx context.Context, gql *graphql.Client, filters []ListTargetFilter) ([]R, error) {
	gql.Log().Print(log.Trace)

	var cursor string
	var nodes []R
	for {
		var result R
		query, params := result.ListQuery(cursor, filters)
		buf, err := gql.Request(ctx, query, params)
		if err != nil {
			return nil, graphql.RequestError(query, err)
		}

		var payload struct {
			Data struct {
				Result struct {
					Nodes    []R `json:"nodes"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"result"`
			} `json:"data"`
		}
		if err := json.Unmarshal(buf, &payload); err != nil {
			return nil, graphql.UnmarshalError(query, err)
		}
		for _, node := range payload.Data.Result.Nodes {
			if err := node.Validate(); err == nil {
				nodes = append(nodes, node)
			}
		}
		if !payload.Data.Result.PageInfo.HasNextPage {
			break
		}
		cursor = payload.Data.Result.PageInfo.EndCursor
	}

	return nodes, nil
}

// CreateTargetParams represents the valid type parameters for a target create
// operation.
type CreateTargetParams interface {
	CreateAWSTargetParams
}

// CreateTargetResult represents the result of a target create operation.
type CreateTargetResult[P CreateTargetParams] interface {
	CreateQuery(createParams P) (string, any)
	Validate() (uuid.UUID, error)
}

// CreateTarget creates a data center archival location.
func CreateTarget[R CreateTargetResult[P], P CreateTargetParams](ctx context.Context, gql *graphql.Client, createParams P) (uuid.UUID, error) {
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

// UpdateTargetParams represents the valid type parameters for a target update
// operation.
type UpdateTargetParams interface {
	UpdateAWSTargetParams
}

// UpdateTargetResult represents the result of a target update operation.
type UpdateTargetResult[P UpdateTargetParams] interface {
	UpdateQuery(targetID uuid.UUID, updateParams P) (string, any)
}

// UpdateTarget updates the data center archival location with the specified
// target ID.
func UpdateTarget[R UpdateTargetResult[P], P UpdateTargetParams](ctx context.Context, gql *graphql.Client, targetID uuid.UUID, updateParams P) error {
	gql.Log().Print(log.Trace)

	var result R
	query, queryParams := result.UpdateQuery(targetID, updateParams)
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

// DisableTarget disables the target with the specified target ID. Data center
// archival locations are also referred to as targets.
func DisableTarget(ctx context.Context, gql *graphql.Client, targetID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	query := disableTargetQuery
	buf, err := gql.Request(ctx, query, struct {
		ID uuid.UUID `json:"id"`
	}{ID: targetID})
	if err != nil {
		return graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result struct {
				ID uuid.UUID `json:"locationId"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}

	return nil
}

// DeleteTarget deletes the target with the specified target ID. Data center
// archival locations are also referred to as targets.
func DeleteTarget(ctx context.Context, gql *graphql.Client, targetID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	query := deleteTargetQuery
	buf, err := gql.Request(ctx, query, struct {
		ID uuid.UUID `json:"id"`
	}{ID: targetID})
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
