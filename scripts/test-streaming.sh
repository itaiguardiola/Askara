#!/bin/bash

# Askara Streaming API Integration Test
# This script tests the streaming endpoint to verify SSE functionality

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${ASKARA_BASE_URL:-http://localhost:8100}"
STREAM_ENDPOINT="$BASE_URL/api/questions/stream"
REGULAR_ENDPOINT="$BASE_URL/api/questions"

echo -e "${YELLOW}=== Askara Streaming Integration Test ===${NC}\n"

# Function to print test status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $2"
    else
        echo -e "${RED}✗ FAIL${NC}: $2"
    fi
}

# Test 1: Check if server is running
echo "Test 1: Checking if server is running..."
if curl -s -o /dev/null -w "%{http_code}" "$BASE_URL" | grep -q "200\|302"; then
    print_status 0 "Server is running at $BASE_URL"
else
    print_status 1 "Server is not running at $BASE_URL"
    echo -e "${RED}Please start the server with 'npm start' before running tests${NC}"
    exit 1
fi

echo ""

# Test 2: Test streaming endpoint with missing parameters
echo "Test 2: Testing streaming endpoint validation (missing question)..."
RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$STREAM_ENDPOINT" \
    -F "model=GPT Turbo" \
    -F "uuid=test-uuid-123")

HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)

if [ "$HTTP_CODE" != "200" ]; then
    print_status 0 "Validation correctly rejected missing question (HTTP $HTTP_CODE)"
else
    print_status 1 "Should have rejected request with missing question"
fi

echo ""

# Test 3: Test streaming endpoint headers
echo "Test 3: Testing SSE headers on streaming endpoint..."
HEADERS=$(curl -s -I -X POST "$STREAM_ENDPOINT" \
    -F "question=What is AI?" \
    -F "model=GPT Turbo" \
    -F "uuid=test-uuid-123" 2>&1)

if echo "$HEADERS" | grep -qi "content-type.*text/event-stream"; then
    print_status 0 "SSE Content-Type header is set correctly"
else
    print_status 1 "SSE Content-Type header is missing or incorrect"
fi

if echo "$HEADERS" | grep -qi "cache-control.*no-cache"; then
    print_status 0 "Cache-Control header is set correctly"
else
    print_status 1 "Cache-Control header is missing or incorrect"
fi

echo ""

# Test 4: Test regular (non-streaming) endpoint
echo "Test 4: Testing regular question endpoint..."
echo -e "${YELLOW}Note: This will fail without valid OpenAI API keys${NC}"

# Note: This test will likely fail without proper API keys, but we can check the structure
RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "$REGULAR_ENDPOINT" \
    -F "question=What is 2+2?" \
    -F "model=GPT Turbo" \
    -F "uuid=test-uuid-123")

HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)

if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "500" ]; then
    print_status 0 "Regular endpoint is reachable (HTTP $HTTP_CODE)"
else
    print_status 1 "Regular endpoint returned unexpected status (HTTP $HTTP_CODE)"
fi

echo ""

# Summary
echo -e "${YELLOW}=== Test Summary ===${NC}"
echo -e "Tests check:"
echo -e "  • Server availability"
echo -e "  • Input validation"
echo -e "  • SSE headers configuration"
echo -e "  • Endpoint accessibility"
echo ""
echo -e "${GREEN}Note:${NC} Full streaming tests require valid OpenAI API keys."
echo -e "      Set OPENAI_API_KEY environment variable or create secret/openai_api_key file."
echo ""
echo -e "${YELLOW}To test with a real question:${NC}"
echo -e "  curl -N -X POST $STREAM_ENDPOINT \\"
echo -e "    -F \"question=What is artificial intelligence?\" \\"
echo -e "    -F \"model=GPT Turbo\" \\"
echo -e "    -F \"uuid=test-\$(uuidgen)\""
echo ""
