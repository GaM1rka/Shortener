package config

import (
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr      string
	PublicBaseURL string
	CodeLength    int
	LogLevel      slog.Level
}

func Load() Config {
	return Config{
		HTTPAddr:      env("HTTP_ADDR", ":8080"),
		PublicBaseURL: env("PUBLIC_BASE_URL", "http://localhost:8080"),
		CodeLength:    envInt("SHORT_CODE_LENGTH", 7),
		LogLevel:      envLogLevel("LOG_LEVEL", slog.LevelInfo),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envLogLevel(key string, fallback slog.Level) slog.Level {
	switch os.Getenv(key) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return fallback
	}
}
