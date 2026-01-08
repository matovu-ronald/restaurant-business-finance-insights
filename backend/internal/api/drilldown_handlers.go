package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lakehouse/restaurant-finance/internal/auth"
)

// DrilldownHandler handles sales drill-down requests
type DrilldownHandler struct {
	db *pgxpool.Pool
}

// NewDrilldownHandler creates a new drilldown handler
func NewDrilldownHandler(db *pgxpool.Pool) *DrilldownHandler {
	return &DrilldownHandler{db: db}
}

// SaleRow represents a single sale for drill-down view
type SaleRow struct {
	ID            string  `json:"id"`
	OccurredAt    string  `json:"occurred_at"`
	Channel       string  `json:"channel"`
	Daypart       string  `json:"daypart"`
	Total         float64 `json:"total"`
	Subtotal      float64 `json:"subtotal"`
	Tax           float64 `json:"tax"`
	Discounts     float64 `json:"discounts"`
	Comps         float64 `json:"comps"`
	PaymentMethod string  `json:"payment_method"`
}

// DrilldownResponse represents the paginated drill-down response
type DrilldownResponse struct {
	Data       []SaleRow `json:"data"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
}

// HandleSales handles GET /kpi/drilldown/sales requests
func (h *DrilldownHandler) HandleSales(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get location ID from claims if authenticated, otherwise use default
	var locationID string
	claims := auth.GetUserClaims(ctx)
	if claims != nil {
		locationID = claims.LocationID.String()
	} else {
		// Use default location for public access - query from database
		row := h.db.QueryRow(ctx, "SELECT id FROM locations LIMIT 1")
		if err := row.Scan(&locationID); err != nil {
			http.Error(w, "No location configured", http.StatusInternalServerError)
			return
		}
	}

	// Parse query parameters
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	channel := r.URL.Query().Get("channel")
	daypart := r.URL.Query().Get("daypart")
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	// Parse pagination
	page := 1
	pageSize := 50
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Parse dates (default to last 30 days)
	loc, _ := time.LoadLocation("Australia/Brisbane")
	endDate := time.Now().In(loc)
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = t
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = t
		}
	}

	// Build query
	baseQuery := `
		FROM sales s
		LEFT JOIN service_channels sc ON s.channel_id = sc.id
		LEFT JOIN dayparts d ON s.daypart_id = d.id
		WHERE s.location_id = $1
		AND s.occurred_at >= $2
		AND s.occurred_at <= $3
	`
	args := []interface{}{locationID, startDate, endDate.Add(24 * time.Hour)}
	argIdx := 4

	if channel != "" {
		baseQuery += ` AND sc.code = $` + strconv.Itoa(argIdx)
		args = append(args, channel)
		argIdx++
	}
	if daypart != "" {
		baseQuery += ` AND d.code = $` + strconv.Itoa(argIdx)
		args = append(args, daypart)
		argIdx++
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) ` + baseQuery
	if err := h.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		http.Error(w, "Failed to count sales", http.StatusInternalServerError)
		return
	}

	// Get paginated data
	dataQuery := `
		SELECT s.id, s.occurred_at, COALESCE(sc.display_name, '-'), COALESCE(d.display_name, '-'),
			s.total, s.subtotal, s.tax, s.discounts, s.comps, COALESCE(s.payment_method, '-')
	` + baseQuery + `
		ORDER BY s.occurred_at DESC
		LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)

	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := h.db.Query(ctx, dataQuery, args...)
	if err != nil {
		http.Error(w, "Failed to fetch sales", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var data []SaleRow
	for rows.Next() {
		var row SaleRow
		var occurredAt time.Time
		err := rows.Scan(
			&row.ID, &occurredAt, &row.Channel, &row.Daypart,
			&row.Total, &row.Subtotal, &row.Tax, &row.Discounts, &row.Comps, &row.PaymentMethod,
		)
		if err != nil {
			continue
		}
		row.OccurredAt = occurredAt.In(loc).Format("2006-01-02 15:04")
		data = append(data, row)
	}

	totalPages := (total + pageSize - 1) / pageSize
	response := DrilldownResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
