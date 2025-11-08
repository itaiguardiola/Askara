package codesourcetrust

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"time"
)

// CodeParser extracts code elements from document text
type CodeParser struct {
	config ExtractionConfig
}

// NewCodeParser creates a new parser with the given configuration
func NewCodeParser(config ExtractionConfig) *CodeParser {
	return &CodeParser{config: config}
}

// ParseDocument extracts code elements from document text
func (p *CodeParser) ParseDocument(documentID, documentName, uuid, text string) (*DocumentEnrichment, error) {
	enrichment := &DocumentEnrichment{
		DocumentID:       documentID,
		UUID:             uuid,
		DocumentName:     documentName,
		Symbols:          []CodeSymbol{},
		CodeBlocks:       []CodeBlock{},
		APIEndpoints:     []APIEndpoint{},
		Concepts:         []string{},
		Keywords:         []string{},
		LastUpdated:      time.Now(),
		ProcessedVersion: "1.0.0",
	}

	// Extract code blocks
	if p.config.EnableCodeBlocks {
		enrichment.CodeBlocks = p.extractCodeBlocks(documentID, documentName, text)
	}

	// Extract symbols from code blocks and inline code
	if p.config.EnableSymbols {
		enrichment.Symbols = p.extractSymbols(documentID, documentName, text, enrichment.CodeBlocks)
	}

	// Extract API endpoints
	if p.config.EnableAPIs {
		enrichment.APIEndpoints = p.extractAPIEndpoints(documentID, documentName, text)
	}

	// Extract concepts and keywords
	if p.config.EnableConcepts {
		enrichment.Concepts = p.extractConcepts(text)
		enrichment.Keywords = p.extractKeywords(text)
	}

	return enrichment, nil
}

// extractCodeBlocks finds code blocks in markdown-style fenced blocks
func (p *CodeParser) extractCodeBlocks(documentID, documentName, text string) []CodeBlock {
	var blocks []CodeBlock

	// Match markdown code fences: ```language\ncode\n```
	// Also match indented code blocks (4 spaces or tab)
	fencePattern := regexp.MustCompile("(?s)```([a-zA-Z0-9]*)\n(.*?)```")
	matches := fencePattern.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		language := strings.ToLower(strings.TrimSpace(match[1]))
		code := strings.TrimSpace(match[2])

		// Skip if language not supported
		if len(p.config.SupportedLanguages) > 0 && language != "" {
			if !contains(p.config.SupportedLanguages, language) {
				continue
			}
		}

		// Skip if too short
		lines := strings.Split(code, "\n")
		if len(lines) < p.config.MinCodeBlockLines {
			continue
		}

		// Extract context around the code block
		codePos := strings.Index(text, match[0])
		contextStart := maxInt(0, codePos-p.config.MaxContextChars)
		contextEnd := minInt(len(text), codePos+len(match[0])+p.config.MaxContextChars)
		context := strings.TrimSpace(text[contextStart:codePos] + " " + text[codePos+len(match[0]):contextEnd])

		// Extract symbols used in this block
		symbols := p.extractSymbolsFromCode(code, language)

		// Generate unique ID
		hash := sha256.Sum256([]byte(documentID + code))
		blockID := hex.EncodeToString(hash[:])[:16]

		blocks = append(blocks, CodeBlock{
			ID:           blockID,
			DocumentID:   documentID,
			DocumentName: documentName,
			Language:     language,
			Code:         code,
			Context:      context,
			Symbols:      symbols,
			LineStart:    strings.Count(text[:codePos], "\n") + 1,
			LineEnd:      strings.Count(text[:codePos], "\n") + len(lines) + 1,
			CreatedAt:    time.Now(),
		})
	}

	return blocks
}

// extractSymbols finds function/class/method definitions
func (p *CodeParser) extractSymbols(documentID, documentName, text string, codeBlocks []CodeBlock) []CodeSymbol {
	var symbols []CodeSymbol

	// Extract from code blocks
	for _, block := range codeBlocks {
		blockSymbols := p.extractSymbolsFromCodeBlock(documentID, documentName, block)
		symbols = append(symbols, blockSymbols...)
	}

	// Extract inline code references (backtick code)
	inlinePattern := regexp.MustCompile("`([^`]+)`")
	inlineMatches := inlinePattern.FindAllStringSubmatch(text, -1)

	seenSymbols := make(map[string]bool)
	for _, match := range inlineMatches {
		inlineCode := match[1]

		// Look for function/method patterns
		if strings.Contains(inlineCode, "(") {
			symbolName := strings.Split(inlineCode, "(")[0]
			symbolName = strings.TrimSpace(symbolName)

			// Skip if already seen or too short
			if seenSymbols[symbolName] || len(symbolName) < 2 {
				continue
			}

			// Find context around this symbol
			pos := strings.Index(text, match[0])
			contextStart := maxInt(0, pos-200)
			contextEnd := minInt(len(text), pos+len(match[0])+200)
			context := strings.TrimSpace(text[contextStart:contextEnd])

			symbols = append(symbols, CodeSymbol{
				Symbol:       symbolName,
				Type:         "function",
				Language:     "unknown",
				DocumentID:   documentID,
				DocumentName: documentName,
				Context:      context,
				Signature:    inlineCode,
			})

			seenSymbols[symbolName] = true
		}
	}

	return symbols
}

// extractSymbolsFromCodeBlock extracts symbols from a code block
func (p *CodeParser) extractSymbolsFromCodeBlock(documentID, documentName string, block CodeBlock) []CodeSymbol {
	var symbols []CodeSymbol

	switch block.Language {
	case "go", "golang":
		symbols = append(symbols, p.extractGoSymbols(documentID, documentName, block)...)
	case "javascript", "js", "typescript", "ts":
		symbols = append(symbols, p.extractJSSymbols(documentID, documentName, block)...)
	case "python", "py":
		symbols = append(symbols, p.extractPythonSymbols(documentID, documentName, block)...)
	case "java":
		symbols = append(symbols, p.extractJavaSymbols(documentID, documentName, block)...)
	}

	return symbols
}

// extractGoSymbols extracts Go function and type definitions
func (p *CodeParser) extractGoSymbols(documentID, documentName string, block CodeBlock) []CodeSymbol {
	var symbols []CodeSymbol

	// Match function definitions: func FuncName(...) ...
	funcPattern := regexp.MustCompile(`func\s+(?:\([^)]*\)\s+)?([A-Za-z0-9_]+)\s*\([^)]*\)`)
	matches := funcPattern.FindAllStringSubmatch(block.Code, -1)

	for _, match := range matches {
		funcName := match[1]
		symbols = append(symbols, CodeSymbol{
			Symbol:       funcName,
			Type:         "function",
			Language:     "go",
			DocumentID:   documentID,
			DocumentName: documentName,
			Context:      block.Context,
			Signature:    strings.TrimSpace(match[0]),
		})
	}

	// Match type definitions: type TypeName ...
	typePattern := regexp.MustCompile(`type\s+([A-Za-z0-9_]+)\s+(?:struct|interface)`)
	typeMatches := typePattern.FindAllStringSubmatch(block.Code, -1)

	for _, match := range typeMatches {
		typeName := match[1]
		symbols = append(symbols, CodeSymbol{
			Symbol:       typeName,
			Type:         "type",
			Language:     "go",
			DocumentID:   documentID,
			DocumentName: documentName,
			Context:      block.Context,
			Signature:    strings.TrimSpace(match[0]),
		})
	}

	return symbols
}

// extractJSSymbols extracts JavaScript/TypeScript function and class definitions
func (p *CodeParser) extractJSSymbols(documentID, documentName string, block CodeBlock) []CodeSymbol {
	var symbols []CodeSymbol

	// Match function declarations: function funcName(...) or const funcName = (...) =>
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`function\s+([A-Za-z0-9_$]+)\s*\(`),
		regexp.MustCompile(`const\s+([A-Za-z0-9_$]+)\s*=\s*(?:async\s*)?\([^)]*\)\s*=>`),
		regexp.MustCompile(`(?:export\s+)?class\s+([A-Za-z0-9_$]+)`),
		regexp.MustCompile(`(?:export\s+)?interface\s+([A-Za-z0-9_$]+)`),
	}

	for i, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(block.Code, -1)
		symbolType := "function"
		if i == 2 {
			symbolType = "class"
		} else if i == 3 {
			symbolType = "interface"
		}

		for _, match := range matches {
			symbols = append(symbols, CodeSymbol{
				Symbol:       match[1],
				Type:         symbolType,
				Language:     block.Language,
				DocumentID:   documentID,
				DocumentName: documentName,
				Context:      block.Context,
				Signature:    strings.TrimSpace(match[0]),
			})
		}
	}

	return symbols
}

// extractPythonSymbols extracts Python function and class definitions
func (p *CodeParser) extractPythonSymbols(documentID, documentName string, block CodeBlock) []CodeSymbol {
	var symbols []CodeSymbol

	// Match function definitions: def func_name(...):
	funcPattern := regexp.MustCompile(`def\s+([A-Za-z0-9_]+)\s*\(`)
	matches := funcPattern.FindAllStringSubmatch(block.Code, -1)

	for _, match := range matches {
		symbols = append(symbols, CodeSymbol{
			Symbol:       match[1],
			Type:         "function",
			Language:     "python",
			DocumentID:   documentID,
			DocumentName: documentName,
			Context:      block.Context,
			Signature:    strings.TrimSpace(match[0]),
		})
	}

	// Match class definitions: class ClassName:
	classPattern := regexp.MustCompile(`class\s+([A-Za-z0-9_]+)(?:\([^)]*\))?:`)
	classMatches := classPattern.FindAllStringSubmatch(block.Code, -1)

	for _, match := range classMatches {
		symbols = append(symbols, CodeSymbol{
			Symbol:       match[1],
			Type:         "class",
			Language:     "python",
			DocumentID:   documentID,
			DocumentName: documentName,
			Context:      block.Context,
			Signature:    strings.TrimSpace(match[0]),
		})
	}

	return symbols
}

// extractJavaSymbols extracts Java class and method definitions
func (p *CodeParser) extractJavaSymbols(documentID, documentName string, block CodeBlock) []CodeSymbol {
	var symbols []CodeSymbol

	// Match method definitions: public/private/protected ... methodName(...)
	methodPattern := regexp.MustCompile(`(?:public|private|protected|static)\s+(?:\w+\s+)*?([A-Za-z0-9_]+)\s*\(`)
	matches := methodPattern.FindAllStringSubmatch(block.Code, -1)

	for _, match := range matches {
		symbols = append(symbols, CodeSymbol{
			Symbol:       match[1],
			Type:         "method",
			Language:     "java",
			DocumentID:   documentID,
			DocumentName: documentName,
			Context:      block.Context,
			Signature:    strings.TrimSpace(match[0]),
		})
	}

	// Match class definitions
	classPattern := regexp.MustCompile(`(?:public\s+)?class\s+([A-Za-z0-9_]+)`)
	classMatches := classPattern.FindAllStringSubmatch(block.Code, -1)

	for _, match := range classMatches {
		symbols = append(symbols, CodeSymbol{
			Symbol:       match[1],
			Type:         "class",
			Language:     "java",
			DocumentID:   documentID,
			DocumentName: documentName,
			Context:      block.Context,
			Signature:    strings.TrimSpace(match[0]),
		})
	}

	return symbols
}

// extractSymbolsFromCode extracts symbol names from code
func (p *CodeParser) extractSymbolsFromCode(code, language string) []string {
	var symbols []string
	seen := make(map[string]bool)

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`func\s+([A-Za-z0-9_]+)`),                     // Go functions
		regexp.MustCompile(`function\s+([A-Za-z0-9_$]+)`),                // JS functions
		regexp.MustCompile(`class\s+([A-Za-z0-9_$]+)`),                   // Classes
		regexp.MustCompile(`def\s+([A-Za-z0-9_]+)`),                      // Python functions
		regexp.MustCompile(`const\s+([A-Za-z0-9_$]+)\s*=`),               // Constants
		regexp.MustCompile(`(?:export\s+)?interface\s+([A-Za-z0-9_$]+)`), // Interfaces
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(code, -1)
		for _, match := range matches {
			symbol := match[1]
			if !seen[symbol] && len(symbol) > 1 {
				symbols = append(symbols, symbol)
				seen[symbol] = true
			}
		}
	}

	return symbols
}

// extractAPIEndpoints finds API endpoint definitions
func (p *CodeParser) extractAPIEndpoints(documentID, documentName, text string) []APIEndpoint {
	var endpoints []APIEndpoint
	seen := make(map[string]bool)

	// Common API patterns
	patterns := []*regexp.Regexp{
		// HTTP method followed by path: GET /api/users
		regexp.MustCompile(`(?i)(GET|POST|PUT|DELETE|PATCH)\s+(/[^\s\)}\]]+)`),
		// RESTful route definitions: router.get('/api/users', ...)
		regexp.MustCompile(`(?i)\.(?:get|post|put|delete|patch)\s*\(\s*['"]([^'"]+)['"]`),
		// Go HTTP handlers: http.HandleFunc("/api/...", ...)
		regexp.MustCompile(`(?i)Handle(?:Func)?\s*\(\s*"([^"]+)"`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(text, -1)

		for _, match := range matches {
			var method, path string

			if len(match) == 3 {
				// Pattern with method and path
				method = strings.ToUpper(match[1])
				path = match[2]
			} else if len(match) == 2 {
				// Pattern with only path
				method = "GET" // Default
				path = match[1]
			}

			key := method + " " + path
			if seen[key] {
				continue
			}

			// Extract context/description
			pos := strings.Index(text, match[0])
			contextStart := maxInt(0, pos-300)
			contextEnd := minInt(len(text), pos+len(match[0])+300)
			context := strings.TrimSpace(text[contextStart:contextEnd])

			// Extract description from surrounding text
			description := extractDescription(context)

			endpoints = append(endpoints, APIEndpoint{
				Method:       method,
				Path:         path,
				Description:  description,
				DocumentID:   documentID,
				DocumentName: documentName,
			})

			seen[key] = true
		}
	}

	return endpoints
}

// extractConcepts identifies high-level concepts from document
func (p *CodeParser) extractConcepts(text string) []string {
	var concepts []string
	seen := make(map[string]bool)

	// Common concept patterns in technical documentation
	conceptPatterns := []string{
		"authentication", "authorization", "security",
		"caching", "performance", "optimization",
		"database", "storage", "persistence",
		"API", "REST", "GraphQL", "gRPC",
		"testing", "deployment", "monitoring",
		"logging", "error handling", "validation",
		"configuration", "settings", "environment",
		"user management", "session", "cookie",
		"encryption", "hashing", "token",
		"microservice", "architecture", "design pattern",
	}

	textLower := strings.ToLower(text)

	for _, concept := range conceptPatterns {
		if strings.Contains(textLower, concept) && !seen[concept] {
			concepts = append(concepts, concept)
			seen[concept] = true
		}
	}

	return concepts
}

// extractKeywords extracts important technical terms
func (p *CodeParser) extractKeywords(text string) []string {
	// For now, use concept extraction as keywords
	// Could be enhanced with NLP/TF-IDF analysis
	return p.extractConcepts(text)
}

// Helper functions

func extractDescription(context string) string {
	// Try to find sentences that describe what something does
	sentences := strings.Split(context, ".")
	for _, sentence := range sentences {
		lower := strings.ToLower(sentence)
		if strings.Contains(lower, "returns") ||
			strings.Contains(lower, "handles") ||
			strings.Contains(lower, "creates") ||
			strings.Contains(lower, "retrieves") ||
			strings.Contains(lower, "deletes") ||
			strings.Contains(lower, "updates") {
			return strings.TrimSpace(sentence)
		}
	}

	// Return first sentence if no descriptive sentence found
	if len(sentences) > 0 {
		return strings.TrimSpace(sentences[0])
	}

	return ""
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
