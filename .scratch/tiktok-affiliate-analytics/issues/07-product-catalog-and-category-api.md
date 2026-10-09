## Parent

#1

## What to build

Public Go REST API endpoints `GET /api/v1/categories` and `GET /api/v1/products` providing a rich, paginated product catalog search engine supporting keyword search (`q`), category filtering (`category_id`), price range sliders (`min_price`, `max_price`), minimum commission threshold (`min_commission`), and multi-column sorting (by score, velocity, total sales, commission rate, or price).

## Acceptance criteria

- [x] Endpoint `GET /api/v1/categories` returns all active product categories ordered by name.
- [x] Endpoint `GET /api/v1/products` supports query parameters: `q`, `category_id`, `min_price`, `max_price`, `min_commission`, `sort_by`, `sort_order`, `page`, and `limit`.
- [x] Title search (`q`) executes case-insensitive ILIKE keyword matching.
- [x] Returns paginated response with total counts, total pages, current page, and product records paired with their latest snapshot metrics.
- [x] SQL query utilizes `idx_products_category`, `idx_snapshots_lookup`, and `idx_snapshots_winning` for fast response (< 150ms).

## Blocked by

- #6: Ticket 05: Top 10 Winning Leaderboard & Badge API
