#!/bin/bash
# 20 admin-panel e2e scenarios (users, products, orders, rates, settings).
# Run: ./admin-ops.sh [staging|prod]
set -u
cd "$(dirname "$0")"
source ./common.sh "${1:-staging}"
SUITE="admin-ops"
TS=$(date +%s)
A="$ADMIN_TOKEN"

echo "=== $SUITE on $ENV_NAME ($BASE) ==="

# A01/A02 auth enforcement
http_code GET "$API/admin/users"
code="$HTTP_CODE"
assert_http A01 "admin users without token rejected" 401 "$code"
http_code GET "$API/admin/users" "" "wrong-token"
code="$HTTP_CODE"
[ "$code" = "401" ] || [ "$code" = "403" ] && ok "A02 admin users with bad token rejected ($code)" || bad "A02 admin users with bad token rejected" "got $code"

# A03 list users
http_code GET "$API/admin/users" "" "$A"
code="$HTTP_CODE"
assert_http A03 "list users" 200 "$code"
assert_contains A03 "users array present" '"users"' "$BODY"

# A04 test user for admin user-ops
UEMAIL="admintest-$TS@example.com"
UDATA=$(curl -s --max-time 10 -X POST "$API/register" -H "Content-Type: application/json" -d "{\"email\":\"$UEMAIL\",\"password\":\"Pass1234\",\"name\":\"admintest\"}")
TUID=$(echo "$UDATA" | python3 -c "import sys,json;d=json.load(sys.stdin);print(d.get('user',{}).get('id',''))")
[ -n "$TUID" ] && ok "A04a helper user created (id=$TUID)" || bad "A04a helper user created" "no id"

# A05 reset password
http_code POST "$API/admin/users/$TUID/reset-password" "" "$A"
code="$HTTP_CODE"
assert_http A05 "reset user password" 200 "$code"
NEWPW=$(jget "$BODY" "['new_password']")
[ -n "$NEWPW" ] && ok "A05 new password returned" || bad "A05 new password returned" "missing new_password"

# A06 login with reset password
http_code POST "$API/login" "{\"email\":\"$UEMAIL\",\"password\":\"$NEWPW\"}"
code="$HTTP_CODE"
assert_http A06 "login with reset password" 200 "$code"

# A07 delete user
http_code DELETE "$API/admin/users/$TUID" "" "$A"
code="$HTTP_CODE"
assert_http A07 "delete user" 200 "$code"
http_code POST "$API/login" "{\"email\":\"$UEMAIL\",\"password\":\"$NEWPW\"}"
code="$HTTP_CODE"
assert_http A07b "deleted user cannot login" 401 "$code"

# A08 create product
http_code POST "$API/admin/products" "{\"title\":\"E2E Widget $TS\",\"slug\":\"e2e-widget-$TS\",\"description\":\"created by e2e\",\"price_usd\":7.5,\"status\":\"active\"}" "$A"
code="$HTTP_CODE"
[ "$code" = "200" ] || [ "$code" = "201" ] && ok "A08 create product ($code)" || bad "A08 create product" "got $code: $(echo "$BODY" | head -c 150)"
PID=$(jget "$BODY" "['product']['id']")
if [ -z "$PID" ]; then PID=$(jget "$BODY" "['id']"); fi

# A09 create product invalid
http_code POST "$API/admin/products" '{"title":"","slug":"x","description":"y","price_usd":-1}' "$A"
code="$HTTP_CODE"
[ "$code" = "400" ] && ok "A09 invalid product rejected" || bad "A09 invalid product rejected" "got $code"

# A10 update product
http_code PUT "$API/admin/products/$PID" '{"title":"E2E Widget updated","price_usd":9.99}' "$A"
code="$HTTP_CODE"
assert_http A10 "update product" 200 "$code"
assert_contains A10 "price updated" "9.99" "$BODY"

# A11 new product visible in shop
http_code GET "$API/products/e2e-widget-$TS"
code="$HTTP_CODE"
assert_http A11 "new product visible in shop" 200 "$code"

# A12 product stats
http_code GET "$API/admin/products/stats" "" "$A"
code="$HTTP_CODE"
assert_http A12 "product stats" 200 "$code"
assert_contains A12 "total_products present" "total_products" "$BODY"

# A13 delete product
http_code DELETE "$API/admin/products/$PID" "" "$A"
code="$HTTP_CODE"
assert_http A13 "delete product" 200 "$code"
http_code GET "$API/products/e2e-widget-$TS"
code="$HTTP_CODE"
assert_http A13b "deleted product gone from shop" 404 "$code"

# A14 list orders
http_code GET "$API/admin/orders" "" "$A"
code="$HTTP_CODE"
assert_http A14 "list all orders" 200 "$code"
assert_contains A14 "orders array present" '"orders"' "$BODY"
OID=$(jget "$BODY" "['orders'][0]['id']")

# A15 order detail
if [ -n "$OID" ]; then
  http_code GET "$API/admin/orders/$OID" "" "$A"
code="$HTTP_CODE"
  assert_http A15 "get any order by id" 200 "$code"
else
  bad "A15 get any order by id" "no orders exist yet"
fi
http_code GET "$API/admin/orders/999999" "" "$A"
code="$HTTP_CODE"
assert_http A15b "unknown order 404" 404 "$code"

# A16 rates list + get
http_code GET "$API/admin/exchange-rates" "" "$A"
code="$HTTP_CODE"
assert_http A16 "list exchange rates" 200 "$code"
assert_contains A16 "bsc rate present" '"bsc"' "$BODY"
http_code GET "$API/admin/exchange-rates/bsc" "" "$A"
code="$HTTP_CODE"
assert_http A16b "get bsc rate" 200 "$code"

# A17 set rate + restore
http_code PUT "$API/admin/exchange-rates/bsc" '{"symbol":"BNB","rate_to_usd":"400"}' "$A"
code="$HTTP_CODE"
assert_http A17 "set bsc rate" 200 "$code"
http_code GET "$API/admin/exchange-rates/bsc" "" "$A"
code="$HTTP_CODE"
assert_contains A17b "rate updated to 400" "400" "$BODY"
http_code PUT "$API/admin/exchange-rates/bsc" '{"symbol":"BNB","rate_to_usd":"350"}' "$A"
code="$HTTP_CODE"
assert_http A17c "restore bsc rate" 200 "$code"

# A18 invalid rate rejected
http_code PUT "$API/admin/exchange-rates/bsc" '{"symbol":"BNB","rate_to_usd":"-5"}' "$A"
code="$HTTP_CODE"
assert_http A18 "negative rate rejected" 400 "$code"

# A19 settings: payment address round-trip
http_code GET "$API/admin/settings" "" "$A"
code="$HTTP_CODE"
assert_http A19 "list settings" 200 "$code"
assert_contains A19b "payment_address key present" "payment_address" "$BODY"
http_code PUT "$API/admin/settings/payment_address" '{"value":"0xE2Etestwallet"}' "$A"
code="$HTTP_CODE"
assert_http A19c "set payment address" 200 "$code"
UE="addr-$TS@example.com"
UT=$(curl -s --max-time 10 -X POST "$API/register" -H "Content-Type: application/json" -d "{\"email\":\"$UE\",\"password\":\"Pass1234\",\"name\":\"addr\"}" | python3 -c "import sys,json;print(json.load(sys.stdin).get('token',''))")
OBODY=$(curl -s --max-time 10 -X POST "$API/orders" -H "Authorization: Bearer $UT" -H "Content-Type: application/json" -d '{"items":[{"product_id":1,"quantity":1}]}')
assert_contains A19d "new order uses updated address" "0xE2Etestwallet" "$OBODY"
http_code PUT "$API/admin/settings/payment_address" '{"value":"0xYourBSCWelcomeAddressHere"}' "$A"
code="$HTTP_CODE"
assert_http A19e "restore payment address" 200 "$code"

# A20 admin page (domain-based routing: staging domain serves admin MFE)
t=$(curl -sk --max-time 10 -o /dev/null -w "%{http_code}" "$BASE/admin")
[ "$t" = "200" ] && ok "A20 admin page /admin" || bad "A20 admin page /admin" "got $t"

summary
