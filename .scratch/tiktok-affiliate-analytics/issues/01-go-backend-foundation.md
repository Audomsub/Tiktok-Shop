## Parent

#1

## What to build

A high-performance Go backend HTTP service foundation utilizing the Chi router, establishing a robust database connection pool with Supabase PostgreSQL via pgxpool, loading structured configuration from environment variables, enabling CORS, and exposing an operational healthcheck endpoint that verifies database ping connectivity.

## Acceptance criteria

- [ ] Go module is initialized with Chi router, pgxpool v5, and godotenv dependencies.
- [ ] Application configuration safely loads DATABASE_URL, PORT, and environment secrets with fallback defaults.
- [ ] Database connection pool connects to Supabase PostgreSQL and successfully executes ping checks.
- [ ] GET /health returns HTTP 200 with database status JSON payload `{"status":"ok","database":"connected"}`.
- [ ] Proper structured logging and graceful shutdown handlers are established.

## Blocked by

None (can start immediately)
