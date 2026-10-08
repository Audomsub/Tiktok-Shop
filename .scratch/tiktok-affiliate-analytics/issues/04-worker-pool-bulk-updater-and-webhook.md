## Parent

#1

## What to build

A high-throughput Worker Pool in Go that buffers and executes chunked bulk SQL updates (`UPDATE ... FROM (VALUES ...)`) to persist computed delta sales, velocity, and winning scores back to Supabase within a 5-second SLA, exposed via an authenticated webhook endpoint `POST /api/v1/jobs/compute-scores?crawl_log_id=...` that updates `crawl_logs` status to `SUCCESS` or `FAILED`.

## Acceptance criteria

- [x] Worker pool of 5–10 goroutines processes scored snapshot records in batches of 100–200 items.
- [x] Uses chunked bulk SQL update (`UPDATE product_snapshots AS s SET delta_sales = u.delta_sales, velocity_per_hour = u.velocity_per_hour, winning_score = u.winning_score FROM (VALUES ...) AS u(id, delta_sales, velocity_per_hour, winning_score) WHERE s.id = u.id`) avoiding individual per-row update statements.
- [x] Successfully persists up to 800 snapshots back to the database in under 5 seconds.
- [x] Endpoint `POST /api/v1/jobs/compute-scores` requires internal API key header (`X-API-Key`).
- [x] Upon completion, `crawl_logs.status` is updated to `SUCCESS` (or `FAILED` with error message if execution fails).

## Blocked by

- #4: Ticket 03: Batch Min-Max Winning Score Normalization Engine
