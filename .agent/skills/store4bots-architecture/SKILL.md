---
name: store4bots-architecture
description: Store4bots microservice and micro-frontend architecture, framework choices, and service boundaries. Use when adding a service, MFE, or shared contract.
---

# Architecture

Nine Go HTTP services + seven Alpine.js MFEs + host Postgres/Mongo + k3s on RED. BLUE builds; RED runs.

## Backend

Go 1.22 + Gin. Each service owns a DB `appdb_<name>_<env>`.

| Service | Port | Owns |
|---------|------|------|
| identity-service | 8081 | users, sessions, JWT, `/auth`, `/users` |
| product-service | 8082 | products, categories, search, `/products`, `/categories` |
| commerce-service | 8083 | cart, wishlist, compare, recent, coupons, `/cart`… |
| community-service | 8084 | posts, comments, people, `/community`… |
| review-service | 8085 | reviews, `/reviews` |
| payment-service | 8086 | checkout intents, `/payments` |
| admin-service | 8087 | admin catalog, settings, roles, `/admin` |
| media-service | 8088 | uploads, `/media` |
| events-service | 8089 | Mongo event log, `/admin/events` |

Public path prefix: `/api/v1`. Ingress maps those paths to the owning service. `API_BASE` is `/api/v1` on both hostnames.

## Frontend

Alpine.js MFEs, shared `shared/theme`, `shared/lib/api.js` (per-service namespaces), `shared/lib/events.js`. Each MFE is an nginx Alpine image serving static files.

| MFE | Staging paths |
|-----|----------------|
| shop-mfe | `/` |
| auth-mfe | `/login` `/register` `/auth` |
| product-mfe | `/product` |
| account-mfe | `/account` `/cart` `/wishlist` `/compare` `/recent` `/guest` `/referrals` |
| checkout-mfe | `/checkout` |
| community-mfe | `/community` `/post` `/profile` `/people` |
| admin-app | `/admin` |

Production uses the same path map on `$PRODUCTION_HOST`.

## Why this split

- Independent image tags and rollouts (shop CSS without rebuilding identity).
- Shared theme/tokens so MFEs look like one site.
- Service DBs stay isolated; no cross-DB table copies.

## Out of scope here

k3s YAML details → `store4bots-kubernetes`. DNS/TLS → `store4bots-network`. Seed/schema → `store4bots-databases`.
