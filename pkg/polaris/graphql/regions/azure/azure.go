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
// DEALINGS IN THE SOFTWARE.package azure

package azure

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	RegionSource  Region = -1
	RegionUnknown Region = iota
	RegionAustraliaCentral
	RegionAustraliaCentral2
	RegionAustraliaEast
	RegionAustraliaSoutheast
	RegionBrazilSouth
	RegionBrazilSoutheast
	RegionCanadaCentral
	RegionCanadaEast
	RegionCentralIndia
	RegionCentralUS
	RegionChinaEast
	RegionChinaEast2
	RegionChinaNorth
	RegionChinaNorth2
	RegionEastAsia
	RegionEastUS
	RegionEastUS2
	RegionFranceCentral
	RegionFranceSouth
	RegionGermanyNorth
	RegionGermanyWestCentral
	RegionIsraelCentral
	RegionItalyNorth
	RegionJapanEast
	RegionJapanWest
	RegionJioIndiaCentral
	RegionJioIndiaWest
	RegionKoreaCentral
	RegionKoreaSouth
	RegionMexicoCentral
	RegionNorthCentralUS
	RegionNorthEurope
	RegionNorwayEast
	RegionNorwayWest
	RegionPolandCentral
	RegionQatarCentral
	RegionSouthAfricaNorth
	RegionSouthAfricaWest
	RegionSouthCentralUS
	RegionSoutheastAsia
	RegionSouthIndia
	RegionSwedenCentral
	RegionSwitzerlandNorth
	RegionSwitzerlandWest
	RegionUAECentral
	RegionUAENorth
	RegionUKSouth
	RegionUKWest
	RegionUSDoDCentral
	RegionUSDoDEast
	RegionUSGovArizona
	RegionUSGovTexas
	RegionUSGovVirginia
	RegionWestCentralUS
	RegionWestEurope
	RegionWestIndia
	RegionWestUS
	RegionWestUS2
	RegionWestUS3
	RegionSpainCentral
	RegionSwedenSouth
)

// Region represents an Azure region in RSC. When reading a Region from a JSON
// document or writing a Region to a JSON document, use one of the specialized
// region enum types, to guarantee that the correct enum value is used.
type Region int

// Name returns the name of the region.
func (region Region) Name() string {
	return regionInfoMap[region].name
}

// DisplayName returns the display name of the region.
func (region Region) DisplayName() string {
	return regionInfoMap[region].displayName
}

// RegionalDisplayName returns the regional display name of the region.
func (region Region) RegionalDisplayName() string {
	return regionInfoMap[region].regionalDisplayName
}

// ToCloudAccountRegionEnum returns the RSC GraphQL AzureCloudAccountRegion enum
// value for the region.
func (region Region) ToCloudAccountRegionEnum() CloudAccountRegionEnum {
	return CloudAccountRegionEnum{Region: region}
}

// ToCloudAccountRegionEnumPtr returns the RSC GraphQL AzureCloudAccountRegion
// enum value for the region as a pointer. If the region is unknown, nil is
// returned.
func (region Region) ToCloudAccountRegionEnumPtr() *CloudAccountRegionEnum {
	if region == RegionUnknown {
		return nil
	}
	return &CloudAccountRegionEnum{Region: region}
}

// ToNativeRegionEnum returns the RSC GraphQL AzureNativeRegion enum value for
// the region.
func (region Region) ToNativeRegionEnum() NativeRegionEnum {
	return NativeRegionEnum{Region: region}
}

// ToNativeRegionEnumPtr returns the RSC GraphQL AzureNativeRegion enum value
// for the region as a pointer. If the region is unknown, nil is returned.
func (region Region) ToNativeRegionEnumPtr() *NativeRegionEnum {
	if region == RegionUnknown {
		return nil
	}
	return &NativeRegionEnum{Region: region}
}

// ToRegionEnum returns the RSC GraphQL AzureRegion enum value for the region.
func (region Region) ToRegionEnum() RegionEnum {
	return RegionEnum{Region: region}
}

// ToRegionEnumPtr returns the RSC GraphQL AzureRegion enum value for the region
// as a pointer. If the region is unknown, nil is returned.
func (region Region) ToRegionEnumPtr() *RegionEnum {
	if region == RegionUnknown {
		return nil
	}
	return &RegionEnum{Region: region}
}

// ToRegionForReplicationEnum returns the RSC GraphQL AzureRegionForReplication enum value for the region.
func (region Region) ToRegionForReplicationEnum() RegionForReplicationEnum {
	return RegionForReplicationEnum{Region: region}
}

// String returns the name of the region.
func (region Region) String() string {
	return region.Name()
}

const (
	FromAny                      = iota // Parse the value as any of the below formats.
	FromCloudAccountRegionEnum          // Parse the value as a GraphQL AzureCloudAccountRegion enum value.
	FromDisplayName                     // Parse the value as a region display name.
	FromName                            // Parse the value as a region name.
	FromNativeRegionEnum                // Parse the value as a GraphQL AzureNativeRegion enum value.
	FromRegionalDisplayName             // Parse the value as a region regional display name.
	FromRegionEnum                      // Parse the value as a GraphQL AzureRegion enum value.
	FromRegionForReplicationEnum        // Parse the value as a GraphQL AzureRegionForReplication enum value.
)

// RegionFrom parses the value as a region identifier in the specified format.
// If the value isn't recognized, RegionUnknown is returned.
func RegionFrom(value string, valueFormat int) Region {
	if value == "" || value == "n/a" {
		return RegionUnknown
	}
	for r, info := range regionInfoMap {
		switch {
		case (valueFormat == FromAny || valueFormat == FromName) && info.name == value:
			return r
		case (valueFormat == FromAny || valueFormat == FromCloudAccountRegionEnum) && info.cloudAccountRegionEnum == value:
			return r
		case (valueFormat == FromAny || valueFormat == FromNativeRegionEnum) && info.nativeRegionEnum == value:
			return r
		case (valueFormat == FromAny || valueFormat == FromRegionEnum) && info.regionEnum == value:
			return r
		case (valueFormat == FromAny || valueFormat == FromDisplayName) && info.displayName == value:
			return r
		case (valueFormat == FromAny || valueFormat == FromRegionalDisplayName) && info.regionalDisplayName == value:
			return r
		case (valueFormat == FromAny || valueFormat == FromRegionForReplicationEnum) && info.regionForReplicationEnum == value:
			return r
		}
	}

	return RegionUnknown
}

// RegionFromAny parses the value as any region identifier that matches.
func RegionFromAny(value string) Region {
	return RegionFrom(value, FromAny)
}

// RegionFromName parses the value as a region name.
func RegionFromName(value string) Region {
	return RegionFrom(value, FromName)
}

// RegionFromDisplayName parses the value as a region display name.
func RegionFromDisplayName(value string) Region {
	return RegionFrom(value, FromDisplayName)
}

// RegionFromRegionalDisplayName parses the value as a region regional display
// name.
func RegionFromRegionalDisplayName(value string) Region {
	return RegionFrom(value, FromRegionalDisplayName)
}

// RegionFromCloudAccountRegionEnum parses the value as a GraphQL
// AzureCloudAccountRegion enum value.
func RegionFromCloudAccountRegionEnum(value string) Region {
	return RegionFrom(value, FromCloudAccountRegionEnum)
}

// RegionFromNativeRegionEnum parses the value as a GraphQL AzureNativeRegion
// enum.
func RegionFromNativeRegionEnum(value string) Region {
	return RegionFrom(value, FromNativeRegionEnum)
}

// RegionFromRegionEnum parses the value as a GraphQL AzureRegion enum value.
func RegionFromRegionEnum(value string) Region {
	return RegionFrom(value, FromRegionEnum)
}

// RegionFromRegionForReplicationEnum parses the value as a GraphQL
// AzureRegionForReplication enum value.
func RegionFromRegionForReplicationEnum(value string) Region {
	return RegionFrom(value, FromRegionForReplicationEnum)
}

// RegionEnum represents the GraphQL AzureRegion enum type.
type RegionEnum struct{ Region }

// MarshalJSON returns the region as a JSON string.
func (region RegionEnum) MarshalJSON() ([]byte, error) {
	return json.Marshal(regionInfoMap[region.Region].regionEnum)
}

// UnmarshalJSON parses the region from a JSON string.
func (region *RegionEnum) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	region.Region = RegionFromRegionEnum(s)
	return nil
}

// CloudAccountRegionEnum represents the GraphQL AzureCloudAccountRegion enum
// type.
type CloudAccountRegionEnum struct{ Region }

// MarshalJSON returns the region as a JSON string.
func (region CloudAccountRegionEnum) MarshalJSON() ([]byte, error) {
	return json.Marshal(regionInfoMap[region.Region].cloudAccountRegionEnum)
}

// UnmarshalJSON parses the region from a JSON string.
func (region *CloudAccountRegionEnum) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	region.Region = RegionFromCloudAccountRegionEnum(s)
	return nil
}

// NativeRegionEnum represents the GraphQL AzureNativeRegion enum type.
type NativeRegionEnum struct{ Region }

// MarshalJSON returns the region as a JSON string.
func (region NativeRegionEnum) MarshalJSON() ([]byte, error) {
	return json.Marshal(regionInfoMap[region.Region].nativeRegionEnum)
}

// UnmarshalJSON parses the region from a JSON string.
func (region *NativeRegionEnum) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	region.Region = RegionFromNativeRegionEnum(s)
	return nil
}

// RegionForReplicationEnum represents the GraphQL AzureRegionForReplication enum type.
type RegionForReplicationEnum struct{ Region }

// MarshalJSON returns the region as a JSON string.
func (region RegionForReplicationEnum) MarshalJSON() ([]byte, error) {
	return json.Marshal(regionInfoMap[region.Region].regionForReplicationEnum)
}

// UnmarshalJSON parses the region from a JSON string.
func (region *RegionForReplicationEnum) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	region.Region = RegionFromRegionForReplicationEnum(s)
	return nil
}

// AllRegionNames returns all the recognized region names.
func AllRegionNames() []string {
	regions := make([]string, 0, len(regionInfoMap))
	for _, info := range regionInfoMap {
		if info.name != "" {
			regions = append(regions, info.name)
		}
	}

	return regions
}

var regionInfoMap = map[Region]struct {
	name                     string
	displayName              string
	regionalDisplayName      string
	regionEnum               string
	cloudAccountRegionEnum   string
	nativeRegionEnum         string
	regionForReplicationEnum string
}{
	RegionSource: {
		name:                     "n/a",
		displayName:              "Same as source",
		regionalDisplayName:      "n/a",
		regionEnum:               "n/a",
		cloudAccountRegionEnum:   "n/a",
		nativeRegionEnum:         "n/a",
		regionForReplicationEnum: "SOURCE_REGION",
	},
	RegionUnknown: {
		name:                     "",
		displayName:              "<Unknown>",
		regionalDisplayName:      "<Unknown>",
		regionEnum:               "UNKNOWN_AZURE_REGION",
		cloudAccountRegionEnum:   "UNKNOWN_AZURE_REGION",
		nativeRegionEnum:         "NOT_SPECIFIED",
		regionForReplicationEnum: "NOT_DEFINED",
	},
	RegionAustraliaCentral: {
		name:                     "australiacentral",
		displayName:              "Australia Central",
		regionalDisplayName:      "(Asia Pacific) Australia Central",
		regionEnum:               "AUSTRALIA_CENTRAL",
		cloudAccountRegionEnum:   "AUSTRALIACENTRAL",
		nativeRegionEnum:         "AUSTRALIA_CENTRAL",
		regionForReplicationEnum: "AUSTRALIA_CENTRAL",
	},
	RegionAustraliaCentral2: {
		name:                     "australiacentral2",
		displayName:              "Australia Central 2",
		regionalDisplayName:      "(Asia Pacific) Australia Central 2",
		regionEnum:               "AUSTRALIA_CENTRAL2",
		cloudAccountRegionEnum:   "AUSTRALIACENTRAL2",
		nativeRegionEnum:         "AUSTRALIA_CENTRAL2",
		regionForReplicationEnum: "AUSTRALIA_CENTRAL2",
	},
	RegionAustraliaEast: {
		name:                     "australiaeast",
		displayName:              "Australia East",
		regionalDisplayName:      "(Asia Pacific) Australia East",
		regionEnum:               "AUSTRALIA_EAST",
		cloudAccountRegionEnum:   "AUSTRALIAEAST",
		nativeRegionEnum:         "AUSTRALIA_EAST",
		regionForReplicationEnum: "AUSTRALIA_EAST",
	},
	RegionAustraliaSoutheast: {
		name:                     "australiasoutheast",
		displayName:              "Australia Southeast",
		regionalDisplayName:      "(Asia Pacific) Australia Southeast",
		regionEnum:               "AUSTRALIA_SOUTHEAST",
		cloudAccountRegionEnum:   "AUSTRALIASOUTHEAST",
		nativeRegionEnum:         "AUSTRALIA_SOUTHEAST",
		regionForReplicationEnum: "AUSTRALIA_SOUTHEAST",
	},
	RegionBrazilSouth: {
		name:                     "brazilsouth",
		displayName:              "Brazil South",
		regionalDisplayName:      "(South America) Brazil South",
		regionEnum:               "BRAZIL_SOUTH",
		cloudAccountRegionEnum:   "BRAZILSOUTH",
		nativeRegionEnum:         "BRAZIL_SOUTH",
		regionForReplicationEnum: "BRAZIL_SOUTH",
	},
	RegionBrazilSoutheast: {
		name:                     "brazilsoutheast",
		displayName:              "Brazil Southeast",
		regionalDisplayName:      "(South America) Brazil Southeast",
		regionEnum:               "BRAZIL_SOUTHEAST",
		cloudAccountRegionEnum:   "BRAZILSOUTHEAST",
		nativeRegionEnum:         "BRAZIL_SOUTHEAST",
		regionForReplicationEnum: "BRAZIL_SOUTHEAST",
	},
	RegionCanadaCentral: {
		name:                     "canadacentral",
		displayName:              "Canada Central",
		regionalDisplayName:      "(Canada) Canada Central",
		regionEnum:               "CANADA_CENTRAL",
		cloudAccountRegionEnum:   "CANADACENTRAL",
		nativeRegionEnum:         "CANADA_CENTRAL",
		regionForReplicationEnum: "CANADA_CENTRAL",
	},
	RegionCanadaEast: {
		name:                     "canadaeast",
		displayName:              "Canada East",
		regionalDisplayName:      "(Canada) Canada East",
		regionEnum:               "CANADA_EAST",
		cloudAccountRegionEnum:   "CANADAEAST",
		nativeRegionEnum:         "CANADA_EAST",
		regionForReplicationEnum: "CANADA_EAST",
	},
	RegionCentralIndia: {
		name:                     "centralindia",
		displayName:              "Central India",
		regionalDisplayName:      "(Asia Pacific) Central India",
		regionEnum:               "INDIA_CENTRAL",
		cloudAccountRegionEnum:   "CENTRALINDIA",
		nativeRegionEnum:         "CENTRAL_INDIA",
		regionForReplicationEnum: "CENTRAL_INDIA",
	},
	RegionCentralUS: {
		name:                     "centralus",
		displayName:              "Central US",
		regionalDisplayName:      "(US) Central US",
		regionEnum:               "US_CENTRAL",
		cloudAccountRegionEnum:   "CENTRALUS",
		nativeRegionEnum:         "CENTRAL_US",
		regionForReplicationEnum: "CENTRAL_US",
	},
	RegionChinaEast: {
		name:                     "chinaeast",
		displayName:              "China East",
		regionalDisplayName:      "(China) China East",
		regionEnum:               "CHINA_EAST",
		cloudAccountRegionEnum:   "CHINAEAST",
		nativeRegionEnum:         "CHINA_EAST",
		regionForReplicationEnum: "CHINA_EAST",
	},
	RegionChinaEast2: {
		name:                     "chinaeast2",
		displayName:              "China East 2",
		regionalDisplayName:      "(China) China East 2",
		regionEnum:               "CHINA_EAST2",
		cloudAccountRegionEnum:   "CHINAEAST2",
		nativeRegionEnum:         "CHINA_EAST2",
		regionForReplicationEnum: "CHINA_EAST2",
	},
	RegionChinaNorth: {
		name:                     "chinanorth",
		displayName:              "China North",
		regionalDisplayName:      "(China) China North",
		regionEnum:               "CHINA_NORTH",
		cloudAccountRegionEnum:   "CHINANORTH",
		nativeRegionEnum:         "CHINA_NORTH",
		regionForReplicationEnum: "CHINA_NORTH",
	},
	RegionChinaNorth2: {
		name:                     "chinanorth2",
		displayName:              "China North 2",
		regionalDisplayName:      "(China) China North 2",
		regionEnum:               "CHINA_NORTH2",
		cloudAccountRegionEnum:   "CHINANORTH2",
		nativeRegionEnum:         "CHINA_NORTH2",
		regionForReplicationEnum: "CHINA_NORTH2",
	},
	RegionEastAsia: {
		name:                     "eastasia",
		displayName:              "East Asia",
		regionalDisplayName:      "(Asia Pacific) East Asia",
		regionEnum:               "ASIA_EAST",
		cloudAccountRegionEnum:   "EASTASIA",
		nativeRegionEnum:         "EAST_ASIA",
		regionForReplicationEnum: "EAST_ASIA",
	},
	RegionEastUS: {
		name:                     "eastus",
		displayName:              "East US",
		regionalDisplayName:      "(US) East US",
		regionEnum:               "US_EAST",
		cloudAccountRegionEnum:   "EASTUS",
		nativeRegionEnum:         "EAST_US",
		regionForReplicationEnum: "EAST_US",
	},
	RegionEastUS2: {
		name:                     "eastus2",
		displayName:              "East US 2",
		regionalDisplayName:      "(US) East US 2",
		regionEnum:               "US_EAST2",
		cloudAccountRegionEnum:   "EASTUS2",
		nativeRegionEnum:         "EAST_US2",
		regionForReplicationEnum: "EAST_US2",
	},
	RegionFranceCentral: {
		name:                     "francecentral",
		displayName:              "France Central",
		regionalDisplayName:      "(Europe) France Central",
		regionEnum:               "FRANCE_CENTRAL",
		cloudAccountRegionEnum:   "FRANCECENTRAL",
		nativeRegionEnum:         "FRANCE_CENTRAL",
		regionForReplicationEnum: "FRANCE_CENTRAL",
	},
	RegionFranceSouth: {
		name:                     "francesouth",
		displayName:              "France South",
		regionalDisplayName:      "(Europe) France South",
		regionEnum:               "FRANCE_SOUTH",
		cloudAccountRegionEnum:   "FRANCESOUTH",
		nativeRegionEnum:         "FRANCE_SOUTH",
		regionForReplicationEnum: "FRANCE_SOUTH",
	},
	RegionGermanyNorth: {
		name:                     "germanynorth",
		displayName:              "Germany North",
		regionalDisplayName:      "(Europe) Germany North",
		regionEnum:               "GERMANY_NORTH",
		cloudAccountRegionEnum:   "GERMANYNORTH",
		nativeRegionEnum:         "GERMANY_NORTH",
		regionForReplicationEnum: "GERMANY_NORTH",
	},
	RegionGermanyWestCentral: {
		name:                     "germanywestcentral",
		displayName:              "Germany West Central",
		regionalDisplayName:      "(Europe) Germany West Central",
		regionEnum:               "GERMANY_WEST_CENTRAL",
		cloudAccountRegionEnum:   "GERMANYWESTCENTRAL",
		nativeRegionEnum:         "GERMANY_WEST_CENTRAL",
		regionForReplicationEnum: "GERMANY_WEST_CENTRAL",
	},
	RegionIsraelCentral: {
		name:                     "israelcentral",
		displayName:              "Israel Central",
		regionalDisplayName:      "(Middle East) Israel Central",
		regionEnum:               "ISRAEL_CENTRAL",
		cloudAccountRegionEnum:   "ISRAELCENTRAL",
		nativeRegionEnum:         "ISRAEL_CENTRAL",
		regionForReplicationEnum: "ISRAEL_CENTRAL",
	},
	RegionItalyNorth: {
		name:                     "italynorth",
		displayName:              "Italy North",
		regionalDisplayName:      "(Europe) Italy North",
		regionEnum:               "ITALY_NORTH",
		cloudAccountRegionEnum:   "ITALYNORTH",
		nativeRegionEnum:         "ITALY_NORTH",
		regionForReplicationEnum: "ITALY_NORTH",
	},
	RegionJapanEast: {
		name:                     "japaneast",
		displayName:              "Japan East",
		regionalDisplayName:      "(Asia Pacific) Japan East",
		regionEnum:               "JAPAN_EAST",
		cloudAccountRegionEnum:   "JAPANEAST",
		nativeRegionEnum:         "JAPAN_EAST",
		regionForReplicationEnum: "JAPAN_EAST",
	},
	RegionJapanWest: {
		name:                     "japanwest",
		displayName:              "Japan West",
		regionalDisplayName:      "(Asia Pacific) Japan West",
		regionEnum:               "JAPAN_WEST",
		cloudAccountRegionEnum:   "JAPANWEST",
		nativeRegionEnum:         "JAPAN_WEST",
		regionForReplicationEnum: "JAPAN_WEST",
	},
	RegionJioIndiaCentral: {
		name:                     "jioindiacentral",
		displayName:              "Jio India Central",
		regionalDisplayName:      "(Asia Pacific) Jio India Central",
		regionEnum:               "JIO_INDIA_CENTRAL",
		cloudAccountRegionEnum:   "JIOINDIACENTRAL",
		nativeRegionEnum:         "JIO_INDIA_CENTRAL",
		regionForReplicationEnum: "n/a",
	},
	RegionJioIndiaWest: {
		name:                     "jioindiawest",
		displayName:              "Jio India West",
		regionalDisplayName:      "(Asia Pacific) Jio India West",
		regionEnum:               "JIO_INDIA_WEST",
		cloudAccountRegionEnum:   "JIOINDIAWEST",
		nativeRegionEnum:         "JIO_INDIA_WEST",
		regionForReplicationEnum: "n/a",
	},
	RegionKoreaCentral: {
		name:                     "koreacentral",
		displayName:              "Korea Central",
		regionalDisplayName:      "(Asia Pacific) Korea Central",
		regionEnum:               "KOREA_CENTRAL",
		cloudAccountRegionEnum:   "KOREACENTRAL",
		nativeRegionEnum:         "KOREA_CENTRAL",
		regionForReplicationEnum: "KOREA_CENTRAL",
	},
	RegionKoreaSouth: {
		name:                     "koreasouth",
		displayName:              "Korea South",
		regionalDisplayName:      "(Asia Pacific) Korea South",
		regionEnum:               "KOREA_SOUTH",
		cloudAccountRegionEnum:   "KOREASOUTH",
		nativeRegionEnum:         "KOREA_SOUTH",
		regionForReplicationEnum: "KOREA_SOUTH",
	},
	RegionMexicoCentral: {
		name:                     "mexicocentral",
		displayName:              "Mexico Central",
		regionalDisplayName:      "(Mexico) Mexico Central",
		regionEnum:               "MEXICO_CENTRAL",
		cloudAccountRegionEnum:   "MEXICOCENTRAL",
		nativeRegionEnum:         "MEXICO_CENTRAL",
		regionForReplicationEnum: "MEXICO_CENTRAL",
	},
	RegionNorthCentralUS: {
		name:                     "northcentralus",
		displayName:              "North Central US",
		regionalDisplayName:      "(US) North Central US",
		regionEnum:               "US_NORTH_CENTRAL",
		cloudAccountRegionEnum:   "NORTHCENTRALUS",
		nativeRegionEnum:         "NORTH_CENTRAL_US",
		regionForReplicationEnum: "NORTH_CENTRAL_US",
	},
	RegionNorthEurope: {
		name:                     "northeurope",
		displayName:              "North Europe",
		regionalDisplayName:      "(Europe) North Europe",
		regionEnum:               "EUROPE_NORTH",
		cloudAccountRegionEnum:   "NORTHEUROPE",
		nativeRegionEnum:         "NORTH_EUROPE",
		regionForReplicationEnum: "NORTH_EUROPE",
	},
	RegionNorwayEast: {
		name:                     "norwayeast",
		displayName:              "Norway East",
		regionalDisplayName:      "(Europe) Norway East",
		regionEnum:               "NORWAY_EAST",
		cloudAccountRegionEnum:   "NORWAYEAST",
		nativeRegionEnum:         "NORWAY_EAST",
		regionForReplicationEnum: "NORWAY_EAST",
	},
	RegionNorwayWest: {
		name:                     "norwaywest",
		displayName:              "Norway West",
		regionalDisplayName:      "(Europe) Norway West",
		regionEnum:               "NORWAY_WEST",
		cloudAccountRegionEnum:   "NORWAYWEST",
		nativeRegionEnum:         "NORWAY_WEST",
		regionForReplicationEnum: "NORWAY_WEST",
	},
	RegionPolandCentral: {
		name:                     "polandcentral",
		displayName:              "Poland Central",
		regionalDisplayName:      "(Europe) Poland Central",
		regionEnum:               "POLAND_CENTRAL",
		cloudAccountRegionEnum:   "POLANDCENTRAL",
		nativeRegionEnum:         "POLAND_CENTRAL",
		regionForReplicationEnum: "POLAND_CENTRAL",
	},
	RegionQatarCentral: {
		name:                     "qatarcentral",
		displayName:              "Qatar Central",
		regionalDisplayName:      "(Middle East) Qatar Central",
		regionEnum:               "QATAR_CENTRAL",
		cloudAccountRegionEnum:   "QATARCENTRAL",
		nativeRegionEnum:         "QATAR_CENTRAL",
		regionForReplicationEnum: "QATAR_CENTRAL",
	},
	RegionSouthAfricaNorth: {
		name:                     "southafricanorth",
		displayName:              "South Africa North",
		regionalDisplayName:      "(Africa) South Africa North",
		regionEnum:               "SOUTH_AFRICA_NORTH",
		cloudAccountRegionEnum:   "SOUTHAFRICANORTH",
		nativeRegionEnum:         "SOUTH_AFRICA_NORTH",
		regionForReplicationEnum: "SOUTH_AFRICA_NORTH",
	},
	RegionSouthAfricaWest: {
		name:                     "southafricawest",
		displayName:              "South Africa West",
		regionalDisplayName:      "(Africa) South Africa West",
		regionEnum:               "SOUTH_AFRICA_WEST",
		cloudAccountRegionEnum:   "SOUTHAFRICAWEST",
		nativeRegionEnum:         "SOUTH_AFRICA_WEST",
		regionForReplicationEnum: "SOUTH_AFRICA_WEST",
	},
	RegionSouthCentralUS: {
		name:                     "southcentralus",
		displayName:              "South Central US",
		regionalDisplayName:      "(US) South Central US",
		regionEnum:               "US_SOUTH_CENTRAL",
		cloudAccountRegionEnum:   "SOUTHCENTRALUS",
		nativeRegionEnum:         "SOUTH_CENTRAL_US",
		regionForReplicationEnum: "SOUTH_CENTRAL_US",
	},
	RegionSoutheastAsia: {
		name:                     "southeastasia",
		displayName:              "Southeast Asia",
		regionalDisplayName:      "(Asia Pacific) Southeast Asia",
		regionEnum:               "ASIA_SOUTHEAST",
		cloudAccountRegionEnum:   "SOUTHEASTASIA",
		nativeRegionEnum:         "SOUTHEAST_ASIA",
		regionForReplicationEnum: "SOUTHEAST_ASIA",
	},
	RegionSouthIndia: {
		name:                     "southindia",
		displayName:              "South India",
		regionalDisplayName:      "(Asia Pacific) South India",
		regionEnum:               "INDIA_SOUTH",
		cloudAccountRegionEnum:   "SOUTHINDIA",
		nativeRegionEnum:         "SOUTH_INDIA",
		regionForReplicationEnum: "SOUTH_INDIA",
	},
	RegionSwedenCentral: {
		name:                     "swedencentral",
		displayName:              "Sweden Central",
		regionalDisplayName:      "(Europe) Sweden Central",
		regionEnum:               "SWEDEN_CENTRAL",
		cloudAccountRegionEnum:   "SWEDENCENTRAL",
		nativeRegionEnum:         "SWEDEN_CENTRAL",
		regionForReplicationEnum: "SWEDEN_CENTRAL",
	},
	RegionSwitzerlandNorth: {
		name:                     "switzerlandnorth",
		displayName:              "Switzerland North",
		regionalDisplayName:      "(Europe) Switzerland North",
		regionEnum:               "SWITZERLAND_NORTH",
		cloudAccountRegionEnum:   "SWITZERLANDNORTH",
		nativeRegionEnum:         "SWITZERLAND_NORTH",
		regionForReplicationEnum: "SWITZERLAND_NORTH",
	},
	RegionSwitzerlandWest: {
		name:                     "switzerlandwest",
		displayName:              "Switzerland West",
		regionalDisplayName:      "(Europe) Switzerland West",
		regionEnum:               "SWITZERLAND_WEST",
		cloudAccountRegionEnum:   "SWITZERLANDWEST",
		nativeRegionEnum:         "SWITZERLAND_WEST",
		regionForReplicationEnum: "SWITZERLAND_WEST",
	},
	RegionUAECentral: {
		name:                     "uaecentral",
		displayName:              "UAE Central",
		regionalDisplayName:      "(Middle East) UAE Central",
		regionEnum:               "UAE_CENTRAL",
		cloudAccountRegionEnum:   "UAECENTRAL",
		nativeRegionEnum:         "UAE_CENTRAL",
		regionForReplicationEnum: "UAE_CENTRAL",
	},
	RegionUAENorth: {
		name:                     "uaenorth",
		displayName:              "UAE North",
		regionalDisplayName:      "(Middle East) UAE North",
		regionEnum:               "UAE_NORTH",
		cloudAccountRegionEnum:   "UAENORTH",
		nativeRegionEnum:         "UAE_NORTH",
		regionForReplicationEnum: "UAE_NORTH",
	},
	RegionUKSouth: {
		name:                     "uksouth",
		displayName:              "UK South",
		regionalDisplayName:      "(Europe) UK South",
		regionEnum:               "UK_SOUTH",
		cloudAccountRegionEnum:   "UKSOUTH",
		nativeRegionEnum:         "UK_SOUTH",
		regionForReplicationEnum: "UK_SOUTH",
	},
	RegionUKWest: {
		name:                     "ukwest",
		displayName:              "UK West",
		regionalDisplayName:      "(Europe) UK West",
		regionEnum:               "UK_WEST",
		cloudAccountRegionEnum:   "UKWEST",
		nativeRegionEnum:         "UK_WEST",
		regionForReplicationEnum: "UK_WEST",
	},
	RegionUSDoDCentral: {
		name:                     "usdodcentral",
		displayName:              "US DoD Central",
		regionalDisplayName:      "(US Gov) US DoD Central",
		regionEnum:               "GOV_US_DOD_CENTRAL",
		cloudAccountRegionEnum:   "USDODCENTRAL",
		nativeRegionEnum:         "US_DOD_CENTRAL",
		regionForReplicationEnum: "n/a",
	},
	RegionUSDoDEast: {
		name:                     "usdodeast",
		displayName:              "US DoD East",
		regionalDisplayName:      "(US Gov) US DoD East",
		regionEnum:               "GOV_US_DOD_EAST",
		cloudAccountRegionEnum:   "USDODEAST",
		nativeRegionEnum:         "US_DOD_EAST",
		regionForReplicationEnum: "n/a",
	},
	RegionUSGovArizona: {
		name:                     "usgovarizona",
		displayName:              "US Gov Arizona",
		regionalDisplayName:      "(US Gov) US Gov Arizona",
		regionEnum:               "GOV_US_ARIZONA",
		cloudAccountRegionEnum:   "USGOVARIZONA",
		nativeRegionEnum:         "US_GOV_ARIZONA",
		regionForReplicationEnum: "US_GOV_ARIZONA",
	},
	RegionUSGovTexas: {
		name:                     "usgovtexas",
		displayName:              "US Gov Texas",
		regionalDisplayName:      "(US Gov) US Gov Texas",
		regionEnum:               "GOV_US_TEXAS",
		cloudAccountRegionEnum:   "USGOVTEXAS",
		nativeRegionEnum:         "US_GOV_TEXAS",
		regionForReplicationEnum: "US_GOV_TEXAS",
	},
	RegionUSGovVirginia: {
		name:                     "usgovvirginia",
		displayName:              "US Gov Virginia",
		regionalDisplayName:      "(US Gov) US Gov Virginia",
		regionEnum:               "GOV_US_VIRGINIA",
		cloudAccountRegionEnum:   "USGOVVIRGINIA",
		nativeRegionEnum:         "US_GOV_VIRGINIA",
		regionForReplicationEnum: "US_GOV_VIRGINIA",
	},
	RegionWestCentralUS: {
		name:                     "westcentralus",
		displayName:              "West Central US",
		regionalDisplayName:      "(US) West Central US",
		regionEnum:               "US_WEST_CENTRAL",
		cloudAccountRegionEnum:   "WESTCENTRALUS",
		nativeRegionEnum:         "WEST_CENTRAL_US",
		regionForReplicationEnum: "WEST_CENTRAL_US",
	},
	RegionWestEurope: {
		name:                     "westeurope",
		displayName:              "West Europe",
		regionalDisplayName:      "(Europe) West Europe",
		regionEnum:               "EUROPE_WEST",
		cloudAccountRegionEnum:   "WESTEUROPE",
		nativeRegionEnum:         "WEST_EUROPE",
		regionForReplicationEnum: "WEST_EUROPE",
	},
	RegionWestIndia: {
		name:                     "westindia",
		displayName:              "West India",
		regionalDisplayName:      "(Asia Pacific) West India",
		regionEnum:               "INDIA_WEST",
		cloudAccountRegionEnum:   "WESTINDIA",
		nativeRegionEnum:         "WEST_INDIA",
		regionForReplicationEnum: "WEST_INDIA",
	},
	RegionWestUS: {
		name:                     "westus",
		displayName:              "West US",
		regionalDisplayName:      "(US) West US",
		regionEnum:               "US_WEST",
		cloudAccountRegionEnum:   "WESTUS",
		nativeRegionEnum:         "WEST_US",
		regionForReplicationEnum: "WEST_US",
	},
	RegionWestUS2: {
		name:                     "westus2",
		displayName:              "West US 2",
		regionalDisplayName:      "(US) West US 2",
		regionEnum:               "US_WEST2",
		cloudAccountRegionEnum:   "WESTUS2",
		nativeRegionEnum:         "WEST_US2",
		regionForReplicationEnum: "WEST_US2",
	},
	RegionWestUS3: {
		name:                     "westus3",
		displayName:              "West US 3",
		regionalDisplayName:      "(US) West US 3",
		regionEnum:               "WEST_US3",
		cloudAccountRegionEnum:   "WESTUS3",
		nativeRegionEnum:         "WEST_US3",
		regionForReplicationEnum: "WEST_US3",
	},
	RegionSpainCentral: {
		name:                     "n/a",
		displayName:              "Spain Central",
		regionalDisplayName:      "n/a",
		regionEnum:               "n/a",
		cloudAccountRegionEnum:   "n/a",
		nativeRegionEnum:         "n/a",
		regionForReplicationEnum: "SPAIN_CENTRAL",
	},
	RegionSwedenSouth: {
		name:                     "n/a",
		displayName:              "Sweden South",
		regionalDisplayName:      "n/a",
		regionEnum:               "n/a",
		cloudAccountRegionEnum:   "n/a",
		nativeRegionEnum:         "n/a",
		regionForReplicationEnum: "SWEDEN_SOUTH",
	},
}

// Deprecated: use Region.Name.
func FormatRegion(region Region) string {
	return region.Name()
}

// Deprecated: no replacement.
func FormatRegions(regions []Region) []string {
	regs := make([]string, 0, len(regions))
	for _, region := range regions {
		regs = append(regs, region.Name())
	}

	return regs
}

// Deprecated: no replacement.
var validRegions = map[Region]struct{}{
	RegionAustraliaCentral:   {},
	RegionAustraliaCentral2:  {},
	RegionAustraliaEast:      {},
	RegionAustraliaSoutheast: {},
	RegionBrazilSouth:        {},
	RegionBrazilSoutheast:    {},
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
	RegionIsraelCentral:      {},
	RegionItalyNorth:         {},
	RegionJapanEast:          {},
	RegionJapanWest:          {},
	RegionJioIndiaCentral:    {},
	RegionJioIndiaWest:       {},
	RegionKoreaCentral:       {},
	RegionKoreaSouth:         {},
	RegionMexicoCentral:      {},
	RegionNorthCentralUS:     {},
	RegionNorthEurope:        {},
	RegionNorwayEast:         {},
	RegionNorwayWest:         {},
	RegionPolandCentral:      {},
	RegionQatarCentral:       {},
	RegionSouthAfricaNorth:   {},
	RegionSouthAfricaWest:    {},
	RegionSouthCentralUS:     {},
	RegionSoutheastAsia:      {},
	RegionSouthIndia:         {},
	RegionSwedenCentral:      {},
	RegionSwitzerlandNorth:   {},
	RegionSwitzerlandWest:    {},
	RegionUAECentral:         {},
	RegionUAENorth:           {},
	RegionUKSouth:            {},
	RegionUKWest:             {},
	RegionUSDoDCentral:       {},
	RegionUSDoDEast:          {},
	RegionUSGovArizona:       {},
	RegionUSGovTexas:         {},
	RegionUSGovVirginia:      {},
	RegionWestCentralUS:      {},
	RegionWestEurope:         {},
	RegionWestIndia:          {},
	RegionWestUS:             {},
	RegionWestUS2:            {},
	RegionWestUS3:            {},
}

// Deprecated: use RegionFromName or RegionFromCloudAccountRegionEnum.
func ParseRegion(value string) (Region, error) {
	// Polaris region name.
	region := RegionFromCloudAccountRegionEnum(value)
	if _, ok := validRegions[region]; ok {
		return region, nil
	}

	// Azure region name.
	region = RegionFromName(value)
	if _, ok := validRegions[region]; ok {
		return region, nil
	}

	return RegionUnknown, errors.New("invalid azure region")
}

// Deprecated: no replacement.
func ParseRegions(values []string) ([]Region, error) {
	regions := make([]Region, 0, len(values))

	for _, r := range values {
		region, err := ParseRegion(r)
		if err != nil {
			return nil, fmt.Errorf("failed to parse region: %v", err)
		}

		regions = append(regions, region)
	}

	return regions, nil
}

// Deprecated: use RegionFromName.
func ParseRegionNoValidation(value string) Region {
	return RegionFromName(value)
}

// Deprecated: no replacement.
func ParseRegionsNoValidation(values []string) []Region {
	regions := make([]Region, 0, len(values))
	for _, value := range values {
		regions = append(regions, RegionFromName(value))
	}

	return regions
}
