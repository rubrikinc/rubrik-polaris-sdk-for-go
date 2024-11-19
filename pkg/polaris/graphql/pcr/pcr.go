//go:generate go run ../queries_gen.go pcr

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

package pcr

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// GetRegistryParams holds the parameters for a private container registry get
// operation.
type GetRegistryParams[R GetRegistryResult] interface {
	GetQuery() (string, any, R)
}

// GetRegistryResult holds the result of a private container registry get
// operation.
type GetRegistryResult interface {
	AWSRegistry | AzureRegistry
}

// GetRegistry returns the private container registry details.
func GetRegistry[P GetRegistryParams[R], R GetRegistryResult](ctx context.Context, gql *graphql.Client, params P) (R, error) {
	gql.Log().Print(log.Trace)

	query, queryParams, defValue := params.GetQuery()
	buf, err := gql.Request(ctx, query, queryParams)
	if err != nil {
		return defValue, graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result R `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return defValue, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// SetRegistryParams holds the parameters for a private container registry set
// operation.
type SetRegistryParams[R SetRegistryResult] interface {
	SetQuery() (string, any, R)
}

// SetRegistryResult holds the result of a private container registry set
// operation.
type SetRegistryResult interface {
	SetAWSRegistryResult | SetAzureRegistryResult
}

// SetRegistry sets the private container registry.
func SetRegistry[P SetRegistryParams[R], R SetRegistryResult](ctx context.Context, gql *graphql.Client, params P) error {
	gql.Log().Print(log.Trace)

	query, queryParams, _ := params.SetQuery()
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

// RemoveRegistry removes the private container registry for the specified cloud
// account ID.
func RemoveRegistry(ctx context.Context, gql *graphql.Client, cloudAccountID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	query := deregisterPrivateContainerRegistryQuery
	buf, err := gql.Request(ctx, query, struct {
		ID uuid.UUID `json:"exocomputeAccountId"`
	}{ID: cloudAccountID})
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
