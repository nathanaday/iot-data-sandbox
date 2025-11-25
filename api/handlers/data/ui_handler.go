package data

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/nathanaday/iot-data-sandbox/api/handlers"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/timeseries"
)

type UIHandler struct {
	store *persistence.Store
}

func NewUIHandler(store *persistence.Store) *UIHandler {
	return &UIHandler{
		store: store,
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

	// Load and validate CSV directly from reader (no filesystem storage)
	tsData, err := timeseries.LoadAndValidateCSVFromReader(file)
	if err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to parse CSV: %v", err), http.StatusBadRequest)
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

	handlers.RespondJSON(w, response, http.StatusOK)
}
