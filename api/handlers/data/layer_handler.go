package data

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nathanaday/iot-data-sandbox/api/handlers"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/services"
)

type LayerHandler struct {
	store            *persistence.Store
	dataLayerService *services.DataLayerService
}

func NewLayerHandler(store *persistence.Store) *LayerHandler {
	dataframeService := services.NewDataFrameService(store)
	dataLayerService := services.NewDataLayerService(store, dataframeService)

	return &LayerHandler{
		store:            store,
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
		handlers.RespondError(w, "Invalid layer ID", http.StatusBadRequest)
		return
	}

	layer, err := h.dataLayerService.LoadByID(id)
	if err != nil {
		handlers.RespondError(w, "Layer not found", http.StatusNotFound)
		return
	}

	response := modelToLayerResponse(layer)
	handlers.RespondJSON(w, response, http.StatusOK)
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
		handlers.RespondError(w, "Invalid layer ID", http.StatusBadRequest)
		return
	}

	if err := h.dataLayerService.Delete(id); err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to delete layer: %v", err), http.StatusInternalServerError)
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
		handlers.RespondError(w, "Invalid layer ID", http.StatusBadRequest)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		handlers.RespondError(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		handlers.RespondError(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		handlers.RespondError(w, "File must be a CSV", http.StatusBadRequest)
		return
	}

	// Load CSV directly into layer (no filesystem storage)
	// CSV data is parsed and inserted directly into SQLite
	if err := h.dataLayerService.LoadFromCSV(id, file); err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to load CSV: %v", err), http.StatusBadRequest)
		return
	}

	// Return updated layer
	layer, err := h.dataLayerService.LoadByID(id)
	if err != nil {
		handlers.RespondError(w, "Failed to load updated layer", http.StatusInternalServerError)
		return
	}

	response := modelToLayerResponse(layer)
	handlers.RespondJSON(w, response, http.StatusOK)
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
		handlers.RespondError(w, "Invalid layer ID", http.StatusBadRequest)
		return
	}

	var req UpdateColorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Color == "" {
		handlers.RespondError(w, "Color is required", http.StatusBadRequest)
		return
	}

	if err := h.dataLayerService.UpdateColor(id, req.Color); err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to update color: %v", err), http.StatusInternalServerError)
		return
	}

	layer, err := h.dataLayerService.LoadByID(id)
	if err != nil {
		handlers.RespondError(w, "Failed to load updated layer", http.StatusInternalServerError)
		return
	}

	response := modelToLayerResponse(layer)
	handlers.RespondJSON(w, response, http.StatusOK)
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
		handlers.RespondError(w, "Invalid layer ID", http.StatusBadRequest)
		return
	}

	var req UpdateVisibilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.dataLayerService.SetVisibility(id, req.IsVisible); err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to update visibility: %v", err), http.StatusInternalServerError)
		return
	}

	layer, err := h.dataLayerService.LoadByID(id)
	if err != nil {
		handlers.RespondError(w, "Failed to load updated layer", http.StatusInternalServerError)
		return
	}

	response := modelToLayerResponse(layer)
	handlers.RespondJSON(w, response, http.StatusOK)
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
		handlers.RespondError(w, "Invalid layer ID", http.StatusBadRequest)
		return
	}

	var req DuplicateLayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.NewName == "" {
		handlers.RespondError(w, "New name is required", http.StatusBadRequest)
		return
	}

	newLayer, err := h.dataLayerService.Duplicate(id, req.NewName)
	if err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to duplicate layer: %v", err), http.StatusInternalServerError)
		return
	}

	response := modelToLayerResponse(newLayer)
	handlers.RespondJSON(w, response, http.StatusCreated)
}

// GetLayerDataMetadata godoc
// @Summary Get layer data source metadata
// @Description Get metadata about the data source associated with a layer
// @Tags layers
// @Produce json
// @Param id path int true "Layer ID"
// @Success 200 {object} DataSourceMetadata
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/layers/{id}/data/metadata [get]
func (h *LayerHandler) GetLayerDataMetadata(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handlers.RespondError(w, "Invalid layer ID", http.StatusBadRequest)
		return
	}

	layer, dataframeSchema, err := h.store.LoadLayerWithDataFrame(id)
	if err != nil {
		handlers.RespondError(w, "Layer not found", http.StatusNotFound)
		return
	}

	if layer.DataFrameId == nil || dataframeSchema == nil {
		handlers.RespondError(w, "Layer has no associated dataframe", http.StatusNotFound)
		return
	}

	metadata := DataFrameMetadata{
		DataFrameId: dataframeSchema.DataFrameId,
		ProjectId:   dataframeSchema.ProjectId,
		Name:        dataframeSchema.Name,
		Description: dataframeSchema.Description,
		RowCount:    dataframeSchema.RowCount,
		StartTime:   dataframeSchema.StartTime,
		EndTime:     dataframeSchema.EndTime,
		CreatedAt:   dataframeSchema.CreatedAt,
	}

	handlers.RespondJSON(w, metadata, http.StatusOK)
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
		handlers.RespondError(w, "Invalid layer ID", http.StatusBadRequest)
		return
	}

	layer, err := h.dataLayerService.LoadWithDataFrame(id)
	if err != nil {
		handlers.RespondError(w, "Layer not found", http.StatusNotFound)
		return
	}

	if layer.DataFrame == nil {
		handlers.RespondJSON(w, DataQueryResponse{
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
			handlers.RespondError(w, "Invalid start_time format, use RFC3339", http.StatusBadRequest)
			return
		}
		startTime = &t
	}

	if endStr := r.URL.Query().Get("end_time"); endStr != "" {
		t, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			handlers.RespondError(w, "Invalid end_time format, use RFC3339", http.StatusBadRequest)
			return
		}
		endTime = &t
	}

	// Extract data from Gota DataFrame
	df := layer.DataFrame.Data
	timestampSeries := df.Col("timestamp")
	timestamps := timestampSeries.Records()

	// Get first value column (support multi-column in future)
	var valueSeries []string
	columnNames := df.Names()
	for _, colName := range columnNames {
		if colName != "timestamp" {
			valueSeries = df.Col(colName).Records()
			break
		}
	}

	if valueSeries == nil {
		handlers.RespondJSON(w, DataQueryResponse{
			Data:     []DataPoint{},
			RowCount: 0,
		}, http.StatusOK)
		return
	}

	// Filter data based on time range
	dataPoints := make([]DataPoint, 0)
	var actualStart, actualEnd time.Time

	// Skip header row (index 0)
	for i := 1; i < len(timestamps); i++ {
		ts, err := time.Parse(time.RFC3339, timestamps[i])
		if err != nil {
			continue
		}

		// Apply time range filter
		if startTime != nil && ts.Before(*startTime) {
			continue
		}
		if endTime != nil && ts.After(*endTime) {
			continue
		}

		val, err := strconv.ParseFloat(valueSeries[i], 64)
		if err != nil {
			continue
		}

		dataPoints = append(dataPoints, DataPoint{
			Timestamp: ts,
			Value:     val,
		})

		if len(dataPoints) == 1 {
			actualStart = ts
		}
		actualEnd = ts
	}

	response := DataQueryResponse{
		Data:     dataPoints,
		RowCount: len(dataPoints),
	}

	if len(dataPoints) > 0 {
		response.StartTime = actualStart
		response.EndTime = actualEnd
	}

	handlers.RespondJSON(w, response, http.StatusOK)
}
