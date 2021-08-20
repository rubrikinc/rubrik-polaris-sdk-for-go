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
	SubnetID string    `json:"subnetNativeId"`
	Message  string    `json:"message"`

	// When true Polaris will manage the security groups.
	IsPolarisManaged bool `json:"isPolarisManaged"`
}

// ExocomputeConfigsForAccount holds all exocompute configs for a specific
// account.
type ExocomputeConfigsForAccount struct {
	Account         CloudAccount       `json:"azureCloudAccount"`
	Configs         []ExocomputeConfig `json:"configs"`
	EligibleRegions []string           `json:"exocomputeEligibleRegions"`
	Feature         Feature            `json:"featureDetails"`
}

// ExocomputeConfigs returns all exocompute configs matching the specified
// filter. The filter can be used to search for account name or account id.
func (a API) ExocomputeConfigs(ctx context.Context, filter string) ([]ExocomputeConfigsForAccount, error) {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/azure.ExocomputeConfigs")

	buf, err := a.GQL.Request(ctx, allAzureExocomputeConfigsInAccountQuery, struct {
		Filter string `json:"azureExocomputeSearchQuery"`
	}{Filter: filter})
	if err != nil {
		return nil, err
	}

	a.GQL.Log().Printf(log.Debug, "allAzureExocomputeConfigsInAccount(%q): %s", filter, string(buf))

	var payload struct {
		Data struct {
			Result []ExocomputeConfigsForAccount `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, err
	}

	return payload.Data.Result, nil
}

// ExocomputeConfigCreate represents an exocompute config to be created by
// Polaris.
type ExocomputeConfigCreate struct {
	Region   Region `json:"region"`
	SubnetID string `json:"subnetNativeId"`

	// When true Polaris will manage the security groups.
	IsPolarisManaged bool `json:"isPolarisManaged"`
}

// AddCloudAccountExocomputeConfigurations creates a new exocompute config for
// the account with the specified Polaris cloud account id. Returns the created
// exocompute config
func (a API) AddCloudAccountExocomputeConfigurations(ctx context.Context, id uuid.UUID, config ExocomputeConfigCreate) (ExocomputeConfig, error) {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/azure.AddCloudAccountExocomputeConfigurations")

	buf, err := a.GQL.Request(ctx, addAzureCloudAccountExocomputeConfigurationsQuery, struct {
		ID      uuid.UUID                `json:"cloudAccountId"`
		Configs []ExocomputeConfigCreate `json:"azureExocomputeRegionConfigs"`
	}{ID: id, Configs: []ExocomputeConfigCreate{config}})
	if err != nil {
		return ExocomputeConfig{}, err
	}

	a.GQL.Log().Printf(log.Debug, "addAzureCloudAccountExocomputeConfigurations(%q, %v): %s", id, config, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				Configs []ExocomputeConfig `json:"configs"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return ExocomputeConfig{}, err
	}
	if len(payload.Data.Result.Configs) != 1 {
		return ExocomputeConfig{}, errors.New("polaris: createAwsExocomputeConfigs: no result")
	}

	return payload.Data.Result.Configs[0], nil
}

// DeleteCloudAccountExocomputeConfigurations deletes the exocompute config
// with the specified Polaris exocompute config id.
func (a API) DeleteCloudAccountExocomputeConfigurations(ctx context.Context, id uuid.UUID) error {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/azure.DeleteCloudAccountExocomputeConfigurations")

	buf, err := a.GQL.Request(ctx, deleteAzureCloudAccountExocomputeConfigurationsQuery, struct {
		IDs []uuid.UUID `json:"cloudAccountIds"`
	}{IDs: []uuid.UUID{id}})
	if err != nil {
		return err
	}

	a.GQL.Log().Printf(log.Debug, "deleteAzureCloudAccountExocomputeConfigurations(%q): %s", id, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				FailIDs    []uuid.UUID `json:"deletionFailedIds"`
				SuccessIDs []uuid.UUID `json:"deletionSuccessIds"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return err
	}
	if ids := payload.Data.Result.SuccessIDs; len(ids) == 1 && ids[0] == id {
		return nil
	}

	return fmt.Errorf("polaris: azureExocomputeConfigsDelete: failed to delete %s", id)
}
