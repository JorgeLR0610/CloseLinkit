package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/JorgeLR0610/CloseLinkit/internal/generator"
	"github.com/JorgeLR0610/CloseLinkit/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrInvalidURLScheme = errors.New("invalid URL scheme")
var ErrNoHost = errors.New("invalid host")
var ErrInvalidURL = errors.New("invalid URL")
var ErrNoURLFound = errors.New("short URL not found")

var	ErrCouldNotGenerateUniqueShortCode = errors.New("could not generate unique short code")

const uniqueViolation = "23505"

const maxRetries = 5

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

func (s *URLService) CreateShortCode(ctx context.Context, originalURL string) (repository.CreateURLRow, error) {
	parsedURL, err := url.Parse(strings.TrimSpace(originalURL))
	if err != nil {
		return repository.CreateURLRow{}, ErrInvalidURL
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return repository.CreateURLRow{}, ErrInvalidURL
	}

	if parsedURL.Hostname() == "" {
		return repository.CreateURLRow{}, ErrNoHost
	}

	// Try to create and store short code, up to the defined number of attempts
	for range maxRetries {
		shortCode, err := s.generator.GenerateShortCode()
		if err != nil {
			return repository.CreateURLRow{}, fmt.Errorf("error generating short code: %w", err)
		}

		createdURL, err := s.repo.CreateURL(ctx, repository.CreateURLParams{
			OriginalUrl: parsedURL.String(),
			ShortCode: shortCode,
		})
		if err != nil {
			if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
				if pgErr.SQLState() == uniqueViolation && pgErr.ConstraintName == "urls_short_code_unique" {
					continue
				}
			}
			return repository.CreateURLRow{}, fmt.Errorf("could not insert URL to database: %w", err)
		}

		return createdURL, nil
	}
	return repository.CreateURLRow{}, ErrCouldNotGenerateUniqueShortCode
}

func (s *URLService) ResolveShortCode(ctx context.Context, shortCode string) (string, error) {
	retrievedURL, err := s.repo.GetURL(ctx, shortCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNoURLFound
		}

		return "", fmt.Errorf("error retrieving URL: %w", err)
	}

	if err := s.repo.IncrementClickCount(ctx, shortCode); err != nil {
		return "", fmt.Errorf("error incrementing click count: %w", err)
	}

	return retrievedURL, nil
}

