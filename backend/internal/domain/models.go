package domain

import "time"

type Link struct {
	ID          string    `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortCode   string    `json:"short_code"`
	ShortURL    string    `json:"short_url"`
	CreatedAt   time.Time `json:"created_at"`
}

type Click struct {
	ID        string    `json:"id"`
	LinkID    string    `json:"link_id"`
	ShortCode string    `json:"short_code"`
	UserAgent string    `json:"user_agent"`
	IP        string    `json:"ip"`
	ClickedAt time.Time `json:"clicked_at"`
}

type Analytics struct {
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	TotalClicks int       `json:"total_clicks"`
	Clicks      []Click   `json:"clicks"`
	ByDay       []Bucket  `json:"by_day"`
	ByMonth     []Bucket  `json:"by_month"`
	ByUserAgent []Bucket  `json:"by_user_agent"`
	GeneratedAt time.Time `json:"generated_at"`
}

type Bucket struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}
