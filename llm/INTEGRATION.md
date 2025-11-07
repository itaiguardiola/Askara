# Integration Guide: LLM Provider Package

This guide explains how to integrate the new LLM provider package into the existing Askara codebase.

## Quick Start

### 1. Update main.go

**Before:**
```go
// vault-web-server/main.go
openaiApiKey := os.Getenv("OPENAI_API_KEY")
if len(openaiApiKey) == 0 {
    log.Fatalln("MISSING OPENAI API KEY ENV VARIABLE")
}
openaiClient := openai.NewClient(openaiApiKey)

handlerContext := postapi.NewHandlerContext(openaiClient, vectorDB, docStore)
```

**After:**
```go
// vault-web-server/main.go
import "github.com/itaiguardiola/askara/llm"

// Create LLM provider from environment variables
// Supports both OpenAI and Ollama based on LLM_PROVIDER env var
llmProvider, err := llm.NewProviderFromEnv()
if err != nil {
    log.Fatalln("ERROR INITIALIZING LLM PROVIDER:", err)
}

handlerContext := postapi.NewHandlerContext(llmProvider, vectorDB, docStore)
```

### 2. Update HandlerContext

**Before:**
```go
// vault-web-server/postapi/handlercontext.go
type HandlerContext struct {
    openAIClient *openai.Client
    cache        *cache.Cache
    vectorDB     vectordb.VectorDB
    docStore     storage.DocumentStore
}

func NewHandlerContext(openAIClient *openai.Client, vectorDB vectordb.VectorDB,
    docStore storage.DocumentStore) *HandlerContext {
    return &HandlerContext{
        openAIClient: openAIClient,
        cache:        cache.New(cache.NoExpiration, cache.NoExpiration),
        vectorDB:     vectorDB,
        docStore:     docStore,
    }
}
```

**After:**
```go
// vault-web-server/postapi/handlercontext.go
import "github.com/itaiguardiola/askara/llm"

type HandlerContext struct {
    llmProvider  llm.LLMProvider  // New unified interface
    cache        *cache.Cache
    vectorDB     vectordb.VectorDB
    docStore     storage.DocumentStore
}

func NewHandlerContext(provider llm.LLMProvider, vectorDB vectordb.VectorDB,
    docStore storage.DocumentStore) *HandlerContext {
    return &HandlerContext{
        llmProvider: provider,
        cache:       cache.New(cache.NoExpiration, cache.NoExpiration),
        vectorDB:    vectorDB,
        docStore:    docStore,
    }
}
```

### 3. Update Question Handler Functions

**Before:**
```go
// vault-web-server/postapi/questions.go
questionEmbedding, err := getEmbedding(h.openAIClient, form.Question, openai.AdaEmbeddingV2)
if err != nil {
    log.Println("[QuestionHandler ERR] OpenAI get embedding request error\n", err.Error())
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}
```

**After:**
```go
// vault-web-server/postapi/questions.go
questionEmbedding, err := h.llmProvider.GenerateEmbedding(form.Question)
if err != nil {
    log.Println("[QuestionHandler ERR] LLM get embedding request error\n", err.Error())
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}
```

**Before:**
```go
// Generate response using OpenAI
response, tokens, err := callOpenAI(h.openAIClient, prompt, model, instructions, maxTokens)
```

**After:**
```go
// Generate response using LLM provider
response, err := h.llmProvider.GenerateCompletion(prompt, contextTexts)
// Note: Token counting would need to be added separately if required
```

### 4. Update Streaming Handlers

**Before:**
```go
err = useChatCompletionStreamAPI(h.openAIClient, prompt, model, instructions,
    temperature, maxTokens, topP, frequencyPenalty, presencePenalty, stop,
    func(chunk string) error {
        fmt.Fprintf(w, "data: %s\n\n", chunk)
        w.(http.Flusher).Flush()
        return nil
    })
```

**After:**
```go
err = h.llmProvider.StreamCompletion(prompt, contextTexts, func(chunk string) {
    fmt.Fprintf(w, "data: %s\n\n", chunk)
    w.(http.Flusher).Flush()
})
```

## Environment Variables

### For OpenAI (Default - Backward Compatible)
```bash
# Required
export OPENAI_API_KEY=sk-...

# Optional (uses defaults if not set)
export LLM_PROVIDER=openai
export OPENAI_MODEL=gpt-3.5-turbo
export OPENAI_EMBED_MODEL=text-embedding-ada-002
```

### For Ollama (New)
```bash
# Required
export LLM_PROVIDER=ollama

# Optional (uses defaults if not set)
export OLLAMA_BASE_URL=http://localhost:11434
export OLLAMA_MODEL=llama2
export OLLAMA_EMBED_MODEL=nomic-embed-text
```

## Migration Steps

### Phase 1: Add LLM Provider (Backward Compatible)
1. Keep existing OpenAI code as-is
2. Add LLM provider package alongside
3. Test in development environment

### Phase 2: Gradual Migration
1. Update `HandlerContext` to accept both old and new interfaces
   ```go
   type HandlerContext struct {
       openAIClient *openai.Client    // Keep for now
       llmProvider  llm.LLMProvider    // Add new
       // ... rest of fields
   }
   ```
2. Create adapter functions that use LLM provider when available
3. Test thoroughly with existing functionality

### Phase 3: Complete Migration
1. Update all handler functions to use `llmProvider`
2. Remove `openAIClient` field from `HandlerContext`
3. Remove old OpenAI-specific helper functions
4. Update tests

## Testing the Integration

### Test with OpenAI (Existing Behavior)
```bash
export OPENAI_API_KEY=sk-...
export LLM_PROVIDER=openai
go run vault-web-server/main.go
```

### Test with Ollama (New Capability)
```bash
# Start Ollama
ollama serve &

# Pull required models
ollama pull llama2
ollama pull nomic-embed-text

# Run application
export LLM_PROVIDER=ollama
export OLLAMA_BASE_URL=http://localhost:11434
export OLLAMA_MODEL=llama2
export OLLAMA_EMBED_MODEL=nomic-embed-text
go run vault-web-server/main.go
```

## Code Examples

### Simple Migration Example

**Old Code:**
```go
func (h *HandlerContext) processQuestion(question string) (string, error) {
    // Get embedding
    embedding, err := getEmbedding(h.openAIClient, question, openai.AdaEmbeddingV2)
    if err != nil {
        return "", err
    }

    // Query vector DB
    results, err := h.vectorDB.Query(embedding, 5)
    if err != nil {
        return "", err
    }

    // Build prompt and get response
    prompt := buildPrompt(results, question)
    response, _, err := callOpenAI(h.openAIClient, prompt, "gpt-3.5-turbo",
        "You are a helpful assistant.", 2000)
    return response, err
}
```

**New Code:**
```go
func (h *HandlerContext) processQuestion(question string) (string, error) {
    // Get embedding
    embedding, err := h.llmProvider.GenerateEmbedding(question)
    if err != nil {
        return "", err
    }

    // Query vector DB
    results, err := h.vectorDB.Query(embedding, 5)
    if err != nil {
        return "", err
    }

    // Extract context texts
    contextTexts := make([]string, len(results))
    for i, r := range results {
        contextTexts[i] = r.Text
    }

    // Get response
    response, err := h.llmProvider.GenerateCompletion(question, contextTexts)
    return response, err
}
```

## Benefits

1. **Provider Flexibility**: Easy to switch between OpenAI and Ollama
2. **Cost Optimization**: Use Ollama for development, OpenAI for production
3. **Privacy**: Keep sensitive data local with Ollama
4. **Simplified API**: Cleaner interface for common operations
5. **Future-Proof**: Easy to add new providers (Claude, Cohere, etc.)

## Considerations

### Token Counting
The new interface doesn't return token counts directly. If you need token tracking:

```go
// Option 1: Use tiktoken for estimation
import "github.com/pkoukk/tiktoken-go"

// Option 2: Add token counting to provider implementations
// Option 3: Track at application level using response length estimates
```

### Model Configuration
The provider uses configuration from:
1. Environment variables (highest priority)
2. Config struct passed to factory
3. Default values (lowest priority)

### Error Handling
All provider methods return descriptive errors that wrap underlying issues:
```go
embedding, err := provider.GenerateEmbedding(text)
if err != nil {
    // Error contains full context: "failed to generate embedding: <underlying error>"
    log.Printf("Error: %v", err)
}
```

## Troubleshooting

### Ollama Connection Issues
```bash
# Check if Ollama is running
curl http://localhost:11434/api/version

# Check available models
ollama list

# Pull missing model
ollama pull llama2
```

### OpenAI API Issues
```bash
# Verify API key is set
echo $OPENAI_API_KEY

# Test with curl
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

## Next Steps

1. Review the package documentation in `/home/user/Askara/llm/README.md`
2. Run the examples in `/home/user/Askara/llm/example_test.go`
3. Start migration with a single handler function
4. Expand to all endpoints once tested
5. Add comprehensive tests for both providers
