package jobs

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusInProgress JobStatus = "in_progress"
	JobStatusSuccess    JobStatus = "success"
	JobStatusFailed     JobStatus = "failed"
)

// LayerProgress tracks the progress of uploading a single layer
type LayerProgress struct {
	LayerName     string  `json:"layer_name"`
	RowsWritten   int     `json:"rows_written"`
	TotalRows     int     `json:"total_rows"`
	PercentComplete float64 `json:"percent_complete"`
	Status        JobStatus `json:"status"`
}

// UploadJob represents a CSV upload job
type UploadJob struct {
	JobID      string          `json:"job_id"`
	ProjectID  int64           `json:"project_id"`
	Status     JobStatus       `json:"status"`
	Layers     []LayerProgress `json:"layers"`
	Error      string          `json:"error,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}

// JobManager manages upload jobs
type JobManager struct {
	jobs map[string]*UploadJob
	mu   sync.RWMutex
}

// NewJobManager creates a new job manager
func NewJobManager() *JobManager {
	return &JobManager{
		jobs: make(map[string]*UploadJob),
	}
}

// CreateJob creates a new upload job
func (jm *JobManager) CreateJob(projectID int64, layerNames []string) *UploadJob {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jobID := uuid.New().String()
	now := time.Now()

	layers := make([]LayerProgress, len(layerNames))
	for i, name := range layerNames {
		layers[i] = LayerProgress{
			LayerName:       name,
			RowsWritten:     0,
			TotalRows:       0,
			PercentComplete: 0,
			Status:          JobStatusPending,
		}
	}

	job := &UploadJob{
		JobID:     jobID,
		ProjectID: projectID,
		Status:    JobStatusPending,
		Layers:    layers,
		CreatedAt: now,
		UpdatedAt: now,
	}

	jm.jobs[jobID] = job
	return job
}

// GetJob retrieves a job by ID
func (jm *JobManager) GetJob(jobID string) (*UploadJob, bool) {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	job, exists := jm.jobs[jobID]
	return job, exists
}

// UpdateJobStatus updates the overall job status
func (jm *JobManager) UpdateJobStatus(jobID string, status JobStatus) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if job, exists := jm.jobs[jobID]; exists {
		job.Status = status
		job.UpdatedAt = time.Now()
		if status == JobStatusSuccess || status == JobStatusFailed {
			now := time.Now()
			job.CompletedAt = &now
		}
	}
}

// UpdateLayerProgress updates the progress for a specific layer
func (jm *JobManager) UpdateLayerProgress(jobID string, layerIndex int, rowsWritten, totalRows int) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if job, exists := jm.jobs[jobID]; exists {
		if layerIndex >= 0 && layerIndex < len(job.Layers) {
			layer := &job.Layers[layerIndex]
			layer.RowsWritten = rowsWritten
			layer.TotalRows = totalRows
			if totalRows > 0 {
				layer.PercentComplete = float64(rowsWritten) / float64(totalRows) * 100
			}
			layer.Status = JobStatusInProgress
			job.UpdatedAt = time.Now()
		}
	}
}

// UpdateLayerStatus updates the status for a specific layer
func (jm *JobManager) UpdateLayerStatus(jobID string, layerIndex int, status JobStatus) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if job, exists := jm.jobs[jobID]; exists {
		if layerIndex >= 0 && layerIndex < len(job.Layers) {
			job.Layers[layerIndex].Status = status
			job.UpdatedAt = time.Now()
		}
	}
}

// SetJobError sets an error message for the job
func (jm *JobManager) SetJobError(jobID string, errorMsg string) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if job, exists := jm.jobs[jobID]; exists {
		job.Status = JobStatusFailed
		job.Error = errorMsg
		job.UpdatedAt = time.Now()
		now := time.Now()
		job.CompletedAt = &now
	}
}

// CleanupOldJobs removes jobs older than the specified duration
func (jm *JobManager) CleanupOldJobs(olderThan time.Duration) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	for jobID, job := range jm.jobs {
		if job.CompletedAt != nil && job.CompletedAt.Before(cutoff) {
			delete(jm.jobs, jobID)
		}
	}
}
