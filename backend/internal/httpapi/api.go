package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"

	"shortener/backend/internal/service"
	"shortener/backend/pkg/response"
)

type API struct {
	shortener *service.Shortener
	logger    *slog.Logger
}

type shortenRequest struct {
	URL         string `json:"url"`
	CustomAlias string `json:"custom_alias"`
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) createShortLink(w http.ResponseWriter, r *http.Request) {
	var request shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	link, err := a.shortener.CreateLink(r.Context(), service.CreateLinkInput{
		OriginalURL: request.URL,
		CustomAlias: request.CustomAlias,
	})
	if err != nil {
		a.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, link)
}

func (a *API) redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("short_url")
	link, err := a.shortener.Resolve(r.Context(), service.RegisterClickInput{
		ShortCode: code,
		UserAgent: r.UserAgent(),
		IP:        clientIP(r),
	})
	if err != nil {
		a.writeServiceError(w, err)
		return
	}

	http.Redirect(w, r, link.OriginalURL, http.StatusFound)
}

func (a *API) analytics(w http.ResponseWriter, r *http.Request) {
	analytics, err := a.shortener.Analytics(r.Context(), r.PathValue("short_url"))
	if err != nil {
		a.writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, analytics)
}

func (a *API) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidURL):
		response.Error(w, http.StatusBadRequest, "url must be an absolute http or https URL")
	case errors.Is(err, service.ErrInvalidCustomAlias):
		response.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrShortCodeExists):
		response.Error(w, http.StatusConflict, "short alias is already in use")
	case errors.Is(err, service.ErrShortCodeNotFound):
		response.Error(w, http.StatusNotFound, "short url not found")
	default:
		a.logger.Error("request failed", "error", err)
		response.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func clientIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		return forwardedFor
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
