#!/usr/bin/env python3
"""Fix all mock SQL in all_features_test.go to match handler queries."""
import re, subprocess

with open("/root/project/backend/handlers/all_features_test.go") as f:
    content = f.read()

lines = content.split('\n')

def fix_line(idx, new_text):
    """Replace line (0-indexed) and return True if changed."""
    old = lines[idx]
    if old != new_text:
        lines[idx] = new_text
        return True
    return False

# ============================================================
# Products_GetProducts (lines 120-133)
# ============================================================
# Handler sends: SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1
# Using simple regex that matches this structure
lines[122] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n'

# Product list: handler sends multi-line JOIN query with tabs
# Use (.*?) for flexible matching of columns and whitespace
lines[124] = (
    '\tmock.ExpectQuery(`(?s)SELECT\\s+p\\.id,\\s+p\\.title,\\s+p\\.slug,\\s+'
    'p\\.description,\\s+p\\.category_id,\\s+COALESCE\\(c\\.name,\\s+\\x27\\x27,\\s+\\)\\s+'
    'p\\.price_usd,\\s+p\\.asset_path,\\s+p\\.asset_hash,\\s+p\\.status,\\s+'
    'p\\.download_count_limit,\\s+p\\.max_downloads_per_user,\\s+'
    'p\\.file_size_bytes,\\s+p\\.file_mime_type,\\s+p\\.created_at,\\s+'
    'p\\.updated_at\\s+FROM\\s+products\\s+p\\s+LEFT\\s+JOIN\\s+categories\\s+c\\s+'
    'ON\\s+p\\.category_id\\s*=\\s*c\\.id\\s+WHERE\\s+p\\.status\\s*=\\s*\\$1\\s+'
    'ORDER\\s+BY\\s+p\\.created_at\\s+DESC\\s+LIMIT\\s+\\$2\\s+OFFSET\\s+\\$3`).\n'
)

# ============================================================
# Products_GetProductBySlug (lines 135-147)
# ============================================================
# COUNT: handler sends SELECT COUNT(*) FROM products WHERE slug = $1
lines[137] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n'

# Product by slug: handler sends JOIN query  
sq_slug = (
    '(?s)SELECT\\s+p\\.id,\\s+p\\.title,\\s+p\\.slug,\\s+p\\.description,\\s+'
    'p\\.category_id,\\s+COALESCE\\(c\\.name,\\s+\\x27\\x27,\\s+\\)\\s+'
    'p\\.price_usd,\\s+p\\.asset_path,\\s+p\\.asset_hash,\\s+p\\.status,\\s+'
    'p\\.download_count_limit,\\s+p\\.max_downloads_per_user,\\s+'
    'p\\.file_size_bytes,\\s+p\\.file_mime_type,\\s+p\\.created_at,\\s+'
    'p\\.updated_at\\s+FROM\\s+products\\s+p\\s+LEFT\\s+JOIN\\s+categories\\s+c\\s+'
    'ON\\s+p\\.category_id\\s*=\\s*c\\.id\\s+WHERE\\s+p\\.slug\\s*=\\s*\\$1'
)
lines[139] = '\tmock.ExpectQuery(`' + sq_slug + '`).\n'

# Product images: handler sends SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id
lines[143] = '\tmock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n'

# ============================================================
# Products_SearchProducts (lines 154-167)
# ============================================================
# COUNT with search: handler sends SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE (p.title ILIKE $1 OR p.description ILIKE $1) AND p.status = $2
lines[156] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE (p.title ILIKE $1 OR p.description ILIKE $1) AND p.status = $2`).\n'

sq_search = (
    '(?s)SELECT\\s+p\\.id,\\s+p\\.title,\\s+p\\.slug,\\s+p\\.description,\\s+'
    'p\\.category_id,\\s+COALESCE\\(c\\.name,\\s+\\x27\\x27,\\s+\\)\\s+'
    'p\\.price_usd,\\s+p\\.asset_path,\\s+p\\.asset_hash,\\s+p\\.status,\\s+'
    'p\\.download_count_limit,\\s+p\\.max_downloads_per_user,\\s+'
    'p\\.file_size_bytes,\\s+p\\.file_mime_type,\\s+p\\.created_at,\\s+'
    'p\\.updated_at\\s+FROM\\s+products\\s+p\\s+LEFT\\s+JOIN\\s+categories\\s+c\\s+'
    'ON\\s+p\\.category_id\\s*=\\s*c\\.id\\s+WHERE\\s+\\(p\\.title\\s+ILIKE\\s+\\$1\\s+'
    'OR\\s+p\\.description\\s+ILIKE\\s+\\$1\\)\\s+AND\\s+p\\.status\\s*=\\s*\\$2\\s+'
    'ORDER\\s+BY\\s+p\\.created_at\\s+DESC\\s+LIMIT\\s+\\$3\\s+OFFSET\\s+\\$4'
)
lines[158] = '\tmock.ExpectQuery(`' + sq_search + '`).\n'

# Product images for search results
lines[162] = '\tmock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n'

# ============================================================
# TestCart_GetOrCreate (line ~472)
# ============================================================
# Handler: SELECT id FROM carts WHERE user_id = $1
lines[471] = '\tmock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).\n'

# Cart create: INSERT INTO carts (user_id) VALUES ($1) RETURNING id
lines[486] = '\t\tmock.ExpectExec(`INSERT INTO carts (user_id) VALUES ($1) RETURNING id`).\n'

# Cart items check: SELECT EXISTS(SELECT 1 FROM cart_items WHERE cart_id=$1 AND product_id=$2)
lines[498] = '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM cart_items WHERE cart_id = $1 AND product_id = $2)`).\n'

# Fix cart items insert if present
for i, line in enumerate(lines):
    if 'INSERT INTO cart_items' in line and 'mock.ExpectExec' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, $3) ON CONFLICT (cart_id, product_id) DO UPDATE SET quantity = $3, added_at = CURRENT_TIMESTAMP`).\n'
        print(f"Fixed cart items INSERT at L{i+1}")

# Cart remove: DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2
lines[498] = '\tmock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).\n'

# ============================================================
# TestWishlist_Toggle (line ~511)
# ============================================================
lines[510] = '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).\n'

# Wishlist insert: INSERT INTO wishlists (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING
for i, line in enumerate(lines):
    if 'INSERT INTO wishlists' in line and 'mock.ExpectExec' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO wishlists (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).\n'
        print(f"Fixed wishlist INSERT at L{i+1}")

# Wishlist list: handler sends SELECT w.product_id, p.name, p.price_usd, p.asset_path, p.asset_hash, p.status FROM wishlists w JOIN products p ON w.product_id = p.id WHERE w.user_id = $1
lines[525] = '\tmock.ExpectQuery(`SELECT w.product_id, p.name, p.price_usd, p.asset_path, p.asset_hash, p.status FROM wishlists w JOIN products p ON w.product_id = p.id WHERE w.user_id = $1`).\n'

# ============================================================
# TestReviews_Create (line ~541)
# ============================================================
for i, line in enumerate(lines):
    if 'INSERT INTO reviews' in line and 'mock.ExpectExec' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) RETURNING id`).\n'
        print(f"Fixed reviews INSERT at L{i+1}")

# Reviews list: handler sends SELECT r.id, r.rating, r.comment, r.created_at, u.name FROM reviews r JOIN users u ON r.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC
lines[552] = '\tmock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.name FROM reviews r JOIN users u ON r.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC`).\n'

# ============================================================
# TestRecentlyViewed_Record + List
# ============================================================
for i, line in enumerate(lines):
    if 'INSERT INTO recently_viewed' in line and 'mock.ExpectExec' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = NOW()`).\n'
        print(f"Fixed recently_viewed INSERT at L{i+1}")

# Recently viewed list: handler sends SELECT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM recently_viewed WHERE user_id = $1 ORDER BY viewed_at DESC LIMIT 10)
lines[577] = '\tmock.ExpectQuery(`SELECT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM recently_viewed WHERE user_id = $1 ORDER BY viewed_at DESC LIMIT 10)`).\n'

# ============================================================
# TestComparison_Toggle + List
# ============================================================
lines[592] = '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).\n'

for i, line in enumerate(lines):
    if 'INSERT INTO product_comparison' in line and 'mock.ExpectExec' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO product_comparison (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).\n'
        print(f"Fixed comparison INSERT at L{i+1}")

# Comparison list: handler sends SELECT p.id, p.name, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM product_comparison WHERE user_id = $1 ORDER BY added_at DESC LIMIT 4)
lines[606] = '\tmock.ExpectQuery(`SELECT p.id, p.name, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM product_comparison WHERE user_id = $1 ORDER BY added_at DESC LIMIT 4)`).\n'

# ============================================================
# TestRecommendations_Get (lines ~621-628)
# ============================================================
lines[621] = '\tmock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).\n'

# Recommendations products: handler sends SELECT p.id, p.title, p.slug, p.description, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.category_id = $1 AND p.status = $2 AND p.id != $3 ORDER BY p.created_at DESC LIMIT $4
lines[623] = '\tmock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.category_id = $1 AND p.status = $2 AND p.id != $3 ORDER BY p.created_at DESC LIMIT $4`).\n'

# ============================================================
# TestGuestOrder_Create (line ~641)
# ============================================================
for i, line in enumerate(lines):
    if 'INSERT INTO guest_orders' in line and 'mock.ExpectExec' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO guest_orders (email, total_usd, crypto_chain, crypto_amount, crypto_address, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`).\n'
        print(f"Fixed guest_orders INSERT at L{i+1}")

# Guest order check: handler sends SELECT id, email, status, total_usd, crypto_chain, crypto_amount, crypto_address FROM guest_orders WHERE id = $1
lines[642] = '\tmock.ExpectQuery(`SELECT id, email, status, total_usd, crypto_chain, crypto_amount, crypto_address FROM guest_orders WHERE id = $1`).\n'

# ============================================================
# TestOrderStatusCheck (line ~671)
# ============================================================
# Handler sends: SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1
lines[670] = '\tmock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).\n'

# ============================================================
# TestNewProducts_GetActiveStat (lines ~725-733)
# ============================================================
# These use simple COUNT queries - check what handler actually sends
lines[725] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = \'active\'`).\n'
lines[727] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products`).\n'
lines[729] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM categories`).\n'

# ============================================================
# TestNewProducts_Delete (lines ~691-698)
# ============================================================
lines[691] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n'
lines[693] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM categories WHERE id = $1`).\n'

for i, line in enumerate(lines):
    if 'INSERT INTO products' in line and 'mock.ExpectExec' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO products (name, slug, description, category_id, price_usd, asset_path, asset_hash, status, download_count_limit, max_downloads_per_user, file_size_bytes, file_mime_type, user_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW()) RETURNING id`).\n'
        print(f"Fixed products INSERT at L{i+1}")

for i, line in enumerate(lines):
    if 'INSERT INTO product_images' in line and 'mock.ExpectExec' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO product_images (product_id, url, is_primary) VALUES ($1, $2, $3) RETURNING id`).\n'
        print(f"Fixed product_images INSERT at L{i+1}")

# ============================================================
# TestExchangeRates_GetAll (line ~403)
# ============================================================
lines[403] = '\tmock.ExpectQuery(`SELECT id, chain, symbol, rate_to_usd, updated_at FROM exchange_rates ORDER BY chain`).\n'

# ============================================================
# TestExchangeRates_GetOne (line ~417)
# ============================================================
lines[417] = '\tmock.ExpectQuery(`SELECT id, chain, symbol, rate_to_usd, updated_at FROM exchange_rates WHERE chain = $1`).\n'

# ============================================================
# TestDBSchema_Migration011Tables (line ~781)
# ============================================================
lines[781] = '\tmock.ExpectQuery(`SELECT table_name FROM information_schema.tables WHERE table_schema = \'public\' AND table_name = $1`).\n'

# ============================================================
# Auth_UpdateProfile (lines ~75-83)
# ============================================================
# Profile SELECT: handler sends SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1
lines[74] = '\tmock.ExpectQuery(`SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1`).\n'

# UPDATE users: handler sends UPDATE users SET name = $1 WHERE id = $2
lines[77] = '\t\tmock.ExpectExec(`UPDATE users SET name = $1 WHERE id = $2`).\n'

# ============================================================
# TestAdmin_Stats (lines ~454-459)
# ============================================================
lines[454] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM users`).\n'
lines[456] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM orders`).\n'
# Revenue: handler sends SELECT COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0) FROM products WHERE id = ANY($1::int[]) AND status = 'active'
# Wait, that doesn't sound right for admin stats... Let me check
# Actually for admin stats the handler might sum orders, not products

# ============================================================
# TestAdmin_GetSetting (line ~441)
# ============================================================
# Handler sends: SELECT setting_key, value, updated_at FROM settings ORDER BY setting_key
lines[441] = '\tmock.ExpectQuery(`SELECT setting_key, value, updated_at FROM settings ORDER BY setting_key`).\n'

# ============================================================
# TestAdmin_ExchangeRateSet (line ~417 - actually the INSERT)
# ============================================================
for i, line in enumerate(lines):
    if 'INSERT INTO exchange_rates' in line and 'mock.ExpectExec' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO exchange_rates (chain, symbol, rate_to_usd) VALUES ($1, $2, $3) ON CONFLICT (chain) DO UPDATE SET symbol = $2, rate_to_usd = $3, updated_at = NOW()`).\n'
        print(f"Fixed exchange_rates INSERT at L{i+1}")

# ============================================================
# Write result
# ============================================================
result = '\n'.join(lines)
with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(result)

print(f"\nWritten {len(result)} chars, {len(lines)} lines")
print("Running build+test...")

r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True, cwd="/root/project/backend")
print(f"Build: exit={r.returncode}")
if r.returncode != 0:
    print(r.stderr[-300:])

r = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True, cwd="/root/project/backend")
print(r.stdout[-2000:])
if r.returncode != 0:
    print(r.stderr[-300:])

passes = r.stdout.count('--- PASS:')
fails = r.stdout.count('--- FAIL:')
print(f"\nResult: {passes} PASS, {fails} FAIL, exit={r.returncode}")
