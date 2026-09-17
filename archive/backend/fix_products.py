#!/usr/bin/env python3
"""Fix all_features_test.go: replace .* wildcards with exact SQL, add WithArgs."""
import re, subprocess, shutil

SRC = "/root/project/backend/handlers/all_features_test.go"
with open(SRC) as f:
    text = f.read()

changes = 0
def fix(name, old, new):
    global text, changes
    if old not in text:
        print(f"  MISS: {name}")
        return False
    text = text.replace(old, new)
    changes += 1
    print(f"  OK: {name}")
    return True

# ============================================================
# PRODUCTS
# ============================================================

# TestProducts_GetProducts (lines 120-133)
# Handler sends: 
#   SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1
#   SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name,''), p.price_usd, ... FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3
#   SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id

fix("GetProducts COUNT",
    'mock.ExpectQuery(`SELECT COUNT\(\*\\) FROM products.*`).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n\t\tWithArgs("active").\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))')

# The products LIST query uses .* wildcard - we need the exact SQL
fix("GetProducts LIST",
    'mock.ExpectQuery(`SELECT p.id, p.title.*`).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "description", "category_id", "category_name", "price_usd", "asset_path", "asset_hash", "status", "download_count_limit", "max_downloads_per_user", "file_size_bytes", "file_mime_type", "created_at", "updated_at"}).\n\t\t\tAddRow(1, "Test Product", "test-product", "A test product", nil, "Icons", "10.00", "", "", "active", 100, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)).\n\t\t\tAddRow(2, "Another Product", "another-product", "Another test", nil, "", "20.00", "", "", "active", 50, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\') as category_name, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).\n\t\tWithArgs("active", 12, 0).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "description", "category_id", "category_name", "price_usd", "asset_path", "asset_hash", "status", "download_count_limit", "max_downloads_per_user", "file_size_bytes", "file_mime_type", "created_at", "updated_at"}).\n\t\t\tAddRow(1, "Test Product", "test-product", "A test product", nil, "Icons", "10.00", "", "", "active", 100, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)).\n\t\t\tAddRow(2, "Another Product", "another-product", "Another test", nil, "", "20.00", "", "", "active", 50, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))')

# Images query - handler fetches for each product (2 products -> 2 queries with different product_ids)
fix("GetProducts images 1",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "url", "is_primary", "created_at"}).\n\t\t\tAddRow(1, 1, "http://example.com/img.png", true, mockTime(2024, time.January, 1)))',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "url", "is_primary", "created_at"}).\n\t\t\tAddRow(1, 1, "http://example.com/img.png", true, mockTime(2024, time.January, 1)))')

# Need a second images query for product 2
fix("GetProducts images 2",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(2).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "url", "is_primary", "created_at"}).\n\t\t\tAddRow(2, 2, "http://example.com/img2.png", true, mockTime(2024, time.January, 1)))',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(2).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "url", "is_primary", "created_at"}).\n\t\t\tAddRow(2, 2, "http://example.com/img2.png", true, mockTime(2024, time.January, 1)))')

# ============================================================
# GetProductBySlug
# ============================================================

fix("GetProductBySlug COUNT",
    'mock.ExpectQuery(`SELECT COUNT\(\*\\) FROM products WHERE slug = \$1`).\n\t\tWithArgs("test-product").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n\t\tWithArgs("test-product").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))')

fix("GetProductBySlug LIST",
    'mock.ExpectQuery(`SELECT p.id, p.title.*`).\n\t\tWithArgs("test-product").\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "description", "category_id", "category_name", "price_usd", "asset_path", "asset_hash", "status", "download_count_limit", "max_downloads_per_user", "file_size_bytes", "file_mime_type", "created_at", "updated_at"}).\n\t\t\tAddRow(1, "Test Product", "test-product", "A test product", nil, "Icons", "10.00", "", "", "active", 100, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\') as category_name, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.slug = $1 AND p.status = \'active\'`).\n\t\tWithArgs("test-product").\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "description", "category_id", "category_name", "price_usd", "asset_path", "asset_hash", "status", "download_count_limit", "max_downloads_per_user", "file_size_bytes", "file_mime_type", "created_at", "updated_at"}).\n\t\t\tAddRow(1, "Test Product", "test-product", "A test product", nil, "Icons", "10.00", "", "", "active", 100, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))')

fix("GetProductBySlug images",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "url", "is_primary", "created_at"}).\n\t\t\tAddRow(1, 1, "http://example.com/img.png", true, mockTime(2024, time.January, 1)))',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "url", "is_primary", "created_at"}).\n\t\t\tAddRow(1, 1, "http://example.com/img.png", true, mockTime(2024, time.January, 1)))')

# ============================================================
# SearchProducts
# ============================================================

fix("SearchProducts COUNT",
    'mock.ExpectQuery(`SELECT COUNT\(\*\\) FROM products.*`).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND (p.name LIKE $2 OR p.description LIKE $2 OR p.slug LIKE $2)`).\n\t\tWithArgs("active", "%test%").\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))')

fix("SearchProducts LIST",
    'mock.ExpectQuery(`SELECT p.id, p.title.*`).\n\t\tWithArgs("%test%").\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "description", "category_id", "category_name", "price_usd", "asset_path", "asset_hash", "status", "download_count_limit", "max_downloads_per_user", "file_size_bytes", "file_mime_type", "created_at", "updated_at"}).\n\t\t\tAddRow(1, "Test Product", "test-product", "A test product", nil, "Icons", "10.00", "", "", "active", 100, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))',
    'mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \'\') as category_name, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND (p.name LIKE $2 OR p.description LIKE $2 OR p.slug LIKE $2) ORDER BY p.created_at DESC LIMIT $3 OFFSET $4`).\n\t\tWithArgs("active", "%test%", 12, 0).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "description", "category_id", "category_name", "price_usd", "asset_path", "asset_hash", "status", "download_count_limit", "max_downloads_per_user", "file_size_bytes", "file_mime_type", "created_at", "updated_at"}).\n\t\t\tAddRow(1, "Test Product", "test-product", "A test product", nil, "Icons", "10.00", "", "", "active", 100, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))')

# ============================================================
# Write and test
# ============================================================
with open(SRC, 'w') as f:
    f.write(text)

print(f"\nApplied {changes} fixes")

r = subprocess.run(['go', 'build', './handlers/'], capture_output=True, text=True, cwd='/root/project/backend')
print(f"Build: {'OK' if r.returncode == 0 else 'FAIL'}")
if r.returncode:
    print(r.stderr[:300])
    exit(1)

r2 = subprocess.run(['go', 'test', './handlers/', '-run', 'TestProducts_GetProducts$', '-count=1'], capture_output=True, text=True, cwd='/root/project/backend', timeout=30)
print(f"\nGetProducts test: {r2.stdout.strip()}")
if r2.returncode:
    print(r2.stderr[:500])
