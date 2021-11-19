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

package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Subnet represents an AWS subnet.
type Subnet struct {
	ID               string `json:"subnetId"`
	AvailabilityZone string `json:"availabilityZone"`
}

// ExocomputeConfig represents a single exocompute config.
type ExocomputeConfig struct {
	ID      uuid.UUID `json:"configUuid"`
	Region  Region    `json:"region"`
	VPCID   string    `json:"vpcId"`
	Subnet1 Subnet    `json:"subnet1"`
	Subnet2 Subnet    `json:"subnet2"`
	Message string    `json:"message"`

	// When true Polaris manages the security groups.
	IsPolarisManaged bool `json:"areSecurityGroupsPolarisManaged"`

	// Security group ids of cluster control plane and worker node. Only needs
	// to be specified if IsPolarisManaged is false.
	ClusterSecurityGroupID string `json:"clusterSecurityGroupId"`
	NodeSecurityGroupID    string `json:"nodeSecurityGroupId"`
}

// ExocomputeConfigsForAccount holds all exocompute configs for a specific
// account.
type ExocomputeConfigsForAccount struct {
	Account         CloudAccount       `json:"awsCloudAccount"`
	Configs         []ExocomputeConfig `json:"configs"`
	EligibleRegions []string           `json:"exocomputeEligibleRegions"`
	Feature         Feature            `json:"featureDetail"`
}

// ExocomputeConfigs returns all exocompute configs matching the specified
// filter. The filter can be used to search for account name or account id.
func (a API) ExocomputeConfigs(ctx context.Context, filter string) ([]ExocomputeConfigsForAccount, error) {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, allAwsExocomputeConfigsQuery, struct {
		Filter string `json:"awsNativeAccountIdOrNamePrefix"`
	}{Filter: filter})
	if err != nil {
		return nil, fmt.Errorf("failed to request ExocomputeConfigs: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "allAwsExocomputeConfigs(%q): %s", filter, string(buf))

	var payload struct {
		Data struct {
			Result []ExocomputeConfigsForAccount `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ExocomputeConfigs: %v", err)
	}

	return payload.Data.Result, nil
}

// ExocomputeConfigCreate represents an exocompute config to be created by
// Polaris.
type ExocomputeConfigCreate struct {
	Region  Region   `json:"region"`
	VPCID   string   `json:"vpcId"`
	Subnets []Subnet `json:"subnets"`

	// When true Polaris will manage the security groups.
	IsPolarisManaged bool `json:"isPolarisManaged"`

	// Security group ids of cluster control plane and worker node. Only needs
	// to be specified if IsPolarisManaged is false.
	ClusterSecurityGroupId string `json:"clusterSecurityGroupId"`
	NodeSecurityGroupId    string `json:"nodeSecurityGroupId"`
}

// CreateExocomputeConfig creates a new exocompute config for the account with
// the specified Polaris cloud account id. Returns the created exocompute config
func (a API) CreateExocomputeConfig(ctx context.Context, id uuid.UUID, config ExocomputeConfigCreate) (ExocomputeConfig, error) {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, createAwsExocomputeConfigsQuery, struct {
		ID      uuid.UUID                `json:"cloudAccountId"`
		Configs []ExocomputeConfigCreate `json:"configs"`
	}{ID: id, Configs: []ExocomputeConfigCreate{config}})
	if err != nil {
		return ExocomputeConfig{}, fmt.Errorf("failed to request CreateExocomputeConfig: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "createAwsExocomputeConfigs(%q, %v): %s", id, config, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				Configs []ExocomputeConfig `json:"configs"`
			} `json:"createAwsExocomputeConfigs"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return ExocomputeConfig{}, fmt.Errorf("failed to unmarshal CreateExocomputeConfig: %v", err)
	}
	if len(payload.Data.Query.Configs) != 1 {
		return ExocomputeConfig{}, errors.New("expected a single result")
	}
	if payload.Data.Query.Configs[0].Message != "" {
		return ExocomputeConfig{}, errors.New(payload.Data.Query.Configs[0].Message)
	}

	return payload.Data.Query.Configs[0], nil
}

// DeleteExocomputeConfig deletes the exocompute config with the specified
// Polaris exocompute config id.
func (a API) DeleteExocomputeConfig(ctx context.Context, id uuid.UUID) error {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, deleteAwsExocomputeConfigsQuery, struct {
		IDs []uuid.UUID `json:"configIdsToBeDeleted"`
	}{IDs: []uuid.UUID{id}})
	if err != nil {
		return fmt.Errorf("failed to request DeleteExocomputeConfig: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "deleteAwsExocomputeConfigs(%q): %s", id, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				Status []struct {
					ID      uuid.UUID `json:"exocomputeConfigId"`
					Success bool      `json:"success"`
				} `json:"deletionStatus"`
			} `json:"deleteAwsExocomputeConfigs"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal DeleteExocomputeConfig: %v", err)
	}
	if len(payload.Data.Query.Status) != 1 {
		return errors.New("expected a single result")
	}
	if !payload.Data.Query.Status[0].Success {
		return errors.New("delete exocompute config failed")
	}

	return nil
}

// StartExocomputeDisableJob starts a task chain job to disables the Exocompute
// feature for the accout with the specified Polaris native account id. Returns
// the Polaris task chain id.
func (a API) StartExocomputeDisableJob(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	a.GQL.Log().Print(log.Trace)

	buf, err := a.GQL.Request(ctx, startAwsExocomputeDisableJobQuery, struct {
		ID uuid.UUID `json:"cloudAccountId"`
	}{ID: id})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to request StartExocomputeDisableJob: %v", err)
	}

	a.GQL.Log().Printf(log.Debug, "startAwsExocomputeDisableJob(%q): %s", id, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				Error string    `json:"error"`
				JobID uuid.UUID `json:"jobId"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal StartExocomputeDisableJob: %v", err)
	}
	if payload.Data.Result.Error != "" {
		return uuid.Nil, errors.New(payload.Data.Result.Error)
	}

	return payload.Data.Result.JobID, nil
}
