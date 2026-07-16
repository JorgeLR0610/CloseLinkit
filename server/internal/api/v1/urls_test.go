package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JorgeLR0610/CloseLinkit/internal/api/v1"
	"github.com/JorgeLR0610/CloseLinkit/internal/repository"
	"github.com/JorgeLR0610/CloseLinkit/internal/service"
	"github.com/jackc/pgx/v5/pgtype"
)

// mockURLService implements api.URLServicer
type mockURLService struct {
	CreateShortCodeFunc  func(ctx context.Context, originalURL string) (repository.CreateURLRow, error)
	ResolveShortCodeFunc func(ctx context.Context, shortCode string) (string, error)
	GetURLStatsFunc      func(ctx context.Context, shortCode string) (repository.GetURLStatsRow, error)
}

func (m *mockURLService) CreateShortCode(ctx context.Context, originalURL string) (repository.CreateURLRow, error) {
	if m.CreateShortCodeFunc != nil {
		return m.CreateShortCodeFunc(ctx, originalURL)
	}
	return repository.CreateURLRow{}, nil
}

func (m *mockURLService) ResolveShortCode(ctx context.Context, shortCode string) (string, error) {
	if m.ResolveShortCodeFunc != nil {
		return m.ResolveShortCodeFunc(ctx, shortCode)
	}
	return "", nil
}

func (m *mockURLService) GetURLStats(ctx context.Context, shortCode string) (repository.GetURLStatsRow, error) {
	if m.GetURLStatsFunc != nil {
		return m.GetURLStatsFunc(ctx, shortCode)
	}
	return repository.GetURLStatsRow{}, nil
}

func TestURLHandler_HandlerCreateURL(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name           string
		requestBody    string
		setupMock      func() *mockURLService
		expectedStatus int
	}{
		{
			name:        "Valid JSON",
			requestBody: `{"url":"https://example.com"}`,
			setupMock: func() *mockURLService {
				return &mockURLService{
					CreateShortCodeFunc: func(ctx context.Context, originalURL string) (repository.CreateURLRow, error) {
						var uuid pgtype.UUID
						uuid.Scan("123e4567-e89b-12d3-a456-426614174000")
						return repository.CreateURLRow{
							ID:          uuid,
							OriginalUrl: "https://example.com",
							ShortCode:   "abcdef",
							CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
						}, nil
					},
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:        "Invalid JSON",
			requestBody: `{"url":}`,
			setupMock: func() *mockURLService {
				return &mockURLService{}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Unknown field",
			requestBody: `{"url":"https://example.com", "unknown":"field"}`,
			setupMock: func() *mockURLService {
				return &mockURLService{}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Invalid URL (Service Error)",
			requestBody: `{"url":"://invalid"}`,
			setupMock: func() *mockURLService {
				return &mockURLService{
					CreateShortCodeFunc: func(ctx context.Context, originalURL string) (repository.CreateURLRow, error) {
						return repository.CreateURLRow{}, service.ErrInvalidURL
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Internal Server Error",
			requestBody: `{"url":"https://example.com"}`,
			setupMock: func() *mockURLService {
				return &mockURLService{
					CreateShortCodeFunc: func(ctx context.Context, originalURL string) (repository.CreateURLRow, error) {
						return repository.CreateURLRow{}, errors.New("db connection lost")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setupMock()
			handler := api.NewURLHandler(svc, logger)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.HandlerCreateURL(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if w.Code == http.StatusCreated {
				if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", contentType)
				}

				var resp api.CreateURLResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("could not unmarshal response: %v", err)
				}
				if resp.ShortCode == "" || resp.OriginalURL == "" {
					t.Errorf("expected response to have populated fields, got: %+v", resp)
				}
			}
		})
	}
}

func TestURLHandler_HandlerGetURL(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name             string
		shortCode        string
		setupMock        func() *mockURLService
		expectedStatus   int
		expectedLocation string
	}{
		{
			name:      "Existing shortcode",
			shortCode: "abcdef",
			setupMock: func() *mockURLService {
				return &mockURLService{
					ResolveShortCodeFunc: func(ctx context.Context, shortCode string) (string, error) {
						return "https://example.com", nil
					},
				}
			},
			expectedStatus:   http.StatusFound,
			expectedLocation: "https://example.com",
		},
		{
			name:      "Non-existing shortcode",
			shortCode: "notfnd",
			setupMock: func() *mockURLService {
				return &mockURLService{
					ResolveShortCodeFunc: func(ctx context.Context, shortCode string) (string, error) {
						return "", service.ErrNoURLFound
					},
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:      "Internal server error",
			shortCode: "errr",
			setupMock: func() *mockURLService {
				return &mockURLService{
					ResolveShortCodeFunc: func(ctx context.Context, shortCode string) (string, error) {
						return "", errors.New("db error")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setupMock()
			handler := api.NewURLHandler(svc, logger)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/"+tt.shortCode, nil)
			req.SetPathValue("shortCode", tt.shortCode)
			w := httptest.NewRecorder()

			handler.HandlerGetURL(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if w.Code == http.StatusFound {
				if loc := w.Header().Get("Location"); loc != tt.expectedLocation {
					t.Errorf("expected Location %s, got %s", tt.expectedLocation, loc)
				}
			}
		})
	}
}

func TestURLHandler_HandlerGetURLStats(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name           string
		shortCode      string
		setupMock      func() *mockURLService
		expectedStatus int
	}{
		{
			name:      "Existing shortcode",
			shortCode: "abcdef",
			setupMock: func() *mockURLService {
				return &mockURLService{
					GetURLStatsFunc: func(ctx context.Context, shortCode string) (repository.GetURLStatsRow, error) {
						return repository.GetURLStatsRow{
							OriginalUrl: "https://example.com",
							ClickCount:  10,
							CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
						}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Non-existing shortcode",
			shortCode: "notfnd",
			setupMock: func() *mockURLService {
				return &mockURLService{
					GetURLStatsFunc: func(ctx context.Context, shortCode string) (repository.GetURLStatsRow, error) {
						return repository.GetURLStatsRow{}, service.ErrNoURLFound
					},
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setupMock()
			handler := api.NewURLHandler(svc, logger)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/"+tt.shortCode+"/stats", nil)
			req.SetPathValue("shortCode", tt.shortCode)
			w := httptest.NewRecorder()

			handler.HandlerGetURLStats(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if w.Code == http.StatusOK {
				var resp api.GetURLStatsResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("could not unmarshal response: %v", err)
				}
				if resp.OriginalURL != "https://example.com" || resp.ClickCount != 10 {
					t.Errorf("expected response to have populated fields correctly, got: %+v", resp)
				}
			}
		})
	}
}
