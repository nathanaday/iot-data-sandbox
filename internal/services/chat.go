package services

import (
	"context"
	"fmt"
	"time"

	"github.com/nathanaday/iot-data-sandbox/internal/jobs"
	"github.com/nathanaday/iot-data-sandbox/internal/llm"
	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/tmc/langchaingo/llms"
)

const (
	defaultMaxTokens     = 128000
	defaultMaxIterations = 10
)

// toolCallServiceAdapter wraps ToolCallService to implement llm.ToolCallServiceInterface
type toolCallServiceAdapter struct {
	svc *ToolCallService
}

func (a *toolCallServiceAdapter) ExecuteToolOnLayer(toolName string, layerID int64, projectID int64, outputName string, params map[string]interface{}) (*llm.ToolExecutionResult, error) {
	result, err := a.svc.ExecuteToolOnLayer(toolName, layerID, projectID, outputName, params)
	if err != nil {
		return nil, err
	}
	return &llm.ToolExecutionResult{
		ResultType: result.ResultType,
		Message:    result.Message,
		RowCount:   result.RowCount,
		Layer:      result.Layer,
	}, nil
}

// dataLayerServiceAdapter wraps DataLayerService to implement llm.DataLayerServiceInterface
type dataLayerServiceAdapter struct {
	svc *DataLayerService
}

func (a *dataLayerServiceAdapter) LoadWithDataFrame(layerID int64) (*models.DataLayer, error) {
	return a.svc.LoadWithDataFrame(layerID)
}

type ChatService struct {
	store               *persistence.Store
	chatJobManager      *jobs.ChatJobManager
	conversationManager *llm.ConversationManager
	agentExecutor       *llm.AgentExecutor
}

func NewChatService(store *persistence.Store, chatJobManager *jobs.ChatJobManager) *ChatService {
	// Create services for the agent executor
	dataframeService := NewDataFrameService(store)
	dataLayerService := NewDataLayerService(store, dataframeService)
	toolCallService := NewToolCallService(store)

	// Create adapters that implement the llm interfaces
	toolCallAdapter := &toolCallServiceAdapter{svc: toolCallService}
	dataLayerAdapter := &dataLayerServiceAdapter{svc: dataLayerService}

	return &ChatService{
		store:               store,
		chatJobManager:      chatJobManager,
		conversationManager: llm.NewConversationManager(),
		agentExecutor:       llm.NewAgentExecutor(store, toolCallAdapter, dataLayerAdapter),
	}
}

// SubmitMessageResult contains the result of submitting a message
type SubmitMessageResult struct {
	JobID          string
	ConversationID string
}

// SubmitMessage starts an async chat completion job with tool calling capabilities
func (s *ChatService) SubmitMessage(conversationID string, userMessage string, providerID int64, projectID int64) (*SubmitMessageResult, error) {
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
	s.chatJobManager.SetMaxIterations(job.JobID, defaultMaxIterations)

	// Start async processing with agentic loop
	go s.processAgenticChat(job.JobID, conversationID, providerID, projectID)

	return &SubmitMessageResult{
		JobID:          job.JobID,
		ConversationID: conversationID,
	}, nil
}

// processAgenticChat handles chat with tool calling in an agentic loop
func (s *ChatService) processAgenticChat(jobID string, conversationID string, providerID int64, projectID int64) {
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

	// Get agent tools
	agentTools := llm.GetAgentTools()

	// Build system prompt with project context
	systemPrompt := s.buildSystemPrompt(projectID)

	// Agentic loop
	for iteration := 0; iteration < defaultMaxIterations; iteration++ {
		s.chatJobManager.SetIteration(jobID, iteration)
		fmt.Printf("[Agent] Iteration %d starting\n", iteration)

		// Get current message history
		messages, err := s.conversationManager.GetLangChainMessages(ctx, conversationID)
		if err != nil {
			s.chatJobManager.FailJob(jobID, fmt.Sprintf("failed to get messages: %v", err))
			return
		}

		// Prepend system message
		messagesWithSystem := append([]llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		}, messages...)

		// Mark job as tool calling
		s.chatJobManager.SetToolCalling(jobID)

		fmt.Printf("[Agent] Calling LLM with %d tools\n", len(agentTools))

		// Call LLM with tools
		response, err := llmClient.GenerateContent(
			ctx,
			messagesWithSystem,
			llms.WithTools(agentTools),
		)
		if err != nil {
			s.chatJobManager.FailJob(jobID, fmt.Sprintf("LLM error: %v", err))
			return
		}

		// Check for empty response
		if len(response.Choices) == 0 {
			s.chatJobManager.FailJob(jobID, "LLM returned no choices")
			return
		}

		// Log all choices to understand the response structure
		fmt.Printf("[Agent] Response received - Total choices: %d\n", len(response.Choices))

		// Collect all tool calls from all choices
		var allToolCalls []llms.ToolCall
		for i, c := range response.Choices {
			fmt.Printf("[Agent] Choice[%d]: ToolCalls=%d, ContentLen=%d, StopReason=%s\n",
				i, len(c.ToolCalls), len(c.Content), c.StopReason)
			if len(c.ToolCalls) > 0 {
				for j, tc := range c.ToolCalls {
					fmt.Printf("[Agent] Choice[%d].ToolCall[%d]: ID=%s, Name=%s\n", i, j, tc.ID, tc.FunctionCall.Name)
				}
				allToolCalls = append(allToolCalls, c.ToolCalls...)
			}
			if c.Content != "" {
				fmt.Printf("[Agent] Choice[%d] Content preview: %.200s...\n", i, c.Content)
			}
		}

		// Check for tool calls across all choices
		if len(allToolCalls) > 0 {
			// Add assistant message with tool calls to history
			if err := s.conversationManager.AddAssistantToolCalls(ctx, conversationID, allToolCalls); err != nil {
				s.chatJobManager.FailJob(jobID, fmt.Sprintf("failed to add tool calls: %v", err))
				return
			}

			// Execute each tool call
			for _, toolCall := range allToolCalls {
				result, execErr := s.agentExecutor.ExecuteTool(
					ctx,
					projectID,
					toolCall.FunctionCall.Name,
					toolCall.FunctionCall.Arguments,
				)

				// Record tool call
				record := jobs.ToolCallRecord{
					ToolName:   toolCall.FunctionCall.Name,
					Arguments:  toolCall.FunctionCall.Arguments,
					ExecutedAt: time.Now(),
				}

				if execErr != nil {
					record.Success = false
					record.Result = execErr.Error()
					result = fmt.Sprintf(`{"error": "%s"}`, execErr.Error())
				} else {
					record.Success = true
					record.Result = result
				}

				s.chatJobManager.AddToolCall(jobID, record)

				// Add tool result to conversation
				if err := s.conversationManager.AddToolResult(ctx, conversationID, toolCall.ID, toolCall.FunctionCall.Name, result); err != nil {
					s.chatJobManager.FailJob(jobID, fmt.Sprintf("failed to add tool result: %v", err))
					return
				}
			}

			// Continue loop for next iteration
			continue
		}

		// No tool calls - this is the final response
		// Stream the final response
		s.chatJobManager.SetStreaming(jobID)

		// Get updated messages with tool results
		finalMessages, err := s.conversationManager.GetLangChainMessages(ctx, conversationID)
		if err != nil {
			s.chatJobManager.FailJob(jobID, fmt.Sprintf("failed to get final messages: %v", err))
			return
		}

		finalMessagesWithSystem := append([]llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		}, finalMessages...)

		// Create streaming callback
		streamingFunc := func(ctx context.Context, chunk []byte) error {
			s.chatJobManager.AppendResponse(jobID, string(chunk))
			return nil
		}

		// Re-call with streaming for final response
		finalResponse, err := llmClient.GenerateContent(
			ctx,
			finalMessagesWithSystem,
			llms.WithStreamingFunc(streamingFunc),
		)
		if err != nil {
			s.chatJobManager.FailJob(jobID, fmt.Sprintf("streaming error: %v", err))
			return
		}

		// Extract response text
		var responseText string
		if len(finalResponse.Choices) > 0 {
			responseText = finalResponse.Choices[0].Content
		}

		// Get the job to retrieve the accumulated response
		job, _ := s.chatJobManager.GetJob(jobID)
		if job != nil && job.ResponseText != "" {
			responseText = job.ResponseText
		}

		// Estimate token counts
		outputTokens := len(responseText) / 4

		// Add assistant response to conversation
		if err := s.conversationManager.AddAssistantMessage(ctx, conversationID, responseText, outputTokens); err != nil {
			fmt.Printf("Warning: failed to add assistant message to conversation: %v\n", err)
		}

		// Complete the job
		s.chatJobManager.CompleteJob(jobID, 0, outputTokens)
		return
	}

	// Max iterations reached without final response
	s.chatJobManager.FailJob(jobID, "max iterations reached without final response")
}

// buildSystemPrompt creates the system prompt with project context
func (s *ChatService) buildSystemPrompt(projectID int64) string {
	return fmt.Sprintf(`You are an IoT data analysis assistant. You help users analyze time series data through natural language.

Current Project ID: %d

You have access to the following tools:
1. list_layers - Get all data layers in the current project with their metadata
2. get_layer_data - Retrieve actual time series data from a layer for analysis
3. execute_analysis_tool - Run analysis tools that create new visualization layers

Available analysis tools:
- CreateTrend: Creates a linear trend line from the data
- OutlierDetection: Finds anomalies using z-score (threshold parameter adjusts sensitivity)
- MovingAverage: Smooths data with a rolling average (requires window_size)
- MinMax: Normalizes values to 0-1 range
- StandardDeviation: Calculates rolling standard deviation
- Smoothing: Applies smoothing algorithms (sma, ema, gaussian methods)
- Derivative: Calculates rate of change

When users ask to analyze data or create visualizations:
1. First use list_layers to see what data is available
2. Use get_layer_data if you need to examine actual values to answer questions
3. Use execute_analysis_tool to create new layers with analysis results

Always explain what you're doing and what the results mean in plain language.
When you create a new layer, let the user know the name and that it will appear in their Layer Manager.`, projectID)
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
