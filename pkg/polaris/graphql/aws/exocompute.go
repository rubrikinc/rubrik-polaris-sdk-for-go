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

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// ExoConfigsForAccount holds all exocompute configurations for a specific
// account.
type ExoConfigsForAccount struct {
	Account         CloudAccount          `json:"awsCloudAccount"`
	Configs         []ExoConfig           `json:"exocomputeConfigs"`
	EligibleRegions []string              `json:"exocomputeEligibleRegions"`
	Feature         Feature               `json:"featureDetail"`
	MappedAccounts  []CloudAccountDetails `json:"mappedCloudAccounts"`
}

func (r ExoConfigsForAccount) ListQuery(filter string) (string, any) {
	return allAwsExocomputeConfigsQuery, struct {
		Filter string `json:"awsNativeAccountIdOrNamePrefix"`
	}{Filter: filter}
}

// ExoConfig represents a single exocompute configuration.
type ExoConfig struct {
	ID      string `json:"configUuid"`
	Region  Region `json:"region"`
	VPCID   string `json:"vpcId"`
	Subnet1 Subnet `json:"subnet1"`
	Subnet2 Subnet `json:"subnet2"`
	Message string `json:"message"`

	// When true, Polaris manages the security groups.
	IsManagedByRubrik bool `json:"areSecurityGroupsRscManaged"`

	// The security group IDs of the cluster control plane and worker nodes.
	// Only needs to be specified if IsPolarisManaged is false.
	ClusterSecurityGroupID string `json:"clusterSecurityGroupId"`
	NodeSecurityGroupID    string `json:"nodeSecurityGroupId"`

	HealthCheckStatus struct {
		Status        string `json:"status"`
		FailureReason string `json:"failureReason"`
		LastUpdatedAt string `json:"lastUpdatedAt"`
		TaskchainID   string `json:"taskchainId"`
	}
}

// CloudAccountDetails holds the details about an exocompute application account
// mapping.
type CloudAccountDetails struct {
	ID       uuid.UUID `json:"id"`
	NativeID string    `json:"nativeId"`
	Name     string    `json:"name"`
}

// Subnet represents an AWS subnet.
type Subnet struct {
	ID               string `json:"subnetId"`
	AvailabilityZone string `json:"availabilityZone"`
}

// ExoCreateParams represents the parameters required to create an AWS
// exocompute configuration.
type ExoCreateParams struct {
	Region Region `json:"region"`

	// Only required for RSC managed clusters
	VPCID   string   `json:"vpcId,omitempty"`
	Subnets []Subnet `json:"subnets,omitempty"`

	// When true, Rubrik will manage the security groups.
	IsManagedByRubrik bool `json:"isRscManaged"`

	// The security group IDs of the cluster control plane and worker nodes.
	// Only needs to be specified if IsPolarisManaged is false.
	ClusterSecurityGroupId string `json:"clusterSecurityGroupId,omitempty"`
	NodeSecurityGroupId    string `json:"nodeSecurityGroupId,omitempty"`
}

// ExoCreateResult represents the result of creating an AWS exocompute
// configuration.
type ExoCreateResult struct {
	Configs []ExoConfig `json:"exocomputeConfigs"`
}

func (r ExoCreateResult) CreateQuery(cloudAccountID uuid.UUID, createParams ExoCreateParams) (string, any) {
	return createAwsExocomputeConfigsQuery, struct {
		ID      uuid.UUID         `json:"cloudAccountId"`
		Configs []ExoCreateParams `json:"configs"`
	}{ID: cloudAccountID, Configs: []ExoCreateParams{createParams}}
}

func (r ExoCreateResult) Validate() (uuid.UUID, error) {
	if len(r.Configs) != 1 {
		return uuid.Nil, errors.New("expected a single create result")
	}
	if msg := r.Configs[0].Message; msg != "" {
		return uuid.Nil, errors.New(msg)
	}
	id, err := uuid.Parse(r.Configs[0].ID)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

// ExoUpdateParams represents the parameters required to update an AWS
// exocompute configuration.
type ExoUpdateParams ExoCreateParams

// ExoUpdateResult represents the result of updating an AWS exocompute
// configuration.
type ExoUpdateResult ExoCreateResult

func (r ExoUpdateResult) UpdateQuery(cloudAccountID uuid.UUID, updateParams ExoUpdateParams) (string, any) {
	return updateAwsExocomputeConfigsQuery, struct {
		ID     uuid.UUID       `json:"cloudAccountId"`
		Config ExoUpdateParams `json:"configs"`
	}{ID: cloudAccountID, Config: updateParams}
}

func (r ExoUpdateResult) Validate() (uuid.UUID, error) {
	if len(r.Configs) != 1 {
		return uuid.Nil, errors.New("expected a single update result")
	}
	if msg := r.Configs[0].Message; msg != "" {
		return uuid.Nil, errors.New(msg)
	}
	id, err := uuid.Parse(r.Configs[0].ID)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

// ExoDeleteResult represents the result of deleting an AWS exocompute
// configuration.
type ExoDeleteResult struct {
	Status []struct {
		ID      uuid.UUID `json:"exocomputeConfigId"`
		Success bool      `json:"success"`
	} `json:"deletionStatus"`
}

func (r ExoDeleteResult) DeleteQuery(configID uuid.UUID) (string, any) {
	return deleteAwsExocomputeConfigsQuery, struct {
		IDs []uuid.UUID `json:"configIdsToBeDeleted"`
	}{IDs: []uuid.UUID{configID}}
}

func (r ExoDeleteResult) Validate() (uuid.UUID, error) {
	if len(r.Status) != 1 {
		return uuid.Nil, errors.New("expected a single delete result")
	}
	if !r.Status[0].Success {
		return uuid.Nil, errors.New("failed to delete exocompute config")
	}

	return r.Status[0].ID, nil
}

// ExoMapResult represents the result of mapping an AWS application cloud
// account to an AWS host cloud account.
type ExoMapResult struct {
	Success bool `json:"isSuccess"`
}

func (r ExoMapResult) MapQuery(hostCloudAccountID, appCloudAccountID uuid.UUID) (string, any) {
	return mapCloudAccountExocomputeAccountQuery, struct {
		HostCloudAccountID uuid.UUID   `json:"exocomputeCloudAccountId"`
		AppCloudAccountIDs []uuid.UUID `json:"cloudAccountIds"`
	}{HostCloudAccountID: hostCloudAccountID, AppCloudAccountIDs: []uuid.UUID{appCloudAccountID}}
}

func (r ExoMapResult) Validate() error {
	if !r.Success {
		return errors.New("failed to map application cloud account")
	}

	return nil
}

// ExoUnmapResult represents the result of unmapping an AWS application cloud
// account.
type ExoUnmapResult struct {
	Success bool `json:"isSuccess"`
}

func (r ExoUnmapResult) UnmapQuery(appCloudAccountID uuid.UUID) (string, any) {
	return unmapCloudAccountExocomputeAccountQuery, struct {
		AppCloudAccountIDs []uuid.UUID `json:"cloudAccountIds"`
	}{AppCloudAccountIDs: []uuid.UUID{appCloudAccountID}}
}

func (r ExoUnmapResult) Validate() error {
	if !r.Success {
		return errors.New("failed to unmap application cloud account")
	}

	return nil
}

// StartExocomputeDisableJob starts a task chain job to disable the Exocompute
// feature for the account with the specified RSC native account id. Returns the
// RSC task chain id.
func (a API) StartExocomputeDisableJob(ctx context.Context, nativeID uuid.UUID) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	query := startAwsExocomputeDisableJobQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ID uuid.UUID `json:"cloudAccountId"`
	}{ID: nativeID})
	if err != nil {
		return uuid.Nil, graphql.RequestError(query, err)
	}
	graphql.LogResponse(a.log, query, buf)

	var payload struct {
		Data struct {
			Result struct {
				Error string    `json:"error"`
				JobID uuid.UUID `json:"jobId"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, graphql.UnmarshalError(query, err)
	}
	if payload.Data.Result.Error != "" {
		return uuid.Nil, errors.New(payload.Data.Result.Error)
	}

	return payload.Data.Result.JobID, nil
}

// ConnectExocomputeCluster connects the named cluster to a specified exocompute
// configuration. The cluster ID and two different ways to connect the cluster
// are returned. The first way to connect the cluster is the kubectl connection
// command, and the second way is the k8s spec (YAML).
func (a API) ConnectExocomputeCluster(ctx context.Context, configID uuid.UUID, clusterName string) (uuid.UUID, string, string, error) {
	a.log.Print(log.Trace)

	query := awsExocomputeClusterConnectQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ConfigID    uuid.UUID `json:"exocomputeConfigId"`
		ClusterName string    `json:"clusterName"`
	}{ConfigID: configID, ClusterName: clusterName})
	if err != nil {
		return uuid.Nil, "", "", graphql.RequestError(query, err)
	}
	graphql.LogResponse(a.log, query, buf)

	var payload struct {
		Data struct {
			Result struct {
				ID        uuid.UUID `json:"clusterUuid"`
				Command   string    `json:"connectionCommand"`
				SetupYAML string    `json:"clusterSetupYaml"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return uuid.Nil, "", "", graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.ID, payload.Data.Result.Command, payload.Data.Result.SetupYAML, nil
}

// DisconnectExocomputeCluster disconnects the exocompute cluster with the
// specified ID from RSC.
func (a API) DisconnectExocomputeCluster(ctx context.Context, clusterID uuid.UUID) error {
	a.log.Print(log.Trace)

	query := disconnectAwsExocomputeClusterQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ClusterID uuid.UUID `json:"clusterId"`
	}{ClusterID: clusterID})
	if err != nil {
		return graphql.RequestError(query, err)
	}
	graphql.LogResponse(a.log, query, buf)

	return nil
}

// ClusterConnectionInfo returns information about the connected cluster,
// specifically the Kubernetes manifest, containing the cluster gateway spec.
func (a API) ClusterConnectionInfo(ctx context.Context, configID uuid.UUID) (string, string, error) {
	a.log.Print(log.Trace)

	query := awsExocomputeGetClusterConnectionInfoQuery
	buf, err := a.GQL.Request(ctx, query, struct {
		ConfigID uuid.UUID `json:"exocomputeConfigId"`
	}{ConfigID: configID})
	if err != nil {
		return "", "", graphql.RequestError(query, err)
	}
	graphql.LogResponse(a.log, query, buf)

	var payload struct {
		Data struct {
			Result struct {
				Command   string `json:"connectionCommand"`
				SetupYAML string `json:"clusterSetupYaml"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return "", "", graphql.UnmarshalError(query, err)
	}

	return payload.Data.Result.Command, payload.Data.Result.SetupYAML, nil
}
