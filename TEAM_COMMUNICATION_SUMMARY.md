# Team Communication System - Summary

## Overview

This document summarizes the team communication systems implemented for AsKara and ML-Worker integration.

## What Was Accomplished

### 1. AsKara Feedback API (Implemented & Working)
Created a complete feedback system for AsKara team communication:

**Endpoints**:
- `POST /api/feedback` - Submit feedback (feature requests, bug reports)
- `GET /api/feedback` - List all feedback
- `GET /api/feedback/{id}` - View specific feedback

**Files Created**:
- `vault-web-server/postapi/feedback.go` - Feedback handlers
- `submit_feedback.ps1` - Generic feedback submission script
- `submit_ml_worker_feedback.ps1` - ML-Worker specific submission script
- `FEEDBACK_SYSTEM.md` - Complete documentation

**Storage**:
- Feedback stored as markdown in `./data/feedback/`
- Persisted via Docker volumes
- Easily reviewed by AsKara team

**Status**: ✅ Fully implemented and tested

### 2. ML-Worker Feature Request Document
Created comprehensive feature request document:

**File**: `ML_WORKER_FEATURE_REQUESTS.md`

**Contents**:
- Current limitations (337KB PDF data URL limit, no multipart upload)
- 5 prioritized feature requests with API examples
- Backwards-compatible implementation suggestions
- Expected accuracy improvements (70%+ for complex PDFs)

**Top Priorities**:
1. **Multipart file upload** - Solves immediate 337KB PDF issue
2. **Hybrid mode** - Auto OCR fallback for poor quality text
3. **Chunked processing** - Handle 100+ page documents
4. **Quality confidence metrics** - Better fallback decisions
5. **Streaming/progressive parsing** - UX improvements

### 3. ML-Worker Team Communication API
Per your information, ML-Worker has team communication endpoints:

**Expected Endpoints** (from user):
- `POST /docs/team/message` - Submit messages to ML Worker team
- `GET /docs/team/messages` - List all messages
- `GET /docs/team/message/{id}` - View specific message
- `GET /docs/team/message/{id}/raw` - Get raw markdown
- `GET /docs/team/communication-info` - Communication statistics

**Files Created**:
- `submit_to_mlworker_team.ps1` - Direct submission to ML-Worker team API
- `ML_WORKER_COMMUNICATION_GUIDE.md` - Usage documentation

**Current Status**:
- Endpoints return 404 on current ML-Worker instance (http://192.168.0.250:6161)
- May be planned/upcoming feature or different version
- Script ready to use once endpoints are available

### 4. Enhanced PDF Processing
Implemented ML-Worker OCR fallback for poor quality PDF text:

**File**: `chunk/ml_pdf_processor.go`

**Functionality**:
- Uses `pdftoppm` to convert PDF pages to images
- Sends each page to ML-Worker OCR endpoint
- Combines results into full text
- Falls back when docconv produces garbled text

**Extraction Flow**:
1. Try ML-Worker parse (Docling) - skipped if file > 2000 chars base64
2. Try docconv (pdftotext) - fast text extraction
3. **NEW**: If docconv quality is poor → try ML-Worker OCR page-by-page
4. Only accept poor docconv text as last resort

## How to Submit Feature Requests

### Option 1: ML-Worker Team Communication API (When Available)

```powershell
cd Askara
powershell -ExecutionPolicy Bypass -File submit_to_mlworker_team.ps1
```

This submits directly to ML-Worker's team communication system at:
`POST http://192.168.0.250:6161/docs/team/message`

**Benefits**:
- Direct communication to ML-Worker team
- Centralized message tracking
- Built-in viewing/listing capabilities

**Current Status**: Endpoints return 404, may be pending implementation

### Option 2: AsKara Feedback API (Current Working Solution)

```powershell
cd Askara
powershell -ExecutionPolicy Bypass -File submit_ml_worker_feedback.ps1
```

This submits to AsKara's feedback API:
`POST http://localhost:8100/api/feedback`

**Benefits**:
- Working right now
- AsKara team can review and forward
- Serves as backup/archive

**Storage**: `./data/feedback/feature_request_*.md`

### Option 3: Direct Communication

Share the `ML_WORKER_FEATURE_REQUESTS.md` document via:
- Email to ML-Worker team
- Slack/Teams channels
- GitHub issues (if public repo exists)
- Shared documentation system

## Verification & Testing

### Test AsKara Feedback API:

```bash
# Submit test feedback
curl -X POST http://localhost:8100/api/feedback \
  -H "Content-Type: application/json" \
  -d '{
    "type": "feature_request",
    "title": "Test Feature Request",
    "content": "# Test\nThis is a test request."
  }'

# List all feedback
curl http://localhost:8100/api/feedback | python -m json.tool

# View specific feedback
curl http://localhost:8100/api/feedback/{id}
```

### Test ML-Worker Team API (When Available):

```bash
# Submit message
curl -X POST http://192.168.0.250:6161/docs/team/message \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Feature Request",
    "content": "...",
    "type": "feature_request"
  }'

# List messages
curl http://192.168.0.250:6161/docs/team/messages

# View specific message
curl http://192.168.0.250:6161/docs/team/message/{id}
```

### Test Enhanced PDF Processing:

```bash
# Upload a complex PDF (337KB+) and check logs
docker logs askara-web --tail 50

# Look for:
# [ExtractTextFromPDF] Docconv text quality is poor, trying ML Worker OCR
# [extractTextWithOCRLocal] Using ML Worker OCR with pdftoppm
# [extractTextWithOCRLocal] Processing X pages with ML Worker OCR
# [extractTextWithOCRLocal] Page 1: OCR success (95.2% confidence)
```

## File Organization

```
Askara/
├── ML_WORKER_FEATURE_REQUESTS.md         # Feature request document
├── FEEDBACK_SYSTEM.md                     # AsKara feedback API docs
├── ML_WORKER_COMMUNICATION_GUIDE.md       # ML-Worker communication guide
├── TEAM_COMMUNICATION_SUMMARY.md          # This file
├── submit_feedback.ps1                    # Generic feedback submission
├── submit_ml_worker_feedback.ps1          # Submit to AsKara API
├── submit_to_mlworker_team.ps1            # Submit to ML-Worker team API
├── test_feedback.ps1                      # Test AsKara feedback API
├── vault-web-server/
│   └── postapi/
│       └── feedback.go                    # Feedback endpoint implementation
├── chunk/
│   ├── fileprocessing.go                  # Enhanced PDF extraction logic
│   └── ml_pdf_processor.go                # ML-Worker OCR integration
└── data/
    └── feedback/                          # Stored feedback files
        └── feature_request_*.md
```

## Next Steps

1. **Verify ML-Worker Team Endpoints**:
   - Check if `/docs/team/message` endpoints are available
   - May need ML-Worker version update or configuration

2. **Submit Feature Requests**:
   - Use working AsKara API as immediate solution
   - Use ML-Worker team API when available
   - Or forward document via direct communication

3. **Monitor Implementation**:
   - Track ML-Worker API updates for new features
   - Test multipart upload support when added
   - Validate hybrid mode implementation

4. **Test PDF Processing**:
   - Upload 337KB Safeco PDF
   - Verify ML-Worker OCR is triggered for poor quality text
   - Compare answer accuracy vs. previous garbled text

## API Endpoints Summary

### AsKara (localhost:8100)
```
POST   /api/feedback              - Submit feedback
GET    /api/feedback              - List all feedback
GET    /api/feedback/{id}         - Get specific feedback
```

### ML-Worker (192.168.0.250:6161)
```
# Document Processing
POST   /askara/parse              - Parse documents (Docling)
POST   /askara/ocr                - OCR text extraction
POST   /askara/enhance            - Image enhancement
POST   /askara/caption            - Image captioning

# Team Communication (Expected)
POST   /docs/team/message         - Submit team message
GET    /docs/team/messages        - List messages
GET    /docs/team/message/{id}    - View message
GET    /docs/team/message/{id}/raw - Get raw markdown
GET    /docs/team/communication-info - Stats

# Documentation
GET    /docs/askara/list          - List docs
GET    /docs/askara/view/{file}   - View doc
GET    /docs/askara/raw/{file}    - Get raw doc
```

## Contact Information

- **AsKara API**: http://localhost:8100
- **AsKara Feedback**: http://localhost:8100/api/feedback
- **ML-Worker API**: http://192.168.0.250:6161
- **ML-Worker Docs**: http://192.168.0.250:6161/docs
- **ML-Worker Swagger**: http://192.168.0.250:6161/docs (interactive)

## Success Criteria

- ✅ AsKara feedback API implemented and working
- ✅ ML-Worker feature request document created
- ✅ Enhanced PDF processing with OCR fallback
- ✅ Scripts for submitting to both systems
- ✅ Comprehensive documentation
- ⏳ ML-Worker team communication endpoints (pending verification)
- ⏳ Feature requests submitted to ML-Worker team (pending)

## Conclusion

We've created a comprehensive communication infrastructure that allows AsKara and ML-Worker teams to exchange feature requests, bug reports, and general feedback. The system is designed to work with both:

1. **AsKara's local feedback API** (working now)
2. **ML-Worker's team communication API** (ready to use when available)

All documentation and scripts are in place. The next step is to verify ML-Worker's team communication endpoints are available and submit the feature requests through the appropriate channel.
