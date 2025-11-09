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

// GeminiProvider implements the LLMProvider interface for Google Gemini.
type GeminiProvider struct {
	config     *GeminiConfig
	httpClient *http.Client
}

// GeminiContent represents content in Gemini API format
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

// GeminiPart represents a part of content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiRequest represents a request to Gemini API
type GeminiRequest struct {
	Contents          []GeminiContent         `json:"contents"`
	SystemInstruction *GeminiContent          `json:"systemInstruction,omitempty"`
	GenerationConfig  *GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

// GeminiGenerationConfig holds generation configuration
type GeminiGenerationConfig struct {
	Temperature     float32 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	TopP            float32 `json:"topP,omitempty"`
	TopK            int     `json:"topK,omitempty"`
}

// GeminiResponse represents a response from Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []GeminiPart `json:"parts"`
			Role  string       `json:"role"`
		} `json:"content"`
		FinishReason  string `json:"finishReason"`
		Index         int    `json:"index"`
		SafetyRatings []struct {
			Category    string `json:"category"`
			Probability string `json:"probability"`
		} `json:"safetyRatings"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

// GeminiEmbeddingRequest represents a request to Gemini embeddings API
type GeminiEmbeddingRequest struct {
	Content  GeminiContent `json:"content"`
	TaskType string        `json:"taskType,omitempty"`
}

// GeminiEmbeddingResponse represents a response from Gemini embeddings API
type GeminiEmbeddingResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

// NewGeminiProvider creates a new Gemini provider.
func NewGeminiProvider(config *GeminiConfig) (*GeminiProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	httpClient := &http.Client{
		Timeout: 120 * time.Second,
	}

	provider := &GeminiProvider{
		config:     config,
		httpClient: httpClient,
	}

	log.Printf("[GeminiProvider] Initialized with model: %s", config.Model)

	return provider, nil
}

// GenerateEmbedding generates embeddings using Gemini.
func (g *GeminiProvider) GenerateEmbedding(text string) ([]float32, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s",
		g.config.EmbedModel, g.config.APIKey)

	reqBody := GeminiEmbeddingRequest{
		Content: GeminiContent{
			Parts: []GeminiPart{{Text: text}},
		},
		TaskType: "RETRIEVAL_DOCUMENT",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		log.Printf("[GeminiProvider] Embedding request failed: %v", err)
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[GeminiProvider] Embedding API error (status %d): %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("embedding API error: status %d: %s", resp.StatusCode, string(body))
	}

	var embedResp GeminiEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}

	if len(embedResp.Embedding.Values) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embedResp.Embedding.Values, nil
}

// GenerateCompletion generates a completion using Gemini.
func (g *GeminiProvider) GenerateCompletion(prompt string, context []string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		g.config.Model, g.config.APIKey)

	contents := g.buildContents(prompt, context)

	reqBody := GeminiRequest{
		Contents: contents,
		GenerationConfig: &GeminiGenerationConfig{
			Temperature:     g.config.Temperature,
			MaxOutputTokens: g.config.MaxTokens,
		},
	}

	if g.config.Instructions != "" {
		reqBody.SystemInstruction = &GeminiContent{
			Parts: []GeminiPart{{Text: g.config.Instructions}},
		}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		log.Printf("[GeminiProvider] Completion request failed: %v", err)
		return "", fmt.Errorf("completion request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[GeminiProvider] API error (status %d): %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("API error: status %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// StreamCompletion generates a streaming completion using Gemini.
func (g *GeminiProvider) StreamCompletion(prompt string, context []string, onChunk func(string)) error {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s",
		g.config.Model, g.config.APIKey)

	contents := g.buildContents(prompt, context)

	reqBody := GeminiRequest{
		Contents: contents,
		GenerationConfig: &GeminiGenerationConfig{
			Temperature:     g.config.Temperature,
			MaxOutputTokens: g.config.MaxTokens,
		},
	}

	if g.config.Instructions != "" {
		reqBody.SystemInstruction = &GeminiContent{
			Parts: []GeminiPart{{Text: g.config.Instructions}},
		}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		log.Printf("[GeminiProvider] Streaming request failed: %v", err)
		return fmt.Errorf("streaming request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[GeminiProvider] API error (status %d): %s", resp.StatusCode, string(body))
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

				var geminiResp GeminiResponse
				if err := json.Unmarshal([]byte(jsonStr), &geminiResp); err != nil {
					log.Printf("[GeminiProvider] Failed to parse stream event: %v", err)
					continue
				}

				// Extract text from the response
				if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
					text := geminiResp.Candidates[0].Content.Parts[0].Text
					if text != "" {
						onChunk(text)
					}
				}
			}
		}
	}

	return nil
}

// buildContents constructs the contents array for Gemini API
func (g *GeminiProvider) buildContents(prompt string, context []string) []GeminiContent {
	contents := []GeminiContent{}

	// If context is provided, add it as the first user message
	if len(context) > 0 {
		contextStr := "Here is the relevant context:\n\n"
		for i, ctx := range context {
			contextStr += fmt.Sprintf("Context %d:\n%s\n\n", i+1, ctx)
		}

		contents = append(contents, GeminiContent{
			Parts: []GeminiPart{{Text: contextStr}},
			Role:  "user",
		})

		// Add model acknowledgment
		contents = append(contents, GeminiContent{
			Parts: []GeminiPart{{Text: "I've received the context. How can I help you with it?"}},
			Role:  "model",
		})
	}

	// Add the actual user prompt
	contents = append(contents, GeminiContent{
		Parts: []GeminiPart{{Text: prompt}},
		Role:  "user",
	})

	return contents
}
