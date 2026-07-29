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
)

const internalErrorMsg = "There was an error on our end. Please try again later"

type URLServicer interface {
	CreateShortCode(ctx context.Context, originalURL string) (repository.CreateURLRow, error)
	ResolveShortCode(ctx context.Context, shortCode string) (string, error)
	GetURLStats(ctx context.Context, shortCode string) (repository.GetURLStatsRow, error)
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

func (h *URLHandler) HandlerCreateURL(w http.ResponseWriter, r *http.Request) {

	type urlCreationParams struct {
		OriginalURL string `json:"url"`
	}

	var urlParams urlCreationParams
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&urlParams); err != nil {
		response.WriteError(w, http.StatusBadRequest, "The provided URL is invalid or malformed")
		return
	}

	newURL, err := h.service.CreateShortCode(r.Context(), urlParams.OriginalURL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURLScheme) || errors.Is(err, service.ErrNoHost) || errors.Is(err, service.ErrInvalidURL) {
			response.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		response.WriteError(w, http.StatusInternalServerError, internalErrorMsg)
		h.logger.Error(
			"could not create URL",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
		return
	}

	if err := response.WriteJSON(w, http.StatusCreated, CreateURLResponse{
		ShortURL: h.baseURL + "/" + newURL.ShortCode,
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
			if err := web.ServeNotFoundPage(w); err != nil {
				h.logger.Error("could not write 404 page",
					slog.Any("error", err),
				)
			}
			return
		}

		response.WriteError(w, http.StatusInternalServerError, internalErrorMsg)
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

		response.WriteError(w, http.StatusInternalServerError, internalErrorMsg)
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
