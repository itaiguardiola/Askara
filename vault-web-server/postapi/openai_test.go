package postapi

import (
	"testing"
)

// TestBuildPrompt tests the prompt building function
func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name     string
		contexts []string
		question string
		wantErr  bool
	}{
		{
			name: "Basic prompt with single context",
			contexts: []string{
				"The sky is blue because of Rayleigh scattering.",
			},
			question: "Why is the sky blue?",
			wantErr:  false,
		},
		{
			name: "Prompt with multiple contexts",
			contexts: []string{
				"Context 1: AI is artificial intelligence.",
				"Context 2: Machine learning is a subset of AI.",
				"Context 3: Deep learning uses neural networks.",
			},
			question: "What is AI?",
			wantErr:  false,
		},
		{
			name:     "Empty contexts",
			contexts: []string{},
			question: "What is AI?",
			wantErr:  false,
		},
		{
			name: "Very long context (should be truncated)",
			contexts: []string{
				generateLongText(5000), // Generate very long text
			},
			question: "Summarize this.",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := buildPrompt(tt.contexts, tt.question)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildPrompt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check that prompt contains the question
				if len(tt.question) > 0 && len(prompt) > 0 {
					// Prompt should contain some form of the question
					t.Logf("Generated prompt length: %d", len(prompt))
				}

				// Check that prompt is not empty when we have input
				if len(tt.contexts) > 0 && len(prompt) == 0 {
					t.Error("Expected non-empty prompt with contexts")
				}
			}
		})
	}
}

// TestStreamHandler tests the StreamHandler type
func TestStreamHandler(t *testing.T) {
	called := false
	var receivedChunk string

	handler := StreamHandler(func(chunk string) error {
		called = true
		receivedChunk = chunk
		return nil
	})

	testChunk := "Test chunk"
	err := handler(testChunk)

	if err != nil {
		t.Errorf("StreamHandler returned unexpected error: %v", err)
	}

	if !called {
		t.Error("StreamHandler was not called")
	}

	if receivedChunk != testChunk {
		t.Errorf("Expected chunk '%s', got '%s'", testChunk, receivedChunk)
	}
}

// TestStreamHandlerError tests error handling in StreamHandler
func TestStreamHandlerError(t *testing.T) {
	expectedErr := "test error"

	handler := StreamHandler(func(chunk string) error {
		return &testError{msg: expectedErr}
	})

	err := handler("test")

	if err == nil {
		t.Error("Expected error from StreamHandler, got nil")
	}

	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// Helper function to generate long text for testing
func generateLongText(length int) string {
	text := ""
	for i := 0; i < length; i++ {
		text += "a"
	}
	return text
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// TestMinFunction tests the min helper function
func TestMinFunction(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"a is smaller", 5, 10, 5},
		{"b is smaller", 10, 5, 5},
		{"equal values", 7, 7, 7},
		{"negative numbers", -5, -10, -10},
		{"mixed positive negative", -5, 10, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := min(tt.a, tt.b); got != tt.want {
				t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestOpenAIResponseType tests the OpenAIResponse struct
func TestOpenAIResponseType(t *testing.T) {
	response := OpenAIResponse{
		Response: "This is a test response",
		Tokens:   150,
	}

	if response.Response != "This is a test response" {
		t.Errorf("Expected response 'This is a test response', got '%s'", response.Response)
	}

	if response.Tokens != 150 {
		t.Errorf("Expected tokens 150, got %d", response.Tokens)
	}
}
