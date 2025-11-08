package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// CohereProvider implements the LLMProvider interface for Cohere.
// Cohere is optimized for RAG (Retrieval-Augmented Generation) applications.
type CohereProvider struct {
	config     *CohereConfig
	httpClient *http.Client
}

// CohereChatRequest represents a request to the Cohere Chat API
type CohereChatRequest struct {
	Model       string   `json:"model"`
	Message     string   `json:"message"`
	ChatHistory []struct {
		Role    string `json:"role"`
		Message string `json:"message"`
	} `json:"chat_history,omitempty"`
	Temperature      float32  `json:"temperature,omitempty"`
	MaxTokens        int      `json:"max_tokens,omitempty"`
	Stream           bool     `json:"stream,omitempty"`
	Preamble         string   `json:"preamble,omitempty"`
	PromptTruncation string   `json:"prompt_truncation,omitempty"`
	Documents        []string `json:"documents,omitempty"` // For RAG
}

// CohereChatResponse represents a response from the Cohere Chat API
type CohereChatResponse struct {
	ResponseID   string `json:"response_id"`
	Text         string `json:"text"`
	GenerationID string `json:"generation_id"`
	FinishReason string `json:"finish_reason"`
	Meta         struct {
		APIVersion struct {
			Version string `json:"version"`
		} `json:"api_version"`
		BilledUnits struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"billed_units"`
	} `json:"meta"`
}

// CohereStreamEvent represents a streaming event from Cohere Chat API
type CohereStreamEvent struct {
	EventType    string `json:"event_type"`
	Text         string `json:"text,omitempty"`
	IsFinished   bool   `json:"is_finished,omitempty"`
	FinishReason string `json:"finish_reason,omitempty"`
}

// CohereEmbedRequest represents a request to the Cohere Embed API
type CohereEmbedRequest struct {
	Texts         []string `json:"texts"`
	Model         string   `json:"model"`
	InputType     string   `json:"input_type"`
	TruncateInput string   `json:"truncate,omitempty"`
}

// CohereEmbedResponse represents a response from the Cohere Embed API
type CohereEmbedResponse struct {
	ID         string      `json:"id"`
	Embeddings [][]float32 `json:"embeddings"`
	Texts      []string    `json:"texts"`
	Meta       struct {
		APIVersion struct {
			Version string `json:"version"`
		} `json:"api_version"`
		BilledUnits struct {
			InputTokens int `json:"input_tokens"`
		} `json:"billed_units"`
	} `json:"meta"`
}

// NewCohereProvider creates a new Cohere provider.
func NewCohereProvider(config *CohereConfig) (*CohereProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	httpClient := &http.Client{
		Timeout: 120 * time.Second,
	}

	provider := &CohereProvider{
		config:     config,
		httpClient: httpClient,
	}

	log.Printf("[CohereProvider] Initialized with model: %s", config.Model)

	return provider, nil
}

// GenerateEmbedding generates an embedding for the given text using Cohere.
func (c *CohereProvider) GenerateEmbedding(text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	log.Printf("[CohereProvider] Generating embedding for text (length: %d)", len(text))

	reqBody := CohereEmbedRequest{
		Texts:         []string{text},
		Model:         c.config.EmbedModel,
		InputType:     "search_document", // For indexing documents
		TruncateInput: "END",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.cohere.ai/v1/embed", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var embedResp CohereEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}

	if len(embedResp.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned from Cohere")
	}

	embedding := embedResp.Embeddings[0]
	log.Printf("[CohereProvider] Generated embedding with dimension: %d", len(embedding))

	return embedding, nil
}

// buildPromptWithContext creates a prompt with context for Cohere.
func (c *CohereProvider) buildPromptWithContext(prompt string, contextTexts []string) (string, []string) {
	// Cohere supports passing documents separately, which is better for RAG
	if len(contextTexts) > 0 {
		return prompt, contextTexts
	}
	return prompt, nil
}

// GenerateCompletion generates a completion for the given prompt with context using Cohere.
func (c *CohereProvider) GenerateCompletion(prompt string, contextTexts []string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	message, documents := c.buildPromptWithContext(prompt, contextTexts)

	reqBody := CohereChatRequest{
		Model:            c.config.Model,
		Message:          message,
		Temperature:      c.config.Temperature,
		MaxTokens:        c.config.MaxTokens,
		Preamble:         c.config.Instructions,
		PromptTruncation: "AUTO",
		Documents:        documents, // Cohere's native RAG support
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal chat request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.cohere.ai/v1/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create chat request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	log.Printf("[CohereProvider] Generating completion with model: %s", c.config.Model)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("chat request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("chat request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp CohereChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode chat response: %w", err)
	}

	log.Printf("[CohereProvider] Completion generated (length: %d)", len(chatResp.Text))

	return chatResp.Text, nil
}

// StreamCompletion generates a completion with streaming support using Cohere.
func (c *CohereProvider) StreamCompletion(prompt string, contextTexts []string, onChunk func(string)) error {
	if prompt == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	if onChunk == nil {
		return fmt.Errorf("onChunk callback cannot be nil")
	}

	message, documents := c.buildPromptWithContext(prompt, contextTexts)

	reqBody := CohereChatRequest{
		Model:            c.config.Model,
		Message:          message,
		Temperature:      c.config.Temperature,
		MaxTokens:        c.config.MaxTokens,
		Stream:           true,
		Preamble:         c.config.Instructions,
		PromptTruncation: "AUTO",
		Documents:        documents,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal chat request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.cohere.ai/v1/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create chat request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("Accept", "text/event-stream")

	log.Printf("[CohereProvider] Starting streaming completion with model: %s", c.config.Model)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("streaming chat request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("streaming chat request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read Server-Sent Events (SSE)
	buf := make([]byte, 4096)
	var remainder string

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			data := remainder + string(buf[:n])
			lines := strings.Split(data, "\n")

			// Keep the last incomplete line
			if !strings.HasSuffix(data, "\n") {
				remainder = lines[len(lines)-1]
				lines = lines[:len(lines)-1]
			} else {
				remainder = ""
			}

			// Process complete lines
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || !strings.HasPrefix(line, "data: ") {
					continue
				}

				// Remove "data: " prefix
				jsonStr := strings.TrimPrefix(line, "data: ")

				var event CohereStreamEvent
				if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
					log.Printf("[CohereProvider] Failed to parse streaming event: %v", err)
					continue
				}

				// Send text chunks to callback
				if event.EventType == "text-generation" && event.Text != "" {
					onChunk(event.Text)
				}

				// Check if finished
				if event.IsFinished {
					log.Printf("[CohereProvider] Streaming completion finished")
					return nil
				}
			}
		}

		if err != nil {
			if err == io.EOF {
				log.Printf("[CohereProvider] Streaming completion finished (EOF)")
				return nil
			}
			return fmt.Errorf("error reading stream: %w", err)
		}
	}
}
