# Pawradise Microservice Architecture

## Quick Start

```bash
# Build all services
make build

# Push to registry
make push

# Deploy to k3s
make deploy

# Run tests
make test
```

## Architecture

```
pawradise/
├── archive/                    # Old monolith code (preserved for reference)
│   ├── backend/
│   └── frontend/
├── docker/                     # Docker Compose for local dev
├── k8s/                        # Kubernetes manifests
│   ├── production/
│   └── staging/
├── frontend/                   # Microfrontends (8 MFEs + Shell)
│   ├── shell/
│   ├── shop/
│   ├── product-detail/
│   ├── community/
│   ├── account/
│   ├── checkout/
│   ├── auth/
│   └── admin/
├── services/                   # Backend microservices (8 services + shared)
│   ├── shared/                 # Shared packages
│   │   ├── auth/              # JWT, password hashing
│   │   ├── database/          # DB connection utilities
│   │   ├── events/            # Event bus (Redis/PostgreSQL)
│   │   ├── grpc/              # gRPC proto + interceptors
│   │   ├── middleware/        # Auth, CORS, recovery, logging
│   │   └── models/            # Shared data models
│   ├── identity-service/      # Auth, users, profiles, referrals
│   ├── product-service/       # Products, categories, bundles, images
│   ├── commerce-service/      # Orders, cart, wishlist, coupons
│   ├── community-service/     # Posts, comments, likes, follows
│   ├── review-service/        # Product ratings
│   ├── payment-service/       # Payments, exchange rates
│   ├── admin-service/         # Admin operations, stats, settings
│   └── media-service/         # File upload/download, image processing
└── tests/                      # TDD test suite
    ├── unit/                  # 83 unit tests (sqlmock)
    ├── integration/           # 14 integration tests
    └── e2e-suite.js           # 104 E2E tests (Node.js)
```

## Service Ports

| Service | HTTP | gRPC |
|---------|------|------|
| Identity | 8081 | 9081 |
| Product | 8082 | 9082 |
| Commerce | 8083 | 9083 |
| Community | 8084 | 9084 |
| Review | 8085 | 9085 |
| Payment | 8086 | 9086 |
| Admin | 8087 | 9087 |
| Media | 8088 | 9088 |

## TDD Workflow

1. All 207 tests FAIL (by design)
2. Implement feature in the corresponding microservice
3. Run tests → some PASS, some FAIL
4. Fix failures until all tests PASS
5. Next feature

## Shared Packages

All services import shared packages from `services/shared/`:
- `auth/jwt.go` — JWT generation and validation
- `auth/password.go` — bcrypt password hashing
- `database/db.go` — PostgreSQL connection pool
- `events/publisher.go` — Event bus for cross-service consistency
- `grpc/server.go` — gRPC server/client infrastructure
- `middleware/` — Auth, CORS, recovery, logging middleware
- `models/` — Shared data models (User, Product, Order, etc.)

## License

Proprietary — Pawradise
