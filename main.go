package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	apiKeyEnv        = "YOUTUBE_API_KEY"
	serverUrlEnv     = "SERVER_URL"
	portEnv          = "PORT"
	filterPatternEnv = "FILTER_PATTERN"
	convertToMp3Env  = "CONVERT_TO_MP3"

	cacheTTL = 3600 // Default cache expiration time (in seconds)
	audioDir = "./audio_cache"
)

func main() {
	_ = godotenv.Load()

	port := getEnvAsInt(portEnv, 8080)
	cacheTTLSeconds := getEnvAsInt("CACHE_TTL", cacheTTL)
	cacheExpiry = time.Duration(cacheTTLSeconds) * time.Second

	if _, err := os.Stat(audioDir); os.IsNotExist(err) {
		if err := os.Mkdir(audioDir, os.ModePerm); err != nil {
			log.Fatalf("Error creating audio cache directory: %v", err)
		} else {
			log.Printf("Created audio cache directory: %s", audioDir)
		}
	}

	serverUrl := os.Getenv(serverUrlEnv)
	if serverUrl == "" {
		serverUrl = fmt.Sprintf("http://localhost:%d", port)
	}

	apiKey := os.Getenv(apiKeyEnv)
	if apiKey == "" {
		log.Fatalf("YouTube API key not set")
		return
	}

	server := &Server{
		apiUrl:        serverUrl,
		youtubeApiKey: apiKey,
		filterPattern: os.Getenv(filterPatternEnv),
		convertToMp3:  os.Getenv(convertToMp3Env) == "true",
	}
	http.Handle("/", server)

	serverPort := fmt.Sprintf(":%d", port)
	log.Printf("Server started at 0.0.0.0:%d", port)

	log.Fatal(http.ListenAndServe(serverPort, nil))
}

// Helper function to get environment variable or default
func getEnvAsInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultVal
}
