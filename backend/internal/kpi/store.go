package kpi

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// KPIAggregate represents aggregated KPI data
type KPIAggregate struct {
	ID                 uuid.UUID `json:"id"`
	Date               time.Time `json:"date"`
	LocationID         uuid.UUID `json:"location_id"`
	ChannelID          *uuid.UUID `json:"channel_id,omitempty"`
	DaypartID          *uuid.UUID `json:"daypart_id,omitempty"`
	Revenue            float64   `json:"revenue"`
	COGS               float64   `json:"cogs"`
	GrossMargin        float64   `json:"gross_margin"`
	LaborCost          float64   `json:"labor_cost"`
	LaborPct           float64   `json:"labor_pct"`
	Opex               float64   `json:"opex"`
	NetProfit          float64   `json:"net_profit"`
	Covers             int       `json:"covers"`
	AvgCheck           float64   `json:"avg_check"`
	Discounts          float64   `json:"discounts"`
	Comps              float64   `json:"comps"`
	FreshnessTimestamp time.Time `json:"freshness_timestamp"`
	ChannelCode        *string   `json:"channel_code,omitempty"`
	ChannelName        *string   `json:"channel_name,omitempty"`
	DaypartCode        *string   `json:"daypart_code,omitempty"`
	DaypartName        *string   `json:"daypart_name,omitempty"`
}

// KPITotals represents totaled KPI data for a period
type KPITotals struct {
	Revenue            float64   `json:"revenue"`
	COGS               float64   `json:"cogs"`
	GrossMargin        float64   `json:"gross_margin"`
	LaborCost          float64   `json:"labor_cost"`
	LaborPct           float64   `json:"labor_pct"`
	Opex               float64   `json:"opex"`
	NetProfit          float64   `json:"net_profit"`
	Covers             int       `json:"covers"`
	AvgCheck           float64   `json:"avg_check"`
	Discounts          float64   `json:"discounts"`
	Comps              float64   `json:"comps"`
	FreshnessTimestamp time.Time `json:"freshness_timestamp"`
}

// Store handles KPI data persistence
type Store struct {
	db *pgxpool.Pool
}

// NewStore creates a new KPI store
func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// GetAggregates retrieves KPI aggregates for a date range
func (s *Store) GetAggregates(ctx context.Context, startDate, endDate time.Time) ([]KPIAggregate, error) {
	query := `
		SELECT k.id, k.date, k.location_id, k.channel_id, k.daypart_id,
			k.revenue, k.cogs, k.gross_margin, k.labor_cost, k.labor_pct,
			k.opex, k.net_profit, k.covers, k.avg_check, k.discounts, k.comps,
			k.freshness_timestamp,
			sc.code as channel_code, sc.display_name as channel_name,
			d.code as daypart_code, d.display_name as daypart_name
		FROM kpi_aggregates k
		LEFT JOIN service_channels sc ON k.channel_id = sc.id
		LEFT JOIN dayparts d ON k.daypart_id = d.id
		WHERE k.date >= $1 AND k.date <= $2
		ORDER BY k.date DESC, sc.display_name, d.start_time
	`

	rows, err := s.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aggregates []KPIAggregate
	for rows.Next() {
		var agg KPIAggregate
		err := rows.Scan(
			&agg.ID, &agg.Date, &agg.LocationID, &agg.ChannelID, &agg.DaypartID,
			&agg.Revenue, &agg.COGS, &agg.GrossMargin, &agg.LaborCost, &agg.LaborPct,
			&agg.Opex, &agg.NetProfit, &agg.Covers, &agg.AvgCheck, &agg.Discounts, &agg.Comps,
			&agg.FreshnessTimestamp,
			&agg.ChannelCode, &agg.ChannelName, &agg.DaypartCode, &agg.DaypartName,
		)
		if err != nil {
			return nil, err
		}
		aggregates = append(aggregates, agg)
	}

	return aggregates, rows.Err()
}

// GetTotals retrieves totaled KPIs for a date range
func (s *Store) GetTotals(ctx context.Context, startDate, endDate time.Time) (*KPITotals, error) {
	query := `
		SELECT
			COALESCE(SUM(revenue), 0) as revenue,
			COALESCE(SUM(cogs), 0) as cogs,
			COALESCE(SUM(gross_margin), 0) as gross_margin,
			COALESCE(SUM(labor_cost), 0) as labor_cost,
			CASE WHEN SUM(revenue) > 0 THEN SUM(labor_cost) / SUM(revenue) * 100 ELSE 0 END as labor_pct,
			COALESCE(SUM(opex), 0) as opex,
			COALESCE(SUM(net_profit), 0) as net_profit,
			COALESCE(SUM(covers), 0) as covers,
			CASE WHEN SUM(covers) > 0 THEN SUM(revenue) / SUM(covers) ELSE 0 END as avg_check,
			COALESCE(SUM(discounts), 0) as discounts,
			COALESCE(SUM(comps), 0) as comps,
			COALESCE(MAX(freshness_timestamp), NOW()) as freshness_timestamp
		FROM kpi_aggregates
		WHERE date >= $1 AND date <= $2
	`

	var totals KPITotals
	err := s.db.QueryRow(ctx, query, startDate, endDate).Scan(
		&totals.Revenue, &totals.COGS, &totals.GrossMargin,
		&totals.LaborCost, &totals.LaborPct, &totals.Opex,
		&totals.NetProfit, &totals.Covers, &totals.AvgCheck,
		&totals.Discounts, &totals.Comps, &totals.FreshnessTimestamp,
	)
	if err != nil {
		return nil, err
	}

	return &totals, nil
}

// GetByChannel retrieves KPIs grouped by channel for a date range
func (s *Store) GetByChannel(ctx context.Context, startDate, endDate time.Time) ([]KPISummary, error) {
	query := `
		SELECT
			sc.code as label,
			sc.display_name as display_name,
			COALESCE(SUM(k.revenue), 0) as revenue,
			COALESCE(SUM(k.cogs), 0) as cogs,
			COALESCE(SUM(k.gross_margin), 0) as gross_margin,
			COALESCE(SUM(k.labor_cost), 0) as labor_cost,
			CASE WHEN SUM(k.revenue) > 0 THEN SUM(k.labor_cost) / SUM(k.revenue) * 100 ELSE 0 END as labor_pct,
			COALESCE(SUM(k.opex), 0) as opex,
			COALESCE(SUM(k.net_profit), 0) as net_profit,
			COALESCE(SUM(k.covers), 0) as covers,
			CASE WHEN SUM(k.covers) > 0 THEN SUM(k.revenue) / SUM(k.covers) ELSE 0 END as avg_check,
			COALESCE(SUM(k.discounts), 0) as discounts,
			COALESCE(SUM(k.comps), 0) as comps
		FROM kpi_aggregates k
		JOIN service_channels sc ON k.channel_id = sc.id
		WHERE k.date >= $1 AND k.date <= $2 AND k.channel_id IS NOT NULL
		GROUP BY sc.code, sc.display_name
		ORDER BY sc.display_name
	`

	rows, err := s.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSummaries(rows)
}

// GetByDaypart retrieves KPIs grouped by daypart for a date range
func (s *Store) GetByDaypart(ctx context.Context, startDate, endDate time.Time) ([]KPISummary, error) {
	query := `
		SELECT
			d.code as label,
			d.display_name as display_name,
			COALESCE(SUM(k.revenue), 0) as revenue,
			COALESCE(SUM(k.cogs), 0) as cogs,
			COALESCE(SUM(k.gross_margin), 0) as gross_margin,
			COALESCE(SUM(k.labor_cost), 0) as labor_cost,
			CASE WHEN SUM(k.revenue) > 0 THEN SUM(k.labor_cost) / SUM(k.revenue) * 100 ELSE 0 END as labor_pct,
			COALESCE(SUM(k.opex), 0) as opex,
			COALESCE(SUM(k.net_profit), 0) as net_profit,
			COALESCE(SUM(k.covers), 0) as covers,
			CASE WHEN SUM(k.covers) > 0 THEN SUM(k.revenue) / SUM(k.covers) ELSE 0 END as avg_check,
			COALESCE(SUM(k.discounts), 0) as discounts,
			COALESCE(SUM(k.comps), 0) as comps
		FROM kpi_aggregates k
		JOIN dayparts d ON k.daypart_id = d.id
		WHERE k.date >= $1 AND k.date <= $2 AND k.daypart_id IS NOT NULL
		GROUP BY d.code, d.display_name, d.start_time
		ORDER BY d.start_time
	`

	rows, err := s.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSummaries(rows)
}

// KPISummary represents KPI data with a label (for channel/daypart breakdowns)
type KPISummary struct {
	Label       string  `json:"label"`
	DisplayName string  `json:"display_name"`
	Revenue     float64 `json:"revenue"`
	COGS        float64 `json:"cogs"`
	GrossMargin float64 `json:"gross_margin"`
	LaborCost   float64 `json:"labor_cost"`
	LaborPct    float64 `json:"labor_pct"`
	Opex        float64 `json:"opex"`
	NetProfit   float64 `json:"net_profit"`
	Covers      int     `json:"covers"`
	AvgCheck    float64 `json:"avg_check"`
	Discounts   float64 `json:"discounts"`
	Comps       float64 `json:"comps"`
}

func scanSummaries(rows pgx.Rows) ([]KPISummary, error) {
	var summaries []KPISummary
	for rows.Next() {
		var s KPISummary
		err := rows.Scan(
			&s.Label, &s.DisplayName, &s.Revenue, &s.COGS, &s.GrossMargin,
			&s.LaborCost, &s.LaborPct, &s.Opex, &s.NetProfit,
			&s.Covers, &s.AvgCheck, &s.Discounts, &s.Comps,
		)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, rows.Err()
}
