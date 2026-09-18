# Test report

**Date:** 2026-09-18  
**Target:** staging `https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/`  
**Runner:** BLUE (`ssh -p 2222`), Playwright `workers=1`  
**Images:** `events-service:proto-v8`, `identity-service:proto-v8`, `admin-service:proto-v8`, `community-service:proto-v8`, `product-service:proto-v8`, `commerce-service:proto-v8.2`, `shop-mfe:proto-v8` (`/assets` includes `events.js`), `account-mfe:proto-v8`, `checkout-mfe:proto-v8`, `admin-app:proto-v8`, `product-mfe:proto-v8`, `community-mfe:proto-v8`

This is the live status file. Older pass/fail dumps in `docs/` and `release-notes/` are historical.

## This pass (events collector + checkout)

Verified on BLUE against live staging:

| Suite | Pass | Fail | Total | Pass rate |
|---|---:|---:|---:|---:|
| Go unit (events-service, identity, commerce, admin, community, product, shared) | all | 0 | — | 100% |
| Node API (`e2e-suite.js`) | 134 | 0 | 134 | 100% |
| Playwright `session-checkout.spec.cjs` (cookie referrals, header Log out, `product_view`) | 3 | 0 | 3 | 100% |

MongoDB is a host Docker `mongo:7` on RED `:27017` (same sharing pattern as host Postgres). `POST /api/v1/events` returns 202 for `product_view` / `checkout_click`. Admin Events TTL default is 1 hour.

## Last full matrix (unchanged suites not re-run in this pass)

| Suite | Pass | Fail | Total | Pass rate |
|---|---:|---:|---:|---:|
| Go unit (`tests/unit`) | 22 | 0 | 22 | 100% |
| Go integration (`tests/integration`) | 11 | 0 | 11 | 100% |
| Playwright (`e2e.spec.cjs`) | 12 | 0 | 12 | 100% |
| Playwright (`tests/frontend` remainder) | 127 | 0 | 127 | 100% |

## Commands

```bash
# handler + shared unit tests
for d in shared identity-service product-service community-service admin-service \
         payment-service review-service media-service commerce-service events-service; do
  (cd services/$d && go test ./... -count=1)
done

# tests module (spec unit + staging integration)
cd tests && go test ./unit/... ./integration -count=1

# HTTP API
node e2e-suite.js

# Playwright
npx playwright test --config playwright.config.cjs
npx playwright test --config frontend/playwright.config.cjs
```

## Spec coverage

| Spec area | Unit | Integration | API | Playwright |
|---|---|---|---|---|
| Auth register/login/JWT role+tabs | identity extras, shared jwt | RBAC login | e2e-suite Auth + RBAC | theme-rbac, e2e.spec |
| Cookie session vs Bearer | identity GetReferrals context | — | e2e-suite referrals Bearer | session-checkout `#root` after loginAs |
| Appearance palettes/fonts/icons/dims | product extras + tests/unit/product | appearance public/PUT | e2e-suite Appearance | theme-rbac, store4bots styling |
| Homepage banner slider | product extras | banner contract | e2e-suite banner PUT | banner-mobile, e2e.spec |
| RBAC staff tabs / Access operator token | middleware rbac, admin access | access operator | e2e-suite RBAC | theme-rbac Access |
| Access search + 1/2/3 col grid | — (client) | users list roles | users list role | Access search + wide/phone grid |
| Order pipeline / ops desk | admin order_steps, rbac Steps | order-steps + staff write 403 | e2e-suite order-steps CRUD/filter | theme-rbac Orders chips + Steps tab |
| Zero-due checkout + commerce coupons | commerce applyCouponDiscount | — | e2e-suite 100% coupon paid | — |
| Event collector (Mongo TTL) | events-service handlers | — | e2e-suite Events | session-checkout product_view |
| Recommendations 404 | product extras | missing product | e2e-suite 404 | fix-009-014 |
| Volume pagination / Show more | community extras | catalog/community volume | bundles/products pages | banner-mobile pagers |
| Cart/wishlist/compare auth | commerce extras | guest cart | e2e-suite | — |
| Reviews rating bounds | review handlers | — | e2e-suite reviews | — |
| Exchange rates | payment rates | — | e2e-suite rates | — |
| Media allow-list | media files | — | — | — |
