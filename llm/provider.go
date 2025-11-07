package llm

// LLMProvider defines the interface for language model providers.
// Implementations include Ollama and OpenAI providers.
type LLMProvider interface {
	// GenerateEmbedding generates an embedding vector for the given text.
	// Returns a slice of float32 values representing the embedding.
	GenerateEmbedding(text string) ([]float32, error)

	// GenerateCompletion generates a completion for the given prompt with optional context.
	// The context parameter provides additional information to inform the completion.
	// Returns the generated text completion.
	GenerateCompletion(prompt string, context []string) (string, error)

	// StreamCompletion generates a completion with streaming support.
	// The onChunk callback is called for each chunk of text as it's generated.
	// This enables real-time response streaming to clients.
	StreamCompletion(prompt string, context []string, onChunk func(string)) error
}
