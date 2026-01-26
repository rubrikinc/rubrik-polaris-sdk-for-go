// Copyright 2026 Rubrik, Inc.
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
	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/gcp"
)

// GCPConfigurationsFilter holds the filter for a GCP exocompute configuration
// list operation.
type GCPConfigurationsFilter struct {
	CloudAccountID uuid.UUID                    `json:"cloudAccountId"`
	Regions        []gcp.CloudAccountRegionEnum `json:"regions,omitempty"`
}

func (p GCPConfigurationsFilter) ListQuery() (string, any, GCPConfigurations) {
	return gcpExocomputeConfigsQuery, p, GCPConfigurations{}
}

// GCPConfigurations holds the result of a GCP exocompute configuration list
// operation.
type GCPConfigurations struct {
	Configs []GCPConfiguration `json:"exocomputeConfigs"`
}

// GCPConfiguration holds a single GCP exocompute configuration.
type GCPConfiguration struct {
	ID                uuid.UUID         `json:"configId"`
	Config            GCPRegionalConfig `json:"regionalExocomputeConfig"`
	HealthCheckStatus struct {
		Status        string `json:"status"`
		FailureReason string `json:"failureReason"`
		LastUpdatedAt string `json:"lastUpdatedAt"`
		TaskchainID   string `json:"taskchainId"`
	} `json:"healthCheckStatus"`
}

// GCPRegionalConfig holds the configuration for a GCP region. A GCP Exocompute
// configuration consist of a set of regional configurations.
type GCPRegionalConfig struct {
	Region         gcp.CloudAccountRegionEnum `json:"region"`
	SubnetName     string                     `json:"subnetName"`
	VPCNetworkName string                     `json:"vpcNetworkName"`
}

// UpdateGCPConfigurationParams holds the parameters for a GCP exocompute
// configuration update operation.
type UpdateGCPConfigurationParams struct {
	CloudAccountID     uuid.UUID           `json:"cloudAccountId"`
	RegionalConfigs    []GCPRegionalConfig `json:"regionalConfigs"`
	TriggerHealthCheck bool                `json:"triggerHealthCheck"`
}

func (p UpdateGCPConfigurationParams) UpdateQuery() (string, any, UpdateGCPConfigurationResult) {
	return setGcpExocomputeConfigsQuery, p, UpdateGCPConfigurationResult{}
}

// UpdateGCPConfigurationResult holds the result of a GCP exocompute
// configuration update operation. Note, the return type of the GraphQL query is
// Void.
type UpdateGCPConfigurationResult struct{}

func (r UpdateGCPConfigurationResult) Validate() (uuid.UUID, error) {
	return uuid.Nil, nil
}

// DeleteGCPConfigurationParams holds the parameters for a GCP exocompute
// configuration delete operation.
type DeleteGCPConfigurationParams struct {
	CloudAccountID uuid.UUID `json:"cloudAccountId"`
}

func (p DeleteGCPConfigurationParams) DeleteQuery() (string, any, DeleteGCPConfigurationResult) {
	params := struct {
		DeleteGCPConfigurationParams
		RegionalConfigs    []GCPRegionalConfig `json:"regionalConfigs"`
		TriggerHealthCheck bool                `json:"triggerHealthCheck"`
	}{DeleteGCPConfigurationParams: p, RegionalConfigs: []GCPRegionalConfig{}, TriggerHealthCheck: true}
	return setGcpExocomputeConfigsQuery, params, DeleteGCPConfigurationResult{}
}

// DeleteGCPConfigurationResult holds the result of a GCP exocompute
// configuration delete operation. Note, the return type of the GraphQL query is
// Void.
type DeleteGCPConfigurationResult struct{}

func (r DeleteGCPConfigurationResult) Validate() error {
	return nil
}
