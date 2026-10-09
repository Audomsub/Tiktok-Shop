# n8n FastMoss Scraper & Analytics Ingestion Pipeline

This directory contains the production-ready **n8n automation service** and workflow templates for the TikTok Affiliate Winning Product Analytics Platform.

---

## 📋 Overview

The ingestion pipeline automates continuous and deep product crawling against FastMoss using advanced anti-ban rate limiting, session cookie protection, multi-attribute Quality Gate Pre-Filtering, idempotent Supabase persistence, and an automated webhook handshake to the Go Analytics Engine.

### Key Capabilities
- **Dual Schedule Cadence (Asia/Bangkok UTC+7)**:
  - **Shallow Crawl**: Triggered at `06:00`, `12:00`, and `18:00` daily (`0 6,12,18 * * *`) — crawls top 20 pages (~400 products).
  - **Deep Crawl**: Triggered at `00:00 Midnight` daily (`0 0 * * *`) — crawls top 80 pages (~1,600 products).
- **Anti-Ban Jitter Delay**:
  - Injects randomized delays between **3.0s and 7.0s** before each page transition.
- **Cool-Off Rate Limiter**:
  - Enforces a mandatory **25-second cool-off pause** after every 10 fetched pages (before fetching pages 11, 21, 31, etc.).
- **Session Protection & Anti-Ban Detector**:
  - Intercepts HTTP `403 Forbidden` and `429 Too Many Requests`.
  - Immediately halts workflow execution via `stopAndError` node to protect the FastMoss account session.
  - Updates Supabase `crawl_logs` with HTTP error code and marks round as `PARTIAL` or `FAILED`.
  - Sends immediate critical alert webhook to Discord / Slack.
- **Pre-Filtering Quality Gate**:
  - Drops low-margin and dead-stock products before database persistence:
    - **Commission Rate** $\ge 10\%$
    - **Price Range** between $80$ and $1,500$ THB
    - **Total Cumulative Sales** $\ge 30$ units
- **Idempotent Supabase Multi-Entity Upsert**:
  - Auto-upserts taxonomy into `categories` using `ON CONFLICT (slug)`.
  - Auto-upserts product metadata into `products` using `ON CONFLICT (source_id)`.
  - Inserts immutable time-series snapshots into `product_snapshots` using `ON CONFLICT (product_id, crawl_log_id)`.
- **Transactional Audit Trail**:
  - Updates `crawl_logs` with accurate metrics: `total_pages_requested`, `total_pages_success`, `raw_products_scraped`, and `filtered_products_saved`.
- **Go Analytics Engine Handshake**:
  - Sends authenticated HTTP POST webhook to Go backend `POST /api/v1/jobs/compute-scores?crawl_log_id=...` with `X-API-Key` to immediately compute velocity and 0–100 winning scores.

---

## 🚀 Quick Start & Deployment

### 1. Configure Environment Variables
Copy `.env.example` to `.env` and fill in your credentials:
```bash
cp n8n/.env.example n8n/.env
```

| Variable | Description | Example / Default |
|---|---|---|
| `FASTMOSS_COOKIE` | Authenticated FastMoss cookie header | `session_id=...; token=...` |
| `FASTMOSS_API_BASE_URL` | FastMoss API Base URL | `https://api.fastmoss.com` |
| `SUPABASE_URL` | Supabase project API URL | `https://evyltfcspqzctgdehxff.supabase.co` |
| `SUPABASE_SERVICE_ROLE_KEY` | Supabase service role key (bypasses RLS) | `eyJhbGci...` |
| `GO_BACKEND_URL` | Go backend analytics engine host | `http://host.docker.internal:8080` |
| `INTERNAL_API_KEY` | Shared secret for backend ingestion webhook | `dev_internal_api_key_secret_2026` |
| `ALERT_WEBHOOK_URL` | Emergency alert webhook (Discord/Slack) | `https://discord.com/api/webhooks/...` |

### 2. Start n8n with Docker Compose
Run the following from the project root:
```bash
docker compose -f n8n/docker-compose.yml up -d
```
Access the n8n UI at **http://localhost:5678**.

### 3. Import Workflow
1. In the n8n web dashboard, navigate to **Workflows** -> **Import from File**.
2. Select `n8n/workflows/fastmoss_scraper_workflow.json`.
3. Activate the workflow to enable the automated schedule triggers.

---

## 🧪 Automated Testing & Validation

The entire pipeline (triggers, anti-ban rate limiting, 403/429 abort handler, pre-filtering quality gate, Supabase `ON CONFLICT` queries, and Go webhook handshake) is verified programmatically:

### Standalone Node.js Validator
```bash
node n8n/tests/validate_workflow.mjs
```

### Go Test Suite Integration
```bash
cd backend
go test -v -run TestFastMossScraperWorkflow ./internal/usecase/...
```

---

## 🔄 End-to-End Pipeline Diagram

```
[Schedule Trigger 06:00, 12:00, 18:00] ---\
[Schedule Trigger 00:00 Midnight] --------+--> [Initialize Crawl Scope]
[Manual / Test Trigger] ------------------/           |
                                                      v
                                        [Supabase: Init crawl_logs (RUNNING)]
                                                      |
                                                      v
                                        [Generate Page Tasks Queue (1..N)]
                                                      |
                                                      v
                                      +-----> [Page Processing Loop]
                                      |               | (next page)
                                      |               v
                                      |     [Calculate Jitter & Cool-off]
                                      |               |
                                      |               v
                                      |     [Wait Node (3-7s + 25s pause)]
                                      |               |
                                      |               v
                                      |     [HTTP: Fetch FastMoss Page]
                                      |               |
                                      |               v
                                      |     [Inspect Response & Status]
                                      |               |
                                      |      /-----------------\
                         (HTTP 200)   |     /                   \  (HTTP 403 / 429)
                                      |    v                     v
                        [Collect Page Data]       [Supabase: Update crawl_logs FAILED/PARTIAL]
                                      |                          |
                                      |                          v
                                      |              [Emergency Alert Webhook]
                                      |                          |
                                      |                          v
                                      |              [Stop & Terminate (Abort)]
                                      |
                     (all pages processed)
                              v
             [Aggregate Raw Products across pages]
                              |
                              v
             [Pre-Filter Quality Gate & Taxonomy Extraction]
             (Commission >= 10%, Price 80-1,500 THB, Sales >= 30)
                              |
                     /-----------------\
      (Items Passed) v                 v (0 Items Passed)
  [Supabase: Upsert Categories]        |
              |                        |
              v                        |
  [Map Category IDs to Products]       |
              |                        |
              v                        |
  [Supabase: Upsert Products]          |
              |                        |
              v                        |
  [Prepare Snapshots Payload]          |
              |                        |
              v                        |
  [Supabase: Insert Snapshots]         |
              |                        |
              \------------------------/
                              |
                              v
             [Format Audit Counts Update]
                              |
                              v
             [Supabase: Update crawl_logs Audit Counts]
             (total_pages, raw_scraped, filtered_saved)
                              |
                              v
             [HTTP POST Webhook: Trigger Go Analytics Engine]
             (POST /api/v1/jobs/compute-scores + X-API-Key)
                              |
                              v
             [Workflow Execution Summary (SUCCESS)]
```
