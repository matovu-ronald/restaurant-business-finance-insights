-- 001_init.down.sql
DROP TABLE IF EXISTS export_jobs;
DROP TABLE IF EXISTS kpi_aggregates;
DROP TABLE IF EXISTS inventory_snapshots;
DROP TABLE IF EXISTS payroll_periods;
DROP TABLE IF EXISTS sale_lines;
DROP TABLE IF EXISTS sales;
DROP TABLE IF EXISTS import_anomalies;
DROP TABLE IF EXISTS import_jobs;
DROP TABLE IF EXISTS mapping_profiles;
DROP TABLE IF EXISTS menu_items;
DROP TABLE IF EXISTS dayparts;
DROP TABLE IF EXISTS service_channels;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS export_type;
DROP TYPE IF EXISTS export_status;
DROP TYPE IF EXISTS source_type;
DROP TYPE IF EXISTS anomaly_code;
DROP TYPE IF EXISTS import_status;
DROP TYPE IF EXISTS user_role;

DROP EXTENSION IF EXISTS "uuid-ossp";
