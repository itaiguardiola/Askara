package postapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/itaiguardiola/askara/form"
	"github.com/itaiguardiola/askara/llm"
	"github.com/itaiguardiola/askara/mlworker"
	"github.com/itaiguardiola/askara/vectordb"
)

type Context struct {
	Text  string `json:"text"`
	Title string `json:"title"`
}

type Answer struct {
	Answer  string    `json:"answer"`
	Context []Context `json:"context"`
	Tokens  int       `json:"tokens"`
}

// Handle Requests For Question
func (ctx *HandlerContext) QuestionHandler(w http.ResponseWriter, r *http.Request) {
	form := new(form.QuestionForm)

	if errs := FormParseVerify(form, "QuestionForm", w, r); errs != nil {
		return
	}

	log.Println("[QuestionHandler] Question:", form.Question)
	log.Println("[QuestionHandler] Model:", form.Model)
	log.Println("[QuestionHandler] UUID:", form.UUID)
	log.Println("[QuestionHandler] ApiKey:", form.ApiKey)

	providerToUse := ctx.llmProvider
	if form.ApiKey != "" {
		log.Println("[QuestionHandler] Using provided custom API key")
		// Create temporary OpenAI provider with custom key
		customProvider, err := llm.NewProvider(&llm.Config{
			Provider: "openai",
			OpenAIConfig: &llm.OpenAIConfig{
				APIKey: form.ApiKey,
			},
		})
		if err != nil {
			log.Println("[QuestionHandler ERR] Failed to create custom provider:", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		providerToUse = customProvider
	}

	// Optional: Rewrite query for better retrieval
	queriesToSearch := []string{form.Question}
	if ctx.queryRewriter.IsEnabled() {
		rewriteResult, err := ctx.queryRewriter.RewriteQuery(form.Question)
		if err != nil {
			log.Printf("[QuestionHandler WARN] Query rewriting failed, using original: %v", err)
		} else {
			queriesToSearch = rewriteResult.Variations
			log.Printf("[QuestionHandler] Using %d query variations for retrieval", len(queriesToSearch))
		}
	}

	// step 1: Generate embeddings for all query variations and retrieve matches using hybrid search
	allMatches := make([]vectordb.QueryMatch, 0)
	seenChunks := make(map[string]bool) // For deduplication

	for i, query := range queriesToSearch {
		queryEmbedding, err := providerToUse.GenerateEmbedding(query)
		if err != nil {
			log.Printf("[QuestionHandler WARN] Embedding failed for query %d: %v", i, err)
			continue
		}

		// Use hybrid search (vector + FTS) instead of just vector search
		matches, err := ctx.vectorDB.HybridSearch(queryEmbedding, query, 8, form.UUID)
		if err != nil {
			log.Printf("[QuestionHandler WARN] Hybrid search failed for query %d: %v", i, err)
			continue
		}

		// Deduplicate matches based on chunk ID
		for _, match := range matches {
			chunkID := match.ID
			if !seenChunks[chunkID] {
				seenChunks[chunkID] = true
				allMatches = append(allMatches, match)
			}
		}
	}

	log.Printf("[QuestionHandler] Retrieved %d unique matches using hybrid search", len(allMatches))

	// Optional: Rerank using ML Worker for better relevance
	mlClient := mlworker.GetClient()
	if mlClient.IsFeatureEnabled("rerank") && len(allMatches) > 0 {
		// Prepare documents for reranking
		rerankDocs := make([]mlworker.RerankDocument, len(allMatches))
		for i, match := range allMatches {
			rerankDocs[i] = mlworker.RerankDocument{
				ID:   match.ID,
				Text: match.Metadata["text"],
			}
		}

		// Rerank to get top 6 results
		reranked, err := mlClient.Rerank(form.Question, rerankDocs, 6)
		if err != nil {
			log.Printf("[QuestionHandler WARN] Reranking failed, using hybrid search order: %v", err)
		} else {
			// Reorder matches based on reranking results
			reorderedMatches := make([]vectordb.QueryMatch, 0, len(reranked.Results))
			for _, result := range reranked.Results {
				// Find the match with this ID
				for _, match := range allMatches {
					if match.ID == result.ID {
						// Update score with reranker score
						match.Score = float32(result.RelevanceScore)
						reorderedMatches = append(reorderedMatches, match)
						break
					}
				}
			}
			allMatches = reorderedMatches
			log.Printf("[QuestionHandler] Reranked to %d results using ML Worker", len(allMatches))
		}
	}

	// Limit to top 6 matches if we have more (after reranking or from hybrid search)
	matches := allMatches
	if len(matches) > 6 {
		matches = matches[:6]
	}

	// Extract context text and titles from the matches
	contexts := make([]Context, len(matches))
	for i, match := range matches {
		contexts[i].Text = match.Metadata["text"]
		contexts[i].Title = match.Metadata["title"]
	}
	log.Println("[QuestionHandler] Retrieved context from vector DB:\n", contexts)

	// Enhance contexts with code-aware grounding (if enabled)
	if ctx.codeTrustSvc != nil {
		// Convert contexts to map format for enhancement
		contextMaps := make([]map[string]interface{}, len(contexts))
		for i, c := range contexts {
			contextMaps[i] = map[string]interface{}{
				"text":  c.Text,
				"title": c.Title,
			}
		}

		enhancedMaps, err := ctx.codeTrustSvc.EnhanceContext(form.UUID, form.Question, contextMaps)
		if err != nil {
			log.Printf("[QuestionHandler WARN] Code grounding failed: %v", err)
		} else if len(enhancedMaps) > len(contextMaps) {
			// Grounding metadata was added - convert back to Context format
			contexts = make([]Context, len(enhancedMaps))
			for i, m := range enhancedMaps {
				contexts[i].Text = m["text"].(string)
				contexts[i].Title = m["title"].(string)
			}
			log.Printf("[QuestionHandler] Enhanced with code-aware grounding")
		}
	}

	// step 3: Structure the prompt with a context section + question, using top x results from vector DB as the context
	contextTexts := make([]string, len(contexts))
	for i, context := range contexts {
		contextTexts[i] = context.Text
	}

	log.Printf("[QuestionHandler] Sending LLM api request for question: %s\n", form.Question)
	llmResponse, err := providerToUse.GenerateCompletion(form.Question, contextTexts)

	if err != nil {
		log.Println("[QuestionHandler ERR] LLM answer questions request error\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("[QuestionHandler] LLM response:\n", llmResponse)
	// Note: Token counting is not supported in the LLM interface for all providers
	tokens := 0

	answer := Answer{llmResponse, contexts, tokens}
	jsonResponse, err := json.Marshal(answer)
	if err != nil {
		log.Println("[QuestionHandler ERR] OpenAI response marshalling error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

// StreamingQuestionHandler handles streaming question requests using Server-Sent Events
func (ctx *HandlerContext) StreamingQuestionHandler(w http.ResponseWriter, r *http.Request) {
	form := new(form.QuestionForm)

	if errs := FormParseVerify(form, "QuestionForm", w, r); errs != nil {
		return
	}

	log.Println("[StreamingQuestionHandler] Question:", form.Question)
	log.Println("[StreamingQuestionHandler] Model:", form.Model)
	log.Println("[StreamingQuestionHandler] UUID:", form.UUID)

	// Set headers for Server-Sent Events
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	providerToUse := ctx.llmProvider
	if form.ApiKey != "" {
		log.Println("[StreamingQuestionHandler] Using provided custom API key")
		// Create temporary OpenAI provider with custom key
		customProvider, err := llm.NewProvider(&llm.Config{
			Provider: "openai",
			OpenAIConfig: &llm.OpenAIConfig{
				APIKey: form.ApiKey,
			},
		})
		if err != nil {
			log.Println("[StreamingQuestionHandler ERR] Failed to create custom provider:", err.Error())
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
			flusher.Flush()
			return
		}
		providerToUse = customProvider
	}

	// Optional: Rewrite query for better retrieval
	queriesToSearch := []string{form.Question}
	if ctx.queryRewriter.IsEnabled() {
		rewriteResult, err := ctx.queryRewriter.RewriteQuery(form.Question)
		if err != nil {
			log.Printf("[StreamingQuestionHandler WARN] Query rewriting failed, using original: %v", err)
		} else {
			queriesToSearch = rewriteResult.Variations
			log.Printf("[StreamingQuestionHandler] Using %d query variations for retrieval", len(queriesToSearch))
		}
	}

	// step 1: Generate embeddings for all query variations and retrieve matches using hybrid search
	allMatches := make([]vectordb.QueryMatch, 0)
	seenChunks := make(map[string]bool) // For deduplication

	for i, query := range queriesToSearch {
		queryEmbedding, err := providerToUse.GenerateEmbedding(query)
		if err != nil {
			log.Printf("[StreamingQuestionHandler WARN] Embedding failed for query %d: %v", i, err)
			continue
		}

		// Use hybrid search (vector + FTS) instead of just vector search
		matches, err := ctx.vectorDB.HybridSearch(queryEmbedding, query, 8, form.UUID)
		if err != nil {
			log.Printf("[StreamingQuestionHandler WARN] Hybrid search failed for query %d: %v", i, err)
			continue
		}

		// Deduplicate matches based on chunk ID
		for _, match := range matches {
			chunkID := match.ID
			if !seenChunks[chunkID] {
				seenChunks[chunkID] = true
				allMatches = append(allMatches, match)
			}
		}
	}

	log.Printf("[StreamingQuestionHandler] Retrieved %d unique matches using hybrid search", len(allMatches))

	// Optional: Rerank using ML Worker for better relevance
	mlClient := mlworker.GetClient()
	if mlClient.IsFeatureEnabled("rerank") && len(allMatches) > 0 {
		// Prepare documents for reranking
		rerankDocs := make([]mlworker.RerankDocument, len(allMatches))
		for i, match := range allMatches {
			rerankDocs[i] = mlworker.RerankDocument{
				ID:   match.ID,
				Text: match.Metadata["text"],
			}
		}

		// Rerank to get top 6 results
		reranked, err := mlClient.Rerank(form.Question, rerankDocs, 6)
		if err != nil {
			log.Printf("[StreamingQuestionHandler WARN] Reranking failed, using hybrid search order: %v", err)
		} else {
			// Reorder matches based on reranking results
			reorderedMatches := make([]vectordb.QueryMatch, 0, len(reranked.Results))
			for _, result := range reranked.Results {
				// Find the match with this ID
				for _, match := range allMatches {
					if match.ID == result.ID {
						// Update score with reranker score
						match.Score = float32(result.RelevanceScore)
						reorderedMatches = append(reorderedMatches, match)
						break
					}
				}
			}
			allMatches = reorderedMatches
			log.Printf("[StreamingQuestionHandler] Reranked to %d results using ML Worker", len(allMatches))
		}
	}

	// Limit to top 6 matches if we have more (after reranking or from hybrid search)
	matches := allMatches
	if len(matches) > 6 {
		matches = matches[:6]
	}

	// Extract context text and titles from the matches
	contexts := make([]Context, len(matches))
	for i, match := range matches {
		contexts[i].Text = match.Metadata["text"]
		contexts[i].Title = match.Metadata["title"]
	}

	// Enhance contexts with code-aware grounding (if enabled)
	if ctx.codeTrustSvc != nil {
		// Convert contexts to map format for enhancement
		contextMaps := make([]map[string]interface{}, len(contexts))
		for i, c := range contexts {
			contextMaps[i] = map[string]interface{}{
				"text":  c.Text,
				"title": c.Title,
			}
		}

		enhancedMaps, err := ctx.codeTrustSvc.EnhanceContext(form.UUID, form.Question, contextMaps)
		if err != nil {
			log.Printf("[StreamingQuestionHandler WARN] Code grounding failed: %v", err)
		} else if len(enhancedMaps) > len(contextMaps) {
			// Grounding metadata was added - convert back to Context format
			contexts = make([]Context, len(enhancedMaps))
			for i, m := range enhancedMaps {
				contexts[i].Text = m["text"].(string)
				contexts[i].Title = m["title"].(string)
			}
			log.Printf("[StreamingQuestionHandler] Enhanced with code-aware grounding")
		}
	}

	// Send context to the client first
	contextsJSON, err := json.Marshal(contexts)
	if err != nil {
		log.Println("[StreamingQuestionHandler ERR] Context marshalling error", err)
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}
	fmt.Fprintf(w, "event: context\ndata: %s\n\n", string(contextsJSON))
	flusher.Flush()

	// step 3: Structure the prompt with a context section + question
	contextTexts := make([]string, len(contexts))
	for i, context := range contexts {
		contextTexts[i] = context.Text
	}

	log.Printf("[StreamingQuestionHandler] Sending LLM streaming request...\n")

	// Use streaming API
	err = providerToUse.StreamCompletion(
		form.Question,
		contextTexts,
		func(chunk string) {
			// Send each chunk as an SSE message
			fmt.Fprintf(w, "event: chunk\ndata: %s\n\n", chunk)
			flusher.Flush()
		},
	)

	if err != nil {
		log.Println("[StreamingQuestionHandler ERR] LLM streaming error\n", err.Error())
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	// Send done event
	fmt.Fprintf(w, "event: done\ndata: {}\n\n")
	flusher.Flush()
	log.Println("[StreamingQuestionHandler] Stream completed")
}
