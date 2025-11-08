# LLM Provider System - Complete Guide

## Overview

Askara now features an advanced **multi-provider LLM system** with intelligent auto-selection, comprehensive analytics, and automatic failover capabilities. This allows you to:

- 🎯 **Use multiple AI providers** simultaneously
- 📊 **Track performance metrics** (latency, costs, success rates)
- 🤖 **Auto-select the best provider** based on real-time analytics
- 🔄 **Automatic failover** when providers fail
- 💰 **Monitor costs** across all providers
- 🏥 **Health monitoring** for all providers

---

## Supported Providers

| Provider | Type | Cost | Speed | Embeddings | Best For |
|----------|------|------|-------|------------|----------|
| **Ollama** | Local | FREE | Medium | ✅ Yes | Privacy, development, RTX GPUs |
| **OpenAI** | Cloud | $$$ | Medium | ✅ Yes | Production, highest quality |
| **Claude** | Cloud | $$ | Fast | ✅ Yes (Voyage) | Reasoning, long context |
| **Gemini** | Cloud | $ | Very Fast | ✅ Yes | Speed + affordability |
| **Groq** | Cloud | $ | ULTRA Fast | ❌ No | Ultra-low latency completions |

---

## Quick Start

### 1. Single Provider Mode (Simple)

Use one provider at a time (traditional mode):

```bash
# .env
LLM_PROVIDER=claude
LLM_PROVIDER_MODE=single
CLAUDE_API_KEY=your_api_key_here
```

### 2. Auto-Selection Mode (Recommended)

Let Askara automatically pick the best provider based on performance:

```bash
# .env
LLM_PROVIDER_MODE=auto
LLM_METRICS_ENABLED=true

# Configure multiple providers
OPENAI_API_KEY=your_openai_key
CLAUDE_API_KEY=your_claude_key
GEMINI_API_KEY=your_gemini_key
```

Askara will:
- Track latency, success rate, and costs for each provider
- Automatically route requests to the best performing provider
- Skip unhealthy providers
- Optimize for speed, reliability, and cost

### 3. Failover Mode

Try providers in order until one succeeds:

```bash
# .env
LLM_PROVIDER_MODE=failover

# Primary, secondary, tertiary providers
OPENAI_API_KEY=your_openai_key
CLAUDE_API_KEY=your_claude_key
OLLAMA_HOST=http://localhost:11434  # Fallback to local
```

---

## Provider Configuration

### Ollama (Local, FREE)

**Best for:** Development, privacy, RTX GPU owners

```bash
# .env
LLM_PROVIDER=ollama
OLLAMA_HOST=http://localhost:11434
OLLAMA_MODEL=llama2
OLLAMA_EMBEDDING_MODEL=nomic-embed-text
```

**Setup:**
```bash
# Install Ollama
curl https://ollama.ai/install.sh | sh

# Pull models
ollama pull llama2
ollama pull nomic-embed-text

# Start server
ollama serve
```

**GPU Support:** Automatic with NVIDIA GPUs (CUDA), AMD (ROCm), Metal (Mac)

---

### OpenAI (Cloud, Paid)

**Best for:** Production applications, highest quality

```bash
# .env
LLM_PROVIDER=openai
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-4
OPENAI_EMBEDDING_MODEL=text-embedding-ada-002
```

**Pricing (approximate):**
- GPT-4: $0.03/1K input, $0.06/1K output
- GPT-3.5: $0.0015/1K input, $0.002/1K output
- Embeddings: $0.0001/1K tokens

**Get API Key:** https://platform.openai.com/api-keys

---

### Anthropic Claude (Cloud, Paid)

**Best for:** Reasoning, long conversations, detailed analysis

```bash
# .env
LLM_PROVIDER=claude
CLAUDE_API_KEY=sk-ant-...
CLAUDE_MODEL=claude-3-5-sonnet-20241022
CLAUDE_EMBED_MODEL=voyage-large-2-instruct
```

**Pricing:**
- Claude 3.5 Sonnet: $0.003/1K input, $0.015/1K output
- Embeddings (Voyage): ~$0.0001/1K tokens

**Features:**
- 200K context window
- Excellent reasoning capabilities
- Constitutional AI safety

**Get API Key:** https://console.anthropic.com/

---

### Google Gemini (Cloud, Paid)

**Best for:** Speed + affordability, multimodal tasks

```bash
# .env
LLM_PROVIDER=gemini
GEMINI_API_KEY=AIza...
GEMINI_MODEL=gemini-1.5-flash
GEMINI_EMBED_MODEL=text-embedding-004
```

**Pricing:**
- Gemini 1.5 Flash: $0.00025/1K input, $0.0005/1K output
- Gemini 1.5 Pro: $0.0035/1K input, $0.0105/1K output
- Embeddings: $0.00001/1K tokens

**Features:**
- Very fast inference
- Lowest cost per token
- 1M token context window (Pro)

**Get API Key:** https://makersuite.google.com/app/apikey

---

### Groq (Cloud, Paid/Free Tier)

**Best for:** Ultra-low latency, real-time applications

```bash
# .env
LLM_PROVIDER=groq
GROQ_API_KEY=gsk_...
GROQ_MODEL=mixtral-8x7b-32768

# Groq doesn't provide embeddings - use another provider
OLLAMA_HOST=http://localhost:11434  # For embeddings
```

**Pricing:** Free tier available, then pay-as-you-go

**Features:**
- **Fastest inference available** (350+ tokens/sec)
- LPU (Language Processing Unit) hardware
- OpenAI-compatible API

**Limitations:**
- No embedding support (use Ollama/OpenAI for embeddings)
- Smaller model selection

**Get API Key:** https://console.groq.com/

---

## Analytics & Metrics

### Viewing Metrics

**API Endpoints:**

```bash
# Get all provider statistics
GET /api/providers/stats

# Get recent metrics (last 100 requests)
GET /api/providers/metrics

# Get provider health status
GET /api/providers/health
```

**Example Response:**
```json
{
  "providers": {
    "claude": {
      "total_requests": 1500,
      "successful_requests": 1485,
      "success_rate": 99.0,
      "avg_latency_ms": 850,
      "p95_latency_ms": 1200,
      "total_cost": 0.45,
      "is_healthy": true,
      "health_score": 95.5
    },
    "gemini": {
      "total_requests": 2000,
      "successful_requests": 1990,
      "success_rate": 99.5,
      "avg_latency_ms": 450,
      "p95_latency_ms": 600,
      "total_cost": 0.12,
      "is_healthy": true,
      "health_score": 98.2
    }
  },
  "mode": "auto",
  "active_provider": "gemini"
}
```

### Metrics Tracked

For each provider, Askara tracks:

- **Performance:**
  - Request count
  - Success/failure rate
  - Average latency
  - P95/P99 latency percentiles

- **Cost:**
  - Total tokens (input/output)
  - Estimated cost per provider
  - Cost per request

- **Health:**
  - Recent failure rate
  - Time since last success
  - Overall health score (0-100)

### Metrics Storage

Metrics are persisted to disk for analysis:

```bash
# Default location
./data/llm_metrics.json

# Configure custom path
LLM_METRICS_PATH=./custom/path/metrics.json
```

Metrics are automatically:
- Saved every 10 requests
- Loaded on startup
- Trimmed to last 10,000 entries

---

## Auto-Selection Algorithm

When `LLM_PROVIDER_MODE=auto`, Askara uses a weighted scoring algorithm:

### Selection Criteria

| Factor | Weight | Description |
|--------|--------|-------------|
| Health Score | 40% | Overall provider health (0-100) |
| Latency | 30% | Lower latency = higher score |
| Success Rate | 20% | Recent success percentage |
| Cost | 10% | Lower cost = higher score |

### Health Calculation

Provider health (0-100) considers:

1. **Success Rate (40%):** Recent request success percentage
2. **Latency Score (30%):** Speed (100ms = best, 5000ms = worst)
3. **Recency Score (30%):** Time since last successful request

### Provider Selection Process

```
1. Get metrics for all registered providers
2. Calculate weighted score for each provider
3. Skip providers with health_score < 50
4. Select provider with highest score
5. If all unhealthy, use fallback order
```

### Example Auto-Selection

```
Provider Scores (higher is better):
┌─────────┬────────┬─────────┬──────────┬──────┬─────────┐
│ Provider│ Health │ Latency │ Success  │ Cost │ Final   │
│         │ (40%)  │ (30%)   │ (20%)    │ (10%)│ Score   │
├─────────┼────────┼─────────┼──────────┼──────┼─────────┤
│ Gemini  │  38.0  │  27.5   │  19.9    │ 9.8  │  95.2 ✓ │
│ Claude  │  36.5  │  24.0   │  19.8    │ 9.5  │  89.8   │
│ OpenAI  │  35.0  │  22.0   │  19.5    │ 8.0  │  84.5   │
│ Ollama  │  30.0  │  18.0   │  18.0    │ 10.0 │  76.0   │
└─────────┴────────┴─────────┴──────────┴──────┴─────────┘

→ Selected: Gemini (best overall score)
```

---

## Failover & Reliability

### Automatic Failover

In `failover` mode, providers are tried in order:

```go
// Automatic fallback chain
1. Try primary provider (e.g., OpenAI)
2. On failure, try secondary (e.g., Claude)
3. On failure, try tertiary (e.g., Gemini)
4. On failure, try local fallback (e.g., Ollama)
```

### Health Monitoring

Providers are marked unhealthy if:
- Recent failure rate > 50% (last 5 minutes)
- No successful request in last 24 hours
- Health score < 50

Unhealthy providers are automatically skipped in auto-selection.

---

## Cost Management

### Cost Tracking

View real-time cost estimates:

```bash
curl http://localhost:8100/api/providers/stats
```

Response includes per-provider costs:
```json
{
  "claude": {
    "total_requests": 1000,
    "total_cost": 0.45
  },
  "gemini": {
    "total_requests": 2000,
    "total_cost": 0.12
  }
}
```

### Cost Optimization Tips

1. **Use Gemini for bulk operations** (lowest cost)
2. **Use Groq for real-time interactions** (fast + cheap)
3. **Use Claude for complex reasoning** (expensive but high quality)
4. **Use Ollama for development** (FREE)
5. **Enable auto-selection with 10% cost weight** for cost-aware routing

---

## Architecture

### Component Overview

```
┌──────────────────────────────────────────────────────────┐
│                    ProviderManager                        │
│  ┌────────────┐  ┌────────────────┐  ┌────────────────┐ │
│  │   Single   │  │  Auto-Selector │  │    Failover    │ │
│  │    Mode    │  │   (Analytics)  │  │      Mode      │ │
│  └────────────┘  └────────────────┘  └────────────────┘ │
└──────────────────────────────────────────────────────────┘
           │                    │                    │
           ▼                    ▼                    ▼
┌──────────────────────────────────────────────────────────┐
│              InstrumentedProviders (Metrics)              │
├─────────┬─────────┬─────────┬─────────┬──────────────────┤
│ Ollama  │ OpenAI  │ Claude  │ Gemini  │ Groq             │
└─────────┴─────────┴─────────┴─────────┴──────────────────┘
           │
           ▼
┌──────────────────────────────────────────────────────────┐
│             MetricsCollector                              │
│  - Latency tracking                                       │
│  - Success/failure rates                                  │
│  - Token counting                                         │
│  - Cost estimation                                        │
│  - Health scoring                                         │
└──────────────────────────────────────────────────────────┘
```

### Files Structure

```
llm/
├── provider.go              # LLMProvider interface
├── factory.go               # Provider factory (creates providers)
├── config.go                # Configuration structs
│
├── ollama.go                # Ollama implementation
├── openai.go                # OpenAI implementation
├── claude.go                # Claude implementation ✨ NEW
├── gemini.go                # Gemini implementation ✨ NEW
├── groq.go                  # Groq implementation ✨ NEW
│
├── metrics.go               # Metrics collection ✨ NEW
├── instrumented_provider.go # Provider wrapper with metrics ✨ NEW
├── provider_manager.go      # Multi-provider management ✨ NEW
│
└── README_PROVIDERS.md      # This file
```

---

## API Reference

### Provider Management Endpoints

#### GET `/api/providers/stats`

Returns statistics for all providers.

**Response:**
```json
{
  "providers": {
    "provider_name": {
      "total_requests": 1000,
      "successful_requests": 995,
      "failed_requests": 5,
      "success_rate": 99.5,
      "avg_latency_ms": 450.5,
      "p95_latency_ms": 650.0,
      "p99_latency_ms": 800.0,
      "total_tokens_in": 50000,
      "total_tokens_out": 75000,
      "total_cost": 0.45,
      "is_healthy": true,
      "health_score": 95.5,
      "last_success": "2024-01-15T10:30:00Z",
      "last_failure": "2024-01-15T09:15:00Z"
    }
  },
  "mode": "auto",
  "active_provider": "gemini"
}
```

#### GET `/api/providers/metrics`

Returns recent metrics (last 100 requests).

**Response:**
```json
{
  "metrics": [
    {
      "provider": "claude",
      "metric_type": "completion",
      "timestamp": "2024-01-15T10:30:00Z",
      "latency_ms": 850,
      "success": true,
      "tokens_input": 150,
      "tokens_output": 200,
      "model": "claude-3-5-sonnet-20241022",
      "cost_estimate": 0.0045
    }
  ],
  "count": 100
}
```

#### GET `/api/providers/health`

Returns health status for all providers.

**Response:**
```json
{
  "providers": {
    "claude": {
      "is_healthy": true,
      "health_score": 95.5,
      "success_rate": 99.5,
      "avg_latency": 850.0,
      "total_requests": 1000,
      "failed_requests": 5
    }
  },
  "mode": "auto"
}
```

---

## Troubleshooting

### Provider Not Working

**Check API key:**
```bash
# Test OpenAI
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"

# Test Claude
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: $CLAUDE_API_KEY"

# Test Gemini
curl "https://generativelanguage.googleapis.com/v1beta/models?key=$GEMINI_API_KEY"
```

**Check provider health:**
```bash
curl http://localhost:8100/api/providers/health
```

### Auto-Selection Not Working

1. **Verify mode:**
   ```bash
   LLM_PROVIDER_MODE=auto  # Must be set
   ```

2. **Check multiple providers configured:**
   ```bash
   # Need at least 2 providers
   OPENAI_API_KEY=...
   CLAUDE_API_KEY=...
   ```

3. **Verify metrics enabled:**
   ```bash
   LLM_METRICS_ENABLED=true
   ```

4. **Check metrics file:**
   ```bash
   ls -lh ./data/llm_metrics.json
   ```

### High Costs

**View cost breakdown:**
```bash
curl http://localhost:8100/api/providers/stats | jq '.providers[].total_cost'
```

**Optimize costs:**
1. Use Gemini for bulk operations (cheapest)
2. Use Ollama for development (free)
3. Set cost weight higher in auto-selection:
   - Edit `llm/instrumented_provider.go`
   - Increase `CostScore` weight

### Slow Performance

**Check latency stats:**
```bash
curl http://localhost:8100/api/providers/stats | jq '.providers[].avg_latency_ms'
```

**Optimize:**
1. Use Groq for low latency
2. Enable auto-selection (routes to fastest provider)
3. Check network latency to cloud providers

---

## Migration Guide

### From Single Provider to Auto-Selection

**Before:**
```bash
LLM_PROVIDER=openai
OPENAI_API_KEY=...
```

**After:**
```bash
# Enable auto mode
LLM_PROVIDER_MODE=auto
LLM_METRICS_ENABLED=true

# Configure multiple providers
OPENAI_API_KEY=...
CLAUDE_API_KEY=...
GEMINI_API_KEY=...
```

**No code changes required!** The API remains the same.

---

## Best Practices

### Development
- Use **Ollama** (free, local, fast iteration)
- Enable metrics for testing
- Test with multiple providers

### Staging
- Use **auto-selection** with 2-3 providers
- Monitor costs and latency
- Set up health monitoring

### Production
- Use **auto-selection** or **failover** mode
- Configure at least 2 providers for redundancy
- Monitor metrics via `/api/providers/stats`
- Set up alerts for provider health
- Review cost metrics weekly

### Cost Optimization
- Use **Gemini** for bulk operations
- Use **Groq** for real-time interactions
- Use **Claude** for complex reasoning only
- Keep **Ollama** as free fallback

---

## Advanced Configuration

### Custom Selector Weights

Edit `llm/instrumented_provider.go`:

```go
func DefaultSelectorWeights() SelectorWeights {
    return SelectorWeights{
        HealthScore:  0.30,  // Reduce health importance
        LatencyScore: 0.50,  // Increase latency importance
        CostScore:    0.10,  // Keep cost low priority
        SuccessRate:  0.10,  // Reduce success importance
    }
}
```

### Custom Metrics Retention

```bash
# Keep more metrics
LLM_METRICS_PATH=./data/llm_metrics.json

# Edit llm/provider_manager.go
collector := NewMetricsCollector(100000, metricsPath)  // 100K instead of 10K
```

---

## Support & Contributing

### Issues
Report issues at: https://github.com/itaiguardiola/askara/issues

### Feature Requests
Want a new provider? Open an issue with:
- Provider name and API documentation
- Use case and benefits
- Pricing information

### Contributing
See `CONTRIBUTING.md` for guidelines on adding new providers.

---

## License

This feature is part of Askara, licensed under the same terms as the main project.

---

**Happy LLM Provider Management! 🚀**
