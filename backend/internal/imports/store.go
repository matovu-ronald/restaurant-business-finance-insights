package imports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ImportStore handles import job and anomaly persistence
type ImportStore struct {
	db *pgxpool.Pool
}

// NewImportStore creates a new import store
func NewImportStore(db *pgxpool.Pool) *ImportStore {
	return &ImportStore{db: db}
}

// CreateJob creates a new import job
func (s *ImportStore) CreateJob(ctx context.Context, job *ImportJob) error {
	query := `
		INSERT INTO import_jobs (id, source_type, status, file_name, file_hash, total_rows, processed_rows, error_rows, location_id, mapping_id, created_by_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := s.db.Exec(ctx, query,
		job.ID,
		job.SourceType,
		job.Status,
		job.FileName,
		job.FileHash,
		job.TotalRows,
		job.ProcessedRows,
		job.ErrorRows,
		job.LocationID,
		job.MappingID,
		job.CreatedByID,
		job.CreatedAt,
	)
	return err
}

// GetJobByID retrieves an import job by ID
func (s *ImportStore) GetJobByID(ctx context.Context, id uuid.UUID) (*ImportJob, error) {
	query := `
		SELECT id, source_type, status, file_name, file_hash, total_rows, processed_rows, error_rows, location_id, mapping_id, created_by_id, created_at, completed_at, error_message
		FROM import_jobs
		WHERE id = $1
	`

	var job ImportJob
	err := s.db.QueryRow(ctx, query, id).Scan(
		&job.ID,
		&job.SourceType,
		&job.Status,
		&job.FileName,
		&job.FileHash,
		&job.TotalRows,
		&job.ProcessedRows,
		&job.ErrorRows,
		&job.LocationID,
		&job.MappingID,
		&job.CreatedByID,
		&job.CreatedAt,
		&job.CompletedAt,
		&job.ErrorMessage,
	)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// GetByFileHash retrieves an import job by file hash
func (s *ImportStore) GetByFileHash(ctx context.Context, fileHash string, locationID uuid.UUID) (*ImportJob, error) {
	query := `
		SELECT id, source_type, status, file_name, file_hash, total_rows, processed_rows, error_rows, location_id, mapping_id, created_by_id, created_at, completed_at, error_message
		FROM import_jobs
		WHERE file_hash = $1 AND location_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var job ImportJob
	err := s.db.QueryRow(ctx, query, fileHash, locationID).Scan(
		&job.ID,
		&job.SourceType,
		&job.Status,
		&job.FileName,
		&job.FileHash,
		&job.TotalRows,
		&job.ProcessedRows,
		&job.ErrorRows,
		&job.LocationID,
		&job.MappingID,
		&job.CreatedByID,
		&job.CreatedAt,
		&job.CompletedAt,
		&job.ErrorMessage,
	)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// UpdateJobStatus updates the status of an import job
func (s *ImportStore) UpdateJobStatus(ctx context.Context, id uuid.UUID, status string, errorMsg string) error {
	var query string
	var args []interface{}

	if status == "completed" || status == "failed" {
		now := time.Now()
		query = `UPDATE import_jobs SET status = $1, error_message = $2, completed_at = $3 WHERE id = $4`
		args = []interface{}{status, errorMsg, now, id}
	} else {
		query = `UPDATE import_jobs SET status = $1, error_message = $2 WHERE id = $3`
		args = []interface{}{status, errorMsg, id}
	}

	_, err := s.db.Exec(ctx, query, args...)
	return err
}

// UpdateJob updates an import job
func (s *ImportStore) UpdateJob(ctx context.Context, job *ImportJob) error {
	query := `
		UPDATE import_jobs
		SET status = $1, total_rows = $2, processed_rows = $3, error_rows = $4, completed_at = $5, error_message = $6
		WHERE id = $7
	`
	_, err := s.db.Exec(ctx, query,
		job.Status,
		job.TotalRows,
		job.ProcessedRows,
		job.ErrorRows,
		job.CompletedAt,
		job.ErrorMessage,
		job.ID,
	)
	return err
}

// ListJobs retrieves import jobs for a location
func (s *ImportStore) ListJobs(ctx context.Context, locationID uuid.UUID, limit int) ([]ImportJob, error) {
	query := `
		SELECT id, source_type, status, file_name, file_hash, total_rows, processed_rows, error_rows, location_id, mapping_id, created_by_id, created_at, completed_at, error_message
		FROM import_jobs
		WHERE location_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := s.db.Query(ctx, query, locationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []ImportJob
	for rows.Next() {
		var job ImportJob
		err := rows.Scan(
			&job.ID,
			&job.SourceType,
			&job.Status,
			&job.FileName,
			&job.FileHash,
			&job.TotalRows,
			&job.ProcessedRows,
			&job.ErrorRows,
			&job.LocationID,
			&job.MappingID,
			&job.CreatedByID,
			&job.CreatedAt,
			&job.CompletedAt,
			&job.ErrorMessage,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

// CreateAnomaly creates an import anomaly record
func (s *ImportStore) CreateAnomaly(ctx context.Context, anomaly *ImportAnomaly) error {
	query := `
		INSERT INTO import_anomalies (id, import_job_id, line_number, severity, message, raw_data, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.db.Exec(ctx, query,
		anomaly.ID,
		anomaly.ImportJobID,
		anomaly.LineNumber,
		anomaly.Severity,
		anomaly.Message,
		anomaly.RawData,
		anomaly.CreatedAt,
	)
	return err
}

// GetAnomaliesForJob retrieves anomalies for an import job
func (s *ImportStore) GetAnomaliesForJob(ctx context.Context, jobID uuid.UUID) ([]ImportAnomaly, error) {
	query := `
		SELECT id, import_job_id, line_number, severity, message, raw_data, created_at
		FROM import_anomalies
		WHERE import_job_id = $1
		ORDER BY line_number
	`

	rows, err := s.db.Query(ctx, query, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var anomalies []ImportAnomaly
	for rows.Next() {
		var a ImportAnomaly
		err := rows.Scan(
			&a.ID,
			&a.ImportJobID,
			&a.LineNumber,
			&a.Severity,
			&a.Message,
			&a.RawData,
			&a.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		anomalies = append(anomalies, a)
	}
	return anomalies, rows.Err()
}
