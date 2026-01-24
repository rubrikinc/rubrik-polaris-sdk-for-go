package graphql

import (
	"errors"
	"regexp"
	"strings"
)

// ErrNoMatch is returned when none of the provided versions matches the
// version format.
var ErrNoMatch = errors.New("none of the provided versions matches format")

// Version represents an RSC version, e.g. latest, master-54647 or v20230227-3.
type Version string

// Before returns true if the version is older than, lexicographically before,
// version tags of the same version format. Returns ErrNoMatch iff none of the
// provided versions matches the version format. Note that "latest" always
// returns false with no error.
func (v Version) Before(versionTags ...string) (bool, error) {
	version := strings.TrimSpace(string(v))

	// Special handling for "latest" - it's never before any version.
	if version == "latest" {
		return false, nil
	}

	for _, pattern := range versionPatterns {
		if !pattern.MatchString(version) {
			continue
		}
		for _, versionTag := range versionTags {
			versionTag = strings.TrimSpace(versionTag)
			if pattern.MatchString(versionTag) {
				return version < versionTag, nil
			}
		}
		// Version matches a pattern but no tag matches it.
		return false, ErrNoMatch
	}

	// Version doesn't match any known pattern.
	return false, ErrNoMatch
}

var versionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^master-\d+$`),
	regexp.MustCompile(`^v\d{8}(?:|-\d+)$`),
}

// Deprecated: use Version.Before.
func VersionOlderThan(version string, versionTags ...string) bool {
	before, _ := Version(version).Before(versionTags...)
	return before
}
