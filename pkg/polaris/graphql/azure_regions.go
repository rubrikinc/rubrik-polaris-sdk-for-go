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

package graphql

import (
	"errors"
	"strings"
)

// AzureRegion -
type AzureRegion string

const (
	AzureRegionUnknown            AzureRegion = "UNKNOWN_AZURE_REGION"
	AzureRegionAustraliaCentral   AzureRegion = "AUSTRALIACENTRAL"
	AzureRegionAustraliaCentral2  AzureRegion = "AUSTRALIACENTRAL2"
	AzureRegionAustraliaEast      AzureRegion = "AUSTRALIAEAST"
	AzureRegionAustraliaSouthEast AzureRegion = "AUSTRALIASOUTHEAST"
	AzureRegionBrazilSouth        AzureRegion = "BRAZILSOUTH"
	AzureRegionCanadaCentral      AzureRegion = "CANADACENTRAL"
	AzureRegionCanadaEast         AzureRegion = "CANADAEAST"
	AzureRegionCentralIndia       AzureRegion = "CENTRALINDIA"
	AzureRegionCentralUS          AzureRegion = "CENTRALUS"
	AzureRegionChinaEast          AzureRegion = "CHINAEAST"
	AzureRegionChinaEast2         AzureRegion = "CHINAEAST2"
	AzureRegionChinaNorth         AzureRegion = "CHINANORTH"
	AzureRegionChinaNorth2        AzureRegion = "CHINANORTH2"
	AzureRegionEastAsia           AzureRegion = "EASTASIA"
	AzureRegionEastUS             AzureRegion = "EASTUS"
	AzureRegionEastUS2            AzureRegion = "EASTUS2"
	AzureRegionFranceCentral      AzureRegion = "FRANCECENTRAL"
	AzureRegionFranceSouth        AzureRegion = "FRANCESOUTH"
	AzureRegionGermanyNorth       AzureRegion = "GERMANYNORTH"
	AzureRegionGermanyWestCentral AzureRegion = "GERMANYWESTCENTRAL"
	AzureRegionJapanEast          AzureRegion = "JAPANEAST"
	AzureRegionJapanWest          AzureRegion = "JAPANWEST"
	AzureRegionKoreaCentral       AzureRegion = "KOREACENTRAL"
	AzureRegionKoreaSouth         AzureRegion = "KOREASOUTH"
	AzureRegionNorthCentralUs     AzureRegion = "NORTHCENTRALUS"
	AzureRegionNorthEurope        AzureRegion = "NORTHEUROPE"
	AzureRegionNorwayEast         AzureRegion = "NORWAYEAST"
	AzureRegionNorwayWest         AzureRegion = "NORWAYWEST"
	AzureRegionSouthAfricaNorth   AzureRegion = "SOUTHAFRICANORTH"
	AzureRegionSouthAfricaWest    AzureRegion = "SOUTHAFRICAWEST"
	AzureRegionSouthCentralUS     AzureRegion = "SOUTHCENTRALUS"
	AzureRegionSouthEastAsia      AzureRegion = "SOUTHEASTASIA"
	AzureRegionSouthIndia         AzureRegion = "SOUTHINDIA"
	AzureRegionSwitzerlandNorth   AzureRegion = "SWITZERLANDNORTH"
	AzureRegionSwitzerlandWest    AzureRegion = "SWITZERLANDWEST"
	AzureRegionUAECentral         AzureRegion = "UAECENTRAL"
	AzureRegionUAENorth           AzureRegion = "UAENORTH"
	AzureRegionUKSouth            AzureRegion = "UKSOUTH"
	AzureRegionUKWest             AzureRegion = "UKWEST"
	AzureRegionWestCentralUS      AzureRegion = "WESTCENTRALUS"
	AzureRegionWestEurope         AzureRegion = "WESTEUROPE"
	AzureRegionWestIndia          AzureRegion = "WESTINDIA"
	AzureRegionWestUS             AzureRegion = "WESTUS"
	AzureRegionWestUS2            AzureRegion = "WESTUS2"
)

// AzureRegions -
var AzureRegions = []AzureRegion{
	AzureRegionAustraliaCentral,
	AzureRegionAustraliaCentral2,
	AzureRegionAustraliaEast,
	AzureRegionAustraliaSouthEast,
	AzureRegionBrazilSouth,
	AzureRegionCanadaCentral,
	AzureRegionCanadaEast,
	AzureRegionCentralIndia,
	AzureRegionCentralUS,
	AzureRegionChinaEast,
	AzureRegionChinaEast2,
	AzureRegionChinaNorth,
	AzureRegionChinaNorth2,
	AzureRegionEastAsia,
	AzureRegionEastUS,
	AzureRegionEastUS2,
	AzureRegionFranceCentral,
	AzureRegionFranceSouth,
	AzureRegionGermanyNorth,
	AzureRegionGermanyWestCentral,
	AzureRegionJapanEast,
	AzureRegionJapanWest,
	AzureRegionKoreaCentral,
	AzureRegionKoreaSouth,
	AzureRegionNorthCentralUs,
	AzureRegionNorthEurope,
	AzureRegionNorwayEast,
	AzureRegionNorwayWest,
	AzureRegionSouthAfricaNorth,
	AzureRegionSouthAfricaWest,
	AzureRegionSouthCentralUS,
	AzureRegionSouthEastAsia,
	AzureRegionSouthIndia,
	AzureRegionSwitzerlandNorth,
	AzureRegionSwitzerlandWest,
	AzureRegionUAECentral,
	AzureRegionUAENorth,
	AzureRegionUKSouth,
	AzureRegionUKWest,
	AzureRegionWestCentralUS,
	AzureRegionWestEurope,
	AzureRegionWestIndia,
	AzureRegionWestUS,
	AzureRegionWestUS2,
}

// AzureFormatRegion -
func AzureFormatRegion(region AzureRegion) string {
	return strings.ToLower(string(region))
}

// AzureParseRegion -
func AzureParseRegion(region string) (AzureRegion, error) {
	region = strings.ToUpper(region)

	for _, azureRegion := range AzureRegions {
		if AzureRegion(region) == azureRegion {
			return azureRegion, nil
		}
	}

	return AzureRegionUnknown, errors.New("polaris: invalid azure region")
}
