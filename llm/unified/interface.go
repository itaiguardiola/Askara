package unified

import (
	"context"
	"time"
)

// UnifiedProvider is the interface that all AI providers must implement
// This allows seamless integration of OpenAI, Gemini, Anthropic, Ollama, etc.
type UnifiedProvider interface {
	// Metadata
	GetName() string
	GetCapabilities() []Capability
	GetPricingModel() PricingModel
	GetRateLimits() RateLimits

	// Core operations (auto-instrumented when wrapped)
	GenerateEmbedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)
	GenerateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
	StreamCompletion(ctx context.Context, req *CompletionRequest, onChunk func(string)) error

	// Health
	HealthCheck(ctx context.Context) error
}

// EmbeddingRequest represents a request to generate embeddings
type EmbeddingRequest struct {
	Text     string
	Model    string
	UserUUID string
	Metadata map[string]interface{}
}

// EmbeddingResponse contains the embedding and performance metrics
type EmbeddingResponse struct {
	Embedding     []float32
	DimensionSize int

	// Performance metrics (populated by provider)
	DurationMs int64
	TokensUsed int
	CostUSD    float64

	// Provider-specific metadata
	Model    string
	Metadata map[string]interface{}
}

// CompletionRequest represents a request for text completion
type CompletionRequest struct {
	Prompt      string
	Context     []string
	MaxTokens   int
	Temperature float32
	Model       string
	UserUUID    string
	Streaming   bool
	Metadata    map[string]interface{}
}

// CompletionResponse contains the completion and performance metrics
type CompletionResponse struct {
	Text string

	// Performance metrics
	DurationMs       int64
	TimeToFirstToken int64 // For streaming (milliseconds)
	InputTokens      int
	OutputTokens     int
	CostUSD          float64

	// Provider-specific metadata
	Model         string
	FinishReason  string
	Metadata      map[string]interface{}
}

// Capability represents provider features
type Capability string

const (
	CapabilityEmbedding    Capability = "embedding"
	CapabilityCompletion   Capability = "completion"
	CapabilityStreaming    Capability = "streaming"
	CapabilityVision       Capability = "vision"
	CapabilityFunctionCall Capability = "function_calling"
	CapabilityLongContext  Capability = "long_context" // >32k tokens
	CapabilityJSON         Capability = "json_mode"
	CapabilityTools        Capability = "tools"
)

// PricingModel describes the cost structure of a provider
type PricingModel struct {
	Type                 PricingType
	EmbeddingPer1KTokens float64
	InputPer1MTokens     float64
	OutputPer1MTokens    float64
	FixedPerRequest      float64
}

// PricingType indicates how the provider charges
type PricingType string

const (
	PricingFree         PricingType = "free"         // Ollama (local)
	PricingPayPerUse    PricingType = "pay_per_use"  // OpenAI, Gemini, Anthropic
	PricingSubscription PricingType = "subscription" // Fixed monthly cost
)

// RateLimits describes the provider's rate limiting
type RateLimits struct {
	RequestsPerMinute int
	TokensPerMinute   int
	TokensPerDay      int
	ConcurrentRequests int
}

// ProviderConfig contains common configuration for providers
type ProviderConfig struct {
	APIKey      string
	Endpoint    string
	Model       string
	Timeout     time.Duration
	MaxRetries  int
	RetryDelay  time.Duration
}
