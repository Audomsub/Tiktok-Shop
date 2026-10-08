## Parent

#1

## What to build

A modern, responsive Next.js App Router presentation dashboard with Tailwind CSS and shadcn/ui components, including a root layout header/navigation bar, and a real-time Top 10 Winning Leaderboard grid showcasing product cards with thumbnail images, live scores, velocity badges (🔥, 💎, 💰, 🏆), and direct outbound links to TikTok Shop.

## Acceptance criteria

- [ ] Next.js 14/15 App Router project is initialized with TypeScript, Tailwind CSS, Lucide icons, and shadcn/ui.
- [ ] Root navigation header is rendered with application branding, navigation links, and theme container.
- [ ] Leaderboard view fetches live data from Go backend `GET /api/v1/leaderboard`.
- [ ] Renders top 10 products with rank medals (1st, 2nd, 3rd badges), winning score circular progress/badge, price in THB, sales velocity, and commission rate.
- [ ] Visual status badges (🔥 Viral Surge, 💎 High Commission, 💰 High Yield, 🏆 Winning Top Pick) render correctly based on API flags.
- [ ] Includes clickable external link opening TikTok Shop product page in a new tab.

## Blocked by

- #6: Ticket 05: Top 10 Winning Leaderboard & Badge API
