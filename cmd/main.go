// Yamusic-dl - a console utility for downloading music from Yandex Music.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Kud1nov/yamusic-dl/internal/logger"
	"github.com/Kud1nov/yamusic-dl/internal/utils"
	"github.com/Kud1nov/yamusic-dl/pkg/yamusic"
)

func main() {
	// Define command line parameters
	trackInput := flag.String("track", "", "Track ID or Yandex Music URL")
	accessToken := flag.String("token", "", "Access token for Yandex Music API")
	qualityStr := flag.String("quality", string(yamusic.AudioQualityMax),
		"Track quality (min, normal, max)")
	outputDir := flag.String("output", "", "Directory for saving files")
	verbose := flag.Bool("verbose", false, "Output debug messages")

	// Parse parameters
	flag.Parse()

	// Check required parameters
	if *trackInput == "" || *accessToken == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Extract track ID from input (URL or ID)
	trackID := utils.ExtractTrackID(*trackInput)

	// Check quality
	quality := yamusic.AudioQuality(*qualityStr)
	if quality != yamusic.AudioQualityMin &&
		quality != yamusic.AudioQualityNormal &&
		quality != yamusic.AudioQualityMax {
		fmt.Println("Error: invalid quality. Valid values: min, normal, max")
		os.Exit(1)
	}

	// Configure logger
	log := logger.New(*verbose)

	// Create directory for saving if needed
	if *outputDir != "" {
		if err := os.MkdirAll(*outputDir, 0755); err != nil {
			log.Error("Error creating directory: %v", err)
			os.Exit(1)
		}
	}

	// Create Yandex Music client
	client := yamusic.NewClient(*accessToken, yamusic.DefaultSignKey, log)

	// Download track
	_, err := client.DownloadTrack(trackID, quality, *outputDir)
	if err != nil {
		log.Error("Error: %v", err)
		os.Exit(1)
	}
}
