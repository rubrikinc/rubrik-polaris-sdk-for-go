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

package cluster

import "testing"

func TestCDMVersionCompare(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		other    string
		expected int
	}{
		{"equal versions", "9.4.0", "9.4.0", 0},
		{"equal with suffix", "9.4.0-p2-30507", "9.4.0", 0},
		{"less than major", "8.4.0", "9.4.0", -1},
		{"greater than major", "10.4.0", "9.4.0", 1},
		{"less than minor", "9.3.0", "9.4.0", -1},
		{"greater than minor", "9.5.0", "9.4.0", 1},
		{"less than patch", "9.4.0", "9.4.1", -1},
		{"greater than patch", "9.4.2", "9.4.1", 1},
		{"compare with suffix", "9.4.0-p2-30507", "9.5", -1},
		{"partial version comparison", "9.5.0-p1-12345", "9.5", 0},
		{"major only comparison", "10.0.0", "9", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := ParseCDMVersion(tt.version)
			if err != nil {
				t.Fatalf("ParseCDMVersion(%q) failed: %v", tt.version, err)
			}
			result := v.Compare(tt.other)
			if result != tt.expected {
				t.Errorf("ParseCDMVersion(%q).Compare(%q) = %d, expected %d", tt.version, tt.other, result, tt.expected)
			}
		})
	}
}

func TestCDMVersionAtLeast(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		minVersion string
		expected   bool
	}{
		{"version equals minimum", "9.5.0", "9.5", true},
		{"version above minimum", "9.6.0-p2-30507", "9.5", true},
		{"version below minimum", "9.4.0-p2-30507", "9.5", false},
		{"major version above", "10.0.0", "9.5", true},
		{"major version below", "8.9.9", "9.5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := ParseCDMVersion(tt.version)
			if err != nil {
				t.Fatalf("ParseCDMVersion(%q) failed: %v", tt.version, err)
			}
			result := v.AtLeast(tt.minVersion)
			if result != tt.expected {
				t.Errorf("ParseCDMVersion(%q).AtLeast(%q) = %v, expected %v", tt.version, tt.minVersion, result, tt.expected)
			}
		})
	}
}

func TestCDMVersionLessThan(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		other    string
		expected bool
	}{
		{"version equals", "9.5.0", "9.5", false},
		{"version above", "9.6.0", "9.5", false},
		{"version below", "9.4.0-p2-30507", "9.5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := ParseCDMVersion(tt.version)
			if err != nil {
				t.Fatalf("ParseCDMVersion(%q) failed: %v", tt.version, err)
			}
			result := v.LessThan(tt.other)
			if result != tt.expected {
				t.Errorf("ParseCDMVersion(%q).LessThan(%q) = %v, expected %v", tt.version, tt.other, result, tt.expected)
			}
		})
	}
}

func TestCDMVersionGreaterThan(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		other    string
		expected bool
	}{
		{"version equals", "9.5.0", "9.5", false},
		{"version above", "9.6.0", "9.5", true},
		{"version below", "9.4.0-p2-30507", "9.5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := ParseCDMVersion(tt.version)
			if err != nil {
				t.Fatalf("ParseCDMVersion(%q) failed: %v", tt.version, err)
			}
			result := v.GreaterThan(tt.other)
			if result != tt.expected {
				t.Errorf("ParseCDMVersion(%q).GreaterThan(%q) = %v, expected %v", tt.version, tt.other, result, tt.expected)
			}
		})
	}
}
