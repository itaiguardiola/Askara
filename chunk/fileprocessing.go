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
	"github.com/gen2brain/go-fitz"
	"github.com/otiai10/gosseract/v2"
	"github.com/itaiguardiola/askara/mlworker"
	"image/png"
	"os"
	"path/filepath"
	"encoding/base64"
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
		fitzDoc, err := fitz.NewFromReader(bytes.NewReader(content))
		if err != nil {
			return "", fmt.Errorf("error reading .epub file: %v", err)
		}
		defer fitzDoc.Close()
		for i := 0; i < fitzDoc.NumPage(); i++ {
			pageText, err := fitzDoc.Text(i)
			if err != nil {
				return "", fmt.Errorf("error getting text from page %d: %v", i, err)
			}
			// Preprocess the text by replacing newline characters with spaces
			pageText = strings.ReplaceAll(pageText, "\n", " ")
			text += pageText
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

// extractTextWithOCR uses go-fitz to render pages and Tesseract for OCR
func extractTextWithOCR(f multipart.File) (string, error) {
	log.Println("[extractTextWithOCR] Using OCR fallback for PDF")

	// Reset file position
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("seek error: %v", err)
	}

	// Read file content
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("read error: %v", err)
	}

	// Open PDF with go-fitz
	doc, err := fitz.NewFromMemory(content)
	if err != nil {
		return "", fmt.Errorf("fitz error: %v", err)
	}
	defer doc.Close()

	var fullText strings.Builder
	numPages := doc.NumPage()
	log.Printf("[extractTextWithOCR] Processing %d pages with OCR", numPages)

	// Create temp directory for images
	tempDir, err := ioutil.TempDir("", "pdf-ocr-*")
	if err != nil {
		return "", fmt.Errorf("temp dir error: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Get ML Worker client
	mlClient := mlworker.GetClient()

	// Process each page
	for pageNum := 0; pageNum < numPages; pageNum++ {
		// Render page as image (150 DPI for good OCR quality)
		img, err := doc.Image(pageNum)
		if err != nil {
			log.Printf("[extractTextWithOCR] Page %d render error: %v", pageNum, err)
			continue
		}

		// Save image to temp file
		imgPath := filepath.Join(tempDir, fmt.Sprintf("page_%d.png", pageNum))
		imgFile, err := os.Create(imgPath)
		if err != nil {
			log.Printf("[extractTextWithOCR] Page %d create file error: %v", pageNum, err)
			continue
		}

		err = png.Encode(imgFile, img)
		imgFile.Close()
		if err != nil {
			log.Printf("[extractTextWithOCR] Page %d encode error: %v", pageNum, err)
			continue
		}

		var pageText string

		// Try ML Worker OCR if enabled
		if mlClient.IsFeatureEnabled("ocr") {
			log.Printf("[extractTextWithOCR] Page %d: Attempting ML Worker OCR", pageNum)

			// Read image and convert to base64 data URL
			imgData, err := ioutil.ReadFile(imgPath)
			if err == nil {
				b64Data := base64.StdEncoding.EncodeToString(imgData)
				dataURL := "data:image/png;base64," + b64Data

				// Call ML Worker OCR
				ocrResp, err := mlClient.OCR(dataURL, "eng", true)
				if err == nil && ocrResp != nil {
					pageText = ocrResp.Text
					log.Printf("[extractTextWithOCR] Page %d: ML Worker OCR success (%.2f%% confidence)",
						pageNum, ocrResp.Confidence*100)
				} else {
					log.Printf("[extractTextWithOCR] Page %d: ML Worker OCR failed: %v, falling back to Tesseract",
						pageNum, err)
				}
			}
		}

		// Fall back to local Tesseract if ML Worker didn't work
		if pageText == "" {
			log.Printf("[extractTextWithOCR] Page %d: Using local Tesseract OCR", pageNum)
			client := gosseract.NewClient()
			defer client.Close()

			client.SetImage(imgPath)
			client.SetLanguage("eng")

			pageText, err = client.Text()
			if err != nil {
				log.Printf("[extractTextWithOCR] Page %d OCR error: %v", pageNum, err)
				os.Remove(imgPath)
				continue
			}
		}

		// Add page text to result
		fullText.WriteString(pageText)
		fullText.WriteString("\n")

		// Clean up image file
		os.Remove(imgPath)
	}

	result := strings.TrimSpace(fullText.String())
	log.Printf("[extractTextWithOCR] OCR extracted %d characters", len(result))

	return result, nil
}

// tryMLWorkerParsing attempts to parse document using ML Worker (Docling)
func tryMLWorkerParsing(content []byte, fileType string) (string, error) {
	mlClient := mlworker.GetClient()

	if !mlClient.IsFeatureEnabled("parse") {
		return "", fmt.Errorf("ML Worker parse feature not enabled")
	}

	log.Println("[tryMLWorkerParsing] Attempting ML Worker document parsing (Docling)")

	// Convert file content to base64 data URL
	b64Data := base64.StdEncoding.EncodeToString(content)
	dataURL := fmt.Sprintf("data:%s;base64,%s", fileType, b64Data)

	// Call ML Worker parse endpoint
	parseResp, err := mlClient.ParseDocument(dataURL, true, true)
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

// extract human-readable text from a given pdf with hybrid approach: try ML Worker (Docling) -> docconv -> OCR
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

	// Reset file position for docconv
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	// Fallback to fast text extraction with docconv
	log.Println("[ExtractTextFromPDF] Attempting fast text extraction (docconv)")
	bodyResult, _, err := docconv.ConvertPDF(f)

	// If extraction succeeded and quality is good, use it
	if err == nil && !isTextExtractionPoor(bodyResult) {
		text := strings.TrimSpace(bodyResult)
		log.Printf("[ExtractTextFromPDF] Fast extraction successful: %d characters", len(text))
		return text, nil
	}

	// Text extraction failed or quality is poor - use OCR
	log.Println("[ExtractTextFromPDF] Text extraction poor or failed, falling back to OCR")
	return extractTextWithOCR(f)
}
