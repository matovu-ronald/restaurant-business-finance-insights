# Data Model — Lakehouse Restaurant Finance MVP

## Entities

- Location
  - id, name, timezone (AEST/AEDT), seating_capacity_indoor, seating_capacity_patio
- ServiceChannel
  - id, code (dine-in, takeaway, pickup, catering), display_name
- Daypart
  - id, code (breakfast, lunch, dinner), start_time, end_time
- MenuItem
  - id, name, category, recipe_cost, price, is_active
- Sale
  - id, occurred_at (UTC), location_id, channel_id, daypart_id, subtotal, discounts, comps, tax, total, payment_method, check_number, source_file_hash
- SaleLine
  - id, sale_id, menu_item_id, quantity, unit_price, line_subtotal, line_discounts, line_comps
- PayrollPeriod
  - id, start_date, end_date, role_category, labor_cost, hours, source_file_hash
- InventorySnapshot
  - id, snapshot_date, menu_item_id, item_cost, source_file_hash
- ImportJob
  - id, source_type (pos, payroll, inventory), file_hash, filename, mapping_profile_id, status, row_count, anomaly_count, started_at, completed_at, user_id, notes
- ImportAnomaly
  - id, import_job_id, row_number, field, code (missing_channel, negative_total, bad_date, duplicate_row), details
- MappingProfile
  - id, source_type, name, column_mappings (JSON), created_by, created_at
- KPIAggregate
  - id, date, location_id, channel_id, daypart_id, revenue, cogs, gross_margin, labor_cost, labor_pct, opex, net_profit, covers, avg_check, discounts, comps, freshness_timestamp
- User
  - id, email, role (owner_admin, manager, accountant, viewer), password_hash (or external auth id), created_at, last_login
- ExportJob
  - id, export_type (pnl, channel_summary), period_start, period_end, status, file_path, requested_by, requested_at, completed_at

## Relationships

- Sale has many SaleLines; Sale belongs to ServiceChannel, Daypart, Location.
- SaleLine references MenuItem and inherits channel/daypart via Sale.
- PayrollPeriod associates labor costs to date ranges and aggregates per daypart during processing.
- InventorySnapshot ties item cost to MenuItem for COGS calculations.
- ImportJob has many ImportAnomaly; ImportJob may reference MappingProfile.
- KPIAggregate derived from Sales, PayrollPeriod, InventorySnapshot grouped by date/channel/daypart/location.
- ExportJob references generated CSVs based on KPIAggregate and transactional detail.
- User performs ImportJob and ExportJob actions (audit trail).

## Notes

- Timestamps stored in UTC; displayed in Australia/Brisbane.
- Monetary fields stored as decimal with currency AUD; avoid floating point for totals.
- Idempotency via file_hash + natural keys (date/channel/register/check_number) per source type.
- Daypart boundaries configurable but default to breakfast/lunch/dinner windows.
