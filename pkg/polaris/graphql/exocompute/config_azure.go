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
	"errors"
	"fmt"

	"github.com/google/uuid"
	gqlazure "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/azure"
)

// AzureConfigurationsFilter holds the filter for an Azure exocompute
// configuration list operation.
type AzureConfigurationsFilter struct {
	SearchQuery string `json:"azureExocomputeSearchQuery"`
}

func (p AzureConfigurationsFilter) ListQuery() (string, any, AzureConfigurationsForCloudAccount) {
	return allAzureExocomputeConfigsInAccountQuery, p, AzureConfigurationsForCloudAccount{}
}

// AzureConfigurationsForCloudAccount holds the result of an Azure exocompute
// configuration list operation.
type AzureConfigurationsForCloudAccount struct {
	CloudAccount gqlazure.CloudAccount `json:"azureCloudAccount"`
	Configs      []AzureConfiguration  `json:"configs"`
}

// AzureConfiguration holds a single Azure exocompute configuration.
type AzureConfiguration struct {
	ID                    uuid.UUID                    `json:"configUuid"`
	Region                azure.CloudAccountRegionEnum `json:"region"`
	SubnetID              string                       `json:"subnetNativeId"`
	ManagedByRubrik       bool                         `json:"isRscManaged"`
	PodOverlayNetworkCIDR string                       `json:"podOverlayNetworkCidr"`
	PodSubnetID           string                       `json:"podSubnetNativeId"`

	// HealthCheckStatus represents the health status of an exocompute cluster.
	HealthCheckStatus struct {
		Status        string `json:"status"`
		FailureReason string `json:"failureReason"`
		LastUpdatedAt string `json:"lastUpdatedAt"`
		TaskchainID   string `json:"taskchainId"`
	} `json:"healthCheckStatus"`
}

// CreateAzureConfigurationParams holds the parameters for an Azure exocompute
// configuration create operation.
type CreateAzureConfigurationParams struct {
	CloudAccountID uuid.UUID                    `json:"-"`
	Region         azure.CloudAccountRegionEnum `json:"region"`

	// When true, Rubrik will manage the security groups.
	IsManagedByRubrik bool `json:"isRscManaged"`

	SubnetID              string `json:"subnetNativeId,omitempty"`
	PodOverlayNetworkCIDR string `json:"podOverlayNetworkCidr,omitempty"`
	PodSubnetID           string `json:"podSubnetNativeId,omitempty"`

	// When true, a health check will be triggered after the configuration is
	// created.
	TriggerHealthCheck bool `json:"-"`
}

func (p CreateAzureConfigurationParams) CreateQuery() (string, any, CreateAzureConfigurationResult) {
	params := struct {
		CloudAccountID     uuid.UUID                      `json:"cloudAccountId"`
		Config             CreateAzureConfigurationParams `json:"azureExocomputeRegionConfig"`
		TriggerHealthCheck bool                           `json:"triggerHealthCheck"`
	}{CloudAccountID: p.CloudAccountID, Config: p, TriggerHealthCheck: p.TriggerHealthCheck}
	return addAzureCloudAccountExocomputeConfigurationsQuery, params, CreateAzureConfigurationResult{}
}

// CreateAzureConfigurationResult holds the result of an Azure exocompute
// configuration create operation.
type CreateAzureConfigurationResult struct {
	Configs []struct {
		ID      string `json:"configUuid"`
		Message string `json:"message"`
	} `json:"configs"`
}

func (r CreateAzureConfigurationResult) Validate() (uuid.UUID, error) {
	if len(r.Configs) != 1 {
		return uuid.Nil, errors.New("expected a single configuration to be created")
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

// DeleteAzureConfigurationParams holds the parameters for an Azure exocompute
// configuration delete operation.
type DeleteAzureConfigurationParams struct {
	ConfigID uuid.UUID `json:"cloudAccountId"`
}

func (p DeleteAzureConfigurationParams) DeleteQuery() (string, any, DeleteAzureConfigurationResult) {
	return deleteAzureCloudAccountExocomputeConfigurationsQuery, p, DeleteAzureConfigurationResult{}
}

// DeleteAzureConfigurationResult holds the result of an Azure exocompute
// configuration delete operation.
type DeleteAzureConfigurationResult struct {
	FailIDs    []uuid.UUID `json:"deletionFailedIds"`
	SuccessIDs []uuid.UUID `json:"deletionSuccessIds"`
}

func (r DeleteAzureConfigurationResult) Validate() error {
	if n := len(r.FailIDs); n > 0 {
		return fmt.Errorf("expected no delete failures: %d", n)
	}
	if len(r.SuccessIDs) != 1 {
		return errors.New("expected a single configuration to be deleted")
	}
	return nil
}
