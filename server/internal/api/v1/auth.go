package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/JorgeLR0610/CloseLinkit/internal/response"
	"github.com/JorgeLR0610/CloseLinkit/internal/service"
)

const (
	RefreshTokenCookieName = "refresh_token"
	RefreshTokenCookiePath = "/api/v1/auth" // #nosec G101 -- false positive
	DefaultCookieMaxAge    = 7 * 24 * 3600  // 7 days in seconds
)

type AuthServicer interface {
	Register(ctx context.Context, email, password string) (*service.UserResponse, error)
	Login(ctx context.Context, email, password string) (*service.TokenPair, *service.UserResponse, error)
	RefreshToken(ctx context.Context, rawRefreshToken string) (*service.TokenPair, *service.UserResponse, error)
	Logout(ctx context.Context, rawRefreshToken string) error
}

type AuthHandler struct {
	service      AuthServicer
	logger       *slog.Logger
	secureCookie bool
}

func NewAuthHandler(svc AuthServicer, logger *slog.Logger) *AuthHandler {
	secure := os.Getenv("COOKIE_SECURE") != "false"
	return &AuthHandler{
		service: svc,
		logger: logger.With(
			slog.String("component", "auth_handler"),
		),
		secureCookie: secure,
	}
}

func (h *AuthHandler) SetSecureCookie(secure bool) {
	h.secureCookie = secure
}

func (h *AuthHandler) setRefreshTokenCookie(w http.ResponseWriter, token string) {
	maxAge := DefaultCookieMaxAge
	if ttlGetter, ok := h.service.(interface{ RefreshTokenTTL() time.Duration }); ok {
		maxAge = int(ttlGetter.RefreshTokenTTL().Seconds())
	}

	http.SetCookie(w, &http.Cookie{ // #nosec G124 -- Mitigated: Cookie set fields are secure
		Name:     RefreshTokenCookieName,
		Value:    token,
		Path:     RefreshTokenCookiePath,
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	})
}

func (h *AuthHandler) clearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{ // #nosec G124 -- Mitigated: Cookie set fields are secure
		Name:     RefreshTokenCookieName,
		Value:    "",
		Path:     RefreshTokenCookiePath,
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func (h *AuthHandler) writeErrorLogged(w http.ResponseWriter, code int, msg string) {
	if err := response.WriteError(w, code, msg); err != nil {
		h.logger.Error("could not write error response", slog.Any("error", err))
	}
}

func (h *AuthHandler) HandlerRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.writeErrorLogged(w, http.StatusBadRequest, "The provided request body is invalid or malformed")
		return
	}

	user, err := h.service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidEmail) ||
			errors.Is(err, service.ErrPasswordTooShort) ||
			errors.Is(err, service.ErrPasswordTooLong) {
			h.writeErrorLogged(w, http.StatusBadRequest, err.Error())
			return
		}

		if errors.Is(err, service.ErrUserAlreadyExists) {
			h.writeErrorLogged(w, http.StatusConflict, err.Error())
			return
		}

		h.writeErrorLogged(w, http.StatusInternalServerError, InternalErrorMsg)
		h.logger.Error(
			"could not register user",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
		return
	}

	if err := response.WriteJSON(w, http.StatusCreated, user); err != nil {
		h.logger.Error(
			"could not send register response",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
	}
}

func (h *AuthHandler) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.writeErrorLogged(w, http.StatusBadRequest, "The provided request body is invalid or malformed")
		return
	}

	tokens, user, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			h.writeErrorLogged(w, http.StatusUnauthorized, err.Error())
			return
		}

		h.writeErrorLogged(w, http.StatusInternalServerError, InternalErrorMsg)
		h.logger.Error(
			"could not login user",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
		return
	}

	h.setRefreshTokenCookie(w, tokens.RefreshToken)

	if err := response.WriteJSON(w, http.StatusOK, LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		User:         *user,
	}); err != nil {
		h.logger.Error(
			"could not send login response",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
	}
}

func (h *AuthHandler) extractRefreshToken(r *http.Request) (string, error) {
	if cookie, err := r.Cookie(RefreshTokenCookieName); err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	if r.Body != nil {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err == nil {
			if req.RefreshToken != "" {
				return req.RefreshToken, nil
			}
		} else if !errors.Is(err, io.EOF) {
			return "", err
		}
	}

	return "", errors.New("missing refresh token")
}

func (h *AuthHandler) HandlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	rawRefreshToken, err := h.extractRefreshToken(r)
	if err != nil || rawRefreshToken == "" {
		h.writeErrorLogged(w, http.StatusBadRequest, "The provided request body is invalid or malformed")
		return
	}

	tokens, user, err := h.service.RefreshToken(r.Context(), rawRefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) ||
			errors.Is(err, service.ErrExpiredRefreshToken) ||
			errors.Is(err, service.ErrUserNotFound) {
			h.writeErrorLogged(w, http.StatusUnauthorized, err.Error())
			return
		}

		h.writeErrorLogged(w, http.StatusInternalServerError, InternalErrorMsg)
		h.logger.Error(
			"could not refresh token",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
		return
	}

	h.setRefreshTokenCookie(w, tokens.RefreshToken)

	if err := response.WriteJSON(w, http.StatusOK, RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		User:         *user,
	}); err != nil {
		h.logger.Error(
			"could not send refresh token response",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
	}
}

func (h *AuthHandler) HandlerLogout(w http.ResponseWriter, r *http.Request) {
	rawRefreshToken, err := h.extractRefreshToken(r)
	if err != nil || rawRefreshToken == "" {
		h.writeErrorLogged(w, http.StatusBadRequest, "The provided request body is invalid or malformed")
		return
	}

	h.clearRefreshTokenCookie(w)

	if err := h.service.Logout(r.Context(), rawRefreshToken); err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			h.writeErrorLogged(w, http.StatusBadRequest, err.Error())
			return
		}

		h.writeErrorLogged(w, http.StatusInternalServerError, InternalErrorMsg)
		h.logger.Error(
			"could not logout",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
		return
	}

	response.WriteNoContent(w)
}
