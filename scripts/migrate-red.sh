#!/usr/bin/env bash
# Copy RED data toward a new serving VM. Does not change DNS.
#   scripts/migrate-red.sh dump              # write dumps under ./migrate-red-data/
#   scripts/migrate-red.sh restore <host>    # after site.env RED_HOST points at the new VM
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/site-env.sh
. "$SCRIPT_DIR/lib/site-env.sh"

need() { [[ -n "${!1:-}" ]] || { echo "missing \$$1" >&2; exit 1; }; }
need DB_USER
need MONGO_USER
need MONGO_PASSWORD
: "${MONGO_DB:=events}"

CMD="${1:-dump}"
DATA="${MIGRATE_RED_DIR:-$REPO_ROOT/migrate-red-data}"
mkdir -p "$DATA"

dump() {
  echo "dumping Postgres from $RED_HOST"
  red_ssh "docker exec postgres pg_dumpall -U ${DB_USER}" >"$DATA/postgres.sql"
  echo "dumping Mongo"
  red_ssh "docker exec mongo mongodump -u '${MONGO_USER}' -p '${MONGO_PASSWORD}' --authenticationDatabase admin --archive --gzip" >"$DATA/mongo.archive.gz"
  echo "wrote $DATA/postgres.sql and $DATA/mongo.archive.gz"
  echo "Registry blobs stay on the old PVC. After the new RED is up, re-push images with scripts/build-push-deploy.sh (or copy the PVC yourself)."
  echo "Next: install the new VM (prepare-red.sh), point config/site.env RED_HOST at it, then: scripts/migrate-red.sh restore"
  echo "Keep the old server until https://\$STAGING_HOST returns the shop. Then retarget DNS A records and retire the old VM."
}

restore() {
  need POSTGRES_PASSWORD
  echo "restoring onto $RED_HOST (must already have postgres/mongo containers)"
  red_ssh "docker exec -i postgres psql -U ${DB_USER} -d postgres" <"$DATA/postgres.sql"
  red_ssh "docker exec -i mongo mongorestore -u '${MONGO_USER}' -p '${MONGO_PASSWORD}' --authenticationDatabase admin --archive --gzip --drop" <"$DATA/mongo.archive.gz"
  echo "restore done. Confirm https://\$STAGING_HOST then switch DNS if it still points at the old IP."
}

case "$CMD" in
  dump) dump ;;
  restore) restore ;;
  *) echo "usage: migrate-red.sh dump|restore" >&2; exit 1 ;;
esac
