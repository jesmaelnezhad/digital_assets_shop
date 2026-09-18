#!/usr/bin/env bash
# Source from any script:  . "$(dirname "$0")/lib/site-env.sh"
# or:  . scripts/lib/site-env.sh   (from repo root)
set -euo pipefail
_SITE_ENV_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
if [[ ! -f "$_SITE_ENV_ROOT/config/site.env" ]]; then
  echo "Missing $_SITE_ENV_ROOT/config/site.env — copy config/site.env.example" >&2
  exit 1
fi
set -a
# shellcheck disable=SC1091
. "$_SITE_ENV_ROOT/config/site.env"
if [[ -f "$_SITE_ENV_ROOT/config/site.secrets.env" ]]; then
  # shellcheck disable=SC1091
  . "$_SITE_ENV_ROOT/config/site.secrets.env"
fi
set +a
REPO_ROOT="$_SITE_ENV_ROOT"
IMAGE() { echo "${REGISTRY_HOST}/${IMAGE_REPO}/$1"; }
blue_ssh() { ssh -p "${BLUE_SSH_PORT}" "${BLUE_SSH_USER}@${BLUE_HOST}" "$@"; }
red_ssh() { ssh -p "${RED_SSH_PORT}" "${RED_SSH_USER}@${RED_HOST}" "$@"; }
