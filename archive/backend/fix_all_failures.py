#!/usr/bin/env python3
"""Fix all SQL mock mismatches in one pass."""
import re

with open("/root/project/backend/handlers/all_features_test.go") as f:
    content = f.read()

lines = content.split('\n')

# For each failing test, fix the mock SQL to match the handler's actual query.
# Based on reading the handler source files.

fixes = {
    # TestProducts_GetProducts - lines 123,125
    122: "mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).",
    124: "mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).",
    
    # TestProducts_GetProductBySlug - line 138
    137: "mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).",
    
    # TestProducts_SearchProducts - line 157
    156: "mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE (p.title ILIKE $1 OR p.description ILIKE $1) AND p.status = $2`).",
    
    # TestAuth_UpdateProfile_Valid - lines 80-81 (UPDATE users)
    # Find the UPDATE users SET line between lines 78-88
    # Fix Cart_GetOrCreate - SELECT id FROM carts
    # Fix Cart_AddItem - INSERT INTO cart_items, SELECT EXISTS
    # Fix Cart_RemoveItem - DELETE FROM cart_items
    # Fix Wishlist_Toggle - INSERT, SELECT EXISTS
    # Fix Wishlist_List - SELECT w.product_id
    # Fix Reviews_Create - INSERT INTO reviews
    # Fix Reviews_List - SELECT r.id
    # Fix RecentlyViewed_Record - INSERT INTO recently_viewed
    # Fix RecentlyViewed_List - SELECT p.id
    # Fix Comparison_Toggle - INSERT, SELECT EXISTS
    # Fix Comparison_List - SELECT p.id
    # Fix Recommendations_Get - SELECT category_id, SELECT p.id/p.name
    # Fix GuestOrder_Create - INSERT INTO guest_orders
    # Fix GuestOrder_Check - SELECT id,email,status FROM guest_orders
    # Fix OrderStatusCheck - SELECT o.id,o.status FROM orders
    # Fix NewProducts_GetActiveStat - SELECT COUNT FROM products WHERE status='active', SELECT COUNT(*) FROM products, SELECT COALESCE
    # Fix NewProducts_Delete - SELECT COUNT FROM products WHERE slug, DELETE FROM product_images, DELETE FROM products
    # Fix ExchangeRates_GetOne - SELECT chain,rate FROM exchange_rates WHERE chain
    # Fix ExchangeRates_GetAll - SELECT chain,rate FROM exchange_rates ORDER BY
    # Fix Admin_Stats - SELECT COUNT(*) FROM users, SELECT COUNT(*) FROM orders, SELECT COUNT(*) FROM categories, SELECT COALESCE
    # Fix Admin_OrderStatus* - UPDATE orders SET status
    # Fix Admin_GuestOrderCheck - SELECT o.id,o.status FROM orders
    # Fix Admin_GetSetting - SELECT setting_key,value FROM settings WHERE setting_key
    # Fix Admin_ExchangeRateSet - INSERT INTO exchange_rates
    # Fix Admin_ExchangeRateList - SELECT chain,rate FROM exchange_rates ORDER BY
    # Fix Admin_DeleteUser - DELETE FROM users WHERE id
    # Fix Admin_DeleteCommunityPost - DELETE FROM community_posts WHERE id
    # Fix Admin_CommunityPostList - SELECT cp.id,cp.user_id FROM community_posts
    # Fix Admin_CommunityUserList - SELECT COUNT(*) FROM users, SELECT u.id,u.email,u.name
    # Fix Admin_BulkProductStatus/Category/Toggle - UPDATE products SET
}

# Apply fixes by searching for patterns and replacing
def replace_line(pattern, replacement, start_line=0, end_line=None):
    """Replace first matching line with replacement."""
    for i in range(start_line, len(lines) if end_line is None else end_line):
        if pattern in lines[i]:
            old = lines[i]
            lines[i] = replacement
            print(f"Line {i+1}: {old[:60]}... -> {replacement[:60]}...")
            return True
    return False

# Products
replace_line('SELECT COUNT\\(\\*\\) FROM products.*`', 
             'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).')
replace_line('SELECT p.id, p.title.*`',
             'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).')
replace_line('SELECT COUNT\\(\\*\\) FROM products WHERE slug',
             'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).')
replace_line('SELECT COUNT\\(\\*\\) FROM products.*`,\n\t\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))',
             'PLACEHOLDER')  # Search products count

# Auth Update Profile
replace_line('mock.ExpectExec(`UPDATE users SET.*`',
             'mock.ExpectExec(`UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).')
replace_line('SELECT bio, wallet_address FROM user_profiles WHERE user_id',
             'mock.ExpectQuery(`SELECT id, user_id, bio, avatar_url, wallet_address, created_at, updated_at FROM user_profiles WHERE user_id = $1`).')

# Cart
replace_line('SELECT id FROM carts WHERE user_id',
             'mock.ExpectQuery(`SELECT id, user_id, product_count FROM carts WHERE user_id = $1`).')
replace_line('INSERT INTO cart_items.*`',
             'mock.ExpectExec(`INSERT INTO cart_items (cart_id, product_id, quantity, created_at) VALUES ($1, $2, $3, NOW()) RETURNING id`).')
replace_line('SELECT EXISTS\\(SELECT 1 FROM cart_items WH',
             'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM cart_items WHERE cart_id = $1 AND product_id = $2)`).')
replace_line('DELETE FROM cart_items WHERE cart_id',
             'mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).')

# Wishlist
replace_line('INSERT INTO wishlists.*`',
             'mock.ExpectExec(`INSERT INTO wishlists (user_id, product_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT (user_id, product_id) DO NOTHING`).')
replace_line('SELECT EXISTS\\(SELECT 1 FROM wishlists WH',
             'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).')
replace_line('SELECT w.product_id, p.name, p.slug.*`',
             'mock.ExpectQuery(`SELECT w.product_id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type FROM wishlists w JOIN products p ON w.product_id = p.id WHERE w.user_id = $1 ORDER BY w.created_at DESC`).')

# Reviews
replace_line('INSERT INTO reviews.*`',
             'mock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment, created_at) VALUES ($1, $2, $3, $4, NOW()) RETURNING id`).')
replace_line('SELECT r.id, r.rating, r.comment.*`',
             'mock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.name FROM reviews r LEFT JOIN users u ON r.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC`).')

# Recently Viewed
replace_line('INSERT INTO recently_viewed.*`',
             'mock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT (user_id, product_id) DO UPDATE SET created_at = NOW()`).')
replace_line('SELECT p.id, p.name, p.slug.*`.\n\t\t\tAddRow(1, "Test Product"',
             'mock.ExpectQuery(`SELECT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id = $1`).')

# Comparison
replace_line('SELECT EXISTS\\(SELECT 1 FROM product_comparison WH',
             'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).')
replace_line('INSERT INTO product_comparison.*`',
             'mock.ExpectExec(`INSERT INTO product_comparison (user_id, product_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT (user_id, product_id) DO NOTHING`).')
replace_line('SELECT p.id, p.name, p.slug.*`.\n\t\t\tAddRow(1, "Test Product"',
             'mock.ExpectQuery(`SELECT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id = $1`).', end_line=600)

# Recommendations
replace_line('SELECT category_id FROM products WHERE id',
             'mock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).')
replace_line('SELECT p.id, p.title.*FROM products p.*`',
             'mock.ExpectQuery(`SELECT DISTINCT p.id, p.title, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.category_id = $1 AND p.status = $2 ORDER BY p.created_at DESC LIMIT $3 OFFSET $4`).')
replace_line('SELECT p.id, p.name.*FROM products p.*`',
             'mock.ExpectQuery(`SELECT DISTINCT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.category_id = $1 AND p.status = $2 ORDER BY p.created_at DESC LIMIT $3 OFFSET $4`).')

# Guest Orders
replace_line('INSERT INTO guest_orders.*`',
             'mock.ExpectExec(`INSERT INTO guest_orders (email, total_usd, crypto_chain, crypto_amount, crypto_address, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, NOW()) RETURNING id`).')
replace_line('SELECT id, email, status.*FROM guest_orders WHERE id',
             'mock.ExpectQuery(`SELECT id, email, status, total_usd, crypto_chain, crypto_amount, crypto_address, created_at FROM guest_orders WHERE id = $1`).')

# Order Status Check
replace_line('SELECT o.id, o.status.*FROM orders WHERE id',
             'mock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).')

# New Products
replace_line("SELECT COUNT\\(\\*\\) FROM products WHERE status = 'active'`",
             "mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = 'active'`).")
replace_line('SELECT COUNT\\(\\*\\) FROM products`,\n\t\t\tWillReturnRows(sqlmock.NewRows',
             'mock.ExpectQuery(`SELECT COUNT(*) FROM products`).', end_line=710)
replace_line('SELECT COALESCE.*`,\n\t\t\tWillReturnRows(sqlmock.NewRows([]string{"coalesce"})',
             'mock.ExpectQuery(`SELECT COALESCE(SUM(total_usd), 0) FROM orders WHERE status = $1`).')

# Exchange Rates
replace_line('SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`',
             'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).')
replace_line('SELECT chain, rate, updated_at FROM exchange_rates WHERE chain',
             'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1`).')
replace_line('INSERT INTO exchange_rates.*`',
             'mock.ExpectExec(`INSERT INTO exchange_rates (chain, rate, updated_at) VALUES ($1, $2, NOW()) ON CONFLICT (chain) DO UPDATE SET rate = $2, updated_at = NOW()`).')

# Admin Settings
replace_line('SELECT value FROM settings WHERE key',
             'mock.ExpectQuery(`SELECT setting_key, value, updated_at FROM settings WHERE setting_key = $1`).')
replace_line('INSERT INTO settings.*`',
             'mock.ExpectExec(`INSERT INTO settings (setting_key, value, updated_at) VALUES ($1, $2, NOW()) ON CONFLICT (setting_key) DO UPDATE SET value = $2, updated_at = NOW()`).')

# Admin Stats
replace_line('SELECT COUNT\\(\\*\\) FROM users`,\n\t\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(25))',
             'mock.ExpectQuery(`SELECT COUNT(*) FROM users`).', end_line=740)
replace_line('SELECT COUNT\\(\\*\\) FROM orders`,\n\t\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))',
             'mock.ExpectQuery(`SELECT COUNT(*) FROM orders`).', end_line=740)
replace_line('SELECT COUNT\\(\\*\\) FROM categories`,\n\t\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))',
             'mock.ExpectQuery(`SELECT COUNT(*) FROM categories`).', end_line=740)
replace_line('SELECT COALESCE.*`,\n\t\t\tWillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow("100.00"))',
             'mock.ExpectQuery(`SELECT COALESCE(SUM(total_usd), 0) FROM orders WHERE status = $1`).', end_line=740)

# Admin Order Status
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`UPDATE orders SET.*`' in line:
        lines[i] = '\t\tmock.ExpectExec(`UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).'
        print(f"Line {i+1}: Fixed UPDATE orders SET")

# Admin Guest Order Check
replace_line('SELECT o.id, o.status.*FROM orders WHERE id = \\\\$1`,\n\t\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "status", "total_usd", "created_at"})',
             'mock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).')

# Admin Get Setting
replace_line('SELECT setting_key, value, updated_at FROM settings WHERE setting_key = $1`,\n\t\t\tWillReturnRows',
             'mock.ExpectQuery(`SELECT setting_key, value, updated_at FROM settings WHERE setting_key = $1`).')

# Admin Delete User
replace_line('DELETE FROM users WHERE id = \\\\$1`,\n\t\tWithArgs(1).WillReturnResult',
             'mock.ExpectExec(`DELETE FROM users WHERE id = $1`).')

# Admin Delete Community Post
replace_line('DELETE FROM community_posts WHERE id = \\\\$1`,\n\t\tWithArgs(1).WillReturnResult',
             'mock.ExpectExec(`DELETE FROM community_posts WHERE id = $1`).')

# Admin Community Post List
replace_line('SELECT cp.id, cp.user_id.*`,\n\t\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "user_id"',
             'mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url, COALESCE(cl.count_likes, 0), COALESCE(cc.count_comments, 0), EXISTS(SELECT 1 FROM follows WHERE follower_id = $2 AND following_id = cp.user_id) FROM community_posts cp LEFT JOIN users u ON cp.user_id = u.id LEFT JOIN user_profiles up ON u.id = up.user_id LEFT JOIN (SELECT post_id, COUNT(*) as count_likes FROM likes WHERE type = 'post' GROUP BY post_id) cl ON cp.id = cl.post_id LEFT JOIN (SELECT post_id, COUNT(*) as count_comments FROM comments WHERE type = 'post' GROUP BY post_id) cc ON cp.id = cc.post_id WHERE cp.type = 'post' ORDER BY cp.created_at DESC LIMIT $3 OFFSET $4`).')

# Admin Community User List
replace_line('SELECT COUNT\\(\\*\\) FROM users`,\n\t\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(25))',
             'mock.ExpectQuery(`SELECT COUNT(*) FROM users`).')
replace_line('SELECT u.id, u.email, u.name.*`,\n\t\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "count"})',
             'mock.ExpectQuery(`SELECT u.id, u.email, u.name, COUNT(cp.id) FROM users u LEFT JOIN community_posts cp ON u.id = cp.user_id WHERE cp.type = 'post' GROUP BY u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`).')

# Admin Bulk Product Status/Category/Toggle
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`UPDATE products SET.*`' in line and i > 200:
        lines[i] = '\t\tmock.ExpectExec(`UPDATE products SET status = $1, category_id = $2, updated_at = CURRENT_TIMESTAMP WHERE id = ANY($3)`).'
        print(f"Line {i+1}: Fixed UPDATE products SET")

# New Products Delete
replace_line('SELECT COUNT\\(\\*\\) FROM products WHERE slug = \\\\$1`,\n\t\t\tWithArgs("test-product").WillReturnRows',
             'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).')
replace_line('DELETE FROM product_images WHERE product_id = \\\\$1`,\n\t\tWithArgs(1).WillReturnResult',
             'mock.ExpectExec(`DELETE FROM product_images WHERE product_id = $1`).')
replace_line('DELETE FROM products WHERE id = \\\\$1`,\n\t\tWithArgs(1).WillReturnResult',
             'mock.ExpectExec(`DELETE FROM products WHERE id = $1`).')

# New Products GetActiveStat - second products list query
replace_line('SELECT p.id, p.name, p.slug.*`,\n\t\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug"})',
             'mock.ExpectQuery(`SELECT p.id, p.name, p.slug FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).', end_line=720)

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write('\n'.join(lines))

print("\nDone. Testing...")
import subprocess
result = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True, cwd="/root/project/backend")
print(result.stdout[-1000:])
print(result.stderr[-500:])
