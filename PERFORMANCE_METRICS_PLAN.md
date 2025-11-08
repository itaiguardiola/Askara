# Performance Metrics & Multi-Provider Plugin System
## Phased Implementation Plan

**Goal:** Build a provider-agnostic performance measurement system for ML-Worker operations, AI providers (OpenAI, Gemini, Anthropic, etc.), and optional features with minimal disruption to production.

**Strategy:**
- **Phases 1-2:** Completely independent development (no main branch changes)
- **Phases 3-4:** Light integration with feature flags (optional, non-breaking)
- **Phases 5-6:** Deeper integration, still feature-flagged (low risk)
- **Phases 7+:** Full deployment with minimal code changes (safe merges)

---

## PHASE 1: Standalone Metrics Framework (Week 1)
**Branch:** `feature/metrics-framework-standalone`
**Main Branch Impact:** 🟢 ZERO (completely independent)

### Deliverables

#### 1.1 Core Metrics Package (`/metrics`)
Create new package with no dependencies on existing code:

```
/metrics/
├── types.go              # Data structures
├── interface.go          # MetricsCollector interface
├── collector.go          # Default implementation
├── storage.go            # In-memory storage
├── exporter.go           # CSV/JSON export
└── metrics_test.go       # Unit tests
```

**Key Files:**

**`metrics/types.go`**
```go
package metrics

import "time"

// Core metric types - provider agnostic
type TaskMetric struct {
    TaskID          string
    TaskType        TaskType
    Provider        string
    Timestamp       time.Time
    DurationMs      int64
    Success         bool
    ErrorType       string

    // Performance
    InputTokens     int
    OutputTokens    int
    InputSize       int64
    OutputSize      int64

    // Cost
    CostUSD         float64

    // Quality
    QualityScore    float64

    // Context
    UserUUID        string
    Metadata        map[string]interface{}
}

type TaskType string
const (
    TaskTypeEmbedding      TaskType = "embedding"
    TaskTypeCompletion     TaskType = "completion"
    TaskTypeRerank         TaskType = "rerank"
    TaskTypeMLParse        TaskType = "ml_parse"
    TaskTypeQueryRewrite   TaskType = "query_rewrite"
    TaskTypeVectorSearch   TaskType = "vector_search"
)

type ProviderMetric struct {
    Provider        string
    Timestamp       time.Time

    // Aggregated stats
    RequestCount    int64
    SuccessCount    int64
    ErrorCount      int64

    // Latency
    AvgLatencyMs    float64
    P50LatencyMs    float64
    P95LatencyMs    float64
    P99LatencyMs    float64

    // Cost
    TotalCostUSD    float64

    // Availability
    Uptime          float64
}

type DecisionMetric struct {
    DecisionID      string
    DecisionType    string
    Timestamp       time.Time

    ChosenOption    string
    Alternatives    []string
    Reason          string

    // Outcome
    Success         bool
    DurationMs      int64
    CostSavings     float64
}
```

**`metrics/interface.go`**
```go
package metrics

import (
    "context"
    "time"
)

type MetricsCollector interface {
    // Task tracking
    RecordTask(metric TaskMetric) error
    RecordDecision(metric DecisionMetric) error

    // Retrieval
    GetTaskMetrics(filter TaskFilter) ([]TaskMetric, error)
    GetProviderStats(provider string, timeRange TimeRange) (*ProviderMetric, error)
    GetAllProviderStats(timeRange TimeRange) (map[string]*ProviderMetric, error)

    // Export
    ExportToCSV(filename string, timeRange TimeRange) error
    ExportToJSON(filename string, timeRange TimeRange) error

    // Cleanup
    PurgeOldMetrics(olderThan time.Duration) error
}

type TaskFilter struct {
    TaskType    *TaskType
    Provider    *string
    UserUUID    *string
    StartTime   *time.Time
    EndTime     *time.Time
    SuccessOnly bool
}

type TimeRange struct {
    Start time.Time
    End   time.Time
}
```

**`metrics/collector.go`**
```go
package metrics

import (
    "encoding/csv"
    "encoding/json"
    "fmt"
    "os"
    "sort"
    "sync"
    "time"
)

// InMemoryCollector - simple, dependency-free implementation
type InMemoryCollector struct {
    tasks     []TaskMetric
    decisions []DecisionMetric
    mu        sync.RWMutex
}

func NewInMemoryCollector() *InMemoryCollector {
    return &InMemoryCollector{
        tasks:     make([]TaskMetric, 0),
        decisions: make([]DecisionMetric, 0),
    }
}

func (c *InMemoryCollector) RecordTask(metric TaskMetric) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if metric.Timestamp.IsZero() {
        metric.Timestamp = time.Now()
    }

    c.tasks = append(c.tasks, metric)
    return nil
}

func (c *InMemoryCollector) RecordDecision(metric DecisionMetric) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if metric.Timestamp.IsZero() {
        metric.Timestamp = time.Now()
    }

    c.decisions = append(c.decisions, metric)
    return nil
}

func (c *InMemoryCollector) GetTaskMetrics(filter TaskFilter) ([]TaskMetric, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    result := make([]TaskMetric, 0)

    for _, task := range c.tasks {
        if c.matchesFilter(task, filter) {
            result = append(result, task)
        }
    }

    return result, nil
}

func (c *InMemoryCollector) matchesFilter(task TaskMetric, filter TaskFilter) bool {
    if filter.TaskType != nil && task.TaskType != *filter.TaskType {
        return false
    }
    if filter.Provider != nil && task.Provider != *filter.Provider {
        return false
    }
    if filter.UserUUID != nil && task.UserUUID != *filter.UserUUID {
        return false
    }
    if filter.StartTime != nil && task.Timestamp.Before(*filter.StartTime) {
        return false
    }
    if filter.EndTime != nil && task.Timestamp.After(*filter.EndTime) {
        return false
    }
    if filter.SuccessOnly && !task.Success {
        return false
    }
    return true
}

func (c *InMemoryCollector) GetProviderStats(provider string, timeRange TimeRange) (*ProviderMetric, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    var relevantTasks []TaskMetric
    for _, task := range c.tasks {
        if task.Provider == provider &&
           task.Timestamp.After(timeRange.Start) &&
           task.Timestamp.Before(timeRange.End) {
            relevantTasks = append(relevantTasks, task)
        }
    }

    if len(relevantTasks) == 0 {
        return nil, fmt.Errorf("no metrics found for provider %s", provider)
    }

    return c.aggregateMetrics(provider, relevantTasks), nil
}

func (c *InMemoryCollector) aggregateMetrics(provider string, tasks []TaskMetric) *ProviderMetric {
    metric := &ProviderMetric{
        Provider:  provider,
        Timestamp: time.Now(),
    }

    metric.RequestCount = int64(len(tasks))

    var totalLatency int64
    var latencies []int64
    var totalCost float64

    for _, task := range tasks {
        if task.Success {
            metric.SuccessCount++
        } else {
            metric.ErrorCount++
        }

        totalLatency += task.DurationMs
        latencies = append(latencies, task.DurationMs)
        totalCost += task.CostUSD
    }

    metric.AvgLatencyMs = float64(totalLatency) / float64(len(tasks))
    metric.TotalCostUSD = totalCost
    metric.Uptime = float64(metric.SuccessCount) / float64(metric.RequestCount) * 100

    // Calculate percentiles
    sort.Slice(latencies, func(i, j int) { return latencies[i] < latencies[j] })
    metric.P50LatencyMs = float64(latencies[len(latencies)/2])
    metric.P95LatencyMs = float64(latencies[int(float64(len(latencies))*0.95)])
    metric.P99LatencyMs = float64(latencies[int(float64(len(latencies))*0.99)])

    return metric
}

func (c *InMemoryCollector) GetAllProviderStats(timeRange TimeRange) (map[string]*ProviderMetric, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    providerTasks := make(map[string][]TaskMetric)

    for _, task := range c.tasks {
        if task.Timestamp.After(timeRange.Start) && task.Timestamp.Before(timeRange.End) {
            providerTasks[task.Provider] = append(providerTasks[task.Provider], task)
        }
    }

    result := make(map[string]*ProviderMetric)
    for provider, tasks := range providerTasks {
        result[provider] = c.aggregateMetrics(provider, tasks)
    }

    return result, nil
}

func (c *InMemoryCollector) ExportToCSV(filename string, timeRange TimeRange) error {
    c.mu.RLock()
    defer c.mu.RUnlock()

    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Header
    writer.Write([]string{
        "TaskID", "TaskType", "Provider", "Timestamp", "DurationMs",
        "Success", "ErrorType", "InputTokens", "OutputTokens", "CostUSD",
        "QualityScore", "UserUUID",
    })

    // Data
    for _, task := range c.tasks {
        if task.Timestamp.After(timeRange.Start) && task.Timestamp.Before(timeRange.End) {
            writer.Write([]string{
                task.TaskID,
                string(task.TaskType),
                task.Provider,
                task.Timestamp.Format(time.RFC3339),
                fmt.Sprintf("%d", task.DurationMs),
                fmt.Sprintf("%t", task.Success),
                task.ErrorType,
                fmt.Sprintf("%d", task.InputTokens),
                fmt.Sprintf("%d", task.OutputTokens),
                fmt.Sprintf("%.4f", task.CostUSD),
                fmt.Sprintf("%.2f", task.QualityScore),
                task.UserUUID,
            })
        }
    }

    return nil
}

func (c *InMemoryCollector) ExportToJSON(filename string, timeRange TimeRange) error {
    c.mu.RLock()
    defer c.mu.RUnlock()

    var exportTasks []TaskMetric
    for _, task := range c.tasks {
        if task.Timestamp.After(timeRange.Start) && task.Timestamp.Before(timeRange.End) {
            exportTasks = append(exportTasks, task)
        }
    }

    data, err := json.MarshalIndent(exportTasks, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(filename, data, 0644)
}

func (c *InMemoryCollector) PurgeOldMetrics(olderThan time.Duration) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    cutoff := time.Now().Add(-olderThan)

    // Filter tasks
    var filteredTasks []TaskMetric
    for _, task := range c.tasks {
        if task.Timestamp.After(cutoff) {
            filteredTasks = append(filteredTasks, task)
        }
    }
    c.tasks = filteredTasks

    // Filter decisions
    var filteredDecisions []DecisionMetric
    for _, decision := range c.decisions {
        if decision.Timestamp.After(cutoff) {
            filteredDecisions = append(filteredDecisions, decision)
        }
    }
    c.decisions = filteredDecisions

    return nil
}
```

#### 1.2 Testing Framework
Create comprehensive tests (can run in isolation):

**`metrics/metrics_test.go`**
```go
package metrics

import (
    "os"
    "testing"
    "time"
)

func TestInMemoryCollector(t *testing.T) {
    collector := NewInMemoryCollector()

    // Test recording tasks
    task1 := TaskMetric{
        TaskID:     "task1",
        TaskType:   TaskTypeEmbedding,
        Provider:   "openai",
        DurationMs: 150,
        Success:    true,
        CostUSD:    0.001,
    }

    err := collector.RecordTask(task1)
    if err != nil {
        t.Errorf("Failed to record task: %v", err)
    }

    // Test retrieval
    filter := TaskFilter{
        TaskType: ptrTaskType(TaskTypeEmbedding),
    }

    tasks, err := collector.GetTaskMetrics(filter)
    if err != nil {
        t.Errorf("Failed to retrieve tasks: %v", err)
    }

    if len(tasks) != 1 {
        t.Errorf("Expected 1 task, got %d", len(tasks))
    }
}

func TestProviderStats(t *testing.T) {
    collector := NewInMemoryCollector()

    // Add multiple tasks
    for i := 0; i < 10; i++ {
        collector.RecordTask(TaskMetric{
            TaskID:     fmt.Sprintf("task%d", i),
            TaskType:   TaskTypeCompletion,
            Provider:   "openai",
            DurationMs: int64(100 + i*10),
            Success:    i < 9, // 1 failure
            CostUSD:    0.01,
        })
    }

    timeRange := TimeRange{
        Start: time.Now().Add(-1 * time.Hour),
        End:   time.Now().Add(1 * time.Hour),
    }

    stats, err := collector.GetProviderStats("openai", timeRange)
    if err != nil {
        t.Errorf("Failed to get provider stats: %v", err)
    }

    if stats.RequestCount != 10 {
        t.Errorf("Expected 10 requests, got %d", stats.RequestCount)
    }

    if stats.SuccessCount != 9 {
        t.Errorf("Expected 9 successes, got %d", stats.SuccessCount)
    }

    if stats.Uptime != 90.0 {
        t.Errorf("Expected 90%% uptime, got %.2f%%", stats.Uptime)
    }
}

func TestExportCSV(t *testing.T) {
    collector := NewInMemoryCollector()

    collector.RecordTask(TaskMetric{
        TaskID:     "export_test",
        TaskType:   TaskTypeEmbedding,
        Provider:   "gemini",
        DurationMs: 100,
        Success:    true,
        CostUSD:    0.0001,
    })

    filename := "/tmp/metrics_test.csv"
    defer os.Remove(filename)

    timeRange := TimeRange{
        Start: time.Now().Add(-1 * time.Hour),
        End:   time.Now().Add(1 * time.Hour),
    }

    err := collector.ExportToCSV(filename, timeRange)
    if err != nil {
        t.Errorf("Failed to export CSV: %v", err)
    }

    // Verify file exists
    if _, err := os.Stat(filename); os.IsNotExist(err) {
        t.Error("CSV file was not created")
    }
}

func ptrTaskType(t TaskType) *TaskType {
    return &t
}
```

#### 1.3 Documentation

**`metrics/README.md`**
```markdown
# Metrics Collection Package

Standalone metrics collection system for tracking performance across AI providers.

## Features

- **Provider-agnostic**: Works with OpenAI, Gemini, Anthropic, Ollama, etc.
- **Zero dependencies**: Pure Go, no external libraries
- **Thread-safe**: Concurrent recording and retrieval
- **Export formats**: CSV and JSON
- **Percentile calculations**: P50, P95, P99 latency tracking

## Usage

```go
import "github.com/itaiguardiola/Askara/metrics"

// Create collector
collector := metrics.NewInMemoryCollector()

// Record a task
collector.RecordTask(metrics.TaskMetric{
    TaskID:     "emb_001",
    TaskType:   metrics.TaskTypeEmbedding,
    Provider:   "openai",
    DurationMs: 150,
    Success:    true,
    CostUSD:    0.001,
})

// Get stats
stats, _ := collector.GetProviderStats("openai", timeRange)
fmt.Printf("Avg latency: %.2fms\n", stats.AvgLatencyMs)

// Export
collector.ExportToCSV("metrics.csv", timeRange)
```

## Testing

```bash
go test ./metrics -v
```
```

### Testing & Validation
```bash
# Independent testing - no need for running server
cd /home/user/Askara
go test ./metrics -v -cover

# Should see:
# ✓ TestInMemoryCollector
# ✓ TestProviderStats
# ✓ TestExportCSV
# Coverage: 85%+
```

### Deliverable Checklist
- [ ] `/metrics` package created
- [ ] All interfaces defined
- [ ] InMemoryCollector implemented
- [ ] Unit tests passing (>80% coverage)
- [ ] Documentation complete
- [ ] **No changes to existing code**

---

## PHASE 2: Provider Interface Definitions (Week 1-2)
**Branch:** `feature/unified-provider-interface`
**Main Branch Impact:** 🟢 ZERO (pure abstraction layer)

### Deliverables

#### 2.1 Unified Provider Interface

**`/llm/unified/interface.go`**
```go
package unified

import (
    "context"
    "time"

    "github.com/itaiguardiola/Askara/metrics"
)

// UnifiedProvider - abstraction for all AI providers
type UnifiedProvider interface {
    // Metadata
    GetName() string
    GetCapabilities() []Capability
    GetPricingModel() PricingModel
    GetRateLimits() RateLimits

    // Core operations (instrumented)
    GenerateEmbedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)
    GenerateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    StreamCompletion(ctx context.Context, req *CompletionRequest, onChunk func(string)) error

    // Health
    HealthCheck(ctx context.Context) error
}

// Request/Response types
type EmbeddingRequest struct {
    Text     string
    Model    string
    UserUUID string
    Metadata map[string]interface{}
}

type EmbeddingResponse struct {
    Embedding     []float32
    DimensionSize int

    // Performance metrics
    DurationMs int64
    TokensUsed int
    CostUSD    float64

    // Provider-specific
    Model    string
    Metadata map[string]interface{}
}

type CompletionRequest struct {
    Prompt      string
    Context     []string
    MaxTokens   int
    Temperature float32
    Model       string
    UserUUID    string
    Streaming   bool
}

type CompletionResponse struct {
    Text string

    // Performance metrics
    DurationMs       int64
    TimeToFirstToken int64  // For streaming
    InputTokens      int
    OutputTokens     int
    CostUSD          float64

    // Provider-specific
    Model       string
    FinishReason string
    Metadata    map[string]interface{}
}

// Capability represents provider features
type Capability string

const (
    CapabilityEmbedding      Capability = "embedding"
    CapabilityCompletion     Capability = "completion"
    CapabilityStreaming      Capability = "streaming"
    CapabilityVision         Capability = "vision"
    CapabilityFunctionCall   Capability = "function_calling"
    CapabilityLongContext    Capability = "long_context"  // >32k tokens
)

type PricingModel struct {
    Type                 PricingType
    EmbeddingPer1KTokens float64
    InputPer1MTokens     float64
    OutputPer1MTokens    float64
    FixedPerRequest      float64
}

type PricingType string

const (
    PricingFree       PricingType = "free"
    PricingPayPerUse  PricingType = "pay_per_use"
    PricingSubscription PricingType = "subscription"
)

type RateLimits struct {
    RequestsPerMinute int
    TokensPerMinute   int
    TokensPerDay      int
}
```

#### 2.2 Provider Wrapper with Auto-Instrumentation

**`/llm/unified/instrumented_provider.go`**
```go
package unified

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/itaiguardiola/Askara/metrics"
)

// InstrumentedProvider wraps any provider with automatic metrics collection
type InstrumentedProvider struct {
    provider  UnifiedProvider
    collector metrics.MetricsCollector
}

func NewInstrumentedProvider(provider UnifiedProvider, collector metrics.MetricsCollector) *InstrumentedProvider {
    return &InstrumentedProvider{
        provider:  provider,
        collector: collector,
    }
}

func (ip *InstrumentedProvider) GetName() string {
    return ip.provider.GetName()
}

func (ip *InstrumentedProvider) GetCapabilities() []Capability {
    return ip.provider.GetCapabilities()
}

func (ip *InstrumentedProvider) GetPricingModel() PricingModel {
    return ip.provider.GetPricingModel()
}

func (ip *InstrumentedProvider) GetRateLimits() RateLimits {
    return ip.provider.GetRateLimits()
}

func (ip *InstrumentedProvider) GenerateEmbedding(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
    taskID := uuid.New().String()
    start := time.Now()

    // Call actual provider
    resp, err := ip.provider.GenerateEmbedding(ctx, req)

    duration := time.Since(start)

    // Record metrics
    metric := metrics.TaskMetric{
        TaskID:       taskID,
        TaskType:     metrics.TaskTypeEmbedding,
        Provider:     ip.provider.GetName(),
        Timestamp:    start,
        DurationMs:   duration.Milliseconds(),
        Success:      err == nil,
        UserUUID:     req.UserUUID,
    }

    if err == nil {
        metric.InputTokens = resp.TokensUsed
        metric.CostUSD = resp.CostUSD
        metric.InputSize = int64(len(req.Text))
        metric.OutputSize = int64(len(resp.Embedding) * 4) // 4 bytes per float32
    } else {
        metric.ErrorType = err.Error()
    }

    ip.collector.RecordTask(metric)

    return resp, err
}

func (ip *InstrumentedProvider) GenerateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    taskID := uuid.New().String()
    start := time.Now()

    // Call actual provider
    resp, err := ip.provider.GenerateCompletion(ctx, req)

    duration := time.Since(start)

    // Record metrics
    metric := metrics.TaskMetric{
        TaskID:       taskID,
        TaskType:     metrics.TaskTypeCompletion,
        Provider:     ip.provider.GetName(),
        Timestamp:    start,
        DurationMs:   duration.Milliseconds(),
        Success:      err == nil,
        UserUUID:     req.UserUUID,
    }

    if err == nil {
        metric.InputTokens = resp.InputTokens
        metric.OutputTokens = resp.OutputTokens
        metric.CostUSD = resp.CostUSD
    } else {
        metric.ErrorType = err.Error()
    }

    ip.collector.RecordTask(metric)

    return resp, err
}

func (ip *InstrumentedProvider) StreamCompletion(ctx context.Context, req *CompletionRequest, onChunk func(string)) error {
    taskID := uuid.New().String()
    start := time.Now()

    var firstTokenTime *time.Time
    var totalTokens int

    // Wrap callback to track TTFT
    wrappedCallback := func(chunk string) {
        if firstTokenTime == nil {
            now := time.Now()
            firstTokenTime = &now
        }
        totalTokens++
        onChunk(chunk)
    }

    // Call actual provider
    err := ip.provider.StreamCompletion(ctx, req, wrappedCallback)

    duration := time.Since(start)

    // Record metrics
    metric := metrics.TaskMetric{
        TaskID:     taskID,
        TaskType:   metrics.TaskTypeCompletion,
        Provider:   ip.provider.GetName(),
        Timestamp:  start,
        DurationMs: duration.Milliseconds(),
        Success:    err == nil,
        UserUUID:   req.UserUUID,
    }

    if firstTokenTime != nil {
        ttft := firstTokenTime.Sub(start).Milliseconds()
        metric.Metadata = map[string]interface{}{
            "time_to_first_token_ms": ttft,
            "streaming": true,
        }
    }

    if err != nil {
        metric.ErrorType = err.Error()
    }

    ip.collector.RecordTask(metric)

    return err
}

func (ip *InstrumentedProvider) HealthCheck(ctx context.Context) error {
    return ip.provider.HealthCheck(ctx)
}
```

#### 2.3 Example Implementation: Gemini Provider

**`/llm/unified/providers/gemini/gemini.go`**
```go
package gemini

import (
    "context"
    "fmt"
    "time"

    "github.com/itaiguardiola/Askara/llm/unified"
)

// GeminiProvider implements UnifiedProvider interface
type GeminiProvider struct {
    apiKey     string
    model      string
    endpoint   string
}

func NewGeminiProvider(apiKey string, model string) *GeminiProvider {
    return &GeminiProvider{
        apiKey:   apiKey,
        model:    model,
        endpoint: "https://generativelanguage.googleapis.com/v1",
    }
}

func (g *GeminiProvider) GetName() string {
    return "gemini"
}

func (g *GeminiProvider) GetCapabilities() []unified.Capability {
    return []unified.Capability{
        unified.CapabilityEmbedding,
        unified.CapabilityCompletion,
        unified.CapabilityStreaming,
        unified.CapabilityVision,
        unified.CapabilityLongContext,  // 1M token context!
    }
}

func (g *GeminiProvider) GetPricingModel() unified.PricingModel {
    return unified.PricingModel{
        Type:                 unified.PricingPayPerUse,
        EmbeddingPer1KTokens: 0.00001,  // 10x cheaper than OpenAI
        InputPer1MTokens:     3.50,      // Gemini 1.5 Pro pricing
        OutputPer1MTokens:    10.50,
    }
}

func (g *GeminiProvider) GetRateLimits() unified.RateLimits {
    return unified.RateLimits{
        RequestsPerMinute: 60,
        TokensPerMinute:   1000000,
        TokensPerDay:      50000000,
    }
}

func (g *GeminiProvider) GenerateEmbedding(ctx context.Context, req *unified.EmbeddingRequest) (*unified.EmbeddingResponse, error) {
    start := time.Now()

    // TODO: Actual Gemini API call
    // For now, return stub

    duration := time.Since(start)

    return &unified.EmbeddingResponse{
        Embedding:     []float32{}, // TODO: actual embedding
        DimensionSize: 768,
        DurationMs:    duration.Milliseconds(),
        TokensUsed:    len(req.Text) / 4, // Rough estimate
        CostUSD:       float64(len(req.Text)/4) * 0.00001 / 1000,
        Model:         g.model,
    }, nil
}

func (g *GeminiProvider) GenerateCompletion(ctx context.Context, req *unified.CompletionRequest) (*unified.CompletionResponse, error) {
    start := time.Now()

    // TODO: Actual Gemini API call

    duration := time.Since(start)

    return &unified.CompletionResponse{
        Text:         "", // TODO: actual response
        DurationMs:   duration.Milliseconds(),
        InputTokens:  len(req.Prompt) / 4,
        OutputTokens: 100,
        CostUSD:      0.001, // TODO: calculate based on actual tokens
        Model:        g.model,
    }, nil
}

func (g *GeminiProvider) StreamCompletion(ctx context.Context, req *unified.CompletionRequest, onChunk func(string)) error {
    // TODO: Implement Gemini streaming
    return fmt.Errorf("streaming not yet implemented")
}

func (g *GeminiProvider) HealthCheck(ctx context.Context) error {
    // TODO: Ping Gemini API
    return nil
}
```

#### 2.4 Example: Wrapping Existing OpenAI Provider

**`/llm/unified/adapters/openai_adapter.go`**
```go
package adapters

import (
    "context"

    "github.com/itaiguardiola/Askara/llm"
    "github.com/itaiguardiola/Askara/llm/unified"
)

// OpenAIAdapter wraps existing OpenAI provider to implement UnifiedProvider
type OpenAIAdapter struct {
    provider *llm.OpenAIProvider
}

func NewOpenAIAdapter(provider *llm.OpenAIProvider) *OpenAIAdapter {
    return &OpenAIAdapter{provider: provider}
}

func (a *OpenAIAdapter) GetName() string {
    return "openai"
}

func (a *OpenAIAdapter) GetCapabilities() []unified.Capability {
    return []unified.Capability{
        unified.CapabilityEmbedding,
        unified.CapabilityCompletion,
        unified.CapabilityStreaming,
        unified.CapabilityVision,
        unified.CapabilityFunctionCall,
    }
}

func (a *OpenAIAdapter) GetPricingModel() unified.PricingModel {
    return unified.PricingModel{
        Type:                 unified.PricingPayPerUse,
        EmbeddingPer1KTokens: 0.0001,
        InputPer1MTokens:     10.0,
        OutputPer1MTokens:    30.0,
    }
}

func (a *OpenAIAdapter) GetRateLimits() unified.RateLimits {
    return unified.RateLimits{
        RequestsPerMinute: 500,
        TokensPerMinute:   90000,
        TokensPerDay:      -1, // No daily limit
    }
}

func (a *OpenAIAdapter) GenerateEmbedding(ctx context.Context, req *unified.EmbeddingRequest) (*unified.EmbeddingResponse, error) {
    // Call existing OpenAI provider
    embedding, err := a.provider.GenerateEmbedding(req.Text)
    if err != nil {
        return nil, err
    }

    // Convert to unified response
    return &unified.EmbeddingResponse{
        Embedding:     embedding,
        DimensionSize: len(embedding),
        TokensUsed:    len(req.Text) / 4,
        CostUSD:       float64(len(req.Text)/4) * 0.0001 / 1000,
        Model:         "text-embedding-ada-002",
    }, nil
}

func (a *OpenAIAdapter) GenerateCompletion(ctx context.Context, req *unified.CompletionRequest) (*unified.CompletionResponse, error) {
    // Call existing OpenAI provider
    text, err := a.provider.GenerateCompletion(req.Prompt, req.Context)
    if err != nil {
        return nil, err
    }

    return &unified.CompletionResponse{
        Text:         text,
        InputTokens:  len(req.Prompt) / 4,
        OutputTokens: len(text) / 4,
        Model:        "gpt-4",
    }, nil
}

func (a *OpenAIAdapter) StreamCompletion(ctx context.Context, req *unified.CompletionRequest, onChunk func(string)) error {
    return a.provider.StreamCompletion(req.Prompt, req.Context, onChunk)
}

func (a *OpenAIAdapter) HealthCheck(ctx context.Context) error {
    // Simple health check
    _, err := a.provider.GenerateEmbedding("test")
    return err
}
```

### Testing & Validation
```bash
# Test unified interface implementations
go test ./llm/unified/... -v

# Test adapters work with existing code
go test ./llm/unified/adapters/... -v
```

### Deliverable Checklist
- [ ] `/llm/unified` package created
- [ ] UnifiedProvider interface defined
- [ ] InstrumentedProvider wrapper implemented
- [ ] Example Gemini provider (stub implementation)
- [ ] OpenAI adapter (wraps existing provider)
- [ ] Tests passing
- [ ] **No changes to production code**

---

## PHASE 3: Opt-in Instrumentation (Week 2-3)
**Branch:** `feature/optional-metrics-integration`
**Main Branch Impact:** 🟡 LOW (feature-flagged, optional)

### Strategy
Add metrics collection as **opt-in feature** - disabled by default.

### Deliverables

#### 3.1 Environment Variable Configuration

**`.env.example` additions:**
```bash
# Performance Metrics (Optional)
METRICS_ENABLED=false
METRICS_STORAGE=memory              # memory, file, prometheus
METRICS_EXPORT_PATH=./metrics       # Where to store exports
METRICS_RETENTION_DAYS=30           # Auto-purge old data
METRICS_LOG_LEVEL=info              # debug, info, warn
```

#### 3.2 Metrics Initialization in Main

**`vault-web-server/main.go` additions (feature-flagged):**
```go
package main

import (
    // ... existing imports
    "github.com/itaiguardiola/Askara/metrics"
    "github.com/itaiguardiola/Askara/llm/unified"
)

func main() {
    // ... existing code

    // OPTIONAL: Initialize metrics (disabled by default)
    var metricsCollector metrics.MetricsCollector
    if os.Getenv("METRICS_ENABLED") == "true" {
        log.Println("[Metrics] Enabling performance tracking...")
        metricsCollector = metrics.NewInMemoryCollector()

        // Start background export job
        go startMetricsExporter(metricsCollector)
    } else {
        log.Println("[Metrics] Performance tracking disabled (set METRICS_ENABLED=true to enable)")
        metricsCollector = &metrics.NoOpCollector{} // Does nothing
    }

    // Pass to handler context
    handlerContext := &postapi.HandlerContext{
        // ... existing fields
        MetricsCollector: metricsCollector, // NEW field
    }

    // ... rest of main
}

func startMetricsExporter(collector metrics.MetricsCollector) {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()

    for range ticker.C {
        exportPath := os.Getenv("METRICS_EXPORT_PATH")
        if exportPath == "" {
            exportPath = "./metrics"
        }

        filename := fmt.Sprintf("%s/metrics_%s.csv", exportPath, time.Now().Format("2006-01-02"))

        timeRange := metrics.TimeRange{
            Start: time.Now().Add(-24 * time.Hour),
            End:   time.Now(),
        }

        err := collector.ExportToCSV(filename, timeRange)
        if err != nil {
            log.Printf("[Metrics] Export failed: %v", err)
        } else {
            log.Printf("[Metrics] Exported to %s", filename)
        }
    }
}
```

#### 3.3 Handler Context Update

**`vault-web-server/postapi/handlercontext.go`:**
```go
package postapi

import (
    // ... existing imports
    "github.com/itaiguardiola/Askara/metrics"
)

type HandlerContext struct {
    // ... existing fields
    LLMProvider        llm.LLMProvider
    VectorDB           vectordb.VectorDB
    DocumentStore      *storage.JSONDocumentStore
    Cache              *sync.Map
    MLWorkerClient     *mlworker.Client
    QueryRewriter      *queryrewriter.QueryRewriter

    // NEW: Optional metrics collector
    MetricsCollector   metrics.MetricsCollector  // nil-safe, can be NoOp
}
```

#### 3.4 NoOp Collector (for when disabled)

**`metrics/noop.go`:**
```go
package metrics

// NoOpCollector does nothing - used when metrics are disabled
type NoOpCollector struct{}

func (n *NoOpCollector) RecordTask(metric TaskMetric) error           { return nil }
func (n *NoOpCollector) RecordDecision(metric DecisionMetric) error   { return nil }
func (n *NoOpCollector) GetTaskMetrics(filter TaskFilter) ([]TaskMetric, error) { return nil, nil }
func (n *NoOpCollector) GetProviderStats(provider string, timeRange TimeRange) (*ProviderMetric, error) { return nil, nil }
func (n *NoOpCollector) GetAllProviderStats(timeRange TimeRange) (map[string]*ProviderMetric, error) { return nil, nil }
func (n *NoOpCollector) ExportToCSV(filename string, timeRange TimeRange) error { return nil }
func (n *NoOpCollector) ExportToJSON(filename string, timeRange TimeRange) error { return nil }
func (n *NoOpCollector) PurgeOldMetrics(olderThan time.Duration) error { return nil }
```

#### 3.5 Lightweight Instrumentation Examples

**Example: Instrument Query Rewriting (minimal change):**

**`queryrewriter/rewriter.go` modification:**
```go
func (qr *QueryRewriter) RewriteQuery(ctx context.Context, query string, metricsCollector metrics.MetricsCollector) ([]string, error) {
    taskID := uuid.New().String()
    start := time.Now()

    // Existing logic
    variations, err := qr.generateVariations(query)

    // OPTIONAL: Record metrics (only if collector is not nil)
    if metricsCollector != nil {
        metricsCollector.RecordTask(metrics.TaskMetric{
            TaskID:     taskID,
            TaskType:   metrics.TaskTypeQueryRewrite,
            Provider:   "local_llm",
            Timestamp:  start,
            DurationMs: time.Since(start).Milliseconds(),
            Success:    err == nil,
            InputSize:  int64(len(query)),
            OutputSize: int64(len(variations)),
        })
    }

    return variations, err
}
```

**Key point:** Existing code still works if `metricsCollector` is nil or NoOp.

### Merge Strategy
1. Add new files (no conflicts)
2. Update `.env.example` (append only)
3. Modify `main.go` (add feature flag)
4. Update `handlercontext.go` (add optional field)
5. **Default:** Metrics disabled, zero performance impact

### Testing
```bash
# Test with metrics DISABLED (default)
METRICS_ENABLED=false go run vault-web-server/main.go
# Should work exactly as before

# Test with metrics ENABLED
METRICS_ENABLED=true go run vault-web-server/main.go
# Should log: [Metrics] Enabling performance tracking...
```

### Deliverable Checklist
- [ ] NoOpCollector implemented
- [ ] Environment variables added
- [ ] Main.go updated (feature-flagged)
- [ ] HandlerContext updated
- [ ] Example instrumentation (query rewriter)
- [ ] Tests pass with metrics enabled/disabled
- [ ] **Default behavior unchanged**

---

## PHASE 4: Provider Registry & Selection (Week 3-4)
**Branch:** `feature/provider-registry`
**Main Branch Impact:** 🟡 LOW (new parallel system, feature-flagged)

### Deliverables

#### 4.1 Provider Registry

**`/llm/registry/registry.go`:**
```go
package registry

import (
    "context"
    "fmt"
    "sync"

    "github.com/itaiguardiola/Askara/llm/unified"
    "github.com/itaiguardiola/Askara/metrics"
)

type ProviderRegistry struct {
    providers map[string]unified.UnifiedProvider
    metrics   metrics.MetricsCollector
    mu        sync.RWMutex
}

func NewProviderRegistry(metricsCollector metrics.MetricsCollector) *ProviderRegistry {
    return &ProviderRegistry{
        providers: make(map[string]unified.UnifiedProvider),
        metrics:   metricsCollector,
    }
}

func (r *ProviderRegistry) Register(provider unified.UnifiedProvider) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    name := provider.GetName()
    if _, exists := r.providers[name]; exists {
        return fmt.Errorf("provider %s already registered", name)
    }

    // Wrap with instrumentation
    instrumented := unified.NewInstrumentedProvider(provider, r.metrics)
    r.providers[name] = instrumented

    log.Printf("[Registry] Registered provider: %s (capabilities: %v)", name, provider.GetCapabilities())

    return nil
}

func (r *ProviderRegistry) Get(name string) (unified.UnifiedProvider, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    provider, exists := r.providers[name]
    if !exists {
        return nil, fmt.Errorf("provider %s not found", name)
    }

    return provider, nil
}

func (r *ProviderRegistry) List() []string {
    r.mu.RLock()
    defer r.mu.RUnlock()

    names := make([]string, 0, len(r.providers))
    for name := range r.providers {
        names = append(names, name)
    }
    return names
}

func (r *ProviderRegistry) GetByCapability(cap unified.Capability) []unified.UnifiedProvider {
    r.mu.RLock()
    defer r.mu.RUnlock()

    result := make([]unified.UnifiedProvider, 0)
    for _, provider := range r.providers {
        for _, c := range provider.GetCapabilities() {
            if c == cap {
                result = append(result, provider)
                break
            }
        }
    }

    return result
}

func (r *ProviderRegistry) HealthCheckAll(ctx context.Context) map[string]error {
    r.mu.RLock()
    defer r.mu.RUnlock()

    results := make(map[string]error)
    for name, provider := range r.providers {
        results[name] = provider.HealthCheck(ctx)
    }

    return results
}
```

#### 4.2 Simple Provider Selector

**`/llm/registry/selector.go`:**
```go
package registry

import (
    "context"
    "fmt"
    "time"

    "github.com/itaiguardiola/Askara/llm/unified"
    "github.com/itaiguardiola/Askara/metrics"
)

type SelectionCriteria struct {
    RequiredCapability unified.Capability
    MaxLatencyMs       int64
    MaxCostUSD         float64
    PreferredProvider  string
    OptimizeFor        OptimizationGoal
}

type OptimizationGoal string

const (
    OptimizeForCost    OptimizationGoal = "cost"
    OptimizeForSpeed   OptimizationGoal = "speed"
    OptimizeForBalance OptimizationGoal = "balance"
)

type ProviderSelector struct {
    registry *ProviderRegistry
    metrics  metrics.MetricsCollector
}

func NewProviderSelector(registry *ProviderRegistry, metricsCollector metrics.MetricsCollector) *ProviderSelector {
    return &ProviderSelector{
        registry: registry,
        metrics:  metricsCollector,
    }
}

func (s *ProviderSelector) SelectProvider(ctx context.Context, criteria SelectionCriteria) (unified.UnifiedProvider, error) {
    // Get candidates with required capability
    candidates := s.registry.GetByCapability(criteria.RequiredCapability)
    if len(candidates) == 0 {
        return nil, fmt.Errorf("no providers support capability: %s", criteria.RequiredCapability)
    }

    // If preferred provider specified and available, use it
    if criteria.PreferredProvider != "" {
        for _, candidate := range candidates {
            if candidate.GetName() == criteria.PreferredProvider {
                return candidate, nil
            }
        }
    }

    // Get recent performance metrics
    timeRange := metrics.TimeRange{
        Start: time.Now().Add(-1 * time.Hour),
        End:   time.Now(),
    }

    providerStats, _ := s.metrics.GetAllProviderStats(timeRange)

    // Score each candidate
    type scoredProvider struct {
        provider unified.UnifiedProvider
        score    float64
    }

    scored := make([]scoredProvider, 0)

    for _, candidate := range candidates {
        score := s.scoreProvider(candidate, criteria, providerStats)
        scored = append(scored, scoredProvider{
            provider: candidate,
            score:    score,
        })
    }

    // Sort by score (higher is better)
    sort.Slice(scored, func(i, j int) bool {
        return scored[i].score > scored[j].score
    })

    if len(scored) == 0 {
        return nil, fmt.Errorf("no suitable provider found")
    }

    selected := scored[0].provider

    // Record decision
    alternatives := make([]string, 0)
    for i := 1; i < len(scored) && i < 3; i++ {
        alternatives = append(alternatives, scored[i].provider.GetName())
    }

    s.metrics.RecordDecision(metrics.DecisionMetric{
        DecisionType:    "provider_selection",
        ChosenOption:    selected.GetName(),
        Alternatives:    alternatives,
        Reason:          fmt.Sprintf("optimize_for=%s", criteria.OptimizeFor),
    })

    return selected, nil
}

func (s *ProviderSelector) scoreProvider(
    provider unified.UnifiedProvider,
    criteria SelectionCriteria,
    stats map[string]*metrics.ProviderMetric,
) float64 {
    score := 100.0

    name := provider.GetName()
    pricing := provider.GetPricingModel()

    // Get recent stats if available
    providerStats, hasStats := stats[name]

    switch criteria.OptimizeFor {
    case OptimizeForCost:
        // Lower cost = higher score
        if pricing.Type == unified.PricingFree {
            score += 100
        } else {
            // Simple cost penalty
            score -= pricing.EmbeddingPer1KTokens * 10000
        }

    case OptimizeForSpeed:
        // Lower latency = higher score
        if hasStats {
            score -= float64(providerStats.P95LatencyMs) / 10
        }

    case OptimizeForBalance:
        // Balance cost and speed
        if pricing.Type == unified.PricingFree {
            score += 50
        } else {
            score -= pricing.EmbeddingPer1KTokens * 5000
        }

        if hasStats {
            score -= float64(providerStats.P95LatencyMs) / 20
        }
    }

    // Penalize if latency constraint violated
    if hasStats && criteria.MaxLatencyMs > 0 {
        if providerStats.P95LatencyMs > float64(criteria.MaxLatencyMs) {
            score -= 50
        }
    }

    // Penalize low uptime
    if hasStats && providerStats.Uptime < 95.0 {
        score -= (95.0 - providerStats.Uptime) * 2
    }

    return score
}
```

#### 4.3 Integration Example (Optional Usage)

**`vault-web-server/main.go` additions:**
```go
// OPTIONAL: Provider registry (only if PROVIDER_REGISTRY_ENABLED=true)
var providerRegistry *registry.ProviderRegistry
if os.Getenv("PROVIDER_REGISTRY_ENABLED") == "true" {
    log.Println("[Registry] Initializing provider registry...")

    providerRegistry = registry.NewProviderRegistry(metricsCollector)

    // Register existing OpenAI provider
    openaiAdapter := adapters.NewOpenAIAdapter(llmProvider.(*llm.OpenAIProvider))
    providerRegistry.Register(openaiAdapter)

    // Register Ollama if available
    if os.Getenv("LLM_PROVIDER") == "ollama" {
        ollamaAdapter := adapters.NewOllamaAdapter(llmProvider.(*llm.OllamaProvider))
        providerRegistry.Register(ollamaAdapter)
    }

    log.Printf("[Registry] Registered providers: %v", providerRegistry.List())
} else {
    log.Println("[Registry] Provider registry disabled")
}

handlerContext.ProviderRegistry = providerRegistry // Can be nil
```

### Testing
```bash
# Test registry independently
go test ./llm/registry/... -v

# Test with existing code (registry disabled)
PROVIDER_REGISTRY_ENABLED=false go run vault-web-server/main.go

# Test with registry enabled
PROVIDER_REGISTRY_ENABLED=true go run vault-web-server/main.go
```

### Deliverable Checklist
- [ ] Provider registry implemented
- [ ] Provider selector with scoring logic
- [ ] Health checking for all providers
- [ ] Optional integration in main.go
- [ ] Tests passing
- [ ] **Existing code unaffected when disabled**

---

## PHASE 5: Deep Integration Points (Week 4-5)
**Branch:** `feature/metrics-deep-integration`
**Main Branch Impact:** 🟡 MEDIUM (modifies existing handlers, but feature-flagged)

### Strategy
Instrument critical paths with metrics collection, all behind feature flags.

### Deliverables

#### 5.1 Question Handler Instrumentation

**`vault-web-server/postapi/questions.go` modifications:**

```go
func handleQuestion(w http.ResponseWriter, r *http.Request, ctx *HandlerContext) {
    // ... existing code for parsing request

    overallStart := time.Now()
    requestID := uuid.New().String()

    // OPTIONAL: Track query rewriting
    var queryVariations []string
    if ctx.QueryRewriter != nil && os.Getenv("QUERY_REWRITE_ENABLED") == "true" {
        rewriteStart := time.Now()
        queryVariations, err = ctx.QueryRewriter.RewriteQuery(r.Context(), question)

        if ctx.MetricsCollector != nil {
            ctx.MetricsCollector.RecordTask(metrics.TaskMetric{
                TaskID:     requestID + "_rewrite",
                TaskType:   metrics.TaskTypeQueryRewrite,
                Provider:   "local_llm",
                DurationMs: time.Since(rewriteStart).Milliseconds(),
                Success:    err == nil,
                UserUUID:   uuid,
            })
        }
    } else {
        queryVariations = []string{question}
    }

    // ... embedding generation (already instrumented if using unified provider)

    // OPTIONAL: Track vector search
    searchStart := time.Now()
    allMatches, err := ctx.VectorDB.HybridSearch(questionEmbedding, question, topK, uuid)

    if ctx.MetricsCollector != nil {
        ctx.MetricsCollector.RecordTask(metrics.TaskMetric{
            TaskID:     requestID + "_search",
            TaskType:   metrics.TaskTypeVectorSearch,
            Provider:   "qdrant", // or get from ctx
            DurationMs: time.Since(searchStart).Milliseconds(),
            Success:    err == nil,
            OutputSize: int64(len(allMatches)),
            UserUUID:   uuid,
        })
    }

    // ... rest of existing code

    // OPTIONAL: Track overall request
    if ctx.MetricsCollector != nil {
        ctx.MetricsCollector.RecordTask(metrics.TaskMetric{
            TaskID:     requestID,
            TaskType:   "question_answering",
            Provider:   "system",
            DurationMs: time.Since(overallStart).Milliseconds(),
            Success:    err == nil,
            UserUUID:   uuid,
            Metadata: map[string]interface{}{
                "query_variations": len(queryVariations),
                "matches_found":    len(allMatches),
                "streaming":        false,
            },
        })
    }
}
```

**Key Points:**
- Metrics collection is optional (checks `ctx.MetricsCollector != nil`)
- No change to logic flow
- Only timing/counting added

#### 5.2 ML-Worker Instrumentation

**`mlworker/client.go` modifications:**

```go
func (c *Client) Rerank(ctx context.Context, query string, documents []string, metricsCollector metrics.MetricsCollector) (*RerankResponse, error) {
    taskID := uuid.New().String()
    start := time.Now()

    // Existing API call logic
    resp, err := c.callRerankAPI(ctx, query, documents)

    // OPTIONAL: Record metrics
    if metricsCollector != nil {
        metricsCollector.RecordTask(metrics.TaskMetric{
            TaskID:     taskID,
            TaskType:   metrics.TaskTypeRerank,
            Provider:   "ml_worker",
            DurationMs: time.Since(start).Milliseconds(),
            Success:    err == nil,
            InputSize:  int64(len(documents)),
            OutputSize: int64(len(resp.Results)),
            Metadata: map[string]interface{}{
                "model_used":         resp.ModelUsed,
                "processing_time_ms": resp.ProcessingTimeMS,
            },
        })
    }

    return resp, err
}
```

#### 5.3 Document Processing Instrumentation

**`chunk/fileprocessing.go` modifications:**

```go
func ProcessDocument(content []byte, contentType string, metricsCollector metrics.MetricsCollector) (string, error) {
    taskID := uuid.New().String()
    overallStart := time.Now()

    var text string
    var err error
    var methodUsed string

    // Try ML-Worker
    if mlClient != nil && mlClient.IsFeatureEnabled("parse") {
        parseStart := time.Now()
        text, err = tryMLWorkerParsing(content, contentType)

        if metricsCollector != nil {
            metricsCollector.RecordTask(metrics.TaskMetric{
                TaskID:     taskID + "_mlworker",
                TaskType:   metrics.TaskTypeMLParse,
                Provider:   "ml_worker",
                DurationMs: time.Since(parseStart).Milliseconds(),
                Success:    err == nil,
                InputSize:  int64(len(content)),
                OutputSize: int64(len(text)),
            })
        }

        if err == nil {
            methodUsed = "ml_worker"
        }
    }

    // Fallback to docconv
    if text == "" {
        docconvStart := time.Now()
        text, err = docconvParse(content)

        if metricsCollector != nil {
            metricsCollector.RecordTask(metrics.TaskMetric{
                TaskID:     taskID + "_docconv",
                TaskType:   metrics.TaskTypeMLParse,
                Provider:   "docconv",
                DurationMs: time.Since(docconvStart).Milliseconds(),
                Success:    err == nil,
                InputSize:  int64(len(content)),
                OutputSize: int64(len(text)),
            })
        }

        if err == nil {
            methodUsed = "docconv"
        }
    }

    // Record decision
    if metricsCollector != nil {
        metricsCollector.RecordDecision(metrics.DecisionMetric{
            DecisionType: "document_parsing",
            ChosenOption: methodUsed,
            Alternatives: []string{"ml_worker", "docconv", "ocr"},
            Reason:       fmt.Sprintf("method_used=%s", methodUsed),
            Success:      err == nil,
            DurationMs:   time.Since(overallStart).Milliseconds(),
        })
    }

    return text, err
}
```

### Merge Strategy
1. **Review changes carefully** - These touch existing handlers
2. **Ensure backwards compatibility** - All metrics calls are nil-safe
3. **Feature flag everything** - METRICS_ENABLED controls all recording
4. **Test both modes:**
   - With metrics disabled: Performance unchanged
   - With metrics enabled: Data collected correctly

### Testing
```bash
# Integration test with metrics disabled
./scripts/test-streaming.sh
# Should pass as before

# Integration test with metrics enabled
METRICS_ENABLED=true ./scripts/test-streaming.sh
# Should pass AND generate metrics

# Verify metrics collected
ls ./metrics/
# Should see metrics_YYYY-MM-DD.csv
```

### Deliverable Checklist
- [ ] Question handler instrumented
- [ ] ML-Worker calls instrumented
- [ ] Document processing instrumented
- [ ] All metrics calls nil-safe
- [ ] Integration tests pass (both modes)
- [ ] **No performance degradation when disabled**

---

## PHASE 6: Metrics Dashboard & API (Week 5-6)
**Branch:** `feature/metrics-api`
**Main Branch Impact:** 🟢 LOW (new endpoints only)

### Deliverables

#### 6.1 Metrics API Endpoints

**`vault-web-server/postapi/metrics_api.go` (new file):**

```go
package postapi

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/itaiguardiola/Askara/metrics"
)

func handleGetProviderStats(w http.ResponseWriter, r *http.Request, ctx *HandlerContext) {
    if ctx.MetricsCollector == nil {
        http.Error(w, "Metrics not enabled", http.StatusNotImplemented)
        return
    }

    // Parse query params
    provider := r.URL.Query().Get("provider")
    hours := r.URL.Query().Get("hours")
    if hours == "" {
        hours = "24"
    }

    hoursInt, _ := strconv.Atoi(hours)

    timeRange := metrics.TimeRange{
        Start: time.Now().Add(-time.Duration(hoursInt) * time.Hour),
        End:   time.Now(),
    }

    var stats interface{}
    var err error

    if provider != "" {
        stats, err = ctx.MetricsCollector.GetProviderStats(provider, timeRange)
    } else {
        stats, err = ctx.MetricsCollector.GetAllProviderStats(timeRange)
    }

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}

func handleExportMetrics(w http.ResponseWriter, r *http.Request, ctx *HandlerContext) {
    if ctx.MetricsCollector == nil {
        http.Error(w, "Metrics not enabled", http.StatusNotImplemented)
        return
    }

    format := r.URL.Query().Get("format") // csv or json
    if format == "" {
        format = "csv"
    }

    hours := r.URL.Query().Get("hours")
    if hours == "" {
        hours = "24"
    }

    hoursInt, _ := strconv.Atoi(hours)

    timeRange := metrics.TimeRange{
        Start: time.Now().Add(-time.Duration(hoursInt) * time.Hour),
        End:   time.Now(),
    }

    tmpFile := fmt.Sprintf("/tmp/metrics_export_%d.%s", time.Now().Unix(), format)

    var err error
    if format == "csv" {
        err = ctx.MetricsCollector.ExportToCSV(tmpFile, timeRange)
        w.Header().Set("Content-Type", "text/csv")
    } else {
        err = ctx.MetricsCollector.ExportToJSON(tmpFile, timeRange)
        w.Header().Set("Content-Type", "application/json")
    }

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    http.ServeFile(w, r, tmpFile)
    os.Remove(tmpFile)
}

func handleProviderComparison(w http.ResponseWriter, r *http.Request, ctx *HandlerContext) {
    if ctx.MetricsCollector == nil {
        http.Error(w, "Metrics not enabled", http.StatusNotImplemented)
        return
    }

    // Get stats for all providers
    timeRange := metrics.TimeRange{
        Start: time.Now().Add(-24 * time.Hour),
        End:   time.Now(),
    }

    allStats, err := ctx.MetricsCollector.GetAllProviderStats(timeRange)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Format as comparison table
    comparison := make(map[string]interface{})

    for provider, stats := range allStats {
        comparison[provider] = map[string]interface{}{
            "requests":      stats.RequestCount,
            "success_rate":  stats.Uptime,
            "avg_latency":   stats.AvgLatencyMs,
            "p95_latency":   stats.P95LatencyMs,
            "total_cost":    stats.TotalCostUSD,
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(comparison)
}
```

**`vault-web-server/main.go` route additions:**

```go
// OPTIONAL: Metrics API endpoints (only if metrics enabled)
if os.Getenv("METRICS_ENABLED") == "true" {
    http.HandleFunc("/api/metrics/providers", func(w http.ResponseWriter, r *http.Request) {
        postapi.HandleGetProviderStats(w, r, handlerContext)
    })

    http.HandleFunc("/api/metrics/export", func(w http.ResponseWriter, r *http.Request) {
        postapi.HandleExportMetrics(w, r, handlerContext)
    })

    http.HandleFunc("/api/metrics/comparison", func(w http.ResponseWriter, r *http.Request) {
        postapi.HandleProviderComparison(w, r, handlerContext)
    })

    log.Println("[Metrics] API endpoints enabled at /api/metrics/*")
}
```

#### 6.2 Simple HTML Dashboard

**`vault-web-server/static/metrics.html` (new file):**

```html
<!DOCTYPE html>
<html>
<head>
    <title>Askara Metrics Dashboard</title>
    <style>
        body { font-family: monospace; padding: 20px; background: #1e1e1e; color: #d4d4d4; }
        table { border-collapse: collapse; width: 100%; margin: 20px 0; }
        th, td { border: 1px solid #444; padding: 12px; text-align: left; }
        th { background: #2d2d30; }
        .good { color: #4ec9b0; }
        .warn { color: #dcdcaa; }
        .bad { color: #f48771; }
        h1 { color: #569cd6; }
        .refresh-btn { background: #007acc; color: white; padding: 10px 20px; border: none; cursor: pointer; }
    </style>
</head>
<body>
    <h1>📊 Askara Performance Metrics</h1>
    <button class="refresh-btn" onclick="loadMetrics()">Refresh</button>

    <h2>Provider Comparison (Last 24h)</h2>
    <div id="comparison"></div>

    <h2>All Providers</h2>
    <div id="providers"></div>

    <script>
        function loadMetrics() {
            fetch('/api/metrics/comparison')
                .then(r => r.json())
                .then(data => {
                    renderComparison(data);
                });

            fetch('/api/metrics/providers')
                .then(r => r.json())
                .then(data => {
                    renderProviders(data);
                });
        }

        function renderComparison(data) {
            let html = '<table><tr><th>Provider</th><th>Requests</th><th>Success Rate</th><th>Avg Latency</th><th>P95 Latency</th><th>Total Cost</th></tr>';

            for (const [provider, stats] of Object.entries(data)) {
                const successClass = stats.success_rate > 95 ? 'good' : stats.success_rate > 90 ? 'warn' : 'bad';
                const latencyClass = stats.p95_latency < 500 ? 'good' : stats.p95_latency < 2000 ? 'warn' : 'bad';

                html += `<tr>
                    <td><strong>${provider}</strong></td>
                    <td>${stats.requests}</td>
                    <td class="${successClass}">${stats.success_rate.toFixed(2)}%</td>
                    <td class="${latencyClass}">${stats.avg_latency.toFixed(0)}ms</td>
                    <td class="${latencyClass}">${stats.p95_latency.toFixed(0)}ms</td>
                    <td>$${stats.total_cost.toFixed(4)}</td>
                </tr>`;
            }

            html += '</table>';
            document.getElementById('comparison').innerHTML = html;
        }

        function renderProviders(data) {
            let html = '<table><tr><th>Provider</th><th>Requests</th><th>Success/Error</th><th>Uptime</th><th>P50/P95/P99</th></tr>';

            for (const [provider, stats] of Object.entries(data)) {
                html += `<tr>
                    <td><strong>${provider}</strong></td>
                    <td>${stats.RequestCount}</td>
                    <td>${stats.SuccessCount} / ${stats.ErrorCount}</td>
                    <td class="${stats.Uptime > 95 ? 'good' : 'warn'}">${stats.Uptime.toFixed(2)}%</td>
                    <td>${stats.P50LatencyMs.toFixed(0)} / ${stats.P95LatencyMs.toFixed(0)} / ${stats.P99LatencyMs.toFixed(0)} ms</td>
                </tr>`;
            }

            html += '</table>';
            document.getElementById('providers').innerHTML = html;
        }

        // Load on page load
        loadMetrics();

        // Auto-refresh every 30 seconds
        setInterval(loadMetrics, 30000);
    </script>
</body>
</html>
```

**Route for dashboard:**

```go
http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "vault-web-server/static/metrics.html")
})
```

### Testing
```bash
# Start server with metrics
METRICS_ENABLED=true go run vault-web-server/main.go

# Generate some traffic
curl -X POST http://localhost:8100/api/questions -d '{"question":"test", "uuid":"test"}'

# View dashboard
open http://localhost:8100/metrics

# Test API
curl http://localhost:8100/api/metrics/providers
curl http://localhost:8100/api/metrics/comparison
curl http://localhost:8100/api/metrics/export?format=csv > metrics.csv
```

### Deliverable Checklist
- [ ] Metrics API endpoints implemented
- [ ] HTML dashboard created
- [ ] Export functionality working
- [ ] Provider comparison API
- [ ] Documentation updated
- [ ] **All optional, no impact on existing endpoints**

---

## PHASE 7: Production Deployment (Week 6+)
**Branch:** Merge to `main`
**Main Branch Impact:** 🟢 MINIMAL (well-tested, feature-flagged)

### Pre-Merge Checklist

**Code Quality:**
- [ ] All tests passing
- [ ] Code coverage >80% for new packages
- [ ] No linter warnings
- [ ] Documentation complete

**Backwards Compatibility:**
- [ ] Existing tests still pass
- [ ] Performance benchmarks show no degradation when disabled
- [ ] Default behavior unchanged (metrics disabled)

**Configuration:**
- [ ] `.env.example` updated
- [ ] README updated with metrics documentation
- [ ] Migration guide for teams

**Testing:**
- [ ] Integration tests with metrics enabled
- [ ] Integration tests with metrics disabled
- [ ] Load testing (no performance impact)
- [ ] Memory leak testing (long-running collection)

### Merge Strategy

**Step 1: Merge infrastructure (low risk)**
```bash
git checkout main
git merge feature/metrics-framework-standalone
git merge feature/unified-provider-interface
# These add new packages, zero conflicts
```

**Step 2: Merge optional integrations**
```bash
git merge feature/optional-metrics-integration
git merge feature/provider-registry
# Feature-flagged additions
```

**Step 3: Merge deep integrations**
```bash
git merge feature/metrics-deep-integration
# Carefully review handler changes
# Ensure nil-safety
```

**Step 4: Merge dashboard**
```bash
git merge feature/metrics-api
# New endpoints only
```

### Post-Merge Validation

**Smoke Tests:**
```bash
# Test 1: Default behavior (metrics disabled)
go run vault-web-server/main.go
# Should NOT see metrics logs
# Performance should be unchanged

# Test 2: Enable metrics
METRICS_ENABLED=true go run vault-web-server/main.go
# Should see: [Metrics] Enabling performance tracking...

# Test 3: Upload document + ask question
# Check metrics collected

# Test 4: View dashboard
open http://localhost:8100/metrics
```

**Performance Validation:**
```bash
# Benchmark without metrics
METRICS_ENABLED=false go test -bench=. -benchtime=10s

# Benchmark with metrics
METRICS_ENABLED=true go test -bench=. -benchtime=10s

# Compare results - should be <5% difference
```

### Rollout Plan

**Week 1: Canary deployment**
- Enable metrics on 1 instance
- Monitor for 7 days
- Check for memory leaks, performance issues

**Week 2: Gradual rollout**
- Enable on 25% of instances
- Validate data quality
- Check storage requirements

**Week 3: Full deployment**
- Enable on all instances
- Set up automated exports
- Train team on dashboard

**Week 4: Optimization**
- Analyze collected data
- Tune retention policies
- Optimize storage

### Documentation

**Update `README.md`:**

```markdown
## Performance Metrics (Optional)

Askara includes optional performance tracking for AI providers and ML-Worker operations.

### Enable Metrics

Add to `.env`:
```bash
METRICS_ENABLED=true
METRICS_EXPORT_PATH=./metrics
METRICS_RETENTION_DAYS=30
```

### View Dashboard

Visit `http://localhost:8100/metrics` to see real-time provider comparison.

### Export Data

```bash
# CSV export
curl http://localhost:8100/api/metrics/export?format=csv > metrics.csv

# JSON export
curl http://localhost:8100/api/metrics/export?format=json > metrics.json
```

### API Endpoints

- `GET /api/metrics/providers` - Provider statistics
- `GET /api/metrics/comparison` - Side-by-side comparison
- `GET /api/metrics/export` - Export historical data
```

### Deliverable Checklist
- [ ] All phases merged to main
- [ ] Tests passing in main branch
- [ ] Documentation updated
- [ ] Team trained on new features
- [ ] Monitoring set up
- [ ] **Production ready**

---

## SUCCESS CRITERIA

**Phase 1-2:** ✅ Independent packages working
- Metrics collection functional
- Provider interfaces defined
- Zero main branch changes

**Phase 3-4:** ✅ Optional integration
- Can enable/disable metrics
- Provider registry functional
- Existing code unaffected

**Phase 5-6:** ✅ Full instrumentation
- All key paths instrumented
- Dashboard functional
- API endpoints working

**Phase 7:** ✅ Production deployment
- No performance degradation
- Data collection validated
- Team adoption

---

## COORDINATION WITH GEMINI TEAM

**Shared Deliverables:**
- `llm/unified/interface.go` - Both teams implement this
- `metrics/types.go` - Shared metric schema
- Test suite - Common benchmarks

**Communication:**
- Weekly sync meetings
- Shared Slack channel for metrics discussions
- Joint dashboard for provider comparison

**Integration Points:**
- Gemini team implements `UnifiedProvider` interface
- Both register providers in same registry
- Metrics automatically collected for all providers

---

## TIMELINE SUMMARY

| Week | Phase | Impact | Merge Risk |
|------|-------|--------|-----------|
| 1 | Phase 1-2 | 🟢 Zero | 🟢 None |
| 2-3 | Phase 3-4 | 🟡 Low | 🟡 Low |
| 4-5 | Phase 5-6 | 🟡 Medium | 🟡 Medium |
| 6+ | Phase 7 | 🟢 Minimal | 🟢 Low |

**Total: 6-8 weeks to full production deployment**

**Quick wins available in Week 2:** Dashboard showing current provider performance!
