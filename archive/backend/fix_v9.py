#!/usr/bin/env python3
"""Fix all_features_test.go: fix \\ to \ in raw strings, add WithArgs."""
import re, subprocess, shutil

SRC = "/root/project/backend/handlers/all_features_test.go"

with open(SRC) as f:
    text = f.read()

# Step 1: Fix \\$ -> \$ inside backtick raw strings
out = []
i = 0
depth = 0
changes = 0
while i < len(text):
    if text[i] == '`':
        depth += 1
        out.append('`')
        i += 1
        while i < len(text) and text[i] != '`':
            if depth % 2 == 1 and i + 2 < len(text) and text[i:i+3] == '\\\\$':
                out.append('\\$')
                i += 3
                changes += 1
            elif depth % 2 == 1 and i + 1 < len(text) and text[i:i+2] == '\\\\':
                # Check if followed by $ or other special char
                if i + 3 < len(text) and text[i+2] == '$':
                    out.append('\\$')
                    i += 3
                elif i + 2 < len(text) and text[i+2] == '*':
                    out.append('\\*')
                    i += 3  
                else:
                    out.append('\\')
                    i += 1
            elif depth % 2 == 1 and i + 1 < len(text) and text[i:i+2] == '\\(':
                out.append('\\(')
                i += 2
            elif depth % 2 == 1 and i + 1 < len(text) and text[i:i+2] == '\\)':
                out.append('\\)')
                i += 2
            else:
                out.append(text[i])
                i += 1
        if i < len(text):
            out.append('`')
            i += 1
        depth -= 1
    else:
        out.append(text[i])
        i += 1

text = ''.join(out)
print(f"Step 1: Fixed {changes} escape sequences in raw strings")

# Step 2: Write and verify build
with open(SRC, 'w') as f:
    f.write(text)

r = subprocess.run(['go', 'build', './handlers/'], capture_output=True, text=True, cwd='/root/project/backend')
print(f"Build: {'OK' if r.returncode == 0 else 'FAIL'}")
if r.returncode:
    print(r.stderr[:500])
    exit(1)

# Step 3: Run tests to see current state
r2 = subprocess.run(['go', 'test', './handlers/', '-count=1', '-v'], capture_output=True, text=True, cwd='/root/project/backend', timeout=90)
lines = (r2.stdout + r2.stderr).split('\n')
passes = sum(1 for l in lines if '--- PASS:' in l)
fails = sum(1 for l in lines if '--- FAIL:' in l)
print(f"\nTest baseline: {passes} PASS, {fails} FAIL")

# Step 4: Apply targeted fixes for each failing test
fixes_applied = 0

def apply_fix(name, old, new):
    global text, fixes_applied
    if old in text:
        text = text.replace(old, new)
        fixes_applied += 1
        print(f"  FIXED: {name}")
        return True
    else:
        print(f"  MISS: {name}")
        return False

# Fix 1: GetProducts - COUNT needs WithArgs("active")
apply_fix("GetProducts COUNT WithArgs",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n\t\tWithArgs("active").\n\t\tWillReturnRows(')

# Fix 2: GetProducts - images need ORDER BY + WithArgs
apply_fix("GetProducts images 1",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

apply_fix("GetProducts images 2",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(2).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(2).\n\t\tWillReturnRows(')

# Fix 3: GetProductBySlug - COUNT needs WithArgs
apply_fix("GetProductBySlug COUNT WithArgs",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n\t\tWithArgs("test-product").\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n\t\tWithArgs("test-product").\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))')

apply_fix("GetProductBySlug images",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# Fix 4: SearchProducts - COUNT WHERE clause 
apply_fix("SearchProducts COUNT WHERE",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND (p.name LIKE $2 OR p.description LIKE $2 OR p.slug LIKE $2)`).\n\t\tWithArgs("active", "test").\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))')

apply_fix("SearchProducts images 1",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

apply_fix("SearchProducts images 2",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(2).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(2).\n\t\tWillReturnRows(')

# Fix 5: Auth_UpdateProfile - various queries need proper SQL and WithArgs
apply_fix("UpdateProfile SELECT users",
    'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

apply_fix("UpdateProfile SELECT profile",
    'mock.ExpectQuery(`SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

apply_fix("UpdateProfile UPDATE users",
    'mock.ExpectExec(`UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).\n\t\tWithArgs(',
    'mock.ExpectExec(`UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).\n\t\tWithArgs("New Name", 1).WillReturnResult(')

apply_fix("UpdateProfile INSERT profile",
    'mock.ExpectExec(`INSERT INTO user_profiles (user_id, bio, wallet_address) VALUES ($1, $2, $3) ON CONFLICT (user_id) DO NOTHING`).\n\t\tWithArgs(',
    'mock.ExpectExec(`INSERT INTO user_profiles (user_id, bio, wallet_address) VALUES ($1, $2, $3) ON CONFLICT (user_id) DO NOTHING`).\n\t\tWithArgs(1, "New bio", "0xNewAddress").WillReturnResult(')

# Fix 6: Admin_BulkProductStatus UPDATE
apply_fix("AdminBulkStatus UPDATE",
    'mock.ExpectExec(`UPDATE products SET status = $1 WHERE id = ANY($2)`).\n\t\tWithArgs(1, 1).\n\t\tWillReturnResult(',
    'mock.ExpectExec(`UPDATE products SET status = $1 WHERE id = ANY($2)`).\n\t\tWithArgs("active", []int{1, 2}).WillReturnResult(')

# Fix 7: Admin order status UPDATEs - fix the 4 occurrences
for status in ['paid', 'completed', 'cancelled', 'refunded']:
    extra = ", paid_at = NOW()" if status == "paid" else ""
    apply_fix(f"Admin_{status} UPDATE",
        f'mock.ExpectExec(`UPDATE orders SET status = \'{status}\'{extra} WHERE id = $1`).\n\t\tWillReturnResult(',
        f'mock.ExpectExec(`UPDATE orders SET status = \'{status}\'{extra} WHERE id = $1`).\n\t\tWithArgs(1).WillReturnResult(')

# Fix 8: Admin_BulkProductCategory UPDATE
apply_fix("AdminBulkCategory UPDATE",
    'mock.ExpectExec(`UPDATE products SET category_id = $1 WHERE id = ANY($2)`).\n\t\tWithArgs(1, 1).\n\t\tWillReturnResult(',
    'mock.ExpectExec(`UPDATE products SET category_id = $1 WHERE id = ANY($2)`).\n\t\tWithArgs(1, []int{1, 2}).WillReturnResult(')

# Fix 9: Admin_BulkProductToggle UPDATE
apply_fix("AdminBulkToggle UPDATE",
    'mock.ExpectExec(`UPDATE products SET status = $1 WHERE id = ANY($2)`).\n\t\tWithArgs(1, 1).\n\t\tWillReturnResult(',
    'mock.ExpectExec(`UPDATE products SET status = $1 WHERE id = ANY($2)`).\n\t\tWithArgs("active", []int{1, 2}).WillReturnResult(')

# Fix 10: Admin_DeleteUser
apply_fix("AdminDeleteUser SELECT",
    'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

apply_fix("AdminDeleteUser DELETE",
    'mock.ExpectExec(`DELETE FROM users WHERE id = $1`).\n\t\tWithArgs(1).',
    'mock.ExpectExec(`DELETE FROM users WHERE id = $1`).\n\t\tWithArgs(1).')

# Fix 11: Admin_DeleteCommunityPost
apply_fix("AdminDeletePost EXISTS",
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM community_posts WHERE id = $1)`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM community_posts WHERE id = $1)`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

apply_fix("AdminDeletePost DELETE",
    'mock.ExpectExec(`DELETE FROM community_posts WHERE id = $1`).\n\t\tWithArgs(1).',
    'mock.ExpectExec(`DELETE FROM community_posts WHERE id = $1`).\n\t\tWithArgs(1).')

# Fix 12: Admin_CommunityPostList
apply_fix("AdminCommunityPosts COUNT",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM community_posts WHERE 1=1`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM community_posts WHERE 1=1`).\n\t\tWillReturnRows(')

apply_fix("AdminCommunityPosts LIST",
    'mock.ExpectQuery(`SELECT cp.id, cp.user_id, u.name, u.email, cp.content, cp.created_at, cp.updated_at FROM community_posts cp LEFT JOIN users u ON u.id = cp.user_id WHERE 1=1 ORDER BY cp.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(20, 0).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT cp.id, cp.user_id, u.name, u.email, cp.content, cp.created_at, cp.updated_at FROM community_posts cp LEFT JOIN users u ON u.id = cp.user_id WHERE 1=1 ORDER BY cp.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(20, 0).\n\t\tWillReturnRows(')

# Fix 13: Admin_CommunityUserList
apply_fix("AdminCommunityUsers COUNT",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM users`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM users`).\n\t\tWillReturnRows(')

apply_fix("AdminCommunityUsers LIST",
    'mock.ExpectQuery(`SELECT u.id, u.name, u.email, u.created_at, u.updated_at, COALESCE(up.bio, \'\') as bio, COALESCE(up.avatar_url, \'\') as avatar_url, COALESCE(up.wallet_address, \'\') as wallet_address FROM users u LEFT JOIN user_profiles up ON up.user_id = u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(20, 0).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT u.id, u.name, u.email, u.created_at, u.updated_at, COALESCE(up.bio, \'\') as bio, COALESCE(up.avatar_url, \'\') as avatar_url, COALESCE(up.wallet_address, \'\') as wallet_address FROM users u LEFT JOIN user_profiles up ON up.user_id = u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(20, 0).\n\t\tWillReturnRows(')

# Fix 14: Admin_GuestOrderCheck
apply_fix("AdminGuestOrderCheck",
    'mock.ExpectQuery(`SELECT guest_email FROM orders WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT guest_email FROM orders WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# Fix 15: Admin_ExchangeRateList
apply_fix("AdminExchangeRateList",
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).\n\t\tWithArgs().\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).\n\t\tWithArgs().\n\t\tWillReturnRows(')

# Fix 16: Admin_ExchangeRateSet
apply_fix("AdminExchangeRateSet",
    'mock.ExpectExec(`INSERT INTO exchange_rates (chain, symbol, rate_to_usd) VALUES ($1, $2, $3) ON CONFLICT (chain) DO UPDATE SET symbol = $2, rate_to_usd = $3, updated_at = NOW()`).\n\t\tWithArgs(',
    'mock.ExpectExec(`INSERT INTO exchange_rates (chain, symbol, rate_to_usd) VALUES ($1, $2, $3) ON CONFLICT (chain) DO UPDATE SET symbol = $2, rate_to_usd = $3, updated_at = NOW()`).\n\t\tWithArgs("BSC", "BNB", "0.03").WillReturnResult(')

# Fix 17: Admin_GetSetting (was Admin_Setting_Get)
apply_fix("AdminGetSetting",
    'mock.ExpectQuery(`SELECT value FROM settings WHERE key = $1`).\n\t\tWithArgs("payment_address").\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT value FROM settings WHERE key = $1`).\n\t\tWithArgs("payment_address").\n\t\tWillReturnRows(')

# Fix 18: Admin_Stats
apply_fix("AdminStats users",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM users`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM users`).\n\t\tWithArgs().\n\t\tWillReturnRows(')

apply_fix("AdminStats orders",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM orders`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM orders`).\n\t\tWithArgs().\n\t\tWillReturnRows(')

apply_fix("AdminStats revenue",
    'mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(total_usd AS NUMERIC)), 0) FROM orders WHERE status = \'completed\'`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(total_usd AS NUMERIC)), 0) FROM orders WHERE status = \'completed\'`).\n\t\tWithArgs().\n\t\tWillReturnRows(')

# Fix 19: Cart_GetOrCreate
apply_fix("Cart_GetOrCreate",
    'mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# Fix 20: Cart_AddItem - check features.go for exact SQL
# From features.go: SELECT id FROM carts WHERE user_id=$1, INSERT INTO carts, SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id=$1, INSERT INTO cart_items, SELECT COALESCE(SUM(quantity),0) FROM cart_items WHERE cart_id=$1

apply_fix("Cart_AddItem stock",
    'mock.ExpectQuery(`SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# Fix 21: Cart_RemoveItem
apply_fix("Cart_RemoveItem",
    'mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).\n\t\tWithArgs(',
    'mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).\n\t\tWithArgs(1, 1).WillReturnResult(')

# Fix 22: Wishlist_Toggle
apply_fix("Wishlist_Toggle EXISTS",
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).\n\t\tWithArgs(1, 1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).\n\t\tWithArgs(1, 1).')

apply_fix("Wishlist_Toggle INSERT",
    'mock.ExpectExec(`INSERT INTO wishlists (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).\n\t\tWithArgs(',
    'mock.ExpectExec(`INSERT INTO wishlists (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).\n\t\tWithArgs(1, 1).WillReturnResult(')

apply_fix("Wishlist_List SQL",
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, \'\') as primary_image FROM wishlists w JOIN products p ON p.id = w.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE w.user_id = $1 ORDER BY w.created_at DESC`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, \'\') as primary_image FROM wishlists w JOIN products p ON p.id = w.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE w.user_id = $1 ORDER BY w.created_at DESC`).\n\t\tWithArgs(1).')

# Fix 23: Reviews_Create
apply_fix("Reviews_Create INSERT",
    'mock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) ON CONFLICT (product_id, user_id) DO UPDATE SET rating = $3, comment = $4, updated_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(1, 1, 5, "Great product!").WillReturnResult(',
    'mock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) ON CONFLICT (product_id, user_id) DO UPDATE SET rating = $3, comment = $4, updated_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(1, 1, 5, "Great product!").WillReturnResult(')

apply_fix("Reviews_List SQL",
    'mock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.id, u.name, u.email, COALESCE(up.avatar_url, \'\') as avatar_url FROM reviews r JOIN users u ON u.id = r.user_id LEFT JOIN user_profiles up ON up.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC LIMIT $2 OFFSET $3`).\n\t\tWithArgs(1, 20, 0).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.id, u.name, u.email, COALESCE(up.avatar_url, \'\') as avatar_url FROM reviews r JOIN users u ON u.id = r.user_id LEFT JOIN user_profiles up ON up.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC LIMIT $2 OFFSET $3`).\n\t\tWithArgs(1, 20, 0).\n\t\tWillReturnRows(')

# Fix 24: RecentlyViewed_Record
apply_fix("RecentlyViewed_Insert",
    'mock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id, viewed_at) VALUES ($1, $2, CURRENT_TIMESTAMP) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(1, 1).WillReturnResult(',
    'mock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id, viewed_at) VALUES ($1, $2, CURRENT_TIMESTAMP) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(1, 1).WillReturnResult(')

apply_fix("RecentlyViewed_List SQL",
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, \'\') as primary_image, rv.viewed_at FROM recently_viewed rv JOIN products p ON p.id = rv.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE rv.user_id = $1 ORDER BY rv.viewed_at DESC LIMIT $2`).\n\t\tWithArgs(1, 20).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, \'\') as primary_image, rv.viewed_at FROM recently_viewed rv JOIN products p ON p.id = rv.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE rv.user_id = $1 ORDER BY rv.viewed_at DESC LIMIT $2`).\n\t\tWithArgs(1, 20).')

# Fix 25: Comparison_Toggle
apply_fix("Comparison_Toggle EXISTS",
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).\n\t\tWithArgs(1, 1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).\n\t\tWithArgs(1, 1).')

apply_fix("Comparison_Toggle INSERT",
    'mock.ExpectExec(`INSERT INTO product_comparison (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).\n\t\tWithArgs(',
    'mock.ExpectExec(`INSERT INTO product_comparison (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).\n\t\tWithArgs(1, 1).WillReturnResult(')

apply_fix("Comparison_List SQL",
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.description, COALESCE(p.asset_path, \'\') as asset_path, COALESCE(p.asset_hash, \'\') as asset_hash, COALESCE(p.file_mime_type, \'\') as file_mime_type, COALESCE(created_at::text, \'\') as created_at FROM product_comparison pc JOIN products p ON p.id = pc.product_id WHERE pc.user_id = $1 ORDER BY pc.added_at DESC`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.description, COALESCE(p.asset_path, \'\') as asset_path, COALESCE(p.asset_hash, \'\') as asset_hash, COALESCE(p.file_mime_type, \'\') as file_mime_type, COALESCE(created_at::text, \'\') as created_at FROM product_comparison pc JOIN products p ON p.id = pc.product_id WHERE pc.user_id = $1 ORDER BY pc.added_at DESC`).\n\t\tWithArgs(1).')

# Fix 26: Recommendations
apply_fix("Recommendations cat lookup",
    'mock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"category_id"}).AddRow(nil))',
    'mock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"category_id"}).AddRow(nil))')

apply_fix("Recommendations global SQL",
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, \'\') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.id != $1 AND p.status = \'active\' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $2 OFFSET $3`).\n\t\tWithArgs(1, 5, 0).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, \'\') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.id != $1 AND p.status = \'active\' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $2 OFFSET $3`).\n\t\tWithArgs(1, 5, 0).\n\t\tWillReturnRows(')

# Fix 27: GuestOrder_Create
apply_fix("GuestOrder_Create sum",
    'mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0) FROM products WHERE id = ANY($1::int[]) AND status = \'active\'`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0) FROM products WHERE id = ANY($1::int[]) AND status = \'active\'`).\n\t\tWithArgs([]int{1}).WillReturnRows(')

apply_fix("GuestOrder_Create INSERT",
    'mock.ExpectExec(`INSERT INTO orders (email, total_usd, crypto_chain, status, guest_order) VALUES ($1, $2, $3, \'pending\', true) RETURNING id`).\n\t\tWithArgs(',
    'mock.ExpectExec(`INSERT INTO orders (email, total_usd, crypto_chain, status, guest_order) VALUES ($1, $2, $3, \'pending\', true) RETURNING id`).\n\t\tWithArgs("test@example.com", "10.00", "BSC").WillReturnResult(')

apply_fix("GuestOrder_Check SQL",
    'mock.ExpectQuery(`SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = $1 AND guest_order = true`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = $1 AND guest_order = true`).\n\t\tWithArgs(1).')

# Fix 28: OrderStatusCheck
apply_fix("OrderStatusCheck SQL",
    'mock.ExpectQuery(`SELECT id, user_id, status, total_crypto, crypto_chain, payment_address, payment_tx_hash, payment_confirmations FROM orders WHERE id = $1 AND user_id = $2`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT id, user_id, status, total_crypto, crypto_chain, payment_address, payment_tx_hash, payment_confirmations FROM orders WHERE id = $1 AND user_id = $2`).\n\t\tWithArgs(1, 1).')

# Fix 29: Products_CreateProduct
apply_fix("CreateProduct slug check",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n\t\tWithArgs("test-product").\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n\t\tWithArgs("test-product").')

apply_fix("CreateProduct category check",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM categories WHERE id = $1`).\n\t\tWithArgs(1).',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM categories WHERE id = $1`).\n\t\tWithArgs(1).')

# Fix 30: NewProducts_GetActiveStat
apply_fix("NewProducts_GetActiveStat",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = $1`).\n\t\tWithArgs("active").\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = $1`).\n\t\tWithArgs("active").')

# Fix 31: NewProducts_Delete
apply_fix("NewProducts_Delete images",
    'mock.ExpectExec(`DELETE FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(1).',
    'mock.ExpectExec(`DELETE FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(1).')

apply_fix("NewProducts_Delete products",
    'mock.ExpectExec(`DELETE FROM products WHERE id = $1`).\n\t\tWithArgs(1).',
    'mock.ExpectExec(`DELETE FROM products WHERE id = $1`).\n\t\tWithArgs(1).')

# Fix 32: ExchangeRates_GetAll
apply_fix("ExchangeRates_GetAll",
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).\n\t\tWithArgs().\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).\n\t\tWithArgs().')

apply_fix("ExchangeRates_GetOne",
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1`).\n\t\tWithArgs("BSC").\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1`).\n\t\tWithArgs("BSC").')

# Fix 33: DBSchema_Migration011Tables
apply_fix("DBSchema_Migration tables",
    'mock.ExpectQuery(`SELECT table_name FROM information_schema.tables WHERE table_schema = \'public\' AND table_name = $1`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT table_name FROM information_schema.tables WHERE table_schema = \'public\' AND table_name = $1`).\n\t\tWithArgs(')

# Fix 34: Community_Feed - check exact SQL
# From earlier extraction: SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url, (SELECT COUNT(*) FROM post_likes WHERE post_id = cp.id) as count_likes, (SELECT COUNT(*) FROM post_comments WHERE post_id = cp.id) as count_comments, EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = cp.user_id) as following FROM community_posts cp JOIN users u ON u.id = cp.user_id LEFT JOIN user_profiles up ON up.user_id = u.id LEFT JOIN post_likes l ON l.post_id = cp.id LEFT JOIN post_comments c ON c.post_id = cp.id WHERE cp.user_id = $1 GROUP BY cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url

apply_fix("Community_Feed SQL",
    'mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url, (SELECT COUNT(*) FROM post_likes WHERE post_id = cp.id) as count_likes, (SELECT COUNT(*) FROM post_comments WHERE post_id = cp.id) as count_comments, EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = cp.user_id) as following FROM community_posts cp JOIN users u ON u.id = cp.user_id LEFT JOIN user_profiles up ON up.user_id = u.id LEFT JOIN post_likes l ON l.post_id = cp.id LEFT JOIN post_comments c ON c.post_id = cp.id WHERE cp.user_id = $1 GROUP BY cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url, (SELECT COUNT(*) FROM post_likes WHERE post_id = cp.id) as count_likes, (SELECT COUNT(*) FROM post_comments WHERE post_id = cp.id) as count_comments, EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = cp.user_id) as following FROM community_posts cp JOIN users u ON u.id = cp.user_id LEFT JOIN user_profiles up ON up.user_id = u.id LEFT JOIN post_likes l ON l.post_id = cp.id LEFT JOIN post_comments c ON c.post_id = cp.id WHERE cp.user_id = $1 GROUP BY cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url`).\n\t\tWithArgs(1).')

# Fix 35: Community_Profile SQL 
# SELECT id, name FROM users WHERE id = $1
# SELECT bio, avatar_url, wallet_address, created_at FROM user_profiles WHERE user_id = $1  
# SELECT COUNT(*) FROM community_posts WHERE user_id = $1
# SELECT COUNT(*) FROM follows WHERE following_id = $1
# SELECT COUNT(*) FROM follows WHERE follower_id = $1

apply_fix("Community_Profile user SELECT",
    'mock.ExpectQuery(`SELECT id, name FROM users WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, name FROM users WHERE id = $1`).\n\t\tWithArgs(1).')

apply_fix("Community_Profile profile SELECT",
    'mock.ExpectQuery(`SELECT bio, avatar_url, wallet_address, created_at FROM user_profiles WHERE user_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT bio, avatar_url, wallet_address, created_at FROM user_profiles WHERE user_id = $1`).\n\t\tWithArgs(1).')

apply_fix("Community_Profile post count",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM community_posts WHERE user_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM community_posts WHERE user_id = $1`).\n\t\tWithArgs(1).')

apply_fix("Community_Profile follower count",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM follows WHERE following_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM follows WHERE following_id = $1`).\n\t\tWithArgs(1).')

apply_fix("Community_Profile following count",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM follows WHERE follower_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM follows WHERE follower_id = $1`).\n\t\tWithArgs(1).')

# Write all fixes
with open(SRC, 'w') as f:
    f.write(text)

print(f"\nApplied {fixes_applied} targeted fixes")

# Rebuild and test
r = subprocess.run(['go', 'build', './handlers/'], capture_output=True, text=True, cwd='/root/project/backend')
print(f"Build: {'OK' if r.returncode == 0 else 'FAIL'}")
if r.returncode:
    print(r.stderr[:300])
    exit(1)

r2 = subprocess.run(['go', 'test', './handlers/', '-count=1'], capture_output=True, text=True, cwd='/root/project/backend', timeout=120)
lines = (r2.stdout + r2.stderr).split('\n')
passes = sum(1 for l in lines if '--- PASS:' in l)
fails = sum(1 for l in lines if '--- FAIL:' in l)
print(f"\nFinal result: {passes} PASS, {fails} FAIL, exit={r2.returncode}")
for l in lines:
    if '--- FAIL:' in l:
        print(f"  FAIL: {l.strip()}")
    elif 'unfulfilled' in l and 'Expectation' in l:
        print(f"  MOCK: {l.strip()[:140]}")
