#!/usr/bin/env python3
"""Fix all mock SQL mismatches - comprehensive fix based on handler analysis."""
with open("/root/project/backend/handlers/all_features_test.go", "r") as f:
    text = f.read()

lines = text.split('\n')
out = []

i = 0
while i < len(lines):
    line = lines[i]
    
    # Fix COUNT(\\\\*) -> COUNT(\\*)  (the v5 script over-escaped these)
    line = line.replace('COUNT(\\\\\\\\*)', 'COUNT(\\\\*)')
    
    # Fix 1: Products_GetProducts - COUNT + product list queries
    if i+1 < len(lines) and 'func TestProducts_GetProducts' in lines[i]:
        # Fix the COUNT query (line i+3 = line 123)
        if i+3 < len(lines) and 'COUNT' in lines[i+3]:
            lines[i+3] = lines[i+3].replace(
                'mock.ExpectQuery(`SELECT COUNT(\\\\*) FROM products.*`).',
                'mock.ExpectQuery(`SELECT COUNT(\\*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).'
            )
            # Add WithArgs
            if i+4 < len(lines) and 'WillReturnRows' in lines[i+4]:
                lines[i+4] = lines[i+4].replace(
                    '\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))',
                    '\t\tWithArgs("active").\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))'
                )
        # Fix the product list query (line i+4 or i+5)
        for j in range(i, min(i+10, len(lines))):
            if 'mock.ExpectQuery(`SELECT p.id, p.title.*`).' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT p.id, p.title.*`).',
                    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\') as name, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).'
                )
                # Add WithArgs if missing
                if j+1 < len(lines) and 'WillReturnRows' in lines[j+1] and 'WithArgs' not in lines[j+1]:
                    lines[j+1] = lines[j+1].replace(
                        '\t\tWillReturnRows',
                        '\t\tWithArgs("active", 12, 0).\n\t\tWillReturnRows'
                    )
                break
    
    # Fix 2: GetProductBySlug
    if i+1 < len(lines) and 'func TestProducts_GetProductBySlug' in lines[i]:
        for j in range(i, min(i+15, len(lines))):
            if 'COUNT.*products WHERE slug' in lines[j]:
                lines[j] = lines[j].replace(
                    'AddRow(0)',
                    'AddRow(1)'
                ).replace(
                    'mock.ExpectQuery(`SELECT COUNT(\\\\*) FROM products WHERE slug = $1`).',
                    'mock.ExpectQuery(`SELECT COUNT(\\*) FROM products WHERE slug = $1`).'
                )
            if 'mock.ExpectQuery(`SELECT p.id, p.title.*`).' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT p.id, p.title.*`).',
                    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\') as name, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.slug = $1 AND p.status = \'"\'active\'"\'`).'
                )
    
    # Fix 3: SearchProducts
    if i+1 < len(lines) and 'func TestProducts_SearchProducts' in lines[i]:
        for j in range(i, min(i+15, len(lines))):
            if 'COUNT.*products.*' in lines[j] and 'AddRow(1)' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT COUNT(\\\\*) FROM products.*`).',
                    'mock.ExpectQuery(`SELECT COUNT(\\*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND p.search_vector @@ plainto_tsquery(\'english\', $2)`).'
                )
                if j+1 < len(lines) and 'WillReturnRows' in lines[j+1]:
                    lines[j+1] = lines[j+1].replace(
                        '\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))',
                        '\t\tWithArgs("active", "%test%").\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))'
                    )
            if 'mock.ExpectQuery(`SELECT p.id, p.title.*`).' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT p.id, p.title.*`).',
                    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\') as name, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND p.search_vector @@ plainto_tsquery(\'english\', $2) ORDER BY p.created_at DESC LIMIT $3 OFFSET $4`).'
                )
    
    # Fix 4: Admin order status transitions - need per-status handling
    if 'func TestAdmin_OrderStatusPaid' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'mock.ExpectExec(`UPDATE orders SET.*`).WillReturnResult(sqlmock.NewResult(1, 1))' in lines[j]:
                lines[j] = (lines[j]
                    .replace('mock.ExpectExec(`UPDATE orders SET.*`).WillReturnResult(sqlmock.NewResult(1, 1))',
                             'mock.ExpectQuery(`SELECT status FROM orders WHERE id = $1`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("pending"))\n\t\tmock.ExpectExec(`UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).WithArgs("paid", 1).WillReturnResult(sqlmock.NewResult(1, 1))'))
                break
    
    if 'func TestAdmin_OrderStatusCompleted' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'mock.ExpectExec(`UPDATE orders SET.*`).WillReturnResult(sqlmock.NewResult(1, 1))' in lines[j]:
                lines[j] = (lines[j]
                    .replace('mock.ExpectExec(`UPDATE orders SET.*`).WillReturnResult(sqlmock.NewResult(1, 1))',
                             'mock.ExpectQuery(`SELECT status FROM orders WHERE id = $1`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("paid"))\n\t\tmock.ExpectExec(`UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).WithArgs("completed", 1).WillReturnResult(sqlmock.NewResult(1, 1))'))
                break
    
    if 'func TestAdmin_OrderStatusCancelled' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'mock.ExpectExec(`UPDATE orders SET.*`).WillReturnResult(sqlmock.NewResult(1, 1))' in lines[j]:
                lines[j] = (lines[j]
                    .replace('mock.ExpectExec(`UPDATE orders SET.*`).WillReturnResult(sqlmock.NewResult(1, 1))',
                             'mock.ExpectQuery(`SELECT status FROM orders WHERE id = $1`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("pending"))\n\t\tmock.ExpectExec(`UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).WithArgs("cancelled", 1).WillReturnResult(sqlmock.NewResult(1, 1))'))
                break
    
    if 'func TestAdmin_OrderStatusRefunded' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'mock.ExpectExec(`UPDATE orders SET.*`).WillReturnResult(sqlmock.NewResult(1, 1))' in lines[j]:
                lines[j] = (lines[j]
                    .replace('mock.ExpectExec(`UPDATE orders SET.*`).WillReturnResult(sqlmock.NewResult(1, 1))',
                             'mock.ExpectQuery(`SELECT status FROM orders WHERE id = $1`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("paid"))\n\t\tmock.ExpectExec(`UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).WithArgs("refunded", 1).WillReturnResult(sqlmock.NewResult(1, 1))'))
                break
    
    # Fix 5: Admin_BulkProductStatus - per-ID updates
    if 'func TestAdmin_BulkProductStatus' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'mock.ExpectExec(`UPDATE products SET.*`).WillReturnResult(sqlmock.NewResult(2, 2))' in lines[j]:
                lines[j] = (lines[j]
                    .replace('mock.ExpectExec(`UPDATE products SET.*`).WillReturnResult(sqlmock.NewResult(2, 2))',
                             'mock.ExpectExec(`UPDATE products SET status = $1 WHERE id = $2`).WithArgs("active", 1).WillReturnResult(sqlmock.NewResult(1, 1))\n\t\tmock.ExpectExec(`UPDATE products SET status = $1 WHERE id = $2`).WithArgs("active", 2).WillReturnResult(sqlmock.NewResult(1, 1))'))
                break
    
    # Fix 6: Admin_BulkProductCategory
    if 'func TestAdmin_BulkProductCategory' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'mock.ExpectExec(`UPDATE products SET.*`).WillReturnResult(sqlmock.NewResult(1, 1))' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectExec(`UPDATE products SET.*`).WillReturnResult(sqlmock.NewResult(1, 1))',
                    'mock.ExpectExec(`UPDATE products SET category_id = $1 WHERE id = $2`).WithArgs(2, 1).WillReturnResult(sqlmock.NewResult(1, 1))'
                )
                break
    
    # Fix 7: Admin_BulkProductToggle
    if 'func TestAdmin_BulkProductToggle' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'mock.ExpectExec(`UPDATE products SET.*`).WithArgs("active", 1).WillReturnResult' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectExec(`UPDATE products SET.*`).WithArgs("active", 1).WillReturnResult(sqlmock.NewResult(1, 1))',
                    'mock.ExpectExec(`UPDATE products SET status = $1 WHERE id = $2`).WithArgs("active", 1).WillReturnResult(sqlmock.NewResult(1, 1))'
                )
                break
    
    # Fix 8: Admin_CreateProduct_Unauthorized need product check mocks
    # Actually looking at Auth_CreateProduct_Unauthorized - handler checks auth first, returns 401 before DB
    # So this should pass. The issue might be that auth middleware does a DB query.
    # Let me check... the middleware queries users table. The test doesn't mock that.
    # But other auth tests (Login, Register) work. Let me check what's different.
    # Actually TestAuth_CreateProduct_Unauthorized sends no auth header, so middleware returns 401
    # without querying DB. The mock should not be needed. Let me check the actual error.
    
    # Fix 9: Cart_AddItem - needs stock check
    if 'func TestCart_AddItem' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'mock.ExpectExec(`INSERT INTO cart_items.*`).WithArgs(1, 1, 3)' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectExec(`INSERT INTO cart_items.*`).WithArgs(1, 1, 3).WillReturnResult(sqlmock.NewResult(1, 1))',
                    'mock.ExpectQuery(`SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id = $1`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"stock_quantity", "CAST(price_usd AS INTEGER)"}).AddRow(100, 10))\n\t\tmock.ExpectExec(`INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, $3) ON CONFLICT (cart_id, product_id) DO UPDATE SET quantity = $3, added_at = CURRENT_TIMESTAMP`).WithArgs(1, 1, 3).WillReturnResult(sqlmock.NewResult(1, 1))'
                )
                break
    
    # Fix 10: Settings_GetAll - key -> setting_key
    if 'func TestSettings_GetAll' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'SELECT key, value, updated_at FROM settings' in lines[j]:
                lines[j] = lines[j].replace(
                    'SELECT key, value, updated_at FROM settings ORDER BY key',
                    'SELECT setting_key, value, updated_at FROM settings ORDER BY setting_key'
                ).replace(
                    'sqlmock.NewRows([]string{"key", "value", "updated_at"})',
                    'sqlmock.NewRows([]string{"setting_key", "value", "updated_at"})'
                )
                break
    
    # Fix 11: Admin_GetSetting - key -> setting_key
    if 'func TestAdmin_GetSetting' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'SELECT value FROM settings WHERE key = $1' in lines[j]:
                lines[j] = lines[j].replace(
                    'SELECT value FROM settings WHERE key = $1',
                    'SELECT value FROM settings WHERE setting_key = $1'
                )
                break
    
    # Fix 12: Admin_Stats
    if 'func TestAdmin_Stats' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'COUNT.*users' in lines[j]:
                lines[j] = lines[j].replace('COUNT(\\\\*)', 'COUNT(\\*)')
            if 'COUNT.*orders' in lines[j]:
                lines[j] = lines[j].replace('COUNT(\\\\*)', 'COUNT(\\*)')
            if 'SELECT COALESCE.*' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT COALESCE.*`).',
                    'mock.ExpectQuery(`SELECT COALESCE(SUM(total_usd)::TEXT, \'0.00\') FROM orders WHERE status = \'completed\'`).'
                ).replace(
                    'sqlmock.NewRows([]string{"total"}).AddRow("100.00")',
                    'sqlmock.NewRows([]string{"COALESCE(SUM(total_usd)::TEXT, \'0.00\')"}).AddRow("100.00")'
                )
    
    # Fix 13: ExchangeRates
    if 'func TestExchangeRates_GetAll' in line or 'func TestExchangeRates_GetOne' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'COUNT' in lines[j]:
                lines[j] = lines[j].replace('COUNT(\\\\*)', 'COUNT(\\*)')
    
    # Fix 14: NewProducts_GetActiveStat
    if 'func TestNewProducts_GetActiveStat' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'COUNT' in lines[j]:
                lines[j] = lines[j].replace('COUNT(\\\\*)', 'COUNT(\\*)')
    
    # Fix 15: DBSchema
    if 'func TestDBSchema_Migration011Tables' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'COUNT' in lines[j]:
                lines[j] = lines[j].replace('COUNT(\\\\*)', 'COUNT(\\*)')
    
    # Fix 16: Admin_DeleteUser - fix escaping
    if 'func TestAdmin_DeleteUser' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'DELETE FROM users WHERE id = $1' in lines[j]:
                lines[j] = lines[j].replace('DELETE FROM users WHERE id = $1', 'DELETE FROM users WHERE id = $1')
    
    # Fix 17: Admin_DeleteCommunityPost
    if 'func TestAdmin_DeleteCommunityPost' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'DELETE FROM community_posts WHERE id = $1' in lines[j]:
                pass  # Already correct
    
    # Fix 18: Admin_CommunityPostList
    if 'func TestAdmin_CommunityPostList' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'SELECT cp.id, cp.user_id.*FROM community_posts cp.*' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT cp.id, cp.user_id.*FROM community_posts cp.*`).',
                    'mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, COALESCE(up.avatar_url, \'\') as avatar_url, (SELECT COUNT(*) FROM community_likes WHERE post_id = cp.id) as count_likes, (SELECT COUNT(*) FROM community_comments WHERE post_id = cp.id) as count_comments, EXISTS(SELECT 1 FROM community_follows WHERE follower_id = $2 AND following_id = cp.user_id) as following FROM community_posts cp JOIN users u ON u.id = cp.user_id LEFT JOIN user_profiles up ON up.user_id = u.id ORDER BY cp.created_at DESC LIMIT $3 OFFSET $4`).'
                )
                break
    
    # Fix 19: Admin_CommunityUserList
    if 'func TestAdmin_CommunityUserList' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'SELECT COUNT.*users' in lines[j]:
                lines[j] = lines[j].replace('COUNT(\\\\*)', 'COUNT(\\*)')
            if 'SELECT u.id, u.email, u.name.*' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT u.id, u.email, u.name.*`).',
                    'mock.ExpectQuery(`SELECT u.id, u.email, u.name, COALESCE(up.bio, \'\') as bio, COALESCE(up.avatar_url, \'\') as avatar_url, (SELECT COUNT(*) FROM community_follows WHERE follower_id = u.id) as followers_count, (SELECT COUNT(*) FROM community_follows WHERE following_id = u.id) as following_count, (SELECT COUNT(*) FROM community_posts WHERE user_id = u.id) as posts_count FROM users u LEFT JOIN user_profiles up ON up.user_id = u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`).'
                ).replace(
                    'sqlmock.NewRows([]string{"id", "email", "name", "count"})',
                    'sqlmock.NewRows([]string{"id", "email", "name", "bio", "avatar_url", "followers_count", "following_count", "posts_count"})'
                ).replace(
                    'AddRow(1, "test@test.com", "Test User", 10).',
                    'AddRow(1, "test@test.com", "Test User", "", "", 0, 0, 5).'
                ).replace(
                    'AddRow(2, "user2@test.com", "User 2", 5)',
                    'AddRow(2, "user2@test.com", "User 2", "", "", 0, 0, 3)'
                )
    
    # Fix 20: Admin_GuestOrderCheck
    if 'func TestAdmin_GuestOrderCheck' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'SELECT o.id, o.status.*FROM orders WHERE id = $1' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT o.id, o.status.*FROM orders WHERE id = $1`).',
                    'mock.ExpectQuery(`SELECT id, email, status, total_usd, crypto_chain, payment_tx_hash, payment_confirmations, paid_at, created_at FROM orders WHERE id = $1 AND guest_order = true`).'
                ).replace(
                    'sqlmock.NewRows([]string{"id", "status", "total_usd", "created_at"})',
                    'sqlmock.NewRows([]string{"id", "email", "status", "total_usd", "crypto_chain", "payment_tx_hash", "payment_confirmations", "paid_at", "created_at"})'
                ).replace(
                    'AddRow(1, "completed", "10.00", mockTime(2024, time.January, 1))',
                    'AddRow(1, "test@example.com", "completed", "10.00", "BSC", "", 0, mockTime(2024, time.January, 1), mockTime(2024, time.January, 1))'
                )
    
    # Fix 21: Admin_ExchangeRateList - looks OK already
    # Fix 22: Admin_ExchangeRateSet
    if 'func TestAdmin_ExchangeRateSet' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'INSERT INTO exchange_rates.*' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectExec(`INSERT INTO exchange_rates.*`).',
                    'mock.ExpectExec(`INSERT INTO exchange_rates (chain, rate) VALUES ($1, $2) ON CONFLICT (chain) DO UPDATE SET rate = $2, updated_at = CURRENT_TIMESTAMP`).'
                )
                break
    
    # Fix 23: Admin_SetSetting
    if 'func TestAdmin_SetSetting' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'INSERT INTO settings.*' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectExec(`INSERT INTO settings.*`).',
                    'mock.ExpectExec(`INSERT INTO settings (setting_key, value) VALUES ($1, $2) ON CONFLICT (setting_key) DO UPDATE SET value = $2, updated_at = CURRENT_TIMESTAMP`).'
                )
                break
    
    # Fix 24: Wishlist_Toggle
    if 'func TestWishlist_Toggle' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'SELECT EXISTS.*wishlists' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT EXISTS\\(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2\\)`).',
                    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).'
                )
            if 'INSERT INTO wishlists.*' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectExec(`INSERT INTO wishlists.*`).',
                    'mock.ExpectExec(`INSERT INTO wishlists (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).'
                )
                break
    
    # Fix 25: Wishlist_List
    if 'func TestWishlist_List' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'SELECT w.product_id, p.name, p.slug.*' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT w.product_id, p.name, p.slug.*`).',
                    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, \'\') as primary_image FROM wishlists w JOIN products p ON p.id = w.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE w.user_id = $1 ORDER BY w.created_at DESC`).'
                ).replace(
                    'sqlmock.NewRows([]string{"product_id", "name", "slug", "price_usd", "image_url"})',
                    'sqlmock.NewRows([]string{"id", "title", "slug", "price_usd", "stock_quantity", "primary_image"})'
                ).replace(
                    'AddRow(1, "Test Product", "test-product", "10.00", "http://example.com/img.png")',
                    'AddRow(1, "Test Product", "test-product", "10.00", 100, "http://example.com/img.png")'
                )
    
    # Fix 26: Reviews_Create
    if 'func TestReviews_Create' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'INSERT INTO reviews.*' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectExec(`INSERT INTO reviews.*`).WithArgs(1, 1, 5, "Great product!").WillReturnResult(sqlmock.NewResult(1, 1))',
                    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))\n\t\tmock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) ON CONFLICT (product_id, user_id) DO UPDATE SET rating = $3, comment = $4, updated_at = CURRENT_TIMESTAMP`).WithArgs(1, 1, 5, "Great product!").WillReturnResult(sqlmock.NewResult(1, 1))'
                )
                break
    
    # Fix 27: Reviews_List
    if 'func TestReviews_List' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'SELECT r.id, r.rating, r.comment.*' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT r.id, r.rating, r.comment.*`).',
                    'mock.ExpectQuery(`SELECT COUNT(*) FROM reviews WHERE product_id = $1`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))\n\t\tmock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.id, u.name, u.email, COALESCE(up.avatar_url, \'\') as avatar_url FROM reviews r JOIN users u ON u.id = r.user_id LEFT JOIN user_profiles up ON up.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC LIMIT $2 OFFSET $3`).'
                ).replace(
                    'sqlmock.NewRows([]string{"id", "product_id", "rating", "comment", "created_at", "email", "name"})',
                    'sqlmock.NewRows([]string{"id", "rating", "comment", "created_at", "id", "name", "email", "avatar_url"})'
                ).replace(
                    'AddRow(1, 1, 5, "Great!", mockTime(2024, time.January, 1), "test@test.com", "Test User")',
                    'AddRow(1, 5, "Great!", mockTime(2024, time.January, 1), 1, "Test User", "test@test.com", "")'
                )
    
    # Fix 28: RecentlyViewed_Record
    if 'func TestRecentlyViewed_Record' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'INSERT INTO recently_viewed.*' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectExec(`INSERT INTO recently_viewed.*`).',
                    'mock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id, viewed_at) VALUES ($1, $2, CURRENT_TIMESTAMP) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = CURRENT_TIMESTAMP`).'
                )
                break
    
    # Fix 29: RecentlyViewed_List
    if 'func TestRecentlyViewed_List' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'SELECT p.id, p.name, p.slug.*' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT p.id, p.name, p.slug.*`).',
                    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, \'\') as primary_image, rv.viewed_at FROM recently_viewed rv JOIN products p ON p.id = rv.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE rv.user_id = $1 ORDER BY rv.viewed_at DESC LIMIT $2`).'
                ).replace(
                    'sqlmock.NewRows([]string{"id", "name", "slug", "price_usd", "description"})',
                    'sqlmock.NewRows([]string{"id", "title", "slug", "price_usd", "stock_quantity", "primary_image", "viewed_at"})'
                ).replace(
                    'AddRow(1, "Test Product", "test-product", "10.00", "A test product")',
                    'AddRow(1, "Test Product", "test-product", "10.00", 100, "http://example.com/img.png", mockTime(2024, time.January, 1))'
                )
    
    # Fix 30: Comparison_Toggle
    if 'func TestComparison_Toggle' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'SELECT EXISTS.*product_comparison' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT EXISTS\\(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2\\)`).',
                    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).'
                )
            if 'INSERT INTO product_comparison.*' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectExec(`INSERT INTO product_comparison.*`).',
                    'mock.ExpectExec(`INSERT INTO product_comparison (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).'
                )
                break
    
    # Fix 31: Comparison_List
    if 'func TestComparison_List' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'SELECT p.id, p.name, p.slug.*' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT p.id, p.name, p.slug.*`).',
                    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.description, COALESCE(p.asset_path, \'\') as asset_path, COALESCE(p.asset_hash, \'\') as asset_hash, COALESCE(p.file_mime_type, \'\') as file_mime_type, COALESCE(created_at::text, \'\') as created_at FROM product_comparison pc JOIN products p ON p.id = pc.product_id WHERE pc.user_id = $1 ORDER BY pc.added_at DESC`).'
                ).replace(
                    'sqlmock.NewRows([]string{"id", "name", "slug", "price_usd", "description"})',
                    'sqlmock.NewRows([]string{"id", "title", "slug", "price_usd", "stock_quantity", "description", "asset_path", "asset_hash", "file_mime_type", "created_at"})'
                ).replace(
                    'AddRow(1, "Test Product", "test-product", "10.00", "A test product")',
                    'AddRow(1, "Test Product", "test-product", "10.00", 100, "A test product", "", "", "", mockTime(2024, time.January, 1))'
                )
    
    # Fix 32: Recommendations_Get
    if 'func TestRecommendations_Get' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'SELECT category_id FROM products WHERE id = $1' in lines[j]:
                lines[j] = lines[j].replace('SELECT category_id FROM products WHERE id = $1', 'SELECT category_id FROM products WHERE id = $1')
            if 'SELECT p.id, p.title.*FROM products p.*' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT p.id, p.title.*FROM products p.*`).',
                    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, \'\') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.id != $1 AND p.status = \'"\'active\'"\' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $2 OFFSET $3`).'
                ).replace(
                    'AddRow(2, "Related Product", "related-product", "15.00", "Related to test")',
                    'AddRow(2, "Related Product", "related-product", "15.00", 50, "active", "http://example.com/img.png")'
                ).replace(
                    'sqlmock.NewRows([]string{"id", "name", "slug", "price_usd", "description"})',
                    'sqlmock.NewRows([]string{"id", "title", "slug", "price_usd", "stock_quantity", "status", "primary_image"})'
                )
    
    # Fix 33: GuestOrder_Create
    if 'func TestGuestOrder_Create' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'INSERT INTO guest_orders.*' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectExec(`INSERT INTO guest_orders.*`).',
                    'mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0) FROM products WHERE id = ANY($1::int[]) AND status = \'active\'`).WithArgs([]int{1}).WillReturnRows(sqlmock.NewRows([]string{"COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0)"}).AddRow("10.00"))\n\t\tmock.ExpectExec(`INSERT INTO orders (email, total_usd, crypto_chain, status, guest_order) VALUES ($1, $2, $3, \'pending\', true) RETURNING id`).'
                )
            if 'SELECT id, email, status.*FROM guest_orders WHERE id = $1' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT id, email, status.*FROM guest_orders WHERE id = $1`).',
                    'mock.ExpectQuery(`SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = $1 AND guest_order = true`).'
                ).replace(
                    'sqlmock.NewRows([]string{"id", "email", "status", "total_usd", "crypto_chain", "created_at"})',
                    'sqlmock.NewRows([]string{"email", "status", "total_usd", "crypto_chain"})'
                ).replace(
                    'AddRow(1, "test@example.com", "pending", "10.00", "BSC", mockTime(2024, time.January, 1))',
                    'AddRow("test@example.com", "pending", "10.00", "BSC")'
                )
    
    # Fix 34: GuestOrder_Check
    if 'func TestGuestOrder_Check' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'SELECT id, email, status.*FROM guest_orders WHERE id = $1' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT id, email, status.*FROM guest_orders WHERE id = $1`).',
                    'mock.ExpectQuery(`SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = $1 AND guest_order = true`).'
                ).replace(
                    'sqlmock.NewRows([]string{"id", "email", "status", "total_usd", "crypto_chain", "created_at"})',
                    'sqlmock.NewRows([]string{"email", "status", "total_usd", "crypto_chain"})'
                ).replace(
                    'AddRow(1, "test@example.com", "completed", "10.00", "BSC", mockTime(2024, time.January, 1))',
                    'AddRow("test@example.com", "completed", "10.00", "BSC")'
                )
    
    # Fix 35: OrderStatusCheck
    if 'func TestOrderStatusCheck' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'SELECT id, status, total_usd.*FROM orders WHERE id = $1' in lines[j]:
                lines[j] = lines[j].replace(
                    'mock.ExpectQuery(`SELECT id, status, total_usd.*FROM orders WHERE id = $1`).',
                    'mock.ExpectQuery(`SELECT id, status, total_usd, crypto_chain, payment_tx_hash, payment_confirmations, paid_at FROM orders WHERE id = $1`).'
                )
    
    # Fix 36: Products_CreateProduct
    if 'func TestProducts_CreateProduct' in line:
        for j in range(i, min(i+15, len(lines))):
            if 'COUNT.*categories WHERE id' in lines[j]:
                lines[j] = lines[j].replace('COUNT(\\\\*)', 'COUNT(\\*)')
    
    # Fix 37: NewProducts_Delete
    if 'func TestNewProducts_Delete' in line:
        for j in range(i, min(i+10, len(lines))):
            if 'DELETE FROM products WHERE id = $1' in lines[j]:
                pass  # Already correct
    
    out.append(line)
    i += 1

# Post-processing: fix any remaining COUNT(\\\\*) patterns
result = '\n'.join(out)
result = result.replace('COUNT(\\\\\\\\*)', 'COUNT(\\\\*)')

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(result)

print("Done. Lines:", len(out))
