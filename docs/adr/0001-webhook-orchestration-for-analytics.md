# Webhook Orchestration for Analytics Engine

n8n triggers the Go Analytics Engine via an authenticated HTTP POST webhook with `crawl_log_id` immediately upon completing ingestion. We chose a direct webhook trigger over database polling or a dedicated message broker to eliminate idle polling latency and operational overhead while maintaining a decoupled architecture.
