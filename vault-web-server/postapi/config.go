package postapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// ConfigResponse represents the current configuration
type ConfigResponse struct {
	LLMProvider      string   `json:"llm_provider"`
	OpenAIModel      string   `json:"openai_model,omitempty"`
	OpenAIEmbedding  string   `json:"openai_embedding,omitempty"`
	OllamaHost       string   `json:"ollama_host,omitempty"`
	OllamaModel      string   `json:"ollama_model,omitempty"`
	OllamaEmbedding  string   `json:"ollama_embedding,omitempty"`
	VectorDB         string   `json:"vector_db"`
	QdrantEndpoint   string   `json:"qdrant_endpoint,omitempty"`
	PineconeEndpoint string   `json:"pinecone_endpoint,omitempty"`
	Port             string   `json:"port"`
}

// OllamaModel represents a model available in Ollama
type OllamaModel struct {
	Name         string      `json:"name"`
	Model        string      `json:"model"`
	ModifiedAt   string      `json:"modified_at"`
	Size         int64       `json:"size"`
	Digest       string      `json:"digest"`
	Details      ModelDetails `json:"details"`
}

type ModelDetails struct {
	Format           string `json:"format"`
	Family           string `json:"family"`
	ParameterSize    string `json:"parameter_size"`
	QuantizationLevel string `json:"quantization_level"`
}

// OllamaModelsResponse represents the response from Ollama's /api/tags endpoint
type OllamaModelsResponse struct {
	Models []OllamaModel `json:"models"`
}

// ConnectionTestRequest represents a request to test a connection
type ConnectionTestRequest struct {
	Type     string `json:"type"` // "ollama", "openai", "qdrant", "pinecone"
	Endpoint string `json:"endpoint,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
	Model    string `json:"model,omitempty"`
}

// ConnectionTestResponse represents the result of a connection test
type ConnectionTestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// GetConfigHandler returns the current configuration
func (ctx *HandlerContext) GetConfigHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[GetConfigHandler] Fetching current configuration")

	config := ConfigResponse{
		LLMProvider:      getEnv("LLM_PROVIDER", "ollama"),
		OpenAIModel:      getEnv("OPENAI_MODEL", ""),
		OpenAIEmbedding:  getEnv("OPENAI_EMBEDDING_MODEL", ""),
		OllamaHost:       getEnv("OLLAMA_HOST", ""),
		OllamaModel:      getEnv("OLLAMA_MODEL", ""),
		OllamaEmbedding:  getEnv("OLLAMA_EMBEDDING_MODEL", ""),
		Port:             getEnv("PORT", "8100"),
	}

	// Determine vector DB
	if os.Getenv("QDRANT_API_ENDPOINT") != "" {
		config.VectorDB = "qdrant"
		config.QdrantEndpoint = os.Getenv("QDRANT_API_ENDPOINT")
	} else if os.Getenv("PINECONE_API_ENDPOINT") != "" {
		config.VectorDB = "pinecone"
		config.PineconeEndpoint = os.Getenv("PINECONE_API_ENDPOINT")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
	log.Println("[GetConfigHandler] Configuration returned successfully")
}

// ListOllamaModelsHandler lists available models from Ollama
func (ctx *HandlerContext) ListOllamaModelsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[ListOllamaModelsHandler] Fetching Ollama models")

	ollamaHost := getEnv("OLLAMA_HOST", "http://host.docker.internal:11434")
	url := fmt.Sprintf("%s/api/tags", ollamaHost)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[ListOllamaModelsHandler ERR] Failed to connect to Ollama: %v", err)
		http.Error(w, fmt.Sprintf("Failed to connect to Ollama: %v", err), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[ListOllamaModelsHandler ERR] Ollama returned status %d: %s", resp.StatusCode, string(body))
		http.Error(w, fmt.Sprintf("Ollama API error: %s", string(body)), resp.StatusCode)
		return
	}

	var modelsResp OllamaModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		log.Printf("[ListOllamaModelsHandler ERR] Failed to decode response: %v", err)
		http.Error(w, fmt.Sprintf("Failed to decode Ollama response: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[ListOllamaModelsHandler] Successfully returned %d models", len(modelsResp.Models))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(modelsResp)
}

// TestConnectionHandler tests a connection to a service
func (ctx *HandlerContext) TestConnectionHandler(w http.ResponseWriter, r *http.Request) {
	var req ConnectionTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[TestConnectionHandler] Testing connection to: %s", req.Type)

	var result ConnectionTestResponse

	switch req.Type {
	case "ollama":
		result = testOllamaConnection(req.Endpoint)
	case "qdrant":
		result = testQdrantConnection(req.Endpoint)
	case "pinecone":
		result = testPineconeConnection(req.Endpoint, req.APIKey)
	default:
		result = ConnectionTestResponse{
			Success: false,
			Message: fmt.Sprintf("Unknown connection type: %s", req.Type),
		}
	}

	log.Printf("[TestConnectionHandler] Test result for %s: success=%v", req.Type, result.Success)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func testOllamaConnection(endpoint string) ConnectionTestResponse {
	if endpoint == "" {
		endpoint = getEnv("OLLAMA_HOST", "http://host.docker.internal:11434")
	}

	url := fmt.Sprintf("%s/api/tags", endpoint)
	resp, err := http.Get(url)
	if err != nil {
		return ConnectionTestResponse{
			Success: false,
			Message: "Failed to connect to Ollama",
			Details: err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ConnectionTestResponse{
			Success: false,
			Message: "Ollama returned an error",
			Details: string(body),
		}
	}

	var modelsResp OllamaModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return ConnectionTestResponse{
			Success: false,
			Message: "Failed to decode Ollama response",
			Details: err.Error(),
		}
	}

	return ConnectionTestResponse{
		Success: true,
		Message: fmt.Sprintf("Connected successfully. Found %d models.", len(modelsResp.Models)),
	}
}

func testQdrantConnection(endpoint string) ConnectionTestResponse {
	if endpoint == "" {
		endpoint = getEnv("QDRANT_API_ENDPOINT", "http://qdrant:6333")
	}

	url := fmt.Sprintf("%s/collections", endpoint)
	resp, err := http.Get(url)
	if err != nil {
		return ConnectionTestResponse{
			Success: false,
			Message: "Failed to connect to Qdrant",
			Details: err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ConnectionTestResponse{
			Success: false,
			Message: "Qdrant returned an error",
			Details: string(body),
		}
	}

	return ConnectionTestResponse{
		Success: true,
		Message: "Connected to Qdrant successfully",
	}
}

func testPineconeConnection(endpoint, apiKey string) ConnectionTestResponse {
	if endpoint == "" {
		endpoint = getEnv("PINECONE_API_ENDPOINT", "")
	}
	if apiKey == "" {
		apiKey = getEnv("PINECONE_API_KEY", "")
	}

	if endpoint == "" || apiKey == "" {
		return ConnectionTestResponse{
			Success: false,
			Message: "Pinecone endpoint and API key are required",
		}
	}

	req, err := http.NewRequest("GET", endpoint+"/describe_index_stats", nil)
	if err != nil {
		return ConnectionTestResponse{
			Success: false,
			Message: "Failed to create request",
			Details: err.Error(),
		}
	}

	req.Header.Set("Api-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ConnectionTestResponse{
			Success: false,
			Message: "Failed to connect to Pinecone",
			Details: err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ConnectionTestResponse{
			Success: false,
			Message: "Pinecone returned an error",
			Details: string(body),
		}
	}

	return ConnectionTestResponse{
		Success: true,
		Message: "Connected to Pinecone successfully",
	}
}
