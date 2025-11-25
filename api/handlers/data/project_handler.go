package data

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nathanaday/iot-data-sandbox/api/handlers"
	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/services"
)

type ProjectHandler struct {
	store          *persistence.Store
	projectService *services.ProjectService
}

func NewProjectHandler(store *persistence.Store) *ProjectHandler {
	dataframeService := services.NewDataFrameService(store)
	dataLayerService := services.NewDataLayerService(store, dataframeService)
	projectService := services.NewProjectService(store, dataLayerService)

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
