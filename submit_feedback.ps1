# Submit Feedback to AsKara Team
# Usage: .\submit_feedback.ps1 -Type "feature_request" -Title "ML-Worker Integration Improvements" -File "ML_WORKER_FEATURE_REQUESTS.md"

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("feature_request", "bug_report", "general_feedback")]
    [string]$Type,

    [Parameter(Mandatory=$true)]
    [string]$Title,

    [Parameter(Mandatory=$true)]
    [string]$File,

    [string]$AskaraURL = "http://localhost:8100"
)

# Read markdown file content
if (-not (Test-Path $File)) {
    Write-Error "File not found: $File"
    exit 1
}

$content = Get-Content $File -Raw

# Create request payload
$payload = @{
    type = $Type
    title = $Title
    content = $content
    metadata = @{
        version = "1.0"
        platform = "Windows"
        submitted_by = "AsKara User"
        submitted_at = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
    }
} | ConvertTo-Json -Depth 10

# Submit to API
try {
    Write-Host "Submitting feedback to $AskaraURL/api/feedback..." -ForegroundColor Cyan
    Write-Host "Type: $Type" -ForegroundColor Gray
    Write-Host "Title: $Title" -ForegroundColor Gray
    Write-Host "Content length: $($content.Length) characters" -ForegroundColor Gray
    Write-Host ""

    $response = Invoke-RestMethod -Uri "$AskaraURL/api/feedback" `
        -Method POST `
        -Body $payload `
        -ContentType "application/json"

    if ($response.success) {
        Write-Host "[SUCCESS] Feedback submitted successfully!" -ForegroundColor Green
        Write-Host "  ID: $($response.id)" -ForegroundColor Gray
        Write-Host "  Timestamp: $($response.timestamp)" -ForegroundColor Gray
        Write-Host "  Saved to: ./data/feedback/$($response.id).md" -ForegroundColor Gray
    } else {
        Write-Host "[FAILED] Failed to submit feedback: $($response.message)" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "[ERROR] Error submitting feedback: $_" -ForegroundColor Red
    exit 1
}
