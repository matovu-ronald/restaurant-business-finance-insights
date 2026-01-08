# Restaurant Finance Web App — Minimal Constitution

## Purpose & Scope

- Serve The Lakehouse Restaurant (fictional cafe/restaurant in Brisbane) with clear, reliable finance insights.
- Support decisions on pricing, menu mix, labor planning, cost control, and cash planning across dine-in, patio, takeaway, online pickup, catering, and small gatherings.

## Core Metrics (non-negotiable)

- Revenue, discounts, comps; COGS; Gross Margin; Labor Cost %; Operating Expense %; Net Profit.
- Cash balance and recent cash flow; data freshness timestamp on every view.
- Time windows: daily, weekly, monthly, YTD, trailing 12 months; location-level rollups (indoor vs. patio if relevant) and overall.
- Volume anchors: covers, average check, item/mix performance, daypart (breakfast/lunch/dinner), and service channel (dine-in, takeaway, pickup, catering).

## Data Sources (MVP)

- POS sales with tenders/discounts; payroll summaries; accounting ledger (GL/P&L); inventory snapshots and recipe costing; optional reservations/online orders CSVs.
- Start with CSV import; add API connectors later without breaking contracts.
- Idempotent imports; never silently drop records—flag and surface anomalies with clear remediation steps.

## Access & Roles

- Roles: Owner/Admin, Manager, Accountant, Viewer (least-privilege by default).
- Audit significant changes (imports, mapping rules, calculated metric definitions, retention settings).

## Privacy & Security

- Minimize PII; avoid collecting sensitive personal data unless essential.
- Encrypt in transit and at rest; configurable data retention and export/delete on request.
- Transparent calculations: formulas visible or easily inspectable.

## Reliability & Performance

- KPI views target < 1s load on typical datasets; heavy ETL runs in background.
- Clear status for data freshness and processing; retries with backoff for connectors.

## UX Principles

- Plain language labels; consistent units and definitions across all views.
- Single source of truth per metric; drill-down to underlying transactions where available.
- Highlight daypart/channel splits and patio vs. indoor seating impact where data allows.

## Compliance

- Respect applicable regional privacy laws (e.g., GDPR/CCPA) based on user locale.
- Not a system of record for tax filing; provide exports to accounting systems.

## Non-Goals

- Not a full accounting system; not inventory management; not HR/payroll processing.
- Avoid complex custom forecasting beyond simple trendlines in MVP.

## Governance

- This constitution sets non-negotiables for design and implementation.
- Amendments require rationale, user impact assessment, and migration plan.

**Version**: 0.1.0 | **Ratified**: 2026-01-08 | **Last Amended**: 2026-01-08
