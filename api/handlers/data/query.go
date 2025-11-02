package data

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/timeseries"
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

	schema, err := h.store.LoadDataSource(id)
	if err != nil {
		respondError(w, "Datasource not found", http.StatusNotFound)
		return
	}

	ds := &models.DataSource{}
	ds.FromSchema(schema)

	filePath := h.fileStore.GetFilePath(ds.DataSourcePath)
	if !h.fileStore.FileExists(ds.DataSourcePath) {
		respondError(w, "Data file not found", http.StatusNotFound)
		return
	}

	tsData, err := timeseries.LoadAndValidateCSV(filePath)
	if err != nil {
		respondError(w, fmt.Sprintf("Failed to load data: %v", err), http.StatusInternalServerError)
		return
	}

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

	filteredData, err := timeseries.FilterByTimeRange(tsData, startTime, endTime)
	if err != nil {
		respondError(w, fmt.Sprintf("Failed to filter data: %v", err), http.StatusInternalServerError)
		return
	}

	dataPoints := make([]DataPoint, 0, filteredData.RowCount)

	timestampRecords := filteredData.DataFrame.Col("timestamp").Records()
	valueRecords := filteredData.DataFrame.Col("value").Records()

	for i := 1; i < len(timestampRecords); i++ {
		ts, err := time.Parse(time.RFC3339, timestampRecords[i])
		if err != nil {
			continue
		}

		val, err := strconv.ParseFloat(valueRecords[i], 64)
		if err != nil {
			continue
		}

		dataPoints = append(dataPoints, DataPoint{
			Timestamp: ts,
			Value:     val,
		})
	}

	response := DataQueryResponse{
		Data:     dataPoints,
		RowCount: len(dataPoints),
	}

	if len(dataPoints) > 0 {
		response.StartTime = filteredData.StartTime
		response.EndTime = filteredData.EndTime
	}

	respondJSON(w, response, http.StatusOK)
}
