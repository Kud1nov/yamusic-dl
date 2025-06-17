// Package yamusic предоставляет клиент для работы с API Яндекс Музыки.
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
	"time"

	"github.com/google/uuid"

	"github.com/Kud1nov/yamusic-dl/internal/api"
	"github.com/Kud1nov/yamusic-dl/internal/crypto"
	"github.com/Kud1nov/yamusic-dl/internal/logger"
	"github.com/Kud1nov/yamusic-dl/internal/utils"
)

// Псевдонимы для типов из пакета api для обратной совместимости
type (
	// AudioQuality определяет качество трека для скачивания
	AudioQuality = api.DownloadQuality

	// ApiTrackQuality определяет качество трека в API Яндекс Музыки
	ApiTrackQuality = api.TrackQuality
)

// Константы качества для обратной совместимости
const (
	AudioQualityMin    = api.QualityMin
	AudioQualityNormal = api.QualityStandard
	AudioQualityMax    = api.QualityHigh

	ApiTrackQualityLow      = api.QualityLow
	ApiTrackQualityNormal   = api.QualityNormal
	ApiTrackQualityLossless = api.QualityLossless

	// DefaultSignKey ключ по умолчанию
	DefaultSignKey = api.DefaultSignKey
)

// Client предоставляет методы для работы с API Яндекс Музыки
type Client struct {
	accessToken string
	signKey     string
	headers     map[string]string
	logger      *logger.Logger
	httpClient  *http.Client
}

// NewClient создает новый клиент для работы с API Яндекс Музыки
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

// GetTrackInfo получает метаданные трека
func (c *Client) GetTrackInfo(trackID string) (map[string]interface{}, error) {
	c.logger.Debug("Получение метаданных трека %s", trackID)

	// Создаем multipart форму
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Добавляем поля в форму
	_ = writer.WriteField("trackIds", trackID)
	_ = writer.WriteField("removeDuplicates", "false")
	_ = writer.WriteField("withProgress", "true")
	writer.Close()

	// Создаем запрос
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/tracks", api.BaseURL), body)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем заголовки
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API вернул ошибку: %s", resp.Status)
	}

	// Разбираем ответ
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	// Проверяем наличие результата
	result, ok := response["result"].([]interface{})
	if !ok || len(result) == 0 {
		return nil, fmt.Errorf("неверный формат ответа: %v", response)
	}

	trackInfo, ok := result[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("неверный формат ответа: %v", response)
	}

	return trackInfo, nil
}

// GetDownloadInfo получает информацию для скачивания трека
func (c *Client) GetDownloadInfo(trackID string, quality ApiTrackQuality) (map[string]interface{}, error) {
	c.logger.Debug("Получение информации для скачивания трека %s", trackID)

	// Формируем параметры запроса
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	// Параметры запроса в виде мапы для удобства формирования URL
	params := map[string]string{
		"ts":         ts,
		"trackId":    trackID,
		"quality":    string(quality),
		"codecs":     api.Codecs,
		"transports": api.Transport,
	}

	// Генерируем подпись
	// Важно: собираем строку данных в правильном порядке
	dataString := ts + trackID + string(quality) + api.Codecs + api.Transport
	params["sign"] = crypto.GenerateSignature(dataString, c.signKey)

	// Логирование параметров и подписи
	c.logger.Debug("Параметры запроса: ts=%s, trackId=%s, quality=%s", ts, trackID, quality)
	c.logger.Debug("Сгенерирована подпись: %s", params["sign"])

	// Формируем URL с параметрами
	baseURL := fmt.Sprintf("%s/get-file-info", api.BaseURL)
	reqURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка формирования URL: %w", err)
	}

	// Добавляем параметры запроса
	query := reqURL.Query()
	for key, value := range params {
		query.Add(key, value)
	}
	reqURL.RawQuery = query.Encode()

	// Создаем запрос
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем заголовки
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API вернул ошибку: %s", resp.Status)
	}

	// Разбираем ответ
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	// Проверяем наличие результата и downloadInfo
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("неверный формат ответа: %v", response)
	}

	downloadInfo, ok := result["downloadInfo"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("неверный формат ответа: %v", response)
	}

	return downloadInfo, nil
}

// GetAPIQuality преобразует уровень качества для пользователя в параметр API
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

// DownloadTrack скачивает и расшифровывает трек
func (c *Client) DownloadTrack(trackID string, quality AudioQuality, outputDir string) (string, error) {
	// Получаем метаданные трека
	trackInfo, err := c.GetTrackInfo(trackID)
	if err != nil {
		c.logger.Error("Ошибка при получении метаданных трека %s: %v", trackID, err)
		return "", err
	}

	// Формируем имя файла из метаданных
	artist := "Unknown"
	title := "Unknown"

	// Извлекаем имя исполнителя
	if artists, ok := trackInfo["artists"].([]interface{}); ok && len(artists) > 0 {
		if artistMap, ok := artists[0].(map[string]interface{}); ok {
			if artistName, ok := artistMap["name"].(string); ok {
				artist = artistName
			}
		}
	}

	// Извлекаем название трека
	if trackTitle, ok := trackInfo["title"].(string); ok {
		title = trackTitle
	}

	// Очищаем имена от недопустимых символов
	safeArtist := utils.CleanFileName(artist)
	safeTitle := utils.CleanFileName(title)
	fileName := fmt.Sprintf("%s - %s [%s].m4a", safeArtist, safeTitle, trackID)

	c.logger.Info("Получена информация: %s", fileName)

	// Получаем информацию для скачивания с учетом выбранного качества
	apiQuality := c.GetAPIQuality(quality)
	downloadInfo, err := c.GetDownloadInfo(trackID, apiQuality)
	if err != nil {
		c.logger.Error("Ошибка при получении информации для скачивания трека %s: %v", trackID, err)
		return "", err
	}

	fileURL, ok := downloadInfo["url"].(string)
	if !ok {
		return "", fmt.Errorf("URL для скачивания не найден")
	}

	decryptionKey, ok := downloadInfo["key"].(string)
	if !ok {
		return "", fmt.Errorf("ключ для расшифровки не найден")
	}

	// Создаем временные файлы
	tempID := uuid.New().String()
	encryptedPath := fmt.Sprintf("encrypted_%s.raw", tempID)

	// Создаем отложенную функцию для очистки временных файлов
	// Она будет вызвана при любом выходе из функции - как нормальном, так и при ошибке или панике
	defer func() {
		if _, err := os.Stat(encryptedPath); err == nil {
			c.logger.Debug("Удаление временного файла: %s", encryptedPath)
			os.Remove(encryptedPath)
		}
	}()

	// Проверяем и создаем директорию для сохранения, если нужно
	if outputDir == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("ошибка получения текущей директории: %w", err)
		}
		outputDir = currentDir
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("ошибка создания директории: %w", err)
	}

	outputPath := filepath.Join(outputDir, fileName)

	// Скачиваем зашифрованный файл
	c.logger.Info("Скачивание трека...")
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("ошибка при скачивании файла: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка при скачивании файла, статус: %s", resp.Status)
	}

	// Сохраняем зашифрованный файл
	encryptedFile, err := os.Create(encryptedPath)
	if err != nil {
		return "", fmt.Errorf("ошибка создания временного файла: %w", err)
	}

	_, err = io.Copy(encryptedFile, resp.Body)
	encryptedFile.Close()
	if err != nil {
		return "", fmt.Errorf("ошибка при сохранении зашифрованного файла: %w", err)
	}

	// Расшифровываем файл
	c.logger.Debug("Ключ дешифровки: %s", decryptionKey)

	// Читаем зашифрованные данные
	encryptedData, err := os.ReadFile(encryptedPath)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения зашифрованного файла: %w", err)
	}

	// Расшифровываем данные
	decrypted, err := crypto.DecryptAesCtr(encryptedData, decryptionKey)
	if err != nil {
		return "", fmt.Errorf("ошибка расшифровки файла: %w", err)
	}

	c.logger.Info("Сохранение файла...")

	// Сохраняем расшифрованный файл
	err = os.WriteFile(outputPath, decrypted, 0644)
	if err != nil {
		return "", fmt.Errorf("ошибка при сохранении расшифрованного файла: %w", err)
	}

	c.logger.Info("Готово: %s", outputPath)
	return outputPath, nil
}
