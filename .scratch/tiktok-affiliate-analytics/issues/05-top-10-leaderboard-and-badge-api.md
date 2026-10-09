## Parent

#1

## What to build

A public Go REST API endpoint `GET /api/v1/leaderboard` that retrieves the Top 10 products with the highest winning scores from the most recent completed crawl round, enriched with product details, calculated velocity, commission yield, and dynamic badge flags (🔥 Viral Surge, 💎 High Commission, 💰 High Yield, 🏆 Winning Top Pick).

## Acceptance criteria

- [x] Endpoint `GET /api/v1/leaderboard` is publicly accessible without requiring an authentication token.
- [x] Returns exactly Top 10 ranked products based on `winning_score DESC` of the latest successful crawl round.
- [x] Each item includes product name, source URL, image URL, category, current price, total sales, hourly velocity, commission rate, and winning score.
- [x] Computes badge flags:
  - `is_viral_surge`: Velocity $\ge 10$ pcs/hr or top 15% in batch.
  - `is_high_commission`: Commission Rate $\ge 20\%$.
  - `is_high_yield`: Expected Return $\ge 100$ THB.
  - `is_winning_pick`: Winning Score $\ge 80$.
- [x] Integration test verifies response payload schema and latency under 100ms using index `idx_snapshots_winning`.

## Blocked by

- #5: Ticket 04: Worker Pool Bulk SQL Updater & Ingestion Webhook
