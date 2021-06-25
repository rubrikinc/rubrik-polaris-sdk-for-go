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

package azure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// ExocomputeConfig represents a single exocompute config.
type ExocomputeConfig struct {
	ID       uuid.UUID `json:"configUuid"`
	Region   Region    `json:"region"`
	SubnetID string    `json:"subnets"`
	Message  string    `json:"message"`

	// When true Polaris will manage the security groups.
	IsPolarisManaged bool `json:"isPolarisManaged"`
}

// ExocomputeConfigs holds all exocompute configs for a specific account.
type ExocomputeConfigSelector struct {
	Account         CloudAccount       `json:"azureCloudAccount"`
	Configs         []ExocomputeConfig `json:"configs"`
	EligibleRegions []string           `json:"exocomputeEligibleRegions"`
	Feature         Feature            `json:"featureDetails"`
}

// ExocomputeConfigs returns all exocompute configs matching the specified
// filter. The filter can be used to search for account name or account id.
func (a API) ExocomputeConfigs(ctx context.Context, filter string) ([]ExocomputeConfigSelector, error) {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/azure.ExocomputeConfigs")

	buf, err := a.GQL.Request(ctx, azureExocomputeConfigsQuery, struct {
		Filter string `json:"azureExocomputeSearchQueryArg"`
	}{Filter: filter})
	if err != nil {
		return nil, err
	}

	a.GQL.Log().Printf(log.Debug, "azureExocomputeConfigs(%q): %s", filter, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				Selectors []ExocomputeConfigSelector `json:"configs"`
			} `json:"azureExocomputeConfigs"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, err
	}

	return payload.Data.Query.Selectors, nil
}

// ExocomputeConfigCreate represents an exocompute config to be created by
// Polaris.
type ExocomputeConfigCreate struct {
	Region   Region
	SubnetID string

	// When true Polaris will manage the security groups.
	IsPolarisManaged bool
}

// ExocomputeAdd creates a new exocompute config for the account with the
// specified Polaris cloud account id. Returns the created exocompute config
func (a API) ExocomputeAdd(ctx context.Context, id uuid.UUID, config ExocomputeConfigCreate) (ExocomputeConfig, error) {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/azure.ExocomputeAdd")

	buf, err := a.GQL.Request(ctx, azureExocomputeAddQuery, struct {
		ID      uuid.UUID                `json:"cloudAccountUuid"`
		Configs []ExocomputeConfigCreate `json:"azureExocomputeAddRequests"`
	}{ID: id, Configs: []ExocomputeConfigCreate{config}})
	if err != nil {
		return ExocomputeConfig{}, err
	}

	a.GQL.Log().Printf(log.Debug, "azureExocomputeAdd(%q, %v): %s", id, config, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				Configs []ExocomputeConfig `json:"configs"`
			} `json:"azureExocomputeAdd"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return ExocomputeConfig{}, err
	}
	if len(payload.Data.Query.Configs) != 1 {
		return ExocomputeConfig{}, errors.New("polaris: createAwsExocomputeConfigs: no result")
	}

	return payload.Data.Query.Configs[0], nil
}

// ExocomputeConfigsDelete deletes the exocompute config with the specified
// Polaris exocompute config id.
func (a API) ExocomputeConfigsDelete(ctx context.Context, id uuid.UUID) error {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/azure.ExocomputeConfigsDelete")

	buf, err := a.GQL.Request(ctx, azureExocomputeConfigsDeleteQuery, struct {
		IDs []uuid.UUID `json:"azureExocomputeConfigIdsArg"`
	}{IDs: []uuid.UUID{id}})
	if err != nil {
		return err
	}

	a.GQL.Log().Printf(log.Debug, "azureExocomputeConfigsDelete(%q): %s", id, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				FailIDs    []uuid.UUID `json:"deletionFailedIds"`
				SuccessIDs []uuid.UUID `json:"deletionSuccessIds"`
			} `json:"azureExocomputeConfigsDelete"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return err
	}
	if ids := payload.Data.Query.SuccessIDs; len(ids) == 1 && ids[0] == id {
		return nil
	}

	return fmt.Errorf("polaris: azureExocomputeConfigsDelete: failed to delete %s", id)
}
