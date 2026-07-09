package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/JorgeLR0610/CloseLinkit/internal/generator"
	"github.com/JorgeLR0610/CloseLinkit/internal/repository"
)

var ErrInvalidURL = errors.New("invalid URL scheme")
var ErrNoHost = errors.New("no such host")

type URLService struct {
	repo 	  *repository.Queries
	generator *generator.ShortCodeGenerator
}

func NewURLService(repo *repository.Queries, generator *generator.ShortCodeGenerator) *URLService {
	return &URLService{
		repo: repo,
		generator: generator,
	}
}

func (s *URLService) CreateURL(ctx context.Context, originalURL string) (repository.Url, error) {
	parsedURL, err := url.Parse(strings.TrimSpace(originalURL))
	if err != nil {
		return repository.Url{}, fmt.Errorf("Error parsing URL: %w", err)
	}

	if parsedURL.Scheme != "http" || parsedURL.Scheme != "https" {
		return repository.Url{}, ErrInvalidURL
	}

	_, err = net.LookupHost(parsedURL.Host)
	if err != nil {
		return repository.Url{}, ErrNoHost
	}

	// Generate short code
	shortCode, err := s.generator.GenerateShortCode()
	if err != nil {
		return repository.Url{}, fmt.Errorf("error generating short code: %w", err)
	}

	
	
}

