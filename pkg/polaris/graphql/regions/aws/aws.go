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

package aws

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	RegionSource  Region = -1
	RegionUnknown Region = iota
	RegionAfSouth1
	RegionApEast1
	RegionApNorthEast1
	RegionApNorthEast2
	RegionApNorthEast3
	RegionApSouthEast1
	RegionApSouthEast2
	RegionApSouthEast3
	RegionApSouthEast4
	RegionApSouthEast5
	RegionApSouth1
	RegionApSouth2
	RegionCaCentral1
	RegionCaWest1
	RegionCnNorthWest1
	RegionCnNorth1
	RegionEuCentral1
	RegionEuCentral2
	RegionEuNorth1
	RegionEuSouth1
	RegionEuSouth2
	RegionEuWest1
	RegionEuWest2
	RegionEuWest3
	RegionIlCentral1
	RegionMeCentral1
	RegionMeSouth1
	RegionSaEast1
	RegionUsEast1
	RegionUsEast2
	RegionUsGovEast1
	RegionUsGovWest1
	RegionUsWest1
	RegionUsWest2
	RegionUsISOEast1
	RegionUsISOWest1
	RegionUsISOBEast1
)

// Region represents an AWS region in RSC. When reading a Region from a JSON
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

// ToNativeRegionEnum returns the RSC GraphQL AwsNativeRegion enum value for the
// region.
func (region Region) ToNativeRegionEnum() NativeRegionEnum {
	return NativeRegionEnum{Region: region}
}

// ToNativeRegionEnumPtr returns the RSC GraphQL AwsNativeRegion enum value for
// the region as a pointer. If the region is unknown, nil is returned.
func (region Region) ToNativeRegionEnumPtr() *NativeRegionEnum {
	if region == RegionUnknown {
		return nil
	}
	return &NativeRegionEnum{Region: region}
}

// ToRegionEnum returns the RSC GraphQL AwsRegion enum value for the region.
func (region Region) ToRegionEnum() RegionEnum {
	return RegionEnum{Region: region}
}

// ToRegionEnumPtr returns the RSC GraphQL AwsNativeRegion enum value for the
// region as a pointer. If the region is unknown, nil is returned.
func (region Region) ToRegionEnumPtr() *RegionEnum {
	if region == RegionUnknown {
		return nil
	}
	return &RegionEnum{Region: region}
}

// ToRegionForReplicationEnum returns the RSC GraphQL AwsRegionForReplication enum value for the region.
func (region Region) ToRegionForReplicationEnum() RegionForReplicationEnum {
	return RegionForReplicationEnum{Region: region}
}

// String returns the name of the region.
func (region Region) String() string {
	return region.Name()
}

const (
	FromAny                      = iota // Parse the value as any of the below formats.
	FromDisplayName                     // Parse the value as a region display name.
	FromName                            // Parse the value as a region name.
	FromNativeRegionEnum                // Parse the value as a GraphQL AwsNativeRegion enum value.
	FromRegionEnum                      // Parse the value as a GraphQL AwsRegion enum value.
	FromRegionForReplicationEnum        // Parse the value as a GraphQL AwsRegionForReplication enum value.
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
		case (valueFormat == FromAny || valueFormat == FromNativeRegionEnum) && info.nativeRegionEnum == value:
			return r
		case (valueFormat == FromAny || valueFormat == FromRegionEnum) && info.regionEnum == value:
			return r
		case (valueFormat == FromAny || valueFormat == FromDisplayName) && info.displayName == value:
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

// RegionFromNativeRegionEnum parses the value as a GraphQL AwsNativeRegion
// enum.
func RegionFromNativeRegionEnum(value string) Region {
	return RegionFrom(value, FromNativeRegionEnum)
}

// RegionFromRegionEnum parses the value as a GraphQL AwsRegion enum value.
func RegionFromRegionEnum(value string) Region {
	return RegionFrom(value, FromRegionEnum)
}

// RegionFromRegionForReplicationEnum parses the value as a GraphQL AwsRegionForReplication enum value.
func RegionFromRegionForReplicationEnum(value string) Region {
	return RegionFrom(value, FromRegionForReplicationEnum)
}

// RegionEnum represents the GraphQL AwsRegion enum type.
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

// NativeRegionEnum represents the GraphQL AwsNativeRegion enum type.
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

// RegionForReplicationEnum represents the GraphQL AwsRegionForReplication enum type.
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
	regionEnum               string
	nativeRegionEnum         string
	regionForReplicationEnum string
}{
	RegionSource: {
		name:                     "n/a",
		displayName:              "Same as source",
		regionEnum:               "n/a",
		nativeRegionEnum:         "n/a",
		regionForReplicationEnum: "SOURCE_REGION",
	},
	RegionUnknown: {
		name:                     "n/a",
		displayName:              "<Unknown>",
		regionEnum:               "UNKNOWN_AWS_REGION",
		nativeRegionEnum:         "NOT_SPECIFIED",
		regionForReplicationEnum: "NOT_DEFINED",
	},
	RegionAfSouth1: {
		name:                     "af-south-1",
		displayName:              "Africa (Cape Town)",
		regionEnum:               "AF_SOUTH_1",
		nativeRegionEnum:         "AF_SOUTH_1",
		regionForReplicationEnum: "AF_SOUTH_1",
	},
	RegionApEast1: {
		name:                     "ap-east-1",
		displayName:              "Asia Pacific (Hong Kong)",
		regionEnum:               "AP_EAST_1",
		nativeRegionEnum:         "AP_EAST_1",
		regionForReplicationEnum: "AP_EAST_1",
	},
	RegionApNorthEast1: {
		name:                     "ap-northeast-1",
		displayName:              "Asia Pacific (Tokyo)",
		regionEnum:               "AP_NORTHEAST_1",
		nativeRegionEnum:         "AP_NORTHEAST_1",
		regionForReplicationEnum: "AP_NORTHEAST_1",
	},
	RegionApNorthEast2: {
		name:                     "ap-northeast-2",
		displayName:              "Asia Pacific (Seoul)",
		regionEnum:               "AP_NORTHEAST_2",
		nativeRegionEnum:         "AP_NORTHEAST_2",
		regionForReplicationEnum: "AP_NORTHEAST_2",
	},
	RegionApNorthEast3: {
		name:                     "ap-northeast-3",
		displayName:              "Asia Pacific (Osaka)",
		regionEnum:               "AP_NORTHEAST_3",
		nativeRegionEnum:         "AP_NORTHEAST_3",
		regionForReplicationEnum: "AP_NORTHEAST_3",
	},
	RegionApSouthEast1: {
		name:                     "ap-southeast-1",
		displayName:              "Asia Pacific (Singapore)",
		regionEnum:               "AP_SOUTHEAST_1",
		nativeRegionEnum:         "AP_SOUTHEAST_1",
		regionForReplicationEnum: "AP_SOUTHEAST_1",
	},
	RegionApSouthEast2: {
		name:                     "ap-southeast-2",
		displayName:              "Asia Pacific (Sydney)",
		regionEnum:               "AP_SOUTHEAST_2",
		nativeRegionEnum:         "AP_SOUTHEAST_2",
		regionForReplicationEnum: "AP_SOUTHEAST_2",
	},
	RegionApSouthEast3: {
		name:                     "ap-southeast-3",
		displayName:              "Asia Pacific (Jakarta)",
		regionEnum:               "AP_SOUTHEAST_3",
		nativeRegionEnum:         "AP_SOUTHEAST_3",
		regionForReplicationEnum: "AP_SOUTHEAST_3",
	},
	RegionApSouthEast4: {
		name:                     "ap-southeast-4",
		displayName:              "Asia Pacific (Melbourne)",
		regionEnum:               "AP_SOUTHEAST_4",
		nativeRegionEnum:         "AP_SOUTHEAST_4",
		regionForReplicationEnum: "AP_SOUTHEAST_4",
	},
	RegionApSouthEast5: {
		name:                     "ap-southeast-5",
		displayName:              "Asia Pacific (Malaysia)",
		regionEnum:               "AP_SOUTHEAST_5",
		nativeRegionEnum:         "AP_SOUTHEAST_5",
		regionForReplicationEnum: "",
	},
	RegionApSouth1: {
		name:                     "ap-south-1",
		displayName:              "Asia Pacific (Mumbai)",
		regionEnum:               "AP_SOUTH_1",
		nativeRegionEnum:         "AP_SOUTH_1",
		regionForReplicationEnum: "AP_SOUTH_1",
	},
	RegionApSouth2: {
		name:                     "ap-south-2",
		displayName:              "Asia Pacific (Hyderabad)",
		regionEnum:               "AP_SOUTH_2",
		nativeRegionEnum:         "AP_SOUTH_2",
		regionForReplicationEnum: "AP_SOUTH_2",
	},
	RegionCaCentral1: {
		name:                     "ca-central-1",
		displayName:              "Canada (Central)",
		regionEnum:               "CA_CENTRAL_1",
		nativeRegionEnum:         "CA_CENTRAL_1",
		regionForReplicationEnum: "CA_CENTRAL_1",
	},
	RegionCaWest1: {
		name:                     "ca-west-1",
		displayName:              "Canada West (Calgary)",
		regionEnum:               "CA_WEST_1",
		nativeRegionEnum:         "CA_WEST_1",
		regionForReplicationEnum: "CA_WEST_1",
	},
	RegionCnNorthWest1: {
		name:                     "cn-northwest-1",
		displayName:              "China (Ningxia)",
		regionEnum:               "CN_NORTHWEST_1",
		nativeRegionEnum:         "CN_NORTHWEST_1",
		regionForReplicationEnum: "CN_NORTHWEST_1",
	},
	RegionCnNorth1: {
		name:                     "cn-north-1",
		displayName:              "China (Beijing)",
		regionEnum:               "CN_NORTH_1",
		nativeRegionEnum:         "CN_NORTH_1",
		regionForReplicationEnum: "CN_NORTH_1",
	},
	RegionEuCentral1: {
		name:                     "eu-central-1",
		displayName:              "Europe (Frankfurt)",
		regionEnum:               "EU_CENTRAL_1",
		nativeRegionEnum:         "EU_CENTRAL_1",
		regionForReplicationEnum: "EU_CENTRAL_1",
	},
	RegionEuCentral2: {
		name:                     "eu-central-2",
		displayName:              "Europe (Zurich)",
		regionEnum:               "EU_CENTRAL_2",
		nativeRegionEnum:         "EU_CENTRAL_2",
		regionForReplicationEnum: "",
	},
	RegionEuNorth1: {
		name:                     "eu-north-1",
		displayName:              "Europe (Stockholm)",
		regionEnum:               "EU_NORTH_1",
		nativeRegionEnum:         "EU_NORTH_1",
		regionForReplicationEnum: "EU_NORTH_1",
	},
	RegionEuSouth1: {
		name:                     "eu-south-1",
		displayName:              "Europe (Milan)",
		regionEnum:               "EU_SOUTH_1",
		nativeRegionEnum:         "EU_SOUTH_1",
		regionForReplicationEnum: "EU_SOUTH_1",
	},
	RegionEuSouth2: {
		name:                     "eu-south-2",
		displayName:              "Europe (Spain)",
		regionEnum:               "EU_SOUTH_2",
		nativeRegionEnum:         "EU_SOUTH_2",
		regionForReplicationEnum: "EU_SOUTH_2",
	},
	RegionEuWest1: {
		name:                     "eu-west-1",
		displayName:              "Europe (Ireland)",
		regionEnum:               "EU_WEST_1",
		nativeRegionEnum:         "EU_WEST_1",
		regionForReplicationEnum: "EU_WEST_1",
	},
	RegionEuWest2: {
		name:                     "eu-west-2",
		displayName:              "Europe (London)",
		regionEnum:               "EU_WEST_2",
		nativeRegionEnum:         "EU_WEST_2",
		regionForReplicationEnum: "EU_WEST_2",
	},
	RegionEuWest3: {
		name:                     "eu-west-3",
		displayName:              "Europe (Paris)",
		regionEnum:               "EU_WEST_3",
		nativeRegionEnum:         "EU_WEST_3",
		regionForReplicationEnum: "EU_WEST_3",
	},
	RegionIlCentral1: {
		name:                     "il-central-1",
		displayName:              "Israel (Tel Aviv)",
		regionEnum:               "IL_CENTRAL_1",
		nativeRegionEnum:         "IL_CENTRAL_1",
		regionForReplicationEnum: "IL_CENTRAL_1",
	},
	RegionMeCentral1: {
		name:                     "me-central-1",
		displayName:              "Middle East (UAE)",
		regionEnum:               "ME_CENTRAL_1",
		nativeRegionEnum:         "ME_CENTRAL_1",
		regionForReplicationEnum: "ME_CENTRAL_1",
	},
	RegionMeSouth1: {
		name:                     "me-south-1",
		displayName:              "Middle East (Bahrain)",
		regionEnum:               "ME_SOUTH_1",
		nativeRegionEnum:         "ME_SOUTH_1",
		regionForReplicationEnum: "ME_SOUTH_1",
	},
	RegionSaEast1: {
		name:                     "sa-east-1",
		displayName:              "South America (São Paulo)",
		regionEnum:               "SA_EAST_1",
		nativeRegionEnum:         "SA_EAST_1",
		regionForReplicationEnum: "SA_EAST_1",
	},
	RegionUsEast1: {
		name:                     "us-east-1",
		displayName:              "US East (N. Virginia)",
		regionEnum:               "US_EAST_1",
		regionForReplicationEnum: "US_EAST_1",
		nativeRegionEnum:         "US_EAST_1",
	},
	RegionUsEast2: {
		name:                     "us-east-2",
		displayName:              "US East (Ohio)",
		regionEnum:               "US_EAST_2",
		nativeRegionEnum:         "US_EAST_2",
		regionForReplicationEnum: "US_EAST_2",
	},
	RegionUsGovEast1: {
		name:                     "us-gov-east-1",
		displayName:              "AWS GovCloud (US-East)",
		regionEnum:               "US_GOV_EAST_1",
		nativeRegionEnum:         "US_GOV_EAST_1",
		regionForReplicationEnum: "US_GOV_EAST_1",
	},
	RegionUsGovWest1: {
		name:                     "us-gov-west-1",
		displayName:              "AWS GovCloud (US-West)",
		regionEnum:               "US_GOV_WEST_1",
		nativeRegionEnum:         "US_GOV_WEST_1",
		regionForReplicationEnum: "US_GOV_WEST_1",
	},
	RegionUsWest1: {
		name:                     "us-west-1",
		displayName:              "US West (N. California)",
		regionEnum:               "US_WEST_1",
		nativeRegionEnum:         "US_WEST_1",
		regionForReplicationEnum: "US_WEST_1",
	},
	RegionUsWest2: {
		name:                     "us-west-2",
		displayName:              "US West (Oregon)",
		regionEnum:               "US_WEST_2",
		nativeRegionEnum:         "US_WEST_2",
		regionForReplicationEnum: "US_WEST_2",
	},
	RegionUsISOEast1: {
		name:                     "us-iso-east-1",
		displayName:              "US ISO East",
		regionEnum:               "n/a",
		nativeRegionEnum:         "n/a",
		regionForReplicationEnum: "US_ISO_EAST_1",
	},
	RegionUsISOWest1: {
		name:                     "us-iso-west-1",
		displayName:              "US ISO West",
		regionEnum:               "n/a",
		nativeRegionEnum:         "n/a",
		regionForReplicationEnum: "US_ISO_WEST_1",
	},
	RegionUsISOBEast1: {
		name:                     "us-isob-east-1",
		displayName:              "US ISOB East",
		regionEnum:               "n/a",
		nativeRegionEnum:         "n/a",
		regionForReplicationEnum: "US_ISOB_EAST_1",
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
	RegionAfSouth1:     {},
	RegionApEast1:      {},
	RegionApNorthEast1: {},
	RegionApNorthEast2: {},
	RegionApNorthEast3: {},
	RegionApSouthEast1: {},
	RegionApSouthEast2: {},
	RegionApSouthEast3: {},
	RegionApSouthEast4: {},
	RegionApSouthEast5: {},
	RegionApSouth1:     {},
	RegionApSouth2:     {},
	RegionCaCentral1:   {},
	RegionCaWest1:      {},
	RegionCnNorthWest1: {},
	RegionCnNorth1:     {},
	RegionEuCentral1:   {},
	RegionEuCentral2:   {},
	RegionEuNorth1:     {},
	RegionEuSouth1:     {},
	RegionEuSouth2:     {},
	RegionEuWest1:      {},
	RegionEuWest2:      {},
	RegionEuWest3:      {},
	RegionIlCentral1:   {},
	RegionMeCentral1:   {},
	RegionMeSouth1:     {},
	RegionSaEast1:      {},
	RegionUsEast1:      {},
	RegionUsEast2:      {},
	RegionUsGovEast1:   {},
	RegionUsGovWest1:   {},
	RegionUsWest1:      {},
	RegionUsWest2:      {},
}

// Deprecated: use RegionFromName or RegionFromRegionEnum.
func ParseRegion(value string) (Region, error) {
	// Polaris region name.
	region := RegionFromRegionEnum(value)
	if _, ok := validRegions[region]; ok {
		return region, nil
	}

	// AWS region name.
	region = RegionFromName(value)
	if _, ok := validRegions[region]; ok {
		return region, nil
	}

	return RegionUnknown, errors.New("invalid aws region")
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
