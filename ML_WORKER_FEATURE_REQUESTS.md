# ML-Worker Feature Requests for AsKara Integration

## Current Limitations

### 1. Parse Endpoint File Size Limit
- **Issue**: 337KB PDF generates ~450KB base64 data URL, exceeds 2083 char limit
- **Current Workaround**: Skip ML-Worker parse, fall back to docconv (poor quality)
- **Impact**: Cannot use advanced Docling parser for large/complex PDFs

### 2. No Multipart Upload Support
- **Issue**: Parse endpoint only accepts JSON with `document_url` field
- **Tested**: Multipart form-data returns Pydantic validation error
- **Impact**: Cannot send large files efficiently

### 3. OCR Page-by-Page Inefficiency
- **Issue**: Must convert PDF to images locally, send each page separately
- **Current**: Use pdftoppm → send N images → combine results
- **Impact**: Multiple network calls, slower processing, more error-prone

## Recommended Feature Requests

### Priority 1: Multipart File Upload for Parse Endpoint

**Request**: Add multipart/form-data support to `/askara/parse` endpoint

**Benefits**:
- Send large PDFs (multi-MB) without base64 encoding overhead
- 33% smaller payload (no base64 inflation)
- Standard HTTP file upload pattern
- Works with Docker file size limits

**Proposed API**:
```bash
curl -X POST http://ml-worker:6161/askara/parse \
  -F "document=@large_file.pdf" \
  -F "preserve_structure=true" \
  -F "extract_tables=true"
```

**Response**: Same as current (markdown, tables, metadata)

### Priority 2: Hybrid Parse Mode (OCR + Text Extraction)

**Request**: Add `hybrid_mode` option to parse endpoint

**Use Case**: PDFs with mixed content (embedded text + scanned images)

**Proposed API**:
```json
{
  "document_url": "data:application/pdf;base64,...",
  "preserve_structure": true,
  "extract_tables": true,
  "hybrid_mode": true,          // NEW: Enable OCR for low-quality text
  "quality_threshold": 0.6       // NEW: OCR if text quality < 60%
}
```

**Benefits**:
- Automatic quality detection (like AsKara's `isTextExtractionPoor`)
- Falls back to OCR for garbled/poor text pages
- Best of both worlds: fast text extraction + accurate OCR
- Single API call instead of fallback chain

### Priority 3: PDF Chunked Processing

**Request**: Support page-range parsing for large PDFs

**Use Case**: Process 100+ page documents incrementally

**Proposed API**:
```json
{
  "document_url": "data:application/pdf;base64,...",
  "page_range": "1-10",          // NEW: Process specific pages
  "preserve_structure": true
}
```

**Benefits**:
- AsKara can split large PDFs into smaller chunks
- Process page ranges in parallel
- Reduce memory usage on ML-Worker
- Better error recovery (retry single page ranges)

### Priority 4: Quality Confidence Metrics

**Request**: Add quality/confidence scores to parse response

**Proposed Response**:
```json
{
  "markdown": "...",
  "text": "...",
  "tables": [...],
  "page_count": 25,
  "processing_time_ms": 1234,
  "quality_metrics": {           // NEW
    "text_confidence": 0.95,     // Overall text extraction confidence
    "page_confidences": [0.98, 0.92, ...],  // Per-page scores
    "extraction_method": "hybrid" // "text", "ocr", or "hybrid"
  }
}
```

**Benefits**:
- AsKara can decide whether to retry with OCR
- Track quality metrics for analytics
- Identify problematic pages/documents
- Better user feedback ("Document quality: 95%")

### Priority 5: Streaming/Progressive Parsing

**Request**: Support Server-Sent Events (SSE) for progressive results

**Use Case**: Large documents that take 30+ seconds to process

**Proposed API**:
```bash
curl -N -X POST http://ml-worker:6161/askara/parse/stream \
  -F "document=@large.pdf"
```

**Response** (SSE):
```
event: page
data: {"page": 1, "text": "...", "confidence": 0.95}

event: page
data: {"page": 2, "text": "...", "confidence": 0.92}

event: complete
data: {"total_pages": 25, "total_chars": 12345}
```

**Benefits**:
- AsKara can show upload progress to users
- Start chunking/vectorizing while document still processing
- Better UX for large documents
- Reduces perceived latency

## Implementation Priority

1. **Multipart Upload** (Priority 1) - Solves immediate 337KB PDF issue
2. **Hybrid Mode** (Priority 2) - Improves accuracy for mixed-content PDFs
3. **Quality Metrics** (Priority 4) - Enables better fallback decisions
4. **Chunked Processing** (Priority 3) - Scales to very large documents
5. **Streaming** (Priority 5) - UX improvement for long processing

## Current AsKara Workarounds

Until ML-Worker implements these features, AsKara uses:

1. **Large files**: Skip ML-Worker parse → docconv → OCR fallback
2. **Poor quality**: pdftoppm page-by-page → ML-Worker OCR per page
3. **Quality detection**: Local `isTextExtractionPoor()` function
4. **No progress**: Silent processing (user sees "uploading..." spinner)

## Expected Impact on Accuracy

With **Priority 1 + 2** implemented:
- 337KB Safeco PDF would use ML-Worker Docling parser (best accuracy)
- Hybrid mode would auto-detect garbled text and apply OCR
- Single API call instead of 3-step fallback chain
- **Estimated improvement**: 70% better answer accuracy for complex PDFs

## Testing Documents

Documents that would benefit most from these features:
1. `Safeco25Policy.pdf` (337KB) - Currently fails ML-Worker parse
2. Large technical manuals (1-5MB) - Need chunked processing
3. Mixed content PDFs - Need hybrid mode
4. Scanned documents with tables - Need OCR + table extraction

## API Compatibility

All proposed features are **backwards compatible**:
- New fields are optional
- Existing API behavior unchanged
- AsKara can feature-detect via `/health` endpoint

Example feature detection:
```json
GET /askara/health
{
  "status": "healthy",
  "features": {
    "parse": true,
    "parse_multipart": true,    // NEW
    "parse_hybrid": true,        // NEW
    "ocr": true,
    "enhance": true,
    "caption": true
  }
}
```

AsKara checks `parse_multipart` before deciding to send large files.

## Questions for ML-Worker Team

1. What's the max file size you'd recommend for multipart uploads?
2. Would hybrid mode significantly increase processing time?
3. Is Docling capable of quality scoring for extracted text?
4. What's the expected memory usage for 100+ page PDF processing?
5. Would streaming results require significant architecture changes?

## References

- Current ML-Worker endpoint: `http://192.168.0.250:6161/askara`
- Current features enabled: `ocr,enhance,caption,parse`
- AsKara extraction code: `chunk/fileprocessing.go`, `chunk/ml_pdf_processor.go`
