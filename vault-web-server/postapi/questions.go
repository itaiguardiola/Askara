package postapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/itaiguardiola/askara/form"
	openai "github.com/sashabaranov/go-openai"
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

	clientToUse := ctx.openAIClient
	if form.ApiKey != "" {
		log.Println("[QuestionHandler] Using provided custom API key:", form.ApiKey)
		clientToUse = openai.NewClient(form.ApiKey)
	}

	// step 1: Feed question to openai embeddings api to get an embedding back
	questionEmbedding, err := getEmbedding(clientToUse, form.Question, openai.AdaEmbeddingV2)
	if err != nil {
		log.Println("[QuestionHandler ERR] OpenAI get embedding request error\n", err.Error())
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
	prompt, err := buildPrompt(contextTexts, form.Question)
	if prompt == "" {
		prompt = form.Question
	}
	if err != nil {
		log.Println("[QuestionHandler ERR] Error building prompt\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	model := openai.GPT3Dot5Turbo
	if form.Model == "GPT Davinci" {
		model = openai.GPT3TextDavinci003
	}

	log.Printf("[QuestionHandler] Sending OpenAI api request...\nPrompt:%s\n", prompt)
	openAIResponse, tokens, err := callOpenAI(clientToUse, prompt, model,
		"You are a helpful assistant answering questions based on the context provided.",
		512)

	if err != nil {
		log.Println("[QuestionHandler ERR] OpenAI answer questions request error\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("[QuestionHandler] OpenAI response:\n", openAIResponse)
	response := OpenAIResponse{openAIResponse, tokens}

	answer := Answer{response.Response, contexts, response.Tokens}
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

	clientToUse := ctx.openAIClient
	if form.ApiKey != "" {
		log.Println("[StreamingQuestionHandler] Using provided custom API key")
		clientToUse = openai.NewClient(form.ApiKey)
	}

	// step 1: Feed question to openai embeddings api to get an embedding back
	questionEmbedding, err := getEmbedding(clientToUse, form.Question, openai.AdaEmbeddingV2)
	if err != nil {
		log.Println("[StreamingQuestionHandler ERR] OpenAI get embedding request error\n", err.Error())
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
	prompt, err := buildPrompt(contextTexts, form.Question)
	if prompt == "" {
		prompt = form.Question
	}
	if err != nil {
		log.Println("[StreamingQuestionHandler ERR] Error building prompt\n", err.Error())
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	model := openai.GPT3Dot5Turbo
	if form.Model == "GPT Davinci" {
		model = openai.GPT3TextDavinci003
	}

	// Streaming parameters
	temperature := float32(0.7)
	topP := float32(1.0)
	frequencyPenalty := float32(0.0)
	presencePenalty := float32(0.6)
	stop := []string{"Human:", "AI:"}

	log.Printf("[StreamingQuestionHandler] Sending OpenAI streaming request...\n")

	// Use streaming API
	err = useChatCompletionStreamAPI(
		clientToUse,
		prompt,
		model,
		"You are a helpful assistant answering questions based on the context provided.",
		temperature,
		512,
		topP,
		frequencyPenalty,
		presencePenalty,
		stop,
		func(chunk string) error {
			// Send each chunk as an SSE message
			fmt.Fprintf(w, "event: chunk\ndata: %s\n\n", chunk)
			flusher.Flush()
			return nil
		},
	)

	if err != nil {
		log.Println("[StreamingQuestionHandler ERR] OpenAI streaming error\n", err.Error())
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	// Send done event
	fmt.Fprintf(w, "event: done\ndata: {}\n\n")
	flusher.Flush()
	log.Println("[StreamingQuestionHandler] Stream completed")
}
