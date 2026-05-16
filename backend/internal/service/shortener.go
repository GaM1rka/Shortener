package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"shortener/backend/internal/domain"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var customAliasPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{3,64}$`)

type Options struct {
	PublicBaseURL string
	CodeLength    int
}

type Shortener struct {
	store         Store
	publicBaseURL string
	codeLength    int
}

type CreateLinkInput struct {
	OriginalURL string
	CustomAlias string
}

type RegisterClickInput struct {
	ShortCode string
	UserAgent string
	IP        string
}

func NewShortener(store Store, opts Options) *Shortener {
	codeLength := opts.CodeLength
	if codeLength <= 0 {
		codeLength = 7
	}

	return &Shortener{
		store:         store,
		publicBaseURL: strings.TrimRight(opts.PublicBaseURL, "/"),
		codeLength:    codeLength,
	}
}

func (s *Shortener) CreateLink(ctx context.Context, input CreateLinkInput) (domain.Link, error) {
	originalURL, err := normalizeURL(input.OriginalURL)
	if err != nil {
		return domain.Link{}, err
	}

	now := time.Now().UTC()

	if input.CustomAlias != "" {
		if !customAliasPattern.MatchString(input.CustomAlias) {
			return domain.Link{}, ErrInvalidCustomAlias
		}

		link := domain.Link{
			ID:          newID(),
			OriginalURL: originalURL,
			ShortCode:   input.CustomAlias,
			ShortURL:    s.shortURL(input.CustomAlias),
			CreatedAt:   now,
		}
		if err := s.store.CreateLink(ctx, link); err != nil {
			return domain.Link{}, err
		}
		return link, nil
	}

	for attempt := 0; attempt < 5; attempt++ {
		code, err := randomCode(s.codeLength)
		if err != nil {
			return domain.Link{}, err
		}

		link := domain.Link{
			ID:          newID(),
			OriginalURL: originalURL,
			ShortCode:   code,
			ShortURL:    s.shortURL(code),
			CreatedAt:   now,
		}

		err = s.store.CreateLink(ctx, link)
		if err == nil {
			return link, nil
		}
		if !errors.Is(err, ErrShortCodeExists) {
			return domain.Link{}, err
		}
	}

	return domain.Link{}, ErrUnableToCreateAlias
}

func (s *Shortener) Resolve(ctx context.Context, input RegisterClickInput) (domain.Link, error) {
	link, err := s.store.GetLinkByCode(ctx, input.ShortCode)
	if err != nil {
		return domain.Link{}, err
	}

	click := domain.Click{
		ID:        newID(),
		LinkID:    link.ID,
		ShortCode: link.ShortCode,
		UserAgent: input.UserAgent,
		IP:        input.IP,
		ClickedAt: time.Now().UTC(),
	}

	if err := s.store.SaveClick(ctx, click); err != nil {
		return domain.Link{}, err
	}

	return link, nil
}

func (s *Shortener) Analytics(ctx context.Context, code string) (domain.Analytics, error) {
	link, err := s.store.GetLinkByCode(ctx, code)
	if err != nil {
		return domain.Analytics{}, err
	}

	clicks, err := s.store.ListClicks(ctx, code)
	if err != nil {
		return domain.Analytics{}, err
	}

	sort.Slice(clicks, func(i, j int) bool {
		return clicks[i].ClickedAt.After(clicks[j].ClickedAt)
	})

	return domain.Analytics{
		ShortCode:   link.ShortCode,
		OriginalURL: link.OriginalURL,
		TotalClicks: len(clicks),
		Clicks:      clicks,
		ByDay:       aggregate(clicks, func(click domain.Click) string { return click.ClickedAt.Format("2006-01-02") }),
		ByMonth:     aggregate(clicks, func(click domain.Click) string { return click.ClickedAt.Format("2006-01") }),
		ByUserAgent: aggregate(clicks, func(click domain.Click) string { return fallback(click.UserAgent, "unknown") }),
		GeneratedAt: time.Now().UTC(),
	}, nil
}

func (s *Shortener) shortURL(code string) string {
	if s.publicBaseURL == "" {
		return "/s/" + code
	}
	return s.publicBaseURL + "/s/" + code
}

func normalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return "", ErrInvalidURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", ErrInvalidURL
	}
	if parsed.Host == "" {
		return "", ErrInvalidURL
	}
	return parsed.String(), nil
}

func randomCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.Grow(length)
	for _, b := range bytes {
		builder.WriteByte(alphabet[int(b)%len(alphabet)])
	}
	return builder.String(), nil
}

func newID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(bytes)
}

func aggregate(clicks []domain.Click, keyFunc func(domain.Click) string) []domain.Bucket {
	counts := make(map[string]int)
	for _, click := range clicks {
		counts[keyFunc(click)]++
	}

	buckets := make([]domain.Bucket, 0, len(counts))
	for key, count := range counts {
		buckets = append(buckets, domain.Bucket{Key: key, Count: count})
	}

	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Key < buckets[j].Key
	})

	return buckets
}

func fallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
