// Package yamusic provides a client for working with the Yandex Music API.
package yamusic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Kud1nov/yamusic-dl/internal/api"
	"github.com/Kud1nov/yamusic-dl/internal/crypto"
	"github.com/Kud1nov/yamusic-dl/internal/logger"
	"github.com/Kud1nov/yamusic-dl/internal/utils"
)

// Type aliases from the api package for backward compatibility
type (
	// AudioQuality defines the quality of the track for downloading
	AudioQuality = api.DownloadQuality

	// ApiTrackQuality defines the quality of the track in the Yandex Music API
	ApiTrackQuality = api.TrackQuality
)

// Quality constants for backward compatibility
const (
	AudioQualityMin    = api.QualityMin
	AudioQualityNormal = api.QualityStandard
	AudioQualityMax    = api.QualityHigh

	ApiTrackQualityLow      = api.QualityLow
	ApiTrackQualityNormal   = api.QualityNormal
	ApiTrackQualityLossless = api.QualityLossless

	// DefaultSignKey default key
	DefaultSignKey = api.DefaultSignKey
)

// Client provides methods for working with the Yandex Music API
type Client struct {
	accessToken string
	signKey     string
	headers     map[string]string
	logger      *logger.Logger
	httpClient  *http.Client
}

// NewClient creates a new client for working with the Yandex Music API
func NewClient(accessToken, signKey string, log *logger.Logger) *Client {
	if signKey == "" {
		signKey = api.DefaultSignKey
	}

	if log == nil {
		log = logger.New(false)
	}

	client := &Client{
		accessToken: accessToken,
		signKey:     signKey,
		logger:      log,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}

	client.headers = map[string]string{
		"Accept-Language":       "ru",
		"Authorization":         fmt.Sprintf("OAuth %s", accessToken),
		"x-yandex-music-client": "YandexMusicAndroid/24023621",
		"User-Agent":            "Ya-Music-DL/1.0",
	}

	return client
}

// GetTrackInfo retrieves track metadata
func (c *Client) GetTrackInfo(trackID string) (map[string]interface{}, error) {
	c.logger.Debug("Getting track metadata %s", trackID)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add fields to the form
	_ = writer.WriteField("trackIds", trackID)
	_ = writer.WriteField("removeDuplicates", "false")
	_ = writer.WriteField("withProgress", "true")
	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/tracks", api.BaseURL), body)
	if err != nil {
		return nil, fmt.Errorf("request creation error: %w", err)
	}

	// Set headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request execution error: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		c.logger.Debug("API error response: %s", string(responseBody))
		return nil, fmt.Errorf("API returned an error: %s", resp.Status)
	}

	// Read response body for debugging and parsing
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Log raw response for debugging
	c.logger.Debug("Raw API response: %s", string(responseData))

	// First try to parse as generic map to inspect structure
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(responseData, &rawResponse); err != nil {
		return nil, fmt.Errorf("error parsing raw response: %w", err)
	}

	// Check raw response structure
	result, ok := rawResponse["result"].([]interface{})
	if !ok || len(result) == 0 {
		return nil, fmt.Errorf("invalid result format in raw response")
	}

	// Log raw track info for debugging
	trackMap, ok := result[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid track format in raw response")
	}

	c.logger.Debug("Raw track title: %v", trackMap["title"])
	c.logger.Debug("Raw artists: %v", trackMap["artists"])

	// Parse response using our defined structure
	var trackResponse api.TrackResponse
	if err := json.Unmarshal(responseData, &trackResponse); err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	// Validate response
	if len(trackResponse.Result) == 0 {
		return nil, fmt.Errorf("no track information found in API response")
	}

	// Log structured data for debugging
	c.logger.Debug("Structured track title: %s", trackResponse.Result[0].Title)
	if len(trackResponse.Result[0].Artists) > 0 {
		c.logger.Debug("Structured artist name: %s", trackResponse.Result[0].Artists[0].Name)
	} else {
		c.logger.Debug("No artists found in structured response")
	}

	// Create trackInfo from raw response instead of structured
	trackInfo := make(map[string]interface{})

	// Basic track info
	trackInfo["id"] = trackMap["id"]
	trackInfo["title"] = trackMap["title"]
	trackInfo["durationMs"] = trackMap["durationMs"]

	// Add artists from raw response
	if rawArtists, ok := trackMap["artists"].([]interface{}); ok && len(rawArtists) > 0 {
		artists := make([]map[string]interface{}, len(rawArtists))
		for i, rawArtist := range rawArtists {
			if artistMap, ok := rawArtist.(map[string]interface{}); ok {
				artist := make(map[string]interface{})
				artist["id"] = artistMap["id"]
				artist["name"] = artistMap["name"]
				artists[i] = artist
				c.logger.Debug("Adding artist: %v", artist["name"])
			}
		}
		trackInfo["artists"] = artists
	}

	// Add albums from raw response
	if rawAlbums, ok := trackMap["albums"].([]interface{}); ok && len(rawAlbums) > 0 {
		albums := make([]map[string]interface{}, len(rawAlbums))
		for i, rawAlbum := range rawAlbums {
			if albumMap, ok := rawAlbum.(map[string]interface{}); ok {
				album := make(map[string]interface{})
				album["id"] = albumMap["id"]
				album["title"] = albumMap["title"]
				albums[i] = album
			}
		}
		trackInfo["albums"] = albums
	}

	// Verify the trackInfo map has the expected values
	c.logger.Debug("Track title in trackInfo: %v", trackInfo["title"])
	if artists, ok := trackInfo["artists"].([]map[string]interface{}); ok && len(artists) > 0 {
		c.logger.Debug("First artist name in trackInfo: %v", artists[0]["name"])
	}

	return trackInfo, nil
}

// GetDownloadInfo retrieves information for downloading a track
func (c *Client) GetDownloadInfo(trackID string, quality ApiTrackQuality) (map[string]interface{}, error) {
	c.logger.Debug("Getting download info for track %s", trackID)

	// Form request parameters
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	// Request parameters as a map for URL formation convenience
	params := map[string]string{
		"ts":         ts,
		"trackId":    trackID,
		"quality":    string(quality),
		"codecs":     api.Codecs,
		"transports": api.Transport,
	}

	// Generate signature
	// Important: assemble the data string in the correct order
	dataString := ts + trackID + string(quality) + api.Codecs + api.Transport
	params["sign"] = crypto.GenerateSignature(dataString, c.signKey)

	// Log parameters and signature
	c.logger.Debug("Request parameters: ts=%s, trackId=%s, quality=%s", ts, trackID, quality)
	c.logger.Debug("Generated signature: %s", params["sign"])

	// Form URL with parameters
	baseURL := fmt.Sprintf("%s/get-file-info", api.BaseURL)
	reqURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("URL formation error: %w", err)
	}

	// Add request parameters
	query := reqURL.Query()
	for key, value := range params {
		query.Add(key, value)
	}
	reqURL.RawQuery = query.Encode()

	// Create request
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("request creation error: %w", err)
	}

	// Set headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request execution error: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned an error: %s", resp.Status)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	// Check for result and downloadInfo presence
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: %v", response)
	}

	downloadInfo, ok := result["downloadInfo"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: %v", response)
	}

	return downloadInfo, nil
}

// GetAPIQuality converts user quality level to API parameter
func (c *Client) GetAPIQuality(quality AudioQuality) ApiTrackQuality {
	switch quality {
	case AudioQualityMin:
		return ApiTrackQualityLow
	case AudioQualityNormal:
		return ApiTrackQualityNormal
	case AudioQualityMax:
		return ApiTrackQualityLossless
	default:
		return ApiTrackQualityLossless
	}
}

// DownloadTrack downloads and decrypts a track
func (c *Client) DownloadTrack(trackID string, quality AudioQuality, outputDir string) (string, error) {
	// Get track metadata
	trackInfo, err := c.GetTrackInfo(trackID)
	if err != nil {
		c.logger.Error("Error getting track metadata %s: %v", trackID, err)
		return "", err
	}

	// Form filename from metadata
	artist := "Unknown"
	title := "Unknown"
	albumsStr := "Unknown"
	artists := []string{}
	albums := []string{}

	// Log the trackInfo structure for debugging
	c.logger.Debug("TrackInfo structure: %+v", trackInfo)

	// Extract track title
	if trackTitle, ok := trackInfo["title"].(string); ok {
		title = trackTitle
		c.logger.Debug("Found track title: %s", title)
	}

	// Extract artists - collect all artists
	if artistsList, ok := trackInfo["artists"].([]map[string]interface{}); ok && len(artistsList) > 0 {
		// Format returned by our improved GetTrackInfo
		for _, artistMap := range artistsList {
			if artistName, ok := artistMap["name"].(string); ok {
				artists = append(artists, artistName)
				c.logger.Debug("Found artist: %s", artistName)
			}
		}
	} else if artistsList, ok := trackInfo["artists"].([]interface{}); ok && len(artistsList) > 0 {
		// Original format
		for _, a := range artistsList {
			if artistMap, ok := a.(map[string]interface{}); ok {
				if artistName, ok := artistMap["name"].(string); ok {
					artists = append(artists, artistName)
					c.logger.Debug("Found artist: %s", artistName)
				}
			}
		}
	}

	// Join artists with &
	if len(artists) > 0 {
		artist = strings.Join(artists, " & ")
		c.logger.Debug("Combined artists: %s", artist)
	}

	// Extract album titles
	if albumsList, ok := trackInfo["albums"].([]map[string]interface{}); ok && len(albumsList) > 0 {
		for _, albumMap := range albumsList {
			if albumTitle, ok := albumMap["title"].(string); ok {
				albums = append(albums, albumTitle)
				c.logger.Debug("Found album: %s", albumTitle)
			}
		}
	} else if albumsList, ok := trackInfo["albums"].([]interface{}); ok && len(albumsList) > 0 {
		for _, a := range albumsList {
			if albumMap, ok := a.(map[string]interface{}); ok {
				if albumTitle, ok := albumMap["title"].(string); ok {
					albums = append(albums, albumTitle)
					c.logger.Debug("Found album: %s", albumTitle)
				}
			}
		}
	}

	// Join albums with comma

	if len(albums) > 0 {
		albumsStr = strings.Join(albums, ", ")
		c.logger.Debug("Combined albums: %s", albumsStr)
	}

	// If still no title, artist or albums, try to get them directly from the API again
	if title == "Unknown" || artist == "Unknown" || (len(albums) == 0 && albumsStr == "Unknown") {
		c.logger.Debug("Missing title, artist or album, trying direct API access")

		// This is a fallback method to get track info if the structured approach failed
		apiTitle, apiArtist, apiAlbum := c.getTrackInfoFallback(trackID)
		if title == "Unknown" && apiTitle != "" {
			title = apiTitle
			c.logger.Debug("Using fallback title: %s", title)
		}
		if artist == "Unknown" && apiArtist != "" {
			artist = apiArtist
			c.logger.Debug("Using fallback artist: %s", artist)
		}
		if albumsStr == "Unknown" && apiAlbum != "" {
			albumsStr = apiAlbum
			c.logger.Debug("Using fallback album: %s", albumsStr)
		}
	}

	// Clean names from invalid characters
	safeTitle := utils.CleanFileName(title)
	safeArtist := utils.CleanFileName(artist)
	safeAlbums := utils.CleanFileName(albumsStr)

	// Format: Track Title - Artist1 & Artist2 (Album1, Album2) [ID трека]
	fileName := fmt.Sprintf("%s - %s (%s) [%s].m4a", safeTitle, safeArtist, safeAlbums, trackID)

	c.logger.Info("Got information: %s", fileName)

	// Get download information considering the selected quality
	apiQuality := c.GetAPIQuality(quality)
	downloadInfo, err := c.GetDownloadInfo(trackID, apiQuality)
	if err != nil {
		c.logger.Error("Error getting download information for track %s: %v", trackID, err)
		return "", err
	}

	fileURL, ok := downloadInfo["url"].(string)
	if !ok {
		return "", fmt.Errorf("download URL not found")
	}

	decryptionKey, ok := downloadInfo["key"].(string)
	if !ok {
		return "", fmt.Errorf("decryption key not found")
	}

	// Create temporary files
	tempID := uuid.New().String()
	encryptedPath := fmt.Sprintf("encrypted_%s.raw", tempID)

	// Create a deferred function for cleaning up temporary files
	// It will be called at any exit from the function - both normal and on error or panic
	defer func() {
		if _, err := os.Stat(encryptedPath); err == nil {
			c.logger.Debug("Deleting temporary file: %s", encryptedPath)
			os.Remove(encryptedPath)
		}
	}()

	// Check and create directory for saving, if needed
	if outputDir == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("error getting current directory: %w", err)
		}
		outputDir = currentDir
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("error creating directory: %w", err)
	}

	outputPath := filepath.Join(outputDir, fileName)

	// Download encrypted file
	c.logger.Info("Downloading track...")
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("error downloading file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error downloading file, status: %s", resp.Status)
	}

	// Save encrypted file
	encryptedFile, err := os.Create(encryptedPath)
	if err != nil {
		return "", fmt.Errorf("error creating temporary file: %w", err)
	}

	_, err = io.Copy(encryptedFile, resp.Body)
	encryptedFile.Close()
	if err != nil {
		return "", fmt.Errorf("error saving encrypted file: %w", err)
	}

	// Decrypt file
	c.logger.Debug("Decryption key: %s", decryptionKey)

	// Read encrypted data
	encryptedData, err := os.ReadFile(encryptedPath)
	if err != nil {
		return "", fmt.Errorf("error reading encrypted file: %w", err)
	}

	// Decrypt data
	decrypted, err := crypto.DecryptAesCtr(encryptedData, decryptionKey)
	if err != nil {
		return "", fmt.Errorf("error decrypting file: %w", err)
	}

	c.logger.Info("Saving file...")

	// Save decrypted file
	err = os.WriteFile(outputPath, decrypted, 0644)
	if err != nil {
		return "", fmt.Errorf("error saving decrypted file: %w", err)
	}

	c.logger.Info("Done: %s", outputPath)
	return outputPath, nil
}

// getTrackInfoFallback is a simpler method to extract basic track info when the main method fails
func (c *Client) getTrackInfoFallback(trackID string) (title, artist, album string) {
	// Create a simple GET request instead of POST with multipart form
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/tracks/%s", api.BaseURL, trackID), nil)
	if err != nil {
		c.logger.Debug("Fallback request creation error: %v", err)
		return "", "", ""
	}

	// Set headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Debug("Fallback request execution error: %v", err)
		return "", "", ""
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		c.logger.Debug("Fallback API returned error status: %s", resp.Status)
		return "", "", ""
	}

	// Parse response as generic map
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.logger.Debug("Fallback response parsing error: %v", err)
		return "", "", ""
	}

	// Try to extract track info
	var extractedTitle, extractedArtist, extractedAlbum string

	// Extract from result
	if result, ok := response["result"].(map[string]interface{}); ok {
		// Extract title
		if trackTitle, ok := result["title"].(string); ok {
			extractedTitle = trackTitle
			c.logger.Debug("Fallback found title: %s", extractedTitle)
		}

		// Extract artist
		if artists, ok := result["artists"].([]interface{}); ok && len(artists) > 0 {
			if artistMap, ok := artists[0].(map[string]interface{}); ok {
				if artistName, ok := artistMap["name"].(string); ok {
					extractedArtist = artistName
					c.logger.Debug("Fallback found artist: %s", extractedArtist)
				}
			}
		}

		// Extract album
		if albums, ok := result["albums"].([]interface{}); ok && len(albums) > 0 {
			if albumMap, ok := albums[0].(map[string]interface{}); ok {
				if albumTitle, ok := albumMap["title"].(string); ok {
					extractedAlbum = albumTitle
					c.logger.Debug("Fallback found album: %s", extractedAlbum)
				}
			}
		}
	}

	return extractedTitle, extractedArtist, extractedAlbum
}
