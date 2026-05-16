package app

import (
	"log/slog"
	"net/http"

	"shortener/backend/internal/config"
	"shortener/backend/internal/httpapi"
	"shortener/backend/internal/service"
	"shortener/backend/internal/storage/memory"
)

type App struct {
	handler http.Handler
}

func New(cfg config.Config, logger *slog.Logger) *App {
	store := memory.New()
	shortener := service.NewShortener(store, service.Options{
		PublicBaseURL: cfg.PublicBaseURL,
		CodeLength:    cfg.CodeLength,
	})

	return &App{
		handler: httpapi.NewRouter(shortener, logger),
	}
}

func (a *App) Handler() http.Handler {
	return a.handler
}
