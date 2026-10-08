## Parent

#1

## What to build

A core analytics calculation service in Go that takes a batch of raw snapshots for a crawl round, retrieves preceding snapshots per product within a 24-hour Lookback Window, identifies newly discovered products as Baseline Snapshots (setting Delta Sales = 0 and Sales Velocity = 0), and accurately computes raw Delta Sales and hourly Sales Velocity for recurring products.

## Acceptance criteria

- [ ] Query logic retrieves immediate preceding snapshots per product within a 24-hour Lookback Window.
- [ ] Products appearing for the first time receive baseline values (`delta_sales = 0`, `velocity_per_hour = 0`).
- [ ] Products whose previous snapshot is older than 24 hours are treated as fresh baselines to prevent diluted hourly velocity.
- [ ] Delta Sales ($\Delta\text{Sales} = \text{TotalSales}_{\text{current}} - \text{TotalSales}_{\text{previous}}$) and hourly Sales Velocity ($\text{Velocity} = \Delta\text{Sales} / \Delta\text{Hours}$) are accurately calculated.
- [ ] Unit tests verify all calculation edge cases (cold start, zero sales change, negative corrections, missing prior history).

## Blocked by

- #2: Ticket 01: Go Backend Foundation, Database Pool & Healthcheck
