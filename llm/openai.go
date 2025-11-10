package llm

import (
	"context"
	"fmt"
	"log"

	"github.com/itaiguardiola/askara/prompt"
	openai "github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the LLMProvider interface using OpenAI.
// It wraps the existing OpenAI client for backward compatibility.
type OpenAIProvider struct {
	client *openai.Client
	config *OpenAIConfig
}

// NewOpenAIProvider creates a new OpenAI provider with the given client and configuration.
func NewOpenAIProvider(client *openai.Client, config *OpenAIConfig) (*OpenAIProvider, error) {
	if client == nil {
		return nil, fmt.Errorf("OpenAI client cannot be nil")
	}

	if config == nil {
		config = DefaultOpenAIConfig("")
	}

	return &OpenAIProvider{
		client: client,
		config: config,
	}, nil
}

// GenerateEmbedding generates an embedding for the given text using OpenAI.
func (o *OpenAIProvider) GenerateEmbedding(text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	log.Printf("[OpenAIProvider] Generating embedding for text (length: %d)", len(text))

	// Convert string model name to EmbeddingModel type
	var embedModel openai.EmbeddingModel
	if err := embedModel.UnmarshalText([]byte(o.config.EmbedModel)); err != nil {
		return nil, fmt.Errorf("invalid embedding model: %w", err)
	}

	resp, err := o.client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: []string{text},
			Model: embedModel,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned from OpenAI")
	}

	log.Printf("[OpenAIProvider] Generated embedding with dimension: %d", len(resp.Data[0].Embedding))

	return resp.Data[0].Embedding, nil
}

// buildPromptWithContext creates an optimized prompt using the prompt builder package.
func (o *OpenAIProvider) buildPromptWithContext(userPrompt string, contextTexts []string) (string, string) {
	if len(contextTexts) == 0 {
		// No context, return simple prompt with default system instructions
		return userPrompt, "You are a helpful assistant. Answer questions accurately and concisely."
	}

	// Use the prompt builder to create an optimized prompt
	builder := prompt.NewBuilder()
	result := builder.Build(userPrompt, contextTexts)

	log.Printf("[OpenAIProvider] Using %s question type with temperature %.2f",
		result.QuestionType.String(), result.Temperature)

	return result.FinalPrompt, result.SystemPrompt
}

// GenerateCompletion generates a completion for the given prompt with context using OpenAI.
func (o *OpenAIProvider) GenerateCompletion(prompt string, contextTexts []string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	fullPrompt, systemPrompt := o.buildPromptWithContext(prompt, contextTexts)

	// Use custom system prompt if provided in config, otherwise use the generated one
	systemInstructions := systemPrompt
	if o.config.Instructions != "" {
		systemInstructions = o.config.Instructions
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemInstructions,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fullPrompt,
		},
	}

	log.Printf("[OpenAIProvider] Generating completion with model: %s", o.config.Model)

	resp, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       o.config.Model,
			Messages:    messages,
			Temperature: o.config.Temperature,
			MaxTokens:   o.config.MaxTokens,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned from OpenAI")
	}

	completion := resp.Choices[0].Message.Content
	log.Printf("[OpenAIProvider] Completion generated (length: %d, tokens: %d)",
		len(completion), resp.Usage.TotalTokens)

	return completion, nil
}

// StreamCompletion generates a completion with streaming support using OpenAI.
func (o *OpenAIProvider) StreamCompletion(prompt string, contextTexts []string, onChunk func(string)) error {
	if prompt == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	if onChunk == nil {
		return fmt.Errorf("onChunk callback cannot be nil")
	}

	fullPrompt, systemPrompt := o.buildPromptWithContext(prompt, contextTexts)

	// Use custom system prompt if provided in config, otherwise use the generated one
	systemInstructions := systemPrompt
	if o.config.Instructions != "" {
		systemInstructions = o.config.Instructions
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemInstructions,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fullPrompt,
		},
	}

	log.Printf("[OpenAIProvider] Starting streaming completion with model: %s", o.config.Model)

	ctx := context.Background()
	stream, err := o.client.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:       o.config.Model,
			Messages:    messages,
			Temperature: o.config.Temperature,
			MaxTokens:   o.config.MaxTokens,
			Stream:      true,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to create streaming completion: %w", err)
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err != nil {
			// Check if stream is done
			if err.Error() == "EOF" {
				log.Printf("[OpenAIProvider] Streaming completion finished")
				return nil
			}
			return fmt.Errorf("error receiving stream: %w", err)
		}

		if len(response.Choices) > 0 {
			chunk := response.Choices[0].Delta.Content
			if chunk != "" {
				onChunk(chunk)
			}
		}
	}
}
