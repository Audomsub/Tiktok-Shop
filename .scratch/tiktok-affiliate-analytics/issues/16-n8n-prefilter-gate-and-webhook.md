## Parent

#1

## What to build

The downstream stages of the n8n ingestion workflow including a Code Node implementing the Pre-Filtering quality gate (dropping items with commission < 10%, price outside 80–1,500 THB, or total sales < 30), database upsert into `products` and `product_snapshots`, audit trail logging in `crawl_logs`, and an authenticated HTTP POST webhook call to trigger the Go Analytics Engine.

## Acceptance criteria

- [ ] Code Node rejects items not meeting quality criteria (Commission $\ge 10\%$, Price 80–1,500 THB, Total Sales $\ge 30$).
- [ ] Auto-upserts categories and products into Supabase using `ON CONFLICT` clauses.
- [ ] Inserts snapshots using `crawl_log_id` and idempotent composite key.
- [ ] Updates `crawl_logs` with counts: `total_pages_requested`, `total_pages_success`, `raw_products_scraped`, and `filtered_products_saved`.
- [ ] Sends HTTP POST webhook to Go backend `POST /api/v1/jobs/compute-scores?crawl_log_id=...` with `X-API-Key` upon completion.

## Blocked by

- #16: Ticket 15: n8n Scraper Workflow Template & Anti-Ban Rate Limiter
