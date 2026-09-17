#!/usr/bin/env python3
"""
Direct fix: replace specific mock SQL lines with correct versions matching handlers.
"""
import re

with open("handlers/all_features_test.go") as f:
    content = f.read()

lines = content.split('\n')

# Fix 1: Products_GetProducts - COUNT query (line 123, 0-indexed 122)
# Handler sends: SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1
# Mock needs: exact match + WithArgs("active")
old_123 = lines[122]
lines[122] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).'
old_125 = lines[124]
lines[124] = '\tmock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).'

# Fix 2: Products_GetProductBySlug - line 138
lines[137] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).'

# Fix 3: Products_SearchProducts - line 157
lines[156] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE (p.title ILIKE $1 OR p.description ILIKE $1) AND p.status = $2`).'

# Fix 4: Cart_GetOrCreate - line ~472
# Find and fix cart queries
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`)' in line:
        lines[i] = line.replace('SELECT id FROM carts WHERE user_id = $1', 
                                'SELECT id, user_id, product_count FROM carts WHERE user_id = $1')
    if 'mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`)' in line:
        lines[i] = line.replace('SELECT id FROM carts WHERE user_id = $1',
                                'SELECT id, user_id, product_count FROM carts WHERE user_id = $1')

# Fix 5: TestAuth_UpdateProfile_Valid - UPDATE users query
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`UPDATE users SET.*`)' in line and i > 70 and i < 90:
        lines[i] = '\t\tmock.ExpectExec(`UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).'

# Fix 6: Admin order status - UPDATE orders queries
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`UPDATE orders SET.*`)' in line:
        lines[i] = '\t\tmock.ExpectExec(`UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).'

# Fix 7: Admin settings - INSERT INTO settings
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`INSERT INTO settings.*`)' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO settings (setting_key, value, updated_at) VALUES ($1, $2, NOW()) ON CONFLICT (setting_key) DO UPDATE SET value = $2, updated_at = NOW()`).'

# Fix 8: Exchange rates
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`INSERT INTO exchange_rates.*`)' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO exchange_rates (chain, rate, updated_at) VALUES ($1, $2, NOW()) ON CONFLICT (chain) DO UPDATE SET rate = $2, updated_at = NOW()`).'
    if 'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).'
    if 'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1`).'

# Fix 9: Guest orders
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`INSERT INTO guest_orders.*`)' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO guest_orders (email, total_usd, crypto_chain, crypto_amount, crypto_address, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, NOW()) RETURNING id`).'
    if 'mock.ExpectQuery(`SELECT id, email, status.*FROM guest_orders WHERE id = $1`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT id, email, status, total_usd, crypto_chain, crypto_amount, crypto_address, created_at FROM guest_orders WHERE id = $1`).'

# Fix 10: Wishlist
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`INSERT INTO wishlists.*`)' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO wishlists (user_id, product_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT (user_id, product_id) DO NOTHING`).'
    if 'mock.ExpectQuery(`SELECT EXISTS\\(SELECT 1 FROM wishlists WH' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).'

# Fix 11: Reviews  
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`INSERT INTO reviews.*`)' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment, created_at) VALUES ($1, $2, $3, $4, NOW()) RETURNING id`).'
    if 'mock.ExpectQuery(`SELECT r.id, r.rating, r.comment.*`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.name FROM reviews r LEFT JOIN users u ON r.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC`).'

# Fix 12: Recently viewed
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`INSERT INTO recently_viewed.*`)' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT (user_id, product_id) DO UPDATE SET created_at = NOW()`).'

# Fix 13: Comparison
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT EXISTS\\(SELECT 1 FROM product_comp' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).'
    if 'mock.ExpectExec(`INSERT INTO product_comparison.*`)' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO product_comparison (user_id, product_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT (user_id, product_id) DO NOTHING`).'

# Fix 14: Recommendations
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).'
    if 'mock.ExpectQuery(`SELECT p.id, p.title.*FROM products p.*`)' in line and 'LEFT JOIN' not in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT DISTINCT p.id, p.title, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.category_id = $1 AND p.status = $2 ORDER BY p.created_at DESC LIMIT $3 OFFSET $4`).'
    if 'mock.ExpectQuery(`SELECT p.id, p.name.*FROM products p.*`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT DISTINCT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.category_id = $1 AND p.status = $2 ORDER BY p.created_at DESC LIMIT $3 OFFSET $4`).'

# Fix 15: Admin stats
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT COUNT\\(\\*\\) FROM users`)' in line and i > 700:
        lines[i] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM users`).'
    if 'mock.ExpectQuery(`SELECT COUNT\\(\\*\\) FROM orders`)' in line and i > 700:
        lines[i] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM orders`).'
    if 'mock.ExpectQuery(`SELECT COUNT\\(\\*\\) FROM categories`)' in line and i > 700:
        lines[i] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM categories`).'
    if 'mock.ExpectQuery(`SELECT COALESCE.*`)' in line and i > 700:
        lines[i] = '\tmock.ExpectQuery(`SELECT COALESCE(SUM(total_usd), 0) FROM orders WHERE status = $1`).'
    if 'mock.ExpectQuery(`SELECT COUNT.*FROM products WHERE status = \\'active\\'`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = 'active'`).'

# Fix 16: Auth update profile - SELECT bio, wallet_address
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT id, user_id, bio, avatar_url, wallet_address, created_at, updated_at FROM user_profiles WHERE user_id = $1`).'

# Fix 17: Auth update profile - SELECT id, email after UPDATE
for i, line in enumerate(lines):
    if i >= 78 and i <= 85 and 'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`)' in line:
        lines[i] = '\t\tmock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).'

# Fix 18: Order status check
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT o.id, o.status.*FROM orders WHERE id = $1`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).'

# Fix 19: New products tests
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT p.id, p.name.*FROM products p.*`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT p.id, p.name, p.slug FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).'

# Fix 20: Admin settings GET
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT value FROM settings WHERE key = $1`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT setting_key, value, updated_at FROM settings WHERE setting_key = $1`).'

# Fix 21: Admin exchange rate set/ get one
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1`).'

# Fix 22: TestNewProducts_GetActiveStat
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT COUNT.*FROM products WHERE status = \\'active\\'`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = 'active'`).'
    if 'mock.ExpectQuery(`SELECT COUNT.*FROM products`)' in line and i > 700:
        lines[i] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products`).'
    if 'mock.ExpectQuery(`SELECT COALESCE.*`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT COALESCE(SUM(total_usd), 0) FROM orders WHERE status = $1`).'

# Fix 23: TestNewProducts_Delete
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT COUNT\\(\\*\\) FROM products WHERE slug = $1`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).'
    if 'mock.ExpectExec(`DELETE FROM product_images WHERE product_id = $1`)' in line:
        lines[i] = '\t\tmock.ExpectExec(`DELETE FROM product_images WHERE product_id = $1`).'
    if 'mock.ExpectExec(`DELETE FROM products WHERE id = $1`)' in line:
        lines[i] = '\t\tmock.ExpectExec(`DELETE FROM products WHERE id = $1`).'

# Fix 24: Category products queries in admin
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT p.id, p.title.*FROM products p.*`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).'

# Fix 25: Orders query in admin guest order check
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT o.id, o.status.*FROM orders WHERE id = $1`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).'

# Fix 26: Admin community post list
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT cp.id, cp.user_id.*FROM community_posts cp.*`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url, COALESCE(cl.count_likes, 0), COALESCE(cc.count_comments, 0), EXISTS(SELECT 1 FROM follows WHERE follower_id = $2 AND following_id = cp.user_id) FROM community_posts cp LEFT JOIN users u ON cp.user_id = u.id LEFT JOIN user_profiles up ON u.id = up.user_id LEFT JOIN (SELECT post_id, COUNT(*) as count_likes FROM likes WHERE type = 'post' GROUP BY post_id) cl ON cp.id = cl.post_id LEFT JOIN (SELECT post_id, COUNT(*) as count_comments FROM comments WHERE type = 'post' GROUP BY post_id) cc ON cp.id = cc.post_id WHERE cp.type = 'post' ORDER BY cp.created_at DESC LIMIT $3 OFFSET $4`).'

# Fix 27: Admin community user list
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT COUNT\\(\\*\\) FROM users`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM users`).'
    if 'mock.ExpectQuery(`SELECT u.id, u.email, u.name.*`)' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT u.id, u.email, u.name, COUNT(cp.id) FROM users u LEFT JOIN community_posts cp ON u.id = cp.user_id WHERE cp.type = 'post' GROUP BY u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`).'

with open("handlers/all_features_test.go", "w") as f:
    f.write('\n'.join(lines))

print("Applied fixes to all mock SQL patterns")
