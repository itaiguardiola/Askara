package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// MetricType represents the type of metric being tracked
type MetricType string

const (
	MetricTypeEmbedding   MetricType = "embedding"
	MetricTypeCompletion  MetricType = "completion"
	MetricTypeStreaming   MetricType = "streaming"
)

// ProviderMetric represents a single metric data point
type ProviderMetric struct {
	Provider      ProviderType  `json:"provider"`
	MetricType    MetricType    `json:"metric_type"`
	Timestamp     time.Time     `json:"timestamp"`
	LatencyMs     int64         `json:"latency_ms"`
	Success       bool          `json:"success"`
	ErrorMessage  string        `json:"error_message,omitempty"`
	TokensInput   int           `json:"tokens_input,omitempty"`
	TokensOutput  int           `json:"tokens_output,omitempty"`
	Model         string        `json:"model,omitempty"`
	CostEstimate  float64       `json:"cost_estimate,omitempty"`
}

// ProviderStats represents aggregated statistics for a provider
type ProviderStats struct {
	Provider         ProviderType  `json:"provider"`
	TotalRequests    int64         `json:"total_requests"`
	SuccessfulReqs   int64         `json:"successful_requests"`
	FailedReqs       int64         `json:"failed_requests"`
	SuccessRate      float64       `json:"success_rate"`
	AvgLatencyMs     float64       `json:"avg_latency_ms"`
	P95LatencyMs     float64       `json:"p95_latency_ms"`
	P99LatencyMs     float64       `json:"p99_latency_ms"`
	TotalTokensIn    int64         `json:"total_tokens_in"`
	TotalTokensOut   int64         `json:"total_tokens_out"`
	TotalCost        float64       `json:"total_cost"`
	LastSuccess      *time.Time    `json:"last_success,omitempty"`
	LastFailure      *time.Time    `json:"last_failure,omitempty"`
	IsHealthy        bool          `json:"is_healthy"`
	HealthScore      float64       `json:"health_score"` // 0-100
}

// MetricsCollector collects and aggregates provider metrics
type MetricsCollector struct {
	metrics        []ProviderMetric
	mutex          sync.RWMutex
	maxMetrics     int
	persistencePath string

	// Stats cache
	statsCache     map[ProviderType]*ProviderStats
	statsMutex     sync.RWMutex
	lastStatUpdate time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(maxMetrics int, persistencePath string) *MetricsCollector {
	mc := &MetricsCollector{
		metrics:         make([]ProviderMetric, 0, maxMetrics),
		maxMetrics:      maxMetrics,
		persistencePath: persistencePath,
		statsCache:      make(map[ProviderType]*ProviderStats),
	}

	// Load existing metrics if available
	mc.loadMetrics()

	return mc
}

// RecordMetric records a new metric data point
func (mc *MetricsCollector) RecordMetric(metric ProviderMetric) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// Add metric
	mc.metrics = append(mc.metrics, metric)

	// Trim old metrics if exceeding max
	if len(mc.metrics) > mc.maxMetrics {
		// Remove oldest 20%
		removeCount := mc.maxMetrics / 5
		mc.metrics = mc.metrics[removeCount:]
	}

	// Invalidate stats cache
	mc.statsMutex.Lock()
	mc.statsCache = make(map[ProviderType]*ProviderStats)
	mc.lastStatUpdate = time.Time{}
	mc.statsMutex.Unlock()

	// Persist every 10 metrics
	if len(mc.metrics)%10 == 0 {
		go mc.persistMetrics()
	}
}

// GetStats returns aggregated statistics for a provider
func (mc *MetricsCollector) GetStats(provider ProviderType) *ProviderStats {
	mc.statsMutex.RLock()
	cached, exists := mc.statsCache[provider]
	mc.statsMutex.RUnlock()

	// Return cached if recent (< 30 seconds old)
	if exists && time.Since(mc.lastStatUpdate) < 30*time.Second {
		return cached
	}

	// Calculate fresh stats
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	stats := &ProviderStats{
		Provider: provider,
	}

	var latencies []int64
	var totalLatency int64

	for _, m := range mc.metrics {
		if m.Provider != provider {
			continue
		}

		stats.TotalRequests++
		if m.Success {
			stats.SuccessfulReqs++
			stats.LastSuccess = &m.Timestamp
		} else {
			stats.FailedReqs++
			stats.LastFailure = &m.Timestamp
		}

		latencies = append(latencies, m.LatencyMs)
		totalLatency += m.LatencyMs
		stats.TotalTokensIn += int64(m.TokensInput)
		stats.TotalTokensOut += int64(m.TokensOutput)
		stats.TotalCost += m.CostEstimate
	}

	// Calculate rates and averages
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(stats.SuccessfulReqs) / float64(stats.TotalRequests) * 100
		stats.AvgLatencyMs = float64(totalLatency) / float64(stats.TotalRequests)

		// Calculate percentiles
		if len(latencies) > 0 {
			stats.P95LatencyMs = calculatePercentile(latencies, 95)
			stats.P99LatencyMs = calculatePercentile(latencies, 99)
		}
	}

	// Calculate health
	stats.IsHealthy = mc.checkProviderHealth(provider, stats)
	stats.HealthScore = mc.calculateHealthScore(stats)

	// Cache the result
	mc.statsMutex.Lock()
	mc.statsCache[provider] = stats
	mc.lastStatUpdate = time.Now()
	mc.statsMutex.Unlock()

	return stats
}

// GetAllStats returns statistics for all providers
func (mc *MetricsCollector) GetAllStats() map[ProviderType]*ProviderStats {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	providers := make(map[ProviderType]bool)
	for _, m := range mc.metrics {
		providers[m.Provider] = true
	}

	result := make(map[ProviderType]*ProviderStats)
	for provider := range providers {
		result[provider] = mc.GetStats(provider)
	}

	return result
}

// GetRecentMetrics returns the most recent N metrics
func (mc *MetricsCollector) GetRecentMetrics(count int) []ProviderMetric {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	if count > len(mc.metrics) {
		count = len(mc.metrics)
	}

	result := make([]ProviderMetric, count)
	copy(result, mc.metrics[len(mc.metrics)-count:])

	return result
}

// checkProviderHealth checks if a provider is currently healthy
func (mc *MetricsCollector) checkProviderHealth(provider ProviderType, stats *ProviderStats) bool {
	// Check recent failures (last 5 minutes)
	recentFailures := 0
	recentRequests := 0
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)

	mc.mutex.RLock()
	for i := len(mc.metrics) - 1; i >= 0; i-- {
		m := mc.metrics[i]
		if m.Provider != provider {
			continue
		}
		if m.Timestamp.Before(fiveMinutesAgo) {
			break
		}

		recentRequests++
		if !m.Success {
			recentFailures++
		}
	}
	mc.mutex.RUnlock()

	// If no recent requests, check if we had success in the past
	if recentRequests == 0 {
		return stats.LastSuccess != nil && (stats.LastFailure == nil || stats.LastSuccess.After(*stats.LastFailure))
	}

	// Consider healthy if < 50% failure rate in last 5 minutes
	recentFailureRate := float64(recentFailures) / float64(recentRequests)
	return recentFailureRate < 0.5
}

// calculateHealthScore calculates a 0-100 health score based on multiple factors
func (mc *MetricsCollector) calculateHealthScore(stats *ProviderStats) float64 {
	if stats.TotalRequests == 0 {
		return 50.0 // Neutral score for untested providers
	}

	score := 0.0

	// Success rate (40 points)
	score += stats.SuccessRate * 0.4

	// Latency score (30 points) - lower is better
	// Assume 5000ms is worst (0 points), 100ms is best (30 points)
	latencyScore := 30.0
	if stats.AvgLatencyMs > 100 {
		latencyScore = 30.0 * (1.0 - min((stats.AvgLatencyMs-100)/4900, 1.0))
	}
	score += latencyScore

	// Recency score (30 points) - recent success is important
	if stats.LastSuccess != nil {
		hoursSinceSuccess := time.Since(*stats.LastSuccess).Hours()
		if hoursSinceSuccess < 1 {
			score += 30.0
		} else if hoursSinceSuccess < 24 {
			score += 30.0 * (1.0 - hoursSinceSuccess/24)
		}
	}

	return min(max(score, 0), 100)
}

// persistMetrics saves metrics to disk
func (mc *MetricsCollector) persistMetrics() error {
	if mc.persistencePath == "" {
		return nil
	}

	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	data, err := json.MarshalIndent(mc.metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	if err := os.WriteFile(mc.persistencePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metrics file: %w", err)
	}

	return nil
}

// loadMetrics loads metrics from disk
func (mc *MetricsCollector) loadMetrics() error {
	if mc.persistencePath == "" {
		return nil
	}

	data, err := os.ReadFile(mc.persistencePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's ok
		}
		return fmt.Errorf("failed to read metrics file: %w", err)
	}

	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if err := json.Unmarshal(data, &mc.metrics); err != nil {
		return fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	return nil
}

// Helper functions
func calculatePercentile(values []int64, percentile int) float64 {
	if len(values) == 0 {
		return 0
	}

	// Simple percentile calculation (not sorting to avoid modifying slice)
	sorted := make([]int64, len(values))
	copy(sorted, values)

	// Bubble sort (fine for small datasets)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)) * float64(percentile) / 100.0)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return float64(sorted[index])
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
