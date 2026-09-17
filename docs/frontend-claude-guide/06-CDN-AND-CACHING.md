# Pawradise — CDN & Caching Guide

> Phase 6, after correctness is done. A CDN sits in front of the ingress
> (e.g. Cloudflare's free tier pointed at RED's public IP/domain) and caches
> **GET responses only**, honoring `Cache-Control` set by the origin. This
> doc classifies every endpoint and gives the exact headers to emit.

---

## 1. Classification

| Class | Endpoints | Cache-Control | Notes |
|---|---|---|---|
| **Public catalog (cacheable, short TTL)** | `GET /api/v1/products`, `/products/:slug`, `/categories`, `/bundles`, `/bundles/:id`, `/reviews/:productId` | `public, max-age=60, stale-while-revalidate=300` | admin edits become visible within ~60s at the CDN edge without needing a purge call — acceptable staleness for a catalog, not for money |
| **Public, rarely-changing** | `GET /api/v1/exchange-rates` | `public, max-age=300, stale-while-revalidate=600` | admin-set rates change infrequently |
| **Generated previews/thumbnails** | `GET /api/v1/media/:id` (preview/thumbnail variants) | `public, max-age=86400, immutable` if filename is content-hashed; otherwise `max-age=3600` | never cache the **full-resolution asset download** endpoint this way — see next row |
| **Never cache** | anything under `/orders`, `/cart`, `/wishlist`, `/recently-viewed`, `/compare`, `/referrals`, `/me`, `/admin/*`, `/guest-orders/:id`, the asset **download** endpoint `/orders/:id/download/:itemId`, all coupon validation, all POST/PUT/DELETE | `private, no-store` | auth-scoped, payment-scoped, or mutating — a CDN caching any of these is a data leak or a correctness bug, not a perf win |
| **Static frontend assets** (JS/CSS/images with content-hashed filenames, e.g. `app.a1b2c3.js`) | any MFE's `/assets/*` | `public, max-age=31536000, immutable` | safe because the filename changes when content changes |
| **HTML pages (`index.html`, route entrypoints)** | every MFE route | `no-cache` (revalidate every time, cheap because the page is small and mostly chrome) | ensures a redeploy is visible immediately even with a CDN in front |
| **`env.js`** | every MFE | `no-store` | must always reflect the live Deployment env vars, never cached |
| **Sitemap / robots** | `/sitemap.xml`, `/robots.txt` | `public, max-age=3600` | |

## 2. Applying headers on the backend (Go/Gin middleware)

```go
// middleware/cache.go
func CacheControl(directive string) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Cache-Control", directive)
        c.Next()
    }
}

// wiring, per route group:
public := router.Group("/api/v1")
{
    public.GET("/products", CacheControl("public, max-age=60, stale-while-revalidate=300"), h.ListProducts)
    public.GET("/products/:slug", CacheControl("public, max-age=60, stale-while-revalidate=300"), h.GetProduct)
    public.GET("/exchange-rates", CacheControl("public, max-age=300, stale-while-revalidate=600"), h.ListRates)
}

authed := router.Group("/api/v1")
authed.Use(middleware.RequireAuth())
{
    authed.GET("/orders", CacheControl("private, no-store"), h.ListOrders)
    // repeat no-store explicitly on every authed/mutating route rather than
    // relying on a default — being explicit here is cheap insurance
}
```

## 3. Applying headers on static assets (nginx, per MFE)

```nginx
# in each MFE's nginx.conf, inside the server block
location /assets/ {
    add_header Cache-Control "public, max-age=31536000, immutable" always;
}
location = /env.js {
    add_header Cache-Control "no-store" always;
}
location / {
    add_header Cache-Control "no-cache" always;   # HTML entrypoints
    try_files $uri $uri/ /index.html;
}
```

## 4. CDN configuration notes (example: Cloudflare free tier)

1. Point DNS for `pawradise.ir` at RED (proxied/orange-cloud).
2. Cloudflare respects origin `Cache-Control` by default for static
   extensions and for anything explicitly cacheable — no special page rules
   needed for the "Never cache" and "HTML pages" rows above, since
   `no-store`/`no-cache`/`private` are already correct signals.
3. For the "Public catalog" and "rarely-changing" rows, if you find the
   default edge behavior isn't caching them (Cloudflare's default cache
   level sometimes only caches static extensions), add a **Cache Rule**
   scoped to `pawradise.ir/api/v1/products*`, `/api/v1/categories*`,
   `/api/v1/bundles*`, `/api/v1/exchange-rates*` → "Eligible for cache,"
   respecting origin TTL.
4. Do **not** create a blanket "Cache Everything" rule for `/api/v1/*` — it
   would catch the never-cache endpoints too. Scope rules to the specific
   paths above.
5. Staging (`/staging/*`) should generally be excluded from CDN caching
   entirely (staging is for verification, staleness there is actively
   unhelpful) — add a Cache Rule matching `/staging/*` → "Bypass cache."
6. No purge-on-publish integration is required for v1 — the `max-age=60`
   ceiling on catalog data is the invalidation strategy. If admin edits
   need to appear instantly later, add a Cloudflare API purge call
   (`POST /zones/:id/purge_cache` with the specific URL) from
   `admin-service` after a product/bundle/coupon mutation — this is a
   small, isolated addition and does not need to be built now.

## 5. Verifying it worked

Add to the Playwright suite (or a small curl script) a check that hits a
catalog endpoint twice and confirms a `CF-Cache-Status: HIT` (or your CDN's
equivalent header) on the second request, and confirms `private, no-store`
plus a cache-miss/bypass status on an authenticated endpoint. This belongs
in `tests/frontend-e2e/` alongside the functional specs, run only against
the staging or production `BASE_URL`, not local dev (no CDN in front of
local dev).
