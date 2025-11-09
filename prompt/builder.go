package prompt

import (
	"fmt"
	"strings"
)

// Builder is the main entry point for building prompts
type Builder struct {
	config PromptConfig
}

// NewBuilder creates a new prompt builder with default configuration
func NewBuilder() *Builder {
	return &Builder{
		config: DefaultPromptConfig(),
	}
}

// NewBuilderWithConfig creates a new prompt builder with custom configuration
func NewBuilderWithConfig(config PromptConfig) *Builder {
	return &Builder{
		config: config,
	}
}

// WithConfig returns a new builder with updated configuration
func (b *Builder) WithConfig(config PromptConfig) *Builder {
	return &Builder{config: config}
}

// WithSourceInfo enables/disables source information in context
func (b *Builder) WithSourceInfo(include bool) *Builder {
	newConfig := b.config
	newConfig.IncludeSourceInfo = include
	return &Builder{config: newConfig}
}

// WithMaxContextLength sets the maximum context length
func (b *Builder) WithMaxContextLength(length int) *Builder {
	newConfig := b.config
	newConfig.MaxContextLength = length
	return &Builder{config: newConfig}
}

// WithTemperature sets the suggested temperature
func (b *Builder) WithTemperature(temp float32) *Builder {
	newConfig := b.config
	newConfig.Temperature = temp
	return &Builder{config: newConfig}
}

// WithSystemInstructions adds custom system instructions
func (b *Builder) WithSystemInstructions(instructions string) *Builder {
	newConfig := b.config
	newConfig.SystemInstructions = instructions
	return &Builder{config: newConfig}
}

// Build generates a complete prompt from question and context
func (b *Builder) Build(question string, contextTexts []string) PromptResult {
	// Convert context texts to ContextChunks
	chunks := make([]ContextChunk, len(contextTexts))
	for i, text := range contextTexts {
		chunks[i] = ContextChunk{
			Text:     text,
			Position: i,
			Score:    1.0, // Default score if not provided
		}
	}

	return b.BuildWithChunks(question, chunks)
}

// BuildWithChunks generates a complete prompt from question and context chunks with metadata
func (b *Builder) BuildWithChunks(question string, chunks []ContextChunk) PromptResult {
	// Detect question type
	questionType := DetectQuestionType(question)

	// Update temperature based on question type if not explicitly set
	config := b.config
	if config.Temperature == 0.7 { // Default temperature
		config.Temperature = SuggestTemperature(questionType)
	}

	// Build template
	templateBuilder := NewTemplateBuilder(config)
	finalPrompt := templateBuilder.BuildPrompt(question, chunks, questionType)
	systemPrompt := templateBuilder.BuildSystemPrompt(questionType)

	// Calculate rough token estimate (rough: 4 chars per token)
	estimatedTokens := len(finalPrompt) / 4

	return PromptResult{
		FinalPrompt:     finalPrompt,
		SystemPrompt:    systemPrompt,
		QuestionType:    questionType,
		ContextUsed:     len(chunks),
		EstimatedTokens: estimatedTokens,
		Config:          config,
	}
}

// BuildSimple creates a simple prompt without advanced features (for backward compatibility)
func (b *Builder) BuildSimple(question string, contextTexts []string) string {
	result := b.Build(question, contextTexts)
	return result.FinalPrompt
}

// Helper functions for common use cases

// QuickPrompt creates a prompt with default settings
func QuickPrompt(question string, contextTexts []string) PromptResult {
	builder := NewBuilder()
	return builder.Build(question, contextTexts)
}

// QuickPromptString creates a simple prompt string with default settings
func QuickPromptString(question string, contextTexts []string) string {
	builder := NewBuilder()
	return builder.BuildSimple(question, contextTexts)
}

// OptimizedPrompt creates a prompt with optimized settings based on question analysis
func OptimizedPrompt(question string, chunks []ContextChunk) PromptResult {
	builder := NewBuilder()
	return builder.BuildWithChunks(question, chunks)
}

// estimateTokens provides a rough estimate of token count
func estimateTokens(text string) int {
	// Rough estimate: average of 4 characters per token
	// More sophisticated tokenizers could be used here
	return len(text) / 4
}

// TruncateContext truncates context to fit within token limit
func TruncateContext(contexts []string, maxTokens int) []string {
	var result []string
	totalTokens := 0

	for _, ctx := range contexts {
		tokens := estimateTokens(ctx)
		if totalTokens+tokens > maxTokens {
			break
		}
		result = append(result, ctx)
		totalTokens += tokens
	}

	return result
}

// FormatContextWithSources formats context chunks with source attribution
func FormatContextWithSources(chunks []ContextChunk) string {
	var parts []string
	for i, chunk := range chunks {
		var formatted string
		if chunk.Source != "" {
			formatted = fmt.Sprintf("[Context %d - Source: %s]\n%s", i+1, chunk.Source, chunk.Text)
		} else {
			formatted = fmt.Sprintf("[Context %d]\n%s", i+1, chunk.Text)
		}
		parts = append(parts, formatted)
	}
	return strings.Join(parts, "\n\n---\n\n")
}
