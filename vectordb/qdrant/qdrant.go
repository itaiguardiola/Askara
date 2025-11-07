package qdrant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/itaiguardiola/askara/chunk"
	"github.com/itaiguardiola/askara/vectordb"
	cache "github.com/patrickmn/go-cache"
)

const (
	VECTOR_SIZE     = 1536 // ada002
	VECTOR_DISTANCE = "Cosine"
	BATCH_SIZE      = 500
)

type Qdrant struct {
	Endpoint string
	cache    *cache.Cache
}

type Point struct {
	ID      int               `json:"id"`
	Vector  []float32         `json:"vector"`
	Payload map[string]string `json:"payload,omitempty"`
}

type Match struct {
	ID      int               `json:"id"`
	Score   float32           `json:"score"`
	Payload map[string]string `json:"payload"`
	Version int               `json:"version"`
}

type SearchResult struct {
	Result []Match `json:"result"`
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

type NamespaceConfig struct {
	Vectors struct {
		Size     int    `json:"size"`
		Distance string `json:"distance"`
	} `json:"vectors"`
}

func New(endpoint string) (*Qdrant, error) {
	return &Qdrant{
		Endpoint: endpoint,
		cache:    cache.New(5*time.Minute, 10*time.Minute),
	}, nil
}

func (q *Qdrant) NamespaceExists(uuid string) (bool, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/collections/%s", q.Endpoint, uuid), nil)
	if err != nil {
		return false, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		_, _ = resp.Body.Read(body)
		return false, fmt.Errorf("failed to check namespace, status code: %d, %s", resp.StatusCode, body)
	}

	return true, nil
}

func (q *Qdrant) CreateNamespace(uuid string) error {
	if _, found := q.cache.Get(uuid); found {
		return nil
	}

	if exists, err := q.NamespaceExists(uuid); err != nil {
		return err
	} else if exists {
		return nil
	}

	config := NamespaceConfig{}
	config.Vectors.Size = VECTOR_SIZE
	config.Vectors.Distance = VECTOR_DISTANCE

	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/collections/%s", q.Endpoint, uuid), bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create namespace, status code: %d", resp.StatusCode)
	}

	q.cache.Set(uuid, true, cache.DefaultExpiration)

	return nil
}

func (q *Qdrant) UpsertEmbeddings(embeddings [][]float32, chunks []chunk.Chunk, uuid string) error {
	// For backward compatibility, use timestamp-based doc ID
	docID := fmt.Sprintf("doc-%d", time.Now().UnixNano())
	return q.UpsertEmbeddingsWithDocID(embeddings, chunks, uuid, docID)
}

func (q *Qdrant) UpsertEmbeddingsWithDocID(embeddings [][]float32, chunks []chunk.Chunk, uuid string, docID string) error {
	if err := q.CreateNamespace(uuid); err != nil {
		return err
	}

	points := make([]Point, len(embeddings))

	for i, embedding := range embeddings {
		points[i].ID = i
		points[i].Vector = embedding
		if i < len(chunks) {
			points[i].Payload = map[string]string{
				"uuid":      uuid,
				"doc_id":    docID,
				"start":     fmt.Sprintf("%d", chunks[i].Start),
				"end":       fmt.Sprintf("%d", chunks[i].End),
				"title":     chunks[i].Title,
				"text":      chunks[i].Text,
				"file_name": chunks[i].Title,
			}
		}
	}

	for i := 0; i < len(points); i += BATCH_SIZE {
		end := i + BATCH_SIZE
		if end > len(points) {
			end = len(points)
		}

		data := map[string][]Point{"points": points[i:end]}
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/collections/%s/points", q.Endpoint, uuid), bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to upsert embeddings, status code: %d", resp.StatusCode)
		}
	}

	return nil
}

func (q *Qdrant) Retrieve(questionEmbedding []float32, topK int, uuid string) ([]vectordb.QueryMatch, error) {
	data := map[string]interface{}{
		"vector":       questionEmbedding,
		"top":          topK,
		"with_payload": true,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/collections/%s/points/search", q.Endpoint, uuid), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve embeddings, status code: %d", resp.StatusCode)
	}

	var searchResult SearchResult
	err = json.NewDecoder(resp.Body).Decode(&searchResult)
	if err != nil {
		return nil, err
	}

	// Convert qdrantMatch to QueryMatch
	queryMatches := make([]vectordb.QueryMatch, len(searchResult.Result))
	for i, result := range searchResult.Result {
		queryMatches[i].ID = fmt.Sprintf("%d", result.ID)
		queryMatches[i].Score = result.Score
		queryMatches[i].Metadata = result.Payload
	}

	return queryMatches, nil
}

// DeleteByDocumentID deletes all points associated with a document ID
func (q *Qdrant) DeleteByDocumentID(docID string, uuid string) error {
	// Delete points by filter
	data := map[string]interface{}{
		"filter": map[string]interface{}{
			"must": []map[string]interface{}{
				{
					"key": "doc_id",
					"match": map[string]string{
						"value": docID,
					},
				},
				{
					"key": "uuid",
					"match": map[string]string{
						"value": uuid,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal delete request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/collections/%s/points/delete", q.Endpoint, uuid), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		_, _ = resp.Body.Read(body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListDocumentIDs lists all unique document IDs in a collection
func (q *Qdrant) ListDocumentIDs(uuid string) ([]string, error) {
	// Use scroll to get all points
	data := map[string]interface{}{
		"limit":        100,
		"with_payload": true,
		"with_vector":  false,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scroll request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/collections/%s/points/scroll", q.Endpoint, uuid), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create scroll request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send scroll request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 1024)
		_, _ = resp.Body.Read(body)
		return nil, fmt.Errorf("scroll failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Result struct {
			Points []struct {
				Payload map[string]string `json:"payload"`
			} `json:"points"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode scroll response: %w", err)
	}

	// Extract unique doc_ids
	docIDMap := make(map[string]bool)
	for _, point := range result.Result.Points {
		if docID, ok := point.Payload["doc_id"]; ok {
			docIDMap[docID] = true
		}
	}

	docIDs := make([]string, 0, len(docIDMap))
	for docID := range docIDMap {
		docIDs = append(docIDs, docID)
	}

	return docIDs, nil
}
