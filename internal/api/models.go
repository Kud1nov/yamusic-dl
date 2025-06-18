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
	ID          json.Number `json:"id"`
	Name        string      `json:"name"`
	Various     bool        `json:"various,omitempty"`
	Composer    bool        `json:"composer,omitempty"`
	Available   bool        `json:"available,omitempty"`
	Cover       Cover       `json:"cover,omitempty"`
	Genres      []string    `json:"genres"`
	Disclaimers []string    `json:"disclaimers"`
}

// Album represents album information
type Album struct {
	ID                       json.Number   `json:"id"`
	Title                    string        `json:"title"`
	MetaType                 string        `json:"metaType,omitempty"`
	ContentWarning           string        `json:"contentWarning,omitempty"`
	Year                     int           `json:"year,omitempty"`
	ReleaseDate              string        `json:"releaseDate,omitempty"`
	CoverUri                 string        `json:"coverUri,omitempty"`
	OgImage                  string        `json:"ogImage,omitempty"`
	Genre                    string        `json:"genre,omitempty"`
	TrackCount               int           `json:"trackCount,omitempty"`
	LikesCount               int           `json:"likesCount,omitempty"`
	Recent                   bool          `json:"recent,omitempty"`
	VeryImportant            bool          `json:"veryImportant,omitempty"`
	Artists                  []Artist      `json:"artists,omitempty"`
	Labels                   []Label       `json:"labels,omitempty"`
	Available                bool          `json:"available,omitempty"`
	AvailableForPremiumUsers bool          `json:"availableForPremiumUsers,omitempty"`
	AvailableForOptions      []string      `json:"availableForOptions,omitempty"`
	AvailableForMobile       bool          `json:"availableForMobile,omitempty"`
	AvailablePartially       bool          `json:"availablePartially,omitempty"`
	Bests                    []json.Number `json:"bests,omitempty"`
	Disclaimers              []string      `json:"disclaimers,omitempty"`
	ListeningFinished        bool          `json:"listeningFinished,omitempty"`
	TrackPosition            TrackPosition `json:"trackPosition,omitempty"`
}

// TrackResponse represents the API response for track information
type TrackResponse struct {
	InvocationInfo InvocationInfo `json:"invocationInfo"`
	Result         []TrackInfo    `json:"result"`
}

// TrackInfo represents detailed information about a track
type TrackInfo struct {
	ID                       string        `json:"id"`
	RealID                   string        `json:"realId"`
	Title                    string        `json:"title"`
	Major                    Major         `json:"major,omitempty"`
	Available                bool          `json:"available"`
	AvailableForPremiumUsers bool          `json:"availableForPremiumUsers"`
	AvailableForOptions      []string      `json:"availableForOptions"`
	Disclaimers              []string      `json:"disclaimers"`
	StorageDir               string        `json:"storageDir"`
	DurationMs               int           `json:"durationMs"`
	FileSize                 int           `json:"fileSize"`
	R128                     R128          `json:"r128,omitempty"`
	Fade                     Fade          `json:"fade,omitempty"`
	PreviewDurationMs        int           `json:"previewDurationMs"`
	Artists                  []Artist      `json:"artists"`
	Albums                   []Album       `json:"albums"`
	CoverUri                 string        `json:"coverUri"`
	OgImage                  string        `json:"ogImage"`
	LyricsAvailable          bool          `json:"lyricsAvailable"`
	LyricsInfo               LyricsInfo    `json:"lyricsInfo,omitempty"`
	Type                     string        `json:"type"`
	RememberPosition         bool          `json:"rememberPosition"`
	TrackSharingFlag         string        `json:"trackSharingFlag"`
	TrackSource              string        `json:"trackSource"`
	DerivedColors            DerivedColors `json:"derivedColors"`
	SpecialAudioResources    []string      `json:"specialAudioResources"`
}

// Major represents label information
type Major struct {
	ID   json.Number `json:"id"`
	Name string      `json:"name"`
}

// Label represents a music label
type Label struct {
	ID   json.Number `json:"id"`
	Name string      `json:"name"`
}

// Cover represents artist or album cover
type Cover struct {
	Type   string `json:"type"`
	Uri    string `json:"uri"`
	Prefix string `json:"prefix,omitempty"`
}

// R128 represents loudness information
type R128 struct {
	I  float64 `json:"i"`
	Tp float64 `json:"tp"`
}

// Fade represents track fade in/out information
type Fade struct {
	InStart  float64 `json:"inStart"`
	InStop   float64 `json:"inStop"`
	OutStart float64 `json:"outStart"`
	OutStop  float64 `json:"outStop"`
}

// TrackPosition represents the track's position in an album
type TrackPosition struct {
	Volume int `json:"volume"`
	Index  int `json:"index"`
}

// DerivedColors represents colors extracted from album art
type DerivedColors struct {
	Average    string `json:"average"`
	WaveText   string `json:"waveText"`
	MiniPlayer string `json:"miniPlayer"`
	Accent     string `json:"accent"`
}

// LyricsInfo represents information about lyrics availability
type LyricsInfo struct {
	HasAvailableSyncLyrics bool `json:"hasAvailableSyncLyrics"`
	HasAvailableTextLyrics bool `json:"hasAvailableTextLyrics"`
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
