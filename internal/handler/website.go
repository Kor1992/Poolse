package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	service "pulse/internal"
	"pulse/internal/domain"
	"pulse/internal/repository"
	"strconv"
	"time"
)

type WebsiteService interface {
	CreateWebsite(ctx context.Context, url string) (int64, error)
	GetActiveWebsites(ctx context.Context) ([]domain.Website, error)
	GetWebsiteChecks(ctx context.Context, websiteID int64, limit int) ([]domain.CheckResult, error)
	GetWebsiteByID(ctx context.Context, id int64) (domain.Website, error)
	UpdateWebsiteActive(ctx context.Context, id int64, isActive bool) error
	DeleteWebsite(ctx context.Context, id int64) error
}

type UpdateWebsiteRequest struct {
	IsActive *bool `json:"is_active"`
}

type WebsiteHandler struct {
	service WebsiteService
}

func NewWebHandler(servise WebsiteService) *WebsiteHandler {
	return &WebsiteHandler{service: servise}
}

func (h *WebsiteHandler) CreateWebsite(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type CreateWebsiteRequest struct {
		URL string `json:"url"`
	}
	var req CreateWebsiteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON decode error", http.StatusBadRequest)
		return
	}

	id, err := h.service.CreateWebsite(r.Context(), req.URL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			http.Error(w, "invalid URL", http.StatusBadRequest)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	type CreateWebsiteResponse struct {
		ID int64 `json:"id"`
	}

	response := CreateWebsiteResponse{
		ID: id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

}

func (h *WebsiteHandler) GetActiveWebsites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := h.service.GetActiveWebsites(r.Context())
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

type CheckResponse struct {
	WebsiteID  int64     `json:"website_id"`
	StatusCode int       `json:"status_code"`
	DurationMs int64     `json:"duration_ms"`
	IsUp       bool      `json:"is_up"`
	Error      string    `json:"error"`
	CheckedAt  time.Time `json:"checked_at"`
}

func (h *WebsiteHandler) GetWebsiteChecks(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	websiteId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "id parse error", http.StatusBadRequest)
		return
	}

	limit := 0
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
	}

	checks, err := h.service.GetWebsiteChecks(r.Context(), websiteId, limit)
	if err != nil {
		if errors.Is(err, service.ErrInvalidID) {
			http.Error(w, "invalid website id", http.StatusBadRequest)
			return
		}
		http.Error(w, "server", http.StatusInternalServerError)
		return
	}

	response := make([]CheckResponse, 0, len(checks))

	for _, v := range checks {
		check := CheckResponse{WebsiteID: v.WebsiteID, StatusCode: v.StatusCode, DurationMs: v.Duration.Milliseconds(), IsUp: v.IsUp, Error: v.Error, CheckedAt: v.CheckedAt}
		response = append(response, check)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func (h *WebsiteHandler) GetWebsiteByID(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "id parse error", http.StatusBadRequest)
		return
	}

	website, err := h.service.GetWebsiteByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "website not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrInvalidID) {
			http.Error(w, "invalid website id", http.StatusBadRequest)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(website)
}

func (h *WebsiteHandler) UpdateWebsiteActive(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "id parse error", http.StatusBadRequest)
		return
	}

	var request UpdateWebsiteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "JSON decode error", http.StatusBadRequest)
		return
	}
	if request.IsActive == nil {
		http.Error(w, "is_active is required", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateWebsiteActive(r.Context(), id, *request.IsActive); err != nil {
		if errors.Is(err, service.ErrInvalidID) {
			http.Error(w, "invalid website id", http.StatusBadRequest)
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "website not found", http.StatusNotFound)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WebsiteHandler) DeleteWebsite(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "id parse error", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteWebsite(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrInvalidID) {
			http.Error(w, "invalid website id", http.StatusBadRequest)
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "website not found", http.StatusNotFound)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
