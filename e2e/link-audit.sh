#!/bin/bash
# Link audit: every staging page must link only to staging paths (never production),
# every production page must never link to /staging. Checks href/src attributes.
# Run: ./link-audit.sh
set -u
RED="http://194.5.206.106"
PASS=0; FAIL=0; FAILED=()
ok()  { PASS=$((PASS+1)); echo "  PASS $1"; }
bad() { FAIL=$((FAIL+1)); FAILED+=("$1"); echo "  FAIL $1 -- $2"; }

echo "=== link-audit ==="
for page in "community" "account" "checkout" "login"; do
  # staging page: no href="/..." that doesn't start with /staging (allow # anchors)
  leaks=$(curl -s --max-time 10 "$RED/staging/$page" | grep -oE '(href|src|action)="/[^"]*"' | grep -vE '(href|src|action)="/staging' || true)
  if [ -z "$leaks" ]; then ok "staging/$page has no production links"; else bad "staging/$page has no production links" "$leaks"; fi
  # production page: no /staging links
  leaks=$(curl -s --max-time 10 "$RED/$page" | grep -oE '(href|src|action)="/staging[^"]*"' || true)
  if [ -z "$leaks" ]; then ok "$page has no staging links"; else bad "$page has no staging links" "$leaks"; fi
done
# product pages both envs
for url in "$RED/product/pixel-icon-pack-500" "$RED/staging/product/pixel-icon-pack-500"; do
  code=$(curl -s --max-time 10 -o /dev/null -w "%{http_code}" "$url")
  [ "$code" = "200" ] && ok "product page $url" || bad "product page $url" "got $code"
done
# nav essentials present on shop pages
for url in "$RED/" "$RED/staging/"; do
  body=$(curl -s --max-time 10 "$url")
  echo "$body" | grep -q 'id="authBtn"' && ok "auth button on $url" || bad "auth button on $url" "missing"
  echo "$body" | grep -q 'data-link="/community"' && ok "community nav on $url" || bad "community nav on $url" "missing"
done
echo ""
echo "=== link-audit: $PASS passed, $FAIL failed ==="
[ "$FAIL" -gt 0 ] && printf 'FAILED: %s\n' "${FAILED[@]}" && exit 1
exit 0
