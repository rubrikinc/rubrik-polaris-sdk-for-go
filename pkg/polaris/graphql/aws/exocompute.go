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
	ID      string `json:"configUuid"`
	Region  Region `json:"region"`
	VPCID   string `json:"vpcId"`
	Subnet1 Subnet `json:"subnet1"`
	Subnet2 Subnet `json:"subnet2"`
	Message string `json:"message"`

	// When true Polaris manages the security groups.
	IsManagedByRubrik bool `json:"areSecurityGroupsRscManaged"`

	// Security group ids of cluster control plane and worker node. Only needs
	// to be specified if IsPolarisManaged is false.
	ClusterSecurityGroupID string `json:"clusterSecurityGroupId"`
	NodeSecurityGroupID    string `json:"nodeSecurityGroupId"`

	HealthCheckStatus struct {
		Status        string `json:"status"`
		FailureReason string `json:"failureReason"`
		LastUpdatedAt string `json:"lastUpdatedAt"`
		TaskchainID   string `json:"taskchainId"`
	}
}

// ExocomputeConfigsForAccount holds all exocompute configs for a specific
// account.
type ExocomputeConfigsForAccount struct {
	Account         CloudAccount          `json:"awsCloudAccount"`
	Configs         []ExocomputeConfig    `json:"exocomputeConfigs"`
	EligibleRegions []string              `json:"exocomputeEligibleRegions"`
	Feature         Feature               `json:"featureDetail"`
	MappedAccounts  []CloudAccountDetails `json:"mappedCloudAccounts"`
}

// CloudAccountDetails holds the details about an exocompute application account
// mapping.
type CloudAccountDetails struct {
	ID       uuid.UUID `json:"id"`
	NativeID string    `json:"nativeId"`
	Name     string    `json:"name"`
}

// ExocomputeConfigs returns all exocompute configs matching the specified
// filter. The filter can be used to search for account name or account id.
func (a API) ExocomputeConfigs(ctx context.Context, filter string) ([]ExocomputeConfigsForAccount, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, allAwsExocomputeConfigsQuery, struct {
		Filter string `json:"awsNativeAccountIdOrNamePrefix"`
	}{Filter: filter})
	if err != nil {
		return nil, fmt.Errorf("failed to request allAwsExocomputeConfigs: %w", err)
	}
	a.log.Printf(log.Debug, "allAwsExocomputeConfigs(%q): %s", filter, string(buf))

	var payload struct {
		Data struct {
			Result []ExocomputeConfigsForAccount `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allAwsExocomputeConfigs: %v", err)
	}

	return payload.Data.Result, nil
}

// ExocomputeConfigCreate represents an exocompute config to be created by
// Polaris.
type ExocomputeConfigCreate struct {
	Region Region `json:"region"`

	// Only required for RSC managed clusters
	VPCID   string   `json:"vpcId,omitempty"`
	Subnets []Subnet `json:"subnets,omitempty"`

	// When true Rubrik will manage the security groups.
	IsManagedByRubrik bool `json:"isRscManaged"`

	// Security group ids of cluster control plane and worker node. Only needs
	// to be specified if IsPolarisManaged is false.
	ClusterSecurityGroupId string `json:"clusterSecurityGroupId"`
	NodeSecurityGroupId    string `json:"nodeSecurityGroupId"`
}

// CreateExocomputeConfig creates a new exocompute config for the account with
// the specified RSC cloud account id. Returns the created exocompute config
func (a API) CreateExocomputeConfig(ctx context.Context, cloudAccountID uuid.UUID, config ExocomputeConfigCreate) (ExocomputeConfig, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, createAwsExocomputeConfigsQuery, struct {
		ID      uuid.UUID                `json:"cloudAccountId"`
		Configs []ExocomputeConfigCreate `json:"configs"`
	}{ID: cloudAccountID, Configs: []ExocomputeConfigCreate{config}})
	if err != nil {
		return ExocomputeConfig{}, fmt.Errorf("failed to request createAwsExocomputeConfigs: %w", err)
	}
	a.log.Printf(log.Debug, "createAwsExocomputeConfigs(%q, %v): %s", cloudAccountID, config, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				Configs []ExocomputeConfig `json:"configs"`
			} `json:"createAwsExocomputeConfigs"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return ExocomputeConfig{}, fmt.Errorf("failed to unmarshal createAwsExocomputeConfigs: %v", err)
	}
	if len(payload.Data.Query.Configs) != 1 {
		return ExocomputeConfig{}, errors.New("expected a single result")
	}
	if payload.Data.Query.Configs[0].Message != "" {
		return ExocomputeConfig{}, errors.New(payload.Data.Query.Configs[0].Message)
	}

	return payload.Data.Query.Configs[0], nil
}

// UpdateExocomputeConfig updates an exocompute config for the account with
// the specified RSC cloud account id. Returns the updated exocompute config.
func (a API) UpdateExocomputeConfig(ctx context.Context, cloudAccountID uuid.UUID, config ExocomputeConfigCreate) (ExocomputeConfig, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, updateAwsExocomputeConfigsQuery, struct {
		ID      uuid.UUID                `json:"cloudAccountId"`
		Configs []ExocomputeConfigCreate `json:"configs"`
	}{ID: cloudAccountID, Configs: []ExocomputeConfigCreate{config}})
	if err != nil {
		return ExocomputeConfig{}, fmt.Errorf("failed to request updateAwsExocomputeConfigs: %w", err)
	}
	a.log.Printf(log.Debug, "updateAwsExocomputeConfigs(%q, %v): %s", cloudAccountID, config, string(buf))

	var payload struct {
		Data struct {
			Query struct {
				Configs []ExocomputeConfig `json:"configs"`
			} `json:"updateAwsExocomputeConfigs"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return ExocomputeConfig{}, fmt.Errorf("failed to unmarshal updateAwsExocomputeConfigs: %v", err)
	}
	if len(payload.Data.Query.Configs) != 1 {
		return ExocomputeConfig{}, errors.New("expected a single result")
	}
	if payload.Data.Query.Configs[0].Message != "" {
		return ExocomputeConfig{}, errors.New(payload.Data.Query.Configs[0].Message)
	}

	return payload.Data.Query.Configs[0], nil
}

// DeleteExocomputeConfig deletes the exocompute config with the specified RSC
// exocompute config id.
func (a API) DeleteExocomputeConfig(ctx context.Context, configID uuid.UUID) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, deleteAwsExocomputeConfigsQuery, struct {
		IDs []uuid.UUID `json:"configIdsToBeDeleted"`
	}{IDs: []uuid.UUID{configID}})
	if err != nil {
		return fmt.Errorf("failed to request deleteAwsExocomputeConfigs: %w", err)
	}
	a.log.Printf(log.Debug, "deleteAwsExocomputeConfigs(%q): %s", configID, string(buf))

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
		return fmt.Errorf("failed to unmarshal deleteAwsExocomputeConfigs: %v", err)
	}
	if len(payload.Data.Query.Status) != 1 {
		return errors.New("expected a single result")
	}
	if !payload.Data.Query.Status[0].Success {
		return errors.New("delete exocompute config failed")
	}

	return nil
}

// StartExocomputeDisableJob starts a task chain job to disable the Exocompute
// feature for the account with the specified RSC native account id. Returns the
// RSC task chain id.
func (a API) StartExocomputeDisableJob(ctx context.Context, nativeID uuid.UUID) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, startAwsExocomputeDisableJobQuery, struct {
		ID uuid.UUID `json:"cloudAccountId"`
	}{ID: nativeID})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to request startAwsExocomputeDisableJob: %w", err)
	}
	a.log.Printf(log.Debug, "startAwsExocomputeDisableJob(%q): %s", nativeID, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				Error string    `json:"error"`
				JobID uuid.UUID `json:"jobId"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal startAwsExocomputeDisableJob: %v", err)
	}
	if payload.Data.Result.Error != "" {
		return uuid.Nil, errors.New(payload.Data.Result.Error)
	}

	return payload.Data.Result.JobID, nil
}

// MapCloudAccountExocomputeAccount maps the slice of exocompute application
// accounts to the specified exocompute host account.
func (a API) MapCloudAccountExocomputeAccount(ctx context.Context, hostID uuid.UUID, appIDs []uuid.UUID) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, mapCloudAccountExocomputeAccountQuery, struct {
		HostID uuid.UUID   `json:"exocomputeCloudAccountId"`
		AppIDs []uuid.UUID `json:"cloudAccountIds"`
	}{HostID: hostID, AppIDs: appIDs})
	if err != nil {
		return fmt.Errorf("failed to request mapCloudAccountExocomputeAccount: %w", err)
	}
	a.log.Printf(log.Debug, "mapCloudAccountExocomputeAccount(%q, %v): %s", hostID, appIDs, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				Success bool `json:"isSuccess"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal mapCloudAccountExocomputeAccount: %v", err)
	}
	if !payload.Data.Result.Success {
		return errors.New("")
	}

	return nil
}

// UnmapCloudAccountExocomputeAccount unmaps the slice of exocompute application
// accounts.
func (a API) UnmapCloudAccountExocomputeAccount(ctx context.Context, appIDs []uuid.UUID) error {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, unmapCloudAccountExocomputeAccountQuery, struct {
		AppIDs []uuid.UUID `json:"cloudAccountIds"`
	}{AppIDs: appIDs})
	if err != nil {
		return fmt.Errorf("failed to request unmapCloudAccountExocomputeAccount: %w", err)
	}
	a.log.Printf(log.Debug, "unmapCloudAccountExocomputeAccount(%v): %s", appIDs, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				Success bool `json:"isSuccess"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal unmapCloudAccountExocomputeAccount: %v", err)
	}
	if !payload.Data.Result.Success {
		return errors.New("")
	}

	return nil
}

// ConnectExocomputeCluster connects the named cluster to specified exocompute
// configration. The cluster ID and connection command are returned.
func (a API) ConnectExocomputeCluster(ctx context.Context, configID uuid.UUID, clusterName string) (uuid.UUID, string, error) {
	a.log.Print(log.Trace)

	buf, err := a.GQL.Request(ctx, awsExocomputeClusterConnectQuery, struct {
		ConfigID    uuid.UUID `json:"exocomputeConfigId"`
		ClusterName string    `json:"clusterName"`
	}{ConfigID: configID, ClusterName: clusterName})
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to request awsExocomputeClusterConnect: %w", err)
	}
	a.log.Printf(log.Debug, "awsExocomputeClusterConnect(%q, %q): %s", configID, clusterName, string(buf))

	var payload struct {
		Data struct {
			Result struct {
				ID      uuid.UUID `json:"clusterUuid"`
				Command string    `json:"connectionCommand"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to unmarshal awsExocomputeClusterConnect: %v", err)
	}

	return payload.Data.Result.ID, payload.Data.Result.Command, nil
}

// DisconnectExocomputeCluster disconnects the exocomptue cluster with the
// specified ID from RSC.
func (a API) DisconnectExocomputeCluster(ctx context.Context, clusterID uuid.UUID) error {
	a.log.Print(log.Trace)

	_, err := a.GQL.Request(ctx, disconnectAwsExocomputeClusterQuery, struct {
		ClusterID uuid.UUID `json:"clusterId"`
	}{ClusterID: clusterID})
	if err != nil {
		return fmt.Errorf("failed to request disconnectAwsExocomputeCluster: %w", err)
	}
	a.log.Printf(log.Debug, "disconnectAwsExocomputeCluster(%q)", clusterID)

	return nil
}
