package data

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nathanaday/iot-data-sandbox/internal/models"
)

// QueryData godoc
// @Summary Query time series data
// @Description Query time series data from a datasource with optional time range filtering
// @Tags datasources
// @Produce json
// @Param id path int true "Datasource ID"
// @Param start_time query string false "Start time in RFC3339 format (e.g., 2024-01-01T00:00:00Z)"
// @Param end_time query string false "End time in RFC3339 format (e.g., 2024-01-01T23:59:59Z)"
// @Success 200 {object} DataQueryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/datasources/{id}/data [get]
func (h *DataSourceHandler) QueryData(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, "Invalid datasource ID", http.StatusBadRequest)
		return
	}

	// Load datasource with data
	ds, err := models.LoadFromStorage(id, h.store, h.fileStore)
	if err != nil {
		respondError(w, "Datasource not found", http.StatusNotFound)
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
	dataPoints := make([]DataPoint, 0, len(ds.Data))
	var actualStart, actualEnd time.Time

	for _, entry := range ds.Data {
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
