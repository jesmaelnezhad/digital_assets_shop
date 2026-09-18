# Pawradise Release Note 18 — Frontend Feature Completion + Demo Data Expansion

**Date:** 2026-09-04
**Release:** 18
**Scope:** Frontend (new pages + checkout auto-poll) + staging DB expansion
**Env affected:** Staging (demo data), Production (frontend only — same pages deployed)

---

## Summary

Added the three most visible missing frontend features and expanded the staging database with significantly more demo data to support a convincing demonstration.

## Frontend changes

### New pages deployed to both production + staging
- **`/profile.html`** — Public user profile page: avatar (placeholder or URL), display name, email handle, bio, join date, BSC wallet address (if connected), follower/following/post count stats, follow/unfollow button (logged-in users only, with login prompt for guests), and a list of the user's recent community posts with like/comment counts. Navigates to `/profile/:id` from post authors and follow buttons.
- **`/post.html`** — Dedicated post detail page: full post content in a highlighted terminal-style card, author info with link to their profile, like toggle button (with live count update), comment count + expandable comment section, add-comment form, delete-own-comment button, follow/unfollow author button. Navigates to `/post/:id` from feed links (future enrichment) and can be reached directly via URL.

### Checkout enhancement
- **`/checkout.html`** — Added auto-polling: when a pending order loads, the page starts polling `GET /api/v1/orders/:id/status` every 15 seconds. A visible "auto-polling every 15s — waiting for blockchain confirmation..." indicator with spinner appears. When the backend reports `status: paid`, the page shows "PAYMENT CONFIRMED!" with a truncated tx hash and auto-reloads after 3 seconds to display download buttons. Manual "CHECK PAYMENT STATUS" button still works alongside auto-poll.

### Bug fixes
- Fixed `timeAgo()` JS runtime error (`nowthen` → `now - then`) in both `profile.html` and `post.html` — would have crashed all relative-time rendering.
- Fixed async `myId` resolution in `profile.html` — was using a Promise synchronously (`myIdRes.id` on a pending Promise → `undefined`), causing follow button logic to always treat the viewer as a guest.
- Fixed follow endpoint URL in both `profile.html` and `post.html` — was calling `/community/users/:id/follow` (404); corrected to `/community/follow/:id` matching the actual backend route.
- Fixed comment delete ownership check in `post.html` — was calling `getMyUserId()` synchronously (returns Promise) inside a `map()` template; replaced with a cached `myUserId` variable set via `initMyUserId()` after login.

### nginx
- Added explicit per-page locations for `/profile/` and `/post/` (and `/staging/profile/`, `/staging/post/`) to the host nginx config, matching the existing pattern for shop/community/account/checkout/login/product/admin. Both environments serve the new pages correctly.

## Staging database expansion

Applied refined demo seed scripts (`backend/seeds/staging_demo_part1.sql`, `backend/seeds/staging_demo_part2.sql`) to `appdb_staging` only (production untouched). Cleanup + reseed ensures stable IDs and consistent demo state for the e2e suites.

**Post-seed counts (staging):**
| Table | Records | Notes |
|-------|---------|-------|
| users | 25 | 16 @pawradise.demo addresses (demo1234 password), 9 pre-existing/functional test users |
| products | 32 | Across 8 categories: Pixel Icons, JS Tools, UI Kits, Fonts, DevOps Tools, API Clients, Terminal Utils, Code Snippets |
| categories | 8 | |
| community_posts | 62 | By 17 distinct authors, seeded with realistic dev/shop/crypto content, timestamps spread over 10 days |
| post_likes | 182 | ~3 per post, deduplicated |
| post_comments | 32 | Hand-written realistic comments across popular posts |
| follows | 48 | Each demo user follows 3 others cyclically |
| orders | 12 | Mix of pending (4) and paid (8) with BSC-style txids, confirmations 15-25 on paid, realistic total_usd values |
| order_items | 15 | Linked to orders + products, download counts tracking |
| product_images | 30 | One primary image per product, placehold.co placeholders |
| user_profiles | 18 | Bios, avatar URLs, wallet addresses for demo users |
| exchange_rates | 1 | BNB/USD rate |
| settings | 1 | payment_address = demo seller wallet |
| invalidated_tokens | 1 | From e2e logout tests |

## Verification

Current pass/fail counts live in [`../tests/REPORT.md`](../tests/REPORT.md).

### Frontend page HTTP status (staging + production)
All 9 pages return 200 with correct titles on both environments:
`/`, `/community`, `/account`, `/checkout`, `/login`, `/product/pixel-icon-pack-500`, `/profile/10`, `/post/10`, `/admin`

### Database audit (staging)
```text
  users                   25
  products                32
  categories               8
  community_posts         62
  post_likes             182
  post_comments           32
  follows                 48
  orders                  12
  order_items             15
  product_images         30
  user_profiles           18
  exchange_rates           1
  settings                 1
  invalidated_tokens       1
```
14 tables total. Production database NOT modified.

## Known remaining gaps (Phase 6, still open)
- Wallet connect UI (MetaMask-style connect + address verify + save to profile) — account page has disconnect button and wallet display but no connect flow
- Cart (localStorage, add/remove/qty, cart page/drawer, checkout from cart) — currently single-product direct-to-order flow only
- Admin UI completion: no product image upload, no community post moderation, no order status manual override

## Files changed
- `frontend/public/profile.html` (new — 342 lines)
- `frontend/public/post.html` (new — 617 lines)
- `frontend/public/checkout.html` (modified — added auto-poll)
- `backend/seeds/staging_demo_part1.sql` (new — users, profiles, products, images)
- `backend/seeds/staging_demo_part2.sql` (new — posts, likes, comments, follows, orders, items)
- `PLAN-ecommerce.md` (updated Phase 6 status)
- nginx config on RED (added /profile/ and /post/ locations)

## Deployment
- frontend files synced to RED `/var/www/production/` and `/var/www/staging/`
- nginx reloaded on RED
- No backend rebuild needed (no backend code changes in this release)
