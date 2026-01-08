package config

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

// FileUploadConfig holds upload safety settings
type FileUploadConfig struct {
	MaxFileSize       int64    // Maximum file size in bytes
	AllowedExtensions []string // Allowed file extensions (lowercase, with dot)
	AllowedMIMETypes  []string // Allowed MIME types
}

// DefaultFileUploadConfig returns safe defaults for CSV imports
func DefaultFileUploadConfig() FileUploadConfig {
	return FileUploadConfig{
		MaxFileSize:       10 * 1024 * 1024, // 10 MB
		AllowedExtensions: []string{".csv", ".txt"},
		AllowedMIMETypes:  []string{"text/csv", "text/plain", "application/csv", "application/octet-stream"},
	}
}

// ValidateConfig performs comprehensive validation of all config values
func ValidateConfig(cfg *Config) error {
	var errs []error

	// Database validation
	if cfg.Database.URL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	} else if !strings.HasPrefix(cfg.Database.URL, "postgres://") && !strings.HasPrefix(cfg.Database.URL, "postgresql://") {
		errs = append(errs, errors.New("DATABASE_URL must be a valid PostgreSQL connection string"))
	}

	// Server validation
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		errs = append(errs, fmt.Errorf("SERVER_PORT must be between 1 and 65535, got %d", cfg.Server.Port))
	}

	// JWT validation
	if cfg.JWT.Secret == "" {
		errs = append(errs, errors.New("JWT_SECRET is required"))
	} else if len(cfg.JWT.Secret) < 32 {
		errs = append(errs, errors.New("JWT_SECRET must be at least 32 characters for security"))
	}
	if cfg.JWT.ExpireHours < 1 {
		errs = append(errs, errors.New("JWT_EXPIRE_HOURS must be at least 1"))
	}

	// Storage path validation
	if cfg.StoragePath == "" {
		errs = append(errs, errors.New("STORAGE_PATH is required"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// ValidateFileUpload checks file against upload safety rules
func ValidateFileUpload(header *multipart.FileHeader, cfg FileUploadConfig) error {
	// Check file size
	if header.Size > cfg.MaxFileSize {
		return fmt.Errorf("file size %d bytes exceeds maximum %d bytes", header.Size, cfg.MaxFileSize)
	}
	if header.Size == 0 {
		return errors.New("file is empty")
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !containsString(cfg.AllowedExtensions, ext) {
		return fmt.Errorf("file extension %q not allowed; must be one of: %v", ext, cfg.AllowedExtensions)
	}

	// Sanitize filename - reject path traversal attempts
	if strings.Contains(header.Filename, "..") || strings.Contains(header.Filename, "/") || strings.Contains(header.Filename, "\\") {
		return errors.New("invalid filename: path traversal characters not allowed")
	}

	return nil
}

// ValidateFileContent performs additional content-based validation
func ValidateFileContent(file multipart.File, cfg FileUploadConfig) error {
	// Read first 512 bytes for MIME detection
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Seek back to start so file can be read again
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	// Detect MIME type
	mimeType := http.DetectContentType(buffer[:n])
	if !containsString(cfg.AllowedMIMETypes, mimeType) {
		// Allow if it's text/plain (common for CSV)
		if !strings.HasPrefix(mimeType, "text/") {
			return fmt.Errorf("file type %q not allowed; must be one of: %v", mimeType, cfg.AllowedMIMETypes)
		}
	}

	return nil
}

// SanitizeFilename removes potentially dangerous characters from filename
func SanitizeFilename(filename string) string {
	// Get just the base name, removing any path components
	base := filepath.Base(filename)

	// Replace any non-alphanumeric, non-dot, non-hyphen, non-underscore characters
	var sanitized strings.Builder
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			sanitized.WriteRune(r)
		}
	}

	result := sanitized.String()
	if result == "" || result == "." || result == ".." {
		return "upload"
	}

	return result
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
