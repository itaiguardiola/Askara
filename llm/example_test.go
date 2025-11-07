package llm_test

import (
	"fmt"
	"log"
	"os"

	"github.com/itaiguardiola/askara/llm"
	openai "github.com/sashabaranov/go-openai"
)

// ExampleNewProviderFromEnv demonstrates creating a provider from environment variables.
func ExampleNewProviderFromEnv() {
	// Set environment variables (typically done outside the application)
	os.Setenv("LLM_PROVIDER", "openai")
	os.Setenv("OPENAI_API_KEY", "sk-...")

	provider, err := llm.NewProviderFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	// Generate embedding
	embedding, err := provider.GenerateEmbedding("Hello, world!")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Generated embedding with %d dimensions\n", len(embedding))
}

// ExampleNewProvider_ollama demonstrates creating an Ollama provider with custom config.
func ExampleNewProvider_ollama() {
	config := &llm.Config{
		Provider: "ollama",
		OllamaConfig: &llm.OllamaConfig{
			BaseURL:     "http://localhost:11434",
			Model:       "llama2",
			EmbedModel:  "nomic-embed-text",
			Temperature: 0.7,
			MaxTokens:   2000,
		},
	}

	provider, err := llm.NewProvider(config)
	if err != nil {
		log.Fatal(err)
	}

	// Generate completion
	response, err := provider.GenerateCompletion("What is AI?", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response)
}

// ExampleNewProvider_openai demonstrates creating an OpenAI provider.
func ExampleNewProvider_openai() {
	apiKey := os.Getenv("OPENAI_API_KEY")

	config := &llm.Config{
		Provider: "openai",
		OpenAIConfig: &llm.OpenAIConfig{
			APIKey:       apiKey,
			Model:        "gpt-3.5-turbo",
			EmbedModel:   "text-embedding-ada-002",
			Temperature:  0.7,
			MaxTokens:    2000,
			Instructions: "You are a helpful assistant.",
		},
	}

	provider, err := llm.NewProvider(config)
	if err != nil {
		log.Fatal(err)
	}

	// Generate completion with context
	contexts := []string{
		"The Earth orbits around the Sun.",
		"It takes approximately 365.25 days for one complete orbit.",
	}

	response, err := provider.GenerateCompletion(
		"How long does it take Earth to orbit the Sun?",
		contexts,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response)
}

// ExampleOpenAIProvider_StreamCompletion demonstrates streaming completions.
func ExampleOpenAIProvider_StreamCompletion() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(apiKey)

	config := &llm.OpenAIConfig{
		APIKey:       apiKey,
		Model:        "gpt-3.5-turbo",
		Temperature:  0.7,
		MaxTokens:    2000,
		Instructions: "You are a helpful assistant.",
	}

	provider, err := llm.NewOpenAIProvider(client, config)
	if err != nil {
		log.Fatal(err)
	}

	// Stream completion
	err = provider.StreamCompletion("Tell me a short story", nil, func(chunk string) {
		fmt.Print(chunk)
	})
	if err != nil {
		log.Fatal(err)
	}
}

// ExampleOllamaProvider_StreamCompletion demonstrates streaming with Ollama.
func ExampleOllamaProvider_StreamCompletion() {
	config := &llm.OllamaConfig{
		BaseURL:     "http://localhost:11434",
		Model:       "llama2",
		Temperature: 0.7,
		MaxTokens:   2000,
	}

	provider, err := llm.NewOllamaProvider(config)
	if err != nil {
		log.Fatal(err)
	}

	// Stream completion with context
	contexts := []string{
		"You are writing a story about a brave knight.",
		"The knight is on a quest to save the kingdom.",
	}

	err = provider.StreamCompletion("Continue the story", contexts, func(chunk string) {
		fmt.Print(chunk)
	})
	if err != nil {
		log.Fatal(err)
	}
}
