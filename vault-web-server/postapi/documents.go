package postapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/itaiguardiola/askara/storage"
)

// DocumentListResponse represents the response for listing documents
type DocumentListResponse struct {
	Documents []storage.Document `json:"documents"`
}

// DocumentDeleteResponse represents the response for deleting a document
type DocumentDeleteResponse struct {
	Message string `json:"message"`
	DocID   string `json:"doc_id"`
}

// DocumentStatsResponse represents the response for document statistics
type DocumentStatsResponse struct {
	Stats *storage.DocumentStats `json:"stats"`
}

// ListDocumentsHandler handles GET /api/documents
func (ctx *HandlerContext) ListDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		log.Println("[ListDocumentsHandler ERR] Missing UUID parameter")
		http.Error(w, "Missing UUID parameter", http.StatusBadRequest)
		return
	}

	log.Printf("[ListDocumentsHandler] Fetching documents for UUID=%s", uuid)

	documents, err := ctx.docStore.GetDocuments(uuid)
	if err != nil {
		log.Printf("[ListDocumentsHandler ERR] Failed to get documents: %v", err)
		http.Error(w, fmt.Sprintf("Failed to retrieve documents: %v", err), http.StatusInternalServerError)
		return
	}

	response := DocumentListResponse{
		Documents: documents,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ListDocumentsHandler ERR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("[ListDocumentsHandler] Successfully returned %d documents", len(documents))
}

// DeleteDocumentHandler handles DELETE /api/documents/{documentId}
func (ctx *HandlerContext) DeleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	docID := vars["documentId"]

	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		log.Println("[DeleteDocumentHandler ERR] Missing UUID parameter")
		http.Error(w, "Missing UUID parameter", http.StatusBadRequest)
		return
	}

	if docID == "" {
		log.Println("[DeleteDocumentHandler ERR] Missing document ID")
		http.Error(w, "Missing document ID", http.StatusBadRequest)
		return
	}

	log.Printf("[DeleteDocumentHandler] Deleting document ID=%s for UUID=%s", docID, uuid)

	// Get the document to verify ownership
	doc, err := ctx.docStore.GetDocument(uuid, docID)
	if err != nil {
		log.Printf("[DeleteDocumentHandler ERR] Document not found: %v", err)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if doc.UUID != uuid {
		log.Printf("[DeleteDocumentHandler ERR] UUID mismatch: expected %s, got %s", doc.UUID, uuid)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete from vector database
	if err := ctx.vectorDB.DeleteByDocumentID(docID, uuid); err != nil {
		log.Printf("[DeleteDocumentHandler ERR] Failed to delete from vector DB: %v", err)
		http.Error(w, fmt.Sprintf("Failed to delete document vectors: %v", err), http.StatusInternalServerError)
		return
	}

	// Delete from metadata store
	if err := ctx.docStore.DeleteDocument(uuid, docID); err != nil {
		log.Printf("[DeleteDocumentHandler ERR] Failed to delete from metadata store: %v", err)
		http.Error(w, fmt.Sprintf("Failed to delete document metadata: %v", err), http.StatusInternalServerError)
		return
	}

	response := DocumentDeleteResponse{
		Message: "Document deleted successfully",
		DocID:   docID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[DeleteDocumentHandler ERR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("[DeleteDocumentHandler] Successfully deleted document ID=%s", docID)
}

// GetDocumentStatsHandler handles GET /api/documents/stats
func (ctx *HandlerContext) GetDocumentStatsHandler(w http.ResponseWriter, r *http.Request) {
	uuid := r.URL.Query().Get("uuid")
	if uuid == "" {
		log.Println("[GetDocumentStatsHandler ERR] Missing UUID parameter")
		http.Error(w, "Missing UUID parameter", http.StatusBadRequest)
		return
	}

	log.Printf("[GetDocumentStatsHandler] Fetching stats for UUID=%s", uuid)

	stats, err := ctx.docStore.GetStats(uuid)
	if err != nil {
		log.Printf("[GetDocumentStatsHandler ERR] Failed to get stats: %v", err)
		http.Error(w, fmt.Sprintf("Failed to retrieve statistics: %v", err), http.StatusInternalServerError)
		return
	}

	response := DocumentStatsResponse{
		Stats: stats,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[GetDocumentStatsHandler ERR] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("[GetDocumentStatsHandler] Successfully returned stats: %d docs, %d chunks", stats.TotalDocuments, stats.TotalChunks)
}
