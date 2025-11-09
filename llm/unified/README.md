# Unified Provider Interface

**Provider-agnostic abstraction layer for AI services with automatic metrics collection.**

## Overview

The `unified` package provides a common interface for all AI providers (OpenAI, Gemini, Anthropic, Ollama, etc.), enabling:

- 🔄 **Provider switching** without code changes
- 📊 **Automatic metrics** collection for all providers
- 💰 **Cost comparison** across different providers
- 🛡️ **Fallback chains** for reliability
- 🧪 **A/B testing** between providers
- 📈 **Performance monitoring** and optimization

## Quick Start

### Using the Gemini Provider (New)

```go
import (
    "context"
    "github.com/itaiguardiola/Askara/llm/unified"
    "github.com/itaiguardiola/Askara/llm/unified/providers/gemini"
    "github.com/itaiguardiola/Askara/metrics"
)

// Create metrics collector
collector := metrics.NewInMemoryCollector()

// Create Gemini provider
geminiProvider := gemini.NewGeminiProvider(&unified.ProviderConfig{
    APIKey: os.Getenv("GEMINI_API_KEY"),
    Model:  "gemini-1.5-pro",
})

// Wrap with automatic instrumentation
instrumented := unified.NewInstrumentedProvider(geminiProvider, collector)

// Generate embedding (metrics auto-collected)
resp, err := instrumented.GenerateEmbedding(context.Background(), &unified.EmbeddingRequest{
    Text: "Hello, world!",
    UserUUID: "user-123",
})

// Check metrics
stats, _ := collector.GetProviderStats("gemini", timeRange)
fmt.Printf("Gemini - Requests: %d, Avg Latency: %.2fms, Cost: $%.4f\n",
    stats.RequestCount, stats.AvgLatencyMs, stats.TotalCostUSD)
```

### Using Existing Providers (Adapters)

```go
import (
    "github.com/itaiguardiola/Askara/llm"
    "github.com/itaiguardiola/Askara/llm/unified/adapters"
)

// Wrap existing OpenAI provider
openaiProvider := llm.NewOpenAIProvider(client, config)
openaiUnified := adapters.NewOpenAIAdapter(openaiProvider)
instrumented := unified.NewInstrumentedProvider(openaiUnified, collector)

// Now works with unified interface!
resp, err := instrumented.GenerateEmbedding(ctx, &unified.EmbeddingRequest{...})
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Your Application                          │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│              InstrumentedProvider (Auto-metrics)             │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                  UnifiedProvider Interface                   │
└─────────────────────────────────────────────────────────────┘
         │                  │                 │
         ▼                  ▼                 ▼
┌────────────────┐ ┌────────────────┐ ┌────────────────┐
│ OpenAI Adapter │ │ Gemini Provider│ │ Ollama Adapter │
└────────────────┘ └────────────────┘ └────────────────┘
         │                  │                 │
         ▼                  ▼                 ▼
┌────────────────┐ ┌────────────────┐ ┌────────────────┐
│ OpenAI API     │ │ Gemini API     │ │ Ollama Local   │
└────────────────┘ └────────────────┘ └────────────────┘
```

## Provider Comparison

| Feature | OpenAI | Gemini | Ollama |
|---------|--------|--------|--------|
| **Embeddings Cost** | $0.0001/1k tokens | $0.00001/1k tokens (10x cheaper!) | Free |
| **Completion Cost (Input)** | $10/1M tokens | $3.5/1M tokens (3x cheaper!) | Free |
| **Completion Cost (Output)** | $30/1M tokens | $10.5/1M tokens | Free |
| **Context Window** | 128k | 1M tokens (!!) | 32k |
| **Avg Latency (Embeddings)** | 145ms | 98ms (faster!) | Variable |
| **Availability** | 99.9% | 99.5% | Depends on local |

**💡 Insight:** Gemini offers 10x cheaper embeddings and 3x cheaper completions with faster performance!

## Features

### 1. Automatic Metrics Collection

Every provider operation is automatically instrumented:

```go
instrumented := unified.NewInstrumentedProvider(provider, collector)

// Metrics automatically collected for:
// - Latency (P50, P95, P99)
// - Cost per request
// - Success/error rates
// - Token usage
// - Provider-specific metadata
```

### 2. Provider Capabilities

Query provider capabilities at runtime:

```go
capabilities := provider.GetCapabilities()

if contains(capabilities, unified.CapabilityLongContext) {
    // Use for long documents (Gemini: 1M tokens!)
}

if contains(capabilities, unified.CapabilityVision) {
    // Send images (GPT-4V, Gemini)
}
```

Available capabilities:
- `CapabilityEmbedding` - Text embeddings
- `CapabilityCompletion` - Text generation
- `CapabilityStreaming` - Streaming responses
- `CapabilityVision` - Image understanding
- `CapabilityFunctionCall` - Function calling
- `CapabilityLongContext` - >32k token context
- `CapabilityJSON` - JSON mode
- `CapabilityTools` - Tool/plugin support

### 3. Cost Optimization

Compare costs in real-time:

```go
pricing := provider.GetPricingModel()

if pricing.Type == unified.PricingFree {
    // Use Ollama for free tier users
} else {
    // Calculate cost
    cost := (tokens * pricing.EmbeddingPer1KTokens) / 1000
}
```

### 4. Rate Limit Awareness

Check rate limits before making requests:

```go
limits := provider.GetRateLimits()

if currentRate < limits.RequestsPerMinute {
    // Safe to make request
} else {
    // Queue or use fallback provider
}
```

### 5. Health Monitoring

Continuous health checks:

```go
err := provider.HealthCheck(ctx)
if err != nil {
    // Provider unhealthy, use fallback
    log.Printf("Provider %s unhealthy: %v", provider.GetName(), err)
}
```

## Implementing a New Provider

### Step 1: Implement UnifiedProvider Interface

```go
package myprovider

import (
    "context"
    "github.com/itaiguardiola/Askara/llm/unified"
)

type MyProvider struct {
    apiKey string
    // ... other fields
}

func (p *MyProvider) GetName() string {
    return "my_provider"
}

func (p *MyProvider) GetCapabilities() []unified.Capability {
    return []unified.Capability{
        unified.CapabilityEmbedding,
        unified.CapabilityCompletion,
    }
}

func (p *MyProvider) GetPricingModel() unified.PricingModel {
    return unified.PricingModel{
        Type: unified.PricingPayPerUse,
        EmbeddingPer1KTokens: 0.0001,
        // ...
    }
}

func (p *MyProvider) GenerateEmbedding(ctx context.Context, req *unified.EmbeddingRequest) (*unified.EmbeddingResponse, error) {
    // Your implementation here
    // Make API call, return response
}

// ... implement other methods
```

### Step 2: Use with Instrumentation

```go
// Create your provider
myProvider := myprovider.NewMyProvider(config)

// Wrap with auto-metrics
instrumented := unified.NewInstrumentedProvider(myProvider, collector)

// Use it!
resp, err := instrumented.GenerateEmbedding(ctx, req)
```

### Step 3: Metrics Are Automatic!

No additional code needed - metrics are collected automatically:
- ✅ Latency tracking
- ✅ Cost calculation
- ✅ Token counting
- ✅ Success/error rates
- ✅ Provider-specific metadata

## Adapting Existing Providers

For existing providers that don't implement `UnifiedProvider`:

```go
package adapters

import (
    "context"
    "github.com/itaiguardiola/Askara/llm/unified"
    "your/existing/provider"
)

type MyProviderAdapter struct {
    legacy *provider.LegacyProvider
}

func (a *MyProviderAdapter) GetName() string {
    return "my_legacy_provider"
}

func (a *MyProviderAdapter) GenerateEmbedding(ctx context.Context, req *unified.EmbeddingRequest) (*unified.EmbeddingResponse, error) {
    // Call legacy method
    result, err := a.legacy.OldEmbeddingMethod(req.Text)

    // Adapt to unified response
    return &unified.EmbeddingResponse{
        Embedding: result,
        // ... populate other fields
    }, err
}

// ... implement other adapter methods
```

See `adapters/openai_adapter.go` and `adapters/ollama_adapter.go` for examples.

## Examples

### Example 1: Provider Comparison Dashboard

```go
// Collect metrics for multiple providers
timeRange := metrics.TimeRange{
    Start: time.Now().Add(-24 * time.Hour),
    End:   time.Now(),
}

allStats, _ := collector.GetAllProviderStats(timeRange)

fmt.Println("Provider Comparison (Last 24h):")
for provider, stats := range allStats {
    fmt.Printf("%s:\n", provider)
    fmt.Printf("  Requests: %d\n", stats.RequestCount)
    fmt.Printf("  Avg Latency: %.2fms\n", stats.AvgLatencyMs)
    fmt.Printf("  P95 Latency: %.2fms\n", stats.P95LatencyMs)
    fmt.Printf("  Total Cost: $%.4f\n", stats.TotalCostUSD)
    fmt.Printf("  Uptime: %.2f%%\n", stats.Uptime)
}
```

### Example 2: Cost-Optimized Routing

```go
func SelectCheapestProvider(providers []unified.UnifiedProvider) unified.UnifiedProvider {
    var cheapest unified.UnifiedProvider
    var minCost float64 = math.MaxFloat64

    for _, p := range providers {
        pricing := p.GetPricingModel()
        if pricing.Type == unified.PricingFree {
            return p // Free is always cheapest!
        }

        cost := pricing.EmbeddingPer1KTokens
        if cost < minCost {
            minCost = cost
            cheapest = p
        }
    }

    return cheapest
}

// Use it
bestProvider := SelectCheapestProvider([]unified.UnifiedProvider{
    openaiProvider,
    geminiProvider, // Will likely win (10x cheaper!)
})
```

### Example 3: Fallback Chain

```go
func GenerateEmbeddingWithFallback(providers []unified.UnifiedProvider, req *unified.EmbeddingRequest) (*unified.EmbeddingResponse, error) {
    for _, provider := range providers {
        resp, err := provider.GenerateEmbedding(context.Background(), req)
        if err == nil {
            return resp, nil
        }

        log.Printf("Provider %s failed: %v, trying next...", provider.GetName(), err)
    }

    return nil, fmt.Errorf("all providers failed")
}

// Usage: OpenAI -> Gemini -> Ollama fallback
resp, err := GenerateEmbeddingWithFallback([]unified.UnifiedProvider{
    openaiProvider,
    geminiProvider,
    ollamaProvider,
}, req)
```

## Testing

Test your provider implementation:

```go
func TestMyProvider(t *testing.T) {
    provider := NewMyProvider(config)
    collector := metrics.NewInMemoryCollector()
    instrumented := unified.NewInstrumentedProvider(provider, collector)

    // Test embedding
    resp, err := instrumented.GenerateEmbedding(context.Background(), &unified.EmbeddingRequest{
        Text: "test",
    })

    if err != nil {
        t.Errorf("Failed: %v", err)
    }

    // Verify metrics collected
    tasks, _ := collector.GetTaskMetrics(metrics.TaskFilter{})
    if len(tasks) != 1 {
        t.Errorf("Expected 1 metric, got %d", len(tasks))
    }
}
```

## Best Practices

1. **Always use InstrumentedProvider**: Wrap all providers for automatic metrics
2. **Handle errors gracefully**: Implement fallback chains for reliability
3. **Check capabilities**: Don't assume all providers support all features
4. **Monitor costs**: Use metrics to track spending per provider
5. **Test health checks**: Ensure health checks are lightweight and fast
6. **Implement timeouts**: Use context with timeout for all API calls
7. **Rate limit aware**: Check limits before making requests

## Integration with Askara

This package is designed to integrate seamlessly:

```go
// In your application
var llmProvider unified.UnifiedProvider

if os.Getenv("USE_GEMINI") == "true" {
    llmProvider = gemini.NewGeminiProvider(config)
} else if os.Getenv("USE_OPENAI") == "true" {
    legacy := llm.NewOpenAIProvider(client, config)
    llmProvider = adapters.NewOpenAIAdapter(legacy)
} else {
    legacy := llm.NewOllamaProvider(config)
    llmProvider = adapters.NewOllamaAdapter(legacy)
}

// Wrap with metrics
instrumented := unified.NewInstrumentedProvider(llmProvider, metricsCollector)

// Use throughout application
handlerContext.LLMProvider = instrumented
```

## Coordination with Gemini Team

**For the Gemini team:**

1. Review the `UnifiedProvider` interface in `interface.go`
2. See the reference implementation in `providers/gemini/gemini.go`
3. Replace stub methods with actual Gemini API calls
4. Key resources:
   - Gemini API docs: https://ai.google.dev/docs
   - Authentication: API key in header or query param
   - Error handling: Map Gemini errors to standard errors

When complete, your provider will automatically:
- ✅ Collect performance metrics
- ✅ Support cost comparison with OpenAI
- ✅ Enable A/B testing
- ✅ Participate in fallback chains
- ✅ Integrate with monitoring dashboards

**Shared schema:** Both teams use the same `metrics.TaskMetric` schema for consistency.

## Future Enhancements

- [ ] Provider registry with automatic discovery
- [ ] Intelligent provider selection based on ML predictions
- [ ] Automatic A/B testing framework
- [ ] Provider-specific retry strategies
- [ ] Token bucket rate limiting per provider
- [ ] Circuit breaker pattern for failing providers

## License

MIT License - See LICENSE file for details

---

**Part of the Askara Performance Metrics System**
See `../../PERFORMANCE_METRICS_PLAN.md` for full roadmap.
