package service

import (
	"context"
	"time"

	"github.com/JorgeLR0610/CloseLinkit/internal/repository"
)

type URLService struct {
	repo *repository.Queries
}

func NewURLService(repo *repository.Queries) *URLService {
	return &URLService{
		repo: repo,
	}
}

func (s *URLService) CreateURL(ctx context.Context, originalURL, shortCode string, expiresAt time.Time) {
	
}