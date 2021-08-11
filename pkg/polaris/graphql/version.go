package graphql

import (
	"regexp"
	"strings"
)

var queryPattern *regexp.Regexp = regexp.MustCompile(`^(?:mutation|query) +SdkGolang(.+?) *?(?:\(|{)`)

// QueryName returns the name of the specified GraphQL query.
func QueryName(query string) string {
	groups := queryPattern.FindStringSubmatch(query)
	if len(groups) != 2 {
		return "<invalid-query>"
	}

	return strings.ToLower(groups[1][:1]) + groups[1][1:]
}

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
