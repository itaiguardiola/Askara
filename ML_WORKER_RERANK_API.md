# ML Worker - Reranker API Specification

## Overview
The Askara orchestrator calls this endpoint for semantic reranking of retrieved documents. This provides cross-encoder based reranking to improve retrieval quality after initial hybrid search (vector + FTS).

## Endpoint

```
POST /rerank
Content-Type: application/json
```

## Request

```json
{
  "query": "What is the capital of France?",
  "documents": [
    {
      "id": "0",
      "text": "Paris is the capital and most populous city of France..."
    },
    {
      "id": "1",
      "text": "London is the capital of the United Kingdom..."
    },
    {
      "id": "2",
      "text": "The city of Paris is known for the Eiffel Tower..."
    }
  ],
  "top_k": 8,
  "model": "BAAI/bge-reranker-v2-m3"
}
```

### Request Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `query` | string | Yes | The user's search query |
| `documents` | array | Yes | List of documents to rerank (ID + text) |
| `top_k` | integer | Yes | Number of top results to return after reranking |
| `model` | string | No | Reranker model to use (defaults to bge-reranker-v2-m3) |

### Document Object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier for the document (matches Qdrant point ID) |
| `text` | string | Yes | Full text content to rerank |

## Response

```json
{
  "results": [
    {
      "id": "0",
      "relevance_score": 0.95,
      "rank": 1
    },
    {
      "id": "2",
      "relevance_score": 0.87,
      "rank": 2
    },
    {
      "id": "1",
      "relevance_score": 0.23,
      "rank": 3
    }
  ],
  "model_used": "BAAI/bge-reranker-v2-m3",
  "processing_time_ms": 45
}
```

### Response Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `results` | array | Yes | Reranked documents with scores (sorted by relevance) |
| `model_used` | string | Yes | The reranker model that was used |
| `processing_time_ms` | number | No | Time taken to rerank (milliseconds) |

### Result Object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Document ID from request |
| `relevance_score` | float | Yes | Relevance score (0.0 to 1.0) |
| `rank` | integer | Yes | Rank position (1-indexed) |

## Error Response

```json
{
  "error": "Error message",
  "details": "Detailed error information"
}
```

HTTP Status Codes:
- `200 OK`: Successful reranking
- `400 Bad Request`: Invalid request (missing fields, invalid data)
- `500 Internal Server Error`: Reranking failed

## Implementation Notes

### Recommended: BGE-Reranker-v2-m3

```python
from FlagEmbedding import FlagReranker

# Initialize reranker (load once, reuse)
reranker = FlagReranker('BAAI/bge-reranker-v2-m3', use_fp16=True)

# Prepare pairs for reranking
pairs = [[query, doc['text']] for doc in documents]

# Get scores
scores = reranker.compute_score(pairs, normalize=True)

# Sort by score descending
results = [
    {
        "id": documents[i]['id'],
        "relevance_score": float(scores[i]),
        "rank": rank + 1
    }
    for rank, i in enumerate(sorted(range(len(scores)), key=lambda i: scores[i], reverse=True)[:top_k])
]
```

### Model Options

Primary recommendation:
- **BAAI/bge-reranker-v2-m3** - Best quality/speed tradeoff, multilingual

Alternatives:
- **BAAI/bge-reranker-base** - Faster, slightly lower quality
- **BAAI/bge-reranker-large** - Highest quality, slower

### Performance Considerations

- Load model once at startup, keep in memory
- Use FP16 for faster inference on GPU
- Batch documents when possible (up to 100 pairs)
- Set timeout to 30 seconds (configurable via `ML_WORKER_TIMEOUT`)
- Return top_k results only (don't send all scores)

## Integration with Askara

### Client Code (To Be Implemented)

```go
// In Go orchestrator
import "github.com/itaiguardiola/askara/mlworker"

mlClient := mlworker.GetClient()

// Prepare documents from Qdrant matches
docs := make([]map[string]string, len(matches))
for i, match := range matches {
    docs[i] = map[string]string{
        "id":   match.ID,
        "text": match.Metadata["text"],
    }
}

// Rerank if feature enabled
if mlClient.IsFeatureEnabled("rerank") {
    reranked, err := mlClient.Rerank(query, docs, 8)
    if err != nil {
        log.Printf("Reranking failed, using original order: %v", err)
    } else {
        // Use reranked results
        matches = reorderMatchesByRerank(matches, reranked)
    }
}
```

### Retrieval Flow with Reranking

1. **Hybrid Search** - Retrieve 50 candidates (vector + FTS)
2. **Optional Reranking** - Use ML Worker to rerank top 50 → 8
3. **Return to User** - Final 8 results with best relevance

### Feature Flag

Enable in `.env`:
```bash
ML_WORKER_ENABLED=true
ML_WORKER_FEATURES=ocr,enhance,caption,parse,rerank
```

## Testing

### Test Request

```bash
curl -X POST http://192.168.0.250:6161/askara/rerank \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What is machine learning?",
    "documents": [
      {"id": "0", "text": "Machine learning is a subset of artificial intelligence..."},
      {"id": "1", "text": "Paris is the capital of France..."},
      {"id": "2", "text": "Deep learning uses neural networks for pattern recognition..."}
    ],
    "top_k": 2
  }'
```

### Expected Response

```json
{
  "results": [
    {"id": "0", "relevance_score": 0.98, "rank": 1},
    {"id": "2", "relevance_score": 0.85, "rank": 2}
  ],
  "model_used": "BAAI/bge-reranker-v2-m3",
  "processing_time_ms": 42
}
```

## Comparison: Local Ollama vs ML Worker Reranking

### Local Ollama Only (Baseline)
- Hybrid search (vector + FTS) with Qdrant
- No cross-encoder reranking
- Fast (single Qdrant query)
- Good for general queries

### ML Worker Reranking (Enhanced)
- Same hybrid search retrieves 50 candidates
- BGE reranker cross-encodes each candidate with query
- More accurate relevance scores
- Slower (network + GPU inference)
- Better for complex/ambiguous queries

### When to Use ML Worker Reranking

Enable for:
- Complex queries requiring deep semantic understanding
- Ambiguous queries with many potential meanings
- High-stakes queries where precision matters

Keep disabled for:
- Simple keyword lookups
- Real-time/low-latency requirements
- Cost-sensitive deployments

## Dependencies

```bash
pip install FlagEmbedding
pip install torch  # For GPU acceleration
```

## Questions?

Contact the Askara orchestrator team for integration support.
