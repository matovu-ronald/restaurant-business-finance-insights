# Quickstart — Lakehouse Restaurant Finance MVP

## Prerequisites

- Docker + Docker Compose (PostgreSQL 15)
- Node.js 20+ (frontend)
- Go 1.22+ (backend)

## 1. Start Database

```bash
# From repository root
docker compose -f docker/docker-compose.yml up -d postgres
```

Wait for health check to pass:

```bash
docker compose -f docker/docker-compose.yml ps
```

## 2. Configure Environment

```bash
# Backend environment
cp docker/env/.env.example docker/env/.env
# Edit docker/env/.env and set:
#   DATABASE_URL=postgres://lakehouse:lakehouse_dev@localhost:5432/lakehouse?sslmode=disable
#   JWT_SECRET=your-secret-key-at-least-32-chars
#   STORAGE_PATH=./uploads
```

## 3. Run Migrations

```bash
cd backend
go run ./cmd/migrate up
```

This creates all tables and seeds default dayparts/channels.

## 4. Start Backend API

```bash
cd backend
go run ./cmd/api
# Server starts on :8080
```

## 5. Start Frontend

```bash
cd frontend
npm install
npm run dev
# App starts on :3000
```

## 6. Import Sample Data

Sample CSV files are in `specs/001-restaurant-finance/fixtures/`:

- `sample-pos.csv` — 24 sales transactions
- `sample-payroll.csv` — 10 payroll records over 2 periods
- `sample-inventory.csv` — 16 inventory items

Import via UI at http://localhost:3000/imports or via API:

```bash
# Import POS data
curl -X POST http://localhost:8080/imports \
  -F "file=@specs/001-restaurant-finance/fixtures/sample-pos.csv" \
  -F "source_type=pos"

# Import payroll
curl -X POST http://localhost:8080/imports \
  -F "file=@specs/001-restaurant-finance/fixtures/sample-payroll.csv" \
  -F "source_type=payroll"

# Import inventory
curl -X POST http://localhost:8080/imports \
  -F "file=@specs/001-restaurant-finance/fixtures/sample-inventory.csv" \
  -F "source_type=inventory"
```

## API Endpoints

### KPI Dashboard

```bash
# Get daily KPIs (default 30 days)
GET /kpi/daily

# With filters
GET /kpi/daily?range=ytd&channel=dine_in
GET /kpi/daily?start=2024-01-01&end=2024-01-31
```

### Imports

```bash
# Upload CSV (multipart form)
POST /imports
  - file: CSV file
  - source_type: pos | payroll | inventory
  - mapping_profile_id: (optional) UUID

# Get import status
GET /imports/{id}

# List recent imports
GET /imports
```

### Drill-down

```bash
# Query sales with filters and pagination
GET /kpi/drilldown/sales?start=2024-01-01&end=2024-01-31&channel=dine_in&page=1&per_page=50
```

### Exports

```bash
# Generate P&L export (returns CSV directly)
GET /exports/pnl?start=2024-01-01&end=2024-01-31

# List export jobs
GET /exports
```

### Mappings

```bash
# Get default mappings for a source type
GET /mappings?source_type=pos

# Save a custom mapping profile
POST /mappings
  {"name": "Custom POS", "source_type": "pos", "mappings": {...}}
```

## Verification

1. Open http://localhost:3000/dashboard — see KPI cards with freshness timestamp
2. Check channel/daypart breakdown charts
3. Navigate to /imports — upload sample CSVs
4. Navigate to /drilldown — filter and paginate sales
5. Use ExportPanel to download P&L as CSV

## Testing

```bash
# Backend tests
cd backend && go test ./...

# Frontend tests (when added)
cd frontend && npm test

# Run all with fresh database
docker compose -f docker/docker-compose.yml down -v
docker compose -f docker/docker-compose.yml up -d postgres
cd backend && go run ./cmd/migrate up
go test ./...
```

## Timezone

All dates/times use **Australia/Brisbane** (AEST/AEDT). Currency is **AUD**.

## Troubleshooting

- **Database connection refused**: Ensure postgres container is healthy
- **Migration fails**: Check DATABASE_URL in environment
- **Import errors**: Verify CSV columns match expected headers (see fixtures/README.md)
- **CORS errors**: Backend CORS middleware allows localhost:3000 by default
