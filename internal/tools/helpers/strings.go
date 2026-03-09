package helpers

import (
	"strings"
)

// TrimStringFields normalizes string pointers in place so handlers can trim once at the boundary.
func TrimStringFields(values ...*string) {
	for _, value := range values {
		if value == nil {
			continue
		}
		*value = strings.TrimSpace(*value)
	}
}

// TrimStrings normalizes each string element and drops entries that become empty after trimming.
func TrimStrings(values []string) []string {
	if values == nil {
		return nil
	}

	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}

	return normalized
}
