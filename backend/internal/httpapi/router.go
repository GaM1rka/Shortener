package httpapi

import (
	"log/slog"
	"net/http"

	"shortener/backend/internal/service"
)

func NewRouter(shortener *service.Shortener, logger *slog.Logger) http.Handler {
	api := &API{
		shortener: shortener,
		logger:    logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", api.health)
	mux.HandleFunc("POST /shorten", api.createShortLink)
	mux.HandleFunc("GET /s/{short_url}", api.redirect)
	mux.HandleFunc("GET /analytics/{short_url}", api.analytics)

	return api.withMiddleware(mux)
}
