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
