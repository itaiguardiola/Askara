package queryrewriter

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/itaiguardiola/askara/llm"
)

// QueryRewriter handles query expansion and rewriting
type QueryRewriter struct {
	provider llm.LLMProvider
	enabled  bool
}

// RewriteResult contains the original query and its variations
type RewriteResult struct {
	Original   string   `json:"original"`
	Variations []string `json:"variations"`
}

// NewQueryRewriter creates a new query rewriter instance
func NewQueryRewriter(provider llm.LLMProvider, enabled bool) *QueryRewriter {
	return &QueryRewriter{
		provider: provider,
		enabled:  enabled,
	}
}

// IsEnabled returns whether query rewriting is enabled
func (qr *QueryRewriter) IsEnabled() bool {
	return qr.enabled
}

// RewriteQuery expands a user query into multiple semantic variations
// This improves retrieval by generating alternative phrasings
func (qr *QueryRewriter) RewriteQuery(query string) (*RewriteResult, error) {
	if !qr.enabled {
		return &RewriteResult{
			Original:   query,
			Variations: []string{query},
		}, nil
	}

	log.Printf("[QueryRewriter] Rewriting query: %s", query)

	// Create a simpler prompt for Ollama/local LLMs
	prompt := fmt.Sprintf(`Rephrase this question in 2 different ways. Keep the same meaning. Return only a JSON array of the 2 rephrased questions, nothing else.

Question: %s

JSON array:`, query)

	// Use the LLM to generate variations
	response, err := qr.provider.GenerateCompletion(prompt, []string{})
	if err != nil {
		log.Printf("[QueryRewriter] Error generating variations: %v", err)
		// Fallback to original query on error
		return &RewriteResult{
			Original:   query,
			Variations: []string{query},
		}, nil
	}

	log.Printf("[QueryRewriter] Raw LLM response: %s", response)

	// Parse the JSON response
	variations, err := qr.parseVariations(response)
	if err != nil {
		log.Printf("[QueryRewriter] Error parsing variations: %v", err)
		// Fallback to original query on error
		return &RewriteResult{
			Original:   query,
			Variations: []string{query},
		}, nil
	}

	// Always include the original query
	allQueries := append([]string{query}, variations...)

	log.Printf("[QueryRewriter] Generated %d query variations (including original)", len(allQueries))
	for i, q := range allQueries {
		log.Printf("[QueryRewriter]   %d: %s", i+1, q)
	}

	return &RewriteResult{
		Original:   query,
		Variations: allQueries,
	}, nil
}

// parseVariations extracts query variations from LLM response
func (qr *QueryRewriter) parseVariations(response string) ([]string, error) {
	// Try to find JSON array in the response
	response = strings.TrimSpace(response)

	// Find the first '[' and last ']'
	startIdx := strings.Index(response, "[")
	endIdx := strings.LastIndex(response, "]")

	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return nil, fmt.Errorf("no JSON array found in response")
	}

	jsonStr := response[startIdx : endIdx+1]

	var variations []string
	err := json.Unmarshal([]byte(jsonStr), &variations)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Filter out empty strings
	filtered := make([]string, 0, len(variations))
	for _, v := range variations {
		v = strings.TrimSpace(v)
		if v != "" {
			filtered = append(filtered, v)
		}
	}

	return filtered, nil
}
