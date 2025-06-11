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

package exocompute

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// CloudAccountMapping represents a mapping between an application cloud account
// and a host cloud account.
type CloudAccountMapping struct {
	AppCloudAccountID  uuid.UUID `json:"applicationCloudAccountId"`
	HostCloudAccountID uuid.UUID `json:"exocomputeCloudAccountId"`
}

// ListCloudAccountMappings return all cloud account mappings for the specified
// cloud vendor.
func ListCloudAccountMappings(ctx context.Context, gql *graphql.Client, cloudVendor core.CloudVendor) ([]CloudAccountMapping, error) {
	gql.Log().Print(log.Trace)

	query := allCloudAccountExocomputeMappingsQuery
	buf, err := gql.Request(ctx, query, struct {
		CloudVendor core.CloudVendor `json:"cloudVendor"`
	}{CloudVendor: cloudVendor})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}

	var payload struct {
		Data struct {
			Result []CloudAccountMapping `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// MapCloudAccountsParams holds the parameters for an application cloud accounts
// map operation.
type MapCloudAccountsParams[R MapCloudAccountsResult] interface {
	MapQuery() (string, any, R)
}

// MapCloudAccountsResult holds the result of an application cloud accounts map
// operation.
type MapCloudAccountsResult interface {
	MapAWSCloudAccountsResult | MapAzureCloudAccountsResult
	Validate() error
}

// MapCloudAccounts maps the application cloud accounts to the host cloud
// account.
func MapCloudAccounts[P MapCloudAccountsParams[R], R MapCloudAccountsResult](ctx context.Context, gql *graphql.Client, params P) error {
	gql.Log().Print(log.Trace)

	query, queryParams, _ := params.MapQuery()
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
	if err := payload.Data.Result.Validate(); err != nil {
		return graphql.ResponseError(query, err)
	}

	return nil
}

// UnmapCloudAccountsParams holds the parameters for an application cloud
// accounts unmap operation.
type UnmapCloudAccountsParams[R UnmapCloudAccountsResult] interface {
	UnmapQuery() (string, any, R)
}

// UnmapCloudAccountsResult holds the result of an application cloud accounts
// unmap operation.
type UnmapCloudAccountsResult interface {
	UnmapAWSCloudAccountsResult | UnmapAzureCloudAccountsResult
	Validate() error
}

// UnmapCloudAccounts unmaps the application cloud accounts.
func UnmapCloudAccounts[P UnmapCloudAccountsParams[R], R UnmapCloudAccountsResult](ctx context.Context, gql *graphql.Client, params P) error {
	gql.Log().Print(log.Trace)

	query, queryParams, _ := params.UnmapQuery()
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
	if err := payload.Data.Result.Validate(); err != nil {
		return graphql.ResponseError(query, err)
	}

	return nil
}
