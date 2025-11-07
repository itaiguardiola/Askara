package llm

// Config holds the configuration for LLM providers.
type Config struct {
	// Provider specifies which LLM provider to use ("ollama" or "openai")
	Provider string

	// OllamaConfig holds Ollama-specific configuration
	OllamaConfig *OllamaConfig

	// OpenAIConfig holds OpenAI-specific configuration
	OpenAIConfig *OpenAIConfig
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
