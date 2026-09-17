#!/usr/bin/env python3
"""Fix broken mock chains in all_features_test.go"""
import re

test_file = "/root/project/backend/handlers/all_features_test.go"
with open(test_file) as f:
    content = f.read()

fixes = 0

# Fix 1: Products_GetProducts line 129 - add WithArgs/WillReturnRows chain
old = "mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = \\$1 ORDER BY p.created_at DESC LIMIT \\$2 OFFSET \\$3`).\nmock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \\$1`)."
new = "mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = \\$1 ORDER BY p.created_at DESC LIMIT \\$2 OFFSET \\$3`).\n\t\tWithArgs(\"active\", 12, 0).\n\t\tWillReturnRows(sqlmock.NewRows([]string{\"id\", \"title\", \"slug\", \"description\", \"category_id\", \"category_name\", \"price_usd\", \"asset_path\", \"asset_hash\", \"status\", \"download_count_limit\", \"max_downloads_per_user\", \"file_size_bytes\", \"file_mime_type\", \"created_at\", \"updated_at\"}).\n\t\t\tAddRow(1, \"Test Product\", \"test-product\", \"A test product\", nil, \"Icons\", \"10.00\", \"\", \"\", \"active\", 100, 5, nil, \"\", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)).\n\t\t\tAddRow(2, \"Another Product\", \"another-product\", \"Another test\", nil, \"\", \"20.00\", \"\", \"\", \"active\", 50, 5, nil, \"\", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))\nmock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \\$1`)."
if old in content:
    content = content.replace(old, new)
    fixes += 1
    print("FIXED: Products_GetProducts chain")
else:
    print("MISS: Products_GetProducts")

with open(test_file, "w") as f:
    f.write(content)

print(f"Fixes: {fixes}")
