package services

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/nathanaday/iot-data-sandbox/internal/jobs"
	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/timeseries"
)

type ProjectService struct {
	store            *persistence.Store
	dataLayerService *DataLayerService
	jobManager       *jobs.JobManager
}

func NewProjectService(store *persistence.Store, dataLayerService *DataLayerService, jobManager *jobs.JobManager) *ProjectService {
	return &ProjectService{
		store:            store,
		dataLayerService: dataLayerService,
		jobManager:       jobManager,
	}
}

// Create creates a new empty project
func (s *ProjectService) Create(name string) (*models.Project, error) {
	project := &models.Project{
		Name:        name,
		WhenCreated: time.Now(),
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

// LoadMultiColumnCSV loads a multi-column CSV and creates separate DataFrames and layers for each value column
// Each value column (non-timestamp) becomes its own layer with a dedicated DataFrame
func (s *ProjectService) LoadMultiColumnCSV(projectID int64, csvReader io.Reader) ([]*models.DataLayer, error) {
	startTime := time.Now()
	fmt.Printf("[Project Service] Starting LoadMultiColumnCSV for project %d at %s\n", projectID, startTime.Format("15:04:05.000"))

	// Verify project exists
	_, err := s.store.LoadProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Split the CSV into separate time series (one per value column)
	splitStartTime := time.Now()
	tsDataList, err := timeseries.LoadAndSplitMultiColumnCSV(csvReader)
	splitTime := time.Since(splitStartTime)
	fmt.Printf("[Project Service] CSV split completed in %v, got %d time series\n", splitTime, len(tsDataList))

	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(tsDataList) == 0 {
		return nil, fmt.Errorf("no data columns found in CSV")
	}

	// Get the current max z-index for this project
	existingLayers, err := s.store.LoadLayersByProjectId(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to load existing layers: %w", err)
	}

	maxZIndex := -1
	for _, layer := range existingLayers {
		if layer.ZIndex > maxZIndex {
			maxZIndex = layer.ZIndex
		}
	}

	// Create a DataFrame and layer for each time series
	var createdLayers []*models.DataLayer
	dataframeService := NewDataFrameService(s.store)

	createStartTime := time.Now()
	fmt.Printf("[Project Service] Starting to create %d DataFrames and layers\n", len(tsDataList))

	for i, tsData := range tsDataList {
		dfStartTime := time.Now()
		fmt.Printf("[Project Service] Creating DataFrame %d/%d (%s)...\n", i+1, len(tsDataList), tsData.ValueLabel)

		// Create DataFrame
		dataframe, err := dataframeService.CreateFromGotaDataFrame(projectID, tsData.ValueLabel, tsData)
		dfTime := time.Since(dfStartTime)
		fmt.Printf("[Project Service] DataFrame %d/%d created in %v (id: %d)\n", i+1, len(tsDataList), dfTime, dataframe.DataFrameId)
		if err != nil {
			// Rollback: delete already created dataframes and layers
			for _, layer := range createdLayers {
				if layer.DataFrameId != nil {
					dataframeService.Delete(*layer.DataFrameId)
				}
				s.dataLayerService.Delete(layer.DataLayerId)
			}
			return nil, fmt.Errorf("failed to create dataframe for column '%s': %w", tsData.ValueLabel, err)
		}

		// Create layer with the value column name
		layer := &models.DataLayer{
			ProjectId:   projectID,
			DataFrameId: &dataframe.DataFrameId,
			Name:        tsData.ValueLabel,
			Color:       "#3b82f6", // default blue
			ZIndex:      maxZIndex + 1 + i,
			IsVisible:   true,
		}

		if err := s.store.SaveLayer(layer); err != nil {
			// Rollback: delete already created dataframes and layers
			dataframeService.Delete(dataframe.DataFrameId)
			for _, l := range createdLayers {
				if l.DataFrameId != nil {
					dataframeService.Delete(*l.DataFrameId)
				}
				s.dataLayerService.Delete(l.DataLayerId)
			}
			return nil, fmt.Errorf("failed to create layer for column '%s': %w", tsData.ValueLabel, err)
		}

		createdLayers = append(createdLayers, layer)
	}

	createTime := time.Since(createStartTime)
	totalTime := time.Since(startTime)
	fmt.Printf("[Project Service] All DataFrames and layers created in %v (total: %v)\n", createTime, totalTime)

	return createdLayers, nil
}

// LoadMultiColumnCSVAsync starts an async job to load a multi-column CSV
// Returns a job ID immediately, processing happens in background
func (s *ProjectService) LoadMultiColumnCSVAsync(projectID int64, csvReader io.Reader) (string, error) {
	// First, we need to read the CSV to get column names to initialize the job
	// We'll need to buffer the CSV data since we can only read the reader once
	csvData, err := io.ReadAll(csvReader)
	if err != nil {
		return "", fmt.Errorf("failed to read CSV data: %w", err)
	}

	// Parse CSV to get column names (for job initialization)
	previewReader := io.NopCloser(bytes.NewReader(csvData))
	tsDataList, err := timeseries.LoadAndSplitMultiColumnCSV(previewReader)
	if err != nil {
		return "", fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(tsDataList) == 0 {
		return "", fmt.Errorf("no data columns found in CSV")
	}

	// Get layer names
	layerNames := make([]string, len(tsDataList))
	for i, ts := range tsDataList {
		layerNames[i] = ts.ValueLabel
	}

	// Create job
	job := s.jobManager.CreateJob(projectID, layerNames)
	fmt.Printf("[Project Service] Created async job %s for project %d with %d layers\n", job.JobID, projectID, len(layerNames))

	// Start background processing
	go s.processCSVUploadJob(job.JobID, projectID, csvData, tsDataList)

	return job.JobID, nil
}

// processCSVUploadJob processes the CSV upload in the background
func (s *ProjectService) processCSVUploadJob(jobID string, projectID int64, csvData []byte, tsDataList []*timeseries.TimeSeriesData) {
	startTime := time.Now()
	fmt.Printf("[Project Service] Starting background job %s at %s\n", jobID, startTime.Format("15:04:05.000"))

	// Update job status to in progress
	s.jobManager.UpdateJobStatus(jobID, jobs.JobStatusInProgress)

	// Get the current max z-index for this project
	existingLayers, err := s.store.LoadLayersByProjectId(projectID)
	if err != nil {
		s.jobManager.SetJobError(jobID, fmt.Sprintf("failed to load existing layers: %v", err))
		return
	}

	maxZIndex := -1
	for _, layer := range existingLayers {
		if layer.ZIndex > maxZIndex {
			maxZIndex = layer.ZIndex
		}
	}

	// Create a DataFrame and layer for each time series
	var createdLayers []*models.DataLayer
	dataframeService := NewDataFrameService(s.store)

	for i, tsData := range tsDataList {
		layerIndex := i
		fmt.Printf("[Project Service] Processing layer %d/%d (%s)...\n", i+1, len(tsDataList), tsData.ValueLabel)

		// Update layer total rows in job
		s.jobManager.UpdateLayerProgress(jobID, layerIndex, 0, tsData.RowCount-1)

		// Create progress callback for this layer
		progressCallback := func(rowsWritten, totalRows int) {
			s.jobManager.UpdateLayerProgress(jobID, layerIndex, rowsWritten, totalRows)
		}

		// Create DataFrame with progress reporting
		dataframe, err := dataframeService.CreateFromGotaDataFrameWithProgress(projectID, tsData.ValueLabel, tsData, progressCallback)
		if err != nil {
			// Rollback: delete already created dataframes and layers
			for _, layer := range createdLayers {
				if layer.DataFrameId != nil {
					dataframeService.Delete(*layer.DataFrameId)
				}
				s.dataLayerService.Delete(layer.DataLayerId)
			}
			s.jobManager.SetJobError(jobID, fmt.Sprintf("failed to create dataframe for column '%s': %v", tsData.ValueLabel, err))
			return
		}

		// Mark layer as complete
		s.jobManager.UpdateLayerStatus(jobID, layerIndex, jobs.JobStatusSuccess)

		// Create layer with the value column name
		layer := &models.DataLayer{
			ProjectId:   projectID,
			DataFrameId: &dataframe.DataFrameId,
			Name:        tsData.ValueLabel,
			Color:       "#3b82f6", // default blue
			ZIndex:      maxZIndex + 1 + i,
			IsVisible:   true,
		}

		if err := s.store.SaveLayer(layer); err != nil {
			// Rollback: delete already created dataframes and layers
			dataframeService.Delete(dataframe.DataFrameId)
			for _, l := range createdLayers {
				if l.DataFrameId != nil {
					dataframeService.Delete(*l.DataFrameId)
				}
				s.dataLayerService.Delete(l.DataLayerId)
			}
			s.jobManager.SetJobError(jobID, fmt.Sprintf("failed to create layer for column '%s': %v", tsData.ValueLabel, err))
			return
		}

		createdLayers = append(createdLayers, layer)
	}

	// Mark job as complete
	s.jobManager.UpdateJobStatus(jobID, jobs.JobStatusSuccess)

	totalTime := time.Since(startTime)
	fmt.Printf("[Project Service] Job %s completed in %v\n", jobID, totalTime)
}

// GetJobStatus retrieves the status of a job
func (s *ProjectService) GetJobStatus(jobID string) (*jobs.UploadJob, error) {
	job, exists := s.jobManager.GetJob(jobID)
	if !exists {
		return nil, fmt.Errorf("job not found")
	}
	return job, nil
}

