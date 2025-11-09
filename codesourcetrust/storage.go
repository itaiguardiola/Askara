package codesourcetrust

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Storage handles persistence of code source of truth data
type Storage struct {
	baseDir string
	mu      sync.RWMutex
	cache   map[string]*SourceOfTruthIndex // UUID -> Index
}

// NewStorage creates a new storage instance
func NewStorage(baseDir string) (*Storage, error) {
	// Create base directory if it doesn't exist
	enrichmentsDir := filepath.Join(baseDir, "enrichments")
	if err := os.MkdirAll(enrichmentsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create enrichments directory: %w", err)
	}

	return &Storage{
		baseDir: baseDir,
		cache:   make(map[string]*SourceOfTruthIndex),
	}, nil
}

// SaveEnrichment stores document enrichment data
func (s *Storage) SaveEnrichment(enrichment *DocumentEnrichment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Load or create index for this user
	index, err := s.loadIndexUnlocked(enrichment.UUID)
	if err != nil {
		// Create new index if doesn't exist
		index = &SourceOfTruthIndex{
			UUID:          enrichment.UUID,
			SymbolIndex:   make(map[string][]CodeSymbol),
			APIIndex:      make(map[string][]APIEndpoint),
			ConceptIndex:  make(map[string]*Concept),
			DocumentIndex: make(map[string]*DocumentEnrichment),
			LastUpdated:   time.Now(),
		}
	}

	// Update document index
	index.DocumentIndex[enrichment.DocumentID] = enrichment

	// Update symbol index
	for _, symbol := range enrichment.Symbols {
		key := strings.ToLower(symbol.Symbol)
		index.SymbolIndex[key] = append(index.SymbolIndex[key], symbol)
	}

	// Update API index
	for _, endpoint := range enrichment.APIEndpoints {
		key := strings.ToLower(endpoint.Path)
		index.APIIndex[key] = append(index.APIIndex[key], endpoint)
	}

	// Update concept index
	for _, conceptName := range enrichment.Concepts {
		key := strings.ToLower(conceptName)
		concept, exists := index.ConceptIndex[key]
		if !exists {
			concept = &Concept{
				Name:        conceptName,
				Keywords:    []string{},
				DocumentIDs: []string{},
				Symbols:     []string{},
				Description: "",
			}
			index.ConceptIndex[key] = concept
		}

		// Add document ID if not already present
		if !contains(concept.DocumentIDs, enrichment.DocumentID) {
			concept.DocumentIDs = append(concept.DocumentIDs, enrichment.DocumentID)
		}
	}

	index.LastUpdated = time.Now()

	// Save to disk
	if err := s.saveIndexUnlocked(index); err != nil {
		return err
	}

	// Update cache
	s.cache[enrichment.UUID] = index

	return nil
}

// GetEnrichment retrieves enrichment data for a specific document
func (s *Storage) GetEnrichment(uuid, documentID string) (*DocumentEnrichment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index, err := s.loadIndexUnlocked(uuid)
	if err != nil {
		return nil, err
	}

	enrichment, exists := index.DocumentIndex[documentID]
	if !exists {
		return nil, fmt.Errorf("enrichment not found for document %s", documentID)
	}

	return enrichment, nil
}

// DeleteEnrichment removes enrichment data for a document
func (s *Storage) DeleteEnrichment(uuid, documentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, err := s.loadIndexUnlocked(uuid)
	if err != nil {
		return err
	}

	// Get enrichment to remove references
	enrichment, exists := index.DocumentIndex[documentID]
	if !exists {
		return nil // Already deleted
	}

	// Remove from document index
	delete(index.DocumentIndex, documentID)

	// Remove symbols
	for _, symbol := range enrichment.Symbols {
		key := strings.ToLower(symbol.Symbol)
		symbols := index.SymbolIndex[key]
		filtered := []CodeSymbol{}
		for _, s := range symbols {
			if s.DocumentID != documentID {
				filtered = append(filtered, s)
			}
		}
		if len(filtered) > 0 {
			index.SymbolIndex[key] = filtered
		} else {
			delete(index.SymbolIndex, key)
		}
	}

	// Remove API endpoints
	for _, endpoint := range enrichment.APIEndpoints {
		key := strings.ToLower(endpoint.Path)
		endpoints := index.APIIndex[key]
		filtered := []APIEndpoint{}
		for _, e := range endpoints {
			if e.DocumentID != documentID {
				filtered = append(filtered, e)
			}
		}
		if len(filtered) > 0 {
			index.APIIndex[key] = filtered
		} else {
			delete(index.APIIndex, key)
		}
	}

	// Remove from concepts
	for _, conceptName := range enrichment.Concepts {
		key := strings.ToLower(conceptName)
		if concept, exists := index.ConceptIndex[key]; exists {
			filtered := []string{}
			for _, docID := range concept.DocumentIDs {
				if docID != documentID {
					filtered = append(filtered, docID)
				}
			}
			concept.DocumentIDs = filtered

			// Remove concept if no documents reference it
			if len(concept.DocumentIDs) == 0 {
				delete(index.ConceptIndex, key)
			}
		}
	}

	index.LastUpdated = time.Now()

	// Save to disk
	if err := s.saveIndexUnlocked(index); err != nil {
		return err
	}

	// Update cache
	s.cache[uuid] = index

	return nil
}

// QuerySymbols searches for symbols by name
func (s *Storage) QuerySymbols(uuid, symbolName string) ([]CodeSymbol, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index, err := s.loadIndexUnlocked(uuid)
	if err != nil {
		return nil, err
	}

	key := strings.ToLower(symbolName)
	symbols, exists := index.SymbolIndex[key]
	if !exists {
		// Try partial match
		var matches []CodeSymbol
		for k, syms := range index.SymbolIndex {
			if strings.Contains(k, key) {
				matches = append(matches, syms...)
			}
		}
		return matches, nil
	}

	return symbols, nil
}

// QueryAPIs searches for API endpoints by path
func (s *Storage) QueryAPIs(uuid, path string) ([]APIEndpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index, err := s.loadIndexUnlocked(uuid)
	if err != nil {
		return nil, err
	}

	key := strings.ToLower(path)
	endpoints, exists := index.APIIndex[key]
	if !exists {
		// Try partial match
		var matches []APIEndpoint
		for k, eps := range index.APIIndex {
			if strings.Contains(k, key) {
				matches = append(matches, eps...)
			}
		}
		return matches, nil
	}

	return endpoints, nil
}

// QueryConcepts searches for concepts
func (s *Storage) QueryConcepts(uuid, conceptName string) (*Concept, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index, err := s.loadIndexUnlocked(uuid)
	if err != nil {
		return nil, err
	}

	key := strings.ToLower(conceptName)
	concept, exists := index.ConceptIndex[key]
	if !exists {
		return nil, fmt.Errorf("concept not found: %s", conceptName)
	}

	return concept, nil
}

// GetIndex returns the complete index for a user
func (s *Storage) GetIndex(uuid string) (*SourceOfTruthIndex, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.loadIndexUnlocked(uuid)
}

// GroundQuery analyzes a query and finds relevant code elements
func (s *Storage) GroundQuery(uuid, query string) (*QueryGroundingResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index, err := s.loadIndexUnlocked(uuid)
	if err != nil {
		return nil, err
	}

	result := &QueryGroundingResult{
		Query:            query,
		DetectedSymbols:  []string{},
		DetectedAPIs:     []string{},
		DetectedConcepts: []string{},
		MatchedSymbols:   []CodeSymbol{},
		MatchedAPIs:      []APIEndpoint{},
		MatchedConcepts:  []Concept{},
		RecommendedDocs:  []string{},
		GroundingScore:   0.0,
	}

	queryLower := strings.ToLower(query)
	words := strings.Fields(queryLower)

	docScores := make(map[string]float64) // Document ID -> score

	// Detect and match symbols
	for symbolKey, symbols := range index.SymbolIndex {
		for _, word := range words {
			if strings.Contains(word, symbolKey) || strings.Contains(symbolKey, word) {
				if !contains(result.DetectedSymbols, symbols[0].Symbol) {
					result.DetectedSymbols = append(result.DetectedSymbols, symbols[0].Symbol)
				}
				result.MatchedSymbols = append(result.MatchedSymbols, symbols...)

				// Boost document scores
				for _, symbol := range symbols {
					docScores[symbol.DocumentID] += 2.0
				}
			}
		}
	}

	// Detect and match APIs
	for apiKey, endpoints := range index.APIIndex {
		if strings.Contains(queryLower, apiKey) || strings.Contains(apiKey, strings.ReplaceAll(queryLower, " ", "")) {
			if !contains(result.DetectedAPIs, endpoints[0].Path) {
				result.DetectedAPIs = append(result.DetectedAPIs, endpoints[0].Path)
			}
			result.MatchedAPIs = append(result.MatchedAPIs, endpoints...)

			// Boost document scores
			for _, endpoint := range endpoints {
				docScores[endpoint.DocumentID] += 2.5
			}
		}
	}

	// Detect and match concepts
	for conceptKey, concept := range index.ConceptIndex {
		if strings.Contains(queryLower, conceptKey) {
			if !contains(result.DetectedConcepts, concept.Name) {
				result.DetectedConcepts = append(result.DetectedConcepts, concept.Name)
			}
			result.MatchedConcepts = append(result.MatchedConcepts, *concept)

			// Boost document scores
			for _, docID := range concept.DocumentIDs {
				docScores[docID] += 1.0
			}
		}
	}

	// Calculate grounding score
	matchCount := len(result.DetectedSymbols) + len(result.DetectedAPIs) + len(result.DetectedConcepts)
	if matchCount > 0 {
		result.GroundingScore = float64(matchCount) / (float64(len(words)) + 1.0)
		if result.GroundingScore > 1.0 {
			result.GroundingScore = 1.0
		}
	}

	// Sort documents by score and recommend top ones
	type docScore struct {
		docID string
		score float64
	}
	var scores []docScore
	for docID, score := range docScores {
		scores = append(scores, docScore{docID, score})
	}

	// Simple bubble sort (fine for small lists)
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score > scores[i].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// Take top 5 documents
	maxDocs := 5
	if len(scores) < maxDocs {
		maxDocs = len(scores)
	}
	for i := 0; i < maxDocs; i++ {
		result.RecommendedDocs = append(result.RecommendedDocs, scores[i].docID)
	}

	return result, nil
}

// GetStats returns statistics about the code source of truth
func (s *Storage) GetStats(uuid string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	index, err := s.loadIndexUnlocked(uuid)
	if err != nil {
		return nil, err
	}

	totalCodeBlocks := 0
	languageDistribution := make(map[string]int)

	for _, enrichment := range index.DocumentIndex {
		totalCodeBlocks += len(enrichment.CodeBlocks)
		for _, block := range enrichment.CodeBlocks {
			languageDistribution[block.Language]++
		}
	}

	return map[string]interface{}{
		"total_documents":       len(index.DocumentIndex),
		"total_symbols":         len(index.SymbolIndex),
		"total_api_endpoints":   len(index.APIIndex),
		"total_concepts":        len(index.ConceptIndex),
		"total_code_blocks":     totalCodeBlocks,
		"language_distribution": languageDistribution,
		"last_updated":          index.LastUpdated,
	}, nil
}

// Private helper methods

func (s *Storage) loadIndexUnlocked(uuid string) (*SourceOfTruthIndex, error) {
	// Check cache first
	if index, exists := s.cache[uuid]; exists {
		return index, nil
	}

	// Load from disk
	indexPath := filepath.Join(s.baseDir, "enrichments", fmt.Sprintf("%s.json", uuid))
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("index not found for user %s", uuid)
		}
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	var index SourceOfTruthIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}

	// Cache it
	s.cache[uuid] = &index

	return &index, nil
}

func (s *Storage) saveIndexUnlocked(index *SourceOfTruthIndex) error {
	indexPath := filepath.Join(s.baseDir, "enrichments", fmt.Sprintf("%s.json", index.UUID))

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := os.WriteFile(indexPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	return nil
}
