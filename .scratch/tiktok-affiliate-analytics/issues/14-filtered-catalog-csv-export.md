## Parent

#1

## What to build

A high-performance CSV export feature streaming filtered product records from Go `GET /api/v1/products/export` with UTF-8 BOM encoding for seamless Microsoft Excel and Numbers compatibility, paired with an "Export to CSV" button in the Next.js filter toolbar that passes active filter parameters to download the matching dataset.

## Acceptance criteria

- [ ] Endpoint `GET /api/v1/products/export` accepts same filter parameters as catalog search (`category_id`, `min_price`, `max_price`, `min_commission`, `q`).
- [ ] Streams CSV response using `encoding/csv` without buffering entire dataset in memory.
- [ ] Writes UTF-8 Byte Order Mark (BOM `\xEF\xBB\xBF`) at start of file so Thai product names display cleanly in Excel.
- [ ] Columns: Product ID, Name, Category, Price, Commission Rate, Expected Return THB, Total Sales, Velocity Per Hour, Winning Score, TikTok Product URL.
- [ ] Next.js toolbar includes "Export CSV" button with downloading spinner state.

## Blocked by

- #8: Ticket 07: Product Catalog & Dynamic Category Search API
- #9: Ticket 08: Filterable Product Table & Range Slider Controls UI
