# Pawradise — Roadmap & Architecture Decisions

> Read this file first. It sequences the work and records the key decisions
> (with reasoning) that every other document assumes. Give this whole folder
> to your implementation agent, in this order.

---

## 1. Why this order

You cannot design good UI before you know exactly what the UI has to *do*,
and you cannot write good code without a test that defines "done." So the
agent should work in this sequence, not jump straight to styling:

| Phase | Doc | Output |
|---|---|---|
| 0 | this file | shared decisions, no ambiguity later |
| 1 | `02-UX-FLOWS.md` | plain-language flows + testable acceptance criteria, **no styling** |
| 2 | `03-DESIGN-SYSTEM.md` | tokens, look & feel, responsive rules |
| 3 | `04-TDD-GUIDE.md` | failing tests written from Phase 1's acceptance criteria |
| 4 | `01-FRONTEND-ARCHITECTURE.md` | implement pages to make tests pass, using Phase 2's tokens |
| 5 | `05-ADMIN-PANEL-SPEC.md` | same TDD loop, applied to the standalone admin app |
| 6 | `06-CDN-AND-CACHING.md` | perf pass once functionality is correct |

Within Phase 4, build one MFE at a time, in this order (cheapest to
riskiest, and each one unblocks the next in manual testing):
**auth-mfe → shop-mfe → product-mfe → account-mfe (cart/wishlist) →
checkout-mfe → community-mfe → admin-app**.

---

## 2. Architecture Decision Records (ADRs)

### ADR-1 — Microfrontends are MPA pages behind ingress, not a client-composed SPA shell

**Decision:** Each MFE is a full, independent HTML page served by its own
nginx:alpine container. Navigating between them is a normal browser
navigation (full page load), not client-side routing. This is a Multi-Page
App made of independently-deployable pages, matching your instruction that
frontends should be routed by ingress "the same way as backend services,"
not via nginx-per-path static hosting.

**Why not Module Federation / single-spa / a JS-composed shell:** those
solve a problem you don't have (many independent teams shipping into one
live SPA session) and cost you a real problem you do have: RED has ~2GB RAM
and a weak/free coding model as the implementer. A runtime JS composition
framework adds a whole class of failure modes (shared dependency version
skew, lifecycle bugs, memory leaks across mounts) that are hard for a
less-capable agent to debug. Full page navigation between MFEs is invisible
to the user for a shop like this — nobody is switching between
Shop/Checkout/Community fast enough to need SPA-smooth transitions — and it
makes auth, caching, and independent deploys trivial (see ADR-2, ADR-3).

**Consequence:** shared chrome (header/footer/nav/theme) cannot be "mounted"
at runtime by a host app. It's assembled at **build time** — see ADR-4.

### ADR-2 — Auth is a same-origin, httpOnly cookie; MFEs do not manage tokens

**Decision:** `identity-service` continues to return `{ token, user }` in
the JSON body (unchanged, for API/admin/Bearer-token clients), but on
`/register` and `/login` it **also** sets the JWT as a cookie:

```
Set-Cookie: pawradise_session=<jwt>; Path=/; HttpOnly; Secure; SameSite=Lax; Max-Age=2592000
```

Every MFE, and every fetch call, uses `credentials: 'include'`. Because all
pages and all API routes live under the same origin (`pawradise.ir`), the
browser attaches this cookie automatically — no MFE needs to read, store, or
pass around a token, and there is no cross-MFE JS state to synchronize.

`middleware.go`'s JWT check is extended to accept the token from **either**
the `Authorization: Bearer` header (admin tooling, tests, future mobile app)
**or** the `pawradise_session` cookie (browser pages) — whichever is
present, cookie checked second.

**Why not localStorage + manual header injection per MFE:** that requires
every MFE to import the same auth module and every page navigation to
re-hydrate it, and it's readable by any injected script (XSS risk). httpOnly
cookies remove an entire category of bugs and an entire shared-library
dependency, at zero runtime cost, because the pages are already same-origin.

**Logout:** clears the cookie (`Max-Age=0`) *and* still invalidates the
token server-side (existing `invalidated_tokens` table), so a stolen
Bearer-header token is also killed.

### ADR-3 — Client interactivity: Alpine.js, vendored, no bundler required

**Decision:** every MFE and the Admin app use [Alpine.js](https://alpinejs.dev)
(~15KB) for interactivity — loops, forms, toggles, fetch-driven state — via
a single vendored `/shared/vendor/alpine.min.js` file (not a CDN `<script
src>`, so production doesn't depend on a third party being up, matching
your "no external font loading / substance over style" direction). Pages are
otherwise plain HTML/CSS + small hand-written JS modules.

**Why Alpine over React/Vue/Svelte with a build step:** your backend is a
JSON API, not server-rendered HTML, so htmx-style approaches would require
backend changes. Alpine gives declarative, JSON-driven UI without a bundler,
a `node_modules` tree, or build tooling the agent has to get right — every
page is inspectable as plain HTML. A build step becomes a **choice**, not a
requirement: use Vite optionally for the Admin app only if the CRUD screens
get unwieldy (see `05-ADMIN-PANEL-SPEC.md`).

### ADR-4 — Shared chrome is assembled at build time, not via runtime SSI or subrequests

**Decision:** a `shared/chrome/` directory (header, footer, nav, theme CSS,
`app.js` for the auth/cart badges) lives once in the monorepo. Each MFE's
build script (run on BLUE, before `docker build`) copies these files into
that MFE's `dist/` alongside its own page markup, so every deployed
container is fully self-contained static files — no pod ever calls another
pod to render a page.

**Why not nginx SSI (`<!--#include virtual -->`)** even though the original
architecture doc mentions it: SSI would require every MFE pod to reach the
shell pod at request time, which means a shell outage takes every page down
with it, and it adds nginx config surface (`ssi on;`, internal proxy_pass,
resolver directives) that's easy to misconfigure. A copy-step in a ~15-line
build script is far more robust and just as capable of being iterated
(`shared/chrome/` changes propagate to all MFEs the next time each is
rebuilt and deployed — independent deploys are preserved).

### ADR-5 — Look & feel is data, not code

**Decision:** all colors, spacing, radii, fonts, and glow intensity are CSS
custom properties generated from one `theme.tokens.json` file (see
`03-DESIGN-SYSTEM.md`). Changing the whole site's look is editing one JSON
file and rerunning a tiny generator script — no HTML/CSS is touched in any
MFE. This directly satisfies "styling should be configurable, we shouldn't
have to stick to one thing."

### ADR-6 — No WebSockets in v1; reserve the pattern

**Decision:** the product spec explicitly excludes real-time chat/push
notifications from scope, and checkout uses **polling**
(`GET /orders/:id/status`) for payment confirmation, which is fine at this
scale. Do not build a socket layer now.

**If/when it's needed** (e.g., live payment confirmation instead of
polling): put a single lightweight events gateway behind
`wss://pawradise.ir/ws`, routed through the same ingress. Because pages are
MPA (ADR-1), only the **currently open page** needs a connection — there is
no cross-MFE socket-sharing problem to solve, since only one MFE page is
ever loaded in the tab at a time. A page opens the socket on load, closes it
on unload. If you later add multi-tab awareness, use `BroadcastChannel`
(same-origin, no library needed) so only one tab holds the live connection.

### ADR-7 — Frontend routing mirrors backend routing exactly

**Decision:** frontend Ingress rules use the identical pattern already
proven for the backend in `DEPLOYMENT-ARCHITECTURE.md` §3: production paths
route directly, staging paths are prefixed `/staging/...` and rewritten
(`rewrite-target: /$1`) before hitting the staging MFE pod. See
`01-FRONTEND-ARCHITECTURE.md` §3 for the full routing table and YAML.

Because the rewrite happens server-side, **the browser's address bar still
shows `/staging/...`**. Combined with a build-time-injected `env.js` (see
ADR-8), the staging deployment doesn't need to sniff the URL at all — it's
simply told at deploy time that it's staging.

### ADR-8 — Runtime environment is injected via `env.js`, not baked into the JS bundle

**Decision:** every MFE image ships an `env.template.js`. The container
entrypoint runs `envsubst` against it using Deployment-level env vars
(`API_BASE`, `IS_STAGING`, `ENV_NAME`) and writes `env.js` before nginx
starts — the same image is deployed to both staging and production, only
the env vars differ. Pages `<script src="/env.js">` before any other script.

```js
// env.template.js -> becomes env.js at container start
window.__PAWRADISE_ENV__ = {
  apiBase: "${API_BASE}",     // "/api/v1" or "/staging/api/v1"
  isStaging: ${IS_STAGING},   // true / false
  envName: "${ENV_NAME}"      // "production" / "staging"
};
```

This also means: **the same Docker image is promoted from staging to
production**, never rebuilt per environment — a real safety property worth
keeping.

### ADR-9 — Storage matrix

| Data | Where | Why |
|---|---|---|
| JWT / session | httpOnly cookie (ADR-2) | never touchable by page JS |
| Admin Bearer token | `sessionStorage` (Admin app only) | privileged, cleared on tab close; see `05-ADMIN-PANEL-SPEC.md` |
| Theme preference (light/dark/etc. if you add one later) | `localStorage` | non-sensitive, must survive full-page navigations |
| Cart/wishlist counts for the header badge | refetched per page load from API, cached in `sessionStorage` for the tab session with a short TTL, invalidated via a `CustomEvent` on same-page mutation | avoids a redundant call on every click without any cross-MFE shared memory |
| Recently-viewed / compare-list echo | `localStorage`, mirrors server state, server is source of truth | fast optimistic UI |
| Anything with PII or payment data | never in browser storage | server + cookie only |

### ADR-10 — CDN sits in front of ingress, caching GETs only

See `06-CDN-AND-CACHING.md` for the full endpoint classification and header
recipes. Short version: static assets are cached aggressively (hashed
filenames, `immutable`), public catalog GETs are cached briefly with
`stale-while-revalidate`, and anything authenticated or mutating is never
cached.

---

## 3. Testing stack (detail in `04-TDD-GUIDE.md`)

- **Backend:** unchanged — Go `testing` + `sqlmock` for unit,
  `testcontainers` for integration, per the layout already defined in
  `MICROSERVICE-ARCHITECTURE.md` §7.
- **Frontend E2E:** Playwright — a real browser driving real full-page
  navigations is a perfect fit for the MPA architecture in ADR-1.
- **Frontend unit:** Vitest, for the handful of pure-JS modules that have
  real logic worth isolating (currency formatting, coupon math mirror, the
  shared `api.js` fetch wrapper's error handling).

## 4. Repository layout

```
/root/project/
  services/                  # existing, unchanged
    identity/ product/ commerce/ community/ review/ payment/ admin/ media/
  shared/
    chrome/                  # header.html, footer.html, nav.html, app.js
    theme/                   # theme.tokens.json, generate-theme-css.js, theme.css (generated)
    vendor/alpine.min.js
    lib/api.js                # shared fetch wrapper (copied into every MFE build)
  frontend/
    auth-mfe/       (login, register)
    shop-mfe/       (/, /category/:slug)
    product-mfe/    (/product/:slug, /bundle/:id)
    account-mfe/    (/account, /referrals, /cart, /wishlist)
    checkout-mfe/   (/checkout)
    community-mfe/  (/community, /post/:id, /profile/:id)
    admin-app/       (/admin — standalone, not an MFE, see ADR notes in its own doc)
  tests/
    unit/               # existing Go unit tests, per service
    integration/        # existing Go integration tests
    e2e/                 # existing e2e-suite.js
    frontend-e2e/        # NEW — Playwright specs, one file per flow in 02-UX-FLOWS.md
    frontend-unit/       # NEW — Vitest specs for shared/lib modules
  k8s/
    ...existing manifests...
    frontend-*.yaml        # NEW — one Deployment+Service per MFE, per 01-FRONTEND-ARCHITECTURE.md
```
