package postapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/itaiguardiola/askara/storage"
	"github.com/itaiguardiola/askara/validator"
)

// UpdateMetadataRequest represents a request to update document metadata
type UpdateMetadataRequest struct {
	UUID        string   `json:"uuid"`
	Title       string   `json:"title,omitempty"`
	Creator     string   `json:"creator,omitempty"`
	Description string   `json:"description,omitempty"`
	Category    string   `json:"category,omitempty"`
	CustomTags  []string `json:"custom_tags,omitempty"`
}

// AddTagRequest represents a request to add a tag to a document
type AddTagRequest struct {
	UUID string `json:"uuid"`
	Tag  string `json:"tag"`
}

// SearchByTagsRequest represents a request to search documents by tags
type SearchByTagsRequest struct {
	UUID string   `json:"uuid"`
	Tags []string `json:"tags"`
}

// GetMetadata handles GET /api/documents/:documentId/metadata
func (ctx *HandlerContext) GetMetadata(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	documentID := vars["documentId"]
	uuid := r.URL.Query().Get("uuid")

	// Validate UUID
	if err := validator.ValidateUUID(uuid); err != nil {
		log.Println("[GetMetadata ERR] Invalid UUID:", err)
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}

	// Get document
	doc, err := ctx.docStore.GetDocument(uuid, documentID)
	if err != nil {
		log.Println("[GetMetadata ERR] Failed to get document:", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Return metadata
	w.Header().Set("Content-Type", "application/json")
	if doc.Metadata == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"metadata": storage.NewDocumentMetadata(doc.Filename),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"metadata": doc.Metadata,
	})
}

// UpdateMetadata handles PUT /api/documents/:documentId/metadata
func (ctx *HandlerContext) UpdateMetadata(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	documentID := vars["documentId"]

	var req UpdateMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("[UpdateMetadata ERR] Invalid request:", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate UUID
	if err := validator.ValidateUUID(req.UUID); err != nil {
		log.Println("[UpdateMetadata ERR] Invalid UUID:", err)
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}

	// Get existing document
	doc, err := ctx.docStore.GetDocument(req.UUID, documentID)
	if err != nil {
		log.Println("[UpdateMetadata ERR] Document not found:", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Initialize metadata if it doesn't exist
	if doc.Metadata == nil {
		doc.Metadata = storage.NewDocumentMetadata(doc.Filename)
	}

	// Update user-editable fields
	if req.Title != "" {
		doc.Metadata.Title = req.Title
	}
	if req.Creator != "" {
		doc.Metadata.Creator = req.Creator
	}
	if req.Description != "" {
		doc.Metadata.Description = req.Description
	}
	if req.Category != "" {
		doc.Metadata.Category = req.Category
	}
	if req.CustomTags != nil {
		doc.Metadata.CustomTags = req.CustomTags
	}

	// Save updated metadata
	if err := ctx.docStore.UpdateMetadata(req.UUID, documentID, doc.Metadata); err != nil {
		log.Println("[UpdateMetadata ERR] Failed to update metadata:", err)
		http.Error(w, "Failed to update metadata", http.StatusInternalServerError)
		return
	}

	log.Printf("[UpdateMetadata] Successfully updated metadata for document %s", documentID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Metadata updated successfully",
		"metadata": doc.Metadata,
	})
}

// AddTag handles POST /api/documents/:documentId/tags
func (ctx *HandlerContext) AddTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	documentID := vars["documentId"]

	var req AddTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("[AddTag ERR] Invalid request:", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate UUID
	if err := validator.ValidateUUID(req.UUID); err != nil {
		log.Println("[AddTag ERR] Invalid UUID:", err)
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}

	// Validate tag
	if req.Tag == "" {
		http.Error(w, "Tag cannot be empty", http.StatusBadRequest)
		return
	}

	// Get existing document
	doc, err := ctx.docStore.GetDocument(req.UUID, documentID)
	if err != nil {
		log.Println("[AddTag ERR] Document not found:", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Initialize metadata if it doesn't exist
	if doc.Metadata == nil {
		doc.Metadata = storage.NewDocumentMetadata(doc.Filename)
	}

	// Add tag
	doc.Metadata.AddCustomTag(req.Tag)

	// Save updated metadata
	if err := ctx.docStore.UpdateMetadata(req.UUID, documentID, doc.Metadata); err != nil {
		log.Println("[AddTag ERR] Failed to add tag:", err)
		http.Error(w, "Failed to add tag", http.StatusInternalServerError)
		return
	}

	log.Printf("[AddTag] Successfully added tag '%s' to document %s", req.Tag, documentID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Tag added successfully",
		"tags":    doc.Metadata.CustomTags,
	})
}

// RemoveTag handles DELETE /api/documents/:documentId/tags/:tag
func (ctx *HandlerContext) RemoveTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	documentID := vars["documentId"]
	tag := vars["tag"]
	uuid := r.URL.Query().Get("uuid")

	// Validate UUID
	if err := validator.ValidateUUID(uuid); err != nil {
		log.Println("[RemoveTag ERR] Invalid UUID:", err)
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}

	// Get existing document
	doc, err := ctx.docStore.GetDocument(uuid, documentID)
	if err != nil {
		log.Println("[RemoveTag ERR] Document not found:", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	if doc.Metadata == nil {
		http.Error(w, "No metadata found", http.StatusNotFound)
		return
	}

	// Remove tag
	doc.Metadata.RemoveCustomTag(tag)

	// Save updated metadata
	if err := ctx.docStore.UpdateMetadata(uuid, documentID, doc.Metadata); err != nil {
		log.Println("[RemoveTag ERR] Failed to remove tag:", err)
		http.Error(w, "Failed to remove tag", http.StatusInternalServerError)
		return
	}

	log.Printf("[RemoveTag] Successfully removed tag '%s' from document %s", tag, documentID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Tag removed successfully",
		"tags":    doc.Metadata.CustomTags,
	})
}

// SearchByTags handles POST /api/documents/search/tags
func (ctx *HandlerContext) SearchByTags(w http.ResponseWriter, r *http.Request) {
	var req SearchByTagsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("[SearchByTags ERR] Invalid request:", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate UUID
	if err := validator.ValidateUUID(req.UUID); err != nil {
		log.Println("[SearchByTags ERR] Invalid UUID:", err)
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}

	// Search documents by tags
	documents, err := ctx.docStore.SearchByTags(req.UUID, req.Tags)
	if err != nil {
		log.Println("[SearchByTags ERR] Search failed:", err)
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	log.Printf("[SearchByTags] Found %d documents matching tags %v", len(documents), req.Tags)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"documents": documents,
		"count":     len(documents),
	})
}

// GetAllTags handles GET /api/documents/tags
func (ctx *HandlerContext) GetAllTags(w http.ResponseWriter, r *http.Request) {
	uuid := r.URL.Query().Get("uuid")

	// Validate UUID
	if err := validator.ValidateUUID(uuid); err != nil {
		log.Println("[GetAllTags ERR] Invalid UUID:", err)
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}

	// Get all documents
	documents, err := ctx.docStore.GetDocuments(uuid)
	if err != nil {
		log.Println("[GetAllTags ERR] Failed to get documents:", err)
		http.Error(w, "Failed to get documents", http.StatusInternalServerError)
		return
	}

	// Collect all unique tags
	tagMap := make(map[string]int) // tag -> count
	for _, doc := range documents {
		if doc.Metadata != nil {
			for _, tag := range doc.Metadata.GetAllTags() {
				tagMap[tag]++
			}
		}
	}

	// Convert to list with counts
	type TagCount struct {
		Tag   string `json:"tag"`
		Count int    `json:"count"`
	}

	tags := make([]TagCount, 0, len(tagMap))
	for tag, count := range tagMap {
		tags = append(tags, TagCount{Tag: tag, Count: count})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tags":  tags,
		"total": len(tags),
	})
}
