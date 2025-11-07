package llm

import (
	"os"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

func TestDefaultOllamaConfig(t *testing.T) {
	config := DefaultOllamaConfig()

	if config.BaseURL != "http://localhost:11434" {
		t.Errorf("Expected BaseURL to be http://localhost:11434, got %s", config.BaseURL)
	}

	if config.Model != "llama2" {
		t.Errorf("Expected Model to be llama2, got %s", config.Model)
	}

	if config.EmbedModel != "nomic-embed-text" {
		t.Errorf("Expected EmbedModel to be nomic-embed-text, got %s", config.EmbedModel)
	}

	if config.Temperature != 0.7 {
		t.Errorf("Expected Temperature to be 0.7, got %f", config.Temperature)
	}

	if config.MaxTokens != 2000 {
		t.Errorf("Expected MaxTokens to be 2000, got %d", config.MaxTokens)
	}
}

func TestDefaultOpenAIConfig(t *testing.T) {
	apiKey := "test-api-key"
	config := DefaultOpenAIConfig(apiKey)

	if config.APIKey != apiKey {
		t.Errorf("Expected APIKey to be %s, got %s", apiKey, config.APIKey)
	}

	if config.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected Model to be gpt-3.5-turbo, got %s", config.Model)
	}

	if config.EmbedModel != "text-embedding-ada-002" {
		t.Errorf("Expected EmbedModel to be text-embedding-ada-002, got %s", config.EmbedModel)
	}

	if config.Temperature != 0.7 {
		t.Errorf("Expected Temperature to be 0.7, got %f", config.Temperature)
	}

	if config.MaxTokens != 2000 {
		t.Errorf("Expected MaxTokens to be 2000, got %d", config.MaxTokens)
	}
}

func TestNewOllamaProvider(t *testing.T) {
	config := DefaultOllamaConfig()
	provider, err := NewOllamaProvider(config)

	if err != nil {
		t.Fatalf("Failed to create Ollama provider: %v", err)
	}

	if provider == nil {
		t.Fatal("Provider should not be nil")
	}

	if provider.config.BaseURL != config.BaseURL {
		t.Errorf("Provider config BaseURL doesn't match")
	}
}

func TestNewOllamaProvider_NilConfig(t *testing.T) {
	provider, err := NewOllamaProvider(nil)

	if err != nil {
		t.Fatalf("Should create provider with nil config: %v", err)
	}

	if provider == nil {
		t.Fatal("Provider should not be nil")
	}

	// Should use defaults
	if provider.config.BaseURL != "http://localhost:11434" {
		t.Errorf("Expected default BaseURL")
	}
}

func TestNewOpenAIProvider(t *testing.T) {
	client := openai.NewClient("test-api-key")
	config := DefaultOpenAIConfig("test-api-key")

	provider, err := NewOpenAIProvider(client, config)

	if err != nil {
		t.Fatalf("Failed to create OpenAI provider: %v", err)
	}

	if provider == nil {
		t.Fatal("Provider should not be nil")
	}
}

func TestNewOpenAIProvider_NilClient(t *testing.T) {
	config := DefaultOpenAIConfig("test-api-key")

	provider, err := NewOpenAIProvider(nil, config)

	if err == nil {
		t.Fatal("Should return error for nil client")
	}

	if provider != nil {
		t.Fatal("Provider should be nil when client is nil")
	}
}

func TestNewProvider_Ollama(t *testing.T) {
	config := &Config{
		Provider:     "ollama",
		OllamaConfig: DefaultOllamaConfig(),
	}

	provider, err := NewProvider(config)

	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if provider == nil {
		t.Fatal("Provider should not be nil")
	}

	// Verify it implements the interface
	var _ LLMProvider = provider
}

func TestNewProvider_OpenAI(t *testing.T) {
	config := &Config{
		Provider:     "openai",
		OpenAIConfig: DefaultOpenAIConfig("test-api-key"),
	}

	provider, err := NewProvider(config)

	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if provider == nil {
		t.Fatal("Provider should not be nil")
	}

	// Verify it implements the interface
	var _ LLMProvider = provider
}

func TestNewProvider_InvalidProvider(t *testing.T) {
	config := &Config{
		Provider: "invalid",
	}

	provider, err := NewProvider(config)

	if err == nil {
		t.Fatal("Should return error for invalid provider")
	}

	if provider != nil {
		t.Fatal("Provider should be nil for invalid provider type")
	}
}

func TestNewProvider_NilConfig(t *testing.T) {
	provider, err := NewProvider(nil)

	if err == nil {
		t.Fatal("Should return error for nil config")
	}

	if provider != nil {
		t.Fatal("Provider should be nil when config is nil")
	}
}

func TestNewProviderFromEnv_Ollama(t *testing.T) {
	// Save original values
	origProvider := os.Getenv("LLM_PROVIDER")
	origBaseURL := os.Getenv("OLLAMA_BASE_URL")

	// Set test values
	os.Setenv("LLM_PROVIDER", "ollama")
	os.Setenv("OLLAMA_BASE_URL", "http://test:11434")

	// Restore original values after test
	defer func() {
		if origProvider != "" {
			os.Setenv("LLM_PROVIDER", origProvider)
		} else {
			os.Unsetenv("LLM_PROVIDER")
		}
		if origBaseURL != "" {
			os.Setenv("OLLAMA_BASE_URL", origBaseURL)
		} else {
			os.Unsetenv("OLLAMA_BASE_URL")
		}
	}()

	provider, err := NewProviderFromEnv()

	if err != nil {
		t.Fatalf("Failed to create provider from env: %v", err)
	}

	if provider == nil {
		t.Fatal("Provider should not be nil")
	}

	// Verify it's an Ollama provider
	if _, ok := provider.(*OllamaProvider); !ok {
		t.Fatal("Expected OllamaProvider")
	}
}

func TestOllamaProvider_BuildPromptWithContext(t *testing.T) {
	provider, _ := NewOllamaProvider(DefaultOllamaConfig())

	tests := []struct {
		name     string
		prompt   string
		contexts []string
		contains []string
	}{
		{
			name:     "No context",
			prompt:   "What is AI?",
			contexts: nil,
			contains: []string{"What is AI?"},
		},
		{
			name:   "With context",
			prompt: "What is AI?",
			contexts: []string{
				"AI stands for Artificial Intelligence",
				"It is a branch of computer science",
			},
			contains: []string{
				"Use the following context",
				"What is AI?",
				"AI stands for Artificial Intelligence",
				"It is a branch of computer science",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.buildPromptWithContext(tt.prompt, tt.contexts)

			for _, expected := range tt.contains {
				if !contains(result, expected) {
					t.Errorf("Expected prompt to contain '%s', but it didn't. Got: %s", expected, result)
				}
			}
		})
	}
}

func TestOpenAIProvider_BuildPromptWithContext(t *testing.T) {
	client := openai.NewClient("test-key")
	provider, _ := NewOpenAIProvider(client, DefaultOpenAIConfig("test-key"))

	tests := []struct {
		name     string
		prompt   string
		contexts []string
		contains []string
	}{
		{
			name:     "No context",
			prompt:   "What is AI?",
			contexts: nil,
			contains: []string{"What is AI?"},
		},
		{
			name:   "With context",
			prompt: "What is AI?",
			contexts: []string{
				"AI stands for Artificial Intelligence",
				"It is a branch of computer science",
			},
			contains: []string{
				"Use the following context",
				"What is AI?",
				"AI stands for Artificial Intelligence",
				"It is a branch of computer science",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.buildPromptWithContext(tt.prompt, tt.contexts)

			for _, expected := range tt.contains {
				if !contains(result, expected) {
					t.Errorf("Expected prompt to contain '%s', but it didn't. Got: %s", expected, result)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
