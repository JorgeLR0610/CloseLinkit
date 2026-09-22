package api

import (
	"time"

	"github.com/JorgeLR0610/CloseLinkit/internal/service"
)

type CreateURLResponse struct {
	ShortURL string `json:"short_url"`
}

type GetURLStatsResponse struct {
	ClickCount int       `json:"click_count"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserURLResponse struct {
	OriginalURL string    `json:"original_url"`
	ShortCode   string    `json:"short_code"`
	ShortURL    string    `json:"short_url"`
	CreatedAt   time.Time `json:"created_at"`
	ClickCount  int       `json:"click_count"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse = service.UserResponse

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token"`
	ExpiresIn    int64                `json:"expires_in"`
	User         service.UserResponse `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}
