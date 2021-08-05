package core

import "regexp"

var versionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^master-\d+$`),
	regexp.MustCompile(`^v\d{8}(?:|-\d+)$`),
}

// VersionOlderThan returns true if the specified version is older than
// (lexicographically before) version tags of the same version format. Note
// that "latest" is never older than any version tag.
func VersionOlderThan(version string, versionTags ...string) bool {
	for _, pattern := range versionPatterns {
		if pattern.MatchString(version) {
			for _, versionTag := range versionTags {
				if pattern.MatchString(versionTag) {
					return version < versionTag
				}
			}
		}
	}

	return false
}
