package prompt

import (
	"strings"
)

// DetectQuestionType analyzes a question and determines its type
func DetectQuestionType(question string) QuestionType {
	questionLower := strings.ToLower(strings.TrimSpace(question))

	// Summary patterns (check first as they're usually explicit)
	summaryPatterns := []string{
		"summarize", "summary", "overview", "brief", "tldr",
		"main points", "key points", "in short", "gist",
		"outline", "recap", "digest",
	}
	for _, pattern := range summaryPatterns {
		if strings.Contains(questionLower, pattern) {
			return QuestionTypeSummary
		}
	}

	// Comparison patterns (check early as they're also explicit)
	comparisonPatterns := []string{
		"compare", "comparison", "difference", "differ",
		"versus", "vs", "vs.", "contrast", "similar",
		"better than", "worse than", "advantage", "disadvantage",
		"which is better",
	}
	for _, pattern := range comparisonPatterns {
		if strings.Contains(questionLower, pattern) {
			return QuestionTypeComparison
		}
	}

	// Extraction patterns (check before analytical to catch "list all" before "all")
	extractionPatterns := []string{
		"list all", "find all", "extract", "identify all",
		"what are the", "which are", "enumerate", "name all",
		"show me all", "give me all", "find mentions of",
		"list the",
	}
	for _, pattern := range extractionPatterns {
		if strings.Contains(questionLower, pattern) {
			return QuestionTypeExtraction
		}
	}

	// Analytical patterns (why, how, explain) - but exclude conversational helpers
	analyticalPatterns := []string{
		"why ", "how does", "how can", "how would",
		"what causes", "what leads to", "reason for",
		"analyze", "analysis", "evaluate", "assess", "implications",
	}
	for _, pattern := range analyticalPatterns {
		if strings.Contains(questionLower, pattern) {
			return QuestionTypeAnalytical
		}
	}

	// Check for "explain" only if not preceded by conversational words
	if strings.Contains(questionLower, "explain") &&
	   !strings.Contains(questionLower, "please") &&
	   !strings.Contains(questionLower, "could you") &&
	   !strings.Contains(questionLower, "can you") {
		return QuestionTypeAnalytical
	}

	// Conversational patterns (check later to avoid false positives)
	// Only match if they're prominent (at start or clear intent)
	if strings.HasPrefix(questionLower, "thanks") ||
	   strings.HasPrefix(questionLower, "thank you") ||
	   strings.HasPrefix(questionLower, "hello") ||
	   strings.HasPrefix(questionLower, "hi ") ||
	   strings.HasPrefix(questionLower, "hi,") ||
	   strings.Contains(questionLower, "help me understand") {
		return QuestionTypeConversational
	}

	// Default to direct question
	return QuestionTypeDirect
}

// SuggestTemperature returns recommended temperature based on question type
func SuggestTemperature(questionType QuestionType) float32 {
	switch questionType {
	case QuestionTypeExtraction:
		return 0.2 // Very deterministic for factual extraction
	case QuestionTypeDirect:
		return 0.3 // Mostly factual
	case QuestionTypeComparison:
		return 0.5 // Balanced
	case QuestionTypeSummary:
		return 0.6 // Slight creativity for conciseness
	case QuestionTypeAnalytical:
		return 0.7 // More creative for insights
	case QuestionTypeConversational:
		return 0.8 // Natural conversation
	default:
		return 0.7
	}
}

// GetSystemInstructions returns appropriate system instructions for question type
func GetSystemInstructions(questionType QuestionType) string {
	switch questionType {
	case QuestionTypeSummary:
		return "You are a helpful assistant specializing in creating clear, concise summaries. Focus on the most important points and maintain accuracy."
	case QuestionTypeComparison:
		return "You are a helpful assistant specializing in comparative analysis. Provide balanced, structured comparisons highlighting key similarities and differences."
	case QuestionTypeExtraction:
		return "You are a helpful assistant specializing in information extraction. Be thorough, accurate, and organize extracted information clearly."
	case QuestionTypeAnalytical:
		return "You are a helpful assistant specializing in analytical thinking. Provide deep insights, explain reasoning, and explore implications."
	case QuestionTypeDirect:
		return "You are a helpful assistant providing accurate, factual answers. Be clear and concise while ensuring completeness."
	case QuestionTypeConversational:
		return "You are a helpful, friendly assistant. Provide natural, conversational responses while being informative."
	default:
		return "You are a helpful assistant. Answer questions accurately based on the provided context."
	}
}
