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

// ListAccountFilter is used to filter data center archival cloud accounts.
// Common field values are:
//
//   - NAME - The prefix or name of the cloud account.
//
//   - ACCOUNT_PROVIDER_TYPE - The type of the cloud account, e.g.
//     CLOUD_ACCOUNT_AWS or CLOUD_ACCOUNT_AZURE.
//
//   - IS_KEY_BASED - Whether the cloud account uses key-based authentication
//     or not.
type ListAccountFilter struct {
	Field string `json:"field"`
	Text  string `json:"text"`
}

// ListAccountResult holds the result of a data center cloud account list
// operation.
type ListAccountResult interface {
	ListQuery(filters []ListAccountFilter) (string, any)
	Validate() bool
}

// ListAccounts returns all data center cloud accounts matching the specified
// filters.
func ListAccounts[R ListAccountResult](ctx context.Context, gql *graphql.Client, filters []ListAccountFilter) ([]R, error) {
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

	var accounts []R
	for _, account := range payload.Data.Result {
		if account.Validate() {
			accounts = append(accounts, account)
		}
	}

	return accounts, nil
}

// CreateCloudAccountParams represents the valid type parameters for a cloud
// account create operation.
type CreateCloudAccountParams interface {
	CreateAWSCloudAccountParams | CreateAzureCloudAccountParams
}

// CreateCloudAccountResult represents the result of a cloud account create
// operation.
type CreateCloudAccountResult[P CreateCloudAccountParams] interface {
	CreateQuery(createParams P) (string, any)
	Validate() (uuid.UUID, error)
}

// CreateCloudAccount creates a data center cloud account for use with data
// center archival locations.
func CreateCloudAccount[R CreateCloudAccountResult[P], P CreateCloudAccountParams](ctx context.Context, gql *graphql.Client, createParams P) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	var result R
	query, queryParams := result.CreateQuery(createParams)
	buf, err := gql.RequestWithoutLogging(ctx, query, queryParams)
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

// UpdateCloudAccountParams represents the valid type parameters for a cloud
// account update operation.
type UpdateCloudAccountParams interface {
	UpdateAWSCloudAccountParams | UpdateAzureCloudAccountParams
}

// UpdateCloudAccountResult represents the result of a cloud account update
// operation.
type UpdateCloudAccountResult[P UpdateCloudAccountParams] interface {
	UpdateQuery(cloudAccountID uuid.UUID, updateParams P) (string, any)
}

// UpdateCloudAccount updates a data center cloud account.
func UpdateCloudAccount[R UpdateCloudAccountResult[P], P UpdateCloudAccountParams](ctx context.Context, gql *graphql.Client, cloudAccountID uuid.UUID, updateParams P) error {
	gql.Log().Print(log.Trace)

	var result R
	query, queryParams := result.UpdateQuery(cloudAccountID, updateParams)
	buf, err := gql.RequestWithoutLogging(ctx, query, queryParams)
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

	return nil
}

// DeleteCloudAccountResult represents the result of a cloud account delete
// operation.
type DeleteCloudAccountResult interface {
	DeleteQuery(cloudAccountID uuid.UUID) (string, any)
}

// DeleteCloudAccount deletes a data center cloud account.
func DeleteCloudAccount[R DeleteCloudAccountResult](ctx context.Context, gql *graphql.Client, cloudAccountID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	var result R
	query, queryParams := result.DeleteQuery(cloudAccountID)
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

	return nil
}
