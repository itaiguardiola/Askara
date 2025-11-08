package codesourcetrust

import (
	"fmt"
	"log"
	"sync"
)

// Service provides high-level code source of truth operations
type Service struct {
	parser  *CodeParser
	storage *Storage
	config  ExtractionConfig
	mu      sync.RWMutex
}

// NewService creates a new code source of truth service
func NewService(storageDir string, config ExtractionConfig) (*Service, error) {
	storage, err := NewStorage(storageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	parser := NewCodeParser(config)

	return &Service{
		parser:  parser,
		storage: storage,
		config:  config,
	}, nil
}

// ProcessDocument analyzes a document and extracts code elements
func (s *Service) ProcessDocument(documentID, documentName, uuid, text string) (*DocumentEnrichment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[CodeSourceTrust] Processing document: %s (ID: %s, User: %s)", documentName, documentID, uuid)

	// Parse the document
	enrichment, err := s.parser.ParseDocument(documentID, documentName, uuid, text)
	if err != nil {
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}

	// Save enrichment data
	if err := s.storage.SaveEnrichment(enrichment); err != nil {
		return nil, fmt.Errorf("failed to save enrichment: %w", err)
	}

	log.Printf("[CodeSourceTrust] Processed document %s: %d symbols, %d code blocks, %d APIs, %d concepts",
		documentName,
		len(enrichment.Symbols),
		len(enrichment.CodeBlocks),
		len(enrichment.APIEndpoints),
		len(enrichment.Concepts))

	return enrichment, nil
}

// GetDocumentEnrichment retrieves enrichment data for a document
func (s *Service) GetDocumentEnrichment(uuid, documentID string) (*DocumentEnrichment, error) {
	return s.storage.GetEnrichment(uuid, documentID)
}

// DeleteDocument removes enrichment data for a document
func (s *Service) DeleteDocument(uuid, documentID string) error {
	log.Printf("[CodeSourceTrust] Deleting enrichment for document: %s (User: %s)", documentID, uuid)
	return s.storage.DeleteEnrichment(uuid, documentID)
}

// SearchSymbols finds symbols by name
func (s *Service) SearchSymbols(uuid, symbolName string) ([]CodeSymbol, error) {
	return s.storage.QuerySymbols(uuid, symbolName)
}

// SearchAPIs finds API endpoints by path
func (s *Service) SearchAPIs(uuid, path string) ([]APIEndpoint, error) {
	return s.storage.QueryAPIs(uuid, path)
}

// SearchConcepts finds concepts by name
func (s *Service) SearchConcepts(uuid, conceptName string) (*Concept, error) {
	return s.storage.QueryConcepts(uuid, conceptName)
}

// GroundQuery analyzes a query and returns relevant code elements
func (s *Service) GroundQuery(uuid, query string) (*QueryGroundingResult, error) {
	return s.storage.GroundQuery(uuid, query)
}

// GetUserIndex returns the complete code source of truth index for a user
func (s *Service) GetUserIndex(uuid string) (*SourceOfTruthIndex, error) {
	return s.storage.GetIndex(uuid)
}

// GetStats returns statistics about the code source of truth
func (s *Service) GetStats(uuid string) (map[string]interface{}, error) {
	return s.storage.GetStats(uuid)
}

// EnhanceContext enriches retrieval context with code metadata
// This is called by the question handler to improve answers
func (s *Service) EnhanceContext(uuid, query string, originalContexts []map[string]interface{}) ([]map[string]interface{}, error) {
	// Ground the query to understand what code elements are being asked about
	grounding, err := s.GroundQuery(uuid, query)
	if err != nil {
		// If grounding fails, return original contexts
		log.Printf("[CodeSourceTrust] Failed to ground query: %v", err)
		return originalContexts, nil
	}

	// If no code elements detected, return original contexts
	if grounding.GroundingScore == 0 {
		return originalContexts, nil
	}

	log.Printf("[CodeSourceTrust] Query grounding - Score: %.2f, Symbols: %d, APIs: %d, Concepts: %d",
		grounding.GroundingScore,
		len(grounding.DetectedSymbols),
		len(grounding.DetectedAPIs),
		len(grounding.DetectedConcepts))

	// Enhance each context with code metadata
	enhanced := make([]map[string]interface{}, len(originalContexts))
	copy(enhanced, originalContexts)

	// Add grounding metadata to help the LLM
	groundingInfo := map[string]interface{}{
		"grounding_score":   grounding.GroundingScore,
		"detected_symbols":  grounding.DetectedSymbols,
		"detected_apis":     grounding.DetectedAPIs,
		"detected_concepts": grounding.DetectedConcepts,
		"recommended_docs":  grounding.RecommendedDocs,
	}

	// Add symbol details if found
	if len(grounding.MatchedSymbols) > 0 {
		symbolDetails := make([]map[string]interface{}, len(grounding.MatchedSymbols))
		for i, symbol := range grounding.MatchedSymbols {
			symbolDetails[i] = map[string]interface{}{
				"symbol":        symbol.Symbol,
				"type":          symbol.Type,
				"language":      symbol.Language,
				"signature":     symbol.Signature,
				"document_name": symbol.DocumentName,
			}
		}
		groundingInfo["matched_symbols"] = symbolDetails
	}

	// Add API details if found
	if len(grounding.MatchedAPIs) > 0 {
		apiDetails := make([]map[string]interface{}, len(grounding.MatchedAPIs))
		for i, api := range grounding.MatchedAPIs {
			apiDetails[i] = map[string]interface{}{
				"method":        api.Method,
				"path":          api.Path,
				"description":   api.Description,
				"document_name": api.DocumentName,
			}
		}
		groundingInfo["matched_apis"] = apiDetails
	}

	// Prepend grounding information as a special context
	groundingContext := map[string]interface{}{
		"text":     fmt.Sprintf("Code Source of Truth - Grounding Info: %v", groundingInfo),
		"title":    "[Code Source Metadata]",
		"metadata": groundingInfo,
	}

	result := append([]map[string]interface{}{groundingContext}, enhanced...)

	return result, nil
}

// UpdateConfig updates the extraction configuration
func (s *Service) UpdateConfig(config ExtractionConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config
	s.parser = NewCodeParser(config)
	log.Printf("[CodeSourceTrust] Configuration updated")
}

// GetConfig returns the current extraction configuration
func (s *Service) GetConfig() ExtractionConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.config
}
