package vectordb

import (
	"github.com/itaiguardiola/askara/chunk"
)

type QueryMatch struct {
	ID       string            `json:"id"`
	Score    float32           `json:"score"` // Use "score" instead of "distance"
	Metadata map[string]string `json:"metadata"`
}

type VectorDB interface {
	UpsertEmbeddings(embeddings [][]float32, chunks []chunk.Chunk, uuid string) error
	UpsertEmbeddingsWithDocID(embeddings [][]float32, chunks []chunk.Chunk, uuid string, docID string) error
	Retrieve(questionEmbedding []float32, topK int, uuid string) ([]QueryMatch, error)
	HybridSearch(vectorQuery []float32, textQuery string, topK int, uuid string) ([]QueryMatch, error)
	DeleteByDocumentID(docID string, uuid string) error
	ListDocumentIDs(uuid string) ([]string, error)
}
