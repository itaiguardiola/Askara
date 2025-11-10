# AsKara Development Server Startup Script
# Sets environment variables and starts the vault-web-server

$env:LLM_PROVIDER="ollama"
$env:LLM_PROVIDER_MODE="single"
$env:LLM_METRICS_ENABLED="true"
$env:LLM_METRICS_PATH="./data/llm_metrics.json"

$env:OLLAMA_HOST="http://localhost:11434"
$env:OLLAMA_MODEL="llama2"
$env:OLLAMA_EMBEDDING_MODEL="nomic-embed-text"

$env:QDRANT_API_ENDPOINT="http://localhost:6333"
$env:PORT="8100"
$env:DEBUG_SITE="false"

$env:ML_WORKER_ENABLED="true"
$env:ML_WORKER_ENDPOINT="http://192.168.0.250:6161/askara"
$env:ML_WORKER_FEATURES="ocr,enhance,caption,parse"
$env:ML_WORKER_TIMEOUT="30"

$env:QUERY_REWRITE_ENABLED="false"
$env:CODE_TRUST_ENABLED="true"
$env:CODE_TRUST_PATH="./data/code_trust.json"

Write-Host "Starting AsKara with Ollama provider..." -ForegroundColor Green
Write-Host "LLM Provider: $env:LLM_PROVIDER" -ForegroundColor Cyan
Write-Host "Ollama Host: $env:OLLAMA_HOST" -ForegroundColor Cyan
Write-Host "Port: $env:PORT" -ForegroundColor Cyan

go run vault-web-server/main.go
