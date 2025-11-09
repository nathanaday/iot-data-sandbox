package data

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/storage"
	"github.com/nathanaday/iot-data-sandbox/internal/timeseries"
)

type UIHandler struct {
	store     *persistence.Store
	fileStore *storage.FileStore
}

func NewUIHandler(store *persistence.Store, fileStore *storage.FileStore) *UIHandler {
	return &UIHandler{
		store:     store,
		fileStore: fileStore,
	}
}

// PreviewCSV godoc
// @Summary Preview CSV data before saving
// @Description Upload and preview a CSV file to see metadata without creating a datasource or saving the file
// @Tags ui
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV file to preview"
// @Success 200 {object} PreviewDataResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/ui/preview_data [post]
func (h *UIHandler) PreviewCSV(w http.ResponseWriter, r *http.Request) {
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

	// Save file temporarily
	tempFilename, err := h.fileStore.SaveFile(header.Filename, file, 10<<20)
	if err != nil {
		respondError(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		return
	}
	defer h.fileStore.DeleteFile(tempFilename) // Clean up temp file

	// Load and validate CSV using timeseries package
	filePath := h.fileStore.GetFilePath(tempFilename)
	tsData, err := timeseries.LoadAndValidateCSV(filePath)
	if err != nil {
		respondError(w, fmt.Sprintf("Failed to parse CSV: %v", err), http.StatusBadRequest)
		return
	}

	response := PreviewDataResponse{
		Type:       "csv",
		RowCount:   tsData.RowCount,
		StartTime:  &tsData.StartTime,
		EndTime:    &tsData.EndTime,
		TimeLabel:  tsData.TimeLabel,
		ValueLabel: tsData.ValueLabel,
	}

	respondJSON(w, response, http.StatusOK)
}
