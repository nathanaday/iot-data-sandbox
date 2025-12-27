package llm

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nathanaday/iot-data-sandbox/api/handlers"
	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/services"
)

type ProviderHandler struct {
	service *services.LLMProviderService
}

func NewProviderHandler(service *services.LLMProviderService) *ProviderHandler {
	return &ProviderHandler{
		service: service,
	}
}

// Request types

type CreateProviderRequest struct {
	ProviderType string  `json:"provider_type"`
	Name         string  `json:"name"`
	APIKey       string  `json:"api_key"`
	BaseURL      *string `json:"base_url,omitempty"`
	DefaultModel *string `json:"default_model,omitempty"`
}

type UpdateProviderRequest struct {
	Name         *string `json:"name,omitempty"`
	APIKey       *string `json:"api_key,omitempty"`
	BaseURL      *string `json:"base_url,omitempty"`
	DefaultModel *string `json:"default_model,omitempty"`
	IsEnabled    *bool   `json:"is_enabled,omitempty"`
}

// Response types

type ProviderResponse struct {
	ProviderID   int64   `json:"provider_id"`
	ProviderType string  `json:"provider_type"`
	Name         string  `json:"name"`
	BaseURL      *string `json:"base_url,omitempty"`
	DefaultModel *string `json:"default_model,omitempty"`
	IsEnabled    bool    `json:"is_enabled"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type ProviderListResponse struct {
	Providers []ProviderResponse `json:"providers"`
}

// Create godoc
// @Summary Create a new LLM provider
// @Description Create a new LLM provider configuration with API key
// @Tags llm-providers
// @Accept json
// @Produce json
// @Param provider body CreateProviderRequest true "Provider details"
// @Success 201 {object} ProviderResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/llm/providers [post]
func (h *ProviderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		handlers.RespondError(w, "Provider name is required", http.StatusBadRequest)
		return
	}

	if req.ProviderType == "" {
		handlers.RespondError(w, "Provider type is required", http.StatusBadRequest)
		return
	}

	// Validate provider type
	providerType := models.LLMProviderType(req.ProviderType)
	if !isValidProviderType(providerType) {
		handlers.RespondError(w, "Invalid provider type", http.StatusBadRequest)
		return
	}

	provider := &models.LLMProvider{
		ProviderType: providerType,
		Name:         req.Name,
		APIKey:       req.APIKey,
		BaseURL:      req.BaseURL,
		DefaultModel: req.DefaultModel,
		IsEnabled:    true,
	}

	if err := h.service.Create(provider); err != nil {
		handlers.RespondError(w, "Failed to create provider: "+err.Error(), http.StatusInternalServerError)
		return
	}

	handlers.RespondJSON(w, toProviderResponse(provider), http.StatusCreated)
}

// List godoc
// @Summary List all LLM providers
// @Description Get a list of all configured LLM providers
// @Tags llm-providers
// @Produce json
// @Success 200 {object} ProviderListResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/llm/providers [get]
func (h *ProviderHandler) List(w http.ResponseWriter, r *http.Request) {
	providers, err := h.service.LoadAll()
	if err != nil {
		handlers.RespondError(w, "Failed to load providers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]ProviderResponse, 0, len(providers))
	for _, p := range providers {
		responses = append(responses, toProviderResponse(p))
	}

	handlers.RespondJSON(w, ProviderListResponse{Providers: responses}, http.StatusOK)
}

// Get godoc
// @Summary Get an LLM provider
// @Description Get a specific LLM provider by ID
// @Tags llm-providers
// @Produce json
// @Param id path int true "Provider ID"
// @Success 200 {object} ProviderResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Router /api/llm/providers/{id} [get]
func (h *ProviderHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handlers.RespondError(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	provider, err := h.service.LoadByID(id)
	if err != nil {
		handlers.RespondError(w, "Provider not found", http.StatusNotFound)
		return
	}

	handlers.RespondJSON(w, toProviderResponse(provider), http.StatusOK)
}

// Update godoc
// @Summary Update an LLM provider
// @Description Update an existing LLM provider configuration
// @Tags llm-providers
// @Accept json
// @Produce json
// @Param id path int true "Provider ID"
// @Param provider body UpdateProviderRequest true "Provider updates"
// @Success 200 {object} ProviderResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/llm/providers/{id} [put]
func (h *ProviderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handlers.RespondError(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	provider, err := h.service.LoadByID(id)
	if err != nil {
		handlers.RespondError(w, "Provider not found", http.StatusNotFound)
		return
	}

	var req UpdateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Apply updates
	if req.Name != nil {
		provider.Name = *req.Name
	}
	if req.APIKey != nil {
		provider.APIKey = *req.APIKey
	}
	if req.BaseURL != nil {
		provider.BaseURL = req.BaseURL
	}
	if req.DefaultModel != nil {
		provider.DefaultModel = req.DefaultModel
	}
	if req.IsEnabled != nil {
		provider.IsEnabled = *req.IsEnabled
	}

	if err := h.service.Update(provider); err != nil {
		handlers.RespondError(w, "Failed to update provider: "+err.Error(), http.StatusInternalServerError)
		return
	}

	handlers.RespondJSON(w, toProviderResponse(provider), http.StatusOK)
}

// Delete godoc
// @Summary Delete an LLM provider
// @Description Delete an LLM provider configuration
// @Tags llm-providers
// @Param id path int true "Provider ID"
// @Success 204 "No Content"
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/llm/providers/{id} [delete]
func (h *ProviderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handlers.RespondError(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(id); err != nil {
		handlers.RespondError(w, "Failed to delete provider: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func toProviderResponse(p *models.LLMProvider) ProviderResponse {
	return ProviderResponse{
		ProviderID:   p.LLMProviderId,
		ProviderType: string(p.ProviderType),
		Name:         p.Name,
		BaseURL:      p.BaseURL,
		DefaultModel: p.DefaultModel,
		IsEnabled:    p.IsEnabled,
		CreatedAt:    p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    p.UpdatedAt.Format(time.RFC3339),
	}
}

func isValidProviderType(pt models.LLMProviderType) bool {
	validTypes := models.AllLLMProviderTypes()
	for _, t := range validTypes {
		if t == pt {
			return true
		}
	}
	return false
}
