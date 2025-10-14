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

package aws

import (
	"reflect"
	"testing"
)

func TestFormatRegion(t *testing.T) {
	region := FormatRegion(RegionEuNorth1)
	if region != "eu-north-1" {
		t.Errorf("invalid region: %v", region)
	}

	regions := FormatRegions([]Region{RegionUsEast1, RegionUsWest1})
	if !reflect.DeepEqual(regions, []string{"us-east-1", "us-west-1"}) {
		t.Errorf("invalid regions: %v", regions)
	}
}

func TestParseRegion(t *testing.T) {
	if region := ParseRegionNoValidation("eu-north-1"); region != RegionEuNorth1 {
		t.Errorf("invalid region: %v", region)
	}

	regions := ParseRegionsNoValidation([]string{"us-east-1", "us-west-1"})
	if !reflect.DeepEqual(regions, []Region{RegionUsEast1, RegionUsWest1}) {
		t.Errorf("invalid region: %v", regions)
	}
}

func TestRegionsForReplication(t *testing.T) {
	if region := RegionFromRegionForReplicationEnum("US_WEST_2"); region != RegionUsWest2 {
		t.Errorf("invalid region: %v", region)
	}

	if region := RegionUsWest2.ToRegionForReplicationEnum(); region.Region != RegionUsWest2 {
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

	if region := RegionFromName("n/a"); region != RegionUnknown {
		t.Errorf("invalid region: %v", region)
	}
}
