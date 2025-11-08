package llm

import (
	"fmt"
	"log"
	"strings"
)

// BedrockProvider implements the LLMProvider interface for AWS Bedrock.
// Bedrock provides access to multiple foundation models (Claude, Titan, Cohere, etc.)
// Note: This implementation uses HTTP API calls. For production use with AWS SDK,
// import "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
type BedrockProvider struct {
	config *BedrockConfig
}

// BedrockClaudeRequest represents a request to Claude models on Bedrock
type BedrockClaudeRequest struct {
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Temperature       float32  `json:"temperature,omitempty"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
}

// BedrockClaudeResponse represents a response from Claude models on Bedrock
type BedrockClaudeResponse struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
}

// BedrockTitanEmbedRequest represents a request to Titan Embeddings on Bedrock
type BedrockTitanEmbedRequest struct {
	InputText string `json:"inputText"`
}

// BedrockTitanEmbedResponse represents a response from Titan Embeddings
type BedrockTitanEmbedResponse struct {
	Embedding     []float32 `json:"embedding"`
	InputTextTokenCount int   `json:"inputTextTokenCount"`
}

// NewBedrockProvider creates a new AWS Bedrock provider.
func NewBedrockProvider(config *BedrockConfig) (*BedrockProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.Region == "" {
		return nil, fmt.Errorf("AWS region is required")
	}

	if config.Model == "" {
		return nil, fmt.Errorf("model ID is required")
	}

	// Note: In production, you would initialize AWS SDK here
	// For now, we'll return an error with instructions
	provider := &BedrockProvider{
		config: config,
	}

	log.Printf("[BedrockProvider] Initialized with model: %s in region: %s", config.Model, config.Region)

	// For full implementation, you would need:
	// import "github.com/aws/aws-sdk-go-v2/config"
	// import "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	//
	// cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(config.Region))
	// client := bedrockruntime.NewFromConfig(cfg)

	return provider, nil
}

// GenerateEmbedding generates an embedding using Amazon Titan Embeddings on Bedrock.
func (b *BedrockProvider) GenerateEmbedding(text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	log.Printf("[BedrockProvider] Generating embedding for text (length: %d)", len(text))

	// For AWS Bedrock, you would use the AWS SDK:
	/*
	reqBody := BedrockTitanEmbedRequest{
		InputText: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(b.config.EmbedModel),
		ContentType: aws.String("application/json"),
		Body:        jsonData,
	}

	result, err := b.client.InvokeModel(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke model: %w", err)
	}

	var embedResp BedrockTitanEmbedResponse
	if err := json.Unmarshal(result.Body, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}

	return embedResp.Embedding, nil
	*/

	return nil, fmt.Errorf("AWS Bedrock provider requires AWS SDK integration. " +
		"Please install: go get github.com/aws/aws-sdk-go-v2/service/bedrockruntime " +
		"and configure AWS credentials. See llm/bedrock.go for implementation details")
}

// buildClaudePrompt creates a Claude-formatted prompt with context.
func (b *BedrockProvider) buildClaudePrompt(prompt string, contextTexts []string) string {
	var sb strings.Builder

	sb.WriteString("\n\nHuman: ")

	if len(contextTexts) > 0 {
		sb.WriteString("Here is some context to help answer the question:\n\n")
		for i, ctx := range contextTexts {
			sb.WriteString(fmt.Sprintf("<context%d>\n%s\n</context%d>\n\n", i+1, ctx, i+1))
		}
		sb.WriteString("Question: ")
	}

	sb.WriteString(prompt)
	sb.WriteString("\n\nAssistant:")

	return sb.String()
}

// GenerateCompletion generates a completion using models on AWS Bedrock.
func (b *BedrockProvider) GenerateCompletion(prompt string, contextTexts []string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	log.Printf("[BedrockProvider] Generating completion with model: %s", b.config.Model)

	// Determine model family from model ID
	if strings.Contains(b.config.Model, "anthropic.claude") {
		return b.generateClaudeCompletion(prompt, contextTexts)
	}

	return "", fmt.Errorf("unsupported Bedrock model family: %s. " +
		"Supported: anthropic.claude-*", b.config.Model)
}

// generateClaudeCompletion generates a completion using Claude models on Bedrock.
func (b *BedrockProvider) generateClaudeCompletion(prompt string, contextTexts []string) (string, error) {
	_ = b.buildClaudePrompt(prompt, contextTexts) // Will be used when AWS SDK is integrated

	// For AWS Bedrock with Claude, you would use:
	/*
	reqBody := BedrockClaudeRequest{
		Prompt:            fullPrompt,
		MaxTokensToSample: b.config.MaxTokens,
		Temperature:       b.config.Temperature,
		StopSequences:     []string{"\n\nHuman:"},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(b.config.Model),
		ContentType: aws.String("application/json"),
		Body:        jsonData,
	}

	result, err := b.client.InvokeModel(context.Background(), input)
	if err != nil {
		return "", fmt.Errorf("failed to invoke model: %w", err)
	}

	var claudeResp BedrockClaudeResponse
	if err := json.Unmarshal(result.Body, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return strings.TrimSpace(claudeResp.Completion), nil
	*/

	return "", fmt.Errorf("AWS Bedrock provider requires AWS SDK integration. " +
		"Please install: go get github.com/aws/aws-sdk-go-v2/service/bedrockruntime " +
		"and configure AWS credentials (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)")
}

// StreamCompletion generates a streaming completion using models on AWS Bedrock.
func (b *BedrockProvider) StreamCompletion(prompt string, contextTexts []string, onChunk func(string)) error {
	if prompt == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	if onChunk == nil {
		return fmt.Errorf("onChunk callback cannot be nil")
	}

	log.Printf("[BedrockProvider] Starting streaming completion with model: %s", b.config.Model)

	// For AWS Bedrock streaming, you would use:
	/*
	fullPrompt := b.buildClaudePrompt(prompt, contextTexts)

	reqBody := BedrockClaudeRequest{
		Prompt:            fullPrompt,
		MaxTokensToSample: b.config.MaxTokens,
		Temperature:       b.config.Temperature,
		StopSequences:     []string{"\n\nHuman:"},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	input := &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     aws.String(b.config.Model),
		ContentType: aws.String("application/json"),
		Body:        jsonData,
	}

	output, err := b.client.InvokeModelWithResponseStream(context.Background(), input)
	if err != nil {
		return fmt.Errorf("failed to invoke model: %w", err)
	}
	defer output.GetStream().Close()

	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			var chunk BedrockClaudeResponse
			if err := json.Unmarshal(v.Value.Bytes, &chunk); err != nil {
				return fmt.Errorf("failed to decode chunk: %w", err)
			}
			if chunk.Completion != "" {
				onChunk(chunk.Completion)
			}
		case *types.UnknownUnionMember:
			log.Printf("unknown event: %s", v.Tag)
		default:
			log.Printf("union is nil or unknown type")
		}
	}

	return nil
	*/

	return fmt.Errorf("AWS Bedrock provider requires AWS SDK integration. " +
		"Please install: go get github.com/aws/aws-sdk-go-v2/service/bedrockruntime")
}

// Helper functions for production implementation:

// getBedrockModelFamily extracts the model family from a Bedrock model ID.
func getBedrockModelFamily(modelID string) string {
	if strings.Contains(modelID, "anthropic.claude") {
		return "claude"
	}
	if strings.Contains(modelID, "amazon.titan") {
		return "titan"
	}
	if strings.Contains(modelID, "cohere") {
		return "cohere"
	}
	if strings.Contains(modelID, "meta.llama") {
		return "llama"
	}
	if strings.Contains(modelID, "ai21") {
		return "ai21"
	}
	return "unknown"
}

// Example Bedrock model IDs:
// - anthropic.claude-3-5-sonnet-20241022-v2:0
// - anthropic.claude-3-opus-20240229-v1:0
// - amazon.titan-embed-text-v2:0
// - amazon.titan-text-express-v1
// - cohere.command-r-plus-v1:0
// - meta.llama2-70b-chat-v1
