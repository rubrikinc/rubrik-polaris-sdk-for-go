package graphql

import (
	"regexp"
	"strings"
)

// Version represents an RSC version, e.g. latest, master-54647 or v20230227-3.
type Version string

// Before returns true if the version is older than, lexicographically before,
// version tags of the same version format. Note that "latest" is never older
// than any version tag.
func (v Version) Before(versionTags ...string) bool {
	version := strings.TrimSpace(string(v))

	for _, pattern := range versionPatterns {
		if pattern.MatchString(version) {
			for _, versionTag := range versionTags {
				versionTag = strings.TrimSpace(versionTag)
				if pattern.MatchString(versionTag) {
					return version < versionTag
				}
			}
		}
	}

	return false
}

var versionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^master-\d+$`),
	regexp.MustCompile(`^v\d{8}(?:|-\d+)$`),
}

// Deprecated: use Version.Before.
func VersionOlderThan(version string, versionTags ...string) bool {
	return Version(version).Before(versionTags...)
}
