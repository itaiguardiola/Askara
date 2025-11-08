package postapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/itaiguardiola/askara/chunk"
	"github.com/itaiguardiola/askara/llm"
	"github.com/itaiguardiola/askara/storage"
	"github.com/itaiguardiola/askara/validator"
)

type UploadResponse struct {
	Message             string             `json:"message"`
	NumFilesSucceeded   int                `json:"num_files_succeeded"`
	NumFilesFailed      int                `json:"num_files_failed"`
	SuccessfulFileNames []string           `json:"successful_file_names"`
	FailedFileNames     map[string]string  `json:"failed_file_names"`
	UploadedDocuments   []storage.Document `json:"uploaded_documents"`
}

const MAX_FILE_SIZE int64 = 25 << 20         // 3 MB
const MAX_TOTAL_UPLOAD_SIZE int64 = 50 << 20 // 3 MB

func (ctx *HandlerContext) UploadHandler(w http.ResponseWriter, r *http.Request) {

	// Limit the request body size
	r.Body = http.MaxBytesReader(w, r.Body, MAX_TOTAL_UPLOAD_SIZE)

	err := r.ParseMultipartForm(MAX_TOTAL_UPLOAD_SIZE) // Maximum upload of 3 MB
	if err != nil {
		if err == http.ErrMissingBoundary || err == http.ErrNotMultipart || err == http.ErrNotSupported {
			log.Println("[UploadHandler ERR] Error parsing multipart form:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Println("[UploadHandler ERR] Request body size exceeds the limit:", err)
		http.Error(w, "Request body size exceeds the limit", http.StatusRequestEntityTooLarge)
		return
	}

	files := r.MultipartForm.File["files"]
	uuid := r.FormValue("uuid") // Get the UUID from the form data
	userProvidedOpenApiKey := r.FormValue("apikey")

	// Validate UUID
	if err := validator.ValidateUUID(uuid); err != nil {
		log.Println("[UploadHandler ERR] Invalid UUID:", err)
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}

	log.Println("[UploadHandler] UUID=", uuid)

	providerToUse := ctx.llmProvider
	if userProvidedOpenApiKey != "" {
		log.Println("[UploadHandler] Using provided custom API key")
		// Create temporary OpenAI provider with custom key
		customProvider, err := llm.NewProvider(&llm.Config{
			Provider: "openai",
			OpenAIConfig: &llm.OpenAIConfig{
				APIKey: userProvidedOpenApiKey,
			},
		})
		if err != nil {
			log.Println("[UploadHandler ERR] Failed to create custom provider:", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		providerToUse = customProvider
	}

	responseData := UploadResponse{
		SuccessfulFileNames: make([]string, 0),
		FailedFileNames:     make(map[string]string),
		UploadedDocuments:   make([]storage.Document, 0),
	}

	for _, file := range files {
		fileName := validator.SanitizeFilename(file.Filename)

		if file.Size > MAX_FILE_SIZE {
			errMsg := fmt.Sprintf("File size exceeds the %d bytes limit", MAX_FILE_SIZE)
			log.Println("[UploadHandler ERR]", errMsg, fileName)
			responseData.NumFilesFailed++
			responseData.FailedFileNames[fileName] = errMsg
			continue
		}

		// Read the file in memory
		f, err := file.Open()
		if err != nil {
			errMsg := "Error opening file"
			log.Println("[UploadHandler ERR]", errMsg, err)
			responseData.NumFilesFailed++
			responseData.FailedFileNames[fileName] = errMsg
			continue
		}
		defer f.Close()

		// Get the file name, MIME type, and first 32 characters of the contents
		fileType := file.Header.Get("Content-Type")
		fileContent := ""
		filePreview := ""

		// Check if the file is a PDF
		if fileType == "application/pdf" {
			fileContent, err = chunk.ExtractTextFromPDF(f, file.Size)
			if err != nil {
				errMsg := "Error extracting text from PDF"
				log.Println("[UploadHandler ERR]", errMsg, err)
				responseData.NumFilesFailed++
				responseData.FailedFileNames[fileName] = errMsg
				continue
			}
		} else {
			fileContent, err = chunk.GetTextFromFile(f)
			if err != nil {
				errMsg := "Error reading file"
				log.Println("[UploadHandler ERR]", errMsg, err)
				responseData.NumFilesFailed++
				responseData.FailedFileNames[fileName] = errMsg
				continue
			}
		}

		if len(fileContent) > 32 {
			filePreview = fileContent[:32]
		}
		log.Printf("File Name: %s, File Type: %s, File Content (first 32 characters): %s\n", fileName, fileType, filePreview)

		// Compute content hash for deduplication
		contentHash := storage.ComputeContentHash(fileContent)
		log.Printf("[UploadHandler] Content hash for %s: %s", fileName, contentHash)

		// Check if this content already exists for this user
		existingDoc, err := ctx.docStore.(*storage.JSONDocumentStore).FindDocumentByContentHash(uuid, contentHash)
		if err != nil {
			log.Printf("[UploadHandler WARN] Error checking for duplicates: %v", err)
			// Continue with upload even if duplicate check fails
		} else if existingDoc != nil {
			// Duplicate file found
			log.Printf("[UploadHandler] Duplicate file detected: %s (original: %s)", fileName, existingDoc.Filename)
			responseData.NumFilesFailed++
			responseData.FailedFileNames[fileName] = fmt.Sprintf("Duplicate content (same as '%s' uploaded on %s)",
				existingDoc.Filename, existingDoc.UploadDate.Format("2006-01-02 15:04:05"))
			continue
		}

		// Process the fileBytes into embeddings and store in vector DB here
		chunks, err := chunk.CreateChunks(fileContent, fileName)
		if err != nil {
			errMsg := "Error chunking file"
			log.Println("[UploadHandler ERR]", errMsg, err)
			responseData.NumFilesFailed++
			responseData.FailedFileNames[fileName] = errMsg
			continue
		}

		// Generate embeddings for each chunk using LLM provider
		embeddings := make([][]float32, 0, len(chunks))
		embeddingError := false
		for _, c := range chunks {
			embedding, err := providerToUse.GenerateEmbedding(c.Text)
			if err != nil {
				errMsg := fmt.Sprintf("Error getting embeddings: %v", err)
				log.Println("[UploadHandler ERR]", errMsg)
				responseData.NumFilesFailed++
				responseData.FailedFileNames[fileName] = errMsg
				embeddingError = true
				break
			}
			embeddings = append(embeddings, embedding)
		}

		// Skip this file if embedding generation failed
		if embeddingError {
			continue
		}
		fmt.Printf("Total chunks: %d\n", len(chunks))
		fmt.Printf("Total embeddings: %d\n", len(embeddings))
		fmt.Printf("Embeddings length: %d\n", len(embeddings[0]))

		// Generate document ID
		docID := storage.GenerateDocumentID(fileName)

		// Upsert with document ID
		err = ctx.vectorDB.UpsertEmbeddingsWithDocID(embeddings, chunks, uuid, docID)
		if err != nil {
			errMsg := fmt.Sprintf("Error upserting embeddings to vector DB: %v", err)
			log.Println("[UploadHandler ERR]", errMsg)
			responseData.NumFilesFailed++
			responseData.FailedFileNames[fileName] = errMsg
			continue
		}

		log.Println("Successfully added vector DB embeddings!")

		// Save document metadata
		doc := storage.Document{
			ID:           docID,
			UUID:         uuid,
			Filename:     fileName,
			FileSize:     file.Size,
			ContentHash:  contentHash,
			UploadDate:   time.Now(),
			ChunkCount:   len(chunks),
			ContentType:  fileType,
			FirstChunkID: storage.GenerateChunkID(uuid, docID, 0),
			LastChunkID:  storage.GenerateChunkID(uuid, docID, len(chunks)-1),
		}

		if err := ctx.docStore.SaveDocument(&doc); err != nil {
			log.Printf("[UploadHandler WARN] Failed to save document metadata: %v", err)
			// Continue anyway - the vectors are uploaded
		}

		responseData.NumFilesSucceeded++
		responseData.SuccessfulFileNames = append(responseData.SuccessfulFileNames, fileName)
		responseData.UploadedDocuments = append(responseData.UploadedDocuments, doc)
	}

	if responseData.NumFilesFailed > 0 {
		responseData.Message = "Some files failed to upload and process"
	} else {
		responseData.Message = "All files uploaded and processed successfully"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonResponse, err := json.Marshal(responseData)
	if err != nil {
		log.Println("[UploadHandler ERR] Error writing json response", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonResponse)
}
