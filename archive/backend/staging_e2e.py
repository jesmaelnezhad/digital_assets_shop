import urllib.request, urllib.error, json, sys, time

BASE = "http://127.0.0.1/staging/api/v1"

def req(method, path, data=None, token=None):
    url = f"{BASE}/{path}"
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    body = json.dumps(data).encode() if data else None
    try:
        r = urllib.request.urlopen(urllib.request.Request(url, data=body, headers=headers, method=method), timeout=5)
        return r.status, json.loads(r.read())
    except urllib.error.HTTPError as e:
        try:
            return e.code, json.loads(e.read())
        except:
            return e.code, {}

def check(name, status, expect=None):
    ok = "PASS" if (expect is None or status == expect) else f"FAIL (expected {expect}, got {status})"
    print(f"  [{ok}] {name}: HTTP {status}")
    return status == (expect or status)

# 1. Register
print("=== 1. AUTH: Register ===")
st, d = req("POST", "register", {"email": f"e2e{int(time.time())}@test.com", "password": "testpass123", "name": "E2E Tester"})
token = d.get("token", "")
check("Register returns token", st, 201)
print(f"      Token len: {len(token)}")

# 2. Login with register token
print("\n=== 2. AUTH: GetMe ===")
st, d = req("GET", "me", token=token)
check("GetMe authorized", st, 200)
if st == 200:
    print(f"      User: {d.get('name')} <{d.get('email')}>")
    uid = d.get("id")

# 3. Products
print("\n=== 3. PRODUCTS: List ===")
st, d = req("GET", "products?page=1&per_page=2")
check("Products list", st, 200)
if st == 200:
    print(f"      Total: {d.get('total')}, returned: {len(d.get('products',[]))}")
    for p in d.get("products", [])[:2]:
        print(f"      [{p['id']}] {p['title']} - ${p['price_usd']} ({p['status']})")

# 4. Product by slug
print("\n=== 4. PRODUCT: By slug ===")
st, d = req("GET", "products/chiptune-loop-library")
check("Product by slug", st, 200)
if st == 200:
    print(f"      {d.get('title')} - ${d.get('price_usd')} [{d.get('status')}] dl/user={d.get('max_downloads_per_user')}")

# 5. Categories
print("\n=== 5. CATEGORIES ===")
st, d = req("GET", "categories")
check("Categories", st, 200)
if st == 200:
    cats = d.get("categories", [])
    print(f"      {len(cats)} categories")
    for c in cats:
        print(f"      [{c['id']}] {c['name']} ({c['slug']})")

# 6. Community feed
print("\n=== 6. COMMUNITY: Feed ===")
st, d = req("GET", "community/feed", token=token)
check("Community feed", st, 200)
if st == 200:
    posts = d.get("posts", [])
    print(f"      {len(posts)} posts")
    if posts:
        p = posts[0]
        print(f"      Latest: \"{p['content'][:50]}...\" by {p['user']['name']} likes={p['like_count']} cmts={p['comment_count']}")

# 7. Cart
print("\n=== 7. CART: GetOrCreate ===")
st, d = req("GET", "cart", token=token)
check("Cart get", st, 200)
if st == 200:
    print(f"      Cart ID: {d.get('cart_id')} - {d.get('message')}")

# 8. Cart add
print("\n=== 8. CART: Add item ===")
st, d = req("POST", "cart/add", {"product_id": 1, "quantity": 2}, token=token)
check("Cart add", st, 200)
print(f"      {d}")

# 9. Cart view
print("\n=== 9. CART: View ===")
st, d = req("GET", "cart/items", token=token)
check("Cart view", st, 200)
if st == 200:
    print(f"      Items: {len(d.get('items',[]))}, Total: {d.get('total_usd')}")

# 10. Wishlist toggle
print("\n=== 10. WISHLIST: Toggle ===")
st, d = req("POST", "wishlist/toggle", {"product_id": 1}, token=token)
check("Wishlist toggle", st, 200)
print(f"      Added: {d.get('added')}")

# 11. Wishlist list
print("\n=== 11. WISHLIST: List ===")
st, d = req("GET", "wishlist", token=token)
check("Wishlist list", st, 200)
print(f"      Items: {len(d.get('items',[]))}")

# 12. Reviews create
print("\n=== 12. REVIEWS: Create ===")
st, d = req("POST", "reviews", {"product_id": 1, "rating": 5, "comment": "Great from e2e test"}, token=token)
check("Review create", st, 201)
print(f"      Review ID: {d.get('id')}")

# 13. Reviews list
print("\n=== 13. REVIEWS: List ===")
st, d = req("GET", "reviews?product_id=1")
check("Reviews list", st, 200)
print(f"      Reviews: {len(d.get('reviews',[]))}")

# 14. Recently viewed
print("\n=== 14. RECENTLY VIEWED ===")
st, d = req("POST", "recently-viewed", {"product_id": 1}, token=token)
check("Recently viewed record", st, 200)
st, d = req("GET", "recently-viewed", token=token)
check("Recently viewed list", st, 200)
print(f"      Viewed: {len(d.get('products',[]))}")

# 15. Compare
print("\n=== 15. COMPARE ===")
st, d = req("POST", "compare/toggle", {"product_id": 1}, token=token)
check("Compare toggle", st, 200)
st, d = req("GET", "compare", token=token)
check("Compare list", st, 200)
print(f"      Items: {len(d.get('products',[]))}")

# 16. Recommendations
print("\n=== 16. RECOMMENDATIONS ===")
st, d = req("GET", "recommendations/1")
check("Recommendations", st, 200)
if st == 200:
    recs = d.get('recommendations', [])
    tp = d.get('top_product', {})
    tp_name = tp.get('title', '?') if isinstance(tp, dict) else '?'
    print(f"      Recs: {len(recs)}, top: {tp_name}")

# 17. Guest order
print("\n=== 17. GUEST ORDER ===")
st, d = req("POST", "guest-orders", {"email": "guest-e2e@example.com", "product_ids": "1,2"})
check("Guest order create", st, 201)
print(f"      Order: {d.get('order_id')} status={d.get('status')}")

# 18. Exchange rates
print("\n=== 18. EXCHANGE RATES ===")
st, d = req("GET", "exchange-rates")
check("Exchange rates", st, 200)
if st == 200:
    rates = d.get("rates", [])
    parts = []
    for r in rates:
        parts.append(f"{r.get('chain','?')}={r.get('rate','?')}")
    print(f"      {len(rates)} rates: {', '.join(parts)}")

# 19. Order create
print("\n=== 19. ORDER (auth) ===")
st, d = req("POST", "orders", {"product_ids": "1", "quantities": "1"}, token=token)
check("Order create", st, 201)
print(f"      Order: {d.get('order_id')} status={d.get('status')} total={d.get('total_usd')}")

# 20. Admin
print("\n=== 20. ADMIN ===")
admin = "admin123"
st, ad = req("POST", "login", {"email": "admin@example.com", "password": admin})
admin_token = ad.get("token", "")
check("Admin login", st, 200)
print(f"      Admin token len: {len(admin_token)}")

if admin_token:
    st, d = req("GET", "admin/users", token=admin_token)
    check("Admin users", st, 200)
    print(f"      Users: {len(d.get('users',[]))}")
    
    st, d = req("GET", "admin/products", token=admin_token)
    check("Admin products", st, 200)
    print(f"      Products: {len(d.get('products',[]))}")
    
    st, d = req("GET", "admin/stats", token=admin_token)
    check("Admin stats", st, 200)
    print(f"      Stats: {d}")
    
    st, d = req("GET", "admin/exchange-rates", token=admin_token)
    check("Admin exchange rates", st, 200)
    print(f"      Exchange rates: {len(d.get('rates',[]))}")

print("\n=== DONE ===")
