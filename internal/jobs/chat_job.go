package jobs

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// ChatJobStatus represents the status of a chat job
type ChatJobStatus string

const (
	ChatJobPending   ChatJobStatus = "pending"
	ChatJobStreaming ChatJobStatus = "streaming"
	ChatJobComplete  ChatJobStatus = "complete"
	ChatJobFailed    ChatJobStatus = "failed"
)

// ChatJob represents an active chat completion request
type ChatJob struct {
	JobID          string        `json:"job_id"`
	ConversationID string        `json:"conversation_id"`
	Status         ChatJobStatus `json:"status"`
	ResponseText   string        `json:"response_text"`
	InputTokens    int           `json:"input_tokens"`
	OutputTokens   int           `json:"output_tokens"`
	Error          string        `json:"error,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	CompletedAt    *time.Time    `json:"completed_at,omitempty"`
}

// ChatJobManager manages active chat jobs
type ChatJobManager struct {
	jobs map[string]*ChatJob
	mu   sync.RWMutex
}

// NewChatJobManager creates a new chat job manager
func NewChatJobManager() *ChatJobManager {
	return &ChatJobManager{
		jobs: make(map[string]*ChatJob),
	}
}

// CreateJob creates a new chat job
func (m *ChatJobManager) CreateJob(conversationID string) *ChatJob {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	job := &ChatJob{
		JobID:          uuid.New().String(),
		ConversationID: conversationID,
		Status:         ChatJobPending,
		ResponseText:   "",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	m.jobs[job.JobID] = job
	return job
}

// GetJob retrieves a job by ID
func (m *ChatJobManager) GetJob(jobID string) (*ChatJob, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	jobCopy := *job
	return &jobCopy, true
}

// AppendResponse appends a chunk to the response text
func (m *ChatJobManager) AppendResponse(jobID string, chunk string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if job, exists := m.jobs[jobID]; exists {
		job.ResponseText += chunk
		job.Status = ChatJobStreaming
		job.UpdatedAt = time.Now()
	}
}

// SetStreaming marks the job as streaming
func (m *ChatJobManager) SetStreaming(jobID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if job, exists := m.jobs[jobID]; exists {
		job.Status = ChatJobStreaming
		job.UpdatedAt = time.Now()
	}
}

// CompleteJob marks a job as complete with token counts
func (m *ChatJobManager) CompleteJob(jobID string, inputTokens, outputTokens int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if job, exists := m.jobs[jobID]; exists {
		job.Status = ChatJobComplete
		job.InputTokens = inputTokens
		job.OutputTokens = outputTokens
		job.UpdatedAt = time.Now()
		now := time.Now()
		job.CompletedAt = &now
	}
}

// FailJob marks a job as failed with an error message
func (m *ChatJobManager) FailJob(jobID string, errorMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if job, exists := m.jobs[jobID]; exists {
		job.Status = ChatJobFailed
		job.Error = errorMsg
		job.UpdatedAt = time.Now()
		now := time.Now()
		job.CompletedAt = &now
	}
}

// CleanupOldJobs removes jobs older than the specified duration
func (m *ChatJobManager) CleanupOldJobs(olderThan time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	for jobID, job := range m.jobs {
		if job.CompletedAt != nil && job.CompletedAt.Before(cutoff) {
			delete(m.jobs, jobID)
		}
	}
}
