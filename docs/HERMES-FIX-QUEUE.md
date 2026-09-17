# Hermes fix queue

This file is the remaining-work list for bringing Pawradise in line with `docs/PRODUCT-SPEC.md`. When it contains no `### FIX-` sections, the queue is done.

HTTP 200 on HTML pages is not proof that an item is done.

---

## How to empty this file

1. Take the **lowest remaining FIX-ID** still in this file (skip only if that item’s **Blocked by** IDs are still present).
2. Write a **failing test** that asserts **Expected**. Prefer:
   - Go unit tests under `tests/unit/<service>/`
   - Go integration tests under `tests/integration/` for cross-service behavior
   - `tests/e2e-suite.js` only when the contract is the public HTTP API
3. Run the test. Confirm it fails for the reason in **Observed**, not because of a typo.
4. Change **services** (and k8s/ingress/frontend only when the item’s **Surface** says so) until the test **passes**. Do not weaken the test to match a stub (`{"message":"…"}` is not a download).
5. **Delete that item’s entire `### FIX-NNN` section** from this file (from the heading through the blank line before the next `### FIX-`).
6. Repeat until grep finds no `### FIX-`.

### Rules

- Contract: `docs/PRODUCT-SPEC.md`. Feature decisions: `docs/FEATURE-CANDIDATES.md`.
- Public and service HTTP paths are `/api/v1/...`. Do **not** introduce `/internal/v1`.
- Each service owns its tables. Do **not** copy `products` / `users` / `orders` into another service’s database to make a test pass. Call the owning service or use events.
- Protected routes must reject missing JWT with 401. Do not “pass” by skipping auth.
- After a backend contract change, keep `shared/lib/api.js` and the relevant MFE in sync if the frontend would otherwise call the wrong shape.

---

### FIX-007 — Product list honors `file_type` filter

- **Surface:** product-service
- **Spec:** §4.3 filter by file type
- **Expected:** results only include products whose stored mime/extension matches `file_type`.
- **Observed:** param unused. Add a real file-type field if missing rather than ignoring the param.

### FIX-008 — Product list honors `rating` filter

- **Surface:** product-service
- **Spec:** §4.3 filter by rating
- **Expected:** only products whose average rating is ≥ requested rating (or exact, if you document it). Unrated products are excluded when a rating filter is set.
- **Observed:** param unused. Average may live in review-service — call it or maintain an aggregate without copying the reviews table.

### FIX-009 — Pagination `total` matches the filtered set

- **Surface:** product-service
- **Blocked by:** FIX-001
- **Expected:** `page` / `per_page` slice the **filtered** result. `total` is the filtered count, not all active products.
- **Observed:** `total` is `count(*)` of all active products even when filters are applied (once filters exist).

### FIX-010 — Pinned products appear first, then the requested sort

- **Surface:** product-service
- **Blocked by:** FIX-002
- **Spec:** §2.1 pinned first in listings
- **Expected:** all pinned active products precede unpinned ones; within each group, `sort` applies. `is_pinned` / `pinned` must agree on list and detail.
- **Observed:** pinned-first roughly works on list, but `GET /products/:slug` returned `is_pinned: false` while list had `is_pinned: true` for product-1.

### FIX-011 — Product detail images belong to that product

- **Surface:** product-service
- **Expected:** `GET /api/v1/products/:slug` `images[].product_id` equals the product id. Primary image and gallery URLs are present when stored.
- **Observed:** `images[0].product_id` was 0. Public image URLs were `images.example.com` placeholders (seed quality is separate; the id join is not).

### FIX-012 — Product detail includes tiers with own price and file

- **Surface:** product-service
- **Spec:** §5.4 `product_tiers` have `tier_name`, `file_path`, `price_usd`
- **Expected:** detail payload includes `tiers` (possibly empty). Creating a tier via admin API then fetching the product returns that tier’s **absolute** `price_usd` and file path, not only a `price_modifier`.
- **Observed:** schema uses `price_modifier`; list/detail did not expose spec-shaped tiers.

### FIX-013 — Product detail includes PWYW config

- **Surface:** product-service
- **Spec:** PWYW per product
- **Expected:** when enabled, detail includes `pwyw_enabled` (or `is_pwyw`) and `pwyw_min_price`. List cards can show PWYW.
- **Observed:** columns exist; sampled products false. Ensure create/update admin API persists and detail returns them under stable names matching `shared/lib/api.js`.

### FIX-014 — Related products / recommendations JSON

- **Surface:** product-service, gateway
- **Spec:** §4.8 `GET /api/v1/recommendations/:productId`
- **Expected:** JSON object with a product array (same-category or documented rule). HTTP 200 `Content-Type: application/json`. 404 if product missing.
- **Observed:** `GET /api/v1/recommendations/1` returned the **shop homepage HTML** (ingress `/` catch-all). Service also returned `null` when hit internally.

### FIX-015 — Public bundle list is the product catalog’s bundles

- **Surface:** product-service, admin-service
- **Spec:** §4.3 `GET /api/v1/bundles`
- **Expected:** active bundles created via `POST /api/v1/admin/bundles` appear on `GET /api/v1/bundles`. Empty only when none exist in **product** data, not because admin wrote a different database.
- **Observed:** public `bundles: []` while admin list returned e2e/test bundles.

### FIX-016 — Bundle detail includes component products and prices

- **Surface:** product-service
- **Spec:** §4.3 `GET /api/v1/bundles/:id`
- **Expected:** payload includes component products, bundle `price_usd`, and enough data to compute savings vs individual totals.
- **Observed:** public list empty so detail unused; admin bundle rows not in product-service DB.

### FIX-017 — Admin pin / unpin persist and affect listing

- **Surface:** product-service, admin-service
- **Spec:** §4.12 pin / unpin
- **Expected:** `POST /api/v1/admin/products/:id/pin` then public list has that product among pinned. `DELETE .../pin` unpins. Both services must hit **product** storage, not an admin copy.
- **Observed:** admin stats `pinned_products: 0` while public list showed many `is_pinned: true`.

### FIX-018 — Digital goods are not stock-limited

- **Surface:** product-service, commerce-service
- **Spec:** buy once, download forever; no physical stock
- **Expected:** public product JSON does not require `stock_quantity` for purchase. Orders succeed when the product is `active` even if stock fields are 0. Do not reject checkout for “out of stock” unless you explicitly add a digital license cap (out of spec).
- **Observed:** products expose `stock_quantity`, `max_downloads_per_user`, `download_window_hours` from a physical-goods leftover schema.

### FIX-019 — Categories list is complete and filterable by slug

- **Surface:** product-service
- **Spec:** §4.3 `GET /api/v1/categories`
- **Expected:** each category has `id`, `name`, `slug`, `description`, `parent_id`. Product filter in FIX-005 uses these slugs.
- **Observed:** 10 categories exist; filter by slug did not apply (see FIX-005).

### FIX-020 — Order create computes a real USD total from catalog prices

- **Surface:** commerce-service
- **Spec:** §4.4 `POST /api/v1/orders`
- **Expected:** order total equals sum of (catalog unit price or PWYW/tier price × qty) minus valid coupon. For a cart with product id 1 qty 2, `total_usd` equals `2 * product.price_usd` (or documented tier/PWYW rules). Prices come from **product-service**, not a local `products` table in commerce DB.
- **Observed:** `POST /orders` returned `total_usd: 0`, `total_crypto: 0`.

### FIX-021 — Created order can be fetched by the owner

- **Surface:** commerce-service
- **Blocked by:** FIX-020
- **Expected:** after 201 create, `GET /api/v1/orders/:id` with the same JWT returns that order (items, total, status). `GET /api/v1/orders` includes it.
- **Observed:** create returned `order.id` 392; GET by id 404; list `orders: []`.

### FIX-022 — Order payment and status endpoints

- **Surface:** commerce-service, payment-service
- **Blocked by:** FIX-021
- **Spec:** §4.4 `/orders/:id/payment`, `/orders/:id/status`
- **Expected:** JSON with status (`pending`/`paid`/…) and payment fields for the owner. 404 only if the order does not exist. 403 if it belongs to someone else.
- **Observed:** both 404 immediately after create. Payment-service `GET /payments/order/:id` 404.

### FIX-023 — Coupon code discounts the order total

- **Surface:** commerce-service
- **Blocked by:** FIX-020, FIX-026
- **Spec:** §2.3 coupon flow
- **Expected:** valid `coupon_code` on create reduces `total_usd` by percentage or fixed amount, respects min purchase, increments usage. Invalid code → 400, order not created (or created without silent ignore — pick one, test it, document it; silent ignore is not valid).
- **Observed:** create accepted flow with total 0; usage/discount not applied.

### FIX-024 — Guest order create and status check

- **Surface:** commerce-service
- **Spec:** §4.9
- **Expected:** `POST /api/v1/guest-orders` with email + items creates an order. `GET /api/v1/guest-orders/:id` with email verification returns it. **Do not** expose `GET /api/v1/guest-orders` as a public dump of all guest orders.
- **Observed:** POST 201; GET by id 404. Public GET list returned `{"orders":[]}` (endpoint should not be an unauthenticated catalog of guests).

### FIX-025 — Download requires paid purchase and returns the file

- **Surface:** commerce-service, media-service
- **Blocked by:** FIX-021, FIX-035
- **Spec:** §4.4 `GET /api/v1/orders/:id/download/:itemId`
- **Expected:** 401 if unauthenticated. 403/404 if unpaid or not owner. 200 with file bytes (or redirect to signed URL) when paid. Unlimited re-download. Not a JSON stub.
- **Observed:** `{"message":"download endpoint"}`.

### FIX-026 — Coupon validate uses the same store admin writes

- **Surface:** commerce-service, admin-service
- **Spec:** §4.5 / §4.12
- **Expected:** `POST /api/v1/admin/coupons` then `POST /api/v1/coupons/validate` with that code returns discount details for a sufficient cart total. Unknown code → 404/400.
- **Observed:** admin listed coupon codes; validate of a listed code returned `coupon not found`. Admin create returned `{id, message}` without the code.

### FIX-027 — Cart GET returns the items that were added

- **Surface:** commerce-service
- **Expected:** after `POST /cart/items` `{product_id, quantity}`, `GET /cart` items include that `product_id` and `quantity`. Missing auth → 401.
- **Observed:** add 200; GET showed product_id 1 qty 2 (this path worked). Keep a regression test. If you change cart shape, update `shared/lib/api.js`.

### FIX-028 — Wishlist toggle GET returns real product ids

- **Surface:** commerce-service
- **Spec:** §4.8
- **Expected:** after toggle add, `GET /wishlist` includes `product_id` of the product (not 0). Toggle again removes it.
- **Observed:** `{"products":[{"id":521,"product_id":0}]}`.

### FIX-029 — Compare list returns real product ids

- **Surface:** commerce-service
- **Expected:** same as wishlist for `POST /compare/toggle` and `GET /compare`.
- **Observed:** `product_id: 0`.

### FIX-030 — Recently viewed returns real product ids

- **Surface:** commerce-service
- **Expected:** `POST /recently-viewed` then `GET /recently-viewed` includes the `product_id`. Newest first. Cap list length reasonably.
- **Observed:** `product_id: 0`.

### FIX-031 — Register captures `referral_code`

- **Surface:** identity-service
- **Spec:** §4.2 optional `referral_code`
- **Expected:** register with a valid referrer code records `user_referrals` (referrer → new user). Invalid code does not fail register, but does not invent a referrer. Response or subsequent `GET /referrals` as referrer shows the new user.
- **Observed:** body `referral_code` ignored; handler only binds name/email/password.

### FIX-032 — Public profile by id (no JWT)

- **Surface:** identity-service, gateway
- **Spec:** §4.2 `GET /api/v1/profile/:id` public
- **Expected:** 200 with public fields (name, bio, avatar, not password). 404 if missing. No auth required.
- **Observed:** identity `GET /profile/:id` is behind JWT and returned 404 for the authenticated owner. Community `GET /community/users/:id` returned stub `name: "User"`, `email: "user@example.com"`.

### FIX-033 — Community public profile uses identity data

- **Surface:** community-service, identity-service
- **Blocked by:** FIX-032
- **Expected:** post author and `GET /community/users/:id` show the user’s real display name from identity (HTTP call), not a local stub user row.
- **Observed:** author email `user@example.com` for a just-registered user.

### FIX-034 — Logout revokes referrals and other JWT routes

- **Surface:** identity-service
- **Spec:** §4.2 logout invalidates JWT
- **Expected:** after `POST /logout`, `GET /referrals`, `GET /me`, `GET /cart` with the same bearer all 401.
- **Observed:** `/me` 401 “Token revoked”; `/referrals` still 200. Handler comments admit a bypass for e2e.

### FIX-035 — Confirm payment only for a real pending order

- **Surface:** commerce-service, payment-service
- **Blocked by:** FIX-021
- **Expected:** confirm on missing order → 404. Confirm on pending order → status `paid` (simplified flow). Confirm on already paid is idempotent 200. Cannot confirm another user’s order.
- **Observed:** `POST /orders/:id/confirm` 200 `payment confirmed` even when GET order 404.

### FIX-036 — Paid order of a referred user records commission

- **Surface:** identity-service, commerce-service
- **Blocked by:** FIX-031, FIX-035
- **Spec:** §2.3 referral = purchase total × admin `referral_commission_percent`
- **Expected:** after referred user pays, referrer `GET /referrals/earnings` includes commission_usd. Percent read from settings (admin), default documented.
- **Observed:** events package unused; no commission on purchase.

### FIX-037 — Public GET setting by key

- **Surface:** payment-service or admin-service, gateway
- **Spec:** §4.11 `GET /api/v1/settings/:key` public read
- **Expected:** JSON value for `referral_commission_percent` (and other public keys). Unknown key 404. List-all remains admin.
- **Observed:** `GET /api/v1/settings/referral_commission_percent` → `404 page not found`. Admin GET `/admin/settings` has the key `5.0`.

### FIX-038 — Crypto amount uses exchange rates

- **Surface:** payment-service
- **Blocked by:** FIX-020, FIX-022
- **Spec:** §2.3 exchange rates admin-set
- **Expected:** payment details include destination, `crypto_chain`, `crypto_amount` ≈ `total_usd / rate_to_usd` for the chosen chain. Placeholder `0xPAWRADISE_WALLET_BSC` is not acceptable as the only production address — use settings `payment_address`.
- **Observed:** `total_crypto: 0`, hardcoded wallet string.

### FIX-039 — Review requires a paid purchase

- **Surface:** review-service, commerce-service
- **Spec:** §4.8 verified purchase only, 1–5 stars, no review media
- **Expected:** POST review without a paid order for that product → 403. After paid order → 201. Duplicate review → 409. Do not set `is_verified_purchase: true` without a check. Do not `UPDATE products` in review-service’s database.
- **Observed:** 201 without purchase; response `user_id: 0`; review-service updates a local `products` table.

### FIX-040 — Review list and average for a product

- **Surface:** review-service
- **Expected:** `GET /api/v1/reviews/:productId` lists ratings; average endpoint matches the math. Distribution 1–5.
- **Observed:** list exists; average path may 404 if not registered as `/reviews/:id/average` vs a dedicated route — add a test that hits the spec path.

### FIX-041 — Community feed pagination and create

- **Surface:** community-service
- **Expected:** `GET /api/v1/community/posts` returns `{posts, total, page, per_page}`. Authenticated POST creates a post visible on the next GET. Content over ~500 chars → 400 (already observed working — keep a regression test).
- **Observed:** create 201; public feed was empty before the probe (no seed). After create, GET by id worked. Add a test that list includes the created post.

### FIX-042 — Admin community list/delete uses community-service

- **Surface:** admin-service, community-service
- **Spec:** §4.12
- **Expected:** `GET /api/v1/admin/community/posts` includes posts from community-service. `DELETE` removes them from the public feed.
- **Observed:** admin list `posts: []` after a successful community POST.

### FIX-043 — Admin stats aggregate from owning services

- **Surface:** admin-service
- **Spec:** §4.12 / §6.11 totals
- **Expected:** `total_users` matches identity user count, `total_orders` / `total_revenue` match commerce, `total_products` matches product-service, not admin’s copied tables. Do not hardcode `dbname=appdb_identity_staging`.
- **Observed:** stats `total_users: 0`, `total_orders: 0`, `total_revenue: 0` while public catalog had 300 products and admin **users** list included a newly registered user. `admin-service/main.go` hardcodes identity staging DB.

### FIX-044 — Admin email export lists buyer emails

- **Surface:** admin-service, commerce-service, identity-service
- **Spec:** §4.12 / §6.10
- **Expected:** `POST /api/v1/admin/export/emails` returns buyers (paid orders) as CSV or JSON with email, name, purchase aggregates. Empty only if there are no buyers.
- **Observed:** `{"count":0,"users":[]}` despite users and orders existing in other DBs.

### FIX-045 — Admin order list is commerce orders

- **Surface:** admin-service, commerce-service
- **Expected:** `GET /api/v1/admin/orders` shows orders created through commerce, including guest orders, with status update via `PUT /admin/orders/:id/status` (valid transitions only).
- **Observed:** admin `orders: []` after a user order create.

### FIX-046 — Admin product CRUD writes product-service

- **Surface:** admin-service, product-service
- **Expected:** `POST /api/v1/admin/products` appears on `GET /api/v1/products` when `status=active`. Update/delete/pin/tiers/images hit the same rows as the public catalog.
- **Observed:** admin create 201 in admin DB; public catalog remained the “Product N” seed in product-service.

### FIX-047 — JWT middleware does not treat missing token as anonymous-success on protected routes

- **Surface:** services/shared/middleware, all services
- **Expected:** routes documented as JWT-protected return 401 when there is no bearer and no session cookie. Optional-auth routes (if any) are explicit.
- **Observed:** `JwtAuthMiddleware` calls `c.Next()` when token is absent; correctness depends on every handler checking `user_id`. Several handlers are inconsistent.

### FIX-048 — No implicit default JWT / admin secrets

- **Surface:** services/shared/auth, middleware
- **Expected:** if `JWT_SECRET` or `ADMIN_TOKEN` is unset, the process logs a fatal error in staging/production. Tests may inject secrets.
- **Observed:** fallbacks `your-secret-key-change-in-production` and `admin-secret-token-change-in-production`.

### FIX-049 — `GET /api/v1/health` is JSON from a backend

- **Surface:** gateway, any service
- **Expected:** `GET /api/v1/health` returns JSON `{"status":"ok",...}` with `Content-Type: application/json`, not shop HTML. Ingress must not let `/` steal `/api/v1/health`.
- **Observed:** 200 HTML homepage.

### FIX-050 — Ingress routes every spec API prefix to the owning service

- **Surface:** gateway (`k8s/api-gateway-staging.yaml` and production equivalent)
- **Expected:** `/api/v1/recommendations`, `/api/v1/settings`, `/api/v1/health` (and any other spec path) hit Go services. Response is JSON or a service 404, never MFE `index.html`.
- **Observed:** recommendations and health served by shop-mfe.

### FIX-051 — Follow / unfollow and counts

- **Surface:** community-service
- **Spec:** §4.7
- **Expected:** authenticated follow 200/201; second follow is idempotent; unfollow removes; public profile `follower_count` / `following_count` update. Unauthenticated follow 401 (already true — keep test).
- **Observed:** follow 200; profile counts were 0 with stub user — retest after FIX-033.

### FIX-052 — Post like unlike and comment delete own only

- **Surface:** community-service
- **Expected:** like toggle; unlike; comment delete 403 for another user’s comment; 200 for own.
- **Observed:** like/comment create worked. Cover unlike and ownership in tests.

### FIX-053 — Media upload and preview generation

- **Surface:** media-service, product-service
- **Spec:** §5.7
- **Expected:** admin upload stores a file; generate preview produces watermarked/thumbnail records; buyers see preview URLs, not the full asset, until paid.
- **Observed:** handlers exist; not verified on staging (no real asset pipeline in the probe). Tests should use a temp dir, not require k8s PV.

### FIX-054 — User delete cascade without cross-DB joins

- **Surface:** identity-service + events + community/commerce
- **Spec:** architecture §5.2
- **Expected:** deleting a user via admin invalidates sessions, anonymizes or deletes community posts per documented rule, and does not leave cart rows owned by a missing user in a joinable state. Implement via events or explicit service calls.
- **Observed:** `events` publisher is never called from handlers.

### FIX-055 — Product delete removes cart/wishlist entries via event or API

- **Surface:** product-service, commerce-service
- **Expected:** after product delete/archive, cart and wishlist no longer return that product as purchasable.
- **Observed:** no event emission; commerce holds raw product ids.

### FIX-056 — Admin users delete and password reset hit identity-service

- **Surface:** admin-service, identity-service
- **Expected:** `DELETE /api/v1/admin/users/:id` and `POST .../reset-password` change identity data. Subsequent login behaves accordingly.
- **Observed:** identity already has these routes; admin also implements users against a second DB. Collapse to one path.

### FIX-057 — Exchange rates: public read, admin write, one store

- **Surface:** payment-service, admin-service
- **Expected:** `GET /api/v1/exchange-rates` matches `GET /api/v1/admin/exchange-rates`. Admin PUT/DELETE changes what public GET returns.
- **Observed:** both returned rates, but admin and payment may be separate tables — test write-then-read through **one** owner (payment-service).

### FIX-058 — Sitemap and robots are not the shop HTML

- **Surface:** gateway or a small handler (product-service may generate sitemap)
- **Spec:** §5.8
- **Expected:** `GET /sitemap.xml` is XML urlset (products, categories, bundles, posts). `GET /robots.txt` is a robots file. Neither is `<title>Pawradise — Digital Assets Marketplace</title>`.
- **Observed:** both returned the homepage HTML 200.

### FIX-059 — Product page SEO tags

- **Surface:** product-mfe (and/or server-side inject)
- **Spec:** §5.8 unique title, description, OG, JSON-LD Product
- **Expected:** `GET /product/:slug` HTML contains `og:title`, `og:image`, JSON-LD Product, canonical. Playwright or HTTP body assertion.
- **Observed:** no `og:title` / JSON-LD on product-1 HTML.

### FIX-060 — Social share controls on product detail

- **Surface:** product-mfe
- **Spec:** §2.1 share to X, Instagram, Reddit
- **Expected:** share links or buttons present for those networks (Instagram may be copy-link if no official share URL).
- **Observed:** no share markup.

### FIX-061 — Bundle and request routes do not 403 / nginx-welcome

- **Surface:** gateway, product-mfe
- **Spec:** pages `/bundle/:id` (spec) and current `/product/bundle.html`
- **Expected:** bundle page is the product-mfe bundle view, not shop fallback. Spec URL `/bundle/:id` should work or redirect to the implemented path — pick one, test both if you keep a redirect.
- **Observed:** `/product/bundle.html` 200 in 2026-09-17 probe. `/admin/` trailing slash 403. Add tests for `/admin` and `/admin/`.

### FIX-062 — `/post/:id` and `/profile/:id` render the right MFE page

- **Surface:** community-mfe, gateway
- **Spec:** §3.1 `/post/:id`, `/profile/:id`
- **Expected:** HTML is post-detail or public-profile, not the feed shell with the same title as `/community`. Alpine should load that id from the path.
- **Observed:** `/post/1` and `/profile/1` returned title “Pawradise — Community” and the same ~12kB feed shell.

### FIX-063 — Unknown marketing URLs should 404, not the homepage

- **Surface:** gateway, shop-mfe nginx `try_files`
- **Expected:** `/help`, `/terms`, `/privacy` either exist as real pages (spec footer) or return 404. They must not silently serve `/`.
- **Observed:** all returned the shop homepage 200.

### FIX-064 — Custom product-request is out of spec or must have an API

- **Surface:** product-mfe, gateway, product-service (only if you keep the feature)
- **Expected:** either remove `/product/request.html` from nav, **or** add `POST /api/v1/product-requests` that 201s. Do not leave a form that 405s on nginx.
- **Observed:** page exists; POST `/api/v1/product-requests` → nginx 405.

### FIX-065 — Admin UI includes Users and email export

- **Surface:** frontend/admin-app
- **Spec:** §3.3 admin tabs include Users and Email Export
- **Expected:** users list/delete/reset and export action call the admin APIs. Token connect still required.
- **Observed:** HTML tabs: stats, products, bundles, coupons, orders, referrals, community, exchange, settings. No Users, no export.

### FIX-066 — Order status workflow validation

- **Surface:** commerce-service, admin-service
- **Spec:** §6.5 pending → paid → completed; cancelled / refunded / failed with legal transitions
- **Expected:** illegal transition → 400. Legal transition persisted and visible on GET.
- **Observed:** admin status handler exists; untestable while orders are not readable (FIX-021 / FIX-045).

### FIX-067 — PWYW checkout uses the buyer’s price

- **Surface:** commerce-service, product-service
- **Spec:** §2.3 price ≥ minimum
- **Expected:** below minimum → 400. At/above minimum, order `total_usd` uses the submitted price for that line.
- **Observed:** not exercised; create-order ignores PWYW and zeroed totals.

### FIX-068 — Health on each service still works behind the gateway

- **Surface:** each `services/*/main.go`, gateway
- **Expected:** `GET /health` on the pod JSON-ok. Gateway may expose `/api/v1/health` (FIX-049) without breaking service-local `/health`.
- **Observed:** service mains register `/health`; public `/health` is shop HTML because ingress `/` wins.

---

### FIX-069 — `/admin/` trailing slash serves the admin app

- **Surface:** gateway, frontend/admin-app nginx
- **Expected:** `GET /admin` and `GET /admin/` both return the admin HTML (or one 301s to the other). Neither is 403.
- **Observed:** `/admin` 200; `/admin/` 403.

### FIX-070 — Register sets session cookie and returns JWT

- **Surface:** identity-service
- **Spec:** §4.2
- **Expected:** `POST /api/v1/register` 201 with `{ token, user }`. `Set-Cookie` includes `pawradise_session` (HttpOnly). Duplicate email → 409. Missing/invalid email or short password → 400.
- **Observed:** register works; keep regression tests for cookie name, duplicate email, and validation. Do not drop the cookie when fixing other auth bugs.

### FIX-071 — Login rejects bad credentials and issues a new token

- **Surface:** identity-service
- **Expected:** wrong password → 401, no cookie. Unknown email → 401 (do not leak whether the email exists). Correct password → 200 `{ token, user }` + session cookie.
- **Observed:** login works for happy path; add the failure cases so they stay 401.

### FIX-072 — `GET /api/v1/me` is the current user

- **Surface:** identity-service
- **Expected:** with JWT, 200 public+private-safe fields (id, email, name, no password_hash). Without JWT → 401.
- **Observed:** `/me` exists and 401s after logout. Add a test that login then `/me` returns the same user id/email.

### FIX-073 — Profile update persists

- **Surface:** identity-service
- **Spec:** §5.4 `user_profiles`
- **Expected:** authenticated `PUT` (spec path `/api/v1/profile` or `/users/:id` — pick the one in PRODUCT-SPEC / current client, document it, test it) updates name/bio/avatar_url/wallet_address. Subsequent GET shows them. Cannot update another user’s profile (403).
- **Observed:** not fully exercised in the 2026-09-17 probe.

### FIX-074 — Logout clears the session cookie

- **Surface:** identity-service
- **Blocked by:** FIX-034
- **Expected:** `POST /logout` `Set-Cookie` expires `pawradise_session`. Browser with only that cookie then gets 401 on `/me`.
- **Observed:** token blocklist works for `/me`; cookie-clear not asserted.

### FIX-075 — Inactive / non-existent products are not publicly listed or sold

- **Surface:** product-service, commerce-service
- **Expected:** `status != active` products omitted from `GET /products`. `GET /products/:slug` for missing or inactive → 404. `POST /orders` with a missing or inactive product_id → 400.
- **Observed:** catalog is all seeded active “Product N”; create-order did not validate catalog (total 0).

### FIX-076 — Product slugs are unique

- **Surface:** product-service, admin-service
- **Expected:** creating a second product with an existing slug → 409/400. Public GET by slug returns exactly one product.
- **Observed:** admin create wrote a different DB (FIX-046). After that is fixed, enforce uniqueness in product-service.

### FIX-077 — Pagination rejects nonsense pages

- **Surface:** product-service, community-service
- **Expected:** `per_page` capped (e.g. 100). `page=0` or negative treated as 1 or 400 (pick one and test). Page beyond last returns empty array and the same `total`, not an error.
- **Observed:** untested; list always returned page 1 of 300.

### FIX-078 — Cart item update and delete

- **Surface:** commerce-service
- **Spec:** §4.8
- **Expected:** `PUT /cart/items/:id` changes quantity (≥1). quantity 0 or `DELETE` removes the line. GET no longer includes it. Other user’s item id → 404/403.
- **Observed:** add/get worked; update/delete not asserted.

### FIX-079 — Order lines snapshot title, slug, and unit price from product-service

- **Surface:** commerce-service
- **Blocked by:** FIX-020
- **Expected:** `order_items` store `product_title`, `product_slug`, `unit_price_usd` copied at create time from product-service. Later product price edits do not change historical orders.
- **Observed:** totals were 0; snapshots not verified.

### FIX-080 — Coupon expiry, usage limit, and min purchase

- **Surface:** commerce-service
- **Blocked by:** FIX-026
- **Spec:** §2.3 / FEATURE-CANDIDATES (accepted)
- **Expected:** expired code → 400. `times_used >= usage_limit` → 400. Cart below `min_purchase_usd` → 400. Valid code once increments `times_used`.
- **Observed:** validate could not even find admin-created codes.

### FIX-081 — Product-scoped coupon only discounts that product

- **Surface:** commerce-service
- **Blocked by:** FIX-023, FIX-026
- **Expected:** coupon with `product_id` set only reduces that line; other lines full price. `product_id` null is cart-wide.
- **Observed:** not applied (totals 0).

### FIX-082 — Review rating must be 1–5 integer

- **Surface:** review-service
- **Expected:** rating 0, 6, or non-integer → 400. No file/media fields accepted.
- **Observed:** create succeeded without purchase (FIX-039); bounds not asserted.

### FIX-083 — Admin routes require ADMIN_TOKEN

- **Surface:** admin-service
- **Expected:** missing/wrong `Authorization: Bearer` on `/api/v1/admin/*` → 401. Correct token from env (no hardcoded fallback in staging/prod — FIX-048) → handler runs.
- **Observed:** admin APIs responded with the staging token; unauthenticated case not systematically tested.

### FIX-084 — Admin referrals list is identity data

- **Surface:** admin-service, identity-service
- **Spec:** §4.12
- **Expected:** `GET /api/v1/admin/referrals` shows referral links/commissions from identity-service, including a referral created in FIX-031.
- **Observed:** admin referrals not checked against identity after register-with-code failed.

### FIX-085 — Admin settings write is visible on public GET by key

- **Surface:** admin-service, payment-service
- **Blocked by:** FIX-037
- **Expected:** `PUT /api/v1/admin/settings/:key` then `GET /api/v1/settings/:key` returns the new value. Unknown public keys still 404.
- **Observed:** public GET by key 404; admin list has keys.

### FIX-086 — Exchange rate by chain

- **Surface:** payment-service
- **Blocked by:** FIX-057
- **Spec:** §4.10
- **Expected:** `GET /api/v1/exchange-rates/:chain` returns that chain. Unknown chain 404. Public GET list includes the same rows as admin.
- **Observed:** list endpoints existed; write-then-read-by-chain not proven on one store.

### FIX-087 — Guest-order list is not a public dump

- **Surface:** commerce-service, gateway
- **Expected:** `GET /api/v1/guest-orders` without id is 404 or 401, never a JSON list of guests. Only `GET /guest-orders/:id` with email check (FIX-024).
- **Observed:** unauthenticated GET list returned `{"orders":[]}`.

### FIX-088 — Follow-self is rejected

- **Surface:** community-service
- **Expected:** `POST /community/follow/:ownUserId` → 400. Unfollow of someone you do not follow is idempotent 200 or 404 (pick one, test it).
- **Observed:** untested.

### FIX-089 — Delete own post; cannot delete others

- **Surface:** community-service
- **Expected:** author `DELETE /community/posts/:id` removes it from feed. Other user → 403. Admin delete is FIX-042.
- **Observed:** create worked; ownership delete not asserted.

### FIX-090 — Missing post / missing product JSON 404

- **Surface:** community-service, product-service
- **Expected:** `GET /api/v1/community/posts/999999999` and `GET /api/v1/products/no-such-slug` return JSON 404 (`Content-Type: application/json`), not shop HTML.
- **Observed:** product unknown slug not checked behind catch-all; recommendations already HTML (FIX-014).

### FIX-091 — Commerce schema has no `products` / `users` copies

- **Surface:** commerce-service
- **Expected:** commerce migrations contain orders, cart, wishlist, coupons, compare, recently_viewed — not a `products` catalog table. Prices at checkout are fetched from product-service.
- **Observed:** leftover physical-goods fields on products; review-service updates a local `products` table (FIX-039). Assert commerce/review/admin do not migrate a second catalog.

### FIX-092 — Admin schema has no copied identity/product/order tables used as source of truth

- **Surface:** admin-service
- **Blocked by:** FIX-043, FIX-045, FIX-046
- **Expected:** admin-service may cache nothing required for correctness. Tests that insert a user only in identity-service still see them via admin users API (HTTP to identity).
- **Observed:** admin copies tables; stats/users/orders disagree with owners.

### FIX-093 — Compare list has a small cap

- **Surface:** commerce-service
- **Expected:** toggling more than a documented max (4 is fine) either rejects or drops the oldest. GET never returns `product_id: 0` (FIX-029).
- **Observed:** `product_id: 0`; no cap test.

### FIX-094 — Recently viewed capped and newest-first

- **Surface:** commerce-service
- **Blocked by:** FIX-030
- **Expected:** recording the same product twice does not duplicate; newest first; list length ≤ documented cap (e.g. 20).
- **Observed:** `product_id: 0`.

### FIX-095 — Media upload requires admin auth and stores bytes

- **Surface:** media-service
- **Blocked by:** FIX-053
- **Expected:** unauthenticated upload → 401. Authenticated admin upload of a small file → 201 with path; GET file bytes match. Invalid/empty body → 400.
- **Observed:** not verified on staging.

### FIX-096 — JWT with bad signature is 401

- **Surface:** services/shared/middleware
- **Expected:** `Authorization: Bearer` garbage or HMAC with wrong secret → 401 on `/me` and `/cart`. Empty header → 401 on those routes (FIX-047).
- **Observed:** missing-token falls through `c.Next()`.

### FIX-097 — Register/login JSON errors use a consistent envelope

- **Surface:** identity-service
- **Expected:** 400/401/409 bodies are JSON `{ "error": "..." }` (or the envelope already used elsewhere — one shape). Never HTML.
- **Observed:** mixed handler messages; not standardized.

### FIX-098 — `shared/lib/api.js` paths match the spec after each backend fix

- **Surface:** shared/lib/api.js
- **Expected:** client methods used by MFEs call the same paths/fields the tests assert (search query names, order create body, download URL). After a FIX that changes a response key, update this file in the same change.
- **Observed:** TESTING-GUIDE notes `api.get('/products')` vs `api.products.list()` mismatches historically.

### FIX-099 — Shop search UI sends `search` to the API

- **Surface:** frontend/shop-mfe
- **Blocked by:** FIX-001
- **Expected:** submitting the shop search box performs a request whose query string includes `search=` (Playwright). Results on the page match the filtered API `total`.
- **Observed:** backend ignored `search`; UI wiring not proven.

### FIX-100 — Checkout UI cannot “succeed” on a zero-total stub

- **Surface:** frontend/checkout-mfe
- **Blocked by:** FIX-020, FIX-021, FIX-025
- **Expected:** Playwright: add a priced product, checkout, see a non-zero total, and after pay a download control that is not the JSON stub. Do not mark paid in the UI if `GET /orders/:id` is 404.
- **Observed:** API create returned `total_usd: 0` and GET 404.

---

When the last `### FIX-` section is gone, add a single line under “How to empty this file”: `Queue emptied YYYY-MM-DD` and stop.
