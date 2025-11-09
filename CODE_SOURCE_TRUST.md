# Code Source of Truth for Grounding

## Overview

The **Code Source of Trust** system is an intelligent grounding enhancement for Askara that automatically extracts and indexes code elements from uploaded documentation. This enables more accurate, code-aware responses when users ask questions about APIs, functions, classes, and other code-related topics.

## Features

### 1. **Automatic Code Symbol Extraction**
When documents are uploaded, the system automatically:
- Extracts function, class, and method definitions
- Identifies API endpoints (REST, GraphQL patterns)
- Detects code blocks in multiple languages (Go, JavaScript, Python, Java, TypeScript, Rust, etc.)
- Captures code signatures and context

### 2. **Multi-Language Support**
Supports symbol extraction for:
- **Go**: Functions, types, structs, interfaces
- **JavaScript/TypeScript**: Functions, classes, arrow functions, interfaces
- **Python**: Functions, classes, methods
- **Java**: Classes, methods, constructors
- And more through configurable pattern matching

### 3. **API Endpoint Detection**
Automatically identifies API patterns:
- HTTP methods: GET, POST, PUT, DELETE, PATCH
- Route definitions (router.get('/api/...'), http.HandleFunc, etc.)
- RESTful path patterns

### 4. **Code-Aware Query Grounding**
When users ask questions:
- Analyzes queries for code symbol mentions
- Matches symbols, APIs, and concepts from the index
- Enhances context with structured metadata
- Recommends relevant documents based on code content

### 5. **Enriched Context for LLM**
The system provides the LLM with:
- Symbol definitions and signatures
- API endpoint documentation
- Code examples and usage patterns
- Cross-referenced documentation

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                  Document Upload                        │
│                       ↓                                  │
│              ┌────────────────┐                          │
│              │ Code Parser    │                          │
│              │ - Extract      │                          │
│              │   symbols      │                          │
│              │ - Extract APIs │                          │
│              │ - Extract code │                          │
│              │   blocks       │                          │
│              └────────┬───────┘                          │
│                       ↓                                  │
│          ┌────────────────────────┐                      │
│          │  Document Enrichment   │                      │
│          │  - Symbols[]           │                      │
│          │  - APIEndpoints[]      │                      │
│          │  - CodeBlocks[]        │                      │
│          │  - Concepts[]          │                      │
│          └────────┬───────────────┘                      │
│                   ↓                                      │
│      ┌────────────────────────────┐                      │
│      │  Source of Truth Index     │                      │
│      │  - SymbolIndex             │                      │
│      │  - APIIndex                │                      │
│      │  - ConceptIndex            │                      │
│      │  - DocumentIndex           │                      │
│      └────────────┬───────────────┘                      │
│                   ↓                                      │
│         Stored in data/enrichments/                      │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│                   Query Processing                       │
│                       ↓                                  │
│           ┌──────────────────┐                           │
│           │ Ground Query     │                           │
│           │ - Detect symbols │                           │
│           │ - Match APIs     │                           │
│           │ - Find concepts  │                           │
│           └────────┬─────────┘                           │
│                    ↓                                     │
│         ┌──────────────────────┐                         │
│         │ Enhanced Context     │                         │
│         │ + Code metadata      │                         │
│         │ + Symbol definitions │                         │
│         │ + API documentation  │                         │
│         └──────────┬───────────┘                         │
│                    ↓                                     │
│              LLM Response                                │
└─────────────────────────────────────────────────────────┘
```

## Configuration

### Environment Variables

```bash
# Enable/disable Code Source Trust (enabled by default)
CODE_TRUST_ENABLED=true
```

### Extraction Configuration

The system uses a default configuration that can be customized in code:

```go
config := codesourcetrust.ExtractionConfig{
    EnableCodeBlocks:   true,
    EnableSymbols:      true,
    EnableAPIs:         true,
    EnableConcepts:     true,
    SupportedLanguages: []string{"go", "javascript", "typescript", "python", "java", "rust"},
    MinCodeBlockLines:  3,     // Minimum lines for a code block
    MaxContextChars:    500,   // Context around code to capture
}
```

## API Endpoints

### 1. Get Code Trust Statistics
```
GET /api/codetrust/stats?uuid={uuid}
```

**Response:**
```json
{
  "stats": {
    "total_documents": 15,
    "total_symbols": 142,
    "total_api_endpoints": 38,
    "total_concepts": 12,
    "total_code_blocks": 87,
    "language_distribution": {
      "go": 45,
      "javascript": 25,
      "python": 17
    },
    "last_updated": "2025-11-08T10:30:00Z"
  }
}
```

### 2. Search Symbols
```
GET /api/codetrust/symbols?uuid={uuid}&query={symbolName}
```

**Response:**
```json
{
  "symbols": [
    {
      "symbol": "getUserAPI",
      "type": "function",
      "language": "go",
      "document_id": "doc_abc123",
      "document_name": "api-docs.md",
      "signature": "func getUserAPI(w http.ResponseWriter, r *http.Request)",
      "context": "This function handles GET requests to retrieve user information..."
    }
  ]
}
```

### 3. Search API Endpoints
```
GET /api/codetrust/apis?uuid={uuid}&query={path}
```

**Response:**
```json
{
  "apis": [
    {
      "method": "GET",
      "path": "/api/users/{id}",
      "description": "Retrieves user information by ID",
      "document_id": "doc_abc123",
      "document_name": "api-reference.md",
      "parameters": ["id"],
      "examples": ["curl GET /api/users/123"]
    }
  ]
}
```

### 4. Ground a Query
```
POST /api/codetrust/ground
Content-Type: application/json

{
  "uuid": "user-uuid",
  "query": "How do I use the getUserAPI function?"
}
```

**Response:**
```json
{
  "grounding": {
    "query": "How do I use the getUserAPI function?",
    "detected_symbols": ["getUserAPI"],
    "detected_apis": ["/api/users"],
    "detected_concepts": ["API", "user management"],
    "matched_symbols": [...],
    "matched_apis": [...],
    "matched_concepts": [...],
    "recommended_docs": ["doc_abc123", "doc_xyz789"],
    "grounding_score": 0.85
  }
}
```

### 5. Get Document Enrichment
```
GET /api/codetrust/enrichment/{documentId}?uuid={uuid}
```

**Response:**
```json
{
  "document_id": "doc_abc123",
  "uuid": "user-uuid",
  "document_name": "api-docs.md",
  "symbols": [...],
  "code_blocks": [...],
  "api_endpoints": [...],
  "concepts": ["authentication", "API", "REST"],
  "keywords": ["user", "token", "endpoint"],
  "last_updated": "2025-11-08T10:30:00Z",
  "processed_version": "1.0.0"
}
```

### 6. Get Full Index
```
GET /api/codetrust/index?uuid={uuid}
```

Returns the complete Source of Truth index for a user.

## Data Structures

### CodeSymbol
```go
type CodeSymbol struct {
    Symbol       string   // e.g., "getUserAPI", "User.login()"
    Type         string   // function, class, method, interface, constant
    Language     string   // go, javascript, python, etc.
    DocumentID   string   // Source document
    DocumentName string   // Filename
    Context      string   // Surrounding text
    Signature    string   // Function signature or definition
}
```

### APIEndpoint
```go
type APIEndpoint struct {
    Method       string   // GET, POST, PUT, DELETE, etc.
    Path         string   // /api/users/{id}
    Description  string   // What the endpoint does
    Parameters   []string // Query/path parameters
    DocumentID   string   // Source document
    DocumentName string   // Filename
}
```

### CodeBlock
```go
type CodeBlock struct {
    ID           string    // Unique ID
    DocumentID   string    // Source document
    Language     string    // Detected language
    Code         string    // The actual code
    Context      string    // Text before/after
    Symbols      []string  // Symbols defined/used
}
```

## Integration with RAG Pipeline

The Code Source Trust system integrates seamlessly with Askara's existing RAG pipeline:

1. **Document Upload**: Extraction happens asynchronously during upload
2. **Query Processing**: Grounding occurs before retrieval
3. **Context Enhancement**: Code metadata is added to retrieved contexts
4. **LLM Completion**: Enhanced context leads to more accurate answers

### Example Flow

```
User Query: "How do I call the getUserAPI endpoint?"
      ↓
[Code Grounding]
- Detected: getUserAPI (symbol)
- Detected: /api/users (API endpoint)
- Grounding Score: 0.85
      ↓
[Enhanced Retrieval]
- Prioritize documents: doc_abc123, doc_xyz789
- Add symbol definitions to context
- Add API endpoint documentation
      ↓
[LLM Response]
"The getUserAPI endpoint is available at GET /api/users/{id}.
Here's how to call it:

```bash
curl -X GET http://localhost:8100/api/users/123 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

The function signature is:
func getUserAPI(w http.ResponseWriter, r *http.Request)
..."
```

## Storage

Enrichment data is stored in:
```
data/enrichments/{uuid}.json
```

Each user has a single index file containing all their enriched document metadata.

## Performance

- **Extraction**: Asynchronous (doesn't block upload response)
- **Query Grounding**: < 50ms typical
- **Storage**: JSON-based, efficient for read-heavy workloads
- **Memory**: In-memory cache for frequently accessed indices

## Code Examples

### Using the Service

```go
// Initialize the service
config := codesourcetrust.DefaultExtractionConfig()
service, err := codesourcetrust.NewService("data", config)

// Process a document
enrichment, err := service.ProcessDocument(
    documentID,
    documentName,
    uuid,
    documentText,
)

// Ground a query
grounding, err := service.GroundQuery(uuid, "How do I use the login API?")

// Search for symbols
symbols, err := service.SearchSymbols(uuid, "login")
```

## Future Enhancements

Potential improvements:
1. **Semantic search** within code symbols
2. **Call graph analysis** to show symbol relationships
3. **Version tracking** for documentation changes
4. **Code example extraction** with execution context
5. **Integration with ML-Worker** for advanced parsing
6. **Documentation coverage metrics**
7. **Cross-reference detection** between documents

## Troubleshooting

### Symbols Not Detected
- Check that documents contain code in supported languages
- Verify code blocks use proper markdown fencing (\`\`\`language)
- Increase logging to see extraction details

### Query Not Grounded
- Check grounding score in logs
- Verify symbols are properly extracted
- Try more specific symbol names in queries

### Performance Issues
- Check index file size in data/enrichments/
- Consider implementing pagination for large symbol sets
- Monitor memory usage with many documents

## Related Files

- `/codesourcetrust/*.go` - Core module
- `/vault-web-server/postapi/codetrust.go` - API handlers
- `/vault-web-server/postapi/questions.go` - Query grounding integration
- `/vault-web-server/postapi/fileupload.go` - Upload integration
- `/vault-web-server/main.go` - Service initialization

## License

Part of the Askara project.
