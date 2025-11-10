package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/itaiguardiola/askara/prompt"
)

// OllamaProvider implements the LLMProvider interface using Ollama.
type OllamaProvider struct {
	config     *OllamaConfig
	httpClient *http.Client
}

// NewOllamaProvider creates a new Ollama provider with the given configuration.
func NewOllamaProvider(config *OllamaConfig) (*OllamaProvider, error) {
	if config == nil {
		config = DefaultOllamaConfig()
	}

	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:11434"
	}

	return &OllamaProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for large model responses
		},
	}, nil
}

// ollamaEmbeddingRequest represents the request structure for Ollama embeddings API.
type ollamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// ollamaEmbeddingResponse represents the response structure from Ollama embeddings API.
type ollamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// GenerateEmbedding generates an embedding for the given text using Ollama.
func (o *OllamaProvider) GenerateEmbedding(text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	url := fmt.Sprintf("%s/api/embeddings", o.config.BaseURL)

	reqBody := ollamaEmbeddingRequest{
		Model:  o.config.EmbedModel,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("[OllamaProvider] Generating embedding for text (length: %d)", len(text))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response ollamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert []float64 to []float32
	embedding := make([]float32, len(response.Embedding))
	for i, v := range response.Embedding {
		embedding[i] = float32(v)
	}

	log.Printf("[OllamaProvider] Generated embedding with dimension: %d", len(embedding))

	return embedding, nil
}

// ollamaGenerateRequest represents the request structure for Ollama generate API.
type ollamaGenerateRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Stream      bool    `json:"stream"`
	Temperature float32 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
}

// ollamaGenerateResponse represents the response structure from Ollama generate API.
type ollamaGenerateResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}

// buildPromptWithContext creates an optimized prompt using the prompt builder package.
func (o *OllamaProvider) buildPromptWithContext(userPrompt string, contextTexts []string) string {
	if len(contextTexts) == 0 {
		return userPrompt
	}

	// Use the prompt builder to create an optimized prompt
	builder := prompt.NewBuilder()
	result := builder.Build(userPrompt, contextTexts)

	log.Printf("[OllamaProvider] Using %s question type with temperature %.2f",
		result.QuestionType.String(), result.Temperature)

	// For Ollama, we combine system prompt and user prompt since it doesn't have separate system messages
	// We'll prepend system instructions as context
	if result.SystemPrompt != "" {
		return fmt.Sprintf("System: %s\n\n%s", result.SystemPrompt, result.FinalPrompt)
	}

	return result.FinalPrompt
}

// GenerateCompletion generates a completion for the given prompt with context using Ollama.
func (o *OllamaProvider) GenerateCompletion(prompt string, contextTexts []string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	url := fmt.Sprintf("%s/api/generate", o.config.BaseURL)
	fullPrompt := o.buildPromptWithContext(prompt, contextTexts)

	reqBody := ollamaGenerateRequest{
		Model:       o.config.Model,
		Prompt:      fullPrompt,
		Stream:      false,
		Temperature: o.config.Temperature,
		MaxTokens:   o.config.MaxTokens,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("[OllamaProvider] Generating completion with model: %s", o.config.Model)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("[OllamaProvider] Completion generated (length: %d)", len(response.Response))

	return response.Response, nil
}

// StreamCompletion generates a completion with streaming support using Ollama.
func (o *OllamaProvider) StreamCompletion(prompt string, contextTexts []string, onChunk func(string)) error {
	if prompt == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	if onChunk == nil {
		return fmt.Errorf("onChunk callback cannot be nil")
	}

	url := fmt.Sprintf("%s/api/generate", o.config.BaseURL)
	fullPrompt := o.buildPromptWithContext(prompt, contextTexts)

	reqBody := ollamaGenerateRequest{
		Model:       o.config.Model,
		Prompt:      fullPrompt,
		Stream:      true,
		Temperature: o.config.Temperature,
		MaxTokens:   o.config.MaxTokens,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("[OllamaProvider] Starting streaming completion with model: %s", o.config.Model)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Process streaming response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Debug: log raw response
		log.Printf("[OllamaProvider] Raw line: %s", string(line))

		var streamResp ollamaGenerateResponse
		if err := json.Unmarshal(line, &streamResp); err != nil {
			log.Printf("[OllamaProvider] Warning: failed to decode streaming response: %v", err)
			continue
		}

		log.Printf("[OllamaProvider] Parsed response: model=%s, done=%v, response_len=%d", streamResp.Model, streamResp.Done, len(streamResp.Response))

		if streamResp.Response != "" {
			// Debug: show what we're streaming
			log.Printf("[OllamaProvider] Chunk: %q (len=%d)", streamResp.Response, len(streamResp.Response))
			onChunk(streamResp.Response)
		}

		if streamResp.Done {
			log.Printf("[OllamaProvider] Streaming completion finished")
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading streaming response: %w", err)
	}

	return nil
}
