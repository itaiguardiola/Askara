# Quick Start Guide

## 5-Minute Integration

### 1. Try with Ollama (Local)

```bash
# Install and start Ollama
curl https://ollama.ai/install.sh | sh
ollama serve &

# Pull required models
ollama pull llama2
ollama pull nomic-embed-text

# Set environment and run
export LLM_PROVIDER=ollama
go run vault-web-server/main.go
```

### 2. Try with OpenAI (Cloud)

```bash
# Set environment and run
export LLM_PROVIDER=openai
export OPENAI_API_KEY=sk-your-key-here
go run vault-web-server/main.go
```

## Basic Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/itaiguardiola/askara/llm"
)

func main() {
    // Create provider from environment
    provider, err := llm.NewProviderFromEnv()
    if err != nil {
        log.Fatal(err)
    }

    // Generate embedding
    embedding, err := provider.GenerateEmbedding("Hello, world!")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Embedding dimensions: %d\n", len(embedding))

    // Generate completion
    response, err := provider.GenerateCompletion("What is AI?", nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(response)

    // Stream completion
    err = provider.StreamCompletion("Tell me a story", nil, func(chunk string) {
        fmt.Print(chunk)
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

## Environment Variables

### OpenAI
```bash
export LLM_PROVIDER=openai              # default
export OPENAI_API_KEY=sk-...            # required
export OPENAI_MODEL=gpt-3.5-turbo       # optional
export OPENAI_EMBED_MODEL=text-embedding-ada-002  # optional
```

### Ollama
```bash
export LLM_PROVIDER=ollama              # required
export OLLAMA_BASE_URL=http://localhost:11434  # optional
export OLLAMA_MODEL=llama2              # optional
export OLLAMA_EMBED_MODEL=nomic-embed-text  # optional
```

## Testing

```bash
# Run tests
go test ./llm/...

# Run with verbose output
go test ./llm/... -v

# Run specific test
go test ./llm/... -run TestNewProvider
```

## Common Issues

### "MISSING OPENAI API KEY"
```bash
export OPENAI_API_KEY=sk-your-actual-key
```

### "connection refused" (Ollama)
```bash
# Start Ollama service
ollama serve
```

### "model not found" (Ollama)
```bash
# Pull the model
ollama pull llama2
ollama pull nomic-embed-text
```

## Next Steps

1. Read `/home/user/Askara/llm/README.md` for detailed documentation
2. Review `/home/user/Askara/llm/INTEGRATION.md` for migration guide
3. Check `/home/user/Askara/llm/example_test.go` for more examples
