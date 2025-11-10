# AsKara Feedback System

## Overview

AsKara now includes a feedback API endpoint that allows users to submit feature requests, bug reports, and general feedback. All feedback is stored as markdown files in `./data/feedback/` for easy review by the AsKara team.

## API Endpoints

### 1. Submit Feedback
**Endpoint**: `POST /api/feedback`

**Request Body**:
```json
{
  "type": "feature_request",  // or "bug_report", "general_feedback"
  "title": "Your feedback title",
  "content": "Detailed markdown content",
  "metadata": {
    "version": "1.0",
    "platform": "Docker",
    "any_custom_field": "value"
  }
}
```

**Response**:
```json
{
  "success": true,
  "message": "Feedback submitted successfully",
  "id": "feature_request_20251109_002301_289",
  "timestamp": "2025-11-09T00:23:01Z"
}
```

### 2. List All Feedback
**Endpoint**: `GET /api/feedback`

**Response**:
```json
[
  {
    "id": "feature_request_20251109_002301_289",
    "type": "feature",
    "filename": "feature_request_20251109_002301_289.md",
    "timestamp": "2025-11-09T00:23:01Z",
    "size": 5420
  }
]
```

### 3. Get Specific Feedback
**Endpoint**: `GET /api/feedback/{id}`

**Response**: Returns the markdown file content

## Using the Feedback System

### Method 1: Direct API Call with curl

```bash
curl -X POST http://localhost:8100/api/feedback \
  -H "Content-Type: application/json" \
  -d '{
    "type": "feature_request",
    "title": "Add new feature",
    "content": "# Feature Description\nDetailed description here..."
  }'
```

### Method 2: PowerShell Script

Use the provided `submit_feedback.ps1` script:

```powershell
.\submit_feedback.ps1 `
  -Type "feature_request" `
  -Title "My Feature Request" `
  -File "path/to/markdown/file.md"
```

### Method 3: Submit from Markdown File

For the ML-Worker feature requests:

```powershell
# Read markdown file content
$content = Get-Content "ML_WORKER_FEATURE_REQUESTS.md" -Raw

# Create payload
$payload = @{
    type = "feature_request"
    title = "ML-Worker Integration Improvements"
    content = $content
    metadata = @{
        priority = "High"
        component = "ML-Worker"
    }
} | ConvertTo-Json -Depth 10

# Submit
Invoke-RestMethod -Uri "http://localhost:8100/api/feedback" `
    -Method POST `
    -Body $payload `
    -ContentType "application/json"
```

## Stored Feedback Format

Feedback is saved as markdown files in `/app/data/feedback/` (inside Docker container) or `./data/feedback/` (on host).

Example filename: `feature_request_20251109_002301_289.md`

Example content:
```markdown
# ML-Worker Integration Improvements

**Type**: feature_request
**Submitted**: 2025-11-09 00:23:01 PST
**ID**: feature_request_20251109_002301_289

[Original markdown content here]

---

## Metadata
- **priority**: High
- **component**: ML-Worker
- **version**: 1.0
```

## Storage Location

- **Docker container**: `/app/data/feedback/`
- **Host machine** (via volume mount): `./data/feedback/`

## Review Process

1. AsKara team can list all feedback via API:
   ```bash
   curl http://localhost:8100/api/feedback | python -m json.tool
   ```

2. Read specific feedback:
   ```bash
   curl http://localhost:8100/api/feedback/feature_request_20251109_002301_289
   ```

3. Or access files directly:
   ```bash
   docker exec askara-web cat /app/data/feedback/feature_request_20251109_002301_289.md
   ```

## ML-Worker Feature Request

The comprehensive ML-Worker feature request document has been created at:
- `ML_WORKER_FEATURE_REQUESTS.md`

To submit it to the AsKara team feedback system:

```powershell
cd Askara
powershell -ExecutionPolicy Bypass -File submit_ml_worker_feedback.ps1
```

Or using curl:
```bash
# Note: Escape quotes properly for your shell
curl -X POST http://localhost:8100/api/feedback \
  -H "Content-Type: application/json" \
  -d @- <<'EOF'
{
  "type": "feature_request",
  "title": "ML-Worker Integration Improvements for AsKara",
  "content": "...[paste ML_WORKER_FEATURE_REQUESTS.md content]...",
  "metadata": {
    "priority": "High",
    "component": "ML-Worker Integration"
  }
}
EOF
```

## Integration Points

The feedback system integrates seamlessly with AsKara:
- Same API patterns as other endpoints
- Stored alongside document data in `/app/data/`
- Accessible via Docker volumes for persistence
- No external dependencies required

## Security Notes

- Feedback endpoint accepts any POST request (no authentication)
- Suitable for internal team communication
- For production deployments, consider adding authentication
- All feedback is stored locally on the AsKara server

## Future Enhancements

Potential improvements to the feedback system:
- Email notifications when feedback is submitted
- Web UI for reviewing feedback
- Categorization and tagging
- Status tracking (open, in-progress, resolved)
- Integration with GitHub Issues
- Export to external systems
