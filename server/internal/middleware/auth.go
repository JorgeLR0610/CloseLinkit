package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/JorgeLR0610/CloseLinkit/internal/response"
	"github.com/JorgeLR0610/CloseLinkit/internal/security"
	"github.com/JorgeLR0610/CloseLinkit/internal/service"
	"github.com/google/uuid"
)

type TokenValidator interface {
	ValidateAccessToken(tokenString string) (*security.Claims, error)
}

func writeAuthError(w http.ResponseWriter, logger *slog.Logger, msg string) {
	if err := response.WriteError(w, http.StatusUnauthorized, msg); err != nil && logger != nil {
		logger.Error("could not write auth error response", slog.Any("error", err))
	}
}

func extractBearerToken(authHeader string) (string, bool) {
	if authHeader == "" {
		return "", false
	}

	parts := strings.SplitN(strings.TrimSpace(authHeader), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}

	return token, true
}

// No auth required endpoints have been added yet
func RequireAuth(validator TokenValidator, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			tokenString, ok := extractBearerToken(authHeader)
			if !ok {
				writeAuthError(w, logger, "Missing or invalid authorization header")
				return
			}

			claims, err := validator.ValidateAccessToken(tokenString)
			if err != nil {
				writeAuthError(w, logger, "Invalid or expired token")
				return
			}

			userID, err := claims.UserID()
			if err != nil {
				writeAuthError(w, logger, "Invalid user identifier in token")
				return
			}

			ctx := service.ContextWithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuth(validator TokenValidator, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			tokenString, ok := extractBearerToken(authHeader)
			if !ok {
				writeAuthError(w, logger, "Invalid authorization header format")
				return
			}

			claims, err := validator.ValidateAccessToken(tokenString)
			if err != nil {
				writeAuthError(w, logger, "Invalid or expired token")
				return
			}

			userID, err := claims.UserID()
			if err != nil {
				writeAuthError(w, logger, "Invalid user identifier in token")
				return
			}

			ctx := service.ContextWithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	return service.UserIDFromContext(ctx)
}
