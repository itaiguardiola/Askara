# ML-Worker Communication Guide

## Overview

ML-Worker (http://192.168.0.250:6161) has a comprehensive API but **no direct feedback submission endpoint**. This guide explains how to communicate feature requests to the ML-Worker team.

## Available ML-Worker Endpoints

### Documentation System
- **GET** `/docs/askara/list` - List all shared documentation
- **GET** `/docs/askara/view/{filename}` - View documentation as HTML
- **GET** `/docs/askara/raw/{filename}` - Get raw markdown files
- **GET** `/docs/askara/download/{filename}` - Download documentation

### Askara Integration Endpoints
- **GET** `/askara/` - Get integration info and available features
- **POST** `/askara/ocr` - OCR text extraction
- **POST** `/askara/caption` - Image captioning
- **POST** `/askara/visual-qa` - Visual question answering
- **POST** `/askara/enhance` - Document image enhancement
- **POST** `/askara/parse` - Document parsing (Docling)
- **POST** `/askara/rerank` - Semantic document reranking
- **POST** `/askara/classify-document` - Document classification
- **POST** `/askara/batch-process` - Batch document processing

## How to Submit Feature Requests

### Option 1: Via AsKara's Feedback API (Recommended)

We've implemented a feedback API in AsKara that stores feature requests as markdown files:

```powershell
# Submit ML-Worker feature requests via AsKara
cd Askara
powershell -ExecutionPolicy Bypass -File submit_ml_worker_feedback.ps1
```

This saves to: `./data/feedback/feature_request_*.md`

The AsKara team can then:
1. Review the feedback via AsKara's API: `GET http://localhost:8100/api/feedback`
2. Forward to ML-Worker team through other channels (email, Slack, etc.)

### Option 2: Direct Communication

Since ML-Worker doesn't have a feedback endpoint, communicate directly with the ML-Worker team via:
- **Email** - Send the `ML_WORKER_FEATURE_REQUESTS.md` document
- **Slack/Teams** - Share the feature request document
- **GitHub Issues** - If ML-Worker has a public repo
- **Shared Documentation** - If you have write access to ML-Worker's `/docs/askara` directory

### Option 3: Use AsKara as Communication Bridge

AsKara's feedback API can serve as a central collection point:

1. **Submit to AsKara**:
   ```bash
   curl -X POST http://localhost:8100/api/feedback \
     -H "Content-Type: application/json" \
     -d '{
       "type": "feature_request",
       "title": "ML-Worker: Multipart Upload Support",
       "content": "[Feature request markdown]",
       "metadata": {
         "target": "ML-Worker",
         "priority": "High"
       }
     }'
   ```

2. **AsKara team reviews** via:
   ```bash
   curl http://localhost:8100/api/feedback | python -m json.tool
   ```

3. **Forward to ML-Worker team** through appropriate channels

## ML-Worker Current Capabilities

### Document Processing
- **Parse** (Docling) - Advanced PDF/DOCX parsing
  - Limitation: 2083 char data URL limit (blocks 337KB+ PDFs)
  - **Proposed**: Add multipart upload support
- **OCR** - Image to text extraction
- **Enhance** - Document image quality improvement
- **Classify** - Document type classification

### Visual Understanding
- **Caption** - Generate image descriptions
- **Visual-QA** - Answer questions about images
- **Rerank** - Semantic document reranking

### Batch Operations
- **Batch-process** - Process multiple pages asynchronously

## Current Integration Status

AsKara currently uses ML-Worker for:
1. **PDF Parsing** (large files skip due to size limit)
2. **OCR** (page-by-page for poor quality PDFs)
3. **Image Enhancement** (before OCR)
4. **Image Captioning** (enabled but not actively used)

Configured via `.env`:
```
ML_WORKER_ENABLED=true
ML_WORKER_ENDPOINT=http://192.168.0.250:6161/askara
ML_WORKER_FEATURES=ocr,enhance,caption,parse
ML_WORKER_TIMEOUT=30
```

## Feature Request Summary

The `ML_WORKER_FEATURE_REQUESTS.md` document contains:

### Priority 1: Multipart File Upload
- **Issue**: 337KB PDF exceeds data URL limit
- **Solution**: Add `multipart/form-data` support to `/askara/parse`
- **Impact**: Immediate fix for large PDF processing

### Priority 2: Hybrid Mode (OCR + Text)
- **Issue**: Some PDFs have garbled text extraction
- **Solution**: Auto-detect poor quality text, apply OCR automatically
- **Impact**: 70%+ better accuracy on complex PDFs

### Priority 3: Chunked Processing
- **Solution**: Add page-range parameter to `/askara/parse`
- **Impact**: Process 100+ page documents incrementally

### Priority 4: Quality Metrics
- **Solution**: Return confidence scores in parse response
- **Impact**: Better fallback decisions in AsKara

### Priority 5: Streaming Results
- **Solution**: Add SSE endpoint for progressive parsing
- **Impact**: Better UX for large documents

## Testing ML-Worker Integration

```bash
# Test OCR endpoint
curl -X POST http://192.168.0.250:6161/askara/ocr \
  -H "Content-Type: application/json" \
  -d '{
    "image_url": "data:image/png;base64,...",
    "language": "eng",
    "enhance": true
  }'

# Test parse endpoint (small file)
curl -X POST http://192.168.0.250:6161/askara/parse \
  -H "Content-Type: application/json" \
  -d '{
    "document_url": "data:application/pdf;base64,...",
    "preserve_structure": true,
    "extract_tables": true
  }'

# Check available features
curl http://192.168.0.250:6161/askara/
```

## Next Steps

1. **Review feature requests** in `ML_WORKER_FEATURE_REQUESTS.md`
2. **Submit to AsKara feedback API** using provided scripts
3. **Forward to ML-Worker team** through appropriate communication channels
4. **Test current integration** to validate feature needs
5. **Monitor AsKara logs** to track ML-Worker usage patterns

## Contact Points

- **ML-Worker API**: http://192.168.0.250:6161
- **ML-Worker Docs**: http://192.168.0.250:6161/docs
- **ML-Worker Swagger**: http://192.168.0.250:6161/docs (interactive API explorer)
- **AsKara Feedback API**: http://localhost:8100/api/feedback

## Viewing ML-Worker Documentation

```bash
# List all documentation
curl http://192.168.0.250:6161/docs/askara/list | python -m json.tool

# View specific doc (HTML)
curl http://192.168.0.250:6161/docs/askara/view/ASKARA_RERANK_API_DOCUMENTATION.md

# Download raw markdown
curl http://192.168.0.250:6161/docs/askara/raw/ASKARA_INTEGRATION_GUIDE.md
```

## AsKara-Specific Recommendations

Given ML-Worker's lack of feedback API, we recommend:

1. **Use AsKara's feedback API** as central collection point
2. **Tag feedback** with `"target": "ML-Worker"` in metadata
3. **Periodic review** of submitted ML-Worker feature requests
4. **Manual forwarding** to ML-Worker team via preferred channel
5. **Track implementation** by monitoring ML-Worker API updates

The feedback system we've built serves as a bridge between AsKara users and both AsKara and ML-Worker development teams.
