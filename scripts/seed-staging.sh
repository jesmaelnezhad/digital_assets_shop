#!/usr/bin/env bash
# Load staging demo data. Never run against production.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/site-env.sh
. "$SCRIPT_DIR/lib/site-env.sh"

need() { [[ -n "${!1:-}" ]] || { echo "missing \$$1" >&2; exit 1; }; }
need DB_USER

apply() {
  local db="$1"
  local file="$2"
  echo "$file -> $db"
  red_ssh "docker exec -i postgres psql -U ${DB_USER} -d ${db} -v ON_ERROR_STOP=1" <"$file"
}

apply appdb_identity_staging  "$SCRIPT_DIR/seed-staging-identity.sql"
apply appdb_product_staging   "$SCRIPT_DIR/seed-staging-product.sql"
apply appdb_admin_staging     "$SCRIPT_DIR/seed-staging-product.sql"
apply appdb_admin_staging     "$SCRIPT_DIR/seed-staging-admin-catalog.sql"
apply appdb_commerce_staging  "$SCRIPT_DIR/seed-staging-commerce.sql"
apply appdb_community_staging "$SCRIPT_DIR/seed-staging-community.sql"
apply appdb_community_staging "$SCRIPT_DIR/seed-staging-community-ids.sql"
apply appdb_payment_staging   "$SCRIPT_DIR/seed-staging-settings.sql"
apply appdb_admin_staging     "$SCRIPT_DIR/seed-staging-settings.sql"
apply appdb_identity_staging  "$SCRIPT_DIR/seed-staging-roles.sql"
apply appdb_product_staging   "$SCRIPT_DIR/seed-staging-volume-product.sql"
apply appdb_admin_staging     "$SCRIPT_DIR/seed-staging-volume-product.sql"
apply appdb_identity_staging  "$SCRIPT_DIR/seed-staging-volume-identity.sql"
apply appdb_community_staging "$SCRIPT_DIR/seed-staging-volume-community.sql"

echo "staging seed done"
