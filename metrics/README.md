# Metrics Collection Package

**Standalone metrics collection system for tracking performance across AI providers and ML-Worker operations.**

## Features

- ✅ **Provider-agnostic**: Works with OpenAI, Gemini, Anthropic, Ollama, ML-Worker, etc.
- ✅ **Zero dependencies**: Pure Go, no external libraries
- ✅ **Thread-safe**: Concurrent recording and retrieval
- ✅ **Export formats**: CSV and JSON
- ✅ **Percentile calculations**: P50, P95, P99 latency tracking
- ✅ **No-op mode**: Zero overhead when disabled
- ✅ **In-memory storage**: Fast, suitable for single-node deployments

## Installation

```bash
go get github.com/itaiguardiola/Askara/metrics
```

## Quick Start

```go
package main

import (
    "fmt"
    "time"
    "github.com/itaiguardiola/Askara/metrics"
)

func main() {
    // Create collector
    collector := metrics.NewInMemoryCollector()

    // Record a task
    collector.RecordTask(metrics.TaskMetric{
        TaskID:       "emb_001",
        TaskType:     metrics.TaskTypeEmbedding,
        Provider:     "openai",
        DurationMs:   150,
        Success:      true,
        InputTokens:  100,
        CostUSD:      0.001,
        UserUUID:     "user-123",
    })

    // Get stats for a provider
    timeRange := metrics.TimeRange{
        Start: time.Now().Add(-24 * time.Hour),
        End:   time.Now(),
    }

    stats, _ := collector.GetProviderStats("openai", timeRange)
    fmt.Printf("OpenAI - Requests: %d, Avg Latency: %.2fms, Uptime: %.2f%%\n",
        stats.RequestCount, stats.AvgLatencyMs, stats.Uptime)

    // Export to CSV
    collector.ExportToCSV("metrics.csv", timeRange)
}
```

## Usage Examples

### Recording Different Task Types

```go
// Embedding generation
collector.RecordTask(metrics.TaskMetric{
    TaskID:       "task-1",
    TaskType:     metrics.TaskTypeEmbedding,
    Provider:     "gemini",
    DurationMs:   98,
    Success:      true,
    InputTokens:  120,
    CostUSD:      0.00001,
})

// LLM Completion
collector.RecordTask(metrics.TaskMetric{
    TaskID:       "task-2",
    TaskType:     metrics.TaskTypeCompletion,
    Provider:     "openai",
    DurationMs:   2100,
    Success:      true,
    InputTokens:  1500,
    OutputTokens: 500,
    CostUSD:      0.03,
})

// ML-Worker Reranking
collector.RecordTask(metrics.TaskMetric{
    TaskID:     "task-3",
    TaskType:   metrics.TaskTypeRerank,
    Provider:   "ml_worker",
    DurationMs: 450,
    Success:    true,
    InputSize:  int64(20 * 1024), // 20 documents
    OutputSize: int64(5 * 1024),  // Top 5 results
})

// Query Rewriting
collector.RecordTask(metrics.TaskMetric{
    TaskID:     "task-4",
    TaskType:   metrics.TaskTypeQueryRewrite,
    Provider:   "local_llm",
    DurationMs: 800,
    Success:    true,
})
```

### Recording Decisions

```go
// Track provider selection decision
collector.RecordDecision(metrics.DecisionMetric{
    DecisionID:   "decision-1",
    DecisionType: "provider_selection",
    ChosenOption: "gemini",
    Alternatives: []string{"openai", "ollama"},
    Reason:       "cost_optimization",
    Success:      true,
    CostSavings:  0.0009, // Saved $0.0009 by choosing Gemini
})

// Track feature toggle decision
collector.RecordDecision(metrics.DecisionMetric{
    DecisionID:   "decision-2",
    DecisionType: "feature_toggle",
    ChosenOption: "enable_reranking",
    Alternatives: []string{"skip_reranking"},
    Reason:       "quality_improvement_expected",
    Success:      true,
    DurationMs:   450,
})
```

### Filtering Metrics

```go
// Get all embedding tasks for a specific provider
filter := metrics.TaskFilter{
    TaskType: ptrTaskType(metrics.TaskTypeEmbedding),
    Provider: ptrString("gemini"),
}
tasks, _ := collector.GetTaskMetrics(filter)

// Get successful tasks only
filter = metrics.TaskFilter{
    SuccessOnly: true,
}
successfulTasks, _ := collector.GetTaskMetrics(filter)

// Get tasks in a time range
filter = metrics.TaskFilter{
    StartTime: ptrTime(time.Now().Add(-1 * time.Hour)),
    EndTime:   ptrTime(time.Now()),
}
recentTasks, _ := collector.GetTaskMetrics(filter)

// Helper functions for creating pointers
func ptrTaskType(t metrics.TaskType) *metrics.TaskType { return &t }
func ptrString(s string) *string { return &s }
func ptrTime(t time.Time) *time.Time { return &t }
```

### Comparing Providers

```go
timeRange := metrics.TimeRange{
    Start: time.Now().Add(-24 * time.Hour),
    End:   time.Now(),
}

allStats, _ := collector.GetAllProviderStats(timeRange)

fmt.Println("Provider Comparison (Last 24h):")
fmt.Println("================================")

for provider, stats := range allStats {
    fmt.Printf("%s:\n", provider)
    fmt.Printf("  Requests: %d\n", stats.RequestCount)
    fmt.Printf("  Success Rate: %.2f%%\n", stats.Uptime)
    fmt.Printf("  Avg Latency: %.2fms\n", stats.AvgLatencyMs)
    fmt.Printf("  P95 Latency: %.2fms\n", stats.P95LatencyMs)
    fmt.Printf("  Total Cost: $%.4f\n", stats.TotalCostUSD)
    fmt.Println()
}
```

Example output:
```
Provider Comparison (Last 24h):
================================
openai:
  Requests: 1523
  Success Rate: 98.50%
  Avg Latency: 145.32ms
  P95 Latency: 320.50ms
  Total Cost: $1.5230

gemini:
  Requests: 2104
  Success Rate: 99.20%
  Avg Latency: 98.45ms
  P95 Latency: 210.30ms
  Total Cost: $0.2104

ollama:
  Requests: 450
  Success Rate: 100.00%
  Avg Latency: 4200.15ms
  P95 Latency: 8500.20ms
  Total Cost: $0.0000
```

### Exporting Data

```go
timeRange := metrics.TimeRange{
    Start: time.Now().Add(-7 * 24 * time.Hour), // Last 7 days
    End:   time.Now(),
}

// Export to CSV
err := collector.ExportToCSV("metrics_weekly.csv", timeRange)
if err != nil {
    log.Printf("Failed to export CSV: %v", err)
}

// Export to JSON
err = collector.ExportToJSON("metrics_weekly.json", timeRange)
if err != nil {
    log.Printf("Failed to export JSON: %v", err)
}
```

CSV format includes all fields:
```csv
TaskID,TaskType,Provider,Timestamp,DurationMs,Success,ErrorType,InputTokens,OutputTokens,InputSize,OutputSize,CostUSD,QualityScore,UserUUID
task-1,embedding,gemini,2025-01-08T10:30:00Z,98,true,,120,0,0,0,0.000010,0.00,user-123
task-2,completion,openai,2025-01-08T10:31:00Z,2100,true,,1500,500,0,0,0.030000,0.00,user-123
```

### Maintenance

```go
// Purge metrics older than 30 days
err := collector.PurgeOldMetrics(30 * 24 * time.Hour)
if err != nil {
    log.Printf("Failed to purge old metrics: %v", err)
}

// Get current metrics count
tasks, decisions := collector.GetMetricsCount()
fmt.Printf("Stored metrics: %d tasks, %d decisions\n", tasks, decisions)
```

## No-Op Collector (Disabled Mode)

When metrics are disabled, use the NoOpCollector for zero overhead:

```go
var collector metrics.MetricsCollector

if os.Getenv("METRICS_ENABLED") == "true" {
    collector = metrics.NewInMemoryCollector()
} else {
    collector = metrics.NewNoOpCollector() // Does nothing
}

// All calls work the same way, but NoOp has zero overhead
collector.RecordTask(metrics.TaskMetric{...}) // No-op if disabled
```

## Task Types

The package supports the following task types:

| TaskType | Description | Common Providers |
|----------|-------------|------------------|
| `TaskTypeEmbedding` | Text embedding generation | openai, gemini, voyage, cohere |
| `TaskTypeCompletion` | LLM text completion | openai, gemini, anthropic, ollama |
| `TaskTypeRerank` | Semantic reranking | ml_worker, cohere |
| `TaskTypeMLParse` | Document parsing | ml_worker, docconv |
| `TaskTypeQueryRewrite` | Query expansion/rewriting | local_llm |
| `TaskTypeVectorSearch` | Vector database search | qdrant, pinecone |
| `TaskTypeOCR` | Optical character recognition | ml_worker, tesseract |
| `TaskTypeDocConv` | Document conversion | docconv |

## Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./metrics -v

# Run with coverage
go test ./metrics -cover

# Run specific test
go test ./metrics -run TestInMemoryCollector_GetProviderStats
```

Expected output:
```
=== RUN   TestInMemoryCollector_RecordTask
--- PASS: TestInMemoryCollector_RecordTask (0.00s)
=== RUN   TestInMemoryCollector_GetTaskMetrics_FilterByType
--- PASS: TestInMemoryCollector_GetTaskMetrics_FilterByType (0.00s)
=== RUN   TestInMemoryCollector_GetProviderStats
--- PASS: TestInMemoryCollector_GetProviderStats (0.00s)
...
PASS
coverage: 87.5% of statements
```

## Performance

Benchmarks on typical hardware (results may vary):

- **Recording a metric**: ~500 ns/op
- **Retrieving stats**: ~50 μs for 10,000 metrics
- **Export to CSV**: ~10 ms for 10,000 metrics
- **Memory usage**: ~500 bytes per metric

The InMemoryCollector is suitable for most single-node deployments. For high-volume production use, consider:
- Periodic export to persistent storage
- Regular purging of old metrics
- Integration with time-series databases (Prometheus, InfluxDB)

## Thread Safety

All operations on InMemoryCollector are thread-safe and can be called concurrently from multiple goroutines.

```go
// Safe to use from multiple goroutines
go func() {
    collector.RecordTask(...)
}()

go func() {
    stats, _ := collector.GetProviderStats(...)
}()
```

## Integration with Askara

This package is designed to integrate seamlessly with Askara's existing architecture:

```go
// In vault-web-server/main.go
var metricsCollector metrics.MetricsCollector

if os.Getenv("METRICS_ENABLED") == "true" {
    metricsCollector = metrics.NewInMemoryCollector()
    log.Println("[Metrics] Enabled")
} else {
    metricsCollector = metrics.NewNoOpCollector()
}

// Pass to handler context
handlerContext := &postapi.HandlerContext{
    // ... existing fields
    MetricsCollector: metricsCollector,
}
```

## Future Enhancements

Planned features for future releases:

- [ ] Prometheus exporter
- [ ] ClickHouse backend for ML training datasets
- [ ] Real-time aggregation windows
- [ ] Alerting thresholds
- [ ] HTTP API for querying metrics
- [ ] Grafana dashboard templates

## Contributing

Contributions are welcome! Please ensure:
- All tests pass: `go test ./metrics -v`
- Code coverage > 80%
- Thread-safe implementations
- Clear documentation

## License

MIT License - See LICENSE file for details

## Support

For questions or issues, please file an issue on GitHub or contact the maintainer.

---

**Part of the Askara Performance Metrics System**
See `PERFORMANCE_METRICS_PLAN.md` for full implementation roadmap.
