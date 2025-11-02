package data

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/services"
	"github.com/nathanaday/iot-data-sandbox/internal/storage"
)

type DataSourceHandler struct {
	store             *persistence.Store
	fileStore         *storage.FileStore
	datasourceService *services.DataSourceService
}

func NewDataSourceHandler(store *persistence.Store, fileStore *storage.FileStore) *DataSourceHandler {
	return &DataSourceHandler{
		store:             store,
		fileStore:         fileStore,
		datasourceService: services.NewDataSourceService(store, fileStore),
	}
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
	schemas, err := h.store.FindAll()
	if err != nil {
		respondError(w, fmt.Sprintf("Failed to load datasources: %v", err), http.StatusInternalServerError)
		return
	}

	metadata := make([]DataSourceMetadata, 0, len(schemas))
	for _, schema := range schemas {
		metadata = append(metadata, DataSourceMetadata{
			DataSourceId: schema.DataSourceId,
			Name:         schema.Name,
			Type:         models.DataSourceTypes[schema.DataSourceType],
			RowCount:     schema.RowCount,
			StartTime:    schema.StartTime,
			EndTime:      schema.EndTime,
			TimeLabel:    schema.TimeLabel,
			ValueLabel:   schema.ValueLabel,
			WhenCreated:  schema.WhenCreated,
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

	schema, err := h.store.FindByID(id)
	if err != nil {
		respondError(w, "Datasource not found", http.StatusNotFound)
		return
	}

	metadata := DataSourceMetadata{
		DataSourceId: schema.DataSourceId,
		Name:         schema.Name,
		Type:         models.DataSourceTypes[schema.DataSourceType],
		RowCount:     schema.RowCount,
		StartTime:    schema.StartTime,
		EndTime:      schema.EndTime,
		TimeLabel:    schema.TimeLabel,
		ValueLabel:   schema.ValueLabel,
		WhenCreated:  schema.WhenCreated,
	}

	respondJSON(w, metadata, http.StatusOK)
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

	schema, err := h.store.FindByID(id)
	if err != nil {
		respondError(w, "Datasource not found", http.StatusNotFound)
		return
	}

	if err := h.fileStore.DeleteFile(schema.DataSourcePath); err != nil {
		respondError(w, fmt.Sprintf("Failed to delete file: %v", err), http.StatusInternalServerError)
		return
	}

	if err := h.store.Delete(id); err != nil {
		respondError(w, fmt.Sprintf("Failed to delete datasource: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
