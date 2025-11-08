package mlworker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Client handles ML Worker API interactions
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Enabled    bool
	Features   map[string]bool
}

// OCRRequest represents request to ML Worker OCR endpoint
type OCRRequest struct {
	ImageURL string `json:"image_url"`
	Language string `json:"language"`
	Enhance  bool   `json:"enhance"`
}

// OCRResponse represents response from ML Worker OCR
type OCRResponse struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Language   string  `json:"language"`
}

// EnhanceRequest represents request to ML Worker enhance endpoint
type EnhanceRequest struct {
	ImageURL   string   `json:"image_url"`
	Operations []string `json:"operations"`
}

// EnhanceResponse represents response from ML Worker enhance
type EnhanceResponse struct {
	EnhancedImageBase64 string                 `json:"enhanced_image_base64"`
	Improvements        map[string]interface{} `json:"improvements"`
	OriginalSize        []int                  `json:"original_size"`
	EnhancedSize        []int                  `json:"enhanced_size"`
}

// CaptionRequest represents request to ML Worker caption endpoint
type CaptionRequest struct {
	ImageURL    string `json:"image_url"`
	DetailLevel string `json:"detail_level"`
}

// CaptionResponse represents response from ML Worker caption
type CaptionResponse struct {
	Caption         string   `json:"caption"`
	ObjectsDetected []string `json:"objects_detected"`
	SceneType       *string  `json:"scene_type"`
	Confidence      float64  `json:"confidence"`
}

// ClassifyDocumentRequest represents request to classify document
type ClassifyDocumentRequest struct {
	ImageURL      string   `json:"image_url"`
	PossibleTypes []string `json:"possible_types,omitempty"`
}

// ClassifyDocumentResponse represents response from document classification
type ClassifyDocumentResponse struct {
	DocumentType     string   `json:"document_type"`
	Confidence       float64  `json:"confidence"`
	DetectedFeatures []string `json:"detected_features"`
	SuggestedTags    []string `json:"suggested_tags"`
}

// ParseDocumentRequest represents request to ML Worker document parsing endpoint (Docling)
type ParseDocumentRequest struct {
	DocumentURL       string `json:"document_url"`        // Base64 data URL or file path
	PreserveStructure bool   `json:"preserve_structure"`  // Whether to preserve document structure
	ExtractTables     bool   `json:"extract_tables"`      // Whether to extract tables separately
	Format            string `json:"format,omitempty"`    // Optional: pdf, docx, pptx, html
}

// TableData represents extracted table data
type TableData struct {
	Headers  []string              `json:"headers"`
	Rows     [][]string            `json:"rows"`
	Caption  string                `json:"caption,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ParseDocumentResponse represents response from ML Worker document parsing
type ParseDocumentResponse struct {
	Text             string                 `json:"text"`              // Extracted text content (plain text)
	Markdown         string                 `json:"markdown"`          // Extracted content in markdown format
	Structure        map[string]interface{} `json:"structure"`         // Document structure (headings, sections)
	Tables           []TableData            `json:"tables"`            // Extracted tables
	Metadata         map[string]interface{} `json:"metadata"`          // Document metadata
	PageCount        int                    `json:"page_count"`        // Total number of pages
	ProcessingTimeMS int                    `json:"processing_time_ms"` // Processing time in milliseconds
}

// RerankDocument represents a document to rerank
type RerankDocument struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// RerankRequest represents request to ML Worker rerank endpoint
type RerankRequest struct {
	Query     string           `json:"query"`
	Documents []RerankDocument `json:"documents"`
	TopK      int              `json:"top_k"`
	Model     string           `json:"model,omitempty"`
}

// RerankResult represents a reranked document result
type RerankResult struct {
	ID             string  `json:"id"`
	RelevanceScore float64 `json:"relevance_score"`
	Rank           int     `json:"rank"`
}

// RerankResponse represents response from ML Worker rerank
type RerankResponse struct {
	Results          []RerankResult `json:"results"`
	ModelUsed        string         `json:"model_used"`
	ProcessingTimeMS float64        `json:"processing_time_ms,omitempty"`
}

// NewClient creates a new ML Worker client
func NewClient() *Client {
	enabled := strings.ToLower(os.Getenv("ML_WORKER_ENABLED")) == "true"
	endpoint := os.Getenv("ML_WORKER_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://192.168.0.250:6161/askara"
	}

	// Parse enabled features
	featuresStr := os.Getenv("ML_WORKER_FEATURES")
	if featuresStr == "" {
		featuresStr = "ocr,enhance,caption"
	}
	features := make(map[string]bool)
	for _, feature := range strings.Split(featuresStr, ",") {
		features[strings.TrimSpace(feature)] = true
	}

	// Get timeout
	timeout := 30 * time.Second
	if timeoutStr := os.Getenv("ML_WORKER_TIMEOUT"); timeoutStr != "" {
		if t, err := time.ParseDuration(timeoutStr + "s"); err == nil {
			timeout = t
		}
	}

	client := &Client{
		BaseURL: endpoint,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Enabled:  enabled,
		Features: features,
	}

	if enabled {
		log.Printf("[MLWorker] Initialized - Endpoint: %s, Features: %v", endpoint, features)
	} else {
		log.Println("[MLWorker] Disabled - using local processing")
	}

	return client
}

// IsFeatureEnabled checks if a specific feature is enabled
func (c *Client) IsFeatureEnabled(feature string) bool {
	return c.Enabled && c.Features[feature]
}

// OCR performs OCR on an image
func (c *Client) OCR(imageURL string, language string, enhance bool) (*OCRResponse, error) {
	if !c.IsFeatureEnabled("ocr") {
		return nil, fmt.Errorf("OCR feature not enabled")
	}

	req := OCRRequest{
		ImageURL: imageURL,
		Language: language,
		Enhance:  enhance,
	}

	var resp OCRResponse
	err := c.doRequest("POST", "/ocr", req, &resp)
	if err != nil {
		return nil, err
	}

	log.Printf("[MLWorker] OCR completed - %d chars, %.2f%% confidence", len(resp.Text), resp.Confidence*100)
	return &resp, nil
}

// Enhance enhances an image
func (c *Client) Enhance(imageURL string, operations []string) (*EnhanceResponse, error) {
	if !c.IsFeatureEnabled("enhance") {
		return nil, fmt.Errorf("enhance feature not enabled")
	}

	req := EnhanceRequest{
		ImageURL:   imageURL,
		Operations: operations,
	}

	var resp EnhanceResponse
	err := c.doRequest("POST", "/enhance", req, &resp)
	if err != nil {
		return nil, err
	}

	log.Printf("[MLWorker] Enhancement completed - Applied: %v", operations)
	return &resp, nil
}

// Caption generates a caption for an image
func (c *Client) Caption(imageURL string, detailLevel string) (*CaptionResponse, error) {
	if !c.IsFeatureEnabled("caption") {
		return nil, fmt.Errorf("caption feature not enabled")
	}

	req := CaptionRequest{
		ImageURL:    imageURL,
		DetailLevel: detailLevel,
	}

	var resp CaptionResponse
	err := c.doRequest("POST", "/caption", req, &resp)
	if err != nil {
		return nil, err
	}

	log.Printf("[MLWorker] Caption generated - %d objects detected", len(resp.ObjectsDetected))
	return &resp, nil
}

// ClassifyDocument classifies a document type
func (c *Client) ClassifyDocument(imageURL string, possibleTypes []string) (*ClassifyDocumentResponse, error) {
	if !c.IsFeatureEnabled("classify") {
		return nil, fmt.Errorf("classify feature not enabled")
	}

	req := ClassifyDocumentRequest{
		ImageURL:      imageURL,
		PossibleTypes: possibleTypes,
	}

	var resp ClassifyDocumentResponse
	err := c.doRequest("POST", "/classify-document", req, &resp)
	if err != nil {
		return nil, err
	}

	log.Printf("[MLWorker] Document classified as: %s (%.2f%% confidence)", resp.DocumentType, resp.Confidence*100)
	return &resp, nil
}

// ParseDocument parses document using advanced parsing (Docling) via ML Worker
func (c *Client) ParseDocument(documentURL string, preserveStructure, extractTables bool) (*ParseDocumentResponse, error) {
	if !c.IsFeatureEnabled("parse") {
		return nil, fmt.Errorf("parse feature not enabled")
	}

	req := ParseDocumentRequest{
		DocumentURL:       documentURL,
		PreserveStructure: preserveStructure,
		ExtractTables:     extractTables,
	}

	var resp ParseDocumentResponse
	err := c.doRequest("POST", "/parse", req, &resp)
	if err != nil {
		return nil, err
	}

	log.Printf("[MLWorker] Document parsed - %d pages, %d chars, %d tables (%dms)",
		resp.PageCount, len(resp.Text), len(resp.Tables), resp.ProcessingTimeMS)
	return &resp, nil
}

// Rerank reranks documents using semantic reranking (cross-encoder)
func (c *Client) Rerank(query string, documents []RerankDocument, topK int) (*RerankResponse, error) {
	if !c.IsFeatureEnabled("rerank") {
		return nil, fmt.Errorf("rerank feature not enabled")
	}

	req := RerankRequest{
		Query:     query,
		Documents: documents,
		TopK:      topK,
		Model:     "BAAI/bge-reranker-v2-m3", // Default model
	}

	var resp RerankResponse
	err := c.doRequest("POST", "/rerank", req, &resp)
	if err != nil {
		return nil, err
	}

	log.Printf("[MLWorker] Reranked %d documents to top %d using %s (%.2fms)",
		len(documents), len(resp.Results), resp.ModelUsed, resp.ProcessingTimeMS)
	return &resp, nil
}

// TestConnection tests connectivity to ML Worker
func (c *Client) TestConnection() error {
	if !c.Enabled {
		return fmt.Errorf("ML Worker is disabled")
	}

	resp, err := c.HTTPClient.Get(c.BaseURL + "/")
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	log.Println("[MLWorker] Connection test successful")
	return nil
}

// doRequest performs an HTTP request to ML Worker
func (c *Client) doRequest(method, path string, reqBody, respBody interface{}) error {
	url := c.BaseURL + path

	var body io.Reader
	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ML Worker error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// Global client instance
var globalClient *Client

// InitGlobalClient initializes the global ML Worker client
func InitGlobalClient() {
	globalClient = NewClient()
}

// GetClient returns the global ML Worker client
func GetClient() *Client {
	if globalClient == nil {
		InitGlobalClient()
	}
	return globalClient
}
