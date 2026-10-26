// Copyright 2025 Rubrik, Inc.
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

package gcp

import (
	"encoding/json"
	"fmt"
	"slices"
)

// Note, the region constants is not guaranteed to keep its numbers between
// versions. They are kept sorted alphabetically.
const (
	RegionUnknown Region = iota
	RegionAfricaSouth1
	RegionAsia
	RegionAsia1
	RegionAsiaEast1
	RegionAsiaEast2
	RegionAsiaNorthEast1
	RegionAsiaNorthEast2
	RegionAsiaNorthEast3
	RegionAsiaSouth1
	RegionAsiaSouth2
	RegionAsiaSouthEast1
	RegionAsiaSouthEast2
	RegionAsiaSouthEast3
	RegionAU
	RegionAustraliaSouthEast1
	RegionAustraliaSouthEast2
	RegionCA
	RegionDE
	RegionEU
	RegionEur3
	RegionEur4
	RegionEur5
	RegionEur6
	RegionEur7
	RegionEur8
	RegionEurope
	RegionEuropeCentral2
	RegionEuropeNorth1
	RegionEuropeNorth2
	RegionEuropeSouthWest1
	RegionEuropeWest1
	RegionEuropeWest2
	RegionEuropeWest3
	RegionEuropeWest4
	RegionEuropeWest6
	RegionEuropeWest8
	RegionEuropeWest9
	RegionEuropeWest10
	RegionEuropeWest12
	RegionGlobal
	RegionIN
	RegionIT
	RegionJP
	RegionMECentral1
	RegionMECentral2
	RegionMEWest1
	RegionNAM3
	RegionNAM4
	RegionNAM6
	RegionNAM7
	RegionNAM8
	RegionNAM9
	RegionNAM10
	RegionNAM11
	RegionNAM12
	RegionNAMEurAsia1
	RegionNorthAmericaNorthEast1
	RegionNorthAmericaNorthEast2
	RegionNorthAmericaSouth1
	RegionSouthAmericaEast1
	RegionSouthAmericaWest1
	RegionUS
	RegionUSCentral1
	RegionUSEast1
	RegionUSEast4
	RegionUSEast5
	RegionUSEast7
	RegionUSSouth1
	RegionUSWest1
	RegionUSWest2
	RegionUSWest3
	RegionUSWest4
	RegionUSWest8
)

// Region represents a GCP region in RSC. When reading a Region from a JSON
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

// ToCloudAccountRegionEnum returns the RSC GraphQL GcpCloudAccountRegion enum
// value for the region.
func (region Region) ToCloudAccountRegionEnum() CloudAccountRegionEnum {
	return CloudAccountRegionEnum{Region: region}
}

// ToCloudAccountRegionEnumPtr returns the RSC GraphQL GcpCloudAccountRegion
// enum value for the region as a pointer. If the region is unknown or has no
// GcpCloudAccountRegion value, nil is returned.
func (region Region) ToCloudAccountRegionEnumPtr() *CloudAccountRegionEnum {
	if region == RegionUnknown || regionInfoMap[region].cloudAccountRegionEnum == "" {
		return nil
	}
	return &CloudAccountRegionEnum{Region: region}
}

// ToRegionEnum returns the RSC GraphQL GcpRegion enum value for the region.
func (region Region) ToRegionEnum() RegionEnum {
	return RegionEnum{Region: region}
}

// ToRegionEnumPtr returns the RSC GraphQL GcpRegion enum value for the region
// as a pointer. If the region is unknown or has no GcpRegion value, nil is
// returned.
func (region Region) ToRegionEnumPtr() *RegionEnum {
	if region == RegionUnknown || regionInfoMap[region].regionEnum == "" {
		return nil
	}
	return &RegionEnum{Region: region}
}

// ToRCSRegionEnum returns the RSC GraphQL RcsRegionEnumType enum value for the region.
func (region Region) ToRCSRegionEnum() RCSRegionEnum {
	return RCSRegionEnum{Region: region}
}

// ToRCSRegionEnumPtr returns the RSC GraphQL RcsRegionEnumType enum value for
// the region as a pointer. If the region is unknown or has no RcsRegionEnumType
// value, nil is returned.
func (region Region) ToRCSRegionEnumPtr() *RCSRegionEnum {
	if region == RegionUnknown || regionInfoMap[region].rcsRegionEnum == "" {
		return nil
	}
	return &RCSRegionEnum{Region: region}
}

// ToKMSLocationEnum returns the RSC GraphQL GcpKmsLocation enum value for the
// region.
func (region Region) ToKMSLocationEnum() KMSLocationEnum {
	return KMSLocationEnum{Region: region}
}

// ToKMSLocationEnumPtr returns the RSC GraphQL GcpKmsLocation enum value for
// the region as a pointer. If the region is unknown or has no GcpKmsLocation
// value, nil is returned.
func (region Region) ToKMSLocationEnumPtr() *KMSLocationEnum {
	if region == RegionUnknown || regionInfoMap[region].kmsLocationEnum == "" {
		return nil
	}
	return &KMSLocationEnum{Region: region}
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
	FromAny                    = iota // Parse the value as any of the below formats.
	FromCloudAccountRegionEnum        // Parse the value as a GraphQL GcpCloudAccountRegion enum value.
	FromDisplayName                   // Parse the value as a region display name.
	FromName                          // Parse the value as a region name.
	FromRegionEnum                    // Parse the value as a GraphQL GcpRegion enum value.
	FromRCSRegionEnum                 // Parse the value as a GraphQL RcsRegionEnumType enum value.
	FromKMSLocationEnum               // Parse the value as a GraphQL GcpKmsLocation enum value.
)

// RegionFrom parses the value as a region identifier in the specified format.
// If the value isn't recognized, RegionUnknown is returned.
func RegionFrom(value string, valueFormat int) Region {
	if value == "" || value == "n/a" {
		return RegionUnknown
	}
	for r, info := range regionInfoMap {
		switch {
		case (valueFormat == FromAny || valueFormat == FromCloudAccountRegionEnum) && info.cloudAccountRegionEnum == value:
			return r
		case (valueFormat == FromAny || valueFormat == FromName) && info.name == value:
			return r
		case (valueFormat == FromAny || valueFormat == FromRegionEnum) && info.regionEnum == value:
			return r
		case (valueFormat == FromAny || valueFormat == FromDisplayName) && info.displayName == value:
			return r
		case (valueFormat == FromAny || valueFormat == FromRCSRegionEnum) && info.rcsRegionEnum == value:
			return r
		case (valueFormat == FromAny || valueFormat == FromKMSLocationEnum) && info.kmsLocationEnum == value:
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

// RegionFromCloudAccountRegionEnum parses the value as a GraphQL
// GcpCloudAccountRegion enum value.
func RegionFromCloudAccountRegionEnum(value string) Region {
	return RegionFrom(value, FromCloudAccountRegionEnum)
}

// RegionFromRegionEnum parses the value as a GraphQL GcpRegion enum value.
func RegionFromRegionEnum(value string) Region {
	return RegionFrom(value, FromRegionEnum)
}

// RegionFromRCSRegionEnum parses the value as a GraphQL RcsRegionEnumType enum value.
func RegionFromRCSRegionEnum(value string) Region {
	return RegionFrom(value, FromRCSRegionEnum)
}

// RegionFromKMSLocationEnum parses the value as a GraphQL GcpKmsLocation enum
// value.
func RegionFromKMSLocationEnum(value string) Region {
	return RegionFrom(value, FromKMSLocationEnum)
}

// RegionEnum represents the GraphQL GcpRegion enum type.
type RegionEnum struct{ Region }

// MarshalJSON returns the region as a JSON string. An error is returned if the
// region has no GcpRegion value.
func (region RegionEnum) MarshalJSON() ([]byte, error) {
	value := regionInfoMap[region.Region].regionEnum
	if value == "" {
		return nil, fmt.Errorf("region %q has no GcpRegion value", region.Region.Name())
	}
	return json.Marshal(value)
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

// CloudAccountRegionEnum represents the GraphQL GcpCloudAccountRegion enum
// type.
type CloudAccountRegionEnum struct{ Region }

// MarshalJSON returns the region as a JSON string. An error is returned if the
// region has no GcpCloudAccountRegion value.
func (region CloudAccountRegionEnum) MarshalJSON() ([]byte, error) {
	value := regionInfoMap[region.Region].cloudAccountRegionEnum
	if value == "" {
		return nil, fmt.Errorf("region %q has no GcpCloudAccountRegion value", region.Region.Name())
	}
	return json.Marshal(value)
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

// RCSRegionEnum represents the GraphQL RcsRegionEnumType enum type. This region
// type is primarily used for RCS/RCV GraphQL queries/mutations.
type RCSRegionEnum struct{ Region }

// MarshalJSON returns the region as a JSON string. An error is returned if the
// region has no RcsRegionEnumType value.
func (region RCSRegionEnum) MarshalJSON() ([]byte, error) {
	value := regionInfoMap[region.Region].rcsRegionEnum
	if value == "" {
		return nil, fmt.Errorf("region %q has no RcsRegionEnumType value", region.Region.Name())
	}
	return json.Marshal(value)
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

// KMSLocationEnum represents the GraphQL GcpKmsLocation enum type. This region
// type is used for GCP Cloud KMS key manager GraphQL queries/mutations. Besides
// regions, it holds multi-regions, dual-regions and the global location.
type KMSLocationEnum struct{ Region }

// MarshalJSON returns the region as a JSON string. An error is returned if the
// region has no GcpKmsLocation value. The GcpKmsLocation enum has no value for
// an unknown location, so this includes RegionUnknown.
func (region KMSLocationEnum) MarshalJSON() ([]byte, error) {
	value := regionInfoMap[region.Region].kmsLocationEnum
	if value == "" {
		return nil, fmt.Errorf("region %q has no GcpKmsLocation value", region.Region.Name())
	}
	return json.Marshal(value)
}

// UnmarshalJSON parses the region from a JSON string.
func (region *KMSLocationEnum) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	region.Region = RegionFromKMSLocationEnum(s)
	return nil
}

// String returns the string representation of the KMSLocationEnum. The
// GcpKmsLocation enum has no value for an unknown location, so an empty string
// is returned for a region which isn't a GCP KMS location.
func (region KMSLocationEnum) String() string {
	return regionInfoMap[region.Region].kmsLocationEnum
}

// AllRegionNames returns all the recognized region names.
func AllRegionNames() []string {
	regions := make([]string, 0, len(regionInfoMap))
	for region, info := range regionInfoMap {
		if region != RegionUnknown {
			regions = append(regions, info.name)
		}
	}

	slices.Sort(regions)
	return regions
}

// AllRCSRegionNames returns the names of all regions supported by RCS/RCV.
func AllRCSRegionNames() []string {
	regions := make([]string, 0, len(regionInfoMap))
	for region, info := range regionInfoMap {
		if region != RegionUnknown && info.rcsRegionEnum != "" {
			regions = append(regions, info.name)
		}
	}

	slices.Sort(regions)
	return regions
}

// AllKMSLocationNames returns the names of all regions which are GCP KMS
// locations, e.g. us-east1, europe and global.
func AllKMSLocationNames() []string {
	regions := make([]string, 0, len(regionInfoMap))
	for _, info := range regionInfoMap {
		if info.kmsLocationEnum != "" {
			regions = append(regions, info.name)
		}
	}

	return regions
}

var regionInfoMap = map[Region]struct {
	name                   string
	displayName            string
	regionEnum             string
	cloudAccountRegionEnum string
	rcsRegionEnum          string
	kmsLocationEnum        string
}{
	// regionEnum has been deprecated in the GraphQL API but is still required
	// for some endpoints, like creating a source region archival location.
	RegionUnknown: {
		name:          "",
		displayName:   "<Unknown>",
		regionEnum:    "UNKNOWN_GCP_REGION",
		rcsRegionEnum: "UNKNOWN_GCP_REGION",
	},
	RegionAfricaSouth1: {
		name:                   "africa-south1",
		displayName:            "africa-south1 (Johannesburg, South Africa)",
		regionEnum:             "AFRICA_SOUTH1",
		cloudAccountRegionEnum: "AFRICA_SOUTH1",
		rcsRegionEnum:          "SOUTH_AFRICA_NORTH",
		kmsLocationEnum:        "AFRICA_SOUTH1",
	},
	RegionAsia: {
		name:            "asia",
		displayName:     "Data centers in Asia, excluding Hong Kong and Indonesia",
		regionEnum:      "ASIA",
		kmsLocationEnum: "ASIA",
	},
	RegionAsia1: {
		name:            "asia1",
		displayName:     "asia-northeast1 (Tokyo, Japan, APAC) and asia-northeast2 (Osaka, Japan, APAC)",
		regionEnum:      "ASIA1",
		kmsLocationEnum: "ASIA1",
	},
	RegionAsiaEast1: {
		name:                   "asia-east1",
		displayName:            "asia-east1 (Changhua County, Taiwan, APAC)",
		regionEnum:             "ASIA_EAST1",
		cloudAccountRegionEnum: "ASIA_EAST1",
		rcsRegionEnum:          "ASIA_EAST_TAIWAN",
		kmsLocationEnum:        "ASIA_EAST1",
	},
	RegionAsiaEast2: {
		name:                   "asia-east2",
		displayName:            "asia-east2 (Hong Kong, APAC)",
		regionEnum:             "ASIA_EAST2",
		cloudAccountRegionEnum: "ASIA_EAST2",
		rcsRegionEnum:          "ASIA_EAST_2",
		kmsLocationEnum:        "ASIA_EAST2",
	},
	RegionAsiaNorthEast1: {
		name:                   "asia-northeast1",
		displayName:            "asia-northeast1 (Tokyo, Japan, APAC)",
		regionEnum:             "ASIA_NORTHEAST1",
		cloudAccountRegionEnum: "ASIA_NORTHEAST1",
		rcsRegionEnum:          "JAPAN_EAST",
		kmsLocationEnum:        "ASIA_NORTHEAST1",
	},
	RegionAsiaNorthEast2: {
		name:                   "asia-northeast2",
		displayName:            "asia-northeast2 (Osaka, Japan, APAC)",
		regionEnum:             "ASIA_NORTHEAST2",
		cloudAccountRegionEnum: "ASIA_NORTHEAST2",
		rcsRegionEnum:          "JAPAN_WEST",
		kmsLocationEnum:        "ASIA_NORTHEAST2",
	},
	RegionAsiaNorthEast3: {
		name:                   "asia-northeast3",
		displayName:            "asia-northeast3 (Seoul, South Korea, APAC)",
		regionEnum:             "ASIA_NORTHEAST3",
		cloudAccountRegionEnum: "ASIA_NORTHEAST3",
		rcsRegionEnum:          "KOREA_CENTRAL",
		kmsLocationEnum:        "ASIA_NORTHEAST3",
	},
	RegionAsiaSouth1: {
		name:                   "asia-south1",
		displayName:            "asia-south1 (Mumbai, India, APAC)",
		regionEnum:             "ASIA_SOUTH1",
		cloudAccountRegionEnum: "ASIA_SOUTH1",
		rcsRegionEnum:          "INDIA_WEST",
		kmsLocationEnum:        "ASIA_SOUTH1",
	},
	RegionAsiaSouth2: {
		name:                   "asia-south2",
		displayName:            "asia-south2 (Delhi, India, APAC)",
		regionEnum:             "ASIA_SOUTH2",
		cloudAccountRegionEnum: "ASIA_SOUTH2",
		rcsRegionEnum:          "INDIA_SOUTH_2",
		kmsLocationEnum:        "ASIA_SOUTH2",
	},
	RegionAsiaSouthEast1: {
		name:                   "asia-southeast1",
		displayName:            "asia-southeast1 (Jurong West, Singapore, APAC)",
		regionEnum:             "ASIA_SOUTHEAST1",
		cloudAccountRegionEnum: "ASIA_SOUTHEAST1",
		rcsRegionEnum:          "ASIA_SOUTHEAST",
		kmsLocationEnum:        "ASIA_SOUTHEAST1",
	},
	RegionAsiaSouthEast2: {
		name:                   "asia-southeast2",
		displayName:            "asia-southeast2 (Jakarta, Indonesia, APAC)",
		regionEnum:             "ASIA_SOUTHEAST2",
		cloudAccountRegionEnum: "ASIA_SOUTHEAST2",
		rcsRegionEnum:          "ASIA_PACIFIC_JAKARTA",
		kmsLocationEnum:        "ASIA_SOUTHEAST2",
	},
	RegionAsiaSouthEast3: {
		name:                   "asia-southeast3",
		displayName:            "asia-southeast3 (Bangkok, Thailand, APAC)",
		regionEnum:             "ASIA_SOUTHEAST3",
		cloudAccountRegionEnum: "ASIA_SOUTHEAST3",
		kmsLocationEnum:        "ASIA_SOUTHEAST3",
	},
	RegionAU: {
		name:            "au",
		displayName:     "All regions in Australia",
		kmsLocationEnum: "AU",
	},
	RegionAustraliaSouthEast1: {
		name:                   "australia-southeast1",
		displayName:            "australia-southeast1 (Sydney, Australia, APAC)",
		regionEnum:             "AUSTRALIA_SOUTHEAST1",
		cloudAccountRegionEnum: "AUSTRALIA_SOUTHEAST1",
		rcsRegionEnum:          "ASIA_PACIFIC_SYDNEY",
		kmsLocationEnum:        "AUSTRALIA_SOUTHEAST1",
	},
	RegionAustraliaSouthEast2: {
		name:                   "australia-southeast2",
		displayName:            "australia-southeast2 (Melbourne, Australia, APAC)",
		regionEnum:             "AUSTRALIA_SOUTHEAST2",
		cloudAccountRegionEnum: "AUSTRALIA_SOUTHEAST2",
		rcsRegionEnum:          "ASIA_PACIFIC_MELBOURNE",
		kmsLocationEnum:        "AUSTRALIA_SOUTHEAST2",
	},
	RegionCA: {
		name:            "ca",
		displayName:     "All regions in Canada",
		kmsLocationEnum: "CA",
	},
	RegionDE: {
		name:            "de",
		displayName:     "All regions in Germany",
		kmsLocationEnum: "DE",
	},
	RegionEU: {
		name:        "eu",
		displayName: "Data centers within member states of the European Union",
		regionEnum:  "EU",
	},
	RegionEur3: {
		name:            "eur3",
		displayName:     "St. Ghislain (Belgium) and Eemshaven (Netherlands)",
		kmsLocationEnum: "EUR3",
	},
	RegionEur4: {
		name:            "eur4",
		displayName:     "europe-north1 (Hamina, Finland, Europe) and europe-west4 (Eemshaven, Netherlands, Europe)",
		regionEnum:      "EUR4",
		kmsLocationEnum: "EUR4",
	},
	RegionEur5: {
		name:            "eur5",
		displayName:     "europe-west1 (St. Ghislain, Belgium, Europe) and europe-west2 (London, England, Europe)",
		regionEnum:      "EUR5",
		kmsLocationEnum: "EUR5",
	},
	RegionEur6: {
		name:            "eur6",
		displayName:     "Eemshaven (Netherlands), Frankfurt (Germany), and Zurich (Switzerland)",
		kmsLocationEnum: "EUR6",
	},
	RegionEur7: {
		name:            "eur7",
		displayName:     "europe-west2 (London, England, Europe) and europe-west3 (Frankfurt, Germany, Europe)",
		regionEnum:      "EUR7",
		kmsLocationEnum: "EUR7",
	},
	RegionEur8: {
		name:            "eur8",
		displayName:     "europe-west3 (Frankfurt, Germany, Europe) and europe-west6 (Zurich, Switzerland, Europe)",
		regionEnum:      "EUR8",
		kmsLocationEnum: "EUR8",
	},
	RegionEurope: {
		name:            "europe",
		displayName:     "All regions in the European Union",
		kmsLocationEnum: "EUROPE",
	},
	RegionEuropeCentral2: {
		name:                   "europe-central2",
		displayName:            "europe-central2 (Warsaw, Poland, Europe)",
		regionEnum:             "EUROPE_CENTRAL2",
		cloudAccountRegionEnum: "EUROPE_CENTRAL2",
		rcsRegionEnum:          "POLAND_CENTRAL",
		kmsLocationEnum:        "EUROPE_CENTRAL2",
	},
	RegionEuropeNorth1: {
		name:                   "europe-north1",
		displayName:            "europe-north1 (Hamina, Finland, Europe)",
		regionEnum:             "EUROPE_NORTH1",
		cloudAccountRegionEnum: "EUROPE_NORTH1",
		rcsRegionEnum:          "EUROPE_NORTH_FINLAND",
		kmsLocationEnum:        "EUROPE_NORTH1",
	},
	RegionEuropeNorth2: {
		name:                   "europe-north2",
		displayName:            "europe-north2 (Stockholm, Sweden, Europe)",
		regionEnum:             "EUROPE_NORTH2",
		cloudAccountRegionEnum: "EUROPE_NORTH2",
		kmsLocationEnum:        "EUROPE_NORTH2",
	},
	RegionEuropeSouthWest1: {
		name:                   "europe-southwest1",
		displayName:            "europe-southwest1 (Madrid, Spain, Europe)",
		regionEnum:             "EUROPE_SOUTHWEST1",
		cloudAccountRegionEnum: "EUROPE_SOUTHWEST1",
		rcsRegionEnum:          "SPAIN_CENTRAL",
		kmsLocationEnum:        "EUROPE_SOUTHWEST1",
	},
	RegionEuropeWest1: {
		name:                   "europe-west1",
		displayName:            "europe-west1 (St. Ghislain, Belgium, Europe)",
		regionEnum:             "EUROPE_WEST1",
		cloudAccountRegionEnum: "EUROPE_WEST1",
		rcsRegionEnum:          "EUROPE_WEST_BELGIUM",
		kmsLocationEnum:        "EUROPE_WEST1",
	},
	RegionEuropeWest2: {
		name:                   "europe-west2",
		displayName:            "europe-west2 (London, England, Europe)",
		regionEnum:             "EUROPE_WEST2",
		cloudAccountRegionEnum: "EUROPE_WEST2",
		rcsRegionEnum:          "UK_SOUTH",
		kmsLocationEnum:        "EUROPE_WEST2",
	},
	RegionEuropeWest3: {
		name:                   "europe-west3",
		displayName:            "europe-west3 (Frankfurt, Germany, Europe)",
		regionEnum:             "EUROPE_WEST3",
		cloudAccountRegionEnum: "EUROPE_WEST3",
		rcsRegionEnum:          "GERMANY_WEST_CENTRAL",
		kmsLocationEnum:        "EUROPE_WEST3",
	},
	RegionEuropeWest4: {
		name:                   "europe-west4",
		displayName:            "europe-west4 (Eemshaven, Netherlands, Europe)",
		regionEnum:             "EUROPE_WEST4",
		cloudAccountRegionEnum: "EUROPE_WEST4",
		rcsRegionEnum:          "EUROPE_WEST_4",
		kmsLocationEnum:        "EUROPE_WEST4",
	},
	RegionEuropeWest6: {
		name:                   "europe-west6",
		displayName:            "europe-west6 (Zurich, Switzerland, Europe)",
		regionEnum:             "EUROPE_WEST6",
		cloudAccountRegionEnum: "EUROPE_WEST6",
		rcsRegionEnum:          "SWITZERLAND_NORTH",
		kmsLocationEnum:        "EUROPE_WEST6",
	},
	RegionEuropeWest8: {
		name:                   "europe-west8",
		displayName:            "europe-west8 (Milan, Italy, Europe)",
		regionEnum:             "EUROPE_WEST8",
		cloudAccountRegionEnum: "EUROPE_WEST8",
		rcsRegionEnum:          "ITALY_NORTH",
		kmsLocationEnum:        "EUROPE_WEST8",
	},
	RegionEuropeWest9: {
		name:                   "europe-west9",
		displayName:            "europe-west9 (Paris, France, Europe)",
		regionEnum:             "EUROPE_WEST9",
		cloudAccountRegionEnum: "EUROPE_WEST9",
		rcsRegionEnum:          "FRANCE_CENTRAL",
		kmsLocationEnum:        "EUROPE_WEST9",
	},
	RegionEuropeWest10: {
		name:                   "europe-west10",
		displayName:            "europe-west10 (Berlin, Germany, Europe)",
		regionEnum:             "EUROPE_WEST10",
		cloudAccountRegionEnum: "EUROPE_WEST10",
		rcsRegionEnum:          "GERMANY_NORTH",
		kmsLocationEnum:        "EUROPE_WEST10",
	},
	RegionEuropeWest12: {
		name:                   "europe-west12",
		displayName:            "europe-west12 (Turin, Italy, Europe)",
		regionEnum:             "EUROPE_WEST12",
		cloudAccountRegionEnum: "EUROPE_WEST12",
		rcsRegionEnum:          "EUROPE_WEST_12",
		kmsLocationEnum:        "EUROPE_WEST12",
	},
	RegionGlobal: {
		name:            "global",
		displayName:     "Global",
		kmsLocationEnum: "GLOBAL",
	},
	RegionIN: {
		name:            "in",
		displayName:     "All regions in India",
		kmsLocationEnum: "IN",
	},
	RegionIT: {
		name:            "it",
		displayName:     "All regions in Italy",
		kmsLocationEnum: "IT",
	},
	RegionJP: {
		name:            "jp",
		displayName:     "All regions in Japan",
		kmsLocationEnum: "JP",
	},
	RegionMECentral1: {
		name:                   "me-central1",
		displayName:            "me-central1 (Doha, Qatar, Middle East)",
		regionEnum:             "ME_CENTRAL1",
		cloudAccountRegionEnum: "ME_CENTRAL1",
		rcsRegionEnum:          "QATAR_CENTRAL",
		kmsLocationEnum:        "ME_CENTRAL1",
	},
	RegionMECentral2: {
		name:                   "me-central2",
		displayName:            "me-central2 (Dammam, Saudi Arabia, Middle East)",
		regionEnum:             "ME_CENTRAL2",
		cloudAccountRegionEnum: "ME_CENTRAL2",
		rcsRegionEnum:          "ME_CENTRAL_2",
		kmsLocationEnum:        "ME_CENTRAL2",
	},
	RegionMEWest1: {
		name:                   "me-west1",
		displayName:            "me-west1 (Tel Aviv, Israel, Middle East)",
		regionEnum:             "ME_WEST1",
		cloudAccountRegionEnum: "ME_WEST1",
		rcsRegionEnum:          "ISRAEL_CENTRAL",
		kmsLocationEnum:        "ME_WEST1",
	},
	RegionNAM3: {
		name:            "nam3",
		displayName:     "Ashburn (Virginia) and Moncks Corner (South Carolina)",
		kmsLocationEnum: "NAM3",
	},
	RegionNAM4: {
		name:            "nam4",
		displayName:     "us-central1 (Council Bluffs, Iowa, North America) and us-east1 (Moncks Corner, South Carolina, North America)",
		regionEnum:      "NAM4",
		kmsLocationEnum: "NAM4",
	},
	RegionNAM6: {
		name:            "nam6",
		displayName:     "Council Bluffs (Iowa) and Moncks Corner (South Carolina)",
		kmsLocationEnum: "NAM6",
	},
	RegionNAM7: {
		name:            "nam7",
		displayName:     "Council Bluffs (Iowa), Ashburn (Virginia), and Pryor (Oklahoma)",
		kmsLocationEnum: "NAM7",
	},
	RegionNAM8: {
		name:            "nam8",
		displayName:     "Los Angeles (California), The Dalles (Oregon), and Salt Lake City (Utah)",
		kmsLocationEnum: "NAM8",
	},
	RegionNAM9: {
		name:            "nam9",
		displayName:     "Ashburn (Virginia) and Council Bluffs (Iowa)",
		kmsLocationEnum: "NAM9",
	},
	RegionNAM10: {
		name:            "nam10",
		displayName:     "Council Bluffs (Iowa), Salt Lake City (Utah), and Pryor (Oklahoma)",
		kmsLocationEnum: "NAM10",
	},
	RegionNAM11: {
		name:            "nam11",
		displayName:     "Council Bluffs (Iowa), Moncks Corner (South Carolina), and Pryor (Oklahoma)",
		kmsLocationEnum: "NAM11",
	},
	RegionNAM12: {
		name:            "nam12",
		displayName:     "Council Bluffs (Iowa), Ashburn (Virginia), Pryor (Oklahoma), and The Dalles (Oregon)",
		kmsLocationEnum: "NAM12",
	},
	RegionNAMEurAsia1: {
		name:            "nam-eur-asia1",
		displayName:     "North America, Europe, and Asia (Council Bluffs (Iowa), Pryor (Oklahoma), St. Ghislain (Belgium), and Changhua County (Taiwan))",
		kmsLocationEnum: "NAM_EUR_ASIA1",
	},
	RegionNorthAmericaNorthEast1: {
		name:                   "northamerica-northeast1",
		displayName:            "northamerica-northeast1 (Montréal, Québec, North America)",
		regionEnum:             "NORTHAMERICA_NORTHEAST1",
		cloudAccountRegionEnum: "NORTHAMERICA_NORTHEAST1",
		rcsRegionEnum:          "NORTHAMERICA_NORTHEAST_1",
		kmsLocationEnum:        "NORTHAMERICA_NORTHEAST1",
	},
	RegionNorthAmericaNorthEast2: {
		name:                   "northamerica-northeast2",
		displayName:            "northamerica-northeast2 (Toronto, Ontario, North America)",
		regionEnum:             "NORTHAMERICA_NORTHEAST2",
		cloudAccountRegionEnum: "NORTHAMERICA_NORTHEAST2",
		rcsRegionEnum:          "CANADA_CENTRAL",
		kmsLocationEnum:        "NORTHAMERICA_NORTHEAST2",
	},
	RegionNorthAmericaSouth1: {
		name:                   "northamerica-south1",
		displayName:            "northamerica-south1 (Queretaro, Mexico, North America)",
		regionEnum:             "NORTHAMERICA_SOUTH1",
		cloudAccountRegionEnum: "NORTHAMERICA_SOUTH1",
		rcsRegionEnum:          "NORTHAMERICA_SOUTH_1",
		kmsLocationEnum:        "NORTHAMERICA_SOUTH1",
	},
	RegionSouthAmericaEast1: {
		name:                   "southamerica-east1",
		displayName:            "southamerica-east1 (Osasco, São Paulo, Brazil, South America)",
		regionEnum:             "SOUTHAMERICA_EAST1",
		cloudAccountRegionEnum: "SOUTHAMERICA_EAST1",
		rcsRegionEnum:          "BRAZIL_SOUTH",
		kmsLocationEnum:        "SOUTHAMERICA_EAST1",
	},
	RegionSouthAmericaWest1: {
		name:                   "southamerica-west1",
		displayName:            "southamerica-west1 (Santiago, Chile, South America)",
		regionEnum:             "SOUTHAMERICA_WEST1",
		cloudAccountRegionEnum: "SOUTHAMERICA_WEST1",
		rcsRegionEnum:          "SOUTHAMERICA_WEST_1",
		kmsLocationEnum:        "SOUTHAMERICA_WEST1",
	},
	RegionUS: {
		name:            "us",
		displayName:     "Data centers in the United States",
		regionEnum:      "US",
		kmsLocationEnum: "US",
	},
	RegionUSCentral1: {
		name:                   "us-central1",
		displayName:            "us-central1 (Council Bluffs, Iowa, North America)",
		regionEnum:             "USCENTRAL1",
		cloudAccountRegionEnum: "US_CENTRAL1",
		rcsRegionEnum:          "US_CENTRAL",
		kmsLocationEnum:        "US_CENTRAL1",
	},
	RegionUSEast1: {
		name:                   "us-east1",
		displayName:            "us-east1 (Moncks Corner, South Carolina, North America)",
		regionEnum:             "USEAST1",
		cloudAccountRegionEnum: "US_EAST1",
		rcsRegionEnum:          "US_EAST_1",
		kmsLocationEnum:        "US_EAST1",
	},
	RegionUSEast4: {
		name:                   "us-east4",
		displayName:            "us-east4 (Ashburn, Virginia, North America)",
		regionEnum:             "USEAST4",
		cloudAccountRegionEnum: "US_EAST4",
		rcsRegionEnum:          "US_EAST",
		kmsLocationEnum:        "US_EAST4",
	},
	RegionUSEast5: {
		name:                   "us-east5",
		displayName:            "us-east5 (Columbus, Ohio, North America)",
		regionEnum:             "US_EAST5",
		cloudAccountRegionEnum: "US_EAST5",
		rcsRegionEnum:          "US_EAST_5",
		kmsLocationEnum:        "US_EAST5",
	},
	RegionUSEast7: {
		name:                   "us-east7",
		displayName:            "us-east7 (Bridgeport, Alabama, North America)",
		regionEnum:             "US_EAST7",
		cloudAccountRegionEnum: "US_EAST7",
		rcsRegionEnum:          "US_EAST_7",
	},
	RegionUSSouth1: {
		name:                   "us-south1",
		displayName:            "us-south1 (Dallas, Texas, North America)",
		regionEnum:             "US_SOUTH1",
		cloudAccountRegionEnum: "US_SOUTH1",
		rcsRegionEnum:          "US_SOUTH_1",
		kmsLocationEnum:        "US_SOUTH1",
	},
	RegionUSWest1: {
		name:                   "us-west1",
		displayName:            "us-west1 (The Dalles, Oregon, North America)",
		regionEnum:             "USWEST1",
		cloudAccountRegionEnum: "US_WEST1",
		rcsRegionEnum:          "US_WEST_1",
		kmsLocationEnum:        "US_WEST1",
	},
	RegionUSWest2: {
		name:                   "us-west2",
		displayName:            "us-west2 (Los Angeles, California, North America)",
		regionEnum:             "USWEST2",
		cloudAccountRegionEnum: "US_WEST2",
		rcsRegionEnum:          "US_WEST_LOS_ANGELES",
		kmsLocationEnum:        "US_WEST2",
	},
	RegionUSWest3: {
		name:                   "us-west3",
		displayName:            "us-west3 (Salt Lake City, Utah, North America)",
		regionEnum:             "US_WEST3",
		cloudAccountRegionEnum: "US_WEST3",
		rcsRegionEnum:          "US_WEST_3",
		kmsLocationEnum:        "US_WEST3",
	},
	RegionUSWest4: {
		name:                   "us-west4",
		displayName:            "us-west4 (Las Vegas, Nevada, North America)",
		regionEnum:             "US_WEST4",
		cloudAccountRegionEnum: "US_WEST4",
		rcsRegionEnum:          "US_WEST_4",
		kmsLocationEnum:        "US_WEST4",
	},
	RegionUSWest8: {
		name:                   "us-west8",
		displayName:            "us-west8 (Mesa, Arizona, North America)",
		regionEnum:             "US_WEST8",
		cloudAccountRegionEnum: "US_WEST8",
		rcsRegionEnum:          "US_WEST_8",
	},
}
