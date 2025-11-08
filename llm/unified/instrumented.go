package unified

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/itaiguardiola/Askara/metrics"
)

// InstrumentedProvider wraps any UnifiedProvider with automatic metrics collection
// This ensures all providers are measured consistently without code duplication
type InstrumentedProvider struct {
	provider  UnifiedProvider
	collector metrics.MetricsCollector
}

// NewInstrumentedProvider creates a provider wrapper that auto-collects metrics
func NewInstrumentedProvider(provider UnifiedProvider, collector metrics.MetricsCollector) *InstrumentedProvider {
	return &InstrumentedProvider{
		provider:  provider,
		collector: collector,
	}
}

// GetName returns the provider name
func (ip *InstrumentedProvider) GetName() string {
	return ip.provider.GetName()
}

// GetCapabilities returns the provider capabilities
func (ip *InstrumentedProvider) GetCapabilities() []Capability {
	return ip.provider.GetCapabilities()
}

// GetPricingModel returns the pricing model
func (ip *InstrumentedProvider) GetPricingModel() PricingModel {
	return ip.provider.GetPricingModel()
}

// GetRateLimits returns the rate limits
func (ip *InstrumentedProvider) GetRateLimits() RateLimits {
	return ip.provider.GetRateLimits()
}

// GenerateEmbedding generates embeddings with automatic metrics collection
func (ip *InstrumentedProvider) GenerateEmbedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	taskID := uuid.New().String()
	start := time.Now()

	// Call actual provider
	resp, err := ip.provider.GenerateEmbedding(ctx, req)

	duration := time.Since(start)

	// Record metrics (if collector is not nil/noop)
	if ip.collector != nil {
		metric := metrics.TaskMetric{
			TaskID:     taskID,
			TaskType:   metrics.TaskTypeEmbedding,
			Provider:   ip.provider.GetName(),
			Timestamp:  start,
			DurationMs: duration.Milliseconds(),
			Success:    err == nil,
			UserUUID:   req.UserUUID,
		}

		if err == nil {
			metric.InputTokens = resp.TokensUsed
			metric.CostUSD = resp.CostUSD
			metric.InputSize = int64(len(req.Text))
			metric.OutputSize = int64(len(resp.Embedding) * 4) // 4 bytes per float32

			// Add metadata
			metric.Metadata = map[string]interface{}{
				"model":          resp.Model,
				"dimension_size": resp.DimensionSize,
			}
		} else {
			metric.ErrorType = err.Error()
		}

		ip.collector.RecordTask(metric)
	}

	return resp, err
}

// GenerateCompletion generates completions with automatic metrics collection
func (ip *InstrumentedProvider) GenerateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	taskID := uuid.New().String()
	start := time.Now()

	// Call actual provider
	resp, err := ip.provider.GenerateCompletion(ctx, req)

	duration := time.Since(start)

	// Record metrics
	if ip.collector != nil {
		metric := metrics.TaskMetric{
			TaskID:     taskID,
			TaskType:   metrics.TaskTypeCompletion,
			Provider:   ip.provider.GetName(),
			Timestamp:  start,
			DurationMs: duration.Milliseconds(),
			Success:    err == nil,
			UserUUID:   req.UserUUID,
		}

		if err == nil {
			metric.InputTokens = resp.InputTokens
			metric.OutputTokens = resp.OutputTokens
			metric.CostUSD = resp.CostUSD

			// Add metadata
			metric.Metadata = map[string]interface{}{
				"model":         resp.Model,
				"finish_reason": resp.FinishReason,
				"max_tokens":    req.MaxTokens,
				"temperature":   req.Temperature,
			}
		} else {
			metric.ErrorType = err.Error()
		}

		ip.collector.RecordTask(metric)
	}

	return resp, err
}

// StreamCompletion streams completions with automatic metrics collection
func (ip *InstrumentedProvider) StreamCompletion(ctx context.Context, req *CompletionRequest, onChunk func(string)) error {
	taskID := uuid.New().String()
	start := time.Now()

	var firstTokenTime *time.Time
	var totalChunks int
	var totalChars int

	// Wrap callback to track time-to-first-token
	wrappedCallback := func(chunk string) {
		if firstTokenTime == nil {
			now := time.Now()
			firstTokenTime = &now
		}
		totalChunks++
		totalChars += len(chunk)
		onChunk(chunk)
	}

	// Call actual provider
	err := ip.provider.StreamCompletion(ctx, req, wrappedCallback)

	duration := time.Since(start)

	// Record metrics
	if ip.collector != nil {
		metric := metrics.TaskMetric{
			TaskID:     taskID,
			TaskType:   metrics.TaskTypeCompletion,
			Provider:   ip.provider.GetName(),
			Timestamp:  start,
			DurationMs: duration.Milliseconds(),
			Success:    err == nil,
			UserUUID:   req.UserUUID,
		}

		// Streaming-specific metadata
		metadataMap := map[string]interface{}{
			"streaming":    true,
			"total_chunks": totalChunks,
			"total_chars":  totalChars,
		}

		if firstTokenTime != nil {
			ttft := firstTokenTime.Sub(start).Milliseconds()
			metadataMap["time_to_first_token_ms"] = ttft
		}

		metric.Metadata = metadataMap

		if err != nil {
			metric.ErrorType = err.Error()
		}

		ip.collector.RecordTask(metric)
	}

	return err
}

// HealthCheck performs health check with metrics
func (ip *InstrumentedProvider) HealthCheck(ctx context.Context) error {
	start := time.Now()
	err := ip.provider.HealthCheck(ctx)
	duration := time.Since(start)

	// Record health check as a metric
	if ip.collector != nil {
		metric := metrics.TaskMetric{
			TaskID:     uuid.New().String(),
			TaskType:   "health_check",
			Provider:   ip.provider.GetName(),
			Timestamp:  start,
			DurationMs: duration.Milliseconds(),
			Success:    err == nil,
		}

		if err != nil {
			metric.ErrorType = err.Error()
		}

		ip.collector.RecordTask(metric)
	}

	return err
}
