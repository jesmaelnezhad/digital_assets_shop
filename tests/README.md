# Tests

TDD against `docs/PRODUCT-SPEC.md` and `docs/ARCHITECTURE.md`. How to run: `docs/TESTING.md`.

## Layout

```
tests/
├── e2e-suite.js                 # Node HTTP API (staging)
├── unit/<service>/              # Go + sqlmock
├── integration/                 # Cross-service
└── frontend/                    # Playwright
```

```bash
. scripts/lib/site-env.sh
cd tests
BASE_URL="https://${STAGING_HOST}" node e2e-suite.js
BASE_URL="https://${STAGING_HOST}" npx playwright test --config frontend/playwright.config.cjs --workers=1
cd .. && go test ./tests/integration ./tests/unit/... -count=1
```

Current results: [`REPORT.md`](REPORT.md). Do not copy pass/fail numbers into other docs.

Staging HTTP tests expect `scripts/seed-staging.sh` (volume + roles).
