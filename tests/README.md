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
cd tests && npx playwright test e2e.spec.js
cd tests/unit && go test ./... -count=1
cd tests/integration && go test ./... -count=1
cd tests/frontend && npx playwright test
# handler tests next to each Go service
cd ../services/product-service && go test ./... -count=1
```
