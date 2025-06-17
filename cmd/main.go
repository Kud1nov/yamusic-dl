// Yamusic-dl - консольная утилита для скачивания музыки из Яндекс Музыки.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Kud1nov/yamusic-dl/internal/logger"
	"github.com/Kud1nov/yamusic-dl/pkg/yamusic"
)

func main() {
	// Определяем параметры командной строки
	trackID := flag.String("track", "", "Идентификатор трека")
	accessToken := flag.String("token", "", "Токен доступа к API Яндекс Музыки")
	qualityStr := flag.String("quality", string(yamusic.AudioQualityMax),
		"Качество трека (min, normal, max)")
	outputDir := flag.String("output", "", "Директория для сохранения файлов")
	verbose := flag.Bool("verbose", false, "Вывод отладочных сообщений")

	// Парсим параметры
	flag.Parse()

	// Проверяем обязательные параметры
	if *trackID == "" || *accessToken == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Проверяем качество
	quality := yamusic.AudioQuality(*qualityStr)
	if quality != yamusic.AudioQualityMin &&
		quality != yamusic.AudioQualityNormal &&
		quality != yamusic.AudioQualityMax {
		fmt.Println("Ошибка: некорректное качество. Допустимые значения: min, normal, max")
		os.Exit(1)
	}

	// Настраиваем логгер
	log := logger.New(*verbose)

	// Создаем директорию для сохранения, если нужно
	if *outputDir != "" {
		if err := os.MkdirAll(*outputDir, 0755); err != nil {
			log.Error("Ошибка создания директории: %v", err)
			os.Exit(1)
		}
	}

	// Создаем клиент Яндекс Музыки
	client := yamusic.NewClient(*accessToken, yamusic.DefaultSignKey, log)

	// Скачиваем трек
	_, err := client.DownloadTrack(*trackID, quality, *outputDir)
	if err != nil {
		log.Error("Ошибка: %v", err)
		os.Exit(1)
	}
}
