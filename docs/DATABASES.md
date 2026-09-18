# Databases (current)

Skill: `store4bots-databases`. Hosts and users: `config/site.env`. Passwords: `config/site.secrets.env`.

## Where they run

**Host Docker on RED**, not k8s StatefulSets. Namespace `database` is empty. YAML under `k8s/database/` is unused on this install.

| Engine | Image | Publish | Auth |
|--------|--------|---------|------|
| PostgreSQL 16 | `postgres:16-alpine` | `$RED_HOST:$DB_PORT` | user `$DB_USER` |
| MongoDB 7 | `mongo:7` | `$RED_HOST:$MONGO_PORT` | user `$MONGO_USER`, authSource `admin` |

Pods set `DB_HOST=$RED_HOST`. Prefer UFW so only `$BLUE_HOST` (and localhost) can reach 5432/27017.

Install: `scripts/install-host-db.sh`.

## Postgres logical databases

For each of identity, product, commerce, community, review, payment, admin, media × `{staging,production}` → `appdb_<service>_<env>` (16 DBs).

Events does not use Postgres.

## Mongo

One instance, database `$MONGO_DB` (default `events`). Collections include `events` with a TTL index on `expire_at`. Admin Events tab sets TTL. Used only by events-service.

`MONGO_URI` is in `store4bots-secrets`, not in git.

## Migrations

SQL under `services/<svc>/migrations/`. Apply staging first, production when promoting:

```bash
scripts/migrate.sh identity-service staging
scripts/migrate.sh identity-service production
```

Do not copy `products` / `users` / `orders` into another service’s database.

## Seed (staging only)

```bash
scripts/seed-staging.sh
```

Applies `scripts/seed-staging-*.sql` in order (identity → product/admin catalog → commerce → community → settings → roles → volume). Volume + roles are required for pagination, Show more, and Nia/Leo RBAC tests.

Demo users (password = local-part): nia (admin), leo (staff), maya, owen. Coupons: `SAVE12`, `WELCOME`, `MARBLE`.

Do not run staging demo seeds against production.
