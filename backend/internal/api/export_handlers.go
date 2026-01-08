package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/lakehouse/restaurant-finance/internal/auth"
	"github.com/lakehouse/restaurant-finance/internal/exports"
)

// ExportHandler handles export-related HTTP requests
type ExportHandler struct {
	service *exports.ExportService
	store   *exports.ExportStore
}

// NewExportHandler creates a new export handler
func NewExportHandler(service *exports.ExportService, store *exports.ExportStore) *ExportHandler {
	return &ExportHandler{
		service: service,
		store:   store,
	}
}

// CreateExportRequest represents the export creation request
type CreateExportRequest struct {
	ExportType string `json:"export_type"` // pnl, channel_summary
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
}

// HandlePnL handles POST /exports/pnl requests
func (h *ExportHandler) HandlePnL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get location ID from claims if authenticated, otherwise use default
	var locationID uuid.UUID
	var userID uuid.UUID
	claims := auth.GetUserClaims(ctx)
	if claims != nil {
		locationID = claims.LocationID
		userID = claims.UserID
	} else {
		// Use default location and admin user for public access
		locationID = uuid.MustParse("593bb8d0-36a8-4ce3-bf51-715532cee9ca")
		userID = uuid.MustParse("22222222-2222-2222-2222-222222222222") // admin@lakehouse.com
	}

	var req CreateExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse dates
	loc, _ := time.LoadLocation("Australia/Brisbane")
	endDate := time.Now().In(loc)
	startDate := endDate.AddDate(0, 0, -30)

	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			endDate = t
		}
	}

	params := exports.ExportPnLParams{
		StartDate:  startDate,
		EndDate:    endDate,
		LocationID: locationID,
		UserID:     userID,
	}

	var job *exports.ExportJob
	var data []byte
	var err error

	switch req.ExportType {
	case "channel_summary":
		job, data, err = h.service.GenerateChannelSummary(ctx, params)
	default:
		job, data, err = h.service.GeneratePnLExport(ctx, params)
	}

	if err != nil {
		log.Printf("Export generation error: %v", err)
		http.Error(w, "Failed to generate export: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return CSV directly
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename="+job.FileName)
	w.Write(data)
}

// HandleGet handles GET /exports/{id} requests
func (h *ExportHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid export ID", http.StatusBadRequest)
		return
	}

	job, err := h.store.GetJobByID(ctx, id)
	if err != nil {
		http.Error(w, "Export not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// HandleList handles GET /exports requests
func (h *ExportHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.GetUserClaims(ctx)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	jobs, err := h.store.ListJobs(ctx, claims.LocationID, 50)
	if err != nil {
		http.Error(w, "Failed to list exports", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}
