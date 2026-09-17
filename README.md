# Digital Assets Shop (Pawradise)

Single-seller digital asset marketplace: Go microservices + Alpine.js microfrontends.

## Structure

```
digital_assets_shop/
├── docs/              # Spec, architecture, testing guides
├── frontend/          # 7 MFEs (shop, product, community, account, checkout, auth, admin-app)
├── k8s/               # Kubernetes manifests
├── release-notes/     # Change history
├── services/          # 8 Go microservices + shared library
├── shared/            # Theme, api.js, chrome, nginx helpers
└── tests/             # Go unit/integration, Node API suite, Playwright
```

## Environments

| | Host | API |
|---|---|---|
| Staging | `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir` | `/api/v1` |
| Production | `pawradise.ir` | `/api/v1` |

Environments are split by hostname, not a `/staging` path prefix.

## Frontend

Each MFE is static HTML + vanilla JS (not React, not a composed Shell). Shared client: `shared/lib/api.js`. Alpine.js remains vendored for older tests that look for the file.

## Testing

```bash
cd tests && node e2e-suite.js
cd tests && npx playwright test e2e.spec.js
cd tests/frontend && npx playwright test
cd services/product-service && go test ./...
cd tests && go test ./unit/... ./integration/...
```

Frontend tests should use a real browser against the staging domain (`docs/TESTING-GUIDE.md`, `docs/TESTING-RULE.md`).

## Documentation

- `docs/PRODUCT-SPEC.md` — product, flows, pages, API, schema
- `docs/FEATURE-CANDIDATES.md` — feature in/out decisions
- `docs/MICROSERVICE-ARCHITECTURE.md` — service and MFE boundaries
- `docs/DEPLOYMENT-ARCHITECTURE.md` — BLUE/RED, ingress, build
- `docs/API-INTERFACE.md` — environments and auth for frontend work
- `docs/staging-routing.md` — staging ingress snapshot
