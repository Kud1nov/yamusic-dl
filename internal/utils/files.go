// Package utils provides helper functions.
package utils

import (
	"regexp"
	"strings"
)

// CleanFileName cleans a string from invalid characters for a filename.
// Keeps letters (including Cyrillic), digits, and some special characters.
func CleanFileName(name string) string {
	// Replace characters that are invalid in filenames
	reg := regexp.MustCompile(`[\\/:*?"<>|]`)
	clean := reg.ReplaceAllString(name, "_")

	// Trim spaces and make sure we have something
	clean = strings.TrimSpace(clean)
	if clean == "" {
		return "Unknown"
	}

	return clean
}
