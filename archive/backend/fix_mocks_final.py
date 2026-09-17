#!/usr/bin/env python3
"""Fix all mock SQL mismatches. Read handler SQL, patch test mocks to match exactly."""
import re, subprocess, sys

with open("/root/project/backend/handlers/all_features_test.go") as f:
    content = f.read()

lines = content.split('\n')

def replace_line(pattern, replacement, start=0, end=None):
    """Replace first matching line. Returns True if replaced."""
    for i in range(start, len(lines) if end is None else end):
        if re.search(pattern, lines[i]):
            old = lines[i]
            lines[i] = replacement
            print(f"  L{i+1}: {old[:70]}...")
            return True
    return False

# === AUTH ===
# SELECT id, email, password_hash, name... FROM users WHERE email = $1 (login queries)
replace_line(r'SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email = \\\$1',
             'mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email = $1`).')
replace_line(r'SELECT id, email, name, created_at, updated_at FROM users WHERE id = \\\$1',
             'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).')

# UPDATE users SET for profile update
replace_line(r'mock\.ExpectExec\(`UPDATE users SET.*`\)\.',
             'mock.ExpectExec(`UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).')

# SELECT bio, wallet_address FROM user_profiles
replace_line(r'SELECT bio, wallet_address FROM user_profiles WHERE user_id',
             'mock.ExpectQuery(`SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1`).')

# === PRODUCTS ===
# Products list: COUNT query
replace_line(r'SELECT COUNT\(\\\\\\*\) FROM products WHERE p\.status',
             'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).')
# Also fix COUNT without WHERE
replace_line(r'SELECT COUNT\(\\\\\\*\) FROM products.*`\)\.[\s\S]*?WillReturnRows\(sqlmock\.NewRows\(\[\]string{"count"}\)\.AddRow\(2\)\)[ \t]*\n[\s\S]*?mock\.ExpectQuery\(`SELECT p\.id, p\.title\.\*`\)',
             'PLACEHOLDER_NEEDS_LINE_FIX')

# Products list: full query (this is the main one that fails)
replace_line(r'mock\.ExpectQuery\(`SELECT p\.id, p\.title\..*`\)\.[\s\S]*AddRow\(1, "Test Product".*AddRow\(2, "Another Product"',
             'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).')

# Products by slug: COUNT
replace_line(r'SELECT COUNT\(\\\\\\*\) FROM products WHERE slug = \\\$1',
             'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).')
replace_line(r'SELECT p\.id, p\.title\..*slug = \\\$1',
             'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.slug = $1`).')

# Products search: COUNT  
replace_line(r'SELECT COUNT\(\\\\\\*\) FROM products.*`\)\.[\s\S]*?WillReturnRows\(sqlmock\.NewRows\(\[\]string{"count"}\)\.AddRow\(1\)\)[ \t]*\n\tmock\.ExpectQuery\(`SELECT p\.id, p\.title\.\*',
             'PLACEHOLDER_SEARCH')

# Products search: full query
replace_line(r'SELECT p\.id, p\.title\..*%test%.*ORDER BY',
             'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE (p.title ILIKE $1 OR p.description ILIKE $1) AND p.status = $2 ORDER BY p.created_at DESC LIMIT $3 OFFSET $4`).')

# Create product
replace_line(r'INSERT INTO products.*`\)\.',
             'mock.ExpectExec(`INSERT INTO products (name, slug, description, category_id, price_usd, asset_path, asset_hash, status, download_count_limit, max_downloads_per_user, file_size_bytes, file_mime_type, user_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`).')

# === CART ===
replace_line(r'SELECT id FROM carts WHERE user_id = \\\$1',
             'mock.ExpectQuery(`SELECT id, user_id FROM carts WHERE user_id = $1`).')
replace_line(r'INSERT INTO cart_items.*`\)\.',
             'mock.ExpectExec(`INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, $3) RETURNING id`).')
replace_line(r'SELECT EXISTS\(SELECT 1 FROM cart_items WH',
             'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM cart_items WHERE cart_id = $1 AND product_id = $2)`).')
replace_line(r'DELETE FROM cart_items WHERE cart_id = \\\$1',
             'mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).')

# === WISHLIST ===
replace_line(r'INSERT INTO wishlists.*`\)\.',
             'mock.ExpectExec(`INSERT INTO wishlists (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).')
replace_line(r'SELECT EXISTS\(SELECT 1 FROM wishlists WH',
             'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).')
replace_line(r'SELECT w\.product_id, p\.name, p\.slug.*`\)\.',
             'mock.ExpectQuery(`SELECT w.product_id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status FROM wishlists w JOIN products p ON w.product_id = p.id WHERE w.user_id = $1`).')

# === REVIEWS ===
replace_line(r'INSERT INTO reviews.*`\)\.',
             'mock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) RETURNING id`).')
replace_line(r'SELECT r\.id, r\.rating, r\.comment.*`\)\.',
             'mock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.name FROM reviews r JOIN users u ON r.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC`).')

# === RECENTLY VIEWED ===
replace_line(r'INSERT INTO recently_viewed.*`\)\.',
             'mock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = NOW()`).')
replace_line(r'SELECT p\.id, p\.name, p\.slug.*`\)\.',
             'mock.ExpectQuery(`SELECT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM recently_viewed WHERE user_id = $1 ORDER BY viewed_at DESC LIMIT 10)`).')

# === COMPARISON ===
replace_line(r'SELECT EXISTS\(SELECT 1 FROM product_comparison WH',
             'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).')
replace_line(r'INSERT INTO product_comparison.*`\)\.',
             'mock.ExpectExec(`INSERT INTO product_comparison (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).')

# === RECOMMENDATIONS ===
replace_line(r'SELECT category_id FROM products WHERE id = \\\$1',
             'mock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).')
replace_line(r'SELECT p\.id, p\.title.*FROM products p.*`\)\.',
             'mock.ExpectQuery(`SELECT DISTINCT p.id, p.title, p.slug, p.description, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.category_id = $1 AND p.status = $2 AND p.id != $3 ORDER BY p.created_at DESC LIMIT 4`).')

# === GUEST ORDERS ===
replace_line(r'INSERT INTO guest_orders.*`\)\.',
             'mock.ExpectExec(`INSERT INTO guest_orders (email, total_usd, crypto_chain, crypto_amount, crypto_address, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`).')
replace_line(r'SELECT id, email, status.*FROM guest_orders WHERE id = \\\$1',
             'mock.ExpectQuery(`SELECT id, email, status, total_usd FROM guest_orders WHERE id = $1`).')

# === ORDER STATUS CHECK ===
replace_line(r'SELECT o\.id, o\.status.*FROM orders WHERE id = \\\$1',
             'mock.ExpectQuery(`SELECT o.id, o.status, o.total_usd FROM orders WHERE o.id = $1`).')

# === NEW PRODUCTS ===
replace_line(r"SELECT COUNT\(\\\\\\*\) FROM products WHERE status = 'active'\)",
             "mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = 'active'`).")
replace_line(r'SELECT COUNT\(\\\\\\*\) FROM products`\)\.[\s\S]*?WillReturnRows\(sqlmock\.NewRows\(\[\]string{"count"}\)\.AddRow\(\d+\)\)[ \t]*\n\tmock\.ExpectQuery\(`SELECT .*FROM products',
             'PLACEHOLDER_NEWPPRODUCTS')
replace_line(r'SELECT COALESCE.*`\)\.',
             'mock.ExpectQuery(`SELECT COALESCE(SUM(total_usd), 0) FROM orders WHERE status = $1`).')
replace_line(r'SELECT p\.id, p\.name, p\.slug.*`\)\.',
             'mock.ExpectQuery(`SELECT p.id, p.name, p.slug FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).')
replace_line(r'DELETE FROM product_images WHERE product_id = \\\$1',
             'mock.ExpectExec(`DELETE FROM product_images WHERE product_id = $1`).')
replace_line(r'DELETE FROM products WHERE id = \\\$1',
             'mock.ExpectExec(`DELETE FROM products WHERE id = $1`).')

# === EXCHANGE RATES ===
replace_line(r'SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`\)\.',
             'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).')
replace_line(r'SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = \\\$1',
             'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1`).')
replace_line(r'INSERT INTO exchange_rates.*`\)\.',
             'mock.ExpectExec(`INSERT INTO exchange_rates (chain, rate) VALUES ($1, $2) ON CONFLICT (chain) DO UPDATE SET rate = $2, updated_at = NOW()`.')')

# === ADMIN SETTINGS ===
replace_line(r'SELECT value FROM settings WHERE key = \\\$1',
             'mock.ExpectQuery(`SELECT setting_key, value, updated_at FROM settings WHERE setting_key = $1`).')
replace_line(r'INSERT INTO settings.*`\)\.',
             'mock.ExpectExec(`INSERT INTO settings (setting_key, value) VALUES ($1, $2) ON CONFLICT (setting_key) DO UPDATE SET value = $2`).')

# === ADMIN STATS ===
replace_line(r'SELECT COUNT\(\\\\\\*\) FROM users`\)\.[\s\S]*?AddRow\(25\)',
             'mock.ExpectQuery(`SELECT COUNT(*) FROM users`).')
replace_line(r'SELECT COUNT\(\\\\\\*\) FROM orders`\)\.[\s\S]*?AddRow\(\d+\)',
             'mock.ExpectQuery(`SELECT COUNT(*) FROM orders`).')
replace_line(r'SELECT COUNT\(\\\\\\*\) FROM categories`\)\.[\s\S]*?AddRow\(\d+\)',
             'mock.ExpectQuery(`SELECT COUNT(*) FROM categories`).')
replace_line(r'SELECT COALESCE.*`\)\.[\s\S]*?AddRow\(\'100\.00\'\)',
             'mock.ExpectQuery(`SELECT COALESCE(SUM(total_usd), 0) FROM orders WHERE status = $1`).')

# === ADMIN ORDER STATUS ===
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`UPDATE orders SET' in line:
        lines[i] = '\t\tmock.ExpectExec(`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`).'
        print(f"  L{i+1}: Fixed UPDATE orders")

# === ADMIN GUEST ORDER CHECK ===
replace_line(r'SELECT o\.id, o\.status, o\.total_usd.*FROM orders WHERE id = \\\$1',
             'mock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).')

# === ADMIN COMMUNITY POSTS ===
replace_line(r'SELECT cp\.id, cp\.user_id.*`\)\.',
             'mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, COALESCE(cl.count, 0), COALESCE(cc.count, 0), EXISTS(SELECT 1 FROM follows WHERE follower_id = $2 AND following_id = cp.user_id) FROM community_posts cp LEFT JOIN users u ON cp.user_id = u.id LEFT JOIN (SELECT post_id, COUNT(*) as count FROM likes WHERE type = \'\'post\'\' GROUP BY post_id) cl ON cp.id = cl.post_id LEFT JOIN (SELECT post_id, COUNT(*) as count FROM comments WHERE type = \'\'post\'\' GROUP BY post_id) cc ON cp.id = cc.post_id WHERE cp.type = \'\'post\'\' ORDER BY cp.created_at DESC LIMIT $3 OFFSET $4`).')

# === ADMIN COMMUNITY USERS ===
replace_line(r'SELECT u\.id, u\.email, u\.name.*`\)\.',
             'mock.ExpectQuery(`SELECT u.id, u.email, u.name, COUNT(cp.id) FROM users u LEFT JOIN community_posts cp ON u.id = cp.user_id WHERE cp.type = \'\'post\'\' GROUP BY u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`).')

# === ADMIN BULK OPERATIONS ===
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`UPDATE products SET' in line and i > 200:
        lines[i] = '\t\tmock.ExpectExec(`UPDATE products SET status = $1, category_id = $2, updated_at = NOW() WHERE id = ANY($3)`).'
        print(f"  L{i+1}: Fixed UPDATE products bulk")

# === DB SCHEMA ===
replace_line(r'SELECT table_name FROM information_schema\.tables WHERE table_schema = \'public\' AND table_name = \\\$1',
             'mock.ExpectQuery(`SELECT table_name FROM information_schema.tables WHERE table_schema = \'public\' AND table_name = $1`).')

# Write result
result = '\n'.join(lines)
with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(result)

print(f"\nApplied fixes. Testing...")
result = subprocess.run(["go", "test", "./handlers/", "-count=1", "-v"], 
                       capture_output=True, text=True, cwd="/root/project/backend")
# Extract PASS/FAIL summary
for line in result.stdout.split('\n'):
    if 'PASS:' in line or 'FAIL:' in line:
        print(line.strip())
print(f"\nResult: exit={result.returncode}")
