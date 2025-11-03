package services

import (
	"fmt"
	"time"

	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
)

// DataLayerService provides business operations for DataLayer entities
type DataLayerService struct {
	store             *persistence.Store
	dataSourceService *DataSourceService
}

// NewDataLayerService creates a new DataLayerService
func NewDataLayerService(store *persistence.Store, dataSourceService *DataSourceService) *DataLayerService {
	return &DataLayerService{
		store:             store,
		dataSourceService: dataSourceService,
	}
}

// Create creates a new DataLayer within a project
func (s *DataLayerService) Create(projectId int64, name string) (*models.DataLayer, error) {
	layer := &models.DataLayer{
		ProjectId:  projectId,
		Name:       name,
		Color:      "#3b82f6", // default blue
		ZIndex:     0,
		IsVisible:  true,
	}

	if err := s.store.SaveLayer(layer); err != nil {
		return nil, fmt.Errorf("failed to create layer: %w", err)
	}

	return layer, nil
}

// LoadByID retrieves a DataLayer by ID
func (s *DataLayerService) LoadByID(id int64) (*models.DataLayer, error) {
	layer, err := s.store.LoadLayer(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load layer: %w", err)
	}
	return layer, nil
}

// LoadWithDataSource retrieves a DataLayer with its associated DataSource loaded
func (s *DataLayerService) LoadWithDataSource(id int64) (*models.DataLayer, error) {
	layer, dataSourceSchema, err := s.store.LoadLayerWithDataSource(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load layer with datasource: %w", err)
	}

	// Convert schema to model and load data
	dataSource := &models.DataSource{}
	dataSource.FromSchema(dataSourceSchema)

	// Load the actual time series data from CSV
	filePath := s.dataSourceService.fileStore.GetFilePath(dataSource.DataSourcePath)
	if err := loadDataFromCSV(dataSource, filePath); err != nil {
		return nil, fmt.Errorf("failed to load datasource data: %w", err)
	}

	layer.DataSource = dataSource
	return layer, nil
}

// LoadFromCSV loads CSV data into a layer by creating a new DataSource
func (s *DataLayerService) LoadFromCSV(layerId int64, csvFilename string) error {
	layer, err := s.store.LoadLayer(layerId)
	if err != nil {
		return fmt.Errorf("failed to load layer: %w", err)
	}

	// Create datasource from CSV
	dataSource, err := s.dataSourceService.CreateFromCSV(layer.Name, csvFilename)
	if err != nil {
		return fmt.Errorf("failed to create datasource from CSV: %w", err)
	}

	// Associate datasource with layer
	layer.DataSourceId = dataSource.DataSourceId
	if err := s.store.SaveLayer(layer); err != nil {
		return fmt.Errorf("failed to associate datasource with layer: %w", err)
	}

	return nil
}

// Save persists a DataLayer
func (s *DataLayerService) Save(layer *models.DataLayer) error {
	if err := s.store.SaveLayer(layer); err != nil {
		return fmt.Errorf("failed to save layer: %w", err)
	}
	return nil
}

// Delete removes a DataLayer by ID
func (s *DataLayerService) Delete(id int64) error {
	if err := s.store.DeleteLayer(id); err != nil {
		return fmt.Errorf("failed to delete layer: %w", err)
	}
	return nil
}

// UpdateDisplayWindow updates the display time window for a layer
func (s *DataLayerService) UpdateDisplayWindow(layerId int64, start, end *time.Time) error {
	layer, err := s.store.LoadLayer(layerId)
	if err != nil {
		return fmt.Errorf("failed to load layer: %w", err)
	}

	layer.DisplayStartTime = start
	layer.DisplayEndTime = end

	if err := s.store.SaveLayer(layer); err != nil {
		return fmt.Errorf("failed to update display window: %w", err)
	}

	return nil
}

// SetVisibility updates the visibility state of a layer
func (s *DataLayerService) SetVisibility(layerId int64, visible bool) error {
	layer, err := s.store.LoadLayer(layerId)
	if err != nil {
		return fmt.Errorf("failed to load layer: %w", err)
	}

	layer.IsVisible = visible

	if err := s.store.SaveLayer(layer); err != nil {
		return fmt.Errorf("failed to update visibility: %w", err)
	}

	return nil
}

// Duplicate creates a copy of a layer with a new name
func (s *DataLayerService) Duplicate(layerId int64, newName string) (*models.DataLayer, error) {
	original, err := s.store.LoadLayer(layerId)
	if err != nil {
		return nil, fmt.Errorf("failed to load original layer: %w", err)
	}

	// Create new layer with same properties but different name
	duplicate := &models.DataLayer{
		ProjectId:        original.ProjectId,
		DataSourceId:     original.DataSourceId, // shares same datasource
		Name:             newName,
		Color:            original.Color,
		ZIndex:           original.ZIndex + 1, // place above original
		IsVisible:        original.IsVisible,
		DisplayStartTime: original.DisplayStartTime,
		DisplayEndTime:   original.DisplayEndTime,
	}

	if err := s.store.SaveLayer(duplicate); err != nil {
		return nil, fmt.Errorf("failed to save duplicate layer: %w", err)
	}

	return duplicate, nil
}

// UpdateColor updates the color of a layer
func (s *DataLayerService) UpdateColor(layerId int64, color string) error {
	layer, err := s.store.LoadLayer(layerId)
	if err != nil {
		return fmt.Errorf("failed to load layer: %w", err)
	}

	layer.Color = color

	if err := s.store.SaveLayer(layer); err != nil {
		return fmt.Errorf("failed to update color: %w", err)
	}

	return nil
}

// UpdateZIndex updates the z-index (stacking order) of a layer
func (s *DataLayerService) UpdateZIndex(layerId int64, zIndex int) error {
	layer, err := s.store.LoadLayer(layerId)
	if err != nil {
		return fmt.Errorf("failed to load layer: %w", err)
	}

	layer.ZIndex = zIndex

	if err := s.store.SaveLayer(layer); err != nil {
		return fmt.Errorf("failed to update z-index: %w", err)
	}

	return nil
}

// Note: loadDataFromCSV and parseDataFrameToEntries are defined in datasource.go
// and shared across the services package
