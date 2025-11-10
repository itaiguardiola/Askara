$content = Get-Content "ML_WORKER_FEATURE_REQUESTS.md" -Raw

$payload = @{
    type = "feature_request"
    title = "ML-Worker Integration Improvements for AsKara"
    content = $content
    metadata = @{
        version = "1.0"
        platform = "Docker"
        component = "ML-Worker Integration"
        priority = "High"
        submitter = "AsKara Team"
    }
} | ConvertTo-Json -Depth 10

try {
    Write-Host "Submitting ML-Worker feature requests..." -ForegroundColor Cyan
    $response = Invoke-RestMethod -Uri "http://localhost:8100/api/feedback" `
        -Method POST `
        -Body $payload `
        -ContentType "application/json"

    Write-Host ""
    Write-Host "[SUCCESS] ML-Worker feature requests submitted!" -ForegroundColor Green
    Write-Host "  Feedback ID: $($response.id)" -ForegroundColor Gray
    Write-Host "  Timestamp: $($response.timestamp)" -ForegroundColor Gray
    Write-Host "  File: ./data/feedback/$($response.id).md" -ForegroundColor Gray
    Write-Host ""
    Write-Host "The AsKara team can now review your feature requests." -ForegroundColor White
} catch {
    Write-Host ""
    Write-Host "[ERROR] Failed to submit feedback" -ForegroundColor Red
    Write-Host "  Error: $_" -ForegroundColor Red
    exit 1
}
