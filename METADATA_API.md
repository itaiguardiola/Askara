# Document Metadata API

Askara now supports rich metadata for documents with ML-powered auto-tagging capabilities.

## Features

### Automatic Metadata Extraction
When a document is uploaded, Askara automatically:
- Generates a concise summary using LLM
- Extracts relevant tags and topics
- Detects the document language
- Identifies named entities (people, organizations, locations)
- Categorizes the document
- Calculates word count and reading time

### User-Editable Metadata
Users can customize metadata fields:
- Title (defaults to filename)
- Creator/Author
- Description
- Category
- Custom tags

## Metadata Schema

```json
{
  "title": "Document Title",
  "creator": "Author Name",
  "description": "User-provided description",
  "category": "Research|Technical|Business|Legal|Academic|Financial|Medical|General",
  "custom_tags": ["tag1", "tag2"],
  "auto_tags": [
    {
      "name": "machine-learning",
      "confidence": 0.9,
      "source": "ml"
    }
  ],
  "summary": "AI-generated 2-3 sentence summary",
  "language": "English",
  "topics": ["topic1", "topic2"],
  "entities": [
    {
      "text": "OpenAI",
      "type": "ORG",
      "count": 5
    }
  ],
  "word_count": 1500,
  "reading_time": 8,
  "last_modified": "2025-11-08T10:30:00Z",
  "version": 1
}
```

## API Endpoints

### Get Document Metadata

```http
GET /api/documents/{documentId}/metadata?uuid={uuid}
```

**Response:**
```json
{
  "metadata": { ... }
}
```

---

### Update Document Metadata

```http
PUT /api/documents/{documentId}/metadata
```

**Request Body:**
```json
{
  "uuid": "user-uuid",
  "title": "New Title",
  "creator": "John Doe",
  "description": "Updated description",
  "category": "Research",
  "custom_tags": ["ai", "nlp", "rag"]
}
```

**Response:**
```json
{
  "message": "Metadata updated successfully",
  "metadata": { ... }
}
```

---

### Add Custom Tag

```http
POST /api/documents/{documentId}/tags
```

**Request Body:**
```json
{
  "uuid": "user-uuid",
  "tag": "important"
}
```

**Response:**
```json
{
  "message": "Tag added successfully",
  "tags": ["important", "ai", "research"]
}
```

---

### Remove Custom Tag

```http
DELETE /api/documents/{documentId}/tags/{tag}?uuid={uuid}
```

**Response:**
```json
{
  "message": "Tag removed successfully",
  "tags": ["ai", "research"]
}
```

---

### Search Documents by Tags

```http
POST /api/documents/search/tags
```

**Request Body:**
```json
{
  "uuid": "user-uuid",
  "tags": ["ai", "research"]
}
```

**Response:**
```json
{
  "documents": [
    {
      "id": "doc-abc123",
      "filename": "research-paper.pdf",
      "metadata": { ... }
    }
  ],
  "count": 5
}
```

---

### Get All Tags

```http
GET /api/documents/tags?uuid={uuid}
```

**Response:**
```json
{
  "tags": [
    {
      "tag": "ai",
      "count": 10
    },
    {
      "tag": "research",
      "count": 7
    }
  ],
  "total": 25
}
```

## ML Auto-Tagging

### How It Works

1. **Content Analysis**: When a document is uploaded, the first ~12,000 characters are analyzed
2. **LLM Extraction**: The configured LLM provider (OpenAI or Ollama) extracts structured metadata
3. **Confidence Scores**: Auto-generated tags include confidence scores (0.0-1.0)
4. **Fallback**: If ML extraction fails, basic metadata is created using simple heuristics

### Prompt Template

The metadata extractor uses the following structured prompt:

```
You are a document analysis expert. Analyze the following document and extract structured metadata.

Document: {filename}
Content: {content_sample}

Please analyze this document and provide metadata in JSON format:
{
  "summary": "A concise 2-3 sentence summary",
  "tags": ["tag1", "tag2", ...],
  "topics": ["main topic 1", "main topic 2"],
  "category": "Research|Technical|Business|...",
  "language": "English",
  "entities": {
    "entity name": "type (PERSON, ORG, LOCATION, etc.)"
  },
  "key_concepts": ["concept1", "concept2"]
}
```

### Customization

To use a different LLM for metadata extraction, configure your LLM provider in `.env`:

```bash
# Use OpenAI
LLM_PROVIDER=openai
OPENAI_API_KEY=your-key

# Or use Ollama (local)
LLM_PROVIDER=ollama
OLLAMA_API_ENDPOINT=http://localhost:11434
OLLAMA_MODEL=llama2
```

## Storage

Metadata is stored in the document JSON files at `data/documents/{uuid}.json`:

```json
{
  "documents": [
    {
      "id": "doc-abc123",
      "uuid": "user-uuid",
      "filename": "example.pdf",
      "file_size": 1024000,
      "upload_date": "2025-11-08T10:00:00Z",
      "chunk_count": 10,
      "content_type": "application/pdf",
      "first_chunk_id": "...",
      "last_chunk_id": "...",
      "metadata": {
        "title": "Example Document",
        "auto_tags": [...],
        "summary": "...",
        ...
      }
    }
  ]
}
```

## Integration with RAG

Metadata enhances the RAG (Retrieval Augmented Generation) system:

1. **Better Organization**: Documents can be filtered by tags/category before semantic search
2. **Context Enrichment**: Entity information and topics provide additional context
3. **Smart Retrieval**: Combine metadata filters with vector similarity search
4. **Source Attribution**: Metadata provides better citation information in answers

## Future Enhancements

Planned features:
- [ ] Bulk metadata operations
- [ ] Metadata export (CSV, JSON)
- [ ] Frontend UI for metadata editing
- [ ] Tag suggestions based on existing documents
- [ ] Document relationships and references
- [ ] Metadata-based access control
- [ ] Integration with vector DB metadata filters
- [ ] Custom metadata fields per user

## Examples

### Upload and Auto-Tag

```bash
# Upload a document (metadata automatically extracted)
curl -X POST http://localhost:8080/upload \
  -F "files=@research.pdf" \
  -F "uuid=user-123"

# Response includes auto-generated metadata
{
  "uploaded_documents": [{
    "id": "doc-abc",
    "metadata": {
      "auto_tags": [
        {"name": "machine-learning", "confidence": 0.9},
        {"name": "neural-networks", "confidence": 0.85}
      ],
      "summary": "This paper discusses...",
      "topics": ["AI", "Deep Learning"]
    }
  }]
}
```

### Update Metadata

```bash
# Add custom information
curl -X PUT http://localhost:8080/api/documents/doc-abc/metadata \
  -H "Content-Type: application/json" \
  -d '{
    "uuid": "user-123",
    "title": "Transformer Architecture Deep Dive",
    "creator": "John Doe",
    "category": "Research",
    "custom_tags": ["important", "review-later"]
  }'
```

### Search by Tags

```bash
# Find all ML-related documents
curl -X POST http://localhost:8080/api/documents/search/tags \
  -H "Content-Type: application/json" \
  -d '{
    "uuid": "user-123",
    "tags": ["machine-learning", "ai"]
  }'
```
