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
	RefreshTokenFunc func(ctx context.Context, rawRefreshToken string) (*service.TokenPair, error)
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

func (m *mockAuthService) RefreshToken(ctx context.Context, rawRefreshToken string) (*service.TokenPair, error) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(ctx, rawRefreshToken)
	}
	return nil, nil
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
						return &service.TokenPair{
								AccessToken:  "access.token",
								RefreshToken: "refresh.token",
								ExpiresIn:    900,
							}, &service.UserResponse{
								ID:        testUUID,
								Email:     email,
								CreatedAt: now,
								UpdatedAt: now,
							}, nil
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
			}
		})
	}
}

func TestAuthHandler_HandlerRefreshToken(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupService   func() *mockAuthService
		expectedStatus int
	}{
		{
			name: "successful refresh",
			body: `{"refresh_token":"valid-refresh-token"}`,
			setupService: func() *mockAuthService {
				return &mockAuthService{
					RefreshTokenFunc: func(ctx context.Context, rawRefreshToken string) (*service.TokenPair, error) {
						return &service.TokenPair{
							AccessToken:  "new.access.token",
							RefreshToken: "new.refresh.token",
							ExpiresIn:    900,
						}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid json format",
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
					RefreshTokenFunc: func(ctx context.Context, rawRefreshToken string) (*service.TokenPair, error) {
						return nil, service.ErrInvalidRefreshToken
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
					RefreshTokenFunc: func(ctx context.Context, rawRefreshToken string) (*service.TokenPair, error) {
						return nil, service.ErrExpiredRefreshToken
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
					RefreshTokenFunc: func(ctx context.Context, rawRefreshToken string) (*service.TokenPair, error) {
						return nil, errors.New("db error")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := api.NewAuthHandler(tt.setupService(), testLogger())
			req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			handler.HandlerRefreshToken(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestAuthHandler_HandlerLogout(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupService   func() *mockAuthService
		expectedStatus int
	}{
		{
			name: "successful logout",
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
			name: "invalid json format",
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
			req := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			handler.HandlerLogout(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}
