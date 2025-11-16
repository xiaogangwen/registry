package model

import "strings"

var SupportedSchemaVersions = []string{
	CurrentSchemaVersion,
	LegacySchemaVersion20250929,
}

// IsSupportedSchemaVersion returns true if the provided schema URL references one of the supported schema versions.
func IsSupportedSchemaVersion(schemaURL string) bool {
	if schemaURL == "" {
		return false
	}

	for _, version := range SupportedSchemaVersions {
		if strings.Contains(schemaURL, version) {
			return true
		}
	}

	return false
}
