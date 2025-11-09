package adapters

import (
	"context"
	"time"

	"github.com/itaiguardiola/Askara/llm"
	"github.com/itaiguardiola/Askara/llm/unified"
)

// OllamaAdapter wraps the existing OllamaProvider to implement UnifiedProvider
// Allows the legacy Ollama provider to work with the unified interface
type OllamaAdapter struct {
	provider *llm.OllamaProvider
}

// NewOllamaAdapter creates a new adapter for the Ollama provider
func NewOllamaAdapter(provider *llm.OllamaProvider) *OllamaAdapter {
	return &OllamaAdapter{provider: provider}
}

// GetName returns the provider name
func (a *OllamaAdapter) GetName() string {
	return "ollama"
}

// GetCapabilities returns Ollama's capabilities
func (a *OllamaAdapter) GetCapabilities() []unified.Capability {
	return []unified.Capability{
		unified.CapabilityEmbedding,
		unified.CapabilityCompletion,
		unified.CapabilityStreaming,
		// Ollama supports various models, some with vision
	}
}

// GetPricingModel returns Ollama's pricing (free, self-hosted)
func (a *OllamaAdapter) GetPricingModel() unified.PricingModel {
	return unified.PricingModel{
		Type:                 unified.PricingFree,
		EmbeddingPer1KTokens: 0.0, // Free (self-hosted)
		InputPer1MTokens:     0.0,
		OutputPer1MTokens:    0.0,
	}
}

// GetRateLimits returns Ollama's rate limits
func (a *OllamaAdapter) GetRateLimits() unified.RateLimits {
	return unified.RateLimits{
		RequestsPerMinute:  -1, // No limit (self-hosted)
		TokensPerMinute:    -1, // Limited by GPU/CPU only
		TokensPerDay:       -1,
		ConcurrentRequests: 4, // Typically limited by GPU memory
	}
}

// GenerateEmbedding adapts the legacy embedding method to the unified interface
func (a *OllamaAdapter) GenerateEmbedding(ctx context.Context, req *unified.EmbeddingRequest) (*unified.EmbeddingResponse, error) {
	start := time.Now()

	// Call legacy provider
	embedding, err := a.provider.GenerateEmbedding(req.Text)
	if err != nil {
		return nil, err
	}

	duration := time.Since(start)

	// Estimate tokens
	tokensUsed := len(req.Text) / 4

	return &unified.EmbeddingResponse{
		Embedding:     embedding,
		DimensionSize: len(embedding),
		DurationMs:    duration.Milliseconds(),
		TokensUsed:    tokensUsed,
		CostUSD:       0.0, // Free!
		Model:         "nomic-embed-text",
		Metadata: map[string]interface{}{
			"provider": "ollama",
			"adapter":  true,
			"local":    true,
		},
	}, nil
}

// GenerateCompletion adapts the legacy completion method to the unified interface
func (a *OllamaAdapter) GenerateCompletion(ctx context.Context, req *unified.CompletionRequest) (*unified.CompletionResponse, error) {
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

	return &unified.CompletionResponse{
		Text:         text,
		DurationMs:   duration.Milliseconds(),
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		CostUSD:      0.0, // Free!
		Model:        "llama2", // Or detect from config
		FinishReason: "stop",
		Metadata: map[string]interface{}{
			"provider": "ollama",
			"adapter":  true,
			"local":    true,
		},
	}, nil
}

// StreamCompletion adapts the legacy streaming method to the unified interface
func (a *OllamaAdapter) StreamCompletion(ctx context.Context, req *unified.CompletionRequest, onChunk func(string)) error {
	// Call legacy provider's streaming method
	return a.provider.StreamCompletion(req.Prompt, req.Context, onChunk)
}

// HealthCheck performs a simple health check
func (a *OllamaAdapter) HealthCheck(ctx context.Context) error {
	// Simple health check: try to generate an embedding for a test string
	_, err := a.provider.GenerateEmbedding("test")
	return err
}
