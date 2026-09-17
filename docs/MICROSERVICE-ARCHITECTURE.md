# Pawradise — Microservice & Microfrontend Architecture

> Architecture v1.0 — 2026-09-06
> Defines the target decomposition for the Pawradise digital asset marketplace.
> This document is the source of truth for service boundaries, data ownership,
> API contracts, and microfrontend composition.

---

## 1. Design Principles

### 1.1 Service Decomposition Strategy

Services are decomposed by **business domain** (DDD bounded contexts), not by
technical layer. Each service:

- **Owns its data** — no shared database tables between services
- **Exposes a well-defined API** — RESTful JSON over HTTP
- **Is independently deployable** — each runs in its own k3s deployment
- **Communicates asynchronously** where possible — via events for cross-service
  consistency
- **Is stateless** where practical — state lives in PostgreSQL or Redis

### 1.2 Microfrontend Strategy

Microfrontends are composed **by user journey area**, not by page. Each MFE:

- **Owns its HTML/JS/CSS** — no shared frontend framework or build
- **Is served independently** — each is a static asset set served by its own
  nginx container
- **Communicates via the Shell** — shared state (auth, cart count) through the
  host application
- **Can be developed independently** — each has its own build and deploy

### 1.3 Communication Pattern

```
Client (Browser)
    │
    ▼
┌─────────────────────────────────────────┐
│  Shell (Host Application)               │
│  - Header, Footer, Navigation           │
│  - Auth state, Service routing          │
│  - Composes MFEs into pages             │
└─────────────────────────────────────────┘
    │
    ├──▶ Shop MFE ──────────────▶ Product Service
    ├──▶ Product Detail MFE ────▶ Product Service
    ├──▶ Community MFE ─────────▶ Community Service
    ├──▶ Account MFE ───────────▶ Identity Service
    ├──▶ Checkout MFE ──────────▶ Commerce + Payment Services
    ├──▶ Auth MFE ──────────────▶ Identity Service
    └──▶ Admin MFE ─────────────▶ Admin + All Services
```

---

## 2. Microservices (Backend)

### 2.1 Service Inventory

| # | Service | Domain | Port | Database |
|---|---------|--------|------|----------|
| 1 | Identity Service | User auth, profiles, referrals | 8081 | identity_db |
| 2 | Product Service | Catalog, categories, bundles, images | 8082 | product_db |
| 3 | Commerce Service | Orders, cart, coupons, wishlist | 8083 | commerce_db |
| 4 | Community Service | Posts, comments, likes, follows | 8084 | community_db |
| 5 | Review Service | Product ratings | 8085 | review_db |
| 6 | Payment Service | Payments, exchange rates | 8086 | payment_db |
| 7 | Admin Service | Admin operations, stats, settings | 8087 | admin_db |
| 8 | Media Service | File storage, image processing | 8088 | — (PV) |

### 2.2 Service Detail: Identity Service

**Responsibilities:**
- User registration (email, password, name, referral code capture)
- User login (email + password → JWT)
- JWT token generation and validation
- Token invalidation (logout, password reset)
- User profile CRUD (name, bio, avatar URL, wallet address)
- Password reset
- Referral link generation and tracking
- Commission recording

**API Routes (internal):**

| Method | Path | Description |
|--------|------|-------------|
| POST | /internal/v1/register | Register new user |
| POST | /internal/v1/login | Authenticate, return JWT |
| POST | /internal/v1/logout | Invalidate token |
| GET | /internal/v1/users/:id | Get user profile |
| PUT | /internal/v1/users/:id | Update profile |
| DELETE | /internal/v1/users/:id | Delete user |
| POST | /internal/v1/users/:id/reset-password | Reset password |
| GET | /internal/v1/users/:id/referrals | Get referral stats |
| GET | /internal/v1/users/:id/commissions | Get commission history |
| POST | /internal/v1/referrals/track | Track referral signup |
| POST | /internal/v1/commissions/record | Record commission on purchase |
| POST | /internal/v1/token/validate | Validate JWT |

**Data Ownership:**
- `users` table
- `user_profiles` table
- `invalidated_tokens` table
- `referral_links` table
- `user_referrals` table
- `referral_commissions` table

### 2.3 Service Detail: Product Service

**Responsibilities:**
- Product CRUD (title, slug, description, category, price, status)
- Category CRUD
- Product tier management (multiple file versions at different prices)
- Product bundle CRUD
- Product image gallery management
- Image preview generation (watermarked thumbnails for images/GIFs)
- Product pinning and sorting
- Pay-what-you-want configuration
- Product search and filtering
- Related product recommendations

**API Routes (internal):**

| Method | Path | Description |
|--------|------|-------------|
| GET | /internal/v1/products | List products (search, filter, sort, pagination) |
| GET | /internal/v1/products/:id | Get product detail |
| POST | /internal/v1/products | Create product |
| PUT | /internal/v1/products/:id | Update product |
| DELETE | /internal/v1/products/:id | Delete product |
| GET | /internal/v1/categories | List categories |
| POST | /internal/v1/categories | Create category |
| GET | /internal/v1/bundles | List bundles |
| GET | /internal/v1/bundles/:id | Get bundle detail |
| POST | /internal/v1/bundles | Create bundle |
| PUT | /internal/v1/bundles/:id | Update bundle |
| DELETE | /internal/v1/bundles/:id | Delete bundle |
| POST | /internal/v1/products/:id/tiers | Add product tier |
| PUT | /internal/v1/products/:id/tiers/:tierId | Update tier |
| DELETE | /internal/v1/products/:id/tiers/:tierId | Delete tier |
| POST | /internal/v1/products/:id/images | Add product image |
| DELETE | /internal/v1/products/:id/images/:imageId | Delete image |
| POST | /internal/v1/products/:id/previews | Generate previews |
| POST | /internal/v1/products/:id/pin | Pin product |
| DELETE | /internal/v1/products/:id/pin | Unpin product |
| GET | /internal/v1/recommendations/:productId | Get recommendations |

**Data Ownership:**
- `products` table
- `product_tiers` table
- `product_images` table
- `categories` table
- `bundles` table
- `bundle_items` table

### 2.4 Service Detail: Commerce Service

**Responsibilities:**
- Order creation (from cart or direct)
- Order listing and detail
- Order status workflow (pending → paid → completed / cancelled / failed / refunded)
- Guest order creation and status check
- Shopping cart management
- Wishlist management
- Coupon validation and CRUD
- Recently viewed products tracking
- Product comparison list
- Discount calculation

**API Routes (internal):**

| Method | Path | Description |
|--------|------|-------------|
| POST | /internal/v1/orders | Create order |
| GET | /internal/v1/orders | List user orders |
| GET | /internal/v1/orders/:id | Get order detail |
| PUT | /internal/v1/orders/:id/status | Update order status |
| GET | /internal/v1/orders/:id/items | Get order items |
| POST | /internal/v1/guest-orders | Create guest order |
| GET | /internal/v1/guest-orders/:id | Check guest order status |
| GET | /internal/v1/cart | Get cart |
| POST | /internal/v1/cart/items | Add to cart |
| PUT | /internal/v1/cart/items/:id | Update cart item |
| DELETE | /internal/v1/cart/items/:id | Remove from cart |
| GET | /internal/v1/wishlist | List wishlist |
| POST | /internal/v1/wishlist/toggle | Toggle wishlist item |
| GET | /internal/v1/coupons/validate | Validate coupon |
| GET | /internal/v1/coupons | List coupons |
| POST | /internal/v1/coupons | Create coupon |
| PUT | /internal/v1/coupons/:id | Update coupon |
| DELETE | /internal/v1/coupons/:id | Delete coupon |
| POST | /internal/v1/recently-viewed | Record view |
| GET | /internal/v1/recently-viewed | List recently viewed |
| POST | /internal/v1/compare/toggle | Toggle compare |
| GET | /internal/v1/compare | List compared |

**Data Ownership:**
- `orders` table
- `order_items` table
- `downloads` table
- `guest_orders` table
- `cart_items` table
- `wishlists` table
- `coupons` table
- `coupon_usages` table
- `recently_viewed` table
- `product_comparison` table

### 2.5 Service Detail: Community Service

**Responsibilities:**
- Community post creation and deletion
- Post feed with pagination
- Comment creation and deletion
- Like/unlike toggle for posts and comments
- User follow/unfollow
- Public user profile aggregation

**API Routes (internal):**

| Method | Path | Description |
|--------|------|-------------|
| GET | /internal/v1/posts | List posts (feed) |
| POST | /internal/v1/posts | Create post |
| GET | /internal/v1/posts/:id | Get post detail |
| DELETE | /internal/v1/posts/:id | Delete post |
| POST | /internal/v1/posts/:id/like | Like/unlike post |
| DELETE | /internal/v1/posts/:id/like | Unlike post |
| GET | /internal/v1/posts/:id/comments | List comments |
| POST | /internal/v1/posts/:id/comments | Add comment |
| DELETE | /internal/v1/posts/:id/comments/:commentId | Delete comment |
| POST | /internal/v1/users/:id/follow | Follow user |
| DELETE | /internal/v1/users/:id/follow | Unfollow user |
| GET | /internal/v1/users/:id/profile | Get public profile |

**Data Ownership:**
- `community_posts` table
- `post_comments` table
- `post_likes` table
- `follows` table

### 2.6 Service Detail: Review Service

**Responsibilities:**
- Product rating creation (verified purchase only)
- Rating aggregation (average, count, distribution)
- Rating listing per product

**API Routes (internal):**

| Method | Path | Description |
|--------|------|-------------|
| POST | /internal/v1/ratings | Rate a product |
| GET | /internal/v1/ratings/product/:productId | Get product ratings |
| GET | /internal/v1/ratings/product/:productId/average | Get average rating |

**Data Ownership:**
- `reviews` table

### 2.7 Service Detail: Payment Service

**Responsibilities:**
- Payment detail generation (crypto address, amount, chain)
- Payment status tracking
- Exchange rate management (CRUD)
- Crypto payment verification (future)

**API Routes (internal):**

| Method | Path | Description |
|--------|------|-------------|
| GET | /internal/v1/payments/order/:orderId | Get payment details |
| GET | /internal/v1/payments/order/:orderId/status | Check payment status |
| POST | /internal/v1/payments/order/:orderId/confirm | Confirm payment |
| GET | /internal/v1/exchange-rates | List all rates |
| GET | /internal/v1/exchange-rates/:chain | Get single rate |
| PUT | /internal/v1/exchange-rates/:chain | Set rate |
| DELETE | /internal/v1/exchange-rates/:chain | Delete rate |

**Data Ownership:**
- `exchange_rates` table
- Payment state is derived from `orders` table in Commerce Service

### 2.8 Service Detail: Admin Service

**Responsibilities:**
- Dashboard statistics aggregation
- User management (list, delete, reset password)
- Community moderation (delete posts)
- Email export
- Settings management
- Acts as **gateway/aggregator** for admin operations across services

**API Routes (internal):**

| Method | Path | Description |
|--------|------|-------------|
| GET | /internal/v1/admin/stats | Dashboard stats |
| GET | /internal/v1/admin/users | List all users |
| DELETE | /internal/v1/admin/users/:id | Delete user |
| POST | /internal/v1/admin/users/:id/reset-password | Reset password |
| GET | /internal/v1/admin/products | List all products |
| POST | /internal/v1/admin/products/bulk | Bulk update products |
| GET | /internal/v1/admin/orders | List all orders |
| GET | /internal/v1/admin/community/posts | List all posts |
| DELETE | /internal/v1/admin/community/posts/:id | Delete post |
| GET | /internal/v1/admin/referrals | List referrals |
| GET | /internal/v1/admin/settings | List settings |
| PUT | /internal/v1/admin/settings/:key | Update setting |
| POST | /internal/v1/admin/export/emails | Export emails |

**Data Ownership:**
- `settings` table
- Other data is fetched from respective services via internal APIs

### 2.9 Service Detail: Media Service

**Responsibilities:**
- File upload (product assets, images)
- File download serving (with auth verification)
- Image preview generation (watermarked, thumbnail)
- Asset storage management on PersistentVolume

**API Routes (internal):**

| Method | Path | Description |
|--------|------|-------------|
| POST | /internal/v1/media/upload | Upload file |
| GET | /internal/v1/media/files/:path | Serve file (auth verified) |
| POST | /internal/v1/media/images/:productId/previews | Generate previews |
| DELETE | /internal/v1/media/files/:path | Delete file |

**Data Ownership:**
- Files on PersistentVolume
- Image metadata references stored in Product Service

---

## 3. Microfrontends

### 3.1 MFE Inventory

| # | MFE | Route Prefix | nginx Port | nginx Internal Port |
|---|-----|-------------|------------|---------------------|
| 1 | Shell | / | 30080 | 80 |
| 2 | Shop | / (composed) | — | — |
| 3 | Product Detail | /product/:slug | — | — |
| 4 | Community | /community, /post/:id | — | — |
| 5 | Account | /account, /referrals | — | — |
| 6 | Checkout | /checkout, /cart, /wishlist | — | — |
| 7 | Auth | /login, /register | — | — |
| 8 | Admin | /admin | — | — |

### 3.2 MFE Composition Model

The **Shell** is the host application. It:

1. Renders the shared header, footer, and navigation
2. Manages auth state (JWT storage, login status)
3. Routes requests to the correct MFE
4. Provides shared state to child MFEs (cart count, user info)

Each MFE is a **standalone HTML page** with its own JS/CSS. MFEs are composed
into the Shell via **server-side includes (SSI)** or **nginx subrequests** at
the edge, NOT via JavaScript module federation.

```
┌──────────────────────────────────────────────────────┐
│  Shell (index.html)                                   │
│  ┌────────────────────────────────────────────────┐  │
│  │  Header: Logo, Nav, Auth status, Cart count    │  │
│  └────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────┐  │
│  │  MFE Content Area                              │  │
│  │  (Shop / Product / Community / Account / etc.) │  │
│  └────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────┐  │
│  │  Footer: Copyright, Links                      │  │
│  └────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
```

### 3.3 MFE Detail: Shop

**Route:** `/` (home), `/category/:slug`, `/bundle/:id`
**Pages:** Product grid with search, filter, sort, pagination; bundle listing
**APIs called:** Product Service (products, categories, bundles)
**State shared to Shell:** None
**State received from Shell:** Auth status (to show "Add to cart" or "Login to buy")

### 3.4 MFE Detail: Product Detail

**Route:** `/product/:slug`
**Pages:** Single product view with gallery, tiers, PWYW, buy button, ratings
**APIs called:** Product Service, Review Service, Media Service (images)
**State shared to Shell:** None
**State received from Shell:** Auth status, cart count

### 3.5 MFE Detail: Community

**Route:** `/community`, `/post/:id`, `/profile/:id`
**Pages:** Post feed, post detail with comments, public profile
**APIs called:** Community Service, Identity Service (profile data)
**State shared to Shell:** None
**State received from Shell:** Auth status (to show create/like/comment or login prompt)

### 3.6 MFE Detail: Account

**Route:** `/account`, `/referrals`
**Pages:** Profile edit, purchase history, re-downloads, referral dashboard
**APIs called:** Identity Service, Commerce Service (orders, downloads)
**State shared to Shell:** None
**State received from Shell:** Auth status (redirect to login if not authenticated)

### 3.7 MFE Detail: Checkout

**Route:** `/checkout`, `/cart`, `/wishlist`
**Pages:** Cart review, checkout with coupon input, payment flow, wishlist
**APIs called:** Commerce Service, Payment Service, Identity Service
**State shared to Shell:** Cart count (updated after add/remove)
**State received from Shell:** Auth status

### 3.8 MFE Detail: Auth

**Route:** `/login`, `/register`
**Pages:** Login form, registration form
**APIs called:** Identity Service
**State shared to Shell:** Auth status (after successful login)
**State received from Shell:** None

### 3.9 MFE Detail: Admin

**Route:** `/admin`
**Pages:** Admin dashboard with tabs (Stats, Users, Products, Bundles, Coupons, Orders, Community, Referrals, Exchange Rates, Settings, Email Export)
**APIs called:** Admin Service (which aggregates from all services)
**State shared to Shell:** None
**State received from Shell:** Admin auth status

---

## 4. API Gateway

### 4.1 Gateway Responsibility

An **API Gateway** (nginx or lightweight Go service) sits in front of all
microservices and:

1. **Routes requests** to the correct internal service
2. **Validates JWT** at the edge (rejects unauthenticated requests before they
   reach services)
3. **Rate limits** per IP/token
4. **CORS handling**
5. **Request/response logging**

### 4.2 Routing Table

| External Path | Internal Service | Internal Path |
|---------------|-----------------|---------------|
| /api/v1/identity/* | Identity Service | /internal/v1/* |
| /api/v1/products/* | Product Service | /internal/v1/* |
| /api/v1/categories/* | Product Service | /internal/v1/* |
| /api/v1/bundles/* | Product Service | /internal/v1/* |
| /api/v1/orders/* | Commerce Service | /internal/v1/* |
| /api/v1/cart/* | Commerce Service | /internal/v1/* |
| /api/v1/wishlist/* | Commerce Service | /internal/v1/* |
| /api/v1/coupons/* | Commerce Service | /internal/v1/* |
| /api/v1/guest-orders/* | Commerce Service | /internal/v1/* |
| /api/v1/community/* | Community Service | /internal/v1/* |
| /api/v1/reviews/* | Review Service | /internal/v1/* |
| /api/v1/ratings/* | Review Service | /internal/v1/* |
| /api/v1/payments/* | Payment Service | /internal/v1/* |
| /api/v1/exchange-rates/* | Payment Service | /internal/v1/* |
| /api/v1/admin/* | Admin Service | /internal/v1/* |
| /api/v1/media/* | Media Service | /internal/v1/* |

### 4.3 Staging vs Production

Both environments use identical gateway configs, differing only in:

- Namespace (staging vs production)
- Database logical name (appdb_staging vs appdb_production)
- External URL path prefix (/staging/ for staging)

---

## 5. Data Architecture

### 5.1 Database per Service

Each service has its **own logical database** within the shared PostgreSQL
instance. Cross-service data access is via **API calls**, not JOINs.

| Service | Logical DB | Tables |
|---------|-----------|--------|
| Identity Service | identity_db | users, user_profiles, invalidated_tokens, referral_links, user_referrals, referral_commissions |
| Product Service | product_db | products, product_tiers, product_images, categories, bundles, bundle_items |
| Commerce Service | commerce_db | orders, order_items, downloads, guest_orders, cart_items, wishlists, coupons, coupon_usages, recently_viewed, product_comparison |
| Community Service | community_db | community_posts, post_comments, post_likes, follows |
| Review Service | review_db | reviews |
| Payment Service | payment_db | exchange_rates |
| Admin Service | admin_db | settings |

### 5.2 Cross-Service Data Consistency

When data must be consistent across services, use **eventual consistency**
via a lightweight message bus (Redis pub/sub or PostgreSQL NOTIFY/LISTEN):

- **Order created** → Commerce Service emits event → Identity Service records referral commission
- **Product deleted** → Product Service emits event → Commerce Service removes from carts/wishlists
- **User deleted** → Identity Service emits event → Community Service anonymizes posts

---

## 6. Infrastructure

### 6.1 k3s Deployment Topology (Production)

```
production namespace
├── api-gateway (deployment, NodePort 8080:30080)
├── identity-service (deployment, NodePort 8081)
├── product-service (deployment, NodePort 8082)
├── commerce-service (deployment, NodePort 8083)
├── community-service (deployment, NodePort 8084)
├── review-service (deployment, NodePort 8085)
├── payment-service (deployment, NodePort 8086)
├── admin-service (deployment, NodePort 8087)
├── media-service (deployment, NodePort 8088)
├── shell-mfe (deployment, nginx:alpine, NodePort 80:30084)
├── shop-mfe (deployment, nginx:alpine, NodePort 80:30085)
├── product-mfe (deployment, nginx:alpine, NodePort 80:30086)
├── community-mfe (deployment, nginx:alpine, NodePort 80:30087)
├── account-mfe (deployment, nginx:alpine, NodePort 80:30088)
├── checkout-mfe (deployment, nginx:alpine, NodePort 80:30089)
├── auth-mfe (deployment, nginx:alpine, NodePort 80:30090)
├── admin-mfe (deployment, nginx:alpine, NodePort 80:30091)
```

### 6.2 Resource Constraints

Total services: 16 deployments (8 backend + 8 frontend)
RAM per service: ~64-128Mi
Total estimated: ~1.5-2GB RAM on RED (within 2GB constraint)

---

## 7. Test Architecture

### 7.1 Test Pyramid

```
        /  E2E Tests  \         ← Full user journey (existing)
       /───────────────\        ← 104 tests, Node.js
      / Integration Tests\      ← Cross-service flows
     /───────────────────\      ← Go + test containers
    /    Unit Tests       \    ← Per-service business logic
   /───────────────────────\   ← Go testing, sqlmock
```

### 7.2 Test File Layout

```
tests/
  unit/
    identity-service/
      auth_test.go
      user_test.go
      referral_test.go
      jwt_test.go
    product-service/
      product_test.go
      category_test.go
      tier_test.go
      bundle_test.go
      image_test.go
      search_test.go
    commerce-service/
      order_test.go
      cart_test.go
      wishlist_test.go
      coupon_test.go
      guest_order_test.go
    community-service/
      post_test.go
      comment_test.go
      like_test.go
      follow_test.go
    review-service/
      rating_test.go
    payment-service/
      payment_test.go
      exchange_rate_test.go
    admin-service/
      stats_test.go
      admin_user_test.go
      admin_product_test.go
      settings_test.go
    media-service/
      upload_test.go
      download_test.go
      image_processing_test.go
  integration/
    product-order-flow_test.go
    coupon-discount-flow_test.go
    referral-commission-flow_test.go
    payment-verification-flow_test.go
    user-deletion-cascade_test.go
  e2e/
    e2e-suite.js  (existing)
```

### 7.3 Unit Test Standards

- Use Go's `testing` package
- Use `sqlmock` for database mocking
- Test business logic in isolation
- No real HTTP calls — mock HTTP clients for cross-service calls
- Each test function is independent and idempotent

### 7.4 Integration Test Standards

- Use Go's `testing` package with `testcontainers` or in-memory PostgreSQL
- Test cross-service workflows (e.g., order creation → payment status → download)
- Verify event emission and handling
- Verify data consistency across service boundaries

---

## 8. Migration Path

### 8.1 From Monolith to Microservices

The current monolithic Go backend will be **strangled** into microservices:

1. **Phase 1:** Define service boundaries and API contracts (this document)
2. **Phase 2:** Build microservices alongside existing monolith
3. **Phase 3:** Route new features to microservices, keep monolith for legacy
4. **Phase 4:** Gradually extract features from monolith into services
5. **Phase 5:** Decommission monolith when all features are extracted

### 8.2 From Monolithic Frontend to Microfrontends

1. **Phase 1:** Build Shell with SSI composition
2. **Phase 2:** Extract Shop and Product Detail as separate MFEs
3. **Phase 3:** Extract Community and Account as separate MFEs
4. **Phase 4:** Extract Checkout and Auth as separate MFEs
5. **Phase 5:** Extract Admin as separate MFE

---

*Architecture version: 1.0 — 2026-09-06*
*Status: Defining target architecture. Implementation to follow.*
