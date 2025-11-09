package metadata

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itaiguardiola/askara/llm"
	"github.com/itaiguardiola/askara/storage"
)

// MetadataExtractor uses LLM to extract metadata from document content
type MetadataExtractor struct {
	provider llm.LLMProvider
}

// NewMetadataExtractor creates a new metadata extractor
func NewMetadataExtractor(provider llm.LLMProvider) *MetadataExtractor {
	return &MetadataExtractor{
		provider: provider,
	}
}

// ExtractedMetadata represents the structured output from LLM
type ExtractedMetadata struct {
	Summary     string            `json:"summary"`
	Tags        []string          `json:"tags"`
	Topics      []string          `json:"topics"`
	Category    string            `json:"category"`
	Language    string            `json:"language"`
	Entities    map[string]string `json:"entities"` // entity text -> type
	KeyConcepts []string          `json:"key_concepts"`
}

// ExtractMetadata analyzes document content and extracts metadata using LLM
func (e *MetadataExtractor) ExtractMetadata(content string, filename string) (*storage.DocumentMetadata, error) {
	// Limit content to first ~3000 tokens for analysis (roughly 12000 chars)
	contentSample := content
	if len(content) > 12000 {
		contentSample = content[:12000]
	}

	// Build the prompt for metadata extraction
	prompt := e.buildMetadataPrompt(contentSample, filename)

	// Generate completion with no additional context
	response, err := e.provider.GenerateCompletion(prompt, []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to generate metadata: %w", err)
	}

	// Parse the JSON response
	extracted, err := e.parseMetadataResponse(response)
	if err != nil {
		// If JSON parsing fails, create basic metadata
		return e.createFallbackMetadata(filename, content), nil
	}

	// Convert to DocumentMetadata
	metadata := storage.NewDocumentMetadata(filename)
	metadata.Summary = extracted.Summary
	metadata.Topics = extracted.Topics
	metadata.Category = extracted.Category
	metadata.Language = extracted.Language

	// Add auto-tags with high confidence (from LLM)
	for _, tag := range extracted.Tags {
		metadata.AddAutoTag(tag, 0.9)
	}

	// Add key concepts as tags with medium confidence
	for _, concept := range extracted.KeyConcepts {
		metadata.AddAutoTag(concept, 0.7)
	}

	// Add entities
	for entityText, entityType := range extracted.Entities {
		metadata.AddEntity(entityText, entityType)
	}

	// Calculate word count and reading time
	metadata.WordCount = e.estimateWordCount(content)
	metadata.CalculateReadingTime()

	return metadata, nil
}

// buildMetadataPrompt creates the prompt for LLM to extract metadata
func (e *MetadataExtractor) buildMetadataPrompt(content, filename string) string {
	return fmt.Sprintf(`You are a document analysis expert. Analyze the following document and extract structured metadata.

Document: %s

Content:
%s

Please analyze this document and provide metadata in the following JSON format:
{
  "summary": "A concise 2-3 sentence summary of the document",
  "tags": ["tag1", "tag2", "tag3", "tag4", "tag5"],
  "topics": ["main topic 1", "main topic 2", "main topic 3"],
  "category": "one of: Research, Technical, Business, Legal, Academic, Financial, Medical, General",
  "language": "detected language (e.g., English, Spanish, etc.)",
  "entities": {
    "entity name": "type (PERSON, ORG, LOCATION, DATE, etc.)"
  },
  "key_concepts": ["concept1", "concept2", "concept3"]
}

Guidelines:
- Tags should be specific keywords (5-10 tags)
- Topics should be broader themes (2-4 topics)
- Category should fit one of the standard categories
- Extract up to 10 named entities
- Key concepts are important ideas or terms from the document

Respond ONLY with valid JSON. No additional text or explanation.`, filename, content)
}

// parseMetadataResponse parses the LLM JSON response
func (e *MetadataExtractor) parseMetadataResponse(response string) (*ExtractedMetadata, error) {
	// Clean up response - sometimes LLMs add markdown code blocks
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var extracted ExtractedMetadata
	if err := json.Unmarshal([]byte(response), &extracted); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &extracted, nil
}

// createFallbackMetadata creates basic metadata when LLM extraction fails
func (e *MetadataExtractor) createFallbackMetadata(filename, content string) *storage.DocumentMetadata {
	metadata := storage.NewDocumentMetadata(filename)

	// Basic language detection (very simple)
	if strings.Contains(content, "español") || strings.Contains(content, "México") {
		metadata.Language = "Spanish"
	} else {
		metadata.Language = "English"
	}

	// Calculate word count
	metadata.WordCount = e.estimateWordCount(content)
	metadata.CalculateReadingTime()

	// Add a basic tag from filename
	if ext := strings.LastIndex(filename, "."); ext > 0 {
		metadata.AddAutoTag(filename[:ext], 0.5)
	}

	return metadata
}

// estimateWordCount estimates the number of words in the content
func (e *MetadataExtractor) estimateWordCount(content string) int {
	// Simple word count based on whitespace
	words := strings.Fields(content)
	return len(words)
}

// QuickTag generates a few quick tags from content without full metadata extraction
// This is faster and useful for batch processing
func (e *MetadataExtractor) QuickTag(content string, maxTags int) ([]string, error) {
	contentSample := content
	if len(content) > 5000 {
		contentSample = content[:5000]
	}

	prompt := fmt.Sprintf(`Extract %d relevant tags/keywords from this document content.
Respond with ONLY a JSON array of strings like: ["tag1", "tag2", "tag3"]

Content:
%s

JSON array of tags:`, maxTags, contentSample)

	response, err := e.provider.GenerateCompletion(prompt, []string{})
	if err != nil {
		return nil, err
	}

	// Parse JSON array
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var tags []string
	if err := json.Unmarshal([]byte(response), &tags); err != nil {
		return []string{}, nil // Return empty on parse failure
	}

	return tags, nil
}
