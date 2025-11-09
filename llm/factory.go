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

	// ProviderTypeClaude represents the Anthropic Claude provider.
	ProviderTypeClaude ProviderType = "claude"

	// ProviderTypeGemini represents the Google Gemini provider.
	ProviderTypeGemini ProviderType = "gemini"

	// ProviderTypeGroq represents the Groq provider.
	ProviderTypeGroq ProviderType = "groq"

	// ProviderTypeAzure represents the Azure OpenAI provider.
	ProviderTypeAzure ProviderType = "azure"

	// ProviderTypeBedrock represents the AWS Bedrock provider.
	ProviderTypeBedrock ProviderType = "bedrock"

	// ProviderTypeCohere represents the Cohere provider.
	ProviderTypeCohere ProviderType = "cohere"
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
	case ProviderTypeClaude:
		return newClaudeProviderFromConfig(config)
	case ProviderTypeGemini:
		return newGeminiProviderFromConfig(config)
	case ProviderTypeGroq:
		return newGroqProviderFromConfig(config)
	case ProviderTypeAzure:
		return newAzureProviderFromConfig(config)
	case ProviderTypeBedrock:
		return newBedrockProviderFromConfig(config)
	case ProviderTypeCohere:
		return newCohereProviderFromConfig(config)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s (supported: ollama, openai, claude, gemini, groq, azure, bedrock, cohere)", config.Provider)
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
	// Check OLLAMA_HOST first (preferred), then fall back to OLLAMA_BASE_URL
	if baseURL := os.Getenv("OLLAMA_HOST"); baseURL != "" {
		ollamaConfig.BaseURL = baseURL
		log.Printf("[Factory] Using OLLAMA_HOST from environment: %s", baseURL)
	} else if baseURL := os.Getenv("OLLAMA_BASE_URL"); baseURL != "" {
		ollamaConfig.BaseURL = baseURL
		log.Printf("[Factory] Using OLLAMA_BASE_URL from environment: %s", baseURL)
	}

	if model := os.Getenv("OLLAMA_MODEL"); model != "" {
		ollamaConfig.Model = model
		log.Printf("[Factory] Using OLLAMA_MODEL from environment: %s", model)
	}

	// Check OLLAMA_EMBEDDING_MODEL first, then OLLAMA_EMBED_MODEL
	if embedModel := os.Getenv("OLLAMA_EMBEDDING_MODEL"); embedModel != "" {
		ollamaConfig.EmbedModel = embedModel
		log.Printf("[Factory] Using OLLAMA_EMBEDDING_MODEL from environment: %s", embedModel)
	} else if embedModel := os.Getenv("OLLAMA_EMBED_MODEL"); embedModel != "" {
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

// newClaudeProviderFromConfig creates a Claude provider from the config.
func newClaudeProviderFromConfig(config *Config) (LLMProvider, error) {
	claudeConfig := config.ClaudeConfig
	if claudeConfig == nil {
		// Try to get API key from environment
		apiKey := os.Getenv("CLAUDE_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("Claude API key not provided in config or CLAUDE_API_KEY environment variable")
		}
		claudeConfig = DefaultClaudeConfig(apiKey)
		log.Println("[Factory] Using default Claude configuration with environment API key")
	}

	// Override with environment variables if set
	if apiKey := os.Getenv("CLAUDE_API_KEY"); apiKey != "" && claudeConfig.APIKey == "" {
		claudeConfig.APIKey = apiKey
		log.Println("[Factory] Using CLAUDE_API_KEY from environment")
	}

	if model := os.Getenv("CLAUDE_MODEL"); model != "" {
		claudeConfig.Model = model
		log.Printf("[Factory] Using CLAUDE_MODEL from environment: %s", model)
	}

	if claudeConfig.APIKey == "" {
		return nil, fmt.Errorf("Claude API key is required")
	}

	provider, err := NewClaudeProvider(claudeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Claude provider: %w", err)
	}

	log.Printf("[Factory] Created Claude provider (Model: %s)", claudeConfig.Model)

	return provider, nil
}

// newGeminiProviderFromConfig creates a Gemini provider from the config.
func newGeminiProviderFromConfig(config *Config) (LLMProvider, error) {
	geminiConfig := config.GeminiConfig
	if geminiConfig == nil {
		// Try to get API key from environment
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("Gemini API key not provided in config or GEMINI_API_KEY environment variable")
		}
		geminiConfig = DefaultGeminiConfig(apiKey)
		log.Println("[Factory] Using default Gemini configuration with environment API key")
	}

	// Override with environment variables if set
	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" && geminiConfig.APIKey == "" {
		geminiConfig.APIKey = apiKey
		log.Println("[Factory] Using GEMINI_API_KEY from environment")
	}

	if model := os.Getenv("GEMINI_MODEL"); model != "" {
		geminiConfig.Model = model
		log.Printf("[Factory] Using GEMINI_MODEL from environment: %s", model)
	}

	if geminiConfig.APIKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	provider, err := NewGeminiProvider(geminiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini provider: %w", err)
	}

	log.Printf("[Factory] Created Gemini provider (Model: %s)", geminiConfig.Model)

	return provider, nil
}

// newGroqProviderFromConfig creates a Groq provider from the config.
func newGroqProviderFromConfig(config *Config) (LLMProvider, error) {
	groqConfig := config.GroqConfig
	if groqConfig == nil {
		// Try to get API key from environment
		apiKey := os.Getenv("GROQ_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("Groq API key not provided in config or GROQ_API_KEY environment variable")
		}
		groqConfig = DefaultGroqConfig(apiKey)
		log.Println("[Factory] Using default Groq configuration with environment API key")
	}

	// Override with environment variables if set
	if apiKey := os.Getenv("GROQ_API_KEY"); apiKey != "" && groqConfig.APIKey == "" {
		groqConfig.APIKey = apiKey
		log.Println("[Factory] Using GROQ_API_KEY from environment")
	}

	if model := os.Getenv("GROQ_MODEL"); model != "" {
		groqConfig.Model = model
		log.Printf("[Factory] Using GROQ_MODEL from environment: %s", model)
	}

	if groqConfig.APIKey == "" {
		return nil, fmt.Errorf("Groq API key is required")
	}

	provider, err := NewGroqProvider(groqConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Groq provider: %w", err)
	}

	log.Printf("[Factory] Created Groq provider (Model: %s)", groqConfig.Model)

	return provider, nil
}

// newAzureProviderFromConfig creates an Azure OpenAI provider from the config.
func newAzureProviderFromConfig(config *Config) (LLMProvider, error) {
	azureConfig := config.AzureConfig
	if azureConfig == nil {
		// Try to get configuration from environment
		apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
		endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
		deploymentID := os.Getenv("AZURE_OPENAI_DEPLOYMENT")

		if apiKey == "" || endpoint == "" || deploymentID == "" {
			return nil, fmt.Errorf("Azure OpenAI requires AZURE_OPENAI_API_KEY, AZURE_OPENAI_ENDPOINT, and AZURE_OPENAI_DEPLOYMENT environment variables")
		}

		azureConfig = DefaultAzureConfig(apiKey, endpoint, deploymentID)
		log.Println("[Factory] Using default Azure configuration with environment variables")
	}

	// Override with environment variables if set
	if apiKey := os.Getenv("AZURE_OPENAI_API_KEY"); apiKey != "" && azureConfig.APIKey == "" {
		azureConfig.APIKey = apiKey
		log.Println("[Factory] Using AZURE_OPENAI_API_KEY from environment")
	}

	if endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT"); endpoint != "" && azureConfig.Endpoint == "" {
		azureConfig.Endpoint = endpoint
		log.Printf("[Factory] Using AZURE_OPENAI_ENDPOINT from environment: %s", endpoint)
	}

	if deploymentID := os.Getenv("AZURE_OPENAI_DEPLOYMENT"); deploymentID != "" {
		azureConfig.DeploymentID = deploymentID
		log.Printf("[Factory] Using AZURE_OPENAI_DEPLOYMENT from environment: %s", deploymentID)
	}

	if embedDeploymentID := os.Getenv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT"); embedDeploymentID != "" {
		azureConfig.EmbeddingDeploymentID = embedDeploymentID
		log.Printf("[Factory] Using AZURE_OPENAI_EMBEDDING_DEPLOYMENT from environment: %s", embedDeploymentID)
	}

	if azureConfig.APIKey == "" || azureConfig.Endpoint == "" || azureConfig.DeploymentID == "" {
		return nil, fmt.Errorf("Azure OpenAI requires API key, endpoint, and deployment ID")
	}

	provider, err := NewAzureProvider(azureConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure provider: %w", err)
	}

	log.Printf("[Factory] Created Azure provider (Endpoint: %s, Deployment: %s)", azureConfig.Endpoint, azureConfig.DeploymentID)

	return provider, nil
}

// newBedrockProviderFromConfig creates an AWS Bedrock provider from the config.
func newBedrockProviderFromConfig(config *Config) (LLMProvider, error) {
	bedrockConfig := config.BedrockConfig
	if bedrockConfig == nil {
		// Try to get configuration from environment
		region := os.Getenv("AWS_REGION")
		if region == "" {
			region = "us-east-1" // Default region
		}

		bedrockConfig = DefaultBedrockConfig(region)
		log.Println("[Factory] Using default Bedrock configuration with environment variables")
	}

	// Override with environment variables if set
	if region := os.Getenv("AWS_REGION"); region != "" {
		bedrockConfig.Region = region
		log.Printf("[Factory] Using AWS_REGION from environment: %s", region)
	}

	if model := os.Getenv("BEDROCK_MODEL"); model != "" {
		bedrockConfig.Model = model
		log.Printf("[Factory] Using BEDROCK_MODEL from environment: %s", model)
	}

	if embedModel := os.Getenv("BEDROCK_EMBEDDING_MODEL"); embedModel != "" {
		bedrockConfig.EmbedModel = embedModel
		log.Printf("[Factory] Using BEDROCK_EMBEDDING_MODEL from environment: %s", embedModel)
	}

	provider, err := NewBedrockProvider(bedrockConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Bedrock provider: %w", err)
	}

	log.Printf("[Factory] Created Bedrock provider (Region: %s, Model: %s)", bedrockConfig.Region, bedrockConfig.Model)

	return provider, nil
}

// newCohereProviderFromConfig creates a Cohere provider from the config.
func newCohereProviderFromConfig(config *Config) (LLMProvider, error) {
	cohereConfig := config.CohereConfig
	if cohereConfig == nil {
		// Try to get API key from environment
		apiKey := os.Getenv("COHERE_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("Cohere API key not provided in config or COHERE_API_KEY environment variable")
		}
		cohereConfig = DefaultCohereConfig(apiKey)
		log.Println("[Factory] Using default Cohere configuration with environment API key")
	}

	// Override with environment variables if set
	if apiKey := os.Getenv("COHERE_API_KEY"); apiKey != "" && cohereConfig.APIKey == "" {
		cohereConfig.APIKey = apiKey
		log.Println("[Factory] Using COHERE_API_KEY from environment")
	}

	if model := os.Getenv("COHERE_MODEL"); model != "" {
		cohereConfig.Model = model
		log.Printf("[Factory] Using COHERE_MODEL from environment: %s", model)
	}

	if embedModel := os.Getenv("COHERE_EMBEDDING_MODEL"); embedModel != "" {
		cohereConfig.EmbedModel = embedModel
		log.Printf("[Factory] Using COHERE_EMBEDDING_MODEL from environment: %s", embedModel)
	}

	if cohereConfig.APIKey == "" {
		return nil, fmt.Errorf("Cohere API key is required")
	}

	provider, err := NewCohereProvider(cohereConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cohere provider: %w", err)
	}

	log.Printf("[Factory] Created Cohere provider (Model: %s)", cohereConfig.Model)

	return provider, nil
}

// NewProviderFromEnv creates a new LLM provider based on environment variables.
// This is a convenience function for simple setups.
//
// Environment variables:
//   - LLM_PROVIDER: "ollama", "openai", "claude", "gemini", "groq", "azure", "bedrock", or "cohere" (default: "openai")
//   - OLLAMA_BASE_URL: Ollama API endpoint (default: http://localhost:11434)
//   - OLLAMA_MODEL: Ollama model name (default: llama2)
//   - OLLAMA_EMBED_MODEL: Ollama embedding model (default: nomic-embed-text)
//   - OPENAI_API_KEY: OpenAI API key (required for OpenAI provider)
//   - CLAUDE_API_KEY: Anthropic API key (required for Claude provider)
//   - GEMINI_API_KEY: Google API key (required for Gemini provider)
//   - GROQ_API_KEY: Groq API key (required for Groq provider)
//   - AZURE_OPENAI_API_KEY, AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_DEPLOYMENT: Azure OpenAI configuration
//   - AWS_REGION, BEDROCK_MODEL: AWS Bedrock configuration
//   - COHERE_API_KEY: Cohere API key (required for Cohere provider)
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
		// Check OLLAMA_HOST first, then OLLAMA_BASE_URL, then use default
		baseURL := os.Getenv("OLLAMA_HOST")
		log.Printf("[Factory][DEBUG] OLLAMA_HOST='%s', OLLAMA_BASE_URL='%s'", baseURL, os.Getenv("OLLAMA_BASE_URL"))
		if baseURL == "" {
			baseURL = getEnvOrDefault("OLLAMA_BASE_URL", "http://localhost:11434")
		}
		// Check OLLAMA_EMBEDDING_MODEL first, then OLLAMA_EMBED_MODEL
		embedModel := os.Getenv("OLLAMA_EMBEDDING_MODEL")
		if embedModel == "" {
			embedModel = getEnvOrDefault("OLLAMA_EMBED_MODEL", "nomic-embed-text")
		}
		config.OllamaConfig = &OllamaConfig{
			BaseURL:     baseURL,
			Model:       getEnvOrDefault("OLLAMA_MODEL", "llama2"),
			EmbedModel:  embedModel,
			Temperature: 0.7,
			MaxTokens:   2000,
		}
		log.Printf("[Factory][DEBUG] Created OllamaConfig with BaseURL='%s'", baseURL)
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
	case "claude":
		apiKey := os.Getenv("CLAUDE_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("CLAUDE_API_KEY environment variable is required")
		}
		config.ClaudeConfig = &ClaudeConfig{
			APIKey:       apiKey,
			Model:        getEnvOrDefault("CLAUDE_MODEL", "claude-3-5-sonnet-20241022"),
			EmbedModel:   getEnvOrDefault("CLAUDE_EMBED_MODEL", "voyage-large-2-instruct"),
			Temperature:  0.7,
			MaxTokens:    4096,
			Instructions: "You are a helpful assistant.",
		}
	case "gemini":
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
		}
		config.GeminiConfig = &GeminiConfig{
			APIKey:       apiKey,
			Model:        getEnvOrDefault("GEMINI_MODEL", "gemini-1.5-flash"),
			EmbedModel:   getEnvOrDefault("GEMINI_EMBED_MODEL", "text-embedding-004"),
			Temperature:  0.7,
			MaxTokens:    8192,
			Instructions: "You are a helpful assistant.",
		}
	case "groq":
		apiKey := os.Getenv("GROQ_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("GROQ_API_KEY environment variable is required")
		}
		config.GroqConfig = &GroqConfig{
			APIKey:       apiKey,
			Model:        getEnvOrDefault("GROQ_MODEL", "mixtral-8x7b-32768"),
			EmbedModel:   "",
			Temperature:  0.7,
			MaxTokens:    4096,
			Instructions: "You are a helpful assistant.",
		}
	case "azure":
		apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
		endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
		deploymentID := os.Getenv("AZURE_OPENAI_DEPLOYMENT")

		if apiKey == "" || endpoint == "" || deploymentID == "" {
			return nil, fmt.Errorf("Azure OpenAI requires AZURE_OPENAI_API_KEY, AZURE_OPENAI_ENDPOINT, and AZURE_OPENAI_DEPLOYMENT environment variables")
		}

		config.AzureConfig = &AzureConfig{
			APIKey:                apiKey,
			Endpoint:              endpoint,
			DeploymentID:          deploymentID,
			EmbeddingDeploymentID: getEnvOrDefault("AZURE_OPENAI_EMBEDDING_DEPLOYMENT", "text-embedding-ada-002"),
			APIVersion:            getEnvOrDefault("AZURE_OPENAI_API_VERSION", "2023-05-15"),
			Temperature:           0.7,
			MaxTokens:             2000,
			Instructions:          "You are a helpful assistant.",
		}
	case "bedrock":
		region := getEnvOrDefault("AWS_REGION", "us-east-1")
		config.BedrockConfig = &BedrockConfig{
			Region:       region,
			Model:        getEnvOrDefault("BEDROCK_MODEL", "anthropic.claude-3-5-sonnet-20241022-v2:0"),
			EmbedModel:   getEnvOrDefault("BEDROCK_EMBEDDING_MODEL", "amazon.titan-embed-text-v2:0"),
			Temperature:  0.7,
			MaxTokens:    4096,
			Instructions: "You are a helpful assistant.",
		}
	case "cohere":
		apiKey := os.Getenv("COHERE_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("COHERE_API_KEY environment variable is required")
		}
		config.CohereConfig = &CohereConfig{
			APIKey:       apiKey,
			Model:        getEnvOrDefault("COHERE_MODEL", "command-r-plus"),
			EmbedModel:   getEnvOrDefault("COHERE_EMBEDDING_MODEL", "embed-english-v3.0"),
			Temperature:  0.7,
			MaxTokens:    4096,
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
