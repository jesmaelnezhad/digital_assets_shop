---
name: store4bots-databases
description: Store4bots host Postgres and MongoDB, schemas, migrations, and staging seeds. Use when changing SQL, seeding staging, or installing databases on RED.
---

# Databases

Current install: **host Docker on RED**, published on `$RED_HOST:5432` and `:27017`. k8s `database` namespace is empty.

```bash
. scripts/lib/site-env.sh
red_ssh "docker exec -i postgres psql -U ${DB_USER} -d appdb_identity_staging -c '\\dt'"
```

Password: `POSTGRES_PASSWORD` / `MONGO_PASSWORD` in `config/site.secrets.env`.

## Postgres logical DBs

`appdb_<service>_staging` and `appdb_<service>_production` for identity, product, commerce, community, review, payment, admin, media (16 DBs). Create via `scripts/install-host-db.sh`.

## Mongo

Database `$MONGO_DB` (default `events`), user `$MONGO_USER`, authSource `admin`. Used only by events-service. TTL index on `expire_at`.

## Migrations

`services/<svc>/migrations/`. Apply to staging, then production when promoting:

```bash
scripts/migrate.sh identity-service staging
```

## Seed (staging only)

```bash
scripts/seed-staging.sh
```

Runs the SQL under `scripts/seed-staging-*.sql` in order (identity, product, admin catalog, commerce, community, community-ids, settings, roles, volume*). Tests that need “Show more” / many products require the **volume** seeds.

Community posts gained `image_url` and `link_*` in `services/community-service/migrations/005_post_media.sql`. Apply that file to `appdb_community_staging` (and production when promoting) before rolling a community-service image that SELECTs those columns.

Demo users: nia/leo/maya/owen (see `store4bots-product`).

## Do not

- Point pods at `postgres.database.svc.cluster.local` unless you actually run the in-cluster StatefulSet.
- Seed production with staging demo passwords.
