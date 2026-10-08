## Parent

#1

## What to build

An interactive Product Catalog page on Next.js featuring a comprehensive filter toolbar (category select dropdown, dual-thumb price range slider, min commission percentage slider, and debounced search bar) wired to a TanStack Table with sortable column headers, pagination controls, thumbnail previews, and external TikTok Shop product links.

## Acceptance criteria

- [ ] Category dropdown dynamically populates options from Go API `GET /api/v1/categories`.
- [ ] Price slider filters products within 80–1,500+ THB.
- [ ] Commission slider filters products with commission rates $\ge 10\%$.
- [ ] Search input debounces API calls by 300ms to eliminate redundant network traffic.
- [ ] TanStack Table displays product image, title, category, price, commission %, total sales, hourly velocity, and winning score.
- [ ] Column headers allow sorting by any numeric metric.
- [ ] Pagination controls allow jumping between pages with customizable page size (10, 25, 50).

## Blocked by

- #7: Ticket 06: Next.js Shell & Real-time Leaderboard Cards UI
- #8: Ticket 07: Product Catalog & Dynamic Category Search API
