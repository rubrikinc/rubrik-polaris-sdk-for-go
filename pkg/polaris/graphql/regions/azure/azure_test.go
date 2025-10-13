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
	"encoding/json"
	"reflect"
	"testing"
)

func TestFormatRegion(t *testing.T) {
	region := FormatRegion(RegionNorthEurope)
	if region != "northeurope" {
		t.Errorf("invalid region: %v", region)
	}

	regions := FormatRegions([]Region{RegionEastUS, RegionWestUS})
	if !reflect.DeepEqual(regions, []string{"eastus", "westus"}) {
		t.Errorf("invalid regions: %v", regions)
	}
}

func TestParseRegion(t *testing.T) {
	if region := ParseRegionNoValidation("northeurope"); region != RegionNorthEurope {
		t.Errorf("invalid region: %v", region)
	}

	regions := ParseRegionsNoValidation([]string{"eastus", "westus"})
	if !reflect.DeepEqual(regions, []Region{RegionEastUS, RegionWestUS}) {
		t.Errorf("invalid region: %v", regions)
	}
}

func TestRegionsForReplication(t *testing.T) {
	if region := RegionFromRegionForReplicationEnum("AUSTRALIA_CENTRAL"); region != RegionAustraliaCentral {
		t.Errorf("invalid region: %v", region)
	}
	if region := RegionFromRegionForReplicationEnum("SWEDEN_SOUTH"); region != RegionSwedenSouth {
		t.Errorf("invalid region: %v", region)
	}
	if region := RegionWestUS2.ToRegionForReplicationEnum(); region.Region != RegionWestUS2 {
		t.Errorf("invalid region: %v", region)
	}
	if region := RegionFromRegionForReplicationEnum("SOURCE_REGION"); region != RegionSource {
		t.Errorf("invalid region: %v", region)
	}
	if region := RegionSource.ToRegionForReplicationEnum(); region.Region != RegionSource {
		t.Errorf("invalid region: %v", region)
	}
	if region := RegionFromRegionForReplicationEnum("NOT_DEFINED"); region != RegionUnknown {
		t.Errorf("invalid region: %v", region)
	}
	if region := RegionUnknown.ToRegionForReplicationEnum(); region.Region != RegionUnknown {
		t.Errorf("invalid region: %v", region)
	}
	if region := RegionFromRegionForReplicationEnum(""); region != RegionUnknown {
		t.Errorf("invalid region: %v", region)
	}
	if region := RegionFromRegionForReplicationEnum("n/a"); region != RegionUnknown {
		t.Errorf("invalid region: %v", region)
	}
}

func TestRegionMarshalJSON(t *testing.T) {
	tests := []struct {
		region   Region
		expected string
	}{
		{RegionEastUS, `"eastus"`},
		{RegionWestUS, `"westus"`},
		{RegionNorthEurope, `"northeurope"`},
		{RegionUnknown, `"n/a"`},
	}

	for _, test := range tests {
		data, err := json.Marshal(test.region)
		if err != nil {
			t.Errorf("failed to marshal region %v: %v", test.region, err)
			continue
		}
		if string(data) != test.expected {
			t.Errorf("marshal region %v: expected %s, got %s", test.region, test.expected, string(data))
		}
	}
}

func TestRegionUnmarshalJSON(t *testing.T) {
	tests := []struct {
		json     string
		expected Region
	}{
		{`"eastus"`, RegionEastUS},
		{`"westus"`, RegionWestUS},
		{`"northeurope"`, RegionNorthEurope},
		{`"n/a"`, RegionUnknown},
		{`""`, RegionUnknown},
		{`"invalid-region"`, RegionUnknown},
	}

	for _, test := range tests {
		var region Region
		err := json.Unmarshal([]byte(test.json), &region)
		if err != nil {
			t.Errorf("failed to unmarshal JSON %s: %v", test.json, err)
			continue
		}
		if region != test.expected {
			t.Errorf("unmarshal JSON %s: expected %v, got %v", test.json, test.expected, region)
		}
	}
}

func TestRegionMarshalUnmarshalRoundTrip(t *testing.T) {
	regions := []Region{
		RegionEastUS,
		RegionWestUS,
		RegionNorthEurope,
		RegionAustraliaEast,
		RegionJapanEast,
		RegionUnknown,
	}

	for _, original := range regions {
		// Marshal to JSON
		data, err := json.Marshal(original)
		if err != nil {
			t.Errorf("failed to marshal region %v: %v", original, err)
			continue
		}

		// Unmarshal back to Region
		var unmarshaled Region
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("failed to unmarshal JSON %s: %v", string(data), err)
			continue
		}

		// Check if they match
		if original != unmarshaled {
			t.Errorf("round trip failed for region %v: got %v", original, unmarshaled)
		}
	}
}
