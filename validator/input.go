package validator

import (
	"crypto/rand"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Common validation errors
var (
	ErrInvalidUUID       = errors.New("invalid UUID format")
	ErrInvalidDocumentID = errors.New("invalid document ID format")
	ErrInvalidFilename   = errors.New("invalid filename")
	ErrFilenameTooLong   = errors.New("filename exceeds maximum length")
	ErrPathTraversal     = errors.New("path traversal attempt detected")
)

const (
	MaxFilenameLength  = 255
	MaxDocumentIDLength = 100
	MaxUUIDLength      = 36
)

// UUID validation regex (RFC 4122 compliant)
var uuidRegex = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[1-5][a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`)

// Document ID validation regex (doc-[hexstring])
var docIDRegex = regexp.MustCompile(`^doc-[a-f0-9]{16}$`)

// ValidateUUID checks if a UUID string is valid according to RFC 4122
func ValidateUUID(uuid string) error {
	if uuid == "" {
		return fmt.Errorf("%w: empty UUID", ErrInvalidUUID)
	}

	if len(uuid) > MaxUUIDLength {
		return fmt.Errorf("%w: UUID too long (max %d characters)", ErrInvalidUUID, MaxUUIDLength)
	}

	// Convert to lowercase for validation
	uuid = strings.ToLower(uuid)

	if !uuidRegex.MatchString(uuid) {
		return fmt.Errorf("%w: must match RFC 4122 format", ErrInvalidUUID)
	}

	return nil
}

// ValidateDocumentID checks if a document ID is valid
func ValidateDocumentID(docID string) error {
	if docID == "" {
		return fmt.Errorf("%w: empty document ID", ErrInvalidDocumentID)
	}

	if len(docID) > MaxDocumentIDLength {
		return fmt.Errorf("%w: document ID too long (max %d characters)", ErrInvalidDocumentID, MaxDocumentIDLength)
	}

	if !docIDRegex.MatchString(docID) {
		return fmt.Errorf("%w: must match format 'doc-[hexstring]'", ErrInvalidDocumentID)
	}

	return nil
}

// ValidateFilename checks if a filename is safe and valid
func ValidateFilename(filename string) error {
	if filename == "" {
		return fmt.Errorf("%w: empty filename", ErrInvalidFilename)
	}

	if len(filename) > MaxFilenameLength {
		return fmt.Errorf("%w: filename exceeds %d characters", ErrFilenameTooLong, MaxFilenameLength)
	}

	// Check for path traversal attempts
	if strings.Contains(filename, "..") {
		return fmt.Errorf("%w: filename contains '..'", ErrPathTraversal)
	}

	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("%w: filename contains path separators", ErrInvalidFilename)
	}

	// Check for null bytes
	if strings.Contains(filename, "\x00") {
		return fmt.Errorf("%w: filename contains null bytes", ErrInvalidFilename)
	}

	// Check for control characters
	for _, r := range filename {
		if r < 32 && r != 9 { // Allow tab (9) but not other control chars
			return fmt.Errorf("%w: filename contains control characters", ErrInvalidFilename)
		}
	}

	return nil
}

// SanitizeFilename removes or replaces unsafe characters from a filename
func SanitizeFilename(filename string) string {
	// Remove path components
	filename = filepath.Base(filename)

	// Replace path separators
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")

	// Remove null bytes
	filename = strings.ReplaceAll(filename, "\x00", "")

	// Remove control characters except tab
	result := strings.Builder{}
	for _, r := range filename {
		if r >= 32 || r == 9 {
			result.WriteRune(r)
		}
	}

	filename = result.String()

	// Trim whitespace
	filename = strings.TrimSpace(filename)

	// Limit length
	if len(filename) > MaxFilenameLength {
		// Try to preserve extension
		ext := filepath.Ext(filename)
		nameWithoutExt := strings.TrimSuffix(filename, ext)
		maxNameLength := MaxFilenameLength - len(ext)
		if maxNameLength > 0 && len(nameWithoutExt) > maxNameLength {
			filename = nameWithoutExt[:maxNameLength] + ext
		} else {
			filename = filename[:MaxFilenameLength]
		}
	}

	// If empty after sanitization, use a default name
	if filename == "" {
		filename = "unnamed_file"
	}

	return filename
}

// ValidatePath checks if a path is safe and doesn't attempt traversal
func ValidatePath(basePath, requestedPath string) error {
	// Clean both paths
	cleanBase := filepath.Clean(basePath)
	cleanRequested := filepath.Clean(requestedPath)

	// Ensure requested path is within base path
	if !strings.HasPrefix(cleanRequested, cleanBase) {
		return fmt.Errorf("%w: path outside base directory", ErrPathTraversal)
	}

	// Check for symlink traversal (additional security)
	evalBase, err := filepath.EvalSymlinks(cleanBase)
	if err != nil {
		return fmt.Errorf("failed to evaluate base path: %w", err)
	}

	evalRequested, err := filepath.EvalSymlinks(cleanRequested)
	if err != nil {
		// Path might not exist yet, just check the parent
		parentPath := filepath.Dir(cleanRequested)
		evalParent, err := filepath.EvalSymlinks(parentPath)
		if err != nil {
			return fmt.Errorf("failed to evaluate parent path: %w", err)
		}
		if !strings.HasPrefix(evalParent, evalBase) {
			return fmt.Errorf("%w: symlink outside base directory", ErrPathTraversal)
		}
	} else {
		if !strings.HasPrefix(evalRequested, evalBase) {
			return fmt.Errorf("%w: symlink outside base directory", ErrPathTraversal)
		}
	}

	return nil
}

// GenerateSecureDocumentID generates a cryptographically secure document ID
func GenerateSecureDocumentID(filename string) (string, error) {
	// Generate 16 random bytes (128 bits)
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to hex string
	hexString := fmt.Sprintf("%x", randomBytes)

	// Return in format doc-[hex]
	return fmt.Sprintf("doc-%s", hexString), nil
}
