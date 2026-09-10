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
)

const (
	RegionSource  Region = -1
	RegionUnknown Region = iota
	RegionAustraliaCentral
	RegionAustraliaCentral2
	RegionAustraliaEast
	RegionAustraliaSoutheast
	RegionAustriaEast
	RegionBelgiumCentral
	RegionBrazilSouth
	RegionBrazilSoutheast
	RegionCanadaCentral
	RegionCanadaEast
	RegionCentralIndia
	RegionCentralUS
	RegionChileCentral
	RegionChinaEast
	RegionChinaEast2
	RegionChinaNorth
	RegionChinaNorth2
	RegionEastAsia
	RegionEastUS
	RegionEastUS2
	RegionFranceCentral
	RegionFranceSouth
	RegionGermanyCentral
	RegionGermanyNorth
	RegionGermanyNortheast
	RegionGermanyWestCentral
	RegionIndonesiaCentral
	RegionIsraelCentral
	RegionItalyNorth
	RegionJapanEast
	RegionJapanWest
	RegionJioIndiaCentral
	RegionJioIndiaWest
	RegionKoreaCentral
	RegionKoreaSouth
	RegionMalaysiaWest
	RegionMexicoCentral
	RegionNewZealandNorth
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
	RegionSpainCentral
	RegionSwedenCentral
	RegionSwedenSouth
	RegionSwitzerlandNorth
	RegionSwitzerlandWest
	RegionTaiwanNorth
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
	if regionInfoMap[region].cloudAccountRegionEnum == "" {
		return CloudAccountRegionEnum{Region: RegionUnknown}
	}
	return CloudAccountRegionEnum{Region: region}
}

// ToCloudAccountRegionEnumPtr returns the RSC GraphQL AzureCloudAccountRegion
// enum value for the region as a pointer. If the region is unknown, nil is
// returned.
func (region Region) ToCloudAccountRegionEnumPtr() *CloudAccountRegionEnum {
	if r := region.ToCloudAccountRegionEnum(); r.Region != RegionUnknown {
		return &r
	}
	return nil
}

// ToCommonRegionEnum returns the RSC GraphQL AzureCommonRegion enum value for
// the region.
func (region Region) ToCommonRegionEnum() CommonRegionEnum {
	if regionInfoMap[region].commonRegionEnum == "" {
		return CommonRegionEnum{Region: RegionUnknown}
	}
	return CommonRegionEnum{Region: region}
}

// ToCommonRegionEnumPtr returns the RSC GraphQL AzureCommonRegion enum value for
// the region as a pointer. If the region is unknown, nil is returned.
func (region Region) ToCommonRegionEnumPtr() *CommonRegionEnum {
	if r := region.ToCommonRegionEnum(); r.Region != RegionUnknown {
		return &r
	}
	return nil
}

// ToNativeRegionEnum returns the RSC GraphQL AzureNativeRegion enum value for
// the region.
func (region Region) ToNativeRegionEnum() NativeRegionEnum {
	if regionInfoMap[region].nativeRegionEnum == "" {
		return NativeRegionEnum{Region: RegionUnknown}
	}
	return NativeRegionEnum{Region: region}
}

// ToNativeRegionEnumPtr returns the RSC GraphQL AzureNativeRegion enum value
// for the region as a pointer. If the region is unknown, nil is returned.
func (region Region) ToNativeRegionEnumPtr() *NativeRegionEnum {
	if r := region.ToNativeRegionEnum(); r.Region != RegionUnknown {
		return &r
	}
	return nil
}

// ToRegionEnum returns the RSC GraphQL AzureRegion enum value for the region.
func (region Region) ToRegionEnum() RegionEnum {
	if regionInfoMap[region].regionEnum == "" {
		return RegionEnum{Region: RegionUnknown}
	}
	return RegionEnum{Region: region}
}

// ToRegionEnumPtr returns the RSC GraphQL AzureRegion enum value for the region
// as a pointer. If the region is unknown, nil is returned.
func (region Region) ToRegionEnumPtr() *RegionEnum {
	if r := region.ToRegionEnum(); r.Region != RegionUnknown {
		return &r
	}
	return nil
}

// ToRegionForReplicationEnum returns the RSC GraphQL AzureNativeRegionForReplication
// enum value for the region.
func (region Region) ToRegionForReplicationEnum() RegionForReplicationEnum {
	if regionInfoMap[region].regionForReplicationEnum == "" {
		return RegionForReplicationEnum{Region: RegionUnknown}
	}
	return RegionForReplicationEnum{Region: region}
}

// ToRegionForReplicationEnumPtr returns the RSC GraphQL AzureNativeRegionForReplication
// enum value for the region as a pointer. If the region is unknown, nil is returned.
func (region Region) ToRegionForReplicationEnumPtr() *RegionForReplicationEnum {
	if r := region.ToRegionForReplicationEnum(); r.Region != RegionUnknown {
		return &r
	}
	return nil
}

// ToRCSRegionEnum returns the RSC GraphQL RcsRegionEnumType enum value for the
// region.
func (region Region) ToRCSRegionEnum() RCSRegionEnum {
	if regionInfoMap[region].rcsRegionEnum == "" {
		return RCSRegionEnum{Region: RegionUnknown}
	}
	return RCSRegionEnum{Region: region}
}

// ToRCSRegionEnumPtr returns the RSC GraphQL RcsRegionEnumType enum value for the
// region as a pointer. If the region is unknown, nil is returned.
func (region Region) ToRCSRegionEnumPtr() *RCSRegionEnum {
	if r := region.ToRCSRegionEnum(); r.Region != RegionUnknown {
		return &r
	}
	return nil
}

// MarshalJSON returns the region as a JSON string using the region name.
func (region Region) MarshalJSON() ([]byte, error) {
	return json.Marshal(regionInfoMap[region].name)
}

// UnmarshalJSON parses the region from a JSON string using the region name.
func (region *Region) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*region = RegionFromName(s)
	return nil
}

// String returns the name of the region.
func (region Region) String() string {
	return region.Name()
}

const (
	FromAny                      = iota // Parse the value as any of the below formats.
	FromCloudAccountRegionEnum          // Parse the value as a GraphQL AzureCloudAccountRegion enum value.
	FromCommonRegionEnum                // Parse the value as a GraphQL AzureCommonRegion enum value.
	FromDisplayName                     // Parse the value as a region display name.
	FromName                            // Parse the value as a region name.
	FromNativeRegionEnum                // Parse the value as a GraphQL AzureNativeRegion enum value.
	FromRegionalDisplayName             // Parse the value as a region regional display name.
	FromRegionEnum                      // Parse the value as a GraphQL AzureRegion enum value.
	FromRegionForReplicationEnum        // Parse the value as a GraphQL AzureNativeRegionForReplication enum value.
	FromRCSRegionEnum                   // Parse the value as a GraphQL RcsRegionEnumType enum value.
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
		case (valueFormat == FromAny || valueFormat == FromCommonRegionEnum) && info.commonRegionEnum == value:
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
		case (valueFormat == FromAny || valueFormat == FromRCSRegionEnum) && info.rcsRegionEnum == value:
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

// RegionFromCommonRegionEnum parses the value as a GraphQL AzureCommonRegion
// enum value.
func RegionFromCommonRegionEnum(value string) Region {
	return RegionFrom(value, FromCommonRegionEnum)
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
// AzureNativeRegionForReplication enum value.
func RegionFromRegionForReplicationEnum(value string) Region {
	return RegionFrom(value, FromRegionForReplicationEnum)
}

// RegionFromRCSRegionEnum parses the value as a GraphQL RcsRegionEnumType enum
// value.
func RegionFromRCSRegionEnum(value string) Region {
	return RegionFrom(value, FromRCSRegionEnum)
}

// RegionEnum represents the GraphQL AzureRegion enum type.
type RegionEnum struct{ Region }

// MarshalJSON returns the region as a JSON string.
func (region RegionEnum) MarshalJSON() ([]byte, error) {
	return json.Marshal(region.String())
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

// String returns the string representation of the RegionEnum.
func (region RegionEnum) String() string {
	if r := regionInfoMap[region.Region].regionEnum; r != "" {
		return r
	}
	return regionInfoMap[RegionUnknown].regionEnum
}

// CloudAccountRegionEnum represents the GraphQL AzureCloudAccountRegion enum
// type.
type CloudAccountRegionEnum struct{ Region }

// MarshalJSON returns the region as a JSON string.
func (region CloudAccountRegionEnum) MarshalJSON() ([]byte, error) {
	return json.Marshal(region.String())
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

// String returns the string representation of the CloudAccountRegionEnum.
func (region CloudAccountRegionEnum) String() string {
	if r := regionInfoMap[region.Region].cloudAccountRegionEnum; r != "" {
		return r
	}
	return regionInfoMap[RegionUnknown].cloudAccountRegionEnum
}

// CommonRegionEnum represents the GraphQL AzureCommonRegion enum type.
type CommonRegionEnum struct{ Region }

// MarshalJSON returns the region as a JSON string.
func (region CommonRegionEnum) MarshalJSON() ([]byte, error) {
	return json.Marshal(region.String())
}

// UnmarshalJSON parses the region from a JSON string.
func (region *CommonRegionEnum) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	region.Region = RegionFromCommonRegionEnum(s)
	return nil
}

// String returns the string representation of the CommonRegionEnum.
func (region CommonRegionEnum) String() string {
	if r := regionInfoMap[region.Region].commonRegionEnum; r != "" {
		return r
	}
	return regionInfoMap[RegionUnknown].commonRegionEnum
}

// NativeRegionEnum represents the GraphQL AzureNativeRegion enum type.
type NativeRegionEnum struct{ Region }

// MarshalJSON returns the region as a JSON string.
func (region NativeRegionEnum) MarshalJSON() ([]byte, error) {
	return json.Marshal(region.String())
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

// String returns the string representation of the NativeRegionEnum.
func (region NativeRegionEnum) String() string {
	if r := regionInfoMap[region.Region].nativeRegionEnum; r != "" {
		return r
	}
	return regionInfoMap[RegionUnknown].nativeRegionEnum
}

// RegionForReplicationEnum represents the GraphQL AzureNativeRegionForReplication enum type.
type RegionForReplicationEnum struct{ Region }

// MarshalJSON returns the region as a JSON string.
func (region RegionForReplicationEnum) MarshalJSON() ([]byte, error) {
	return json.Marshal(region.String())
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

// String returns the string representation of the RegionForReplicationEnum.
func (region RegionForReplicationEnum) String() string {
	if r := regionInfoMap[region.Region].regionForReplicationEnum; r != "" {
		return r
	}
	return regionInfoMap[RegionUnknown].regionForReplicationEnum
}

// RCSRegionEnum represents the GraphQL RcsRegionEnumType enum type. This region
// type is primarily used for RCS/RCV GraphQL queries/mutations.
type RCSRegionEnum struct{ Region }

// MarshalJSON returns the region as a JSON string.
func (region RCSRegionEnum) MarshalJSON() ([]byte, error) {
	return json.Marshal(region.String())
}

// UnmarshalJSON parses the region from a JSON string.
func (region *RCSRegionEnum) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	region.Region = RegionFromRCSRegionEnum(s)
	return nil
}

// String returns the string representation of the RCSRegionEnum.
func (region RCSRegionEnum) String() string {
	if r := regionInfoMap[region.Region].rcsRegionEnum; r != "" {
		return r
	}
	return regionInfoMap[RegionUnknown].rcsRegionEnum
}

// AllRegionNames returns all the recognized region names.
func AllRegionNames() []string {
	regions := make([]string, 0, len(regionInfoMap))
	for region, info := range regionInfoMap {
		if region != RegionUnknown && region != RegionSource {
			regions = append(regions, info.name)
		}
	}

	return regions
}

// AllRCSRegionNames returns all the recognized RCS/RCV region names.
func AllRCSRegionNames() []string {
	regions := make([]string, 0, len(regionInfoMap))
	for _, info := range regionInfoMap {
		if info.rcsRegionEnum != "" {
			regions = append(regions, info.rcsRegionEnum)
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
	commonRegionEnum         string
	nativeRegionEnum         string
	regionForReplicationEnum string
	rcsRegionEnum            string
}{
	RegionSource: {
		name:                     "n/a",
		displayName:              "Same as source",
		regionForReplicationEnum: "SOURCE_REGION",
	},
	RegionUnknown: {
		name:                     "",
		displayName:              "<Unknown>",
		regionalDisplayName:      "<Unknown>",
		regionEnum:               "UNKNOWN_AZURE_REGION",
		cloudAccountRegionEnum:   "UNKNOWN_AZURE_REGION",
		commonRegionEnum:         "UNKNOWN_AZURE_REGION",
		nativeRegionEnum:         "NOT_SPECIFIED",
		regionForReplicationEnum: "NOT_DEFINED",
		rcsRegionEnum:            "UNKNOWN_AZURE_REGION",
	},
	RegionAustraliaCentral: {
		name:                     "australiacentral",
		displayName:              "Australia Central",
		regionalDisplayName:      "(Asia Pacific) Australia Central",
		regionEnum:               "AUSTRALIA_CENTRAL",
		cloudAccountRegionEnum:   "AUSTRALIACENTRAL",
		commonRegionEnum:         "AUSTRALIACENTRAL",
		nativeRegionEnum:         "AUSTRALIA_CENTRAL",
		regionForReplicationEnum: "AUSTRALIA_CENTRAL",
		rcsRegionEnum:            "AUSTRALIA_CENTRAL",
	},
	RegionAustraliaCentral2: {
		name:                     "australiacentral2",
		displayName:              "Australia Central 2",
		regionalDisplayName:      "(Asia Pacific) Australia Central 2",
		regionEnum:               "AUSTRALIA_CENTRAL2",
		cloudAccountRegionEnum:   "AUSTRALIACENTRAL2",
		commonRegionEnum:         "AUSTRALIACENTRAL2",
		nativeRegionEnum:         "AUSTRALIA_CENTRAL2",
		regionForReplicationEnum: "AUSTRALIA_CENTRAL2",
		rcsRegionEnum:            "AUSTRALIA_CENTRAL2",
	},
	RegionAustraliaEast: {
		name:                     "australiaeast",
		displayName:              "Australia East",
		regionalDisplayName:      "(Asia Pacific) Australia East",
		regionEnum:               "AUSTRALIA_EAST",
		cloudAccountRegionEnum:   "AUSTRALIAEAST",
		commonRegionEnum:         "AUSTRALIAEAST",
		nativeRegionEnum:         "AUSTRALIA_EAST",
		regionForReplicationEnum: "AUSTRALIA_EAST",
		rcsRegionEnum:            "AUSTRALIA_EAST",
	},
	RegionAustraliaSoutheast: {
		name:                     "australiasoutheast",
		displayName:              "Australia Southeast",
		regionalDisplayName:      "(Asia Pacific) Australia Southeast",
		regionEnum:               "AUSTRALIA_SOUTHEAST",
		cloudAccountRegionEnum:   "AUSTRALIASOUTHEAST",
		commonRegionEnum:         "AUSTRALIASOUTHEAST",
		nativeRegionEnum:         "AUSTRALIA_SOUTHEAST",
		regionForReplicationEnum: "AUSTRALIA_SOUTHEAST",
		rcsRegionEnum:            "AUSTRALIA_SOUTHEAST",
	},
	RegionAustriaEast: {
		name:                     "austriaeast",
		displayName:              "Austria East",
		regionalDisplayName:      "(Europe) Austria East",
		regionEnum:               "AUSTRIA_EAST",
		cloudAccountRegionEnum:   "AUSTRIAEAST",
		commonRegionEnum:         "AUSTRIAEAST",
		nativeRegionEnum:         "AUSTRIA_EAST",
		regionForReplicationEnum: "AUSTRIA_EAST",
		rcsRegionEnum:            "AUSTRIA_EAST",
	},
	RegionBelgiumCentral: {
		name:                     "belgiumcentral",
		displayName:              "Belgium Central",
		regionalDisplayName:      "(Europe) Belgium Central",
		regionEnum:               "BELGIUM_CENTRAL",
		cloudAccountRegionEnum:   "BELGIUMCENTRAL",
		commonRegionEnum:         "BELGIUMCENTRAL",
		nativeRegionEnum:         "BELGIUM_CENTRAL",
		regionForReplicationEnum: "BELGIUM_CENTRAL",
		rcsRegionEnum:            "BELGIUM_CENTRAL",
	},
	RegionBrazilSouth: {
		name:                     "brazilsouth",
		displayName:              "Brazil South",
		regionalDisplayName:      "(South America) Brazil South",
		regionEnum:               "BRAZIL_SOUTH",
		cloudAccountRegionEnum:   "BRAZILSOUTH",
		commonRegionEnum:         "BRAZILSOUTH",
		nativeRegionEnum:         "BRAZIL_SOUTH",
		regionForReplicationEnum: "BRAZIL_SOUTH",
		rcsRegionEnum:            "BRAZIL_SOUTH",
	},
	RegionBrazilSoutheast: {
		name:                     "brazilsoutheast",
		displayName:              "Brazil Southeast",
		regionalDisplayName:      "(South America) Brazil Southeast",
		regionEnum:               "BRAZIL_SOUTHEAST",
		cloudAccountRegionEnum:   "BRAZILSOUTHEAST",
		commonRegionEnum:         "BRAZILSOUTHEAST",
		nativeRegionEnum:         "BRAZIL_SOUTHEAST",
		regionForReplicationEnum: "BRAZIL_SOUTHEAST",
		rcsRegionEnum:            "BRAZIL_SOUTHEAST",
	},
	RegionCanadaCentral: {
		name:                     "canadacentral",
		displayName:              "Canada Central",
		regionalDisplayName:      "(Canada) Canada Central",
		regionEnum:               "CANADA_CENTRAL",
		cloudAccountRegionEnum:   "CANADACENTRAL",
		commonRegionEnum:         "CANADACENTRAL",
		nativeRegionEnum:         "CANADA_CENTRAL",
		regionForReplicationEnum: "CANADA_CENTRAL",
		rcsRegionEnum:            "CANADA_CENTRAL",
	},
	RegionCanadaEast: {
		name:                     "canadaeast",
		displayName:              "Canada East",
		regionalDisplayName:      "(Canada) Canada East",
		regionEnum:               "CANADA_EAST",
		cloudAccountRegionEnum:   "CANADAEAST",
		commonRegionEnum:         "CANADAEAST",
		nativeRegionEnum:         "CANADA_EAST",
		regionForReplicationEnum: "CANADA_EAST",
		rcsRegionEnum:            "CANADA_EAST",
	},
	RegionCentralIndia: {
		name:                     "centralindia",
		displayName:              "Central India",
		regionalDisplayName:      "(Asia Pacific) Central India",
		regionEnum:               "INDIA_CENTRAL",
		cloudAccountRegionEnum:   "CENTRALINDIA",
		commonRegionEnum:         "CENTRALINDIA",
		nativeRegionEnum:         "CENTRAL_INDIA",
		regionForReplicationEnum: "CENTRAL_INDIA",
		rcsRegionEnum:            "INDIA_CENTRAL",
	},
	RegionCentralUS: {
		name:                     "centralus",
		displayName:              "Central US",
		regionalDisplayName:      "(US) Central US",
		regionEnum:               "US_CENTRAL",
		cloudAccountRegionEnum:   "CENTRALUS",
		commonRegionEnum:         "CENTRALUS",
		nativeRegionEnum:         "CENTRAL_US",
		regionForReplicationEnum: "CENTRAL_US",
		rcsRegionEnum:            "US_CENTRAL",
	},
	RegionChileCentral: {
		name:                     "chilecentral",
		displayName:              "Chile Central",
		regionalDisplayName:      "(South America) Chile Central",
		regionEnum:               "CHILE_CENTRAL",
		cloudAccountRegionEnum:   "CHILECENTRAL",
		commonRegionEnum:         "CHILECENTRAL",
		nativeRegionEnum:         "CHILE_CENTRAL",
		regionForReplicationEnum: "CHILE_CENTRAL",
		rcsRegionEnum:            "CHILE_CENTRAL",
	},
	RegionChinaEast: {
		name:                     "chinaeast",
		displayName:              "China East",
		regionalDisplayName:      "(China) China East",
		regionEnum:               "CHINA_EAST",
		cloudAccountRegionEnum:   "CHINAEAST",
		commonRegionEnum:         "CHINAEAST",
		nativeRegionEnum:         "CHINA_EAST",
		regionForReplicationEnum: "CHINA_EAST",
	},
	RegionChinaEast2: {
		name:                     "chinaeast2",
		displayName:              "China East 2",
		regionalDisplayName:      "(China) China East 2",
		regionEnum:               "CHINA_EAST2",
		cloudAccountRegionEnum:   "CHINAEAST2",
		commonRegionEnum:         "CHINAEAST2",
		nativeRegionEnum:         "CHINA_EAST2",
		regionForReplicationEnum: "CHINA_EAST2",
	},
	RegionChinaNorth: {
		name:                     "chinanorth",
		displayName:              "China North",
		regionalDisplayName:      "(China) China North",
		regionEnum:               "CHINA_NORTH",
		cloudAccountRegionEnum:   "CHINANORTH",
		commonRegionEnum:         "CHINANORTH",
		nativeRegionEnum:         "CHINA_NORTH",
		regionForReplicationEnum: "CHINA_NORTH",
	},
	RegionChinaNorth2: {
		name:                     "chinanorth2",
		displayName:              "China North 2",
		regionalDisplayName:      "(China) China North 2",
		cloudAccountRegionEnum:   "CHINANORTH2",
		commonRegionEnum:         "CHINANORTH2",
		nativeRegionEnum:         "CHINA_NORTH2",
		regionForReplicationEnum: "CHINA_NORTH2",
	},
	RegionEastAsia: {
		name:                     "eastasia",
		displayName:              "East Asia",
		regionalDisplayName:      "(Asia Pacific) East Asia",
		regionEnum:               "ASIA_EAST",
		cloudAccountRegionEnum:   "EASTASIA",
		commonRegionEnum:         "EASTASIA",
		nativeRegionEnum:         "EAST_ASIA",
		regionForReplicationEnum: "EAST_ASIA",
		rcsRegionEnum:            "ASIA_EAST",
	},
	RegionEastUS: {
		name:                     "eastus",
		displayName:              "East US",
		regionalDisplayName:      "(US) East US",
		regionEnum:               "US_EAST",
		cloudAccountRegionEnum:   "EASTUS",
		commonRegionEnum:         "EASTUS",
		nativeRegionEnum:         "EAST_US",
		regionForReplicationEnum: "EAST_US",
		rcsRegionEnum:            "US_EAST",
	},
	RegionEastUS2: {
		name:                     "eastus2",
		displayName:              "East US 2",
		regionalDisplayName:      "(US) East US 2",
		regionEnum:               "US_EAST2",
		cloudAccountRegionEnum:   "EASTUS2",
		commonRegionEnum:         "EASTUS2",
		nativeRegionEnum:         "EAST_US2",
		regionForReplicationEnum: "EAST_US2",
		rcsRegionEnum:            "US_EAST_2_VIRGINIA",
	},
	RegionFranceCentral: {
		name:                     "francecentral",
		displayName:              "France Central",
		regionalDisplayName:      "(Europe) France Central",
		regionEnum:               "FRANCE_CENTRAL",
		cloudAccountRegionEnum:   "FRANCECENTRAL",
		commonRegionEnum:         "FRANCECENTRAL",
		nativeRegionEnum:         "FRANCE_CENTRAL",
		regionForReplicationEnum: "FRANCE_CENTRAL",
		rcsRegionEnum:            "FRANCE_CENTRAL",
	},
	RegionFranceSouth: {
		name:                     "francesouth",
		displayName:              "France South",
		regionalDisplayName:      "(Europe) France South",
		regionEnum:               "FRANCE_SOUTH",
		cloudAccountRegionEnum:   "FRANCESOUTH",
		commonRegionEnum:         "FRANCESOUTH",
		nativeRegionEnum:         "FRANCE_SOUTH",
		regionForReplicationEnum: "FRANCE_SOUTH",
		rcsRegionEnum:            "FRANCE_SOUTH",
	},
	RegionGermanyCentral: {
		name:                "germanycentral",
		displayName:         "Germany Central",
		regionalDisplayName: "(Europe) Germany Central",
		regionEnum:          "GERMANY_CENTRAL",
	},
	RegionGermanyNorth: {
		name:                     "germanynorth",
		displayName:              "Germany North",
		regionalDisplayName:      "(Europe) Germany North",
		regionEnum:               "GERMANY_NORTH",
		cloudAccountRegionEnum:   "GERMANYNORTH",
		commonRegionEnum:         "GERMANYNORTH",
		nativeRegionEnum:         "GERMANY_NORTH",
		regionForReplicationEnum: "GERMANY_NORTH",
		rcsRegionEnum:            "GERMANY_NORTH",
	},
	RegionGermanyNortheast: {
		name:                "germanynortheast",
		displayName:         "Germany Northeast",
		regionalDisplayName: "(Europe) Germany Northeast",
		regionEnum:          "GERMANY_NORTHEAST",
	},
	RegionGermanyWestCentral: {
		name:                     "germanywestcentral",
		displayName:              "Germany West Central",
		regionalDisplayName:      "(Europe) Germany West Central",
		regionEnum:               "GERMANY_WEST_CENTRAL",
		cloudAccountRegionEnum:   "GERMANYWESTCENTRAL",
		commonRegionEnum:         "GERMANYWESTCENTRAL",
		nativeRegionEnum:         "GERMANY_WEST_CENTRAL",
		regionForReplicationEnum: "GERMANY_WEST_CENTRAL",
		rcsRegionEnum:            "GERMANY_WEST_CENTRAL",
	},
	RegionIndonesiaCentral: {
		name:                     "indonesiacentral",
		displayName:              "Indonesia Central",
		regionalDisplayName:      "(Asia Pacific) Indonesia Central",
		regionEnum:               "INDONESIA_CENTRAL",
		cloudAccountRegionEnum:   "INDONESIACENTRAL",
		commonRegionEnum:         "INDONESIACENTRAL",
		nativeRegionEnum:         "INDONESIA_CENTRAL",
		regionForReplicationEnum: "INDONESIA_CENTRAL",
		rcsRegionEnum:            "INDONESIA_CENTRAL",
	},
	RegionIsraelCentral: {
		name:                     "israelcentral",
		displayName:              "Israel Central",
		regionalDisplayName:      "(Middle East) Israel Central",
		regionEnum:               "ISRAEL_CENTRAL",
		cloudAccountRegionEnum:   "ISRAELCENTRAL",
		commonRegionEnum:         "ISRAELCENTRAL",
		nativeRegionEnum:         "ISRAEL_CENTRAL",
		regionForReplicationEnum: "ISRAEL_CENTRAL",
		rcsRegionEnum:            "ISRAEL_CENTRAL",
	},
	RegionItalyNorth: {
		name:                     "italynorth",
		displayName:              "Italy North",
		regionalDisplayName:      "(Europe) Italy North",
		regionEnum:               "ITALY_NORTH",
		cloudAccountRegionEnum:   "ITALYNORTH",
		commonRegionEnum:         "ITALYNORTH",
		nativeRegionEnum:         "ITALY_NORTH",
		regionForReplicationEnum: "ITALY_NORTH",
		rcsRegionEnum:            "ITALY_NORTH",
	},
	RegionJapanEast: {
		name:                     "japaneast",
		displayName:              "Japan East",
		regionalDisplayName:      "(Asia Pacific) Japan East",
		regionEnum:               "JAPAN_EAST",
		cloudAccountRegionEnum:   "JAPANEAST",
		commonRegionEnum:         "JAPANEAST",
		nativeRegionEnum:         "JAPAN_EAST",
		regionForReplicationEnum: "JAPAN_EAST",
		rcsRegionEnum:            "JAPAN_EAST",
	},
	RegionJapanWest: {
		name:                     "japanwest",
		displayName:              "Japan West",
		regionalDisplayName:      "(Asia Pacific) Japan West",
		regionEnum:               "JAPAN_WEST",
		cloudAccountRegionEnum:   "JAPANWEST",
		commonRegionEnum:         "JAPANWEST",
		nativeRegionEnum:         "JAPAN_WEST",
		regionForReplicationEnum: "JAPAN_WEST",
		rcsRegionEnum:            "JAPAN_WEST",
	},
	RegionJioIndiaCentral: {
		name:                "jioindiacentral",
		displayName:         "Jio India Central",
		regionalDisplayName: "(Asia Pacific) Jio India Central",
	},
	RegionJioIndiaWest: {
		name:                "jioindiawest",
		displayName:         "Jio India West",
		regionalDisplayName: "(Asia Pacific) Jio India West",
	},
	RegionKoreaCentral: {
		name:                     "koreacentral",
		displayName:              "Korea Central",
		regionalDisplayName:      "(Asia Pacific) Korea Central",
		regionEnum:               "KOREA_CENTRAL",
		cloudAccountRegionEnum:   "KOREACENTRAL",
		commonRegionEnum:         "KOREACENTRAL",
		nativeRegionEnum:         "KOREA_CENTRAL",
		regionForReplicationEnum: "KOREA_CENTRAL",
		rcsRegionEnum:            "KOREA_CENTRAL",
	},
	RegionKoreaSouth: {
		name:                     "koreasouth",
		displayName:              "Korea South",
		regionalDisplayName:      "(Asia Pacific) Korea South",
		regionEnum:               "KOREA_SOUTH",
		cloudAccountRegionEnum:   "KOREASOUTH",
		commonRegionEnum:         "KOREASOUTH",
		nativeRegionEnum:         "KOREA_SOUTH",
		regionForReplicationEnum: "KOREA_SOUTH",
		rcsRegionEnum:            "KOREA_SOUTH",
	},
	RegionMalaysiaWest: {
		name:                     "malaysiawest",
		displayName:              "Malaysia West",
		regionalDisplayName:      "(Asia Pacific) Malaysia West",
		regionEnum:               "MALAYSIA_WEST",
		cloudAccountRegionEnum:   "MALAYSIAWEST",
		commonRegionEnum:         "MALAYSIAWEST",
		nativeRegionEnum:         "MALAYSIA_WEST",
		regionForReplicationEnum: "MALAYSIA_WEST",
		rcsRegionEnum:            "MALAYSIA_WEST",
	},
	RegionMexicoCentral: {
		name:                     "mexicocentral",
		displayName:              "Mexico Central",
		regionalDisplayName:      "(Mexico) Mexico Central",
		regionEnum:               "MEXICO_CENTRAL",
		cloudAccountRegionEnum:   "MEXICOCENTRAL",
		commonRegionEnum:         "MEXICOCENTRAL",
		nativeRegionEnum:         "MEXICO_CENTRAL",
		regionForReplicationEnum: "MEXICO_CENTRAL",
		rcsRegionEnum:            "MEXICO_CENTRAL",
	},
	RegionNewZealandNorth: {
		name:                "newzealandnorth",
		displayName:         "New Zealand North",
		regionalDisplayName: "(Asia Pacific) New Zealand North",
		regionEnum:          "NEW_ZEALAND_NORTH",
		commonRegionEnum:    "NEWZEALANDNORTH",
		rcsRegionEnum:       "NEW_ZEALAND_NORTH",
	},
	RegionNorthCentralUS: {
		name:                     "northcentralus",
		displayName:              "North Central US",
		regionalDisplayName:      "(US) North Central US",
		regionEnum:               "US_NORTH_CENTRAL",
		cloudAccountRegionEnum:   "NORTHCENTRALUS",
		commonRegionEnum:         "NORTHCENTRALUS",
		nativeRegionEnum:         "NORTH_CENTRAL_US",
		regionForReplicationEnum: "NORTH_CENTRAL_US",
		rcsRegionEnum:            "US_NORTH_CENTRAL",
	},
	RegionNorthEurope: {
		name:                     "northeurope",
		displayName:              "North Europe",
		regionalDisplayName:      "(Europe) North Europe",
		regionEnum:               "EUROPE_NORTH",
		cloudAccountRegionEnum:   "NORTHEUROPE",
		commonRegionEnum:         "NORTHEUROPE",
		nativeRegionEnum:         "NORTH_EUROPE",
		regionForReplicationEnum: "NORTH_EUROPE",
		rcsRegionEnum:            "EUROPE_NORTH",
	},
	RegionNorwayEast: {
		name:                     "norwayeast",
		displayName:              "Norway East",
		regionalDisplayName:      "(Europe) Norway East",
		regionEnum:               "NORWAY_EAST",
		cloudAccountRegionEnum:   "NORWAYEAST",
		commonRegionEnum:         "NORWAYEAST",
		nativeRegionEnum:         "NORWAY_EAST",
		regionForReplicationEnum: "NORWAY_EAST",
		rcsRegionEnum:            "NORWAY_EAST",
	},
	RegionNorwayWest: {
		name:                     "norwaywest",
		displayName:              "Norway West",
		regionalDisplayName:      "(Europe) Norway West",
		regionEnum:               "NORWAY_WEST",
		cloudAccountRegionEnum:   "NORWAYWEST",
		commonRegionEnum:         "NORWAYWEST",
		nativeRegionEnum:         "NORWAY_WEST",
		regionForReplicationEnum: "NORWAY_WEST",
		rcsRegionEnum:            "NORWAY_WEST",
	},
	RegionPolandCentral: {
		name:                     "polandcentral",
		displayName:              "Poland Central",
		regionalDisplayName:      "(Europe) Poland Central",
		regionEnum:               "POLAND_CENTRAL",
		cloudAccountRegionEnum:   "POLANDCENTRAL",
		commonRegionEnum:         "POLANDCENTRAL",
		nativeRegionEnum:         "POLAND_CENTRAL",
		regionForReplicationEnum: "POLAND_CENTRAL",
		rcsRegionEnum:            "POLAND_CENTRAL",
	},
	RegionQatarCentral: {
		name:                     "qatarcentral",
		displayName:              "Qatar Central",
		regionalDisplayName:      "(Middle East) Qatar Central",
		regionEnum:               "QATAR_CENTRAL",
		cloudAccountRegionEnum:   "QATARCENTRAL",
		commonRegionEnum:         "QATARCENTRAL",
		nativeRegionEnum:         "QATAR_CENTRAL",
		regionForReplicationEnum: "QATAR_CENTRAL",
		rcsRegionEnum:            "QATAR_CENTRAL",
	},
	RegionSouthAfricaNorth: {
		name:                     "southafricanorth",
		displayName:              "South Africa North",
		regionalDisplayName:      "(Africa) South Africa North",
		regionEnum:               "SOUTH_AFRICA_NORTH",
		cloudAccountRegionEnum:   "SOUTHAFRICANORTH",
		commonRegionEnum:         "SOUTHAFRICANORTH",
		nativeRegionEnum:         "SOUTH_AFRICA_NORTH",
		regionForReplicationEnum: "SOUTH_AFRICA_NORTH",
		rcsRegionEnum:            "SOUTH_AFRICA_NORTH",
	},
	RegionSouthAfricaWest: {
		name:                     "southafricawest",
		displayName:              "South Africa West",
		regionalDisplayName:      "(Africa) South Africa West",
		regionEnum:               "SOUTH_AFRICA_WEST",
		cloudAccountRegionEnum:   "SOUTHAFRICAWEST",
		commonRegionEnum:         "SOUTHAFRICAWEST",
		nativeRegionEnum:         "SOUTH_AFRICA_WEST",
		regionForReplicationEnum: "SOUTH_AFRICA_WEST",
		rcsRegionEnum:            "SOUTH_AFRICA_WEST",
	},
	RegionSouthCentralUS: {
		name:                     "southcentralus",
		displayName:              "South Central US",
		regionalDisplayName:      "(US) South Central US",
		regionEnum:               "US_SOUTH_CENTRAL",
		cloudAccountRegionEnum:   "SOUTHCENTRALUS",
		commonRegionEnum:         "SOUTHCENTRALUS",
		nativeRegionEnum:         "SOUTH_CENTRAL_US",
		regionForReplicationEnum: "SOUTH_CENTRAL_US",
		rcsRegionEnum:            "US_SOUTH_CENTRAL",
	},
	RegionSoutheastAsia: {
		name:                     "southeastasia",
		displayName:              "Southeast Asia",
		regionalDisplayName:      "(Asia Pacific) Southeast Asia",
		regionEnum:               "ASIA_SOUTHEAST",
		cloudAccountRegionEnum:   "SOUTHEASTASIA",
		commonRegionEnum:         "SOUTHEASTASIA",
		nativeRegionEnum:         "SOUTHEAST_ASIA",
		regionForReplicationEnum: "SOUTHEAST_ASIA",
		rcsRegionEnum:            "ASIA_SOUTHEAST",
	},
	RegionSouthIndia: {
		name:                     "southindia",
		displayName:              "South India",
		regionalDisplayName:      "(Asia Pacific) South India",
		regionEnum:               "INDIA_SOUTH",
		cloudAccountRegionEnum:   "SOUTHINDIA",
		commonRegionEnum:         "SOUTHINDIA",
		nativeRegionEnum:         "SOUTH_INDIA",
		regionForReplicationEnum: "SOUTH_INDIA",
		rcsRegionEnum:            "INDIA_SOUTH",
	},
	RegionSpainCentral: {
		name:                     "spaincentral",
		displayName:              "Spain Central",
		regionalDisplayName:      "(Europe) Spain Central",
		regionEnum:               "SPAIN_CENTRAL",
		cloudAccountRegionEnum:   "SPAINCENTRAL",
		commonRegionEnum:         "SPAINCENTRAL",
		nativeRegionEnum:         "SPAIN_CENTRAL",
		regionForReplicationEnum: "SPAIN_CENTRAL",
		rcsRegionEnum:            "SPAIN_CENTRAL",
	},
	RegionSwedenCentral: {
		name:                     "swedencentral",
		displayName:              "Sweden Central",
		regionalDisplayName:      "(Europe) Sweden Central",
		regionEnum:               "SWEDEN_CENTRAL",
		cloudAccountRegionEnum:   "SWEDENCENTRAL",
		commonRegionEnum:         "SWEDENCENTRAL",
		nativeRegionEnum:         "SWEDEN_CENTRAL",
		regionForReplicationEnum: "SWEDEN_CENTRAL",
		rcsRegionEnum:            "SWEDEN_CENTRAL",
	},
	RegionSwedenSouth: {
		name:                     "swedensouth",
		displayName:              "Sweden South",
		regionalDisplayName:      "(Europe) Sweden South",
		regionEnum:               "SWEDEN_SOUTH",
		cloudAccountRegionEnum:   "SWEDENSOUTH",
		commonRegionEnum:         "SWEDENSOUTH",
		nativeRegionEnum:         "SWEDEN_SOUTH",
		regionForReplicationEnum: "SWEDEN_SOUTH",
		rcsRegionEnum:            "SWEDEN_SOUTH",
	},
	RegionSwitzerlandNorth: {
		name:                     "switzerlandnorth",
		displayName:              "Switzerland North",
		regionalDisplayName:      "(Europe) Switzerland North",
		regionEnum:               "SWITZERLAND_NORTH",
		cloudAccountRegionEnum:   "SWITZERLANDNORTH",
		commonRegionEnum:         "SWITZERLANDNORTH",
		nativeRegionEnum:         "SWITZERLAND_NORTH",
		regionForReplicationEnum: "SWITZERLAND_NORTH",
		rcsRegionEnum:            "SWITZERLAND_NORTH",
	},
	RegionSwitzerlandWest: {
		name:                     "switzerlandwest",
		displayName:              "Switzerland West",
		regionalDisplayName:      "(Europe) Switzerland West",
		regionEnum:               "SWITZERLAND_WEST",
		cloudAccountRegionEnum:   "SWITZERLANDWEST",
		commonRegionEnum:         "SWITZERLANDWEST",
		nativeRegionEnum:         "SWITZERLAND_WEST",
		regionForReplicationEnum: "SWITZERLAND_WEST",
		rcsRegionEnum:            "SWITZERLAND_WEST",
	},
	RegionTaiwanNorth: {
		name:                "taiwannorth",
		displayName:         "Taiwan North",
		regionalDisplayName: "(Asia Pacific) Taiwan North",
		commonRegionEnum:    "TAIWANNORTH",
	},
	RegionUAECentral: {
		name:                     "uaecentral",
		displayName:              "UAE Central",
		regionalDisplayName:      "(Middle East) UAE Central",
		regionEnum:               "UAE_CENTRAL",
		cloudAccountRegionEnum:   "UAECENTRAL",
		commonRegionEnum:         "UAECENTRAL",
		nativeRegionEnum:         "UAE_CENTRAL",
		regionForReplicationEnum: "UAE_CENTRAL",
		rcsRegionEnum:            "UAE_CENTRAL",
	},
	RegionUAENorth: {
		name:                     "uaenorth",
		displayName:              "UAE North",
		regionalDisplayName:      "(Middle East) UAE North",
		regionEnum:               "UAE_NORTH",
		cloudAccountRegionEnum:   "UAENORTH",
		commonRegionEnum:         "UAENORTH",
		nativeRegionEnum:         "UAE_NORTH",
		regionForReplicationEnum: "UAE_NORTH",
		rcsRegionEnum:            "UAE_NORTH",
	},
	RegionUKSouth: {
		name:                     "uksouth",
		displayName:              "UK South",
		regionalDisplayName:      "(Europe) UK South",
		regionEnum:               "UK_SOUTH",
		cloudAccountRegionEnum:   "UKSOUTH",
		commonRegionEnum:         "UKSOUTH",
		nativeRegionEnum:         "UK_SOUTH",
		regionForReplicationEnum: "UK_SOUTH",
		rcsRegionEnum:            "UK_SOUTH",
	},
	RegionUKWest: {
		name:                     "ukwest",
		displayName:              "UK West",
		regionalDisplayName:      "(Europe) UK West",
		regionEnum:               "UK_WEST",
		cloudAccountRegionEnum:   "UKWEST",
		commonRegionEnum:         "UKWEST",
		nativeRegionEnum:         "UK_WEST",
		regionForReplicationEnum: "UK_WEST",
		rcsRegionEnum:            "UK_WEST",
	},
	RegionUSDoDCentral: {
		name:                "usdodcentral",
		displayName:         "US DoD Central",
		regionalDisplayName: "(US Gov) US DoD Central",
		regionEnum:          "GOV_US_DOD_CENTRAL",
	},
	RegionUSDoDEast: {
		name:                "usdodeast",
		displayName:         "US DoD East",
		regionalDisplayName: "(US Gov) US DoD East",
		regionEnum:          "GOV_US_DOD_EAST",
	},
	RegionUSGovArizona: {
		name:                     "usgovarizona",
		displayName:              "US Gov Arizona",
		regionalDisplayName:      "(US Gov) US Gov Arizona",
		regionEnum:               "GOV_US_ARIZONA",
		cloudAccountRegionEnum:   "USGOVARIZONA",
		commonRegionEnum:         "USGOVARIZONA",
		nativeRegionEnum:         "US_GOV_ARIZONA",
		regionForReplicationEnum: "US_GOV_ARIZONA",
		rcsRegionEnum:            "GOV_US_ARIZONA",
	},
	RegionUSGovTexas: {
		name:                     "usgovtexas",
		displayName:              "US Gov Texas",
		regionalDisplayName:      "(US Gov) US Gov Texas",
		regionEnum:               "GOV_US_TEXAS",
		cloudAccountRegionEnum:   "USGOVTEXAS",
		commonRegionEnum:         "USGOVTEXAS",
		nativeRegionEnum:         "US_GOV_TEXAS",
		regionForReplicationEnum: "US_GOV_TEXAS",
		rcsRegionEnum:            "GOV_US_TEXAS",
	},
	RegionUSGovVirginia: {
		name:                     "usgovvirginia",
		displayName:              "US Gov Virginia",
		regionalDisplayName:      "(US Gov) US Gov Virginia",
		regionEnum:               "GOV_US_VIRGINIA",
		cloudAccountRegionEnum:   "USGOVVIRGINIA",
		commonRegionEnum:         "USGOVVIRGINIA",
		nativeRegionEnum:         "US_GOV_VIRGINIA",
		regionForReplicationEnum: "US_GOV_VIRGINIA",
		rcsRegionEnum:            "GOV_US_VIRGINIA",
	},
	RegionWestCentralUS: {
		name:                     "westcentralus",
		displayName:              "West Central US",
		regionalDisplayName:      "(US) West Central US",
		regionEnum:               "US_WEST_CENTRAL",
		cloudAccountRegionEnum:   "WESTCENTRALUS",
		commonRegionEnum:         "WESTCENTRALUS",
		nativeRegionEnum:         "WEST_CENTRAL_US",
		regionForReplicationEnum: "WEST_CENTRAL_US",
		rcsRegionEnum:            "US_WEST_CENTRAL",
	},
	RegionWestEurope: {
		name:                     "westeurope",
		displayName:              "West Europe",
		regionalDisplayName:      "(Europe) West Europe",
		regionEnum:               "EUROPE_WEST",
		cloudAccountRegionEnum:   "WESTEUROPE",
		commonRegionEnum:         "WESTEUROPE",
		nativeRegionEnum:         "WEST_EUROPE",
		regionForReplicationEnum: "WEST_EUROPE",
		rcsRegionEnum:            "EUROPE_WEST",
	},
	RegionWestIndia: {
		name:                     "westindia",
		displayName:              "West India",
		regionalDisplayName:      "(Asia Pacific) West India",
		regionEnum:               "INDIA_WEST",
		cloudAccountRegionEnum:   "WESTINDIA",
		commonRegionEnum:         "WESTINDIA",
		nativeRegionEnum:         "WEST_INDIA",
		regionForReplicationEnum: "WEST_INDIA",
		rcsRegionEnum:            "INDIA_WEST",
	},
	RegionWestUS: {
		name:                     "westus",
		displayName:              "West US",
		regionalDisplayName:      "(US) West US",
		regionEnum:               "US_WEST",
		cloudAccountRegionEnum:   "WESTUS",
		commonRegionEnum:         "WESTUS",
		nativeRegionEnum:         "WEST_US",
		regionForReplicationEnum: "WEST_US",
		rcsRegionEnum:            "US_WEST",
	},
	RegionWestUS2: {
		name:                     "westus2",
		displayName:              "West US 2",
		regionalDisplayName:      "(US) West US 2",
		regionEnum:               "US_WEST2",
		cloudAccountRegionEnum:   "WESTUS2",
		commonRegionEnum:         "WESTUS2",
		nativeRegionEnum:         "WEST_US2",
		regionForReplicationEnum: "WEST_US2",
		rcsRegionEnum:            "US_WEST_2",
	},
	RegionWestUS3: {
		name:                     "westus3",
		displayName:              "West US 3",
		regionalDisplayName:      "(US) West US 3",
		regionEnum:               "WEST_US3",
		cloudAccountRegionEnum:   "WESTUS3",
		commonRegionEnum:         "WESTUS3",
		nativeRegionEnum:         "WEST_US3",
		regionForReplicationEnum: "WEST_US3",
		rcsRegionEnum:            "WEST_US3",
	},
}
