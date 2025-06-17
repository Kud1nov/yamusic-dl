package crypto

import (
	"net/url"
	"testing"
)

func TestGenerateSignature(t *testing.T) {
	// Тестовые случаи
	tests := []struct {
		name       string
		dataString string
		signKey    string
		expected   string
	}{
		{
			name:       "Тестовый случай из примера URL",
			dataString: "1750200603138562777losslessflac,flac-mp4,mp3,aac,he-aac,aac-mp4,he-aac-mp4encraw",
			signKey:    "p93jhgh689SBReK6ghtw62",
			expected:   "dIS2WrOe5DP9DOAgm6OGu68yb4hfD0DoQj/mu4vVaGc",
		},
		{
			name:       "Тестовый случай с запятыми",
			dataString: "1750200603138562777losslessflac,flac-mp4,mp3,aac,he-aac,aac-mp4,he-aac-mp4encraw",
			signKey:    "p93jhgh689SBReK6ghtw62",
			expected:   "dIS2WrOe5DP9DOAgm6OGu68yb4hfD0DoQj/mu4vVaGc",
		},
	}

	// Запуск тестов
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSignature(tt.dataString, tt.signKey)

			// Проверяем соответствие ожидаемому результату
			if result != tt.expected {
				t.Errorf("GenerateSignature() = %v, want %v", result, tt.expected)
			}

			// Проверяем URL-безопасность полученной подписи
			_, err := url.Parse("https://api.music.yandex.net/get-file-info?sign=" + result)
			if err != nil {
				t.Errorf("Generated signature is not URL-safe: %v", err)
			}
		})
	}
}

func TestGenerateSignatureFromParams(t *testing.T) {
	// Тестовые данные
	ts := "1750200603"
	trackId := "138562777"
	quality := "lossless"
	codecs := "flac,flac-mp4,mp3,aac,he-aac,aac-mp4,he-aac-mp4"
	transports := "encraw"
	signKey := "p93jhgh689SBReK6ghtw62"
	expected := "dIS2WrOe5DP9DOAgm6OGu68yb4hfD0DoQj/mu4vVaGc"

	// Генерируем подпись
	result := GenerateSignatureFromParams(ts, trackId, quality, codecs, transports, signKey)

	// Проверяем результат
	if result != expected {
		t.Errorf("GenerateSignatureFromParams() = %v, want %v", result, expected)
	}
}

// TestParseAndGenerateFromURL проверяет генерацию подписи из URL
func TestParseAndGenerateFromURL(t *testing.T) {
	// Тестовый URL
	testURL := "https://api.music.yandex.net/get-file-info?ts=1750200603&trackId=138562777&quality=lossless&codecs=flac%2Cflac-mp4%2Cmp3%2Caac%2Che-aac%2Caac-mp4%2Che-aac-mp4&transports=encraw&sign=dIS2WrOe5DP9DOAgm6OGu68yb4hfD0DoQj%2Fmu4vVaGc"

	// Парсим URL
	parsedURL, err := url.Parse(testURL)
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}

	// Извлекаем параметры
	query := parsedURL.Query()

	// Получаем параметры в нужном порядке
	ts := query.Get("ts")
	trackId := query.Get("trackId")
	quality := query.Get("quality")
	codecs := query.Get("codecs")
	transports := query.Get("transports")

	// Получаем ожидаемую подпись
	expectedSign := query.Get("sign")

	// Генерируем подпись
	signKey := "p93jhgh689SBReK6ghtw62"
	generatedSign := GenerateSignatureFromParams(ts, trackId, quality, codecs, transports, signKey)

	// Проверяем соответствие
	if generatedSign != expectedSign {
		t.Errorf("Generated signature %s doesn't match expected %s", generatedSign, expectedSign)
	}
}
