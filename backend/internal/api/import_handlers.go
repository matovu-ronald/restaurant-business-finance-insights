package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/lakehouse/restaurant-finance/internal/auth"
	"github.com/lakehouse/restaurant-finance/internal/config"
	"github.com/lakehouse/restaurant-finance/internal/imports"
)

// ImportHandler handles import-related HTTP requests
type ImportHandler struct {
	pipeline     *imports.Pipeline
	importStore  *imports.ImportStore
	mappingStore *imports.MappingStore
	uploadCfg    config.FileUploadConfig
}

// NewImportHandler creates a new import handler
func NewImportHandler(pipeline *imports.Pipeline, importStore *imports.ImportStore, mappingStore *imports.MappingStore) *ImportHandler {
	return &ImportHandler{
		pipeline:     pipeline,
		importStore:  importStore,
		mappingStore: mappingStore,
		uploadCfg:    config.DefaultFileUploadConfig(),
	}
}

// CreateImportRequest represents the import creation request
type CreateImportRequest struct {
	SourceType string  `json:"source_type"`
	MappingID  *string `json:"mapping_id,omitempty"`
}

// HandleCreate handles POST /imports requests
func (h *ImportHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.GetUserClaims(ctx)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form (10 MB max)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file upload (size, extension, path traversal)
	if err := config.ValidateFileUpload(header, h.uploadCfg); err != nil {
		http.Error(w, "Invalid file: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate file content (MIME type check)
	if err := config.ValidateFileContent(file, h.uploadCfg); err != nil {
		http.Error(w, "Invalid file content: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Sanitize filename
	sanitizedFilename := config.SanitizeFilename(header.Filename)

	// Get source type
	sourceType := r.FormValue("source_type")
	if sourceType == "" {
		sourceType = "pos"
	}

	// Get mapping ID if provided
	var mappingID *uuid.UUID
	if mappingIDStr := r.FormValue("mapping_id"); mappingIDStr != "" {
		id, err := uuid.Parse(mappingIDStr)
		if err == nil {
			mappingID = &id
		}
	}

	// Read file content into buffer for hashing and reuse
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, file); err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Start import
	params := imports.ImportParams{
		SourceType: sourceType,
		FileName:   sanitizedFilename,
		File:       bytes.NewReader(buf.Bytes()),
		LocationID: claims.LocationID,
		MappingID:  mappingID,
		UserID:     claims.UserID,
	}

	job, err := h.pipeline.StartImport(ctx, params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// Process import in background (for now, synchronously)
	go func() {
		h.pipeline.ProcessImport(ctx, job.ID, bytes.NewReader(buf.Bytes()))
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(job)
}

// HandleGet handles GET /imports/{id} requests
func (h *ImportHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid import ID", http.StatusBadRequest)
		return
	}

	job, err := h.importStore.GetJobByID(ctx, id)
	if err != nil {
		http.Error(w, "Import not found", http.StatusNotFound)
		return
	}

	// Get anomalies
	anomalies, _ := h.importStore.GetAnomaliesForJob(ctx, id)

	response := map[string]interface{}{
		"job":       job,
		"anomalies": anomalies,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleList handles GET /imports requests
func (h *ImportHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.GetUserClaims(ctx)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	jobs, err := h.importStore.ListJobs(ctx, claims.LocationID, 50)
	if err != nil {
		http.Error(w, "Failed to list imports", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

// HandleMappingsGet handles GET /mappings requests
func (h *ImportHandler) HandleMappingsGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.GetUserClaims(ctx)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profiles, err := h.mappingStore.GetAll(ctx, claims.LocationID)
	if err != nil {
		http.Error(w, "Failed to list mappings", http.StatusInternalServerError)
		return
	}

	// Include default mappings
	response := map[string]interface{}{
		"profiles": profiles,
		"defaults": imports.DefaultMappings(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateMappingRequest represents a mapping profile creation request
type CreateMappingRequest struct {
	Name       string                 `json:"name"`
	SourceType string                 `json:"source_type"`
	ColumnMaps map[string]string      `json:"column_maps"`
	Defaults   map[string]interface{} `json:"defaults"`
}

// HandleMappingCreate handles POST /mappings requests
func (h *ImportHandler) HandleMappingCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.GetUserClaims(ctx)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateMappingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.SourceType == "" {
		http.Error(w, "Name and source_type are required", http.StatusBadRequest)
		return
	}

	profile := &imports.MappingProfile{
		Name:        req.Name,
		SourceType:  req.SourceType,
		ColumnMaps:  req.ColumnMaps,
		Defaults:    req.Defaults,
		LocationID:  claims.LocationID,
		CreatedByID: claims.UserID,
	}

	if err := h.mappingStore.Create(ctx, profile); err != nil {
		http.Error(w, "Failed to create mapping", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(profile)
}
