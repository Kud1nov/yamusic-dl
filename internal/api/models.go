// Package api contains data models and constants for working with the Yandex Music API.
package api

// Constants for API
const (
	// BaseURL base URL for Yandex Music API
	BaseURL = "https://api.music.yandex.net"

	// DefaultSignKey default key for request signatures
	DefaultSignKey = "p93jhgh689SBReK6ghtw62"

	// Codecs supported codecs
	Codecs = "flac,flac-mp4,mp3,aac,he-aac,aac-mp4,he-aac-mp4"

	// Transport data transfer format
	Transport = "encraw"
)

// TrackQuality defines the quality of the track in Yandex Music API
type TrackQuality string

const (
	// QualityLow - low quality
	QualityLow TrackQuality = "lq"

	// QualityNormal - standard quality
	QualityNormal TrackQuality = "nq"

	// QualityLossless - lossless quality
	QualityLossless TrackQuality = "lossless"
)

// DownloadQuality defines the quality of the track for downloading
type DownloadQuality string

const (
	// QualityMin - minimum quality
	QualityMin DownloadQuality = "min"

	// QualityStandard - standard quality
	QualityStandard DownloadQuality = "normal"

	// QualityHigh - maximum quality
	QualityHigh DownloadQuality = "max"
)

// Track represents track information
type Track struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Artists  []Artist `json:"artists"`
	Albums   []Album  `json:"albums"`
	Duration int      `json:"durationMs"`
}

// Artist represents artist information
type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Album represents album information
type Album struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
