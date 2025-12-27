package llm

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
)

// ChatMessage represents a single message in a conversation
type ChatMessage struct {
	Role       string    `json:"role"` // "user", "assistant", "system"
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
	TokenCount int       `json:"token_count,omitempty"`
}

// Conversation represents a chat session
type Conversation struct {
	ConversationID string        `json:"conversation_id"`
	Messages       []ChatMessage `json:"messages"`
	TotalTokens    int           `json:"total_tokens"`
	MaxTokens      int           `json:"max_tokens"`
	ProviderID     int64         `json:"provider_id"`
	Model          string        `json:"model,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// ActiveConversation holds a conversation with its LangChain memory buffer
type ActiveConversation struct {
	Conversation *Conversation
	Memory       *memory.ConversationBuffer
}

// ConversationManager manages active conversations with memory buffers
type ConversationManager struct {
	conversations map[string]*ActiveConversation
	mu            sync.RWMutex
}

// NewConversationManager creates a new conversation manager
func NewConversationManager() *ConversationManager {
	return &ConversationManager{
		conversations: make(map[string]*ActiveConversation),
	}
}

// CreateConversation creates a new conversation with memory
func (m *ConversationManager) CreateConversation(providerID int64, maxTokens int) *Conversation {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	conv := &Conversation{
		ConversationID: uuid.New().String(),
		Messages:       []ChatMessage{},
		TotalTokens:    0,
		MaxTokens:      maxTokens,
		ProviderID:     providerID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Create LangChain conversation buffer
	chatMemory := memory.NewConversationBuffer()

	m.conversations[conv.ConversationID] = &ActiveConversation{
		Conversation: conv,
		Memory:       chatMemory,
	}

	return conv
}

// GetConversation retrieves an active conversation by ID
func (m *ConversationManager) GetConversation(id string) (*ActiveConversation, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conv, exists := m.conversations[id]
	return conv, exists
}

// DeleteConversation removes a conversation from the manager
func (m *ConversationManager) DeleteConversation(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.conversations, id)
}

// AddUserMessage adds a user message to the conversation
func (m *ConversationManager) AddUserMessage(ctx context.Context, conversationID string, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	active, exists := m.conversations[conversationID]
	if !exists {
		return nil // Conversation not found
	}

	// Add to local message history
	msg := ChatMessage{
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	}
	active.Conversation.Messages = append(active.Conversation.Messages, msg)
	active.Conversation.UpdatedAt = time.Now()

	// Add to LangChain memory
	return active.Memory.ChatHistory.AddUserMessage(ctx, content)
}

// AddAssistantMessage adds an assistant message to the conversation
func (m *ConversationManager) AddAssistantMessage(ctx context.Context, conversationID string, content string, tokenCount int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	active, exists := m.conversations[conversationID]
	if !exists {
		return nil // Conversation not found
	}

	// Add to local message history
	msg := ChatMessage{
		Role:       "assistant",
		Content:    content,
		Timestamp:  time.Now(),
		TokenCount: tokenCount,
	}
	active.Conversation.Messages = append(active.Conversation.Messages, msg)
	active.Conversation.TotalTokens += tokenCount
	active.Conversation.UpdatedAt = time.Now()

	// Add to LangChain memory
	return active.Memory.ChatHistory.AddAIMessage(ctx, content)
}

// GetMessageHistory returns the full message history for a conversation
func (m *ConversationManager) GetMessageHistory(conversationID string) []ChatMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active, exists := m.conversations[conversationID]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions
	messages := make([]ChatMessage, len(active.Conversation.Messages))
	copy(messages, active.Conversation.Messages)
	return messages
}

// GetLangChainMessages returns the LangChain message format for LLM calls
func (m *ConversationManager) GetLangChainMessages(ctx context.Context, conversationID string) ([]llms.MessageContent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active, exists := m.conversations[conversationID]
	if !exists {
		return nil, nil
	}

	// Get messages from LangChain memory
	history, err := active.Memory.ChatHistory.Messages(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to MessageContent format
	messages := make([]llms.MessageContent, len(history))
	for i, msg := range history {
		role := llms.ChatMessageTypeHuman
		if msg.GetType() == llms.ChatMessageTypeAI {
			role = llms.ChatMessageTypeAI
		}
		messages[i] = llms.TextParts(role, msg.GetContent())
	}

	return messages, nil
}

// ClearHistory clears the message history for a conversation
func (m *ConversationManager) ClearHistory(ctx context.Context, conversationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	active, exists := m.conversations[conversationID]
	if !exists {
		return nil
	}

	active.Conversation.Messages = []ChatMessage{}
	active.Conversation.TotalTokens = 0
	active.Conversation.UpdatedAt = time.Now()

	return active.Memory.ChatHistory.Clear(ctx)
}

// GetTotalTokens returns the total token count for a conversation
func (m *ConversationManager) GetTotalTokens(conversationID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active, exists := m.conversations[conversationID]
	if !exists {
		return 0
	}

	return active.Conversation.TotalTokens
}

// CleanupOldConversations removes conversations older than the specified duration
func (m *ConversationManager) CleanupOldConversations(olderThan time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	for id, active := range m.conversations {
		if active.Conversation.UpdatedAt.Before(cutoff) {
			delete(m.conversations, id)
		}
	}
}
