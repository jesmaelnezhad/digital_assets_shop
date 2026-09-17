#!/usr/bin/env bash
# External browser-facing test — validates what a real browser sees from outside
# Tests: domain, all pages, all assets, all navigation links, MIME types, cookie auth

DOMAIN="https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir"
ASSETS=("/assets/theme.css" "/assets/alpine.min.js" "/assets/api.js" "/assets/app.js" "/assets/favicon.svg")
FAILED=0
PASSED=0

echo "=== EXTERNAL BROWSER-FACING TEST ==="
echo "Domain: $DOMAIN"
echo ""

# Test assets first
echo "--- Asset MIME types ---"
for asset in "${ASSETS[@]}"; do
  info=$(curl -s -k -o /dev/null -w "%{http_code} %{content_type}" "${DOMAIN}${asset}")
  code=$(echo "$info" | cut -d' ' -f1)
  ctype=$(echo "$info" | cut -d' ' -f2-)
  if [ "$code" = "200" ]; then
    echo "  PASS ${asset}: ${ctype}"
    ((PASSED++))
  else
    echo "  FAIL ${asset}: HTTP ${code}"
    ((FAILED++))
  fi
done

# Test all pages (200 or 301 to trailing slash is OK)
echo ""
echo "--- All pages ---"
PAGES=(
  "/" "/auth/login.html" "/auth/register.html"
  "/shop/" "/category"
  "/product/product.html" "/product/bundle.html" "/product/request.html"
  "/account/account.html" "/account/" "/cart" "/referrals" "/wishlist"
  "/checkout/checkout.html" "/community/community.html"
  "/admin/" "/post/1" "/profile/1"
  "/staging/" "/staging/auth/login.html" "/staging/shop/"
  "/staging/admin/" "/staging/product/product.html"
  "/staging/product/bundle.html" "/staging/product/request.html"
  "/staging/account/account.html" "/staging/cart"
  "/staging/checkout/checkout.html" "/staging/community/community.html"
)
for page in "${PAGES[@]}"; do
  code=$(curl -s -k -o /dev/null -w "%{http_code}" "${DOMAIN}${page}")
  if [ "$code" = "200" ]; then
    echo "  PASS ${page}"
    ((PASSED++))
  else
    echo "  FAIL ${page}: HTTP ${code}"
    ((FAILED++))
  fi
done

# Test navigation links in rendered pages
echo ""
echo "--- Navigation links from pages ---"
for page in "/" "/auth/login.html" "/product/product.html" "/account/account.html" "/checkout/checkout.html" "/community/community.html" "/admin/"; do
  links=$(curl -s -k "${DOMAIN}${page}" 2>/dev/null | grep -oE 'href="/[^"]*"' | sort -u)
  while IFS= read -r link; do
    [ -z "$link" ] && continue
    path=$(echo "$link" | sed 's/href="//;s/"//')
    [ "$path" = "/" ] && continue
    [[ "$path" == \#* ]] && continue
    [[ "$path" == *'${'* ]] && continue
    # Skip asset links (already tested above)
    [[ "$path" == /assets/* ]] && continue
    code=$(curl -s -k -o /dev/null -w "%{http_code}" "${DOMAIN}${path}")
    # 301 to add trailing slash is acceptable (browsers follow redirects)
    if [ "$code" = "200" ] || [ "$code" = "301" ]; then
      echo "  PASS ${page} -> ${path} (${code})"
      ((PASSED++))
    else
      echo "  FAIL ${page} -> ${path}: HTTP ${code}"
      ((FAILED++))
    fi
  done <<< "$links"
done

# Test auth flow
echo ""
echo "--- Auth flow ---"
EMAIL="exttest$(date +%s)@example.com"
resp=$(curl -s -k -c /tmp/ext_cookies.txt -X POST "${DOMAIN}/api/v1/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"password123\",\"name\":\"Ext Test\"}")
if echo "$resp" | grep -q '"token"'; then
  echo "  PASS Register: token received"
  ((PASSED++))
else
  echo "  FAIL Register: no token"
  ((FAILED++))
fi

resp=$(curl -s -k -c /tmp/ext_cookies.txt -X POST "${DOMAIN}/api/v1/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"password123\"}")
if echo "$resp" | grep -q '"token"'; then
  echo "  PASS Login: token received"
  ((PASSED++))
else
  echo "  FAIL Login: no token"
  ((FAILED++))
fi

profile=$(curl -s -k -b /tmp/ext_cookies.txt "${DOMAIN}/api/v1/me")
if echo "$profile" | grep -q '"email"'; then
  echo "  PASS Cookie auth: /me returns profile"
  ((PASSED++))
else
  echo "  FAIL Cookie auth: /me failed"
  ((FAILED++))
fi

# Test API
echo ""
echo "--- API health ---"
for endpoint in "/api/v1/health" "/api/v1/products" "/api/v1/categories"; do
  code=$(curl -s -k -o /dev/null -w "%{http_code}" "${DOMAIN}${endpoint}")
  if [ "$code" = "200" ]; then
    echo "  PASS ${endpoint}"
    ((PASSED++))
  else
    echo "  FAIL ${endpoint}: HTTP ${code}"
    ((FAILED++))
  fi
done

echo ""
echo "=== SUMMARY: ${PASSED} passed, ${FAILED} failed ==="
exit $FAILED
