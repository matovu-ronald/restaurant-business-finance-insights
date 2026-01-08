package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileStorage handles file operations for imports/exports
type FileStorage struct {
	basePath string
}

// NewFileStorage creates a new file storage instance
func NewFileStorage(basePath string) (*FileStorage, error) {
	// Create base directories
	dirs := []string{
		filepath.Join(basePath, "uploads"),
		filepath.Join(basePath, "exports"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return &FileStorage{basePath: basePath}, nil
}

// SaveUpload saves an uploaded file and returns its hash and path
func (fs *FileStorage) SaveUpload(filename string, reader io.Reader) (hash string, path string, err error) {
	// Create temp file to calculate hash while writing
	tempPath := filepath.Join(fs.basePath, "uploads", filename+".tmp")
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Calculate hash while copying
	hasher := sha256.New()
	tee := io.TeeReader(reader, hasher)

	if _, err := io.Copy(tempFile, tee); err != nil {
		os.Remove(tempPath)
		return "", "", fmt.Errorf("failed to write file: %w", err)
	}

	hash = hex.EncodeToString(hasher.Sum(nil))
	finalPath := filepath.Join(fs.basePath, "uploads", hash+"_"+filename)

	// Rename temp file to final path
	if err := os.Rename(tempPath, finalPath); err != nil {
		os.Remove(tempPath)
		return "", "", fmt.Errorf("failed to rename file: %w", err)
	}

	return hash, finalPath, nil
}

// GetUploadPath returns the full path for an upload
func (fs *FileStorage) GetUploadPath(hash, filename string) string {
	return filepath.Join(fs.basePath, "uploads", hash+"_"+filename)
}

// OpenUpload opens an uploaded file for reading
func (fs *FileStorage) OpenUpload(hash, filename string) (*os.File, error) {
	path := fs.GetUploadPath(hash, filename)
	return os.Open(path)
}

// SaveExport saves an export file and returns its path
func (fs *FileStorage) SaveExport(filename string, data []byte) (string, error) {
	path := filepath.Join(fs.basePath, "exports", filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write export: %w", err)
	}
	return path, nil
}

// GetExportPath returns the full path for an export
func (fs *FileStorage) GetExportPath(filename string) string {
	return filepath.Join(fs.basePath, "exports", filename)
}

// OpenExport opens an export file for reading
func (fs *FileStorage) OpenExport(filename string) (*os.File, error) {
	path := fs.GetExportPath(filename)
	return os.Open(path)
}

// DeleteFile removes a file
func (fs *FileStorage) DeleteFile(path string) error {
	return os.Remove(path)
}

// FileExists checks if a file exists
func (fs *FileStorage) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
