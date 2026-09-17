
# Testing Rule: Always Use Real Browser Against Staging Domain

## The Mistake We Made

### Mistake 1: Curl-only frontend testing
Early testing relied on curl-based checks against API endpoints, which only verified HTTP status codes. This missed critical issues:
- **Nginx welcome pages** returned HTTP 200 but were not the actual application
- **SPA routing failures** — pages with correct HTTP 200 had no rendered content because JavaScript couldn't execute
- **Link rewriting bugs** — env.template.js added `/staging` prefix to all URLs, breaking API calls silently
- **Missing assets** — auth-mfe had missing CSS/JS files causing white screens, invisible to curl

### Mistake 2: HTTP status-only API testing (September 2026)
The e2e-suite.js Node.js HTTP tests initially only checked response status codes and superficial field presence. This missed critical issues:
- **Wrong field paths** — tests expected `data.title` but API returned `data.product.title` (nested under `product` key)
- **Wrong response structure** — tests expected `data.related_products` but API returned `data.recommendations.products`
- **Missing fields assumed present** — tests asserted fields like `added` in wishlist response but API returned `message` instead
- **Status code mismatches** — tests expected 201 but API returned 200 for cart operations
- **Null array handling** — tests called `assertArray(data, 'products')` on null values without null guards

**Root cause:** Tests were written to match an assumed/spec'd API shape without verifying against the actual running API. When the implementation diverged from the spec, tests failed not because of bugs but because of expectation mismatches.

## The Rule
> **ALL frontend tests MUST use a real browser (Playwright) against the staging domain URL.**
> 
> HTTP 200 alone is insufficient — tests must verify:
> 1. Page content is NOT the nginx welcome page
> 2. Page has rendered content (DOM elements, not blank/white screen)
> 3. Browser JavaScript executed (Alpine.js components, API calls)
> 4. User flows work end-to-end (login, register, create post, add to cart, checkout)

## Why Curl-Only Testing Fails
1. **Curl cannot execute JavaScript** — SPA pages render with Alpine.js after page load; curl sees the raw HTML before initialization
2. **Curl sees what the server returns, not what the user sees** — nginx welcome page returns 200; curl reports success; user sees wrong page
3. **Curl cannot detect visual issues** — missing CSS, white screens, broken layouts are invisible to HTTP-level checks
4. **Curl cannot simulate user interaction** — login flows, form submissions, cart operations require browser context

## What curl CAN Be Used For
- Quick API endpoint health checks (200/401/404 status verification)
- Backend service availability verification
- Not for frontend page validation

## Testing Infrastructure
- Playwright tests: `/root/project/tests/frontend/spec-crawl-complete.spec.cjs`
- Run: `cd /root/project/tests/frontend && BASE_URL=<staging-domain> ADMIN_TOKEN=<token> npx playwright test spec-crawl-complete.spec.cjs`
- Tests use Chromium browser via Playwright
- Tests verify rendered content, not just HTTP status
- Tests cover: all public pages, auth pages, product pages, community pages, user pages, checkout, admin, API endpoints, user flows (register, login, post, cart, order, wishlist, follow), design (dark theme, header/footer)

## History
- 2026-09-03: Identified curl-only approach misses nginx welcome pages and SPA rendering issues
- 2026-09-11: Added testing rule to pawradise-backend-dev skill and this docs file
