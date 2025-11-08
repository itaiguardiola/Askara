package llm

import (
	"context"
	"fmt"
	"log"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// AzureProvider implements the LLMProvider interface using Azure OpenAI.
// Azure OpenAI provides enterprise-grade OpenAI models with additional compliance and security features.
type AzureProvider struct {
	client *openai.Client
	config *AzureConfig
}

// NewAzureProvider creates a new Azure OpenAI provider with the given configuration.
func NewAzureProvider(config *AzureConfig) (*AzureProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	if config.Endpoint == "" {
		return nil, fmt.Errorf("Azure endpoint is required")
	}

	if config.DeploymentID == "" {
		return nil, fmt.Errorf("Azure deployment ID is required")
	}

	// Create Azure OpenAI client configuration
	// DefaultAzureConfig expects (apiKey, baseURL, deploymentID) format
	azureConfig := openai.DefaultAzureConfig(config.APIKey, config.Endpoint, config.DeploymentID)
	azureConfig.APIVersion = config.APIVersion

	client := openai.NewClientWithConfig(azureConfig)

	provider := &AzureProvider{
		client: client,
		config: config,
	}

	log.Printf("[AzureProvider] Initialized with endpoint: %s, deployment: %s", config.Endpoint, config.DeploymentID)

	return provider, nil
}

// GenerateEmbedding generates an embedding for the given text using Azure OpenAI.
func (a *AzureProvider) GenerateEmbedding(text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	log.Printf("[AzureProvider] Generating embedding for text (length: %d)", len(text))

	// Azure OpenAI uses deployment IDs instead of model names
	// For Azure, the deployment ID is set in the client config
	resp, err := a.client.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: []string{text},
			Model: openai.AdaEmbeddingV2, // Azure uses deployment ID from config
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned from Azure OpenAI")
	}

	log.Printf("[AzureProvider] Generated embedding with dimension: %d", len(resp.Data[0].Embedding))

	return resp.Data[0].Embedding, nil
}

// buildPromptWithContext creates a prompt that includes the provided context.
func (a *AzureProvider) buildPromptWithContext(prompt string, contextTexts []string) string {
	if len(contextTexts) == 0 {
		return prompt
	}

	var sb strings.Builder
	sb.WriteString("Use the following context to answer the question:\n\n")
	sb.WriteString("Context:\n")
	for i, ctx := range contextTexts {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, ctx))
	}
	sb.WriteString("\n")
	sb.WriteString("Question: ")
	sb.WriteString(prompt)

	return sb.String()
}

// GenerateCompletion generates a completion for the given prompt with context using Azure OpenAI.
func (a *AzureProvider) GenerateCompletion(prompt string, contextTexts []string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	fullPrompt := a.buildPromptWithContext(prompt, contextTexts)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: a.config.Instructions,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fullPrompt,
		},
	}

	log.Printf("[AzureProvider] Generating completion with deployment: %s", a.config.DeploymentID)

	// Azure uses deployment IDs instead of model names
	resp, err := a.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       a.config.DeploymentID, // Azure uses deployment ID
			Messages:    messages,
			Temperature: a.config.Temperature,
			MaxTokens:   a.config.MaxTokens,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned from Azure OpenAI")
	}

	completion := resp.Choices[0].Message.Content
	log.Printf("[AzureProvider] Completion generated (length: %d, tokens: %d)",
		len(completion), resp.Usage.TotalTokens)

	return completion, nil
}

// StreamCompletion generates a completion with streaming support using Azure OpenAI.
func (a *AzureProvider) StreamCompletion(prompt string, contextTexts []string, onChunk func(string)) error {
	if prompt == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	if onChunk == nil {
		return fmt.Errorf("onChunk callback cannot be nil")
	}

	fullPrompt := a.buildPromptWithContext(prompt, contextTexts)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: a.config.Instructions,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fullPrompt,
		},
	}

	log.Printf("[AzureProvider] Starting streaming completion with deployment: %s", a.config.DeploymentID)

	ctx := context.Background()
	stream, err := a.client.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:       a.config.DeploymentID, // Azure uses deployment ID
			Messages:    messages,
			Temperature: a.config.Temperature,
			MaxTokens:   a.config.MaxTokens,
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
				log.Printf("[AzureProvider] Streaming completion finished")
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
