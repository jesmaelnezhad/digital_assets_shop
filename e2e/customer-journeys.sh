#!/bin/bash
# 20 customer-journey e2e scenarios (shop + crypto checkout + community).
# Run: ./customer-journeys.sh [staging|prod]
set -u
cd "$(dirname "$0")"
source ./common.sh "${1:-staging}"
SUITE="customer-journeys"
TS=$(date +%s)
EMAIL="cust-$TS@example.com"
PW="Pass1234"

echo "=== $SUITE on $ENV_NAME ($BASE) ==="

# C01 register
http_code POST "$API/register" "{\"email\":\"$EMAIL\",\"password\":\"$PW\",\"name\":\"cust\"}"
code="$HTTP_CODE"
assert_http C01 "register new user" 201 "$code"
TOKEN=$(jget "$BODY" "['token']")
[ -n "$TOKEN" ] && ok "C01 token issued" || bad "C01 token issued" "no token in response"

# C02 duplicate register rejected
http_code POST "$API/register" "{\"email\":\"$EMAIL\",\"password\":\"$PW\",\"name\":\"cust\"}"
code="$HTTP_CODE"
[ "$code" = "400" ] || [ "$code" = "409" ] && ok "C02 duplicate register rejected ($code)" || bad "C02 duplicate register rejected" "got $code"

# C03 login
http_code POST "$API/login" "{\"email\":\"$EMAIL\",\"password\":\"$PW\"}"
code="$HTTP_CODE"
assert_http C03 "login with correct password" 200 "$code"
TOKEN=$(jget "$BODY" "['token']")

# C04 login wrong password
http_code POST "$API/login" "{\"email\":\"$EMAIL\",\"password\":\"wrong\"}"
code="$HTTP_CODE"
assert_http C04 "login with wrong password rejected" 401 "$code"

# C05 profile
http_code GET "$API/me" "" "$TOKEN"
code="$HTTP_CODE"
assert_http C05 "get own profile" 200 "$code"
assert_contains C05 "profile has email" "$EMAIL" "$BODY"

# C06 browse products
http_code GET "$API/products?per_page=12"
code="$HTTP_CODE"
assert_http C06 "browse products" 200 "$code"
assert_contains C06 "product list non-empty" '"title"' "$BODY"

# C07 search
http_code GET "$API/products?search=pixel"
code="$HTTP_CODE"
assert_http C07 "search products" 200 "$code"
assert_contains C07 "search narrows to pixel pack" "pixel" "$(echo "$BODY" | tr '[:upper:]' '[:lower:]')"

# C08 category filter
http_code GET "$API/categories"
code="$HTTP_CODE"
assert_http C08 "list categories" 200 "$code"
assert_contains C08 "categories non-empty" '"slug"' "$BODY"

# C09 product detail
http_code GET "$API/products/pixel-icon-pack-500"
code="$HTTP_CODE"
assert_http C09 "product detail page data" 200 "$code"
assert_contains C09 "detail has price" "price_usd" "$BODY"
assert_contains C09 "detail has images" '"images"' "$BODY"

# C10 product 404
http_code GET "$API/products/no-such-product-xyz"
code="$HTTP_CODE"
assert_http C10 "unknown product slug 404" 404 "$code"

# C11 unauthenticated order rejected
http_code POST "$API/orders" '{"items":[{"product_id":1,"quantity":1}]}'
code="$HTTP_CODE"
assert_http C11 "create order without token rejected" 401 "$code"

# C12 create order (checkout)
http_code POST "$API/orders" '{"items":[{"product_id":1,"quantity":1}]}' "$TOKEN"
code="$HTTP_CODE"
assert_http C12 "create order" 201 "$code"
ORDER_ID=$(jget "$BODY" "['order']['id']")
assert_contains C12 "payment address present" "payment_address" "$BODY"
assert_contains C12 "crypto amount present" "total_crypto" "$BODY"
assert_contains C12 "memo present" "memo" "$(echo "$BODY" | tr '[:upper:]' '[:lower:]')"

# C13 order detail
http_code GET "$API/orders/$ORDER_ID" "" "$TOKEN"
code="$HTTP_CODE"
assert_http C13 "get order detail" 200 "$code"
assert_contains C13 "order is pending" "pending" "$BODY"

# C14 my orders list
http_code GET "$API/orders" "" "$TOKEN"
code="$HTTP_CODE"
assert_http C14 "list my orders" 200 "$code"
assert_contains C14 "new order listed" "\"id\":$ORDER_ID" "$BODY"

# C15 payment info
http_code GET "$API/orders/$ORDER_ID/payment" "" "$TOKEN"
code="$HTTP_CODE"
assert_http C15 "get payment info" 200 "$code"
assert_contains C15 "payment instructions present" "BNB" "$BODY"

# C16 payment status poll
http_code GET "$API/orders/$ORDER_ID/status" "" "$TOKEN"
code="$HTTP_CODE"
assert_http C16 "check payment status" 200 "$code"
assert_contains C16 "status field present" "status" "$BODY"

# C17 unpaid download blocked
http_code GET "$API/orders/download/$ORDER_ID" "" "$TOKEN"
code="$HTTP_CODE"
[ "$code" != "200" ] && ok "C17 unpaid download blocked ($code)" || bad "C17 unpaid download blocked" "got 200 without payment"

# C18 community feed
http_code GET "$API/community/feed?per_page=5"
code="$HTTP_CODE"
[ "$code" = "200" ] && ok "C18 community feed loads" || bad "C18 community feed loads" "got $code"

# C19 community interactions (post -> like -> comment -> follow -> profile)
http_code POST "$API/community/posts" "{\"content\":\"e2e hello $TS\"}" "$TOKEN"
code="$HTTP_CODE"
assert_http C19a "create post" 201 "$code"
POST_ID=$(jget "$BODY" "['post']['id']")
if [ -z "$POST_ID" ]; then POST_ID=$(jget "$BODY" "['id']"); fi
http_code POST "$API/community/posts/$POST_ID/like" "" "$TOKEN"
code="$HTTP_CODE"
[ "$code" = "200" ] && ok "C19b like post" || bad "C19b like post" "got $code"
http_code POST "$API/community/posts/$POST_ID/comments" "{\"content\":\"nice post\"}" "$TOKEN"
code="$HTTP_CODE"
[ "$code" = "200" ] || [ "$code" = "201" ] && ok "C19c comment on post ($code)" || bad "C19c comment on post" "got $code"
AUTHOR_ID=$(jget "$BODY" "['comment']['user_id']" 2>/dev/null)
ME_ID=$(jget "$(curl -s --max-time 10 "$API/me" -H "Authorization: Bearer $TOKEN")" "['id']")
OTHER_ID="${AUTHOR_ID:-$ME_ID}"
http_code GET "$API/community/users/$OTHER_ID" "" "$TOKEN"
code="$HTTP_CODE"
[ "$code" = "200" ] && ok "C19d view public profile" || bad "C19d view public profile" "got $code"
# second user so we can follow somebody else
U2="cust2-$TS@example.com"
T2=$(curl -s --max-time 10 -X POST "$API/register" -H "Content-Type: application/json" -d "{\"email\":\"$U2\",\"password\":\"$PW\",\"name\":\"cust2\"}" | python3 -c "import sys,json;d=json.load(sys.stdin);print(d.get('token',''))")
U2ID=$(curl -s --max-time 10 "$API/me" -H "Authorization: Bearer $T2" | python3 -c "import sys,json;print(json.load(sys.stdin).get('id',''))")
http_code POST "$API/community/follow/$U2ID" "" "$TOKEN"
code="$HTTP_CODE"
[ "$code" = "200" ] && ok "C19e follow another user" || bad "C19e follow another user" "got $code"
http_code DELETE "$API/community/follow/$U2ID" "" "$TOKEN"
code="$HTTP_CODE"
[ "$code" = "200" ] && ok "C19f unfollow user" || bad "C19f unfollow user" "got $code"

# C20 logout invalidates token
http_code POST "$API/logout" "" "$TOKEN"
code="$HTTP_CODE"
[ "$code" = "200" ] && ok "C20a logout" || bad "C20a logout" "got $code"
http_code GET "$API/me" "" "$TOKEN"
code="$HTTP_CODE"
[ "$code" = "401" ] && ok "C20b old token rejected after logout" || bad "C20b old token rejected after logout" "got $code"

summary
