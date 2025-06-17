// Package utils provides helper functions.
package utils

import (
	"regexp"
	"strings"
)

// CleanFileName cleans a string from invalid characters for a filename.
// Only keeps letters, digits, and some special characters.
func CleanFileName(name string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9 _\-]`)
	clean := reg.ReplaceAllString(name, "")
	return strings.TrimSpace(clean)
}
