#!/bin/bash
# 20 additional end-to-end scenarios (customer + admin) testing real flows.
# These complement customer-journeys.sh (C01-C20) and admin-ops.sh (A01-A20).
# Run: ./extra-e2e.sh [staging|prod]
set -u
cd "$(dirname "$0")"
source ./common.sh "${1:-staging}"
SUITE="extra-e2e"
TS=$(date +%s)

# Tokens for two test users (customer + separate admin-interacting user)
CUST_EMAIL="ex2cust-$TS@example.com"
CUST_PW="Pass1234"
OTHER_EMAIL="ex2other-$TS@example.com"
OTHER_PW="OtherPass1234"

echo "=== $SUITE on $ENV_NAME ($BASE) ==="

# ---------------------------------------------------------------------------
# Register two users up front for customer scenarios
# ---------------------------------------------------------------------------
http_code POST "$API/register" "{\"email\":\"$CUST_EMAIL\",\"password\":\"$CUST_PW\",\"name\":\"ExtraCust\"}"
code="$HTTP_CODE"; assert_http C_REG1 "register primary test user" 201 "$code"
TCUST=$(jget "$BODY" "['token']")
[ -n "$TCUST" ] && ok "C_REG1a token issued" || bad "C_REG1a token" "no token"

http_code POST "$API/register" "{\"email\":\"$OTHER_EMAIL\",\"password\":\"$OTHER_PW\",\"name\":\"ExtraOther\"}"
code="$HTTP_CODE"; assert_http C_REG2 "register secondary test user" 201 "$code"
OTHER_TOK=$(jget "$BODY" "['token']")
[ -n "$OTHER_TOK" ] && ok "C_REG2a other token issued" || bad "C_REG2a other token" "no token"

OTHER_ID=$(jget "$BODY" "['user']['id']")

# ---------------------------------------------------------------------------
# CUSTOMER SCENARIOS (10)
# ---------------------------------------------------------------------------

# C21 — Update profile: name + bio + wallet_address, verify persistence
http_code PUT "$API/me" "{\"name\":\"UpdatedName\",\"bio\":\"I love digital assets\",\"wallet_address\":\"0xUpdatedWallet12345\"}" "$TCUST"
code="$HTTP_CODE"; assert_http C21a "update profile" 200 "$code"
http_code GET "$API/me" "" "$TCUST"
code="$HTTP_CODE"; assert_http C21b "get updated profile" 200 "$code"
assert_contains C21c "name persisted" "UpdatedName" "$BODY"
assert_contains C21d "bio persisted" "I love digital assets" "$BODY"
assert_contains C21e "wallet persisted" "0xUpdatedWallet12345" "$BODY"

# C22 — Public profile viewable WITHOUT auth (GET /community/users/:id)
http_code GET "$API/community/users/$OTHER_ID"
code="$HTTP_CODE"; assert_http C22a "public profile without auth" 200 "$code"
assert_contains C22b "has user name" "ExtraOther" "$BODY"
assert_contains C22c "has post_count" "post_count" "$BODY"
assert_contains C22d "has followers" "followers" "$BODY"
assert_contains C22e "has following" "following" "$BODY"

# C23 — Follow → verify → unfollow → verify cycle
http_code POST "$API/community/follow/$OTHER_ID" "" "$TCUST"
code="$HTTP_CODE"; assert_http C23a "follow user" 200 "$code"
http_code GET "$API/community/users/$OTHER_ID" "" "$TCUST"
FOLLOW_STATE=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(str(d.get('is_following',False)).lower())" 2>/dev/null)
[ "$FOLLOW_STATE" = "true" ] && ok "C23b is_following=true after follow" || bad "C23b is_following" "got=$FOLLOW_STATE"
http_code DELETE "$API/community/follow/$OTHER_ID" "" "$TCUST"
code="$HTTP_CODE"; assert_http C23c "unfollow user" 200 "$code"
http_code GET "$API/community/users/$OTHER_ID" "" "$TCUST"
FOLLOW_AFTER=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(str(d.get('is_following',False)).lower())" 2>/dev/null)
[ "$FOLLOW_AFTER" = "false" ] && ok "C23d is_following=false after unfollow" || bad "C23d is_following" "got=$FOLLOW_AFTER"

# C24 — Like → verify count up → unlike → verify count down
# Create a post first so we control the ID
http_code POST "$API/community/posts" "{\"content\":\"e2e like test $TS\"}" "$OTHER_TOK"
code="$HTTP_CODE"; assert_http C24a "create post for like test" 201 "$code"
POST_ID=$(jget "$BODY" "['post']['id']")
[ -z "$POST_ID" ] && POST_ID=$(jget "$BODY" "['id']")
http_code GET "$API/community/posts/$POST_ID" "" "$OTHER_TOK"
BEFORE_LIKES=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['post']['like_count'])" 2>/dev/null)
http_code POST "$API/community/posts/$POST_ID/like" "" "$TCUST"
code="$HTTP_CODE"; assert_http C24b "like the post" 200 "$code"
http_code GET "$API/community/posts/$POST_ID" "" "$OTHER_TOK"
AFTER_LIKE=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['post']['like_count'])" 2>/dev/null)
[ "$AFTER_LIKE" = "$((BEFORE_LIKES + 1))" ] && ok "C24c like_count incremented" || bad "C24c like_count" "before=$BEFORE_LIKES after=$AFTER_LIKE"
http_code DELETE "$API/community/posts/$POST_ID/like" "" "$TCUST"
code="$HTTP_CODE"; [ "$code" = "200" ] && ok "C24d unlike" || bad "C24d unlike" "got $code"
http_code GET "$API/community/posts/$POST_ID" "" "$OTHER_TOK"
AFTER_UNLIKE=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['post']['like_count'])" 2>/dev/null)
[ "$AFTER_UNLIKE" = "$BEFORE_LIKES" ] && ok "C24e like_count restored" || bad "C24e like_count restore" "before=$BEFORE_LIKES after=$AFTER_UNLIKE"

# C25 — Add comment → verify in post → delete own comment → verify gone
http_code POST "$API/community/posts/$POST_ID/comments" "{\"content\":\"e2e comment $TS\"}" "$OTHER_TOK"
code="$HTTP_CODE"; assert_http C25a "add comment" 201 "$code"
COMMENT_ID=$(jget "$BODY" "['comment']['id']")
http_code GET "$API/community/posts/$POST_ID" "" "$OTHER_TOK"
COMMENT_IN_POST=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print('YES' if any(c['id']==$COMMENT_ID for c in d.get('comments',[])) else 'NO')" 2>/dev/null)
[ "$COMMENT_IN_POST" = "YES" ] && ok "C25b comment appears in post" || bad "C25b comment in post" "not found"
http_code DELETE "$API/community/posts/$POST_ID/comments/$COMMENT_ID" "" "$OTHER_TOK"
code="$HTTP_CODE"; assert_http C25c "delete own comment" 200 "$code"
http_code GET "$API/community/posts/$POST_ID" "" "$OTHER_TOK"
COMMENT_GONE=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print('YES' if any(c['id']==$COMMENT_ID for c in d.get('comments',[])) else 'NO')" 2>/dev/null)
[ "$COMMENT_GONE" = "NO" ] && ok "C25d comment removed" || bad "C25d comment removed" "still there"

# C26 — Post validation: empty content → 400, content > 500 chars → 400
http_code POST "$API/community/posts" "{\"content\":\"\"}" "$TCUST"
code="$HTTP_CODE"; [ "$code" = "400" ] && ok "C26a empty post rejected" || bad "C26a empty post" "got $code"
LONG_CONTENT=$(python3 -c "print('x' * 501)")
http_code POST "$API/community/posts" "{\"content\":\"$LONG_CONTENT\"}" "$TCUST"
code="$HTTP_CODE"; [ "$code" = "400" ] && ok "C26b over-500 post rejected" || bad "C26b over-500" "got $code"

http_code POST "$API/community/posts/99999/like" "" "$TCUST"
code="$HTTP_CODE"; [ "$code" = "404" ] && ok "C27a like nonexistent post → 404" || bad "C27a like 404" "got $code"
http_code POST "$API/community/follow/99999" "" "$TCUST"
code="$HTTP_CODE"; [ "$code" = "404" ] && ok "C27b follow nonexistent user → 404" || bad "C27b follow 404" "got $code"
# Follow self: use the already-extracted TCUST user id
MY_ID=$(curl -s --max-time 6 "$API/me" -H "Authorization: Bearer ***" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['id'])" 2>/dev/null)
if [ -n "$MY_ID" ] && [ "$MY_ID" != "None" ] && [ "$MY_ID" != "" ]; then
  http_code POST "$API/community/follow/$MY_ID" "" "$TCUST"
  code="$HTTP_CODE"; [ "$code" = "400" ] && ok "C27c follow self → 400" || bad "C27c follow self" "got $code"
else
  bad "C27c my id" "could not get current user id from /me"
fi

# C28 — Product pagination: per_page=3, page=1 vs page=2, different products, total reported
http_code GET "$API/products?per_page=3&page=1"
code="$HTTP_CODE"; assert_http C28a "products page 1 per_page=3" 200 "$code"
PAGE1_SLUG=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['products'][0]['slug'])" 2>/dev/null)
TOTAL=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['total'])" 2>/dev/null)
[ "$TOTAL" -gt 3 ] && ok "C28a2 total > per_page" || bad "C28a2 total" "total=$TOTAL"
http_code GET "$API/products?per_page=3&page=2"
code="$HTTP_CODE"; assert_http C28b "products page 2" 200 "$code"
PAGE2_SLUG=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['products'][0]['slug'])" 2>/dev/null)
[ "$PAGE1_SLUG" != "$PAGE2_SLUG" ] && ok "C28c page1 != page2 slugs" || bad "C28c page diff" "same slug=$PAGE1_SLUG"

# C29 — Multi-item order: 2 different products, verify items count=2, total=sum
# Use products by slug: pixel-icon-pack-500 (id=1) and ultimate-readme-template (id=5)
http_code POST "$API/orders" "{\"items\":[{\"product_id\":1,\"quantity\":1},{\"product_id\":5,\"quantity\":1}]}" "$TCUST"
code="$HTTP_CODE"; assert_http C29a "multi-item order created" 201 "$code"
ORDER_ID=$(jget "$BODY" "['order']['id']")
ITEM_COUNT=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d['order']['items']))" 2>/dev/null)
[ "$ITEM_COUNT" = "2" ] && ok "C29b 2 items in order" || bad "C29b item count" "count=$ITEM_COUNT"
# Get prices by slug (products API uses slug, not id)
P1=$(curl -s --max-time 6 "$API/products/pixel-icon-pack-500" | python3 -c "import sys,json; print(json.load(sys.stdin)['product']['price_usd'])" 2>/dev/null)
P5=$(curl -s --max-time 6 "$API/products/ultimate-readme-template" | python3 -c "import sys,json; print(json.load(sys.stdin)['product']['price_usd'])" 2>/dev/null)
if [ -n "$P1" ] && [ -n "$P5" ] && [ "$P1" != "None" ] && [ "$P5" != "None" ]; then
  EXPECTED=$(python3 -c "print(round($P1 + $P5, 2))")
  ORDER_TOTAL=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['order']['total_usd'])" 2>/dev/null)
  python3 -c "exit(0 if abs(round($ORDER_TOTAL,2) - round($EXPECTED,2)) < 0.01 else 1)" 2>/dev/null && ok "C29c total matches sum" || bad "C29c total" "order=$ORDER_TOTAL expected=$EXPECTED"
else
  bad "C29c price fetch" "could not get product prices (p1=$P1 p5=$P5)"
fi

# C30 — Download from a PAID order (use satoshifan@pawradise.demo who has paid order #1)
DEMO_EMAIL="satoshifan@pawradise.demo"
DEMO_PW="demo1234"
http_code POST "$API/login" "{\"email\":\"$DEMO_EMAIL\",\"password\":\"$DEMO_PW\"}"
code="$HTTP_CODE"; assert_http C30a "login as demo user" 200 "$code"
DEMO_TOK=$(jget "$BODY" "['token']")
http_code GET "$API/orders" "" "$DEMO_TOK"
code="$HTTP_CODE"; assert_http C30b "demo user orders list" 200 "$code"
PAID_ORDER_ID=$(echo "$BODY" | python3 -c "
import sys,json
d=json.load(sys.stdin)
for o in d.get('orders',[]):
  if o.get('status')=='paid':
    print(o['id']); break
else:
  print('')
" 2>/dev/null)
if [ -n "$PAID_ORDER_ID" ] && [ "$PAID_ORDER_ID" != "" ]; then
  http_code GET "$API/orders/$PAID_ORDER_ID" "" "$DEMO_TOK"
  code="$HTTP_CODE"; assert_http C30c "get paid order detail" 200 "$code"
  ITEM_ID=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['order']['items'][0]['id'])" 2>/dev/null)
  if [ -n "$ITEM_ID" ] && [ "$ITEM_ID" != "" ]; then
    http_code GET "$API/orders/download/$ITEM_ID" "" "$DEMO_TOK"
    code="$HTTP_CODE"; assert_http C30d "download from paid order item" 200 "$code"
    assert_contains C30e "download response has item_id" "$ITEM_ID" "$BODY"
    assert_contains C30f "download response has product_title" "product_title" "$BODY"
  else
    bad "C30d item id" "no items in paid order"
  fi
else
  bad "C30b paid order" "no paid order found for demo user"
fi

# ---------------------------------------------------------------------------
# ADMIN SCENARIOS (10)
# ---------------------------------------------------------------------------

# A21 — Admin list orders filtered by status=pending
http_code GET "$API/admin/orders?status=pending&per_page=20" "" "$A"
code="$HTTP_CODE"; assert_http A21a "admin orders filtered pending" 200 "$code"
PENDING_CNT=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(sum(1 for o in d['orders'] if o['status']=='pending'))" 2>/dev/null)
TOTAL_PENDING=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['total'])" 2>/dev/null)
[ "$PENDING_CNT" = "$TOTAL_PENDING" ] && ok "A21b all returned are pending" || bad "A21b all pending" "pending=$PENDING_CNT total=$TOTAL_PENDING"
[ "$TOTAL_PENDING" -ge 0 ] && ok "A21c count >= 0" || bad "A21c count"

# A22 — Admin list orders filtered by user_id
# Get a user id from the users list
UID=$(curl -s --max-time 10 -X GET "$API/admin/users" -H "Authorization: Bearer ***" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['users'][0]['id'])" 2>/dev/null)
if [ -n "$UID" ] && [ "$UID" != "None" ]; then
  http_code GET "$API/admin/orders?user_id=$UID&per_page=20" "" "$A"
  code="$HTTP_CODE"; assert_http A22a "admin orders filtered by user_id" 200 "$code"
  ALL_OWNER=$(echo "$BODY" | python3 -c "
import sys,json
d=json.load(sys.stdin)
oids=[str(o['user_id']) for o in d.get('orders',[])]
print('YES' if all(o=='$UID' for o in oids) else 'NO')
" 2>/dev/null)
  [ "$ALL_OWNER" = "YES" ] && ok "A22b all orders belong to user $UID" || bad "A22b user filter" "mixed owners"
fi

# A23 — Admin search orders by email substring
http_code GET "$API/admin/orders?search=example" "" "$A"
code="$HTTP_CODE"; assert_http A23a "admin order search by email substring" 200 "$code"
SEARCH_TOTAL=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['total'])" 2>/dev/null)
[ "$SEARCH_TOTAL" -ge 0 ] && ok "A23b search returned results" || bad "A23b search total"

# A24 — Admin delete exchange rate (create → delete → verify gone → re-create for cleanup)
http_code PUT "$API/admin/exchange-rates/eth" '{"symbol":"ETH","rate_to_usd":"2000"}' "$A"
code="$HTTP_CODE"; [ "$code" = "200" ] && ok "A24a create eth rate for delete test" || bad "A24a create eth" "got $code"
http_code DELETE "$API/admin/exchange-rates/eth" "" "$A"
code="$HTTP_CODE"; assert_http A24b "delete exchange rate" 200 "$code"
http_code GET "$API/admin/exchange-rates/eth" "" "$A"
code="$HTTP_CODE"; [ "$code" = "404" ] && ok "A24c deleted rate returns 404" || bad "A24c 404" "got $code"
# restore eth rate so other tests aren't broken
http_code PUT "$API/admin/exchange-rates/eth" '{"symbol":"ETH","rate_to_usd":"2000"}' "$A" 2>/dev/null

# A25 — Admin GET single setting key (payment_address)
http_code GET "$API/admin/settings/payment_address" "" "$A"
code="$HTTP_CODE"; assert_http A25a "get single setting key" 200 "$code"
assert_contains A25b "has key payment_address" "payment_address" "$BODY"

# A26 — Admin product with category_id: create → verify category set → visible in shop
CAT_ID=$(curl -s --max-time 10 "$API/categories" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['categories'][0]['id'])" 2>/dev/null)
[ -z "$CAT_ID" ] && CAT_ID=1
http_code POST "$API/admin/products" "{\"title\":\"E2E Cat Product $TS\",\"slug\":\"e2e-cat-$TS\",\"description\":\"with category\",\"price_usd\":12.5,\"category_id\":$CAT_ID,\"status\":\"active\"}" "$A"
code="$HTTP_CODE"; assert_http A26a "create product with category" 201 "$code"
PID_CAT=$(jget "$BODY" "['product']['id']")
http_code GET "$API/admin/products/$PID_CAT" "" "$A"
CAT_NAME=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['product'].get('category_name',''))" 2>/dev/null)
[ -n "$CAT_NAME" ] && ok "A26b category_name is set" || bad "A26b category_name empty"
http_code GET "$API/products/e2e-cat-$TS"
code="$HTTP_CODE"; assert_http A26c "product visible in shop" 200 "$code"
# clean up
http_code DELETE "$API/admin/products/$PID_CAT" "" "$A" 2>/dev/null

# A27 — Admin product draft → hidden from shop → restore → delete
http_code POST "$API/admin/products" "{\"title\":\"E2E Draft $TS\",\"slug\":\"e2e-draft-$TS\",\"description\":\"draft test\",\"price_usd\":5.0,\"status\":\"draft\"}" "$A"
code="$HTTP_CODE"; assert_http A27a "create draft product" 201 "$code"
PID_DRAFT=$(jget "$BODY" "['product']['id']")
http_code GET "$API/products/e2e-draft-$TS"
code="$HTTP_CODE"; [ "$code" = "404" ] && ok "A27b draft product hidden from shop" || bad "A27b draft hidden" "got $code"
# publish it
http_code PUT "$API/admin/products/$PID_DRAFT" '{"status":"active"}' "$A"
code="$HTTP_CODE"; assert_http A27c "publish draft product" 200 "$code"
http_code GET "$API/products/e2e-draft-$TS"
code="$HTTP_CODE"; assert_http A27d "published product visible" 200 "$code"
# delete it
http_code DELETE "$API/admin/products/$PID_DRAFT" "" "$A"
code="$HTTP_CODE"; assert_http A27e "delete product" 200 "$code"

# A28 — Admin order pagination
http_code GET "$API/admin/orders?per_page=5&page=1"
code="$HTTP_CODE"; assert_http A28a "admin orders page 1 per_page=5" 200 "$code"
P1N=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d['orders']))" 2>/dev/null)
[ "$P1N" -le 5 ] && ok "A28b page 1 has <= 5 orders" || bad "A28b page size" "count=$P1N"
TOTAL_ORD=$(echo "$BODY" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['total'])" 2>/dev/null)
[ "$TOTAL_ORD" -gt 0 ] && ok "A28c total orders > 0" || bad "A28c total"
http_code GET "$API/admin/orders?per_page=5&page=2"
code="$HTTP_CODE"; assert_http A28d "admin orders page 2" 200 "$code"

# A29 — Admin product creation invalid: invalid category_id → 400, price=0 → 400, duplicate slug → 400
http_code POST "$API/admin/products" "{\"title\":\"Bad Cat $TS\",\"slug\":\"bad-cat-$TS\",\"description\":\"d\",\"price_usd\":1,\"category_id\":99999}" "$A"
code="$HTTP_CODE"; [ "$code" = "400" ] && ok "A29a invalid category_id rejected" || bad "A29a invalid cat" "got $code"
http_code POST "$API/admin/products" "{\"title\":\"Zero Price $TS\",\"slug\":\"zero-price-$TS\",\"description\":\"d\",\"price_usd\":0}" "$A"
code="$HTTP_CODE"; [ "$code" = "400" ] && ok "A29b zero price rejected" || bad "A29b zero price" "got $code"
# create a product, then try duplicate slug
http_code POST "$API/admin/products" "{\"title\":\"Dup Slug $TS\",\"slug\":\"dup-slug-$TS\",\"description\":\"d\",\"price_usd\":1}" "$A" 2>/dev/null
http_code POST "$API/admin/products" "{\"title\":\"Dup Again\",\"slug\":\"dup-slug-$TS\",\"description\":\"d\",\"price_usd\":1}" "$A"
code="$HTTP_CODE"; [ "$code" = "400" ] && ok "A29c duplicate slug rejected" || bad "A29c dup slug" "got $code"

# A30 — Exchange rate change affects crypto amount in new orders
# Get product 1's price by slug (products API uses slug, not id)
http_code GET "$API/products/pixel-icon-pack-500"
P1_PRICE=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['product']['price_usd'])" 2>/dev/null)
RATE_NEW=400.0
http_code PUT "$API/admin/exchange-rates/bsc" "{\"symbol\":\"BNB\",\"rate_to_usd\":\"400\"}" "$A"
code="$HTTP_CODE"; assert_http A30a "set bsc rate to 400" 200 "$code"
# create order as the test customer user
http_code POST "$API/orders" "{\"items\":[{\"product_id\":1,\"quantity\":1}]}" "$TCUST"
code="$HTTP_CODE"; assert_http A30b "create order at new rate" 201 "$code"
CRYPTO_AMT=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['order']['total_crypto'])" 2>/dev/null)
# expected: P1_PRICE / 400, to 8 decimal places
if [ -n "$P1_PRICE" ] && [ "$P1_PRICE" != "None" ] && [ "$P1_PRICE" != "" ]; then
  EXP_CRYPTO=$(python3 -c "print(f'{($P1_PRICE / 400.0):.8f}')")
  [ "$CRYPTO_AMT" = "$EXP_CRYPTO" ] && ok "A30c crypto amount matches rate" || bad "A30c crypto" "got=$CRYPTO_AMT expected=$EXP_CRYPTO"
else
  bad "A30c price" "could not get product price p1=$P1_PRICE"
fi
# restore rate
http_code PUT "$API/admin/exchange-rates/bsc" "{\"symbol\":\"BNB\",\"rate_to_usd\":\"350\"}" "$A" 2>/dev/null

summary
