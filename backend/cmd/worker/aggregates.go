package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Worker refreshes KPI aggregates after imports
func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := RefreshAggregates(ctx, pool); err != nil {
		log.Fatalf("Failed to refresh aggregates: %v", err)
	}

	log.Println("Aggregates refreshed successfully")
}

// RefreshAggregates recalculates KPI aggregates from sales and payroll data
func RefreshAggregates(ctx context.Context, pool *pgxpool.Pool) error {
	// Get location ID
	var locationID uuid.UUID
	err := pool.QueryRow(ctx, "SELECT id FROM locations LIMIT 1").Scan(&locationID)
	if err != nil {
		log.Println("No location found, skipping aggregate refresh")
		return nil
	}

	// Get date range from sales data
	var minDate, maxDate time.Time
	err = pool.QueryRow(ctx, `
		SELECT COALESCE(MIN(DATE(occurred_at)), CURRENT_DATE), COALESCE(MAX(DATE(occurred_at)), CURRENT_DATE)
		FROM sales
	`).Scan(&minDate, &maxDate)
	if err != nil {
		return err
	}

	// Process each day
	for d := minDate; !d.After(maxDate); d = d.AddDate(0, 0, 1) {
		if err := refreshDayAggregates(ctx, pool, locationID, d); err != nil {
			log.Printf("Failed to refresh aggregates for %s: %v", d.Format("2006-01-02"), err)
		}
	}

	return nil
}

func refreshDayAggregates(ctx context.Context, pool *pgxpool.Pool, locationID uuid.UUID, date time.Time) error {
	// Calculate revenue and sales metrics by channel and daypart
	query := `
		INSERT INTO kpi_aggregates (date, location_id, channel_id, daypart_id, revenue, cogs, gross_margin, labor_cost, labor_pct, opex, net_profit, covers, avg_check, discounts, comps, freshness_timestamp)
		SELECT
			DATE(s.occurred_at) as date,
			s.location_id,
			s.channel_id,
			s.daypart_id,
			COALESCE(SUM(s.total), 0) as revenue,
			COALESCE(SUM(sl.quantity * COALESCE(mi.recipe_cost, 0)), 0) as cogs,
			COALESCE(SUM(s.total), 0) - COALESCE(SUM(sl.quantity * COALESCE(mi.recipe_cost, 0)), 0) as gross_margin,
			0 as labor_cost,
			0 as labor_pct,
			0 as opex,
			0 as net_profit,
			COUNT(DISTINCT s.id) as covers,
			CASE WHEN COUNT(DISTINCT s.id) > 0 THEN SUM(s.total) / COUNT(DISTINCT s.id) ELSE 0 END as avg_check,
			COALESCE(SUM(s.discounts), 0) as discounts,
			COALESCE(SUM(s.comps), 0) as comps,
			NOW() as freshness_timestamp
		FROM sales s
		LEFT JOIN sale_lines sl ON s.id = sl.sale_id
		LEFT JOIN menu_items mi ON sl.menu_item_id = mi.id
		WHERE DATE(s.occurred_at) = $1 AND s.location_id = $2
		GROUP BY DATE(s.occurred_at), s.location_id, s.channel_id, s.daypart_id
		ON CONFLICT (date, location_id, channel_id, daypart_id)
		DO UPDATE SET
			revenue = EXCLUDED.revenue,
			cogs = EXCLUDED.cogs,
			gross_margin = EXCLUDED.gross_margin,
			covers = EXCLUDED.covers,
			avg_check = EXCLUDED.avg_check,
			discounts = EXCLUDED.discounts,
			comps = EXCLUDED.comps,
			freshness_timestamp = NOW(),
			updated_at = NOW()
	`

	_, err := pool.Exec(ctx, query, date, locationID)
	if err != nil {
		return err
	}

	// Update labor costs from payroll periods
	laborQuery := `
		UPDATE kpi_aggregates k
		SET
			labor_cost = COALESCE(p.labor_cost, 0),
			labor_pct = CASE WHEN k.revenue > 0 THEN COALESCE(p.labor_cost, 0) / k.revenue * 100 ELSE 0 END,
			net_profit = k.gross_margin - COALESCE(p.labor_cost, 0) - k.opex,
			updated_at = NOW()
		FROM (
			SELECT
				SUM(labor_cost) / (end_date - start_date + 1) as labor_cost
			FROM payroll_periods
			WHERE start_date <= $1 AND end_date >= $1
		) p
		WHERE k.date = $1 AND k.location_id = $2
	`

	_, err = pool.Exec(ctx, laborQuery, date, locationID)
	return err
}
