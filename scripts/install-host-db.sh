#!/usr/bin/env bash
# Host Docker Postgres 16 + Mongo 7 on RED; create appdb_<service>_{staging,production}.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/site-env.sh
. "$SCRIPT_DIR/lib/site-env.sh"

need() { [[ -n "${!1:-}" ]] || { echo "missing \$$1" >&2; exit 1; }; }
need POSTGRES_PASSWORD
need MONGO_PASSWORD
need DB_USER
need MONGO_USER
: "${DB_PORT:=5432}"
: "${MONGO_PORT:=27017}"
: "${MONGO_DB:=events}"

# shellcheck disable=SC2087
red_ssh bash -s <<EOF
set -euo pipefail
if ! docker ps -a --format '{{.Names}}' | grep -qx postgres; then
  docker run -d --name postgres --restart unless-stopped \\
    -e POSTGRES_USER='${DB_USER}' \\
    -e POSTGRES_PASSWORD='${POSTGRES_PASSWORD}' \\
    -e POSTGRES_DB=postgres \\
    -p ${DB_PORT}:5432 \\
    -v store4bots-pg:/var/lib/postgresql/data \\
    postgres:16-alpine
else
  docker start postgres >/dev/null
fi
if ! docker ps -a --format '{{.Names}}' | grep -qx mongo; then
  docker run -d --name mongo --restart unless-stopped \\
    -e MONGO_INITDB_ROOT_USERNAME='${MONGO_USER}' \\
    -e MONGO_INITDB_ROOT_PASSWORD='${MONGO_PASSWORD}' \\
    -p ${MONGO_PORT}:27017 \\
    -v store4bots-mongo:/data/db \\
    mongo:7
else
  docker start mongo >/dev/null
fi

for i in \$(seq 1 30); do
  docker exec postgres pg_isready -U '${DB_USER}' && break
  sleep 2
done

svcs="identity product commerce community review payment admin media"
for env in staging production; do
  for svc in \$svcs; do
    db="appdb_\${svc}_\${env}"
    docker exec postgres psql -U '${DB_USER}' -d postgres -tc "SELECT 1 FROM pg_database WHERE datname='\${db}'" | grep -q 1 \\
      || docker exec postgres psql -U '${DB_USER}' -d postgres -c "CREATE DATABASE \${db}"
  done
done

docker exec mongo mongosh --quiet -u '${MONGO_USER}' -p '${MONGO_PASSWORD}' --authenticationDatabase admin --eval 'db.getSiblingDB("${MONGO_DB}").events.createIndex({expire_at:1},{expireAfterSeconds:0})' || true
echo "host databases ready"
EOF
