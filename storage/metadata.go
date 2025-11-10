package storage

import (
	"time"
)

// Tag represents a document tag with confidence score
type Tag struct {
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"` // 0.0-1.0
	Source     string  `json:"source"`     // "ml" or "user"
}

// Entity represents a named entity extracted from the document
type Entity struct {
	Text  string `json:"text"`
	Type  string `json:"type"`  // "PERSON", "ORG", "LOCATION", "DATE", etc.
	Count int    `json:"count"` // Occurrences in document
}

// DocumentMetadata represents rich metadata about a document
type DocumentMetadata struct {
	// User-editable fields
	Title       string   `json:"title"`                  // User-friendly name (default: filename)
	Creator     string   `json:"creator,omitempty"`      // Author/creator name
	Description string   `json:"description,omitempty"`  // User description
	Category    string   `json:"category,omitempty"`     // e.g., "Research", "Legal", "Finance"
	CustomTags  []string `json:"custom_tags,omitempty"`  // User-added tags

	// ML-generated fields
	AutoTags []Tag    `json:"auto_tags,omitempty"` // ML-generated tags with confidence
	Summary  string   `json:"summary,omitempty"`   // AI-generated summary
	Language string   `json:"language,omitempty"`  // Detected language
	Topics   []string `json:"topics,omitempty"`    // Key topics/themes
	Entities []Entity `json:"entities,omitempty"`  // Named entities (people, orgs, places)

	// Document properties
	PageCount   int `json:"page_count,omitempty"`   // Number of pages (for PDFs)
	WordCount   int `json:"word_count,omitempty"`   // Approximate word count
	ReadingTime int `json:"reading_time,omitempty"` // Estimated minutes

	// System metadata
	LastModified time.Time `json:"last_modified"`
	Version      int       `json:"version"` // For tracking metadata updates
}

// NewDocumentMetadata creates a new metadata object with default values
func NewDocumentMetadata(filename string) *DocumentMetadata {
	return &DocumentMetadata{
		Title:        filename,
		CustomTags:   []string{},
		AutoTags:     []Tag{},
		Topics:       []string{},
		Entities:     []Entity{},
		LastModified: time.Now(),
		Version:      1,
	}
}

// AddCustomTag adds a user-defined tag
func (m *DocumentMetadata) AddCustomTag(tag string) {
	// Check if tag already exists
	for _, t := range m.CustomTags {
		if t == tag {
			return
		}
	}
	m.CustomTags = append(m.CustomTags, tag)
	m.LastModified = time.Now()
	m.Version++
}

// RemoveCustomTag removes a user-defined tag
func (m *DocumentMetadata) RemoveCustomTag(tag string) {
	for i, t := range m.CustomTags {
		if t == tag {
			m.CustomTags = append(m.CustomTags[:i], m.CustomTags[i+1:]...)
			m.LastModified = time.Now()
			m.Version++
			return
		}
	}
}

// AddAutoTag adds an ML-generated tag with confidence score
func (m *DocumentMetadata) AddAutoTag(name string, confidence float64) {
	// Check if tag already exists
	for i, t := range m.AutoTags {
		if t.Name == name {
			// Update confidence if higher
			if confidence > t.Confidence {
				m.AutoTags[i].Confidence = confidence
				m.LastModified = time.Now()
			}
			return
		}
	}
	m.AutoTags = append(m.AutoTags, Tag{
		Name:       name,
		Confidence: confidence,
		Source:     "ml",
	})
	m.LastModified = time.Now()
}

// GetAllTags returns combined list of custom and auto tags
func (m *DocumentMetadata) GetAllTags() []string {
	allTags := make([]string, 0, len(m.CustomTags)+len(m.AutoTags))
	allTags = append(allTags, m.CustomTags...)
	for _, tag := range m.AutoTags {
		allTags = append(allTags, tag.Name)
	}
	return allTags
}

// AddEntity adds or updates a named entity
func (m *DocumentMetadata) AddEntity(text, entityType string) {
	// Check if entity already exists
	for i, e := range m.Entities {
		if e.Text == text && e.Type == entityType {
			m.Entities[i].Count++
			return
		}
	}
	m.Entities = append(m.Entities, Entity{
		Text:  text,
		Type:  entityType,
		Count: 1,
	})
}

// CalculateReadingTime estimates reading time based on word count
// Average reading speed: 200-250 words per minute
func (m *DocumentMetadata) CalculateReadingTime() {
	if m.WordCount > 0 {
		m.ReadingTime = (m.WordCount + 199) / 200 // Round up
	}
}
