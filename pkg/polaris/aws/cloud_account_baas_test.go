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

package aws

import (
	"slices"
	"testing"

	awsregions "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/regions/aws"
)

// TestManagedAccountSupportedRegionsResolve guards against the BaaS region
// allow-list referencing a region that is not present in the SDK region table
// (which would silently marshal to an unknown region). Every entry must
// resolve and round-trip through its name.
func TestManagedAccountSupportedRegionsResolve(t *testing.T) {
	regions := ManagedAccountSupportedRegions()
	if len(regions) == 0 {
		t.Fatal("expected a non-empty BaaS supported region set")
	}

	seen := make(map[awsregions.Region]struct{}, len(regions))
	for _, region := range regions {
		if region == awsregions.RegionUnknown {
			t.Fatal("BaaS supported region set contains an unresolved region")
		}
		if _, ok := seen[region]; ok {
			t.Errorf("BaaS supported region set contains a duplicate: %s", region.Name())
		}
		seen[region] = struct{}{}

		if got := awsregions.RegionFromName(region.Name()); got != region {
			t.Errorf("region %q does not round-trip via name, got %q", region.Name(), got.Name())
		}
	}

	// The two regions added specifically for the BaaS flow must be present.
	for _, region := range []awsregions.Region{awsregions.RegionApSouthEast7, awsregions.RegionMxCentral1} {
		if _, ok := seen[region]; !ok {
			t.Errorf("expected BaaS supported regions to include %q", region.Name())
		}
	}
}

// TestManagedAccountDefaultFeatureNames pins the default BaaS feature set:
// EC2, RDS, S3 and Cloud Discovery.
func TestManagedAccountDefaultFeatureNames(t *testing.T) {
	got := ManagedAccountDefaultFeatureNames()
	want := []string{
		"CLOUD_NATIVE_PROTECTION",
		"RDS_PROTECTION",
		"CLOUD_NATIVE_S3_PROTECTION",
		"CLOUD_DISCOVERY",
	}
	if !slices.Equal(got, want) {
		t.Errorf("unexpected default BaaS features\n got: %v\nwant: %v", got, want)
	}
}
