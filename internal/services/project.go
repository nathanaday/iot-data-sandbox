package services

import (
	"fmt"

	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
)

// ProjectService provides business operations for Project entities
type ProjectService struct {
	store            *persistence.Store
	dataLayerService *DataLayerService
}

// NewProjectService creates a new ProjectService
func NewProjectService(store *persistence.Store, dataLayerService *DataLayerService) *ProjectService {
	return &ProjectService{
		store:            store,
		dataLayerService: dataLayerService,
	}
}

// Create creates a new empty project
func (s *ProjectService) Create(name string) (*models.Project, error) {
	project := &models.Project{
		Name: name,
	}

	if err := s.store.SaveProject(project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

// LoadByID retrieves a Project by ID without loading layers
func (s *ProjectService) LoadByID(id int64) (*models.Project, error) {
	project, err := s.store.LoadProject(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}
	return project, nil
}

// LoadWithLayers retrieves a Project with all its layers loaded
func (s *ProjectService) LoadWithLayers(id int64) (*models.Project, error) {
	project, err := s.store.LoadProject(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	// Load all layers for this project (ordered by z_index)
	layers, err := s.store.LoadLayersByProjectId(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load project layers: %w", err)
	}

	project.Layers = layers
	return project, nil
}

// LoadAll retrieves all projects
func (s *ProjectService) LoadAll() ([]*models.Project, error) {
	projects, err := s.store.LoadAllProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to load all projects: %w", err)
	}
	return projects, nil
}

// Save persists a Project (metadata only, not layers)
func (s *ProjectService) Save(project *models.Project) error {
	if err := s.store.SaveProject(project); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}
	return nil
}

// SaveAll persists a Project and all its layers
// This is a transactional operation - if any save fails, all fail
func (s *ProjectService) SaveAll(project *models.Project) error {
	// TODO: Add transaction support when Store supports BeginTx

	// Save project metadata
	if err := s.store.SaveProject(project); err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	// Save each layer
	for _, layer := range project.Layers {
		if err := s.dataLayerService.Save(layer); err != nil {
			return fmt.Errorf("failed to save layer %s: %w", layer.Name, err)
		}
	}

	return nil
}

// Delete removes a Project by ID (cascade deletes layers)
func (s *ProjectService) Delete(id int64) error {
	if err := s.store.DeleteProject(id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

// AddLayer creates a new layer and adds it to the project
func (s *ProjectService) AddLayer(projectID int64, layerName string) (*models.DataLayer, error) {
	// Verify project exists
	_, err := s.store.LoadProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Create the layer
	layer, err := s.dataLayerService.Create(projectID, layerName)
	if err != nil {
		return nil, fmt.Errorf("failed to add layer: %w", err)
	}

	return layer, nil
}

// ReorderLayer changes the z-index of a layer to reorder it within the project
func (s *ProjectService) ReorderLayer(projectID int64, layerID int64, newZIndex int) error {
	// Load all layers for this project
	layers, err := s.store.LoadLayersByProjectId(projectID)
	if err != nil {
		return fmt.Errorf("failed to load project layers: %w", err)
	}

	// Find the target layer
	var targetLayer *models.DataLayer
	for _, layer := range layers {
		if layer.DataLayerId == layerID {
			targetLayer = layer
			break
		}
	}

	if targetLayer == nil {
		return fmt.Errorf("layer not found in project")
	}

	// Validate new z-index is within bounds
	if newZIndex < 0 {
		newZIndex = 0
	}
	if newZIndex >= len(layers) {
		newZIndex = len(layers) - 1
	}

	oldZIndex := targetLayer.ZIndex

	// If no change, return early
	if oldZIndex == newZIndex {
		return nil
	}

	// Update z-indexes: shift other layers to make room
	if newZIndex > oldZIndex {
		// Moving up: decrement z-index of layers between old and new positions
		for _, layer := range layers {
			if layer.ZIndex > oldZIndex && layer.ZIndex <= newZIndex {
				layer.ZIndex--
				if err := s.store.SaveLayer(layer); err != nil {
					return fmt.Errorf("failed to reorder layers: %w", err)
				}
			}
		}
	} else {
		// Moving down: increment z-index of layers between new and old positions
		for _, layer := range layers {
			if layer.ZIndex >= newZIndex && layer.ZIndex < oldZIndex {
				layer.ZIndex++
				if err := s.store.SaveLayer(layer); err != nil {
					return fmt.Errorf("failed to reorder layers: %w", err)
				}
			}
		}
	}

	// Update target layer
	targetLayer.ZIndex = newZIndex
	if err := s.store.SaveLayer(targetLayer); err != nil {
		return fmt.Errorf("failed to update target layer: %w", err)
	}

	return nil
}

// GetLayerCount returns the number of layers in a project
func (s *ProjectService) GetLayerCount(projectID int64) (int, error) {
	layers, err := s.store.LoadLayersByProjectId(projectID)
	if err != nil {
		return 0, fmt.Errorf("failed to load project layers: %w", err)
	}
	return len(layers), nil
}
