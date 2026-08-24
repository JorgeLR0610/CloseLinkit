package middleware

import (
	"log/slog"
	"net/http"

	"github.com/JorgeLR0610/CloseLinkit/internal/api/v1"
	"github.com/JorgeLR0610/CloseLinkit/internal/response"
)

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if err := response.WriteError(w, http.StatusInternalServerError, api.InternalErrorMsg); err != nil {
						logger.Error(
							"could not write error response",
							slog.Any("error", err),
						)
					}
					logger.Error(
						"panic recovered",
						slog.Any("error", err),
						slog.String("path", r.URL.Path),
					)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
