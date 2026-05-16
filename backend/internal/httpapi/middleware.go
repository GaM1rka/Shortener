package httpapi

import (
	"net/http"
	"time"
)

func (a *API) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)
		a.logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(startedAt).String(),
		)
	})
}
