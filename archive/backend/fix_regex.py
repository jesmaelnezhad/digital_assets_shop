#!/usr/bin/env python3
"""Fix all_features_test.go using proper regex replacements."""
import re

PATH = "/root/project/backend/handlers/all_features_test.go"
with open(PATH) as f:
    content = f.read()

def replace(name, pattern, replacement, flags=0):
    """Replace first occurrence. Report MISS if not found."""
    global content
    new, count = re.subn(pattern, replacement, content, count=1, flags=flags)
    if count == 0:
        print(f"  MISS: {name}")
        return False
    content = new
    print(f"  OK: {name}")
    return True

# ============================================================
# FIX 1: Double-escaped star in COUNT queries
# The test has: COUNT\(\*\\\\)  ->  sqlmock sees \\* which matches literal \*
# Handler sends: COUNT(*)  ->  sqlmock needs: COUNT\(\*\)
# ============================================================
replace("COUNT star escape (products list)",
    r"COUNT\(\*\\)",
    "COUNT\(\*\)")

replace("COUNT star escape (product by slug)",
    r"COUNT\(\*\\)",
    "COUNT\(\*\)")

replace("COUNT star escape (search)",
    r"COUNT\(\*\\)",
    "COUNT\(\*\)")

replace("COUNT star escape (admin products)",
    r"COUNT\(\*\\)",
    "COUNT\(\*\)")

replace("COUNT star escape (admin orders)",
    r"COUNT\(\*\\)",
    "COUNT\(\*\)")

replace("COUNT star escape (admin users)",
    r"COUNT\(\*\\)",
    "COUNT\(\*\)")

replace("COUNT star escape (admin community posts)",
    r"COUNT\(\*\\)",
    "COUNT\(\*\)")

# ============================================================
# FIX 2: Product images query - add ORDER BY
# Handler sends: ORDER BY is_primary DESC, id
# Mock currently: no ORDER BY
# ============================================================
replace("Product images ORDER BY (GetProducts)",
    r"FROM product_images WHERE product_id = \\\$1`\)",
    "FROM product_images WHERE product_id = \\\$1 ORDER BY is_primary DESC, id`)")

replace("Product images ORDER BY (GetProductBySlug)",
    r"FROM product_images WHERE product_id = \\\$1`\)",
    "FROM product_images WHERE product_id = \\\$1 ORDER BY is_primary DESC, id`)")

replace("Product images ORDER BY (SearchProducts)",
    r"FROM product_images WHERE product_id = \\\$1`\)",
    "FROM product_images WHERE product_id = \\\$1 ORDER BY is_primary DESC, id`)")

# Need to fix the 4th occurrence too (for product by slug)
replace("Product images ORDER BY (4th occurrence)",
    r"FROM product_images WHERE product_id = \\\$1`\)",
    "FROM product_images WHERE product_id = \\\$1 ORDER BY is_primary DESC, id`)")

# ============================================================
# FIX 3: Search products WHERE clause
# Handler sends: (p.name LIKE $2 OR p.description LIKE $2 OR p.slug LIKE $2)
# Current mock already has this! But the COUNT mock is wrong (has wrong WHERE)
# ============================================================
# The search COUNT mock currently only has WHERE p.status = $1
# It needs WHERE p.status = $1 AND p.search_vector @@ plainto_tsquery('english', $2)
# But wait - the handler actually uses LIKE not tsquery! Let me check...
# Handler: models.SearchProducts uses ILIKE on name, description, slug
# So the mock's WHERE clause is already correct for search:
#   WHERE p.status = $1 AND (p.name LIKE $2 OR p.description LIKE $2 OR p.slug LIKE $2)
# 
# But the COUNT mock only has: WHERE p.status = $1
# It's missing the AND clause! This is a separate mock call.
# Let me fix the COUNT mock's WHERE clause.
replace("Search products COUNT WHERE",
    r"WHERE p\.status = \\\$1`"\)\.WillReturnRows\(sqlmock\.NewRows\(\[\]string\{\"count\"\}\)\",
    "WHERE p.status = \\$1 AND (p.name LIKE \\$2 OR p.description LIKE \\$2 OR p.slug LIKE \\$2`)).WillReturnRows(sqlmock.NewRows([]string{\"count\"})")

# Wait - this is getting complex. The LIKE needs proper escaping.
# ILIKE vs LIKE: handler uses LIKE (case-sensitive in PG default)
# Let me just match the simpler approach: use $2 as-is for LIKE patterns
replace("Search products COUNT WHERE (2nd try)",
    r"WHERE p\.status = \\\$1 AND \(p\.name LIKE \\\$2 OR p\.description LIKE \\\$2 OR p\.slug LIKE \\\$2\)`\)\.WillReturnRows\(sqlmock\.NewRows\(\[\]string\{\"count\"\}\)",
    "WHERE p.status = \\$1 AND (p.name LIKE \\\\$2 OR p.description LIKE \\\\$2 OR p.slug LIKE \\\\$2`)).WillReturnRows(sqlmock.NewRows([]string{\"count\"})")

# ============================================================
# FIX 4: Cart stock query uses CAST(price_usd AS INTEGER) not NUMERIC
# ============================================================
replace("Cart stock CAST type",
    r"FROM products WHERE id = \\\$1`\)",
    "FROM products WHERE id = \\\$1`)")

# The actual mock line is:
# mock.ExpectQuery(`SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id = $1`).WithArgs(1)...
# This looks correct already!

# ============================================================
# FIX 5: Community feed - comment alias c.id vs cm.id
# Handler: LEFT JOIN post_comments cm ON cm.post_id = cp.id  (alias is 'cm')
# Mock: LEFT JOIN post_comments cm ON cm.post_id = cp.id  (already cm!)
# Then: COUNT(DISTINCT cm.id) as count_comments  (uses cm, correct!)
# But the mock pattern has: COUNT(DISTINCT c.id) - wrong alias!
replace("Community feed comment alias",
    r"COUNT\(DISTINCT c\.id\)",
    "COUNT(DISTINCT cm.id)")

# ============================================================
# FIX 6: Guest order CHECK - missing columns
# Handler: SELECT email, status, total_usd, crypto_chain FROM orders WHERE id=$1 AND guest_order=true
# Mock: SELECT id, email, status(?s).*FROM guest_orders WHERE id=$1
# The (?s).* matches everything after status, including the rest
# But the FROM clause says 'guest_orders' not 'orders'
# And the mock only has 3 columns in the row data
# Let me check what the mock currently has...
replace("Guest order CHECK FROM clause",
    r"FROM guest_orders WHERE id = \\\$1`\)",
    "FROM orders WHERE id = \\\$1 AND guest_order = true`)")

replace("Guest order CHECK columns",
    r"id, email, status\(.*?\)",
    "email, status, total_usd, crypto_chain")

# ============================================================
# FIX 7: Wishlist LIST - fix the .* wildcard to exact columns
# Handler: SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, '') as primary_image FROM wishlists w JOIN products p ON p.id = w.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE w.user_id = $1 ORDER BY w.created_at DESC
# Mock: SELECT w.product_id, p.title, p.slug.*FROM wishlists w.*
# This won't match because:
#   - w.product_id != p.id (alias mismatch)
#   - .* after p.slug won't match the CAST... part properly
#   - FROM wishlists w.* won't match the full JOIN clause
replace("Wishlist LIST query",
    r"SELECT w\.product_id, p\.title, p\.slug\.\*FROM wishlists w\.\*",
    "SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, '') as primary_image FROM wishlists w JOIN products p ON p.id = w.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE w.user_id = $1 ORDER BY w.created_at DESC")

# ============================================================
# FIX 8: Reviews LIST - fix the .* wildcard
# Handler: SELECT r.id, r.rating, r.comment, r.created_at, u.id, u.name, u.email, COALESCE(up.avatar_url, '') as avatar_url FROM reviews r JOIN users u ON u.id = r.user_id LEFT JOIN user_profiles up ON up.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC LIMIT $2 OFFSET $3
# Mock: SELECT r.id, r.product_id, r.rating, r.comment.*FROM reviews r.*
replace("Reviews LIST query",
    r"SELECT r\.id, r\.product_id, r\.rating, r\.comment\.\*FROM reviews r\.\*",
    "SELECT r.id, r.rating, r.comment, r.created_at, u.id, u.name, u.email, COALESCE(up.avatar_url, '') as avatar_url FROM reviews r JOIN users u ON u.id = r.user_id LEFT JOIN user_profiles up ON up.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC LIMIT $2 OFFSET $3")

# ============================================================
# FIX 9: RecentlyViewed LIST - fix the .* wildcard
# Handler: SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, '') as primary_image, rv.viewed_at FROM recently_viewed rv JOIN products p ON p.id = rv.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE rv.user_id = $1 ORDER BY rv.viewed_at DESC LIMIT $2
# Mock: SELECT p.id, p.title, p.slug.*FROM recently_viewed rv.*
replace("RecentlyViewed LIST query",
    r"SELECT p\.id, p\.title, p\.slug\.\*FROM recently_viewed rv\.\*",
    "SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, '') as primary_image, rv.viewed_at FROM recently_viewed rv JOIN products p ON p.id = rv.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE rv.user_id = $1 ORDER BY rv.viewed_at DESC LIMIT $2")

# ============================================================
# FIX 10: Comparison LIST - fix the .* wildcard
# Handler: SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.description, COALESCE(p.asset_path, '') as asset_path, COALESCE(p.asset_hash, '') as asset_hash, COALESCE(p.file_mime_type, '') as file_mime_type, COALESCE(created_at::text, '') as created_at FROM product_comparison pc JOIN products p ON p.id = pc.product_id WHERE pc.user_id = $1 ORDER BY pc.added_at DESC
# Mock: SELECT p.id, p.title, p.slug.*FROM product_comparison pc.*
replace("Comparison LIST query",
    r"SELECT p\.id, p\.title, p\.slug\.\*FROM product_comparison pc\.\*",
    "SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.description, COALESCE(p.asset_path, '') as asset_path, COALESCE(p.asset_hash, '') as asset_hash, COALESCE(p.file_mime_type, '') as file_mime_type, COALESCE(created_at::text, '') as created_at FROM product_comparison pc JOIN products p ON p.id = pc.product_id WHERE pc.user_id = $1 ORDER BY pc.added_at DESC")

# ============================================================
# FIX 11: Recommendations same-cat query - fix the .* wildcard
# Handler: SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, '') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.category_id = $1 AND p.id != $2 AND p.status = 'active' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $3 OFFSET $4
# Mock: SELECT p.id, p.title.*FROM products p.*  ...  p.category_id = $1 AND p.id != $2
replace("Recommendations same-cat query",
    r"SELECT p\.id, p\.title\.\*FROM products p\.\*.*?p\.category_id = \\\$1 AND p\.id != \\\$2",
    "SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, '') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.category_id = $1 AND p.id != $2 AND p.status = 'active' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $3 OFFSET $4")

# ============================================================
# FIX 12: Recommendations global query - fix the .* wildcard
# Handler: SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, '') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.id != $1 AND p.status = 'active' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $2 OFFSET $3
# Mock: SELECT p.id, p.title.*FROM products p.*  ...  p.id != $1
replace("Recommendations global query",
    r"SELECT p\.id, p\.title, COALESCE\(c\.name, ''\)\s*as category_name\.\*FROM products p\.\*.*?p\.id != \\\$1",
    "SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, '') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.id != $1 AND p.status = 'active' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $2 OFFSET $3")

# ============================================================
# FIX 13: Admin products LIST - fix the .* wildcard
# Handler: SELECT p.id, p.title, p.slug, p.description, COALESCE(c.name, '') as category_name, COALESCE(c.slug, '') as category_slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.status, p.stock_quantity, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.asset_path, p.asset_hash, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE 1=1 ORDER BY p.created_at DESC LIMIT $1 OFFSET $2
# Mock: SELECT p.id, p.title.*FROM products p.*
replace("Admin products LIST query",
    r"SELECT p\.id, p\.title\.\*FROM products p\.\*",
    "SELECT p.id, p.title, p.slug, p.description, COALESCE(c.name, '') as category_name, COALESCE(c.slug, '') as category_slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.status, p.stock_quantity, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.asset_path, p.asset_hash, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE 1=1 ORDER BY p.created_at DESC LIMIT $1 OFFSET $2")

# ============================================================
# FIX 14: Admin orders LIST - fix the .* wildcard
# Handler: SELECT o.id, o.user_id, o.status, o.total_usd, o.total_crypto, o.crypto_chain, o.payment_address, o.payment_tx_hash, o.payment_confirmations, o.payment_confirmed_at, o.paid_at, o.created_at, o.updated_at FROM orders o ORDER BY o.created_at DESC LIMIT $1 OFFSET $2
# Mock: SELECT o.id, o.status.*FROM orders o.*
replace("Admin orders LIST query",
    r"SELECT o\.id, o\.status\.\*FROM orders o\.\*",
    "SELECT o.id, o.user_id, o.status, o.total_usd, o.total_crypto, o.crypto_chain, o.payment_address, o.payment_tx_hash, o.payment_confirmations, o.payment_confirmed_at, o.paid_at, o.created_at, o.updated_at FROM orders o ORDER BY o.created_at DESC LIMIT $1 OFFSET $2")

# ============================================================
# FIX 15: Admin order detail SELECT items - fix the .* wildcard
# Handler: SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price_usd, oi.price_crypto, oi.download_count, oi.max_downloads, oi.downloaded_at, oi.created_at, p.title, p.slug FROM order_items oi JOIN products p ON oi.product_id = p.id WHERE oi.order_id = $1
# Mock: SELECT oi.product_id, oi.quantity.*FROM order_items oi WHERE oi.order_id=$1
replace("Admin order items query",
    r"SELECT oi\.product_id, oi\.quantity\.\*FROM order_items oi WHERE oi\.order_id = \\\$1`\)",
    "SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price_usd, oi.price_crypto, oi.download_count, oi.max_downloads, oi.downloaded_at, oi.created_at, p.title, p.slug FROM order_items oi JOIN products p ON oi.product_id = p.id WHERE oi.order_id = $1`)")

# ============================================================
# FIX 16: Admin order detail SELECT order - fix the .* wildcard
# Handler: SELECT o.status, CAST(o.total_usd AS NUMERIC), o.guest_email, o.billing_name, o.shipping_address_json, o.notes, o.created_at, o.paid_at FROM orders o WHERE o.id = $1
# Mock: SELECT o.id, o.status.*FROM orders o WHERE o.id=$1
replace("Admin order detail query",
    r"SELECT o\.id, o\.status\.\*FROM orders o WHERE o\.id = \\\$1`\)",
    "SELECT o.status, CAST(o.total_usd AS NUMERIC), o.guest_email, o.billing_name, o.shipping_address_json, o.notes, o.created_at, o.paid_at FROM orders o WHERE o.id = $1`)")

# ============================================================
# FIX 17: Admin setting GET - missing columns
# Handler: SELECT id, key, value, updated_at FROM settings WHERE key = $1
# Mock: SELECT value FROM settings WHERE key=$1
# The mock only returns 1 column but handler expects 4
replace("Admin setting GET columns",
    r"SELECT value FROM settings WHERE key = \\\$1`\)",
    "SELECT id, key, value, updated_at FROM settings WHERE key = $1`)")

# ============================================================
# FIX 18: Admin exchange rate LIST - fix the .* wildcard  
# Handler: SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain
# Mock: SELECT chain, rate(?s).*FROM exchange_rates(?s).*
# The (?s).* wildcards make this match anything, BUT the row data columns might be wrong
replace("Admin exchange rate LIST query",
    r"SELECT chain, rate\(.*?\)FROM exchange_rates\(.*?\)",
    "SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain")

# ============================================================
# FIX 19: Admin community posts LIST - fix the .* wildcard
# Handler: SELECT cp.id, cp.user_id, u.name, u.email, cp.content, cp.created_at, cp.updated_at FROM community_posts cp LEFT JOIN users u ON u.id = cp.user_id WHERE 1=1 ORDER BY cp.created_at DESC LIMIT $1 OFFSET $2
# Mock: SELECT cp.id, cp.user_id.*FROM community_posts cp.*
replace("Admin community posts LIST query",
    r"SELECT cp\.id, cp\.user_id\.\*FROM community_posts cp\.\*",
    "SELECT cp.id, cp.user_id, u.name, u.email, cp.content, cp.created_at, cp.updated_at FROM community_posts cp LEFT JOIN users u ON u.id = cp.user_id WHERE 1=1 ORDER BY cp.created_at DESC LIMIT $1 OFFSET $2")

# ============================================================
# FIX 20: Admin community users LIST - fix the .* wildcard
# Handler: SELECT u.id, u.name, u.email, u.created_at, u.updated_at, COALESCE(up.bio, ''), COALESCE(up.avatar_url, ''), COALESCE(up.wallet_address, '') FROM users u LEFT JOIN user_profiles up ON up.user_id = u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2
# Mock: SELECT u.id, u.name.*FROM users u.*
replace("Admin community users LIST query",
    r"SELECT u\.id, u\.name\.\*FROM users u\.\*",
    "SELECT u.id, u.name, u.email, u.created_at, u.updated_at, COALESCE(up.bio, ''), COALESCE(up.avatar_url, ''), COALESCE(up.wallet_address, '') FROM users u LEFT JOIN user_profiles up ON up.user_id = u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2")

# ============================================================
# FIX 21: Guest order CHECK goods - fix columns
# Already handled above
# ============================================================
replace("Guest order CHECK - SELECT",
    r"SELECT id, email, status\(.*?\)FROM guest_orders WHERE id = \\\$1`\)",
    "SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = $1 AND guest_order = true`)")

# ============================================================
# FIX 22: Order status check - fix the .* wildcard
# Handler: SELECT id, user_id, status, total_crypto, crypto_chain, payment_address, payment_tx_hash, payment_confirmations FROM orders WHERE id = $1 AND user_id = $2
# Mock: SELECT id, status(?s).*FROM orders WHERE id=$1
replace("Order status check SELECT",
    r"SELECT id, status\(.*?\)FROM orders WHERE id = \\\$1`\)",
    "SELECT id, user_id, status, total_crypto, crypto_chain, payment_address, payment_tx_hash, payment_confirmations FROM orders WHERE id = $1 AND user_id = $2`)")

# ============================================================
# FIX 23: Cart total query - fix exact match
# Handler: SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE cart_id = $1
# Let me check if the mock has the right SQL...
# Actually the mock might already be correct. Let me check with grep.
# ============================================================
replace("Cart total SELECT",
    r"SELECT COALESCE\(SUM\(quantity\)\),\s*0\) FROM cart_items WHERE cart_id = \\\$1`\)",
    "SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE cart_id = $1`)")

# ============================================================
# FIX 24: Admin order status UPDATEs - fix .* wildcards
# Handler: UPDATE orders SET status = 'paid', paid_at = NOW() WHERE id = $1
# Mock: UPDATE orders SET(?s).*
# The (?s).* wildcard makes any UPDATE match, but the handler expects specific columns
# Actually (?s).* matches everything including newlines, so it should match any UPDATE
# But the row data in WillReturnResult might matter...
# These are likely OK since (?s).* matches everything
# ============================================================
# These are probably fine. Skip.

# ============================================================
# WRITE
# ============================================================
with open(PATH, "w") as f:
    f.write(content)

print("\nDone.")
