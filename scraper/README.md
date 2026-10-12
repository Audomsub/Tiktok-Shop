# FastMoss Playwright Scraper Sidecar Service

Automated headless browser ingestion sidecar for FastMoss TikTok Shop product analytics.

## Features
- **Dynamic Signature Bypass**: Automatically solves and evaluates client-side dynamic `fm-sign` and `fm-sig` signatures required by FastMoss WAF.
- **Warm Session Engine**: Keeps an authenticated Chromium browser instance alive in memory with session cookies (`fd_tk`), enabling sub-second API fetches (~400ms).
- **Internal Microservice**: Runs seamlessly on Docker network `app_network` on port `3001`, consumed directly by n8n or any backend worker.

## Endpoints
- `GET /health` — Check browser session readiness
- `GET /api/v1/fastmoss/search?page=1&pageSize=20&region=TH&order=2,2` — Scrapes live FastMoss products
- `POST /api/v1/fastmoss/search` — Scrapes with JSON body parameters

## Environment Variables
- `PORT` (default: `3001`)
- `FASTMOSS_COOKIE` — Session cookies including `fd_tk`
