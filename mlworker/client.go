package mlworker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"mime/multipart"
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
	ExtractImages     bool   `json:"extract_images"`      // Whether to extract images
	HybridMode        bool   `json:"hybrid_mode"`         // Use PyMuPDF fallback for better results
	PageRange         string `json:"page_range,omitempty"` // Optional: "1-10" or "5" or "1,3,5-7"
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

	// Get timeout - default 5 minutes for document processing
	timeout := 300 * time.Second
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

// ParseDocumentMultipart parses document using multipart file upload (for large files)
// This function now uses retry logic and automatic chunking for better reliability
func (c *Client) ParseDocumentMultipart(fileData []byte, filename string, fileType string, preserveStructure, extractTables bool) (*ParseDocumentResponse, error) {
	// Use default production options with retry and chunking
	options := DefaultParseOptions()
	options.PreserveStructure = preserveStructure
	options.ExtractTables = extractTables

	return c.ParseDocumentWithRetry(fileData, filename, options)
}

// ParseDocumentAdvanced parses document with full control over options
func (c *Client) ParseDocumentAdvanced(fileData []byte, filename string, options ParseOptions) (*ParseDocumentResponse, error) {
	return c.ParseDocumentWithRetry(fileData, filename, options)
}

// ParseDocumentAuto automatically determines the best parsing strategy
// based on document size and complexity
func (c *Client) ParseDocumentAuto(fileData []byte, filename string) (*ParseDocumentResponse, error) {
	if !c.IsFeatureEnabled("parse") {
		return nil, fmt.Errorf("parse feature not enabled")
	}

	// First, do a quick parse to get page count
	quickOptions := ParseOptions{
		PreserveStructure: false,
		ExtractTables:     false,
		ExtractImages:     false,
		HybridMode:        false,
		PageRange:         "1", // Just first page
		MaxRetries:        1,
	}

	quickResp, err := c.ParseDocumentWithRetry(fileData, filename, quickOptions)
	if err != nil {
		log.Printf("[MLWorker] Quick page count failed, using standard parse: %v", err)
		// Fallback to standard parse
		return c.ParseDocumentWithRetry(fileData, filename, DefaultParseOptions())
	}

	totalPages := quickResp.PageCount
	log.Printf("[MLWorker] Document has %d pages, determining best strategy", totalPages)

	// Choose strategy based on page count
	options := DefaultParseOptions()

	if totalPages > 100 {
		// Very large document - use chunking
		log.Printf("[MLWorker] Large document (%d pages), using chunked processing", totalPages)
		return c.ParseDocumentChunked(fileData, filename, totalPages, options)
	} else if totalPages > 50 {
		// Medium document - standard parse with retry
		log.Printf("[MLWorker] Medium document (%d pages), using standard parse with retry", totalPages)
		return c.ParseDocumentWithRetry(fileData, filename, options)
	} else {
		// Small document - fast parse
		log.Printf("[MLWorker] Small document (%d pages), using fast parse", totalPages)
		return c.ParseDocumentWithRetry(fileData, filename, options)
	}
}

// ParseDocumentPyMuPDF parses PDF using PyMuPDF4LLM (markdown-optimized extraction)
func (c *Client) ParseDocumentPyMuPDF(fileData []byte, filename string) (*ParseDocumentResponse, error) {
	if !c.IsFeatureEnabled("pymupdf") {
		return nil, fmt.Errorf("PyMuPDF feature not enabled")
	}

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file part
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(fileData); err != nil {
		return nil, fmt.Errorf("write file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	// Create request
	url := c.BaseURL + "/pymupdf"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML Worker error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var parseResp ParseDocumentResponse
	if err := json.NewDecoder(resp.Body).Decode(&parseResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	log.Printf("[MLWorker] PyMuPDF extraction - %d pages, %d chars (%dms)",
		parseResp.PageCount, len(parseResp.Text), parseResp.ProcessingTimeMS)
	return &parseResp, nil
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

// doRequestWithRetry performs an HTTP request with exponential backoff retry logic
func (c *Client) doRequestWithRetry(method, path string, reqBody, respBody interface{}, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := c.doRequest(method, path, reqBody, respBody)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on non-timeout errors
		if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline exceeded") {
			return err
		}

		if attempt < maxRetries-1 {
			// Exponential backoff: 2^attempt seconds (1s, 2s, 4s)
			waitTime := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			log.Printf("[MLWorker] Request timeout, retrying in %v (attempt %d/%d)", waitTime, attempt+1, maxRetries)
			time.Sleep(waitTime)
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", maxRetries, lastErr)
}

// ParseDocumentWithRetry parses document with automatic retry on timeout
func (c *Client) ParseDocumentWithRetry(fileData []byte, filename string, options ParseOptions) (*ParseDocumentResponse, error) {
	if !c.IsFeatureEnabled("parse") {
		return nil, fmt.Errorf("parse feature not enabled")
	}

	startTime := time.Now()

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file part
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(fileData); err != nil {
		return nil, fmt.Errorf("write file data: %w", err)
	}

	// Add form fields
	writer.WriteField("preserve_structure", fmt.Sprintf("%t", options.PreserveStructure))
	writer.WriteField("extract_tables", fmt.Sprintf("%t", options.ExtractTables))
	writer.WriteField("extract_images", fmt.Sprintf("%t", options.ExtractImages))
	writer.WriteField("hybrid_mode", fmt.Sprintf("%t", options.HybridMode))

	if options.PageRange != "" {
		writer.WriteField("page_range", options.PageRange)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	// Retry logic with exponential backoff
	var parseResp *ParseDocumentResponse
	var lastErr error
	maxRetries := options.MaxRetries
	if maxRetries == 0 {
		maxRetries = 2 // Default 2 retries
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Create request
		url := c.BaseURL + "/parse"
		req, err := http.NewRequest("POST", url, bytes.NewReader(body.Bytes()))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())

		// Send request
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err

			// Check if it's a timeout error
			if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
				if attempt < maxRetries-1 {
					waitTime := time.Duration(math.Pow(2, float64(attempt))) * time.Second
					log.Printf("[MLWorker] Parse timeout for %s, retrying in %v (attempt %d/%d)",
						filename, waitTime, attempt+1, maxRetries)
					time.Sleep(waitTime)
					continue
				}
			}
			return nil, fmt.Errorf("HTTP request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("ML Worker error (%d): %s", resp.StatusCode, string(bodyBytes))

			// Retry on 5xx errors
			if resp.StatusCode >= 500 && attempt < maxRetries-1 {
				waitTime := time.Duration(math.Pow(2, float64(attempt))) * time.Second
				log.Printf("[MLWorker] Server error for %s, retrying in %v (attempt %d/%d)",
					filename, waitTime, attempt+1, maxRetries)
				time.Sleep(waitTime)
				continue
			}
			return nil, lastErr
		}

		var result ParseDocumentResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}

		parseResp = &result
		break
	}

	if parseResp == nil {
		return nil, fmt.Errorf("max retries (%d) exceeded: %w", maxRetries, lastErr)
	}

	duration := time.Since(startTime)

	// Log slow documents (>2 minutes)
	if duration.Seconds() > 120 {
		log.Printf("[MLWorker] ⚠️  SLOW PARSE: %.1fs for %s (%d pages, %d chars, %d tables)",
			duration.Seconds(), filename, parseResp.PageCount, len(parseResp.Text), len(parseResp.Tables))
	} else {
		log.Printf("[MLWorker] Document parsed - %s: %d pages, %d chars, %d tables (%.1fs)",
			filename, parseResp.PageCount, len(parseResp.Text), len(parseResp.Tables), duration.Seconds())
	}

	return parseResp, nil
}

// ParseOptions contains options for document parsing
type ParseOptions struct {
	PreserveStructure bool
	ExtractTables     bool
	ExtractImages     bool
	HybridMode        bool
	PageRange         string // "1-10" or "5" or "1,3,5-7"
	MaxRetries        int    // Default 2
	ChunkSize         int    // Pages per chunk (0 = no chunking)
}

// DefaultParseOptions returns recommended production settings
func DefaultParseOptions() ParseOptions {
	return ParseOptions{
		PreserveStructure: true,
		ExtractTables:     true,
		ExtractImages:     false,
		HybridMode:        false,
		MaxRetries:        2,
		ChunkSize:         50, // Chunk documents >50 pages
	}
}

// ParseDocumentChunked parses large documents in chunks
func (c *Client) ParseDocumentChunked(fileData []byte, filename string, totalPages int, options ParseOptions) (*ParseDocumentResponse, error) {
	if options.ChunkSize == 0 || totalPages <= options.ChunkSize {
		// No chunking needed
		return c.ParseDocumentWithRetry(fileData, filename, options)
	}

	log.Printf("[MLWorker] Chunking %s (%d pages) into chunks of %d pages", filename, totalPages, options.ChunkSize)

	// Parse in chunks
	var allText string
	var allMarkdown string
	var allTables []TableData
	var totalProcessingTime int

	for start := 1; start <= totalPages; start += options.ChunkSize {
		end := start + options.ChunkSize - 1
		if end > totalPages {
			end = totalPages
		}

		chunkOptions := options
		chunkOptions.PageRange = fmt.Sprintf("%d-%d", start, end)

		log.Printf("[MLWorker] Processing chunk: pages %d-%d of %s", start, end, filename)

		chunkResp, err := c.ParseDocumentWithRetry(fileData, filename, chunkOptions)
		if err != nil {
			return nil, fmt.Errorf("chunk %d-%d failed: %w", start, end, err)
		}

		// Merge results
		allText += chunkResp.Text + "\n\n"
		allMarkdown += chunkResp.Markdown + "\n\n"
		allTables = append(allTables, chunkResp.Tables...)
		totalProcessingTime += chunkResp.ProcessingTimeMS
	}

	// Return merged response
	return &ParseDocumentResponse{
		Text:             allText,
		Markdown:         allMarkdown,
		Tables:           allTables,
		PageCount:        totalPages,
		ProcessingTimeMS: totalProcessingTime,
		Metadata: map[string]interface{}{
			"chunked":    true,
			"chunk_size": options.ChunkSize,
		},
	}, nil
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
