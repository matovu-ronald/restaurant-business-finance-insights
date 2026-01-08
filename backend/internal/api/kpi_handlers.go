package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lakehouse/restaurant-finance/internal/kpi"
)

// KPIHandler handles KPI-related HTTP requests
type KPIHandler struct {
	service *kpi.Service
}

// NewKPIHandler creates a new KPI handler
func NewKPIHandler(service *kpi.Service) *KPIHandler {
	return &KPIHandler{service: service}
}

// HandleDaily handles GET /kpi/daily requests
func (h *KPIHandler) HandleDaily(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	dateStr := r.URL.Query().Get("date")
	rangeStr := r.URL.Query().Get("range")

	// Default to today if no date specified
	var referenceDate time.Time
	if dateStr != "" {
		var err error
		referenceDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "Invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	} else {
		// Use Brisbane timezone for "today"
		loc, _ := time.LoadLocation("Australia/Brisbane")
		referenceDate = time.Now().In(loc)
	}

	// Default range
	if rangeStr == "" {
		rangeStr = "30d"
	}

	// Parse date range
	startDate, endDate := kpi.ParseDateRange(rangeStr, referenceDate)

	// Get KPI data
	response, err := h.service.GetDailyKPIs(ctx, startDate, endDate, rangeStr)
	if err != nil {
		http.Error(w, "Failed to fetch KPI data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
