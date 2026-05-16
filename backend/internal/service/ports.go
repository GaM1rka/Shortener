package service

import (
	"context"

	"shortener/backend/internal/domain"
)

type Store interface {
	CreateLink(ctx context.Context, link domain.Link) error
	GetLinkByCode(ctx context.Context, code string) (domain.Link, error)
	SaveClick(ctx context.Context, click domain.Click) error
	ListClicks(ctx context.Context, code string) ([]domain.Click, error)
}
