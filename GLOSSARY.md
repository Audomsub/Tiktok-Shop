# TikTok Affiliate Winning Product Analytics

Domain vocabulary for identifying, ingesting, scoring, and tracking viral TikTok affiliate products.

## Language

**Product**:
A distinct TikTok Shop item tracked across scraping cycles, uniquely identified by its source marketplace ID.
_Avoid_: Item, listing, merchandise

**Snapshot**:
An immutable point-in-time capture of a Product's price, commission rate, total sales, velocity, and winning score during a specific Crawl Round.
_Avoid_: Record, historical state, log entry

**Baseline Snapshot**:
The first snapshot recorded for a newly discovered product, with sales delta and velocity initialized to zero.
_Avoid_: Initial snapshot, genesis record, cold start

**Lookback Window**:
The maximum allowable time window (up to 24 hours) to pair consecutive snapshots for calculating hourly velocity.
_Avoid_: History window, comparison range, lookback period

**Crawl Round**:
A scheduled ingestion execution window (06:00, 12:00, 18:00, or 00:00) with a defined target depth and audit trail.
_Avoid_: Batch, run, scrape session

**Partial Crawl**:
An ingestion execution interrupted by external rate limits or network errors where only a subset of pages succeeded.
_Avoid_: Interrupted crawl, failed run, half batch

**Sales Delta**:
The net increase in a Product's cumulative sales count between consecutive snapshots.
_Avoid_: Sales diff, sales growth, period sales

**Sales Velocity**:
The hourly sales rate of a Product computed as Sales Delta divided by elapsed hours between snapshots.
_Avoid_: Hourly rate, sales speed, run rate

**Winning Score**:
A normalized composite score from 0 to 100 ranking a Product by Sales Velocity, Commission Rate, and Expected Return.
_Avoid_: Product rank, popularity index, virality score

**Expected Return**:
The projected earnings in Thai Baht generated per unit sold, calculated as Price multiplied by Commission Rate.
_Avoid_: Commission payout, profit per item, unit commission

**Winning Badge**:
A standardized classification tag (Viral Surge, High Commission, High Yield, Winning Top Pick) assigned based on analytical criteria.
_Avoid_: Product label, status tag, marketing badge

**Pre-Filtering**:
The rejection threshold enforced at ingestion to drop noise before persistence (commission < 10%, price outside 80–1,500 THB, or total sales < 30).
_Avoid_: Ingestion filter, data gating, validation rule
