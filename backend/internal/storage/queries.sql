-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetLocation :one
SELECT * FROM locations LIMIT 1;

-- name: ListServiceChannels :many
SELECT * FROM service_channels ORDER BY display_name;

-- name: GetServiceChannelByCode :one
SELECT * FROM service_channels WHERE code = $1 LIMIT 1;

-- name: ListDayparts :many
SELECT * FROM dayparts ORDER BY start_time;

-- name: GetDaypartByCode :one
SELECT * FROM dayparts WHERE code = $1 LIMIT 1;

-- name: GetDaypartByTime :one
SELECT * FROM dayparts WHERE start_time <= $1 AND end_time > $1 LIMIT 1;

-- name: CreateMappingProfile :one
INSERT INTO mapping_profiles (source_type, name, column_mappings, created_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetMappingProfile :one
SELECT * FROM mapping_profiles WHERE id = $1;

-- name: ListMappingProfiles :many
SELECT * FROM mapping_profiles WHERE source_type = $1 ORDER BY created_at DESC;

-- name: CreateImportJob :one
INSERT INTO import_jobs (source_type, file_hash, filename, mapping_profile_id, user_id, status)
VALUES ($1, $2, $3, $4, $5, 'pending')
RETURNING *;

-- name: GetImportJob :one
SELECT * FROM import_jobs WHERE id = $1;

-- name: UpdateImportJobStatus :exec
UPDATE import_jobs
SET status = $2, row_count = $3, anomaly_count = $4, started_at = $5, completed_at = $6
WHERE id = $1;

-- name: CreateImportAnomaly :one
INSERT INTO import_anomalies (import_job_id, row_number, field, code, details)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListImportAnomalies :many
SELECT * FROM import_anomalies WHERE import_job_id = $1 ORDER BY row_number;

-- name: CreateSale :one
INSERT INTO sales (occurred_at, location_id, channel_id, daypart_id, subtotal, discounts, comps, tax, total, payment_method, check_number, source_file_hash)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: CreateSaleLine :one
INSERT INTO sale_lines (sale_id, menu_item_id, quantity, unit_price, line_subtotal, line_discounts, line_comps)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetSalesByDateRange :many
SELECT s.*, sc.code as channel_code, sc.display_name as channel_name, d.code as daypart_code, d.display_name as daypart_name
FROM sales s
JOIN service_channels sc ON s.channel_id = sc.id
JOIN dayparts d ON s.daypart_id = d.id
WHERE s.occurred_at >= $1 AND s.occurred_at < $2
ORDER BY s.occurred_at DESC;

-- name: CreatePayrollPeriod :one
INSERT INTO payroll_periods (start_date, end_date, role_category, labor_cost, hours, source_file_hash)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetPayrollByDateRange :many
SELECT * FROM payroll_periods
WHERE start_date <= $2 AND end_date >= $1
ORDER BY start_date;

-- name: CreateInventorySnapshot :one
INSERT INTO inventory_snapshots (snapshot_date, menu_item_id, item_cost, source_file_hash)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLatestInventoryCosts :many
SELECT DISTINCT ON (menu_item_id) *
FROM inventory_snapshots
WHERE snapshot_date <= $1
ORDER BY menu_item_id, snapshot_date DESC;

-- name: UpsertKPIAggregate :one
INSERT INTO kpi_aggregates (date, location_id, channel_id, daypart_id, revenue, cogs, gross_margin, labor_cost, labor_pct, opex, net_profit, covers, avg_check, discounts, comps, freshness_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW())
ON CONFLICT (date, location_id, channel_id, daypart_id)
DO UPDATE SET
    revenue = EXCLUDED.revenue,
    cogs = EXCLUDED.cogs,
    gross_margin = EXCLUDED.gross_margin,
    labor_cost = EXCLUDED.labor_cost,
    labor_pct = EXCLUDED.labor_pct,
    opex = EXCLUDED.opex,
    net_profit = EXCLUDED.net_profit,
    covers = EXCLUDED.covers,
    avg_check = EXCLUDED.avg_check,
    discounts = EXCLUDED.discounts,
    comps = EXCLUDED.comps,
    freshness_timestamp = NOW(),
    updated_at = NOW()
RETURNING *;

-- name: GetKPIAggregates :many
SELECT k.*, sc.code as channel_code, sc.display_name as channel_name, d.code as daypart_code, d.display_name as daypart_name
FROM kpi_aggregates k
LEFT JOIN service_channels sc ON k.channel_id = sc.id
LEFT JOIN dayparts d ON k.daypart_id = d.id
WHERE k.date >= $1 AND k.date <= $2
ORDER BY k.date DESC, sc.display_name, d.start_time;

-- name: GetKPITotals :one
SELECT
    SUM(revenue) as revenue,
    SUM(cogs) as cogs,
    SUM(gross_margin) as gross_margin,
    SUM(labor_cost) as labor_cost,
    CASE WHEN SUM(revenue) > 0 THEN SUM(labor_cost) / SUM(revenue) * 100 ELSE 0 END as labor_pct,
    SUM(opex) as opex,
    SUM(net_profit) as net_profit,
    SUM(covers) as covers,
    CASE WHEN SUM(covers) > 0 THEN SUM(revenue) / SUM(covers) ELSE 0 END as avg_check,
    SUM(discounts) as discounts,
    SUM(comps) as comps,
    MAX(freshness_timestamp) as freshness_timestamp
FROM kpi_aggregates
WHERE date >= $1 AND date <= $2;

-- name: CreateExportJob :one
INSERT INTO export_jobs (export_type, period_start, period_end, requested_by, status)
VALUES ($1, $2, $3, $4, 'pending')
RETURNING *;

-- name: GetExportJob :one
SELECT * FROM export_jobs WHERE id = $1;

-- name: UpdateExportJobStatus :exec
UPDATE export_jobs
SET status = $2, file_path = $3, completed_at = $4
WHERE id = $1;

-- name: CreateMenuItem :one
INSERT INTO menu_items (name, category, recipe_cost, price)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetMenuItemByName :one
SELECT * FROM menu_items WHERE name = $1 LIMIT 1;

-- name: ListMenuItems :many
SELECT * FROM menu_items WHERE is_active = true ORDER BY category, name;

-- name: UpdateMenuItemCost :exec
UPDATE menu_items SET recipe_cost = $2, updated_at = NOW() WHERE id = $1;
