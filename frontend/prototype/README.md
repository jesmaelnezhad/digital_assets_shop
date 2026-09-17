# Pawradise frontend prototype

Portable mock of the shop MFEs. No Go services required. Open in a browser.

```bash
cd frontend/prototype
python3 -m http.server 4173
# http://127.0.0.1:4173/
```

`fetch` for `/api/v1` is intercepted in `shared/mock-api.js`. JSON **keys and nesting** match the live Go handlers. Differences: `API-DIVERGENCE.md`. Demo data lives in `localStorage` (`pawradise-proto-v4`).

Walk the product as a person with `SCENARIOS.md` (jobs, edges, and cross-page state — not automated tests).

## Demo accounts

| Role | Email | Password |
|---|---|---|
| Buyer | nia@example.com | nia |
| Buyer | leo@example.com | leo |
| Buyer | maya@example.com | maya |
| Buyer | owen@example.com | owen |
| Admin Bearer | `admin_secret_staging_2026` | also `studio-admin` |

Coupons: `SAVE12`, `WELCOME`, `MARBLE`.

Cart, wishlist, compare, and recently-viewed return **401** until you log in (same as the backend).

## Pages

| MFE | File |
|---|---|
| Shop | `shop-mfe/index.html`, `category.html`, `legal.html` |
| Product | `product-mfe/index.html?slug=…`, `bundles.html`, `bundle.html?id=…`, `request.html` |
| Community | `community-mfe/index.html`, `people.html`, `post.html`, `profile.html?id=&tab=` |
| Account | `account-mfe/account.html`, `order.html?id=…`, `cart.html`, `wishlist.html`, `compare.html`, `recent.html`, `referrals.html`, `guest.html` |
| Checkout | `checkout-mfe/index.html` |
| Auth | `auth-mfe/login.html`, `register.html` |
| Admin | `admin-app/index.html` (token gate, then Stats / Users / Products / Categories / Bundles / Coupons / Orders / Guest / Community / Referrals / Rates / Settings / SEO / Export / Requests) |

SEO: `sitemap.xml`, `robots.txt`. Admin Bearer is entered in the gate (not auto-filled). Design notes: `DESIGN.md`.
