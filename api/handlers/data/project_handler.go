package data

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/nathanaday/iot-data-sandbox/api/handlers"
	"github.com/nathanaday/iot-data-sandbox/internal/jobs"
	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/services"
)

type ProjectHandler struct {
	store          *persistence.Store
	projectService *services.ProjectService
}

func NewProjectHandler(store *persistence.Store, jobManager *jobs.JobManager) *ProjectHandler {
	dataframeService := services.NewDataFrameService(store)
	dataLayerService := services.NewDataLayerService(store, dataframeService)
	projectService := services.NewProjectService(store, dataLayerService, jobManager)

	return &ProjectHandler{
		store:          store,
		projectService: projectService,
	}
}

// CreateProject godoc
// @Summary Create a new project
// @Description Create a new empty project
// @Tags projects
// @Accept json
// @Produce json
// @Param project body CreateProjectRequest true "Project details"
// @Success 201 {object} ProjectResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/projects [post]
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		handlers.RespondError(w, "Project name is required", http.StatusBadRequest)
		return
	}

	project, err := h.projectService.Create(req.Name)
	if err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to create project: %v", err), http.StatusInternalServerError)
		return
	}

	response := ProjectResponse{
		ProjectId:   project.ProjectId,
		Name:        project.Name,
		WhenCreated: project.WhenCreated,
		LayerCount:  0,
	}

	handlers.RespondJSON(w, response, http.StatusCreated)
}

// GetProject godoc
// @Summary Get project details
// @Description Get a specific project by ID with layer count
// @Tags projects
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} ProjectResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/projects/{id} [get]
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handlers.RespondError(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.projectService.LoadByID(id)
	if err != nil {
		handlers.RespondError(w, "Project not found", http.StatusNotFound)
		return
	}

	layerCount, _ := h.projectService.GetLayerCount(id)

	response := ProjectResponse{
		ProjectId:   project.ProjectId,
		Name:        project.Name,
		WhenCreated: project.WhenCreated,
		LayerCount:  layerCount,
	}

	handlers.RespondJSON(w, response, http.StatusOK)
}

// ListProjects godoc
// @Summary List all projects
// @Description Get a list of all projects
// @Tags projects
// @Produce json
// @Success 200 {object} ProjectListResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/projects [get]
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.projectService.LoadAll()
	if err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to load projects: %v", err), http.StatusInternalServerError)
		return
	}

	responses := make([]ProjectResponse, 0, len(projects))
	for _, project := range projects {
		layerCount, _ := h.projectService.GetLayerCount(project.ProjectId)

		responses = append(responses, ProjectResponse{
			ProjectId:   project.ProjectId,
			Name:        project.Name,
			WhenCreated: project.WhenCreated,
			LayerCount:  layerCount,
		})
	}

	handlers.RespondJSON(w, ProjectListResponse{Projects: responses}, http.StatusOK)
}

// DeleteProject godoc
// @Summary Delete a project
// @Description Delete a project and all its layers (cascade)
// @Tags projects
// @Param id path int true "Project ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handlers.RespondError(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	if err := h.projectService.Delete(id); err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to delete project: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddLayer godoc
// @Summary Add a layer to a project
// @Description Create a new layer and add it to the project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param layer body CreateLayerRequest true "Layer details"
// @Success 201 {object} LayerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/projects/{id}/layers [post]
func (h *ProjectHandler) AddLayer(w http.ResponseWriter, r *http.Request) {
	projectId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handlers.RespondError(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var req CreateLayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		handlers.RespondError(w, "Layer name is required", http.StatusBadRequest)
		return
	}

	layer, err := h.projectService.AddLayer(projectId, req.Name)
	if err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to add layer: %v", err), http.StatusInternalServerError)
		return
	}

	response := modelToLayerResponse(layer)
	handlers.RespondJSON(w, response, http.StatusCreated)
}

// GetProjectLayers godoc
// @Summary Get all layers in a project
// @Description Get all layers that belong to a specific project
// @Tags projects
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} LayerListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/projects/{id}/layers [get]
func (h *ProjectHandler) GetProjectLayers(w http.ResponseWriter, r *http.Request) {
	projectId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handlers.RespondError(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.projectService.LoadWithLayers(projectId)
	if err != nil {
		handlers.RespondError(w, "Project not found", http.StatusNotFound)
		return
	}

	responses := make([]LayerResponse, 0, len(project.Layers))
	for _, layer := range project.Layers {
		responses = append(responses, modelToLayerResponse(layer))
	}

	handlers.RespondJSON(w, LayerListResponse{Layers: responses}, http.StatusOK)
}

// LoadCSV godoc
// @Summary Load multi-column CSV into project (async)
// @Description Upload a multi-column CSV file asynchronously. Returns a job ID immediately. Creates separate DataFrames and layers for each value column (non-timestamp). For example, a CSV with columns (ts, humidity, smoke, temp) will create three layers: humidity, smoke, and temp. Use GET /api/projects/{id}/load-csv/status?job={job_id} to check progress.
// @Tags projects
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Project ID"
// @Param file formData file true "CSV file to upload"
// @Success 202 {object} UploadJobResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/projects/{id}/load-csv [post]
func (h *ProjectHandler) LoadCSV(w http.ResponseWriter, r *http.Request) {
	projectId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handlers.RespondError(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Parse multipart form (allow larger files for async processing)
	if err := r.ParseMultipartForm(100 << 20); err != nil { // 100MB max
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

	// Start async CSV load and get job ID
	jobID, err := h.projectService.LoadMultiColumnCSVAsync(projectId, file)
	if err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to start CSV upload: %v", err), http.StatusBadRequest)
		return
	}

	fmt.Printf("[API] Created upload job %s for project %d\n", jobID, projectId)

	// Return job ID immediately
	handlers.RespondJSON(w, UploadJobResponse{JobID: jobID}, http.StatusAccepted)
}

// GetLoadCSVStatus godoc
// @Summary Get CSV upload job status
// @Description Check the progress of an async CSV upload job by job ID
// @Tags projects
// @Produce json
// @Param id path int true "Project ID"
// @Param job query string true "Job ID"
// @Success 200 {object} JobStatusResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/projects/{id}/load-csv/status [get]
func (h *ProjectHandler) GetLoadCSVStatus(w http.ResponseWriter, r *http.Request) {
	projectId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handlers.RespondError(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	jobID := r.URL.Query().Get("job")
	if jobID == "" {
		handlers.RespondError(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	// Get job status
	job, err := h.projectService.GetJobStatus(jobID)
	if err != nil {
		handlers.RespondError(w, "Job not found", http.StatusNotFound)
		return
	}

	// Verify job belongs to this project
	if job.ProjectID != projectId {
		handlers.RespondError(w, "Job does not belong to this project", http.StatusBadRequest)
		return
	}

	// Convert to response format
	layerDetails := make([]LayerStatusDetail, len(job.Layers))
	for i, layer := range job.Layers {
		layerDetails[i] = LayerStatusDetail{
			LayerName:       layer.LayerName,
			RowsWritten:     layer.RowsWritten,
			TotalRows:       layer.TotalRows,
			PercentComplete: layer.PercentComplete,
			Status:          string(layer.Status),
		}
	}

	response := JobStatusResponse{
		JobID:       job.JobID,
		ProjectID:   job.ProjectID,
		Status:      string(job.Status),
		Layers:      layerDetails,
		Error:       job.Error,
		CreatedAt:   job.CreatedAt,
		UpdatedAt:   job.UpdatedAt,
		CompletedAt: job.CompletedAt,
	}

	handlers.RespondJSON(w, response, http.StatusOK)
}

// Helper function to convert model to response
func modelToLayerResponse(layer *models.DataLayer) LayerResponse {
	return LayerResponse{
		DataLayerId: layer.DataLayerId,
		ProjectId:   layer.ProjectId,
		DataFrameId: layer.DataFrameId,
		Name:        layer.Name,
		Color:       layer.Color,
		ZIndex:      layer.ZIndex,
		IsVisible:   layer.IsVisible,
	}
}
