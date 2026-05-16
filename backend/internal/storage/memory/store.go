package memory

import (
	"context"
	"sync"

	"shortener/backend/internal/domain"
	"shortener/backend/internal/service"
)

type Store struct {
	mu     sync.RWMutex
	links  map[string]domain.Link
	clicks map[string][]domain.Click
}

func New() *Store {
	return &Store{
		links:  make(map[string]domain.Link),
		clicks: make(map[string][]domain.Click),
	}
}

func (s *Store) CreateLink(_ context.Context, link domain.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.links[link.ShortCode]; exists {
		return service.ErrShortCodeExists
	}

	s.links[link.ShortCode] = link
	return nil
}

func (s *Store) GetLinkByCode(_ context.Context, code string) (domain.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	link, exists := s.links[code]
	if !exists {
		return domain.Link{}, service.ErrShortCodeNotFound
	}

	return link, nil
}

func (s *Store) SaveClick(_ context.Context, click domain.Click) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.links[click.ShortCode]; !exists {
		return service.ErrShortCodeNotFound
	}

	s.clicks[click.ShortCode] = append(s.clicks[click.ShortCode], click)
	return nil
}

func (s *Store) ListClicks(_ context.Context, code string) ([]domain.Click, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.links[code]; !exists {
		return nil, service.ErrShortCodeNotFound
	}

	clicks := s.clicks[code]
	result := make([]domain.Click, len(clicks))
	copy(result, clicks)
	return result, nil
}
