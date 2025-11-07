package llm

import (
	"fmt"
	"log"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// ProviderType represents the type of LLM provider.
type ProviderType string

const (
	// ProviderTypeOllama represents the Ollama provider.
	ProviderTypeOllama ProviderType = "ollama"

	// ProviderTypeOpenAI represents the OpenAI provider.
	ProviderTypeOpenAI ProviderType = "openai"
)

// NewProvider creates a new LLM provider based on the given configuration.
// This factory function determines which provider to instantiate based on the config.
func NewProvider(config *Config) (LLMProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	providerType := ProviderType(strings.ToLower(config.Provider))

	switch providerType {
	case ProviderTypeOllama:
		return newOllamaProviderFromConfig(config)
	case ProviderTypeOpenAI:
		return newOpenAIProviderFromConfig(config)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s (supported: ollama, openai)", config.Provider)
	}
}

// newOllamaProviderFromConfig creates an Ollama provider from the config.
func newOllamaProviderFromConfig(config *Config) (LLMProvider, error) {
	ollamaConfig := config.OllamaConfig
	if ollamaConfig == nil {
		ollamaConfig = DefaultOllamaConfig()
		log.Println("[Factory] Using default Ollama configuration")
	}

	// Override with environment variables if set
	if baseURL := os.Getenv("OLLAMA_BASE_URL"); baseURL != "" {
		ollamaConfig.BaseURL = baseURL
		log.Printf("[Factory] Using OLLAMA_BASE_URL from environment: %s", baseURL)
	}

	if model := os.Getenv("OLLAMA_MODEL"); model != "" {
		ollamaConfig.Model = model
		log.Printf("[Factory] Using OLLAMA_MODEL from environment: %s", model)
	}

	if embedModel := os.Getenv("OLLAMA_EMBED_MODEL"); embedModel != "" {
		ollamaConfig.EmbedModel = embedModel
		log.Printf("[Factory] Using OLLAMA_EMBED_MODEL from environment: %s", embedModel)
	}

	provider, err := NewOllamaProvider(ollamaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama provider: %w", err)
	}

	log.Printf("[Factory] Created Ollama provider (BaseURL: %s, Model: %s)",
		ollamaConfig.BaseURL, ollamaConfig.Model)

	return provider, nil
}

// newOpenAIProviderFromConfig creates an OpenAI provider from the config.
func newOpenAIProviderFromConfig(config *Config) (LLMProvider, error) {
	openaiConfig := config.OpenAIConfig
	if openaiConfig == nil {
		// Try to get API key from environment
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OpenAI API key not provided in config or OPENAI_API_KEY environment variable")
		}
		openaiConfig = DefaultOpenAIConfig(apiKey)
		log.Println("[Factory] Using default OpenAI configuration with environment API key")
	}

	// Override with environment variable if set
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" && openaiConfig.APIKey == "" {
		openaiConfig.APIKey = apiKey
		log.Println("[Factory] Using OPENAI_API_KEY from environment")
	}

	if openaiConfig.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	client := openai.NewClient(openaiConfig.APIKey)

	provider, err := NewOpenAIProvider(client, openaiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
	}

	log.Printf("[Factory] Created OpenAI provider (Model: %s)", openaiConfig.Model)

	return provider, nil
}

// NewProviderFromEnv creates a new LLM provider based on environment variables.
// This is a convenience function for simple setups.
//
// Environment variables:
//   - LLM_PROVIDER: "ollama" or "openai" (default: "openai")
//   - OLLAMA_BASE_URL: Ollama API endpoint (default: http://localhost:11434)
//   - OLLAMA_MODEL: Ollama model name (default: llama2)
//   - OLLAMA_EMBED_MODEL: Ollama embedding model (default: nomic-embed-text)
//   - OPENAI_API_KEY: OpenAI API key (required for OpenAI provider)
func NewProviderFromEnv() (LLMProvider, error) {
	providerType := os.Getenv("LLM_PROVIDER")
	if providerType == "" {
		providerType = "openai" // Default to OpenAI for backward compatibility
	}

	config := &Config{
		Provider: providerType,
	}

	switch strings.ToLower(providerType) {
	case "ollama":
		config.OllamaConfig = &OllamaConfig{
			BaseURL:     getEnvOrDefault("OLLAMA_BASE_URL", "http://localhost:11434"),
			Model:       getEnvOrDefault("OLLAMA_MODEL", "llama2"),
			EmbedModel:  getEnvOrDefault("OLLAMA_EMBED_MODEL", "nomic-embed-text"),
			Temperature: 0.7,
			MaxTokens:   2000,
		}
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
		}
		config.OpenAIConfig = &OpenAIConfig{
			APIKey:       apiKey,
			Model:        getEnvOrDefault("OPENAI_MODEL", "gpt-3.5-turbo"),
			EmbedModel:   getEnvOrDefault("OPENAI_EMBED_MODEL", "text-embedding-ada-002"),
			Temperature:  0.7,
			MaxTokens:    2000,
			Instructions: "You are a helpful assistant.",
		}
	default:
		return nil, fmt.Errorf("unsupported LLM_PROVIDER: %s", providerType)
	}

	return NewProvider(config)
}

// getEnvOrDefault returns the value of an environment variable or a default value.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
