// Copyright 2026 Rubrik, Inc.
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

package gcp

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestRegion(t *testing.T) {
	for region, info := range regionInfoMap {
		// Lookup.
		if r := RegionFromName(info.name); r != region {
			t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
		}

		// Marshal/unmarshal.
		buf, err := json.Marshal(region)
		if err != nil {
			t.Fatal(err)
		}
		var r Region
		if err := json.Unmarshal(buf, &r); err != nil {
			t.Fatal(err)
		}
		if r != region {
			t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
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

func TestCloudAccountRegionEnum(t *testing.T) {
	for region, info := range regionInfoMap {
		regionSupported := info.cloudAccountRegionEnum != ""

		// Lookup. The correct region should be returned or RegionUnknown for
		// unsupported regions.
		if r := RegionFromCloudAccountRegionEnum(info.cloudAccountRegionEnum); regionSupported {
			if r != region {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if r != RegionUnknown {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), RegionUnknown, RegionUnknown.DisplayName())
			}
		}

		// Marshal/unmarshal. Unsupported regions should fail to marshal.
		buf, err := json.Marshal(region.ToCloudAccountRegionEnum())
		if regionSupported {
			if err != nil {
				t.Fatal(err)
			}

			var enum CloudAccountRegionEnum
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
		if r := region.ToCloudAccountRegionEnumPtr(); regionSupported && region != RegionUnknown {
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

func TestKMSLocationEnum(t *testing.T) {
	for region, info := range regionInfoMap {
		regionSupported := info.kmsLocationEnum != ""

		// Lookup. The correct region should be returned or RegionUnknown for
		// unsupported regions.
		if r := RegionFromKMSLocationEnum(info.kmsLocationEnum); regionSupported {
			if r != region {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if r != RegionUnknown {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), RegionUnknown, RegionUnknown.DisplayName())
			}
		}

		// Marshal/unmarshal. Unsupported regions should fail to marshal.
		buf, err := json.Marshal(region.ToKMSLocationEnum())
		if regionSupported {
			if err != nil {
				t.Fatal(err)
			}

			var enum KMSLocationEnum
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
		// regions.
		if r := region.ToKMSLocationEnumPtr(); regionSupported && region != RegionUnknown {
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

func TestAllRCSRegionNames(t *testing.T) {
	names := AllRCSRegionNames()
	if !slices.Contains(names, "us-west1") {
		t.Errorf("expected us-west1 in %v", names)
	}
	if slices.Contains(names, "US_WEST_1") {
		t.Errorf("expected region names, not RCS region enum values, got %v", names)
	}
	if slices.Contains(names, "") {
		t.Errorf("expected no unknown region in %v", names)
	}
}
