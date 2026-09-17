# Prototype API vs live Go handlers

The mock in `shared/mock-api.js` returns the same JSON **keys and nesting** as the Go handlers under `services/*/handlers`. Pages join extra data themselves (cart `product_id` → `GET /products`). This file lists every intentional difference.

## Extra routes (handler exists, not registered on live `main.go`)

| Method | Path | Live | Prototype |
|---|---|---|---|
| GET | `/api/v1/products/:id/tiers` | `GetProductTiers` exists in product-service but is not routed publicly | `{ "tiers": [ ProductTier ] }` so the product page can show licenses. Live GET product is only `{ product, images }`. |
| PUT | `/api/v1/cart/items/:id` | `UpdateCartItemRequest` exists in commerce models; **not wired** in `commerce-service/main.go` | `{ "message": "quantity updated" }` so cart qty inputs work. Body `{ quantity }`. |

## Missing on live (spec / FEATURE-CANDIDATES / current frontend) — implemented in the mock

These are **not** in the wired Go `main.go` files. The prototype implements them so admin and buyer flows in the spec can be used. Response keys below are what the mock returns.

| Method | Path | Why | Prototype body / response |
|---|---|---|---|
| POST | `/api/v1/product-requests` | Live `product-mfe/src/request.html` posts here; no Go handler | `{ title, description, category, budget_usd, email }` → 201 `{ id, message: "request submitted" }` |
| GET | `/api/v1/product-requests` | Buyer list of own requests | `{ requests: [{ id, title, description, category, budget_usd, email, status, created }] }` |
| GET | `/api/v1/admin/product-requests` | Admin moderation of custom requests | `{ requests: [ ...same ] }` |
| PUT | `/api/v1/admin/product-requests/:id` | Admin status | `{ status }` → `{ message: "request updated" }` |
| GET | `/api/v1/admin/guest-orders` | Spec §4.12 / §6.5; **not** in `admin-service/main.go` | `{ orders: [{ id, email, total_usd, status, crypto_chain, created_at }], total }` |
| GET | `/api/v1/admin/guest-orders/:id` | Spec §4.12 | `{ order, email, items }` |
| POST | `/api/v1/admin/bundles` `product_ids` | Live `CreateBundle` does **not** bind `product_ids` (title/description/price/status only) | Prototype **accepts** `product_ids: number[]` and stores them. Response still `{ id, message: "bundle created" }`. |
| GET | `/api/v1/community/users/:id/followers` | Spec follow graph; no list handler live | `{ users: [Member], total }` — Member has `id, name, email, avatar_url, bio, follower_count, following_count, following, follows_you` |
| GET | `/api/v1/community/users/:id/following` | Same | Same shape |
| GET | `/api/v1/community/users` | Member directory | `{ users: [Member], total }`. Optional `q` filters name/bio/email. Sorted by `follower_count`. |
| GET | `/api/v1/community/suggestions` | People you don't follow yet | `{ users: [Member] }` |
| GET | `/api/v1/admin/categories` | Live public `GET /categories` is active-only; no admin list | `{ categories }` including inactive. Extra key `product_count`. |
| PUT | `/api/v1/categories/:id` | Live wires `POST /categories` only | `{ name, slug, description, parent_id, sort_order, image_url, is_active }` → `{ message: "category updated" }`. Renaming `slug` retargets products. |
| DELETE | `/api/v1/categories/:id` | Live has no delete | `{ message: "category deleted" }`. 400 if products or child categories still use it. |

Site-wide SEO uses existing `PUT /admin/settings/:key` (generic KV). Extra **setting keys** the mock seeds: `site_title`, `site_description`, `site_keywords`, `og_image`, `canonical_host`, `robots_index`, `download_policy`, `default_currency`, `site_name`. Live will store any key if the row is written.

Per-product SEO fields `seo_title`, `seo_description`, `og_image` are extra on `POST/PUT /admin/products` and echoed on public `GET /products/:slug` `product` and admin product list. Live `UpdateProduct` does not bind them.

## Same schema, richer values

| Endpoint | Live | Prototype |
|---|---|---|
| `GET /products` | Ignores `search`, `category`, `sort`, `rating`, `file_type`, `price_*`. `digital_formats` / `is_pwyw` are usually empty because the list SELECT omits them. | Honors those query params (spec). Fills `digital_formats`, `is_pwyw`, `pwyw_min_price` from seed. **No extra keys** (`file_type`, `rating`, `category_name` are omitted on list). |
| `GET /recommendations/:id` | Always `{ "products": [] }` | Same key; fills same-category products so related cards work. |
| `GET /community/posts` | `total` is the current page length; `filter` ignored; `user_id` ignored | `total` is the full count. `filter=following` returns posts from followed users (empty list if logged out, not 401). `user_id` filters a profile's posts. Extra post key `liked`. |
| Feed `author.id` | Scan never sets it (serializes as `0`) | Set to `user_id` so profile links work. Extra author key `follows_you`. |
| `GET /community/users/:id` | `id, email, name, avatar_url, post_count, follower_count, following_count, created_at, updated_at` | Same keys **plus** spec fields `bio`, `wallet_address` and relationship flags `following`, `follows_you`, `is_self` (live profile UI already reads `following`). |
| `POST /community/follow/:id` | 400 `cannot follow yourself`, 409 `already following` | Same errors; mock no longer toggles on POST. |
| `DELETE /community/follow/:id` | 404 `not following` | Same. |
| `POST /admin/products/:id/pin` | Returns `{ "message": "product pinned" }` and does not update the DB | Same message; actually sets `pinned`. |
| `GET /admin/stats` | `{ total_users, total_orders, total_revenue, total_products, active_products, pinned_products }` | Same six keys **plus** spec analytics: `total_categories`, `total_downloads`, `revenue_daily: [{ date, revenue }]`, `top_products: [{ id, title, units, revenue }]`, `conversions: { views, carts, checkouts, purchases, visitor_to_view, view_to_cart, cart_to_checkout, checkout_to_purchase }`. |
| `POST /admin/export/emails` | Lists all users; ignores filters; optional `?format=csv` | Accepts JSON `{ from, to, category, product_id }` (spec §6.10) and extra user keys `total_purchases`, `first_purchase_date`, `last_purchase_date`. Still `{ users, count }` for JSON. |
| `POST /admin/products/bulk` | Wired admin body is `{ ids, status, category_id }` → `{ message: "bulk update applied" }`. Unwired product-service variant uses `{ product_ids, action }`. | Accepts **both**. Response `{ message, action, rows_affected }`. |
| `POST /categories` | Admin Bearer. Body `{ name, slug, description, parent_id }` → 201 `{ message: "category created" }`. Slug required. | Same path and message. Also binds `sort_order`, `image_url`, `is_active`. Empty slug is filled from name. |
| `GET /categories` | Active rows only, `{ categories }` | Same. Inactive categories are omitted until an admin reactivates them. |
| Admin product list | No PWYW / SEO / file type | Extra keys `pwyw_enabled`, `pwyw_min_price`, `digital_formats`, `seo_title`, `seo_description`, `og_image`. |
| `GET /orders/:id` | `{ order }` without line items | Same `order` object **plus** `items: [{ id, order_id, product_id, product_title, product_slug, quantity, unit_price_usd, download_count }]` so account re-downloads work. |

## Request-body vs live path params

Live `RecordView` / `ToggleCompare` / `ToggleWishlist` read `c.Param("id")`, but the routes have no `:id` and `shared/lib/api.js` posts `{ "product_id": n }`. The mock follows the **client body** so the pages work. Response messages still match the handlers (`added to wishlist`, `added to compare`, `view recorded`).

## Auth

| Endpoint | Live | Prototype |
|---|---|---|
| Cart, wishlist, compare, recently-viewed, orders | JWT / session cookie | Same: 401 `{ "error": "unauthorized" }` if not logged in. Token `mock.{userId}.user` or `Paw` session after login. |
| Admin | `AdminAuthMiddleware` | Bearer `admin_secret_staging_2026` or `studio-admin`. Admin UI stores it in `sessionStorage.admin_token` (spec) and `localStorage.pawradise_admin_token`. |

Register accepts `referral_code` in the prototype even though live `Register` does not bind it (identity still creates a referral **link** for the new user).

## Response shapes the UI must not assume

These are **not** on the wire (they used to be in an earlier mock):

- Cart items are `{ id, product_id, quantity }` only — no nested `product`, no `unit_price_usd`.
- Wishlist / compare / recently-viewed are `{ products: [{ id, product_id, ... }] }` — no nested product.
- `GET /me` is a bare `User` (`id`, `email`, `name`, `created_at`, `updated_at`, optional `profile`) — not `{ user: ... }`.
- Login is `{ user, token }` where `user` has no `profile`.
- `PUT /me` returns `{ name, bio, wallet_address }` on live. Prototype **also** binds `avatar_url` on the request and echoes it (data URL or http(s) in the mock). Live is URL-on-profile only; file upload is out of spec.
- `GET /referrals` is `{ referral_link, referrals, total_earnings, total_commissions, total_referrals }`. Prototype `referrals[]` items also include `name` and `email`.
- `GET /exchange-rates/:chain` is `{ rate: ExchangeRate }` with `rate_to_usd` as a **string** (payment-service model). Admin list uses a **number**.
- Coupon validate is `{ valid, discount_type, discount_value }` (`discount_value` is a string). No `discount_usd`.
- Create order: `{ order: { id }, total_usd, payment_address, total_crypto, crypto_chain, memo }`. Guest create: `{ id, message: "guest order created" }` — no address; guest checkout reads `GET /settings/payment_address`.
- `GET /orders/:id/payment` is `{ order_id, status, total_usd }`. Confirm is `{ message: "payment confirmed" }`.
- Create post: `{ id, user_id, content, type }` (201). Get post: `{ post, author, comments, total_comments }` with comment field `user`, not `author`.
- Like / unlike: `{ message: "post liked" | "post unliked" }`.
- Create review: `{ review: { id, product_id, rating } }`. Average: `{ product_id, average_rating, total_reviews, total_rating, distribution }`.
- Admin users: `{ users: [ id, email, name, created_at, updated_at ] }` (the `order_count` variant in `users.go` is not the wired handler).
- Bundles: `items` is a **string** (comma-separated product ids in the prototype). Live SELECT does not populate it, so it is usually `""`.
- `GET /orders/:id/download/:itemId` is `{ message: "download endpoint" }` (live), not a file.
- Admin CUD messages match handlers: `product created` / `product updated` / `product deleted`, `tier added` / `tier updated` / `tier deleted`, `image added` / `image deleted`, `previews generated`, `bundle created` / `bundle updated` / `bundle deleted`, `coupon created` / `coupon updated` / `coupon deleted`, `exchange rate updated` / `exchange rate deleted`.

## Public GET `/settings/:key`

Not registered on live payment-service (settings are admin-only). Prototype returns `{ key, value }` so guest checkout can show the wallet and chrome can read site SEO defaults. E2E already allows 200 or 404 for this path.

## Static files (not API)

Spec §5.8 sitemap / robots are pages, not `/api/v1` routes. Prototype serves `sitemap.xml` and `robots.txt` at the prototype root.
