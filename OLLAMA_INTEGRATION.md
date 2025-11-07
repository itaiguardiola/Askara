# Ollama Integration Guide

This guide explains how to use Askara with Ollama for local, free AI inference.

## Overview

Askara now supports multiple LLM providers through a unified interface:
- **OpenAI** (cloud, paid) - Original provider
- **Ollama** (local, free) - New local inference option

The LLM provider abstraction layer (`llm/` package) allows seamless switching between providers using environment variables.

## Architecture

### LLM Provider Interface

```go
type LLMProvider interface {
    GenerateEmbedding(text string) ([]float32, error)
    GenerateCompletion(prompt string, context []string) (string, error)
    StreamCompletion(prompt string, context []string, onChunk func(string)) error
}
```

### Implementations

1. **OpenAI Provider** (`llm/openai.go`)
   - Wraps the official OpenAI Go SDK
   - Supports embeddings with `text-embedding-ada-002`
   - Supports completions with `gpt-3.5-turbo` or other models
   - Includes streaming support

2. **Ollama Provider** (`llm/ollama.go`)
   - HTTP client for Ollama REST API
   - Configurable base URL and model
   - Default embedding model: `nomic-embed-text`
   - Default completion model: `llama2`
   - Full streaming support

### Factory Pattern

The `llm.NewProviderFromEnv()` function creates providers based on environment variables:

```go
llmProvider, err := llm.NewProviderFromEnv()
```

## Configuration

### Environment Variables

#### Provider Selection

```bash
# Choose your LLM provider
LLM_PROVIDER=ollama  # or "openai" (default)
```

#### OpenAI Configuration

```bash
OPENAI_API_KEY=sk-...
```

#### Ollama Configuration

```bash
OLLAMA_HOST=http://localhost:11434  # Ollama server URL
OLLAMA_MODEL=llama2                 # Completion model
OLLAMA_EMBEDDING_MODEL=nomic-embed-text  # Embedding model
```

### Configuration Files

**`.env` example for Ollama:**
```bash
LLM_PROVIDER=ollama
OLLAMA_HOST=http://localhost:11434
OLLAMA_MODEL=llama2
OLLAMA_EMBEDDING_MODEL=nomic-embed-text
QDRANT_API_ENDPOINT=http://localhost:6333
```

**`.env` example for OpenAI:**
```bash
LLM_PROVIDER=openai
OPENAI_API_KEY=sk-your-key-here
PINECONE_API_KEY=your-pinecone-key
PINECONE_API_ENDPOINT=https://your-index.svc.region.pinecone.io
```

## Setup Instructions

### Prerequisites

1. **Ollama Installation**
   - Windows: Download from https://ollama.ai/download
   - Linux: `curl https://ollama.ai/install.sh | sh`
   - Mac: `brew install ollama`

2. **Pull Required Models**
   ```bash
   ollama pull llama2
   ollama pull nomic-embed-text
   ```

3. **Start Ollama Server**
   ```bash
   ollama serve
   ```
   - Default port: 11434
   - Verify: `curl http://localhost:11434/api/tags`

### Option 1: Local Development Setup

1. **Configure Environment**
   ```bash
   cp .env.example .env
   # Edit .env to set LLM_PROVIDER=ollama
   ```

2. **Start Qdrant (Vector Database)**
   ```bash
   docker run -p 6333:6333 qdrant/qdrant
   ```

3. **Install Dependencies**
   ```bash
   npm install
   go mod download
   ```

4. **Run Development Server**
   ```bash
   # Terminal 1: Start Go server
   npm start

   # Terminal 2: Start webpack dev server
   npm run dev
   ```

5. **Access Application**
   - Open http://localhost:8100
   - Upload documents and ask questions

### Option 2: Docker Deployment

#### Windows (Ollama on Host)

1. **Ensure Ollama is Running**
   ```powershell
   ollama serve
   ```

2. **Create .env File**
   ```bash
   LLM_PROVIDER=ollama
   OLLAMA_HOST=http://host.docker.internal:11434
   OLLAMA_MODEL=llama2
   QDRANT_API_ENDPOINT=http://qdrant:6333
   ```

3. **Start Services**
   ```bash
   docker-compose -f docker-compose.yml -f docker-compose.windows.yml up -d
   ```

#### Linux/Mac (Ollama in Docker)

1. **Add Ollama to docker-compose.yml**
   ```yaml
   services:
     ollama:
       image: ollama/ollama:latest
       ports:
         - "11434:11434"
       volumes:
         - ollama-data:/root/.ollama
   ```

2. **Create .env File**
   ```bash
   LLM_PROVIDER=ollama
   OLLAMA_HOST=http://ollama:11434
   OLLAMA_MODEL=llama2
   QDRANT_API_ENDPOINT=http://qdrant:6333
   ```

3. **Start Services**
   ```bash
   docker-compose up -d
   ```

4. **Pull Models Inside Container**
   ```bash
   docker-compose exec ollama ollama pull llama2
   docker-compose exec ollama ollama pull nomic-embed-text
   ```

## Testing

### Manual Testing

1. **Test Ollama Connection**
   ```bash
   curl http://localhost:11434/api/tags
   ```
   Expected: JSON list of available models

2. **Test Embedding Generation**
   ```bash
   curl http://localhost:11434/api/embeddings -d '{
     "model": "nomic-embed-text",
     "prompt": "The sky is blue"
   }'
   ```
   Expected: JSON with embedding array

3. **Test Completion Generation**
   ```bash
   curl http://localhost:11434/api/generate -d '{
     "model": "llama2",
     "prompt": "Why is the sky blue?",
     "stream": false
   }'
   ```
   Expected: JSON with response text

### Integration Testing

1. **Upload a Document**
   - Open http://localhost:8100
   - Drag and drop a PDF or text file
   - Wait for "All files uploaded and processed successfully"
   - Check server logs for "LLM get embedding request"

2. **Ask a Question**
   - Type a question about your uploaded document
   - Click "Submit"
   - Verify streaming response appears
   - Check server logs for "LLM api request"

3. **Test Document Management**
   - Scroll to "Uploaded Documents" section
   - Verify your document appears in the list
   - Click delete icon and confirm
   - Verify document is removed

### Automated Testing

Run the LLM provider unit tests:

```bash
cd llm
go test -v
```

Expected output:
```
=== RUN   TestOllamaProvider_GenerateEmbedding
--- PASS: TestOllamaProvider_GenerateEmbedding
=== RUN   TestOllamaProvider_GenerateCompletion
--- PASS: TestOllamaProvider_GenerateCompletion
...
PASS
ok      github.com/itaiguardiola/askara/llm     0.XXXs
```

## Troubleshooting

### Ollama Connection Issues

**Error:** `connection refused` or `dial tcp: connect: connection refused`

**Solutions:**
1. Verify Ollama is running: `ps aux | grep ollama`
2. Check port: `netstat -an | grep 11434`
3. Test connectivity: `curl http://localhost:11434/api/tags`
4. Windows Docker: Ensure host.docker.internal is accessible

### Model Not Found

**Error:** `model 'llama2' not found`

**Solutions:**
1. Pull the model: `ollama pull llama2`
2. Verify: `ollama list`
3. Check model name in .env matches exactly

### Embedding Dimension Mismatch

**Error:** `vector dimension mismatch`

**Solutions:**
1. Check embedding model dimensions:
   - `nomic-embed-text`: 768 dimensions
   - `text-embedding-ada-002` (OpenAI): 1536 dimensions
2. Recreate vector database with correct dimensions
3. For Qdrant: Delete collection and let Askara recreate it

### Slow Performance

**Solutions:**
1. Use GPU acceleration (RTX 3090):
   - Install CUDA toolkit
   - Verify: `nvidia-smi`
   - Ollama automatically uses GPU if available
2. Use smaller models:
   - Try `mistral` instead of `llama2`
   - Try `all-minilm` for embeddings
3. Adjust model parameters in `llm/ollama.go`

### Memory Issues

**Solutions:**
1. Reduce context window in Ollama:
   ```bash
   ollama run llama2 --ctx-size 2048
   ```
2. Use quantized models (Q4, Q5)
3. Monitor memory: `nvidia-smi` or Task Manager

## Model Recommendations

### For RTX 3090 (24GB VRAM)

**Best Performance:**
- Completion: `llama2:13b` or `mistral:7b`
- Embedding: `nomic-embed-text`

**Maximum Quality:**
- Completion: `llama2:70b` (requires quantization)
- Embedding: `nomic-embed-text`

**Fastest:**
- Completion: `llama2:7b`
- Embedding: `all-minilm:33m`

### For CPU Only

**Recommended:**
- Completion: `llama2:7b` or `mistral:7b`
- Embedding: `all-minilm:33m`

## Switching Between Providers

### From OpenAI to Ollama

1. Update `.env`:
   ```bash
   # Change from:
   LLM_PROVIDER=openai
   OPENAI_API_KEY=sk-...

   # To:
   LLM_PROVIDER=ollama
   OLLAMA_HOST=http://localhost:11434
   ```

2. Restart server:
   ```bash
   npm start
   ```

3. **Important:** Clear existing embeddings if switching embedding models
   - Embeddings from different models are incompatible
   - Delete vector database data or create new collection

### From Ollama to OpenAI

1. Update `.env`:
   ```bash
   # Change from:
   LLM_PROVIDER=ollama

   # To:
   LLM_PROVIDER=openai
   OPENAI_API_KEY=sk-your-key-here
   ```

2. Restart server

## Performance Comparison

| Provider | Cost | Speed (RTX 3090) | Privacy | Setup |
|----------|------|------------------|---------|-------|
| OpenAI   | $$$  | Very Fast        | Cloud   | Easy  |
| Ollama   | Free | Fast             | Local   | Medium|

## Additional Resources

- [Ollama Documentation](https://github.com/ollama/ollama/blob/main/docs/api.md)
- [Ollama Models](https://ollama.ai/library)
- [Windows Setup Guide](WINDOWS_SETUP.md)
- [Testing Documentation](TESTING.md)

## Support

For issues specific to:
- **Ollama Integration:** Check this document
- **Windows Setup:** See WINDOWS_SETUP.md
- **General Askara Issues:** See main README.md
- **Bug Reports:** Create issue on GitHub
