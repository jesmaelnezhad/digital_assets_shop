#!/usr/bin/env python3
"""
Comprehensive fix for all_features_test.go.
Converts double-quoted mock SQL to backtick raw strings, fixes escaping,
adds WithArgs, and fixes SQL patterns to match handler SQL.
"""
import re, subprocess

SRC = "/root/project/backend/handlers/all_features_test.go"
with open(SRC) as f:
    text = f.read()

# ============================================================
# STEP 1: Convert all double-quoted mock strings to backtick raw strings
# ============================================================
out = []
i = 0
n = 0
while i < len(text):
    pos_q = text.find('mock.ExpectQuery("', i)
    pos_e = text.find('mock.ExpectExec("', i)
    if pos_q < 0 and pos_e < 0:
        out.append(text[i:])
        break
    
    if pos_q >= 0 and (pos_e < 0 or pos_q <= pos_e):
        pos, sym = pos_q, 'ExpectQuery'
    else:
        pos, sym = pos_e, 'ExpectExec'
    
    out.append(text[i:pos])
    
    oq = text.find('("', pos) + 2
    cq = oq
    while cq < len(text) and text[cq] != '"':
        cq += 1
    
    if cq >= len(text):
        out.append(text[pos:])
        break
    
    sql_raw = text[oq:cq]
    
    # Transform to backtick raw string
    # Collapse backslash sequences: \\\\\$ -> \$, \\\\ -> \, \$ -> $
    # In raw strings, we want \$ to match literal $ in sqlmock regex
    # But the source has too many backslashes from Go double-quoted escaping
    
    result = []
    j = 0
    while j < len(sql_raw):
        if sql_raw[j] == '\\':
            start = j
            while j < len(sql_raw) and sql_raw[j] == '\\':
                j += 1
            nbs = j - start
            if j < len(sql_raw):
                nc = sql_raw[j]
                # For $, keep exactly one backslash (raw string: \$ = literal \$)
                if nc == '$':
                    # nbs backslashes + $ -> 1 backslash + $
                    result.append('\\')
                    result.append('$')
                    j += 1
                elif nc == '*':
                    result.append('\\')
                    result.append('*')
                    j += 1
                elif nc in '()':
                    result.append('\\')
                    result.append(nc)
                    j += 1
                elif nc == "'":
                    j += 1  # skip backslash
                else:
                    result.append('\\')
            else:
                result.append('\\')
        else:
            result.append(sql_raw[j])
            j += 1
    
    sql_final = ''.join(result)
    replacement = f'mock.{sym}(`{sql_final}`)'
    out.append(replacement)
    i = cq + 1
    n += 1

text = ''.join(out)
print(f"Step 1: Converted {n} mock strings to backtick raw strings")

# ============================================================
# STEP 2: Fix remaining escape issues in backtick strings
# ============================================================
# Inside raw strings: \\$ should become \$ (the conversion above may have left some)
out = []
i = 0
depth = 0
n2 = 0
while i < len(text):
    if text[i] == '`':
        depth += 1
        out.append('`')
        i += 1
    elif depth % 2 == 1:
        if i + 2 < len(text) and text[i:i+3] == '\\\\$':
            out.append('\\$')
            i += 3
            n2 += 1
        elif i + 2 < len(text) and text[i:i+3] == '\\\\*':
            out.append('\\*')
            i += 3
        elif i + 1 < len(text) and text[i:i+2] == '\\\\':
            if i + 3 < len(text) and text[i+2] == '$':
                out.append('\\$')
                i += 3
            elif i + 2 < len(text) and text[i+2] == '*':
                out.append('\\*')
                i += 3
            else:
                out.append('\\')
                i += 1
        elif i + 1 < len(text) and text[i:i+2] == '\\(':
            out.append('\\(')
            i += 2
        elif i + 1 < len(text) and text[i:i+2] == '\\)':
            out.append('\\)')
            i += 2
        else:
            out.append(text[i])
            i += 1
    else:
        out.append(text[i])
        i += 1

text = ''.join(out)
print(f"Step 2: Fixed {n2} escapes in raw strings")

# ============================================================
# STEP 3: Write and verify build
# ============================================================
with open(SRC, 'w') as f:
    f.write(text)

r = subprocess.run(['go', 'build', './handlers/'], capture_output=True, text=True, cwd='/root/project/backend')
print(f"Build: {'OK' if r.returncode == 0 else 'FAIL'}")
if r.returncode:
    print(r.stderr[:500])
    exit(1)

# ============================================================
# STEP 4: Apply SQL fixes for all failing tests
# ============================================================
fixes = 0
def fix(name, old, new):
    global text, fixes
    if old not in text:
        print(f"  MISS: {name}")
        return
    text = text.replace(old, new, 1)
    fixes += 1
    print(f"  FIXED: {name}")

# --- GetProducts (line 120-133): COUNT needs WithArgs, images need ORDER BY + 2nd query ---

# Current line 123: mock.ExpectQuery(`SELECT COUNT\(\*\\) FROM products.*`).
# The .* matches anything but the handler sends specific SQL.
# Fix: replace with exact SQL + WithArgs("active")
old_123 = "mock.ExpectQuery(`SELECT COUNT\\(\\*\\) FROM products.*`).\n\t\tWillReturnRows(sqlmock.NewRows([]string{\"count\"}).AddRow(2))"
new_123 = "mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n\t\tWithArgs(\"active\").\n\t\tWillReturnRows(sqlmock.NewRows([]string{\"count\"}).AddRow(2))"
fix("GetProducts COUNT", old_123, new_123)

# Current lines 125-129: product LIST query (already has correct SQL with WithArgs)
# Just need to add second images query after line 129
# The first images query (line 134 in original) doesn't have ORDER BY
# After the fix, line numbers shift. Let me find the images queries.

# Find all product_images queries
images_matches = list(re.finditer(r'mock\.ExpectQuery\(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \$1`\)', text))
print(f"\nFound {len(images_matches)} product_images queries (need ORDER BY + WithArgs)")

for idx, m in enumerate(images_matches):
    old_img = text[m.start():m.end()]
    new_img = "mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`)."
    text = text[:m.start()] + new_img + text[m.end():]
    fixes += 1
    print(f"  FIXED: product_images query #{idx+1} (ORDER BY)")

# Add WithArgs to images queries
images_withargs = list(re.finditer(r'mock\.ExpectQuery\(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \$1 ORDER BY is_primary DESC, id`\)\.', text))
print(f"Found {len(images_withargs)} images queries without WithArgs")

for idx, m in enumerate(images_withargs):
    old_wa = text[m.start():m.end()]
    new_wa = old_wa[:-1] + f'\n\t\tWithArgs({idx+1}).'
    # Find the WillReturnRows after this
    after = text[m.end():]
    wr_pos = after.find('WillReturnRows')
    if wr_pos >= 0:
        insert_pos = m.end() + wr_pos
        text = text[:insert_pos] + f'\n\t\tWithArgs({idx+1}).' + text[insert_pos:]
        fixes += 1
        print(f"  FIXED: images query #{idx+1} WithArgs({idx+1})")

# --- GetProductBySlug (line 136+): same fixes ---

# --- SearchProducts (line 154+): WHERE clause needs AND clause ---

# --- All other tests: add WithArgs where missing ---

# Cart
fix("Cart_GetOrCreate", 
    'mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).\n\t\tWithArgs(1).')

# Wishlist  
fix("Wishlist_Toggle EXISTS",
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).\n\t\tWithArgs(1, 1).')

# Reviews
fix("Reviews_Create INSERT",
    'mock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) ON CONFLICT (product_id, user_id) DO UPDATE SET rating = $3, comment = $4, updated_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(1, 1, 5, "Great product!").WillReturnResult(',
    'mock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) ON CONFLICT (product_id, user_id) DO UPDATE SET rating = $3, comment = $4, updated_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(1, 1, 5, "Great product!").')

# RecentlyViewed
fix("RecentlyViewed_Insert",  
    'mock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id, viewed_at) VALUES ($1, $2, CURRENT_TIMESTAMP) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(1, 1).WillReturnResult(',
    'mock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id, viewed_at) VALUES ($1, $2, CURRENT_TIMESTAMP) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(1, 1).')

# Comparison
fix("Comparison_Toggle EXISTS",
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).\n\t\tWithArgs(1, 1).')

# Recommendations
fix("Recommendations cat_lookup",
    'mock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"category_id"}).AddRow(nil))',
    'mock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"category_id"}).AddRow(nil))')

# GuestOrder
fix("GuestOrder sum",
    'mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0) FROM products WHERE id = ANY($1::int[]) AND status = \'active\'`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0) FROM products WHERE id = ANY($1::int[]) AND status = \'active\'`).\n\t\tWithArgs([]int{1}).WillReturnRows(')

fix("GuestOrder INSERT",
    'mock.ExpectExec(`INSERT INTO orders (email, total_usd, crypto_chain, status, guest_order) VALUES ($1, $2, $3, \'pending\', true) RETURNING id`).\n\t\tWithArgs(',
    'mock.ExpectExec(`INSERT INTO orders (email, total_usd, crypto_chain, status, guest_order) VALUES ($1, $2, $3, \'pending\', true) RETURNING id`).\n\t\tWithArgs("test@example.com", "10.00", "BSC").WillReturnResult(')

fix("GuestOrder check",
    'mock.ExpectQuery(`SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = $1 AND guest_order = true`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = $1 AND guest_order = true`).\n\t\tWithArgs(1).')

# OrderStatusCheck
fix("OrderStatusCheck",
    'mock.ExpectQuery(`SELECT id, user_id, status, total_crypto, crypto_chain, payment_address, payment_tx_hash, payment_confirmations FROM orders WHERE id = $1 AND user_id = $2`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT id, user_id, status, total_crypto, crypto_chain, payment_address, payment_tx_hash, payment_confirmations FROM orders WHERE id = $1 AND user_id = $2`).\n\t\tWithArgs(1, 1).')

# Admin products LIST
fix("AdminProducts LIST",
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, COALESCE(c.name, \'\') as category_name, COALESCE(c.slug, \'\') as category_slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.status, p.stock_quantity, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.asset_path, p.asset_hash, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE 1=1 ORDER BY p.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, COALESCE(c.name, \'\') as category_name, COALESCE(c.slug, \'\') as category_slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.status, p.stock_quantity, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.asset_path, p.asset_hash, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE 1=1 ORDER BY p.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(20, 0).')

# Admin orders LIST
fix("AdminOrders LIST",
    'mock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.total_crypto, o.crypto_chain, o.payment_address, o.payment_tx_hash, o.payment_confirmations, o.payment_confirmed_at, o.paid_at, o.created_at, o.updated_at FROM orders o ORDER BY o.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.total_crypto, o.crypto_chain, o.payment_address, o.payment_tx_hash, o.payment_confirmations, o.payment_confirmed_at, o.paid_at, o.created_at, o.updated_at FROM orders o ORDER BY o.created_at DESC LIMIT $1 OFFSET $2`).\n\t\tWithArgs(20, 0).')

# Admin order detail
fix("AdminOrderDetail order",
    'mock.ExpectQuery(`SELECT o.id, o.status.*FROM orders o WHERE o.id = $1`).\n\t\tWithArgs(1).',
    'mock.ExpectQuery(`SELECT o.status, CAST(o.total_usd AS NUMERIC), o.guest_email, o.billing_name, o.shipping_address_json, o.notes, o.created_at, o.paid_at FROM orders o WHERE o.id = $1`).\n\t\tWithArgs(1).')

fix("AdminOrderDetail items",
    'mock.ExpectQuery(`SELECT oi.product_id, oi.quantity.*FROM order_items oi WHERE oi.order_id = $1`).\n\t\tWithArgs(1).',
    'mock.ExpectQuery(`SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price_usd, oi.price_crypto, oi.download_count, oi.max_downloads, oi.downloaded_at, oi.created_at, p.title, p.slug FROM order_items oi JOIN products p ON oi.product_id = p.id WHERE oi.order_id = $1`).\n\t\tWithArgs(1).')

# Admin order status UPDATEs
for status in ['paid', 'completed', 'cancelled', 'refunded']:
    extra = ", paid_at = NOW()" if status == "paid" else ""
    fix(f"Admin_{status} UPDATE",
        f'mock.ExpectExec(`UPDATE orders SET.*`).\n\t\tWithArgs(1).WillReturnResult(',
        f'mock.ExpectExec(`UPDATE orders SET status = \'{status}\'{extra} WHERE id = $1`).\n\t\tWithArgs(1).WillReturnResult(')

# Admin bulk
for name, sql, args in [
    ("AdminBulkStatus", "UPDATE products SET status = $1 WHERE id = ANY($2)", '"active", []int{1, 2}'),
    ("AdminBulkCategory", "UPDATE products SET category_id = $1 WHERE id = ANY($2)", '1, []int{1, 2}'),
]:
    fix(f"{name} UPDATE",
        f'mock.ExpectExec(`{sql}`).\n\t\tWithArgs(1, 1).WillReturnResult(',
        f'mock.ExpectExec(`{sql}`).\n\t\tWithArgs({args}).WillReturnResult(')

# Admin settings
fix("AdminSetting_Get",
    'mock.ExpectQuery(`SELECT value FROM settings WHERE key = $1`).\n\t\tWithArgs("payment_address").\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT value FROM settings WHERE key = $1`).\n\t\tWithArgs("payment_address").')

fix("AdminSetting_Set",
    'mock.ExpectExec(`INSERT INTO settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()`).\n\t\tWithArgs(',
    'mock.ExpectExec(`INSERT INTO settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()`).\n\t\tWithArgs("payment_address", "0xNewAddress").WillReturnResult(')

# Admin exchange rates
fix("AdminExchangeRateList",
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).\n\t\tWithArgs().\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).\n\t\tWithArgs().')

fix("AdminExchangeRateSet",
    'mock.ExpectExec(`INSERT INTO exchange_rates (chain, symbol, rate_to_usd) VALUES ($1, $2, $3) ON CONFLICT (chain) DO UPDATE SET symbol = $2, rate_to_usd = $3, updated_at = NOW()`).\n\t\tWithArgs(',
    'mock.ExpectExec(`INSERT INTO exchange_rates (chain, symbol, rate_to_usd) VALUES ($1, $2, $3) ON CONFLICT (chain) DO UPDATE SET symbol = $2, rate_to_usd = $3, updated_at = NOW()`).\n\t\tWithArgs("BSC", "BNB", "0.03").WillReturnResult(')

fix("ExchangeRate_GetOne",
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1`).\n\t\tWithArgs("BSC").\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1`).\n\t\tWithArgs("BSC").')

# Admin community
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

# Admin delete user
fix("AdminDeleteUser SELECT",
    'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).\n\t\tWithArgs(1).')

# Admin guest order check
fix("AdminGuestOrderCheck",
    'mock.ExpectQuery(`SELECT guest_email FROM orders WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT guest_email FROM orders WHERE id = $1`).\n\t\tWithArgs(1).')

# Admin stats
for label in ["users", "orders"]:
    fix(f"AdminStats {label}",
        f'mock.ExpectQuery(`SELECT COUNT(*) FROM { "users" if label == "users" else "orders" }`).\n\t\tWillReturnRows(',
        f'mock.ExpectQuery(`SELECT COUNT(*) FROM {"users" if label == "users" else "orders"}`).\n\t\tWithArgs().')

fix("AdminStats revenue",
    'mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(total_usd AS NUMERIC)), 0) FROM orders WHERE status = \'completed\'`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(total_usd AS NUMERIC)), 0) FROM orders WHERE status = \'completed\'`).\n\t\tWithArgs().')

# Community profile
for label, sql in [
    ("user", "SELECT id, name FROM users WHERE id = $1"),
    ("profile", "SELECT bio, avatar_url, wallet_address, created_at FROM user_profiles WHERE user_id = $1"),
    ("post_count", "SELECT COUNT(*) FROM community_posts WHERE user_id = $1"),
    ("follower_count", "SELECT COUNT(*) FROM follows WHERE following_id = $1"),
    ("following_count", "SELECT COUNT(*) FROM follows WHERE follower_id = $1"),
]:
    fix(f"Community_Profile {label}",
        f'mock.ExpectQuery(`{sql}`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
        f'mock.ExpectQuery(`{sql}`).\n\t\tWithArgs(1).')

# Community feed
fix("Community_Feed",
    'mock.ExpectQuery(`SELECT cp.id, cp.user_id.*FROM community_posts cp.*`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url, (SELECT COUNT(*) FROM post_likes WHERE post_id = cp.id) as count_likes, (SELECT COUNT(*) FROM post_comments WHERE post_id = cp.id) as count_comments, EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = cp.user_id) as following FROM community_posts cp JOIN users u ON u.id = cp.user_id LEFT JOIN user_profiles up ON up.user_id = u.id LEFT JOIN post_likes l ON l.post_id = cp.id LEFT JOIN post_comments c ON c.post_id = cp.id WHERE cp.user_id = $1 GROUP BY cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url`).\n\t\tWithArgs(1).')

# CreateProduct
fix("CreateProduct slug_check",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n\t\tWithArgs("test-product").\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n\t\tWithArgs("test-product").')

fix("CreateProduct cat_check",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM categories WHERE id = $1`).\n\t\tWithArgs(1).',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM categories WHERE id = $1`).\n\t\tWithArgs(1).')

# NewProducts
fix("NewProducts_GetActiveStat",
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = $1`).\n\t\tWithArgs("active").\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = $1`).\n\t\tWithArgs("active").')

fix("NewProducts_Delete images",
    'mock.ExpectExec(`DELETE FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(1).',
    'mock.ExpectExec(`DELETE FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(1).')

fix("NewProducts_Delete products",
    'mock.ExpectExec(`DELETE FROM products WHERE id = $1`).\n\t\tWithArgs(1).',
    'mock.ExpectExec(`DELETE FROM products WHERE id = $1`).\n\t\tWithArgs(1).')

# DBSchema
fix("DBSchema tables",
    'mock.ExpectQuery(`SELECT table_name FROM information_schema.tables WHERE table_schema = \'public\' AND table_name = $1`).\n\t\tWithArgs(',
    'mock.ExpectQuery(`SELECT table_name FROM information_schema.tables WHERE table_schema = \'public\' AND table_name = $1`).\n\t\tWithArgs(\'reviews\').')

# UpdateProfile
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

# UpdateProfile final SELECT
fix("UpdateProfile final SELECT",
    'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).\n\t\t\tAddRow(1, "test@test.com", "New Name", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))',
    'mock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).\n\t\t\tAddRow(1, "test@test.com", "New Name", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))')

# Cart_RemoveItem
fix("Cart_RemoveItem",
    'mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).\n\t\tWithArgs(',
    'mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).\n\t\tWithArgs(1, 1).WillReturnResult(')

# Cart_AddItem stock
fix("Cart_AddItem stock",
    'mock.ExpectQuery(`SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id = $1`).\n\t\tWithArgs(1).',
    'mock.ExpectQuery(`SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id = $1`).\n\t\tWithArgs(1).')

# Write
with open(SRC, 'w') as f:
    f.write(text)

print(f"\n{'='*60}")
print(f"Total fixes: {fixes}")
print(f"{'='*60}\n")

# Build
r = subprocess.run(['go', 'build', './handlers/'], capture_output=True, text=True, cwd='/root/project/backend')
print(f"Build: {'OK' if r.returncode == 0 else 'FAIL'}")
if r.returncode:
    print(r.stderr[:500])
    exit(1)

# Full test
r2 = subprocess.run(['go', 'test', './handlers/', '-count=1'], capture_output=True, text=True, cwd='/root/project/backend', timeout=120)
lines = (r2.stdout + r2.stderr).split('\n')
passes = sum(1 for l in lines if '--- PASS:' in l)
fails = sum(1 for l in lines if '--- FAIL:' in l)
print(f"\nTest result: {passes} PASS, {fails} FAIL, exit={r2.returncode}")
for l in lines:
    if '--- FAIL:' in l:
        print(f"  FAIL: {l.strip()}")
    elif 'unfulfilled' in l and 'Expectation' in l:
        print(f"  MOCK: {l.strip()[:140]}")
