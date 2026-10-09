## Parent

#1

## What to build

A production-ready n8n workflow template JSON for Docker that automates scheduled scraping against FastMoss with dynamic scopes (06:00, 12:00, 18:00 top 20 pages; 00:00 top 80 pages), randomized jitter delays (3–7 seconds) per page, cool-off pauses (25 seconds every 10 pages), and instant halting with error logging upon receiving HTTP 403 or 429 status codes.

## Acceptance criteria

- [ ] Configures Schedule Triggers: 06:00, 12:00, 18:00 (pages 1–20) and 00:00 (pages 1–80).
- [ ] Injects authenticated FastMoss cookie session in HTTP Request node.
- [ ] Applies randomized Jitter Delay between 3 and 7 seconds before each page transition.
- [ ] Enforces 25-second cool-off pause after every 10 fetched pages.
- [ ] Error handler detects HTTP 403 / 429, immediately terminates workflow to avoid permanent ban, records error in `crawl_logs`, and sends alert notification.

## Blocked by

- #5: Ticket 04: Worker Pool Bulk SQL Updater & Ingestion Webhook
