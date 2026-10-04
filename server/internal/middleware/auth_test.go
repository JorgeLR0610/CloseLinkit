package middleware_test

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JorgeLR0610/CloseLinkit/internal/middleware"
	"github.com/JorgeLR0610/CloseLinkit/internal/security"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type mockTokenValidator struct {
	ValidateAccessTokenFunc func(tokenString string) (*security.Claims, error)
}

func (m *mockTokenValidator) ValidateAccessToken(tokenString string) (*security.Claims, error) {
	if m.ValidateAccessTokenFunc != nil {
		return m.ValidateAccessTokenFunc(tokenString)
	}
	return nil, errors.New("not implemented")
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRequireAuth(t *testing.T) {
	expectedUserID := uuid.New()
	validToken := "valid.jwt.token"
	malformedSubjectToken := "valid.signature.invalid.uuid"

	validator := &mockTokenValidator{
		ValidateAccessTokenFunc: func(tokenString string) (*security.Claims, error) {
			switch tokenString {
			case validToken:
				return &security.Claims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: expectedUserID.String(),
					},
				}, nil
			case malformedSubjectToken:
				return &security.Claims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: "not-a-valid-uuid",
					},
				}, nil
			default:
				return nil, errors.New("invalid token")
			}
		},
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectNext     bool
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
		{
			name:           "malformed authorization header without bearer",
			authHeader:     "Basic 12345",
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
		{
			name:           "empty bearer token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectNext:     true,
		},
		{
			name:           "valid token with lowercase bearer prefix",
			authHeader:     "bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectNext:     true,
		},
		{
			name:           "token with malformed subject UUID",
			authHeader:     "Bearer " + malformedSubjectToken,
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			var capturedUserID uuid.UUID
			var hasUser bool

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				capturedUserID, hasUser = middleware.GetUserID(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware.RequireAuth(validator, testLogger())(next)

			req := httptest.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
			if nextCalled != tt.expectNext {
				t.Errorf("expected next handler called=%v, got %v", tt.expectNext, nextCalled)
			}
			if tt.expectNext {
				if !hasUser {
					t.Error("expected user to be present in context")
				}
				if capturedUserID != expectedUserID {
					t.Errorf("expected user ID %v, got %v", expectedUserID, capturedUserID)
				}
			}
		})
	}
}

func TestOptionalAuth(t *testing.T) {
	expectedUserID := uuid.New()
	validToken := "valid.jwt.token"

	validator := &mockTokenValidator{
		ValidateAccessTokenFunc: func(tokenString string) (*security.Claims, error) {
			if tokenString == validToken {
				return &security.Claims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: expectedUserID.String(),
					},
				}, nil
			}
			return nil, errors.New("invalid token")
		},
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectNext     bool
		expectUser     bool
	}{
		{
			name:           "anonymous request without header",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			expectNext:     true,
			expectUser:     false,
		},
		{
			name:           "invalid header format with non-bearer scheme",
			authHeader:     "Basic 12345",
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
			expectUser:     false,
		},
		{
			name:           "invalid or expired token",
			authHeader:     "Bearer expired.or.invalid",
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
			expectUser:     false,
		},
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectNext:     true,
			expectUser:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			var capturedUserID uuid.UUID
			var hasUser bool

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				capturedUserID, hasUser = middleware.GetUserID(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware.OptionalAuth(validator, testLogger())(next)

			req := httptest.NewRequest("POST", "/shorten", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
			if nextCalled != tt.expectNext {
				t.Errorf("expected next handler called=%v, got %v", tt.expectNext, nextCalled)
			}
			if tt.expectUser {
				if !hasUser {
					t.Error("expected user to be present in context")
				}
				if capturedUserID != expectedUserID {
					t.Errorf("expected user ID %v, got %v", expectedUserID, capturedUserID)
				}
			} else if tt.expectNext {
				if hasUser {
					t.Error("expected user not to be present in context for anonymous call")
				}
			}
		})
	}
}
