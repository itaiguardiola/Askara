# Verify ML-Worker Feature Implementation
# Testing the 4 implemented feature requests

$mlworkerBase = "http://192.168.0.250:6161"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "ML-Worker Feature Verification" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Feature 1: Multipart File Upload
Write-Host "[Feature 1/5] Multipart File Upload Support" -ForegroundColor White
Write-Host "Testing if /askara/parse accepts multipart/form-data..." -ForegroundColor Gray

$testFile = "test_code_document.txt"
if (Test-Path $testFile) {
    try {
        # Create multipart form data
        $boundary = [System.Guid]::NewGuid().ToString()
        $fileContent = Get-Content $testFile -Raw

        $bodyLines = @(
            "--$boundary",
            "Content-Disposition: form-data; name=`"document`"; filename=`"$testFile`"",
            "Content-Type: text/plain",
            "",
            $fileContent,
            "--$boundary",
            "Content-Disposition: form-data; name=`"preserve_structure`"",
            "",
            "true",
            "--$boundary--"
        )

        $body = $bodyLines -join "`r`n"

        $response = Invoke-WebRequest -Uri "$mlworkerBase/askara/parse" `
            -Method POST `
            -Body $body `
            -ContentType "multipart/form-data; boundary=$boundary" `
            -ErrorAction Stop

        Write-Host "  SUCCESS! Multipart upload is supported" -ForegroundColor Green
        Write-Host "  Status: $($response.StatusCode)" -ForegroundColor Gray
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 422) {
            Write-Host "  NOT SUPPORTED - Still expects JSON with document_url" -ForegroundColor Yellow
        } else {
            Write-Host "  ERROR (Status: $statusCode)" -ForegroundColor Red
        }
    }
} else {
    Write-Host "  SKIPPED - test file not found" -ForegroundColor Yellow
}
Write-Host ""

# Feature 2: Hybrid Mode (OCR + Text)
Write-Host "[Feature 2/5] Hybrid Mode (Auto OCR Fallback)" -ForegroundColor White
Write-Host "Checking if parse request accepts hybrid_mode parameter..." -ForegroundColor Gray

try {
    $testPayload = @{
        document_url = "data:application/pdf;base64,JVBERi0xLjQKJeLjz9MK"  # Tiny PDF header
        preserve_structure = $true
        hybrid_mode = $true
        quality_threshold = 0.6
    } | ConvertTo-Json

    $response = Invoke-WebRequest -Uri "$mlworkerBase/askara/parse" `
        -Method POST `
        -Body $testPayload `
        -ContentType "application/json" `
        -ErrorAction Stop

    Write-Host "  SUCCESS! Hybrid mode parameter accepted" -ForegroundColor Green
} catch {
    $errorBody = $_.ErrorDetails.Message
    if ($errorBody -like "*hybrid_mode*" -or $errorBody -like "*Extra inputs are not permitted*") {
        Write-Host "  NOT IMPLEMENTED - hybrid_mode parameter not recognized" -ForegroundColor Yellow
    } else {
        Write-Host "  Could not determine (invalid test data)" -ForegroundColor Gray
    }
}
Write-Host ""

# Feature 3: Chunked/Page-Range Processing
Write-Host "[Feature 3/5] Chunked Processing (Page Ranges)" -ForegroundColor White
Write-Host "Checking if parse request accepts page_range parameter..." -ForegroundColor Gray

try {
    $testPayload = @{
        document_url = "data:application/pdf;base64,JVBERi0xLjQKJeLjz9MK"
        page_range = "1-10"
    } | ConvertTo-Json

    $response = Invoke-WebRequest -Uri "$mlworkerBase/askara/parse" `
        -Method POST `
        -Body $testPayload `
        -ContentType "application/json" `
        -ErrorAction Stop

    Write-Host "  SUCCESS! Page range parameter accepted" -ForegroundColor Green
} catch {
    $errorBody = $_.ErrorDetails.Message
    if ($errorBody -like "*page_range*" -or $errorBody -like "*Extra inputs are not permitted*") {
        Write-Host "  NOT IMPLEMENTED - page_range parameter not recognized" -ForegroundColor Yellow
    } else {
        Write-Host "  Could not determine (invalid test data)" -ForegroundColor Gray
    }
}
Write-Host ""

# Feature 4: Quality Confidence Metrics
Write-Host "[Feature 4/5] Quality Confidence Metrics" -ForegroundColor White
Write-Host "Checking if parse response includes quality metrics..." -ForegroundColor Gray

# Use a valid small text file
$smallDoc = "data:text/plain;base64," + [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes("Test document"))

try {
    $testPayload = @{
        document_url = $smallDoc
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$mlworkerBase/askara/parse" `
        -Method POST `
        -Body $testPayload `
        -ContentType "application/json" `
        -ErrorAction Stop

    if ($response.PSObject.Properties.Name -contains "quality_metrics" -or
        $response.PSObject.Properties.Name -contains "text_confidence" -or
        $response.PSObject.Properties.Name -contains "extraction_method") {
        Write-Host "  SUCCESS! Quality metrics found in response" -ForegroundColor Green
        Write-Host "  Metrics: $($response.quality_metrics | ConvertTo-Json -Compress)" -ForegroundColor Gray
    } else {
        Write-Host "  NOT IMPLEMENTED - No quality metrics in response" -ForegroundColor Yellow
        Write-Host "  Available fields: $($response.PSObject.Properties.Name -join ', ')" -ForegroundColor Gray
    }
} catch {
    Write-Host "  ERROR: Could not test ($($_.Exception.Message))" -ForegroundColor Red
}
Write-Host ""

# Feature 5: Streaming/Progressive Parsing
Write-Host "[Feature 5/5] Streaming/Progressive Parsing (SSE)" -ForegroundColor White
Write-Host "Checking for /askara/parse/stream endpoint..." -ForegroundColor Gray

try {
    $response = Invoke-WebRequest -Uri "$mlworkerBase/askara/parse/stream" `
        -Method POST `
        -Body '{"document_url": "test"}' `
        -ContentType "application/json" `
        -ErrorAction Stop

    Write-Host "  SUCCESS! Streaming endpoint exists" -ForegroundColor Green
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    if ($statusCode -eq 404) {
        Write-Host "  NOT IMPLEMENTED - Streaming endpoint not found" -ForegroundColor Yellow
    } else {
        Write-Host "  ENDPOINT EXISTS (Status: $statusCode)" -ForegroundColor Green
    }
}
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Verification Complete" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
