# Migrations — identity-service

## Strategy

This service uses **PostgreSQL** for persistent storage. Migrations are applied
at service startup using an embedded migration approach.

## Approach

1. SQL migration files are stored in `migrations/sql/` as numbered files:
   - `001_create_users_table.up.sql`
   - `001_create_users_table.down.sql`
   - `002_add_user_avatar.up.sql`
   - `002_add_user_avatar.down.sql`

2. A migration tracking table (`schema_migrations`) records which migrations
   have been applied.

3. Migrations run automatically on service startup before the HTTP/gRPC
   servers begin listening.

## Tables Owned by This Service

users, user_profiles, invalidated_tokens, referral_links, user_referrals, referral_commissions

## Adding a New Migration

1. Create `migrations/sql/NNN_description.up.sql` and `NNN_description.down.sql`
2. The service will apply it on next startup
3. Never modify an existing migration file — always create a new one

## Rollback

To rollback the last migration, run the service with:
```
ROLLBACK_LAST=true ./server
```

## Notes

- Each service owns its own database schema. Cross-service data access is via
  API calls, never JOINs.
- Use `IF NOT EXISTS` / `IF EXISTS` for idempotent DDL.
- Always include a `.down.sql` rollback file.
