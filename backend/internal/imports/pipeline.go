package imports

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ImportJob represents an import job with its status and results
type ImportJob struct {
	ID             uuid.UUID  `json:"id"`
	SourceType     string     `json:"source_type"`
	Status         string     `json:"status"` // pending, processing, completed, failed
	FileName       string     `json:"file_name"`
	FileHash       string     `json:"file_hash"`
	TotalRows      int        `json:"total_rows"`
	ProcessedRows  int        `json:"processed_rows"`
	ErrorRows      int        `json:"error_rows"`
	LocationID     uuid.UUID  `json:"location_id"`
	MappingID      *uuid.UUID `json:"mapping_id,omitempty"`
	CreatedByID    uuid.UUID  `json:"created_by_id"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	ErrorMessage   string     `json:"error_message,omitempty"`
}

// ImportAnomaly represents an anomaly or issue detected during import
type ImportAnomaly struct {
	ID          uuid.UUID `json:"id"`
	ImportJobID uuid.UUID `json:"import_job_id"`
	LineNumber  int       `json:"line_number"`
	Severity    string    `json:"severity"` // error, warning
	Message     string    `json:"message"`
	RawData     string    `json:"raw_data,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Pipeline handles the import process
type Pipeline struct {
	db           *pgxpool.Pool
	store        *ImportStore
	mappingStore *MappingStore
}

// NewPipeline creates a new import pipeline
func NewPipeline(db *pgxpool.Pool) *Pipeline {
	return &Pipeline{
		db:           db,
		store:        NewImportStore(db),
		mappingStore: NewMappingStore(db),
	}
}

// StartImport creates a new import job and begins processing
func (p *Pipeline) StartImport(ctx context.Context, params ImportParams) (*ImportJob, error) {
	// Calculate file hash
	hash := sha256.New()
	if _, err := io.Copy(hash, params.File); err != nil {
		return nil, fmt.Errorf("failed to hash file: %w", err)
	}
	fileHash := hex.EncodeToString(hash.Sum(nil))

	// Check for duplicate import (idempotency)
	existingJob, err := p.store.GetByFileHash(ctx, fileHash, params.LocationID)
	if err == nil && existingJob != nil {
		if existingJob.Status == "completed" {
			return existingJob, fmt.Errorf("file has already been imported (job ID: %s)", existingJob.ID)
		}
	}

	// Create import job
	job := &ImportJob{
		ID:          uuid.New(),
		SourceType:  params.SourceType,
		Status:      "pending",
		FileName:    params.FileName,
		FileHash:    fileHash,
		LocationID:  params.LocationID,
		MappingID:   params.MappingID,
		CreatedByID: params.UserID,
		CreatedAt:   time.Now(),
	}

	if err := p.store.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create import job: %w", err)
	}

	return job, nil
}

// ProcessImport processes an import job
func (p *Pipeline) ProcessImport(ctx context.Context, jobID uuid.UUID, fileReader io.Reader) error {
	// Update job status to processing
	if err := p.store.UpdateJobStatus(ctx, jobID, "processing", ""); err != nil {
		return err
	}

	// Get job details
	job, err := p.store.GetJobByID(ctx, jobID)
	if err != nil {
		return err
	}

	// Get mapping if specified
	var mapping *MappingProfile
	if job.MappingID != nil {
		mapping, err = p.mappingStore.GetByID(ctx, *job.MappingID)
		if err != nil {
			p.store.UpdateJobStatus(ctx, jobID, "failed", fmt.Sprintf("failed to load mapping: %v", err))
			return err
		}
	}

	// Parse the file
	parser := NewParser(job.SourceType, mapping)
	result, err := parser.Parse(fileReader)
	if err != nil {
		p.store.UpdateJobStatus(ctx, jobID, "failed", fmt.Sprintf("failed to parse file: %v", err))
		return err
	}

	// Update job with row counts
	job.TotalRows = result.TotalRows
	job.ErrorRows = result.ErrorRows

	// Process rows
	var processedRows int
	for _, row := range result.Rows {
		if len(row.Errors) > 0 {
			// Record anomalies for error rows
			for _, errMsg := range row.Errors {
				anomaly := &ImportAnomaly{
					ID:          uuid.New(),
					ImportJobID: jobID,
					LineNumber:  row.LineNumber,
					Severity:    "error",
					Message:     errMsg,
					CreatedAt:   time.Now(),
				}
				p.store.CreateAnomaly(ctx, anomaly)
			}
			continue
		}

		// Process valid row based on source type
		var processErr error
		switch job.SourceType {
		case "pos":
			processErr = p.processPOSRow(ctx, job, row)
		case "payroll":
			processErr = p.processPayrollRow(ctx, job, row)
		case "inventory":
			processErr = p.processInventoryRow(ctx, job, row)
		}

		if processErr != nil {
			anomaly := &ImportAnomaly{
				ID:          uuid.New(),
				ImportJobID: jobID,
				LineNumber:  row.LineNumber,
				Severity:    "error",
				Message:     processErr.Error(),
				CreatedAt:   time.Now(),
			}
			p.store.CreateAnomaly(ctx, anomaly)
			job.ErrorRows++
		} else {
			processedRows++
		}
	}

	// Update job as completed
	job.ProcessedRows = processedRows
	now := time.Now()
	job.CompletedAt = &now
	job.Status = "completed"

	if err := p.store.UpdateJob(ctx, job); err != nil {
		return err
	}

	return nil
}

func (p *Pipeline) processPOSRow(ctx context.Context, job *ImportJob, row ParsedRow) error {
	dateStr, _ := row.Mapped["date"].(string)
	date, err := parseDate(dateStr)
	if err != nil {
		return fmt.Errorf("invalid date: %w", err)
	}

	totalStr, _ := row.Mapped["total"].(string)
	total, err := parseAmount(totalStr)
	if err != nil {
		return fmt.Errorf("invalid total: %w", err)
	}

	// Parse optional fields
	var discounts, comps, tax float64
	if v, ok := row.Mapped["discounts"].(string); ok && v != "" {
		discounts, _ = parseAmount(v)
	}
	if v, ok := row.Mapped["comps"].(string); ok && v != "" {
		comps, _ = parseAmount(v)
	}
	if v, ok := row.Mapped["tax"].(string); ok && v != "" {
		tax, _ = parseAmount(v)
	}

	// Get or create channel
	var channelID *uuid.UUID
	if channel, ok := row.Mapped["channel"].(string); ok && channel != "" {
		id, err := p.getOrCreateChannel(ctx, channel, job.LocationID)
		if err == nil {
			channelID = &id
		}
	}

	// Get daypart based on time
	var daypartID *uuid.UUID
	if timeStr, ok := row.Mapped["time"].(string); ok && timeStr != "" {
		id, err := p.getDaypartForTime(ctx, timeStr)
		if err == nil {
			daypartID = &id
		}
	}

	// Upsert sale using date + location + row number as key for idempotency
	query := `
		INSERT INTO sales (id, location_id, channel_id, daypart_id, occurred_at, total, subtotal, tax, discounts, comps, payment_method, import_source, source_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		ON CONFLICT (location_id, import_source, source_id) DO UPDATE SET
			total = EXCLUDED.total,
			subtotal = EXCLUDED.subtotal,
			tax = EXCLUDED.tax,
			discounts = EXCLUDED.discounts,
			comps = EXCLUDED.comps,
			updated_at = NOW()
	`

	subtotal := total - tax
	paymentMethod, _ := row.Mapped["payment_method"].(string)
	sourceID := fmt.Sprintf("%s-%d", job.FileHash[:8], row.LineNumber)

	_, err = p.db.Exec(ctx, query,
		uuid.New(),
		job.LocationID,
		channelID,
		daypartID,
		date,
		total,
		subtotal,
		tax,
		discounts,
		comps,
		paymentMethod,
		"csv-import",
		sourceID,
	)

	return err
}

func (p *Pipeline) processPayrollRow(ctx context.Context, job *ImportJob, row ParsedRow) error {
	startStr, _ := row.Mapped["period_start"].(string)
	startDate, err := parseDate(startStr)
	if err != nil {
		return fmt.Errorf("invalid period_start: %w", err)
	}

	endStr, _ := row.Mapped["period_end"].(string)
	endDate, err := parseDate(endStr)
	if err != nil {
		return fmt.Errorf("invalid period_end: %w", err)
	}

	wagesStr, _ := row.Mapped["total_wages"].(string)
	wages, err := parseAmount(wagesStr)
	if err != nil {
		return fmt.Errorf("invalid total_wages: %w", err)
	}

	// Parse optional fields
	var super, taxWithheld float64
	if v, ok := row.Mapped["superannuation"].(string); ok && v != "" {
		super, _ = parseAmount(v)
	}
	if v, ok := row.Mapped["tax_withheld"].(string); ok && v != "" {
		taxWithheld, _ = parseAmount(v)
	}

	// Upsert payroll period
	query := `
		INSERT INTO payroll_periods (id, location_id, start_date, end_date, labor_cost, superannuation, tax_withheld, import_source, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		ON CONFLICT (location_id, start_date, end_date) DO UPDATE SET
			labor_cost = EXCLUDED.labor_cost,
			superannuation = EXCLUDED.superannuation,
			tax_withheld = EXCLUDED.tax_withheld,
			updated_at = NOW()
	`

	_, err = p.db.Exec(ctx, query,
		uuid.New(),
		job.LocationID,
		startDate,
		endDate,
		wages,
		super,
		taxWithheld,
		"csv-import",
	)

	return err
}

func (p *Pipeline) processInventoryRow(ctx context.Context, job *ImportJob, row ParsedRow) error {
	dateStr, _ := row.Mapped["snapshot_date"].(string)
	date, err := parseDate(dateStr)
	if err != nil {
		return fmt.Errorf("invalid snapshot_date: %w", err)
	}

	itemName, _ := row.Mapped["item_name"].(string)
	if itemName == "" {
		return fmt.Errorf("item_name is required")
	}

	qtyStr, _ := row.Mapped["quantity"].(string)
	qty, err := parseAmount(qtyStr)
	if err != nil {
		return fmt.Errorf("invalid quantity: %w", err)
	}

	costStr, _ := row.Mapped["unit_cost"].(string)
	cost, err := parseAmount(costStr)
	if err != nil {
		return fmt.Errorf("invalid unit_cost: %w", err)
	}

	unit, _ := row.Mapped["unit"].(string)
	if unit == "" {
		unit = "ea"
	}

	totalValue := qty * cost

	// Upsert inventory snapshot
	query := `
		INSERT INTO inventory_snapshots (id, location_id, snapshot_date, item_name, category, quantity, unit, unit_cost, total_value, import_source, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		ON CONFLICT (location_id, snapshot_date, item_name) DO UPDATE SET
			quantity = EXCLUDED.quantity,
			unit_cost = EXCLUDED.unit_cost,
			total_value = EXCLUDED.total_value,
			updated_at = NOW()
	`

	category, _ := row.Mapped["category"].(string)

	_, err = p.db.Exec(ctx, query,
		uuid.New(),
		job.LocationID,
		date,
		itemName,
		category,
		qty,
		unit,
		cost,
		totalValue,
		"csv-import",
	)

	return err
}

func (p *Pipeline) getOrCreateChannel(ctx context.Context, name string, locationID uuid.UUID) (uuid.UUID, error) {
	// Try to find existing channel
	var id uuid.UUID
	query := `SELECT id FROM service_channels WHERE LOWER(display_name) = LOWER($1) AND location_id = $2`
	err := p.db.QueryRow(ctx, query, name, locationID).Scan(&id)
	if err == nil {
		return id, nil
	}

	// Create new channel
	id = uuid.New()
	code := slugify(name)
	insertQuery := `INSERT INTO service_channels (id, code, display_name, location_id, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW())`
	_, err = p.db.Exec(ctx, insertQuery, id, code, name, locationID)
	return id, err
}

func (p *Pipeline) getDaypartForTime(ctx context.Context, timeStr string) (uuid.UUID, error) {
	// Parse time
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		t, err = time.Parse("3:04 PM", timeStr)
		if err != nil {
			return uuid.Nil, err
		}
	}

	// Find daypart that contains this time
	query := `SELECT id FROM dayparts WHERE start_time <= $1 AND end_time > $1 LIMIT 1`
	var id uuid.UUID
	err = p.db.QueryRow(ctx, query, t.Format("15:04:05")).Scan(&id)
	return id, err
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

// ImportParams contains parameters for starting an import
type ImportParams struct {
	SourceType string
	FileName   string
	File       io.Reader
	LocationID uuid.UUID
	MappingID  *uuid.UUID
	UserID     uuid.UUID
}
