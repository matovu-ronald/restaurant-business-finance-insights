package imports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MappingProfile represents a saved column-to-field mapping configuration
type MappingProfile struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	SourceType  string                 `json:"source_type"` // pos, payroll, inventory
	ColumnMaps  map[string]string      `json:"column_maps"` // source column -> target field
	Defaults    map[string]interface{} `json:"defaults"`    // default values for missing columns
	LocationID  uuid.UUID              `json:"location_id"`
	CreatedByID uuid.UUID              `json:"created_by_id"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// MappingStore handles mapping profile persistence
type MappingStore struct {
	db *pgxpool.Pool
}

// NewMappingStore creates a new mapping store
func NewMappingStore(db *pgxpool.Pool) *MappingStore {
	return &MappingStore{db: db}
}

// Create creates a new mapping profile
func (s *MappingStore) Create(ctx context.Context, profile *MappingProfile) error {
	query := `
		INSERT INTO mapping_profiles (id, name, source_type, column_maps, defaults, location_id, created_by_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	profile.ID = uuid.New()
	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()

	_, err := s.db.Exec(ctx, query,
		profile.ID,
		profile.Name,
		profile.SourceType,
		profile.ColumnMaps,
		profile.Defaults,
		profile.LocationID,
		profile.CreatedByID,
		profile.CreatedAt,
		profile.UpdatedAt,
	)
	return err
}

// GetByID retrieves a mapping profile by ID
func (s *MappingStore) GetByID(ctx context.Context, id uuid.UUID) (*MappingProfile, error) {
	query := `
		SELECT id, name, source_type, column_maps, defaults, location_id, created_by_id, created_at, updated_at
		FROM mapping_profiles
		WHERE id = $1
	`

	var profile MappingProfile
	err := s.db.QueryRow(ctx, query, id).Scan(
		&profile.ID,
		&profile.Name,
		&profile.SourceType,
		&profile.ColumnMaps,
		&profile.Defaults,
		&profile.LocationID,
		&profile.CreatedByID,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// GetBySourceType retrieves all mapping profiles for a source type
func (s *MappingStore) GetBySourceType(ctx context.Context, sourceType string, locationID uuid.UUID) ([]MappingProfile, error) {
	query := `
		SELECT id, name, source_type, column_maps, defaults, location_id, created_by_id, created_at, updated_at
		FROM mapping_profiles
		WHERE source_type = $1 AND location_id = $2
		ORDER BY name
	`

	rows, err := s.db.Query(ctx, query, sourceType, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []MappingProfile
	for rows.Next() {
		var profile MappingProfile
		err := rows.Scan(
			&profile.ID,
			&profile.Name,
			&profile.SourceType,
			&profile.ColumnMaps,
			&profile.Defaults,
			&profile.LocationID,
			&profile.CreatedByID,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, rows.Err()
}

// GetAll retrieves all mapping profiles for a location
func (s *MappingStore) GetAll(ctx context.Context, locationID uuid.UUID) ([]MappingProfile, error) {
	query := `
		SELECT id, name, source_type, column_maps, defaults, location_id, created_by_id, created_at, updated_at
		FROM mapping_profiles
		WHERE location_id = $1
		ORDER BY source_type, name
	`

	rows, err := s.db.Query(ctx, query, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []MappingProfile
	for rows.Next() {
		var profile MappingProfile
		err := rows.Scan(
			&profile.ID,
			&profile.Name,
			&profile.SourceType,
			&profile.ColumnMaps,
			&profile.Defaults,
			&profile.LocationID,
			&profile.CreatedByID,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, rows.Err()
}

// Delete deletes a mapping profile
func (s *MappingStore) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM mapping_profiles WHERE id = $1`
	_, err := s.db.Exec(ctx, query, id)
	return err
}
