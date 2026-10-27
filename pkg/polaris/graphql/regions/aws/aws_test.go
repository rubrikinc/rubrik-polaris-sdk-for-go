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
	"testing"
)

func TestRegion(t *testing.T) {
	for region, info := range regionInfoMap {
		// Lookup. Note, the source region returns RegionUnknown.
		if r := RegionFromName(info.name); region != RegionSource {
			if r != region {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if r != RegionUnknown {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), RegionUnknown, RegionUnknown.DisplayName())
			}
		}

		// Marshal/unmarshal. Note, the source region marshals as "n/a", which
		// unmarshal as RegionUnknown.
		buf, err := json.Marshal(region)
		if err != nil {
			t.Fatal(err)
		}
		var r Region
		if err := json.Unmarshal(buf, &r); err != nil {
			t.Fatal(err)
		}
		if region != RegionSource {
			if r != region {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if r != RegionUnknown {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), RegionUnknown, RegionUnknown.DisplayName())
			}
		}
	}
}

func TestRegionEnum(t *testing.T) {
	for region, info := range regionInfoMap {
		regionSupported := info.regionEnum != ""

		// Lookup. The correct region should be returned or RegionUnknown for
		// unsupported regions.
		if r := RegionFromRegionEnum(info.regionEnum); regionSupported {
			if r != region {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if r != RegionUnknown {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), RegionUnknown, RegionUnknown.DisplayName())
			}
		}

		// Marshal/unmarshal. Unsupported regions should fail to marshal.
		buf, err := json.Marshal(region.ToRegionEnum())
		if regionSupported {
			if err != nil {
				t.Fatal(err)
			}

			var enum RegionEnum
			if err := json.Unmarshal(buf, &enum); err != nil {
				t.Fatal(err)
			}
			if r := enum.Region; r != region {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if err == nil {
				t.Errorf("got %s want a marshal error for %q", buf, region)
			}
		}

		// ToPtr. The correct region should be returned or nil for unsupported
		// regions. Note, RegionUnknown can be a supported region but should
		// still return nil.
		if r := region.ToRegionEnumPtr(); regionSupported && region != RegionUnknown {
			if r == nil {
				t.Errorf("got <nil> want %q [%s]", region, region.DisplayName())
			} else if r.Region != region {
				t.Fatalf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if r != nil {
				t.Fatalf("got %q [%s] want <nil>", r, r.DisplayName())
			}
		}
	}
}

func TestNativeRegionEnum(t *testing.T) {
	for region, info := range regionInfoMap {
		regionSupported := info.nativeRegionEnum != ""

		// Lookup. The correct region should be returned or RegionUnknown for
		// unsupported regions.
		if r := RegionFromNativeRegionEnum(info.nativeRegionEnum); regionSupported {
			if r != region {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if r != RegionUnknown {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), RegionUnknown, RegionUnknown.DisplayName())
			}
		}

		// Marshal/unmarshal. Unsupported regions should fail to marshal.
		buf, err := json.Marshal(region.ToNativeRegionEnum())
		if regionSupported {
			if err != nil {
				t.Fatal(err)
			}

			var enum NativeRegionEnum
			if err := json.Unmarshal(buf, &enum); err != nil {
				t.Fatal(err)
			}
			if r := enum.Region; r != region {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if err == nil {
				t.Errorf("got %s want a marshal error for %q", buf, region)
			}
		}

		// ToPtr. The correct region should be returned or nil for unsupported
		// regions. Note, RegionUnknown can be a supported region but should
		// still return nil.
		if r := region.ToNativeRegionEnumPtr(); regionSupported && region != RegionUnknown {
			if r == nil {
				t.Errorf("got <nil> want %q [%s]", region, region.DisplayName())
			} else if r.Region != region {
				t.Fatalf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if r != nil {
				t.Fatalf("got %q [%s] want <nil>", r, r.DisplayName())
			}
		}
	}
}

func TestRegionForReplicationEnum(t *testing.T) {
	for region, info := range regionInfoMap {
		regionSupported := info.regionForReplicationEnum != ""

		// Lookup. The correct region should be returned or RegionUnknown for
		// unsupported regions.
		if r := RegionFromRegionForReplicationEnum(info.regionForReplicationEnum); regionSupported {
			if r != region {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if r != RegionUnknown {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), RegionUnknown, RegionUnknown.DisplayName())
			}
		}

		// Marshal/unmarshal. Unsupported regions should fail to marshal.
		buf, err := json.Marshal(region.ToRegionForReplicationEnum())
		if regionSupported {
			if err != nil {
				t.Fatal(err)
			}

			var enum RegionForReplicationEnum
			if err := json.Unmarshal(buf, &enum); err != nil {
				t.Fatal(err)
			}
			if r := enum.Region; r != region {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if err == nil {
				t.Errorf("got %s want a marshal error for %q", buf, region)
			}
		}

		// ToPtr. The correct region should be returned or nil for unsupported
		// regions. Note, RegionUnknown can be a supported region but should
		// still return nil.
		if r := region.ToRegionForReplicationEnumPtr(); regionSupported && region != RegionUnknown {
			if r == nil {
				t.Errorf("got <nil> want %q [%s]", region, region.DisplayName())
			} else if r.Region != region {
				t.Fatalf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if r != nil {
				t.Fatalf("got %q [%s] want <nil>", r, r.DisplayName())
			}
		}
	}
}

func TestRCSRegionEnum(t *testing.T) {
	for region, info := range regionInfoMap {
		regionSupported := info.rcsRegionEnum != ""

		// Lookup. The correct region should be returned or RegionUnknown for
		// unsupported regions.
		if r := RegionFromRCSRegionEnum(info.rcsRegionEnum); regionSupported {
			if r != region {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if r != RegionUnknown {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), RegionUnknown, RegionUnknown.DisplayName())
			}
		}

		// Marshal/unmarshal. Unsupported regions should fail to marshal.
		buf, err := json.Marshal(region.ToRCSRegionEnum())
		if regionSupported {
			if err != nil {
				t.Fatal(err)
			}

			var enum RCSRegionEnum
			if err := json.Unmarshal(buf, &enum); err != nil {
				t.Fatal(err)
			}
			if r := enum.Region; r != region {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if err == nil {
				t.Errorf("got %s want a marshal error for %q", buf, region)
			}
		}

		// ToPtr. The correct region should be returned or nil for unsupported
		// regions. Note, RegionUnknown can be a supported region but should
		// still return nil.
		if r := region.ToRCSRegionEnumPtr(); regionSupported && region != RegionUnknown {
			if r == nil {
				t.Errorf("got <nil> want %q [%s]", region, region.DisplayName())
			} else if r.Region != region {
				t.Fatalf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if r != nil {
				t.Fatalf("got %q [%s] want <nil>", r, r.DisplayName())
			}
		}
	}
}

func TestBaaSSupportedRegions(t *testing.T) {
	regions := BaaSRegions()
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
