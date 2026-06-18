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

import "testing"

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
