# Tasks: Lakehouse Restaurant Finance MVP

**Input**: Design documents from `/specs/001-restaurant-finance/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/api.yaml

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure) ✅

- [x] T001 Create Docker Compose with Postgres service and volumes in docker/docker-compose.yml
- [x] T002 Initialize Go module and dependencies in backend/go.mod and backend/go.sum
- [x] T003 Scaffold Next.js 14 + Tailwind app in frontend/ (app router, tsconfig, tailwind.config.js, postcss.config.js)
- [x] T004 Add repo tooling configs (.editorconfig, backend/.golangci.yml, frontend/.eslintrc.js, frontend/prettier.config.js)
- [x] T005 Add environment examples for backend and frontend in docker/env/.env.example and frontend/.env.example

---

## Phase 2: Foundational (Blocking Prerequisites) ✅

- [x] T006 Create base SQL migrations for core tables (users, roles, service_channels, dayparts, menu_items, sales, sale_lines, payroll_periods, inventory_snapshots, mapping_profiles, import_jobs, import_anomalies, kpi_aggregates, export_jobs) in backend/migrations/
- [x] T007 Add migration runner command in backend/cmd/migrate/main.go (wire to Postgres DSN env)
- [x] T008 Implement backend config loader (DB, storage paths, JWT/secret) in backend/internal/config/config.go
- [x] T009 Implement RBAC roles and middleware (Owner/Admin, Manager, Accountant, Viewer) in backend/internal/auth/
- [x] T010 Set up HTTP server and router with middleware (logging, recovery, request ID, CORS) in backend/internal/api/router.go
- [x] T011 Implement file storage helpers for uploads/exports in backend/internal/storage/filesystem.go
- [x] T012 Configure sqlc for Postgres and generate stubs in backend/sqlc.yaml and backend/internal/storage/queries.sql
- [x] T013 Seed default dayparts and service channels in backend/migrations/seed_dayparts_channels.sql
- [x] T014 Create frontend layout shell with Tailwind base styles in frontend/app/layout.tsx and frontend/app/globals.css
- [x] T015 Add API client with base URL/env handling and TanStack Query setup in frontend/lib/api-client.ts and frontend/lib/query-client.ts
- [x] T016 Add timezone/format utilities defaulting to Australia/Brisbane in frontend/lib/dates.ts

**Checkpoint**: Foundation ready — user story work can proceed. ✅

---

## Phase 3: User Story 1 - Daily finance snapshot for decisions (Priority: P1) 🎯 MVP ✅

**Goal**: Daily KPI dashboard with revenue/COGS/labor/OpEx/net profit, channel/daypart splits, freshness.
**Independent Test**: Import sample CSVs, load /dashboard, verify KPIs and freshness without other stories.

### Implementation

- [x] T017 [P] [US1] Implement KPI aggregate queries and repository in backend/internal/kpi/store.go
- [x] T018 [US1] Implement KPI service with freshness timestamp and range/channel/daypart filters in backend/internal/kpi/service.go
- [x] T019 [US1] Expose GET /kpi/daily handler in backend/internal/api/kpi_handlers.go
- [x] T020 [P] [US1] Add aggregate refresh worker triggered after imports in backend/cmd/worker/aggregates.go
- [x] T021 [P] [US1] Create KPI daily query hook in frontend/src/hooks/use-kpi.ts
- [x] T022 [P] [US1] Build KPI cards and freshness banner components in frontend/src/components/kpi/KpiCards.tsx
- [x] T023 [P] [US1] Build channel/daypart D3 charts (bar/stacked) in frontend/src/components/kpi/ChannelDaypartCharts.tsx
- [x] T024 [US1] Assemble dashboard page with filters (range, channel, daypart) in frontend/src/app/dashboard/page.tsx

**Checkpoint**: Dashboard loads KPIs with freshness and breakdowns using imported data. ✅

---

## Phase 4: User Story 2 - Import and reconcile source data (Priority: P1) ✅

**Goal**: Reliable CSV imports for POS, payroll, inventory with saved mappings, idempotency, and anomaly surfacing.
**Independent Test**: Upload sample files, confirm row counts and anomalies in import log; no other stories required.

### Implementation

- [x] T025 [P] [US2] Implement mapping profile persistence and retrieval in backend/internal/imports/mappings.go
- [x] T026 [P] [US2] Implement CSV parsers/validators for POS, payroll, inventory in backend/internal/imports/parsers.go
- [x] T027 [US2] Implement import pipeline with idempotency (file hash + keys), upserts, and anomaly recording in backend/internal/imports/pipeline.go
- [x] T028 [US2] Expose POST /imports and GET /imports/{id} handlers in backend/internal/api/import_handlers.go
- [x] T029 [P] [US2] Implement import log repository and status polling support in backend/internal/imports/store.go
- [x] T030 [P] [US2] Build import upload page with source type, mapping picker, file chooser, and status poller in frontend/src/app/imports/page.tsx
- [x] T031 [P] [US2] Build mapping profile editor/selector component in frontend/src/components/imports/MappingProfileForm.tsx
- [x] T032 [US2] Build import log and anomaly table view in frontend/src/components/imports/ImportLogTable.tsx

**Checkpoint**: Imports run idempotently with anomalies surfaced and visible in UI. ✅

---

## Phase 5: User Story 3 - Drill-down and exports (Priority: P2) ✅

**Goal**: Drill into transactions and export P&L/channel summaries with freshness.
**Independent Test**: From existing data, view drill-down table and generate/export CSV matching dashboard totals.

### Implementation

- [x] T033 [P] [US3] Implement sales drill-down query and GET /kpi/drilldown/sales handler in backend/internal/api/drilldown_handlers.go
- [x] T034 [P] [US3] Implement export job creation and CSV writer for P&L/channel summary in backend/internal/exports/service.go
- [x] T035 [US3] Expose GET /exports/pnl and GET /exports/{id} handlers in backend/internal/api/export_handlers.go
- [x] T036 [P] [US3] Build drill-down table page with filters and pagination in frontend/src/app/drilldown/page.tsx
- [x] T037 [P] [US3] Build export trigger/status/download UI in frontend/src/components/exports/ExportPanel.tsx
- [x] T038 [US3] Add export/query hooks ensuring totals match on-screen KPIs in frontend/src/hooks/use-exports.ts

**Checkpoint**: Drill-down and exports function independently atop existing data. ✅

---

## Phase 6: Polish & Cross-Cutting ✅

- [x] T039 [P] Add structured logging and request IDs in backend/internal/api/router.go (via chi middleware)
- [x] T040 [P] Add sample CSV fixtures and notes in specs/001-restaurant-finance/fixtures/ (sample-pos.csv, sample-payroll.csv, sample-inventory.csv, README.md)
- [x] T041 [P] Validate and update quickstart with actual commands in specs/001-restaurant-finance/quickstart.md
- [x] T042 Harden config validation and file upload safety (size limits, extensions) in backend/internal/config/validate.go

---

## Summary

**Total Tasks**: 42  
**Completed**: 42 ✅

### Tasks by Phase

| Phase | Description    | Tasks          | Status      |
| ----- | -------------- | -------------- | ----------- |
| 1     | Setup          | T001-T005 (5)  | ✅ Complete |
| 2     | Foundational   | T006-T016 (11) | ✅ Complete |
| 3     | US1: Dashboard | T017-T024 (8)  | ✅ Complete |
| 4     | US2: Imports   | T025-T032 (8)  | ✅ Complete |
| 5     | US3: Exports   | T033-T038 (6)  | ✅ Complete |
| 6     | Polish         | T039-T042 (4)  | ✅ Complete |

### Key Files Created

**Backend (Go)**

- `backend/cmd/api/main.go` - HTTP server entrypoint
- `backend/cmd/migrate/main.go` - Migration runner
- `backend/cmd/worker/aggregates.go` - KPI aggregate refresh worker
- `backend/internal/api/router.go` - Chi router with middleware
- `backend/internal/api/kpi_handlers.go` - GET /kpi/daily endpoint
- `backend/internal/api/import_handlers.go` - POST/GET /imports endpoints
- `backend/internal/api/export_handlers.go` - Export endpoints
- `backend/internal/api/drilldown_handlers.go` - GET /kpi/drilldown/sales
- `backend/internal/auth/` - JWT, middleware, RBAC roles
- `backend/internal/config/` - Config loader and validation
- `backend/internal/kpi/` - KPI store and service
- `backend/internal/imports/` - Import pipeline, parsers, mappings
- `backend/internal/exports/` - Export service with CSV writer
- `backend/internal/storage/` - File system helpers
- `backend/migrations/` - SQL migrations and seeds

**Frontend (Next.js 14 + TypeScript)**

- `frontend/src/app/dashboard/page.tsx` - Main KPI dashboard
- `frontend/src/app/imports/page.tsx` - CSV import page
- `frontend/src/app/drilldown/page.tsx` - Transaction drill-down
- `frontend/src/components/kpi/` - KpiCards, ChannelDaypartCharts
- `frontend/src/components/imports/` - MappingProfileForm, ImportLogTable
- `frontend/src/components/exports/` - ExportPanel
- `frontend/src/hooks/` - use-kpi, use-imports, use-exports
- `frontend/src/lib/` - API client, query-client, date utilities

**Configuration & Documentation**

- `docker/docker-compose.yml` - PostgreSQL service
- `specs/001-restaurant-finance/quickstart.md` - Setup and usage guide
- `specs/001-restaurant-finance/fixtures/` - Sample CSV files for testing

---

## Dependencies & Execution Order

- Setup (Phase 1) → Foundational (Phase 2) → User stories (Phases 3–5) → Polish (Phase 6).
- User stories can run in parallel after Phase 2; prioritize P1 (US1, US2) before P2 (US3).
- Within each story: storage/models → services → handlers → frontend hooks/components → pages.

## Parallel Execution Examples

- US1: T017, T020, T021, T022, T023 can proceed in parallel; T018/T019 and T024 land after data ready.
- US2: T025, T026, T029, T030, T031 can run in parallel; T027/T028 depend on parsers and mappings; T032 after log API.
- US3: T033 and T034 can run in parallel; T035 depends on both; T036/T037/T038 can run after handlers stubbed.

## Implementation Strategy

- MVP first: complete Phases 1–2, then US1 to deliver usable dashboard; demo and validate.
- Incremental: add US2 imports next for live data; add US3 for drill-down/exports.
- Keep RBAC and audit in place from foundational phase; refresh aggregates after each import.
