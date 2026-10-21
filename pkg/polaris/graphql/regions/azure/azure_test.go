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
	"strings"
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

func TestCommonRegionEnum(t *testing.T) {
	for region, info := range regionInfoMap {
		regionSupported := info.commonRegionEnum != ""

		// Lookup. The correct region should be returned or RegionUnknown for
		// unsupported regions.
		if r := RegionFromCommonRegionEnum(info.commonRegionEnum); regionSupported {
			if r != region {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), region, region.DisplayName())
			}
		} else {
			if r != RegionUnknown {
				t.Errorf("got %q [%s] want %q [%s]", r, r.DisplayName(), RegionUnknown, RegionUnknown.DisplayName())
			}
		}

		// Marshal/unmarshal. Unsupported regions should fail to marshal.
		buf, err := json.Marshal(region.ToCommonRegionEnum())
		if regionSupported {
			if err != nil {
				t.Fatal(err)
			}

			var enum CommonRegionEnum
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
		if r := region.ToCommonRegionEnumPtr(); regionSupported && region != RegionUnknown {
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

// TestCommonRegionEnumRoundTrip covers the AzureCommonRegion enum wiring,
// including the regions that were added for it. New Zealand North and Taiwan
// North exist only in AzureCommonRegion, not AzureCloudAccountRegion, so they
// must resolve via the common enum but not the cloud-account enum.
func TestCommonRegionEnumRoundTrip(t *testing.T) {
	regions := []Region{
		RegionAustriaEast,
		RegionBelgiumCentral,
		RegionChileCentral,
		RegionIndonesiaCentral,
		RegionMalaysiaWest,
		RegionNewZealandNorth,
		RegionTaiwanNorth,
		RegionEastUS,
	}

	for _, original := range regions {
		data, err := json.Marshal(original.ToCommonRegionEnum())
		if err != nil {
			t.Errorf("failed to marshal region %s: %s", original, err)
			continue
		}

		var got CommonRegionEnum
		if err := json.Unmarshal(data, &got); err != nil {
			t.Errorf("failed to unmarshal JSON %s: %s", string(data), err)
			continue
		}
		if got.Region != original {
			t.Errorf("round trip failed for region %s: got %s (data %s)", original, got.Region, string(data))
		}
	}

	if RegionFromCommonRegionEnum("NEWZEALANDNORTH") != RegionNewZealandNorth {
		t.Error("NEWZEALANDNORTH should resolve via the common region enum")
	}
	if RegionFromCloudAccountRegionEnum("NEWZEALANDNORTH") != RegionUnknown {
		t.Error("NEWZEALANDNORTH must not resolve via the cloud account region enum")
	}
}

func TestCommonRegionEnumMarshalError(t *testing.T) {
	// Verify that a region without a value for a specific GraphQL enum type
	// results in a marshal error.
	_, err := json.Marshal(RegionGermanyNortheast.ToCommonRegionEnum())
	if err == nil {
		t.Fatal("expected an error for a region without an AzureCommonRegion value")
	}
	if msg := err.Error(); !strings.Contains(msg, `"germanynortheast"`) || !strings.Contains(msg, "AzureCommonRegion") {
		t.Errorf("error %q should name the region and the enum", msg)
	}

	// The unknown region still marshals to the placeholder.
	buf, err := json.Marshal(RegionUnknown.ToCommonRegionEnum())
	if err != nil {
		t.Fatal(err)
	}
	if got := string(buf); got != `"UNKNOWN_AZURE_REGION"` {
		t.Errorf("got %s want %q", got, "UNKNOWN_AZURE_REGION")
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
		{RegionUnknown, `""`},
	}

	for _, test := range tests {
		data, err := json.Marshal(test.region)
		if err != nil {
			t.Errorf("failed to marshal region %s: %s", test.region, err)
			continue
		}
		if string(data) != test.expected {
			t.Errorf("marshal region %s: expected %s, got %s", test.region, test.expected, string(data))
		}
	}
}

func TestRegionUnmarshalJSON(t *testing.T) {
	tests := []struct {
		native   string
		expected Region
	}{
		{`"eastus"`, RegionEastUS},
		{`"westus"`, RegionWestUS},
		{`"northeurope"`, RegionNorthEurope},
		{`""`, RegionUnknown},
	}

	for _, test := range tests {
		var region Region
		err := json.Unmarshal([]byte(test.native), &region)
		if err != nil {
			t.Errorf("failed to unmarshal JSON %s: %s", test.native, err)
			continue
		}
		if region != test.expected {
			t.Errorf("unmarshal native region %s: expected %s, got %s", test.native, test.expected, region)
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
		data, err := json.Marshal(original)
		if err != nil {
			t.Errorf("failed to marshal region %v: %s", original, err)
			continue
		}

		var unmarshaled Region
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Errorf("failed to unmarshal JSON %s: %s", string(data), err)
			continue
		}
		if original != unmarshaled {
			t.Errorf("round trip failed for region %s: got %s", original, unmarshaled)
		}
	}
}
