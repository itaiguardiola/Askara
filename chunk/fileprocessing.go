package chunk

import (
	"bytes"
	"log"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"strings"

	"github.com/neurosnap/sentences/english"
	tke "github.com/pkoukk/tiktoken-go"
	"golang.org/x/text/transform"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"github.com/gabriel-vasile/mimetype"


	"github.com/saintfish/chardet"


	"io"

	"code.sajari.com/docconv"
	"github.com/itaiguardiola/askara/mlworker"
)

type Chunk struct {
	Start   int
	End     int
	Title   string
	Text    string
	Page    int    // Estimated page number
	Heading string // Last heading before this chunk (from markdown)
}

// MaxTokensPerChunk is the maximum number of tokens allowed in a single chunk for OpenAI embeddings
// MaxTokensPerChunk is the maximum number of tokens allowed in a single chunk for OpenAI embeddings
const MaxTokensPerChunk = 1500
const EmbeddingModel = "text-embedding-ada-002"
const CharsPerPage = 3000 // Rough estimate: ~3000 chars per page

// extractLastHeading finds the last markdown heading before a given position
func extractLastHeading(text string, position int) string {
	// Look backwards from position for markdown headings (# or ##)
	lines := strings.Split(text[:position], "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ") {
			// Remove markdown symbols
			heading := strings.TrimPrefix(line, "## ")
			heading = strings.TrimPrefix(heading, "# ")
			return strings.TrimSpace(heading)
		}
	}
	return ""
}

func CreateChunks(fileContent string, title string) ([]Chunk, error) {
	chunks := []Chunk{}

	// Initialize sentence tokenizer
	tokenizer, _ := english.NewSentenceTokenizer(nil)
	sentences := tokenizer.Tokenize(fileContent)

	// Get tiktoken encoding for the model
	tiktoken, err := tke.EncodingForModel(EmbeddingModel)
	if err != nil {
		return []Chunk{}, fmt.Errorf("getEncoding: %v", err)
	}

	chunkStart := 0

	for chunkStart < len(sentences) {
		tokenCount := 0
		chunkText := ""
		chunkSentences := 0

		for i := chunkStart; i < len(sentences) && tokenCount < MaxTokensPerChunk; i++ {
			sentence := sentences[i].Text
			tiktokens := tiktoken.Encode(sentence, nil, nil)
			sentenceTokenCount := len(tiktokens)

			if sentenceTokenCount > MaxTokensPerChunk {
				continue // Skip sentence if longer than MaxTokensPerChunk
			}

			if tokenCount+sentenceTokenCount <= MaxTokensPerChunk {
				tokenCount += sentenceTokenCount
				chunkText += " " + sentence
				chunkSentences++
			} else {
				break
			}
		}

		trimmedText := strings.TrimSpace(chunkText)
		if len(trimmedText) > 0 {
			// Estimate page number based on character position
			charPosition := 0
			for i := 0; i < chunkStart && i < len(sentences); i++ {
				charPosition += len(sentences[i].Text)
			}
			page := (charPosition / CharsPerPage) + 1

			// Extract last heading before this chunk
			heading := extractLastHeading(fileContent, charPosition)

			chunks = append(chunks, Chunk{
				Start:   chunkStart,
				End:     chunkStart + tokenCount,
				Title:   title,
				Text:    trimmedText,
				Page:    page,
				Heading: heading,
			})
		}

		// Calculate stride dynamically based on chunk sentences
		sentenceStride := chunkSentences / 5
		if sentenceStride == 0 {
			sentenceStride = 1
		}

		// Move chunkStart forward by sentenceStride
		chunkStart += sentenceStride
	}

	if len(chunks) == 0 {
		return nil, errors.New("no chunks created")
	}

	return chunks, nil
}


func GetTextFromFile(f multipart.File) (string, error) {
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	mime := mimetype.Detect(content)
	contentType := mime.String()

	var text string

	log.Println("[GetTextfromFile] ContentType:", contentType)
	switch contentType {
	case "application/msword": // .doc
		log.Println("[GetTextfromFile] .doc file encountered...")
		text, _, err = docconv.ConvertDoc(bytes.NewReader(content))
		if err != nil {
			return "", fmt.Errorf("error converting .doc file")
		}
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document": // .docx
		log.Println("[GetTextfromFile] .docx file encountered...")
		text, _, err = docconv.ConvertDocx(bytes.NewReader(content))
		if err != nil {
			return "", fmt.Errorf("error converting .docx file: %v", err)
		}
	case "application/zip": // .pages
		log.Println("[GetTextfromFile] .pages file encountered...")
		text, _, err = docconv.ConvertPages(bytes.NewReader(content))
		if err != nil {
			return "", fmt.Errorf("error converting .pages file: %v", err)
		}
		text = strings.TrimSpace(text)
	case "application/epub+zip": // .epub
		log.Println("[GetTextfromFile] .epub file encountered...")
		text, err = extractEPUBTextLocal(content)
		if err != nil {
			return "", fmt.Errorf("error reading .epub file: %v", err)
		}
	default: // Assume plain text
		detector := chardet.NewTextDetector()
		result, err := detector.DetectBest(content)
		if err != nil {
			return "", fmt.Errorf("error detecting encoding: %v", err)
		}

		if strings.ToLower(result.Charset) == "utf-8" {
			text = string(content)
		} else {
			var enc encoding.Encoding
			switch strings.ToLower(result.Charset) {
			case "iso-8859-1":
				enc = charmap.ISO8859_1
			case "windows-1252":
				enc = charmap.Windows1252
			// Add more encodings here as needed
			default:
				return "", fmt.Errorf("unsupported encoding: %s", result.Charset)
			}

			text, _, err = transform.String(enc.NewDecoder(), string(content))
			if err != nil {
				return "", fmt.Errorf("error decoding content: %v", err)
			}
		}
	}

	return text, nil
}

// isTextExtractionPoor checks if extracted text is of poor quality
func isTextExtractionPoor(text string) bool {
	trimmed := strings.TrimSpace(text)

	// Empty or very short text indicates poor extraction
	if len(trimmed) < 50 {
		return true
	}

	// Count alphanumeric characters
	alphanumCount := 0
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			alphanumCount++
		}
	}

	// If less than 60% alphanumeric, likely gibberish or failed extraction
	alphanumRatio := float64(alphanumCount) / float64(len(trimmed))
	if alphanumRatio < 0.6 {
		return true
	}

	return false
}

// tryMLWorkerParsing attempts to parse document using ML Worker (Docling)
func tryMLWorkerParsing(content []byte, fileType string) (string, error) {
	mlClient := mlworker.GetClient()

	if !mlClient.IsFeatureEnabled("parse") {
		return "", fmt.Errorf("ML Worker parse feature not enabled")
	}

	log.Printf("[tryMLWorkerParsing] Attempting ML Worker document parsing (Docling) - file size: %d bytes", len(content))

	// Use multipart upload for all files (supports large files)
	// Generate a filename based on file type
	filename := "document"
	switch fileType {
	case "application/pdf":
		filename = "document.pdf"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		filename = "document.docx"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		filename = "document.pptx"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		filename = "document.xlsx"
	default:
		filename = "document.bin"
	}

	// Call ML Worker parse endpoint with multipart upload
	parseResp, err := mlClient.ParseDocumentMultipart(content, filename, fileType, true, true)
	if err != nil {
		return "", fmt.Errorf("ML Worker parsing failed: %w", err)
	}

	log.Printf("[tryMLWorkerParsing] Successfully parsed: %d pages, %d chars, %d tables (%dms)",
		parseResp.PageCount, len(parseResp.Markdown), len(parseResp.Tables), parseResp.ProcessingTimeMS)

	// Use markdown output for better structure preservation
	fullText := parseResp.Markdown

	// If tables were extracted, append them to the text
	if len(parseResp.Tables) > 0 {
		fullText += "\n\n=== EXTRACTED TABLES ===\n"
		for i, table := range parseResp.Tables {
			if table.Caption != "" {
				fullText += fmt.Sprintf("\nTable %d: %s\n", i+1, table.Caption)
			} else {
				fullText += fmt.Sprintf("\nTable %d:\n", i+1)
			}

			// Add headers
			if len(table.Headers) > 0 {
				fullText += strings.Join(table.Headers, " | ") + "\n"
				fullText += strings.Repeat("-", len(strings.Join(table.Headers, " | "))) + "\n"
			}

			// Add rows
			for _, row := range table.Rows {
				fullText += strings.Join(row, " | ") + "\n"
			}
			fullText += "\n"
		}
	}

	return fullText, nil
}

// tryPyMuPDFParsing attempts to parse PDF using PyMuPDF4LLM via ML Worker
func tryPyMuPDFParsing(content []byte) (string, error) {
	mlClient := mlworker.GetClient()

	if !mlClient.IsFeatureEnabled("pymupdf") {
		return "", fmt.Errorf("PyMuPDF feature not enabled")
	}

	log.Printf("[tryPyMuPDFParsing] Attempting PyMuPDF4LLM extraction - file size: %d bytes", len(content))

	parseResp, err := mlClient.ParseDocumentPyMuPDF(content, "document.pdf")
	if err != nil {
		return "", fmt.Errorf("PyMuPDF extraction failed: %w", err)
	}

	log.Printf("[tryPyMuPDFParsing] Successfully parsed: %d pages, %d chars (%dms)",
		parseResp.PageCount, len(parseResp.Markdown), parseResp.ProcessingTimeMS)

	// Use markdown output for better LLM consumption
	return parseResp.Markdown, nil
}

// extract human-readable text from a given pdf with hybrid approach: try ML Worker (Docling) -> PyMuPDF4LLM -> docconv -> OCR
func ExtractTextFromPDF(f multipart.File, fileSize int64) (string, error) {
	// Reset the file reader's position
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	// Read file content for ML Worker
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("read error: %v", err)
	}

	// Try ML Worker parsing first (Docling) if enabled
	text, err := tryMLWorkerParsing(content, "application/pdf")
	if err == nil && !isTextExtractionPoor(text) {
		log.Printf("[ExtractTextFromPDF] ML Worker parsing successful: %d characters", len(text))
		return text, nil
	}
	if err != nil {
		log.Printf("[ExtractTextFromPDF] ML Worker parsing unavailable: %v", err)
	}

	// Try PyMuPDF4LLM as second option (better than docconv)
	text, err = tryPyMuPDFParsing(content)
	if err == nil && !isTextExtractionPoor(text) {
		log.Printf("[ExtractTextFromPDF] PyMuPDF4LLM extraction successful: %d characters", len(text))
		return text, nil
	}
	if err != nil {
		log.Printf("[ExtractTextFromPDF] PyMuPDF4LLM unavailable: %v", err)
	}

	// Reset file position for docconv
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	// Fallback to fast text extraction with docconv
	log.Println("[ExtractTextFromPDF] Attempting fast text extraction (docconv)")
	bodyResult, _, err := docconv.ConvertPDF(f)

	// Log what we got
	if err != nil {
		log.Printf("[ExtractTextFromPDF] Docconv failed: %v", err)
	} else {
		log.Printf("[ExtractTextFromPDF] Docconv extracted %d characters", len(bodyResult))
		isPoor := isTextExtractionPoor(bodyResult)
		log.Printf("[ExtractTextFromPDF] Quality check: isPoor=%v", isPoor)

		// If extraction succeeded and quality is good, use it
		if !isPoor {
			text := strings.TrimSpace(bodyResult)
			log.Printf("[ExtractTextFromPDF] Fast extraction successful: %d characters", len(text))
			return text, nil
		}

		// Quality is poor - try OCR before accepting poor quality text
		log.Println("[ExtractTextFromPDF] Docconv text quality is poor, trying ML Worker OCR")

		// Reset file position for OCR
		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			// Can't reset file, use poor quality docconv text as fallback
			text := strings.TrimSpace(bodyResult)
			if len(text) > 0 {
				log.Printf("[ExtractTextFromPDF] File seek failed, using docconv text: %d characters", len(text))
				return text, nil
			}
			return "", fmt.Errorf("file seek error: %v", err)
		}

		ocrText, ocrErr := extractTextWithOCRLocal(f)
		if ocrErr == nil && !isTextExtractionPoor(ocrText) {
			log.Printf("[ExtractTextFromPDF] ML Worker OCR successful: %d characters", len(ocrText))
			return ocrText, nil
		}

		if ocrErr != nil {
			log.Printf("[ExtractTextFromPDF] ML Worker OCR failed: %v", ocrErr)
		} else {
			log.Printf("[ExtractTextFromPDF] ML Worker OCR text also poor quality")
		}

		// OCR didn't work, use docconv text as last resort
		text := strings.TrimSpace(bodyResult)
		if len(text) > 0 {
			log.Printf("[ExtractTextFromPDF] Using docconv text as last resort: %d characters", len(text))
			return text, nil
		}
	}

	// Text extraction completely failed - return error
	return "", fmt.Errorf("all extraction methods failed")
}
