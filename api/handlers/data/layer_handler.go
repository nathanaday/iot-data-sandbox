package data

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
// @Description Upload and load a CSV file into the layer (creates a new datasource). The CSV must have 'timestamp' and 'value' columns.
// @Tags layers
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Layer ID"
// @Param file formData file true "CSV file to upload"
// @Param name formData string false "Name for the datasource (defaults to filename)"
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

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		respondError(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		respondError(w, "File must be a CSV", http.StatusBadRequest)
		return
	}

	// Get optional datasource name from form
	name := r.FormValue("name")
	if name == "" {
		name = strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	}

	// Save the file
	savedFilename, err := h.fileStore.SaveFile(header.Filename, file, 10<<20)
	if err != nil {
		respondError(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		return
	}

	// Load CSV into layer
	if err := h.dataLayerService.LoadFromCSV(id, savedFilename); err != nil {
		// Clean up the saved file on error
		h.fileStore.DeleteFile(savedFilename)
		respondError(w, fmt.Sprintf("Failed to load CSV: %v", err), http.StatusInternalServerError)
		return
	}

	// Return updated layer
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
	var startTime, endTime *time.Time
	if startStr := r.URL.Query().Get("start_time"); startStr != "" {
		t, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			respondError(w, "Invalid start_time format, use RFC3339", http.StatusBadRequest)
			return
		}
		startTime = &t
	}

	if endStr := r.URL.Query().Get("end_time"); endStr != "" {
		t, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			respondError(w, "Invalid end_time format, use RFC3339", http.StatusBadRequest)
			return
		}
		endTime = &t
	}

	// Filter data based on time range
	dataPoints := make([]DataPoint, 0, len(layer.DataSource.Data))
	var actualStart, actualEnd time.Time

	for _, entry := range layer.DataSource.Data {
		// Apply time range filter
		if startTime != nil && entry.Timestamp.Before(*startTime) {
			continue
		}
		if endTime != nil && entry.Timestamp.After(*endTime) {
			continue
		}

		dataPoints = append(dataPoints, DataPoint{
			Timestamp: entry.Timestamp,
			Value:     entry.Value,
		})

		if len(dataPoints) == 1 {
			actualStart = entry.Timestamp
		}
		actualEnd = entry.Timestamp
	}

	response := DataQueryResponse{
		Data:     dataPoints,
		RowCount: len(dataPoints),
	}

	if len(dataPoints) > 0 {
		response.StartTime = actualStart
		response.EndTime = actualEnd
	}

	respondJSON(w, response, http.StatusOK)
}
