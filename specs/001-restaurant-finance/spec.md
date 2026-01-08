# Feature Specification: Lakehouse Restaurant Finance MVP

**Feature Branch**: `001-restaurant-finance`  
**Created**: 2026-01-08  
**Status**: Draft  
**Input**: User description: "The Lakehouse Restaurant is a fictional casual dining cafe/restaurant in Brisbane, Australia, situated near a lake (such as Northshore Hamilton or Forest Lake) in a 180–200 square meter space with indoor seating for 60–70 guests and an outdoor patio for 30 more, offering scenic lakeside views. Operating as a Proprietary Limited Company, it serves breakfast, lunch, and dinner seven days a week (7:00 AM–8:00 PM weekdays, extended hours on weekends, and shorter on Sundays), featuring approachable cafe classics like eggs, burgers, steaks, pasta, salads, a children's menu, desserts, daily specials, coffee, soft drinks, smoothies, and a limited selection of beer, wine, and cocktails. Services include dine-in, takeaway, online pickup, basic catering, and private gatherings, with an emphasis on consistent quality, family-friendly atmosphere, reasonable pricing, and strong commitment to diversity, inclusion, and sustainability. Initial staffing comprises a small diverse team (manager, head chef, cooks, servers, and support roles), supported by standard commercial kitchen equipment and basic technology systems, with hypothetical startup costs around $415,000 AUD and projected first-year revenue of $750,000."

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Daily finance snapshot for decisions (Priority: P1)

The owner/manager reviews a daily finance snapshot to confirm revenue, labor %, COGS %, cash position, and channel/daypart mix, and decides whether to adjust staffing or specials.

**Why this priority**: Enables same-day actions on cost control and staffing for a single-site cafe with long hours.

**Independent Test**: Upload latest data, load dashboard, verify KPIs and freshness indicator; decisions (e.g., reduce staffing next shift) can be made from this view alone.

**Acceptance Scenarios**:

1. **Given** the latest POS, payroll, and inventory CSVs are imported, **When** the manager opens the dashboard, **Then** revenue, COGS %, labor %, operating expense %, and net profit for the selected day display with a data freshness timestamp.
2. **Given** multiple service channels (dine-in, takeaway, pickup, catering), **When** the manager views channel/daypart breakdowns, **Then** each shows revenue, covers, and average check with percentages of total.

---

### User Story 2 - Import and reconcile source data (Priority: P1)

The accountant uploads POS sales, payroll, and basic inventory/recipe cost CSVs, maps columns once, and confirms import completion with anomaly flags for missing or inconsistent records.

**Why this priority**: Reliable data ingestion is foundational for any insights.

**Independent Test**: Run imports with sample files, verify mapping, see success + anomalies report without needing other features.

**Acceptance Scenarios**:

1. **Given** a POS CSV with sales, discounts, tenders, and service channel indicators, **When** the accountant uploads and applies saved mappings, **Then** the system ingests all rows, reports row counts, and surfaces any anomalies (e.g., negative totals, missing channel) for review.
2. **Given** payroll and inventory cost CSVs, **When** they are uploaded, **Then** labor cost totals and item cost updates are applied to the next KPI refresh and marked in the import log.

---

### User Story 3 - Drill-down and exports (Priority: P2)

The owner drills from KPIs to supporting transactions and exports summarized P&L and channel performance to share with stakeholders or external accounting systems.

**Why this priority**: Transparency builds trust and supports handoff to accounting/tax tools.

**Independent Test**: From an existing dashboard, open drill-down, filter, and export without needing further development.

**Acceptance Scenarios**:

1. **Given** KPI cards show revenue and labor %, **When** the owner clicks drill-down, **Then** itemized sales (with channel/daypart) and labor summaries are displayed with filters for date range and channel.
2. **Given** the owner selects a period, **When** they export P&L and channel summary, **Then** a CSV is generated with the same totals as displayed and a freshness timestamp.

---

### Edge Cases

- Imports with missing or malformed columns should fail fast with a clear mapping error, not partial silent loads.
- Duplicate file uploads should be idempotent, replacing or ignoring duplicates with a clear message.
- Timezone handling should default to Brisbane local time and display it on reports.
- If a source (e.g., payroll) is absent for a period, KPIs should show "missing data" indicators rather than zeros.
- Large single-day promos (high discounts/comps) should flag outliers in reports.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: System MUST ingest POS sales CSVs with tenders, discounts, comps, channel/daypart markers, and map columns via a saved mapping profile.
- **FR-002**: System MUST ingest payroll summary CSVs and associate labor cost to matching dates and dayparts.
- **FR-003**: System MUST ingest inventory/recipe cost CSVs to update item-level COGS used in margin calculations.
- **FR-004**: System MUST present a daily dashboard with revenue, COGS %, labor %, operating expense %, net profit, cash balance, and data freshness.
- **FR-005**: System MUST break down KPIs by service channel (dine-in, takeaway, pickup, catering) and daypart (breakfast, lunch, dinner).
- **FR-006**: System MUST provide drill-down views to underlying transactions (sales lines, labor summaries) with filters for date range, channel, and daypart.
- **FR-007**: System MUST export summarized P&L and channel performance to CSV with matching totals and freshness timestamp.
- **FR-008**: System MUST maintain an import log with status, row counts, anomalies, and who performed the import.
- **FR-009**: System MUST enforce role-based access (Owner/Admin, Manager, Accountant, Viewer) with least privilege defaults.
- **FR-010**: System MUST display Brisbane local timezones on time-based reports and imports.

### Key Entities

- **Location**: Single venue with indoor and patio seating; attributes include name, timezone, seating capacity.
- **Service Channel**: Dine-in, takeaway, pickup, catering; attributes include name, indicators used in imports.
- **Daypart**: Breakfast, lunch, dinner; attributes include time windows used for grouping.
- **Menu Item**: Item name, category, recipe cost, pricing; links to sales lines.
- **Sale**: Transaction with date/time, channel, daypart, items, discounts, tenders, totals.
- **Payroll Period**: Date range, role/category, total labor cost, hours.
- **Inventory Snapshot**: Date, item cost data used for COGS.
- **Import Job**: Source type, mapping profile, status, row counts, anomalies, user, timestamp.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: Daily dashboard loads primary KPIs in under 1 second for 30 days of data and under 3 seconds for trailing 12 months.
- **SC-002**: 100% of imported rows are either ingested or reported as anomalies; no silent drops.
- **SC-003**: At least 90% of imports complete with zero anomalies using saved mapping profiles on second run.
- **SC-004**: Managers can generate and download P&L and channel exports within 10 seconds for a selected month, matching on-screen totals.
- **SC-005**: 90% of first-time manager users report they can find revenue, labor %, and COGS % within 30 seconds of opening the dashboard (usability check).

## Assumptions & Dependencies

- Assumed single venue in Brisbane (AEST/AEDT); currency AUD; business hours as provided; public holidays treated as normal days unless configured.
- Assumed POS, payroll, and inventory systems can export CSVs with stable column headers; mapping profiles persist across uploads.
- Assumed dayparts map to breakfast/lunch/dinner windows and can be overridden in configuration without schema changes.
- Dependency: Basic role directory exists (app-native roles Owner/Admin, Manager, Accountant, Viewer) to enforce access.
- Dependency: Secure storage for uploaded files and processed outputs; audit trail retained per import job.
