# n8n FastMoss Scraper & Rate Limiting Engine

This directory contains the production-ready **n8n automation service** and workflow templates for the TikTok Affiliate Winning Product Analytics Platform.

---

## 📋 Overview

The scraper workflow automates continuous and deep product crawling against FastMoss using advanced anti-ban rate limiting, session cookie protection, and transactional audit logging with Supabase.

### Key Capabilities
- **Dual Schedule Cadence (Asia/Bangkok UTC+7)**:
  - **Shallow Crawl**: Triggered at `06:00`, `12:00`, and `18:00` daily (`0 6,12,18 * * *`) — crawls top 20 pages (400 products).
  - **Deep Crawl**: Triggered at `00:00 Midnight` daily (`0 0 * * *`) — crawls top 80 pages (1,600 products).
- **Anti-Ban Jitter Delay**:
  - Injects randomized delays between **3.0s and 7.0s** before each page transition.
- **Cool-Off Rate Limiter**:
  - Enforces a mandatory **25-second cool-off pause** after every 10 fetched pages (before fetching pages 11, 21, 31, etc.).
- **Session Protection & Anti-Ban Detector**:
  - Intercepts HTTP `403 Forbidden` and `429 Too Many Requests`.
  - Immediately halts workflow execution via `stopAndError` node to protect the FastMoss account session.
  - Updates Supabase `crawl_logs` with HTTP error code and marks round as `PARTIAL` or `FAILED`.
  - Sends immediate critical alert webhook to Discord / Slack.

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

The workflow structure, node connectivity, anti-ban algorithms, and schedule triggers are verified automatically:

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

## 🔄 Workflow Diagram & Flow

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
             [Aggregate Raw Products (Ready for Quality Gate)]
```
