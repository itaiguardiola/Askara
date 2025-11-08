package llm

// Config holds the configuration for LLM providers.
type Config struct {
	// Provider specifies which LLM provider to use ("ollama", "openai", "claude", "gemini", "groq")
	Provider string

	// OllamaConfig holds Ollama-specific configuration
	OllamaConfig *OllamaConfig

	// OpenAIConfig holds OpenAI-specific configuration
	OpenAIConfig *OpenAIConfig

	// ClaudeConfig holds Anthropic Claude-specific configuration
	ClaudeConfig *ClaudeConfig

	// GeminiConfig holds Google Gemini-specific configuration
	GeminiConfig *GeminiConfig

	// GroqConfig holds Groq-specific configuration
	GroqConfig *GroqConfig

	// AzureConfig holds Azure OpenAI-specific configuration
	AzureConfig *AzureConfig

	// BedrockConfig holds AWS Bedrock-specific configuration
	BedrockConfig *BedrockConfig

	// CohereConfig holds Cohere-specific configuration
	CohereConfig *CohereConfig
}

// OllamaConfig holds configuration for the Ollama provider.
type OllamaConfig struct {
	// BaseURL is the Ollama API endpoint (default: http://localhost:11434)
	BaseURL string

	// Model is the model name to use for completions (e.g., "llama2", "mistral")
	Model string

	// EmbedModel is the model name to use for embeddings (e.g., "nomic-embed-text")
	EmbedModel string

	// Temperature controls randomness in generation (0.0 to 1.0)
	Temperature float32

	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int
}

// OpenAIConfig holds configuration for the OpenAI provider.
type OpenAIConfig struct {
	// APIKey is the OpenAI API key
	APIKey string

	// Model is the model name to use for completions (e.g., "gpt-4", "gpt-3.5-turbo")
	Model string

	// EmbedModel is the model name to use for embeddings (e.g., "text-embedding-ada-002")
	EmbedModel string

	// Temperature controls randomness in generation (0.0 to 1.0)
	Temperature float32

	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int

	// Instructions is the system message/instructions for the model
	Instructions string
}

// DefaultOllamaConfig returns a default Ollama configuration.
func DefaultOllamaConfig() *OllamaConfig {
	return &OllamaConfig{
		BaseURL:     "http://localhost:11434",
		Model:       "llama2",
		EmbedModel:  "nomic-embed-text",
		Temperature: 0.7,
		MaxTokens:   2000,
	}
}

// DefaultOpenAIConfig returns a default OpenAI configuration with the given API key.
func DefaultOpenAIConfig(apiKey string) *OpenAIConfig {
	return &OpenAIConfig{
		APIKey:       apiKey,
		Model:        "gpt-3.5-turbo",
		EmbedModel:   "text-embedding-ada-002",
		Temperature:  0.7,
		MaxTokens:    2000,
		Instructions: "You are a helpful assistant.",
	}
}

// ClaudeConfig holds configuration for the Anthropic Claude provider.
type ClaudeConfig struct {
	// APIKey is the Anthropic API key
	APIKey string

	// Model is the model name to use for completions (e.g., "claude-3-5-sonnet-20241022", "claude-3-opus-20240229")
	Model string

	// EmbedModel is the model/service to use for embeddings (using Voyage AI)
	EmbedModel string

	// Temperature controls randomness in generation (0.0 to 1.0)
	Temperature float32

	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int

	// Instructions is the system message/instructions for the model
	Instructions string
}

// GeminiConfig holds configuration for the Google Gemini provider.
type GeminiConfig struct {
	// APIKey is the Google API key
	APIKey string

	// Model is the model name to use for completions (e.g., "gemini-1.5-pro", "gemini-1.5-flash")
	Model string

	// EmbedModel is the model to use for embeddings (e.g., "text-embedding-004")
	EmbedModel string

	// Temperature controls randomness in generation (0.0 to 2.0)
	Temperature float32

	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int

	// Instructions is the system message/instructions for the model
	Instructions string
}

// GroqConfig holds configuration for the Groq provider.
type GroqConfig struct {
	// APIKey is the Groq API key
	APIKey string

	// Model is the model name to use for completions (e.g., "mixtral-8x7b-32768", "llama2-70b-4096")
	Model string

	// EmbedModel is not used by Groq (uses external embedding service or Ollama)
	EmbedModel string

	// Temperature controls randomness in generation (0.0 to 2.0)
	Temperature float32

	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int

	// Instructions is the system message/instructions for the model
	Instructions string
}

// DefaultClaudeConfig returns a default Claude configuration with the given API key.
func DefaultClaudeConfig(apiKey string) *ClaudeConfig {
	return &ClaudeConfig{
		APIKey:       apiKey,
		Model:        "claude-3-5-sonnet-20241022",
		EmbedModel:   "voyage-large-2-instruct", // Voyage AI embedding
		Temperature:  0.7,
		MaxTokens:    4096,
		Instructions: "You are a helpful assistant.",
	}
}

// DefaultGeminiConfig returns a default Gemini configuration with the given API key.
func DefaultGeminiConfig(apiKey string) *GeminiConfig {
	return &GeminiConfig{
		APIKey:       apiKey,
		Model:        "gemini-1.5-flash",
		EmbedModel:   "text-embedding-004",
		Temperature:  0.7,
		MaxTokens:    8192,
		Instructions: "You are a helpful assistant.",
	}
}

// DefaultGroqConfig returns a default Groq configuration with the given API key.
func DefaultGroqConfig(apiKey string) *GroqConfig {
	return &GroqConfig{
		APIKey:       apiKey,
		Model:        "mixtral-8x7b-32768",
		EmbedModel:   "", // Groq doesn't provide embeddings
		Temperature:  0.7,
		MaxTokens:    4096,
		Instructions: "You are a helpful assistant.",
	}
}

// AzureConfig holds configuration for the Azure OpenAI provider.
type AzureConfig struct {
	// APIKey is the Azure OpenAI API key
	APIKey string

	// Endpoint is the Azure OpenAI endpoint (e.g., "https://your-resource.openai.azure.com/")
	Endpoint string

	// DeploymentID is the deployment name for completions
	DeploymentID string

	// EmbeddingDeploymentID is the deployment name for embeddings
	EmbeddingDeploymentID string

	// APIVersion is the Azure OpenAI API version (e.g., "2023-05-15")
	APIVersion string

	// Temperature controls randomness in generation (0.0 to 1.0)
	Temperature float32

	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int

	// Instructions is the system message/instructions for the model
	Instructions string
}

// BedrockConfig holds configuration for the AWS Bedrock provider.
type BedrockConfig struct {
	// Region is the AWS region (e.g., "us-east-1")
	Region string

	// Model is the model ID to use for completions (e.g., "anthropic.claude-3-5-sonnet-20241022-v2:0")
	Model string

	// EmbedModel is the model ID to use for embeddings (e.g., "amazon.titan-embed-text-v2:0")
	EmbedModel string

	// Temperature controls randomness in generation (0.0 to 1.0)
	Temperature float32

	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int

	// Instructions is the system message/instructions for the model
	Instructions string
}

// CohereConfig holds configuration for the Cohere provider.
type CohereConfig struct {
	// APIKey is the Cohere API key
	APIKey string

	// Model is the model name to use for completions (e.g., "command-r-plus", "command-r")
	Model string

	// EmbedModel is the model to use for embeddings (e.g., "embed-english-v3.0")
	EmbedModel string

	// Temperature controls randomness in generation (0.0 to 1.0)
	Temperature float32

	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int

	// Instructions is the system message/instructions for the model
	Instructions string
}

// DefaultAzureConfig returns a default Azure OpenAI configuration.
func DefaultAzureConfig(apiKey, endpoint, deploymentID string) *AzureConfig {
	return &AzureConfig{
		APIKey:                apiKey,
		Endpoint:              endpoint,
		DeploymentID:          deploymentID,
		EmbeddingDeploymentID: "text-embedding-ada-002", // Default embedding deployment
		APIVersion:            "2023-05-15",
		Temperature:           0.7,
		MaxTokens:             2000,
		Instructions:          "You are a helpful assistant.",
	}
}

// DefaultBedrockConfig returns a default AWS Bedrock configuration.
func DefaultBedrockConfig(region string) *BedrockConfig {
	return &BedrockConfig{
		Region:       region,
		Model:        "anthropic.claude-3-5-sonnet-20241022-v2:0",
		EmbedModel:   "amazon.titan-embed-text-v2:0",
		Temperature:  0.7,
		MaxTokens:    4096,
		Instructions: "You are a helpful assistant.",
	}
}

// DefaultCohereConfig returns a default Cohere configuration with the given API key.
func DefaultCohereConfig(apiKey string) *CohereConfig {
	return &CohereConfig{
		APIKey:       apiKey,
		Model:        "command-r-plus",
		EmbedModel:   "embed-english-v3.0",
		Temperature:  0.7,
		MaxTokens:    4096,
		Instructions: "You are a helpful assistant.",
	}
}
