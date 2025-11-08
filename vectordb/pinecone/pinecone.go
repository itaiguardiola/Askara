package pinecone

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/itaiguardiola/askara/chunk"
	"github.com/itaiguardiola/askara/vectordb"
)

type PineconeVector struct {
	ID       string            `json:"id"`
	Values   []float32         `json:"values"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type Pinecone struct {
	Endpoint string
	ApiKey   string
}

func New(endpoint string, apiKey string) (*Pinecone, error) {
	return &Pinecone{
		Endpoint: endpoint,
		ApiKey:   apiKey,
	}, nil
}

func (p *Pinecone) UpsertEmbeddings(embeddings [][]float32, chunks []chunk.Chunk, uuid string) error {
	// For backward compatibility, use timestamp-based doc ID
	docID := fmt.Sprintf("doc-%d", time.Now().UnixNano())
	return p.UpsertEmbeddingsWithDocID(embeddings, chunks, uuid, docID)
}

func (p *Pinecone) UpsertEmbeddingsWithDocID(embeddings [][]float32, chunks []chunk.Chunk, uuid string, docID string) error {
	// Prepare URL
	url := p.Endpoint + "/vectors/upsert"

	// Prepare the vectors
	vectors := make([]PineconeVector, len(embeddings))
	for i, embedding := range embeddings {
		chunk := chunks[i]
		// Use format: {uuid}-{docID}-{chunkIndex}
		vectorID := fmt.Sprintf("%s-%s-%d", uuid, docID, i)
		vectors[i] = PineconeVector{
			ID:     vectorID,
			Values: embedding,
			Metadata: map[string]string{
				"uuid":      uuid,
				"doc_id":    docID,
				"file_name": chunk.Title,
				"start":     strconv.Itoa(chunk.Start),
				"end":       strconv.Itoa(chunk.End),
				"title":     chunk.Title,
				"text":      chunk.Text,
			},
		}
	}

	maxVectorsPerRequest := 100

	// Split vectors into smaller chunks and make multiple upsert requests
	for i := 0; i < len(vectors); i += maxVectorsPerRequest {
		end := i + maxVectorsPerRequest
		if end > len(vectors) {
			end = len(vectors)
		}

		requestBody, err := json.Marshal(struct {
			Vectors   []PineconeVector `json:"vectors"`
			Namespace string           `json:"namespace"`
		}{
			Vectors:   vectors[i:end],
			Namespace: uuid,
		})
		if err != nil {
			return err
		}
		log.Println("[upsertEmbeddingsToPinecone] Created pinecone upsert request with namespace = ", uuid)
		// Create HTTP request
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Api-Key", p.ApiKey)

		// Send HTTP request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			return errors.New(string(body))
		}
	}

	return nil
}

type PineconeQueryRequest struct {
	TopK            int                 `json:"topK"`
	IncludeMetadata bool                `json:"includeMetadata"`
	Namespace       string              `json:"namespace"`
	Queries         []PineconeQueryItem `json:"queries"`
}

type PineconeQueryItem struct {
	Values []float32 `json:"values"`
}

type PineconeQueryResponseResult struct {
	Matches []vectordb.QueryMatch `json:"matches"`
}

type PineconeQueryResponse struct {
	Results []PineconeQueryResponseResult `json:"results"`
}

func (p *Pinecone) Retrieve(questionEmbedding []float32, topK int, uuid string) ([]vectordb.QueryMatch, error) {
	// Prepare the Pinecone query request
	requestBody, _ := json.Marshal(PineconeQueryRequest{
		TopK:            topK,
		IncludeMetadata: true,
		Namespace:       uuid,
		Queries: []PineconeQueryItem{
			{
				Values: questionEmbedding,
			},
		},
	})

	log.Println("[retrieve] Querying pinecone namespace:", uuid)
	// Send the Pinecone query request
	pineconeIndexURL := p.Endpoint + "/query"
	req, _ := http.NewRequest("POST", pineconeIndexURL, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", p.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the Pinecone query response
	body, _ := ioutil.ReadAll(resp.Body)
	var pineconeQueryResponse PineconeQueryResponse
	json.Unmarshal(body, &pineconeQueryResponse)

	// Check if there are any results and return the matches
	if len(pineconeQueryResponse.Results) > 0 {
		return pineconeQueryResponse.Results[0].Matches, nil
	}

	return nil, nil
}

// HybridSearch implements hybrid search (vector + FTS)
// Note: Pinecone doesn't have native FTS, so this falls back to vector-only search
func (p *Pinecone) HybridSearch(vectorQuery []float32, textQuery string, topK int, uuid string) ([]vectordb.QueryMatch, error) {
	log.Printf("[Pinecone] HybridSearch not supported, falling back to vector-only search")
	return p.Retrieve(vectorQuery, topK, uuid)
}

// DeleteByDocumentID deletes all vectors associated with a document ID
func (p *Pinecone) DeleteByDocumentID(docID string, uuid string) error {
	url := p.Endpoint + "/vectors/delete"

	// Delete by filter on metadata
	requestBody, err := json.Marshal(map[string]interface{}{
		"filter": map[string]interface{}{
			"doc_id": map[string]string{
				"$eq": docID,
			},
			"uuid": map[string]string{
				"$eq": uuid,
			},
		},
		"namespace": uuid,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal delete request: %w", err)
	}

	log.Printf("[Pinecone] Deleting vectors for doc_id=%s in namespace=%s", docID, uuid)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", p.ApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("[Pinecone] Successfully deleted vectors for doc_id=%s", docID)
	return nil
}

// ListDocumentIDs lists all unique document IDs in a namespace
// Note: Pinecone doesn't have a direct way to list all vectors, so this is a placeholder
// In a production system, you'd want to maintain a separate index of doc IDs
func (p *Pinecone) ListDocumentIDs(uuid string) ([]string, error) {
	// Pinecone doesn't support listing all vectors directly
	// This would require maintaining a separate index or using the stats API
	// For now, we'll return an error indicating this isn't supported
	log.Println("[Pinecone] ListDocumentIDs not directly supported - use metadata store instead")
	return nil, fmt.Errorf("ListDocumentIDs not supported by Pinecone - use document metadata store")
}

func float32sToBytes(floats []float32) []byte {
	result := make([]byte, 4*len(floats))
	for i, f := range floats {
		binary.BigEndian.PutUint32(result[i*4:], math.Float32bits(f))
	}
	return result
}
