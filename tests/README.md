# Tests

TDD against `docs/PRODUCT-SPEC.md` and `docs/MICROSERVICE-ARCHITECTURE.md`.

## Layout

```
tests/
├── e2e-suite.js                 # Node HTTP API (staging)
├── unit/<service>/              # Go + sqlmock
├── integration/                 # Cross-service
└── frontend/                    # Playwright
```

## Commands

```bash
cd tests && node e2e-suite.js
cd tests && npx playwright test --config playwright.config.cjs
cd tests && npx playwright test --config frontend/playwright.config.cjs
cd tests && go test ./integration ./unit/... -count=1
# handler tests next to each Go service
cd ../services/product-service && go test ./... -count=1
cd ../services/community-service && go test ./... -count=1
cd ../services/events-service && go test ./... -count=1
cd ../services/identity-service && go test ./... -count=1
cd ../services/commerce-service && go test ./... -count=1
```

Current results: [`REPORT.md`](REPORT.md) (counts, pass rate, spec coverage). Do not copy pass/fail numbers into other docs.

Staging HTTP tests (`e2e-suite.js`, `e2e.spec.cjs`, `tests/frontend`, `tests/integration`) expect the volume seed: 2+ banner slides, 4+ shop pages at 12 per page, feed/people **Show more**, and extra bundles. Seed with `scripts/seed-staging-volume.sql` and `scripts/seed-staging-roles.sql` (Nia admin, Leo staff). Access search + wide/mobile grids are covered in `tests/e2e.spec.cjs`, `tests/frontend/theme-rbac.spec.cjs`, and `tests/frontend/banner-mobile.spec.cjs`.

Handler unit tests live next to each Go service (`services/*/handlers/*_test.go`) plus `services/shared`. `tests/unit` covers spec catalogs and JWT/RBAC contracts via `replace github.com/pawradise/shared`.
