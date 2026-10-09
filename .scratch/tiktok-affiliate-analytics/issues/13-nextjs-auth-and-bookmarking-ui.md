## Parent

#1

## What to build

A seamless user authentication and bookmarking experience on Next.js leveraging `@supabase/ssr` / Supabase Auth, integrating login/logout buttons in the top navbar, heart/bookmark action buttons with optimistic UI updates on product cards and table rows, and a dedicated "/favorites" page displaying the creator's saved products.

## Acceptance criteria

- [ ] Supabase Auth client is initialized on Next.js with environment variables.
- [ ] Navbar contains Sign In / Sign Out actions and shows active user email when logged in.
- [ ] Heart/bookmark button on product cards toggles status instantly with optimistic UI.
- [ ] Next.js attaches Supabase session access token to Go API requests for `/favorites`.
- [ ] Dedicated "My Favorites" page lists all bookmarked products with unfavorite action.
- [ ] Prompt appears if an unauthenticated user attempts to bookmark a product, guiding them to sign in.

## Blocked by

- #9: Ticket 08: Filterable Product Table & Range Slider Controls UI
- #13: Ticket 12: Favorite Products CRUD API Endpoints
