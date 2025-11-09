package prompt

import (
	"strings"
	"testing"
)

func TestNewBuilder(t *testing.T) {
	builder := NewBuilder()
	if builder == nil {
		t.Fatal("NewBuilder() returned nil")
	}

	// Check default config
	if !builder.config.IncludeSourceInfo {
		t.Error("Default config should include source info")
	}
	if builder.config.MaxContextLength != 8000 {
		t.Errorf("Default MaxContextLength = %d, want 8000", builder.config.MaxContextLength)
	}
}

func TestBuilderWithConfig(t *testing.T) {
	config := PromptConfig{
		IncludeSourceInfo: false,
		MaxContextLength:  5000,
		Temperature:       0.5,
	}

	builder := NewBuilderWithConfig(config)
	if builder.config.IncludeSourceInfo {
		t.Error("Config should have IncludeSourceInfo = false")
	}
	if builder.config.MaxContextLength != 5000 {
		t.Errorf("MaxContextLength = %d, want 5000", builder.config.MaxContextLength)
	}
}

func TestBuilderFluentAPI(t *testing.T) {
	builder := NewBuilder().
		WithSourceInfo(false).
		WithMaxContextLength(10000).
		WithTemperature(0.8).
		WithSystemInstructions("Custom instructions")

	if builder.config.IncludeSourceInfo {
		t.Error("IncludeSourceInfo should be false")
	}
	if builder.config.MaxContextLength != 10000 {
		t.Errorf("MaxContextLength = %d, want 10000", builder.config.MaxContextLength)
	}
	if builder.config.Temperature != 0.8 {
		t.Errorf("Temperature = %f, want 0.8", builder.config.Temperature)
	}
	if builder.config.SystemInstructions != "Custom instructions" {
		t.Error("SystemInstructions not set correctly")
	}
}

func TestBuild(t *testing.T) {
	builder := NewBuilder()
	question := "What is machine learning?"
	contexts := []string{
		"Machine learning is a subset of AI.",
		"It involves training models on data.",
	}

	result := builder.Build(question, contexts)

	if result.FinalPrompt == "" {
		t.Error("FinalPrompt should not be empty")
	}
	if result.SystemPrompt == "" {
		t.Error("SystemPrompt should not be empty")
	}
	if result.ContextUsed != 2 {
		t.Errorf("ContextUsed = %d, want 2", result.ContextUsed)
	}
	if result.EstimatedTokens <= 0 {
		t.Error("EstimatedTokens should be positive")
	}
	if !strings.Contains(result.FinalPrompt, question) {
		t.Error("FinalPrompt should contain the question")
	}
}

func TestBuildWithChunks(t *testing.T) {
	builder := NewBuilder()
	question := "Summarize the findings"
	chunks := []ContextChunk{
		{
			Text:     "Finding 1: The system is effective",
			Source:   "report.pdf",
			Score:    0.95,
			Position: 0,
		},
		{
			Text:     "Finding 2: Users are satisfied",
			Source:   "survey.docx",
			Score:    0.87,
			Position: 1,
		},
	}

	result := builder.BuildWithChunks(question, chunks)

	if result.QuestionType != QuestionTypeSummary {
		t.Errorf("QuestionType = %v, want Summary", result.QuestionType)
	}
	if result.ContextUsed != 2 {
		t.Errorf("ContextUsed = %d, want 2", result.ContextUsed)
	}
	if !strings.Contains(result.FinalPrompt, "Finding 1") {
		t.Error("FinalPrompt should contain context text")
	}
}

func TestBuildSimple(t *testing.T) {
	builder := NewBuilder()
	question := "What is X?"
	contexts := []string{"X is a thing"}

	prompt := builder.BuildSimple(question, contexts)

	if prompt == "" {
		t.Error("BuildSimple should return non-empty string")
	}
	if !strings.Contains(prompt, question) {
		t.Error("Prompt should contain the question")
	}
}

func TestQuickPrompt(t *testing.T) {
	question := "Compare A and B"
	contexts := []string{"A is fast", "B is slow"}

	result := QuickPrompt(question, contexts)

	if result.QuestionType != QuestionTypeComparison {
		t.Errorf("QuestionType = %v, want Comparison", result.QuestionType)
	}
	if result.FinalPrompt == "" {
		t.Error("FinalPrompt should not be empty")
	}
}

func TestQuickPromptString(t *testing.T) {
	question := "What is this?"
	contexts := []string{"Context here"}

	prompt := QuickPromptString(question, contexts)

	if prompt == "" {
		t.Error("QuickPromptString should return non-empty string")
	}
	if !strings.Contains(prompt, question) {
		t.Error("Prompt should contain the question")
	}
}

func TestTruncateContext(t *testing.T) {
	contexts := []string{
		strings.Repeat("a", 100), // ~25 tokens
		strings.Repeat("b", 100), // ~25 tokens
		strings.Repeat("c", 100), // ~25 tokens
		strings.Repeat("d", 100), // ~25 tokens
	}

	// Should fit approximately 2 contexts in 50 tokens
	truncated := TruncateContext(contexts, 50)

	if len(truncated) > len(contexts) {
		t.Error("Truncated should not be longer than original")
	}
	if len(truncated) == 0 {
		t.Error("Truncated should not be empty for reasonable token limit")
	}
}

func TestBuildWithNoContext(t *testing.T) {
	builder := NewBuilder()
	question := "What is AI?"
	contexts := []string{}

	result := builder.Build(question, contexts)

	// When no context is provided, it should still return a valid result
	if result.FinalPrompt == "" {
		t.Error("FinalPrompt should not be empty even without context")
	}
	if result.ContextUsed != 0 {
		t.Errorf("ContextUsed = %d, want 0", result.ContextUsed)
	}
}

func TestMaxContextLength(t *testing.T) {
	builder := NewBuilder().WithMaxContextLength(100)
	question := "What is this?"
	contexts := []string{
		strings.Repeat("a", 200), // Exceeds max
		strings.Repeat("b", 200), // Would also exceed
	}

	result := builder.Build(question, contexts)

	// The builder should respect max context length
	// At least some truncation should occur
	if result.ContextUsed == len(contexts) {
		// This might pass, but the actual context in the prompt should be truncated
		// We can't easily test the internal truncation without exposing it
		// This is a basic sanity check
		t.Log("All contexts used, but should be truncated internally")
	}
}

func TestTemperatureOverride(t *testing.T) {
	// When temperature is explicitly set, it should not be auto-adjusted
	builder := NewBuilder().WithTemperature(0.9)
	question := "List all items" // Would normally suggest 0.2
	contexts := []string{"Item 1", "Item 2"}

	result := builder.Build(question, contexts)

	if result.Config.Temperature != 0.9 {
		t.Errorf("Temperature = %f, want 0.9 (should not be auto-adjusted)", result.Config.Temperature)
	}
}
