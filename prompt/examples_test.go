package prompt_test

import (
	"fmt"

	"github.com/itaiguardiola/askara/prompt"
)

// Example_basic demonstrates basic usage of the prompt builder
func Example_basic() {
	question := "What are the main features of this product?"
	contexts := []string{
		"The product includes AI-powered search capabilities.",
		"It supports multiple file formats including PDF and DOCX.",
	}

	result := prompt.QuickPrompt(question, contexts)

	fmt.Println("Question Type:", result.QuestionType)
	fmt.Println("Temperature:", result.Config.Temperature)
	fmt.Println("Contexts Used:", result.ContextUsed)
	// Output will include formatted prompt
}

// Example_summary demonstrates summary-type questions
func Example_summary() {
	question := "Summarize the main points of these documents"
	contexts := []string{
		"Document 1 discusses the benefits of cloud computing...",
		"Document 2 outlines the security considerations...",
	}

	result := prompt.QuickPrompt(question, contexts)

	fmt.Println("Detected Type:", result.QuestionType)
	// Output: Detected Type: Summary
}

// Example_comparison demonstrates comparison-type questions
func Example_comparison() {
	question := "Compare the performance of Algorithm A vs Algorithm B"
	contexts := []string{
		"Algorithm A completed in 2.5 seconds with 95% accuracy",
		"Algorithm B completed in 4.1 seconds with 98% accuracy",
	}

	result := prompt.QuickPrompt(question, contexts)

	fmt.Println("Detected Type:", result.QuestionType)
	// Output: Detected Type: Comparison
}

// Example_extraction demonstrates extraction-type questions
func Example_extraction() {
	question := "List all the authors mentioned in these papers"
	contexts := []string{
		"Paper by Dr. Smith and Prof. Johnson",
		"Research conducted by Dr. Williams",
	}

	result := prompt.QuickPrompt(question, contexts)

	fmt.Println("Detected Type:", result.QuestionType)
	fmt.Printf("Temperature: %.1f\n", result.Config.Temperature)
	// Output:
	// Detected Type: Extraction
	// Temperature: 0.2
}

// Example_withMetadata demonstrates using context chunks with metadata
func Example_withMetadata() {
	question := "What are the recommendations?"

	chunks := []prompt.ContextChunk{
		{
			Text:     "We recommend implementing the new system gradually.",
			Source:   "report.pdf",
			Score:    0.95,
			Position: 0,
		},
		{
			Text:     "The analysis suggests a phased approach.",
			Source:   "analysis.docx",
			Score:    0.87,
			Position: 1,
		},
	}

	result := prompt.OptimizedPrompt(question, chunks)

	fmt.Println("Contexts Used:", result.ContextUsed)
	// Output: Contexts Used: 2
}

// Example_customConfig demonstrates custom configuration
func Example_customConfig() {
	config := prompt.PromptConfig{
		IncludeSourceInfo:   true,
		MaxContextLength:    5000,
		Temperature:         0.5,
		SystemInstructions:  "You are a technical expert.",
		ContextSeparator:    "\n\n---\n\n",
		NumberContextChunks: true,
	}

	builder := prompt.NewBuilderWithConfig(config)

	question := "How does this work?"
	contexts := []string{"Technical explanation here..."}

	result := builder.Build(question, contexts)

	fmt.Printf("Custom Temperature: %.1f\n", result.Config.Temperature)
	// Output: Custom Temperature: 0.5
}

// Example_fluentBuilder demonstrates fluent builder pattern
func Example_fluentBuilder() {
	question := "Explain the architecture"
	contexts := []string{
		"The system uses a microservices architecture...",
	}

	result := prompt.NewBuilder().
		WithSourceInfo(true).
		WithMaxContextLength(10000).
		WithTemperature(0.6).
		Build(question, contexts)

	fmt.Printf("Temperature: %.1f\n", result.Config.Temperature)
	// Output: Temperature: 0.6
}

// Example_detectQuestionType demonstrates question type detection
func Example_detectQuestionType() {
	questions := []string{
		"What is the definition?",
		"Summarize the key findings",
		"Compare X and Y",
		"List all the features",
		"Why did this happen?",
	}

	for _, q := range questions {
		qType := prompt.DetectQuestionType(q)
		fmt.Printf("%s: %s\n", q, qType)
	}
	// Output:
	// What is the definition?: Direct
	// Summarize the key findings: Summary
	// Compare X and Y: Comparison
	// List all the features: Extraction
	// Why did this happen?: Analytical
}

// Example_temperatureSuggestions demonstrates temperature suggestions
func Example_temperatureSuggestions() {
	types := []prompt.QuestionType{
		prompt.QuestionTypeExtraction,
		prompt.QuestionTypeDirect,
		prompt.QuestionTypeComparison,
		prompt.QuestionTypeSummary,
		prompt.QuestionTypeAnalytical,
	}

	for _, qType := range types {
		temp := prompt.SuggestTemperature(qType)
		fmt.Printf("%s: %.1f\n", qType, temp)
	}
	// Output:
	// Extraction: 0.2
	// Direct: 0.3
	// Comparison: 0.5
	// Summary: 0.6
	// Analytical: 0.7
}

// Example_truncateContext demonstrates context truncation
func Example_truncateContext() {
	contexts := []string{
		"First context with some content",
		"Second context with more content",
		"Third context with even more content",
	}

	// Truncate to fit within 50 tokens (rough estimate)
	truncated := prompt.TruncateContext(contexts, 50)

	fmt.Println("Original:", len(contexts))
	fmt.Println("Truncated:", len(truncated))
	// Output will depend on actual token counts
}
