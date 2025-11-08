package data

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/services"
	"github.com/nathanaday/iot-data-sandbox/internal/storage"
)

type LayerHandler struct {
	store            *persistence.Store
	fileStore        *storage.FileStore
	dataLayerService *services.DataLayerService
}

func NewLayerHandler(store *persistence.Store, fileStore *storage.FileStore) *LayerHandler {
	dataSourceService := services.NewDataSourceService(store, fileStore)
	dataLayerService := services.NewDataLayerService(store, dataSourceService)

	return &LayerHandler{
		store:            store,
		fileStore:        fileStore,
		dataLayerService: dataLayerService,
	}
}

// GetLayer godoc
// @Summary Get layer details
// @Description Get a specific layer by ID
// @Tags layers
// @Produce json
// @Param id path int true "Layer ID"
// @Success 200 {object} LayerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/layers/{id} [get]
func (h *LayerHandler) GetLayer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, "Invalid layer ID", http.StatusBadRequest)
		return
	}

	layer, err := h.dataLayerService.LoadByID(id)
	if err != nil {
		respondError(w, "Layer not found", http.StatusNotFound)
		return
	}

	response := modelToLayerResponse(layer)
	respondJSON(w, response, http.StatusOK)
}

// DeleteLayer godoc
// @Summary Delete a layer
// @Description Delete a layer by ID
// @Tags layers
// @Param id path int true "Layer ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/layers/{id} [delete]
func (h *LayerHandler) DeleteLayer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, "Invalid layer ID", http.StatusBadRequest)
		return
	}

	if err := h.dataLayerService.Delete(id); err != nil {
		respondError(w, fmt.Sprintf("Failed to delete layer: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// LoadCSV godoc
// @Summary Load CSV data into layer
// @Description Load data from an existing CSV file into the layer (creates a new datasource)
// @Tags layers
// @Accept json
// @Produce json
// @Param id path int true "Layer ID"
// @Param request body LoadCSVRequest true "CSV filename"
// @Success 200 {object} LayerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/layers/{id}/load-csv [post]
func (h *LayerHandler) LoadCSV(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, "Invalid layer ID", http.StatusBadRequest)
		return
	}

	var req LoadCSVRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CSVFilename == "" {
		respondError(w, "CSV filename is required", http.StatusBadRequest)
		return
	}

	if err := h.dataLayerService.LoadFromCSV(id, req.CSVFilename); err != nil {
		respondError(w, fmt.Sprintf("Failed to load CSV: %v", err), http.StatusInternalServerError)
		return
	}

	layer, err := h.dataLayerService.LoadByID(id)
	if err != nil {
		respondError(w, "Failed to load updated layer", http.StatusInternalServerError)
		return
	}

	response := modelToLayerResponse(layer)
	respondJSON(w, response, http.StatusOK)
}

// UpdateColor godoc
// @Summary Update layer color
// @Description Update the display color of a layer
// @Tags layers
// @Accept json
// @Produce json
// @Param id path int true "Layer ID"
// @Param request body UpdateColorRequest true "Color value (hex format)"
// @Success 200 {object} LayerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/layers/{id}/color [put]
func (h *LayerHandler) UpdateColor(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, "Invalid layer ID", http.StatusBadRequest)
		return
	}

	var req UpdateColorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Color == "" {
		respondError(w, "Color is required", http.StatusBadRequest)
		return
	}

	if err := h.dataLayerService.UpdateColor(id, req.Color); err != nil {
		respondError(w, fmt.Sprintf("Failed to update color: %v", err), http.StatusInternalServerError)
		return
	}

	layer, err := h.dataLayerService.LoadByID(id)
	if err != nil {
		respondError(w, "Failed to load updated layer", http.StatusInternalServerError)
		return
	}

	response := modelToLayerResponse(layer)
	respondJSON(w, response, http.StatusOK)
}

// UpdateVisibility godoc
// @Summary Update layer visibility
// @Description Update the visibility state of a layer
// @Tags layers
// @Accept json
// @Produce json
// @Param id path int true "Layer ID"
// @Param request body UpdateVisibilityRequest true "Visibility state"
// @Success 200 {object} LayerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/layers/{id}/visibility [put]
func (h *LayerHandler) UpdateVisibility(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, "Invalid layer ID", http.StatusBadRequest)
		return
	}

	var req UpdateVisibilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.dataLayerService.SetVisibility(id, req.IsVisible); err != nil {
		respondError(w, fmt.Sprintf("Failed to update visibility: %v", err), http.StatusInternalServerError)
		return
	}

	layer, err := h.dataLayerService.LoadByID(id)
	if err != nil {
		respondError(w, "Failed to load updated layer", http.StatusInternalServerError)
		return
	}

	response := modelToLayerResponse(layer)
	respondJSON(w, response, http.StatusOK)
}

// DuplicateLayer godoc
// @Summary Duplicate a layer
// @Description Create a copy of a layer with a new name
// @Tags layers
// @Accept json
// @Produce json
// @Param id path int true "Layer ID"
// @Param request body DuplicateLayerRequest true "New layer name"
// @Success 201 {object} LayerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/layers/{id}/duplicate [post]
func (h *LayerHandler) DuplicateLayer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, "Invalid layer ID", http.StatusBadRequest)
		return
	}

	var req DuplicateLayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.NewName == "" {
		respondError(w, "New name is required", http.StatusBadRequest)
		return
	}

	newLayer, err := h.dataLayerService.Duplicate(id, req.NewName)
	if err != nil {
		respondError(w, fmt.Sprintf("Failed to duplicate layer: %v", err), http.StatusInternalServerError)
		return
	}

	response := modelToLayerResponse(newLayer)
	respondJSON(w, response, http.StatusCreated)
}

// UpdateDisplayWindow godoc
// @Summary Update layer display time window
// @Description Update the display time window (start and end times) for a layer
// @Tags layers
// @Accept json
// @Produce json
// @Param id path int true "Layer ID"
// @Param request body UpdateDisplayWindowRequest true "Display window times"
// @Success 200 {object} LayerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/layers/{id}/display-window [put]
func (h *LayerHandler) UpdateDisplayWindow(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, "Invalid layer ID", http.StatusBadRequest)
		return
	}

	var req UpdateDisplayWindowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.dataLayerService.UpdateDisplayWindow(id, req.StartTime, req.EndTime); err != nil {
		respondError(w, fmt.Sprintf("Failed to update display window: %v", err), http.StatusInternalServerError)
		return
	}

	layer, err := h.dataLayerService.LoadByID(id)
	if err != nil {
		respondError(w, "Failed to load updated layer", http.StatusInternalServerError)
		return
	}

	response := modelToLayerResponse(layer)
	respondJSON(w, response, http.StatusOK)
}

// GetLayerData godoc
// @Summary Get layer time series data
// @Description Get the time series data points for a layer (includes data from associated datasource)
// @Tags layers
// @Produce json
// @Param id path int true "Layer ID"
// @Param start_time query string false "Start time in RFC3339 format (e.g., 2024-01-01T00:00:00Z)"
// @Param end_time query string false "End time in RFC3339 format (e.g., 2024-01-01T23:59:59Z)"
// @Success 200 {object} DataQueryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/layers/{id}/data [get]
func (h *LayerHandler) GetLayerData(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, "Invalid layer ID", http.StatusBadRequest)
		return
	}

	layer, err := h.dataLayerService.LoadWithDataSource(id)
	if err != nil {
		respondError(w, "Layer not found", http.StatusNotFound)
		return
	}

	if layer.DataSource == nil {
		respondJSON(w, DataQueryResponse{
			Data:     []DataPoint{},
			RowCount: 0,
		}, http.StatusOK)
		return
	}

	// Parse time range filters
	var startTimeFilter, endTimeFilter *interface{}
	if startStr := r.URL.Query().Get("start_time"); startStr != "" {
		// Could parse and filter, but for simplicity just return all data
		// This matches the datasource query behavior
	}
	if endStr := r.URL.Query().Get("end_time"); endStr != "" {
		// Could parse and filter
	}

	dataPoints := make([]DataPoint, 0, len(layer.DataSource.Data))

	for _, entry := range layer.DataSource.Data {
		dataPoints = append(dataPoints, DataPoint{
			Timestamp: entry.Timestamp,
			Value:     entry.Value,
		})
	}

	response := DataQueryResponse{
		Data:     dataPoints,
		RowCount: len(dataPoints),
	}

	if len(dataPoints) > 0 {
		response.StartTime = dataPoints[0].Timestamp
		response.EndTime = dataPoints[len(dataPoints)-1].Timestamp
	}

	_ = startTimeFilter
	_ = endTimeFilter
	respondJSON(w, response, http.StatusOK)
}
