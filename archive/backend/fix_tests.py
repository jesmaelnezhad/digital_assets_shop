#!/usr/bin/env python3
"""
Fix all_features_test.go: align mocks with actual handler SQL queries.
The handlers use these exact column names and SQL patterns.
This script applies targeted patches.
"""
import re, sys

PATH = "/root/project/backend/handlers/all_features_test.go"

with open(PATH) as f:
    content = f.read()

fixes = 0

# Helper: apply a regex replace, count if changed
def patch(old, new, label=""):
    global content, fixes
    if old in content:
        content = content.replace(old, new)
        fixes += 1
        print(f"  FIXED: {label}")
        return True
    print(f"  SKIP (not found): {label}")
    return False

# ====== 1. Fix column name mismatches across all mocks ======
# products.go uses: title, price_usd, stock_quantity, category_id (not "name", "price", "stock")
# Wishlist mock: "title" -> "name", "price_usd" -> "price", "stock_quantity" -> "stock"
# Lines 541-544: Wishlist_List mock
old = '''mock.ExpectQuery(`SELECT w.product_id, p.title, p.slug.*FROM wishlists w.*`).
\t\tWithArgs(1).
\t\tWillReturnRows(sqlmock.NewRows([]string{"product_id", "title", "slug", "price_usd", "image_url"}).
\t\t\tAddRow(1, "Test Product", "test-product", "10.00", "http://example.com/img.png"))'''
new = '''mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST\(p.price_usd AS NUMERIC\) as price_usd, p.stock_quantity,.*FROM wishlists w.*FROM products p.*` threaded through wishlists.*).
\t\tWithArgs(1).
\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "price_usd", "stock_quantity", "primary_image"}).
\t\t\tAddRow(1, "Test Product", "test-product", "10.00", 100, "http://example.com/img.png"))'''
if old in content:
    content = content.replace(old, new)
    fixes += 1
    print("  FIXED: Wishlist_List column names")
else:
    print("  SKIP: Wishlist_List (check manually)")

# Reviews mock: "comment" -> "comment" is fine, but columns differ
# Lines 568-571: Reviews_List mock
old = '''mock.ExpectQuery(`SELECT r.id, r.product_id, r.rating, r.comment.*FROM reviews r.*`).
\t\tWithArgs(1).
\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "rating", "comment", "created_at", "email", "name"}).
\t\t\tAddRow(1, 1, 5, "Great!", mockTime(2024, time.January, 1), "test@test.com", "Test User"))'''
new = '''mock.ExpectQuery(`SELECT r.id, r.rating, r.comment, r.created_at,.*FROM reviews r.*JOIN users u ON u.id = r.user_id.*LEFT JOIN user_profiles up.*` threaded).
\t\tWithArgs(1, 10, 0).
\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "rating", "comment", "created_at", "id2", "name", "email", "avatar_url"}).
\t\t\tAddRow(1, 5, "Great!", mockTime(2024, time.January, 1), 1, "Test User", "test@test.com", ""))'''
if old in content:
    content = content.replace(old, new)
    fixes += 1
    print("  FIXED: Reviews_List SQL and columns")
else:
    print("  SKIP: Reviews_List (check manually)")

# RecentlyViewed mock: columns mismatch (title, slug, price_usd, description vs actual)
# Lines 593-596
old = '''mock.ExpectQuery(`SELECT p.id, p.title, p.slug.*FROM recently_viewed rv.*`).
\t\tWithArgs(1).
\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "price_usd", "description"}).
\t\t\tAddRow(1, "Test Product", "test-product", "10.00", "A test product"))'''
new = '''mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST\(p.price_usd AS NUMERIC\) as price_usd, p.stock_quantity,.*FROM recently_viewed rv.*JOIN products p.*LEFT JOIN product_images pi.*` threaded).
\t\tWithArgs(1, 20).
\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "price_usd", "stock_quantity", "primary_image", "viewed_at"}).
\t\t\tAddRow(1, "Test Product", "test-product", "10.00", 100, "http://example.com/img.png", mockTime(2024, time.January, 1).String()))'''
if old in content:
    content = content.replace(old, new)
    fixes += 1
    print("  FIXED: RecentlyViewed_List SQL and columns")
else:
    print("  SKIP: RecentlyViewed_List (check manually)")

# Comparison mock: columns mismatch
# Lines 622-625
old = '''mock.ExpectQuery(`SELECT p.id, p.title, p.slug.*FROM product_comparison pc.*`).
\t\tWithArgs(1).
\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "price_usd", "description"}).
\t\t\tAddRow(1, "Test Product", "test-product", "10.00", "A test product"))'''
new = '''mock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST\(p.price_usd AS NUMERIC\) as price_usd, p.stock_quantity, p.description,.*FROM product_comparison pc.*JOIN products p.*` threaded).
\t\tWithArgs(1).
\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "price_usd", "stock_quantity", "description", "asset_path", "asset_hash", "file_mime_type", "created_at"}).
\t\t\tAddRow(1, "Test Product", "test-product", "10.00", 100, "A test product", "", "", "", ""))'''
if old in content:
    content = content.replace(old, new)
    fixes += 1
    print("  FIXED: Comparison_List SQL and columns")
else:
    print("  SKIP: Comparison_List (check manually)")

# Recommendations mock: SQL structure mismatch (two actual queries vs one mock)
# Lines 639-646: should match actual: SELECT category_id FROM products WHERE id=$1, then either same-cat or global query
# Also adds a third query for product_images
old = '''mock.ExpectQuery(`SELECT category_id FROM products WHERE id=\\$1`).
\t\tWithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"category_id"}).AddRow(nil))
\tmock.ExpectQuery(`SELECT p.id, p.title.*FROM products p.*`).
\t\tWithArgs(1).
\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "price_usd", "description"}).
\t\t\tAddRow(2, "Related Product", "related-product", "15.00", "Related to test"))
\tmock.ExpectQuery(`SELECT p.id, p.title, COALESCE\\(c.name, ''\\) as category_name.*FROM products p.*`).
\t\tWithArgs(1).
\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "title", "category_name"}).
\t\t\tAddRow(1, "Test Product", "Icons"))'''
new = '''mock.ExpectQuery(`SELECT category_id FROM products WHERE id=\\$1`).
\t\tWithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"category_id"}).AddRow(nil))
\tmock.ExpectQuery(`SELECT p.id, p.title, p.slug, CAST\(p.price_usd AS NUMERIC\) as price_usd, p.stock_quantity, p.status,.*FROM products p.*LEFT JOIN product_images pi.*` threaded).
\t\tWithArgs(1, 1, 8, 0).
\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "price_usd", "stock_quantity", "status", "primary_image"}).
\t\t\tAddRow(2, "Related Product", "related-product", "15.00", 50, "active", ""))
\tmock.ExpectQuery(`SELECT COALESCE\(url, ''\\) FROM product_images WHERE product_id=\\$1 AND is_primary=true LIMIT 1`).
\t\tWithArgs(2).WillReturnRows(sqlmock.NewRows([]string{"url"}).AddRow(""))'''
if old in content:
    content = content.replace(old, new)
    fixes += 1
    print("  FIXED: Recommendations_Get SQL structure")
else:
    print("  SKIP: Recommendations_Get (check manually)")

# ====== 2. Fix order status check mock ======
# Lines 686-689: "status" -> "status" is fine but columns include payment_tx_hash, payment_confirmations, paid_at
# This looks OK already - handler uses "status" column. Keep as-is.

# ====== 3. Fix products list mock - add missing image query ======
# TestProducts_GetProducts already has 3 mocks; the handler also queries cart items count? No.
# The handler does: count query, products query, image query per product. Mock has all 3. OK.

# TestProducts_GetProductBySlug: same structure, 3 mocks. OK.

# TestProducts_SearchProducts: 3 mocks. OK.

# ====== 4. Fix admin orders payment mock - column name "status" present? ======
# Lines 802-805: handler queries "SELECT o.id, o.status..." - mock has "id", "status". OK but add missing columns
# The handler scans: id, status, total_usd, crypto_chain, payment_tx_hash, payment_confirmations, paid_at
# Mock provides: id, status, total_usd, crypto_chain, payment_tx_hash, payment_confirmations, paid_at - matches!
# But the handler also queries order_items and products price and category_name
# Lines 806-813: 3 more queries. Mock has them. Should be OK.

# ====== 5. Fix admin products list mock ======
# Lines 843-846: handler's NewProducts_List queries different columns than mock
# Handler (products.go List): SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name,''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at
# Mock provides: id, title, slug, category_name, category_slug, price_usd, crypto_amount, crypto_currency, status, stock_quantity
# MISMATCH - need to fix mock to match handler
old = '''mock.ExpectQuery(`SELECT p.id, p.title.*FROM products p.*`).
\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "category_name", "category_slug", "price_usd", "crypto_amount", "crypto_currency", "status", "stock_quantity"}).
\t\t\tAddRow(1, "Test Product", "test-product", "Icons", "icons", "10.00", "0.03", "BSC", "active", 100).
\t\t\tAddRow(2, "Another Product", "another-product", "", "", "20.00", "0.05", "BSC", "active", 50))'''
new = '''mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE\\(c.name, ''\\) as category_name, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE 1=1` threaded).
\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "description", "category_id", "category_name", "price_usd", "asset_path", "asset_hash", "status", "download_count_limit", "max_downloads_per_user", "file_size_bytes", "file_mime_type", "created_at", "updated_at"}).
\t\t\tAddRow(1, "Test Product", "test-product", "A test product", nil, "Icons", "10.00", "", "", "active", 100, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)).
\t\t\tAddRow(2, "Another Product", "another-product", "Another test", nil, "", "20.00", "", "", "active", 50, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))'''
if old in content:
    content = content.replace(old, new)
    fixes += 1
    print("  FIXED: NewProducts_List SQL and columns")
else:
    print("  SKIP: NewProducts_List (check manually)")

# ====== 6. Settings_Get: handler uses "key" column, mock uses "key" - should match ======
# Handler: SELECT value FROM settings WHERE key = $1
# Mock: SELECT value FROM settings WHERE key = $1  -- matches!
# But test returns 404... let me check the route. The test uses "/api/v1/settings/payment_address"
# Handler registers: settings.GET("/:key", GetSetting)  -->  /api/v1/settings/payment_address
# In R(), the path passed is "/api/v1/settings/payment_address" -- should match.
# Wait - R() doesn't register settings routes! Let me check test_helpers.go R() function.

# ====== 7. GuestOrder mocks: handler creates order with more fields ======
# Lines 656-661: handler INSERT INTO guest_orders includes more columns
# Let me check the actual handler

# ====== 8. DBSchema test: handler queries information_schema, mock queries information_schema - should match
# But returns 404... route issue again?

# ====== 9. Admin stats: 404 means route not registered in R() ======

# ====== 10. ExchangeRates: 404 means route not registered in R() ======

# ====== 11. Cart: 404 means route not registered in R() ======

# The pattern is clear: many 404s because R() doesn't register those routes.
# Let me check what R() registers vs what the tests expect.

with open(PATH, "w") as f:
    f.write(content)

print(f"\nTotal fixes applied: {fixes}")
print("Now check test_helpers.go R() for missing route registrations.")
