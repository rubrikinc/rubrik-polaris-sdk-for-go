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
	"encoding/json"
	"fmt"
	"slices"
	"testing"
)

func TestRegion(t *testing.T) {
	unsupported := []Region{
		RegionSource,
	}

	for region, info := range regionInfoMap {
		want := region
		if slices.Contains(unsupported, region) {
			want = RegionUnknown
		}

		// Lookup.
		if r := RegionFromName(info.name); r != want {
			t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), want, want.DisplayName())
		}

		// Marshal/unmarshal.
		var r Region
		buf, err := json.Marshal(region)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(buf, &r); err != nil {
			t.Fatal(err)
		}
		if r != want {
			t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), want, want.DisplayName())
		}
	}
}

func TestRegionEnum(t *testing.T) {
	unsupported := []Region{
		RegionSource,
		RegionUsISOEast1,
		RegionUsISOWest1,
		RegionUsISOBEast1,
	}

	for region, info := range regionInfoMap {
		want := region
		if slices.Contains(unsupported, region) {
			want = RegionUnknown
		}

		// Lookup.
		if r := RegionFromRegionEnum(info.regionEnum); r != want {
			t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), want, want.DisplayName())
		}

		// Marshal/unmarshal.
		var enum RegionEnum
		buf, err := json.Marshal(region.ToRegionEnum())
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(buf, &enum); err != nil {
			t.Fatal(err)
		}
		if r := enum.Region; r != want {
			t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), want, want.DisplayName())
		}

		// ToPtr.
		if r := region.ToRegionEnumPtr(); (r == nil && want != RegionUnknown) || (r != nil && *r != want.ToRegionEnum()) {
			got := "<nil>"
			if r != nil {
				got = fmt.Sprintf("%q", r.Name())
			}
			t.Errorf("got %s want %q", got, want)
		}
	}
}

func TestNativeRegionEnum(t *testing.T) {
	unsupported := []Region{
		RegionSource,
	}

	for region, info := range regionInfoMap {
		want := region
		if slices.Contains(unsupported, region) {
			want = RegionUnknown
		}

		// Lookup.
		if r := RegionFromNativeRegionEnum(info.nativeRegionEnum); r != want {
			t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), want, want.DisplayName())
		}

		// Marshal/unmarshal.
		var reg NativeRegionEnum
		buf, err := json.Marshal(region.ToNativeRegionEnum())
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(buf, &reg); err != nil {
			t.Fatal(err)
		}
		if r := reg.Region; r != want {
			t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), want, want.DisplayName())
		}

		// ToPtr.
		if r := region.ToNativeRegionEnumPtr(); (r == nil && want != RegionUnknown) || (r != nil && *r != want.ToNativeRegionEnum()) {
			got := "<nil>"
			if r != nil {
				got = fmt.Sprintf("%q", r.Name())
			}
			t.Errorf("got %s want %q", got, want)
		}
	}
}

func TestRegionForReplicationEnum(t *testing.T) {
	unsupported := []Region{
		RegionEuCentral2,
		RegionMxCentral1,
		RegionApSouthEast5,
		RegionApSouthEast7,
	}

	for region, info := range regionInfoMap {
		want := region
		if slices.Contains(unsupported, region) {
			want = RegionUnknown
		}

		// Lookup.
		if r := RegionFromRegionForReplicationEnum(info.regionForReplicationEnum); r != want {
			t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), want, want.DisplayName())
		}

		// Marshal/unmarshal.
		var reg RegionForReplicationEnum
		buf, err := json.Marshal(region.ToRegionForReplicationEnum())
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(buf, &reg); err != nil {
			t.Fatal(err)
		}
		if r := reg.Region; r != want {
			t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), want, want.DisplayName())
		}

		// ToPtr.
		if r := region.ToRegionForReplicationEnumPtr(); (r == nil && want != RegionUnknown) || (r != nil && *r != want.ToRegionForReplicationEnum()) {
			got := "<nil>"
			if r != nil {
				got = fmt.Sprintf("%q", r.Name())
			}
			t.Errorf("got %s want %q", got, want)
		}
	}
}

func TestRCSRegionEnum(t *testing.T) {
	unsupported := []Region{
		RegionSource,
		RegionCnNorth1,
		RegionCnNorthWest1,
		RegionUsISOEast1,
		RegionUsISOWest1,
		RegionUsISOBEast1,
	}

	for region, info := range regionInfoMap {
		want := region
		if slices.Contains(unsupported, region) {
			want = RegionUnknown
		}

		// Lookup.
		if r := RegionFromRCSRegionEnum(info.rcsRegionEnum); r != want {
			t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), want, want.DisplayName())
		}

		// Marshal/unmarshal.
		var enum RCSRegionEnum
		buf, err := json.Marshal(region.ToRCSRegionEnum())
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(buf, &enum); err != nil {
			t.Fatal(err)
		}
		if r := enum.Region; r != want {
			t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), want, want.DisplayName())
		}

		// ToPtr.
		if r := region.ToRCSRegionEnumPtr(); (r == nil && want != RegionUnknown) || (r != nil && *r != want.ToRCSRegionEnum()) {
			got := "<nil>"
			if r != nil {
				got = fmt.Sprintf("%q", r.Name())
			}
			t.Errorf("got %s want %q", got, want)
		}
	}
}

func TestBaaSSupportedRegions(t *testing.T) {
	regions := BaaSSupportedRegions()
	if len(regions) == 0 {
		t.Fatal("expected a non-empty BaaS supported region set")
	}

	// The result must be sorted by name and consistent with the per-region flag.
	for i, region := range regions {
		if !region.BaaSSupported() {
			t.Errorf("region %q returned by BaaSSupportedRegions is not flagged as supported", region.Name())
		}
		if i > 0 && regions[i-1].Name() >= region.Name() {
			t.Errorf("regions are not sorted by name: %q before %q", regions[i-1].Name(), region.Name())
		}
	}

	// Commercial regions supported by BaaS.
	for _, region := range []Region{RegionUsEast1, RegionApSouthEast7, RegionMxCentral1} {
		if !region.BaaSSupported() {
			t.Errorf("expected %q to be BaaS supported", region.Name())
		}
	}

	// GovCloud, China, ISO and Middle East regions are excluded.
	for _, region := range []Region{RegionUsGovEast1, RegionCnNorth1, RegionUsISOEast1, RegionMeCentral1, RegionMeSouth1} {
		if region.BaaSSupported() {
			t.Errorf("expected %q to be excluded from BaaS", region.Name())
		}
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
