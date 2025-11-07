# Askara Testing Guide

This document describes how to run tests for Askara.

## Test Structure

### Unit Tests (Go)

Backend unit tests are written in Go and located alongside the source code:

- `vault-web-server/postapi/questions_test.go` - Tests for question handlers
- `vault-web-server/postapi/openai_test.go` - Tests for OpenAI integration

### Integration Tests

- `scripts/test-streaming.sh` - Integration tests for streaming functionality

## Running Tests

### Backend Unit Tests

Run all Go unit tests:

```bash
go test ./...
```

Run tests for a specific package:

```bash
go test ./vault-web-server/postapi
```

Run tests with verbose output:

```bash
go test -v ./vault-web-server/postapi
```

Run tests with coverage:

```bash
go test -cover ./vault-web-server/postapi
```

Generate coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Integration Tests

The integration tests require a running server.

**Step 1: Start the server**

In one terminal:

```bash
npm start
```

**Step 2: Run integration tests**

In another terminal:

```bash
./scripts/test-streaming.sh
```

Set custom base URL:

```bash
ASKARA_BASE_URL=http://localhost:8100 ./scripts/test-streaming.sh
```

## Test Coverage

### What's Tested

#### Backend (Go)

✅ **Question Handlers**
- Form validation
- SSE header configuration
- Error handling for vector DB failures
- Both streaming and non-streaming modes

✅ **OpenAI Integration**
- Prompt building with various contexts
- Token limit handling
- StreamHandler functionality
- Helper functions (min, etc.)

#### Integration Tests

✅ **Streaming API**
- Server availability
- Endpoint validation
- SSE headers
- Error responses

### What's Not Tested

⚠️ **Requires Live API Keys**
- Actual OpenAI API calls
- Real vector database operations
- End-to-end streaming with real responses

These require valid API keys and are better suited for manual testing or CI/CD with secrets.

## Manual Testing

### Testing Streaming Endpoint

With curl:

```bash
curl -N -X POST http://localhost:8100/api/questions/stream \
  -F "question=What is artificial intelligence?" \
  -F "model=GPT Turbo" \
  -F "uuid=test-$(uuidgen)"
```

Expected output:
```
event: context
data: [{"Text":"...","Title":"..."}]

event: chunk
data: Artificial

event: chunk
data:  intelligence

event: chunk
data:  (AI)

...

event: done
data: {}
```

### Testing Regular Endpoint

```bash
curl -X POST http://localhost:8100/api/questions \
  -F "question=What is artificial intelligence?" \
  -F "model=GPT Turbo" \
  -F "uuid=test-$(uuidgen)"
```

Expected output (JSON):
```json
{
  "answer": "Artificial intelligence (AI) is...",
  "context": [...],
  "tokens": 150
}
```

### Testing Upload Endpoint

```bash
curl -X POST http://localhost:8100/upload \
  -F "files=@/path/to/document.pdf" \
  -F "uuid=test-$(uuidgen)"
```

## Continuous Integration

To integrate tests into CI/CD:

```yaml
# Example GitHub Actions workflow
- name: Run Go tests
  run: go test -v -cover ./...

- name: Start server
  run: npm start &

- name: Wait for server
  run: sleep 5

- name: Run integration tests
  run: ./scripts/test-streaming.sh
```

## Adding New Tests

### Adding a Go Unit Test

1. Create or update `*_test.go` file in the same package
2. Follow the naming convention: `TestFunctionName`
3. Use table-driven tests for multiple cases:

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name string
        input string
        want string
    }{
        {"case 1", "input1", "output1"},
        {"case 2", "input2", "output2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := MyFunction(tt.input)
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Adding an Integration Test

Add new test cases to `scripts/test-streaming.sh`:

```bash
echo "Test N: Description..."
# Test implementation
print_status $? "Test description"
echo ""
```

## Troubleshooting

### Tests Fail with "connection refused"

Make sure the server is running on the expected port:

```bash
npm start
```

### Tests Fail with OpenAI Errors

Some tests require valid API keys. Set them up:

```bash
echo "your_openai_api_key" > secret/openai_api_key
echo "your_pinecone_api_key" > secret/pinecone_api_key
echo "https://your-index.pinecone.io" > secret/pinecone_api_endpoint
```

Or use environment variables:

```bash
export OPENAI_API_KEY="your_key"
export PINECONE_API_KEY="your_key"
export PINECONE_API_ENDPOINT="https://your-index.pinecone.io"
```

### Coverage Reports Not Generated

Install Go tools:

```bash
go install golang.org/x/tools/cmd/cover@latest
```

## Best Practices

1. **Mock External Dependencies** - Use mocks for OpenAI API, vector DB
2. **Test Edge Cases** - Empty inputs, very long inputs, invalid data
3. **Test Error Paths** - Network failures, API errors, timeouts
4. **Keep Tests Fast** - Unit tests should run in milliseconds
5. **Avoid Test Interdependence** - Each test should be independent
6. **Use Descriptive Names** - Test names should describe what's being tested

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Server-Sent Events Spec](https://html.spec.whatwg.org/multipage/server-sent-events.html)
