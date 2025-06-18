// Package utils provides helper functions.
package utils

import (
	"regexp"
	"strings"
)

// ExtractTrackID extracts track ID from different formats:
// - Full URL: https://music.yandex.ru/album/10376938/track/64551568
// - URL with params: https://music.yandex.ru/album/10376938/track/64551568?utm_source=desktop
// - Just track ID: 64551568
func ExtractTrackID(input string) string {
	// If input is already just a track ID (only digits)
	if matched, _ := regexp.MatchString(`^\d+$`, input); matched {
		return input
	}

	// Check if it's a Yandex Music URL
	if strings.Contains(input, "music.yandex") {
		// Extract track ID from URL
		re := regexp.MustCompile(`/track/(\d+)`)
		matches := re.FindStringSubmatch(input)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// Return original input if no pattern matched
	// (this will likely fail later, but we're being lenient)
	return input
}
