package postapi

import (
	"github.com/itaiguardiola/askara/storage"
	"github.com/itaiguardiola/askara/vectordb"

	cache "github.com/patrickmn/go-cache"
	openai "github.com/sashabaranov/go-openai"
)

type HandlerContext struct {
	openAIClient *openai.Client
	cache        *cache.Cache
	vectorDB     vectordb.VectorDB
	docStore     storage.DocumentStore
}

func NewHandlerContext(openAIClient *openai.Client, vectorDB vectordb.VectorDB, docStore storage.DocumentStore) *HandlerContext {
	return &HandlerContext{
		openAIClient: openAIClient,
		cache:        cache.New(cache.NoExpiration, cache.NoExpiration),
		vectorDB:     vectorDB,
		docStore:     docStore,
	}
}
