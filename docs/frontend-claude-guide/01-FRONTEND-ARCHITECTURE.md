# Pawradise — Frontend Architecture Implementation Guide

> Phase 4. Implements the ADRs from `00-ROADMAP-AND-DECISIONS.md`. Read that
> first for *why*; this is the *how*.

---

## 1. MFE inventory & routes

Unchanged from `MICROSERVICE-ARCHITECTURE.md` §3, except Admin is explicitly
**not** an MFE (own doc: `05-ADMIN-PANEL-SPEC.md`):

| MFE | Routes owned | Talks to (services) |
|---|---|---|
| auth-mfe | `/login`, `/register` | Identity |
| shop-mfe | `/`, `/category/:slug` | Product |
| product-mfe | `/product/:slug`, `/bundle/:id` | Product, Review (ratings), Commerce (recommendations) |
| account-mfe | `/account`, `/referrals`, `/cart`, `/wishlist` | Identity, Commerce |
| checkout-mfe | `/checkout` | Commerce, Payment |
| community-mfe | `/community`, `/post/:id`, `/profile/:id` | Community, Identity (public profile) |
| admin-app | `/admin` | Admin (which aggregates all) — separate doc |

## 2. Build-time chrome assembly (ADR-4)

`shared/chrome/` contains:
```
shared/chrome/
  header.html        <!-- nav, logo, auth badge slot, cart badge slot, theme.css link -->
  footer.html
  app.js              // hydrates the auth/cart badges after page load
shared/theme/theme.css  // generated, see 03-DESIGN-SYSTEM.md
shared/vendor/alpine.min.js
shared/lib/api.js       // shared fetch wrapper
shared/lib/coupon-math.js
shared/lib/currency.js
```

Each MFE has a tiny `build.sh` (identical shape across MFEs, copy-paste and
adjust the destination):

```bash
#!/usr/bin/env bash
set -euo pipefail
DIST=dist
rm -rf "$DIST" && mkdir -p "$DIST/assets"

# 1. this MFE's own pages
cp -r src/* "$DIST/"

# 2. shared chrome + theme + vendor + lib, identical across every MFE
cp ../../shared/chrome/*.html "$DIST/assets/"
cp ../../shared/chrome/app.js "$DIST/assets/"
cp ../../shared/theme/theme.css "$DIST/assets/"
cp ../../shared/vendor/alpine.min.js "$DIST/assets/"
cp ../../shared/lib/*.js "$DIST/assets/"

echo "built $DIST"
```

Each page template includes the shared chrome the same way every time:
```html
<head>
  <link rel="stylesheet" href="/assets/theme.css">
  <script src="/env.js"></script>              <!-- see §4 -->
</head>
<body>
  <div id="chrome-header"></div>                <!-- filled by app.js via fetch('/assets/header.html') at load, or simplest: inline the include at build time too -->
  <main>...page content...</main>
  <div id="chrome-footer"></div>
  <script src="/assets/alpine.min.js" defer></script>
  <script src="/assets/api.js"></script>
  <script src="/assets/app.js"></script>
</body>
```
Simplest correct option: have `build.sh` literally inline `header.html`/
`footer.html` into each page's HTML at build time (string substitution on a
`<!-- CHROME_HEADER -->` marker) rather than fetching them client-side —
zero runtime cost, zero flash-of-missing-chrome, and it's still "shared" in
the sense that matters (one source file, copied everywhere on rebuild).

## 3. Ingress routing (ADR-7)

Extend the existing production/staging Ingress resources
(`DEPLOYMENT-ARCHITECTURE.md` §3.2) with one rule per MFE. Production —
no rewrite, direct path match:

```yaml
# k8s/frontend-production-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: frontend-gateway
  namespace: production
  annotations:
    kubernetes.io/ingress.class: nginx
spec:
  rules:
  - http:
      paths:
      - path: /
        pathType: Exact
        backend: { service: { name: shop-mfe, port: { number: 80 } } }
      - path: /category
        pathType: Prefix
        backend: { service: { name: shop-mfe, port: { number: 80 } } }
      - path: /product
        pathType: Prefix
        backend: { service: { name: product-mfe, port: { number: 80 } } }
      - path: /bundle
        pathType: Prefix
        backend: { service: { name: product-mfe, port: { number: 80 } } }
      - path: /account
        pathType: Prefix
        backend: { service: { name: account-mfe, port: { number: 80 } } }
      - path: /referrals
        pathType: Prefix
        backend: { service: { name: account-mfe, port: { number: 80 } } }
      - path: /cart
        pathType: Prefix
        backend: { service: { name: account-mfe, port: { number: 80 } } }
      - path: /wishlist
        pathType: Prefix
        backend: { service: { name: account-mfe, port: { number: 80 } } }
      - path: /checkout
        pathType: Prefix
        backend: { service: { name: checkout-mfe, port: { number: 80 } } }
      - path: /community
        pathType: Prefix
        backend: { service: { name: community-mfe, port: { number: 80 } } }
      - path: /post
        pathType: Prefix
        backend: { service: { name: community-mfe, port: { number: 80 } } }
      - path: /profile
        pathType: Prefix
        backend: { service: { name: community-mfe, port: { number: 80 } } }
      - path: /login
        pathType: Prefix
        backend: { service: { name: auth-mfe, port: { number: 80 } } }
      - path: /register
        pathType: Prefix
        backend: { service: { name: auth-mfe, port: { number: 80 } } }
      - path: /admin
        pathType: Prefix
        backend: { service: { name: admin-app, port: { number: 80 } } }
```

Staging mirrors this with the exact same rewrite mechanism already used for
the backend (`DEPLOYMENT-ARCHITECTURE.md` §3.2), pointed at the staging
Services instead:

```yaml
# k8s/frontend-staging-ingress.yaml — one rule per route, same shape as prod
metadata:
  name: frontend-gateway-staging
  namespace: staging
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/rewrite-target: /$1
spec:
  rules:
  - http:
      paths:
      - path: /staging(/?)$
        pathType: ImplementationSpecific
        backend: { service: { name: shop-mfe, port: { number: 80 } } }
      - path: /staging/product(/.*)?
        pathType: ImplementationSpecific
        backend: { service: { name: product-mfe, port: { number: 80 } } }
      # ...one such rule per route, mirroring the production list above...
```

Each MFE therefore exists as **two Deployments + two Services** (one in
`production`, one in `staging`), exactly like the backend already does —
same pattern, same operational muscle memory.

## 4. Runtime env injection (ADR-8)

Every MFE's Dockerfile:
```dockerfile
FROM nginx:1.27-alpine
COPY dist/ /usr/share/nginx/html/
COPY env.template.js /usr/share/nginx/html/env.template.js
COPY docker-entrypoint-env.sh /docker-entrypoint.d/40-pawradise-env.sh
RUN chmod +x /docker-entrypoint.d/40-pawradise-env.sh
```

```bash
#!/bin/sh
# docker-entrypoint-env.sh — runs automatically, nginx's official image
# executes everything in /docker-entrypoint.d/ before starting nginx.
set -eu
envsubst '${API_BASE} ${IS_STAGING} ${ENV_NAME}' \
  < /usr/share/nginx/html/env.template.js \
  > /usr/share/nginx/html/env.js
```

Deployment env vars (only line that differs between staging/production
manifests, everything else — image, ports, resources — identical):

```yaml
# production
env:
  - { name: API_BASE,   value: "/api/v1" }
  - { name: IS_STAGING, value: "false" }
  - { name: ENV_NAME,   value: "production" }
---
# staging
env:
  - { name: API_BASE,   value: "/staging/api/v1" }
  - { name: IS_STAGING, value: "true" }
  - { name: ENV_NAME,   value: "staging" }
```

`shared/lib/api.js` reads this once:
```js
// shared/lib/api.js
const ENV = window.__PAWRADISE_ENV__ ?? { apiBase: '/api/v1' };

export class ApiError extends Error {
  constructor(status, message) { super(message); this.status = status; }
}

export async function apiFetch(path, options = {}) {
  const res = await fetch(`${ENV.apiBase}${path}`, {
    credentials: 'include',                 // ADR-2: cookie auth, no token handling here
    headers: { 'Content-Type': 'application/json', ...options.headers },
    ...options,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(res.status, body.error ?? res.statusText);
  }
  return res.status === 204 ? null : res.json();
}
```

Every MFE imports this one module — never constructs `API_BASE` itself,
never checks `window.location.pathname` for staging detection. This is the
one and only place environment-awareness lives.

## 5. Auth cookie (ADR-2) — backend-side change required

In `identity-service`'s `handlers/auth.go`, after issuing the JWT in
Register/Login, add:

```go
c.SetCookie(
    "pawradise_session", token,
    30*24*3600,       // 30 days, matches JWT expiry
    "/", "",          // path, domain (empty = current host)
    true, true,        // secure, httpOnly
)
// SameSite=Lax — Gin's SetCookie doesn't expose SameSite directly pre-1.10;
// set it via c.Writer.Header().Set("Set-Cookie", ...) with SameSite=Lax
// appended, or upgrade to a Gin version that supports c.SetSameSite(http.SameSiteLaxMode)
// before calling SetCookie.
```

In `handlers/middleware.go`'s JWT check, fall back to the cookie when no
`Authorization` header is present:

```go
func extractToken(c *gin.Context) string {
    if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
        return strings.TrimPrefix(auth, "Bearer ")
    }
    if cookie, err := c.Cookie("pawradise_session"); err == nil {
        return cookie
    }
    return ""
}
```

Logout clears the cookie (`MaxAge: -1`) in addition to the existing
`invalidated_tokens` insert.

This is the **only** backend change this set of documents requires — every
other frontend concern is solvable entirely within the frontend layer.

## 6. Storage recap

See `00-ROADMAP-AND-DECISIONS.md` ADR-9 for the full matrix; implementation
notes:
- Never write the JWT to `localStorage`/`sessionStorage` anywhere in any
  MFE — if you find yourself doing this, the cookie flow above isn't wired
  correctly yet.
- Cart/wishlist badge counts: `sessionStorage.setItem('badge:cart', JSON.stringify({count, at: Date.now()}))`,
  treated stale after 60s or invalidated immediately by a
  `document.dispatchEvent(new CustomEvent('cart:changed'))` fired by any
  Alpine component that mutates the cart on the current page.

## 7. Same-page reactivity where full navigation isn't right

Two flows in `02-UX-FLOWS.md` explicitly call for same-page updates: liking
a post, and live tier/PWYW price updates on the product page. Handle these
with a plain Alpine component, no extra library:

```html
<div x-data="{ liked: false, count: 3 }">
  <button
    :aria-pressed="liked"
    @click="
      liked = !liked;
      count += liked ? 1 : -1;
      apiFetch(`/community/posts/${postId}/like`, { method: liked ? 'POST' : 'DELETE' })
        .catch(() => { liked = !liked; count += liked ? 1 : -1; }) /* rollback on failure */
    ">
    <span x-text="liked ? 'Liked' : 'Like'"></span> (<span x-text="count"></span>)
  </button>
</div>
```

This is the extent of "SPA-like" behavior in the whole system — optimistic,
local, single-component, rolled back on failure. It does not require, and
should not grow into, cross-MFE state sharing.

## 8. CDN hook

Static asset responses from every MFE's nginx should set aggressive cache
headers for hashed assets and none for HTML — see `06-CDN-AND-CACHING.md`
for the exact `nginx.conf` snippet and the CDN configuration in front of
ingress.
