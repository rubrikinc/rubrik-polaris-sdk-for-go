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
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestRegion(t *testing.T) {
	unsupported := []Region{}

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
		RegionAU,
		RegionCA,
		RegionDE,
		RegionEur3,
		RegionEur6,
		RegionEurope,
		RegionGlobal,
		RegionIN,
		RegionIT,
		RegionJP,
		RegionNAM3,
		RegionNAM6,
		RegionNAM7,
		RegionNAM8,
		RegionNAM9,
		RegionNAM10,
		RegionNAM11,
		RegionNAM12,
		RegionNAMEurAsia1,
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

		// Marshal/unmarshal. Regions without a value for the enum fail to
		// marshal.
		buf, err := json.Marshal(region.ToRegionEnum())
		if want != region {
			if err == nil {
				t.Errorf("got %s want a marshal error for %q", buf, region)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
			var enum RegionEnum
			if err := json.Unmarshal(buf, &enum); err != nil {
				t.Fatal(err)
			}
			if r := enum.Region; r != want {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), want, want.DisplayName())
			}
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

func TestCloudAccountRegionEnum(t *testing.T) {
	unsupported := []Region{
		RegionAsia,
		RegionAsia1,
		RegionAU,
		RegionCA,
		RegionDE,
		RegionEU,
		RegionEur3,
		RegionEur4,
		RegionEur5,
		RegionEur6,
		RegionEur7,
		RegionEur8,
		RegionEurope,
		RegionGlobal,
		RegionIN,
		RegionIT,
		RegionJP,
		RegionNAM3,
		RegionNAM4,
		RegionNAM6,
		RegionNAM7,
		RegionNAM8,
		RegionNAM9,
		RegionNAM10,
		RegionNAM11,
		RegionNAM12,
		RegionNAMEurAsia1,
		RegionUS,
	}

	for region, info := range regionInfoMap {
		want := region
		if slices.Contains(unsupported, region) {
			want = RegionUnknown
		}

		// Lookup.
		if r := RegionFromCloudAccountRegionEnum(info.cloudAccountRegionEnum); r != want {
			t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), want, want.DisplayName())
		}

		// Marshal/unmarshal. Regions without a value for the enum fail to
		// marshal.
		buf, err := json.Marshal(region.ToCloudAccountRegionEnum())
		if want != region {
			if err == nil {
				t.Errorf("got %s want a marshal error for %q", buf, region)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
			var enum CloudAccountRegionEnum
			if err := json.Unmarshal(buf, &enum); err != nil {
				t.Fatal(err)
			}
			if r := enum.Region; r != want {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), want, want.DisplayName())
			}
		}

		// ToPtr.
		if r := region.ToCloudAccountRegionEnumPtr(); (r == nil && want != RegionUnknown) || (r != nil && *r != want.ToCloudAccountRegionEnum()) {
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
		RegionAsia,
		RegionAsia1,
		RegionAsiaSouthEast3,
		RegionAU,
		RegionCA,
		RegionDE,
		RegionEU,
		RegionEur3,
		RegionEur4,
		RegionEur5,
		RegionEur6,
		RegionEur7,
		RegionEur8,
		RegionEurope,
		RegionEuropeNorth2,
		RegionGlobal,
		RegionIN,
		RegionIT,
		RegionJP,
		RegionNAM3,
		RegionNAM4,
		RegionNAM6,
		RegionNAM7,
		RegionNAM8,
		RegionNAM9,
		RegionNAM10,
		RegionNAM11,
		RegionNAM12,
		RegionNAMEurAsia1,
		RegionUS,
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

		// Marshal/unmarshal. Regions without a value for the enum fail to
		// marshal.
		buf, err := json.Marshal(region.ToRCSRegionEnum())
		if want != region {
			if err == nil {
				t.Errorf("got %s want a marshal error for %q", buf, region)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
			var enum RCSRegionEnum
			if err := json.Unmarshal(buf, &enum); err != nil {
				t.Fatal(err)
			}
			if r := enum.Region; r != want {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), want, want.DisplayName())
			}
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

func TestKMSLocationEnum(t *testing.T) {
	unsupported := []Region{
		RegionEU,
		RegionUSEast7,
		RegionUSWest8,
	}

	for region, info := range regionInfoMap {
		want := region
		if slices.Contains(unsupported, region) {
			want = RegionUnknown
		}

		// Lookup.
		if r := RegionFromKMSLocationEnum(info.kmsLocationEnum); r != want {
			t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), want, want.DisplayName())
		}

		// Marshal/unmarshal.
		var enum KMSLocationEnum
		buf, err := json.Marshal(region.ToKMSLocationEnum())
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
		if r := region.ToKMSLocationEnumPtr(); (r == nil && want != RegionUnknown) || (r != nil && *r != want.ToKMSLocationEnum()) {
			got := "<nil>"
			if r != nil {
				got = fmt.Sprintf("%q", r.Name())
			}
			t.Errorf("got %s want %q", got, want)
		}

		// The region name is the Cloud KMS location name.
		if want != RegionUnknown {
			if name := strings.ToLower(strings.ReplaceAll(info.kmsLocationEnum, "_", "-")); name != info.name {
				t.Errorf("got %q want %q", info.name, name)
			}
		}
	}
}
