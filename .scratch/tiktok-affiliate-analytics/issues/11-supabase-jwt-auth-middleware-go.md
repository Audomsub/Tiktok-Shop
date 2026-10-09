## Parent

#1

## What to build

A robust HTTP authentication middleware in Go using `golang-jwt/jwt/v5` that parses and validates Supabase JWT access tokens passed via the `Authorization: Bearer <token>` header against the Supabase JWT Secret (`HS256`), extracts the authenticated `user_id` (`sub` claim) into the request context, and provides optional vs required auth guards.

## Acceptance criteria

- [x] Middleware extracts Bearer token from HTTP `Authorization` header.
- [x] Validates token signature, expiration (`exp`), and issuer using the project Supabase JWT Secret.
- [x] Injects verified `user_id` (UUID) into `r.Context()`.
- [x] Returns HTTP 401 Unauthorized with standard error JSON for expired, forged, or missing tokens on protected endpoints.
- [x] Supports an optional-auth mode for public endpoints (such as `/products`) to detect whether a requesting user is signed in.
- [x] Unit tests verify valid, expired, and maliciously altered JWT tokens.

## Blocked by

- #2: Ticket 01: Go Backend Foundation, Database Pool & Healthcheck
