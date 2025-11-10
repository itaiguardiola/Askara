# Submit Feature Requests to ML-Worker Team
# This script uses ML-Worker's team communication API
# Endpoints: POST /docs/team/message

$content = Get-Content "ML_WORKER_FEATURE_REQUESTS.md" -Raw

$payload = @{
    title = "AsKara Integration - Feature Requests for Improved PDF Processing"
    content = $content
    type = "feature_request"
    priority = "high"
    from = "AsKara Team"
    metadata = @{
        component = "Document Parsing / Askara Integration"
        date = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
        contact = "AsKara Development Team"
        related_endpoints = @("/askara/parse", "/askara/ocr")
    }
} | ConvertTo-Json -Depth 10

$mlworkerEndpoint = "http://192.168.0.250:6161/docs/team/message"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "ML-Worker Team Communication" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Submitting feature requests to ML-Worker team..." -ForegroundColor White
Write-Host "Endpoint: $mlworkerEndpoint" -ForegroundColor Gray
Write-Host "Document: ML_WORKER_FEATURE_REQUESTS.md" -ForegroundColor Gray
Write-Host "Content length: $($content.Length) characters" -ForegroundColor Gray
Write-Host ""

try {
    $response = Invoke-RestMethod -Uri $mlworkerEndpoint `
        -Method POST `
        -Body $payload `
        -ContentType "application/json" `
        -ErrorAction Stop

    Write-Host "[SUCCESS] Feature requests submitted successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Response Details:" -ForegroundColor White
    Write-Host "  Message ID: $($response.id)" -ForegroundColor Gray
    Write-Host "  Timestamp: $($response.timestamp)" -ForegroundColor Gray
    Write-Host "  Status: $($response.status)" -ForegroundColor Gray
    Write-Host ""
    Write-Host "The ML-Worker team can view this message at:" -ForegroundColor White
    Write-Host "  http://192.168.0.250:6161/docs/team/message/$($response.id)" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "List all messages:" -ForegroundColor White
    Write-Host "  http://192.168.0.250:6161/docs/team/messages" -ForegroundColor Cyan
    Write-Host ""

} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    $errorBody = $_.ErrorDetails.Message

    Write-Host "[ERROR] Failed to submit message" -ForegroundColor Red
    Write-Host ""
    Write-Host "Details:" -ForegroundColor White
    Write-Host "  Status Code: $statusCode" -ForegroundColor Gray
    Write-Host "  Error Message: $_" -ForegroundColor Gray

    if ($errorBody) {
        Write-Host "  Response Body: $errorBody" -ForegroundColor Gray
    }

    Write-Host ""

    if ($statusCode -eq 404) {
        Write-Host "NOTE: The team communication endpoints may not be implemented yet." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Alternative: Submit via AsKara's feedback API instead:" -ForegroundColor Yellow
        Write-Host "  powershell -ExecutionPolicy Bypass -File submit_ml_worker_feedback.ps1" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "Or contact ML-Worker team directly with the document:" -ForegroundColor Yellow
        Write-Host "  ML_WORKER_FEATURE_REQUESTS.md" -ForegroundColor Cyan
    }

    exit 1
}
