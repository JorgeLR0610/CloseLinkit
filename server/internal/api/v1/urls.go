package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/JorgeLR0610/CloseLinkit/internal/repository"
	"github.com/JorgeLR0610/CloseLinkit/internal/response"
	"github.com/JorgeLR0610/CloseLinkit/internal/service"
	"github.com/JorgeLR0610/CloseLinkit/web"
	"github.com/google/uuid"
)

const InternalErrorMsg = "There was an error on our end. Please try again later"

type URLServicer interface {
	CreateShortCode(ctx context.Context, originalURL string) (string, error)
	ResolveShortCode(ctx context.Context, shortCode string) (string, error)
	GetURLStats(ctx context.Context, shortCode string) (repository.GetURLStatsRow, error)
	GetURLsByUserID(ctx context.Context, userID uuid.UUID) ([]service.UserURL, error)
}

type URLHandler struct {
	service URLServicer
	logger  *slog.Logger
	baseURL string
}

func NewURLHandler(svc URLServicer, logger *slog.Logger, baseURL string) *URLHandler {
	return &URLHandler{
		service: svc,
		logger: logger.With(
			slog.String("component", "url_handler"),
		),
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (h *URLHandler) writeErrorLogged(w http.ResponseWriter, code int, msg string) {
	if err := response.WriteError(w, code, msg); err != nil {
		h.logger.Error("could not write error response", slog.Any("error", err))
	}
}

func (h *URLHandler) HandlerCreateURL(w http.ResponseWriter, r *http.Request) {

	type urlCreationParams struct {
		OriginalURL string `json:"url"`
	}

	var urlParams urlCreationParams
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&urlParams); err != nil {
		h.writeErrorLogged(w, http.StatusBadRequest, "The provided URL is invalid or malformed")
		return
	}

	shortCode, err := h.service.CreateShortCode(r.Context(), urlParams.OriginalURL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURLScheme) || errors.Is(err, service.ErrNoHost) || errors.Is(err, service.ErrInvalidURL) {
			h.writeErrorLogged(w, http.StatusBadRequest, err.Error())
			return
		}

		h.writeErrorLogged(w, http.StatusInternalServerError, InternalErrorMsg)
		h.logger.Error(
			"could not create URL",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
		return
	}

	if err := response.WriteJSON(w, http.StatusCreated, CreateURLResponse{
		ShortURL: h.baseURL + "/" + shortCode,
	}); err != nil {
		h.logger.Error(
			"could not send shortURL creation response",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
		return
	}
}

func (h *URLHandler) HandlerResolveShortURL(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("shortCode")

	retrievedURL, err := h.service.ResolveShortCode(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, service.ErrNoURLFound) {
			// Serve a static 404 page during the initial release.
			// Future versions will redirect to the React application,
			// which will render the dedicated Not Found page.
			if err := web.ServeNotFoundPage(w); err != nil {
				h.logger.Error("could not write 404 page",
					slog.Any("error", err),
				)
			}
			return
		}

		h.writeErrorLogged(w, http.StatusInternalServerError, InternalErrorMsg)
		h.logger.Error(
			"could not retrieve URL",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
		return
	}

	// #nosec G710 -- Mitigated: URL is sanitized in both React Client and API Service Layer
	http.Redirect(w, r, retrievedURL, http.StatusFound)
}

func (h *URLHandler) HandlerGetURLStats(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("shortCode")

	stats, err := h.service.GetURLStats(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, service.ErrNoURLFound) {
			h.writeErrorLogged(w, http.StatusNotFound, "Sorry, we did not found the page you are looking for")
			return
		}

		h.writeErrorLogged(w, http.StatusInternalServerError, InternalErrorMsg)
		h.logger.Error(
			"could not retrieve URL",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
		return
	}

	if err := response.WriteJSON(w, http.StatusOK, GetURLStatsResponse{
		ClickCount: int(stats.ClickCount),
		CreatedAt:  stats.CreatedAt.Time,
	}); err != nil {
		h.logger.Error(
			"could not send stats response",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
	}
}

func (h *URLHandler) HandlerGetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := service.UserIDFromContext(r.Context())
	if !ok {
		h.writeErrorLogged(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	urls, err := h.service.GetURLsByUserID(r.Context(), userID)
	if err != nil {
		h.writeErrorLogged(w, http.StatusInternalServerError, InternalErrorMsg)
		h.logger.Error(
			"could not retrieve user URLs",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("user_id", userID.String()),
			slog.Any("error", err),
		)
		return
	}

	responseURLs := make([]UserURLResponse, 0, len(urls))
	for _, u := range urls {
		responseURLs = append(responseURLs, UserURLResponse{
			OriginalURL: u.OriginalURL,
			ShortCode:   u.ShortCode,
			ShortURL:    h.baseURL + "/" + u.ShortCode,
			CreatedAt:   u.CreatedAt,
			ClickCount:  u.ClickCount,
		})
	}

	if err := response.WriteJSON(w, http.StatusOK, responseURLs); err != nil {
		h.logger.Error(
			"could not send user URLs response",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
	}
}
