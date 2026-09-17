#!/usr/bin/env python3
"""Fix all mock SQL patterns to match handler queries."""
import re, subprocess

with open("/root/project/backend/handlers/all_features_test.go") as f:
    content = f.read()

lines = content.split('\n')

def replace_line(idx, new_line):
    """Replace line (0-indexed) with new content."""
    if idx < len(lines):
        lines[idx] = new_line
        print(f"  L{idx+1}: patched")
        return True
    return False

def find_and_replace(pattern, replacement, desc="", start=0, end=None):
    """Find first matching line and replace."""
    for i in range(start, len(lines) if end is None else end):
        if re.search(pattern, lines[i]):
            lines[i] = replacement
            print(f"  L{i+1} [{desc}]: patched")
            return True
    print(f"  NOT FOUND: {desc}")
    return False

# ================================================================
# Helper: build product list mocking pattern 
# ================================================================
def prod_list_pattern(where_clause="$1", limit="$2", offset="$3"):
    """Build sqlmock-compatible regex for product list query."""
    return (
        r'(?s)SELECT\s+p\.id,\s+p\.title,\s+p\.slug,\s+p\.description,\s+'
        r'p\.category_id,\s+COALESCE\(c\.name,\s+\'\',\s+\)\s+'
        r'p\.price_usd,\s+p\.asset_path,\s+p\.asset_hash,\s+p\.status,\s+'
        r'p\.download_count_limit,\s+p\.max_downloads_per_user,\s+'
        r'p\.file_size_bytes,\s+p\.file_mime_type,\s+p\.created_at,\s+'
        r'p\.updated_at\s+FROM\s+products\s+p\s+LEFT\s+JOIN\s+categories\s+c\s+'
        r'ON\s+p\.category_id\s*=\s*c\.id\s+' + where_clause + r'\s+'
        r'ORDER\s+BY\s+p\.created_at\s+DESC\s+LIMIT\s+' + limit + r'\s+OFFSET\s+' + offset
    )

# ================================================================
# TestProducts_GetProducts (L120-133) 
# ================================================================
replace_line(122,
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n')
replace_line(124,
    '\tmock.ExpectQuery(`' + prod_list_pattern() + '`).\n')

# ================================================================
# TestProducts_GetProductBySlug (L135-147)
# ================================================================
replace_line(137, '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n')
replace_line(139,
    '\tmock.ExpectQuery(`' + prod_list_pattern('WHERE p.slug = $1', '', '') + '`).\n')
replace_line(143, '\tmock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n')

# ================================================================
# TestProducts_SearchProducts (L154-167)
# ================================================================
replace_line(156,
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE (p.title ILIKE $1 OR p.description ILIKE $1) AND p.status = $2`).\n')
replace_line(158,
    '\tmock.ExpectQuery(`' + prod_list_pattern(
        r'\(p\.title\s+ILIKE\s+\$1\s+OR\s+p\.description\s+ILIKE\s+\$1\)\s+AND\s+p\.status\s*=\s*\$2',
        '$3', '$4') + '`).\n')
replace_line(162, '\tmock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n')

# ================================================================
# TestCart_GetOrCreate (L471-487)
# ================================================================
find_and_replace(r'SELECT id FROM carts WHERE user_id = \\\$1`\)\.',
    '\tmock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).', "Cart GetOrCreate")
find_and_replace(r'INSERT INTO cart_items.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, $3) ON CONFLICT (cart_id, product_id) DO UPDATE SET quantity = $3, added_at = CURRENT_TIMESTAMP`).',
    "Cart AddItem INSERT")
find_and_replace(r'SELECT EXISTS\(SELECT 1 FROM cart_items',
    '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM cart_items WHERE cart_id = $1 AND product_id = $2)`).',
    "Cart AddItem EXISTS check")
find_and_replace(r'DELETE FROM cart_items WHERE cart_id = \\\$1',
    '\tmock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).',
    "Cart RemoveItem")

# ================================================================
# TestWishlist
# ================================================================
find_and_replace(r'SELECT EXISTS\(SELECT 1 FROM wishlists',
    '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).',
    "Wishlist Toggle EXISTS")
find_and_replace(r'INSERT INTO wishlists.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO wishlists (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).',
    "Wishlist Toggle INSERT")
find_and_replace(r'SELECT w\.product_id,\s*p\.name,\s*p\.slug.*`\)\.',
    '\tmock.ExpectQuery(`SELECT w.product_id, p.name, p.price_usd, p.asset_path, p.asset_hash, p.status FROM wishlists w JOIN products p ON w.product_id = p.id WHERE w.user_id = $1`).',
    "Wishlist List")

# ================================================================
# TestReviews
# ================================================================
find_and_replace(r'INSERT INTO reviews.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) RETURNING id`).',
    "Reviews Create INSERT")
find_and_replace(r'SELECT r\.id,\s*r\.rating,\s*r\.comment.*`\)\.',
    '\tmock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.name FROM reviews r JOIN users u ON r.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC`).',
    "Reviews List")

# ================================================================
# TestRecentlyViewed
# ================================================================
find_and_replace(r'INSERT INTO recently_viewed.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = NOW()`).',
    "RecentlyViewed Record INSERT")
find_and_replace(r'SELECT p\.id,\s*p\.name,\s*p\.slug.*FROM products p.*WHERE.*recently_viewed',
    '\tmock.ExpectQuery(`SELECT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM recently_viewed WHERE user_id = $1 ORDER BY viewed_at DESC LIMIT 10)`).',
    "RecentlyViewed List")

# ================================================================
# TestComparison
# ================================================================
find_and_replace(r'SELECT EXISTS\(SELECT 1 FROM product_comparison',
    '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).',
    "Comparison Toggle EXISTS")
find_and_replace(r'INSERT INTO product_comparison.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO product_comparison (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).',
    "Comparison Toggle INSERT")
find_and_replace(r'SELECT p\.id,\s*p\.name,\s*p\.slug.*FROM products p.*WHERE.*product_comparison',
    '\tmock.ExpectQuery(`SELECT p.id, p.name, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM product_comparison WHERE user_id = $1 ORDER BY added_at DESC LIMIT 4)`).',
    "Comparison List")

# ================================================================
# TestRecommendations
# ================================================================
find_and_replace(r'SELECT category_id FROM products WHERE id = \\\$1`\)\.',
    '\tmock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).',
    "Recommendations category_id")
find_and_replace(r'SELECT p\.id,\s*p\.title.*FROM products p.*WHERE.*category_id.*$1.*AND.*status.*$2.*AND.*id.*!.*$3',
    '\tmock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.category_id = $1 AND p.status = $2 AND p.id != $3 ORDER BY p.created_at DESC LIMIT $4`).',
    "Recommendations products")
find_and_replace(r'SELECT p\.id,\s*p\.name.*FROM products p.*WHERE.*id\s*=\s*\$1',
    '\tmock.ExpectQuery(`SELECT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id = $1`).',
    "Recommendations similar product")

# ================================================================
# TestGuestOrder
# ================================================================
find_and_replace(r'INSERT INTO guest_orders.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO guest_orders (email, total_usd, crypto_chain, crypto_amount, crypto_address, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`).',
    "GuestOrder Create INSERT")
find_and_replace(r'SELECT id,\s*email,\s*status.*FROM guest_orders WHERE id = \\\$1`\)\.',
    '\tmock.ExpectQuery(`SELECT id, email, status, total_usd, crypto_chain, crypto_amount, crypto_address FROM guest_orders WHERE id = $1`).',
    "GuestOrder Check SELECT")

# ================================================================
# TestOrderStatusCheck
# ================================================================
find_and_replace(r'SELECT o\.id,\s*o\.status.*FROM orders WHERE id = \\\$1`\)\.',
    '\tmock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).',
    "OrderStatusCheck")

# ================================================================
# TestNewProducts_GetActiveStat (L725-730)
# ================================================================
replace_line(725, "\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = 'active'`).\n")
replace_line(727, "\tmock.ExpectQuery(`SELECT COUNT(*) FROM products`).\n")
replace_line(729, "\tmock.ExpectQuery(`SELECT COUNT(*) FROM categories`).\n")

# ================================================================
# TestNewProducts_Delete (L691-698)
# ================================================================
find_and_replace(r'SELECT COUNT\(\\\\\\*\) FROM products WHERE slug = \\\$1',
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).',
    "NewProducts Delete COUNT")
find_and_replace(r'SELECT COUNT\(\\\\\\*\) FROM categories WHERE id = \\\$1',
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM categories WHERE id = $1`).',
    "NewProducts Delete category COUNT")

# ================================================================
# TestExchangeRates
# ================================================================
find_and_replace(r'SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`\)\.',
    '\tmock.ExpectQuery(`SELECT id, chain, symbol, rate_to_usd, updated_at FROM exchange_rates ORDER BY chain`).',
    "ExchangeRates GetAll")
find_and_replace(r'SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = \\\$1`\)\.',
    '\tmock.ExpectQuery(`SELECT id, chain, symbol, rate_to_usd, updated_at FROM exchange_rates WHERE chain = $1`).',
    "ExchangeRates GetOne")

# ================================================================
# TestDBSchema_Migration011Tables
# ================================================================
find_and_replace(r"SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name = \\\$1`\)\.",
    "\tmock.ExpectQuery(`SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1`).",
    "DBSchema")

# ================================================================
# Auth_UpdateProfile fixes (L71-84)
# ================================================================
replace_line(70, '\tmock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).\n')
replace_line(74, '\tmock.ExpectQuery(`SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1`).\n')
# UPDATE users SET - handler builds dynamically
find_and_replace(r'mock\.ExpectExec\(`UPDATE users SET.*`\)\.',
    '\t\tmock.ExpectExec(`UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).',
    "Auth UpdateProfile UPDATE")

# ================================================================
# Write result
# ================================================================
result = '\n'.join(lines)
with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(result)

print(f"\nDone. {len(result)} chars. Build+test...")

r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True, cwd="/root/project/backend")
print(f"Build: exit={r.returncode}")
if r.returncode != 0:
    print("STDERR:", r.stderr[-500:])

r = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True, cwd="/root/project/backend")
out = r.stdout
p = out.count('--- PASS:')
f = out.count('--- FAIL:')
print(f"\nResult: {p} PASS, {f} FAIL, exit={r.returncode}")
if r.returncode != 0:
    print(out[-800:])
