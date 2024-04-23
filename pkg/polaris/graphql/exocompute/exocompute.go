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
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
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
	logParams(gql.Log(), query, params)

	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return nil, requestError(query, err)
	}
	logResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result []Result `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, unmarshalError(query, err)
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
	logParams(gql.Log(), query, queryParams)

	buf, err := gql.Request(ctx, query, queryParams)
	if err != nil {
		return uuid.Nil, requestError(query, err)
	}
	logResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result Result `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, unmarshalError(query, err)
	}
	id, err := payload.Data.Result.Validate()
	if err != nil {
		return uuid.Nil, responseError(query, err)
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
	logParams(gql.Log(), query, queryParams)

	buf, err := gql.Request(ctx, query, queryParams)
	if err != nil {
		return uuid.Nil, requestError(query, err)
	}
	logResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result Result `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, unmarshalError(query, err)
	}
	id, err := payload.Data.Result.Validate()
	if err != nil {
		return uuid.Nil, responseError(query, err)
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
	logParams(gql.Log(), query, params)

	buf, err := gql.Request(ctx, query, params)
	if err != nil {
		return requestError(query, err)
	}
	logResponse(gql.Log(), query, buf)

	var payload struct {
		Data struct {
			Result Result `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return unmarshalError(query, err)
	}
	id, err := payload.Data.Result.Validate()
	if err != nil {
		return responseError(query, err)
	}
	if id != configID {
		return responseError(query, errors.New("deleted config ID does not match requested"))
	}

	return nil
}

// MapCloudAccount maps exocompute for the application cloud account to the
// host cloud account.
func MapCloudAccount(ctx context.Context, gql *graphql.Client, appCloudAccountID, hostCloudAccountID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	params := struct {
		AppIDs []uuid.UUID `json:"cloudAccountIds"`
		HostID uuid.UUID   `json:"exocomputeCloudAccountId"`
	}{AppIDs: []uuid.UUID{appCloudAccountID}, HostID: hostCloudAccountID}
	logParams(gql.Log(), mapCloudAccountExocomputeAccountQuery, params)

	buf, err := gql.Request(ctx, mapCloudAccountExocomputeAccountQuery, params)
	if err != nil {
		return requestError(mapCloudAccountExocomputeAccountQuery, err)
	}
	logResponse(gql.Log(), mapCloudAccountExocomputeAccountQuery, buf)

	var payload struct {
		Data struct {
			Result struct {
				Success bool `json:"isSuccess"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return unmarshalError(mapCloudAccountExocomputeAccountQuery, err)
	}
	if !payload.Data.Result.Success {
		return responseError(mapCloudAccountExocomputeAccountQuery, errors.New("failed to map cloud account"))
	}

	return nil
}

// UnmapCloudAccount unmaps exocompute for the application cloud account.
func UnmapCloudAccount(ctx context.Context, gql *graphql.Client, appID uuid.UUID) error {
	gql.Log().Print(log.Trace)

	params := struct {
		AppIDs []uuid.UUID `json:"cloudAccountIds"`
	}{AppIDs: []uuid.UUID{appID}}
	logParams(gql.Log(), mapCloudAccountExocomputeAccountQuery, params)

	buf, err := gql.Request(ctx, unmapCloudAccountExocomputeAccountQuery, params)
	if err != nil {
		return requestError(unmapCloudAccountExocomputeAccountQuery, err)
	}
	logResponse(gql.Log(), unmapCloudAccountExocomputeAccountQuery, buf)

	var payload struct {
		Data struct {
			Result struct {
				Success bool `json:"isSuccess"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return unmarshalError(unmapCloudAccountExocomputeAccountQuery, err)
	}
	if !payload.Data.Result.Success {
		return responseError(unmapCloudAccountExocomputeAccountQuery, errors.New("failed to unmap cloud account"))
	}

	return nil
}

func logParams(logger log.Logger, query string, params any) {
	buf, err := json.Marshal(params)
	if err != nil {
		buf = []byte(fmt.Sprintf("marshaling of params failed: %s", err))
	}
	logger.Printf(log.Debug, "%s params: %s", graphql.QueryName(query), string(buf))
}

func logResponse(logger log.Logger, query string, response []byte) {
	logger.Printf(log.Debug, "%s response: %s", query, string(response))
}

func requestError(query string, err error) error {
	return fmt.Errorf("failed to request %s: %w", graphql.QueryName(query), err)
}

func unmarshalError(query string, err error) error {
	return fmt.Errorf("failed to unmarshal %s: %s", graphql.QueryName(query), err)
}

func responseError(query string, err error) error {
	return fmt.Errorf("%s response is an error: %s", graphql.QueryName(query), err)
}
