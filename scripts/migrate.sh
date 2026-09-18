#!/usr/bin/env bash
# Apply SQL in services/<svc>/migrations/ to appdb_<svc>_<env> on RED.
# Usage: scripts/migrate.sh [service|all] [staging|production]
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/site-env.sh
. "$SCRIPT_DIR/lib/site-env.sh"

need() { [[ -n "${!1:-}" ]] || { echo "missing \$$1" >&2; exit 1; }; }
need DB_USER

SERVICES=(identity-service product-service commerce-service community-service review-service payment-service admin-service media-service)
ENV_NAME="staging"
WANT="all"
for arg in "$@"; do
  case "$arg" in
    staging|production) ENV_NAME="$arg" ;;
    all) WANT="all" ;;
    *) WANT="$arg" ;;
  esac
done
if [[ "$WANT" != "all" ]]; then
  SERVICES=("$WANT")
fi

for svc in "${SERVICES[@]}"; do
  short="${svc%-service}"
  db="appdb_${short}_${ENV_NAME}"
  dir="$REPO_ROOT/services/$svc/migrations"
  if [[ ! -d "$dir" ]]; then
    echo "no migrations for $svc"
    continue
  fi
  echo "migrate $svc -> $db"
  shopt -s nullglob
  files=("$dir"/*.sql)
  shopt -u nullglob
  if [[ ${#files[@]} -eq 0 ]]; then
    continue
  fi
  IFS=$'\n' files_sorted=($(sort <<<"${files[*]}"))
  unset IFS
  for f in "${files_sorted[@]}"; do
    echo "  $(basename "$f")"
    red_ssh "docker exec -i postgres psql -U ${DB_USER} -d ${db} -v ON_ERROR_STOP=1" <"$f"
  done
done
