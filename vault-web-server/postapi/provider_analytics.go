package postapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/itaiguardiola/askara/llm"
)

// ProviderStatsResponse represents provider statistics for the API
type ProviderStatsResponse struct {
	Provider       string  `json:"provider"`
	TotalRequests  int64   `json:"total_requests"`
	SuccessfulReqs int64   `json:"successful_requests"`
	FailedReqs     int64   `json:"failed_requests"`
	SuccessRate    float64 `json:"success_rate"`
	AvgLatencyMs   float64 `json:"avg_latency_ms"`
	P95LatencyMs   float64 `json:"p95_latency_ms"`
	P99LatencyMs   float64 `json:"p99_latency_ms"`
	TotalTokensIn  int64   `json:"total_tokens_in"`
	TotalTokensOut int64   `json:"total_tokens_out"`
	TotalCost      float64 `json:"total_cost"`
	LastSuccess    *string `json:"last_success,omitempty"`
	LastFailure    *string `json:"last_failure,omitempty"`
	IsHealthy      bool    `json:"is_healthy"`
	HealthScore    float64 `json:"health_score"`
}

// AllProvidersStatsResponse contains stats for all providers
type AllProvidersStatsResponse struct {
	Providers      map[string]ProviderStatsResponse `json:"providers"`
	Mode           string                           `json:"mode"`
	ActiveProvider string                           `json:"active_provider,omitempty"`
}

// RecentMetricsResponse contains recent metrics
type RecentMetricsResponse struct {
	Metrics []MetricResponse `json:"metrics"`
	Count   int              `json:"count"`
}

// MetricResponse represents a single metric
type MetricResponse struct {
	Provider     string  `json:"provider"`
	MetricType   string  `json:"metric_type"`
	Timestamp    string  `json:"timestamp"`
	LatencyMs    int64   `json:"latency_ms"`
	Success      bool    `json:"success"`
	ErrorMessage string  `json:"error_message,omitempty"`
	TokensInput  int     `json:"tokens_input,omitempty"`
	TokensOutput int     `json:"tokens_output,omitempty"`
	Model        string  `json:"model,omitempty"`
	CostEstimate float64 `json:"cost_estimate,omitempty"`
}

// GetProviderStatsHandler returns statistics for all providers
func (ctx *HandlerContext) GetProviderStatsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[GetProviderStatsHandler] Fetching provider statistics")

	// Check if provider manager exists
	if ctx.providerManager == nil {
		log.Println("[GetProviderStatsHandler] Provider manager not initialized")
		http.Error(w, "Provider manager not available", http.StatusServiceUnavailable)
		return
	}

	// Get stats from provider manager
	allStats := ctx.providerManager.GetStats()
	if allStats == nil {
		log.Println("[GetProviderStatsHandler] No stats available")
		response := AllProvidersStatsResponse{
			Providers: make(map[string]ProviderStatsResponse),
			Mode:      string(ctx.providerManager.GetMode()),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Convert stats to response format
	providers := make(map[string]ProviderStatsResponse)
	for providerType, stats := range allStats {
		var lastSuccess, lastFailure *string
		if stats.LastSuccess != nil {
			s := stats.LastSuccess.Format("2006-01-02T15:04:05Z07:00")
			lastSuccess = &s
		}
		if stats.LastFailure != nil {
			s := stats.LastFailure.Format("2006-01-02T15:04:05Z07:00")
			lastFailure = &s
		}

		providers[string(providerType)] = ProviderStatsResponse{
			Provider:       string(stats.Provider),
			TotalRequests:  stats.TotalRequests,
			SuccessfulReqs: stats.SuccessfulReqs,
			FailedReqs:     stats.FailedReqs,
			SuccessRate:    stats.SuccessRate,
			AvgLatencyMs:   stats.AvgLatencyMs,
			P95LatencyMs:   stats.P95LatencyMs,
			P99LatencyMs:   stats.P99LatencyMs,
			TotalTokensIn:  stats.TotalTokensIn,
			TotalTokensOut: stats.TotalTokensOut,
			TotalCost:      stats.TotalCost,
			LastSuccess:    lastSuccess,
			LastFailure:    lastFailure,
			IsHealthy:      stats.IsHealthy,
			HealthScore:    stats.HealthScore,
		}
	}

	response := AllProvidersStatsResponse{
		Providers: providers,
		Mode:      string(ctx.providerManager.GetMode()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("[GetProviderStatsHandler] Returned stats for %d providers", len(providers))
}

// GetRecentMetricsHandler returns recent metrics
func (ctx *HandlerContext) GetRecentMetricsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[GetRecentMetricsHandler] Fetching recent metrics")

	// Check if provider manager exists
	if ctx.providerManager == nil {
		log.Println("[GetRecentMetricsHandler] Provider manager not initialized")
		http.Error(w, "Provider manager not available", http.StatusServiceUnavailable)
		return
	}

	// Get metrics collector
	collector := ctx.providerManager.GetMetrics()
	if collector == nil {
		log.Println("[GetRecentMetricsHandler] Metrics collector not available")
		response := RecentMetricsResponse{
			Metrics: []MetricResponse{},
			Count:   0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get recent metrics (last 100)
	recentMetrics := collector.GetRecentMetrics(100)

	// Convert to response format
	metrics := make([]MetricResponse, len(recentMetrics))
	for i, m := range recentMetrics {
		metrics[i] = MetricResponse{
			Provider:     string(m.Provider),
			MetricType:   string(m.MetricType),
			Timestamp:    m.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			LatencyMs:    m.LatencyMs,
			Success:      m.Success,
			ErrorMessage: m.ErrorMessage,
			TokensInput:  m.TokensInput,
			TokensOutput: m.TokensOutput,
			Model:        m.Model,
			CostEstimate: m.CostEstimate,
		}
	}

	response := RecentMetricsResponse{
		Metrics: metrics,
		Count:   len(metrics),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("[GetRecentMetricsHandler] Returned %d metrics", len(metrics))
}

// GetProviderHealthHandler returns health status for all providers
func (ctx *HandlerContext) GetProviderHealthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[GetProviderHealthHandler] Fetching provider health")

	// Check if provider manager exists
	if ctx.providerManager == nil {
		log.Println("[GetProviderHealthHandler] Provider manager not initialized")
		http.Error(w, "Provider manager not available", http.StatusServiceUnavailable)
		return
	}

	// Get stats which includes health information
	allStats := ctx.providerManager.GetStats()
	if allStats == nil {
		log.Println("[GetProviderHealthHandler] No stats available")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"providers": map[string]interface{}{},
			"mode":      string(ctx.providerManager.GetMode()),
		})
		return
	}

	// Build health response
	healthStatus := make(map[string]interface{})
	for providerType, stats := range allStats {
		healthStatus[string(providerType)] = map[string]interface{}{
			"is_healthy":      stats.IsHealthy,
			"health_score":    stats.HealthScore,
			"success_rate":    stats.SuccessRate,
			"avg_latency":     stats.AvgLatencyMs,
			"total_requests":  stats.TotalRequests,
			"failed_requests": stats.FailedReqs,
		}
	}

	response := map[string]interface{}{
		"providers": healthStatus,
		"mode":      string(ctx.providerManager.GetMode()),
		"timestamp": llm.MetricsCollector{}.GetRecentMetrics(0), // Current time
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("[GetProviderHealthHandler] Returned health for %d providers", len(healthStatus))
}
