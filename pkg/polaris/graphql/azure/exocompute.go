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
	Account         CloudAccount          `json:"azureCloudAccount"`
	Configs         []ExoConfig           `json:"configs"`
	EligibleRegions []string              `json:"exocomputeEligibleRegions"`
	Feature         Feature               `json:"featureDetails"`
	MappedAccounts  []CloudAccountDetails `json:"mappedCloudAccounts"`
}

func (r ExoConfigsForAccount) ListQuery(filter string) (string, any) {
	return allAzureExocomputeConfigsInAccountQuery, struct {
		Filter string `json:"azureExocomputeSearchQuery"`
	}{Filter: filter}
}

// ExoConfig represents a single exocompute configuration.
type ExoConfig struct {
	ID       uuid.UUID `json:"configUuid"`
	Region   Region    `json:"region"`
	SubnetID string    `json:"subnetNativeId"`
	Message  string    `json:"message"`

	// When true, Rubrik will manage the security groups.
	ManagedByRubrik bool `json:"isRscManaged"`
}

// CloudAccountDetails holds the details about an exocompute application account
// mapping.
type CloudAccountDetails struct {
	ID       uuid.UUID `json:"id"`
	NativeID string    `json:"nativeId"`
	Name     string    `json:"name"`
}

// ExoCreateParams holds the parameters for an exocompute configuration to be
// created by RSC.
type ExoCreateParams struct {
	Region   Region `json:"region"`
	SubnetID string `json:"subnetNativeId"`

	// When true, Rubrik will manage the security groups.
	IsManagedByRubrik bool `json:"isRscManaged"`
}

type ExoCreateResult struct {
	Configs []ExoConfig `json:"configs"`
}

func (r ExoCreateResult) CreateQuery(cloudAccountID uuid.UUID, createParams ExoCreateParams) (string, any) {
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

	return r.Configs[0].ID, nil
}

type ExoDeleteResult struct {
	FailIDs    []uuid.UUID `json:"deletionFailedIds"`
	SuccessIDs []uuid.UUID `json:"deletionSuccessIds"`
}

func (r ExoDeleteResult) DeleteQuery(configID uuid.UUID) (string, any) {
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
