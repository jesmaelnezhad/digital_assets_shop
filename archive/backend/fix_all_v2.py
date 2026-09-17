#!/usr/bin/env python3
"""Comprehensive fix for all_features_test.go - fix all failing tests at once."""
import re, subprocess, os

SRC = "/root/project/backend/handlers/all_features_test.go"
with open(SRC) as f:
    text = f.read()

changes = 0
def fix(name, old, new):
    global text, changes
    if old not in text:
        print(f"  MISS: {name}")
        return False
    text = text.replace(old, new, 1)
    changes += 1
    print(f"  OK: {name}")
    return True

# ============================================================
# 1. GetProducts - fix COUNT mock (currently uses .* wildcard)
# ============================================================
# Current: mock.ExpectQuery(`SELECT COUNT\(\*\\) FROM products.*`).
# Fix: replace with exact SQL + WithArgs
fix("GetProducts COUNT",
    'mock.ExpectQuery(`SELECT COUNT\\(\\*\\) FROM products.*`).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n\t\tWithArgs("active").\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))')

# ============================================================
# 2. GetProducts - fix images queries (need ORDER BY + 2nd image query for product 2)
# ============================================================
# Current images query doesn't have ORDER BY, and there's only 1 image query but handler calls it twice
# Check if images query has ORDER BY already (it was fixed in a previous run)
images_line = None
for i, line in enumerate(text.split('\n')):
    if 'product_images WHERE product_id = $1' in line and 'mock.ExpectQuery' in line:
        images_line = i
        break

if images_line is not None:
    line = text.split('\n')[images_line]
    if 'ORDER BY is_primary' not in line:
        fix("GetProducts images ORDER BY",
            'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
            'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')
    
    # Check for second images query for product 2
    has_second = 'product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(2)' in text
    if not has_second:
        # Need to add second images query after the first one's WillReturnRows
        # Find the end of the first images query block
        first_images_end = text.find('mockTime(2024, time.January, 1)))', images_line)
        if first_images_end >= 0:
            insert_pos = first_images_end + len('mockTime(2024, time.January, 1)))')
            second_images = '\n\tmock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(2).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "url", "is_primary", "created_at"}).\n\t\t\tAddRow(2, 2, "http://example.com/img2.png", true, mockTime(2024, time.January, 1)))'
            text = text[:insert_pos] + second_images + text[insert_pos:]
            print("  OK: GetProducts images 2 (added)")
            changes += 1

# ============================================================
# 3. GetProductBySlug - fix COUNT and images
# ============================================================
fix("GetProductBySlug COUNT",
    'mock.ExpectQuery(`SELECT COUNT\\(\\*\\) FROM products WHERE slug = \\$1`).\n\t\tWithArgs("test-product").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n\t\tWithArgs("test-product").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))')

fix("GetProductBySlug images",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# ============================================================
# 4. SearchProducts - fix WHERE clause
# ============================================================
fix("SearchProducts COUNT",
    'mock.ExpectQuery(`SELECT COUNT\\(\\*\\) FROM products.*`).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND (p.name LIKE $2 OR p.description LIKE $2 OR p.slug LIKE $2)`).\n\t\tWithArgs("active", "%test%").\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))')

fix("SearchProducts images 1",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

fix("SearchProducts images 2",  
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(2).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(2).\n\t\tWillReturnRows(')

# ============================================================
# 5. Cart tests - fix WithArgs
# ============================================================
fix("Cart_GetOrCreate",
    'mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

fix("Cart_AddItem stock",
    'mock.ExpectQuery(`SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id = $1`).\n\t\tWithArgs(1).')

fix("Cart_RemoveItem",
    'mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).\n\t\tWithArgs(',
    'mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).\n\t\tWithArgs(1, 1).WillReturnResult(')

# ============================================================
# 6. Wishlist tests
# ============================================================
fix("Wishlist_Toggle EXISTS",
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).\n\t\tWithArgs(1, 1).')

fix("Wishlist_List",
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, \'\') as primary_image FROM wishlists w JOIN products p ON p.id = w.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE w.user_id = $1 ORDER BY w.created_at DESC`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, \'\') as primary_image FROM wishlists w JOIN products p ON p.id = w.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE w.user_id = $1 ORDER BY w.created_at DESC`).\n\t\tWithArgs(1).')

# ============================================================
# 7. Reviews tests
# ============================================================
fix("Reviews_Create INSERT",
    'mock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) ON CONFLICT (product_id, user_id) DO UPDATE SET rating = $3, comment = $4, updated_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(1, 1, 5, "Great product!").WillReturnResult(',
    'mock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) ON CONFLICT (product_id, user_id) DO UPDATE SET rating = $3, comment = $4, updated_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(1, 1, 5, "Great product!").')

fix("Reviews_List SELECT",
    'mock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.id, u.name, u.email, COALESCE(up.avatar_url, \'\') as avatar_url FROM reviews r JOIN users u ON u.id = r.user_id LEFT JOIN user_profiles up ON up.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC LIMIT $2 OFFSET $3`).\n\t\tWithArgs(1, 20, 0).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.id, u.name, u.email, COALESCE(up.avatar_url, \'\') as avatar_url FROM reviews r JOIN users u ON u.id = r.user_id LEFT JOIN user_profiles up ON up.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC LIMIT $2 OFFSET $3`).\n\t\tWithArgs(1, 20, 0).')

# ============================================================
# 8. RecentlyViewed tests
# ============================================================
fix("RecentlyViewed_Insert",
    'mock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id, viewed_at) VALUES ($1, $2, CURRENT_TIMESTAMP) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(1, 1).WillReturnResult(',
    'mock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id, viewed_at) VALUES ($1, $2, CURRENT_TIMESTAMP) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(1, 1).')

fix("RecentlyViewed_List",
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, \'\') as primary_image, rv.viewed_at FROM recently_viewed rv JOIN products p ON p.id = rv.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE rv.user_id = $1 ORDER BY rv.viewed_at DESC LIMIT $2`).\n\t\tWithArgs(1, 20).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, \'\') as primary_image, rv.viewed_at FROM recently_viewed rv JOIN products p ON p.id = rv.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE rv.user_id = $1 ORDER BY rv.viewed_at DESC LIMIT $2`).\n\t\tWithArgs(1, 20).')

# ============================================================
# 9. Comparison tests
# ============================================================
fix("Comparison_Toggle EXISTS",
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).\n\t\tWithArgs(1, 1).')

fix("Comparison_List",
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.description, COALESCE(p.asset_path, \'\') as asset_path, COALESCE(p.asset_hash, \'\') as asset_hash, COALESCE(p.file_mime_type, \'\') as file_mime_type, COALESCE(created_at::text, \'\') as created_at FROM product_comparison pc JOIN products p ON p.id = pc.product_id WHERE pc.user_id = $1 ORDER BY pc.added_at DESC`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.description, COALESCE(p.asset_path, \'\') as asset_path, COALESCE(p.asset_hash, \'\') as asset_hash, COALESCE(p.file_mime_type, \'\') as file_mime_type, COALESCE(created_at::text, \'\') as created_at FROM product_comparison pc JOIN products p ON p.id = pc.product_id WHERE pc.user_id = $1 ORDER BY pc.added_at DESC`).\n\t\tWithArgs(1).')

# ============================================================
# 10. Recommendations
# ============================================================
fix("Recommendations cat lookup",
    'mock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"category_id"}).AddRow(nil))',
    'mock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"category_id"}).AddRow(nil))')

fix("Recommendations global",
    'mock.ExpectQuery(`SELECT p.id, p.title.*FROM products p.*`).\n\t\tWithArgs(1, 5, 0).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, \'\') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.id != $1 AND p.status = \'active\' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $2 OFFSET $3`).\n\t\tWithArgs(1, 5, 0).\n\t\tWillReturnRows(')

# ============================================================
# 11. GuestOrder
# ============================================================
fix("GuestOrder_Create sum",
    'mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0) FROM products WHERE id = ANY($1::int[]) AND status = \'active\'`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0) FROM products WHERE id = ANY($1::int[]) AND status = \'active\'`).\n\t\tWithArgs([]int{1}).WillReturnRows(')

fix("GuestOrder_Create INSERT",
    'mock.ExpectExec(`INSERT INTO orders (email, total_usd, crypto_chain, status, guest_order) VALUES ($1, $2, $3, \'pending\', true) RETURNING id`).\n\t\tWithArgs(',
    'mock.ExpectExec(`INSERT INTO orders (email, total_usd, crypto_chain, status, guest_order) VALUES ($1, $2, $3, \'pending\', true) RETURNING id`).\n\t\tWithArgs("test@example.com", "10.00", "BSC").WillReturnResult(')

fix("GuestOrder_Check",
    'mock.ExpectQuery(`SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = $1 AND guest_order = true`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = $1 AND guest_order = true`).\n\t\tWithArgs(1).')

# ============================================================
# 12. OrderStatusCheck
# ============================================================
fix("OrderStatusCheck",
    'mock.ExpectQuery(`SELECT id, user_id, status, total_crypto, crypto_chain, payment_address, payment_tx_hash, payment_confirmations FROM orders WHERE id = $1 AND user_id = $2`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT id, user_id, status, total_crypto, crypto_chain, payment_address, payment_tx_hash, payment_confirmations FROM orders WHERE id = $1 AND user_id = $2`).\n\t\tWithArgs(1, 1).')

# ============================================================
# 13. Admin products
# ============================================================
fix("AdminProducts_List COUNT",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE 1=1`).\n\t\tWithArgs().\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE 1=1`).\n\t\tWithArgs().')

fix("AdminProducts_List SELECT",
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, COALESCE(c.name, \'\') as category_name, COALESCE(c.slug, \'\') as category_slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.status, p.stock_quantity, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.asset_path, p.asset_hash, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE 1=1 ORDER BY p.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, COALESCE(c.name, \'\') as category_name, COALESCE(c.slug, \'\') as category_slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.status, p.stock_quantity, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.asset_path, p.asset_hash, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE 1=1 ORDER BY p.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(20, 0).')

# ============================================================
# 14. Admin orders
# ============================================================
fix("AdminOrders_ListAll COUNT",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM orders`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM orders`).\n\t\tWithArgs().')

fix("AdminOrders_ListAll SELECT",
    'mock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.total_crypto, o.crypto_chain, o.payment_address, o.payment_tx_hash, o.payment_confirmations, o.payment_confirmed_at, o.paid_at, o.created_at, o.updated_at FROM orders o ORDER BY o.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.total_crypto, o.crypto_chain, o.payment_address, o.payment_tx_hash, o.payment_confirmations, o.payment_confirmed_at, o.paid_at, o.created_at, o.updated_at FROM orders o ORDER BY o.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(20, 0).')

fix("AdminOrderDetail order",
    'mock.ExpectQuery(`SELECT o.id, o.status.*FROM orders o WHERE o.id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT o.status, CAST(o.total_usd AS NUMERIC), o.guest_email, o.billing_name, o.shipping_address_json, o.notes, o.created_at, o.paid_at FROM orders o WHERE o.id = $1`).\n\t\tWithArgs(1).')

fix("AdminOrderDetail items",
    'mock.ExpectQuery(`SELECT oi.product_id, oi.quantity.*FROM order_items oi WHERE oi.order_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price_usd, oi.price_crypto, oi.download_count, oi.max_downloads, oi.downloaded_at, oi.created_at, p.title, p.slug FROM order_items oi JOIN products p ON oi.product_id = p.id WHERE oi.order_id = $1`).\n\t\tWithArgs(1).')

# ============================================================
# 15. Admin order status UPDATEs
# ============================================================
for status in ['paid', 'completed', 'cancelled', 'refunded']:
    extra = ", paid_at = NOW()" if status == "paid" else ""
    fix(f"Admin_{status} UPDATE",
        f'mock.ExpectExec(`UPDATE orders SET status = \'{status}\'{extra} WHERE id = $1`).\n\t\tWillReturnResult(',
        f'mock.ExpectExec(`UPDATE orders SET status = \'{status}\'{extra} WHERE id = $1`).\n\t\tWithArgs(1).WillReturnResult(')

# ============================================================
# 16. Admin bulk operations
# ============================================================
fix("AdminBulkStatus UPDATE",
    'mock.ExpectExec(`UPDATE products SET status = $1 WHERE id = ANY($2)`).\n\t\tWithArgs(1, 1).\n\t\tWillReturnResult(',
    'mock.ExpectExec(`UPDATE products SET status = $1 WHERE id = ANY($2)`).\n\t\tWithArgs("active", []int{1, 2}).WillReturnResult(')

fix("AdminBulkCategory UPDATE",
    'mock.ExpectExec(`UPDATE products SET category_id = $1 WHERE id = ANY($2)`).\n\t\tWithArgs(1, 1).\n\t\tWillReturnResult(',
    'mock.ExpectExec(`UPDATE products SET category_id = $1 WHERE id = ANY($2)`).\n\t\tWithArgs(1, []int{1, 2}).WillReturnResult(')

# ============================================================
# 17. Admin settings
# ============================================================
fix("AdminSetting_Get",
    'mock.ExpectQuery(`SELECT value FROM settings WHERE key = $1`).\n\t\tWithArgs("payment_address").\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT value FROM settings WHERE key = $1`).\n\t\tWithArgs("payment_address").')

fix("AdminSetting_Set",
    'mock.ExpectExec(`INSERT INTO settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()`).\n\t\tWithArgs(',
    'mock.ExpectExec(`INSERT INTO settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()`).\n\t\tWithArgs("payment_address", "0xNewAddress").WillReturnResult(')

# ============================================================
# 18. Admin exchange rates
# ============================================================
fix("AdminExchangeRateList",
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).\n\t\tWithArgs().\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).\n\t\tWithArgs().')

fix("AdminExchangeRateSet",
    'mock.ExpectExec(`INSERT INTO exchange_rates (chain, symbol, rate_to_usd) VALUES ($1, $2, $3) ON CONFLICT (chain) DO UPDATE SET symbol = $2, rate_to_usd = $3, updated_at = NOW()`).\n\t\tWithArgs(',
    'mock.ExpectExec(`INSERT INTO exchange_rates (chain, symbol, rate_to_usd) VALUES ($1, $2, $3) ON CONFLICT (chain) DO UPDATE SET symbol = $2, rate_to_usd = $3, updated_at = NOW()`).\n\t\tWithArgs("BSC", "BNB", "0.03").WillReturnResult(')

fix("ExchangeRate_GetOne",
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1`).\n\t\tWithArgs("BSC").\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1`).\n\t\tWithArgs("BSC").')

fix("ExchangeRate_GetAll",
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).\n\t\tWithArgs().\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).\n\t\tWithArgs().')

# ============================================================
# 19. Admin community moderation
# ============================================================
fix("AdminCommunityPosts COUNT",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM community_posts WHERE 1=1`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM community_posts WHERE 1=1`).\n\t\tWithArgs().')

fix("AdminCommunityPosts LIST",
    'mock.ExpectQuery(`SELECT cp.id, cp.user_id.*FROM community_posts cp.*`).\n\t\tWithArgs(20, 0).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT cp.id, cp.user_id, u.name, u.email, cp.content, cp.created_at, cp.updated_at FROM community_posts cp LEFT JOIN users u ON u.id = cp.user_id WHERE 1=1 ORDER BY cp.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(20, 0).')

fix("AdminCommunityUsers COUNT",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM users`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM users`).\n\t\tWithArgs().')

fix("AdminCommunityUsers LIST",
    'mock.ExpectQuery(`SELECT u.id, u.name.*FROM users u.*`).\n\t\tWithArgs(20, 0).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT u.id, u.name, u.email, u.created_at, u.updated_at, COALESCE(up.bio, \'\') as bio, COALESCE(up.avatar_url, \'\') as avatar_url, COALESCE(up.wallet_address, \'\') as wallet_address FROM users u LEFT JOIN user_profiles up ON up.user_id = u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(20, 0).')

fix("AdminDeletePost EXISTS",
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM community_posts WHERE id = $1)`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM community_posts WHERE id = $1)`).\n\t\tWithArgs(1).')

# ============================================================
# 20. Admin delete user
# ============================================================
fix("AdminDeleteUser SELECT",
    'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).\n\t\tWithArgs(1).')

# ============================================================
# 21. Admin guest order check
# ============================================================
fix("AdminGuestOrderCheck",
    'mock.ExpectQuery(`SELECT guest_email FROM orders WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT guest_email FROM orders WHERE id = $1`).\n\t\tWithArgs(1).')

# ============================================================
# 22. Admin stats
# ============================================================
fix("AdminStats users",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM users`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM users`).\n\t\tWithArgs().')

fix("AdminStats orders",  
    'mock.ExpectQuery(`SELECT COUNT(*) FROM orders`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM orders`).\n\t\tWithArgs().')

fix("AdminStats revenue",
    'mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(total_usd AS NUMERIC)), 0) FROM orders WHERE status = \'completed\'`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(total_usd AS NUMERIC)), 0) FROM orders WHERE status = \'completed\'`).\n\t\tWithArgs().')

# ============================================================
# 23. Community profile and feed
# ============================================================
fix("Community_Profile user",
    'mock.ExpectQuery(`SELECT id, name FROM users WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, name FROM users WHERE id = $1`).\n\t\tWithArgs(1).')

fix("Community_Profile profile",
    'mock.ExpectQuery(`SELECT bio, avatar_url, wallet_address, created_at FROM user_profiles WHERE user_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT bio, avatar_url, wallet_address, created_at FROM user_profiles WHERE user_id = $1`).\n\t\tWithArgs(1).')

fix("Community_Profile post_count",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM community_posts WHERE user_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM community_posts WHERE user_id = $1`).\n\t\tWithArgs(1).')

fix("Community_Profile follower_count",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM follows WHERE following_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM follows WHERE following_id = $1`).\n\t\tWithArgs(1).')

fix("Community_Profile following_count",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM follows WHERE follower_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM follows WHERE follower_id = $1`).\n\t\tWithArgs(1).')

fix("Community_Feed",
    'mock.ExpectQuery(`SELECT cp.id, cp.user_id.*FROM community_posts cp.*`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url, (SELECT COUNT(*) FROM post_likes WHERE post_id = cp.id) as count_likes, (SELECT COUNT(*) FROM post_comments WHERE post_id = cp.id) as count_comments, EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = cp.user_id) as following FROM community_posts cp JOIN users u ON u.id = cp.user_id LEFT JOIN user_profiles up ON up.user_id = u.id LEFT JOIN post_likes l ON l.post_id = cp.id LEFT JOIN post_comments c ON c.post_id = cp.id WHERE cp.user_id = $1 GROUP BY cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url`).\n\t\tWithArgs(1).')

# ============================================================
# 24. CreateProduct
# ============================================================
fix("CreateProduct slug_check",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n\t\tWithArgs("test-product").\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n\t\tWithArgs("test-product").')

fix("CreateProduct cat_check",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM categories WHERE id = $1`).\n\t\tWithArgs(1).',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM categories WHERE id = $1`).\n\t\tWithArgs(1).')

# ============================================================
# 25. NewProducts
# ============================================================
fix("NewProducts_GetActiveStat",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = $1`).\n\t\tWithArgs("active").\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = $1`).\n\t\tWithArgs("active").')

fix("NewProducts_Delete images",
    'mock.ExpectExec(`DELETE FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(1).',
    'mock.ExpectExec(`DELETE FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(1).')

fix("NewProducts_Delete products",
    'mock.ExpectExec(`DELETE FROM products WHERE id = $1`).\n\t\tWithArgs(1).',
    'mock.ExpectExec(`DELETE FROM products WHERE id = $1`).\n\t\tWithArgs(1).')

# ============================================================
# 26. DBSchema
# ============================================================
fix("DBSchema tables",
    'mock.ExpectQuery(`SELECT table_name FROM information_schema.tables WHERE table_schema = \'public\' AND table_name = $1`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT table_name FROM information_schema.tables WHERE table_schema = \'public\' AND table_name = $1`).\n\t\tWithArgs(')

# ============================================================
# 27. Auth_UpdateProfile
# ============================================================
fix("UpdateProfile SELECT users",
    'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).\n\t\tWithArgs(1).')

fix("UpdateProfile SELECT profile",
    'mock.ExpectQuery(`SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1`).\n\t\tWithArgs(1).')

fix("UpdateProfile UPDATE users",
    'mock.ExpectExec(`UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).\n\t\tWithArgs(',
    'mock.ExpectExec(`UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).\n\t\tWithArgs("New Name", 1).WillReturnResult(')

fix("UpdateProfile INSERT profile",
    'mock.ExpectExec(`INSERT INTO user_profiles (user_id, bio, wallet_address) VALUES ($1, $2, $3) ON CONFLICT (user_id) DO NOTHING`).\n\t\tWithArgs(',
    'mock.ExpectExec(`INSERT INTO user_profiles (user_id, bio, wallet_address) VALUES ($1, $2, $3) ON CONFLICT (user_id) DO NOTHING`).\n\t\tWithArgs(1, "New bio", "0xNewAddress").WillReturnResult(')

# ============================================================
# Write all fixes
# ============================================================
with open(SRC, 'w') as f:
    f.write(text)

print(f"\n{'='*60}")
print(f"Applied {changes} fixes")
print(f"{'='*60}")

# Build
r = subprocess.run(['go', 'build', './handlers/'], capture_output=True, text=True, cwd='/root/project/backend')
print(f"Build: {'OK' if r.returncode == 0 else 'FAIL'}")
if r.returncode:
    print(r.stderr[:500])
    exit(1)

# Full test run
r2 = subprocess.run(['go', 'test', './handlers/', '-count=1', '-v'], capture_output=True, text=True, cwd='/root/project/backend', timeout=120)
lines = (r2.stdout + r2.stderr).split('\n')
passes = sum(1 for l in lines if '--- PASS:' in l)
fails = sum(1 for l in lines if '--- FAIL:' in l)
print(f"\n{'='*60}")
print(f"Test result: {passes} PASS, {fails} FAIL")
print(f"{'='*60}")

# Show summary
for l in lines:
    if '--- FAIL:' in l:
        print(f"  FAIL: {l.strip()}")
    elif 'unfulfilled' in l and 'Expectation' in l:
        print(f"  MOCK: {l.strip()[:140]}")
