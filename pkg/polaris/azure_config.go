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

package polaris

import (
	"encoding/json"
	"os"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
)

// AzureServicePrincipalConfig -
type AzureServicePrincipalConfig struct {
	AppID        string `json:"app_id"`
	AppName      string `json:"app_name"`
	AppSecret    string `json:"app_secret"`
	TenantID     string `json:"tenant_id"`
	TenantDomain string `json:"tenant_domain"`
}

// AzureServicePrincipalFromFile -
func AzureServicePrincipalFromFile(file string) (AzureServicePrincipal, error) {
	buf, err := os.ReadFile(file)
	if err != nil {
		return AzureServicePrincipal{}, err
	}

	var config AzureServicePrincipalConfig
	if err := json.Unmarshal(buf, &config); err != nil {
		return AzureServicePrincipal{}, err
	}

	appID, err := uuid.Parse(config.AppID)
	if err != nil {
		return AzureServicePrincipal{}, err
	}

	tenantID, err := uuid.Parse(config.TenantID)
	if err != nil {
		return AzureServicePrincipal{}, err
	}

	principal := AzureServicePrincipal{
		Cloud:        graphql.AzurePublic,
		AppID:        appID,
		AppName:      config.AppName,
		AppSecret:    config.AppSecret,
		TenantID:     tenantID,
		TenantDomain: config.TenantDomain,
	}

	return principal, nil
}

// AzureDefaultServicePrincipal -
func AzureDefaultServicePrincipal() (AzureServicePrincipal, error) {
	return AzureServicePrincipalFromFile(os.Getenv("AZURE_SERVICEPRINCIPAL_LOCATION"))
}

// AzureSubscriptionConfig -
type AzureSubscriptionConfig struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	TenantDomain string   `json:"tenant_domain"`
	Regions      []string `json:"regions"`
}

// AzureSubscriptionFromFile -
func AzureSubscriptionFromFile(file string) (AzureSubscriptionIn, error) {
	buf, err := os.ReadFile(file)
	if err != nil {
		return AzureSubscriptionIn{}, err
	}

	var config AzureSubscriptionConfig
	if err := json.Unmarshal(buf, &config); err != nil {
		return AzureSubscriptionIn{}, err
	}

	id, err := uuid.Parse(config.ID)
	if err != nil {
		return AzureSubscriptionIn{}, err
	}

	azureRegions := make([]graphql.AzureRegion, 0, 10)
	for _, region := range config.Regions {
		azureRegion, err := graphql.AzureParseRegion(region)
		if err != nil {
			return AzureSubscriptionIn{}, err
		}
		azureRegions = append(azureRegions, azureRegion)
	}

	subscription := AzureSubscriptionIn{
		Cloud:        graphql.AzurePublic,
		ID:           id,
		Name:         config.Name,
		TenantDomain: config.TenantDomain,
		Regions:      azureRegions,
	}

	return subscription, nil
}

// AzureDefaultSubscription -
func AzureDefaultSubscription() (AzureSubscriptionIn, error) {
	return AzureSubscriptionFromFile(os.Getenv("AZURE_SUBSCRIPTION_LOCATION"))
}
