# Pawradise E-Commerce — Product Specification

> This document describes what Pawradise IS and what it SHOULD HAVE, independent
> of what is currently working or passing tests. It defines product specs, user
> flows, pages, routes, modules, architecture, and services.
>
> **Version: 1.2** — Architecture sections aligned with the microservice/MFE
> repo. Product behavior is unchanged from 1.1.

---

## 1. Product Vision & Scope

### 1.1 What Is Pawradise

Pawradise is a single-seller digital asset marketplace — an online shop where one
vendor ("Pawradise") sells digital download products (files, packs, templates,
code, 3D models, audio, graphics, etc.) to visitors who can optionally become
buyers. Alongside the shop, there is a community center where users can post
short text updates, like and comment on posts, and follow other users.

- **Vendor:** Single seller ("Pawradise"). No multi-vendor marketplace.
- **Customer:** Visitors who browse, some of whom purchase and download.
- **Platform operator:** The admin — manages products, orders, community
  moderation, exchange rates, settings.

### 1.2 Product Type

Digital goods ecommerce store + community. Core transaction model: user pays
(crypto wallet or simplified pay flow), receives a download grant for the
purchased item, can re-download indefinitely as long as they have the purchase.

### 1.3 Licensing Model

Simple download model: **buy once, download forever**. Each purchase grants the
buyer unlimited downloads of the purchased product. No license keys, no
activations, no device limits. The asset is yours once purchased. Products may
be offered in multiple tiers (different file versions / formats at different
price points).

### 1.4 Design Direction

- Dark theme background (near-black with slight blue tint)
- Monospace accents for prices, technical labels, code elements
- Neon/synthwave accent colors (cyan, magenta) used sparingly
- CSS glow effects on interactive elements
- System fonts only — no external font loading
- No heavy graphics — substance over style
- ASCII-art style dividers or simple CSS borders
- CSS-only loading spinners

### 1.5 Environments

- **Production:** Served at pawradise.ir (root path /)
- **Staging:** Served at server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir (subdomain, identical paths to production)
- Both environments use identical `/api/v1` paths — environments are separated by hostname (domain-based routing), not path prefix
- Staging is for verification; production is the live customer-facing site
- Host nginx on RED (port 443) terminates SSL and proxies by Host header to k3s ingress-nginx (NodePort 30758)

---

## 2. User Flows & Capabilities

### 2.1 Visitor / Browser Flow (No Auth Required)

1. Lands on shop homepage — sees product grid (paginated, 12 per page)
2. Can search by text (searches title + description)
3. Can filter by category, price range, file type, rating
4. Can sort by newest, price low-high, price high-low
5. Pinned products appear first in listings
6. Clicks a product — sees product detail page with:
   - Image gallery with lightbox/zoom
   - Generated preview/thumbnail for images/GIFs (watermarked/low-res so
     buyers cannot simply save and use without purchasing)
   - Title, full description, category, price in USD
   - Tier selection (if product has multiple tiers — different file versions
     at different prices)
   - Pay-what-you-want price input (if enabled for this product)
   - File info, download policy, related products
7. Can browse categories independently
8. Can browse bundles (discounted multi-product packs)
9. Can browse community feed (read-only)
10. Can view individual community posts (read-only)
11. Can view public user profiles (read-only)
12. Can share products to X, Instagram, Reddit, and other platforms via
    share buttons with Open Graph preview cards

### 2.2 Registered User Flow

1. Registers with email + password (optionally via referral link)
2. Logs in — gets JWT token
3. Can browse shop (same as visitor)
4. Has a unique referral link to share; earns commission on purchases made by
   referred users
5. Can view referral dashboard (link, earnings, referred users)
6. Can create community posts (short text, ~500 char limit)
7. Can like/unlike posts
8. Can comment on posts
9. Can follow/unfollow users
10. Can rate products (1-5 stars) after purchasing
11. Can view own profile and edit (bio, avatar URL, wallet address)
12. Can create orders (select products, quantities, apply coupon codes)
13. Can view own orders (order history)
14. Can check order payment status
15. Can download purchased assets (unlimited, as long as purchased)
16. Can manage wishlist (add/remove products)
17. Can manage cart (add/remove items, quantities)
18. Can view recently viewed products
19. Can compare products

### 2.3 Checkout / Payment Flow

**Simplified payment flow (current scope):**

1. User is logged in, has items in cart or selected products
2. User clicks "Checkout" — sees order summary:
   - Items with tier selections and quantities
   - Subtotal
   - Discount applied (if coupon code entered)
   - Total USD
3. User clicks "Pay" — confirmation popup/modal appears
4. Payment method: crypto wallet (BSC-compatible) OR simplified pay button
5. After payment initiated/confirmed — order marked as paid
6. Download links become available for purchased items
7. User can re-download purchased items from "My Purchases" page (unlimited)

**Crypto wallet payment (extended scope):**

1. User connects wallet (MetaMask or similar) — wallet address saved to profile
2. Backend provides: destination address, amount to send, chain, order ID as memo
3. User sends transaction from their wallet
4. Backend polls block explorer API for transaction confirmation
5. After N confirmations — order marked paid — downloads enabled
6. Exchange rates: manual admin-set rates (chain, symbol, rate_to_usd)

**Guest checkout flow:**

1. Visitor enters email at checkout (no account required)
2. Guest order is created with email for receipt/delivery
3. After payment, download link shown and sent to provided email
4. Guest can re-check order status anytime with order ID + email

**Coupon / discount flow:**

1. User enters coupon code at checkout
2. System validates: not expired, usage limit not met, minimum purchase met,
   applicable to cart items (if product-specific)
3. Discount applied to order total
4. Coupon usage count incremented

**Pay-what-you-want flow:**

1. Product has PWYW enabled by admin
2. Product detail shows suggested price + minimum price
3. User enters their price (must be >= minimum)
4. User checks out with their chosen price

**Referral flow:**

1. User A signs up and gets a unique referral link
2. User A shares link with User B
3. User B signs up via the link — User A is recorded as referrer
4. When User B makes any purchase, User A earns a commission
5. Commission = (purchase total) × (global referral percentage)
6. User A can view earnings in their referral dashboard

### 2.4 Admin Flow

1. Admin logs in with admin token (Bearer token auth)
2. Admin dashboard shows:
   - **Stats:** revenue over time, top-selling products, conversion rates,
     total users, total orders, total products, total categories
   - **Users:** list, delete, reset passwords
   - **Products:** CRUD, tiers, image gallery, preview generation, pinning,
     PWYW toggle, bulk operations
   - **Bundles:** create/edit/delete product bundles with discounted pricing
   - **Coupons:** create/edit/delete discount codes
   - **Orders:** list, view, update status with transition validation
   - **Community:** list posts, delete posts (moderation)
   - **Referrals:** view all referrals and commissions
   - **Exchange rates:** list, set, delete rates
   - **Settings:** view, set settings (including global referral commission %)
   - **Email export:** export buyer emails (CSV/JSON) for email marketing tools

---

## 3. Page Structure & User Interface

### 3.1 Public Pages (No Auth)

| Page | URL | Purpose |
|------|-----|---------|
| Shop home | / | Product grid with search, filter, sort, pagination; pinned products first |
| Product detail | /product/:slug | Single product: image gallery, preview, tier selection, PWYW input, buy button, related products, share buttons |
| Bundle detail | /bundle/:id | Bundle: list of included products, bundle price vs. individual total, buy button |
| Community feed | /community | Posts feed (public read) |
| Post detail | /post/:id | Single post + comments thread |
| User profile (public) | /profile/:id | Public profile: avatar, bio, wallet, posts, follow counts |
| Login | /login | Email/password login form |
| Register | /register | Registration form (with optional referral code) |

### 3.2 Authenticated User Pages

| Page | URL | Purpose |
|------|-----|---------|
| My Account | /account | Profile edit, wallet address, purchase history, re-downloads |
| My Referrals | /referrals | Referral link, earnings, referred users list |
| Checkout | /checkout | Order review, coupon code input, pay button, confirmation popup, payment polling |
| Cart | /cart | Add/remove items, quantities, apply coupon, checkout |
| Wishlist | /wishlist | Wishlisted products |

### 3.3 Admin Pages

| Page | URL | Purpose |
|------|-----|---------|
| Admin panel | /admin | Dashboard with tabs: Stats, Users, Products, Bundles, Coupons, Orders, Community, Referrals, Exchange Rates, Settings, Email Export |

### 3.4 Shared UI Elements

- **Header (all pages):** Logo "PAWRADISE", nav links (Shop, Community,
  Account), wallet status indicator, admin link if admin
- **Footer (all pages):** Copyright, links
- **Product card:** Thumbnail (generated preview), title, category badge,
  price (or "Pay what you want"), buy button, pinned badge if pinned
- **Community post:** Author + timestamp, content, like/comment buttons, counts
- **Loading states:** CSS-only spinners
- **Error/empty states:** For no products, no posts, no purchases, no results

---

## 4. API Surface & Routes

### 4.1 Base URL
All API routes under /api/v1/. Both staging and production use identical paths — environments are separated by hostname (domain-based routing), not path prefix.

### 4.2 Authentication Routes (Public)

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/v1/register | Register new user (email, password, name, optional referral_code) — returns JWT |
| POST | /api/v1/login | Login with email/password — returns JWT |
| POST | /api/v1/logout | Invalidate current JWT session |
| GET | /api/v1/me | Get current authenticated user profile (JWT protected) |
| PUT | /api/v1/me | Update profile fields: name, bio, wallet_address (JWT protected) |
| GET | /api/v1/profile/:id | Get public profile of any user (public) |

### 4.3 Product Routes (Public)

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/products | List products — pagination, search (title/description), filter (category, price_min, price_max, file_type, rating), sort (newest, price_asc, price_desc, popular). Pinned products always appear first. Returns products array + total count |
| GET | /api/v1/products/:slug | Get single product by slug — full detail: images, generated previews, tiers (if any), PWYW config, related products |
| GET | /api/v1/categories | List all categories (id, name, slug, description, parent_id) |
| GET | /api/v1/bundles | List all bundles (id, title, description, price, products in bundle) |
| GET | /api/v1/bundles/:id | Get single bundle detail with component products |

### 4.4 Order Routes (JWT Protected)

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/v1/orders | Create order from cart/selected items. Accepts coupon_code (optional), tier selections. Returns order with items, total, status |
| GET | /api/v1/orders | List current user's orders |
| GET | /api/v1/orders/:id | Get single order detail with items, totals, status, payment info |
| GET | /api/v1/orders/:id/payment | Get payment details for an order |
| GET | /api/v1/orders/:id/status | Check current payment status |
| GET | /api/v1/orders/:id/download/:itemId | Download asset file for an order item (auth + payment verified, unlimited downloads) |

### 4.5 Coupon Routes (Public validate, admin manage)

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/v1/coupons/validate | Validate a coupon code (check expiry, usage, minimum purchase, product applicability) — returns discount details |

### 4.6 Referral Routes (JWT Protected)

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/referrals | Get current user's referral stats: link, earnings, referred users |
| GET | /api/v1/referrals/earnings | Get detailed earnings history |

### 4.7 Community Routes (Public read, JWT for write)

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/community/posts | List posts — feed with pagination |
| POST | /api/v1/community/posts | Create a community post (content, max ~500 chars) |
| GET | /api/v1/community/posts/:id | Get single post with comments |
| POST | /api/v1/community/posts/:id/like | Like/unlike toggle for a post |
| DELETE | /api/v1/community/posts/:id/like | Remove like from a post |
| POST | /api/v1/community/posts/:id/comments | Add comment to a post |
| DELETE | /api/v1/community/posts/:id/comments/:commentId | Delete own comment |
| POST | /api/v1/community/follow/:userId | Follow a user |
| DELETE | /api/v1/community/follow/:userId | Unfollow a user |
| GET | /api/v1/community/users/:id | Get public user profile with post count, follower/following counts |

### 4.8 Feature Routes (JWT Protected)

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/cart | Get current user's cart (or create if doesn't exist) |
| POST | /api/v1/cart/items | Add item to cart (product_id, tier_id, quantity) |
| DELETE | /api/v1/cart/items/:id | Remove item from cart |
| GET | /api/v1/wishlist | List user's wishlist products |
| POST | /api/v1/wishlist/toggle | Add/remove product from wishlist |
| POST | /api/v1/reviews | Rate a product (1-5 stars). Verified purchase required |
| GET | /api/v1/reviews/:productId | List ratings for a product (average, count, distribution) |
| POST | /api/v1/recently-viewed | Record a product as recently viewed |
| GET | /api/v1/recently-viewed | List recently viewed products |
| POST | /api/v1/compare/toggle | Add/remove product from comparison list |
| GET | /api/v1/compare | List products in comparison |
| GET | /api/v1/recommendations/:productId | Get recommended products based on category |

### 4.9 Guest Order Routes (Public)

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/v1/guest-orders | Create a guest order (email, total, crypto details, status) |
| GET | /api/v1/guest-orders/:id | Check guest order status by ID + email verification |

### 4.10 Exchange Rate Routes (Public read, admin write)

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/exchange-rates | List all exchange rates (chain, symbol, rate_to_usd, updated_at) |
| GET | /api/v1/exchange-rates/:chain | Get single exchange rate by chain |
| PUT | /api/v1/exchange-rates/:chain | Set/update exchange rate (admin) |
| DELETE | /api/v1/exchange-rates/:chain | Delete exchange rate (admin) |

### 4.11 Settings Routes (Public read, admin write)

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/settings/:key | Get a single setting value by key |
| GET | /api/v1/settings | List all settings (admin) |
| PUT | /api/v1/settings/:key | Set/update a setting value (admin) |

### 4.12 Admin Routes (Admin Token Auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/admin/users | List all users |
| DELETE | /api/v1/admin/users/:id | Delete a user |
| POST | /api/v1/admin/users/:id/reset-password | Reset a user's password |
| GET | /api/v1/admin/products | List all products (admin) |
| POST | /api/v1/admin/products | Create a product |
| PUT | /api/v1/admin/products/:id | Update product |
| DELETE | /api/v1/admin/products/:id | Delete/draft product |
| POST | /api/v1/admin/products/:id/tiers | Add a tier to a product |
| PUT | /api/v1/admin/products/:id/tiers/:tierId | Update a tier |
| DELETE | /api/v1/admin/products/:id/tiers/:tierId | Delete a tier |
| POST | /api/v1/admin/products/:id/images | Add image to product gallery |
| DELETE | /api/v1/admin/products/:id/images/:imageId | Delete product image |
| POST | /api/v1/admin/products/:id/generate-previews | Trigger preview generation (admin-only, for images/GIFs) |
| POST | /api/v1/admin/products/:id/pin | Pin product to top |
| DELETE | /api/v1/admin/products/:id/pin | Unpin product |
| POST | /api/v1/admin/products/bulk | Bulk update products (status, category) |
| GET | /api/v1/admin/bundles | List all bundles |
| POST | /api/v1/admin/bundles | Create a bundle |
| PUT | /api/v1/admin/bundles/:id | Update a bundle |
| DELETE | /api/v1/admin/bundles/:id | Delete a bundle |
| GET | /api/v1/admin/coupons | List all coupons |
| POST | /api/v1/admin/coupons | Create a coupon |
| PUT | /api/v1/admin/coupons/:id | Update a coupon |
| DELETE | /api/v1/admin/coupons/:id | Delete a coupon |
| GET | /api/v1/admin/orders | List all orders |
| GET | /api/v1/admin/orders/:id | Get order detail |
| PUT | /api/v1/admin/orders/:id/status | Update order status with transition validation |
| GET | /api/v1/admin/community/posts | List community posts (moderation) |
| DELETE | /api/v1/admin/community/posts/:id | Delete a community post |
| GET | /api/v1/admin/referrals | List all referrals and commissions |
| GET | /api/v1/admin/stats | Dashboard stats: revenue (daily/weekly/monthly), top-selling products, conversion rates, totals |
| GET | /api/v1/admin/settings | List all settings |
| PUT | /api/v1/admin/settings/:key | Set setting value (e.g., referral_commission_percent) |
| PUT | /api/v1/admin/exchange-rates/:chain | Set exchange rate |
| DELETE | /api/v1/admin/exchange-rates/:chain | Delete exchange rate |
| GET | /api/v1/admin/guest-orders | List guest orders |
| GET | /api/v1/admin/guest-orders/:id | Get guest order detail |
| POST | /api/v1/admin/export/emails | Export buyer emails (CSV/JSON). Optional filters: date range, category, product |

---

## 5. Architecture & Services

### 5.1 Infrastructure

Details and IPs: `docs/DEPLOYMENT-ARCHITECTURE.md`. Summary:

- **RED** — k3s serving node (ingress-nginx NodePort 30758). Host nginx on 80/443 terminates TLS and routes by Host header.
  - `pawradise.ir` → production namespace
  - `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir` → staging namespace
- **BLUE** — build box (Go, Docker). Images push to RED’s registry.
- PostgreSQL 16 in `database` namespace: **one logical database per service per environment** (`appdb_<service>_staging` / `_production`), not a single `appdb_staging`.
- One Deployment per backend service and per MFE, per namespace.

### 5.2 Backend services

Public routes stay under `/api/v1/` as listed in §4. Each service is a Go/Gin process in `services/<name>/` with its own DB. Shared JWT/CORS live in `services/shared/`.

| Service | Port | Responsibility |
|---------|------|----------------|
| identity-service | 8081 | Register, login, logout, profile, referrals, commissions |
| product-service | 8082 | Products, categories, bundles, tiers, images, pin, PWYW, recommendations |
| commerce-service | 8083 | Orders, guest orders, cart, wishlist, coupons, recently-viewed, compare, downloads |
| community-service | 8084 | Posts, likes, comments, follows, public community profiles |
| review-service | 8085 | Verified-purchase ratings |
| payment-service | 8086 | Payment details/status, exchange rates, public settings read |
| admin-service | 8087 | Admin aggregation, stats, email export, settings write (must call other services — must not copy their tables) |
| media-service | 8088 | Upload, download, watermarked previews |

Cross-service consistency is via HTTP or events (`services/shared/events`), not shared tables.

### 5.3 Frontend modules (microfrontends)

Static HTML + Alpine.js per MFE under `frontend/`. Shared chrome lives in `shared/chrome/` but each MFE currently inlines its own header. Target page URLs remain those in §3.

| MFE | Directory | Routes (implemented prefix) |
|-----|-----------|-------------------------------|
| Shop | `frontend/shop-mfe` | `/`, `/category` |
| Product | `frontend/product-mfe` | `/product/:slug`, bundle/request pages |
| Community | `frontend/community-mfe` | `/community`, `/post/:id`, `/profile/:id` |
| Account | `frontend/account-mfe` | `/account`, `/cart`, `/wishlist`, `/referrals` |
| Checkout | `frontend/checkout-mfe` | `/checkout` |
| Auth | `frontend/auth-mfe` | `/login`, `/register` |
| Admin | `frontend/admin-app` | `/admin` |

API client: `shared/lib/api.js`. Theme: `shared/theme/`.

### 5.4 Database Schema

| Table | Key columns | Purpose |
|-------|------------|---------|
| users | id, email, password_hash, name, referred_by (user_id nullable), created_at, updated_at | User accounts |
| user_profiles | user_id, bio, avatar_url, wallet_address, created_at, updated_at | Extended profile data |
| referral_links | id, user_id, code (unique), created_at | Unique referral codes per user |
| user_referrals | id, referrer_id, referred_id, created_at | Track who referred whom |
| referral_commissions | id, order_id, referrer_id, commission_percent, commission_usd, status, created_at | Commission earned per order |
| products | id, title, slug, description, category_id, price_usd, asset_path, asset_hash, status, file_size_bytes, file_mime_type, pwyw_enabled, pwyw_min_price, pinned, sort_order, user_id, created_at, updated_at | Products for sale |
| product_tiers | id, product_id, tier_name, file_path, file_hash, price_usd, sort_order, created_at | Multiple file versions / tiers per product |
| product_images | id, product_id, url, image_type (full/preview/thumbnail), is_primary, width, height, created_at | Product image gallery and generated previews |
| categories | id, name, slug, description, parent_id, created_at, updated_at | Product categories |
| bundles | id, title, slug, description, price_usd, status, created_at, updated_at | Product bundles |
| bundle_items | id, bundle_id, product_id, quantity, created_at | Products in a bundle |
| coupons | id, code, discount_type (percentage/fixed), discount_value, expires_at, usage_limit, times_used, min_purchase_usd, product_id (nullable — null means cart-wide), is_active, created_by, created_at, updated_at | Discount codes |
| coupon_usages | id, coupon_id, user_id, order_id, discount_usd, created_at | Track coupon usage |
| orders | id, user_id, coupon_id, coupon_discount_usd, total_usd, status, payment_tx_hash, payment_chain, payment_confirmations, guest_email, billing_name, paid_at, crypto_amount, crypto_address, crypto_chain, created_at, updated_at | Customer orders |
| order_items | id, order_id, product_id, product_tier_id, product_title, product_slug, quantity, unit_price_usd, download_count, created_at | Items in an order |
| downloads | id, order_item_id, user_id, downloaded_at, ip_address | Download tracking (unlimited — for analytics only) |
| community_posts | id, user_id, content, type, created_at, updated_at | Community posts |
| post_likes | id, user_id, post_id, type, created_at | Likes on posts or comments |
| post_comments | id, user_id, post_id, content, type, created_at | Comments on posts |
| follows | id, follower_id, following_id, created_at | User follow relationships |
| settings | id, setting_key, value, updated_at | Site settings (key-value) |
| exchange_rates | id, chain, symbol, rate_to_usd, updated_at | Crypto exchange rates |
| invalidated_tokens | id, user_id, token_hash, created_at | JWT blacklist for logout |
| cart_items | id, user_id, product_id, product_tier_id, quantity, added_at | Shopping cart items |
| wishlists | id, user_id, product_id, created_at | Wishlist entries |
| reviews | id, product_id, user_id, rating (1-5), created_at | Product ratings (verified purchase required, no comments/media) |
| recently_viewed | id, user_id, product_id, viewed_at, created_at | Recently viewed products |
| product_comparison | id, user_id, product_id, added_at | Product comparison list |
| guest_orders | id, email, total_usd, crypto_chain, crypto_amount, crypto_address, status, created_at | Guest orders |

### 5.5 External Dependencies

- **PostgreSQL:** On RED, in k3s database namespace. Access via connection
  string from backend.
- **No external APIs required for v1** — exchange rates are manual, no price
  feed API, no blockchain node needed.
- **Wallet integration (future):** MetaMask or similar browser extension for
  user wallet connection. Block explorer APIs (BscScan, Etherscan) for payment
  verification.
- **Image processing:** Go-native image manipulation (imaging, gift, or similar
  library) — no external service needed for preview generation.
- **Hosting:** k3s on RED, single node, low-resource (~2GB RAM).

### 5.6 Asset Storage

- Digital assets stored on a PersistentVolume (local-path-provisioner) in k3s
- Backend serves files through download endpoint with auth + payment verification
- File path stored in products.asset_path
- Asset integrity: SHA-256 hash stored in products.asset_hash
- Upload via admin panel (file upload to backend, stored on PV)
- Generated previews (watermarked/thumbnails) stored in product_images table

### 5.7 Image Processing (Preview Generation)

When a product is created or its images are updated:
- For image files (JPG, PNG, GIF, WebP): the system generates:
  - A **thumbnail** (small crop, ~200px wide)
  - A **preview image** (watermarked or downscaled version, ~800px wide) —
    shown on the product detail page instead of the full-resolution file
- Watermark: subtle overlay of the shop logo or "PAWRADISE" text
- Preview images are stored separately from the full product file
- Buyers see previews; only after purchase do they access the full file
- If the product is not an image/GIF (e.g., ZIP, PDF, audio), no preview is
    generated — only file info is shown

### 5.8 SEO

- Every product page has unique `<title>`, `<meta description>`, and Open Graph
  tags (og:title, og:description, og:image, og:url)
- Category pages have unique meta tags
- XML sitemap at /sitemap.xml (products, categories, bundles, community posts)
- Clean semantic URLs: /product/my-asset, /bundle/ui-kit-pack, /category/icons
- Structured data (JSON-LD Product schema) on product pages
- robots.txt and proper canonical tags

---

## 6. Admin Capabilities Matrix

### 6.1 User Management
- List all users (paginated, with email, name, creation date, referral stats)
- Delete user
- Reset user password

### 6.2 Product Management
- Create product: title, slug (auto-generated), description, category, price
  USD, asset upload, asset hash, status (active/draft/archived), file info
- Edit product: update any field
- Delete product: remove product and its images
- **Product tiers:** add multiple file versions with different prices (optional)
- **Product image gallery:** upload multiple images, set primary image, trigger
  preview generation
- **Product preview:** manually trigger or regenerate watermarked previews/
  thumbnails for images and GIFs
- **Pin product:** pin/unpin products to appear at top of listings
- **Pay-what-you-want:** enable/disable PWYW per product, set minimum price
- **Bulk update:** change status for multiple products, change category for
  multiple products

### 6.3 Bundle Management
- Create bundle: title, description, select component products, set bundle
  price (discounted vs. individual total)
- Edit bundle: change products, pricing
- Delete/draft bundle

### 6.4 Coupon Management
- Create coupon: code, discount type (percentage/fixed), value, expiration,
  usage limits, minimum purchase, product-specific or cart-wide
- Edit coupon
- Delete coupon
- View usage stats (times used, total discount given)

### 6.5 Order Management
- List all orders (paginated, with user, total, status, date)
- View order detail: items, quantities, prices, total, status, coupon used,
  payment info
- Update order status: pending, paid, completed, cancelled, refunded, failed
  (with transition validation)
- View order items with product title, slug, tier, quantity, unit price

### 6.6 Community Moderation
- List all community posts (paginated, with author, content, timestamps,
  like/comment counts)
- Delete any post

### 6.7 Referral Management
- View all referrals: referrer → referred, date, commission earned
- View commission totals per user
- View global referral payouts

### 6.8 Exchange Rate Management
- List all exchange rates (chain, symbol, rate_to_usd, updated_at)
- Set/update rate for a chain (chain, symbol, rate_to_usd)
- Delete a rate

### 6.9 Settings Management
- List all settings (key-value pairs)
- Set/update a setting value by key
- **Key settings include:**
  - referral_commission_percent (global, e.g., 5.0 = 5%)
  - site title, description, keywords
  - default currency display
  - download policy text

### 6.10 Email Export
- Export buyer emails as CSV or JSON
- Filter by: date range, category purchased, product purchased
- Export includes: email, name, total purchases, first purchase date, last
  purchase date

### 6.11 Sales Analytics Dashboard
- **Revenue over time:** daily, weekly, monthly views (chart)
- **Top-selling products:** by revenue, by units sold
- **Conversion rates:**
  - Visitor-to-product view rate
  - Product view-to-cart rate
  - Cart-to-checkout rate
  - Checkout-to-purchase rate
- **Totals:** total users, total orders, total revenue, total products, total
  categories, total downloads

---

## 7. In Scope vs Out of Scope

### In Scope (Current Product)
- Single-seller digital asset shop
- User auth (register/login/logout/profile)
- Product catalog (CRUD, list, search, detail, categories, tiers, gallery,
  previews, pinning, PWYW)
- Product bundles (discounted multi-product packs)
- Coupon/discount code system
- Referral program (referral links + fixed commission %)
- Orders (create, list, detail, payment status, download, guest orders)
- Unlimited downloads for purchased items
- Community (posts, likes, comments, follows, public profiles)
- Product ratings (1-5 stars, verified purchase)
- Features (cart, wishlist, reviews, recently-viewed, compare, recommendations)
- Exchange rates (manual, admin-set)
- Settings (key-value, admin-managed)
- Admin panel (users, products, bundles, coupons, orders, community, referrals,
  stats, settings, email export)
- Image preview generation (watermarked/thumbnails for images and GIFs)
- SEO (meta tags, Open Graph, structured data, sitemap)
- Social sharing (X, Instagram, Reddit, others)
- Digital asset download (auth + payment verified)
- Dark theme, monospace accents, CSS glow, no external fonts

### Out of Scope (Not in current product spec)
- Multi-vendor marketplace
- Subscriptions/recurring payments
- License key management (keys, activations, device limits)
- Physical goods shipping
- Advanced search (Elasticsearch, etc.) — basic DB search sufficient
- Search suggestions/autocomplete
- Real-time chat/WebSocket
- Push notifications
- Email notifications/sequences (email export only — external tool handles sending)
- Automated crypto price feed (manual rates only)
- Automated hot wallet sweeping
- User avatars via file upload (URL-based only)
- Post editing (future)
- Comment likes (future)
- User banning (future)
- Wishlist sharing/gift links (future)
- Multi-language/i18n (future)
- Mobile app
- API rate limiting & security hardening (future work)
- Abandoned cart recovery (future)

---

*Spec version: 1.2 — 2026-09-17*
*Status: Defining the target product. Implementation status tracked separately.*
