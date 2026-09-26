package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JorgeLR0610/CloseLinkit/internal/repository"
	"github.com/JorgeLR0610/CloseLinkit/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// Mock for URLRepository
type mockURLRepository struct {
	CreateURLFunc           func(ctx context.Context, arg repository.CreateURLParams) (string, error)
	GetURLFunc              func(ctx context.Context, shortCode string) (string, error)
	GetURLStatsFunc          func(ctx context.Context, shortCode string) (repository.GetURLStatsRow, error)
	IncrementClickCountFunc func(ctx context.Context, shortCode string) error
	GetURLsByUserIDFunc     func(ctx context.Context, userID pgtype.UUID) ([]repository.GetURLsByUserIDRow, error)
	createURLCalls          int
	incrementClickCalls     int
}

func (m *mockURLRepository) CreateURL(ctx context.Context, arg repository.CreateURLParams) (string, error) {
	m.createURLCalls++
	if m.CreateURLFunc != nil {
		return m.CreateURLFunc(ctx, arg)
	}
	return "", nil
}

func (m *mockURLRepository) GetURL(ctx context.Context, shortCode string) (string, error) {
	if m.GetURLFunc != nil {
		return m.GetURLFunc(ctx, shortCode)
	}
	return "", nil
}

func (m *mockURLRepository) GetURLStats(ctx context.Context, shortCode string) (repository.GetURLStatsRow, error) {
	if m.GetURLStatsFunc != nil {
		return m.GetURLStatsFunc(ctx, shortCode)
	}
	return repository.GetURLStatsRow{}, nil
}

func (m *mockURLRepository) IncrementClickCount(ctx context.Context, shortCode string) error {
	m.incrementClickCalls++
	if m.IncrementClickCountFunc != nil {
		return m.IncrementClickCountFunc(ctx, shortCode)
	}
	return nil
}

func (m *mockURLRepository) GetURLsByUserID(ctx context.Context, userID pgtype.UUID) ([]repository.GetURLsByUserIDRow, error) {
	if m.GetURLsByUserIDFunc != nil {
		return m.GetURLsByUserIDFunc(ctx, userID)
	}
	return nil, nil
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
					CreateURLFunc: func(ctx context.Context, arg repository.CreateURLParams) (string, error) {
						return arg.ShortCode, nil
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
					CreateURLFunc: func(ctx context.Context, arg repository.CreateURLParams) (string, error) {
						callCount++
						if callCount == 1 {
							return "", &pgconn.PgError{
								Code:           "23505",
								ConstraintName: "urls_short_code_unique",
							}
						}
						return arg.ShortCode, nil
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
					CreateURLFunc: func(ctx context.Context, arg repository.CreateURLParams) (string, error) {
						return "", &pgconn.PgError{
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
					CreateURLFunc: func(ctx context.Context, arg repository.CreateURLParams) (string, error) {
						return "", errors.New("connection failed")
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
				if result != tt.expectedShortCode {
					t.Errorf("expected shortCode %s, got %s", tt.expectedShortCode, result)
				}
			}

			if repo.createURLCalls != tt.expectedRepoCalls {
				t.Errorf("expected repo CreateURL to be called %d times, was called %d times", tt.expectedRepoCalls, repo.createURLCalls)
			}
		})
	}
}

func TestURLService_CreateShortCode_WithAuthenticatedUser(t *testing.T) {
	testUserID := uuid.New()
	generator := &mockShortCodeGenerator{
		GenerateShortCodeFunc: func() (string, error) {
			return "abcDEFg", nil
		},
	}

	t.Run("authenticated context sets user_id", func(t *testing.T) {
		var capturedArg repository.CreateURLParams
		repo := &mockURLRepository{
			CreateURLFunc: func(ctx context.Context, arg repository.CreateURLParams) (string, error) {
				capturedArg = arg
				return arg.ShortCode, nil
			},
		}

		srv := service.NewURLService(repo, generator)
		ctx := service.ContextWithUserID(context.Background(), testUserID)

		_, err := srv.CreateShortCode(ctx, "https://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !capturedArg.UserID.Valid {
			t.Fatal("expected UserID to be valid")
		}
		if capturedArg.UserID.Bytes != testUserID {
			t.Errorf("expected UserID %v, got %v", testUserID, capturedArg.UserID.Bytes)
		}
	})

	t.Run("unauthenticated context leaves user_id invalid (null)", func(t *testing.T) {
		var capturedArg repository.CreateURLParams
		repo := &mockURLRepository{
			CreateURLFunc: func(ctx context.Context, arg repository.CreateURLParams) (string, error) {
				capturedArg = arg
				return arg.ShortCode, nil
			},
		}

		srv := service.NewURLService(repo, generator)

		_, err := srv.CreateShortCode(context.Background(), "https://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if capturedArg.UserID.Valid {
			t.Error("expected UserID.Valid to be false for anonymous shorten")
		}
	})
}

func TestURLService_GetURLsByUserID(t *testing.T) {
	generator := &mockShortCodeGenerator{}
	testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	now := time.Now().Truncate(time.Second)

	t.Run("successful retrieval of user URLs", func(t *testing.T) {
		repo := &mockURLRepository{
			GetURLsByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) ([]repository.GetURLsByUserIDRow, error) {
				if !userID.Valid || userID.Bytes != testUserID {
					t.Fatalf("expected userID %v, got %v", testUserID, userID.Bytes)
				}
				return []repository.GetURLsByUserIDRow{
					{
						OriginalUrl: "https://example.com/one",
						ShortCode:   "code111",
						CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
						ClickCount:  5,
					},
					{
						OriginalUrl: "https://example.com/two",
						ShortCode:   "code222",
						CreatedAt:   pgtype.Timestamptz{Time: now.Add(-time.Hour), Valid: true},
						ClickCount:  0,
					},
				}, nil
			},
		}

		srv := service.NewURLService(repo, generator)
		urls, err := srv.GetURLsByUserID(context.Background(), testUserID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(urls) != 2 {
			t.Fatalf("expected 2 URLs, got %d", len(urls))
		}

		if urls[0].OriginalURL != "https://example.com/one" || urls[0].ShortCode != "code111" || urls[0].ClickCount != 5 {
			t.Errorf("unexpected first item: %+v", urls[0])
		}
		if urls[1].OriginalURL != "https://example.com/two" || urls[1].ShortCode != "code222" || urls[1].ClickCount != 0 {
			t.Errorf("unexpected second item: %+v", urls[1])
		}
	})

	t.Run("returns empty slice when user has no URLs", func(t *testing.T) {
		repo := &mockURLRepository{
			GetURLsByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) ([]repository.GetURLsByUserIDRow, error) {
				return []repository.GetURLsByUserIDRow{}, nil
			},
		}

		srv := service.NewURLService(repo, generator)
		urls, err := srv.GetURLsByUserID(context.Background(), testUserID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if urls == nil {
			t.Fatal("expected non-nil empty slice, got nil")
		}
		if len(urls) != 0 {
			t.Fatalf("expected 0 URLs, got %d", len(urls))
		}
	})

	t.Run("returns error on repository failure", func(t *testing.T) {
		dbErr := errors.New("database connection failed")
		repo := &mockURLRepository{
			GetURLsByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) ([]repository.GetURLsByUserIDRow, error) {
				return nil, dbErr
			},
		}

		srv := service.NewURLService(repo, generator)
		urls, err := srv.GetURLsByUserID(context.Background(), testUserID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, dbErr) {
			t.Errorf("expected wrapped dbErr, got %v", err)
		}
		if urls != nil {
			t.Errorf("expected nil result on error, got %+v", urls)
		}
	})
}

func TestURLService_ResolveShortCode(t *testing.T) {
	generator := &mockShortCodeGenerator{}

	t.Run("successful resolution increments click count", func(t *testing.T) {
		repo := &mockURLRepository{
			GetURLFunc: func(ctx context.Context, shortCode string) (string, error) {
				if shortCode != "validCode" {
					t.Fatalf("expected shortCode validCode, got %s", shortCode)
				}
				return "https://example.com/target", nil
			},
			IncrementClickCountFunc: func(ctx context.Context, shortCode string) error {
				if shortCode != "validCode" {
					t.Fatalf("expected shortCode validCode, got %s", shortCode)
				}
				return nil
			},
		}

		srv := service.NewURLService(repo, generator)
		url, err := srv.ResolveShortCode(context.Background(), "validCode")
		if err != nil {
			t.Fatalf("unexpected error resolving short code: %v", err)
		}
		if url != "https://example.com/target" {
			t.Errorf("expected https://example.com/target, got %s", url)
		}
		if repo.incrementClickCalls != 1 {
			t.Errorf("expected 1 increment click call, got %d", repo.incrementClickCalls)
		}
	})

	t.Run("non-existing short code returns ErrNoURLFound", func(t *testing.T) {
		repo := &mockURLRepository{
			GetURLFunc: func(ctx context.Context, shortCode string) (string, error) {
				return "", pgx.ErrNoRows
			},
		}

		srv := service.NewURLService(repo, generator)
		_, err := srv.ResolveShortCode(context.Background(), "nonexistent")
		if !errors.Is(err, service.ErrNoURLFound) {
			t.Errorf("expected ErrNoURLFound, got %v", err)
		}
		if repo.incrementClickCalls != 0 {
			t.Errorf("expected 0 increment click calls, got %d", repo.incrementClickCalls)
		}
	})

	t.Run("repository error on GetURL", func(t *testing.T) {
		dbErr := errors.New("db error")
		repo := &mockURLRepository{
			GetURLFunc: func(ctx context.Context, shortCode string) (string, error) {
				return "", dbErr
			},
		}

		srv := service.NewURLService(repo, generator)
		_, err := srv.ResolveShortCode(context.Background(), "someCode")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, dbErr) {
			t.Errorf("expected wrapped dbErr, got %v", err)
		}
	})

	t.Run("error incrementing click count", func(t *testing.T) {
		incErr := errors.New("increment failed")
		repo := &mockURLRepository{
			GetURLFunc: func(ctx context.Context, shortCode string) (string, error) {
				return "https://example.com", nil
			},
			IncrementClickCountFunc: func(ctx context.Context, shortCode string) error {
				return incErr
			},
		}

		srv := service.NewURLService(repo, generator)
		_, err := srv.ResolveShortCode(context.Background(), "someCode")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, incErr) {
			t.Errorf("expected wrapped incErr, got %v", err)
		}
	})
}

func TestURLService_GetURLStats(t *testing.T) {
	generator := &mockShortCodeGenerator{}
	now := time.Now().Truncate(time.Second)

	t.Run("successful stats retrieval", func(t *testing.T) {
		repo := &mockURLRepository{
			GetURLStatsFunc: func(ctx context.Context, shortCode string) (repository.GetURLStatsRow, error) {
				if shortCode != "statCode" {
					t.Fatalf("expected statCode, got %s", shortCode)
				}
				return repository.GetURLStatsRow{
					ClickCount: 42,
					CreatedAt:  pgtype.Timestamptz{Time: now, Valid: true},
				}, nil
			},
		}

		srv := service.NewURLService(repo, generator)
		stats, err := srv.GetURLStats(context.Background(), "statCode")
		if err != nil {
			t.Fatalf("unexpected error getting stats: %v", err)
		}
		if stats.ClickCount != 42 {
			t.Errorf("expected ClickCount 42, got %d", stats.ClickCount)
		}
		if !stats.CreatedAt.Time.Equal(now) {
			t.Errorf("expected CreatedAt %v, got %v", now, stats.CreatedAt.Time)
		}
	})

	t.Run("non-existing short code returns ErrNoURLFound", func(t *testing.T) {
		repo := &mockURLRepository{
			GetURLStatsFunc: func(ctx context.Context, shortCode string) (repository.GetURLStatsRow, error) {
				return repository.GetURLStatsRow{}, pgx.ErrNoRows
			},
		}

		srv := service.NewURLService(repo, generator)
		_, err := srv.GetURLStats(context.Background(), "missingCode")
		if !errors.Is(err, service.ErrNoURLFound) {
			t.Errorf("expected ErrNoURLFound, got %v", err)
		}
	})

	t.Run("repository error on GetURLStats", func(t *testing.T) {
		dbErr := errors.New("db error")
		repo := &mockURLRepository{
			GetURLStatsFunc: func(ctx context.Context, shortCode string) (repository.GetURLStatsRow, error) {
				return repository.GetURLStatsRow{}, dbErr
			},
		}

		srv := service.NewURLService(repo, generator)
		_, err := srv.GetURLStats(context.Background(), "someCode")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, dbErr) {
			t.Errorf("expected wrapped dbErr, got %v", err)
		}
	})
}

func TestURLService_CreateShortCode_HostValidation(t *testing.T) {
	generator := &mockShortCodeGenerator{
		GenerateShortCodeFunc: func() (string, error) {
			return "validCode", nil
		},
	}
	repo := &mockURLRepository{
		CreateURLFunc: func(ctx context.Context, arg repository.CreateURLParams) (string, error) {
			return arg.ShortCode, nil
		},
	}
	srv := service.NewURLService(repo, generator)

	tests := []struct {
		name        string
		rawURL      string
		expectedErr error
	}{
		{
			name:        "Loopback IPv4 is rejected",
			rawURL:      "https://127.0.0.1/path",
			expectedErr: service.ErrNoHost,
		},
		{
			name:        "Loopback IPv6 is rejected",
			rawURL:      "https://[::1]/path",
			expectedErr: service.ErrNoHost,
		},
		{
			name:        "Private IPv4 10.x is rejected",
			rawURL:      "https://10.0.0.1/path",
			expectedErr: service.ErrNoHost,
		},
		{
			name:        "Private IPv4 192.168.x is rejected",
			rawURL:      "https://192.168.1.1/path",
			expectedErr: service.ErrNoHost,
		},
		{
			name:        "Private IPv4 172.16.x is rejected",
			rawURL:      "https://172.16.0.1/path",
			expectedErr: service.ErrNoHost,
		},
		{
			name:        "Link-local IPv4 is rejected",
			rawURL:      "https://169.254.1.1/path",
			expectedErr: service.ErrNoHost,
		},
		{
			name:        "Valid public IP is accepted",
			rawURL:      "https://8.8.8.8/dns",
			expectedErr: nil,
		},
		{
			name:        "Localhost is rejected",
			rawURL:      "https://localhost/path",
			expectedErr: service.ErrNoHost,
		},
		{
			name:        "Host without dot is rejected",
			rawURL:      "https://example/path",
			expectedErr: service.ErrNoHost,
		},
		{
			name:        "Host with trailing dot is rejected",
			rawURL:      "https://example.com./path",
			expectedErr: service.ErrNoHost,
		},
		{
			name:        "Empty host is rejected",
			rawURL:      "https:///path",
			expectedErr: service.ErrNoHost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := srv.CreateShortCode(context.Background(), tt.rawURL)
			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

