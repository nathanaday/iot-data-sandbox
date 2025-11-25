package data

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nathanaday/iot-data-sandbox/api/handlers"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/services"
)

type DataFrameHandler struct {
	store            *persistence.Store
	dataframeService *services.DataFrameService
}

func NewDataFrameHandler(store *persistence.Store) *DataFrameHandler {
	return &DataFrameHandler{
		store:            store,
		dataframeService: services.NewDataFrameService(store),
	}
}

// ListDataFrames godoc
// @Summary List all dataframes
// @Description Get a list of all registered dataframes with their metadata
// @Tags dataframes
// @Produce json
// @Success 200 {object} DataFrameListResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/dataframes [get]
func (h *DataFrameHandler) ListDataFrames(w http.ResponseWriter, r *http.Request) {
	schemas, err := h.store.LoadAllDataFrames()
	if err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to load dataframes: %v", err), http.StatusInternalServerError)
		return
	}

	metadata := make([]DataFrameMetadata, 0, len(schemas))
	for _, schema := range schemas {
		metadata = append(metadata, DataFrameMetadata{
			DataFrameId:  schema.DataFrameId,
			ProjectId:    schema.ProjectId,
			Name:         schema.Name,
			Description:  schema.Description,
			RowCount:     schema.RowCount,
			StartTime:    schema.StartTime,
			EndTime:      schema.EndTime,
			CreatedAt:    schema.CreatedAt,
		})
	}

	handlers.RespondJSON(w, DataFrameListResponse{DataFrames: metadata}, http.StatusOK)
}

// GetDataFrame godoc
// @Summary Get dataframe metadata
// @Description Get metadata for a specific dataframe by ID
// @Tags dataframes
// @Produce json
// @Param id path int true "DataFrame ID"
// @Success 200 {object} DataFrameMetadata
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/dataframes/{id} [get]
func (h *DataFrameHandler) GetDataFrame(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handlers.RespondError(w, "Invalid dataframe ID", http.StatusBadRequest)
		return
	}

	schema, err := h.store.LoadDataFrame(id)
	if err != nil {
		handlers.RespondError(w, "DataFrame not found", http.StatusNotFound)
		return
	}

	metadata := DataFrameMetadata{
		DataFrameId:  schema.DataFrameId,
		ProjectId:    schema.ProjectId,
		Name:         schema.Name,
		Description:  schema.Description,
		RowCount:     schema.RowCount,
		StartTime:    schema.StartTime,
		EndTime:      schema.EndTime,
		CreatedAt:    schema.CreatedAt,
	}

	handlers.RespondJSON(w, metadata, http.StatusOK)
}

// DeleteDataFrame godoc
// @Summary Delete a dataframe
// @Description Delete a dataframe and its associated data table
// @Tags dataframes
// @Param id path int true "DataFrame ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/dataframes/{id} [delete]
func (h *DataFrameHandler) DeleteDataFrame(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handlers.RespondError(w, "Invalid dataframe ID", http.StatusBadRequest)
		return
	}

	if err := h.dataframeService.Delete(id); err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to delete dataframe: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
