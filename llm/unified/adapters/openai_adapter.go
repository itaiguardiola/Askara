package adapters

import (
	"context"
	"time"

	"github.com/itaiguardiola/Askara/llm"
	"github.com/itaiguardiola/Askara/llm/unified"
)

// OpenAIAdapter wraps the existing OpenAIProvider to implement UnifiedProvider
// This allows the legacy OpenAI provider to work with the new unified interface
// without requiring a rewrite
type OpenAIAdapter struct {
	provider *llm.OpenAIProvider
}

// NewOpenAIAdapter creates a new adapter for the OpenAI provider
func NewOpenAIAdapter(provider *llm.OpenAIProvider) *OpenAIAdapter {
	return &OpenAIAdapter{provider: provider}
}

// GetName returns the provider name
func (a *OpenAIAdapter) GetName() string {
	return "openai"
}

// GetCapabilities returns OpenAI's capabilities
func (a *OpenAIAdapter) GetCapabilities() []unified.Capability {
	return []unified.Capability{
		unified.CapabilityEmbedding,
		unified.CapabilityCompletion,
		unified.CapabilityStreaming,
		unified.CapabilityVision,       // GPT-4V
		unified.CapabilityFunctionCall, // Function calling
		unified.CapabilityJSON,         // JSON mode
		unified.CapabilityTools,        // Tools/plugins
	}
}

// GetPricingModel returns OpenAI's pricing (as of Jan 2025)
func (a *OpenAIAdapter) GetPricingModel() unified.PricingModel {
	return unified.PricingModel{
		Type:                 unified.PricingPayPerUse,
		EmbeddingPer1KTokens: 0.0001, // ada-002
		InputPer1MTokens:     10.0,   // GPT-4 Turbo
		OutputPer1MTokens:    30.0,   // GPT-4 Turbo
	}
}

// GetRateLimits returns OpenAI's rate limits
func (a *OpenAIAdapter) GetRateLimits() unified.RateLimits {
	return unified.RateLimits{
		RequestsPerMinute:  500,
		TokensPerMinute:    90000,
		TokensPerDay:       -1, // No daily limit
		ConcurrentRequests: 50,
	}
}

// GenerateEmbedding adapts the legacy embedding method to the unified interface
func (a *OpenAIAdapter) GenerateEmbedding(ctx context.Context, req *unified.EmbeddingRequest) (*unified.EmbeddingResponse, error) {
	start := time.Now()

	// Call legacy provider
	embedding, err := a.provider.GenerateEmbedding(req.Text)
	if err != nil {
		return nil, err
	}

	duration := time.Since(start)

	// Estimate tokens (rough: 4 chars per token)
	tokensUsed := len(req.Text) / 4

	// Calculate cost (OpenAI ada-002: $0.0001 per 1k tokens)
	costUSD := float64(tokensUsed) * 0.0001 / 1000

	return &unified.EmbeddingResponse{
		Embedding:     embedding,
		DimensionSize: len(embedding),
		DurationMs:    duration.Milliseconds(),
		TokensUsed:    tokensUsed,
		CostUSD:       costUSD,
		Model:         "text-embedding-ada-002",
		Metadata: map[string]interface{}{
			"provider": "openai",
			"adapter":  true,
		},
	}, nil
}

// GenerateCompletion adapts the legacy completion method to the unified interface
func (a *OpenAIAdapter) GenerateCompletion(ctx context.Context, req *unified.CompletionRequest) (*unified.CompletionResponse, error) {
	start := time.Now()

	// Call legacy provider
	text, err := a.provider.GenerateCompletion(req.Prompt, req.Context)
	if err != nil {
		return nil, err
	}

	duration := time.Since(start)

	// Estimate tokens
	inputTokens := len(req.Prompt) / 4
	for _, ctx := range req.Context {
		inputTokens += len(ctx) / 4
	}
	outputTokens := len(text) / 4

	// Calculate cost (GPT-4 Turbo pricing)
	costUSD := (float64(inputTokens) * 10.0 / 1000000) +
		(float64(outputTokens) * 30.0 / 1000000)

	return &unified.CompletionResponse{
		Text:         text,
		DurationMs:   duration.Milliseconds(),
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		CostUSD:      costUSD,
		Model:        "gpt-4-turbo",
		FinishReason: "stop",
		Metadata: map[string]interface{}{
			"provider": "openai",
			"adapter":  true,
		},
	}, nil
}

// StreamCompletion adapts the legacy streaming method to the unified interface
func (a *OpenAIAdapter) StreamCompletion(ctx context.Context, req *unified.CompletionRequest, onChunk func(string)) error {
	// Call legacy provider's streaming method
	return a.provider.StreamCompletion(req.Prompt, req.Context, onChunk)
}

// HealthCheck performs a simple health check
func (a *OpenAIAdapter) HealthCheck(ctx context.Context) error {
	// Simple health check: try to generate an embedding for a test string
	_, err := a.provider.GenerateEmbedding("test")
	return err
}
