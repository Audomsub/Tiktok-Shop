## Parent

#1

## What to build

A Go REST API endpoint `GET /api/v1/products/:id/trends?days=7` that queries the `product_snapshots` table utilizing index `idx_snapshots_trends` to retrieve chronological time-series data points (snapshot timestamp, total sales, hourly velocity, delta sales, price, and winning score) over the preceding 7-day period.

## Acceptance criteria

- [ ] Endpoint `GET /api/v1/products/:id/trends` validates product UUID and optional `days` query param (default: 7, max: 30).
- [ ] Returns chronological list of snapshots ordered by `snapshot_time ASC`.
- [ ] Returns 404 Not Found if product UUID does not exist.
- [ ] Query executes efficiently using `idx_snapshots_trends ON product_snapshots(product_id, snapshot_time ASC)`.
- [ ] Unit/Integration tests verify response schema and correct chronological ordering of points.

## Blocked by

- #8: Ticket 07: Product Catalog & Dynamic Category Search API
