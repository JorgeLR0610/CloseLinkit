package api

import (
	"time"
)

type CreateURLResponse struct {
	ID			string				`json:"id"`
	OriginalURL string				`json:"original_url"`
    ShortCode   string				`json:"short_code"`
    CreatedAt   time.Time			`json:"created_at"`
}