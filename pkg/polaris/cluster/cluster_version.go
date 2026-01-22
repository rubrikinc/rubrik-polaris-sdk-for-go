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

// CDMVersion represents a CDM version string (e.g., "9.4.0-p2-30507").
// Only the major.minor.patch portion is used for comparisons.
type CDMVersion string

// ParseCDMVersion creates a CDMVersion from a version string.
func ParseCDMVersion(version string) CDMVersion {
	return CDMVersion(version)
}

// parseComponents extracts the major, minor, and patch version numbers from
// a CDM version string. It handles formats like "9.4.0", "9.4", "9.4.0-p2-30507".
func (v CDMVersion) parseComponents() (major, minor, patch int) {
	version := string(v)

	// Remove any suffix after the first hyphen (e.g., "-p2-30507")
	if idx := strings.Index(version, "-"); idx != -1 {
		version = version[:idx]
	}

	parts := strings.Split(version, ".")
	if len(parts) >= 1 {
		major, _ = strconv.Atoi(parts[0])
	}
	if len(parts) >= 2 {
		minor, _ = strconv.Atoi(parts[1])
	}
	if len(parts) >= 3 {
		patch, _ = strconv.Atoi(parts[2])
	}

	return major, minor, patch
}

// Compare compares the CDM version with another version string.
// Returns:
//
//	-1 if v < other
//	 0 if v == other
//	 1 if v > other
//
// Only the major.minor.patch portion is compared; suffixes are ignored.
func (v CDMVersion) Compare(other string) int {
	vMajor, vMinor, vPatch := v.parseComponents()
	oMajor, oMinor, oPatch := CDMVersion(other).parseComponents()

	if vMajor != oMajor {
		if vMajor < oMajor {
			return -1
		}
		return 1
	}
	if vMinor != oMinor {
		if vMinor < oMinor {
			return -1
		}
		return 1
	}
	if vPatch != oPatch {
		if vPatch < oPatch {
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
