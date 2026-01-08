# Implementation Plan: Lakehouse Restaurant Finance MVP

**Branch**: `001-restaurant-finance` | **Date**: 2026-01-08 | **Spec**: [specs/001-restaurant-finance/spec.md](specs/001-restaurant-finance/spec.md)
**Input**: Feature specification from `/specs/001-restaurant-finance/spec.md`

## Summary

Build a restaurant finance web app for The Lakehouse with CSV-based ingestion (POS, payroll, inventory), daily KPI dashboard, channel/daypart drill-downs, and exports. Frontend: Next.js + Tailwind + D3 for charts. Backend: Go service with REST API, PostgreSQL (Docker), role-based access, audit/import logs, and idempotent imports.

## Technical Context

**Language/Version**: Go 1.22 (backend), TypeScript/Next.js 14 (frontend).  
**Primary Dependencies**: Go: chi/echo-class router (chi), sqlc/pgx for Postgres, testify for tests; Frontend: Next.js 14, React 18, Tailwind CSS, D3.js, TanStack Query for data fetching.  
**Storage**: PostgreSQL 15 via Docker; blob storage simulated with local filesystem for imports/exports.  
**Testing**: Go test + testify (unit/integration); SQL tests via dockerized Postgres; frontend with Vitest + Testing Library + Playwright for basic flows.  
**Target Platform**: Dockerized services; local dev on macOS; deployable to container runtime.  
**Project Type**: Web app with separate frontend (Next.js) and backend (Go API).  
**Performance Goals**: Dashboard p95 < 1s for 30-day data; exports < 10s for month; imports process typical day CSV (<10k rows) in < 30s.  
**Constraints**: Data freshness visible; idempotent imports; no silent drops; minimal PII; timezone locked to Brisbane in UI and backend.  
**Scale/Scope**: Single venue, low concurrency (<20 active users); data volumes daily POS/payroll/inventory for one site.

## Constitution Check (gate)

- Metrics: Include revenue, discounts/comps, COGS, Gross Margin, labor %, OpEx %, net profit, cash balance, freshness timestamp, channel/daypart splits. → Addressed in spec.
- Data handling: Idempotent CSV imports; anomaly surfacing; no silent drops. → Plan requires import log and anomaly reporting.
- Privacy/Security: Minimize PII; encrypt in transit/at rest; audit imports/mappings/calculation definitions. → Plan includes HTTPS, role-based access, audit logs; assume TLS termination in deployment; local dev via HTTP acceptable.
- Reliability/Performance: KPI views target <1s; background ETL; retries for connectors. → Plan includes p95 targets and background import processing with retries.
- UX: Plain language, consistent units, drill-down, channel/daypart, patio/indoor if available. → Plan includes these displays.
- Compliance: Not a tax system; exports for accounting; respect privacy laws. → Covered by exports + data minimization.

Gate status: PASS (no violations). Re-check after Phase 1 deliverables.

## Project Structure

### Documentation (this feature)

```text
specs/001-restaurant-finance/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md (future /speckit.tasks)
```

### Source Code (intended)

```text
backend/
├── cmd/api/
├── internal/
│   ├── api/        # handlers, routes
│   ├── auth/       # RBAC, roles
│   ├── imports/    # CSV ingestion, mapping, anomalies
│   ├── kpi/        # KPI aggregations, daypart/channel splits
│   ├── storage/    # Postgres access (sqlc/pgx)
│   └── exports/    # P&L/channel exports
├── migrations/
└── tests/

frontend/
├── app/            # Next.js 14 app router
├── components/     # Charts (D3), tables, filters
├── lib/            # API client, hooks (TanStack Query)
└── tests/          # Vitest/RTL, Playwright

docker/
├── docker-compose.yml
└── env/
```

**Structure Decision**: Separate frontend (Next.js) and backend (Go) with shared Docker Compose for Postgres; documentation stays under specs/001-restaurant-finance.

## Complexity Tracking

No constitution violations to justify.

## Phase 0 Output

- Created [research.md](research.md) with stack, storage, import/anomaly, KPI, security, and timezone decisions; no open items.

## Phase 1 Output

- Created [data-model.md](data-model.md) capturing entities (sales, payroll, inventory, imports/anomalies, KPIs, exports, users).
- Added API contract [contracts/api.yaml](contracts/api.yaml) for imports, KPIs, drill-down, exports.
- Added [quickstart.md](quickstart.md) with planned dev runbook and workflows.

## Constitution Check (post-design)

- Metrics, channel/daypart, freshness, idempotent imports, anomaly surfacing, audit, and RBAC reflected in data model and APIs.
- Privacy/security: minimal PII (email only), HTTPS assumed at ingress, audit for imports/exports; data retention configurable later.
- Reliability/performance: pre-aggregation for KPIs; background imports; clear status endpoints.
- UX: channel/daypart splits, drill-down, exports; Brisbane timezone surfaced via API parameters and UI convention.
- Compliance: exports for accounting; not a tax system.

Gate status: PASS.
