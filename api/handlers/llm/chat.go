package llm

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nathanaday/iot-data-sandbox/api/handlers"
	"github.com/nathanaday/iot-data-sandbox/internal/services"
)

type ChatHandler struct {
	chatService *services.ChatService
}

func NewChatHandler(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// Request types

type ChatRequest struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id,omitempty"`
	ProviderID     int64  `json:"provider_id"`
	ProjectID      int64  `json:"project_id"`
}

type ClearHistoryRequest struct {
	ConversationID string `json:"conversation_id"`
}

// Response types

type ChatJobResponse struct {
	JobID          string `json:"job_id"`
	ConversationID string `json:"conversation_id"`
}

type ToolCallInfo struct {
	ToolName   string `json:"tool_name"`
	Arguments  string `json:"arguments"`
	Result     string `json:"result,omitempty"`
	Success    bool   `json:"success"`
	ExecutedAt string `json:"executed_at"`
}

type ChatStatusResponse struct {
	JobID            string         `json:"job_id"`
	ConversationID   string         `json:"conversation_id"`
	Status           string         `json:"status"`
	ResponseText     string         `json:"response_text"`
	InputTokens      int            `json:"input_tokens"`
	OutputTokens     int            `json:"output_tokens"`
	Error            string         `json:"error,omitempty"`
	CreatedAt        string         `json:"created_at"`
	UpdatedAt        string         `json:"updated_at"`
	CompletedAt      *string        `json:"completed_at,omitempty"`
	ToolCalls        []ToolCallInfo `json:"tool_calls,omitempty"`
	CurrentIteration int            `json:"current_iteration,omitempty"`
	MaxIterations    int            `json:"max_iterations,omitempty"`
}

// SubmitMessage godoc
// @Summary Submit a chat message
// @Description Submit a message to the LLM and get a job ID for polling
// @Tags chat
// @Accept json
// @Produce json
// @Param message body ChatRequest true "Chat message"
// @Success 202 {object} ChatJobResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/chat [post]
func (h *ChatHandler) SubmitMessage(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		handlers.RespondError(w, "Message is required", http.StatusBadRequest)
		return
	}

	if req.ProviderID == 0 {
		handlers.RespondError(w, "Provider ID is required", http.StatusBadRequest)
		return
	}

	if req.ProjectID == 0 {
		handlers.RespondError(w, "Project ID is required for tool-enabled chat", http.StatusBadRequest)
		return
	}

	result, err := h.chatService.SubmitMessage(req.ConversationID, req.Message, req.ProviderID, req.ProjectID)
	if err != nil {
		handlers.RespondError(w, "Failed to submit message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	handlers.RespondJSON(w, ChatJobResponse{
		JobID:          result.JobID,
		ConversationID: result.ConversationID,
	}, http.StatusAccepted)
}

// GetStatus godoc
// @Summary Get chat job status
// @Description Poll for the status of a chat completion job
// @Tags chat
// @Produce json
// @Param jobId path string true "Job ID"
// @Success 200 {object} ChatStatusResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Router /api/chat/{jobId} [get]
func (h *ChatHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobId")
	if jobID == "" {
		handlers.RespondError(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	job, err := h.chatService.GetJobStatus(jobID)
	if err != nil {
		handlers.RespondError(w, "Job not found", http.StatusNotFound)
		return
	}

	response := ChatStatusResponse{
		JobID:            job.JobID,
		ConversationID:   job.ConversationID,
		Status:           string(job.Status),
		ResponseText:     job.ResponseText,
		InputTokens:      job.InputTokens,
		OutputTokens:     job.OutputTokens,
		Error:            job.Error,
		CreatedAt:        job.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        job.UpdatedAt.Format(time.RFC3339),
		CurrentIteration: job.CurrentIteration,
		MaxIterations:    job.MaxIterations,
	}

	if job.CompletedAt != nil {
		completedAt := job.CompletedAt.Format(time.RFC3339)
		response.CompletedAt = &completedAt
	}

	// Convert tool calls
	if len(job.ToolCalls) > 0 {
		response.ToolCalls = make([]ToolCallInfo, len(job.ToolCalls))
		for i, tc := range job.ToolCalls {
			response.ToolCalls[i] = ToolCallInfo{
				ToolName:   tc.ToolName,
				Arguments:  tc.Arguments,
				Result:     tc.Result,
				Success:    tc.Success,
				ExecutedAt: tc.ExecutedAt.Format(time.RFC3339),
			}
		}
	}

	handlers.RespondJSON(w, response, http.StatusOK)
}

// ClearHistory godoc
// @Summary Clear chat history
// @Description Clear the conversation history for a specific conversation
// @Tags chat
// @Accept json
// @Param request body ClearHistoryRequest true "Conversation ID to clear"
// @Success 204 "No Content"
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/chat/history [delete]
func (h *ChatHandler) ClearHistory(w http.ResponseWriter, r *http.Request) {
	var req ClearHistoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ConversationID == "" {
		handlers.RespondError(w, "Conversation ID is required", http.StatusBadRequest)
		return
	}

	if err := h.chatService.ClearConversation(req.ConversationID); err != nil {
		handlers.RespondError(w, "Failed to clear history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Also delete the conversation entirely
	h.chatService.DeleteConversation(req.ConversationID)

	w.WriteHeader(http.StatusNoContent)
}
