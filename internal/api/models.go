// Package api содержит модели данных и константы для работы с API Яндекс Музыки.
package api

// Constants for API
const (
	// BaseURL базовый URL для API Яндекс Музыки
	BaseURL = "https://api.music.yandex.net"

	// DefaultSignKey ключ для подписи запросов по умолчанию
	DefaultSignKey = "p93jhgh689SBReK6ghtw62"

	// Codecs поддерживаемые кодеки
	Codecs = "flac,flac-mp4,mp3,aac,he-aac,aac-mp4,he-aac-mp4"

	// Transport формат передачи данных
	Transport = "encraw"
)

// TrackQuality определяет качество трека в API Яндекс Музыки
type TrackQuality string

const (
	// QualityLow - низкое качество
	QualityLow TrackQuality = "lq"

	// QualityNormal - стандартное качество
	QualityNormal TrackQuality = "nq"

	// QualityLossless - без потерь
	QualityLossless TrackQuality = "lossless"
)

// DownloadQuality определяет качество трека для скачивания
type DownloadQuality string

const (
	// QualityMin - минимальное качество
	QualityMin DownloadQuality = "min"

	// QualityStandard - стандартное качество
	QualityStandard DownloadQuality = "normal"

	// QualityHigh - максимальное качество
	QualityHigh DownloadQuality = "max"
)

// Track представляет информацию о треке
type Track struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Artists  []Artist `json:"artists"`
	Albums   []Album  `json:"albums"`
	Duration int      `json:"durationMs"`
}

// Artist представляет информацию об исполнителе
type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Album представляет информацию об альбоме
type Album struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
