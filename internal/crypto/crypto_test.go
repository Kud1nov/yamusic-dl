package crypto

import (
	"net/url"
	"testing"
)

func TestGenerateSignature(t *testing.T) {
	// Test cases
	tests := []struct {
		name       string
		dataString string
		signKey    string
		expected   string
	}{
		{
			name:       "Test case from URL example",
			dataString: "1750200603138562777losslessflac,flac-mp4,mp3,aac,he-aac,aac-mp4,he-aac-mp4encraw",
			signKey:    "p93jhgh689SBReK6ghtw62",
			expected:   "dIS2WrOe5DP9DOAgm6OGu68yb4hfD0DoQj/mu4vVaGc",
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSignature(tt.dataString, tt.signKey)

			// Check the match with the expected result
			if result != tt.expected {
				t.Errorf("GenerateSignature() = %v, want %v", result, tt.expected)
			}

			// Check URL safety of the generated signature
			_, err := url.Parse("https://api.music.yandex.net/get-file-info?sign=" + result)
			if err != nil {
				t.Errorf("Generated signature is not URL-safe: %v", err)
			}
		})
	}
}

func TestGenerateSignatureFromParams(t *testing.T) {
	// Test data
	ts := "1750200603"
	trackId := "138562777"
	quality := "lossless"
	codecs := "flac,flac-mp4,mp3,aac,he-aac,aac-mp4,he-aac-mp4"
	transports := "encraw"
	signKey := "p93jhgh689SBReK6ghtw62"
	expected := "dIS2WrOe5DP9DOAgm6OGu68yb4hfD0DoQj/mu4vVaGc"

	// Generate signature
	result := GenerateSignatureFromParams(ts, trackId, quality, codecs, transports, signKey)

	// Check the result
	if result != expected {
		t.Errorf("GenerateSignatureFromParams() = %v, want %v", result, expected)
	}
}

// TestParseAndGenerateFromURL checks signature generation from URL
func TestParseAndGenerateFromURL(t *testing.T) {
	// Test URL
	testURL := "https://api.music.yandex.net/get-file-info?ts=1750200603&trackId=138562777&quality=lossless&codecs=flac%2Cflac-mp4%2Cmp3%2Caac%2Che-aac%2Caac-mp4%2Che-aac-mp4&transports=encraw&sign=dIS2WrOe5DP9DOAgm6OGu68yb4hfD0DoQj%2Fmu4vVaGc"

	// Parse URL
	parsedURL, err := url.Parse(testURL)
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}

	// Extract parameters
	query := parsedURL.Query()

	// Get parameters in the correct order
	ts := query.Get("ts")
	trackId := query.Get("trackId")
	quality := query.Get("quality")
	codecs := query.Get("codecs")
	transports := query.Get("transports")

	// Get the expected signature
	expectedSign := query.Get("sign")

	// Generate signature
	signKey := "p93jhgh689SBReK6ghtw62"
	generatedSign := GenerateSignatureFromParams(ts, trackId, quality, codecs, transports, signKey)

	// Check the match
	if generatedSign != expectedSign {
		t.Errorf("Generated signature %s doesn't match expected %s", generatedSign, expectedSign)
	}
}
