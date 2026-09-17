# Digital Assets Shop (Pawradise)

Multi-service Go backend + React microfrontend marketplace for digital assets.

## Structure

```
digital_assets_shop/
├── docs/              # Architecture, specs, testing guides
├── e2e/               # End-to-end API tests
├── frontend/          # Microfrontends (8 MFEs)
├── k8s/               # Kubernetes manifests
├── release-notes/     # Change history
├── services/          # Go microservices (8 backends)
├── shared/            # Shared libraries, themes, nginx configs
└── tests/             # Frontend and API tests
```

## Environments

- **Staging:** `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir`
- **Production:** `pawradise.ir`

## Quick Start

```bash
# Build and push all services
cd services && make build-all

# Build and push all frontend MFEs
cd frontend && bash build-all.sh

# Deploy to staging
kubectl apply -f k8s/staging/
```

## Testing

```bash
# API tests
cd e2e && npm test

# Frontend tests
cd tests/frontend && npx playwright test
```

## Documentation

- `docs/TESTING-GUIDE.md` — Rules for frontend testing
- `docs/MICROSERVICE-ARCHITECTURE.md` — Service decomposition
- `docs/staging-routing.md` — Complete routing configuration
