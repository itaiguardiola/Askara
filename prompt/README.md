# Prompt Forming Package

A sophisticated prompt engineering system for optimizing questions about documents in RAG (Retrieval Augmented Generation) applications.

## Features

- **Automatic Question Type Detection**: Identifies question types (summary, comparison, extraction, analytical, etc.)
- **Optimized Templates**: Different prompt templates for each question type
- **Smart Context Formatting**: Better organization of retrieved document chunks
- **Dynamic Temperature**: Suggests optimal temperature based on question type
- **System Instruction Generation**: Creates appropriate system prompts for different tasks
- **Flexible Configuration**: Customizable prompt building behavior

## Quick Start

```go
import "github.com/itaiguardiola/askara/prompt"

// Simple usage
question := "Summarize the main points of these documents"
contexts := []string{
    "Context chunk 1...",
    "Context chunk 2...",
}

result := prompt.QuickPrompt(question, contexts)
fmt.Println(result.QuestionType)  // "Summary"
fmt.Println(result.Temperature)    // 0.6 (optimized for summaries)
fmt.Println(result.FinalPrompt)    // Complete formatted prompt
```

## Question Types

The system automatically detects and optimizes for different question types:

1. **Direct** - Straightforward factual questions ("What is X?")
   - Temperature: 0.3 (factual)
   - Focus: Clear, accurate answers

2. **Summary** - Requests for summaries ("Summarize...", "Give me an overview...")
   - Temperature: 0.6 (slight creativity)
   - Focus: Conciseness and main points

3. **Comparison** - Comparing multiple things ("Compare X and Y", "What are the differences...")
   - Temperature: 0.5 (balanced)
   - Focus: Structured, balanced analysis

4. **Extraction** - Extract specific information ("List all...", "Find mentions of...")
   - Temperature: 0.2 (very deterministic)
   - Focus: Thoroughness and accuracy

5. **Analytical** - Deep analysis ("Why...", "How...", "Explain...")
   - Temperature: 0.7 (creative insights)
   - Focus: Reasoning and implications

6. **Conversational** - General chat or clarification
   - Temperature: 0.8 (natural)
   - Focus: Helpful, friendly responses

## Advanced Usage

### Using the Builder

```go
builder := prompt.NewBuilder()

// Build with custom configuration
builder = builder.
    WithSourceInfo(true).
    WithMaxContextLength(8000).
    WithTemperature(0.5)

result := builder.Build(question, contexts)
```

### Working with Context Chunks

```go
// Create context chunks with metadata
chunks := []prompt.ContextChunk{
    {
        Text:     "Content from document 1",
        Source:   "report.pdf",
        Score:    0.95,
        Position: 0,
    },
    {
        Text:     "Content from document 2",
        Source:   "analysis.docx",
        Score:    0.87,
        Position: 1,
    },
}

result := prompt.OptimizedPrompt(question, chunks)
```

### Custom Configuration

```go
config := prompt.PromptConfig{
    IncludeSourceInfo:   true,
    MaxContextLength:    10000,
    Temperature:         0.7,
    SystemInstructions:  "You are an expert analyst.",
    ContextSeparator:    "\n\n---\n\n",
    NumberContextChunks: true,
}

builder := prompt.NewBuilderWithConfig(config)
result := builder.Build(question, contexts)
```

## Configuration Options

- **IncludeSourceInfo** (bool): Add document source information to context
- **MaxContextLength** (int): Limit total context length in characters
- **Temperature** (float32): Suggested creativity level for LLM (0.0-1.0)
- **SystemInstructions** (string): Additional system-level instructions
- **ContextSeparator** (string): How to separate different context chunks
- **NumberContextChunks** (bool): Add numbers to each context chunk

## Prompt Result

The `PromptResult` struct contains:

```go
type PromptResult struct {
    FinalPrompt     string       // Complete prompt for the LLM
    SystemPrompt    string       // System-level instruction
    QuestionType    QuestionType // Detected question type
    ContextUsed     int          // Number of context chunks used
    EstimatedTokens int          // Rough token count estimate
    Config          PromptConfig // Configuration used
}
```

## Template Examples

### Summary Template

```
Based on the following context from documents, provide a comprehensive summary.

Context:
[Context 1] (Source: file.pdf, Relevance: 0.95)
Content here...

---

[Context 2] (Source: doc.docx, Relevance: 0.87)
More content...

Task: Summarize the main findings

Instructions:
- Create a well-structured summary with main points
- Maintain accuracy - only include information from the context
- Organize information logically
- Be concise but comprehensive
- Use bullet points or paragraphs as appropriate

Summary:
```

### Analytical Template

```
Based on the following context from documents, provide a thorough analysis.

Context:
[Context chunks...]

Question: Why did the project fail?

Instructions:
- Analyze the information deeply and thoughtfully
- Explain reasoning and connections
- Consider multiple perspectives if relevant
- Support your analysis with evidence from the context
- Discuss implications or significance when appropriate
- Be clear about what is stated vs. what is inferred

Analysis:
```

## Integration with LLM Providers

The prompt system is automatically integrated with both OpenAI and Ollama providers in the `llm` package:

```go
// Automatically uses optimized prompts
provider := llm.NewOpenAIProvider(client, config)
response, err := provider.GenerateCompletion("Summarize these documents", contexts)
```

## Helper Functions

### Detect Question Type

```go
questionType := prompt.DetectQuestionType("Compare the two approaches")
// Returns: QuestionTypeComparison
```

### Get Suggested Temperature

```go
temp := prompt.SuggestTemperature(prompt.QuestionTypeExtraction)
// Returns: 0.2 (for factual extraction)
```

### Get System Instructions

```go
instructions := prompt.GetSystemInstructions(prompt.QuestionTypeSummary)
// Returns: "You are a helpful assistant specializing in creating clear, concise summaries..."
```

### Truncate Context

```go
truncated := prompt.TruncateContext(contexts, maxTokens)
// Returns contexts that fit within token limit
```

## Best Practices

1. **Let the system detect question types** - The automatic detection is usually accurate
2. **Use context chunks with metadata** when available for better source attribution
3. **Set MaxContextLength** to prevent token limit issues
4. **Override temperature** only when you have specific requirements
5. **Add custom SystemInstructions** for domain-specific applications

## Examples

### Example 1: Document Summarization

```go
question := "Summarize the key findings from these research papers"
contexts := []string{
    "Paper 1 found that...",
    "Paper 2 concluded...",
}

result := prompt.QuickPrompt(question, contexts)
// Question type: Summary
// Temperature: 0.6
// Optimized prompt with summary-specific instructions
```

### Example 2: Factual Extraction

```go
question := "List all the companies mentioned in these documents"
contexts := loadDocumentChunks()

builder := prompt.NewBuilder().WithSourceInfo(true)
result := builder.Build(question, contexts)
// Question type: Extraction
// Temperature: 0.2
// Includes source attribution for each mention
```

### Example 3: Analytical Question

```go
question := "Why did the merger fail according to these reports?"
contexts := retrieveRelevantChunks(question)

result := prompt.OptimizedPrompt(question, contexts)
// Question type: Analytical
// Temperature: 0.7
// Prompts for deep analysis and reasoning
```

## License

Part of the Askara project.
