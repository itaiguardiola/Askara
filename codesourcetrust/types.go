package codesourcetrust

import "time"

// CodeSymbol represents a code element found in documentation
type CodeSymbol struct {
	Symbol       string   `json:"symbol"`        // e.g., "getUserAPI", "User.login()"
	Type         string   `json:"type"`          // function, class, method, interface, constant, endpoint
	Language     string   `json:"language"`      // go, javascript, python, etc.
	DocumentID   string   `json:"document_id"`   // Source document
	DocumentName string   `json:"document_name"` // Filename for attribution
	ChunkIDs     []string `json:"chunk_ids"`     // Vector DB chunks containing this symbol
	Context      string   `json:"context"`       // Surrounding text for context
	Signature    string   `json:"signature"`     // Function signature or definition
	Description  string   `json:"description"`   // Extracted description if available
	LineNumber   int      `json:"line_number"`   // Line number in document (if available)
}

// CodeBlock represents a code snippet found in documentation
type CodeBlock struct {
	ID           string    `json:"id"`            // Unique ID
	DocumentID   string    `json:"document_id"`   // Source document
	DocumentName string    `json:"document_name"` // Filename
	Language     string    `json:"language"`      // Detected language
	Code         string    `json:"code"`          // The actual code
	Context      string    `json:"context"`       // Text before/after explaining the code
	Symbols      []string  `json:"symbols"`       // Symbols defined/used in this block
	LineStart    int       `json:"line_start"`    // Starting line
	LineEnd      int       `json:"line_end"`      // Ending line
	CreatedAt    time.Time `json:"created_at"`
}

// APIEndpoint represents documented API endpoints
type APIEndpoint struct {
	Method       string   `json:"method"`        // GET, POST, PUT, DELETE, etc.
	Path         string   `json:"path"`          // /api/users/{id}
	Description  string   `json:"description"`   // What the endpoint does
	Parameters   []string `json:"parameters"`    // Query/path parameters
	DocumentID   string   `json:"document_id"`   // Source document
	DocumentName string   `json:"document_name"` // Filename
	Examples     []string `json:"examples"`      // Usage examples
}

// ConceptMapping represents high-level concepts and their documentation
type Concept struct {
	Name        string   `json:"name"`         // e.g., "Authentication", "User Management"
	Keywords    []string `json:"keywords"`     // Related search terms
	DocumentIDs []string `json:"document_ids"` // Documents covering this concept
	Symbols     []string `json:"symbols"`      // Related code symbols
	Description string   `json:"description"`  // Concept explanation
}

// DocumentEnrichment is the complete metadata for a document
type DocumentEnrichment struct {
	DocumentID       string        `json:"document_id"`
	UUID             string        `json:"uuid"` // User ID
	DocumentName     string        `json:"document_name"`
	Symbols          []CodeSymbol  `json:"symbols"`
	CodeBlocks       []CodeBlock   `json:"code_blocks"`
	APIEndpoints     []APIEndpoint `json:"api_endpoints"`
	Concepts         []string      `json:"concepts"` // High-level topics covered
	Keywords         []string      `json:"keywords"` // Searchable terms
	LastUpdated      time.Time     `json:"last_updated"`
	ProcessedVersion string        `json:"processed_version"` // Version of parser used
}

// SourceOfTruthIndex is the master index for all code documentation
type SourceOfTruthIndex struct {
	UUID          string                         `json:"uuid"`           // User ID
	SymbolIndex   map[string][]CodeSymbol        `json:"symbol_index"`   // Symbol name -> Symbols
	APIIndex      map[string][]APIEndpoint       `json:"api_index"`      // Endpoint path -> Endpoints
	ConceptIndex  map[string]*Concept            `json:"concept_index"`  // Concept name -> Concept
	DocumentIndex map[string]*DocumentEnrichment `json:"document_index"` // Document ID -> Enrichment
	LastUpdated   time.Time                      `json:"last_updated"`
}

// QueryGroundingResult provides enriched context for code-related queries
type QueryGroundingResult struct {
	Query            string        `json:"query"`
	DetectedSymbols  []string      `json:"detected_symbols"`  // Symbols mentioned in query
	DetectedAPIs     []string      `json:"detected_apis"`     // API endpoints mentioned
	DetectedConcepts []string      `json:"detected_concepts"` // Concepts identified
	MatchedSymbols   []CodeSymbol  `json:"matched_symbols"`   // Matching symbol definitions
	MatchedAPIs      []APIEndpoint `json:"matched_apis"`      // Matching endpoints
	MatchedConcepts  []Concept     `json:"matched_concepts"`  // Related concepts
	RecommendedDocs  []string      `json:"recommended_docs"`  // Documents to prioritize
	GroundingScore   float64       `json:"grounding_score"`   // Confidence in grounding (0-1)
}

// ExtractionConfig controls the code extraction process
type ExtractionConfig struct {
	EnableCodeBlocks   bool     `json:"enable_code_blocks"`
	EnableSymbols      bool     `json:"enable_symbols"`
	EnableAPIs         bool     `json:"enable_apis"`
	EnableConcepts     bool     `json:"enable_concepts"`
	SupportedLanguages []string `json:"supported_languages"`  // empty = all
	MinCodeBlockLines  int      `json:"min_code_block_lines"` // Minimum lines to consider a code block
	MaxContextChars    int      `json:"max_context_chars"`    // Context to capture around code
}

// DefaultExtractionConfig returns sensible defaults
func DefaultExtractionConfig() ExtractionConfig {
	return ExtractionConfig{
		EnableCodeBlocks:   true,
		EnableSymbols:      true,
		EnableAPIs:         true,
		EnableConcepts:     true,
		SupportedLanguages: []string{"go", "javascript", "typescript", "python", "java", "rust", "c", "cpp", "ruby", "php", "bash", "sql"},
		MinCodeBlockLines:  3,
		MaxContextChars:    500,
	}
}
