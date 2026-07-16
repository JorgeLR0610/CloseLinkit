package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JorgeLR0610/CloseLinkit/internal/repository"
	"github.com/JorgeLR0610/CloseLinkit/internal/service"
	"github.com/jackc/pgx/v5/pgconn"
)

// Mock for URLRepository
type mockURLRepository struct {
	CreateURLFunc           func(ctx context.Context, arg repository.CreateURLParams) (repository.CreateURLRow, error)
	createURLCalls          int
}

func (m *mockURLRepository) CreateURL(ctx context.Context, arg repository.CreateURLParams) (repository.CreateURLRow, error) {
	m.createURLCalls++
	if m.CreateURLFunc != nil {
		return m.CreateURLFunc(ctx, arg)
	}
	return repository.CreateURLRow{}, nil
}

func (m *mockURLRepository) GetURL(ctx context.Context, shortCode string) (string, error) {
	return "", nil
}

func (m *mockURLRepository) GetURLStats(ctx context.Context, shortCode string) (repository.GetURLStatsRow, error) {
	return repository.GetURLStatsRow{}, nil
}

func (m *mockURLRepository) IncrementClickCount(ctx context.Context, shortCode string) error {
	return nil
}

// Mock for ShortCodeGenerator
type mockShortCodeGenerator struct {
	GenerateShortCodeFunc func() (string, error)
}

func (m *mockShortCodeGenerator) GenerateShortCode() (string, error) {
	if m.GenerateShortCodeFunc != nil {
		return m.GenerateShortCodeFunc()
	}
	return "", nil
}

func TestURLService_CreateShortCode(t *testing.T) {
	tests := []struct {
		name              string
		originalURL       string
		setupGenerator    func() *mockShortCodeGenerator
		setupRepo         func() *mockURLRepository
		expectedErr       error
		expectedRepoCalls int
		expectedShortCode string
	}{
		{
			name:        "Valid URL, no collision",
			originalURL: "https://example.com",
			setupGenerator: func() *mockShortCodeGenerator {
				return &mockShortCodeGenerator{
					GenerateShortCodeFunc: func() (string, error) {
						return "abcDEFg", nil
					},
				}
			},
			setupRepo: func() *mockURLRepository {
				return &mockURLRepository{
					CreateURLFunc: func(ctx context.Context, arg repository.CreateURLParams) (repository.CreateURLRow, error) {
						return repository.CreateURLRow{
							ShortCode: arg.ShortCode,
						}, nil
					},
				}
			},
			expectedErr:       nil,
			expectedRepoCalls: 1,
			expectedShortCode: "abcDEFg",
		},
		{
			name:        "Malformed URL",
			originalURL: "://invalid", // Causes url.Parse error
			setupGenerator: func() *mockShortCodeGenerator {
				return &mockShortCodeGenerator{}
			},
			setupRepo: func() *mockURLRepository {
				return &mockURLRepository{}
			},
			expectedErr:       service.ErrInvalidURL,
			expectedRepoCalls: 0,
		},
		{
			name:        "Invalid Scheme (ftp)",
			originalURL: "ftp://example.com",
			setupGenerator: func() *mockShortCodeGenerator {
				return &mockShortCodeGenerator{}
			},
			setupRepo: func() *mockURLRepository {
				return &mockURLRepository{}
			},
			expectedErr:       service.ErrInvalidURLScheme,
			expectedRepoCalls: 0,
		},
		{
			name:        "Short code collision, first attempt fails, second succeeds",
			originalURL: "https://example.com",
			setupGenerator: func() *mockShortCodeGenerator {
				codes := []string{"nmWbCno", "abcDEFg"}
				callCount := 0
				return &mockShortCodeGenerator{
					GenerateShortCodeFunc: func() (string, error) {
						code := codes[callCount]
						callCount++
						return code, nil
					},
				}
			},
			setupRepo: func() *mockURLRepository {
				callCount := 0
				return &mockURLRepository{
					CreateURLFunc: func(ctx context.Context, arg repository.CreateURLParams) (repository.CreateURLRow, error) {
						callCount++
						if callCount == 1 {
							return repository.CreateURLRow{}, &pgconn.PgError{
								Code:           "23505",
								ConstraintName: "urls_short_code_unique",
							}
						}
						return repository.CreateURLRow{
							ShortCode: arg.ShortCode,
						}, nil
					},
				}
			},
			expectedErr:       nil,
			expectedRepoCalls: 2,
			expectedShortCode: "abcDEFg",
		},
		{
			name:        "Five consecutive collisions",
			originalURL: "https://example.com",
			setupGenerator: func() *mockShortCodeGenerator {
				return &mockShortCodeGenerator{
					GenerateShortCodeFunc: func() (string, error) {
						return "abcDEFg", nil
					},
				}
			},
			setupRepo: func() *mockURLRepository {
				return &mockURLRepository{
					CreateURLFunc: func(ctx context.Context, arg repository.CreateURLParams) (repository.CreateURLRow, error) {
						return repository.CreateURLRow{}, &pgconn.PgError{
							Code:           "23505",
							ConstraintName: "urls_short_code_unique",
						}
					},
				}
			},
			expectedErr:       service.ErrCouldNotGenerateUniqueShortCode,
			expectedRepoCalls: 5,
		},
		{
			name:        "Other non-constraint violation PostgreSQL error",
			originalURL: "https://example.com",
			setupGenerator: func() *mockShortCodeGenerator {
				return &mockShortCodeGenerator{
					GenerateShortCodeFunc: func() (string, error) {
						return "abcDEFg", nil
					},
				}
			},
			setupRepo: func() *mockURLRepository {
				return &mockURLRepository{
					CreateURLFunc: func(ctx context.Context, arg repository.CreateURLParams) (repository.CreateURLRow, error) {
						return repository.CreateURLRow{}, errors.New("connection failed")
					},
				}
			},
			// Matching exact wrapped error string returned by URLService
			expectedErr:       errors.New("could not insert URL to database: connection failed"),
			expectedRepoCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := tt.setupGenerator()
			repo := tt.setupRepo()
			srv := service.NewURLService(repo, generator)

			result, err := srv.CreateShortCode(context.Background(), tt.originalURL)

			if tt.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}
				if err.Error() != tt.expectedErr.Error() {
					if !errors.Is(err, tt.expectedErr) {
						t.Errorf("expected error %v, got %v", tt.expectedErr, err)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("did not expect error, got %v", err)
				}
				if result.ShortCode != tt.expectedShortCode {
					t.Errorf("expected shortCode %s, got %s", tt.expectedShortCode, result.ShortCode)
				}
			}

			if repo.createURLCalls != tt.expectedRepoCalls {
				t.Errorf("expected repo CreateURL to be called %d times, was called %d times", tt.expectedRepoCalls, repo.createURLCalls)
			}
		})
	}
}
