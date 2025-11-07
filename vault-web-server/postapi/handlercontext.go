package postapi

import (
	"github.com/itaiguardiola/askara/llm"
	"github.com/itaiguardiola/askara/storage"
	"github.com/itaiguardiola/askara/vectordb"

	cache "github.com/patrickmn/go-cache"
)

type HandlerContext struct {
	llmProvider llm.LLMProvider
	cache       *cache.Cache
	vectorDB    vectordb.VectorDB
	docStore    storage.DocumentStore
}

func NewHandlerContext(llmProvider llm.LLMProvider, vectorDB vectordb.VectorDB, docStore storage.DocumentStore) *HandlerContext {
	return &HandlerContext{
		llmProvider: llmProvider,
		cache:       cache.New(cache.NoExpiration, cache.NoExpiration),
		vectorDB:    vectorDB,
		docStore:    docStore,
	}
}
