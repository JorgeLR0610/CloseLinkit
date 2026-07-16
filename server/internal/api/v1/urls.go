package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/JorgeLR0610/CloseLinkit/internal/repository"
	"github.com/JorgeLR0610/CloseLinkit/internal/response"
	"github.com/JorgeLR0610/CloseLinkit/internal/service"
)

type URLServicer interface {
	CreateShortCode(ctx context.Context, originalURL string) (repository.CreateURLRow, error)
	ResolveShortCode(ctx context.Context, shortCode string) (string, error)
	GetURLStats(ctx context.Context, shortCode string) (repository.GetURLStatsRow, error)
}

type URLHandler struct {
	service URLServicer
	logger  *slog.Logger
}

func NewURLHandler(svc URLServicer, logger *slog.Logger) *URLHandler {
	return &URLHandler{
		service: svc,
		logger: logger.With(
			slog.String("component", "url_handler"),
		),
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
		response.WriteError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	newURL, err := h.service.CreateShortCode(r.Context(), urlParams.OriginalURL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURLScheme) || errors.Is(err, service.ErrNoHost) || errors.Is(err, service.ErrInvalidURL) {
			response.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		response.WriteError(w, http.StatusInternalServerError, "There was an error on our end")
		h.logger.Error(
			"could not create URL",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
		return
	}

	if err := response.WriteJSON(w, http.StatusCreated, CreateURLResponse{
		ID:          newURL.ID.String(),
		OriginalURL: newURL.OriginalUrl,
		ShortCode:   newURL.ShortCode,
		CreatedAt:   newURL.CreatedAt.Time,
	}); err != nil {
		h.logger.Error(
			"could not send shortCode creation response",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
		return
	}
}

func (h *URLHandler) HandlerGetURL(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("shortCode")

	retrievedURL, err := h.service.ResolveShortCode(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, service.ErrNoURLFound) {
			response.WriteError(w, http.StatusNotFound, "Sorry, we did not found the page you are looking for")
			return
		}

		response.WriteError(w, http.StatusInternalServerError, "There was an error on our end")
		h.logger.Error(
			"could not retrieve URL",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
		return
	}

	http.Redirect(w, r, retrievedURL, http.StatusFound)
}

func (h *URLHandler) HandlerGetURLStats(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("shortCode")

	stats, err := h.service.GetURLStats(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, service.ErrNoURLFound) {
			response.WriteError(w, http.StatusNotFound, "Sorry, we did not found the page you are looking for")
			return
		}

		response.WriteError(w, http.StatusInternalServerError, "There was an error on our end")
		h.logger.Error(
			"could not retrieve URL",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
		return
	}

	if err := response.WriteJSON(w, http.StatusOK, GetURLStatsResponse{
		OriginalURL: stats.OriginalUrl,
		ClickCount:  int(stats.ClickCount),
		CreatedAt:   stats.CreatedAt.Time,
	}); err != nil {
		h.logger.Error(
			"could not send stats response",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
	}
}
