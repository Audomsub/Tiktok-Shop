## Parent

#1

## What to build

An interactive product detail modal / sheet on Next.js powered by Recharts that visualizes a product's 7-day performance history with a dual-axis line chart (cumulative total sales on left axis vs hourly sales velocity on right axis), formatted timestamps, and informative tooltips when clicking on any product in the catalog or leaderboard.

## Acceptance criteria

- [x] Clicking any product card or table row opens the Product Trend modal dialog.
- [x] Fetches time-series data from Go API `GET /api/v1/products/:id/trends?days=7`.
- [x] Recharts ResponsiveContainer renders dual-axis lines (Sales curve and Hourly Velocity curve).
- [x] Displays summary metrics at the top: current velocity, 7-day net sales growth, current price, and commission payout.
- [x] Includes loading skeleton state and graceful error handling for products with single baseline snapshots.

## Blocked by

- #9: Ticket 08: Filterable Product Table & Range Slider Controls UI
- #10: Ticket 09: 7-Day Historical Trend Data API
