package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port               int
	TMDBAPIKey         string
	RutrackerUsername   string
	RutrackerPassword  string
	RutrackerMirror    string
	OpenSubtitlesKey   string
	DataDir            string
	TorrentDir         string
	DBPath             string
	MaxCacheGB         int
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:             getEnvInt("PORT", 8080),
		TMDBAPIKey:       os.Getenv("TMDB_API_KEY"),
		RutrackerUsername: os.Getenv("RUTRACKER_USERNAME"),
		RutrackerPassword: os.Getenv("RUTRACKER_PASSWORD"),
		RutrackerMirror:  getEnv("RUTRACKER_MIRROR", "rutracker.org"),
		OpenSubtitlesKey: os.Getenv("OPENSUBTITLES_API_KEY"),
		DataDir:          getEnv("DATA_DIR", "./data"),
		MaxCacheGB:       getEnvInt("MAX_CACHE_GB", 50),
	}

	cfg.TorrentDir = cfg.DataDir + "/torrents"
	cfg.DBPath = cfg.DataDir + "/streambox.db"

	if cfg.TMDBAPIKey == "" {
		return nil, fmt.Errorf("TMDB_API_KEY is required")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return defaultVal
}
