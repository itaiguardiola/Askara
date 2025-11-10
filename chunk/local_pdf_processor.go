//go:build local_pdf_ocr
// +build local_pdf_ocr

package chunk

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
	"github.com/itaiguardiola/askara/mlworker"
	"github.com/otiai10/gosseract/v2"
	"image/png"
)

// extractEPUBTextLocal extracts text from EPUB files using fitz library
func extractEPUBTextLocal(content []byte) (string, error) {
	log.Println("[extractEPUBTextLocal] Using local fitz library for EPUB")
	fitzDoc, err := fitz.NewFromReader(bytes.NewReader(content))
	if err != nil {
		return "", fmt.Errorf("error reading .epub file: %v", err)
	}
	defer fitzDoc.Close()

	var text string
	for i := 0; i < fitzDoc.NumPage(); i++ {
		pageText, err := fitzDoc.Text(i)
		if err != nil {
			return "", fmt.Errorf("error getting text from page %d: %v", i, err)
		}
		// Preprocess the text by replacing newline characters with spaces
		pageText = strings.ReplaceAll(pageText, "\n", " ")
		text += pageText
	}
	return text, nil
}

// extractTextWithOCRLocal uses go-fitz to render pages and Tesseract for OCR (with ML Worker fallback)
func extractTextWithOCRLocal(f multipart.File) (string, error) {
	log.Println("[extractTextWithOCRLocal] Using local fitz + ML Worker OCR")

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
	log.Printf("[extractTextWithOCRLocal] Processing %d pages with OCR", numPages)

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
			log.Printf("[extractTextWithOCRLocal] Page %d render error: %v", pageNum, err)
			continue
		}

		// Save image to temp file
		imgPath := filepath.Join(tempDir, fmt.Sprintf("page_%d.png", pageNum))
		imgFile, err := os.Create(imgPath)
		if err != nil {
			log.Printf("[extractTextWithOCRLocal] Page %d create file error: %v", pageNum, err)
			continue
		}

		err = png.Encode(imgFile, img)
		imgFile.Close()
		if err != nil {
			log.Printf("[extractTextWithOCRLocal] Page %d encode error: %v", pageNum, err)
			continue
		}

		var pageText string

		// Try ML Worker OCR if enabled
		if mlClient.IsFeatureEnabled("ocr") {
			log.Printf("[extractTextWithOCRLocal] Page %d: Attempting ML Worker OCR", pageNum)

			// Read image and convert to base64 data URL
			imgData, err := ioutil.ReadFile(imgPath)
			if err == nil {
				b64Data := base64.StdEncoding.EncodeToString(imgData)
				dataURL := "data:image/png;base64," + b64Data

				// Call ML Worker OCR
				ocrResp, err := mlClient.OCR(dataURL, "eng", true)
				if err == nil && ocrResp != nil {
					pageText = ocrResp.Text
					log.Printf("[extractTextWithOCRLocal] Page %d: ML Worker OCR success (%.2f%% confidence)",
						pageNum, ocrResp.Confidence*100)
				} else {
					log.Printf("[extractTextWithOCRLocal] Page %d: ML Worker OCR failed: %v, falling back to Tesseract",
						pageNum, err)
				}
			}
		}

		// Fall back to local Tesseract if ML Worker didn't work
		if pageText == "" {
			log.Printf("[extractTextWithOCRLocal] Page %d: Using local Tesseract OCR", pageNum)
			client := gosseract.NewClient()
			defer client.Close()

			client.SetImage(imgPath)
			client.SetLanguage("eng")

			pageText, err = client.Text()
			if err != nil {
				log.Printf("[extractTextWithOCRLocal] Page %d OCR error: %v", pageNum, err)
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
	log.Printf("[extractTextWithOCRLocal] OCR extracted %d characters", len(result))

	return result, nil
}
