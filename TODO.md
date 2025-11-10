# AsKara TODO List

Code review completed with devstral on 2025-11-08.

## Critical Priority (Before Production)

### Security
- [ ] Add authentication/authorization middleware to all API endpoints
  - Implement JWT or API key authentication
  - Protect: `/api/questions`, `/upload`, `/api/documents`, `/api/feedback`, etc.
  - Location: `vault-web-server/main.go:138-164`

- [x] Upgrade Go from 1.17 to 1.21+
  - Update `go.mod:3` - DONE
  - Test all dependencies for compatibility - DONE
  - Build verified successfully - DONE
  - Fixed import case sensitivity issues in unified LLM adapters

- [ ] Fix SSRF vulnerability in ML Worker client
  - Add URL validation in `mlworker/client.go:140, 326-332, 380`
  - Whitelist allowed ML Worker endpoints
  - Validate all external URLs before HTTP requests

- [ ] Configure proper CORS policies
  - Replace wildcard `"*"` in `vault-web-server/postapi/questions.go:220`
  - Set specific allowed origins
  - Use environment variable for CORS configuration

- [ ] Implement rate limiting
  - Add rate limiting middleware using `golang.org/x/time/rate`
  - Configure per-IP and per-API-key limits
  - Apply to all public endpoints

- [ ] Remove API key logging
  - Remove plaintext logging in `vault-web-server/postapi/questions.go:37`
  - Redact sensitive data from all log statements
  - Review all log statements for sensitive information

## High Priority

### Security Hardening
- [ ] Add request size validation
  - Implement max request body size middleware
  - Prevent memory exhaustion attacks
  - Configure appropriate limits per endpoint type

- [ ] Sanitize error messages
  - Replace detailed errors with generic messages for clients
  - Keep detailed logging server-side only
  - Locations: `main.go:61, 71, 85, 95` and throughout

- [ ] Enforce TLS/HTTPS
  - Add HTTP to HTTPS redirect middleware
  - Don't rely on port 443 check only (`main.go:173-180`)
  - Consider using Let's Encrypt for certificates

- [ ] Sanitize file upload filenames
  - Validate file extensions
  - Remove path traversal characters
  - Use UUIDs for storage filenames
  - Validate MIME types

- [ ] Review hardcoded endpoints
  - Remove hardcoded ML Worker fallback in `mlworker/client.go:140`
  - Ensure all endpoints come from environment variables
  - Document required environment variables

## Medium Priority

### Performance & Reliability
- [ ] Add HTTP client timeouts
  - Set reasonable timeouts on all HTTP clients
  - Prevent hanging connections
  - Configure per-client based on expected response times

- [ ] Implement cache eviction policy
  - Add LRU or TTL eviction to cache in `qdrant/qdrant.go:57`
  - Prevent memory leaks
  - Monitor cache size and hit rates

- [ ] Add graceful shutdown handling
  - Implement signal handling (SIGTERM, SIGINT)
  - Drain in-flight requests before shutdown
  - Close database connections properly

- [ ] Add security headers
  - Content-Security-Policy
  - X-Frame-Options
  - X-Content-Type-Options
  - Strict-Transport-Security

### Code Quality
- [ ] Replace deprecated `ioutil` package
  - Update `chunk/fileprocessing.go:8`
  - Use `io`, `os`, and `bufio` instead
  - Check for other deprecated package usage

- [ ] Optimize string concatenation
  - Use `strings.Builder` in chunking logic
  - Location: `chunk/fileprocessing.go:91`
  - Profile and optimize hot paths

## Low Priority

### Testing & Documentation
- [ ] Increase test coverage
  - Current: Only 3 test files found
  - Target: 70%+ coverage
  - Focus on critical paths first (auth, file processing, LLM integration)

- [ ] Replace magic numbers with constants
  - `MaxTokensPerChunk = 1500`
  - `BATCH_SIZE = 500`
  - `CharsPerPage = 3000`
  - Create centralized constants package

- [ ] Standardize error handling
  - Define consistent error handling patterns
  - Use structured logging
  - Create error types for different categories

- [ ] Clean up TODO comments
  - Review all TODO comments in code
  - Either implement or create issues
  - Remove resolved TODOs

### Future Enhancements
- [ ] Add metrics and monitoring
  - Prometheus metrics endpoint
  - Request duration tracking
  - Error rate monitoring
  - LLM provider performance metrics

- [ ] Implement circuit breakers
  - For external service calls (LLM providers, ML Worker)
  - Prevent cascading failures
  - Add fallback mechanisms

- [ ] Add OpenAPI/Swagger documentation
  - Document all API endpoints
  - Include request/response schemas
  - Add authentication requirements

- [ ] Implement connection pooling
  - For HTTP clients
  - For database connections
  - Optimize resource usage

- [ ] Add request tracing
  - Distributed tracing with OpenTelemetry
  - Request ID propagation
  - Performance profiling

## Positive Aspects (Keep Doing)

- Modular architecture with clean separation of concerns
- Comprehensive LLM provider abstraction
- Hybrid search implementation (vector + full-text)
- ML Worker integration with multiple features
- Docker multi-stage build optimization
- Query rewriting for improved retrieval
- Streaming support for real-time responses
- Code trust system for enhanced grounding

## Notes

- AsKara is designed to be lightweight and leverage Ollama and ML-worker services
- Focus on modularity and measurability
- May use multiple instances of ML-worker for scalability
- Keep AsKara core as light as possible

## Review Information

- Review Date: 2025-11-08
- Review Tool: devstral via Zen MCP
- Files Examined: 14
- Issues Found: 21 (5 critical, 6 high, 6 medium, 4 low)
- Overall Assessment: Well-structured codebase requiring security hardening before production
