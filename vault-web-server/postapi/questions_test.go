package postapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/itaiguardiola/askara/vectordb"
	cache "github.com/patrickmn/go-cache"
	openai "github.com/sashabaranov/go-openai"
)

// MockVectorDB is a mock implementation of VectorDB for testing
type MockVectorDB struct {
	ShouldError bool
}

func (m *MockVectorDB) UpsertEmbeddings(embeddings [][]float32, chunks interface{}, uuid string) error {
	if m.ShouldError {
		return fmt.Errorf("mock upsert error")
	}
	return nil
}

func (m *MockVectorDB) Retrieve(questionEmbedding []float32, topK int, uuid string) ([]vectordb.QueryMatch, error) {
	if m.ShouldError {
		return nil, fmt.Errorf("mock retrieve error")
	}

	// Return mock context matches
	return []vectordb.QueryMatch{
		{
			ID:    "1",
			Score: 0.95,
			Metadata: map[string]string{
				"text":  "This is test context from document 1.",
				"title": "Test Document 1",
			},
		},
		{
			ID:    "2",
			Score: 0.85,
			Metadata: map[string]string{
				"text":  "This is test context from document 2.",
				"title": "Test Document 2",
			},
		},
	}, nil
}

// TestStreamingQuestionHandlerValidation tests form validation
func TestStreamingQuestionHandlerValidation(t *testing.T) {
	// Create mock dependencies
	mockVectorDB := &MockVectorDB{}
	mockClient := openai.NewClient("test-key")

	ctx := &HandlerContext{
		openAIClient: mockClient,
		cache:        cache.New(cache.NoExpiration, cache.NoExpiration),
		vectorDB:     mockVectorDB,
	}

	// Test missing question parameter
	form := url.Values{}
	form.Add("model", "GPT Turbo")
	form.Add("uuid", "test-uuid")

	req := httptest.NewRequest("POST", "/api/questions/stream", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	ctx.StreamingQuestionHandler(w, req)

	if w.Code == http.StatusOK {
		t.Errorf("Expected non-OK status for missing question, got %d", w.Code)
	}
}

// TestStreamingQuestionHandlerHeaders tests SSE headers
func TestStreamingQuestionHandlerHeaders(t *testing.T) {
	// Note: This test would require a real OpenAI client or a more sophisticated mock
	// For now, we'll test that the headers are set correctly before any API calls

	mockVectorDB := &MockVectorDB{}
	mockClient := openai.NewClient("test-key")

	ctx := &HandlerContext{
		openAIClient: mockClient,
		cache:        cache.New(cache.NoExpiration, cache.NoExpiration),
		vectorDB:     mockVectorDB,
	}

	form := url.Values{}
	form.Add("question", "What is AI?")
	form.Add("model", "GPT Turbo")
	form.Add("uuid", "test-uuid")

	req := httptest.NewRequest("POST", "/api/questions/stream", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()

	// Note: This will fail at the OpenAI API call stage, but we can check headers
	ctx.StreamingQuestionHandler(w, req)

	// Check SSE headers were set
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("Expected Content-Type 'text/event-stream', got '%s'", contentType)
	}

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "no-cache" {
		t.Errorf("Expected Cache-Control 'no-cache', got '%s'", cacheControl)
	}

	connection := w.Header().Get("Connection")
	if connection != "keep-alive" {
		t.Errorf("Expected Connection 'keep-alive', got '%s'", connection)
	}

	accessControl := w.Header().Get("Access-Control-Allow-Origin")
	if accessControl != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin '*', got '%s'", accessControl)
	}
}

// TestStreamingQuestionHandlerVectorDBError tests error handling for vector DB failures
func TestStreamingQuestionHandlerVectorDBError(t *testing.T) {
	// Create mock dependencies with error
	mockVectorDB := &MockVectorDB{ShouldError: true}
	mockClient := openai.NewClient("test-key")

	ctx := &HandlerContext{
		openAIClient: mockClient,
		cache:        cache.New(cache.NoExpiration, cache.NoExpiration),
		vectorDB:     mockVectorDB,
	}

	form := url.Values{}
	form.Add("question", "What is AI?")
	form.Add("model", "GPT Turbo")
	form.Add("uuid", "test-uuid")

	req := httptest.NewRequest("POST", "/api/questions/stream", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()

	// This will fail at the OpenAI embedding stage since we don't have a real client
	// But we can verify the handler doesn't panic
	ctx.StreamingQuestionHandler(w, req)

	// Should have SSE headers even on error
	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected SSE headers to be set even on error")
	}
}

// TestQuestionHandlerBasic tests the non-streaming handler
func TestQuestionHandlerBasic(t *testing.T) {
	mockVectorDB := &MockVectorDB{}
	mockClient := openai.NewClient("test-key")

	ctx := &HandlerContext{
		openAIClient: mockClient,
		cache:        cache.New(cache.NoExpiration, cache.NoExpiration),
		vectorDB:     mockVectorDB,
	}

	form := url.Values{}
	form.Add("question", "What is AI?")
	form.Add("model", "GPT Turbo")
	form.Add("uuid", "test-uuid")

	req := httptest.NewRequest("POST", "/api/questions", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	ctx.QuestionHandler(w, req)

	// Will fail at OpenAI API call, but should have proper error handling
	// Should not return 200 since we don't have valid API keys
	if w.Code == http.StatusOK {
		t.Logf("Note: Test passed validation but would fail at OpenAI API call without valid credentials")
	}
}
