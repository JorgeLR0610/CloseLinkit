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
	"github.com/JorgeLR0610/CloseLinkit/internal/service"
	"github.com/google/uuid"
)

type mockAuthService struct {
	RegisterFunc     func(ctx context.Context, email, password string) (*service.UserResponse, error)
	LoginFunc        func(ctx context.Context, email, password string) (*service.TokenPair, *service.UserResponse, error)
	RefreshTokenFunc func(ctx context.Context, rawRefreshToken string) (*service.TokenPair, *service.UserResponse, error)
	LogoutFunc       func(ctx context.Context, rawRefreshToken string) error
}

func (m *mockAuthService) Register(ctx context.Context, email, password string) (*service.UserResponse, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, email, password)
	}
	return nil, nil
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*service.TokenPair, *service.UserResponse, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, email, password)
	}
	return nil, nil, nil
}

func (m *mockAuthService) RefreshToken(ctx context.Context, rawRefreshToken string) (*service.TokenPair, *service.UserResponse, error) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(ctx, rawRefreshToken)
	}
	return nil, nil, nil
}

func (m *mockAuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	if m.LogoutFunc != nil {
		return m.LogoutFunc(ctx, rawRefreshToken)
	}
	return nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestAuthHandler_HandlerRegister(t *testing.T) {
	testUUID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		body           string
		setupService   func() *mockAuthService
		expectedStatus int
	}{
		{
			name: "successful registration",
			body: `{"email":"user@example.com","password":"password123"}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{
					RegisterFunc: func(ctx context.Context, email, password string) (*service.UserResponse, error) {
						return &service.UserResponse{
							ID:        testUUID,
							Email:     email,
							CreatedAt: now,
							UpdatedAt: now,
						}, nil
					},
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid json format",
			body: `{"email":`,
			setupService: func() *mockAuthService {
				return &mockAuthService{}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "unknown fields in json",
			body: `{"email":"user@example.com","password":"password123","unknown":"field"}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email format",
			body: `{"email":"not-an-email","password":"password123"}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{
					RegisterFunc: func(ctx context.Context, email, password string) (*service.UserResponse, error) {
						return nil, service.ErrInvalidEmail
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "password too short",
			body: `{"email":"user@example.com","password":"123"}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{
					RegisterFunc: func(ctx context.Context, email, password string) (*service.UserResponse, error) {
						return nil, service.ErrPasswordTooShort
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "user already exists",
			body: `{"email":"user@example.com","password":"password123"}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{
					RegisterFunc: func(ctx context.Context, email, password string) (*service.UserResponse, error) {
						return nil, service.ErrUserAlreadyExists
					},
				}
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "internal server error",
			body: `{"email":"user@example.com","password":"password123"}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{
					RegisterFunc: func(ctx context.Context, email, password string) (*service.UserResponse, error) {
						return nil, errors.New("unexpected database error")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := api.NewAuthHandler(tt.setupService(), testLogger())
			req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			handler.HandlerRegister(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestAuthHandler_HandlerLogin(t *testing.T) {
	testUUID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		body           string
		setupService   func() *mockAuthService
		expectedStatus int
	}{
		{
			name: "successful login",
			body: `{"email":"user@example.com","password":"password123"}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{
					LoginFunc: func(ctx context.Context, email, password string) (*service.TokenPair, *service.UserResponse, error) {
						tokens := &service.TokenPair{
							AccessToken:  "access.token",
							RefreshToken: "refresh.token",
							ExpiresIn:    900,
						}
						user := &service.UserResponse{
							ID:        testUUID,
							Email:     email,
							CreatedAt: now,
						}
						return tokens, user, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid json format",
			body: `{"email":`,
			setupService: func() *mockAuthService {
				return &mockAuthService{}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid credentials",
			body: `{"email":"user@example.com","password":"wrongpassword"}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{
					LoginFunc: func(ctx context.Context, email, password string) (*service.TokenPair, *service.UserResponse, error) {
						return nil, nil, service.ErrInvalidCredentials
					},
				}
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "internal server error",
			body: `{"email":"user@example.com","password":"password123"}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{
					LoginFunc: func(ctx context.Context, email, password string) (*service.TokenPair, *service.UserResponse, error) {
						return nil, nil, errors.New("db error")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := api.NewAuthHandler(tt.setupService(), testLogger())
			req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			handler.HandlerLogin(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var res api.LoginResponse
				if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
					t.Fatalf("could not decode response: %v", err)
				}
				if res.AccessToken != "access.token" || res.RefreshToken != "refresh.token" {
					t.Errorf("unexpected tokens: %+v", res)
				}
				if res.User.Email != "user@example.com" {
					t.Errorf("unexpected user email: %s", res.User.Email)
				}

				// Verify HttpOnly cookie was set
				cookies := rec.Result().Cookies()
				var refreshCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == api.RefreshTokenCookieName {
						refreshCookie = c
						break
					}
				}
				if refreshCookie == nil {
					t.Fatal("expected refresh_token cookie to be set")
				}
				if refreshCookie.Value != "refresh.token" {
					t.Errorf("expected cookie value refresh.token, got %s", refreshCookie.Value)
				}
				if !refreshCookie.HttpOnly {
					t.Error("expected cookie to be HttpOnly")
				}
				if refreshCookie.Path != api.RefreshTokenCookiePath {
					t.Errorf("expected cookie path %s, got %s", api.RefreshTokenCookiePath, refreshCookie.Path)
				}
				if refreshCookie.SameSite != http.SameSiteLaxMode {
					t.Errorf("expected SameSite Lax, got %v", refreshCookie.SameSite)
				}
			}
		})
	}
}

func TestAuthHandler_HandlerRefreshToken(t *testing.T) {
	testUUID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		body           string
		cookieToken    string
		setupService   func() *mockAuthService
		expectedStatus int
	}{
		{
			name: "successful refresh via body",
			body: `{"refresh_token":"valid-refresh-token"}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{
					RefreshTokenFunc: func(ctx context.Context, rawRefreshToken string) (*service.TokenPair, *service.UserResponse, error) {
						tokens := &service.TokenPair{
							AccessToken:  "new.access.token",
							RefreshToken: "new.refresh.token",
							ExpiresIn:    900,
						}
						user := &service.UserResponse{
							ID:        testUUID,
							Email:     "user@example.com",
							CreatedAt: now,
							UpdatedAt: now,
						}
						return tokens, user, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful refresh via cookie",
			body:        "",
			cookieToken: "valid-cookie-token",
			setupService: func() *mockAuthService {
				return &mockAuthService{
					RefreshTokenFunc: func(ctx context.Context, rawRefreshToken string) (*service.TokenPair, *service.UserResponse, error) {
						if rawRefreshToken != "valid-cookie-token" {
							return nil, nil, service.ErrInvalidRefreshToken
						}
						tokens := &service.TokenPair{
							AccessToken:  "cookie.access.token",
							RefreshToken: "cookie.refresh.token",
							ExpiresIn:    900,
						}
						user := &service.UserResponse{
							ID:        testUUID,
							Email:     "user@example.com",
							CreatedAt: now,
							UpdatedAt: now,
						}
						return tokens, user, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid json format when no cookie",
			body: `{"refresh_token":`,
			setupService: func() *mockAuthService {
				return &mockAuthService{}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid or revoked refresh token",
			body: `{"refresh_token":"revoked-token"}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{
					RefreshTokenFunc: func(ctx context.Context, rawRefreshToken string) (*service.TokenPair, *service.UserResponse, error) {
						return nil, nil, service.ErrInvalidRefreshToken
					},
				}
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "expired refresh token",
			body: `{"refresh_token":"expired-token"}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{
					RefreshTokenFunc: func(ctx context.Context, rawRefreshToken string) (*service.TokenPair, *service.UserResponse, error) {
						return nil, nil, service.ErrExpiredRefreshToken
					},
				}
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "internal server error",
			body: `{"refresh_token":"valid-token"}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{
					RefreshTokenFunc: func(ctx context.Context, rawRefreshToken string) (*service.TokenPair, *service.UserResponse, error) {
						return nil, nil, errors.New("db error")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := api.NewAuthHandler(tt.setupService(), testLogger())
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBufferString(tt.body))
			} else {
				req = httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
			}
			if tt.cookieToken != "" {
				req.AddCookie(&http.Cookie{
					Name:  api.RefreshTokenCookieName,
					Value: tt.cookieToken,
				})
			}
			rec := httptest.NewRecorder()

			handler.HandlerRefreshToken(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var res api.RefreshTokenResponse
				if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
					t.Fatalf("could not decode response: %v", err)
				}
				if res.User.Email != "user@example.com" {
					t.Errorf("expected user email user@example.com, got %s", res.User.Email)
				}

				// Check rotated cookie
				cookies := rec.Result().Cookies()
				var rotatedCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == api.RefreshTokenCookieName {
						rotatedCookie = c
						break
					}
				}
				if rotatedCookie == nil {
					t.Fatal("expected rotated refresh_token cookie")
				}
				if !rotatedCookie.HttpOnly {
					t.Error("expected rotated cookie to be HttpOnly")
				}
			}
		})
	}
}

func TestAuthHandler_HandlerLogout(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		cookieToken    string
		setupService   func() *mockAuthService
		expectedStatus int
	}{
		{
			name: "successful logout via body",
			body: `{"refresh_token":"valid-refresh-token"}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{
					LogoutFunc: func(ctx context.Context, rawRefreshToken string) error {
						return nil
					},
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:        "successful logout via cookie",
			body:        "",
			cookieToken: "valid-cookie-token",
			setupService: func() *mockAuthService {
				return &mockAuthService{
					LogoutFunc: func(ctx context.Context, rawRefreshToken string) error {
						if rawRefreshToken != "valid-cookie-token" {
							return service.ErrInvalidRefreshToken
						}
						return nil
					},
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "invalid json format and no cookie",
			body: `{"refresh_token":`,
			setupService: func() *mockAuthService {
				return &mockAuthService{}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty refresh token error",
			body: `{"refresh_token":""}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{
					LogoutFunc: func(ctx context.Context, rawRefreshToken string) error {
						return service.ErrInvalidRefreshToken
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "internal server error",
			body: `{"refresh_token":"valid-refresh-token"}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{
					LogoutFunc: func(ctx context.Context, rawRefreshToken string) error {
						return errors.New("db error")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := api.NewAuthHandler(tt.setupService(), testLogger())
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewBufferString(tt.body))
			} else {
				req = httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
			}
			if tt.cookieToken != "" {
				req.AddCookie(&http.Cookie{
					Name:  api.RefreshTokenCookieName,
					Value: tt.cookieToken,
				})
			}
			rec := httptest.NewRecorder()

			handler.HandlerLogout(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusNoContent {
				cookies := rec.Result().Cookies()
				var clearedCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == api.RefreshTokenCookieName {
						clearedCookie = c
						break
					}
				}
				if clearedCookie == nil {
					t.Fatal("expected cleared refresh_token cookie")
				}
				if clearedCookie.MaxAge >= 0 {
					t.Errorf("expected cookie MaxAge < 0, got %d", clearedCookie.MaxAge)
				}
			}
		})
	}
}
