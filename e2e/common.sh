#!/bin/bash
# common helpers for pawradise e2e suites (curl-based, run from BLUE against RED)
# usage: source "$(dirname "$0")/common.sh" [prod|staging]
#
# With domain-based routing (v2):
#   staging → https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir
#   production → https://pawradise.ir (placeholder, not deployed yet)
#
# Tests MUST run from BLUE (the machine with internet access to the domain).
# RED is the k3s host and has no direct route to the public internet for testing.

ENV_NAME="${1:-staging}"
STAGING_DOMAIN="server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir"
PRODUCTION_DOMAIN="pawradise.ir"  # placeholder — real domain TBD

if [ "$ENV_NAME" = "prod" ] || [ "$ENV_NAME" = "production" ]; then
  BASE="https://${PRODUCTION_DOMAIN}"
  ENV_NAME="production"
else
  BASE="https://${STAGING_DOMAIN}"
  ENV_NAME="staging"
fi

API="$BASE/api/v1"
ADMIN_TOKEN="${ADMIN_TOKEN:-admin_secret_staging_2026}"

PASS=0; FAIL=0; FAILED_CASES=()

# run CASE_ID "description" -- runs the assertion command(s) given, counts result
# simpler primitives below:
ok()   { PASS=$((PASS+1)); echo "  PASS $1"; }
bad()  { FAIL=$((FAIL+1)); FAILED_CASES+=("$1"); echo "  FAIL $1 -- $2"; }

# assert_http CASE_ID DESCRIPTION EXPECTED_CODE ACTUAL_CODE
assert_http() {
  if [ "$3" = "$4" ]; then ok "$1 $2"; else bad "$1 $2" "want HTTP $3, got $4"; fi
}

# assert_contains CASE_ID DESCRIPTION NEEDLE HAYSTACK
assert_contains() {
  if echo "$4" | grep -qF "$3"; then ok "$1 $2"; else bad "$1 $2" "missing [$3] in: $(echo "$4" | head -c 200)"; fi
}

# http_code METHOD URL [data] [token] -> sets HTTP_CODE and BODY globals
# (must be called directly, NOT in $() — subshell would lose BODY)
HTTP_CODE=""; BODY=""
http_code() {
  local method="$1" url="$2" data="${3:-}" token="${4:-}"
  local args=(-sk --max-time 10 -o /tmp/e2e_body.txt -w "%{http_code}" -X "$method" "$url")
  [ -n "$data" ] && args+=(-H "Content-Type: application/json" -d "$data")
  [ -n "$token" ] && args+=(-H "Authorization: Bearer $token")
  HTTP_CODE=$(curl "${args[@]}")
  BODY=$(cat /tmp/e2e_body.txt)
}

jget() { echo "$1" | python3 -c "import sys,json;print(json.load(sys.stdin)$2)" 2>/dev/null; }

summary() {
  echo ""
  echo "=== $SUITE ($ENV_NAME): $PASS passed, $FAIL failed ==="
  if [ "$FAIL" -gt 0 ]; then printf 'FAILED: %s\n' "${FAILED_CASES[@]}"; return 1; fi
  return 0
}
