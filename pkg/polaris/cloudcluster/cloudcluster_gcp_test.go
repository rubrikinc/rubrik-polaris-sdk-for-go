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

package cloudcluster

import (
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/cloudcluster"
)

func boolPtr(b bool) *bool { return &b }

func TestValidateGcpBucketRegion(t *testing.T) {
	tests := []struct {
		name          string
		bucketRegion  string
		clusterRegion string
		expectErr     bool
	}{
		{
			name:          "ExactMatch",
			bucketRegion:  "us-east1",
			clusterRegion: "us-east1",
			expectErr:     false,
		},
		{
			name:          "CaseInsensitiveMatch",
			bucketRegion:  "US-EAST1",
			clusterRegion: "us-east1",
			expectErr:     false,
		},
		{
			name:          "RegionMismatch",
			bucketRegion:  "us-central1",
			clusterRegion: "us-east1",
			expectErr:     true,
		},
		{
			name:          "MultiRegionDoesNotMatchRegional",
			bucketRegion:  "US",
			clusterRegion: "us-east1",
			expectErr:     true,
		},
		{
			name:          "EmptyBucketRegionAllowed",
			bucketRegion:  "",
			clusterRegion: "us-east1",
			expectErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGcpBucketRegion(tt.bucketRegion, tt.clusterRegion)
			if tt.expectErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGcpRegionZones(t *testing.T) {
	regions := []cloudcluster.GcpRegionInfo{
		{Name: "us-west1", Zones: []string{"us-west1-a", "us-west1-b", "us-west1-c"}},
		{Name: "us-east4", Zones: []string{"us-east4-a", "us-east4-b"}},
	}

	t.Run("FoundCaseInsensitive", func(t *testing.T) {
		zones, err := gcpRegionZones(regions, "US-WEST1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(zones) != 3 {
			t.Errorf("expected 3 zones, got %d", len(zones))
		}
	})

	t.Run("NotAvailable", func(t *testing.T) {
		if _, err := gcpRegionZones(regions, "europe-west1"); err == nil {
			t.Error("expected error for unavailable region but got none")
		}
	})
}

func TestExpandGcpNetworkConfig(t *testing.T) {
	subnet := cloudcluster.GcpSubnetInput{Name: "subnet-a", Network: "vpc", HostProject: "p", Region: "us-west1"}

	mk := func(numNodes int, azResilient bool, nc []cloudcluster.GcpSubnetInput) cloudcluster.CreateGcpClusterInput {
		in := cloudcluster.CreateGcpClusterInput{
			ClusterConfig: cloudcluster.GcpClusterConfig{NumNodes: numNodes},
			VMConfig:      cloudcluster.GcpVmConfig{NetworkConfig: nc},
		}
		if azResilient {
			in.IsAzResilient = boolPtr(true)
		}
		return in
	}

	t.Run("SingleSubnetMultiNodeFansOut", func(t *testing.T) {
		got := expandGcpNetworkConfig(mk(3, false, []cloudcluster.GcpSubnetInput{subnet}))
		if len(got) != 3 {
			t.Fatalf("expected 3 entries, got %d", len(got))
		}
		for i, s := range got {
			if s != subnet {
				t.Errorf("entry %d = %+v, want %+v", i, s, subnet)
			}
		}
	})

	t.Run("SingleNodeUnchanged", func(t *testing.T) {
		if got := expandGcpNetworkConfig(mk(1, false, []cloudcluster.GcpSubnetInput{subnet})); len(got) != 1 {
			t.Errorf("expected 1 entry, got %d", len(got))
		}
	})

	t.Run("AlreadyPerNodeUnchanged", func(t *testing.T) {
		nc := []cloudcluster.GcpSubnetInput{subnet, subnet, subnet}
		if got := expandGcpNetworkConfig(mk(3, false, nc)); len(got) != 3 {
			t.Errorf("expected 3 entries, got %d", len(got))
		}
	})

	t.Run("AzResilientUnchanged", func(t *testing.T) {
		if got := expandGcpNetworkConfig(mk(3, true, []cloudcluster.GcpSubnetInput{subnet})); len(got) != 1 {
			t.Errorf("expected AZ-resilient to be left unchanged at 1 entry, got %d", len(got))
		}
	})
}

func TestValidateGcpZones(t *testing.T) {
	threeZones := []string{"us-west1-a", "us-west1-b", "us-west1-c"}
	twoZones := []string{"us-east4-a", "us-east4-b"}

	tests := []struct {
		name        string
		input       cloudcluster.CreateGcpClusterInput
		regionZones []string
		expectErr   bool
	}{
		{
			name: "ZoneInRegion",
			input: cloudcluster.CreateGcpClusterInput{
				Region: "us-west1",
				Zone:   "us-west1-a",
			},
			regionZones: threeZones,
			expectErr:   false,
		},
		{
			name: "ZoneInRegionCaseInsensitive",
			input: cloudcluster.CreateGcpClusterInput{
				Region: "us-west1",
				Zone:   "US-WEST1-A",
			},
			regionZones: threeZones,
			expectErr:   false,
		},
		{
			name: "ZoneNotInRegion",
			input: cloudcluster.CreateGcpClusterInput{
				Region: "us-west1",
				Zone:   "us-east4-a",
			},
			regionZones: threeZones,
			expectErr:   true,
		},
		{
			name: "AzResilientValid",
			input: cloudcluster.CreateGcpClusterInput{
				Region:        "us-west1",
				Zone:          "us-west1-a",
				IsAzResilient: boolPtr(true),
				ClusterConfig: cloudcluster.GcpClusterConfig{NumNodes: 3},
				VMConfig: cloudcluster.GcpVmConfig{
					SubnetAzConfigs: []cloudcluster.SubnetAzConfig{
						{AvailabilityZone: "us-west1-a", Subnet: "subnet-a"},
						{AvailabilityZone: "us-west1-b", Subnet: "subnet-b"},
						{AvailabilityZone: "us-west1-c", Subnet: "subnet-c"},
					},
				},
			},
			regionZones: threeZones,
			expectErr:   false,
		},
		{
			name: "AzResilientTooFewRegionZones",
			input: cloudcluster.CreateGcpClusterInput{
				Region:        "us-east4",
				Zone:          "us-east4-a",
				IsAzResilient: boolPtr(true),
				ClusterConfig: cloudcluster.GcpClusterConfig{NumNodes: 3},
			},
			regionZones: twoZones,
			expectErr:   true,
		},
		{
			name: "AzResilientTooFewNodes",
			input: cloudcluster.CreateGcpClusterInput{
				Region:        "us-west1",
				Zone:          "us-west1-a",
				IsAzResilient: boolPtr(true),
				ClusterConfig: cloudcluster.GcpClusterConfig{NumNodes: 1},
			},
			regionZones: threeZones,
			expectErr:   true,
		},
		{
			name: "AzResilientTooFewSubnets",
			input: cloudcluster.CreateGcpClusterInput{
				Region:        "us-west1",
				Zone:          "us-west1-a",
				IsAzResilient: boolPtr(true),
				ClusterConfig: cloudcluster.GcpClusterConfig{NumNodes: 3},
				VMConfig: cloudcluster.GcpVmConfig{
					SubnetAzConfigs: []cloudcluster.SubnetAzConfig{
						{AvailabilityZone: "us-west1-a", Subnet: "subnet-a"},
						{AvailabilityZone: "us-west1-b", Subnet: "subnet-b"},
					},
				},
			},
			regionZones: threeZones,
			expectErr:   true,
		},
		{
			name: "AzResilientSubnetZoneNotInRegion",
			input: cloudcluster.CreateGcpClusterInput{
				Region:        "us-west1",
				Zone:          "us-west1-a",
				IsAzResilient: boolPtr(true),
				ClusterConfig: cloudcluster.GcpClusterConfig{NumNodes: 3},
				VMConfig: cloudcluster.GcpVmConfig{
					SubnetAzConfigs: []cloudcluster.SubnetAzConfig{
						{AvailabilityZone: "us-west1-a", Subnet: "subnet-a"},
						{AvailabilityZone: "us-west1-b", Subnet: "subnet-b"},
						{AvailabilityZone: "us-east4-a", Subnet: "subnet-c"},
					},
				},
			},
			regionZones: threeZones,
			expectErr:   true,
		},
		{
			name: "NonResilientIgnoresNodeAndZoneCount",
			input: cloudcluster.CreateGcpClusterInput{
				Region:        "us-east4",
				Zone:          "us-east4-a",
				IsAzResilient: boolPtr(false),
				ClusterConfig: cloudcluster.GcpClusterConfig{NumNodes: 1},
			},
			regionZones: twoZones,
			expectErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGcpZones(tt.input, tt.regionZones)
			if tt.expectErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
