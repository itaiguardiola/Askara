package postapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// FeedbackRequest represents feedback from users
type FeedbackRequest struct {
	Type     string            `json:"type"`     // feature_request, bug_report, general_feedback
	Title    string            `json:"title"`    // Short title
	Content  string            `json:"content"`  // Markdown content
	Metadata map[string]string `json:"metadata"` // Optional metadata (version, browser, etc.)
}

// FeedbackResponse represents the response after submitting feedback
type FeedbackResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ID        string `json:"id"`        // Unique feedback ID
	Timestamp string `json:"timestamp"` // When feedback was submitted
}

// SubmitFeedbackHandler handles POST /api/feedback
func (ctx *HandlerContext) SubmitFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req FeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Feedback] Invalid request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Type == "" {
		req.Type = "general_feedback"
	}
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Create feedback directory if it doesn't exist
	feedbackDir := filepath.Join(".", "data", "feedback")
	if err := os.MkdirAll(feedbackDir, 0755); err != nil {
		log.Printf("[Feedback] Failed to create feedback directory: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate unique ID and timestamp
	timestamp := time.Now()
	feedbackID := fmt.Sprintf("%s_%s_%d",
		req.Type,
		timestamp.Format("20060102_150405"),
		timestamp.UnixNano()%1000,
	)

	// Create feedback file content
	feedbackContent := fmt.Sprintf(`# %s

**Type**: %s
**Submitted**: %s
**ID**: %s

%s

---

## Metadata
`, req.Title, req.Type, timestamp.Format("2006-01-02 15:04:05 MST"), feedbackID, req.Content)

	// Add metadata if provided
	if len(req.Metadata) > 0 {
		for key, value := range req.Metadata {
			feedbackContent += fmt.Sprintf("- **%s**: %s\n", key, value)
		}
	} else {
		feedbackContent += "No metadata provided\n"
	}

	// Save to file
	filename := filepath.Join(feedbackDir, feedbackID+".md")
	if err := ioutil.WriteFile(filename, []byte(feedbackContent), 0644); err != nil {
		log.Printf("[Feedback] Failed to write feedback file: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("[Feedback] Saved: %s - %s", feedbackID, req.Title)

	// Return success response
	response := FeedbackResponse{
		Success:   true,
		Message:   "Feedback submitted successfully",
		ID:        feedbackID,
		Timestamp: timestamp.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListFeedbackHandler handles GET /api/feedback
func (ctx *HandlerContext) ListFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	feedbackDir := filepath.Join(".", "data", "feedback")

	// Check if feedback directory exists
	if _, err := os.Stat(feedbackDir); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	// Read all feedback files
	files, err := ioutil.ReadDir(feedbackDir)
	if err != nil {
		log.Printf("[Feedback] Failed to read feedback directory: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Build list of feedback items
	type FeedbackItem struct {
		ID        string    `json:"id"`
		Type      string    `json:"type"`
		Filename  string    `json:"filename"`
		Timestamp time.Time `json:"timestamp"`
		Size      int64     `json:"size"`
	}

	feedbackList := []FeedbackItem{}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".md" {
			// Parse filename to extract type and ID
			filenameBase := file.Name()[:len(file.Name())-3] // Remove .md extension

			item := FeedbackItem{
				ID:        filenameBase,
				Filename:  file.Name(),
				Timestamp: file.ModTime(),
				Size:      file.Size(),
			}

			// Extract type from filename (format: type_date_time_nano.md)
			// Example: feature_request_20250109_120530_123.md
			if len(filenameBase) > 0 {
				// Find first underscore to get type
				for i := 0; i < len(filenameBase); i++ {
					if filenameBase[i] == '_' {
						item.Type = filenameBase[:i]
						break
					}
				}
			}

			feedbackList = append(feedbackList, item)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feedbackList)
}

// GetFeedbackHandler handles GET /api/feedback/{id}
func (ctx *HandlerContext) GetFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	// Expected format: /api/feedback/{id}
	path := r.URL.Path
	id := path[len("/api/feedback/"):]

	if id == "" {
		http.Error(w, "Feedback ID is required", http.StatusBadRequest)
		return
	}

	// Read feedback file
	feedbackDir := filepath.Join(".", "data", "feedback")
	filename := filepath.Join(feedbackDir, id+".md")

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Feedback not found", http.StatusNotFound)
		} else {
			log.Printf("[Feedback] Failed to read feedback file: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return markdown content
	w.Header().Set("Content-Type", "text/markdown")
	w.Write(content)
}
