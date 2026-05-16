package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"shortener/backend/internal/config"
	"shortener/backend/internal/httpapi"
	"shortener/backend/internal/migrations"
	"shortener/backend/internal/service"
	"shortener/backend/internal/storage/memory"
	"shortener/backend/internal/storage/postgres"
)

type App struct {
	handler http.Handler
	close   func()
}

func New(cfg config.Config, logger *slog.Logger) *App {
	var store service.Store
	closeStore := func() {}

	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		postgresStore, err := postgres.New(ctx, cfg.DatabaseURL)
		if err != nil {
			logger.Error("postgres unavailable, falling back to in-memory storage", "error", err)
			store = memory.New()
		} else {
			if err := migrations.Up(ctx, postgresStore.Pool(), cfg.MigrationsDir); err != nil {
				logger.Error("migrations failed, falling back to in-memory storage", "error", err)
				postgresStore.Close()
				store = memory.New()
			} else {
				logger.Info("postgres storage enabled")
				store = postgresStore
				closeStore = postgresStore.Close
			}
		}
	} else {
		logger.Info("in-memory storage enabled")
		store = memory.New()
	}

	shortener := service.NewShortener(store, service.Options{
		PublicBaseURL: cfg.PublicBaseURL,
		CodeLength:    cfg.CodeLength,
	})

	return &App{
		handler: httpapi.NewRouter(shortener, logger, cfg.FrontendDir),
		close:   closeStore,
	}
}

func (a *App) Handler() http.Handler {
	return a.handler
}

func (a *App) Close() {
	a.close()
}
