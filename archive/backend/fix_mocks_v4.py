#!/usr/bin/env python3
"""Comprehensive mock SQL fixer - patches all mocks to match exact handler SQL."""
import re, subprocess

with open("/root/project/backend/handlers/all_features_test.go") as f:
    content = f.read()

lines = content.split('\n')

def patch(idx, new_text):
    """Replace line (0-indexed) if different."""
    if idx < len(lines) and lines[idx] != new_text:
        lines[idx] = new_text
        return True
    return False

def find_and_replace(pattern, replacement, desc="", start=0, end=None):
    """Find first line matching pattern and replace."""
    for i in range(start, len(lines) if end is None else end):
        if re.search(pattern, lines[i]):
            lines[i] = replacement
            print(f"  L{i+1} [{desc}]: patched")
            return True
    print(f"  NOT FOUND: {desc}")
    return False

# ============================================================
# Helper: build product list regex pattern
# ============================================================
def product_list_regex(where_extra="", extra_order="", limit_offset=True):
    """Build regex for product list query with JOIN."""
    cols = (
        r'p\.id,\s+p\.title,\s+p\.slug,\s+p\.description,\s+'
        r'p\.category_id,\s+COALESCE\(c\.name,\s+\'\',\s+\)\s+'
        r'p\.price_usd,\s+p\.asset_path,\s+p\.asset_hash,\s+p\.status,\s+'
        r'p\.download_count_limit,\s+p\.max_downloads_per_user,\s+'
        r'p\.file_size_bytes,\s+p\.file_mime_type,\s+p\.created_at,\s+'
        r'p\.updated_at'
    )
    parts = [
        r'SELECT\s+' + cols,
        r'\s+FROM\s+products\s+p',
        r'\s+LEFT\s+JOIN\s+categories\s+c\s+ON\s+p\.category_id\s*=\s*c\.id',
    ]
    if where_extra:
        parts.append(r'\s+' + where_extra)
    parts.append(r'\s+ORDER\s+BY\s+p\.created_at\s+DESC')
    if limit_offset:
        parts.append(r'\s+LIMIT\s+\$2\s+OFFSET\s+\$3')
    return '(?s)' + '\s*'.join(parts)

# ============================================================
# TestProducts_GetProducts (lines 120-133)
# ============================================================
patch(122,
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).')
patch(124,
    '\tmock.ExpectQuery(`' + product_list_regex(r'WHERE\s+p\.status\s*=\s*\$1') + '`).')

# ============================================================
# TestProducts_GetProductBySlug (lines 135-147)
# ============================================================
patch(137, '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).')
patch(139,
    '\tmock.ExpectQuery(`' + product_list_regex(r'WHERE\s+p\.slug\s*=\s*\$1', limit_offset=False) + '`).')
patch(143, '\tmock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).')

# ============================================================
# TestProducts_SearchProducts (lines 154-167)
# ============================================================
patch(156,
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE (p.title ILIKE $1 OR p.description ILIKE $1) AND p.status = $2`).')
patch(158,
    '\tmock.ExpectQuery(`' + product_list_regex(
        r'\(\s*p\.title\s+ILIKE\s+\$1\s+OR\s+p\.description\s+ILIKE\s+\$1\s*\)\s+AND\s+p\.status\s*=\s*\$2'
    ) + '`).')
patch(162, '\tmock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).')

# ============================================================
# TestCart_GetOrCreate (line ~472)
# ============================================================
find_and_replace(r'SELECT id FROM carts WHERE user_id = \\\$1`\)\.',
    '\tmock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).',
    "Cart GetOrCreate")
find_and_replace(r'INSERT INTO cart_items.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, $3) ON CONFLICT (cart_id, product_id) DO UPDATE SET quantity = $3, added_at = CURRENT_TIMESTAMP`).',
    "Cart AddItem INSERT")
find_and_replace(r'SELECT EXISTS\(SELECT 1 FROM cart_items',
    '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM cart_items WHERE cart_id = $1 AND product_id = $2)`).',
    "Cart AddItem EXISTS")
find_and_replace(r'DELETE FROM cart_items WHERE cart_id = \\\$1',
    '\tmock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).',
    "Cart RemoveItem")

# ============================================================
# TestWishlist_Toggle (line ~511)
# ============================================================
find_and_replace(r'SELECT EXISTS\(SELECT 1 FROM wishlists',
    '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).',
    "Wishlist Toggle EXISTS")
find_and_replace(r'INSERT INTO wishlists.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO wishlists (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).',
    "Wishlist Toggle INSERT")
find_and_replace(r'SELECT w\.product_id, p\.name, p\.slug.*`\)\.',
    '\tmock.ExpectQuery(`SELECT w.product_id, p.name, p.price_usd, p.asset_path, p.asset_hash, p.status FROM wishlists w JOIN products p ON w.product_id = p.id WHERE w.user_id = $1`).',
    "Wishlist List")

# ============================================================
# TestReviews_Create + List
# ============================================================
find_and_replace(r'INSERT INTO reviews.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) RETURNING id`).',
    "Reviews Create INSERT")
find_and_replace(r'SELECT r\.id, r\.rating, r\.comment.*`\)\.',
    '\tmock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.name FROM reviews r JOIN users u ON r.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC`).',
    "Reviews List")

# ============================================================
# TestRecentlyViewed
# ============================================================
find_and_replace(r'INSERT INTO recently_viewed.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = NOW()`).',
    "RecentlyViewed INSERT")
find_and_replace(r'SELECT p\.id, p\.name, p\.slug.*product_id FROM recently_viewed',
    '\tmock.ExpectQuery(`SELECT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM recently_viewed WHERE user_id = $1 ORDER BY viewed_at DESC LIMIT 10)`).',
    "RecentlyViewed List")

# ============================================================
# TestComparison
# ============================================================
find_and_replace(r'SELECT EXISTS\(SELECT 1 FROM product_comparison',
    '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).',
    "Comparison Toggle EXISTS")
find_and_replace(r'INSERT INTO product_comparison.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO product_comparison (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).',
    "Comparison Toggle INSERT")
find_and_replace(r'SELECT p\.id, p\.name, p\.slug.*product_id FROM product_comparison',
    '\tmock.ExpectQuery(`SELECT p.id, p.name, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM product_comparison WHERE user_id = $1 ORDER BY added_at DESC LIMIT 4)`).',
    "Comparison List")

# ============================================================
# TestRecommendations
# ============================================================
find_and_replace(r'SELECT category_id FROM products WHERE id = \\\$1',
    '\tmock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).',
    "Recommendations category_id")
find_and_replace(r'SELECT p\.id, p\.title.*FROM products p.*WHERE.*category_id.*$1.*status.*$2.*id.*!.*$3',
    '\tmock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.category_id = $1 AND p.status = $2 AND p.id != $3 ORDER BY p.created_at DESC LIMIT $4`).',
    "Recommendations products")

# ============================================================
# TestGuestOrder
# ============================================================
find_and_replace(r'INSERT INTO guest_orders.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO guest_orders (email, total_usd, crypto_chain, crypto_amount, crypto_address, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`).',
    "GuestOrder INSERT")
find_and_replace(r'SELECT id, email, status.*FROM guest_orders WHERE id = \\\$1',
    '\tmock.ExpectQuery(`SELECT id, email, status, total_usd, crypto_chain, crypto_amount, crypto_address FROM guest_orders WHERE id = $1`).',
    "GuestOrder SELECT")

# ============================================================
# TestOrderStatusCheck
# ============================================================
find_and_replace(r'SELECT o\.id, o\.status.*FROM orders WHERE id = \\\$1',
    '\tmock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).',
    "OrderStatusCheck")

# ============================================================
# TestNewProducts_GetActiveStat
# ============================================================
patch(725, "\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = 'active'`).")
patch(727, "\tmock.ExpectQuery(`SELECT COUNT(*) FROM products`).")
patch(729, "\tmock.ExpectQuery(`SELECT COUNT(*) FROM categories`).")

# ============================================================
# TestNewProducts_Delete
# ============================================================
find_and_replace(r'SELECT COUNT\(\\\\\\*\) FROM products WHERE slug = \\\$1',
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).',
    "NewProducts Delete COUNT")
find_and_replace(r'DELETE FROM product_images WHERE product_id = \\\$1',
    '\tmock.ExpectExec(`DELETE FROM product_images WHERE product_id = $1`).',
    "NewProducts Delete images")
find_and_replace(r'DELETE FROM products WHERE id = \\\$1',
    '\tmock.ExpectExec(`DELETE FROM products WHERE id = $1`).',
    "NewProducts Delete products")

# ============================================================
# TestExchangeRates
# ============================================================
find_and_replace(r'SELECT id, chain, symbol, rate_to_usd.*FROM exchange_rates ORDER BY chain',
    '\tmock.ExpectQuery(`SELECT id, chain, symbol, rate_to_usd, updated_at FROM exchange_rates ORDER BY chain`).',
    "ExchangeRates GetAll")
find_and_replace(r'SELECT id, chain, symbol, rate_to_usd.*FROM exchange_rates WHERE chain = \\\$1',
    '\tmock.ExpectQuery(`SELECT id, chain, symbol, rate_to_usd, updated_at FROM exchange_rates WHERE chain = $1`).',
    "ExchangeRates GetOne")

# ============================================================
# TestDBSchema_Migration011Tables
# ============================================================
find_and_replace(r"SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name = \\\$1`\)\.",
    "\tmock.ExpectQuery(`SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1`).",
    "DBSchema")

# ============================================================
# Auth_UpdateProfile fixes
# ============================================================
# Line 71: initial SELECT from users
patch(70, '\tmock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).')
# Line 75: SELECT from user_profiles
patch(74, '\tmock.ExpectQuery(`SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1`).')
# Line 78: UPDATE users SET
patch(77, '\t\tmock.ExpectExec(`UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).')
# Line 80: the second SELECT after update (line 80 in file was changed, fix it)
# Check current line 80 (idx 79)
if 'SELECT email, name FROM users' in lines[79]:
    lines[79] = '\tmock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).'
    print("  L80: Fixed SELECT after update")

# ============================================================
# Admin tests fixes
# ============================================================

# Admin Stats
find_and_replace(r'SELECT COUNT\(\\\\\\*\) FROM users.*`\)\.[\s\S]*?AddRow\(25\)',
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM users`).',
    "Admin Stats users")
find_and_replace(r'SELECT COUNT\(\\\\\\*\) FROM orders.*`\)\.[\s\S]*?AddRow\(\d+\)',
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM orders`).',
    "Admin Stats orders")
find_and_replace(r'SELECT COUNT\(\\\\\\*\) FROM categories.*`\)\.[\s\S]*?AddRow\(\d+\)',
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM categories`).',
    "Admin Stats categories")
find_and_replace(r'SELECT COALESCE.*`\)\.[\s\S]*?AddRow\(\'100\.00\'\)',
    '\tmock.ExpectQuery(`SELECT COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0) FROM products WHERE id = ANY($1::int[]) AND status = \'active\'`).',
    "Admin Stats revenue")

# Admin Order status transitions
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`UPDATE orders SET' in line:
        lines[i] = '\t\tmock.ExpectExec(`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`).'
        print(f"  L{i+1}: Fixed UPDATE orders SET")

# Admin GuestOrderCheck
find_and_replace(r'SELECT o\.id, o\.status.*FROM orders WHERE id = \\\$1.*`\)\.[\s\S]*?AddRow.*completed',
    '\tmock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).',
    "Admin GuestOrderCheck order")
find_and_replace(r'SELECT id, product_id, product_title.*FROM order_items WHERE order_id=\$1',
    '\tmock.ExpectQuery(`SELECT id, product_id, product_title, product_slug, quantity, CAST(unit_price_usd AS NUMERIC), download_count FROM order_items WHERE order_id = $1 ORDER BY id`).',
    "Admin GuestOrderCheck items")

# Admin GetSetting
find_and_replace(r'SELECT setting_key, value, updated_at FROM settings ORDER BY setting_key',
    '\tmock.ExpectQuery(`SELECT setting_key, value, updated_at FROM settings ORDER BY setting_key`).',
    "Admin GetSettings list")

# Admin SetSetting
find_and_replace(r'INSERT INTO settings.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO settings (setting_key, value) VALUES ($1, $2) ON CONFLICT (setting_key) DO UPDATE SET value = $2`).',
    "Admin SetSetting")

# Admin ExchangeRateSet
find_and_replace(r'INSERT INTO exchange_rates.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO exchange_rates (chain, symbol, rate_to_usd) VALUES ($1, $2, $3) ON CONFLICT (chain) DO UPDATE SET symbol = $2, rate_to_usd = $3, updated_at = NOW()`).',
    "Admin ExchangeRateSet")

# Admin DeleteUser
find_and_replace(r'DELETE FROM users WHERE id = \\\$1.*`\)\.',
    '\tmock.ExpectExec(`DELETE FROM users WHERE id = $1`).',
    "Admin DeleteUser")

# Admin DeleteCommunityPost
find_and_replace(r'DELETE FROM community_posts WHERE id = \\\$1.*`\)\.',
    '\tmock.ExpectExec(`DELETE FROM community_posts WHERE id = $1`).',
    "Admin DeleteCommunityPost")

# Admin CommunityPostList
find_and_replace(r'SELECT cp\.id, cp\.user_id.*FROM community_posts cp.*`\)\.[\s\S]*?AddRow.*post',
    '\tmock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, COALESCE(cl.cnt, 0) + COALESCE(cc.cnt, 0) FROM community_posts cp LEFT JOIN users u ON cp.user_id = u.id LEFT JOIN (SELECT post_id, COUNT(*) as cnt FROM likes WHERE type = \'post\' GROUP BY post_id) cl ON cp.id = cl.post_id LEFT JOIN (SELECT post_id, COUNT(*) as cnt FROM comments WHERE type = \'post\' GROUP BY post_id) cc ON cp.id = cc.post_id WHERE cp.type = \'post\' ORDER BY cp.created_at DESC LIMIT $3 OFFSET $4`).',
    "Admin CommunityPostList", end=380)

# Admin CommunityUserList
find_and_replace(r'SELECT u\.id, u\.email, u\.name.*`\)\.[\s\S]*?AddRow.*test@test.com',
    '\tmock.ExpectQuery(`SELECT u.id, u.email, u.name, COUNT(cp.id) FROM users u LEFT JOIN community_posts cp ON u.id = cp.user_id WHERE cp.type = \'post\' GROUP BY u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`).',
    "Admin CommunityUserList", end=390)

# Admin Bulk operations
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`UPDATE products SET' in line and i > 200:
        lines[i] = '\t\tmock.ExpectExec(`UPDATE products SET status = $1, category_id = $2, updated_at = NOW() WHERE id = ANY($3)`).'
        print(f"  L{i+1}: Fixed bulk product UPDATE")

# ============================================================
# Cart mocks - fix WHERE clause style ($1 vs WHERE $1)
# ============================================================
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT id FROM carts WHERE user_id=$1`' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).'
        print(f"  L{i+1}: Fixed cart SELECT = $1")
    if 'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM cart_items WHERE cart_id=$1' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM cart_items WHERE cart_id = $1 AND product_id = $2)`).'
        print(f"  L{i+1}: Fixed cart EXISTS")
    if 'mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id=$1' in line:
        lines[i] = '\tmock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).'
        print(f"  L{i+1}: Fixed cart DELETE")
    if 'mock.ExpectExec(`INSERT INTO cart_items' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, $3) ON CONFLICT (cart_id, product_id) DO UPDATE SET quantity = $3, added_at = CURRENT_TIMESTAMP`).'
        print(f"  L{i+1}: Fixed cart INSERT")

# ============================================================
# Write result
# ============================================================
result = '\n'.join(lines)
with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(result)

print(f"\nDone. {len(result)} chars. Running build+test...")

r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True, cwd="/root/project/backend")
print(f"Build: exit={r.returncode}")
if r.returncode != 0:
    print(r.stderr[-400:])

r = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True, cwd="/root/project/backend")
out = r.stdout
passes = out.count('--- PASS:')
fails = out.count('--- FAIL:')
print(f"\nResult: {passes} PASS, {fails} FAIL, exit={r.returncode}")
if r.returncode != 0:
    print(out[-800:])
