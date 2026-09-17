#!/usr/bin/env python3
"""Fix all mock SQL patterns to match handler queries."""
import re, subprocess

with open("/root/project/backend/handlers/all_features_test.go") as f:
    content = f.read()

lines = content.split('\n')
fixed = 0

# Fix 1: Products_GetProducts COUNT (L123)
lines[122] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n'

# Fix 2: Products_GetProducts list (L125)
lines[124] = (
    '\tmock.ExpectQuery(`(?s)SELECT\\s+p\\.id,\\s+p\\.title,\\s+p\\.slug,\\s+p\\.description,\\s+'
    'p\\.category_id,\\s+COALESCE\\(c\\.name,\\s+\\'\\',\\s+\\)\\s+p\\.price_usd,\\s+p\\.asset_path,\\s+p\\.asset_hash,\\s+p\\.status,\\s+'
    'p\\.download_count_limit,\\s+p\\.max_downloads_per_user,\\s+p\\.file_size_bytes,\\s+p\\.file_mime_type,\\s+p\\.created_at,\\s+p\\.updated_at\\s+FROM\\s+products\\s+p\\s+LEFT\\s+JOIN\\s+categories\\s+c\\s+'
    'ON\\s+p\\.category_id\\s*=\\s*c\\.id\\s+WHERE\\s+p\\.status\\s*=\\s*\\$1\\s+'
    'ORDER\\s+BY\\s+p\\.created_at\\s+DESC\\s+LIMIT\\s+\\$2\\s+OFFSET\\s+\\$3`).\n'
)
fixed += 2

# Fix 3: Products_GetProductBySlug (L137-139)
lines[137] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n'
lines[139] = (
    '\tmock.ExpectQuery(`(?s)SELECT\\s+p\\.id,\\s+p\\.title,\\s+p\\.slug,\\s+p\\.description,\\s+p\\.category_id,\\s+COALESCE\\(c\\.name,\\s+\\'\\',\\s+\\)\\s+p\\.price_usd,\\s+p\\.asset_path,\\s+p\\.asset_hash,\\s+p\\.status,\\s+p\\.download_count_limit,\\s+p\\.max_downloads_per_user,\\s+p\\.file_size_bytes,\\s+p\\.file_mime_type,\\s+p\\.created_at,\\s+p\\.updated_at\\s+FROM\\s+products\\s+p\\s+LEFT\\s+JOIN\\s+categories\\s+c\\s+'
    'ON\\s+p\\.category_id\\s*=\\s*c\\.id\\s+WHERE\\s+p\\.slug\\s*=\\s*\\$1\\s+ORDER\\s+BY\\s+p\\.created_at\\s+DESC\\s+LIMIT\\s+\\s+OFFSET\\s+').\n'
)
fixed += 2

# Fix 4: Products_SearchProducts (L157-159)
lines[156] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE (p.title ILIKE $1 OR p.description ILIKE $1) AND p.status = $2`).\n'
lines[158] = (
    '\tmock.ExpectQuery(`(?s)SELECT\\s+p\\.id,\\s+p\\.title,\\s+p\\.slug,\\s+p\\.description,\\s+p\\.category_id,\\s+COALESCE\\(c\\.name,\\s+\\'\\',\\s+\\)\\s+p\\.price_usd,\\s+p\\.asset_path,\\s+p\\.asset_hash,\\s+p\\.status,\\s+p\\.download_count_limit,\\s+p\\.max_downloads_per_user,\\s+p\\.file_size_bytes,\\s+p\\.file_mime_type,\\s+p\\.created_at,\\s+p\\.updated_at\\s+FROM\\s+products\\s+p\\s+LEFT\\s+JOIN\\s+categories\\s+c\\s+'
    'ON\\s+p\\.category_id\\s*=\\s*c\\.id\\s+\\(p\\.title\\s+ILIKE\\s+\\$1\\s+OR\\s+p\\.description\\s+ILIKE\\s+\\$1\\)\\s+AND\\s+p\\.status\\s*=\\s*\\$2\\s+ORDER\\s+BY\\s+p\\.created_at\\s+DESC\\s+LIMIT\\s+\\$3\\s+OFFSET\\s+\\$4`).\n'
)
fixed += 2

# Fix 5: Cart_GetOrCreate (L482-487)
lines[482] = '\tmock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).\n'
lines[487] = '\t\tmock.ExpectExec(`INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, $3) ON CONFLICT (cart_id, product_id) DO UPDATE SET quantity = $3, added_at = CURRENT_TIMESTAMP`).\n'
fixed += 2

# Fix 6: Wishlist Toggle (L521-523)
lines[521] = '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`. \n'
lines[523] = '\t\tmock.ExpectExec(`INSERT INTO wishlists (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).\n'
fixed += 2

# Fix 7: Reviews (L551-553)
lines[551] = '\t\tmock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) RETURNING id`).\n'
lines[563] = '\tmock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.name FROM reviews r JOIN users u ON r.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC`).\n'
fixed += 2

# Fix 7b: RecentlyViewed (L576-578)
lines[576] = '\t\tmock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = NOW()`. \n'
lines[578] = '\tmock.ExpectQuery(`SELECT p.id, p.name, p.slug, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM recently_viewed WHERE user_id = $1 ORDER BY viewed_at DESC LIMIT 10)`. \n'
fixed += 2

# Fix 9: Comparison (L595-607)
lines[595] = '\t\tmock.ExpectExec(`INSERT INTO product_comparison (user_id, product_id) VALUES ($1, $2) ON CONFLICT (user_id, product_id) DO NOTHING`).\n'
lines[607] = '\tmock.ExpectQuery(`SELECT p.id, p.name, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.id IN (SELECT product_id FROM product_comparison WHERE user_id = $1 ORDER BY added_at DESC LIMIT 4)`. \n'
fixed += 2

# Fix 10: Recommendations (L622-624)
lines[622] = '\tmock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).\n'
lines[624] = '\tmock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.price_usd, p.asset_path, p.asset_hash, p.status FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.category_id = $1 AND p.status = $2 AND p.id != $3 ORDER BY p.created_at DESC LIMIT $4`).\n'
fixed += 2

# Fix 10: GuestOrder (L641-643)
lines[641] = '\t\tmock.ExpectExec(`INSERT INTO guest_orders (email, total_usd, crypto_chain, crypto_amount, crypto_address, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`).\n'
lines[643] = '\tmock.ExpectQuery(`SELECT id, email, status, total_usd, crypto_chain, crypto_amount, crypto_address FROM guest_orders WHERE id = $1`).\n'
fixed += 2

# Fix 11: OrderStatusCheck (L670)
lines[670] = '\tmock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).\n'
fixed += 1

# Fix 12: NewProducts_GetActiveStat (L725-729)
lines[725] = "\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = 'active'`).\n"
lines[727] = "\tmock.ExpectQuery(`SELECT COUNT(*) FROM products`).\n"
lines[729] = "\tmock.ExpectQuery(`SELECT COUNT(*) FROM categories`).\n"
fixed += 3

# Fix 13: NewProducts_Delete (L691-693)
lines[691] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n'
lines[693] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM categories WHERE id = $1`).\n'
fixed += 2

# Fix 14: ExchangeRates (L403, L417)
lines[403] = '\tmock.ExpectQuery(`SELECT id, chain, symbol, rate_to_usd, updated_at FROM exchange_rates ORDER BY chain`).\n'
lines[417] = '\tmock.ExpectQuery(`SELECT id, chain, symbol, rate_to_usd, updated_at FROM exchange_rates WHERE chain = $1`).\n'
fixed += 2

# Fix 15: DBSchema (L782)
lines[782] = "\tmock.ExpectQuery(`SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1`).\n"
fixed += 1

# Fix 16: Auth_UpdateProfile (L70, L74, L77, L80)
lines[70] = '\tmock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).\n'
lines[74] = '\tmock.ExpectQuery(`SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1`).\n'
lines[77] = '\t\tmock.ExpectExec(`UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).\n'
lines[80] = '\tmock.ExpectQuery(`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`).\n'
fixed += 4

# Fix 16b: Admin Stats COUNT
for target, _ in [("SELECT COUNT(*) FROM users", ""), ("SELECT COUNT(*) FROM orders", ""), ("SELECT COUNT(*) FROM categories", "")]:
    for i, line in enumerate(lines):
        if f'SELECT COUNT(*) FROM {target.split()[2]}' in line and i > 450:
            lines[i] = '\tmock.ExpectQuery(`' + target + '`).\n'
            fixed += 1
            break

# Fix 16c: Admin order status
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`UPDATE orders SET' in line:
        lines[i] = '\t\tmock.ExpectExec(`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`).\n'
        fixed += 1

# Fix 16d: Admin GuestOrderCheck
for i, line in enumerate(lines):
    if 'SELECT o.id, o.status' in line and 'FROM orders WHERE o.id' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT o.id, o.status, o.total_usd, o.created_at FROM orders WHERE o.id = $1`).\n'
        fixed += 1

# Fix 16e: Admin GetSetting
for i, line in enumerate(lines):
    if 'SELECT setting_key, value, updated_at FROM settings ORDER BY setting_key' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT setting_key, value, updated_at FROM settings ORDER BY setting_key`).\n'
        fixed += 1

# Fix 16f: Admin SetSetting
for i, line in enumerate(lines):
    if 'INSERT INTO settings' in line and 'mock.ExpectExec' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO settings (setting_key, value) VALUES ($1, $2) ON CONFLICT (setting_key) DO UPDATE SET value = $2`).\n'
        fixed += 1

# Fix 16g: Admin ExchangeRateSet
for i, line in enumerate(lines):
    if 'INSERT INTO exchange_rates' in line and 'mock.ExpectExec' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO exchange_rates (chain, symbol, rate_to_usd) VALUES ($1, $2, $3) ON CONFLICT (chain) DO UPDATE SET symbol = $2, rate_to_usd = $3, updated_at = NOW()`.\n'
        fixed += 1

# Fix 16h: Admin DeleteUser
for i, line in enumerate(lines):
    if 'DELETE FROM users WHERE id' in line and 'mock.ExpectExec' in line:
        lines[i] = '\tmock.ExpectExec(`DELETE FROM users WHERE id = $1`).\n'
        fixed += 1

# Fix 16h: Admin DeletePost
for i, line in enumerate(lines):
    if 'DELETE FROM community_posts WHERE id' in line and 'mock.ExpectExec' in line:
        lines[i] = '\tmock.ExpectExec(`DELETE FROM community_posts WHERE id = $1`).\n'
        fixed += 1

# Fix 16i: Admin CommunityPostList
find_text = 'SELECT cp.id, cp.user_id FROM community_posts cp'
for i, line in enumerate(lines):
    if find_text in line and 'mock.ExpectQuery' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, COALESCE(cl.cnt, 0) + COALESCE(cc.cnt, 0) FROM community_posts cp LEFT JOIN users u ON cp.user_id = u.id LEFT JOIN (SELECT post_id, COUNT(*) as cnt FROM likes WHERE type = \\'post\\' GROUP BY post_id) cl ON cp.id = cl.post_id LEFT JOIN (SELECT post_id, COUNT(*) as cnt FROM comments WHERE type = \\'post\\' GROUP BY post_id) cc ON cp.id = cc.post_id WHERE cp.type = \\'post\\' ORDER BY cp.created_at DESC LIMIT $3 OFFSET $4`).\n'
        fixed += 1

# Fix 16j: Admin CommunityUserList
find_text = 'SELECT u.id, u.email, u.name FROM users u'
for i, line in enumerate(lines):
    if find_text in line and 'mock.ExpectQuery' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT u.id, u.email, u.name, COUNT(cp.id) FROM users u LEFT JOIN community_posts cp ON u.id = cp.user_id WHERE cp.type = \\'post\\' GROUP BY u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`).\n'
        fixed += 1

# Fix 16j: Admin Bulk Product
for i, line in enumerate(lines):
    if 'mock.ExpectExec(`UPDATE products SET' in line and i > 200:
        lines[i] = '\t\tmock.ExpectExec(`UPDATE products SET status = $1, category_id = $2, updated_at = NOW() WHERE id = ANY($3)`. \n'
        fixed += 1

# Fix 16k: Cart WHERE user_id
for i, line in enumerate(lines):
    if 'mock.ExpectQuery(`SELECT id FROM carts WHERE user_id=' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).\n'
        fixed += 1
    if 'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM cart_items WHERE cart_id=' in line:
        lines[i] = '\tmock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM cart_items WHERE cart_id = $1 AND product_id = $2)`. \n'
        fixed += 1
    if 'mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id=' in line:
        lines[i] = '\tmock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).\n'
        fixed += 1
    if 'mock.ExpectExec(`INSERT INTO cart_items' in line:
        lines[i] = '\t\tmock.ExpectExec(`INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, $3) ON CONFLICT (cart_id, product_id) DO UPDATE SET quantity = $3, added_at = CURRENT_TIMESTAMP`).\n'
        fixed += 1

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write('\n'.join(lines))

print(f"Fixed {fixed} mock SQL lines")

r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True, cwd="/root/project/backend")
print(f"Build: exit={r.returncode}")
if r.returncode != 0:
    print(r.stderr[-500:])

r = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True, cwd="/root/project/backend")
out = r.stdout
p = out.count('--- PASS:')
f = out.count('--- FAIL:')
print(f"\nResult: {p} PASS, {f} FAIL, exit={r.returncode}")
if r.returncode != 0:
    print(out[-800:])
    # Show first few failure details
    for line in out.split('\n'):
        if 'FAIL:' in line:
            print(line)
            # Show next few lines
            break
PYEOF