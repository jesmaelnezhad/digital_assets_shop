# Architecture (current)

BLUE builds. RED serves. Nine **Gin** services (Go 1.22) + seven Alpine.js MFEs. Staging and production share URL paths and differ only by hostname (`$STAGING_HOST` vs `$PRODUCTION_HOST` in `config/site.env`).

Contract: `docs/PRODUCT-SPEC.md`. Decisions: `docs/FEATURE-CANDIDATES.md`. Gaps: `docs/SPEC-GAPS.md`.

## Data plane

```
Browser
  → https://$STAGING_HOST or $PRODUCTION_HOST
  → host nginx on RED (TLS)
  → 127.0.0.1:$INGRESS_HTTP_NODEPORT
  → ingress-nginx
  → Service in namespace staging | production
```

Always call the **hostname**. A raw IP to the NodePort can match the production Ingress first.

## Backend

Each service owns `appdb_<service>_<env>` on host Postgres. Events owns Mongo `$MONGO_DB`. Services talk over HTTP or events — they do not copy another service’s tables.

| Service | Port | Owns |
|---------|------|------|
| identity-service | 8081 | users, JWT, cookie `store4bots_session`, referrals |
| product-service | 8082 | products, categories, bundles, appearance, banner |
| commerce-service | 8083 | cart, wishlist, compare, recent, coupons, orders |
| community-service | 8084 | posts (text, photo, URL preview), comments, people, follows |
| review-service | 8085 | verified-purchase ratings |
| payment-service | 8086 | payments, exchange rates, settings |
| admin-service | 8087 | ops desk, RBAC, stats, email export |
| media-service | 8088 | uploads / downloads |
| events-service | 8089 | Mongo ingest (`product_view`, `checkout_click`), TTL |

Public prefix: `/api/v1`. Ingress maps paths to services. Put `/api/v1/admin/events` **before** `/api/v1/admin`.

Auth: `Authorization: Bearer` JWT and/or session cookie. Operator APIs also accept `X-Admin-Token` (`ADMIN_TOKEN` in secrets).

CORS: `ENV_NAME` + `CORS_ORIGINS` (the https origin for that hostname).

## Frontend

Static nginx Alpine images. Shared `shared/theme/theme.css`, `shared/chrome/chrome.js`, `shared/lib/api.js` (namespaced clients), `shared/lib/events.js`.

| MFE | Paths |
|-----|--------|
| shop-mfe | `/` `/category` |
| auth-mfe | `/login` `/register` `/auth` |
| product-mfe | `/product` |
| account-mfe | `/account` `/cart` `/wishlist` `/compare` `/recent` `/guest` `/referrals` |
| checkout-mfe | `/checkout` |
| community-mfe | `/community` `/post` `/profile` `/people` |
| admin-app | `/admin` |

`API_BASE=/api/v1` in both environments. `frontend/prototype/` is the design mock only.

## Cluster on RED

| Namespace | Role |
|-----------|------|
| `ingress-nginx` | ingress-nginx-controller, NodePort `$INGRESS_HTTP_NODEPORT` |
| `staging` | backends + MFEs + Ingress Host=`$STAGING_HOST` |
| `production` | backends + MFEs + Ingress Host=`$PRODUCTION_HOST` |
| `registry` | `registry:2` + nginx auth proxy; Ingress `/registry` and `/v2` |
| `database` | unused (DBs are host Docker) |

Service selectors: `app: <name>` only.

## Images

- **Built by us:** `$REGISTRY_HOST/$IMAGE_REPO/<name>:<tag>` (authenticated in-cluster registry).
- **Stock:** Docker Hub / registry.k8s.io (`postgres:16-alpine`, `mongo:7`, `registry:2`, `nginx:1.27-alpine`, ingress-nginx controller).

## Host databases

Postgres `:5432` and Mongo `:27017` on `$RED_HOST`. Pods use `DB_HOST=$RED_HOST`. Prefer firewalling DB ports to `$BLUE_HOST`. See `docs/DATABASES.md`.
