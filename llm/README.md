# LLM Provider Package

A unified interface for multiple Large Language Model providers in the Askara project. This package provides a common abstraction for working with different LLM backends including Ollama and OpenAI.

## Features

- **Unified Interface**: Single `LLMProvider` interface for all providers
- **Multiple Backends**: Support for Ollama (local) and OpenAI (cloud)
- **Streaming Support**: Real-time response streaming for all providers
- **Embeddings**: Generate text embeddings for vector search
- **Production Ready**: Proper error handling, logging, and timeouts
- **Backward Compatible**: Wraps existing OpenAI client usage

## Architecture

### Core Components

1. **`provider.go`**: Defines the `LLMProvider` interface
2. **`ollama.go`**: Ollama implementation (HTTP client-based)
3. **`openai.go`**: OpenAI implementation (wraps existing client)
4. **`factory.go`**: Provider factory for easy instantiation
5. **`config.go`**: Configuration structures for all providers

## Usage

### Quick Start with Environment Variables

```go
import "github.com/itaiguardiola/askara/llm"

// Create provider from environment variables
provider, err := llm.NewProviderFromEnv()
if err != nil {
    log.Fatal(err)
}

// Generate embedding
embedding, err := provider.GenerateEmbedding("Hello, world!")
if err != nil {
    log.Fatal(err)
}

// Generate completion
response, err := provider.GenerateCompletion("What is AI?", nil)
if err != nil {
    log.Fatal(err)
}

// Stream completion
err = provider.StreamCompletion("Tell me a story", nil, func(chunk string) {
    fmt.Print(chunk)
})
```

### Environment Variables

#### For Ollama:
```bash
export LLM_PROVIDER=ollama
export OLLAMA_BASE_URL=http://localhost:11434
export OLLAMA_MODEL=llama2
export OLLAMA_EMBED_MODEL=nomic-embed-text
```

#### For OpenAI:
```bash
export LLM_PROVIDER=openai
export OPENAI_API_KEY=sk-...
export OPENAI_MODEL=gpt-3.5-turbo
export OPENAI_EMBED_MODEL=text-embedding-ada-002
```

### Advanced Configuration

```go
import (
    "github.com/itaiguardiola/askara/llm"
    openai "github.com/sashabaranov/go-openai"
)

// Ollama configuration
ollamaConfig := &llm.Config{
    Provider: "ollama",
    OllamaConfig: &llm.OllamaConfig{
        BaseURL:     "http://localhost:11434",
        Model:       "mistral",
        EmbedModel:  "nomic-embed-text",
        Temperature: 0.7,
        MaxTokens:   2000,
    },
}

ollamaProvider, err := llm.NewProvider(ollamaConfig)
if err != nil {
    log.Fatal(err)
}

// OpenAI configuration
openaiClient := openai.NewClient(apiKey)
openaiConfig := &llm.OpenAIConfig{
    APIKey:       apiKey,
    Model:        "gpt-4",
    EmbedModel:   "text-embedding-ada-002",
    Temperature:  0.7,
    MaxTokens:    2000,
    Instructions: "You are a helpful assistant.",
}

openaiProvider, err := llm.NewOpenAIProvider(openaiClient, openaiConfig)
if err != nil {
    log.Fatal(err)
}
```

### Using Context in Completions

```go
// Provide context from vector database results
contexts := []string{
    "The Earth orbits around the Sun.",
    "It takes approximately 365.25 days for one complete orbit.",
}

response, err := provider.GenerateCompletion(
    "How long does it take Earth to orbit the Sun?",
    contexts,
)
```

## Integration with Existing Code

### Migrating from Direct OpenAI Usage

**Before:**
```go
client := openai.NewClient(apiKey)
resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{...})
```

**After:**
```go
config := &llm.Config{
    Provider: "openai",
    OpenAIConfig: &llm.OpenAIConfig{
        APIKey: apiKey,
        Model:  "gpt-3.5-turbo",
    },
}

provider, err := llm.NewProvider(config)
response, err := provider.GenerateCompletion(prompt, context)
```

### Updating HandlerContext

```go
// In handlercontext.go
type HandlerContext struct {
    llmProvider  llm.LLMProvider  // New unified provider
    cache        *cache.Cache
    vectorDB     vectordb.VectorDB
    docStore     storage.DocumentStore
}

func NewHandlerContext(provider llm.LLMProvider, vectorDB vectordb.VectorDB, docStore storage.DocumentStore) *HandlerContext {
    return &HandlerContext{
        llmProvider: provider,
        cache:       cache.New(cache.NoExpiration, cache.NoExpiration),
        vectorDB:    vectorDB,
        docStore:    docStore,
    }
}
```

## Ollama Setup

### Installation

```bash
# Install Ollama
curl https://ollama.ai/install.sh | sh

# Pull models
ollama pull llama2
ollama pull nomic-embed-text
```

### Running Ollama

```bash
# Start Ollama service
ollama serve
```

The Ollama API will be available at `http://localhost:11434`.

## API Reference

### LLMProvider Interface

```go
type LLMProvider interface {
    // Generate embedding vector for text
    GenerateEmbedding(text string) ([]float32, error)

    // Generate completion with optional context
    GenerateCompletion(prompt string, context []string) (string, error)

    // Stream completion with callback for each chunk
    StreamCompletion(prompt string, context []string, onChunk func(string)) error
}
```

## Design Decisions

1. **Interface-Based Design**: Allows easy swapping between providers without changing application code
2. **Factory Pattern**: Simplifies provider instantiation with sensible defaults
3. **Environment Variable Support**: Enables configuration without code changes
4. **Backward Compatibility**: OpenAI wrapper maintains existing client usage patterns
5. **Streaming First**: All providers support streaming for better UX
6. **Context Integration**: Built-in support for RAG (Retrieval-Augmented Generation) patterns
7. **Production Ready**: Comprehensive error handling, logging, and timeouts

## Error Handling

All methods return errors that wrap underlying issues:

```go
embedding, err := provider.GenerateEmbedding(text)
if err != nil {
    // Error contains context about what failed
    log.Printf("Failed to generate embedding: %v", err)
    return err
}
```

## Logging

The package uses structured logging with prefixes:

- `[OllamaProvider]`: Ollama-specific operations
- `[OpenAIProvider]`: OpenAI-specific operations
- `[Factory]`: Provider creation and configuration

## Testing

```bash
# Run tests
go test ./llm/...

# Run with coverage
go test -cover ./llm/...
```

## Performance Considerations

### Ollama
- Local inference (no API costs)
- Latency depends on hardware
- Supports GPU acceleration
- Good for development and privacy-sensitive applications

### OpenAI
- Cloud-based (API costs apply)
- Lower latency for smaller models
- No local hardware requirements
- Better for production at scale

## Future Enhancements

- [ ] Support for additional providers (Anthropic Claude, Cohere, etc.)
- [ ] Token counting and budget management
- [ ] Request caching layer
- [ ] Retry logic with exponential backoff
- [ ] Request/response middleware hooks
- [ ] Performance metrics and monitoring
- [ ] Batch operations for embeddings

## License

Same as the Askara project.
