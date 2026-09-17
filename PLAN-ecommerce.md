# Pawradise E-Commerce Platform — Design & Implementation Plan

**Status:** DRAFT — v1.0  
**Date:** 2026-09-04  
**Author:** Hermes Agent  
**Goal:** Build a cozy online shop for digital assets with crypto payments, community features, and a nerdish/futuristic aesthetic. Running on the existing RED/BLUE infrastructure.

---

## 1. Vision & Product Definition

**What it is:** A digital download shop where one seller (us) sells digital assets (files, packs, licenses, etc.) to visitors. Some visitors become buyers. Everyone can participate in a community center — chat, short posts, likes, comments, follows.

**Vibe:** Nerdish, programming-oriented, futuristic but cozy. Think: terminal aesthetics, monospace where appropriate, dark theme with accent colors, subtle glow effects, no heavy graphics. Placeholder images are fine — substance over style.

**Core user journeys:**
1. **Visitor** → browses shop → searches/browses categories → views product detail → pays with crypto wallet → downloads asset
2. **Visitor** → joins community → posts short text → likes/comments on others' posts → follows people
3. **Buyer** → manages their purchases → re-downloads assets
4. **Admin (us)** → adds/edits products → views sales → manages community (moderation)

---

## 2. Architecture Overview

### 2.1 High-Level Architecture

```
RED (130.185.123.156) — Serving
├── k3s cluster
│   ├── database namespace
│   │   └── PostgreSQL 16 (appdb_pawradise — single DB for everything)
│   ├── production namespace
│   │   ├── pawradise-backend (Go + Gin, port 8080, NodePort 30093)
│   │   └── pawradise-frontend (nginx:alpine, port 80, NodePort 30094)
│   └── staging namespace (identical, path-prefixed)
├── host nginx (port 80) — path-based routing
│   ├── / → production frontend
│   ├── /staging/ → staging frontend
│   ├── /api/v1/ → production backend
│   ├── /staging/api/v1/ → staging backend
│   ├── /admin → production admin UI
│   └── /staging/admin → staging admin UI
└── ufw: 22, 80, 443, 30081-30084, 30093-30094

BLUE (130.185.121.83) — Build/Dev/Admin
├── Go build environment (golang:1.22-alpine Docker image)
├── Docker for building frontend images
├── docker-compose.yml for dev PostgreSQL (optional, for local testing)
├── /root/project/ — source code
├── /root/.kube/red-kubeconfig.yaml — kubectl against RED
├── Helper scripts: red-run.sh, red-status.sh, red-kubectl.sh
└── Agent session lives here
```

### 2.2 Tech Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Backend | Go + Gin framework | Already in use, fast, single binary, low RAM |
| Frontend | Vanilla JS + HTML (no React build pipeline) | Already in use, lightweight, no npm overhead |
| Database | PostgreSQL 16 | Already running on RED, handles relational data well |
| Auth | JWT + bcrypt (existing) + crypto wallet auth (new) |
| Payments | Crypto wallet connection (WalletConnect / direct) + on-chain verification |
| Assets storage | PersistentVolume on k3s (local-path-provisioner) + backend serves files |
| CSS | Custom CSS (no Tailwind/build step) — dark theme, monospace accents |
| Admin | Existing admin panel extended |

### 2.3 Database Schema (Core Tables)

```
users                    — existing (id, email, password_hash, name, created_at)
user_profiles            — (user_id, bio, avatar_url, wallet_address, created_at)
products                 — (id, title, description, category, price_usd, 
                           asset_path, asset_hash, status, created_at, updated_at)
product_images           — (id, product_id, url, is_primary)
categories               — (id, name, slug, description, parent_id)
orders                   — (id, user_id, total_usd, status, payment_tx_hash, 
                           payment_chain, payment_confirmations, created_at)
order_items              — (id, order_id, product_id, quantity, price_at_purchase,
                           download_count, max_downloads)
downloads                — (id, order_item_id, user_id, downloaded_at, ip_address)
community_posts          — (id, user_id, content, created_at)
post_likes               — (id, user_id, post_id, created_at) — UNIQUE(user_id, post_id)
post_comments            — (id, user_id, post_id, content, created_at)
follows                  — (id, follower_id, following_id, created_at) — UNIQUE pair
admin_logs               — (id, admin_user_id, action, details, created_at)
```

### 2.4 API Design

All under `/api/v1/` prefix (consistent with existing structure).

**Auth (existing + extended):**
- `POST /api/v1/register` — email/password registration
- `POST /api/v1/login` — email/password login → JWT
- `POST /api/v1/logout` — invalidate JWT
- `GET /api/v1/me` — current user profile
- `PUT /api/v1/me` — update profile (bio, avatar)
- `POST /api/v1/me/wallet` — connect crypto wallet (address → saved to profile)

**Products (public, no auth needed for browsing):**
- `GET /api/v1/products` — list with pagination, search, category filter
- `GET /api/v1/products/:id` — product detail
- `GET /api/v1/categories` — list categories

**Shop (auth required for purchasing):**
- `POST /api/v1/orders` — create order (product IDs + quantities)
- `GET /api/v1/orders` — user's order history
- `GET /api/v1/orders/:id` — order detail with download links
- `POST /api/v1/orders/:id/pay` — initiate crypto payment (returns payment info)
- `GET /api/v1/orders/:id/status` — check payment status on-chain
- `GET /api/v1/orders/:id/download/:itemId` — download asset (after payment verified)

**Community (auth required for posting):**
- `GET /api/v1/community/posts` — feed (paginated, optional filter: trending, recent, following)
- `POST /api/v1/community/posts` — create post (max length: ~500 chars)
- `GET /api/v1/community/posts/:id` — single post with comments
- `POST /api/v1/community/posts/:id/like` — like/unlike toggle
- `DELETE /api/v1/community/posts/:id/like` — remove like
- `POST /api/v1/community/posts/:id/comments` — add comment
- `DELETE /api/v1/community/posts/:id/comments/:commentId` — delete own comment
- `POST /api/v1/community/follow/:userId` — follow user
- `DELETE /api/v1/community/follow/:userId` — unfollow
- `GET /api/v1/community/users/:id` — public profile (posts, followers, following count)

**Admin (admin token auth):**
- `GET /api/v1/admin/products` — list all products
- `POST /api/v1/admin/products` — create product
- `PUT /api/v1/admin/products/:id` — update product
- `DELETE /api/v1/admin/products/:id` — delete/draft product
- `GET /api/v1/admin/orders` — list all orders
- `GET /api/v1/admin/users` — list users (existing)
- `GET /api/v1/admin/community/posts` — list community posts (moderation)
- `DELETE /api/v1/admin/community/posts/:id` — remove post (moderation)
- `GET /api/v1/admin/stats` — dashboard stats (sales, users, revenue)

---

## 3. Feature Breakdown

### 3.1 Shop (Core E-Commerce)

**Product listing page (`/`)**
- Grid of product cards with: thumbnail, title, price, short description, category badge
- Search bar (text search across title + description)
- Category filter sidebar/tabs
- Sort: newest, price low-high, price high-low
- Pagination (12 per page)

**Product detail page (`/product/:id` or `/products/:slug`)**
- Large image, title, full description, category
- Price in USD (static reference price)
- "Connect Wallet & Buy" button
- Seller info (just "Pawradise" for now)
- Related products sidebar
- Download info: file type, size, number of downloads allowed

**Cart / Checkout flow:**
- Simple: no persistent cart needed for digital assets
- User selects product(s) → clicks buy → sees order summary → connects wallet → confirms payment
- After payment confirmed on-chain → download links appear

### 3.2 Crypto Payment

**Wallet connection:**
- User clicks "Connect Wallet" → wallet extension (MetaMask, etc.) prompts
- On connect: wallet address saved to user profile
- Multiple chain support: start with one chain (e.g., Ethereum/BSC/Arbitrum — to be decided)

**Payment flow:**
1. User adds items to order → backend creates order with total in USD + equivalent in crypto (via oracle/price feed or manual rate)
2. Backend returns: destination address, amount to send, expected chain, order ID
3. User sends transaction from their wallet to the destination address with order ID in memo/data
4. Backend monitors the chain (polling or webhook) for the transaction
5. After N confirmations → order status = "paid" → download links activated
6. User can see order status on `/my-purchases` page

**Key design decisions:**
- HOT wallet on backend for receiving payments (small balance, swept periodically)
- Or: user sends to their own deposit address and we credit their account (more complex)
- Start with: simple — we give them our wallet address, they send, we verify on-chain, we enable download
- Price feed: use a simple exchange rate table updated daily (manual or via API call at startup)

### 3.3 Community Center

**Feed page (`/community` or `/feed`)**
- Posts from people the user follows + trending posts
- Each post: author avatar, name, timestamp, content, like count, comment count
- Like button (toggle), comment button → expands comments
- "Follow" button on user profiles

**Post creation:**
- Text area (short form, max ~500 chars — think Twitter-style)
- Submit → appears in feed
- Edit own posts (optional, v1 can skip)

**Comments:**
- Threaded under each post (flat list, not nested — simpler)
- Like comments too (optional)

**User profiles:**
- Public profile page: avatar, bio, follower/following counts, post count
- Follow/Unfollow button
- List of their posts

**Moderation (admin):**
- List all posts, delete inappropriate ones
- Ban users (future — skip for v1)

### 3.4 User Account

**My Account page (`/account` or `/my-account`):**
- Profile info (name, bio, avatar URL, wallet address)
- My purchases (list of orders with download links)
- Re-download assets (with download count tracking)

### 3.5 Admin Panel (existing, extended)

**Existing:** user list, delete user, reset password
**New tabs/sections:**
- Products: CRUD products (title, description, price, category, asset file upload, images)
- Orders: view all orders, see payment status, manual override (mark as paid if needed)
- Community: view/delete posts (moderation)
- Stats: simple dashboard (total users, total sales, revenue, active products)

---

## 4. UI/UX Design Direction

### 4.1 Visual Identity

**Theme:** Dark background (#0a0a0f or similar deep dark), monospace accents for technical feel, neon/synthwave accent colors (cyan #00f0ff, magenta #ff00aa, or green #00ff88) used sparingly.

**Typography:**
- Headings: a bold sans-serif (system fonts: -apple-system, 'Segoe UI', etc. — no external font loading needed)
- Body: system sans-serif for readability
- Code/technical elements: monospace (`Consolas`, 'Courier New', monospace)
- No external font downloads — keeps it fast and self-contained

**Color palette (suggested):**
- Background: #0a0a0f (near-black with slight blue)
- Surface: #14141f (cards, panels)
- Border: #2a2a3a (subtle borders)
- Text primary: #e0e0e0
- Text secondary: #8888aa
- Accent cyan: #00d4ff (buttons, links, highlights)
- Accent magenta: #ff0066 (sale badges, important actions)
- Success: #00ff88
- Error: #ff4444

**Design elements:**
- Subtle glow on interactive elements (box-shadow with accent color)
- Thin borders, no heavy shadows
- Monospace for prices, technical labels, code snippets
- ASCII-art style dividers or simple CSS borders
- Loading states: simple spinners (CSS-only)
- No images needed for UI chrome — pure CSS

### 4.2 Page Layouts

**Header (all pages):**
- Logo/wordmark "PAWRADISE" in monospace/cyber style
- Nav: Shop, Community, My Account
- Wallet status indicator (connected address or "Connect Wallet")
- Admin link (if admin)

**Footer:**
- Simple: copyright, links to shop/community/account

**Product Card:**
```
+------------------+
| [IMAGE/PLACEHOLDER] |
|                   |
|  Title            |
|  Category badge   |
|  $XX.00           |
|  [Buy with Crypto] |
+------------------+
```

**Community Post:**
```
+----------------------------------+
| @username · 2h ago               |
|                                  |
| Post content here...             |
|                                  |
| ❤ 12  💬 3  [Like] [Comment]   |
+----------------------------------+
```

### 4.3 Responsive

- Mobile-first CSS with media queries
- Single column on mobile, grid on desktop
- Touch-friendly buttons (min 44px tap targets)

---

## 5. Implementation Phases

### Phase 1: Foundation & Product Catalog (Days 1-2)

**Database:**
- Create tables: products, categories, product_images (migrations)
- Seed with a few sample categories

**Backend:**
- Product handlers: CRUD, list with search/filter, detail
- Category handlers: list
- File upload handler: accept asset file, store on PV, record path + hash
- Admin product CRUD endpoints

**Frontend pages:**
- Shop landing page (product grid with search/filter)
- Product detail page
- Category page (optional — can be filter on shop page)

**Admin:**
- Product management UI (add/edit/delete products)

**Test:** Verify product CRUD, listing, search, detail page renders.

---

### Phase 2: Orders & Crypto Payments (Days 3-4)

**Database:**
- Create tables: orders, order_items, downloads

**Backend:**
- Order handlers: create, list, detail
- Payment initiation: generate payment request (wallet address, amount, memo)
- Payment verification: poll on-chain (or use a simple block explorer API)
- Download handler: serve asset file, track download count

**Frontend pages:**
- Cart/checkout flow (simple: select → review → pay)
- Payment confirmation UI (show tx details, waiting for confirmation)
- My purchases page (order history + download buttons)

**Admin:**
- Order list with status
- Manual payment confirmation (override)

**Test:** Full purchase flow: select product → create order → "pay" (test mode) → download.

---

### Phase 3: Community Center (Days 5-6)

**Database:**
- Create tables: community_posts, post_likes, post_comments, follows, user_profiles

**Backend:**
- Post handlers: create, list (feed), detail, like, comment
- Follow handlers: follow, unfollow, follower counts
- User profile handler: public profile view

**Frontend pages:**
- Community feed (`/community`)
- Post detail with comments
- User profile page
- Follow/Unfollow UI

**Test:** Create post, like, comment, follow/unfollow, feed updates.

---

### Phase 4: Polish & Integration (Days 7-8)

**Frontend polish:**
- Consistent header/footer across all pages
- Responsive design verification
- Loading states, error states
- Empty states (no products, no posts, no purchases)

**Backend hardening:**
- Input validation on all endpoints
- Rate limiting on community posting (simple: max 1 post per 30s per user)
- SQL injection prevention (use parameterized queries — already doing this with pgx)
- Error handling: consistent JSON error responses

**Admin polish:**
- Stats dashboard
- Better admin UX

**Cross-cutting:**
- Update reference docs
- Update PLAN file with progress
- Ensure staging = production

---

## 6. Crypto Payment Deep Dive

### 6.1 Approach: Direct Wallet-to-Wallet

This is the simplest robust approach for a small shop:

1. **Backend has a hot wallet** (one address per chain) for receiving payments
2. **User flow:** Add items → see total → click "Pay" → backend returns:
   - Destination wallet address (our hot wallet)
   - Amount to send (in crypto, derived from USD price × exchange rate)
   - A unique memo/order ID to include in the transaction
3. **User** sends the transaction from their wallet (MetaMask or similar)
4. **Backend** polls the blockchain (or uses a block explorer API) to find the transaction
5. **After confirmations** → order marked paid → download enabled

### 6.2 Exchange Rate

- Maintain a simple `exchange_rates` table: `chain, symbol, rate_to_usd, updated_at`
- Updated daily (admin can manually update, or fetch from a public API at backend startup)
- For v1: manual update is fine — we set the rate once a day

### 6.3 On-Chain Verification

Options (in order of simplicity):
1. **Poll a block explorer API** (e.g., Etherscan API for Ethereum, BscScan for BSC) — simple HTTP calls, no node needed
2. **Run a light node** — too heavy for 2GB RAM
3. **Use a wallet connect SDK** — more complex, requires more JS

**Chosen: Option 1** — use public block explorer APIs for verification. We don't need to run a node. We just need to read the chain.

### 6.4 Supported Chains (v1)

Start with ONE chain to keep it simple:
- **Ethereum** — most users have MetaMask, Etherscan API is free
- Or **BSC (Binance Smart Chain)** — lower fees, also MetaMask-compatible
- Decision: start with BSC (lower fees = better for digital asset purchases). Can add ETH later.

### 6.5 Security Considerations

- Hot wallet should hold minimal funds — sweep to cold storage periodically (manual for v1)
- Transaction memo/order ID must be validated — prevent someone from sending random ETH and claiming products
- Replay protection: each order has a unique ID, used once
- Amount matching: verify the sent amount matches (within tolerance for gas price fluctuations)

---

## 7. File Storage for Digital Assets

### 7.1 Approach

- Digital assets stored on a PersistentVolume in k3s (local-path-provisioner on RED)
- Backend serves files through a download endpoint (with auth check)
- File path stored in `products.asset_path` (relative to PV mount)
- Asset integrity: store SHA-256 hash, verify on upload

### 7.2 PV Setup

- Create a PersistentVolumeClaim (e.g., `assets-pvc`, 10Gi)
- Mount to backend pod at `/assets`
- Backend download endpoint: `GET /api/v1/download/:orderItemId` → checks auth + payment → serves file

### 7.3 Alternative: Object Storage

For v1, PV is fine. If we outgrow it, we can switch to:
- MinIO (self-hosted S3-compatible) on k3s
- Or an external service (Backblaze B2, Cloudflare R2)

---

## 8. Testing Strategy

### 8.1 Backend Unit Tests

- Continue existing pattern: handlers_test.go for each handler group
- Test product CRUD, order creation, payment initiation, community actions
- Mock DB for unit tests, real DB for integration tests (when accessible)
- Target: 50%+ coverage (same as existing requirement)

### 8.2 Integration Tests (curl-based, from BLUE to RED)

- Test product flow: create product (admin) → list products → view product
- Test purchase flow: register → create order → payment → download
- Test community flow: post → like → comment → follow
- Run after each phase deployment

### 8.3 Frontend Testing

- Visual inspection (manual) — no automated frontend tests for now (per existing constraints)
- Verify all interactive elements work: buttons, forms, navigation

---

## 9. Deployment & Environment

### 9.1 Staging vs Production

- Staging and production are IDENTICAL (same deployment manifests, different namespaces, different databases)
- Develop on staging, verify, then apply same manifests to production
- Staging at `/staging/` path prefix, production at `/`
- Same admin token for both (or different — TBD)

### 9.2 Database Migrations

- Go migrations: a simple migration system that runs SQL files in order
- Migration files in `/root/project/backend/migrations/`
- Run on backend startup (or via akubectl job) — check which tables exist, apply missing ones
- Each migration is an incremental SQL file: `001_create_products.sql`, `002_create_orders.sql`, etc.

### 9.3 Asset Storage Migration

- PVC created once, persists across pod restarts
- Backend pod mounts PVC at `/assets`
- Files uploaded via admin panel → stored on PVC → served via download endpoint

---

## 10. Open Questions & Decisions Needed

1. **Which blockchain for payments?** BSC (low fees) vs Ethereum (more users) vs Polygon (low fees + Ethereum ecosystem). Recommendation: BSC for v1.
2. **Hot wallet management?** Manual sweep for v1 — admin withdraws funds periodically. Automated sweeping is v2.
3. **Price feed?** Manual daily update for v1. Could add CoinGecko API later.
4. **Asset file size limits?** Set a max upload size (e.g., 100MB per file). Larger files need different handling.
5. **Number of download attempts?** Unlimited for v1, or set a limit (e.g., 5 downloads per order item).
6. **Community post length?** 500 chars (Twitter-like) is good for v1.
7. **Avatar system?** URL-based avatars (user provides image URL) for v1. No file upload for avatars yet.
8. **Search?** PostgreSQL full-text search via `tsvector` — good enough for v1, no external search engine needed.

---

## 11. File Structure (Proposed)

```
/root/project/
├── backend/
│   ├── main.go                    # Gin router setup, all route groups
│   ├── handlers/
│   │   ├── auth.go                # register, login, logout, me, wallet connect
│   │   ├── auth_test.go           # unit tests
│   │   ├── products.go            # product CRUD, list, search, detail
│   │   ├── products_test.go
│   │   ├── orders.go              # order creation, payment, download
│   │   ├── orders_test.go
│   │   ├── community.go           # posts, likes, comments, follows
│   │   ├── community_test.go
│   │   ├── admin.go               # admin endpoints (extended)
│   │   ├── admin_test.go
│   │   ├── download.go            # asset download handler
│   │   └── health.go              # health check (existing)
│   ├── database/
│   │   ├── db.go                  # DB connection (existing)
│   │   └── migrate.go             # migration runner
│   ├── models/
│   │   ├── user.go                # existing
│   │   ├── product.go
│   │   ├── order.go
│   │   ├── community.go
│   │   └── exchange_rate.go
│   ├── migrations/
│   │   ├── 001_create_products.sql
│   │   ├── 002_create_categories.sql
│   │   ├── 003_create_orders.sql
│   │   ├── 004_create_community.sql
│   │   └── ...
│   ├── services/
│   │   ├── payment.go             # crypto payment logic
│   │   └── exchange.go            # exchange rate handling
│   ├── middleware/
│   │   └── jwt.go                 # existing JWT middleware
│   ├── uploads/                   # temp upload dir (not in image)
│   ├── Dockerfile
│   └── go.mod / go.sum
├── frontend/
│   ├── public/
│   │   ├── index.html             # shop landing (product grid)
│   │   ├── product.html           # product detail (or handled by JS routing)
│   │   ├── community.html         # community feed
│   │   ├── post.html              # single post view
│   │   ├── profile.html           # user profile
│   │   ├── account.html           # my account / purchases
│   │   ├── admin.html             # admin panel (existing, extended)
│   │   ├── login.html             # login page
│   │   ├── register.html          # register page
│   │   ├── connect-wallet.html    # wallet connection flow
│   │   ├── checkout.html          # checkout / payment flow
│   │   ├── css/
│   │   │   └── style.css          # main stylesheet
│   │   └── js/
│   │       ├── api.js             # API client (fetch wrappers)
│   │       ├── auth.js            # auth state management
│   │       ├── shop.js            # shop page logic
│   │       ├── community.js       # community page logic
│   │       ├── account.js         # account page logic
│   │       └── admin.js           # admin panel logic
│   ├── Dockerfile
│   └── ...
├── tests/
│   └── ... (existing test files)
├── docker-compose.yml             # dev PostgreSQL on BLUE
├── PLAN-ecommerce.md              # this plan file
└── ...
```

---

## 12. Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| RED RAM exhaustion (2GB) | Backend/DB pods OOM | Keep resource limits tight (128Mi backend, 100m CPU), monitor |
| Crypto payment verification lag | User waits for confirmation | Show "pending" state, poll every 30s, notify when confirmed |
| Hot wallet compromise | Loss of funds | Keep minimal balance, manual sweep, use separate wallet per environment |
| Asset file storage fills up | Can't upload new products | Set PVC size with headroom, monitor usage, clean up old files if needed |
| Community spam | Noise in feed | Rate limiting, admin moderation tools |
| Frontend complexity creep | Hard to maintain | Keep it simple — vanilla JS, one page per feature, no SPA framework |

---

## 13. Progress Tracking

This plan will be updated as we progress. Each phase completion gets a release note in `/root/release-notes/`.

**Current state (pre-ecommerce):**
- [x] RED/BLUE split complete
- [x] Auth system working (register/login/logout/profile)
- [x] Admin panel working (user management)
- [x] Staging + production environments identical
- [x] Nginx path routing on port 80

**E-commerce phases (status corrected 2026-09-04 — user review):**
- [x] Phase 1: Foundation & Product Catalog (backend + basic shop pages)
- [x] Phase 2: Orders & Crypto Payments (backend; frontend BUY->order->checkout wired, wallet connect NOT done)
- [x] Phase 3: Community Center (backend + feed page; profile pages NOT done)
- [ ] Phase 4: Polish & Integration — REOPENED. Still missing from frontend:
  wallet connect, cart, public profile pages, post detail page, account
  purchases/downloads UI, admin product/order/community management UI,
  checkout auto-poll, error/empty states. Backend for these exists except
  asset file upload/serving.
- [x] Phase 5 (2026-09-04): routing/auth/e2e hardening — DONE, release note 16
  (explicit nginx locations, login.html, shared header + LINK_BASE,
  product.html, panic fixes, real logout, settings table, admin tabs,
  40 customer + 37 admin e2e assertions passing, link audit 14/14)

**Phase 6 (added 2026-09-04): frontend feature completion + staging demo data**
- [x] Public profile pages (`/profile/:id` — avatar, bio, wallet, posts, follow/unfollow) — DONE 2026-09-04
- [x] Post detail page (`/post/:id` — single post + comments thread + follow author) — DONE 2026-09-04
- [x] Checkout auto-poll payment status (15s interval, auto-reload on confirmation) — DONE 2026-09-04
- [x] Staging demo seed: 25+ products, 15+ users, 60+ posts, likes/comments/follows, sample orders — DONE (32 products, 25 users, 62 posts, 182 likes, 32 comments, 48 follows, 12 orders, 15 order items)
- [ ] Wallet connect UI (MetaMask-style address input + verify, save to profile)
- [ ] Cart (localStorage, add/remove/qty, cart drawer or page, checkout from cart)
- [ ] Admin product/order/community management UI completion (admin.html has tabs but no image upload, no post moderation, no order status override)

---

## 14. Next Immediate Steps

1. Review and approve this plan (or request changes)
2. Decide on blockchain for payments (recommend BSC)
3. Create Phase 1 database migrations
4. Implement product handlers in backend
5. Build product listing + detail pages in frontend
6. Build admin product management UI
7. Deploy to staging, verify, then production
8. Write release notes

---

*This plan is a living document. Update it as decisions are made and features are implemented.*
