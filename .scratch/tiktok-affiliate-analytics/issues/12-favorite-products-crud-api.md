## Parent

#1

## What to build

Authenticated Go REST API endpoints `GET /api/v1/favorites`, `POST /api/v1/favorites`, and `DELETE /api/v1/favorites/:productId` allowing signed-in users to manage their bookmarked product shortlist, plus enriching `GET /api/v1/products` and `GET /api/v1/leaderboard` with a boolean `is_favorited` property when called with an active user token.

## Acceptance criteria

- [ ] `POST /api/v1/favorites` accepts `{"product_id": "uuid"}` and inserts into `favorite_products` using authenticated `user_id`.
- [ ] Safe idempotent insert (`ON CONFLICT (user_id, product_id) DO NOTHING`).
- [ ] `DELETE /api/v1/favorites/:productId` removes bookmark for authenticated user.
- [ ] `GET /api/v1/favorites` returns paginated list of user's favorited products with latest metrics.
- [ ] Public endpoints (`/leaderboard`, `/products`) evaluate `is_favorited: true/false` if an optional valid Bearer token is provided.
- [ ] Integration tests verify that User A cannot view or delete User B's favorites.

## Blocked by

- #8: Ticket 07: Product Catalog & Dynamic Category Search API
- #12: Ticket 11: Supabase JWT Authentication Middleware in Go
