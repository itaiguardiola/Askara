package prompt

import (
	"testing"
)

func TestDetectQuestionType(t *testing.T) {
	tests := []struct {
		question string
		expected QuestionType
	}{
		// Summary type
		{"Summarize the document", QuestionTypeSummary},
		{"Give me a summary of this", QuestionTypeSummary},
		{"What are the main points?", QuestionTypeSummary},
		{"Provide an overview", QuestionTypeSummary},
		{"TLDR", QuestionTypeSummary},

		// Comparison type
		{"Compare X and Y", QuestionTypeComparison},
		{"What are the differences between A and B?", QuestionTypeComparison},
		{"X versus Y", QuestionTypeComparison},
		{"Which is better?", QuestionTypeComparison},
		{"Contrast these approaches", QuestionTypeComparison},

		// Extraction type
		{"List all the authors", QuestionTypeExtraction},
		{"Find all mentions of X", QuestionTypeExtraction},
		{"Extract the key dates", QuestionTypeExtraction},
		{"What are the features?", QuestionTypeExtraction},
		{"Identify all instances", QuestionTypeExtraction},

		// Analytical type
		{"Why did this happen?", QuestionTypeAnalytical},
		{"How does this work?", QuestionTypeAnalytical},
		{"Explain the reasoning", QuestionTypeAnalytical},
		{"What causes this?", QuestionTypeAnalytical},
		{"Analyze the results", QuestionTypeAnalytical},

		// Direct type
		{"What is X?", QuestionTypeDirect},
		{"Who is the author?", QuestionTypeDirect},
		{"Where is the location?", QuestionTypeDirect},

		// Conversational type
		{"Thanks for the help", QuestionTypeConversational},
		{"Hello there", QuestionTypeConversational},
		{"Hi, I need help", QuestionTypeConversational},
	}

	for _, tt := range tests {
		t.Run(tt.question, func(t *testing.T) {
			result := DetectQuestionType(tt.question)
			if result != tt.expected {
				t.Errorf("DetectQuestionType(%q) = %v, want %v", tt.question, result, tt.expected)
			}
		})
	}
}

func TestQuestionTypeString(t *testing.T) {
	tests := []struct {
		qType    QuestionType
		expected string
	}{
		{QuestionTypeDirect, "Direct"},
		{QuestionTypeSummary, "Summary"},
		{QuestionTypeComparison, "Comparison"},
		{QuestionTypeExtraction, "Extraction"},
		{QuestionTypeAnalytical, "Analytical"},
		{QuestionTypeConversational, "Conversational"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.qType.String()
			if result != tt.expected {
				t.Errorf("QuestionType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSuggestTemperature(t *testing.T) {
	tests := []struct {
		qType       QuestionType
		expectedMin float32
		expectedMax float32
	}{
		{QuestionTypeExtraction, 0.1, 0.3},     // Very deterministic
		{QuestionTypeDirect, 0.2, 0.4},         // Mostly factual
		{QuestionTypeComparison, 0.4, 0.6},     // Balanced
		{QuestionTypeSummary, 0.5, 0.7},        // Slight creativity
		{QuestionTypeAnalytical, 0.6, 0.8},     // More creative
		{QuestionTypeConversational, 0.7, 0.9}, // Natural
	}

	for _, tt := range tests {
		t.Run(tt.qType.String(), func(t *testing.T) {
			result := SuggestTemperature(tt.qType)
			if result < tt.expectedMin || result > tt.expectedMax {
				t.Errorf("SuggestTemperature(%v) = %v, want between %v and %v",
					tt.qType, result, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestGetSystemInstructions(t *testing.T) {
	tests := []struct {
		qType QuestionType
	}{
		{QuestionTypeDirect},
		{QuestionTypeSummary},
		{QuestionTypeComparison},
		{QuestionTypeExtraction},
		{QuestionTypeAnalytical},
		{QuestionTypeConversational},
	}

	for _, tt := range tests {
		t.Run(tt.qType.String(), func(t *testing.T) {
			result := GetSystemInstructions(tt.qType)
			if result == "" {
				t.Errorf("GetSystemInstructions(%v) returned empty string", tt.qType)
			}
			if len(result) < 10 {
				t.Errorf("GetSystemInstructions(%v) returned suspiciously short string: %q", tt.qType, result)
			}
		})
	}
}
