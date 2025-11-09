package prompt

import (
	"fmt"
	"strings"
)

// TemplateBuilder builds prompts using templates based on question type
type TemplateBuilder struct {
	config PromptConfig
}

// NewTemplateBuilder creates a new template builder with given config
func NewTemplateBuilder(config PromptConfig) *TemplateBuilder {
	return &TemplateBuilder{config: config}
}

// BuildPrompt creates a complete prompt from question and context
func (tb *TemplateBuilder) BuildPrompt(question string, contexts []ContextChunk, questionType QuestionType) string {
	// Format context chunks
	formattedContext := tb.formatContext(contexts)

	// Get template for question type
	template := tb.getTemplate(questionType)

	// Build complete prompt
	prompt := strings.Replace(template, "{CONTEXT}", formattedContext, 1)
	prompt = strings.Replace(prompt, "{QUESTION}", question, 1)

	return prompt
}

// formatContext formats context chunks according to configuration
func (tb *TemplateBuilder) formatContext(contexts []ContextChunk) string {
	if len(contexts) == 0 {
		return "No relevant context found."
	}

	var parts []string
	totalLength := 0

	for i, ctx := range contexts {
		// Check if adding this chunk would exceed max length
		if tb.config.MaxContextLength > 0 && totalLength+len(ctx.Text) > tb.config.MaxContextLength {
			break
		}

		var chunk string
		if tb.config.NumberContextChunks {
			chunk = fmt.Sprintf("[Context %d]", i+1)
			if tb.config.IncludeSourceInfo && ctx.Source != "" {
				chunk += fmt.Sprintf(" (Source: %s, Relevance: %.2f)", ctx.Source, ctx.Score)
			}
			chunk += "\n" + ctx.Text
		} else {
			if tb.config.IncludeSourceInfo && ctx.Source != "" {
				chunk = fmt.Sprintf("Source: %s\n%s", ctx.Source, ctx.Text)
			} else {
				chunk = ctx.Text
			}
		}

		parts = append(parts, chunk)
		totalLength += len(ctx.Text)
	}

	return strings.Join(parts, tb.config.ContextSeparator)
}

// getTemplate returns the appropriate template for the question type
func (tb *TemplateBuilder) getTemplate(questionType QuestionType) string {
	switch questionType {
	case QuestionTypeSummary:
		return tb.getSummaryTemplate()
	case QuestionTypeComparison:
		return tb.getComparisonTemplate()
	case QuestionTypeExtraction:
		return tb.getExtractionTemplate()
	case QuestionTypeAnalytical:
		return tb.getAnalyticalTemplate()
	case QuestionTypeConversational:
		return tb.getConversationalTemplate()
	default:
		return tb.getDirectTemplate()
	}
}

func (tb *TemplateBuilder) getDirectTemplate() string {
	return `Based on the following context from documents, please provide a clear and accurate answer to the question.

Context:
{CONTEXT}

Question: {QUESTION}

Instructions:
- Answer directly and concisely
- Use only information from the provided context
- If the context doesn't contain enough information, acknowledge this
- Cite specific parts of the context when relevant

Answer:`
}

func (tb *TemplateBuilder) getSummaryTemplate() string {
	return `Based on the following context from documents, provide a comprehensive summary.

Context:
{CONTEXT}

Task: {QUESTION}

Instructions:
- Create a well-structured summary with main points
- Maintain accuracy - only include information from the context
- Organize information logically
- Be concise but comprehensive
- Use bullet points or paragraphs as appropriate

Summary:`
}

func (tb *TemplateBuilder) getComparisonTemplate() string {
	return `Based on the following context from documents, provide a detailed comparison.

Context:
{CONTEXT}

Question: {QUESTION}

Instructions:
- Identify key similarities and differences
- Organize the comparison clearly (consider using a structured format)
- Be balanced and objective
- Base your comparison only on information in the context
- Highlight the most significant points

Comparison:`
}

func (tb *TemplateBuilder) getExtractionTemplate() string {
	return `Based on the following context from documents, extract the requested information.

Context:
{CONTEXT}

Task: {QUESTION}

Instructions:
- Extract all relevant information accurately
- Organize the extracted information in a clear, easy-to-read format
- Use lists or structured formats when appropriate
- Include source references if multiple sources are present
- Only include information explicitly stated in the context

Extracted Information:`
}

func (tb *TemplateBuilder) getAnalyticalTemplate() string {
	return `Based on the following context from documents, provide a thorough analysis.

Context:
{CONTEXT}

Question: {QUESTION}

Instructions:
- Analyze the information deeply and thoughtfully
- Explain reasoning and connections
- Consider multiple perspectives if relevant
- Support your analysis with evidence from the context
- Discuss implications or significance when appropriate
- Be clear about what is stated vs. what is inferred

Analysis:`
}

func (tb *TemplateBuilder) getConversationalTemplate() string {
	return `Use the following context from documents to help answer the question naturally.

Context:
{CONTEXT}

Question: {QUESTION}

Instructions:
- Respond in a natural, conversational way
- Be helpful and informative
- Base your answer on the provided context
- If the context is limited, acknowledge it politely

Response:`
}

// BuildSystemPrompt creates a system prompt based on question type and config
func (tb *TemplateBuilder) BuildSystemPrompt(questionType QuestionType) string {
	baseInstructions := GetSystemInstructions(questionType)

	if tb.config.SystemInstructions != "" {
		return baseInstructions + "\n\n" + tb.config.SystemInstructions
	}

	return baseInstructions
}
