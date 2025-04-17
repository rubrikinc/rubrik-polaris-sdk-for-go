//go:generate go run ../queries_gen.go exocompute

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
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// ListConfigurationsFilter holds the filter for an exocompute configuration
// list operation.
type ListConfigurationsFilter[R ListConfigurationsResult] interface {
	ListQuery() (string, any, R)
}

// ListConfigurationsResult holds the result of an exocompute configuration list
// operation.
type ListConfigurationsResult interface {
	AWSConfigurationsForCloudAccount | AzureConfigurationsForCloudAccount
}

// ListConfigurations return all exocompute configurations matching the
// specified filter.
func ListConfigurations[F ListConfigurationsFilter[R], R ListConfigurationsResult](ctx context.Context, gql *graphql.Client, filter F) ([]R, error) {
	gql.Log().Print(log.Trace)

	query, queryParams, _ := filter.ListQuery()
	buf, err := gql.Request(ctx, query, queryParams)
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

// CreateConfigurationParams holds the parameters for an exocompute
// configuration create operation.
type CreateConfigurationParams[R CreateConfigurationResult] interface {
	CreateQuery() (string, any, R)
}

// CreateConfigurationResult holds the result of an exocompute configuration
// create operation.
type CreateConfigurationResult interface {
	CreateAWSConfigurationResult | CreateAzureConfigurationResult
	Validate() (uuid.UUID, error)
}

// CreateConfiguration creates a new exocompute configuration. Returns the ID
// of the new configuration.
func CreateConfiguration[P CreateConfigurationParams[R], R CreateConfigurationResult](ctx context.Context, gql *graphql.Client, params P) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	query, queryParams, _ := params.CreateQuery()
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

// UpdateConfigurationParams holds the parameters for an exocompute
// configuration update operation.
type UpdateConfigurationParams[R UpdateConfigurationResult] interface {
	UpdateQuery() (string, any, R)
}

// UpdateConfigurationResult holds the result of an exocompute configuration
// update operation.
type UpdateConfigurationResult interface {
	UpdateAWSConfigurationResult
	Validate() (uuid.UUID, error)
}

// UpdateConfiguration updates an existing exocompute configuration. Returns the
// ID of the updated configuration. Note, the configuration ID might change with
// the update.
func UpdateConfiguration[P UpdateConfigurationParams[R], R UpdateConfigurationResult](ctx context.Context, gql *graphql.Client, params P) (uuid.UUID, error) {
	gql.Log().Print(log.Trace)

	query, queryParams, _ := params.UpdateQuery()
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
	configID, err := payload.Data.Result.Validate()
	if err != nil {
		return uuid.Nil, graphql.ResponseError(query, err)
	}

	return configID, nil
}

// DeleteConfigurationParams holds the parameters for an exocompute
// configuration delete operation.
type DeleteConfigurationParams[R DeleteConfigurationResult] interface {
	DeleteQuery() (string, any, R)
}

// DeleteConfigurationResult holds the result of an exocompute configuration
// delete operation.
type DeleteConfigurationResult interface {
	DeleteAWSConfigurationResult | DeleteAzureConfigurationResult
	Validate() error
}

// DeleteConfiguration deletes the exocompute configuration.
func DeleteConfiguration[P DeleteConfigurationParams[R], R DeleteConfigurationResult](ctx context.Context, gql *graphql.Client, params P) error {
	gql.Log().Print(log.Trace)

	query, queryParams, _ := params.DeleteQuery()
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
