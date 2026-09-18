---
name: store4bots-backend
description: Develop Store4bots Go microservices, APIs, and SQL migrations. Use when changing services/*, handlers, gin routes, or per-service Postgres schema.
---

# Backend

Nine Gin services. Ports: identity 8081, product 8082, commerce 8083, community 8084, review 8085, payment 8086, admin 8087, media 8088, events 8089.

Events uses Mongo (`$MONGO_DB`). All others use `appdb_<service>_staging|production`.

## Build

```bash
. scripts/lib/site-env.sh
cd services/<svc>
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test ./handlers -count=1
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o <svc>-binary .
docker build -t "$(IMAGE <svc>):<tag>" .
docker push "$(IMAGE <svc>):<tag>"
red_ssh "k3s kubectl -n ${STAGING_NS} set image deploy/<svc> <svc>=${REGISTRY_HOST}/${IMAGE_REPO}/<svc>:<tag>"
```

Alpine images need a static binary (`CGO_ENABLED=0`).

## Migrations

Files: `services/<svc>/migrations/*.sql` (and `migrations/sql/` in some services). Apply on RED:

```bash
red_ssh "docker exec -i postgres psql -U ${DB_USER} -d appdb_<svc>_staging" < services/<svc>/migrations/<file>.sql
```

Apply the same file to `_production` when promoting.

## Rules

- Public paths are `/api/v1/...`. Add the path to `k8s/api-gateway-staging.yaml` (and production ingress) **before** the more general `/api/v1/admin` prefix when needed (events admin is an example).
- Do not duplicate another service’s tables. Call the owner over HTTP.
- 401 on missing auth. Cookie `store4bots_session` and `Authorization: Bearer` are both used (identity).
- Operator APIs also accept `X-Admin-Token` / `ADMIN_TOKEN`.

## Pitfalls

- Pinned list order: `ORDER BY pinned DESC, <requested sort>`.
- Count queries for pagination must use the **same filters** as the list query, without LIMIT.
- `kubectl` Service selectors must be `app: <name>` only — a stray `namespace:` label in the selector yields empty Endpoints and 503.
- After `set image` with the same tag, delete the pod if `imagePullPolicy` is IfNotPresent and the digest changed.
- Community `POST /community/posts` accepts optional `image_url` (http(s) or `data:image/{png,jpeg,jpg,gif,webp};base64,…`, ~900KB). The first public `http(s)` URL in `content` is fetched for Open Graph tags (4s timeout, 256KB, no loopback/private IPs). Migration `005_post_media.sql` adds `image_url` / `link_*` on `community_posts`.

## Spec gaps

Work `docs/SPEC-GAPS.md` lowest FIX-ID: failing test first, then code, then delete that section.
