package postapi

import (
	"github.com/itaiguardiola/askara/codesourcetrust"
	"github.com/itaiguardiola/askara/llm"
	"github.com/itaiguardiola/askara/queryrewriter"
	"github.com/itaiguardiola/askara/storage"
	"github.com/itaiguardiola/askara/vectordb"

	cache "github.com/patrickmn/go-cache"
)

type HandlerContext struct {
	llmProvider     llm.LLMProvider
	cache           *cache.Cache
	vectorDB        vectordb.VectorDB
	docStore        storage.DocumentStore
	queryRewriter   *queryrewriter.QueryRewriter
	providerManager *llm.ProviderManager
	codeTrustSvc    *codesourcetrust.Service
}

func NewHandlerContext(llmProvider llm.LLMProvider, vectorDB vectordb.VectorDB, docStore storage.DocumentStore, queryRewriter *queryrewriter.QueryRewriter, codeTrustSvc *codesourcetrust.Service) *HandlerContext {
	return &HandlerContext{
		llmProvider:   llmProvider,
		cache:         cache.New(cache.NoExpiration, cache.NoExpiration),
		vectorDB:      vectorDB,
		docStore:      docStore,
		queryRewriter: queryRewriter,
		codeTrustSvc:  codeTrustSvc,
	}
}

// NewHandlerContextWithManager creates a handler context with a provider manager
func NewHandlerContextWithManager(providerManager *llm.ProviderManager, vectorDB vectordb.VectorDB, docStore storage.DocumentStore, queryRewriter *queryrewriter.QueryRewriter, codeTrustSvc *codesourcetrust.Service) *HandlerContext {
	// Get the initial provider from manager
	provider, _ := providerManager.GetProvider()

	return &HandlerContext{
		llmProvider:     provider,
		cache:           cache.New(cache.NoExpiration, cache.NoExpiration),
		vectorDB:        vectorDB,
		docStore:        docStore,
		queryRewriter:   queryRewriter,
		providerManager: providerManager,
		codeTrustSvc:    codeTrustSvc,
	}
}
