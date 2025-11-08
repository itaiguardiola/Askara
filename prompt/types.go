package prompt

// QuestionType represents different types of questions that can be asked about documents
type QuestionType int

const (
	// QuestionTypeDirect represents straightforward factual questions
	QuestionTypeDirect QuestionType = iota
	// QuestionTypeSummary represents requests for summaries or overviews
	QuestionTypeSummary
	// QuestionTypeComparison represents questions comparing multiple things
	QuestionTypeComparison
	// QuestionTypeExtraction represents requests to extract specific information
	QuestionTypeExtraction
	// QuestionTypeAnalytical represents questions requiring deep analysis
	QuestionTypeAnalytical
	// QuestionTypeConversational represents general conversation or clarification
	QuestionTypeConversational
)

// String returns the string representation of a QuestionType
func (qt QuestionType) String() string {
	return [...]string{
		"Direct",
		"Summary",
		"Comparison",
		"Extraction",
		"Analytical",
		"Conversational",
	}[qt]
}

// PromptConfig holds configuration for prompt generation
type PromptConfig struct {
	// IncludeSourceInfo adds document source information to context
	IncludeSourceInfo bool
	// MaxContextLength limits the total context length in characters
	MaxContextLength int
	// Temperature suggests creativity level for LLM (0.0-1.0)
	Temperature float32
	// SystemInstructions provides additional system-level instructions
	SystemInstructions string
	// ContextSeparator defines how to separate different context chunks
	ContextSeparator string
	// NumberContextChunks adds numbers to each context chunk
	NumberContextChunks bool
}

// DefaultPromptConfig returns sensible default configuration
func DefaultPromptConfig() PromptConfig {
	return PromptConfig{
		IncludeSourceInfo:   true,
		MaxContextLength:    8000,
		Temperature:         0.7,
		SystemInstructions:  "",
		ContextSeparator:    "\n\n---\n\n",
		NumberContextChunks: true,
	}
}

// ContextChunk represents a piece of context with metadata
type ContextChunk struct {
	Text     string
	Source   string
	Score    float32
	Position int
}

// PromptResult contains the generated prompt and metadata
type PromptResult struct {
	// FinalPrompt is the complete prompt to send to the LLM
	FinalPrompt string
	// SystemPrompt is the system-level instruction
	SystemPrompt string
	// QuestionType is the detected type of question
	QuestionType QuestionType
	// ContextUsed is the number of context chunks used
	ContextUsed int
	// EstimatedTokens is rough estimate of token count
	EstimatedTokens int
	// Config is the configuration used
	Config PromptConfig
}
