package models

import "time"

// LLMProviderType represents the type of LLM provider
type LLMProviderType string

const (
	LLMProviderOpenAI      LLMProviderType = "openai"
	LLMProviderAzureOpenAI LLMProviderType = "azure_openai"
	LLMProviderAnthropic   LLMProviderType = "anthropic"
	LLMProviderGoogleAI    LLMProviderType = "google_ai"
	LLMProviderVertexAI    LLMProviderType = "vertex_ai"
	LLMProviderOllama      LLMProviderType = "ollama"
	LLMProviderHuggingFace LLMProviderType = "huggingface"
)

// AllLLMProviderTypes returns all supported LLM provider types
func AllLLMProviderTypes() []LLMProviderType {
	return []LLMProviderType{
		LLMProviderOpenAI,
		LLMProviderAzureOpenAI,
		LLMProviderAnthropic,
		LLMProviderGoogleAI,
		LLMProviderVertexAI,
		LLMProviderOllama,
		LLMProviderHuggingFace,
	}
}

// LLMProvider represents a configured LLM provider with its credentials
type LLMProvider struct {
	LLMProviderId int64
	ProviderType  LLMProviderType
	Name          string // User-defined name for this configuration (e.g., "My OpenAI Key")
	APIKey        string // Encrypted API key (stored encrypted, decrypted when loaded)
	BaseURL       *string // Optional custom endpoint (for Azure OpenAI, custom deployments)
	DefaultModel  *string // Optional default model to use
	IsEnabled     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ProviderDisplayName returns a human-readable name for the provider type
func (p *LLMProvider) ProviderDisplayName() string {
	switch p.ProviderType {
	case LLMProviderOpenAI:
		return "OpenAI"
	case LLMProviderAzureOpenAI:
		return "Azure OpenAI"
	case LLMProviderAnthropic:
		return "Anthropic"
	case LLMProviderGoogleAI:
		return "Google AI (Gemini)"
	case LLMProviderVertexAI:
		return "Google Vertex AI"
	case LLMProviderOllama:
		return "Ollama"
	case LLMProviderHuggingFace:
		return "Hugging Face"
	default:
		return string(p.ProviderType)
	}
}

// RequiresAPIKey returns whether this provider type requires an API key
func (p *LLMProvider) RequiresAPIKey() bool {
	switch p.ProviderType {
	case LLMProviderOllama:
		return false // Ollama typically runs locally without API key
	case LLMProviderVertexAI:
		return false // Vertex AI uses service account credentials, not API key
	default:
		return true
	}
}
