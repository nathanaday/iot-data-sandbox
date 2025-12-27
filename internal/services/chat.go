package services

import (
	"context"
	"fmt"

	"github.com/nathanaday/iot-data-sandbox/internal/jobs"
	"github.com/nathanaday/iot-data-sandbox/internal/llm"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/tmc/langchaingo/llms"
)

const defaultMaxTokens = 128000

type ChatService struct {
	store               *persistence.Store
	chatJobManager      *jobs.ChatJobManager
	conversationManager *llm.ConversationManager
}

func NewChatService(store *persistence.Store, chatJobManager *jobs.ChatJobManager) *ChatService {
	return &ChatService{
		store:               store,
		chatJobManager:      chatJobManager,
		conversationManager: llm.NewConversationManager(),
	}
}

// SubmitMessageResult contains the result of submitting a message
type SubmitMessageResult struct {
	JobID          string
	ConversationID string
}

// SubmitMessage starts an async chat completion job
func (s *ChatService) SubmitMessage(conversationID string, userMessage string, providerID int64) (*SubmitMessageResult, error) {
	ctx := context.Background()

	// Get or create conversation
	_, exists := s.conversationManager.GetConversation(conversationID)
	if !exists {
		// Create new conversation
		conv := s.conversationManager.CreateConversation(providerID, defaultMaxTokens)
		conversationID = conv.ConversationID
	}

	// Add user message to conversation
	if err := s.conversationManager.AddUserMessage(ctx, conversationID, userMessage); err != nil {
		return nil, fmt.Errorf("failed to add user message: %w", err)
	}

	// Create job
	job := s.chatJobManager.CreateJob(conversationID)

	// Start async processing
	go s.processChat(job.JobID, conversationID, providerID)

	return &SubmitMessageResult{
		JobID:          job.JobID,
		ConversationID: conversationID,
	}, nil
}

func (s *ChatService) processChat(jobID string, conversationID string, providerID int64) {
	ctx := context.Background()

	// Load provider
	provider, err := s.store.LoadLLMProvider(providerID)
	if err != nil {
		s.chatJobManager.FailJob(jobID, fmt.Sprintf("failed to load provider: %v", err))
		return
	}

	// Create LLM client
	llmClient, err := llm.CreateLLMClient(ctx, provider)
	if err != nil {
		s.chatJobManager.FailJob(jobID, fmt.Sprintf("failed to create LLM client: %v", err))
		return
	}

	// Get message history
	messages, err := s.conversationManager.GetLangChainMessages(ctx, conversationID)
	if err != nil {
		s.chatJobManager.FailJob(jobID, fmt.Sprintf("failed to get message history: %v", err))
		return
	}

	// Mark job as streaming
	s.chatJobManager.SetStreaming(jobID)

	// Create streaming callback
	streamingFunc := func(ctx context.Context, chunk []byte) error {
		s.chatJobManager.AppendResponse(jobID, string(chunk))
		return nil
	}

	// Call LLM with streaming
	response, err := llmClient.GenerateContent(
		ctx,
		messages,
		llms.WithStreamingFunc(streamingFunc),
	)

	if err != nil {
		s.chatJobManager.FailJob(jobID, fmt.Sprintf("LLM error: %v", err))
		return
	}

	// Extract response text
	var responseText string
	if len(response.Choices) > 0 {
		responseText = response.Choices[0].Content
	}

	// Get the job to retrieve the accumulated response
	job, _ := s.chatJobManager.GetJob(jobID)
	if job != nil && job.ResponseText != "" {
		responseText = job.ResponseText
	}

	// Estimate token counts (rough approximation - 1 token per 4 chars)
	inputTokens := 0
	for _, msg := range messages {
		for _, part := range msg.Parts {
			if textPart, ok := part.(llms.TextContent); ok {
				inputTokens += len(textPart.Text) / 4
			}
		}
	}
	outputTokens := len(responseText) / 4

	// Add assistant response to conversation
	if err := s.conversationManager.AddAssistantMessage(ctx, conversationID, responseText, outputTokens); err != nil {
		// Log error but don't fail the job
		fmt.Printf("Warning: failed to add assistant message to conversation: %v\n", err)
	}

	// Mark job complete
	s.chatJobManager.CompleteJob(jobID, inputTokens, outputTokens)
}

// GetJobStatus retrieves the current status of a chat job
func (s *ChatService) GetJobStatus(jobID string) (*jobs.ChatJob, error) {
	job, exists := s.chatJobManager.GetJob(jobID)
	if !exists {
		return nil, fmt.Errorf("job not found")
	}
	return job, nil
}

// ClearConversation clears the conversation history
func (s *ChatService) ClearConversation(conversationID string) error {
	ctx := context.Background()
	return s.conversationManager.ClearHistory(ctx, conversationID)
}

// DeleteConversation removes a conversation entirely
func (s *ChatService) DeleteConversation(conversationID string) {
	s.conversationManager.DeleteConversation(conversationID)
}

// GetConversationHistory returns the message history for a conversation
func (s *ChatService) GetConversationHistory(conversationID string) []llm.ChatMessage {
	return s.conversationManager.GetMessageHistory(conversationID)
}

// GetConversationTokenCount returns the total token count for a conversation
func (s *ChatService) GetConversationTokenCount(conversationID string) int {
	return s.conversationManager.GetTotalTokens(conversationID)
}
