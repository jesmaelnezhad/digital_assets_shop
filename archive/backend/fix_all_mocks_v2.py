#!/usr/bin/env python3
"""
Fix all mock SQL mismatches. Read handler source, patch test file.
"""
import re, subprocess

with open("/root/project/backend/handlers/all_features_test.go") as f:
    content = f.read()

lines = content.split('\n')

def replace_matching(pattern, replacement, start=0, end=None):
    """Replace first line matching regex pattern."""
    for i in range(start, len(lines) if end is None else end):
        if re.search(pattern, lines[i]):
            old = lines[i]
            lines[i] = replacement
            print(f"  L{i+1}: FIXED")
            return True
    return False

def replace_line(num, replacement):
    """Replace exact line number (1-indexed)."""
    idx = num - 1
    if idx < len(lines):
        old = lines[idx]
        lines[idx] = replacement
        print(f"  L{num}: {old[:80]}... -> {replacement[:80]}...")
        return True
    return False

# ===== PRODUCTS_GETPRODUCTS (lines 123, 125 in 1-indexed) =====
# Handler sends: SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1
replace_line(123, '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).')
# Handler sends full query with all columns, JOIN, WHERE, ORDER BY, LIMIT, OFFSET
replace_line(125, '\tmock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).')

# ===== PRODUCTS_GETPRODUCTBYSLUG (lines 138, 140) =====
replace_line(138, '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).')
replace_line(140, '\tmock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.slug = $1`).')

# ===== PRODUCTS_SEARCHPRODUCTS (lines 157, 159) =====
replace_line(157, '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE (p.title ILIKE $1 OR p.description ILIKE $1) AND p.status = $2`).')
replace_line(159, '\tmock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE (p.title ILIKE $1 OR p.description ILIKE $1) AND p.status = $2 ORDER BY p.created_at DESC LIMIT $3 OFFSET $4`).')

# ===== CART_GETORCREATE (lines ~472) =====
replace_line(472, '\tmock.ExpectQuery(`SELECT id, user_id FROM carts WHERE user_id = $1`).')

# ===== CART_ADDITEM (lines ~485-487) =====
replace_line(487, '\tmock.ExpectExec(`INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, $3) RETURNING id`).')

# ===== CART_REMOVEITEM (lines ~499) =====
replace_line(499, '\tmock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).')

# ===== WISHLIST_TOGGLE (lines ~511-513) =====
replace_line(511, '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`.')
replace_line(513, '\tmock.ExpectExec(`INSERT INTO wishlists (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).')

# ===== WISHLIST_LIST (line ~526) =====
replace_line(526, '\tmock.ExpectQuery(`SELECT w.product_id, p.name, p.slug, p.price_usd FROM wishlists w JOIN products p ON w.product_id = p.id WHERE w.user_id = $1`).')

# ===== REVIEWS_CREATE (line ~541) =====
replace_line(541, '\tmock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) RETURNING id`).')

# ===== REVIEWS_LIST (line ~553) =====
replace_line(553, '\tmock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.name FROM reviews r JOIN users u ON r.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC`).')

# ===== RECENTLYVIEWED_RECORD (line ~566) =====
replace_line(566, '\tmock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = NOW()`).')

# ===== RECENTLYVIEWED_LIST (line ~578) =====
replace_line(578, '\tmock.ExpectQuery(`SELECT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM recently_viewed WHERE user_id = $1 ORDER BY viewed_at DESC LIMIT 10)`).')

# ===== COMPARISON_TOGGLE (lines ~593-595) =====
replace_line(593, '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`.')
replace_line(595, '\tmock.ExpectExec(`INSERT INTO product_comparison (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).')

# ===== COMPARISON_LIST (line ~607) =====
replace_line(607, '\tmock.ExpectQuery(`SELECT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM product_comparison WHERE user_id = $1 ORDER BY added_at DESC LIMIT 4)`).')

# ===== RECOMMENDATIONS_GET (lines ~622-628) =====
replace_line(622, '\tmock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).')
replace_line(624, '\tmock.ExpectQuery(`SELECT DISTINCT p.id, p.title, p.slug, p.description, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.category_id = $1 AND p.status = $2 AND p.id != $3 ORDER BY p.created_at DESC LIMIT 4`).')

# ===== GUESTORDER_CREATE (line ~641) =====
replace_line(641, '\tmock.ExpectExec(`INSERT INTO guest_orders (email, total_usd, crypto_chain, crypto_amount, crypto_address, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`).')

# ===== GUESTORDER_CHECK (line ~643) =====
replace_line(643, '\tmock.ExpectQuery(`SELECT id, email, status, total_usd, crypto_chain, crypto_amount, crypto_address FROM guest_orders WHERE id = $1`).')

# ===== ORDERSTATUSCHECK (line ~671) =====
replace_line(671, '\tmock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).')

# ===== NEWPRODUCTS_GETACTIVESTAT (lines ~726-730) =====
replace_line(726, "\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = 'active'`).")
replace_line(728, '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products`).')
replace_line(730, '\tmock.ExpectQuery(`SELECT COUNT(*) FROM categories`).')

# Need to also fix the COALESCE query for admin stats
for i, line in enumerate(lines):
    if i > 700 and 'COALESCE' in line and 'AddRow' in line:
        lines[i] = '\t\tmock.ExpectQuery(`SELECT COALESCE(SUM(total_usd), 0) FROM orders WHERE status = $1`).'
        print(f"  L{i+1}: Fixed COALESCE query")

# Also fix the products list query for NewProducts_GetActiveStat
for i, line in enumerate(lines):
    if i > 700 and 'SELECT p.id, p.name, p.slug' in line and 'LEFT JOIN' not in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT p.id, p.name, p.slug FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).'
        print(f"  L{i+1}: Fixed NewProducts products list query")

# ===== EXCHANGE RATES =====
for i, line in enumerate(lines):
    if i > 400 and 'SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).'
        print(f"  L{i+1}: Fixed exchange rates list query")

for i, line in enumerate(lines):
    if i > 400 and 'SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1`).'
        print(f"  L{i+1}: Fixed exchange rate get one query")

for i, line in enumerate(lines):
    if i > 400 and 'INSERT INTO exchange_rates' in line and 'mock.ExpectExec' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO exchange_rates (chain, rate) VALUES ($1, $2) ON CONFLICT (chain) DO UPDATE SET rate = $2, updated_at = NOW()`.').replace('`.)', '`).')
        print(f"  L{i+1}: Fixed exchange rate insert")

# ===== ADMIN SETTINGS =====
for i, line in enumerate(lines):
    if i > 400 and 'SELECT value FROM settings WHERE key = $1' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT setting_key, value, updated_at FROM settings WHERE setting_key = $1`).'
        print(f"  L{i+1}: Fixed settings get query")

for i, line in enumerate(lines):
    if i > 400 and 'INSERT INTO settings' in line and 'mock.ExpectExec' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO settings (setting_key, value) VALUES ($1, $2) ON CONFLICT (setting_key) DO UPDATE SET value = $2`.').replace('`.)', '`).')
        print(f"  L{i+1}: Fixed settings insert")

# ===== ADMIN STATS =====
for i, line in enumerate(lines):
    if i > 700 and 'SELECT COUNT(*) FROM users' in line and 'AddRow' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM users`).'
        print(f"  L{i+1}: Fixed admin stats users count")
        break

for i, line in enumerate(lines):
    if i > 700 and 'SELECT COUNT(*) FROM orders' in line and 'AddRow' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM orders`).'
        print(f"  L{i+1}: Fixed admin stats orders count")
        break

for i, line in enumerate(lines):
    if i > 700 and 'SELECT COUNT(*) FROM categories' in line and 'AddRow' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM categories`).'
        print(f"  L{i+1}: Fixed admin stats categories count")
        break

# ===== ADMIN ORDER STATUS =====
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`UPDATE orders SET' in line:
        lines[i] = '\t\tmock.ExpectExec(`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`).'
        print(f"  L{i+1}: Fixed UPDATE orders SET")

# ===== ADMIN GUEST ORDER CHECK =====
for i, line in enumerate(lines):
    if i > 300 and 'SELECT o.id, o.status' in line and 'FROM orders WHERE id' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).'
        print(f"  L{i+1}: Fixed admin guest order check query")
        break

# ===== ADMIN COMMUNITY POSTS =====
for i, line in enumerate(lines):
    if i > 300 and 'SELECT cp.id, cp.user_id' in line and 'FROM community_posts cp' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, COALESCE(cl.cnt, 0), COALESCE(cc.cnt, 0) FROM community_posts cp LEFT JOIN users u ON cp.user_id = u.id LEFT JOIN (SELECT post_id, COUNT(*) as cnt FROM likes WHERE type = \'post\' GROUP BY post_id) cl ON cp.id = cl.post_id LEFT JOIN (SELECT post_id, COUNT(*) as cnt FROM comments WHERE type = \'post\' GROUP BY post_id) cc ON cp.id = cc.post_id WHERE cp.type = \'post\' ORDER BY cp.created_at DESC LIMIT $3 OFFSET $4`).'
        print(f"  L{i+1}: Fixed admin community post list query")
        break

# ===== ADMIN COMMUNITY USERS =====
for i, line in enumerate(lines):
    if i > 300 and 'SELECT u.id, u.email, u.name' in line and 'count' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT u.id, u.email, u.name, COUNT(cp.id) FROM users u LEFT JOIN community_posts cp ON u.id = cp.user_id WHERE cp.type = \'post\' GROUP BY u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`).'
        print(f"  L{i+1}: Fixed admin community user list query")
        break

# ===== ADMIN BULK OPERATIONS =====
for i, line in enumerate(lines):
    if i > 200 and 'mock.ExpectExec(`UPDATE products SET' in line:
        lines[i] = '\t\tmock.ExpectExec(`UPDATE products SET status = $1, category_id = $2, updated_at = NOW() WHERE id = ANY($3)`).'
        print(f"  L{i+1}: Fixed bulk product update")

# ===== ADMIN DELETE USER =====
for i, line in enumerate(lines):
    if i > 300 and 'DELETE FROM users WHERE id = $1' in line and 'mock.ExpectExec' in line:
        lines[i] = '\t\tmock.ExpectExec(`DELETE FROM users WHERE id = $1`).'
        print(f"  L{i+1}: Fixed delete user")

# ===== ADMIN DELETE COMMUNITY POST =====
for i, line in enumerate(lines):
    if i > 300 and 'DELETE FROM community_posts WHERE id = $1' in line and 'mock.ExpectExec' in line:
        lines[i] = '\t\tmock.ExpectExec(`DELETE FROM community_posts WHERE id = $1`).'
        print(f"  L{i+1}: Fixed delete community post")

# ===== DB SCHEMA MIGRATION011 =====
for i, line in enumerate(lines):
    if i > 700 and 'information_schema.tables' in line:
        lines[i] = '\t\tmock.ExpectQuery(`SELECT table_name FROM information_schema.tables WHERE table_schema = \'public\' AND table_name = $1`).'
        print(f"  L{i+1}: Fixed schema migration query")
        break

# ===== PRODUCTS_CREATE =====
for i, line in enumerate(lines):
    if i > 130 and 'INSERT INTO products' in line and 'mock.ExpectExec' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO products (name, slug, description, category_id, price_usd, asset_path, asset_hash, status, download_count_limit, max_downloads_per_user, file_size_bytes, file_mime_type, user_id, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW()) RETURNING id`).'
        print(f"  L{i+1}: Fixed create product query")
        break

# ===== AUTH UPDATE PROFILE =====
for i, line in enumerate(lines):
    if i > 70 and 'mock.ExpectExec(`UPDATE users SET' in line:
        lines[i] = '\t\tmock.ExpectExec(`UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).'
        print(f"  L{i+1}: Fixed update users profile query")
        break

for i, line in enumerate(lines):
    if i > 70 and 'SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`.\n\t\tWithArgs(1).\n\t\tWillReturnRows' in line:
        # This is the second SELECT in UpdateProfile (after the UPDATE)
        pass  # Keep as-is, it's correct

# Write result
result = '\n'.join(lines)
with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(result)

print(f"\nApplied fixes. Running build+test...")
r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True, cwd="/root/project/backend")
print(f"Build: exit={r.returncode}")
if r.returncode != 0:
    print(r.stderr[-500:])

r = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True, cwd="/root/project/backend")
print(r.stdout[-1000:])
if r.returncode != 0:
    print("STDERR:", r.stderr[-500:])
