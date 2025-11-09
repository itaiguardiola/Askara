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

// ClaudeProvider implements the LLMProvider interface for Anthropic Claude.
type ClaudeProvider struct {
	config     *ClaudeConfig
	httpClient *http.Client
	apiVersion string
}

// ClaudeMessage represents a message in the Claude API format
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeRequest represents a request to the Claude API
type ClaudeRequest struct {
	Model       string          `json:"model"`
	Messages    []ClaudeMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float32         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	System      string          `json:"system,omitempty"`
}

// ClaudeResponse represents a response from the Claude API
type ClaudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// ClaudeStreamEvent represents a streaming event from Claude API
type ClaudeStreamEvent struct {
	Type  string `json:"type"`
	Index int    `json:"index,omitempty"`
	Delta *struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta,omitempty"`
	Message      *ClaudeResponse `json:"message,omitempty"`
	ContentBlock *struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content_block,omitempty"`
}

// VoyageEmbeddingRequest represents a request to Voyage AI embeddings API
type VoyageEmbeddingRequest struct {
	Input         []string `json:"input"`
	Model         string   `json:"model"`
	InputType     string   `json:"input_type,omitempty"`
	TruncateInput bool     `json:"truncate,omitempty"`
}

// VoyageEmbeddingResponse represents a response from Voyage AI
type VoyageEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// NewClaudeProvider creates a new Claude provider.
func NewClaudeProvider(config *ClaudeConfig) (*ClaudeProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	httpClient := &http.Client{
		Timeout: 120 * time.Second,
	}

	provider := &ClaudeProvider{
		config:     config,
		httpClient: httpClient,
		apiVersion: "2023-06-01",
	}

	log.Printf("[ClaudeProvider] Initialized with model: %s", config.Model)

	return provider, nil
}

// GenerateEmbedding generates embeddings using Voyage AI (Claude's recommended embedding service)
func (c *ClaudeProvider) GenerateEmbedding(text string) ([]float32, error) {
	// Use Voyage AI for embeddings
	voyageAPIKey := c.config.APIKey // You may want to use a separate Voyage API key

	reqBody := VoyageEmbeddingRequest{
		Input:         []string{text},
		Model:         c.config.EmbedModel,
		InputType:     "document",
		TruncateInput: true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.voyageai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", voyageAPIKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[ClaudeProvider] Embedding request failed: %v", err)
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[ClaudeProvider] Embedding API error (status %d): %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("embedding API error: status %d: %s", resp.StatusCode, string(body))
	}

	var embedResp VoyageEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}

	if len(embedResp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embedResp.Data[0].Embedding, nil
}

// GenerateCompletion generates a completion using Claude.
func (c *ClaudeProvider) GenerateCompletion(prompt string, context []string) (string, error) {
	messages := c.buildMessages(prompt, context)

	reqBody := ClaudeRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		Stream:      false,
		System:      c.config.Instructions,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", c.apiVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[ClaudeProvider] Completion request failed: %v", err)
		return "", fmt.Errorf("completion request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[ClaudeProvider] API error (status %d): %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("API error: status %d: %s", resp.StatusCode, string(body))
	}

	var claudeResp ClaudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return claudeResp.Content[0].Text, nil
}

// StreamCompletion generates a streaming completion using Claude.
func (c *ClaudeProvider) StreamCompletion(prompt string, context []string, onChunk func(string)) error {
	messages := c.buildMessages(prompt, context)

	reqBody := ClaudeRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		Stream:      true,
		System:      c.config.Instructions,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", c.apiVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[ClaudeProvider] Streaming request failed: %v", err)
		return fmt.Errorf("streaming request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[ClaudeProvider] API error (status %d): %s", resp.StatusCode, string(body))
		return fmt.Errorf("API error: status %d: %s", resp.StatusCode, string(body))
	}

	// Parse SSE stream
	reader := resp.Body
	buffer := make([]byte, 4096)

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading stream: %w", err)
		}

		data := string(buffer[:n])
		lines := strings.Split(data, "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "data: ") {
				jsonStr := strings.TrimPrefix(line, "data: ")

				// Skip [DONE] signal
				if jsonStr == "[DONE]" {
					continue
				}

				var event ClaudeStreamEvent
				if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
					log.Printf("[ClaudeProvider] Failed to parse stream event: %v", err)
					continue
				}

				// Handle content delta
				if event.Type == "content_block_delta" && event.Delta != nil {
					if event.Delta.Text != "" {
						onChunk(event.Delta.Text)
					}
				}
			}
		}
	}

	return nil
}

// buildMessages constructs the messages array for Claude API
func (c *ClaudeProvider) buildMessages(prompt string, context []string) []ClaudeMessage {
	messages := []ClaudeMessage{}

	// Add context as user message if provided
	if len(context) > 0 {
		contextStr := "Here is the relevant context:\n\n"
		for i, ctx := range context {
			contextStr += fmt.Sprintf("Context %d:\n%s\n\n", i+1, ctx)
		}
		messages = append(messages, ClaudeMessage{
			Role:    "user",
			Content: contextStr,
		})

		// Add assistant acknowledgment
		messages = append(messages, ClaudeMessage{
			Role:    "assistant",
			Content: "I've received the context. How can I help you with it?",
		})
	}

	// Add the actual user prompt
	messages = append(messages, ClaudeMessage{
		Role:    "user",
		Content: prompt,
	})

	return messages
}
