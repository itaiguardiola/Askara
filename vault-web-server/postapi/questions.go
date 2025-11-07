package postapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/itaiguardiola/askara/form"
	"github.com/itaiguardiola/askara/llm"
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

	// step 1: Feed question to LLM to get an embedding back
	questionEmbedding, err := providerToUse.GenerateEmbedding(form.Question)
	if err != nil {
		log.Println("[QuestionHandler ERR] LLM get embedding request error\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("[QuestionHandler] Question Embedding Length:", len(questionEmbedding))

	// step 2: Query vector db using questionEmbedding to get context matches
	matches, err := ctx.vectorDB.Retrieve(questionEmbedding, 4, form.UUID)
	if err != nil {
		log.Println("[QuestionHandler ERR] Vector DB query error\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("[QuestionHandler] Got matches from vector DB:", matches)

	// Extract context text and titles from the matches
	contexts := make([]Context, len(matches))
	for i, match := range matches {
		contexts[i].Text = match.Metadata["text"]
		contexts[i].Title = match.Metadata["title"]
	}
	log.Println("[QuestionHandler] Retrieved context from vector DB:\n", contexts)

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

	// step 1: Feed question to LLM to get an embedding back
	questionEmbedding, err := providerToUse.GenerateEmbedding(form.Question)
	if err != nil {
		log.Println("[StreamingQuestionHandler ERR] LLM get embedding request error\n", err.Error())
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	// step 2: Query vector db using questionEmbedding to get context matches
	matches, err := ctx.vectorDB.Retrieve(questionEmbedding, 4, form.UUID)
	if err != nil {
		log.Println("[StreamingQuestionHandler ERR] Vector DB query error\n", err.Error())
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	// Extract context text and titles from the matches
	contexts := make([]Context, len(matches))
	for i, match := range matches {
		contexts[i].Text = match.Metadata["text"]
		contexts[i].Title = match.Metadata["title"]
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
