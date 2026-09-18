# Release 16 — Frontend routing fix, auth UX, admin expansion, 40 e2e scenarios

Date: 2026-09-04. User review showed the site was far from complete:
no login page, COMMUNITY button re-served the shop page, staging links
leaked to production, no product detail pages, and no UI path created an order.

## Bugs fixed (root causes, not symptoms)

- BUG-1 nginx had no locations for /community /account /checkout /login —
  all fell through to index.html. Fix: one explicit location per page,
  both envs (`/etc/nginx/sites-available/pawradise`, backup at `.bak`).
- BUG-2 staging block used `root` instead of per-file roots, so /staging/*
  fell back to production files. Fix: explicit staging locations.
- BUG-3 no login page existed. Fix: new `login.html` (login+register tabs,
  token storage, safe same-env `?redirect=` handling).
- BUG-4 nav links hardcoded to production paths, no login/logout affordance.
  Fix: shared header (SHOP/COMMUNITY/ACCOUNT + auth-aware LOGIN/LOGOUT)
  with `LINK_BASE` prefix on index/community/account/checkout/login/product.
- BUG-5 index linked to /product/:slug with no detail page. Fix: new
  `product.html` (gallery, price, BUY NOW creates order -> checkout).
- Dead BUY flow: no page ever POSTed /orders. Product page now creates the
  order and redirects to checkout?order=ID.
- Backend panic class: 18x `userID.(float64)` casts on a context value the
  middleware stores as int (12 community.go, 6 orders.go) -> 500s. Fixed.
- Orders group had no `JwtAuthMiddleware` -> every order call 401. Fixed.
- Products with NULL category broke scans (`c.name` NULL into string):
  silent zero structs on create/update, 500 on detail. Fixed with
  COALESCE(c.name,'') in all 4 product queries.
- Logout was a no-op (stateless JWT stayed valid 24h). Fix: migration 009
  `invalidated_tokens` blocklist, middleware revocation check, real Logout.
- Seller wallet was a compile-time const. Fix: migration 008 `settings`
  table + admin GET/PUT /settings + orders read live value per order.
- New admin endpoints: GET /admin/orders, GET /admin/orders/:id.
- Admin panel expanded: Users + Products (CRUD, stats) + Orders (list) +
  Rates & Settings tabs.

## Migrations (applied to appdb_staging AND appdb_production)

- 008_create_settings.sql — settings KV, seeded payment_address placeholder
- 009_invalidate_tokens.sql — JWT blocklist + expiry index

## E2E (curl-based, BLUE -> RED)

Current pass/fail counts live in [`../tests/REPORT.md`](../tests/REPORT.md). Historical customer-journeys.sh / admin-ops.sh scripts are not the live suite.

- e2e/common.sh — shared harness (HTTP_CODE/BODY globals, assertions)
- e2e/customer-journeys.sh — C01..C20: register/login/profile, browse,
  search, categories, detail, detail-404, order auth, create order + payment
  info, order detail, my orders, payment info, status poll, unpaid download
  blocked (403), feed, post/like/comment/profile/follow/unfollow, logout
  invalidates token. Result: 40/40 staging, 40/40 production.
- e2e/admin-ops.sh — A01..A20: auth enforcement, users, reset+login,
  delete, product CRUD + visibility + stats + invalid rejection, orders
  list/detail/404, rates get/set/restore/invalid rejection, settings
  round-trip incl. new-order-uses-new-address, admin pages serve.
  Result: 37/37 staging, 37/37 production.
- e2e/link-audit.sh — no staging page links to production paths and vice
  versa, product pages serve both envs, nav essentials present.
  Result: 14/14.
- Test residue cleaned from both envs (test users deleted, posts cascaded).

## Verification

- 14/14 page routes return 200 with correct titles (7 prod + 7 staging).
- Backend pods ready in staging + production; nginx -t clean.
- Admin token and secrets redacted from all reports ([REDACTED] convention).
