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

import (
	"strconv"
	"strings"
)

// CDMVersion represents a parsed CDM version (e.g., "9.4.0-p2-30507").
// Comparisons use the major.minor.patch portion and, as a final tiebreaker,
// the trailing build number, so releases sharing the same major.minor.patch
// (e.g. "9.5.0-p1-35914" and "9.5.0-p2-36035") order by build.
type CDMVersion struct {
	major int
	minor int
	patch int
	build int
}

// ParseCDMVersion parses a version string and creates a CDMVersion.
// It handles formats like "9.4.0", "9.4", "9.4.0-30189", "9.4.0-p2-30507".
func ParseCDMVersion(version string) (CDMVersion, error) {
	// Split the major.minor.patch core from any "-pN-build" suffix.
	core, _, _ := strings.Cut(version, "-")

	var major, minor, patch int
	var err error
	parts := strings.Split(core, ".")
	if len(parts) >= 1 {
		major, err = strconv.Atoi(parts[0])
		if err != nil {
			return CDMVersion{}, err
		}
	}
	if len(parts) >= 2 {
		minor, err = strconv.Atoi(parts[1])
		if err != nil {
			return CDMVersion{}, err
		}
	}
	if len(parts) >= 3 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return CDMVersion{}, err
		}
	}

	// The build number is the final hyphen-separated segment, when numeric
	// (e.g. "9.4.0-p2-30507" -> 30507, "9.4.0-30189" -> 30189). A version with
	// no build, such as "9.4.0" or a bare patch level "9.3.3-p9", leaves it at 0.
	var build int
	if i := strings.LastIndex(version, "-"); i != -1 {
		build, _ = strconv.Atoi(version[i+1:])
	}

	return CDMVersion{major: major, minor: minor, patch: patch, build: build}, nil
}

// Compare compares the CDM version with another version string.
// Returns:
//
//	-1 if v < other
//	 0 if v == other
//	 1 if v > other
//
// The major.minor.patch portion is compared first; the trailing build number
// breaks ties between versions sharing the same major.minor.patch. Other
// suffix tokens (such as the "pN" patch level) are not compared.
func (v CDMVersion) Compare(other string) int {
	o, err := ParseCDMVersion(other)
	if err != nil {
		return -1
	}

	if v.major != o.major {
		if v.major < o.major {
			return -1
		}
		return 1
	}
	if v.minor != o.minor {
		if v.minor < o.minor {
			return -1
		}
		return 1
	}
	if v.patch != o.patch {
		if v.patch < o.patch {
			return -1
		}
		return 1
	}
	if v.build != o.build {
		if v.build < o.build {
			return -1
		}
		return 1
	}

	return 0
}

// AtLeast returns true if the CDM version is greater than or equal to the
// specified minimum version.
func (v CDMVersion) AtLeast(minVersion string) bool {
	return v.Compare(minVersion) >= 0
}

// LessThan returns true if the CDM version is less than the specified version.
func (v CDMVersion) LessThan(version string) bool {
	return v.Compare(version) < 0
}

// GreaterThan returns true if the CDM version is greater than the specified
// version.
func (v CDMVersion) GreaterThan(version string) bool {
	return v.Compare(version) > 0
}
