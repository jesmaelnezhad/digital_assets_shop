#!/usr/bin/env python3
"""
Comprehensive mock SQL fixer for all_features_test.go.
Fixes SQL patterns to match exact handler queries.
"""
import re

with open("/root/project/backend/handlers/all_features_test.go") as f:
    content = f.read()

lines = content.split('\n')

def patch_line(lnum, new_text):
    """Replace line (1-indexed) with new text."""
    idx = lnum - 1
    if idx < len(lines):
        old = lines[idx]
        lines[idx] = new_text
        print(f"  L{lnum}: {old[:80]}")
        return True
    return False

def find_and_patch(pattern, replacement, desc="", start=0, end=None):
    """Find first line matching pattern and replace."""
    for i in range(start, len(lines) if end is None else end):
        if re.search(pattern, lines[i]):
            old = lines[i]
            lines[i] = replacement
            print(f"  L{i+1} [{desc}]: patched")
            return True
    print(f"  NOT FOUND: {desc}")
    return False

# ============================================================
# Helper: build multi-line SQL for product list queries
# ============================================================
def product_list_sql(where_clause="$1", extra_order=""):
    return (
        "SELECT p.id, p.title, p.slug, p.description, p.category_id, "
        "COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, "
        "p.status, p.download_count_limit, p.max_downloads_per_user, "
        "p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at\n"
        "FROM products p\n"
        "LEFT JOIN categories c ON p.category_id = c.id\n"
        f"{where_clause}\n"
        f"ORDER BY p.created_at DESC{extra_order}\n"
        "LIMIT $2 OFFSET $3"
    )

# ============================================================
# TestProducts_GetProducts (lines 120-133)
# ============================================================
# Fix COUNT mock - handler sends: SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1
patch_line(123,
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).'
)

# Fix product list mock - use (?s) flag for multiline matching
# Handler sends multi-line query with tabs/newlines
patch_line(125,
    '\tmock.ExpectQuery(`(?s)SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).'
)

# ============================================================
# TestProducts_GetProductBySlug (lines 135-147)
# ============================================================
patch_line(138,
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).'
)

patch_line(140,
    '\tmock.ExpectQuery(`(?s)SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.slug = $1`).'
)

patch_line(144,
    '\tmock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).'
)

# ============================================================
# TestProducts_SearchProducts (lines 154-167)
# ============================================================
patch_line(157,
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE (p.title ILIKE $1 OR p.description ILIKE $1) AND p.status = $2`).'
)

patch_line(159,
    '\tmock.ExpectQuery(`(?s)SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE (p.title ILIKE $1 OR p.description ILIKE $1) AND p.status = $2 ORDER BY p.created_at DESC LIMIT $3 OFFSET $4`).'
)

patch_line(163,
    '\tmock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).'
)

# ============================================================
# TestCart_GetOrCreate (line ~472)
# ============================================================
find_and_patch(r'SELECT id FROM carts WHERE user_id = \\\$1`',
    '\tmock.ExpectQuery(`SELECT id, user_id, product_count FROM carts WHERE user_id = $1`).',
    "Cart GetOrCreate - SELECT cart")

# Also fix the INSERT INTO cart_items
find_and_patch(r'INSERT INTO cart_items.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, $3) RETURNING id`).',
    "Cart AddItem - INSERT cart_items")

# Fix cart item existence check
find_and_patch(r'SELECT EXISTS\(SELECT 1 FROM cart_items WH',
    '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM cart_items WHERE cart_id = $1 AND product_id = $2)`).',
    "Cart AddItem - EXISTS check")

# Fix cart remove
find_and_patch(r'DELETE FROM cart_items WHERE cart_id = \\\$1',
    '\tmock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).',
    "Cart RemoveItem - DELETE")

# ============================================================
# TestWishlist_Toggle (lines ~511-513)
# ============================================================
find_and_patch(r'SELECT EXISTS\(SELECT 1 FROM wishlists WH',
    '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).',
    "Wishlist Toggle - EXISTS")

find_and_patch(r'INSERT INTO wishlists.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO wishlists (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).',
    "Wishlist Toggle - INSERT")

# ============================================================
# TestWishlist_List (line ~526)
# ============================================================
find_and_patch(r'SELECT w\.product_id, p\.name, p\.slug.*FROM wishlists w.*`\)\.',
    '\tmock.ExpectQuery(`SELECT w.product_id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status FROM wishlists w JOIN products p ON w.product_id = p.id WHERE w.user_id = $1`).',
    "Wishlist List - SELECT")

# ============================================================
# TestReviews_Create (line ~541)
# ============================================================
find_and_patch(r'INSERT INTO reviews.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) RETURNING id`).',
    "Reviews Create - INSERT")

# ============================================================
# TestReviews_List (line ~553)
# ============================================================
find_and_patch(r'SELECT r\.id, r\.rating, r\.comment.*`\)\.',
    '\tmock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.name FROM reviews r JOIN users u ON r.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC`).',
    "Reviews List - SELECT")

# ============================================================
# TestRecentlyViewed_Record (line ~566)
# ============================================================
find_and_patch(r'INSERT INTO recently_viewed.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = NOW()`).',
    "RecentlyViewed Record - INSERT")

# ============================================================
# TestRecentlyViewed_List (line ~578)
# ============================================================
find_and_patch(r'SELECT p\.id, p\.name, p\.slug.*FROM products p.*`\)\.',
    '\tmock.ExpectQuery(`SELECT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM recently_viewed WHERE user_id = $1 ORDER BY viewed_at DESC LIMIT 10)`).',
    "RecentlyViewed List - SELECT", end=600)

# ============================================================
# TestComparison_Toggle (lines ~593-595)
# ============================================================
find_and_patch(r'SELECT EXISTS\(SELECT 1 FROM product_comparison WH',
    '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).',
    "Comparison Toggle - EXISTS")

find_and_patch(r'INSERT INTO product_comparison.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO product_comparison (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).',
    "Comparison Toggle - INSERT")

# ============================================================
# TestComparison_List (line ~607)
# ============================================================
find_and_patch(r'SELECT p\.id, p\.name, p\.slug.*FROM products p.*`\)\.',
    '\tmock.ExpectQuery(`SELECT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM product_comparison WHERE user_id = $1 ORDER BY added_at DESC LIMIT 4)`).',
    "Comparison List - SELECT", end=620)

# ============================================================
# TestRecommendations_Get (lines ~622-628)
# ============================================================
find_and_patch(r'SELECT category_id FROM products WHERE id = \\\$1`',
    '\tmock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).',
    "Recommendations - category_id")

find_and_patch(r'SELECT p\.id, p\.title.*FROM products p.*`\)\.',
    '\tmock.ExpectQuery(`(?s)SELECT DISTINCT p.id, p.title, p.slug, p.description, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.category_id = $1 AND p.status = $2 AND p.id != $3 ORDER BY p.created_at DESC LIMIT 4`).',
    "Recommendations - products")

# ============================================================
# TestGuestOrder_Create (line ~641)
# ============================================================
find_and_patch(r'INSERT INTO guest_orders.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO guest_orders (email, total_usd, crypto_chain, crypto_amount, crypto_address, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`).',
    "GuestOrder Create - INSERT")

# ============================================================
# TestGuestOrder_Check (line ~643)
# ============================================================
find_and_patch(r'SELECT id, email, status.*FROM guest_orders WHERE id = \\\$1`\)',
    '\tmock.ExpectQuery(`SELECT id, email, status, total_usd, crypto_chain, crypto_amount, crypto_address FROM guest_orders WHERE id = $1`).',
    "GuestOrder Check - SELECT")

# ============================================================
# TestOrderStatusCheck (line ~671)
# ============================================================
find_and_patch(r'SELECT o\.id, o\.status.*FROM orders WHERE id = \\\$1`\)',
    '\tmock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).',
    "OrderStatusCheck - SELECT")

# ============================================================
# TestNewProducts_GetActiveStat (lines ~726-730)
# ============================================================
patch_line(726,
    "\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = 'active'`)."
)

patch_line(728,
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products`).'
)

patch_line(730,
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM categories`).'
)

# Fix the COALESCE query (find it after line 730)
find_and_patch(r'SELECT COALESCE.*`\)\.[\s\S]*?AddRow\(\'100\.00\'\)',
    '\tmock.ExpectQuery(`SELECT COALESCE(SUM(total_usd), 0) FROM orders WHERE status = $1`).',
    "NewProducts - COALESCE", start=730)

# Fix the products list for NewProducts_GetActiveStat
find_and_patch(r'SELECT p\.id, p\.name, p\.slug.*`\)\.[\s\S]*?AddRow\(1, "Test Product"',
    '\tmock.ExpectQuery(`(?s)SELECT p.id, p.name, p.slug FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).',
    "NewProducts - products list", start=730)

# ============================================================
# TestNewProducts_Delete (lines ~692-698)
# ============================================================
find_and_patch(r'SELECT COUNT\(\\\\\\*\) FROM products WHERE slug = \\\$1`\)',
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).',
    "NewProducts Delete - COUNT")

find_and_patch(r'DELETE FROM product_images WHERE product_id = \\\$1`\)',
    '\tmock.ExpectExec(`DELETE FROM product_images WHERE product_id = $1`).',
    "NewProducts Delete - DELETE images")

find_and_patch(r'DELETE FROM products WHERE id = \\\$1`\)',
    '\tmock.ExpectExec(`DELETE FROM products WHERE id = $1`).',
    "NewProducts Delete - DELETE products")

# ============================================================
# TestExchangeRates_GetAll (line ~404)
# ============================================================
find_and_patch(r'SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`\)\.',
    '\tmock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).',
    "ExchangeRates GetAll")

# ============================================================
# TestExchangeRates_GetOne (line ~418 or similar)
# ============================================================
find_and_patch(r'SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = \\\$1`\)\.',
    '\tmock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1`).',
    "ExchangeRates GetOne")

# ============================================================
# TestAdmin_Stats (lines ~756-769)
# ============================================================
find_and_patch(r'SELECT COUNT\(\\\\\\*\) FROM users`\)\.[\s\S]*?AddRow\(25\)',
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM users`).',
    "Admin Stats - users count")

find_and_patch(r'SELECT COUNT\(\\\\\\*\) FROM orders`\)\.[\s\S]*?AddRow\(\d+\)',
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM orders`).',
    "Admin Stats - orders count")

find_and_patch(r'SELECT COUNT\(\\\\\\*\) FROM categories`\)\.[\s\S]*?AddRow\(\d+\)',
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM categories`).',
    "Admin Stats - categories count")

find_and_patch(r'SELECT COALESCE.*`\)\.[\s\S]*?AddRow\(\'100\.00\'\)',
    '\tmock.ExpectQuery(`SELECT COALESCE(SUM(total_usd), 0) FROM orders WHERE status = $1`).',
    "Admin Stats - revenue")

# ============================================================
# Admin Order Status transitions (lines ~246-283)
# ============================================================
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`UPDATE orders SET' in line:
        lines[i] = '\t\tmock.ExpectExec(`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`).'
        print(f"  L{i+1}: Fixed UPDATE orders SET")

# ============================================================
# Admin GuestOrderCheck (lines ~389-401)
# ============================================================
find_and_patch(r'SELECT o\.id, o\.status.*FROM orders WHERE id = \\\$1.*`\)\.[\s\S]*?AddRow\(1, "completed".*`\)\.',
    '\tmock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).',
    "Admin GuestOrderCheck - SELECT")

# Also the second SELECT for guest order items
find_and_patch(r'SELECT id, product_id, product_title.*FROM order_items WHERE order_id=\$1',
    '\tmock.ExpectQuery(`SELECT id, product_id, product_title, product_slug, quantity, CAST(unit_price_usd AS NUMERIC), download_count FROM order_items WHERE order_id = $1 ORDER BY id`).',
    "Admin GuestOrderCheck - items")

# ============================================================
# Admin GetSetting (line ~712)
# ============================================================
find_and_patch(r'SELECT key, value, updated_at FROM settings ORDER BY key`\)\.',
    '\tmock.ExpectQuery(`SELECT setting_key, value, updated_at FROM settings ORDER BY setting_key`).',
    "Admin GetSettings - list")

find_and_patch(r'SELECT value FROM settings WHERE key = \\\$1`\)\.',
    '\tmock.ExpectQuery(`SELECT setting_key, value, updated_at FROM settings WHERE setting_key = $1`).',
    "Admin GetSetting - one")

# ============================================================
# Admin ExchangeRateSet + ExchangeRateList (lines ~757-769)
# ============================================================
find_and_patch(r'INSERT INTO exchange_rates.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO exchange_rates (chain, rate) VALUES ($1, $2) ON CONFLICT (chain) DO UPDATE SET rate = $2, updated_at = NOW()`).',
    "Admin ExchangeRateSet - INSERT")

# ============================================================
# Admin SetSetting (line ~430)
# ============================================================
find_and_patch(r'INSERT INTO settings.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO settings (setting_key, value) VALUES ($1, $2) ON CONFLICT (setting_key) DO UPDATE SET value = $2`).',
    "Admin SetSetting - INSERT")

# ============================================================
# Admin DeleteUser (line ~320)
# ============================================================
find_and_patch(r'DELETE FROM users WHERE id = \\\$1`\)\.',
    '\tmock.ExpectExec(`DELETE FROM users WHERE id = $1`).',
    "Admin DeleteUser")

# ============================================================
# Admin DeleteCommunityPost (line ~346)
# ============================================================
find_and_patch(r'DELETE FROM community_posts WHERE id = \\\$1`\)\.',
    '\tmock.ExpectExec(`DELETE FROM community_posts WHERE id = $1`).',
    "Admin DeleteCommunityPost")

# ============================================================
# Admin CommunityPostList (line ~358)
# ============================================================
find_and_patch(r'SELECT cp\.id, cp\.user_id.*FROM community_posts cp.*`\)\.',
    '\tmock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, COALESCE(cl.cnt, 0) + COALESCE(cc.cnt, 0) FROM community_posts cp LEFT JOIN users u ON cp.user_id = u.id LEFT JOIN (SELECT post_id, COUNT(*) as cnt FROM likes WHERE type = \'post\' GROUP BY post_id) cl ON cp.id = cl.post_id LEFT JOIN (SELECT post_id, COUNT(*) as cnt FROM comments WHERE type = \'post\' GROUP BY post_id) cc ON cp.id = cc.post_id WHERE cp.type = \'post\' ORDER BY cp.created_at DESC LIMIT $3 OFFSET $4`).',
    "Admin CommunityPostList", end=380)

# ============================================================
# Admin CommunityUserList (lines ~372-374)
# ============================================================
find_and_patch(r'SELECT COUNT\(\\\\\\*\) FROM users`\)\.[\s\S]*?AddRow\(25\)',
    '\tmock.ExpectQuery(`SELECT COUNT(*) FROM users`).',
    "Admin CommunityUserList - count")

find_and_patch(r'SELECT u\.id, u\.email, u\.name.*`\)\.[\s\S]*?AddRow\(1, "test@test.com".*`\)\.',
    '\tmock.ExpectQuery(`SELECT u.id, u.email, u.name, COUNT(cp.id) FROM users u LEFT JOIN community_posts cp ON u.id = cp.user_id WHERE cp.type = \'post\' GROUP BY u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`).',
    "Admin CommunityUserList - users", end=390)

# ============================================================
# Admin Bulk Operations (lines ~233-308)
# ============================================================
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`UPDATE products SET' in line and i > 200:
        lines[i] = '\t\tmock.ExpectExec(`UPDATE products SET status = $1, category_id = $2, updated_at = NOW() WHERE id = ANY($3)`).'
        print(f"  L{i+1}: Fixed bulk product UPDATE")

# ============================================================
# DB Schema Migration011 (line ~782)
# ============================================================
find_and_patch(r'SELECT table_name FROM information_schema\.tables WHERE table_schema = \'public\' AND table_name = \\\$1`\)\.',
    '\tmock.ExpectQuery(`SELECT table_name FROM information_schema.tables WHERE table_schema = \'public\' AND table_name = $1`).',
    "DB Schema Migration011")

# ============================================================
# Products Create (line ~196)
# ============================================================
find_and_patch(r'INSERT INTO products.*`\)\.',
    '\t\tmock.ExpectExec(`INSERT INTO products (name, slug, description, category_id, price_usd, asset_path, asset_hash, status, download_count_limit, max_downloads_per_user, file_size_bytes, file_mime_type, user_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW()) RETURNING id`).',
    "Products Create", start=190, end=210)

# ============================================================
# Auth UpdateProfile fixes
# ============================================================
# Fix the UPDATE users SET mock
find_and_patch(r'mock\.ExpectExec\(`UPDATE users SET.*`\)\.',
    '\t\tmock.ExpectExec(`UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).',
    "Auth UpdateProfile - UPDATE users")

# Fix the SELECT for bio/wallet_address - should match exactly
find_and_patch(r'SELECT bio, wallet_address FROM user_profiles WHERE user_id = \\\$1`\)\.',
    '\tmock.ExpectQuery(`SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1`).',
    "Auth UpdateProfile - SELECT bio/wallet")

# Fix the second SELECT id,email,name after UPDATE (should also have WithArgs)
# This is already correct in most cases

# ============================================================
# Write result
# ============================================================
result = '\n'.join(lines)
with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(result)

print(f"\nDone. Written {len(result)} chars. Running build+test...")

import subprocess
r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True, cwd="/root/project/backend")
print(f"Build: exit={r.returncode}")
if r.returncode != 0:
    print("STDERR:", r.stderr[-500:])

r = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True, cwd="/root/project/backend")
print(r.stdout[-1500:])
if r.returncode != 0:
    print("STDERR:", r.stderr[-500:])

# Count passes/fails
passes = r.stdout.count('--- PASS:')
fails = r.stdout.count('--- FAIL:')
print(f"\nResult: {passes} PASS, {fails} FAIL, exit={r.returncode}")
