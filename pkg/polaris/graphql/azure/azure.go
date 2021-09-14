//go:generate go run ../queries_gen.go azure

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
	"strings"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// Cloud represents the Azure cloud type.
type Cloud string

const (
	ChinaCloud  Cloud = "AZURECHINACLOUD"
	PublicCloud Cloud = "AZUREPUBLICCLOUD"
)

// Region represents an Azure region in Polaris.
type Region string

const (
	RegionUnknown            Region = "UNKNOWN_AZURE_REGION"
	RegionAustraliaCentral   Region = "AUSTRALIACENTRAL"
	RegionAustraliaCentral2  Region = "AUSTRALIACENTRAL2"
	RegionAustraliaEast      Region = "AUSTRALIAEAST"
	RegionAustraliaSouthEast Region = "AUSTRALIASOUTHEAST"
	RegionBrazilSouth        Region = "BRAZILSOUTH"
	RegionCanadaCentral      Region = "CANADACENTRAL"
	RegionCanadaEast         Region = "CANADAEAST"
	RegionCentralIndia       Region = "CENTRALINDIA"
	RegionCentralUS          Region = "CENTRALUS"
	RegionChinaEast          Region = "CHINAEAST"
	RegionChinaEast2         Region = "CHINAEAST2"
	RegionChinaNorth         Region = "CHINANORTH"
	RegionChinaNorth2        Region = "CHINANORTH2"
	RegionEastAsia           Region = "EASTASIA"
	RegionEastUS             Region = "EASTUS"
	RegionEastUS2            Region = "EASTUS2"
	RegionFranceCentral      Region = "FRANCECENTRAL"
	RegionFranceSouth        Region = "FRANCESOUTH"
	RegionGermanyNorth       Region = "GERMANYNORTH"
	RegionGermanyWestCentral Region = "GERMANYWESTCENTRAL"
	RegionJapanEast          Region = "JAPANEAST"
	RegionJapanWest          Region = "JAPANWEST"
	RegionKoreaCentral       Region = "KOREACENTRAL"
	RegionKoreaSouth         Region = "KOREASOUTH"
	RegionNorthCentralUS     Region = "NORTHCENTRALUS"
	RegionNorthEurope        Region = "NORTHEUROPE"
	RegionNorwayEast         Region = "NORWAYEAST"
	RegionNorwayWest         Region = "NORWAYWEST"
	RegionSouthAfricaNorth   Region = "SOUTHAFRICANORTH"
	RegionSouthAfricaWest    Region = "SOUTHAFRICAWEST"
	RegionSouthCentralUS     Region = "SOUTHCENTRALUS"
	RegionSouthEastAsia      Region = "SOUTHEASTASIA"
	RegionSouthIndia         Region = "SOUTHINDIA"
	RegionSwitzerlandNorth   Region = "SWITZERLANDNORTH"
	RegionSwitzerlandWest    Region = "SWITZERLANDWEST"
	RegionUAECentral         Region = "UAECENTRAL"
	RegionUAENorth           Region = "UAENORTH"
	RegionUKSouth            Region = "UKSOUTH"
	RegionUKWest             Region = "UKWEST"
	RegionWestCentralUS      Region = "WESTCENTRALUS"
	RegionWestEurope         Region = "WESTEUROPE"
	RegionWestIndia          Region = "WESTINDIA"
	RegionWestUS             Region = "WESTUS"
	RegionWestUS2            Region = "WESTUS2"
	RegionWestUS3            Region = "WESTUS3"
)

// FormatRegion returns the Region as a string formatted in Azure's style, i.e.
// lower case.
func FormatRegion(region Region) string {
	return strings.ToLower(string(region))
}

// FormatRegions returns the Regions as a slice of strings formatted in Azure's
// style, i.e. lower case.
func FormatRegions(regions []Region) []string {
	regs := make([]string, 0, len(regions))
	for _, region := range regions {
		regs = append(regs, FormatRegion(region))
	}

	return regs
}

// AzureRegions -
var validRegions = map[Region]struct{}{
	RegionAustraliaCentral:   {},
	RegionAustraliaCentral2:  {},
	RegionAustraliaEast:      {},
	RegionAustraliaSouthEast: {},
	RegionBrazilSouth:        {},
	RegionCanadaCentral:      {},
	RegionCanadaEast:         {},
	RegionCentralIndia:       {},
	RegionCentralUS:          {},
	RegionChinaEast:          {},
	RegionChinaEast2:         {},
	RegionChinaNorth:         {},
	RegionChinaNorth2:        {},
	RegionEastAsia:           {},
	RegionEastUS:             {},
	RegionEastUS2:            {},
	RegionFranceCentral:      {},
	RegionFranceSouth:        {},
	RegionGermanyNorth:       {},
	RegionGermanyWestCentral: {},
	RegionJapanEast:          {},
	RegionJapanWest:          {},
	RegionKoreaCentral:       {},
	RegionKoreaSouth:         {},
	RegionNorthCentralUS:     {},
	RegionNorthEurope:        {},
	RegionNorwayEast:         {},
	RegionNorwayWest:         {},
	RegionSouthAfricaNorth:   {},
	RegionSouthAfricaWest:    {},
	RegionSouthCentralUS:     {},
	RegionSouthEastAsia:      {},
	RegionSouthIndia:         {},
	RegionSwitzerlandNorth:   {},
	RegionSwitzerlandWest:    {},
	RegionUAECentral:         {},
	RegionUAENorth:           {},
	RegionUKSouth:            {},
	RegionUKWest:             {},
	RegionWestCentralUS:      {},
	RegionWestEurope:         {},
	RegionWestIndia:          {},
	RegionWestUS:             {},
	RegionWestUS2:            {},
	RegionWestUS3:            {},
}

// ParseRegion returns the Region matching the given region. Accepts both
// Polaris and Azure style region names.
func ParseRegion(region string) (Region, error) {
	// Polaris region name.
	r := Region(region)
	if _, ok := validRegions[r]; ok {
		return r, nil
	}

	// AWS region name.
	r = Region(strings.ToUpper(region))
	if _, ok := validRegions[r]; ok {
		return r, nil
	}

	return RegionUnknown, errors.New("polaris: invalid azure region")
}

// ParseRegions returns the Regions matching the given regions. Accepts both
// Polaris and Azure style region names.
func ParseRegions(regions []string) ([]Region, error) {
	regs := make([]Region, 0, len(regions))

	for _, r := range regions {
		region, err := ParseRegion(r)
		if err != nil {
			return nil, err
		}

		regs = append(regs, region)
	}

	return regs, nil
}

// API wraps around GraphQL clients to give them the Polaris Azure API.
type API struct {
	GQL *graphql.Client
}

// Wrap the GraphQL client in the Azure API.
func Wrap(gql *graphql.Client) API {
	return API{GQL: gql}
}

// SetCloudAccountCustomerAppCredentials sets the credentials for the customer
// application for the specified tenant domain. If the tenant domain is empty,
// set it for all the tenants of the customer.
func (a API) SetCloudAccountCustomerAppCredentials(ctx context.Context, cloud Cloud, appID, appTenantID uuid.UUID, appName, appTenantDomain, appSecretKey string) error {
	a.GQL.Log().Print(log.Trace, "polaris/graphql/azure.SetCloudAccountCustomerAppCredentials")

	buf, err := a.GQL.Request(ctx, setAzureCloudAccountCustomerAppCredentialsQuery, struct {
		Cloud        Cloud     `json:"azureCloudType"`
		ID           uuid.UUID `json:"appId"`
		Name         string    `json:"appName"`
		SecretKey    string    `json:"appSecretKey"`
		TenantID     uuid.UUID `json:"appTenantId"`
		TenantDomain string    `json:"tenantDomainName"`
	}{Cloud: cloud, ID: appID, Name: appName, TenantID: appTenantID, TenantDomain: appTenantDomain, SecretKey: appSecretKey})
	if err != nil {
		return err
	}

	a.GQL.Log().Printf(log.Debug, "setAzureCloudAccountCustomerAppCredentials(%q, %q, %q, %q, %q, %q): %s", cloud, appID, appName,
		appSecretKey, appTenantID, appTenantDomain, string(buf))

	var payload struct {
		Data struct {
			Result bool `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return err
	}
	if !payload.Data.Result {
		return errors.New("polaris: failed to set azure customer credentials")
	}

	return nil
}
