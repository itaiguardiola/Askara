# ML Worker Client - Production-Ready Document Processing

Enhanced ML Worker client with retry logic, chunked processing, and intelligent error handling based on ML Worker team recommendations.

## Features Implemented

### 1. Automatic Retry Logic with Exponential Backoff

All document parsing operations now include automatic retry on timeout or server errors:

```go
// Retry sequence: 1s → 2s → 4s wait between attempts
options := mlworker.DefaultParseOptions()
options.MaxRetries = 3  // Default is 2
resp, err := client.ParseDocumentWithRetry(fileData, filename, options)
```

### 2. Chunked Processing for Large Documents

Documents >50 pages are automatically processed in chunks to prevent timeouts:

```go
// Automatically chunks documents >100 pages
resp, err := client.ParseDocumentAuto(fileData, filename)

// Manual chunking with custom chunk size
options := mlworker.DefaultParseOptions()
options.ChunkSize = 20  // Process 20 pages at a time
resp, err := client.ParseDocumentChunked(fileData, filename, totalPages, options)
```

### 3. Performance Monitoring

Slow document processing is automatically logged:

```
[MLWorker] ⚠️  SLOW PARSE: 215.3s for report.pdf (150 pages, 245000 chars, 25 tables)
```

### 4. Enhanced Parse Options

Full control over document processing parameters:

```go
options := mlworker.ParseOptions{
    PreserveStructure: true,    // Extract headings/structure
    ExtractTables:     true,     // Extract table data
    ExtractImages:     false,    // Extract images (resource intensive)
    HybridMode:        false,    // Use PyMuPDF fallback
    PageRange:         "1-10",   // Process specific pages
    MaxRetries:        2,        // Number of retry attempts
    ChunkSize:         50,       // Pages per chunk (0 = no chunking)
}

resp, err := client.ParseDocumentAdvanced(fileData, filename, options)
```

## Usage Examples

### Simple Usage (Recommended)

```go
// Uses default production settings with retry and chunking
client := mlworker.GetClient()
resp, err := client.ParseDocumentMultipart(fileData, "document.pdf", "application/pdf", true, true)
```

### Automatic Strategy Selection

```go
// Automatically determines best strategy based on document size
resp, err := client.ParseDocumentAuto(fileData, "document.pdf")
```

### Advanced Usage with Custom Options

```go
options := mlworker.ParseOptions{
    PreserveStructure: true,
    ExtractTables:     true,
    ExtractImages:     false,
    HybridMode:        false,
    PageRange:         "1-50",  // First 50 pages only
    MaxRetries:        3,
    ChunkSize:         0,        // No chunking
}

resp, err := client.ParseDocumentAdvanced(fileData, "document.pdf", options)
```

### Large Document Chunking

```go
// For 200-page document, process in 25-page chunks
options := mlworker.DefaultParseOptions()
options.ChunkSize = 25

resp, err := client.ParseDocumentChunked(fileData, "large-doc.pdf", 200, options)
// Results are automatically merged
```

## Expected Response Times

| Document Type       | Pages  | Typical Time | Strategy               |
|---------------------|--------|--------------|------------------------|
| Simple text PDF     | 10-20  | 10-30s       | Fast parse             |
| Complex with tables | 20-50  | 1-2 min      | Standard parse + retry |
| Heavy tables/images | 50-100 | 2-4 min      | Standard parse + retry |
| Very complex        | 100+   | 4-5 min      | Chunked processing     |

## Error Handling

### Timeout Errors

Automatically retried with exponential backoff:

```go
// First attempt: immediate
// Retry 1: wait 1 second
// Retry 2: wait 2 seconds
// Retry 3: wait 4 seconds (if MaxRetries=3)
```

### Server Errors (5xx)

Automatically retried on temporary server issues:

```
[MLWorker] Server error for document.pdf, retrying in 2s (attempt 2/3)
```

### Client Errors (4xx)

Not retried - returned immediately as these indicate invalid requests.

## Configuration

### Environment Variables

```env
ML_WORKER_ENABLED=true
ML_WORKER_ENDPOINT=http://192.168.0.250:6161/askara
ML_WORKER_FEATURES=ocr,enhance,caption,parse,rerank
ML_WORKER_TIMEOUT=300  # 5 minutes (increased from 30s)
```

### Default Production Settings

```go
options := mlworker.DefaultParseOptions()
// Returns:
// {
//   PreserveStructure: true,
//   ExtractTables:     true,
//   ExtractImages:     false,  // Disabled for performance
//   HybridMode:        false,  // Only use if docling fails
//   MaxRetries:        2,
//   ChunkSize:         50,     // Chunk docs >50 pages
// }
```

## Page Range Syntax

Process specific pages using the `PageRange` option:

```go
options.PageRange = "1-10"      // Pages 1 through 10
options.PageRange = "5"         // Only page 5
options.PageRange = "1,3,5-7"   // Pages 1, 3, 5, 6, 7
```

## Best Practices

### 1. Use Auto Mode for Unknown Documents

```go
resp, err := client.ParseDocumentAuto(fileData, filename)
```

### 2. Monitor Slow Documents

Check logs for warnings about slow processing:

```go
if duration > 120*time.Second {
    log.Printf("Consider chunking this document in future")
}
```

### 3. Adjust Chunk Size for Your Use Case

- **Small chunks (10-20 pages)**: Faster individual requests, more API calls
- **Large chunks (50+ pages)**: Fewer API calls, longer individual requests
- **Default (50 pages)**: Balanced approach

### 4. Handle Errors Gracefully

```go
resp, err := client.ParseDocumentWithRetry(fileData, filename, options)
if err != nil {
    if strings.Contains(err.Error(), "timeout") {
        // Consider chunking or reducing page range
        log.Println("Timeout - try chunking next time")
    } else if strings.Contains(err.Error(), "max retries") {
        // All retries exhausted
        log.Println("Max retries exceeded - document may be too complex")
    }
    return err
}
```

## API Functions

### Core Functions

- `ParseDocumentMultipart()` - Standard parse with retry (recommended)
- `ParseDocumentAuto()` - Automatic strategy selection
- `ParseDocumentAdvanced()` - Full control with custom options
- `ParseDocumentWithRetry()` - Parse with retry logic
- `ParseDocumentChunked()` - Chunked processing for large docs

### Helper Functions

- `DefaultParseOptions()` - Get recommended production settings
- `doRequestWithRetry()` - Generic retry wrapper for any ML Worker request

## Migration Guide

### Before (Old API)

```go
resp, err := client.ParseDocumentMultipart(fileData, filename, "application/pdf", true, true)
```

### After (New API - Same Function Signature!)

```go
// No changes needed! Function automatically uses retry and monitoring
resp, err := client.ParseDocumentMultipart(fileData, filename, "application/pdf", true, true)

// Or use new auto mode for even better handling
resp, err := client.ParseDocumentAuto(fileData, filename)
```

## Timeout Configuration

The default timeout has been increased from 30 seconds to 5 minutes:

- **Code default**: 300 seconds (5 minutes)
- **Environment override**: `ML_WORKER_TIMEOUT=300`
- **Per-request timeout**: Set via HTTP client

## Production Checklist

- [x] 5-minute timeout configured
- [x] Retry logic enabled (2 retries default)
- [x] Slow document monitoring active
- [x] Chunking available for large documents
- [x] Error handling with exponential backoff
- [x] Page range support for selective processing
- [x] Hybrid mode option for fallback
- [x] Image extraction control

## Troubleshooting

### "Timeout - max retries exceeded"

**Solution**: Document is too large. Use chunked processing:

```go
options := mlworker.DefaultParseOptions()
options.ChunkSize = 25  // Smaller chunks
resp, err := client.ParseDocumentChunked(fileData, filename, totalPages, options)
```

### "Slow parse warnings in logs"

**Solution**: Normal for complex documents. If consistent, consider:

1. Reducing `ExtractImages` to `false`
2. Using chunked processing
3. Limiting page range for testing

### "Server error (500)"

**Solution**: Automatically retried. If persistent, check ML Worker health:

```go
err := client.TestConnection()
```

## Performance Metrics

All parsing operations now log detailed metrics:

```
[MLWorker] Document parsed - report.pdf: 85 pages, 124000 chars, 15 tables (143.2s)
```

Use these metrics to:

- Identify slow documents
- Optimize chunk size
- Adjust timeout values
- Monitor ML Worker performance

## Support

For issues or questions:

1. Check ML Worker logs for detailed error messages
2. Enable DEBUG logging in .env: `DEBUG_SITE=true`
3. Review slow parse warnings for optimization opportunities
4. Contact ML Worker team for API issues

## Version History

- **v2.0** - Added retry logic, chunked processing, monitoring
- **v1.0** - Initial ML Worker integration
