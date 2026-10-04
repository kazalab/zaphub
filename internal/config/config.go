package config

import (
	"os"
	"time"
)

type Config struct {
	Port                  string
	LogLevel              string
	EPGURL                string
	EPGRefreshInterval    time.Duration
	YTRefreshInterval     time.Duration
	YTChannelsFile        string
	CinemaURL             string
	CinemaProvider        string
	CinemaRefreshInterval time.Duration
	DataDir               string
}

func Load() Config {
	return Config{
		Port:                  env("PORT", "8080"),
		LogLevel:              env("LOG_LEVEL", "info"),
		EPGURL:                env("EPG_URL", "https://xmltvfr.fr/xmltv/xmltv_tnt.xml.gz"),
		EPGRefreshInterval:    envDuration("EPG_REFRESH_INTERVAL", 6*time.Hour),
		YTRefreshInterval:     envDuration("YT_REFRESH_INTERVAL", 15*time.Minute),
		YTChannelsFile:        env("YT_CHANNELS_FILE", "/data/youtube_channels.json"),
		CinemaURL:             env("CINEMA_URL", "https://www.allocine.fr/film/sorties-semaine/"),
		CinemaProvider:        env("CINEMA_PROVIDER", "allocine"),
		CinemaRefreshInterval: envDuration("CINEMA_REFRESH_INTERVAL", 24*time.Hour),
		DataDir:               env("DATA_DIR", "/data"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
