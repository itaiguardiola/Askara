package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Document represents metadata about an uploaded document
type Document struct {
	ID           string    `json:"id"`
	UUID         string    `json:"uuid"`
	Filename     string    `json:"filename"`
	FileSize     int64     `json:"file_size"`
	UploadDate   time.Time `json:"upload_date"`
	ChunkCount   int       `json:"chunk_count"`
	ContentType  string    `json:"content_type"`
	FirstChunkID string    `json:"first_chunk_id"`
	LastChunkID  string    `json:"last_chunk_id"`
}

// DocumentStats represents usage statistics for a user
type DocumentStats struct {
	TotalDocuments int       `json:"total_documents"`
	TotalChunks    int       `json:"total_chunks"`
	TotalSize      int64     `json:"total_size"`
	OldestUpload   time.Time `json:"oldest_upload,omitempty"`
	NewestUpload   time.Time `json:"newest_upload,omitempty"`
}

// DocumentStore defines the interface for document metadata storage
type DocumentStore interface {
	SaveDocument(doc *Document) error
	GetDocuments(uuid string) ([]Document, error)
	GetDocument(uuid, docID string) (*Document, error)
	DeleteDocument(uuid, docID string) error
	GetStats(uuid string) (*DocumentStats, error)
}

// JSONDocumentStore implements DocumentStore using JSON files
type JSONDocumentStore struct {
	basePath string
	mu       sync.RWMutex
}

// userDocuments represents the JSON structure stored in each user's file
type userDocuments struct {
	Documents []Document `json:"documents"`
}

// NewJSONDocumentStore creates a new JSON-based document store
func NewJSONDocumentStore(basePath string) (*JSONDocumentStore, error) {
	// Ensure the base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &JSONDocumentStore{
		basePath: basePath,
	}, nil
}

// GenerateDocumentID creates a unique document ID based on filename and timestamp
func GenerateDocumentID(filename string) string {
	timestamp := time.Now().UnixNano()
	data := fmt.Sprintf("%s-%d", filename, timestamp)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("doc-%s", hex.EncodeToString(hash[:8]))
}

// GenerateChunkID creates a vector ID for a chunk
func GenerateChunkID(uuid, docID string, chunkIndex int) string {
	return fmt.Sprintf("%s-%s-%d", uuid, docID, chunkIndex)
}

// getUserFilePath returns the path to a user's document metadata file
func (s *JSONDocumentStore) getUserFilePath(uuid string) string {
	return filepath.Join(s.basePath, fmt.Sprintf("%s.json", uuid))
}

// loadUserDocuments loads documents for a user from their JSON file
func (s *JSONDocumentStore) loadUserDocuments(uuid string) (*userDocuments, error) {
	filePath := s.getUserFilePath(uuid)

	// If file doesn't exist, return empty structure
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &userDocuments{Documents: []Document{}}, nil
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read document file: %w", err)
	}

	var docs userDocuments
	if err := json.Unmarshal(data, &docs); err != nil {
		return nil, fmt.Errorf("failed to parse document file: %w", err)
	}

	return &docs, nil
}

// saveUserDocuments saves documents for a user to their JSON file
func (s *JSONDocumentStore) saveUserDocuments(uuid string, docs *userDocuments) error {
	filePath := s.getUserFilePath(uuid)

	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal documents: %w", err)
	}

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write document file: %w", err)
	}

	return nil
}

// SaveDocument saves a document to the store
func (s *JSONDocumentStore) SaveDocument(doc *Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	docs, err := s.loadUserDocuments(doc.UUID)
	if err != nil {
		return err
	}

	// Check if document with this ID already exists
	for i, existingDoc := range docs.Documents {
		if existingDoc.ID == doc.ID {
			// Update existing document
			docs.Documents[i] = *doc
			return s.saveUserDocuments(doc.UUID, docs)
		}
	}

	// Add new document
	docs.Documents = append(docs.Documents, *doc)

	return s.saveUserDocuments(doc.UUID, docs)
}

// GetDocuments retrieves all documents for a user
func (s *JSONDocumentStore) GetDocuments(uuid string) ([]Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	docs, err := s.loadUserDocuments(uuid)
	if err != nil {
		return nil, err
	}

	// Sort by upload date (newest first)
	sort.Slice(docs.Documents, func(i, j int) bool {
		return docs.Documents[i].UploadDate.After(docs.Documents[j].UploadDate)
	})

	return docs.Documents, nil
}

// GetDocument retrieves a specific document by ID
func (s *JSONDocumentStore) GetDocument(uuid, docID string) (*Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	docs, err := s.loadUserDocuments(uuid)
	if err != nil {
		return nil, err
	}

	for _, doc := range docs.Documents {
		if doc.ID == docID {
			return &doc, nil
		}
	}

	return nil, fmt.Errorf("document not found")
}

// DeleteDocument removes a document from the store
func (s *JSONDocumentStore) DeleteDocument(uuid, docID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	docs, err := s.loadUserDocuments(uuid)
	if err != nil {
		return err
	}

	// Find and remove the document
	for i, doc := range docs.Documents {
		if doc.ID == docID {
			// Remove from slice
			docs.Documents = append(docs.Documents[:i], docs.Documents[i+1:]...)
			return s.saveUserDocuments(uuid, docs)
		}
	}

	return fmt.Errorf("document not found")
}

// GetStats calculates usage statistics for a user
func (s *JSONDocumentStore) GetStats(uuid string) (*DocumentStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	docs, err := s.loadUserDocuments(uuid)
	if err != nil {
		return nil, err
	}

	stats := &DocumentStats{
		TotalDocuments: len(docs.Documents),
	}

	if len(docs.Documents) == 0 {
		return stats, nil
	}

	// Calculate totals
	for _, doc := range docs.Documents {
		stats.TotalChunks += doc.ChunkCount
		stats.TotalSize += doc.FileSize
	}

	// Find oldest and newest
	stats.OldestUpload = docs.Documents[0].UploadDate
	stats.NewestUpload = docs.Documents[0].UploadDate

	for _, doc := range docs.Documents {
		if doc.UploadDate.Before(stats.OldestUpload) {
			stats.OldestUpload = doc.UploadDate
		}
		if doc.UploadDate.After(stats.NewestUpload) {
			stats.NewestUpload = doc.UploadDate
		}
	}

	return stats, nil
}
