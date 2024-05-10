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
	"errors"

	"github.com/google/uuid"
)

// ExoConfigsForAccount holds all exocompute configurations for a specific
// account.
type ExoConfigsForAccount struct {
	Account         CloudAccount `json:"azureCloudAccount"`
	Configs         []ExoConfig  `json:"configs"`
	EligibleRegions []string     `json:"exocomputeEligibleRegions"`
	Feature         Feature      `json:"featureDetails"`
}

func (r ExoConfigsForAccount) ListQuery(filter string) (string, any) {
	return allAzureExocomputeConfigsInAccountQuery, struct {
		Filter string `json:"azureExocomputeSearchQuery"`
	}{Filter: filter}
}

// ExoConfig represents a single exocompute configuration.
type ExoConfig struct {
	ID                    string `json:"configUuid"`
	Region                Region `json:"region"`
	SubnetID              string `json:"subnetNativeId"`
	Message               string `json:"message"`
	ManagedByRubrik       bool   `json:"isRscManaged"` // When true, Rubrik will manage the security groups.
	PodOverlayNetworkCIDR string `json:"podOverlayNetworkCidr"`
	PodSubnetID           string `json:"podSubnetNativeId"`

	// HealthCheckStatus represents the health status of an exocompute cluster.
	HealthCheckStatus struct {
		Status        string `json:"status"`
		FailureReason string `json:"failureReason"`
		LastUpdatedAt string `json:"lastUpdatedAt"`
		TaskchainID   string `json:"taskchainId"`
	} `json:"healthCheckStatus"`
}

// ExoCreateParams represents the parameters required to create an Azure
// exocompute configuration.
type ExoCreateParams struct {
	Region                Region `json:"region"`
	SubnetID              string `json:"subnetNativeId"`
	IsManagedByRubrik     bool   `json:"isRscManaged"` // When true, Rubrik will manage the security groups.
	PodOverlayNetworkCIDR string `json:"podOverlayNetworkCidr,omitempty"`
	PodSubnetID           string `json:"podSubnetNativeId,omitempty"`
}

// ExoCreateResult represents the result of creating an Azure exocompute
// configuration.
type ExoCreateResult struct {
	Configs []ExoConfig `json:"configs"`
}

func (ExoCreateResult) CreateQuery(cloudAccountID uuid.UUID, createParams ExoCreateParams) (string, any) {
	return addAzureCloudAccountExocomputeConfigurationsQuery, struct {
		ID      uuid.UUID         `json:"cloudAccountId"`
		Configs []ExoCreateParams `json:"azureExocomputeRegionConfigs"`
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

// ExoDeleteResult represents the result of deleting an Azure exocompute
// configuration.
type ExoDeleteResult struct {
	FailIDs    []uuid.UUID `json:"deletionFailedIds"`
	SuccessIDs []uuid.UUID `json:"deletionSuccessIds"`
}

func (ExoDeleteResult) DeleteQuery(configID uuid.UUID) (string, any) {
	return deleteAzureCloudAccountExocomputeConfigurationsQuery, struct {
		IDs []uuid.UUID `json:"cloudAccountIds"`
	}{IDs: []uuid.UUID{configID}}
}

func (r ExoDeleteResult) Validate() (uuid.UUID, error) {
	if len(r.FailIDs) > 0 {
		return uuid.Nil, errors.New("expected no delete failures")
	}
	if len(r.SuccessIDs) != 1 {
		return uuid.Nil, errors.New("expected a single delete result")
	}

	return r.SuccessIDs[0], nil
}

// ExoMapResult represents the result of mapping an Azure application cloud
// account to an Azure host cloud account.
type ExoMapResult struct {
	Success bool `json:"isSuccess"`
}

func (ExoMapResult) MapQuery(hostCloudAccountID, appCloudAccountID uuid.UUID) (string, any) {
	return mapAzureCloudAccountExocomputeSubscriptionQuery, struct {
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

// ExoUnmapResult represents the result of unmapping an Azure application cloud
// account.
type ExoUnmapResult struct {
	Success bool `json:"isSuccess"`
}

func (ExoUnmapResult) UnmapQuery(appCloudAccountID uuid.UUID) (string, any) {
	return unmapAzureCloudAccountExocomputeSubscriptionQuery, struct {
		AppCloudAccountIDs []uuid.UUID `json:"cloudAccountIds"`
	}{AppCloudAccountIDs: []uuid.UUID{appCloudAccountID}}
}

func (r ExoUnmapResult) Validate() error {
	if !r.Success {
		return errors.New("failed to unmap application cloud account")
	}

	return nil
}
