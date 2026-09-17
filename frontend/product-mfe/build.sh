#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
exec "$ROOT/shared/build-mfe.sh" "$(basename "$(cd "$(dirname "$0")" && pwd)")"
