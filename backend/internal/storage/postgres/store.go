package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"shortener/backend/internal/domain"
	"shortener/backend/internal/service"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *Store) CreateLink(ctx context.Context, link domain.Link) error {
	const query = `
		INSERT INTO links (id, original_url, short_code, short_url, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.pool.Exec(ctx, query, link.ID, link.OriginalURL, link.ShortCode, link.ShortURL, link.CreatedAt)
	if isUniqueViolation(err) {
		return service.ErrShortCodeExists
	}
	return err
}

func (s *Store) GetLinkByCode(ctx context.Context, code string) (domain.Link, error) {
	const query = `
		SELECT id, original_url, short_code, short_url, created_at
		FROM links
		WHERE short_code = $1
	`

	var link domain.Link
	err := s.pool.QueryRow(ctx, query, code).Scan(
		&link.ID,
		&link.OriginalURL,
		&link.ShortCode,
		&link.ShortURL,
		&link.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Link{}, service.ErrShortCodeNotFound
	}
	return link, err
}

func (s *Store) SaveClick(ctx context.Context, click domain.Click) error {
	const query = `
		INSERT INTO clicks (id, link_id, short_code, user_agent, ip, clicked_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := s.pool.Exec(ctx, query, click.ID, click.LinkID, click.ShortCode, click.UserAgent, click.IP, click.ClickedAt)
	if isForeignKeyViolation(err) {
		return service.ErrShortCodeNotFound
	}
	return err
}

func (s *Store) ListClicks(ctx context.Context, code string) ([]domain.Click, error) {
	const query = `
		SELECT c.id, c.link_id, c.short_code, c.user_agent, c.ip, c.clicked_at
		FROM clicks c
		INNER JOIN links l ON l.id = c.link_id
		WHERE l.short_code = $1
		ORDER BY c.clicked_at DESC
	`

	rows, err := s.pool.Query(ctx, query, code)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clicks []domain.Click
	for rows.Next() {
		var click domain.Click
		if err := rows.Scan(
			&click.ID,
			&click.LinkID,
			&click.ShortCode,
			&click.UserAgent,
			&click.IP,
			&click.ClickedAt,
		); err != nil {
			return nil, err
		}
		clicks = append(clicks, click)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	linkExists, err := s.linkExists(ctx, code)
	if err != nil {
		return nil, err
	}
	if !linkExists {
		return nil, service.ErrShortCodeNotFound
	}

	return clicks, nil
}

func (s *Store) linkExists(ctx context.Context, code string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM links WHERE short_code = $1)`

	var exists bool
	err := s.pool.QueryRow(ctx, query, code).Scan(&exists)
	return exists, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
