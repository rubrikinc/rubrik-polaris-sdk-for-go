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
)

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
	RegionAustraliaSouthEast1
	RegionAustraliaSouthEast2
	RegionEU
	RegionEur4
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
	RegionMECentral1
	RegionMECentral2
	RegionMEWest1
	RegionNAM4
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
	RegionUSSouth1
	RegionUSWest1
	RegionUSWest2
	RegionUSWest3
	RegionUSWest4
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

// ToRegionEnum returns the RSC GraphQL GcpRegion enum value for the region.
func (region Region) ToRegionEnum() RegionEnum {
	return RegionEnum{Region: region}
}

// ToRegionEnumPtr returns the RSC GraphQL GcpRegion enum value for the region
// as a pointer. If the region is unknown, nil is returned.
func (region Region) ToRegionEnumPtr() *RegionEnum {
	if region == RegionUnknown {
		return nil
	}
	return &RegionEnum{Region: region}
}

// String returns the name of the region.
func (region Region) String() string {
	return region.Name()
}

const (
	FromAny         = iota // Parse the value as any of the below formats.
	FromDisplayName        // Parse the value as a region display name.
	FromName               // Parse the value as a region name.
	FromRegionEnum         // Parse the value as a GraphQL GcpRegion enum value.
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
		case (valueFormat == FromAny || valueFormat == FromRegionEnum) && info.regionEnum == value:
			return r
		case (valueFormat == FromAny || valueFormat == FromDisplayName) && info.displayName == value:
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

// RegionFromRegionEnum parses the value as a GraphQL GcpRegion enum value.
func RegionFromRegionEnum(value string) Region {
	return RegionFrom(value, FromRegionEnum)
}

// RegionEnum represents the GraphQL GcpRegion enum type.
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
	name        string
	displayName string
	regionEnum  string
}{
	RegionUnknown: {
		name:        "",
		displayName: "<Unknown>",
		regionEnum:  "UNKNOWN_GCP_REGION",
	},
	RegionAfricaSouth1: {
		name:        "africa-south1",
		displayName: "africa-south1 (Johannesburg, South Africa)",
		regionEnum:  "AFRICA_SOUTH1",
	},
	RegionAsia: {
		name:        "asia",
		displayName: "Data centers in Asia, excluding Hong Kong and Indonesia",
		regionEnum:  "ASIA",
	},
	RegionAsia1: {
		name:        "asia1",
		displayName: "asia-northeast1 (Tokyo, Japan, APAC) and asia-northeast2 (Osaka, Japan, APAC)",
		regionEnum:  "ASIA1",
	},
	RegionAsiaEast1: {
		name:        "asia-east1",
		displayName: "asia-east1 (Changhua County, Taiwan, APAC)",
		regionEnum:  "ASIA_EAST1",
	},
	RegionAsiaEast2: {
		name:        "asia-east2",
		displayName: "asia-east2 (Hong Kong, APAC)",
		regionEnum:  "ASIA_EAST2",
	},
	RegionAsiaNorthEast1: {
		name:        "asia-northeast1",
		displayName: "asia-northeast1 (Tokyo, Japan, APAC)",
		regionEnum:  "ASIA_NORTHEAST1",
	},
	RegionAsiaNorthEast2: {
		name:        "asia-northeast2",
		displayName: "asia-northeast2 (Osaka, Japan, APAC)",
		regionEnum:  "ASIA_NORTHEAST2",
	},
	RegionAsiaNorthEast3: {
		name:        "asia-northeast3",
		displayName: "asia-northeast3 (Seoul, South Korea, APAC)",
		regionEnum:  "ASIA_NORTHEAST3",
	},
	RegionAsiaSouth1: {
		name:        "asia-south1",
		displayName: "asia-south1 (Mumbai, India, APAC)",
		regionEnum:  "ASIA_SOUTH1",
	},
	RegionAsiaSouth2: {
		name:        "asia-south2",
		displayName: "asia-south2 (Delhi, India, APAC)",
		regionEnum:  "ASIA_SOUTH2",
	},
	RegionAsiaSouthEast1: {
		name:        "asia-southeast1",
		displayName: "asia-southeast1 (Jurong West, Singapore, APAC)",
		regionEnum:  "ASIA_SOUTHEAST1",
	},
	RegionAsiaSouthEast2: {
		name:        "asia-southeast2",
		displayName: "asia-southeast2 (Jakarta, Indonesia, APAC)",
		regionEnum:  "ASIA_SOUTHEAST2",
	},
	RegionAustraliaSouthEast1: {
		name:        "australia-southeast1",
		displayName: "australia-southeast1 (Sydney, Australia, APAC)",
		regionEnum:  "AUSTRALIA_SOUTHEAST1",
	},
	RegionAustraliaSouthEast2: {
		name:        "australia-southeast2",
		displayName: "australia-southeast2 (Melbourne, Australia, APAC)",
		regionEnum:  "AUSTRALIA_SOUTHEAST2",
	},
	RegionEU: {
		name:        "eu",
		displayName: "Data centers within member states of the European Union",
		regionEnum:  "EU",
	},
	RegionEur4: {
		name:        "eur4",
		displayName: "europe-north1 (Hamina, Finland, Europe) and europe-west4 (Eemshaven, Netherlands, Europe)",
		regionEnum:  "EUR4",
	},
	RegionEuropeCentral2: {
		name:        "europe-central2",
		displayName: "europe-central2 (Warsaw, Poland, Europe)",
		regionEnum:  "EUROPE_CENTRAL2",
	},
	RegionEuropeNorth1: {
		name:        "europe-north1",
		displayName: "europe-north1 (Hamina, Finland, Europe)",
		regionEnum:  "EUROPE_NORTH1",
	},
	RegionEuropeNorth2: {
		name:        "europe-north2",
		displayName: "europe-north2 (Stockholm, Sweden, Europe)",
		regionEnum:  "EUROPE_NORTH2",
	},
	RegionEuropeSouthWest1: {
		name:        "europe-southwest1",
		displayName: "europe-southwest1 (Madrid, Spain, Europe)",
		regionEnum:  "EUROPE_SOUTHWEST1",
	},
	RegionEuropeWest1: {
		name:        "europe-west1",
		displayName: "europe-west1 (St. Ghislain, Belgium, Europe)",
		regionEnum:  "EUROPE_WEST1",
	},
	RegionEuropeWest2: {
		name:        "europe-west2",
		displayName: "europe-west2 (London, England, Europe)",
		regionEnum:  "EUROPE_WEST2",
	},
	RegionEuropeWest3: {
		name:        "europe-west3",
		displayName: "europe-west3 (Frankfurt, Germany, Europe)",
		regionEnum:  "EUROPE_WEST3",
	},
	RegionEuropeWest4: {
		name:        "europe-west4",
		displayName: "europe-west4 (Eemshaven, Netherlands, Europe)",
		regionEnum:  "EUROPE_WEST4",
	},
	RegionEuropeWest6: {
		name:        "europe-west6",
		displayName: "europe-west6 (Zurich, Switzerland, Europe)",
		regionEnum:  "EUROPE_WEST6",
	},
	RegionEuropeWest8: {
		name:        "europe-west8",
		displayName: "europe-west8 (Milan, Italy, Europe)",
		regionEnum:  "EUROPE_WEST8",
	},
	RegionEuropeWest9: {
		name:        "europe-west9",
		displayName: "europe-west9 (Paris, France, Europe)",
		regionEnum:  "EUROPE_WEST9",
	},
	RegionEuropeWest10: {
		name:        "europe-west10",
		displayName: "europe-west10 (Berlin, Germany, Europe)",
		regionEnum:  "EUROPE_WEST10",
	},
	RegionEuropeWest12: {
		name:        "europe-west12",
		displayName: "europe-west12 (Turin, Italy, Europe)",
		regionEnum:  "EUROPE_WEST12",
	},
	RegionMECentral1: {
		name:        "me-central1",
		displayName: "me-central1 (Doha, Qatar, Middle East)",
		regionEnum:  "ME_CENTRAL1",
	},
	RegionMECentral2: {
		name:        "me-central2",
		displayName: "me-central2 (Dammam, Saudi Arabia, Middle East)",
		regionEnum:  "ME_CENTRAL2",
	},
	RegionMEWest1: {
		name:        "me-west1",
		displayName: "me-west1 (Tel Aviv, Israel, Middle East)",
		regionEnum:  "ME_WEST1",
	},
	RegionNAM4: {
		name:        "nam4",
		displayName: "us-central1 (Council Bluffs, Iowa, North America) and us-east1 (Moncks Corner, South Carolina, North America)",
		regionEnum:  "NAM4",
	},
	RegionNorthAmericaNorthEast1: {
		name:        "northamerica-northeast1",
		displayName: "northamerica-northeast1 (Montréal, Québec, North America)",
		regionEnum:  "NORTHAMERICA_NORTHEAST1",
	},
	RegionNorthAmericaNorthEast2: {
		name:        "northamerica-northeast2",
		displayName: "northamerica-northeast2 (Toronto, Ontario, North America)",
		regionEnum:  "NORTHAMERICA_NORTHEAST2",
	},
	RegionNorthAmericaSouth1: {
		name:        "northamerica-south1",
		displayName: "northamerica-south1 (Queretaro, Mexico, North America)",
		regionEnum:  "NORTHAMERICA_SOUTH1",
	},
	RegionSouthAmericaEast1: {
		name:        "southamerica-east1",
		displayName: "southamerica-east1 (Osasco, São Paulo, Brazil, South America)",
		regionEnum:  "SOUTHAMERICA_EAST1",
	},
	RegionSouthAmericaWest1: {
		name:        "southamerica-west1",
		displayName: "southamerica-west1 (Santiago, Chile, South America)",
		regionEnum:  "SOUTHAMERICA_WEST1",
	},
	RegionUS: {
		name:        "us",
		displayName: "Data centers in the United States",
		regionEnum:  "US",
	},
	RegionUSCentral1: {
		name:        "us-central1",
		displayName: "us-central1 (Council Bluffs, Iowa, North America)",
		regionEnum:  "USCENTRAL1",
	},
	RegionUSEast1: {
		name:        "us-east1",
		displayName: "us-east1 (Moncks Corner, South Carolina, North America)",
		regionEnum:  "USEAST1",
	},
	RegionUSEast4: {
		name:        "us-east4",
		displayName: "us-east4 (Ashburn, Virginia, North America)",
		regionEnum:  "USEAST4",
	},
	RegionUSEast5: {
		name:        "us-east5",
		displayName: "us-east5 (Columbus, Ohio, North America)",
		regionEnum:  "US_EAST5",
	},
	RegionUSSouth1: {
		name:        "us-south1",
		displayName: "us-south1 (Dallas, Texas, North America)",
		regionEnum:  "US_SOUTH1",
	},
	RegionUSWest1: {
		name:        "us-west1",
		displayName: "us-west1 (The Dalles, Oregon, North America)",
		regionEnum:  "USWEST1",
	},
	RegionUSWest2: {
		name:        "us-west2",
		displayName: "us-west2 (Los Angeles, California, North America)",
		regionEnum:  "USWEST2",
	},
	RegionUSWest3: {
		name:        "us-west3",
		displayName: "us-west3 (Salt Lake City, Utah, North America)",
		regionEnum:  "US_WEST3",
	},
	RegionUSWest4: {
		name:        "us-west4",
		displayName: "us-west4 (Las Vegas, Nevada, North America)",
		regionEnum:  "US_WEST4",
	},
}
