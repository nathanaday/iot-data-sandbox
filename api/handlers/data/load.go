package data

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/storage"
	"github.com/nathanaday/iot-data-sandbox/internal/timeseries"
)

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

	schema := dataSource.ToSchema()
	if err := h.store.SaveDataSource(schema); err != nil {
		h.fileStore.DeleteFile(savedFilename)
		respondError(w, fmt.Sprintf("Failed to save datasource: %v", err), http.StatusInternalServerError)
		return
	}
	dataSource.DataSourceId = schema.DataSourceId

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