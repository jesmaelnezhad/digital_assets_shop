#!/usr/bin/env python3
"""Comprehensive fix for all_features_test.go - fix all mock SQL to match handlers."""
import re

SRC = "/root/project/backend/handlers/all_features_test.go"
with open(SRC, "r") as f:
    content = f.read()

# ======== Products fixes ========

# 1. TestProducts_GetProducts: COUNT query + product list query
# Handler: SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1
# Handler: SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, ... FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3
old = '''	mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).
		WithArgs("active").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(`SELECT p.id, p.title.*`).'''
new = '''	mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).
		WithArgs("active").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).'''
content = content.replace(old, new, 1)

# 2. TestProducts_GetProductBySlug: slug query + product query 
# Handler product query: WHERE p.slug = $1 AND p.status = 'active' (no ORDER BY/LIMIT)
old2 = '''	mock.ExpectQuery(`SELECT p.id, p.title.*`).
		WithArgs("test-product").'''
new2 = '''	mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.slug = $1 AND p.status = 'active'`).'''
content = content.replace(old2, new2, 1)

# 3. TestProducts_SearchProducts: COUNT + product query with search
# Handler COUNT: SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND p.search_vector @@ plainto_tsquery('english', $2)
# Handler product: same + ORDER BY p.created_at DESC LIMIT $3 OFFSET $4
old3 = '''	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM products.*`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT p.id, p.title.*`).'''
new3 = '''	mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND p.search_vector @@ plainto_tsquery('english', $2)`).
		WithArgs("active", "%test%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND p.search_vector @@ plainto_tsquery('english', $2) ORDER BY p.created_at DESC LIMIT $3 OFFSET $4`).'''
content = content.replace(old3, new3, 1)

# 4. TestProducts_CreateProduct: slug check + category check + INSERT + image INSERT
# Handler slug check: SELECT COUNT(*) FROM products WHERE slug = $1
# Handler cat check: SELECT COUNT(*) FROM categories WHERE id = $1
# Handler INSERT: INSERT INTO products (title, slug, description, category_id, price_usd, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
# Handler image INSERT: INSERT INTO product_images (product_id, url, is_primary) VALUES ($1, $2, $3)
# The test currently uses WithArgs("Test Product", "test-product", "A test product", 1, "10.00", "active") for products INSERT
# But handler does: INSERT INTO products (title, slug, description, category_id, price_usd, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
# That matches! But the category check query needs fixing
old4_cat = '''	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM categories WHERE id = \\$1`).'''
new4_cat = '''	mock.ExpectQuery(`SELECT COUNT(*) FROM categories WHERE id = $1`).'''
content = content.replace(old4_cat, new4_cat, 1)

# 5. TestNewProducts_GetActiveStat
old5 = '''	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM products WHERE status = 'active'`).'''
new5 = '''	mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE status = 'active'`).'''
content = content.replace(old5, new5, 1)

old5b = '''	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM products`).'''
new5b = '''	mock.ExpectQuery(`SELECT COUNT(*) FROM products`).'''
content = content.replace(old5b, new5b, 1)

old5c = '''	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM categories`).'''
new5c = '''	mock.ExpectQuery(`SELECT COUNT(*) FROM categories`).'''
content = content.replace(old5c, new5c, 1)

# ======== Cart fixes ========

# 6. TestCart_GetOrCreate: SELECT id FROM carts WHERE user_id = $1
old6 = '''	mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = \\$1`).'''
new6 = '''	mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).'''
content = content.replace(old6, new6, 1)

# 7. TestCart_AddItem: cart lookup + product stock check + INSERT
# Handler: SELECT id FROM carts WHERE user_id=$1
# Handler: SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id=$1
# Handler: INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, $3) ON CONFLICT...
old7a = '''	mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = \\$1`).'''
# Already handled by fix 6 above for this occurrence... but AddItem has a SECOND cart lookup
# Let me handle Cart_AddItem specifically - it has 2 cart lookups + stock check + insert
# The second cart lookup in AddItem is the same SQL. We need to match it twice.
# Actually looking at the test code, Cart_AddItem test has:
# mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = \$1`).WithArgs(1)... (line 485-486)
# This is the same pattern. The replace_all for fix 6 should have caught both.
# But the stock check query: SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id=$1
# The test doesn't mock this! It goes directly to INSERT.
# Let me check... actually the test mocks INSERT directly without the stock check.
# The handler does stock check BEFORE insert. So we need to add the stock check mock.

# Actually let me re-read the AddItem test to see its current state
# From the file read: lines 482-493 show:
# mock.ExpectQuery(SELECT id FROM carts WHERE user_id = \$1).WithArgs(1)...
# mock.ExpectExec(INSERT INTO cart_items.*).WithArgs(1, 1, 3)...
# Missing: SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id=$1

# Find the Cart_AddItem test and add the missing stock check mock
old7 = '''func TestCart_AddItem(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = \$1`).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec(`INSERT INTO cart_items.*`).
		WithArgs(1, 1, 3).WillReturnResult(sqlmock.NewResult(1, 1))'''
new7 = '''func TestCart_AddItem(t *testing.T) {
	cleanup, mock := setupTestEnv(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id = $1`).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"stock_quantity", "CAST(price_usd AS INTEGER)"}).AddRow(100, 10))
	mock.ExpectExec(`INSERT INTO cart_items \(cart_id, product_id, quantity\) VALUES \(\$1, \$2, \$3\) ON CONFLICT \(cart_id, product_id\) DO UPDATE SET quantity = \$3, added_at = CURRENT_TIMESTAMP`).
		WithArgs(1, 1, 3).WillReturnResult(sqlmock.NewResult(1, 1))'''
if old7 in content:
    content = content.replace(old7, new7, 1)
else:
    print("WARNING: Cart_AddItem pattern not found exactly")

# 8. TestCart_RemoveItem: DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2
old8 = '''	mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = \\$1 AND product_id = \\$2`).'''
new8 = '''	mock.ExpectExec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`).'''
content = content.replace(old8, new8, 1)

# ======== Wishlist fixes ========

# 9. TestWishlist_Toggle: EXISTS check + INSERT
# Handler: SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id=$1 AND product_id=$2)
# Handler: INSERT INTO wishlists (user_id, product_id) VALUES ($1, $2) ON CONFLICT...
old9_query = '''	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM wishlists WHERE user_id = \\$1 AND product_id = \\$2\)`).'''
new9_query = '''	mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).'''
content = content.replace(old9_query, new9_query, 1)

old9_exec = '''	mock.ExpectExec(`INSERT INTO wishlists.*`).'''
new9_exec = '''	mock.ExpectExec(`INSERT INTO wishlists \(user_id, product_id\) VALUES \(\$1, \$2\) ON CONFLICT \(user_id, product_id\) DO NOTHING`).'''
content = content.replace(old9_exec, new9_exec, 1)

# 10. TestWishlist_List: product query
# Handler: SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, '') as primary_image FROM wishlists w JOIN products p ON p.id = w.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE w.user_id = $1 ORDER BY w.created_at DESC
old10 = '''	mock.ExpectQuery(`SELECT w.product_id, p.name, p.slug.*`).'''
new10 = '''	mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, '') as primary_image FROM wishlists w JOIN products p ON p.id = w.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE w.user_id = $1 ORDER BY w.created_at DESC`).'''
content = content.replace(old10, new10, 1)

# Also fix the rows columns for wishlist
old10_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"product_id", "name", "slug", "price_usd", "image_url"}).
		AddRow(1, "Test Product", "test-product", "10.00", "http://example.com/img.png"))'''
new10_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "price_usd", "stock_quantity", "primary_image"}).
		AddRow(1, "Test Product", "test-product", "10.00", 100, "http://example.com/img.png"))'''
content = content.replace(old10_rows, new10_rows, 1)

# ======== Reviews fixes ========

# 11. TestReviews_Create: product exists check + INSERT + SELECT
# Handler: SELECT EXISTS(SELECT 1 FROM products WHERE id=$1)
# Handler: INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) ON CONFLICT...
# Handler: SELECT id, rating, COALESCE(comment, ''), created_at FROM reviews WHERE product_id=$1 AND user_id=$2
old11_exec = '''	mock.ExpectExec(`INSERT INTO reviews.*`).'''
new11_exec = '''	mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`INSERT INTO reviews \(product_id, user_id, rating, comment\) VALUES \(\$1, \$2, \$3, \$4\) ON CONFLICT \(product_id, user_id\) DO UPDATE SET rating = \$3, comment = \$4, updated_at = CURRENT_TIMESTAMP`).'''
content = content.replace(old11_exec, new11_exec, 1)

# 12. TestReviews_List: COUNT + reviews query
# Handler COUNT: SELECT COUNT(*) FROM reviews WHERE product_id=$1
# Handler query: SELECT r.id, r.rating, r.comment, r.created_at, u.id, u.name, u.email, COALESCE(up.avatar_url, '') as avatar_url FROM reviews r JOIN users u ON u.id = r.user_id LEFT JOIN user_profiles up ON up.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC LIMIT $2 OFFSET $3
old12_count = '''	mock.ExpectQuery(`SELECT.*FROM reviews.*`).'''
# The reviews test currently has:
# mock.ExpectQuery(SELECT r.id, r.rating, r.comment.*).WithArgs(1)...
# Missing: COUNT query + proper review list query with user join
old12 = '''	mock.ExpectQuery(`SELECT r.id, r.rating, r.comment.*`).'''
new12 = '''	mock.ExpectQuery(`SELECT COUNT(*) FROM reviews WHERE product_id = $1`).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at, u.id, u.name, u.email, COALESCE(up.avatar_url, '') as avatar_url FROM reviews r JOIN users u ON u.id = r.user_id LEFT JOIN user_profiles up ON up.user_id = u.id WHERE r.product_id = $1 ORDER BY r.created_at DESC LIMIT $2 OFFSET $3`).'''
content = content.replace(old12, new12, 1)

# Fix review rows columns
old12_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "rating", "comment", "created_at", "email", "name"}).
		AddRow(1, 1, 5, "Great!", mockTime(2024, time.January, 1), "test@test.com", "Test User"))'''
new12_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"id", "rating", "comment", "created_at", "id", "name", "email", "avatar_url"}).
		AddRow(1, 5, "Great!", mockTime(2024, time.January, 1), 1, "Test User", "test@test.com", ""))'''
content = content.replace(old12_rows, new12_rows, 1)

# ======== RecentlyViewed fixes ========

# 13. TestRecentlyViewed_Record: INSERT + DELETE cleanup
# Handler: INSERT INTO recently_viewed (user_id, product_id, viewed_at) VALUES ($1, $2, CURRENT_TIMESTAMP) ON CONFLICT...
# Handler: DELETE FROM recently_viewed WHERE user_id=$1 AND id NOT IN (SELECT id FROM recently_viewed WHERE user_id=$1 ORDER BY viewed_at DESC LIMIT 20)
old13 = '''	mock.ExpectExec(`INSERT INTO recently_viewed.*`).'''
new13 = '''	mock.ExpectExec(`INSERT INTO recently_viewed \(user_id, product_id, viewed_at\) VALUES \(\$1, \$2, CURRENT_TIMESTAMP\) ON CONFLICT \(user_id, product_id\) DO UPDATE SET viewed_at = CURRENT_TIMESTAMP`).'''
content = content.replace(old13, new13, 1)

# 14. TestRecentlyViewed_List: query
# Handler: SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, '') as primary_image, rv.viewed_at FROM recently_viewed rv JOIN products p ON p.id = rv.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE rv.user_id = $1 ORDER BY rv.viewed_at DESC LIMIT $2
old14 = '''	mock.ExpectQuery(`SELECT p.id, p.name, p.slug.*`).'''
new14 = '''	mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, '') as primary_image, rv.viewed_at FROM recently_viewed rv JOIN products p ON p.id = rv.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE rv.user_id = $1 ORDER BY rv.viewed_at DESC LIMIT $2`).'''
content = content.replace(old14, new14, 1)

old14_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "price_usd", "description"}).
		AddRow(1, "Test Product", "test-product", "10.00", "A test product"))'''
new14_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "price_usd", "stock_quantity", "primary_image", "viewed_at"}).
		AddRow(1, "Test Product", "test-product", "10.00", 100, "http://example.com/img.png", mockTime(2024, time.January, 1)))'''
content = content.replace(old14_rows, new14_rows, 1)

# ======== Comparison fixes ========

# 15. TestComparison_Toggle: EXISTS + INSERT
# Handler: SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id=$1 AND product_id=$2)
# Handler: INSERT INTO product_comparison (user_id, product_id) VALUES ($1, $2) ON CONFLICT...
old15_query = '''	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM product_comparison WHERE user_id = \\$1 AND product_id = \\$2\)`).'''
new15_query = '''	mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).'''
content = content.replace(old15_query, new15_query, 1)

old15_exec = '''	mock.ExpectExec(`INSERT INTO product_comparison.*`).'''
new15_exec = '''	mock.ExpectExec(`INSERT INTO product_comparison \(user_id, product_id\) VALUES \(\$1, \$2\) ON CONFLICT \(user_id, product_id\) DO NOTHING`).'''
content = content.replace(old15_exec, new15_exec, 1)

# 16. TestComparison_List: query + per-item image query
# Handler: SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.description, COALESCE(p.asset_path, '') as asset_path, COALESCE(p.asset_hash, '') as asset_hash, COALESCE(p.file_mime_type, '') as file_mime_type, COALESCE(created_at::text, '') as created_at FROM product_comparison pc JOIN products p ON p.id = pc.product_id WHERE pc.user_id = $1 ORDER BY pc.added_at DESC
# Per-item: SELECT COALESCE(url, '') FROM product_images WHERE product_id=$1 AND is_primary=true LIMIT 1
old16 = '''	mock.ExpectQuery(`SELECT p.id, p.name, p.slug.*`).'''
new16 = '''	mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.description, COALESCE(p.asset_path, '') as asset_path, COALESCE(p.asset_hash, '') as asset_hash, COALESCE(p.file_mime_type, '') as file_mime_type, COALESCE(created_at::text, '') as created_at FROM product_comparison pc JOIN products p ON p.id = pc.product_id WHERE pc.user_id = $1 ORDER BY pc.added_at DESC`).'''
content = content.replace(old16, new16, 1)

old16_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "price_usd", "description"}).
		AddRow(1, "Test Product", "test-product", "10.00", "A test product"))'''
new16_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "price_usd", "stock_quantity", "description", "asset_path", "asset_hash", "file_mime_type", "created_at"}).
		AddRow(1, "Test Product", "test-product", "10.00", 100, "A test product", "", "", "", mockTime(2024, time.January, 1)))'''
content = content.replace(old16_rows, new16_rows, 1)

# Also need to add the per-item image query mock for comparison list
# The handler does: SELECT COALESCE(url, '') FROM product_images WHERE product_id=$1 AND is_primary=true LIMIT 1
# This is called for EACH item in the list. Add it before the main query result check.
old16b = '''	mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.description, COALESCE(p.asset_path, '') as asset_path, COALESCE(p.asset_hash, '') as asset_hash, COALESCE(p.file_mime_type, '') as file_mime_type, COALESCE(created_at::text, '') as created_at FROM product_comparison pc JOIN products p ON p.id = pc.product_id WHERE pc.user_id = $1 ORDER BY pc.added_at DESC`).'''
new16b = '''	mock.ExpectQuery(`SELECT COALESCE(url, '') FROM product_images WHERE product_id = $1 AND is_primary = true LIMIT 1`).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"COALESCE(url, '')"}).AddRow("http://example.com/img.png"))
	mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.description, COALESCE(p.asset_path, '') as asset_path, COALESCE(p.asset_hash, '') as asset_hash, COALESCE(p.file_mime_type, '') as file_mime_type, COALESCE(created_at::text, '') as created_at FROM product_comparison pc JOIN products p ON p.id = pc.product_id WHERE pc.user_id = $1 ORDER BY pc.added_at DESC`).'''
content = content.replace(old16b, new16b, 1)

# ======== Recommendations fixes ========

# 17. TestRecommendations_Get: category lookup + product query (same category or global)
# Handler: SELECT category_id FROM products WHERE id=$1
# If catID.Valid: SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, '') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.category_id = $1 AND p.id != $2 AND p.status = 'active' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $3 OFFSET $4
# Else: SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, '') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.id != $1 AND p.status = 'active' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $2 OFFSET $3
old17_cat = '''	mock.ExpectQuery(`SELECT category_id FROM products WHERE id = \\$1`).'''
new17_cat = '''	mock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).'''
content = content.replace(old17_cat, new17_cat, 1)

# Fix the same-category recommendation query (catID.Valid = true path, but test sets nil...)
# Actually the test sets AddRow(nil) for category_id, which means catID.Valid=false, so it takes the global path
# Global path: WHERE p.id != $1 AND p.status = 'active' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $2 OFFSET $3
old17_query = '''	mock.ExpectQuery(`SELECT p.id, p.title.*FROM products p.*`).'''
new17_query = '''	mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, p.status, COALESCE(pi.url, '') as primary_image FROM products p LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE p.id != $1 AND p.status = 'active' ORDER BY p.views_count DESC, p.created_at DESC LIMIT $2 OFFSET $3`).'''
content = content.replace(old17_query, new17_query, 1)

old17_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "price_usd", "description"}).
		AddRow(2, "Related Product", "related-product", "15.00", "Related to test"))'''
new17_rows = '''		WithArgs(1, 12, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "price_usd", "stock_quantity", "status", "primary_image"}).
		AddRow(2, "Related Product", "related-product", "15.00", 50, "active", "http://example.com/img.png"))'''
content = content.replace(old17_rows, new17_rows, 1)

# Remove the second mock query for recommendations (the test has an extra one for the product itself)
old17_extra = '''	mock.ExpectQuery(`SELECT p.id, p.name.*FROM products p.*`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "category_name"}).
		AddRow(1, "Test Product", "test-product", "Icons"))'''
# Remove this entirely - it's not needed
content = content.replace(old17_extra, '', 1)

# Fix the recommendations WithArgs - the global query needs (productID, perPage, offset)
old17_args = '''		WithArgs(1).'''
# This appears in the recommendations test after the query fix. Need to find it in context.
# Actually the WithArgs(1) is for the cat lookup, which is correct. The product query needs WithArgs(1, 12, 0)

# ======== Guest Order fixes ========

# 18. TestGuestOrder_Create: SUM query + INSERT into orders + SELECT
# Handler SUM: SELECT COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0) FROM products WHERE id = ANY($1::int[]) AND status = 'active'
# Handler INSERT: INSERT INTO orders (email, total_usd, crypto_chain, status, guest_order) VALUES ($1, $2, $3, 'pending', true) RETURNING id
# Note: test has WithArgs("test@example.com", "10.00", "BSC") but handler does INSERT INTO orders (email, total_usd, crypto_chain, status, guest_order) VALUES ($1, $2, $3, 'pending', true)
# The test's mock has wrong args count - handler uses $1=$email, $2=$totalUSD, $3=$cryptoChain
old18_sum = '''	mock.ExpectExec(`INSERT INTO guest_orders.*`).'''
new18_sum = '''	mock.ExpectQuery(`SELECT COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0) FROM products WHERE id = ANY($1::int[]) AND status = 'active'`).
		WithArgs([]int{1}).WillReturnRows(sqlmock.NewRows([]string{"COALESCE(SUM(CAST(price_usd AS NUMERIC)), 0)"}).AddRow("10.00"))
	mock.ExpectExec(`INSERT INTO orders \(email, total_usd, crypto_chain, status, guest_order\) VALUES \(\$1, \$2, \$3, 'pending', true\) RETURNING id`).'''
content = content.replace(old18_sum, new18_sum, 1)

# 19. TestGuestOrder_Create: SELECT after insert
# Handler: SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = $1 AND guest_order = true
old18_select = '''	mock.ExpectQuery(`SELECT id, email, status.*FROM guest_orders WHERE id = \\$1`).'''
new18_select = '''	mock.ExpectQuery(`SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = $1 AND guest_order = true`).'''
content = content.replace(old18_select, new18_select, 1)

old18_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "status", "total_usd", "crypto_chain", "created_at"}).
		AddRow(1, "test@example.com", "pending", "10.00", "BSC", mockTime(2024, time.January, 1)))'''
new18_rows = '''		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"email", "status", "total_usd", "crypto_chain"}).
		AddRow("test@example.com", "pending", "10.00", "BSC"))'''
content = content.replace(old18_rows, new18_rows, 1)

# 20. TestGuestOrder_Check: SELECT
# Handler: SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = $1 AND guest_order = true
old20 = '''	mock.ExpectQuery(`SELECT id, email, status.*FROM guest_orders WHERE id = \\$1`).'''
new20 = '''	mock.ExpectQuery(`SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = $1 AND guest_order = true`).'''
content = content.replace(old20, new20, 1)

old20_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "status", "total_usd", "crypto_chain", "created_at"}).
		AddRow(1, "test@example.com", "completed", "10.00", "BSC", mockTime(2024, time.January, 1)))'''
new20_rows = '''		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"email", "status", "total_usd", "crypto_chain"}).
		AddRow("test@example.com", "completed", "10.00", "BSC"))'''
content = content.replace(old20_rows, new20_rows, 1)

# 21. TestOrderStatusCheck: SELECT
# Handler: SELECT id, status, total_usd, crypto_chain, payment_tx_hash, payment_confirmations, paid_at FROM orders WHERE id = $1
old21 = '''	mock.ExpectQuery(`SELECT id, status, total_usd.*FROM orders WHERE id = \\$1`).'''
new21 = '''	mock.ExpectQuery(`SELECT id, status, total_usd, crypto_chain, payment_tx_hash, payment_confirmations, paid_at FROM orders WHERE id = $1`).'''
content = content.replace(old21, new21, 1)

old21_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "total_usd", "crypto_chain", "payment_tx_hash", "payment_confirmations", "paid_at"}).
		AddRow(1, "paid", "10.00", "BSC", "0xabc...", 12, mockTime(2024, time.January, 1)))'''
# This one looks correct actually - same columns. But need WithArgs
old21_withargs = '''		WithArgs(1).
		WillReturnRows'''
# It might already have WithArgs. Let me check... Looking at line 672-674:
# mock.ExpectQuery(SELECT id, status, total_usd.*FROM orders WHERE id = \$1).
#   WithArgs(1).
#   WillReturnRows(...)
# So WithArgs is there. The issue is the SQL pattern match.

# ======== Admin fixes ========

# 22. TestAdmin_BulkProductStatus: UPDATE products SET...
# Handler builds dynamic UPDATE: UPDATE products SET status = $1 WHERE id = $2 (for each product)
# The test sends status="active", ids="1,2" which becomes product_ids=[{ProductID:1},{ProductID:2}]
# Handler loops over IDs and does individual UPDATEs
# Test expects 1 Exec but handler does 2 (one per product ID)
old22 = '''	mock.ExpectExec(`UPDATE products SET.*`).
		WillReturnResult(sqlmock.NewResult(2, 2))'''
new22 = '''	mock.ExpectExec(`UPDATE products SET status = \$1 WHERE id = \$2`).
		WithArgs("active", 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE products SET status = \$1 WHERE id = \$2`).
		WithArgs("active", 2).WillReturnResult(sqlmock.NewResult(1, 1))'''
content = content.replace(old22, new22, 1)

# 23. TestAdmin_OrderStatusPaid: UPDATE orders SET...
# Handler: SELECT status FROM orders WHERE id=$1 (to check current status)
# Then: UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2
# Test only mocks the UPDATE, missing the SELECT
old23 = '''	mock.ExpectExec(`UPDATE orders SET.*`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, result := R("PUT", "/api/v1/admin/orders/1/status/paid", nil'''
new23 = '''	mock.ExpectQuery(`SELECT status FROM orders WHERE id = $1`).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("pending"))
	mock.ExpectExec(`UPDATE orders SET status = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).
		WithArgs("paid", 1).WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, result := R("PUT", "/api/v1/admin/orders/1/status/paid", nil'''
content = content.replace(old23, new23, 1)

# 24. TestAdmin_OrderStatusCompleted
old24 = '''	mock.ExpectExec(`UPDATE orders SET.*`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("PUT", "/api/v1/admin/orders/1/status/completed"'''
new24 = '''	mock.ExpectQuery(`SELECT status FROM orders WHERE id = $1`).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("paid"))
	mock.ExpectExec(`UPDATE orders SET status = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).
		WithArgs("completed", 1).WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("PUT", "/api/v1/admin/orders/1/status/completed"'''
content = content.replace(old24, new24, 1)

# 25. TestAdmin_OrderStatusCancelled
old25 = '''	mock.ExpectExec(`UPDATE orders SET.*`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("PUT", "/api/v1/admin/orders/1/status/cancelled"'''
new25 = '''	mock.ExpectQuery(`SELECT status FROM orders WHERE id = $1`).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("pending"))
	mock.ExpectExec(`UPDATE orders SET status = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).
		WithArgs("cancelled", 1).WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("PUT", "/api/v1/admin/orders/1/status/cancelled"'''
content = content.replace(old25, new25, 1)

# 26. TestAdmin_OrderStatusRefunded
old26 = '''	mock.ExpectExec(`UPDATE orders SET.*`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("PUT", "/api/v1/admin/orders/1/status/refunded"'''
new26 = '''	mock.ExpectQuery(`SELECT status FROM orders WHERE id = $1`).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("paid"))
	mock.ExpectExec(`UPDATE orders SET status = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).
		WithArgs("refunded", 1).WillReturnResult(sqlmock.NewResult(1, 1))
	code, _, _ := R("PUT", "/api/v1/admin/orders/1/status/refunded"'''
content = content.replace(old26, new26, 1)

# 27. TestAdmin_BulkProductCategory: UPDATE products SET...
# Similar to bulk status - handler loops and does individual UPDATEs
old27 = '''	mock.ExpectExec(`UPDATE products SET.*`).
		WillReturnResult(sqlmock.NewResult(1, 1))'''
new27 = '''	mock.ExpectExec(`UPDATE products SET category_id = \$1 WHERE id = \$2`).
		WithArgs(2, 1).WillReturnResult(sqlmock.NewResult(1, 1))'''
content = content.replace(old27, new27, 1)

# 28. TestAdmin_BulkProductToggle: UPDATE products SET...
old28 = '''	mock.ExpectExec(`UPDATE products SET.*`).
		WithArgs("active", 1).WillReturnResult(sqlmock.NewResult(1, 1))'''
# Wait, let me check... the test sends status="active", ids="1"
# Handler: UPDATE products SET status = $1 WHERE id = $2
new28 = '''	mock.ExpectExec(`UPDATE products SET status = \$1 WHERE id = \$2`).
		WithArgs("active", 1).WillReturnResult(sqlmock.NewResult(1, 1))'''
content = content.replace(old28, new28, 1)

# 29. TestAdmin_DeleteUser: DELETE FROM users WHERE id = $1
old29 = '''	mock.ExpectExec(`DELETE FROM users WHERE id = \\$1`).'''
new29 = '''	mock.ExpectExec(`DELETE FROM users WHERE id = $1`).'''
content = content.replace(old29, new29, 1)

# 30. TestAdmin_DeleteCommunityPost: DELETE FROM community_posts WHERE id = $1
old30 = '''	mock.ExpectExec(`DELETE FROM community_posts WHERE id = \\$1`).'''
new30 = '''	mock.ExpectExec(`DELETE FROM community_posts WHERE id = $1`).'''
content = content.replace(old30, new30, 1)

# 31. TestAdmin_CommunityPostList: query
# Handler: SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, COALESCE(up.avatar_url, '') as avatar_url, (SELECT COUNT(*) FROM community_likes WHERE post_id = cp.id) as count_likes, (SELECT COUNT(*) FROM community_comments WHERE post_id = cp.id) as count_comments, EXISTS(SELECT 1 FROM community_follows WHERE follower_id = $2 AND following_id = cp.user_id) as following FROM community_posts cp JOIN users u ON u.id = cp.user_id LEFT JOIN user_profiles up ON up.user_id = u.id ORDER BY cp.created_at DESC LIMIT $3 OFFSET $4
# The test mock has: SELECT cp.id, cp.user_id.*FROM community_posts cp.* 
# This won't match the complex query with subqueries
old31 = '''	mock.ExpectQuery(`SELECT cp.id, cp.user_id.*FROM community_posts cp.*`).'''
new31 = '''	mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, COALESCE(up.avatar_url, '') as avatar_url, \(SELECT COUNT\(\*\) FROM community_likes WHERE post_id = cp.id\) as count_likes, \(SELECT COUNT\(\*\) FROM community_comments WHERE post_id = cp.id\) as count_comments, EXISTS\(SELECT 1 FROM community_follows WHERE follower_id = \$2 AND following_id = cp.user_id\) as following FROM community_posts cp JOIN users u ON u.id = cp.user_id LEFT JOIN user_profiles up ON up.user_id = u.id ORDER BY cp.created_at DESC LIMIT \$3 OFFSET \$4`).'''
content = content.replace(old31, new31, 1)

# Fix the rows for community post list
old31_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "created_at", "email", "name", "avatar_url", "count_likes", "count_comments", "following"}).
		AddRow(1, 1, "Post 1", mockTime(2024, time.January, 1), "test@test.com", "Test User", "", 0, 0, false))'''
new31_rows = '''		WithArgs(1, 1, 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "created_at", "email", "name", "avatar_url", "count_likes", "count_comments", "following"}).
		AddRow(1, 1, "Post 1", mockTime(2024, time.January, 1), "test@test.com", "Test User", "", 0, 0, false))'''
content = content.replace(old31_rows, new31_rows, 1)

# 32. TestAdmin_CommunityUserList: COUNT + users query
# Handler COUNT: SELECT COUNT(*) FROM users
# Handler query: SELECT u.id, u.email, u.name, COALESCE(up.bio, '') as bio, COALESCE(up.avatar_url, '') as avatar_url, (SELECT COUNT(*) FROM community_follows WHERE follower_id = u.id) as followers_count, (SELECT COUNT(*) FROM community_follows WHERE following_id = u.id) as following_count, (SELECT COUNT(*) FROM community_posts WHERE user_id = u.id) as posts_count FROM users u LEFT JOIN user_profiles up ON up.user_id = u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2
old32_count = '''	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM users`).'''
new32_count = '''	mock.ExpectQuery(`SELECT COUNT(*) FROM users`).'''
content = content.replace(old32_count, new32_count, 1)

old32_query = '''	mock.ExpectQuery(`SELECT u.id, u.email, u.name.*`).'''
new32_query = '''	mock.ExpectQuery(`SELECT u.id, u.email, u.name, COALESCE(up.bio, '') as bio, COALESCE(up.avatar_url, '') as avatar_url, \(SELECT COUNT\(\*\) FROM community_follows WHERE follower_id = u.id\) as followers_count, \(SELECT COUNT\(\*\) FROM community_follows WHERE following_id = u.id\) as following_count, \(SELECT COUNT\(\*\) FROM community_posts WHERE user_id = u.id\) as posts_count FROM users u LEFT JOIN user_profiles up ON up.user_id = u.id ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`).'''
content = content.replace(old32_query, new32_query, 1)

# Fix rows for community user list
old32_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "count"}).
		AddRow(1, "test@test.com", "Test User", 10).
		AddRow(2, "user2@test.com", "User 2", 5))'''
new32_rows = '''		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name", "bio", "avatar_url", "followers_count", "following_count", "posts_count"}).
		AddRow(1, "test@test.com", "Test User", "", "", 0, 0, 5).
		AddRow(2, "user2@test.com", "User 2", "", "", 0, 0, 3))'''
content = content.replace(old32_rows, new32_rows, 1)

# 33. TestAdmin_GuestOrderCheck: SELECT
# Handler: SELECT id, email, status, total_usd, crypto_chain, payment_tx_hash, payment_confirmations, paid_at, created_at FROM orders WHERE id = $1 AND guest_order = true
old33 = '''	mock.ExpectQuery(`SELECT o.id, o.status.*FROM orders WHERE id = \\$1`).'''
new33 = '''	mock.ExpectQuery(`SELECT id, email, status, total_usd, crypto_chain, payment_tx_hash, payment_confirmations, paid_at, created_at FROM orders WHERE id = $1 AND guest_order = true`).'''
content = content.replace(old33, new33, 1)

old33_rows = '''		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "total_usd", "created_at"}).
		AddRow(1, "completed", "10.00", mockTime(2024, time.January, 1)))'''
new33_rows = '''		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "status", "total_usd", "crypto_chain", "payment_tx_hash", "payment_confirmations", "paid_at", "created_at"}).
		AddRow(1, "test@example.com", "completed", "10.00", "BSC", "", 0, mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))'''
content = content.replace(old33_rows, new33_rows, 1)

# 34. TestAdmin_ExchangeRateList: SELECT
# Handler: SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain
old34 = '''	mock.ExpectQuery(`SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain`).'''
# This one is already correct! No fix needed.

# 35. TestAdmin_ExchangeRateSet: INSERT
# Handler: INSERT INTO exchange_rates (chain, rate) VALUES ($1, $2) ON CONFLICT (chain) DO UPDATE SET rate = $2, updated_at = CURRENT_TIMESTAMP
old35 = '''	mock.ExpectExec(`INSERT INTO exchange_rates.*`).'''
new35 = '''	mock.ExpectExec(`INSERT INTO exchange_rates \(chain, rate\) VALUES \(\$1, \$2\) ON CONFLICT \(chain\) DO UPDATE SET rate = \$2, updated_at = CURRENT_TIMESTAMP`).'''
content = content.replace(old35, new35, 1)

# 36. TestAdmin_SetSetting: INSERT
# Handler: INSERT INTO settings (setting_key, value) VALUES ($1, $2) ON CONFLICT (setting_key) DO UPDATE SET value = $2, updated_at = CURRENT_TIMESTAMP
old36 = '''	mock.ExpectExec(`INSERT INTO settings.*`).'''
new36 = '''	mock.ExpectExec(`INSERT INTO settings \(setting_key, value\) VALUES \(\$1, \$2\) ON CONFLICT \(setting_key\) DO UPDATE SET value = \$2, updated_at = CURRENT_TIMESTAMP`).'''
content = content.replace(old36, new36, 1)

# 37. TestAdmin_GetSetting: SELECT
# Handler: SELECT value FROM settings WHERE setting_key = $1
old37 = '''	mock.ExpectQuery(`SELECT value FROM settings WHERE key = \\$1`).'''
new37 = '''	mock.ExpectQuery(`SELECT value FROM settings WHERE setting_key = $1`).'''
content = content.replace(old37, new37, 1)

# 38. TestAdmin_Stats: COUNT queries + revenue
# Handler: SELECT COUNT(*) FROM users
# Handler: SELECT COUNT(*) FROM orders
# Handler: SELECT COALESCE(SUM(total_usd)::TEXT, '0.00') FROM orders WHERE status = 'completed'
old38a = '''	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM users`).'''
new38a = '''	mock.ExpectQuery(`SELECT COUNT(*) FROM users`).'''
content = content.replace(old38a, new38a, 1)

old38b = '''	mock.ExpectQuery(`SELECT COUNT\(\\*) FROM orders`).'''
new38b = '''	mock.ExpectQuery(`SELECT COUNT(*) FROM orders`).'''
content = content.replace(old38b, new38b, 1)

old38c = '''	mock.ExpectQuery(`SELECT COALESCE.*`).'''
new38c = '''	mock.ExpectQuery(`SELECT COALESCE(SUM(total_usd)::TEXT, '0.00') FROM orders WHERE status = 'completed'`).'''
content = content.replace(old38c, new38c, 1)

# Fix the revenue row - handler returns TEXT not numeric
old38_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow("100.00"))'''
new38_rows = '''		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"COALESCE(SUM(total_usd)::TEXT, '0.00')"}).AddRow("100.00"))'''
content = content.replace(old38_rows, new38_rows, 1)

# 39. TestAdmin_ListUsers: SELECT
# Handler: SELECT id, email, name, created_at, updated_at FROM users ORDER BY created_at DESC
# This one looks correct already.

# 40. TestAdmin_GetSetting: already fixed above (setting_key)

# 41. TestSettings_GetAll: SELECT
# Handler: SELECT setting_key, value, updated_at FROM settings ORDER BY setting_key
old41 = '''	mock.ExpectQuery(`SELECT key, value, updated_at FROM settings ORDER BY key`).'''
new41 = '''	mock.ExpectQuery(`SELECT setting_key, value, updated_at FROM settings ORDER BY setting_key`).'''
content = content.replace(old41, new41, 1)

# Fix rows for settings
old41_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"key", "value", "updated_at"}).
		AddRow("payment_address", "0xTestAddress", mockTime(2024, time.January, 1)))'''
new41_rows = '''		WillReturnRows(sqlmock.NewRows([]string{"setting_key", "value", "updated_at"}).
		AddRow("payment_address", "0xTestAddress", mockTime(2024, time.January, 1)))'''
content = content.replace(old41_rows, new41_rows, 1)

# 42. TestExchangeRates_GetAll: SELECT
# Handler: SELECT chain, rate, updated_at FROM exchange_rates ORDER BY chain
# Already correct.

# 43. TestExchangeRates_GetOne: SELECT
# Handler: SELECT chain, rate, updated_at FROM exchange_rates WHERE chain = $1
# Already correct.

# 44. TestDBSchema_Migration011Tables: SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1
# Already correct.

# 45. TestAuth_UpdateProfile_Valid: SELECT + SELECT + UPDATE + SELECT
# Handler: SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1
# Handler: SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1
# Handler: UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2
# Handler: INSERT INTO user_profiles (user_id, bio, wallet_address) VALUES ($1, $2, $3) ON CONFLICT...
# The test has UPDATE users SET.* which won't match the specific query
old45_update = '''	mock.ExpectExec(`UPDATE users SET.*`).'''
new45_update = '''	mock.ExpectExec(`UPDATE users SET name = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).'''
content = content.replace(old45_update, new45_update, 1)

# Add profile insert mock (handler does INSERT INTO user_profiles ON CONFLICT DO UPDATE)
old45_profile = '''	mock.ExpectQuery(`SELECT bio, wallet_address FROM user_profiles WHERE user_id = \\$1`).'''
new45_profile = '''	mock.ExpectQuery(`SELECT bio, wallet_address FROM user_profiles WHERE user_id = $1`).'''
content = content.replace(old45_profile, new45_profile, 1)

# The UPDATE users mock needs WithArgs
old45_update_args = '''	mock.ExpectExec(`UPDATE users SET name = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).'''
new45_update_args = '''	mock.ExpectExec(`UPDATE users SET name = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).
		WithArgs("New Name", 1).'''
# Wait, the original test already has WillReturnResult right after. Let me check the exact pattern.
# From the file: mock.ExpectExec(UPDATE users SET.*).WillReturnResult(sqlmock.NewResult(1, 1))
# After my fix it becomes: mock.ExpectExec(UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2).WillReturnResult(...)
# But it needs WithArgs("New Name", 1) before WillReturnResult
# Let me fix this more carefully
old45_full = '''	mock.ExpectExec(`UPDATE users SET name = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).
		WillReturnResult(sqlmock.NewResult(1, 1))'''
new45_full = '''	mock.ExpectExec(`UPDATE users SET name = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).
		WithArgs("New Name", 1).WillReturnResult(sqlmock.NewResult(1, 1))'''
content = content.replace(old45_full, new45_full, 1)

# Also need to add profile INSERT mock for UpdateProfile
# Handler does: INSERT INTO user_profiles (user_id, bio, wallet_address) VALUES ($1, $2, $3) ON CONFLICT (user_id) DO UPDATE SET bio = $2, wallet_address = $3
# The test doesn't mock this! Need to add it.
old45_insert_profile = '''	mock.ExpectExec(`UPDATE users SET name = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).
		WithArgs("New Name", 1).WillReturnResult(sqlmock.NewResult(1, 1))'''
new45_insert_profile = '''	mock.ExpectExec(`UPDATE users SET name = \$1, updated_at = CURRENT_TIMESTAMP WHERE id = \$2`).
		WithArgs("New Name", 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO user_profiles \(user_id, bio, wallet_address\) VALUES \(\$1, \$2, \$3\) ON CONFLICT \(user_id\) DO UPDATE SET bio = \$2, wallet_address = \$3`).
		WithArgs(1, nil, nil).WillReturnResult(sqlmock.NewResult(1, 1))'''
content = content.replace(old45_insert_profile, new45_insert_profile, 1)

with open(SRC, "w") as f:
    f.write(content)
print("All fixes applied")
