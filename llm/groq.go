package llm

import (
	"context"
	"fmt"
	"io"
	"log"

	openai "github.com/sashabaranov/go-openai"
)

// GroqProvider implements the LLMProvider interface for Groq.
// Groq provides ultra-fast inference with an OpenAI-compatible API.
type GroqProvider struct {
	client *openai.Client
	config *GroqConfig
}

// NewGroqProvider creates a new Groq provider.
func NewGroqProvider(config *GroqConfig) (*GroqProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Create OpenAI client configured for Groq
	clientConfig := openai.DefaultConfig(config.APIKey)
	clientConfig.BaseURL = "https://api.groq.com/openai/v1"

	client := openai.NewClientWithConfig(clientConfig)

	provider := &GroqProvider{
		client: client,
		config: config,
	}

	log.Printf("[GroqProvider] Initialized with model: %s", config.Model)

	return provider, nil
}

// GenerateEmbedding generates embeddings.
// Note: Groq doesn't provide embeddings directly, so we'll return an error
// In production, you might want to fallback to another provider (like Ollama) for embeddings
func (g *GroqProvider) GenerateEmbedding(text string) ([]float32, error) {
	// Groq doesn't provide embeddings API
	// You could implement a fallback to Ollama or another provider here
	return nil, fmt.Errorf("Groq does not support embeddings. Please use a different provider for embeddings (e.g., Ollama, OpenAI)")
}

// GenerateCompletion generates a completion using Groq.
func (g *GroqProvider) GenerateCompletion(prompt string, context []string) (string, error) {
	messages := g.buildMessages(prompt, context)

	req := openai.ChatCompletionRequest{
		Model:       g.config.Model,
		Messages:    messages,
		MaxTokens:   g.config.MaxTokens,
		Temperature: g.config.Temperature,
	}

	resp, err := g.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		log.Printf("[GroqProvider] Completion request failed: %v", err)
		return "", fmt.Errorf("completion request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}

// StreamCompletion generates a streaming completion using Groq.
func (g *GroqProvider) StreamCompletion(prompt string, context []string, onChunk func(string)) error {
	messages := g.buildMessages(prompt, context)

	req := openai.ChatCompletionRequest{
		Model:       g.config.Model,
		Messages:    messages,
		MaxTokens:   g.config.MaxTokens,
		Temperature: g.config.Temperature,
		Stream:      true,
	}

	stream, err := g.client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		log.Printf("[GroqProvider] Streaming request failed: %v", err)
		return fmt.Errorf("streaming request failed: %w", err)
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("[GroqProvider] Stream error: %v", err)
			return fmt.Errorf("stream error: %w", err)
		}

		if len(response.Choices) > 0 {
			content := response.Choices[0].Delta.Content
			if content != "" {
				onChunk(content)
			}
		}
	}

	return nil
}

// buildMessages constructs the messages array for Groq API (OpenAI format)
func (g *GroqProvider) buildMessages(prompt string, context []string) []openai.ChatCompletionMessage {
	messages := []openai.ChatCompletionMessage{}

	// Add system instruction if configured
	if g.config.Instructions != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: g.config.Instructions,
		})
	}

	// Add context if provided
	if len(context) > 0 {
		contextStr := "Here is the relevant context:\n\n"
		for i, ctx := range context {
			contextStr += fmt.Sprintf("Context %d:\n%s\n\n", i+1, ctx)
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: contextStr,
		})

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: "I've received the context. How can I help you with it?",
		})
	}

	// Add the actual user prompt
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})

	return messages
}
