package data

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nathanaday/iot-data-sandbox/api/handlers"
)

const MaxFileSize = 100 * 1024 * 1024 // 100MB

// UploadCSV godoc
// @Summary Upload a CSV dataframe
// @Description Upload a CSV file containing time series data. The CSV must have 'timestamp' column and one or more value columns. Supports various timestamp formats (ISO8601, Unix, Julian Day). Data is stored directly in SQLite.
// @Tags dataframes
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV file to upload"
// @Param name formData string false "Name for the dataframe (defaults to filename)"
// @Param project_id formData int true "Project ID to associate the dataframe with"
// @Success 201 {object} UploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/dataframes [post]
func (h *DataFrameHandler) UploadCSV(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(MaxFileSize); err != nil {
		handlers.RespondError(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		handlers.RespondError(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		handlers.RespondError(w, "File must be a CSV", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		name = strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	}

	projectIdStr := r.FormValue("project_id")
	if projectIdStr == "" {
		handlers.RespondError(w, "project_id is required", http.StatusBadRequest)
		return
	}

	projectId, err := strconv.ParseInt(projectIdStr, 10, 64)
	if err != nil {
		handlers.RespondError(w, "Invalid project_id", http.StatusBadRequest)
		return
	}

	// Create dataframe directly from CSV (no filesystem storage)
	// CSV data is parsed and inserted directly into SQLite
	dataframe, err := h.dataframeService.CreateFromCSV(projectId, name, file)
	if err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to create dataframe: %v", err), http.StatusBadRequest)
		return
	}

	response := UploadResponse{
		DataFrameId: dataframe.DataFrameId,
		Name:        dataframe.Name,
		RowCount:    dataframe.RowCount,
		StartTime:   dataframe.StartTime,
		EndTime:     dataframe.EndTime,
		CreatedAt:   dataframe.CreatedAt,
	}

	handlers.RespondJSON(w, response, http.StatusCreated)
}
