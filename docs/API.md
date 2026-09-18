# API notes (current)

Catalog and schemas: `docs/PRODUCT-SPEC.md` §4–§5. Gaps: `docs/SPEC-GAPS.md`. Gin services, prefix `/api/v1`.

Base URL: `https://$STAGING_HOST/api/v1` or `https://$PRODUCTION_HOST/api/v1` (`config/site.env`). Ingress maps paths; services themselves still listen on `/api/v1/...`.

## Auth

1. `POST /register` or `POST /login` → `{ token, user }`
2. Later requests: `Authorization: Bearer <token>` and/or cookie `store4bots_session`
3. `POST /logout` blocklists the JWT

JWT HS256 claims: `user_id`, `email`, `role`, optional `tabs`, `exp`. Roles: `customer` (legacy `user` maps here), `staff`, `admin`. Staff `tabs` is a comma-separated desk list from `users.staff_tabs`.

Operator: `X-Admin-Token: $ADMIN_TOKEN` (from secrets). `Authorization: Bearer $ADMIN_TOKEN` is a full-admin API bypass. Do not put live tokens in docs.

Missing auth on a protected route is **401**, not a skipped check.

## Service map

| Prefix | Service |
|--------|---------|
| `/register` `/login` `/logout` `/me` `/referrals` `/commissions` | identity 8081 |
| `/products` `/categories` `/bundles` `/recommendations` | product 8082 |
| `/cart` `/wishlist` `/compare` `/recently-viewed` `/orders` `/guest-orders` `/coupons` | commerce 8083 |
| `/community` `/posts` `/profile` `/users` `/follow` | community 8084 |
| `/reviews` | review 8085 |
| `/payments` `/exchange-rates` `/settings` | payment 8086 |
| `/admin` (except `/admin/events`) | admin 8087 |
| `/media` | media 8088 |
| `/events` `/admin/events` | events 8089 |

Community also exposes wrapper paths (`/posts`, `/profile`, `/users/:id`, `/follow/:userId`) that forward to `/community/...`.

## Routes the MFEs rely on

| Method | Path | Notes |
|--------|------|--------|
| GET | `/products/:slug` | slug or numeric id |
| POST | `/community/posts` | `{content,type,is_public,image_url}` — text ≤500 or a photo; first public URL gets OG title/description/image (SSRF-blocked private hosts) |
| POST | `/wishlist/toggle` `/compare/toggle` | `{product_id}` → `{added, message}` (compare max 4) |
| GET/PUT | `/products/appearance` | public GET; staff/admin PUT look dimensions |
| PUT | `/products/banner` | `{product_ids:[…]}` |
| GET | `/products/:id/tiers` | |
| GET | `/categories?all=1` | includes inactive + `product_count` |
| POST | `/products/:id/pin` `/unpin` | admin |
| GET | `/admin/orders?status=&q=` | ops desk; `by_step` |
| PUT | `/admin/orders/:id/status` | `{status}` pipeline slug |
| GET/POST/PUT/DELETE | `/admin/order-steps` | staff Orders tab may GET |
| PUT | `/admin/users/:id/access` | `{role,staff_tabs}` needs admin JWT + `X-Admin-Token` |
| POST | `/events` | `{name,session_id,path,properties}`; 202 |
| GET/PUT | `/admin/events` `/admin/events/ttl` | ingress before `/admin` |

Frontend clients: `Store4bots.api.<ns>.<method>()` in `shared/lib/api.js`, never ad-hoc `api.get('/categories')`.
