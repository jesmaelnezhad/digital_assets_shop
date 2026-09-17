#!/usr/bin/env python3
"""Fix all mock SQL mismatches - comprehensive batch fix."""
import re, subprocess

with open("/root/project/backend/handlers/all_features_test.go") as f:
    content = f.read()

lines = content.split('\n')

replacements = [
    # Auth: SELECT id, email, password_hash... FROM users WHERE email = \$1
    (r'mock\.ExpectQuery\(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email = \\\$1`\)\.',
     'mock.ExpectQuery(`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email = $1`).'),
    
    # Auth: SELECT id, email, name... FROM users WHERE id = $1 (for GetMe and UpdateProfile read)
    (r'mock\.ExpectQuery\(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = \\\$1`\)\.',
     'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).'),
    
    # Auth: UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2
    (r'mock\.ExpectExec\(`UPDATE users SET.*`\)\.',
     'mock.ExpectExec(`UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).'),
    
    # Auth: SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1
    (r'mock\.ExpectQuery\(`SELECT bio, wallet_address FROM user_profiles WHERE user_id = \\\$1`\)\.',
     'mock.ExpectQuery(`SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1`).'),
    
    # Products: COUNT query for list
    (r'mock\.ExpectQuery\(`SELECT COUNT\(\\\\\\*\) FROM products.*`\)\.',
     'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).'),
    
    # Products: Full product list query
    (r'mock\.ExpectQuery\(`SELECT p\.id, p\.title\..*`\)\.',
     'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).'),
    
    # Products: GetBySlug COUNT
    (r'mock\.ExpectQuery\(`SELECT COUNT\(\\\\\\*\) FROM products WHERE slug = \\\$1`\)\.',
     'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).'),
    
    # Products: GetBySlug full query
    (r'mock\.ExpectQuery\(`SELECT p\.id, p\.title\..*slug = \\\$1.*`\)\.',
     'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.slug = $1`).'),
    
    # Products: Search COUNT
    (r'mock\.ExpectQuery\(`SELECT COUNT\(\\\\\\*\) FROM products.*`\)\.\n\t\tWillReturnRows\(sqlmock\.NewRows\(\[\]string{"count"}\)\.AddRow\(1\)\)',
     'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE (p.title ILIKE $1 OR p.description ILIKE $1) AND p.status = $2`).'),
    
    # Products: Search full query
    (r'mock\.ExpectQuery\(`SELECT p\.id, p\.title\..*%test%.*`\)\.',
     'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE (p.title ILIKE $1 OR p.description ILIKE $1) AND p.status = $2 ORDER BY p.created_at DESC LIMIT $3 OFFSET $4`).'),
    
    # Cart: GetOrCreate - SELECT id FROM carts
    (r'mock\.ExpectQuery\(`SELECT id FROM carts WHERE user_id = \\\$1`\)\.',
     'mock.ExpectQuery(`SELECT id, user_id, product_count FROM carts WHERE user_id = $1`).'),
    
    # Cart: AddItem - INSERT INTO cart_items  
    (r'mock\.ExpectExec\(`INSERT INTO cart_items\..*`\)\.',
     'mock.ExpectExec(`INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, $3) ON CONFLICT (cart_id, product_id) DO UPDATE SET quantity = cart_items.quantity + $3 RETURNING id`).'),
    
    # Cart: RemoveItem - DELETE FROM cart_items
    (r'mock\.ExpectExec\(`DELETE FROM cart_items WHERE cart_id = \\\$1 AND product_id = \\\$2`\)\.',
     'mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).'),
    
    # Wishlist: Toggle - INSERT INTO wishlists
    (r'mock\.ExpectExec\(`INSERT INTO wishlists\..*`\)\.',
     'mock.ExpectExec(`INSERT INTO wishlists (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).'),
    
    # Wishlist: Toggle - SELECT EXISTS
    (r'mock\.ExpectQuery\(`SELECT EXISTS\(SELECT 1 FROM wishlists WHERE user_id = \\\$1 AND product_id = \\\$2\)`\)\.',
     'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).'),
    
    # Wishlist: List
    (r'mock\.ExpectQuery\(`SELECT w\.product_id, p\.name, p\.slug\..*`\)\.',
     'mock.ExpectQuery(`SELECT w.product_id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status FROM wishlists w JOIN products p ON w.product_id = p.id WHERE w.user_id = $1`).'),
    
    # Reviews: Create - INSERT INTO reviews
    (r'mock\.ExpectExec\(`INSERT INTO reviews\..*`\)\.',
     'mock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) RETURNING id`).'),
    
    # Reviews: List
    (r'mock\.ExpectQuery\(`SELECT r\.id, r\.rating, r\.comment\..*`\)\.',
     'mock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.name FROM reviews r JOIN users u ON r.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC`).'),
    
    # Recently Viewed: Record
    (r'mock\.ExpectExec\(`INSERT INTO recently_viewed\..*`\)\.',
     'mock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = NOW()`).'),
    
    # Recently Viewed: List
    (r'mock\.ExpectQuery\(`SELECT p\.id, p\.name, p\.slug\..*`\)\.',
     'mock.ExpectQuery(`SELECT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM recently_viewed WHERE user_id = $1 ORDER BY viewed_at DESC LIMIT 10)`).'),
    
    # Comparison: Toggle
    (r'mock\.ExpectExec\(`INSERT INTO product_comparison\..*`\)\.',
     'mock.ExpectExec(`INSERT INTO product_comparison (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).'),
    
    # Comparison: List
    (r'mock\.ExpectQuery\(`SELECT p\.id, p\.name, p\.slug\..*`\)\.',
     'mock.ExpectQuery(`SELECT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM product_comparison WHERE user_id = $1 ORDER BY added_at DESC LIMIT 4)`).'),
    
    # Recommendations
    (r'mock\.ExpectQuery\(`SELECT category_id FROM products WHERE id = \\\$1`\)\.',
     'mock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).'),
    
    # Recommendations: products query (two variants - same category and global)
    (r'mock\.ExpectQuery\(`SELECT p\.id, p\.title\..*FROM products p\..*`\)\.',
     'mock.ExpectQuery(`SELECT DISTINCT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.category_id = $1 AND p.status = $2 AND p.id != $3 ORDER BY p.created_at DESC LIMIT $4 OFFSET $5`).'),
    
    # Guest Order: Create
    (r'mock\.ExpectExec\(`INSERT INTO guest_orders\..*`\)\.',
     'mock.ExpectExec(`INSERT INTO guest_orders (email, total_usd, crypto_chain, crypto_amount, crypto_address, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`).'),
    
    # Guest Order: Check
    (r'mock\.ExpectQuery\(`SELECT id, email, status\..*FROM guest_orders WHERE id = \\\$1`\)\.',
     'mock.ExpectQuery(`SELECT id, email, status, total_usd, crypto_chain, crypto_amount, crypto_address FROM guest_orders WHERE id = $1`).'),
    
    # Order Status Check
    (r'mock\.ExpectQuery\(`SELECT o\.id, o\.status\..*FROM orders WHERE id = \\\$1`\)\.',
     'mock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).'),
    
    # New Products: GetActiveStat
    (r'mock\.ExpectQuery\(`SELECT COUNT\(\\\\\\*\) FROM products WHERE status = 'active'`\)\.',
     'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = ''''active''''`).'),
    (r'mock\.ExpectQuery\(`SELECT COUNT\(\\\\\\*\) FROM products`\)\.',
     'mock.ExpectQuery(`SELECT COUNT(*) FROM products`).'),
    (r'mock\.ExpectQuery\(`SELECT COALESCE.*`\)\.',
     'mock.ExpectQuery(`SELECT COALESCE(SUM(total_usd), 0) FROM orders WHERE status = $1`).'),
    
    # New Products: Delete  
    (r'mock\.ExpectQuery\(`SELECT COUNT\(\\\\\\*\) FROM products WHERE slug = \\\$1`\)\.',
     'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).'),
    (r'mock\.ExpectExec\(`DELETE FROM product_images WHERE product_id = \\\$1`\)\.',
     'mock.ExpectExec(`DELETE FROM product_images WHERE product_id = $1`).'),
    (r'mock\.ExpectExec\(`DELETE FROM products WHERE id = \\\$1`\)\.',
     'mock.ExpectExec(`DELETE FROM products WHERE id = $1`).'),
    
    # Exchange Rates: GetOne
    (r'mock\.ExpectQuery\(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = \\\$1`\)\.',
     'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1`).'),
    
    # Exchange Rates: GetAll
    (r'mock\.ExpectQuery\(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`\)\.',
     'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).'),
    
    # Admin: Stats
    (r'mock\.ExpectQuery\(`SELECT COUNT\(\\\\\\*\) FROM users`\)\.',
     'mock.ExpectQuery(`SELECT COUNT(*) FROM users`).'),
    (r'mock\.ExpectQuery\(`SELECT COUNT\(\\\\\\*\) FROM orders`\)\.',
     'mock.ExpectQuery(`SELECT COUNT(*) FROM orders`).'),
    (r'mock\.ExpectQuery\(`SELECT COUNT\(\\\\\\*\) FROM categories`\)\.',
     'mock.ExpectQuery(`SELECT COUNT(*) FROM categories`).'),
    (r'mock\.ExpectQuery\(`SELECT COALESCE.*`\)\.\n.*AddRow\(\''100\.00\'\)',
     'mock.ExpectQuery(`SELECT COALESCE(SUM(total_usd), 0) FROM orders WHERE status = $1`).'),
    
    # Admin: OrderStatus transitions
    (r'mock\.ExpectExec\(`UPDATE orders SET.*`\)\.',
     'mock.ExpectExec(`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`).'),
    
    # Admin: GuestOrderCheck
    (r'mock\.ExpectQuery\(`SELECT o\.id, o\.status.*FROM orders WHERE id = \\\$1`\)\.',
     'mock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).'),
    
    # Admin: GetSetting
    (r'mock\.ExpectQuery\(`SELECT value FROM settings WHERE key = \\\$1`\)\.',
     'mock.ExpectQuery(`SELECT setting_key, value, updated_at FROM settings WHERE setting_key = $1`).'),
    
    # Admin: ExchangeRateSet
    (r'mock\.ExpectExec\(`INSERT INTO exchange_rates.*`\)\.',
     'mock.ExpectExec(`INSERT INTO exchange_rates (chain, rate) VALUES ($1, $2) ON CONFLICT (chain) DO UPDATE SET rate = $2, updated_at = NOW()`).'),
    
    # Admin: ExchangeRateList  
    (r'mock\.ExpectQuery\(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`\)\.',
     'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).'),
    
    # Admin: SetSetting
    (r'mock\.ExpectExec\(`INSERT INTO settings.*`\)\.',
     'mock.ExpectExec(`INSERT INTO settings (setting_key, value) VALUES ($1, $2) ON CONFLICT (setting_key) DO UPDATE SET value = $2, updated_at = NOW()`).'),
    
    # Admin: DeleteUser
    (r'mock\.ExpectExec\(`DELETE FROM users WHERE id = \\\$1`\)\.',
     'mock.ExpectExec(`DELETE FROM users WHERE id = $1`).'),
    
    # Admin: DeleteCommunityPost
    (r'mock\.ExpectExec\(`DELETE FROM community_posts WHERE id = \\\$1`\)\.',
     'mock.ExpectExec(`DELETE FROM community_posts WHERE id = $1`).'),
    
    # Admin: CommunityPostList
    (r'mock\.ExpectQuery\(`SELECT cp\.id, cp\.user_id\..*FROM community_posts cp.*`\)\.',
     'mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url, COALESCE(cl.count, 0), COALESCE(cc.count, 0), EXISTS(SELECT 1 FROM follows WHERE follower_id = $2 AND following_id = cp.user_id) FROM community_posts cp LEFT JOIN users u ON cp.user_id = u.id LEFT JOIN user_profiles up ON cp.user_id = up.user_id LEFT JOIN (SELECT post_id, COUNT(*) as count FROM likes WHERE type = ''''''''post'''''''' GROUP BY post_id) cl ON cp.id = cl.post_id LEFT JOIN (SELECT post_id, COUNT(*) as count FROM comments WHERE type = ''''''''post'''''''' GROUP BY post_id) cc ON cp.id = cc.post_id WHERE cp.type = ''''''''post'''''''' ORDER BY cp.created_at DESC LIMIT $3 OFFSET $4`).'),
    
    # Admin: CommunityUserList
    (r'mock\.ExpectQuery\(`SELECT COUNT\(\\\\\\*\) FROM users`\)\.',
     'mock.ExpectQuery(`SELECT COUNT(*) FROM users`).'),
    
    # DBSchema: Migration011Tables
    (r'mock\.ExpectQuery\(`SELECT table_name FROM information_schema\.tables WHERE table_schema = ''''public'''' AND table_name = \\\$1'`\)\.',
     'mock.ExpectQuery(`SELECT table_name FROM information_schema.tables WHERE table_schema = ''''public'''' AND table_name = $1`).'),
]

applied = 0
for pattern, replacement in replacements:
    if re.search(pattern, content):
        content = re.sub(pattern, replacement, content, flags=re.DOTALL)
        applied += 1

print(f"Applied {applied} replacements")

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(content)

print("Running build...")
result = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True, cwd="/root/project/backend")
print(f"Build: exit={result.returncode}")
if result.returncode != 0:
    print(result.stderr[-500:])

print("\nRunning tests...")
result = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True, cwd="/root/project/backend")
print(result.stdout[-1500:])
if result.returncode != 0:
    print("STDERR:", result.stderr[-500:])
