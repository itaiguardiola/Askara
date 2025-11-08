package postapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/itaiguardiola/askara/validator"
)

// CodeTrustStatsResponse represents the response for code source of truth statistics
type CodeTrustStatsResponse struct {
	Stats map[string]interface{} `json:"stats"`
}

// CodeTrustSymbolsResponse represents the response for symbol searches
type CodeTrustSymbolsResponse struct {
	Symbols []interface{} `json:"symbols"`
}

// CodeTrustAPIsResponse represents the response for API endpoint searches
type CodeTrustAPIsResponse struct {
	APIs []interface{} `json:"apis"`
}

// CodeTrustGroundingResponse represents the response for query grounding
type CodeTrustGroundingResponse struct {
	Grounding interface{} `json:"grounding"`
}

// CodeTrustIndexResponse represents the response for the full index
type CodeTrustIndexResponse struct {
	Index interface{} `json:"index"`
}

// GetCodeTrustStatsHandler handles GET /api/codetrust/stats
func (ctx *HandlerContext) GetCodeTrustStatsHandler(w http.ResponseWriter, r *http.Request) {
	if ctx.codeTrustSvc == nil {
		http.Error(w, "Code Source Trust is not enabled", http.StatusServiceUnavailable)
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		log.Println("[GetCodeTrustStatsHandler ERR] Missing UUID parameter")
		http.Error(w, "Missing UUID parameter", http.StatusBadRequest)
		return
	}

	// Validate UUID format
	if err := validator.ValidateUUID(uuid); err != nil {
		log.Printf("[GetCodeTrustStatsHandler ERR] Invalid UUID: %v", err)
		http.Error(w, fmt.Sprintf("Invalid UUID format: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("[GetCodeTrustStatsHandler] Fetching code trust stats for UUID=%s", uuid)

	stats, err := ctx.codeTrustSvc.GetStats(uuid)
	if err != nil {
		log.Printf("[GetCodeTrustStatsHandler ERR] Failed to get stats: %v", err)
		http.Error(w, fmt.Sprintf("Failed to retrieve code trust stats: %v", err), http.StatusInternalServerError)
		return
	}

	response := CodeTrustStatsResponse{
		Stats: stats,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[GetCodeTrustStatsHandler ERR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("[GetCodeTrustStatsHandler] Successfully returned stats")
}

// SearchSymbolsHandler handles GET /api/codetrust/symbols?uuid=xxx&query=xxx
func (ctx *HandlerContext) SearchSymbolsHandler(w http.ResponseWriter, r *http.Request) {
	if ctx.codeTrustSvc == nil {
		http.Error(w, "Code Source Trust is not enabled", http.StatusServiceUnavailable)
		return
	}

	uuid := r.URL.Query().Get("uuid")
	query := r.URL.Query().Get("query")

	if uuid == "" {
		log.Println("[SearchSymbolsHandler ERR] Missing UUID parameter")
		http.Error(w, "Missing UUID parameter", http.StatusBadRequest)
		return
	}

	if query == "" {
		log.Println("[SearchSymbolsHandler ERR] Missing query parameter")
		http.Error(w, "Missing query parameter", http.StatusBadRequest)
		return
	}

	// Validate UUID format
	if err := validator.ValidateUUID(uuid); err != nil {
		log.Printf("[SearchSymbolsHandler ERR] Invalid UUID: %v", err)
		http.Error(w, fmt.Sprintf("Invalid UUID format: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("[SearchSymbolsHandler] Searching for symbol '%s' for UUID=%s", query, uuid)

	symbols, err := ctx.codeTrustSvc.SearchSymbols(uuid, query)
	if err != nil {
		log.Printf("[SearchSymbolsHandler ERR] Failed to search symbols: %v", err)
		http.Error(w, fmt.Sprintf("Failed to search symbols: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to interface{} slice for JSON encoding
	symbolsInterface := make([]interface{}, len(symbols))
	for i, s := range symbols {
		symbolsInterface[i] = s
	}

	response := CodeTrustSymbolsResponse{
		Symbols: symbolsInterface,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[SearchSymbolsHandler ERR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("[SearchSymbolsHandler] Found %d symbols", len(symbols))
}

// SearchAPIsHandler handles GET /api/codetrust/apis?uuid=xxx&query=xxx
func (ctx *HandlerContext) SearchAPIsHandler(w http.ResponseWriter, r *http.Request) {
	if ctx.codeTrustSvc == nil {
		http.Error(w, "Code Source Trust is not enabled", http.StatusServiceUnavailable)
		return
	}

	uuid := r.URL.Query().Get("uuid")
	query := r.URL.Query().Get("query")

	if uuid == "" {
		log.Println("[SearchAPIsHandler ERR] Missing UUID parameter")
		http.Error(w, "Missing UUID parameter", http.StatusBadRequest)
		return
	}

	if query == "" {
		log.Println("[SearchAPIsHandler ERR] Missing query parameter")
		http.Error(w, "Missing query parameter", http.StatusBadRequest)
		return
	}

	// Validate UUID format
	if err := validator.ValidateUUID(uuid); err != nil {
		log.Printf("[SearchAPIsHandler ERR] Invalid UUID: %v", err)
		http.Error(w, fmt.Sprintf("Invalid UUID format: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("[SearchAPIsHandler] Searching for API '%s' for UUID=%s", query, uuid)

	apis, err := ctx.codeTrustSvc.SearchAPIs(uuid, query)
	if err != nil {
		log.Printf("[SearchAPIsHandler ERR] Failed to search APIs: %v", err)
		http.Error(w, fmt.Sprintf("Failed to search APIs: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to interface{} slice for JSON encoding
	apisInterface := make([]interface{}, len(apis))
	for i, a := range apis {
		apisInterface[i] = a
	}

	response := CodeTrustAPIsResponse{
		APIs: apisInterface,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[SearchAPIsHandler ERR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("[SearchAPIsHandler] Found %d APIs", len(apis))
}

// GroundQueryHandler handles POST /api/codetrust/ground with body {uuid, query}
func (ctx *HandlerContext) GroundQueryHandler(w http.ResponseWriter, r *http.Request) {
	if ctx.codeTrustSvc == nil {
		http.Error(w, "Code Source Trust is not enabled", http.StatusServiceUnavailable)
		return
	}

	var requestBody struct {
		UUID  string `json:"uuid"`
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		log.Printf("[GroundQueryHandler ERR] Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if requestBody.UUID == "" {
		log.Println("[GroundQueryHandler ERR] Missing UUID")
		http.Error(w, "Missing UUID", http.StatusBadRequest)
		return
	}

	if requestBody.Query == "" {
		log.Println("[GroundQueryHandler ERR] Missing query")
		http.Error(w, "Missing query", http.StatusBadRequest)
		return
	}

	// Validate UUID format
	if err := validator.ValidateUUID(requestBody.UUID); err != nil {
		log.Printf("[GroundQueryHandler ERR] Invalid UUID: %v", err)
		http.Error(w, fmt.Sprintf("Invalid UUID format: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("[GroundQueryHandler] Grounding query '%s' for UUID=%s", requestBody.Query, requestBody.UUID)

	grounding, err := ctx.codeTrustSvc.GroundQuery(requestBody.UUID, requestBody.Query)
	if err != nil {
		log.Printf("[GroundQueryHandler ERR] Failed to ground query: %v", err)
		http.Error(w, fmt.Sprintf("Failed to ground query: %v", err), http.StatusInternalServerError)
		return
	}

	response := CodeTrustGroundingResponse{
		Grounding: grounding,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[GroundQueryHandler ERR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("[GroundQueryHandler] Grounding score: %.2f", grounding.GroundingScore)
}

// GetCodeTrustIndexHandler handles GET /api/codetrust/index?uuid=xxx
func (ctx *HandlerContext) GetCodeTrustIndexHandler(w http.ResponseWriter, r *http.Request) {
	if ctx.codeTrustSvc == nil {
		http.Error(w, "Code Source Trust is not enabled", http.StatusServiceUnavailable)
		return
	}

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		log.Println("[GetCodeTrustIndexHandler ERR] Missing UUID parameter")
		http.Error(w, "Missing UUID parameter", http.StatusBadRequest)
		return
	}

	// Validate UUID format
	if err := validator.ValidateUUID(uuid); err != nil {
		log.Printf("[GetCodeTrustIndexHandler ERR] Invalid UUID: %v", err)
		http.Error(w, fmt.Sprintf("Invalid UUID format: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("[GetCodeTrustIndexHandler] Fetching code trust index for UUID=%s", uuid)

	index, err := ctx.codeTrustSvc.GetUserIndex(uuid)
	if err != nil {
		log.Printf("[GetCodeTrustIndexHandler ERR] Failed to get index: %v", err)
		http.Error(w, fmt.Sprintf("Failed to retrieve code trust index: %v", err), http.StatusInternalServerError)
		return
	}

	response := CodeTrustIndexResponse{
		Index: index,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[GetCodeTrustIndexHandler ERR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("[GetCodeTrustIndexHandler] Successfully returned index")
}

// GetDocumentEnrichmentHandler handles GET /api/codetrust/enrichment/{documentId}?uuid=xxx
func (ctx *HandlerContext) GetDocumentEnrichmentHandler(w http.ResponseWriter, r *http.Request) {
	if ctx.codeTrustSvc == nil {
		http.Error(w, "Code Source Trust is not enabled", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	docID := vars["documentId"]

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		log.Println("[GetDocumentEnrichmentHandler ERR] Missing UUID parameter")
		http.Error(w, "Missing UUID parameter", http.StatusBadRequest)
		return
	}

	// Validate UUID format
	if err := validator.ValidateUUID(uuid); err != nil {
		log.Printf("[GetDocumentEnrichmentHandler ERR] Invalid UUID: %v", err)
		http.Error(w, fmt.Sprintf("Invalid UUID format: %v", err), http.StatusBadRequest)
		return
	}

	if docID == "" {
		log.Println("[GetDocumentEnrichmentHandler ERR] Missing document ID")
		http.Error(w, "Missing document ID", http.StatusBadRequest)
		return
	}

	// Validate document ID format
	if err := validator.ValidateDocumentID(docID); err != nil {
		log.Printf("[GetDocumentEnrichmentHandler ERR] Invalid document ID: %v", err)
		http.Error(w, fmt.Sprintf("Invalid document ID format: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("[GetDocumentEnrichmentHandler] Fetching enrichment for document ID=%s, UUID=%s", docID, uuid)

	enrichment, err := ctx.codeTrustSvc.GetDocumentEnrichment(uuid, docID)
	if err != nil {
		log.Printf("[GetDocumentEnrichmentHandler ERR] Failed to get enrichment: %v", err)
		http.Error(w, fmt.Sprintf("Failed to retrieve document enrichment: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(enrichment); err != nil {
		log.Printf("[GetDocumentEnrichmentHandler ERR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("[GetDocumentEnrichmentHandler] Successfully returned enrichment")
}
