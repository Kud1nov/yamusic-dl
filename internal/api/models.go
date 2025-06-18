// Package api contains data models and constants for working with the Yandex Music API.
package api

import "encoding/json"

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
	ID   json.Number `json:"id"` // Changed from string to json.Number for flexibility
	Name string      `json:"name"`
}

// Album represents album information
type Album struct {
	ID    json.Number `json:"id"` // Changed from string to json.Number
	Title string      `json:"title"`
}

// TrackResponse represents the API response for track information
type TrackResponse struct {
	InvocationInfo InvocationInfo `json:"invocationInfo"`
	Result         []TrackInfo    `json:"result"`
}

// TrackInfo represents detailed information about a track
type TrackInfo struct {
	ID                    string            `json:"id"`
	RealID                string            `json:"realId"`
	Title                 string            `json:"title"`
	Available             bool              `json:"available"`
	DurationMs            int               `json:"durationMs"`
	PreviewDurationMs     int               `json:"previewDurationMs"`
	Artists               []Artist          `json:"artists"`
	Albums                []Album           `json:"albums"`
	CoverUri              string            `json:"coverUri"`
	OgImage               string            `json:"ogImage"`
	LyricsAvailable       bool              `json:"lyricsAvailable"`
	Type                  string            `json:"type"`
	RememberPosition      bool              `json:"rememberPosition"`
	TrackSharingFlag      string            `json:"trackSharingFlag"`
	TrackSource           string            `json:"trackSource"`
	DerivedColors         map[string]string `json:"derivedColors"`
	SpecialAudioResources []string          `json:"specialAudioResources"`
}

// DownloadInfoResponse represents the API response for download information
type DownloadInfoResponse struct {
	Result         DownloadInfoResult `json:"result"`
	InvocationInfo InvocationInfo     `json:"invocationInfo"`
}

// DownloadInfoResult wraps the download information
type DownloadInfoResult struct {
	DownloadInfo DownloadInfo `json:"downloadInfo"`
}

// DownloadInfo contains details needed to download a track
type DownloadInfo struct {
	TrackID   string   `json:"trackId"`
	Quality   string   `json:"quality"`
	Codec     string   `json:"codec"`
	Bitrate   int      `json:"bitrate"`
	Transport string   `json:"transport"`
	Key       string   `json:"key"`
	Size      int      `json:"size"`
	Gain      bool     `json:"gain"`
	Urls      []string `json:"urls"`
	Url       string   `json:"url"`
	RealID    string   `json:"realId"`
}

// InvocationInfo contains metadata about the API request
type InvocationInfo struct {
	ReqID              string `json:"req-id"`
	Hostname           string `json:"hostname"`
	ExecDurationMillis int    `json:"exec-duration-millis"`
	AppName            string `json:"app-name,omitempty"`
}
