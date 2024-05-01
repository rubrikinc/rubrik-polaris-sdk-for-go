//go:generate go run ../queries_gen.go exocompute

// Copyright 2021 Rubrik, Inc.
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
	"errors"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// ListResult represents the result of a list operation.
type ListResult interface {
	ListQuery(filter string) (string, any)
}

// ListConfigurations return all exocompute configurations matching the
// specified filter.
func ListConfigurations[Result ListResult](ctx context.Context, gql *graphql.Client, filter string) ([]Result, error) {
	gql.Log().Print(log.Trace)

	var result Result
	query, params := result.ListQuery(filter)
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result []Result `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result, nil
}

// CreateParams represents the valid type parameters for a CreateConfiguration
// call.
type CreateParams interface {
	aws.ExoCreateParams | azure.ExoCreateParams
}

// CreateResult represents the result of a create operation.
type CreateResult[Params CreateParams] interface {
	CreateQuery(cloudAccountID uuid.UUID, createParams Params) (string, any)
	Validate() (uuid.UUID, error)
}

// CreateConfiguration creates a new exocompute configuration in the account
// with the specified RSC cloud account id. Returns the ID of the configuration.
func CreateConfiguration[Result CreateResult[Params], Params CreateParams](ctx context.Context, gql *graphql.Client, cloudAccountID uuid.UUID, createParams Params) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	var result Result
	query, queryParams := result.CreateQuery(cloudAccountID, createParams)
	buf, err := gql.Request(ctx, query, queryParams)
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result Result `json:"result"`
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

// UpdateParams represents the valid type parameters for an UpdateConfiguration
// call.
type UpdateParams interface {
	aws.ExoUpdateParams
}

type UpdateResult[Params UpdateParams] interface {
	UpdateQuery(cloudAccountID uuid.UUID, updateParams Params) (string, any)
	Validate() (uuid.UUID, error)
}

// UpdateConfiguration updates an existing exocompute configuration in the
// account with the specified RSC cloud account id.
func UpdateConfiguration[Result UpdateResult[Params], Params UpdateParams](ctx context.Context, gql *graphql.Client, cloudAccountID uuid.UUID, updateParams Params) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	var result Result
	query, queryParams := result.UpdateQuery(cloudAccountID, updateParams)
	buf, err := gql.Request(ctx, query, queryParams)
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result Result `json:"result"`
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

// DeleteResult represents the result of a delete operation.
type DeleteResult interface {
	DeleteQuery(configID uuid.UUID) (string, any)
	Validate() (uuid.UUID, error)
}

// DeleteConfiguration deletes the exocompute configuration with the specified
// configuration ID.
func DeleteConfiguration[Result DeleteResult](ctx context.Context, gql *graphql.Client, configID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	var result Result
	query, params := result.DeleteQuery(configID)
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result Result `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return graphql.UnmarshalError(query, err)
	}
	id, err := payload.Data.Result.Validate()
	if err != nil {
		return graphql.ResponseError(query, err)
	}
	if id != configID {
		return graphql.ResponseError(query, errors.New("deleted config ID does not match requested"))
	}

	return nil
}

// MapResult represents the result of a map operation.
type MapResult interface {
	MapQuery(hostCloudAccountID, appCloudAccountIDs uuid.UUID) (string, any)
	Validate() error
}

// MapCloudAccount maps the application cloud account to the host cloud account.
func MapCloudAccount[Result MapResult](ctx context.Context, gql *graphql.Client, hostCloudAccountID, appCloudAccountID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	var result Result
	query, params := result.MapQuery(hostCloudAccountID, appCloudAccountID)
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result Result `json:"result"`
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

// UnmapResult represents the result of an unmap operation.
type UnmapResult interface {
	UnmapQuery(appCloudAccountIDs uuid.UUID) (string, any)
	Validate() error
}

// UnmapCloudAccount unmaps the application cloud account.
func UnmapCloudAccount[Result UnmapResult](ctx context.Context, gql *graphql.Client, appCloudAccountID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	var result Result
	query, params := result.UnmapQuery(appCloudAccountID)
	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result Result `json:"result"`
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

// CloudAccountMapping represents a mapping between an exocompute application
// cloud account and a host cloud account.
type CloudAccountMapping struct {
	AppCloudAccountID  uuid.UUID `json:"applicationCloudAccountId"`
	HostCloudAccountID uuid.UUID `json:"exocomputeCloudAccountId"`
}

// AllCloudAccountMappings returns all exocompute cloud account mappings for
// the specified cloud vendor. Note that only AWS and Azure are supported by
// RSC.
func AllCloudAccountMappings(ctx context.Context, gql *graphql.Client, cloudVendor core.CloudVendor) ([]CloudAccountMapping, error) {
	gql.Log().Print(log.Trace)

	query := allCloudAccountExocomputeMappingsQuery
	buf, err := gql.Request(ctx, query, struct {
		CloudVendor core.CloudVendor `json:"cloudVendor"`
	}{CloudVendor: cloudVendor})
	if err != nil {
		return nil, graphql.RequestError(query, err)
	}
	graphql.LogResponse(gql.Log(), query, buf)

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
