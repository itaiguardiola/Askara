package llm

import (
	"fmt"
	"log"
	"time"
)

// InstrumentedProvider wraps an LLMProvider with metrics collection
type InstrumentedProvider struct {
	provider  LLMProvider
	providerType ProviderType
	collector *MetricsCollector
	model     string
}

// NewInstrumentedProvider creates a new instrumented provider
func NewInstrumentedProvider(provider LLMProvider, providerType ProviderType, collector *MetricsCollector, model string) *InstrumentedProvider {
	return &InstrumentedProvider{
		provider:     provider,
		providerType: providerType,
		collector:    collector,
		model:        model,
	}
}

// GenerateEmbedding generates embeddings with metrics tracking
func (ip *InstrumentedProvider) GenerateEmbedding(text string) ([]float32, error) {
	start := time.Now()

	embedding, err := ip.provider.GenerateEmbedding(text)

	// Record metric
	metric := ProviderMetric{
		Provider:     ip.providerType,
		MetricType:   MetricTypeEmbedding,
		Timestamp:    time.Now(),
		LatencyMs:    time.Since(start).Milliseconds(),
		Success:      err == nil,
		Model:        ip.model,
		TokensInput:  len(text) / 4, // Rough estimate
	}

	if err != nil {
		metric.ErrorMessage = err.Error()
	} else {
		metric.CostEstimate = ip.estimateEmbeddingCost(text)
	}

	ip.collector.RecordMetric(metric)

	return embedding, err
}

// GenerateCompletion generates completions with metrics tracking
func (ip *InstrumentedProvider) GenerateCompletion(prompt string, context []string) (string, error) {
	start := time.Now()

	response, err := ip.provider.GenerateCompletion(prompt, context)

	// Estimate tokens
	inputTokens := ip.estimateTokens(prompt, context)
	outputTokens := len(response) / 4 // Rough estimate

	// Record metric
	metric := ProviderMetric{
		Provider:     ip.providerType,
		MetricType:   MetricTypeCompletion,
		Timestamp:    time.Now(),
		LatencyMs:    time.Since(start).Milliseconds(),
		Success:      err == nil,
		TokensInput:  inputTokens,
		TokensOutput: outputTokens,
		Model:        ip.model,
	}

	if err != nil {
		metric.ErrorMessage = err.Error()
	} else {
		metric.CostEstimate = ip.estimateCompletionCost(inputTokens, outputTokens)
	}

	ip.collector.RecordMetric(metric)

	return response, err
}

// StreamCompletion generates streaming completions with metrics tracking
func (ip *InstrumentedProvider) StreamCompletion(prompt string, context []string, onChunk func(string)) error {
	start := time.Now()
	totalChunks := 0
	totalOutput := ""

	// Wrap the chunk handler to track output
	wrappedHandler := func(chunk string) {
		totalChunks++
		totalOutput += chunk
		onChunk(chunk)
	}

	err := ip.provider.StreamCompletion(prompt, context, wrappedHandler)

	// Estimate tokens
	inputTokens := ip.estimateTokens(prompt, context)
	outputTokens := len(totalOutput) / 4 // Rough estimate

	// Record metric
	metric := ProviderMetric{
		Provider:     ip.providerType,
		MetricType:   MetricTypeStreaming,
		Timestamp:    time.Now(),
		LatencyMs:    time.Since(start).Milliseconds(),
		Success:      err == nil,
		TokensInput:  inputTokens,
		TokensOutput: outputTokens,
		Model:        ip.model,
	}

	if err != nil {
		metric.ErrorMessage = err.Error()
	} else {
		metric.CostEstimate = ip.estimateCompletionCost(inputTokens, outputTokens)
	}

	ip.collector.RecordMetric(metric)

	return err
}

// estimateTokens estimates token count for prompt and context
func (ip *InstrumentedProvider) estimateTokens(prompt string, context []string) int {
	total := len(prompt) / 4 // Rough estimate: 1 token ≈ 4 chars

	for _, ctx := range context {
		total += len(ctx) / 4
	}

	return total
}

// estimateEmbeddingCost estimates the cost of an embedding request
func (ip *InstrumentedProvider) estimateEmbeddingCost(text string) float64 {
	tokens := len(text) / 4

	switch ip.providerType {
	case ProviderTypeOpenAI:
		// text-embedding-ada-002: $0.0001 per 1K tokens
		return float64(tokens) / 1000.0 * 0.0001
	case ProviderTypeClaude:
		// Voyage embeddings: ~$0.0001 per 1K tokens
		return float64(tokens) / 1000.0 * 0.0001
	case ProviderTypeGemini:
		// Gemini embeddings: $0.00001 per 1K tokens
		return float64(tokens) / 1000.0 * 0.00001
	case ProviderTypeOllama, ProviderTypeGroq:
		// Free/local
		return 0.0
	default:
		return 0.0
	}
}

// estimateCompletionCost estimates the cost of a completion request
func (ip *InstrumentedProvider) estimateCompletionCost(inputTokens, outputTokens int) float64 {
	switch ip.providerType {
	case ProviderTypeOpenAI:
		// GPT-4: $0.03 input, $0.06 output per 1K tokens
		// GPT-3.5: $0.0015 input, $0.002 output per 1K tokens
		// Using GPT-4 pricing as default
		inputCost := float64(inputTokens) / 1000.0 * 0.03
		outputCost := float64(outputTokens) / 1000.0 * 0.06
		return inputCost + outputCost

	case ProviderTypeClaude:
		// Claude 3 Sonnet: $0.003 input, $0.015 output per 1K tokens
		inputCost := float64(inputTokens) / 1000.0 * 0.003
		outputCost := float64(outputTokens) / 1000.0 * 0.015
		return inputCost + outputCost

	case ProviderTypeGemini:
		// Gemini Pro: $0.00025 input, $0.0005 output per 1K tokens
		inputCost := float64(inputTokens) / 1000.0 * 0.00025
		outputCost := float64(outputTokens) / 1000.0 * 0.0005
		return inputCost + outputCost

	case ProviderTypeGroq:
		// Groq is free tier pricing - nominal
		return 0.0

	case ProviderTypeOllama:
		// Local, no cost
		return 0.0

	default:
		return 0.0
	}
}

// AutoSelector automatically selects the best provider based on metrics
type AutoSelector struct {
	providers  map[ProviderType]*InstrumentedProvider
	collector  *MetricsCollector
	fallbackOrder []ProviderType
	weights    SelectorWeights
}

// SelectorWeights defines the weights for auto-selection criteria
type SelectorWeights struct {
	HealthScore   float64 // 0-1
	LatencyScore  float64 // 0-1
	CostScore     float64 // 0-1
	SuccessRate   float64 // 0-1
}

// DefaultSelectorWeights returns sensible default weights
func DefaultSelectorWeights() SelectorWeights {
	return SelectorWeights{
		HealthScore:  0.40, // 40% weight on health
		LatencyScore: 0.30, // 30% weight on speed
		CostScore:    0.10, // 10% weight on cost
		SuccessRate:  0.20, // 20% weight on reliability
	}
}

// NewAutoSelector creates a new auto-selector
func NewAutoSelector(collector *MetricsCollector, weights SelectorWeights) *AutoSelector {
	return &AutoSelector{
		providers:     make(map[ProviderType]*InstrumentedProvider),
		collector:     collector,
		fallbackOrder: []ProviderType{},
		weights:       weights,
	}
}

// RegisterProvider registers a provider with the auto-selector
func (as *AutoSelector) RegisterProvider(providerType ProviderType, provider *InstrumentedProvider) {
	as.providers[providerType] = provider
	as.fallbackOrder = append(as.fallbackOrder, providerType)
	log.Printf("[AutoSelector] Registered provider: %s", providerType)
}

// SelectBestProvider selects the best provider based on current metrics
func (as *AutoSelector) SelectBestProvider() (*InstrumentedProvider, ProviderType, error) {
	if len(as.providers) == 0 {
		return nil, "", fmt.Errorf("no providers registered")
	}

	allStats := as.collector.GetAllStats()

	var bestProvider ProviderType
	var bestScore float64 = -1

	for providerType, provider := range as.providers {
		stats, exists := allStats[providerType]

		// If no stats exist, give it a neutral score
		if !exists || stats.TotalRequests == 0 {
			score := 50.0 // Neutral score for untested providers
			log.Printf("[AutoSelector] Provider %s: no metrics, neutral score %.2f", providerType, score)

			if score > bestScore {
				bestScore = score
				bestProvider = providerType
			}
			continue
		}

		// Skip unhealthy providers
		if !stats.IsHealthy {
			log.Printf("[AutoSelector] Skipping unhealthy provider: %s (health score: %.2f)", providerType, stats.HealthScore)
			continue
		}

		// Calculate weighted score
		score := as.calculateProviderScore(stats)

		log.Printf("[AutoSelector] Provider %s: score %.2f (health: %.2f, latency: %.2fms, success: %.2f%%, cost: $%.4f)",
			providerType, score, stats.HealthScore, stats.AvgLatencyMs, stats.SuccessRate, stats.TotalCost)

		if score > bestScore {
			bestScore = score
			bestProvider = providerType
		}
	}

	// If no healthy provider found, use fallback order
	if bestProvider == "" {
		log.Printf("[AutoSelector] No healthy provider found, using fallback")
		for _, providerType := range as.fallbackOrder {
			if _, exists := as.providers[providerType]; exists {
				bestProvider = providerType
				break
			}
		}
	}

	if bestProvider == "" {
		return nil, "", fmt.Errorf("no suitable provider found")
	}

	selectedProvider := as.providers[bestProvider]
	log.Printf("[AutoSelector] Selected provider: %s (score: %.2f)", bestProvider, bestScore)

	return selectedProvider, bestProvider, nil
}

// calculateProviderScore calculates a weighted score for a provider
func (as *AutoSelector) calculateProviderScore(stats *ProviderStats) float64 {
	score := 0.0

	// Health score (0-100) -> normalize to 0-100
	score += stats.HealthScore * as.weights.HealthScore

	// Success rate (0-100) -> already 0-100
	score += stats.SuccessRate * as.weights.SuccessRate

	// Latency score - lower is better (invert)
	// Assume: 100ms = 100 points, 5000ms = 0 points
	latencyScore := 100.0
	if stats.AvgLatencyMs > 100 {
		latencyScore = 100.0 * (1.0 - min((stats.AvgLatencyMs-100)/4900, 1.0))
	}
	score += latencyScore * as.weights.LatencyScore

	// Cost score - lower is better (invert)
	// Normalize cost to 0-100 scale (lower cost = higher score)
	// Assume: $0 = 100 points, $1+ = 0 points
	costScore := 100.0
	if stats.TotalCost > 0 {
		costScore = 100.0 * (1.0 - min(stats.TotalCost, 1.0))
	}
	score += costScore * as.weights.CostScore

	return score
}

// GetProvider returns a specific provider
func (as *AutoSelector) GetProvider(providerType ProviderType) (*InstrumentedProvider, error) {
	provider, exists := as.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerType)
	}
	return provider, nil
}

// GetAllProviders returns all registered providers
func (as *AutoSelector) GetAllProviders() map[ProviderType]*InstrumentedProvider {
	return as.providers
}
