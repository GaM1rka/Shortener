package httpapi

import (
	"log/slog"
	"net/http"
	"os"

	"shortener/backend/internal/service"
)

func NewRouter(shortener *service.Shortener, logger *slog.Logger, frontendDir string) http.Handler {
	api := &API{
		shortener: shortener,
		logger:    logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", api.health)
	mux.HandleFunc("POST /shorten", api.createShortLink)
	mux.HandleFunc("GET /s/{short_url}", api.redirect)
	mux.HandleFunc("GET /analytics/{short_url}", api.analytics)

	if frontendDir != "" {
		if _, err := os.Stat(frontendDir); err == nil {
			mux.Handle("GET /", http.FileServer(http.Dir(frontendDir)))
			logger.Info("frontend enabled", "dir", frontendDir)
		} else {
			logger.Warn("frontend disabled", "dir", frontendDir, "error", err)
		}
	}

	return api.withMiddleware(mux)
}
