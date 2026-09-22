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
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// mockURLService implements api.URLServicer
type mockURLService struct {
	CreateShortCodeFunc  func(ctx context.Context, originalURL string) (string, error)
	ResolveShortCodeFunc func(ctx context.Context, shortCode string) (string, error)
	GetURLStatsFunc      func(ctx context.Context, shortCode string) (repository.GetURLStatsRow, error)
	GetURLsByUserIDFunc  func(ctx context.Context, userID uuid.UUID) ([]service.UserURL, error)
}

func (m *mockURLService) CreateShortCode(ctx context.Context, originalURL string) (string, error) {
	if m.CreateShortCodeFunc != nil {
		return m.CreateShortCodeFunc(ctx, originalURL)
	}
	return "", nil
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

func (m *mockURLService) GetURLsByUserID(ctx context.Context, userID uuid.UUID) ([]service.UserURL, error) {
	if m.GetURLsByUserIDFunc != nil {
		return m.GetURLsByUserIDFunc(ctx, userID)
	}
	return nil, nil
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
			name:        "Valid short code",
			requestBody: `{"url":"https://example.com"}`,
			setupMock: func() *mockURLService {
				return &mockURLService{
					CreateShortCodeFunc: func(ctx context.Context, originalURL string) (string, error) {
						var uuid pgtype.UUID
						if err := uuid.Scan("123e4567-e89b-12d3-a456-426614174000"); err != nil {
							t.Fatalf("could not write response: %v", err)
						}
						return "abcdef", nil
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
					CreateShortCodeFunc: func(ctx context.Context, originalURL string) (string, error) {
						return "", service.ErrInvalidURL
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
					CreateShortCodeFunc: func(ctx context.Context, originalURL string) (string, error) {
						return "", errors.New("db connection lost")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setupMock()
			handler := api.NewURLHandler(svc, logger, "http://closelinkit.test")

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
				if resp.ShortURL != "http://closelinkit.test/abcdef" {
					t.Errorf("expected ShortURL to be 'http://closelinkit.test/abcdef', got: %s", resp.ShortURL)
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
			handler := api.NewURLHandler(svc, logger, "http://closelinkit.test")

			req := httptest.NewRequest(http.MethodGet, "/api/v1/"+tt.shortCode, nil)
			req.SetPathValue("shortCode", tt.shortCode)
			w := httptest.NewRecorder()

			handler.HandlerResolveShortURL(w, req)

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
							ClickCount: 10,
							CreatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
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
			handler := api.NewURLHandler(svc, logger, "http://closelinkit.test")

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
				if resp.ClickCount != 10 {
					t.Errorf("expected response to have populated fields correctly, got: %+v", resp)
				}
			}
		})
	}
}

func TestURLHandler_HandlerGetUserURLs(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	testUserID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name           string
		authenticated  bool
		setupMock      func() *mockURLService
		expectedStatus int
		verifyBody     func(t *testing.T, body []byte)
	}{
		{
			name:          "Successful retrieval with user URLs",
			authenticated: true,
			setupMock: func() *mockURLService {
				return &mockURLService{
					GetURLsByUserIDFunc: func(ctx context.Context, userID uuid.UUID) ([]service.UserURL, error) {
						if userID != testUserID {
							t.Fatalf("expected userID %v, got %v", testUserID, userID)
						}
						return []service.UserURL{
							{
								OriginalURL: "https://example.com/one",
								ShortCode:   "code111",
								CreatedAt:   now,
								ClickCount:  5,
							},
							{
								OriginalURL: "https://example.com/two",
								ShortCode:   "code222",
								CreatedAt:   now.Add(-time.Hour),
								ClickCount:  0,
							},
						}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
			verifyBody: func(t *testing.T, body []byte) {
				var resp []api.UserURLResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("could not unmarshal response: %v", err)
				}
				if len(resp) != 2 {
					t.Fatalf("expected 2 items, got %d", len(resp))
				}
				if resp[0].ShortURL != "http://closelinkit.test/code111" {
					t.Errorf("expected short_url http://closelinkit.test/code111, got %s", resp[0].ShortURL)
				}
				if resp[0].OriginalURL != "https://example.com/one" || resp[0].ShortCode != "code111" || resp[0].ClickCount != 5 {
					t.Errorf("unexpected first item: %+v", resp[0])
				}
				if resp[1].ShortURL != "http://closelinkit.test/code222" {
					t.Errorf("expected short_url http://closelinkit.test/code222, got %s", resp[1].ShortURL)
				}
			},
		},
		{
			name:          "User has no URLs returns empty array",
			authenticated: true,
			setupMock: func() *mockURLService {
				return &mockURLService{
					GetURLsByUserIDFunc: func(ctx context.Context, userID uuid.UUID) ([]service.UserURL, error) {
						return []service.UserURL{}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
			verifyBody: func(t *testing.T, body []byte) {
				var resp []api.UserURLResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("could not unmarshal response: %v", err)
				}
				if resp == nil {
					t.Fatal("expected empty array, got nil")
				}
				if len(resp) != 0 {
					t.Fatalf("expected 0 items, got %d", len(resp))
				}
				trimmed := bytes.TrimSpace(body)
				if string(trimmed) != "[]" {
					t.Errorf("expected JSON [], got %s", string(trimmed))
				}
			},
		},
		{
			name:          "Unauthenticated request returns 401",
			authenticated: false,
			setupMock: func() *mockURLService {
				return &mockURLService{}
			},
			expectedStatus: http.StatusUnauthorized,
			verifyBody: func(t *testing.T, body []byte) {
				var errResp map[string]string
				if err := json.Unmarshal(body, &errResp); err != nil {
					t.Fatalf("could not unmarshal error response: %v", err)
				}
				if errResp["error"] != "Unauthorized" {
					t.Errorf("expected error 'Unauthorized', got %v", errResp["error"])
				}
			},
		},
		{
			name:          "Service error returns 500",
			authenticated: true,
			setupMock: func() *mockURLService {
				return &mockURLService{
					GetURLsByUserIDFunc: func(ctx context.Context, userID uuid.UUID) ([]service.UserURL, error) {
						return nil, errors.New("db error")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
			verifyBody:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setupMock()
			handler := api.NewURLHandler(svc, logger, "http://closelinkit.test")

			req := httptest.NewRequest(http.MethodGet, "/api/v1/urls", nil)
			if tt.authenticated {
				ctx := service.ContextWithUserID(req.Context(), testUserID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			handler.HandlerGetUserURLs(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.verifyBody != nil {
				tt.verifyBody(t, w.Body.Bytes())
			}
		})
	}
}
