package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/storage"
	"github.com/nathanaday/iot-data-sandbox/internal/timeseries"
)

type DataSourceHandler struct {
	store     *persistence.Store
	fileStore *storage.FileStore
}

func NewDataSourceHandler(store *persistence.Store, fileStore *storage.FileStore) *DataSourceHandler {
	return &DataSourceHandler{
		store:     store,
		fileStore: fileStore,
	}
}

type UploadResponse struct {
	DataSourceId int64      `json:"data_source_id"`
	Name         string     `json:"name"`
	RowCount     int        `json:"row_count"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	TimeLabel    string     `json:"time_label"`
	ValueLabel   string     `json:"value_label"`
	WhenCreated  time.Time  `json:"when_created"`
}

type DataSourceListResponse struct {
	DataSources []DataSourceMetadata `json:"data_sources"`
}

type DataSourceMetadata struct {
	DataSourceId int64      `json:"data_source_id"`
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	RowCount     int        `json:"row_count"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	TimeLabel    string     `json:"time_label"`
	ValueLabel   string     `json:"value_label"`
	WhenCreated  time.Time  `json:"when_created"`
}

type DataQueryResponse struct {
	Data      []DataPoint `json:"data"`
	RowCount  int         `json:"row_count"`
	StartTime time.Time   `json:"start_time"`
	EndTime   time.Time   `json:"end_time"`
}

type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// UploadCSV godoc
// @Summary Upload a CSV datasource
// @Description Upload a CSV file containing time series data. The CSV must have 'timestamp' and 'value' columns. Supports various timestamp formats (ISO8601, Unix, Julian Day).
// @Tags datasources
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV file to upload"
// @Param name formData string false "Name for the datasource (defaults to filename)"
// @Success 201 {object} UploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/datasources [post]
func (h *DataSourceHandler) UploadCSV(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(storage.MaxFileSize); err != nil {
		respondError(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		respondError(w, "File must be a CSV", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		name = strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	}

	savedFilename, err := h.fileStore.SaveFile(header.Filename, file, storage.MaxFileSize)
	if err != nil {
		respondError(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		return
	}

	filePath := h.fileStore.GetFilePath(savedFilename)

	tsData, err := timeseries.LoadAndValidateCSV(filePath)
	if err != nil {
		h.fileStore.DeleteFile(savedFilename)
		respondError(w, fmt.Sprintf("Invalid CSV: %v", err), http.StatusBadRequest)
		return
	}

	dataSource := &models.DataSource{
		Name:           name,
		DataSourceType: 0,
		DataSourcePath: savedFilename,
		RowCount:       tsData.RowCount,
		TimeLabel:      tsData.TimeLabel,
		ValueLabel:     tsData.ValueLabel,
		WhenCreated:    time.Now(),
	}

	if tsData.RowCount > 0 {
		dataSource.StartTime = &tsData.StartTime
		dataSource.EndTime = &tsData.EndTime
	}

	if err := h.store.SaveDataSource(dataSource); err != nil {
		h.fileStore.DeleteFile(savedFilename)
		respondError(w, fmt.Sprintf("Failed to save datasource: %v", err), http.StatusInternalServerError)
		return
	}

	response := UploadResponse{
		DataSourceId: dataSource.DataSourceId,
		Name:         dataSource.Name,
		RowCount:     dataSource.RowCount,
		StartTime:    dataSource.StartTime,
		EndTime:      dataSource.EndTime,
		TimeLabel:    dataSource.TimeLabel,
		ValueLabel:   dataSource.ValueLabel,
		WhenCreated:  dataSource.WhenCreated,
	}

	respondJSON(w, response, http.StatusCreated)
}

// ListDataSources godoc
// @Summary List all datasources
// @Description Get a list of all registered datasources with their metadata
// @Tags datasources
// @Produce json
// @Success 200 {object} DataSourceListResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/datasources [get]
func (h *DataSourceHandler) ListDataSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.store.LoadAllDataSources()
	if err != nil {
		respondError(w, fmt.Sprintf("Failed to load datasources: %v", err), http.StatusInternalServerError)
		return
	}

	metadata := make([]DataSourceMetadata, 0, len(sources))
	for _, ds := range sources {
		metadata = append(metadata, DataSourceMetadata{
			DataSourceId: ds.DataSourceId,
			Name:         ds.Name,
			Type:         models.DataSourceTypes[ds.DataSourceType],
			RowCount:     ds.RowCount,
			StartTime:    ds.StartTime,
			EndTime:      ds.EndTime,
			TimeLabel:    ds.TimeLabel,
			ValueLabel:   ds.ValueLabel,
			WhenCreated:  ds.WhenCreated,
		})
	}

	respondJSON(w, DataSourceListResponse{DataSources: metadata}, http.StatusOK)
}

// GetDataSource godoc
// @Summary Get datasource metadata
// @Description Get metadata for a specific datasource by ID
// @Tags datasources
// @Produce json
// @Param id path int true "Datasource ID"
// @Success 200 {object} DataSourceMetadata
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/datasources/{id} [get]
func (h *DataSourceHandler) GetDataSource(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, "Invalid datasource ID", http.StatusBadRequest)
		return
	}

	ds, err := h.store.LoadDataSource(id)
	if err != nil {
		respondError(w, "Datasource not found", http.StatusNotFound)
		return
	}

	metadata := DataSourceMetadata{
		DataSourceId: ds.DataSourceId,
		Name:         ds.Name,
		Type:         models.DataSourceTypes[ds.DataSourceType],
		RowCount:     ds.RowCount,
		StartTime:    ds.StartTime,
		EndTime:      ds.EndTime,
		TimeLabel:    ds.TimeLabel,
		ValueLabel:   ds.ValueLabel,
		WhenCreated:  ds.WhenCreated,
	}

	respondJSON(w, metadata, http.StatusOK)
}

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

	ds, err := h.store.LoadDataSource(id)
	if err != nil {
		respondError(w, "Datasource not found", http.StatusNotFound)
		return
	}

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

// DeleteDataSource godoc
// @Summary Delete a datasource
// @Description Delete a datasource and its associated CSV file
// @Tags datasources
// @Param id path int true "Datasource ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/datasources/{id} [delete]
func (h *DataSourceHandler) DeleteDataSource(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, "Invalid datasource ID", http.StatusBadRequest)
		return
	}

	ds, err := h.store.LoadDataSource(id)
	if err != nil {
		respondError(w, "Datasource not found", http.StatusNotFound)
		return
	}

	if err := h.fileStore.DeleteFile(ds.DataSourcePath); err != nil {
		respondError(w, fmt.Sprintf("Failed to delete file: %v", err), http.StatusInternalServerError)
		return
	}

	if err := h.store.DeleteDataSource(id); err != nil {
		respondError(w, fmt.Sprintf("Failed to delete datasource: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
