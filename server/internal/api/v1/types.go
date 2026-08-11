package api

import (
	"time"
)

type CreateURLResponse struct {
	ShortURL string `json:"short_url"`
}

type GetURLStatsResponse struct {
	ClickCount  int       `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
}
