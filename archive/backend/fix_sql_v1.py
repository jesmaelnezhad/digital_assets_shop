#!/usr/bin/env python3
"""
Fix all_features_test.go mock SQL to match real handler SQL.
"""
import re, subprocess

SRC = "/root/project/backend/handlers/all_features_test.go"
with open(SRC) as f:
    text = f.read()

changes = 0

def fix(name, old, new):
    global text, changes
    if old not in text:
        print(f"  SKIP: {name}")
        return False
    text = text.replace(old, new, 1)
    changes += 1
    print(f"  OK: {name}")
    return True

def fix_all(name, old, new):
    global text, changes
    n = text.count(old)
    if n == 0:
        print(f"  SKIP: {name}")
        return
    text = text.replace(old, new)
    changes += n
    print(f"  OK ({n}x): {name}")

# ============================================================
# Fix 1: Products_GetProducts - COUNT needs WithArgs
# ============================================================
fix("GetProducts COUNT WithArgs",
    'mock.ExpectQuery(`SELECT COUNT(\*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = \$1`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n\t\tWithArgs("active").\n\t\tWillReturnRows(')

# ============================================================
# Fix 2: Products_GetProducts - images need WithArgs and ORDER BY
# ============================================================
fix("GetProducts images 1 WithArgs",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \$1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# Second images query for product 2
fix("GetProducts images 2",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \$1`).\n\t\tWithArgs(2).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(2).\n\t\tWillReturnRows(')

# ============================================================
# Fix 3: Products_GetProductBySlug - COUNT and images
# ============================================================
fix("GetProductBySlug COUNT WithArgs",
    'mock.ExpectQuery(`SELECT COUNT(\*) FROM products WHERE slug = \$1`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n\t\tWithArgs("test-product").\n\t\tWillReturnRows(')

fix("GetProductBySlug images WithArgs",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \$1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# ============================================================
# Fix 4: Products_SearchProducts - WHERE clause + images
# ============================================================
fix("SearchProducts COUNT WHERE",
    'mock.ExpectQuery(`SELECT COUNT(\*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = \$1`).\n\t\tWithArgs("active").\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND (p.name LIKE $2 OR p.description LIKE $2 OR p.slug LIKE $2)`).\n\t\tWithArgs("active", "test").\n\t\tWillReturnRows(')

fix("SearchProducts images 1",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \$1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

fix("SearchProducts images 2",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \$1`).\n\t\tWithArgs(2).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(2).\n\t\tWillReturnRows(')

# ============================================================
# Fix 5: Cart_GetOrCreate - need WithArgs
# ============================================================
fix("Cart_GetOrCreate WithArgs",
    'mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = \$1`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# ============================================================
# Fix 6: Cart_AddItem - stock query + items insert + total
# ============================================================
fix("Cart_AddItem stock WithArgs",
    'mock.ExpectQuery(`SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id = \$1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# ============================================================
# Fix 7: Wishlist_Toggle - EXISTS query
# ============================================================
fix("Wishlist_Toggle EXISTS WithArgs",
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = \$1 AND product_id = \$2)`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).\n\t\tWithArgs(1, 1).\n\t\tWillReturnRows(')

# ============================================================
# Fix 8: Wishlist_List - SQL
# ============================================================
fix("Wishlist_List SQL",
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, \'\') as primary_image FROM wishlists w JOIN products p ON p.id = w.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE w.user_id = \$1 ORDER BY w.created_at DESC`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, \'\') as primary_image FROM wishlists w JOIN products p ON p.id = w.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE w.user_id = $1 ORDER BY w.created_at DESC`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# ============================================================
# Fix 9: Reviews - CREATE and LIST
# ============================================================
fix("Reviews_Create WithArgs",
    'mock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES (\$1, \$2, \$3, \$4) ON CONFLICT (product_id, user_id) DO UPDATE SET rating = \$3, comment = \$4, updated_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(',
    'mock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) ON CONFLICT (product_id, user_id) DO UPDATE SET rating = $3, comment = $4, updated_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(1, 1, 5, "Great product!").WillReturnResult(')

fix("Reviews_List SQL",
    'mock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.id, u.name, u.email, COALESCE(up.avatar_url, \'\') as avatar_url FROM reviews r JOIN users u ON u.id = r.user_id LEFT JOIN user_profiles up ON up.user_id = u.id WHERE r.product_id = \$1 ORDER BY r.created_at DESC LIMIT \$2 OFFSET \$3`).\n\t\tWithArgs(1, 20, 0).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.id, u.name, u.email, COALESCE(up.avatar_url, \'\') as avatar_url FROM reviews r JOIN users u ON u.id = r.user_id LEFT JOIN user_profiles up ON up.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC LIMIT $2 OFFSET $3`).\n\t\tWithArgs(1, 20, 0).\n\t\tWillReturnRows(')

# ============================================================
# Fix 10: RecentlyViewed
# ============================================================
fix("RecentlyViewed_Insert WithArgs",
    'mock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id, viewed_at) VALUES (\$1, \$2, CURRENT_TIMESTAMP) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(',
    'mock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id, viewed_at) VALUES ($1, $2, CURRENT_TIMESTAMP) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(1, 1).WillReturnResult(')

fix("RecentlyViewed_List SQL",
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, \'\') as primary_image, rv.viewed_at FROM recently_viewed rv JOIN products p ON p.id = rv.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE rv.user_id = \$1 ORDER BY rv.viewed_at DESC LIMIT \$2`).\n\t\tWithArgs(1, 20).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, \'\') as primary_image, rv.viewed_at FROM recently_viewed rv JOIN products p ON p.id = rv.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE rv.user_id = $1 ORDER BY rv.viewed_at DESC LIMIT $2`).\n\t\tWithArgs(1, 20).\n\t\tWillReturnRows(')

# ============================================================
# Fix 11: Comparison
# ============================================================
fix("Comparison_Toggle EXISTS WithArgs",
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = \$1 AND product_id = \$2)`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).\n\t\tWithArgs(1, 1).\n\t\tWillReturnRows(')

fix("Comparison_List SQL",
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.description, COALESCE(p.asset_path, \'\') as asset_path, COALESCE(p.asset_hash, \'\') as asset_hash, COALESCE(p.file_mime_type, \'\') as file_mime_type, COALESCE(created_at::text, \'\') as created_at FROM product_comparison pc JOIN products p ON p.id = pc.product_id WHERE pc.user_id = \$1 ORDER BY pc.added_at DESC`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.description, COALESCE(p.asset_path, \'\') as asset_path, COALESCE(p.asset_hash, \'\') as asset_hash, COALESCE(p.file_mime_type, \'\') as file_mime_type, COALESCE(created_at::text, \'\') as created_at FROM product_comparison pc JOIN products p ON p.id = pc.product_id WHERE pc.user_id = $1 ORDER BY pc.added_at DESC`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# ============================================================
# Fix 12: Recommendations
# ============================================================
fix("Recommendations cat lookup WithArgs",
    'mock.ExpectQuery(`SELECT category_id FROM products WHERE id = \$1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

fix("Recommendations same-cat SQL",
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, \'\') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.category_id = \$1 AND p.id != \$2 AND p.status = \'active\' ORDER BY p.views_count DESC, p.created_at DESC LIMIT \$3 OFFSET \$4`).\n\t\tWithArgs(1, 1, 5, 0).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, \'\') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.category_id = $1 AND p.id != $2 AND p.status = \'active\' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $3 OFFSET $4`).\n\t\tWithArgs(1, 1, 5, 0).\n\t\tWillReturnRows(')

fix("Recommendations global SQL",
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, \'\') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.id != \$1 AND p.status = \'active\' ORDER BY p.views_count DESC, p.created_at DESC LIMIT \$2 OFFSET \$3`).\n\t\tWithArgs(1, 5, 0).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, \'\') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.id != $1 AND p.status = \'active\' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $2 OFFSET $3`).\n\t\tWithArgs(1, 5, 0).\n\t\tWillReturnRows(')

# ============================================================
# Fix 13: Admin order status UPDATEs
# ============================================================
for status in ['paid', 'completed', 'cancelled', 'refunded']:
    extra = ", paid_at = NOW()" if status == "paid" else ""
    fix(f"Admin_{status} UPDATE SQL",
        f'mock.ExpectExec(`UPDATE orders SET.*`).\n\t\tWithArgs(1).\n\t\tWillReturnResult(',
        f'mock.ExpectExec(`UPDATE orders SET status = \'{status}\'{extra} WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnResult(')

# ============================================================
# Fix 14: Admin bulk product UPDATEs
# ============================================================
fix("Admin_BulkProductStatus UPDATE",
    'mock.ExpectExec(`UPDATE products SET.*`).\n\t\tWithArgs(1, 1).\n\t\tWillReturnResult(',
    'mock.ExpectExec(`UPDATE products SET status = $1 WHERE id = ANY($2)`).\n\t\tWithArgs("active", []int{1, 2}).\n\t\tWillReturnResult(')

fix("Admin_BulkProductCategory UPDATE",
    'mock.ExpectExec(`UPDATE products SET.*`).\n\t\tWithArgs(1, 1).\n\t\tWillReturnResult(',
    'mock.ExpectExec(`UPDATE products SET category_id = $1 WHERE id = ANY($2)`).\n\t\tWithArgs(1, []int{1, 2}).\n\t\tWillReturnResult(')

# ============================================================
# Fix 15: Admin products LIST
# ============================================================
fix("Admin_Products_List SQL",
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, \'\') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.category_id = \$1 AND p.id != \$2 AND p.status = \'active\' ORDER BY p.views_count DESC, p.created_at DESC LIMIT \$3 OFFSET \$4`).\n\t\tWithArgs(1, 1, 5, 0).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, \'\') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.category_id = $1 AND p.id != $2 AND p.status = \'active\' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $3 OFFSET $4`).\n\t\tWithArgs(1, 1, 5, 0).\n\t\tWillReturnRows(')

# ============================================================
# Fix 16: Admin orders LIST
# ============================================================
fix("Admin_Orders_List SQL",
    'mock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.total_crypto, o.crypto_chain, o.payment_address, o.payment_tx_hash, o.payment_confirmations, o.payment_confirmed_at, o.paid_at, o.created_at, o.updated_at FROM orders o ORDER BY o.created_at DESC LIMIT \$1 OFFSET \$2`).\n\t\tWithArgs(20, 0).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.total_crypto, o.crypto_chain, o.payment_address, o.payment_tx_hash, o.payment_confirmations, o.payment_confirmed_at, o.paid_at, o.created_at, o.updated_at FROM orders o ORDER BY o.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(20, 0).\n\t\tWillReturnRows(')

# ============================================================
# Fix 17: Admin settings GET
# ============================================================
fix("Admin_Setting_GET SQL",
    'mock.ExpectQuery(`SELECT id, key, value, updated_at FROM settings WHERE key = \$1`).\n\t\tWithArgs("payment_address").\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, key, value, updated_at FROM settings WHERE key = $1`).\n\t\tWithArgs("payment_address").\n\t\tWillReturnRows(')

# ============================================================
# Fix 18: Admin exchange rates LIST
# ============================================================
fix("Admin_ExchangeRate_List WithArgs",
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).\n\t\tWithArgs().\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).\n\t\tWithArgs().\n\t\tWillReturnRows(')

# ============================================================
# Fix 19: Admin guest order check
# ============================================================
fix("Admin_GuestOrderCheck SQL",
    'mock.ExpectQuery(`SELECT guest_email FROM orders WHERE id = \$1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT guest_email FROM orders WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# ============================================================
# Fix 20: Community feed
# ============================================================
fix("Community_Feed WithArgs",
    'mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, COALESCE(up.avatar_url, \'\') as avatar_url, (SELECT COUNT(*) FROM post_likes WHERE post_id = cp.id) as count_likes, (SELECT COUNT(*) FROM post_comments WHERE post_id = cp.id) as count_comments, EXISTS(SELECT 1 FROM follows WHERE follower_id = \$1 AND followee_id = cp.user_id) as following FROM community_posts cp JOIN users u ON u.id = cp.user_id LEFT JOIN user_profiles up ON up.user_id = u.id LEFT JOIN post_likes l ON l.post_id = cp.id LEFT JOIN post_comments c ON c.post_id = cp.id WHERE cp.user_id = \$1 GROUP BY cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, COALESCE(up.avatar_url, \'\') as avatar_url, (SELECT COUNT(*) FROM post_likes WHERE post_id = cp.id) as count_likes, (SELECT COUNT(*) FROM post_comments WHERE post_id = cp.id) as count_comments, EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = cp.user_id) as following FROM community_posts cp JOIN users u ON u.id = cp.user_id LEFT JOIN user_profiles up ON up.user_id = u.id LEFT JOIN post_likes l ON l.post_id = cp.id LEFT JOIN post_comments c ON c.post_id = cp.id WHERE cp.user_id = $1 GROUP BY cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# ============================================================
# Fix 21: Community profile
# ============================================================
fix("Community_Profile profile SELECT WithArgs",
    'mock.ExpectQuery(`SELECT bio, avatar_url, wallet_address, created_at FROM user_profiles WHERE user_id = \$1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT bio, avatar_url, wallet_address, created_at FROM user_profiles WHERE user_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# ============================================================
# Fix 22: Delete user
# ============================================================
fix("Admin_DeleteUser SELECT WithArgs",
    'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = \$1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# ============================================================
# Fix 23: Delete community post
# ============================================================
fix("Admin_DeletePost EXISTS WithArgs",
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM community_posts WHERE id = \$1)`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM community_posts WHERE id = $1)`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# ============================================================
# Fix 24: Admin community posts LIST
# ============================================================
fix("Admin_CommunityPosts_List SQL",
    'mock.ExpectQuery(`SELECT cp.id, cp.user_id, u.name, u.email, cp.content, cp.created_at, cp.updated_at FROM community_posts cp LEFT JOIN users u ON u.id = cp.user_id WHERE 1=1 ORDER BY cp.created_at DESC LIMIT \$1 OFFSET \$2`).\n\t\tWithArgs(20, 0).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT cp.id, cp.user_id, u.name, u.email, cp.content, cp.created_at, cp.updated_at FROM community_posts cp LEFT JOIN users u ON u.id = cp.user_id WHERE 1=1 ORDER BY cp.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(20, 0).\n\t\tWillReturnRows(')

# ============================================================
# Fix 25: Admin community users LIST
# ============================================================
fix("Admin_CommunityUsers_List SQL",
    'mock.ExpectQuery(`SELECT u.id, u.name, u.email, u.created_at, u.updated_at, COALESCE(up.bio, \'\') as bio, COALESCE(up.avatar_url, \'\') as avatar_url, COALESCE(up.wallet_address, \'\') as wallet_address FROM users u LEFT JOIN user_profiles up ON up.user_id = u.id ORDER BY u.created_at DESC LIMIT \$1 OFFSET \$2`).\n\t\tWithArgs(20, 0).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT u.id, u.name, u.email, u.created_at, u.updated_at, COALESCE(up.bio, \'\') as bio, COALESCE(up.avatar_url, \'\') as avatar_url, COALESCE(up.wallet_address, \'\') as wallet_address FROM users u LEFT JOIN user_profiles up ON up.user_id = u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(20, 0).\n\t\tWillReturnRows(')

# ============================================================
# Fix 26: Guest order create
# ============================================================
fix("GuestOrder_Create sum query",
    'mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0) FROM products WHERE id = ANY(\$1::int[]) AND status = \이었습니다active\'`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0) FROM products WHERE id = ANY($1::int[]) AND status = \'active\'`).\n\t\tWithArgs(')

# ============================================================
# Fix 27: Guest order check
# ============================================================
fix("GuestOrder_Check SQL",
    'mock.ExpectQuery(`SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = \$1 AND guest_order = true`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = $1 AND guest_order = true`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# ============================================================
# Fix 28: Order status check
# ============================================================
fix("OrderStatusCheck SQL",
    'mock.ExpectQuery(`SELECT id, user_id, status, total_crypto, crypto_chain, payment_address, payment_tx_hash, payment_confirmations FROM orders WHERE id = \$1 AND user_id = \$2`).\n\t\tWithArgs(1, 1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, user_id, status, total_crypto, crypto_chain, payment_address, payment_tx_hash, payment_confirmations FROM orders WHERE id = $1 AND user_id = $2`).\n\t\tWithArgs(1, 1).\n\t\tWillReturnRows(')

# ============================================================
# Fix 29: Create product
# ============================================================
fix("CreateProduct slug check WithArgs",
    'mock.ExpectQuery(`SELECT COUNT(\*) FROM products WHERE slug = \$1`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n\t\tWithArgs("test-product").\n\t\tWillReturnRows(')

fix("CreateProduct category check WithArgs",
    'mock.ExpectQuery(`SELECT COUNT(\*) FROM categories WHERE id = \$1`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM categories WHERE id = $1`).\n\t\tWithArgs(1).')

# ============================================================
# Write
# ============================================================
with open(SRC, 'w') as f:
    f.write(text)

print(f"\nApplied {changes} fixes")

# Build
r = subprocess.run(['go', 'build', './handlers/'], capture_output=True, text=True, cwd='/root/project/backend')
print(f"Build: {'OK' if r.returncode == 0 else 'FAIL'}\n{r.stderr[:300] if r.returncode else ''}")

if r.returncode == 0:
    # Test
    r2 = subprocess.run(['go', 'test', './handlers/', '-count=1', '-v'], capture_output=True, text=True, cwd='/root/project/backend', timeout=120)
    lines = (r2.stdout + r2.stderr).split('\n')
    passes = sum(1 for l in lines if '--- PASS:' in l)
    fails = sum(1 for l in lines if '--- FAIL:' in l)
    print(f"\nResult: {passes} PASS, {fails} FAIL, exit={r2.returncode}")
    for l in lines:
        if '--- FAIL:' in l or ('unfulfilled' in l and 'Expectation' in l and 'matches sql' in l):
            print(f"  {l.strip()[:140]}")
