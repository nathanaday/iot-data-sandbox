package llm

import (
	"context"
	"fmt"

	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
)

// CreateLLMClient creates an LLM client from a provider configuration
func CreateLLMClient(ctx context.Context, provider *models.LLMProvider) (llms.Model, error) {
	switch provider.ProviderType {
	case models.LLMProviderOpenAI:
		return createOpenAIClient(provider)
	case models.LLMProviderAzureOpenAI:
		return createAzureOpenAIClient(provider)
	case models.LLMProviderAnthropic:
		return createAnthropicClient(provider)
	case models.LLMProviderGoogleAI:
		return createGoogleAIClient(ctx, provider)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", provider.ProviderType)
	}
}

func createOpenAIClient(provider *models.LLMProvider) (llms.Model, error) {
	opts := []openai.Option{
		openai.WithToken(provider.APIKey),
	}

	if provider.BaseURL != nil && *provider.BaseURL != "" {
		opts = append(opts, openai.WithBaseURL(*provider.BaseURL))
	}

	if provider.DefaultModel != nil && *provider.DefaultModel != "" {
		opts = append(opts, openai.WithModel(*provider.DefaultModel))
	}

	return openai.New(opts...)
}

func createAzureOpenAIClient(provider *models.LLMProvider) (llms.Model, error) {
	opts := []openai.Option{
		openai.WithToken(provider.APIKey),
		openai.WithAPIType(openai.APITypeAzure),
	}

	if provider.BaseURL != nil && *provider.BaseURL != "" {
		opts = append(opts, openai.WithBaseURL(*provider.BaseURL))
	}

	if provider.DefaultModel != nil && *provider.DefaultModel != "" {
		opts = append(opts, openai.WithModel(*provider.DefaultModel))
	}

	return openai.New(opts...)
}

func createAnthropicClient(provider *models.LLMProvider) (llms.Model, error) {
	opts := []anthropic.Option{
		anthropic.WithToken(provider.APIKey),
	}

	if provider.DefaultModel != nil && *provider.DefaultModel != "" {
		opts = append(opts, anthropic.WithModel(*provider.DefaultModel))
	}

	return anthropic.New(opts...)
}

func createGoogleAIClient(ctx context.Context, provider *models.LLMProvider) (llms.Model, error) {
	opts := []googleai.Option{
		googleai.WithAPIKey(provider.APIKey),
	}

	if provider.DefaultModel != nil && *provider.DefaultModel != "" {
		opts = append(opts, googleai.WithDefaultModel(*provider.DefaultModel))
	}

	return googleai.New(ctx, opts...)
}
