# ML Worker - Document Parsing API Specification

## Overview
The Askara orchestrator calls this endpoint for advanced document parsing using Docling. This provides structured document parsing with table extraction and layout preservation.

## Endpoint

```
POST /parse
Content-Type: application/json
```

## Request

```json
{
  "document_url": "data:application/pdf;base64,JVBERi0xLj...",
  "preserve_structure": true,
  "extract_tables": true,
  "extract_images": true
}
```

### Request Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `document_url` | string (URI) | Yes | Base64 data URL of the document (`data:<mime>;base64,<data>`) or HTTP URL |
| `preserve_structure` | boolean | No (default: true) | Whether to preserve document structure (headings, sections) |
| `extract_tables` | boolean | No (default: true) | Whether to extract tables as structured data |
| `extract_images` | boolean | No (default: true) | Whether to extract images from the document |

## Response

```json
{
  "text": "Full extracted text content...",
  "markdown": "# Introduction\n\n## Background\n\nFull extracted content in markdown format...",
  "structure": {
    "headings": [
      {"level": 1, "text": "Introduction", "page": 1},
      {"level": 2, "text": "Background", "page": 1}
    ],
    "sections": [...]
  },
  "tables": [
    {
      "headers": ["Column 1", "Column 2", "Column 3"],
      "rows": [
        ["Value 1", "Value 2", "Value 3"],
        ["Value 4", "Value 5", "Value 6"]
      ],
      "caption": "Table 1: Summary Statistics",
      "metadata": {
        "page": 3,
        "confidence": 0.95
      }
    }
  ],
  "metadata": {
    "pages": 10,
    "language": "en",
    "file_size": 1234567
  },
  "page_count": 10,
  "processing_time_ms": 1542
}
```

### Response Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `text` | string | Yes | Full extracted text content (plain text) |
| `markdown` | string | Yes | Full extracted content in markdown format with preserved structure |
| `structure` | object | Yes | Document structure (headings, sections) |
| `tables` | array | Yes | Extracted tables (empty array if none found) |
| `metadata` | object | Yes | Document metadata (language, file info, etc.) |
| `page_count` | integer | Yes | Total number of pages in the document |
| `processing_time_ms` | integer | Yes | Processing time in milliseconds |

### Table Object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `headers` | string[] | Yes | Table column headers |
| `rows` | string[][] | Yes | Table rows (array of arrays) |
| `caption` | string | No | Table caption/title |
| `metadata` | object | No | Additional table metadata (page number, confidence) |

## Error Response

```json
{
  "error": "Error message",
  "details": "Detailed error information"
}
```

HTTP Status Codes:
- `200 OK`: Successful parsing
- `400 Bad Request`: Invalid request (missing fields, invalid data URL)
- `500 Internal Server Error`: Parsing failed

## Implementation Notes

### Recommended: Docling

```python
from docling.document_converter import DocumentConverter

converter = DocumentConverter()

# Parse document
result = converter.convert(document_path)

# Extract text
text = result.document.export_to_markdown()

# Extract tables
tables = []
for table in result.document.tables:
    tables.append({
        "headers": table.data[0] if table.data else [],
        "rows": table.data[1:] if len(table.data) > 1 else [],
        "caption": table.caption,
        "metadata": {
            "page": table.page_number,
            "confidence": table.confidence
        }
    })

# Extract structure
structure = {
    "headings": [
        {
            "level": heading.level,
            "text": heading.text,
            "page": heading.page_number
        }
        for heading in result.document.headings
    ]
}
```

### Fallback Strategies

If Docling fails, implement fallbacks:
1. PyMuPDF (fitz) for PDF text extraction
2. python-docx for DOCX files
3. python-pptx for PPTX files
4. BeautifulSoup for HTML files

Always return `"method": "fallback"` when not using Docling.

## Integration with Askara

### Client Code (Already Implemented)

```go
// In Go orchestrator (already done)
import "github.com/itaiguardiola/askara/mlworker"

mlClient := mlworker.GetClient()

// Parse document
resp, err := mlClient.ParseDocument(dataURL, true, true)
if err != nil {
    // Handle error - fall back to docconv
}

// Use parsed text
text := resp.Text

// Access tables
for _, table := range resp.Tables {
    // Process table data
}
```

### Parsing Flow

1. **Try ML Worker (Docling)** - If `parse` feature enabled
2. **Fallback to docconv** - Fast local text extraction
3. **Fallback to OCR** - For scanned/image-based PDFs

### Feature Flag

Enable in `.env`:
```bash
ML_WORKER_ENABLED=true
ML_WORKER_FEATURES=ocr,enhance,caption,parse
```

## Testing

### Test Request

```bash
curl -X POST http://192.168.0.250:6161/askara/parse \
  -H "Content-Type: application/json" \
  -d '{
    "document_url": "data:application/pdf;base64,JVBERi0xLj...",
    "preserve_structure": true,
    "extract_tables": true
  }'
```

### Expected Response

```json
{
  "text": "Document text content with preserved structure...",
  "tables": [
    {
      "headers": ["Name", "Age", "City"],
      "rows": [
        ["John", "30", "NYC"],
        ["Jane", "25", "LA"]
      ]
    }
  ],
  "method": "docling"
}
```

## Performance Considerations

- Cache parsed results if document hasn't changed
- Use async processing for large documents
- Set timeout to 30 seconds (configurable via `ML_WORKER_TIMEOUT`)
- Return partial results on timeout rather than failing completely

## Dependencies

```bash
pip install docling
pip install pymupdf  # For fallback
pip install python-docx  # For DOCX files
```

## Questions?

Contact the Askara orchestrator team for integration support.
