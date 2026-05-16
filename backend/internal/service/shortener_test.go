package service_test

import (
	"context"
	"errors"
	"testing"

	"shortener/backend/internal/service"
	"shortener/backend/internal/storage/memory"
)

func TestCreateLinkWithCustomAlias(t *testing.T) {
	shortener := service.NewShortener(memory.New(), service.Options{PublicBaseURL: "http://localhost:8080"})

	link, err := shortener.CreateLink(context.Background(), service.CreateLinkInput{
		OriginalURL: "https://example.com/page?q=1",
		CustomAlias: "example-page",
	})
	if err != nil {
		t.Fatalf("CreateLink returned error: %v", err)
	}

	if link.ShortCode != "example-page" {
		t.Fatalf("expected custom alias, got %q", link.ShortCode)
	}
	if link.ShortURL != "http://localhost:8080/s/example-page" {
		t.Fatalf("unexpected short url: %q", link.ShortURL)
	}
}

func TestCreateLinkRejectsInvalidURL(t *testing.T) {
	shortener := service.NewShortener(memory.New(), service.Options{})

	_, err := shortener.CreateLink(context.Background(), service.CreateLinkInput{OriginalURL: "ftp://example.com"})
	if !errors.Is(err, service.ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}
}

func TestResolveTracksClicksAndAnalytics(t *testing.T) {
	shortener := service.NewShortener(memory.New(), service.Options{})
	link, err := shortener.CreateLink(context.Background(), service.CreateLinkInput{
		OriginalURL: "https://example.com",
		CustomAlias: "docs",
	})
	if err != nil {
		t.Fatalf("CreateLink returned error: %v", err)
	}

	_, err = shortener.Resolve(context.Background(), service.RegisterClickInput{
		ShortCode: link.ShortCode,
		UserAgent: "Go-http-client/1.1",
		IP:        "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}

	analytics, err := shortener.Analytics(context.Background(), link.ShortCode)
	if err != nil {
		t.Fatalf("Analytics returned error: %v", err)
	}

	if analytics.TotalClicks != 1 {
		t.Fatalf("expected one click, got %d", analytics.TotalClicks)
	}
	if len(analytics.ByUserAgent) != 1 || analytics.ByUserAgent[0].Key != "Go-http-client/1.1" {
		t.Fatalf("unexpected user-agent aggregation: %+v", analytics.ByUserAgent)
	}
}
