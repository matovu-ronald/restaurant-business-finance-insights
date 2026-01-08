package exports

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ExportJob represents an export job
type ExportJob struct {
	ID          uuid.UUID  `json:"id"`
	ExportType  string     `json:"export_type"` // pnl, channel_summary, daypart_summary
	PeriodStart time.Time  `json:"period_start"`
	PeriodEnd   time.Time  `json:"period_end"`
	Status      string     `json:"status"` // pending, processing, completed, failed
	FileName    string     `json:"file_name"`
	FilePath    string     `json:"file_path,omitempty"`
	RequestedBy uuid.UUID  `json:"requested_by"`
	RequestedAt time.Time  `json:"requested_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// ExportService handles export operations
type ExportService struct {
	db    *pgxpool.Pool
	store *ExportStore
}

// NewExportService creates a new export service
func NewExportService(db *pgxpool.Pool) *ExportService {
	return &ExportService{
		db:    db,
		store: NewExportStore(db),
	}
}

// ExportPnLParams contains parameters for P&L export
type ExportPnLParams struct {
	StartDate  time.Time
	EndDate    time.Time
	LocationID uuid.UUID
	UserID     uuid.UUID
}

// GeneratePnLExport creates a P&L CSV export
func (s *ExportService) GeneratePnLExport(ctx context.Context, params ExportPnLParams) (*ExportJob, []byte, error) {
	// Create export job
	job := &ExportJob{
		ID:          uuid.New(),
		ExportType:  "pnl",
		PeriodStart: params.StartDate,
		PeriodEnd:   params.EndDate,
		Status:      "processing",
		FileName:    fmt.Sprintf("pnl_%s_%s.csv", params.StartDate.Format("20060102"), params.EndDate.Format("20060102")),
		RequestedBy: params.UserID,
		RequestedAt: time.Now(),
	}

	if err := s.store.CreateJob(ctx, job); err != nil {
		return nil, nil, err
	}

	// Query KPI aggregates
	query := `
		SELECT
			DATE(k.date) as date,
			COALESCE(sc.display_name, 'Total') as channel,
			COALESCE(d.display_name, 'All Day') as daypart,
			SUM(k.revenue) as revenue,
			SUM(k.cogs) as cogs,
			SUM(k.gross_margin) as gross_margin,
			SUM(k.labor_cost) as labor_cost,
			CASE WHEN SUM(k.revenue) > 0 THEN SUM(k.labor_cost) / SUM(k.revenue) * 100 ELSE 0 END as labor_pct,
			SUM(k.opex) as opex,
			SUM(k.net_profit) as net_profit,
			SUM(k.covers) as covers,
			CASE WHEN SUM(k.covers) > 0 THEN SUM(k.revenue) / SUM(k.covers) ELSE 0 END as avg_check,
			SUM(k.discounts) as discounts,
			SUM(k.comps) as comps
		FROM kpi_aggregates k
		LEFT JOIN service_channels sc ON k.channel_id = sc.id
		LEFT JOIN dayparts d ON k.daypart_id = d.id
		WHERE k.location_id = $1
		AND k.date >= $2
		AND k.date <= $3
		GROUP BY DATE(k.date), sc.display_name, d.display_name, d.start_time
		ORDER BY DATE(k.date), sc.display_name, d.start_time
	`

	rows, err := s.db.Query(ctx, query, params.LocationID, params.StartDate, params.EndDate)
	if err != nil {
		s.store.UpdateJobStatus(ctx, job.ID, "failed", err.Error())
		return nil, nil, err
	}
	defer rows.Close()

	// Generate CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"Date", "Channel", "Daypart", "Revenue", "COGS", "Gross Margin", "Labor Cost", "Labor %", "OpEx", "Net Profit", "Covers", "Avg Check", "Discounts", "Comps"}
	writer.Write(header)

	// Write data rows
	for rows.Next() {
		var date time.Time
		var channel, daypart string
		var revenue, cogs, grossMargin, laborCost, laborPct, opex, netProfit, avgCheck, discounts, comps float64
		var covers int

		err := rows.Scan(&date, &channel, &daypart, &revenue, &cogs, &grossMargin, &laborCost, &laborPct, &opex, &netProfit, &covers, &avgCheck, &discounts, &comps)
		if err != nil {
			continue
		}

		row := []string{
			date.Format("2006-01-02"),
			channel,
			daypart,
			fmt.Sprintf("%.2f", revenue),
			fmt.Sprintf("%.2f", cogs),
			fmt.Sprintf("%.2f", grossMargin),
			fmt.Sprintf("%.2f", laborCost),
			fmt.Sprintf("%.1f%%", laborPct),
			fmt.Sprintf("%.2f", opex),
			fmt.Sprintf("%.2f", netProfit),
			fmt.Sprintf("%d", covers),
			fmt.Sprintf("%.2f", avgCheck),
			fmt.Sprintf("%.2f", discounts),
			fmt.Sprintf("%.2f", comps),
		}
		writer.Write(row)
	}

	writer.Flush()

	// Update job as completed
	now := time.Now()
	job.Status = "completed"
	job.CompletedAt = &now
	s.store.UpdateJob(ctx, job)

	return job, buf.Bytes(), nil
}

// GenerateChannelSummary creates a channel summary CSV export
func (s *ExportService) GenerateChannelSummary(ctx context.Context, params ExportPnLParams) (*ExportJob, []byte, error) {
	job := &ExportJob{
		ID:          uuid.New(),
		ExportType:  "channel_summary",
		PeriodStart: params.StartDate,
		PeriodEnd:   params.EndDate,
		Status:      "processing",
		FileName:    fmt.Sprintf("channel_summary_%s_%s.csv", params.StartDate.Format("20060102"), params.EndDate.Format("20060102")),
		RequestedBy: params.UserID,
		RequestedAt: time.Now(),
	}

	if err := s.store.CreateJob(ctx, job); err != nil {
		return nil, nil, err
	}

	query := `
		SELECT
			COALESCE(sc.display_name, 'Unknown') as channel,
			SUM(k.revenue) as revenue,
			SUM(k.cogs) as cogs,
			SUM(k.gross_margin) as gross_margin,
			SUM(k.covers) as covers,
			CASE WHEN SUM(k.covers) > 0 THEN SUM(k.revenue) / SUM(k.covers) ELSE 0 END as avg_check
		FROM kpi_aggregates k
		LEFT JOIN service_channels sc ON k.channel_id = sc.id
		WHERE k.location_id = $1
		AND k.date >= $2
		AND k.date <= $3
		AND k.channel_id IS NOT NULL
		GROUP BY sc.display_name
		ORDER BY SUM(k.revenue) DESC
	`

	rows, err := s.db.Query(ctx, query, params.LocationID, params.StartDate, params.EndDate)
	if err != nil {
		s.store.UpdateJobStatus(ctx, job.ID, "failed", err.Error())
		return nil, nil, err
	}
	defer rows.Close()

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	header := []string{"Channel", "Revenue", "COGS", "Gross Margin", "Covers", "Avg Check"}
	writer.Write(header)

	for rows.Next() {
		var channel string
		var revenue, cogs, grossMargin, avgCheck float64
		var covers int

		err := rows.Scan(&channel, &revenue, &cogs, &grossMargin, &covers, &avgCheck)
		if err != nil {
			continue
		}

		row := []string{
			channel,
			fmt.Sprintf("%.2f", revenue),
			fmt.Sprintf("%.2f", cogs),
			fmt.Sprintf("%.2f", grossMargin),
			fmt.Sprintf("%d", covers),
			fmt.Sprintf("%.2f", avgCheck),
		}
		writer.Write(row)
	}

	writer.Flush()

	now := time.Now()
	job.Status = "completed"
	job.CompletedAt = &now
	s.store.UpdateJob(ctx, job)

	return job, buf.Bytes(), nil
}

// ExportStore handles export job persistence
type ExportStore struct {
	db *pgxpool.Pool
}

// NewExportStore creates a new export store
func NewExportStore(db *pgxpool.Pool) *ExportStore {
	return &ExportStore{db: db}
}

// CreateJob creates a new export job
func (s *ExportStore) CreateJob(ctx context.Context, job *ExportJob) error {
	query := `
		INSERT INTO export_jobs (id, export_type, period_start, period_end, status, file_path, requested_by, requested_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := s.db.Exec(ctx, query,
		job.ID,
		job.ExportType,
		job.PeriodStart,
		job.PeriodEnd,
		job.Status,
		job.FilePath,
		job.RequestedBy,
		job.RequestedAt,
	)
	return err
}
// GetJobByID retrieves an export job by ID
func (s *ExportStore) GetJobByID(ctx context.Context, id uuid.UUID) (*ExportJob, error) {
	query := `
		SELECT id, export_type, period_start, period_end, status, file_path, requested_by, requested_at, completed_at
		FROM export_jobs
		WHERE id = $1
	`

	var job ExportJob
	err := s.db.QueryRow(ctx, query, id).Scan(
		&job.ID,
		&job.ExportType,
		&job.PeriodStart,
		&job.PeriodEnd,
		&job.Status,
		&job.FilePath,
		&job.RequestedBy,
		&job.RequestedAt,
		&job.CompletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// UpdateJob updates an export job
func (s *ExportStore) UpdateJob(ctx context.Context, job *ExportJob) error {
	query := `
		UPDATE export_jobs
		SET status = $1, file_path = $2, completed_at = $3
		WHERE id = $4
	`
	_, err := s.db.Exec(ctx, query, job.Status, job.FilePath, job.CompletedAt, job.ID)
	return err
}

// UpdateJobStatus updates just the status of an export job
func (s *ExportStore) UpdateJobStatus(ctx context.Context, id uuid.UUID, status, errorMsg string) error {
	query := `UPDATE export_jobs SET status = $1 WHERE id = $2`
	_, err := s.db.Exec(ctx, query, status, id)
	return err
}

// ListJobs retrieves export jobs (no location filter for now)
func (s *ExportStore) ListJobs(ctx context.Context, locationID uuid.UUID, limit int) ([]ExportJob, error) {
	query := `
		SELECT id, export_type, period_start, period_end, status, file_path, requested_by, requested_at, completed_at
		FROM export_jobs
		ORDER BY requested_at DESC
		LIMIT $1
	`

	rows, err := s.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []ExportJob
	for rows.Next() {
		var job ExportJob
		err := rows.Scan(
			&job.ID,
			&job.ExportType,
			&job.PeriodStart,
			&job.PeriodEnd,
			&job.Status,
			&job.FilePath,
			&job.RequestedBy,
			&job.RequestedAt,
			&job.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}
