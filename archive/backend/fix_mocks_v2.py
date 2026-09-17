#!/usr/bin/env python3
"""Fix all_features_test.go mock SQL to match real handler SQL."""
import re

PATH = "/root/project/backend/handlers/all_features_test.go"
with open(PATH) as f:
    content = f.read()

def replace_sql_in_mock(name, old_sql, new_sql):
    """Replace SQL pattern inside a mock.ExpectQuery/ExpectExec call."""
    global content
    if old_sql not in content:
        print(f"  MISS: {name}")
        return False
    content = content.replace(old_sql, new_sql, 1)
    print(f"  OK: {name}")
    return True

# ============================================================
# PRODUCTS - exact SQL from products.go
# ============================================================
replace_sql_in_mock("GET products COUNT",
    "SELECT COUNT(\*\\) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = \$1",
    "SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1")

replace_sql_in_mock("GET products LIST",
    "SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3",
    "p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3")

replace_sql_in_mock("GET product images",
    "SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \$1 ORDER BY is_primary DESC, id",
    "id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id")

replace_sql_in_mock("GET product by slug COUNT",
    "SELECT COUNT(\*\\) FROM products WHERE slug = \$1",
    "SELECT COUNT(*) FROM products WHERE slug = $1")

replace_sql_in_mock("GET product by slug LIST",
    "SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.slug = $1 AND p.status = 'active'",
    "p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.slug = $1 AND p.status = 'active'")

replace_sql_in_mock("Search products COUNT",
    "SELECT COUNT(\*\\) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = \$1 AND p.search_vector @@ plainto_tsquery('english', \$2)",
    "SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND p.search_vector @@ plainto_tsquery('english', $2)")

replace_sql_in_mock("Search products LIST",
    "SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND p.search_vector @@ plainto_tsquery('english', $2) ORDER BY p.created_at DESC LIMIT $3 OFFSET $4",
    "p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND p.search_vector @@ plainto_tsquery('english', $2) ORDER BY p.created_at DESC LIMIT $3 OFFSET $4")

# ============================================================
# CART - exact SQL from features.go AddToCart
# ============================================================
replace_sql_in_mock("Cart stock query",
    "SELECT stock_quantity(?s).*FROM products WHERE id=\$1",
    "SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id = $1")

replace_sql_in_mock("Cart items INSERT",
    "INSERT INTO cart_items(?s).*",
    "INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, $3) ON CONFLICT (cart_id, product_id) DO UPDATE SET quantity = $3, added_at = CURRENT_TIMESTAMP")

replace_sql_in_mock("Cart total query",
    "SELECT COALESCE\(SUM\(quantity\)\),\s*0\) FROM cart_items WHERE cart_id = \$1",
    "SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE cart_id = $1")

# ============================================================
# CART RemoveItem
# ============================================================
replace_sql_in_mock("Cart DELETE items",
    "DELETE FROM cart_items WHERE cart_id = \$1 AND product_id = \$2",
    "DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2")

# ============================================================
# WISHLIST
# ============================================================
replace_sql_in_mock("Wishlist EXISTS",
    "SELECT EXISTS\(SELECT 1 FROM wishlists WHERE user_id = \$1 AND product_id = \$2\)",
    "SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)")

replace_sql_in_mock("Wishlist INSERT",
    "INSERT INTO wishlists(?s).*",
    "INSERT INTO wishlists (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING")

replace_sql_in_mock("Wishlist LIST",
    "SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, '') as primary_image FROM wishlists w JOIN products p ON p.id = w.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE w.user_id = rf...[truncated]",
    "p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, '') as primary_image FROM wishlists w JOIN products p ON p.id = w.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE w.user_id = $1 ORDER BY w.created_at DESC")

# ============================================================
# REVIEWS
# ============================================================
replace_sql_in_mock("Reviews INSERT",
    "INSERT INTO reviews(?s).*",
    "INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) ON CONFLICT (product_id, user_id) DO UPDATE SET rating = $3, comment = $4, updated_at = CURRENT_TIMESTAMP")

replace_sql_in_mock("Reviews build INSERT",
    "INSERT INTO reviews(?s).*",
    "INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) ON CONFLICT (product_id, user_id) DO UPDATE SET rating = $3, comment = $4, updated_at = CURRENT_TIMESTAMP")

replace_sql_in_mock("Reviews list SELECT",
    "SELECT r.id, r.product_id, r.rating, r.comment(?s).*FROM reviews r(?s).*",
    "r.id, r.rating, r.comment, r.created_at, u.id, u.name, u.email, COALESCE(up.avatar_url, '') as avatar_url FROM reviews r JOIN users u ON u.id = r.user_id LEFT JOIN user_profiles up ON up.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC LIMIT $2 OFFSET $3")

# ============================================================
# RECENTLY VIEWED
# ============================================================
replace_sql_in_mock("RecentlyViewed INSERT",
    "INSERT INTO recently_viewed(?s).*",
    "INSERT INTO recently_viewed (user_id, product_id, viewed_at) VALUES ($1, $2, CURRENT_TIMESTAMP) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = CURRENT_TIMESTAMP")

replace_sql_in_mock("RecentlyViewed DELETE cleanup",
    "DELETE FROM recently_viewed WHERE user_id = \$1 AND id NOT IN \(SELECT id FROM recently_viewed WHERE user_id = \$1 ORDER BY viewed_at DESC LIMIT 20\)",
    "DELETE FROM recently_viewed WHERE user_id = $1 AND id NOT IN (SELECT id FROM recently_viewed WHERE user_id = $1 ORDER BY viewed_at DESC LIMIT 20)")

replace_sql_in_mock("RecentlyViewed LIST",
    "SELECT p.id, p.title, p.slug(?s).*FROM recently_viewed rv(?s).*",
    "p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, '') as primary_image, rv.viewed_at FROM recently_viewed rv JOIN products p ON p.id = rv.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE rv.user_id = $1 ORDER BY rv.viewed_at DESC LIMIT $2")

# ============================================================
# COMPARISON
# ============================================================
replace_sql_in_mock("Comparison EXISTS",
    "SELECT EXISTS\(SELECT 1 FROM product_comparison WHERE user_id = \$1 AND product_id = \$2\)",
    "SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)")

replace_sql_in_mock("Comparison INSERT",
    "INSERT INTO product_comparison(?s).*",
    "INSERT INTO product_comparison (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING")

replace_sql_in_mock("Comparison LIST",
    "SELECT p.id, p.title, p.slug(?s).*FROM product_comparison pc(?s).*",
    "p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.description, COALESCE(p.asset_path, '') as asset_path, COALESCE(p.asset_hash, '') as asset_hash, COALESCE(p.file_mime_type, '') as file_mime_type, COALESCE(created_at::text, '') as created_at FROM product_comparison pc JOIN products p ON p.id = pc.product_id WHERE pc.user_id = $1 ORDER BY pc.added_at DESC")

replace_sql_in_mock("Comparison image query",
    "SELECT COALESCE\(url, ''\)\s*FROM product_images WHERE product_id = \$1 AND is_primary = true LIMIT 1",
    "SELECT COALESCE(url, '') FROM product_images WHERE product_id = $1 AND is_primary = true LIMIT 1")

# ============================================================
# RECOMMENDATIONS
# ============================================================
replace_sql_in_mock("Recommendations category lookup",
    "SELECT category_id FROM products WHERE id = \$1",
    "SELECT category_id FROM products WHERE id = $1")

replace_sql_in_mock("Recommendations same-cat query",
    "SELECT p.id, p.title, p.slug(?s).*FROM products p(?s).*p\.category_id = \$1 AND p\.id != \$2",
    "p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, '') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.category_id = $1 AND p.id != $2 AND p.status = 'active' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $3 OFFSET $4")

replace_sql_in_mock("Recommendations global query",
    "SELECT p.id, p.title(?s).*FROM products p(?s).*p\.id != \$1",
    "p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, '') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.id != $1 AND p.status = 'active' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $2 OFFSET $3")

# ============================================================
# GUEST ORDER
# ============================================================
replace_sql_in_mock("Guest order sum query",
    "SELECT COALESCE\(SUM\(CAST\(price_usd AS NUMERIC\)\)\), 0\) FROM products WHERE id = ANY\(\$1::int\[\]\) AND status = 'active'",
    "SELECT COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0) FROM products WHERE id = ANY($1::int[]) AND status = 'active'")

replace_sql_in_mock("Guest order INSERT orders",
    "INSERT INTO guest_orders(?s).*",
    "INSERT INTO orders (email, total_usd, crypto_chain, status, guest_order) VALUES ($1, $2, $3, 'pending', true) RETURNING id")

replace_sql_in_mock("Guest order CHECK goods",
    "SELECT id, email, status(?s).*FROM guest_orders WHERE id = \$1",
    "SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = $1 AND guest_order = true")

# ============================================================
# ORDER STATUS CHECK
# ============================================================
replace_sql_in_mock("Order status check SELECT",
    "SELECT id, status(?s).*FROM orders WHERE id = \$1",
    "SELECT id, user_id, status, total_crypto, crypto_chain, payment_address, payment_tx_hash, payment_confirmations FROM orders WHERE id = $1 AND user_id = $2")

# ============================================================
# ADMIN ORDERS
# ============================================================
replace_sql_in_mock("Admin order detail SELECT order",
    "SELECT o.id, o.status(?s).*FROM orders o WHERE o.id = \$1",
    "SELECT o.status, CAST(o.total_usd AS NUMERIC), o.guest_email, o.billing_name, o.shipping_address_json, o.notes, o.created_at, o.paid_at FROM orders o WHERE o.id = $1")

replace_sql_in_mock("Admin order detail SELECT items",
    "SELECT oi.product_id, oi.quantity(?s).*FROM order_items oi WHERE oi.order_id = \$1",
    "SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price_usd, oi.price_crypto, oi.download_count, oi.max_downloads, oi.downloaded_at, oi.created_at, p.title, p.slug FROM order_items oi JOIN products p ON oi.product_id = p.id WHERE oi.order_id = $1")

replace_sql_in_mock("Admin orders COUNT",
    "SELECT COUNT(\*\\) FROM orders WHERE 1=1",
    "SELECT COUNT(*) FROM orders")

replace_sql_in_mock("Admin orders LIST",
    "SELECT o.id, o.status(?s).*FROM orders o(?s).*",
    "o.id, o.user_id, o.status, o.total_usd, o.total_crypto, o.crypto_chain, o.payment_address, o.payment_tx_hash, o.payment_confirmations, o.payment_confirmed_at, o.paid_at, o.created_at, o.updated_at FROM orders o ORDER BY o.created_at DESC LIMIT $1 OFFSET $2")

# ============================================================
# ADMIN PRODUCTS (NewProducts)
# ============================================================
replace_sql_in_mock("Admin products COUNT",
    "SELECT COUNT(\*\\) FROM products WHERE 1=1",
    "SELECT COUNT(*) FROM products WHERE 1=1")

replace_sql_in_mock("Admin products LIST",
    "SELECT p.id, p.title(?s).*FROM products p(?s).*",
    "p.id, p.title, p.slug, p.description, COALESCE(c.name, '') as category_name, COALESCE(c.slug, '') as category_slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.status, p.stock_quantity, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.asset_path, p.asset_hash, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE 1=1 ORDER BY p.created_at DESC LIMIT $1 OFFSET $2")

# ============================================================
# ADMIN BULK OPERATIONS
# ============================================================
replace_sql_in_mock("Admin bulk UPDATE status",
    "UPDATE products SET(?s).*",
    "UPDATE products SET status = $1 WHERE id = ANY($2)")

replace_sql_in_mock("Admin bulk UPDATE category",
    "UPDATE products SET(?s).*",
    "UPDATE products SET category_id = $1 WHERE id = ANY($2)")

replace_sql_in_mock("Admin bulk UPDATE toggle",
    "UPDATE products SET(?s).*",
    "UPDATE products SET status = $1 WHERE id = ANY($2)")

# ============================================================
# ADMIN EXCHANGE RATES
# ============================================================
replace_sql_in_mock("Admin exchange rate INSERT",
    "INSERT INTO exchange_rates(?s).*",
    "INSERT INTO exchange_rates (chain, symbol, rate_to_usd) VALUES ($1, $2, $3) ON CONFLICT (chain) DO UPDATE SET symbol = $2, rate_to_usd = $3, updated_at = NOW()")

replace_sql_in_mock("Admin exchange rate SELECT",
    "SELECT chain, rate\(?s\).*FROM exchange_rates(?s).*",
    "chain, rate, updated_at FROM exchange_rates ORDER BY chain")

# ============================================================
# ADMIN SETTINGS
# ============================================================
replace_sql_in_mock("Admin setting SET",
    "INSERT INTO settings(?s).*",
    "INSERT INTO settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()")

replace_sql_in_mock("Admin setting GET",
    "SELECT id, key, value(?s).*FROM settings WHERE key = \$1",
    "SELECT id, key, value, updated_at FROM settings WHERE key = $1")

# ============================================================
# ADMIN USER / COMMUNITY MODERATION
# ============================================================
replace_sql_in_mock("Admin delete user SELECT",
    "SELECT id, email, name, created_at, updated_at FROM users WHERE id = \$1",
    "SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1")

replace_sql_in_mock("Admin delete post EXISTS",
    "SELECT EXISTS\(SELECT 1 FROM community_posts WHERE id = \$1\)",
    "SELECT EXISTS(SELECT 1 FROM community_posts WHERE id = $1)")

replace_sql_in_mock("Admin community posts COUNT",
    "SELECT COUNT(\*\\) FROM community_posts WHERE 1=1",
    "SELECT COUNT(*) FROM community_posts WHERE 1=1")

replace_sql_in_mock("Admin community posts LIST",
    "SELECT cp.id, cp.user_id(?s).*FROM community_posts cp(?s).*",
    "cp.id, cp.user_id, u.name, u.email, cp.content, cp.created_at, cp.updated_at FROM community_posts cp LEFT JOIN users u ON u.id = cp.user_id WHERE 1=1 ORDER BY cp.created_at DESC LIMIT $1 OFFSET $2")

replace_sql_in_mock("Admin community users COUNT",
    "SELECT COUNT(\*\\) FROM users WHERE 1=1",
    "SELECT COUNT(*) FROM users")

replace_sql_in_mock("Admin community users LIST",
    "SELECT u.id, u.name(?s).*FROM users u(?s).*",
    "u.id, u.name, u.email, u.created_at, u.updated_at, COALESCE(up.bio, ''), COALESCE(up.avatar_url, ''), COALESCE(up.wallet_address, '') FROM users u LEFT JOIN user_profiles up ON up.user_id = u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2")

# ============================================================
# ADMIN GUEST ORDER CHECK
# ============================================================
replace_sql_in_mock("Admin guest order check SELECT",
    "SELECT guest_email FROM orders WHERE id = \$1",
    "SELECT guest_email FROM orders WHERE id = $1")

# ============================================================
# ADMIN ORDER STATUS SET
# ============================================================
replace_sql_in_mock("Admin order status UPDATE paid",
    "UPDATE orders SET status = 'paid'(?s).*",
    "UPDATE orders SET status = 'paid', paid_at = NOW() WHERE id = $1")

replace_sql_in_mock("Admin order status UPDATE completed",
    "UPDATE orders SET status = 'completed'(?s).*",
    "UPDATE orders SET status = 'completed' WHERE id = $1")

replace_sql_in_mock("Admin order status UPDATE cancelled",
    "UPDATE orders SET status = 'cancelled'(?s).*",
    "UPDATE orders SET status = 'cancelled' WHERE id = $1")

replace_sql_in_mock("Admin order status UPDATE refunded",
    "UPDATE orders SET status = 'refunded'(?s).*",
    "UPDATE orders SET status = 'refunded' WHERE id = $1")

# ============================================================
# COMMUNITY PROFILE / FEED
# ============================================================
replace_sql_in_mock("Community profile user_profiles SELECT",
    "SELECT id, user_id, bio, avatar_url, wallet_address, created_at, updated_at FROM user_profiles WHERE user_id = \$1",
    "SELECT bio, avatar_url, wallet_address, created_at FROM user_profiles WHERE user_id = $1")

replace_sql_in_mock("Community feed SELECT",
    "SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up\.avatar_url, COUNT\(DISTINCT l\.id\) as count_likes, COUNT\(DISTINCT c\.id\) as count_comments, EXISTS\(SELECT 1 FROM follows WHERE follower_id = \$1 AND followee_id = cp\.user_id\) as following(?s).*",
    "cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, COALESCE(up.avatar_url, '') as avatar_url, (SELECT COUNT(*) FROM post_likes WHERE post_id = cp.id) as count_likes, (SELECT COUNT(*) FROM post_comments WHERE post_id = cp.id) as count_comments, EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = cp.user_id) as following FROM community_posts cp JOIN users u ON u.id = cp.user_id LEFT JOIN user_profiles up ON up.user_id = u.id LEFT JOIN post_likes l ON l.post_id = cp.id LEFT JOIN post_comments c ON c.post_id = cp.id WHERE cp.user_id = $1 GROUP BY cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url")

with open(PATH, "w") as f:
    f.write(content)

print("\nDone.")
