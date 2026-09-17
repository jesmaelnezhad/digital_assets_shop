#!/usr/bin/env python3
"""Fix all remaining mock SQL mismatches in all_features_test.go."""
import re

with open("/root/project/backend/handlers/all_features_test.go", "r") as f:
    text = f.read()

lines = text.split('\n')
out = []

for line in lines:
    # Fix 1: COUNT(\\*) -> COUNT(\*) 
    line = line.replace('COUNT(\\\\*)', 'COUNT(\\*)')
    
    # Fix 2: SearchProducts product query
    if 'mock.ExpectQuery(`SELECT p.id, p.title.*`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT p.id, p.title.*`).',
            'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\') as name, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = \$1 AND p.search_vector @@ plainto_tsquery(\'english\', \$2) ORDER BY p.created_at DESC LIMIT \$3 OFFSET \$4`).'
        )
    
    # Fix 3: Admin_CommunityPostList query
    if 'mock.ExpectQuery(`SELECT cp.id, cp.user_id.*FROM community_posts cp.*`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT cp.id, cp.user_id.*FROM community_posts cp.*`).',
            'mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, COALESCE(up.avatar_url, \'\') as avatar_url, (SELECT COUNT(*) FROM community_likes WHERE post_id = cp.id) as count_likes, (SELECT COUNT(*) FROM community_comments WHERE post_id = cp.id) as count_comments, EXISTS(SELECT 1 FROM community_follows WHERE follower_id = \$2 AND following_id = cp.user_id) as following FROM community_posts cp JOIN users u ON u.id = cp.user_id LEFT JOIN user_profiles up ON up.user_id = u.id ORDER BY cp.created_at DESC LIMIT \$3 OFFSET \$4`).'
        )
    
    # Fix 4: Admin_CommunityUserList query
    if 'mock.ExpectQuery(`SELECT u.id, u.email, u.name.*`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT u.id, u.email, u.name.*`).',
            'mock.ExpectQuery(`SELECT u.id, u.email, u.name, COALESCE(up.bio, \'\') as bio, COALESCE(up.avatar_url, \'\') as avatar_url, (SELECT COUNT(*) FROM community_follows WHERE follower_id = u.id) as followers_count, (SELECT COUNT(*) FROM community_follows WHERE following_id = u.id) as following_count, (SELECT COUNT(*) FROM community_posts WHERE user_id = u.id) as posts_count FROM users u LEFT JOIN user_profiles up ON up.user_id = u.id ORDER BY u.created_at DESC LIMIT \$1 OFFSET \$2`).'
        )
    
    # Fix 5: Admin_GuestOrderCheck query
    if 'mock.ExpectQuery(`SELECT o.id, o.status.*FROM orders WHERE id = \$1`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT o.id, o.status.*FROM orders WHERE id = \$1`).',
            'mock.ExpectQuery(`SELECT id, email, status, total_usd, crypto_chain, payment_tx_hash, payment_confirmations, paid_at, created_at FROM orders WHERE id = \$1 AND guest_order = true`).'
        )
    
    # Fix 6: Admin_GetSetting key -> setting_key
    if 'mock.ExpectQuery(`SELECT value FROM settings WHERE key = \$1`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT value FROM settings WHERE key = \$1`).',
            'mock.ExpectQuery(`SELECT value FROM settings WHERE setting_key = \$1`).'
        )
    
    # Fix 7: Admin_Stats COALESCE query
    if 'mock.ExpectQuery(`SELECT COALESCE.*`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT COALESCE.*`).',
            'mock.ExpectQuery(`SELECT COALESCE(SUM(total_usd)::TEXT, \'0.00\') FROM orders WHERE status = \'completed\'`).'
        )
    
    # Fix 8: Cart_AddItem - add stock check before INSERT
    if 'mock.ExpectExec(`INSERT INTO cart_items.*`).WithArgs(1, 1, 3)' in line:
        line = line.replace(
            'mock.ExpectExec(`INSERT INTO cart_items.*`).WithArgs(1, 1, 3)',
            'mock.ExpectQuery(`SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id = \$1`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"stock_quantity", "CAST(price_usd AS INTEGER)"}).AddRow(100, 10))\nmock.ExpectExec(`INSERT INTO cart_items (cart_id, product_id, quantity) VALUES (\$1, \$2, \$3) ON CONFLICT (cart_id, product_id) DO UPDATE SET quantity = \$3, added_at = CURRENT_TIMESTAMP`).WithArgs(1, 1, 3)'
        )
    
    # Fix 9: Wishlist_Toggle EXISTS
    if 'mock.ExpectQuery(`SELECT EXISTS\\(SELECT 1 FROM wishlists WHERE user_id = \$1 AND product_id = \$2\\)`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT EXISTS\\(SELECT 1 FROM wishlists WHERE user_id = \$1 AND product_id = \$2\\)`).',
            'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = \$1 AND product_id = \$2)`).'
        )
    
    # Fix 10: Wishlist_List query
    if 'mock.ExpectQuery(`SELECT w.product_id, p.name, p.slug.*`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT w.product_id, p.name, p.slug.*`).',
            'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, \'\') as primary_image FROM wishlists w JOIN products p ON p.id = w.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE w.user_id = \$1 ORDER BY w.created_at DESC`).'
        )
    
    # Fix 11: Reviews_Create - add product check
    if 'mock.ExpectExec(`INSERT INTO reviews.*`).WithArgs(1, 1, 5, "Great product!")' in line:
        line = line.replace(
            'mock.ExpectExec(`INSERT INTO reviews.*`).WithArgs(1, 1, 5, "Great product!")',
            'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM products WHERE id = \$1)`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))\nmock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES (\$1, \$2, \$3, \$4) ON CONFLICT (product_id, user_id) DO UPDATE SET rating = \$3, comment = \$4, updated_at = CURRENT_TIMESTAMP`).WithArgs(1, 1, 5, "Great product!")'
        )
    
    # Fix 12: Reviews_List - add COUNT + fix query
    if 'mock.ExpectQuery(`SELECT r.id, r.rating, r.comment.*`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT r.id, r.rating, r.comment.*`).',
            'mock.ExpectQuery(`SELECT COUNT(*) FROM reviews WHERE product_id = \$1`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))\nmock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.id, u.name, u.email, COALESCE(up.avatar_url, \'\') as avatar_url FROM reviews r JOIN users u ON u.id = r.user_id LEFT JOIN user_profiles up ON up.user_id = u.id WHERE r.product_id = \$1 ORDER BY r.created_at DESC LIMIT \$2 OFFSET \$3`).'
        )
    
    # Fix 13: RecentlyViewed_Record INSERT
    if 'mock.ExpectExec(`INSERT INTO recently_viewed.*`).' in line:
        line = line.replace(
            'mock.ExpectExec(`INSERT INTO recently_viewed.*`).',
            'mock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id, viewed_at) VALUES (\$1, \$2, CURRENT_TIMESTAMP) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = CURRENT_TIMESTAMP`).'
        )
    
    # Fix 14: RecentlyViewed_List query (context: inside RecentlyViewed_List function)
    if 'mock.ExpectQuery(`SELECT p.id, p.name, p.slug.*`).' in line:
        # Check if we're in RecentlyViewed_List or Comparison_List by looking at context
        ctx_start = max(0, lines.index(line) - 10)
        ctx = '\n'.join(lines[ctx_start:lines.index(line)])
        if 'Recently' in ctx:
            line = line.replace(
                'mock.ExpectQuery(`SELECT p.id, p.name, p.slug.*`).',
                'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, \'\') as primary_image, rv.viewed_at FROM recently_viewed rv JOIN products p ON p.id = rv.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE rv.user_id = \$1 ORDER BY rv.viewed_at DESC LIMIT \$2`).'
            )
        elif 'Comparison' in ctx or 'Comparison_List' in ctx:
            line = line.replace(
                'mock.ExpectQuery(`SELECT p.id, p.name, p.slug.*`).',
                'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.description, COALESCE(p.asset_path, \'\') as asset_path, COALESCE(p.asset_hash, \'\') as asset_hash, COALESCE(p.file_mime_type, \'\') as file_mime_type, COALESCE(created_at::text, \'\') as created_at FROM product_comparison pc JOIN products p ON p.id = pc.product_id WHERE pc.user_id = \$1 ORDER BY pc.added_at DESC`).'
            )
    
    # Fix 15: Comparison_Toggle EXISTS
    if 'mock.ExpectQuery(`SELECT EXISTS\\(SELECT 1 FROM product_comparison WHERE user_id = \$1 AND product_id = \$2\\)`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT EXISTS\\(SELECT 1 FROM product_comparison WHERE user_id = \$1 AND product_id = \$2\\)`).',
            'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = \$1 AND product_id = \$2)`).'
        )
    
    # Fix 16: GuestOrder_Create - fix INSERT (orders not guest_orders) + add SUM
    if 'mock.ExpectExec(`INSERT INTO guest_orders.*`).' in line:
        line = line.replace(
            'mock.ExpectExec(`INSERT INTO guest_orders.*`).',
            'mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0) FROM products WHERE id = ANY(\$1::int[]) AND status = \'active\'`).WithArgs([]int{1}).WillReturnRows(sqlmock.NewRows([]string{"COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0)"}).AddRow("10.00"))\nmock.ExpectExec(`INSERT INTO orders (email, total_usd, crypto_chain, status, guest_order) VALUES (\$1, \$2, \$3, \'pending\', true) RETURNING id`).'
        )
    
    # Fix 17: Guest order SELECT queries (both Create and Check)
    if 'mock.ExpectQuery(`SELECT id, email, status.*FROM guest_orders WHERE id = \$1`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT id, email, status.*FROM guest_orders WHERE id = \$1`).',
            'mock.ExpectQuery(`SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = \$1 AND guest_order = true`).'
        )
    
    # Fix 18: OrderStatusCheck query
    if 'mock.ExpectQuery(`SELECT id, status, total_usd.*FROM orders WHERE id = \$1`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT id, status, total_usd.*FROM orders WHERE id = \$1`).',
            'mock.ExpectQuery(`SELECT id, status, total_usd, crypto_chain, payment_tx_hash, payment_confirmations, paid_at FROM orders WHERE id = \$1`).'
        )
    
    # Fix 19: Admin_BulkProductStatus - per-ID updates
    if 'mock.ExpectExec(`UPDATE products SET.*`).WillReturnResult(sqlmock.NewResult(2, 2))' in line:
        line = line.replace(
            'mock.ExpectExec(`UPDATE products SET.*`).WillReturnResult(sqlmock.NewResult(2, 2))',
            'mock.ExpectExec(`UPDATE products SET status = \$1 WHERE id = \$2`).WithArgs("active", 1).WillReturnResult(sqlmock.NewResult(1, 1))\nmock.ExpectExec(`UPDATE products SET status = \$1 WHERE id = \$2`).WithArgs("active", 2).WillReturnResult(sqlmock.NewResult(1, 1))'
        )
    
    # Fix 20: Admin order status transitions (paid, completed, cancelled, refunded)
    for new_status, prev_status in [('paid', 'pending'), ('completed', 'paid'), ('cancelled', 'pending'), ('refunded', 'paid')]:
        old = 'mock.ExpectExec(`UPDATE orders SET.*`).WillReturnResult(sqlmock.NewResult(1, 1))\n\t\tcode, _, result := R("PUT", "/api/v1/admin/orders/1/status/' + new_status + '"'
        new = ('mock.ExpectQuery(`SELECT status FROM orders WHERE id = $1`).WithArgs(1).'
               'WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("' + prev_status + '"))\n\t\t'
               'mock.ExpectExec(`UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).'
               'WithArgs("' + new_status + '", 1).WillReturnResult(sqlmock.NewResult(1, 1))\n\t\t'
               'code, _, result := R("PUT", "/api/v1/admin/orders/1/status/' + new_status + '"')
        line = line.replace(old, new)
    
    # Fix 21: Admin_BulkProductCategory
    if 'mock.ExpectExec(`UPDATE products SET.*`).WillReturnResult(sqlmock.NewResult(1, 1))' in line and 'category_id": "2"' in ''.join(lines[max(0,lines.index(line)-3):lines.index(line)+1]):
        line = line.replace(
            'mock.ExpectExec(`UPDATE products SET.*`).WillReturnResult(sqlmock.NewResult(1, 1))',
            'mock.ExpectExec(`UPDATE products SET category_id = \$1 WHERE id = \$2`).WithArgs(2, 1).WillReturnResult(sqlmock.NewResult(1, 1))'
        )
    
    # Fix 22: Admin_BulkProductToggle
    if 'mock.ExpectExec(`UPDATE products SET.*`).WithArgs("active", 1).WillReturnResult(sqlmock.NewResult(1, 1))' in line:
        line = line.replace(
            'mock.ExpectExec(`UPDATE products SET.*`).WithArgs("active", 1).WillReturnResult(sqlmock.NewResult(1, 1))',
            'mock.ExpectExec(`UPDATE products SET status = \$1 WHERE id = \$2`).WithArgs("active", 1).WillReturnResult(sqlmock.NewResult(1, 1))'
        )
    
    # Fix 23: Admin_ExchangeRateSet INSERT
    if 'mock.ExpectExec(`INSERT INTO exchange_rates.*`).' in line:
        line = line.replace(
            'mock.ExpectExec(`INSERT INTO exchange_rates.*`).',
            'mock.ExpectExec(`INSERT INTO exchange_rates (chain, rate) VALUES (\$1, \$2) ON CONFLICT (chain) DO UPDATE SET rate = \$2, updated_at = CURRENT_TIMESTAMP`).'
        )
    
    # Fix 24: Admin_SetSetting INSERT
    if 'mock.ExpectExec(`INSERT INTO settings.*`).' in line:
        line = line.replace(
            'mock.ExpectExec(`INSERT INTO settings.*`).',
            'mock.ExpectExec(`INSERT INTO settings (setting_key, value) VALUES (\$1, \$2) ON CONFLICT (setting_key) DO UPDATE SET value = \$2, updated_at = CURRENT_TIMESTAMP`).'
        )
    
    # Fix 25: Settings_GetAll key -> setting_key
    if 'mock.ExpectQuery(`SELECT key, value, updated_at FROM settings ORDER BY key`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT key, value, updated_at FROM settings ORDER BY key`).',
            'mock.ExpectQuery(`SELECT setting_key, value, updated_at FROM settings ORDER BY setting_key`).'
        )
    
    # Fix 26: NewProducts_GetActiveStat COUNT queries
    if 'mock.ExpectQuery(`SELECT COUNT\\(\\\\*) FROM products WHERE status = \'active\'`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT COUNT\\(\\\\*) FROM products WHERE status = \'active\'`).',
            'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = \'active\'`).'
        )
    if 'mock.ExpectQuery(`SELECT COUNT\\(\\\\*) FROM products`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT COUNT\\(\\\\*) FROM products`).',
            'mock.ExpectQuery(`SELECT COUNT(*) FROM products`).'
        )
    if 'mock.ExpectQuery(`SELECT COUNT\\(\\\\*) FROM categories`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT COUNT\\(\\\\*) FROM categories`).',
            'mock.ExpectQuery(`SELECT COUNT(*) FROM categories`).'
        )
    
    # Fix 27: CreateProduct category check
    if 'mock.ExpectQuery(`SELECT COUNT\\(\\\\*) FROM categories WHERE id = \$1`).' in line:
        line = line.replace(
            'mock.ExpectQuery(`SELECT COUNT\\(\\\\*) FROM categories WHERE id = \$1`).',
            'mock.ExpectQuery(`SELECT COUNT(*) FROM categories WHERE id = \$1`).'
        )
    
    # Fix 28: UpdateProfile - profile INSERT already handled
    
    out.append(line)

result = '\n'.join(out)

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(result)

print("Done. Line count:", len(out))
