package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/JorgeLR0610/CloseLinkit/internal/response"
	"github.com/JorgeLR0610/CloseLinkit/internal/service"
)

type URLHandler struct {
	service *service.URLService
}

func NewURLHandler(svc *service.URLService) *URLHandler {
	return &URLHandler{
		service: svc,
	}
}

func (h *URLHandler) HandlerCreateURL(w http.ResponseWriter, r *http.Request) {

	type urlCreationParams struct {
		OriginalURL	string `json:"url"`
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
		log.Printf("There was an error creating a URL: %v", err)
		return
	}

	if err := response.WriteJSON(w, http.StatusCreated, CreateURLResponse{
		ID: newURL.ID.String(),
		OriginalURL: newURL.OriginalUrl,
		ShortCode: newURL.ShortCode,
		CreatedAt: newURL.CreatedAt.Time,
	}); err != nil {
		log.Printf("Could not send response creation: %v", err)
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
		log.Printf("There was an error retrieving a URL: %v", err)
		return
	}

	http.Redirect(w, r, retrievedURL, http.StatusFound)

}