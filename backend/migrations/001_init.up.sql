-- 001_init.up.sql
-- Core schema for Lakehouse Restaurant Finance MVP

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enum types
CREATE TYPE user_role AS ENUM ('owner_admin', 'manager', 'accountant', 'viewer');
CREATE TYPE import_status AS ENUM ('pending', 'processing', 'completed', 'failed');
CREATE TYPE export_status AS ENUM ('pending', 'processing', 'completed', 'failed');
CREATE TYPE anomaly_code AS ENUM ('missing_channel', 'negative_total', 'bad_date', 'duplicate_row', 'missing_column');
CREATE TYPE source_type AS ENUM ('pos', 'payroll', 'inventory');
CREATE TYPE export_type AS ENUM ('pnl', 'channel_summary');

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role user_role NOT NULL DEFAULT 'viewer',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login TIMESTAMPTZ
);

-- Location (single venue for MVP)
CREATE TABLE locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'Australia/Brisbane',
    seating_capacity_indoor INT NOT NULL DEFAULT 70,
    seating_capacity_patio INT NOT NULL DEFAULT 30,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Service Channels
CREATE TABLE service_channels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL
);

-- Dayparts
CREATE TABLE dayparts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL
);

-- Menu Items
CREATE TABLE menu_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    recipe_cost DECIMAL(10, 2) NOT NULL DEFAULT 0,
    price DECIMAL(10, 2) NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Mapping Profiles (for CSV imports)
CREATE TABLE mapping_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_type source_type NOT NULL,
    name VARCHAR(255) NOT NULL,
    column_mappings JSONB NOT NULL DEFAULT '{}',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Import Jobs
CREATE TABLE import_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_type source_type NOT NULL,
    file_hash VARCHAR(64) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    mapping_profile_id UUID REFERENCES mapping_profiles(id),
    status import_status NOT NULL DEFAULT 'pending',
    row_count INT NOT NULL DEFAULT 0,
    anomaly_count INT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    user_id UUID REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Import Anomalies
CREATE TABLE import_anomalies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    import_job_id UUID NOT NULL REFERENCES import_jobs(id) ON DELETE CASCADE,
    row_number INT NOT NULL,
    field VARCHAR(100),
    code anomaly_code NOT NULL,
    details TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Sales
CREATE TABLE sales (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    occurred_at TIMESTAMPTZ NOT NULL,
    location_id UUID NOT NULL REFERENCES locations(id),
    channel_id UUID NOT NULL REFERENCES service_channels(id),
    daypart_id UUID NOT NULL REFERENCES dayparts(id),
    subtotal DECIMAL(10, 2) NOT NULL DEFAULT 0,
    discounts DECIMAL(10, 2) NOT NULL DEFAULT 0,
    comps DECIMAL(10, 2) NOT NULL DEFAULT 0,
    tax DECIMAL(10, 2) NOT NULL DEFAULT 0,
    total DECIMAL(10, 2) NOT NULL DEFAULT 0,
    payment_method VARCHAR(50),
    check_number VARCHAR(50),
    source_file_hash VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_sales_occurred_at ON sales(occurred_at);
CREATE INDEX idx_sales_channel_id ON sales(channel_id);
CREATE INDEX idx_sales_daypart_id ON sales(daypart_id);

-- Sale Lines
CREATE TABLE sale_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sale_id UUID NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
    menu_item_id UUID REFERENCES menu_items(id),
    quantity DECIMAL(10, 2) NOT NULL DEFAULT 1,
    unit_price DECIMAL(10, 2) NOT NULL DEFAULT 0,
    line_subtotal DECIMAL(10, 2) NOT NULL DEFAULT 0,
    line_discounts DECIMAL(10, 2) NOT NULL DEFAULT 0,
    line_comps DECIMAL(10, 2) NOT NULL DEFAULT 0
);
CREATE INDEX idx_sale_lines_sale_id ON sale_lines(sale_id);

-- Payroll Periods
CREATE TABLE payroll_periods (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    role_category VARCHAR(100),
    labor_cost DECIMAL(10, 2) NOT NULL DEFAULT 0,
    hours DECIMAL(10, 2) NOT NULL DEFAULT 0,
    source_file_hash VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_payroll_periods_dates ON payroll_periods(start_date, end_date);

-- Inventory Snapshots
CREATE TABLE inventory_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    snapshot_date DATE NOT NULL,
    menu_item_id UUID NOT NULL REFERENCES menu_items(id),
    item_cost DECIMAL(10, 2) NOT NULL DEFAULT 0,
    source_file_hash VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_inventory_snapshots_date ON inventory_snapshots(snapshot_date);

-- KPI Aggregates (materialized daily summaries)
CREATE TABLE kpi_aggregates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date DATE NOT NULL,
    location_id UUID NOT NULL REFERENCES locations(id),
    channel_id UUID REFERENCES service_channels(id),
    daypart_id UUID REFERENCES dayparts(id),
    revenue DECIMAL(12, 2) NOT NULL DEFAULT 0,
    cogs DECIMAL(12, 2) NOT NULL DEFAULT 0,
    gross_margin DECIMAL(12, 2) NOT NULL DEFAULT 0,
    labor_cost DECIMAL(12, 2) NOT NULL DEFAULT 0,
    labor_pct DECIMAL(5, 2) NOT NULL DEFAULT 0,
    opex DECIMAL(12, 2) NOT NULL DEFAULT 0,
    net_profit DECIMAL(12, 2) NOT NULL DEFAULT 0,
    covers INT NOT NULL DEFAULT 0,
    avg_check DECIMAL(10, 2) NOT NULL DEFAULT 0,
    discounts DECIMAL(12, 2) NOT NULL DEFAULT 0,
    comps DECIMAL(12, 2) NOT NULL DEFAULT 0,
    freshness_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (date, location_id, channel_id, daypart_id)
);
CREATE INDEX idx_kpi_aggregates_date ON kpi_aggregates(date);

-- Export Jobs
CREATE TABLE export_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    export_type export_type NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    status export_status NOT NULL DEFAULT 'pending',
    file_path VARCHAR(500),
    requested_by UUID REFERENCES users(id),
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);
