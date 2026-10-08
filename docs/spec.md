# System Specification: TikTok Affiliate Winning Product Analytics Platform

## Problem Statement

TikTok affiliate creators and digital marketers struggle to identify high-potential "winning products" in real time. Manually browsing marketplace analytics like FastMoss is slow, noisy, and overwhelming with thousands of low-margin or slow-moving items. Furthermore, raw metrics like total sales can be deceptive, as legacy products with high cumulative sales may currently be stagnant, while explosive viral products with rapid hourly momentum and lucrative commission rates get buried. Without automated ingestion, velocity tracking, normalized composite scoring, and an interactive dashboard, creators miss viral product waves, waste marketing budget on dead stock, and lose out on affiliate commissions.

## Solution

An end-to-end analytics platform that automates the discovery, calculation, and visualization of viral TikTok Shop affiliate products:
1. **Automated Hybrid Ingestion (n8n)**: Periodically scrapes FastMoss with dynamic scheduling (shallow daytime crawls for viral momentum, deep midnight crawls for daily archives), rate control, and early rejection gates (Pre-Filtering).
2. **Persistent Time-Series Storage (Supabase PostgreSQL)**: Maintains master Product catalogs, immutable point-in-time Snapshots, audit Crawl Logs, and user bookmarks with Row Level Security (RLS).
3. **High-Performance Analytics Engine (Go)**: Triggered via webhook to compute Sales Delta, hourly Sales Velocity, and a 0–100 Winning Score via batch Min-Max Normalization across velocity, commission rate, and expected return, persisting updates within 5 seconds.
4. **Interactive Web Dashboard (Next.js)**: Communicates exclusively via the Go REST API to deliver a real-time Winning Leaderboard, rich multi-criteria filters, 7-day historical trend charts, user bookmarking, and CSV exports.

## User Stories

1. As an affiliate creator, I want to view a real-time Top 10 Winning Leaderboard of products with the highest winning scores, so that I can immediately identify the most lucrative products to promote today.
2. As an affiliate creator, I want each winning product on the leaderboard to display standardized status badges (🔥 Viral Surge, 💎 High Commission, 💰 High Yield, 🏆 Winning Top Pick), so that I can quickly assess why a product is trending.
3. As an affiliate creator, I want to filter products by category, so that I only see products relevant to my specific TikTok content niche.
4. As an affiliate creator, I want to filter products by price range using an interactive slider (e.g. 80–1,500 THB), so that I can select products that fit my audience's purchasing power.
5. As an affiliate creator, I want to set a minimum commission rate threshold, so that I only spend time promoting products that offer acceptable profit margins.
6. As an affiliate creator, I want to search products by keyword or title, so that I can check whether a specific brand or item is gaining traction.
7. As an affiliate creator, I want to click on any product to open a 7-day historical growth line chart, so that I can see whether its sales velocity is accelerating, plateauing, or declining before investing time in filming video content.
8. As an affiliate creator, I want to bookmark (favorite) high-potential products, so that I have a curated shortlist of products ready for upcoming video production.
9. As an affiliate creator, I want to view and manage my bookmarked products in a dedicated Favorites tab, so that I can review and remove items once videos are completed.
10. As an affiliate creator, I want to export filtered product lists into a CSV file, so that I can perform offline analysis or share candidates with my video production team.
11. As an affiliate creator, I want product cards and tables to display direct links to the TikTok Shop / source marketplace page, so that I can easily verify product availability and apply for sample units.
12. As a platform administrator, I want n8n to execute lightweight shallow crawls (pages 1–20, ~200 items) at 06:00, 12:00, and 18:00, so that midday viral surges are captured rapidly without unnecessary proxy bandwidth.
13. As a platform administrator, I want n8n to execute deep crawls (pages 1–80, ~800 items) at 00:00 midnight, so that the platform archives comprehensive daily market history.
14. As a platform administrator, I want the scraper to inject random jitter delays (3–7 seconds) per page and cool-off pauses (25 seconds every 10 pages), so that the scraping session avoids automated bot detection and IP bans.
15. As a platform administrator, I want ingestion to discard noise immediately via Pre-Filtering (commission < 10%, price outside 80–1,500 THB, or total sales < 30 units), so that the database is not polluted with low-value records.
16. As a platform administrator, I want every scraping run to initialize an audit record in `crawl_logs` with status `RUNNING`, updating to `SUCCESS`, `PARTIAL`, or `FAILED` upon completion, so that system operational health is fully transparent.
17. As a platform administrator, I want the scraper to immediately halt and alert when encountering HTTP 403 or 429 status codes, so that external marketplace accounts are protected from permanent bans.
18. As a system operator, I want n8n to send an authenticated webhook to the Go Analytics Engine upon finishing a crawl round, so that score computation begins immediately without wasteful database polling.
19. As an analytics engineer, I want the Go engine to compute Sales Delta and hourly Sales Velocity by comparing current snapshots with the immediate previous snapshot within a 24-hour Lookback Window, so that velocity accurately reflects short-term momentum.
20. As an analytics engineer, I want products discovered for the first time to be treated as a Baseline Snapshot with delta and velocity set to zero, so that large historical total sales do not cause artificial spikes.
21. As an analytics engineer, I want the Go engine to apply batch Min-Max Normalization with 99th percentile clipping to Sales Velocity, Commission Rate, and Expected Return, so that the composite 50/30/20 formula reliably scales from 0 to 100 points.
22. As an analytics engineer, I want the Go engine to update snapshot computations back to Supabase using a worker pool and chunked bulk SQL updates within 5 seconds, so that data is immediately available to users without database connection exhaustion.
23. As an authenticated user, I want my bookmarked products to be protected by Row Level Security (RLS) and verified by JWT in the Go API, so that other users cannot view or manipulate my personal bookmarks.
24. As an unauthenticated visitor, I want to view public leaderboard rankings and search product catalogs without logging in, so that I can experience platform value before signing up.

## Implementation Decisions

### 1. Ingestion Layer (n8n Workflow Engine)
- Configured in Docker with schedule triggers at 06:00, 12:00, 18:00 (pages 1–20) and 00:00 (pages 1–80).
- Applies random delays (3–7s) and 25s cool-off every 10 pages.
- Code Node enforces Pre-Filtering criteria before database emission.
- On HTTP 403/429, execution terminates immediately, logs error code, marks `PARTIAL` (if pages > 0) or `FAILED`, and sends alert.
- Triggers Go backend via HTTP POST webhook with `crawl_log_id` and internal API secret upon completion.

### 2. Persistence Layer (Supabase PostgreSQL)
- **Entities**: `categories`, `products`, `product_snapshots`, `crawl_logs`, `favorite_products`.
- **Idempotency**: Composite unique constraint `UNIQUE (product_id, crawl_log_id)` prevents duplicate records on retry or pagination overlap.
- **Snapshot Immutability**: `product_snapshots` stores `price` and `commission_rate NUMERIC(5, 2)` at snapshot time, ensuring historical calculations remain accurate even if sellers change terms later.
- **Foreign Key & Query Indexing**:
  - `idx_snapshots_crawl_log` on `product_snapshots(crawl_log_id)` for batch analytics lookup.
  - `idx_snapshots_trends` on `product_snapshots(product_id, snapshot_time ASC)` for 7-day chart rendering.
  - `idx_favorite_product` on `favorite_products(product_id)` for efficient cascade deletes.
- **Security (RLS)**:
  - Public read on `categories`, `products`, `product_snapshots`, and `crawl_logs`.
  - User-isolated CRUD on `favorite_products` using cached subquery `((SELECT auth.uid()) = user_id)` restricted `TO authenticated`.

### 3. Analytics Engine (Go Service)
- Built with `go-chi/chi` router and `jackc/pgx/v5` connection pool.
- **Scoring Pipeline**:
  - Pulls current round snapshots via `crawl_log_id`.
  - Queries preceding snapshots per product within 24-hour Lookback Window.
  - Sets Baseline Snapshot (`delta_sales = 0`, `velocity_per_hour = 0`) if no prior snapshot exists or if prior snapshot is > 24 hours old.
  - Computes $\Delta\text{Sales} = \text{TotalSales}_{\text{current}} - \text{TotalSales}_{\text{previous}}$ and $\text{Velocity} = \Delta\text{Sales} / \Delta\text{Hours}$.
  - Computes Expected Return = $\text{Price} \times (\text{Commission Rate} / 100)$.
  - Applies batch Min-Max Normalization (0–100) with 99th percentile clipping for Velocity, Commission Rate, and Expected Return.
  - Computes Winning Score:
    $$\text{Winning Score} = (0.50 \times \text{NormVelocity}) + (0.30 \times \text{NormCommission}) + (0.20 \times \text{NormExpectedReturn})$$
  - Dispatches bulk updates via Worker Pool (5–10 workers) in chunks of 100–200 items using `UPDATE product_snapshots ... FROM (VALUES ...)` within SLA (< 5s).

### 4. API Gateway & Endpoints Contract (Go Backend)
- Next.js calls Go REST API exclusively; no direct Supabase client calls from the frontend.
- Go verifies Supabase JWT (`Authorization: Bearer <token>`) using project JWT Secret (`HS256`).
- **REST Endpoints**:
  - `POST /api/v1/jobs/compute-scores` — Webhook for n8n with internal API key.
  - `GET /api/v1/leaderboard` — Top 10 products of latest round with winning score and badge flags.
  - `GET /api/v1/products` — Paginated catalog with filters (`category_id`, `min_price`, `max_price`, `min_commission`, `q`, `page`, `limit`, `sort_by`).
  - `GET /api/v1/products/:id/trends` — 7-day time-series snapshot data.
  - `GET /api/v1/categories` — Category list for dropdowns.
  - `GET /api/v1/favorites` — Current user's bookmarked products (requires auth).
  - `POST /api/v1/favorites` — Add product to bookmarks (requires auth).
  - `DELETE /api/v1/favorites/:productId` — Remove product from bookmarks (requires auth).
  - `GET /api/v1/products/export` — Stream CSV of filtered products.

### 5. Presentation Layer (Next.js App Router)
- Built with Next.js 14/15, TypeScript, Tailwind CSS, `shadcn/ui`, `recharts`, and `lucide-react`.
- Renders Winning Leaderboard cards with badges:
  - 🔥 **Viral Surge**: Velocity $\ge 10$ pcs/hr or top 15% in batch.
  - 💎 **High Commission**: Commission Rate $\ge 20\%$.
  - 💰 **High Yield**: Expected Return $\ge 100$ THB.
  - 🏆 **Winning Top Pick**: Winning Score $\ge 80$.
- Interactive catalog table with debounced search, range sliders, and multi-select filters.
- Modal or detail view featuring historical velocity/sales line chart over 7 days.
- User bookmark toggle with optimistic UI updates.

## Testing Decisions

### Testing Seam
The primary testing seam across the system is the **Go HTTP API & Webhook Boundary**:
- Tests interact with the Go service via HTTP requests (`net/http/httptest`) executing against an integration PostgreSQL database (or test database container).
- Verifies full pipeline behavior: HTTP request parsing, JWT authentication, query filtering, scoring algorithms, and database batch updates.
- Avoids testing internal private helper functions or implementation details; tests validate strictly external observable behavior.

### Modules Tested
1. **Analytics Engine Batch Scoring**:
   - Verify calculation of $\Delta\text{Sales}$ and hourly velocity.
   - Verify baseline handling for brand-new products.
   - Verify 24-hour lookback cutoff behavior.
   - Verify Min-Max normalization bounds and 0–100 score distribution.
   - Verify batch update idempotency and database write completion within SLA.
2. **API Handlers & Middleware**:
   - Verify public access to Leaderboard, Catalog, and Trends without token.
   - Verify rejected access or 401 Unauthorized on Favorites without valid JWT.
   - Verify authorized CRUD on Favorites with valid Supabase JWT.
   - Verify CSV export formatting and filter compliance.
3. **Frontend Integration Seam**:
   - Verify Next.js components render leaderboard cards, apply filter queries, and display charts against mock API responses.

## Out of Scope
- Automatic video script generation or AI content creation for TikTok videos.
- Direct automated placement of affiliate orders or purchasing of physical samples.
- Direct TikTok Shop Creator OAuth account linking (data sourced via FastMoss marketplace scraping).
- Custom multi-tenant billing or subscription paywalls (initial platform focuses on free/internal affiliate analytics).

## Further Notes
- Automated scrapers must respect cool-off pauses and anti-ban limits to maintain persistent FastMoss session cookie validity.
- Database credentials and Supabase service role keys are managed strictly via environment variables and never checked into source control.
- Git workflow adheres strictly to `main` (Production), `uat` (Staging/Test), and `feature/<name>` branching model with commits per completed function.
