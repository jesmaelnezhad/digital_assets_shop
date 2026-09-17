#!/usr/bin/env python3
"""Fix all failing mocks in all_features_test.go based on actual handler SQL."""
import re

SRC = "/root/project/backend/handlers/all_features_test.go"
with open(SRC, "r") as f:
    content = f.read()

# Fix 1: SearchProducts COUNT - needs to match search query with status + search args
# Handler builds: SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND p.search_vector @@ plainto_tsquery('english', $2)
content = content.replace(
    "mock.ExpectQuery(`SELECT COUNT\\(\\*\\) FROM products.*`).\n\t\tWillReturnRows(sqlmock.NewRows([]string{\"count\"}).AddRow(1))",
    "mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND p.search_vector @@ plainto_tsquery('english', $2)`).\n\t\tWithArgs(\"active\", \"%test%\").\n\t\tWillReturnRows(sqlmock.NewRows([]string{\"count\"}).AddRow(1))"
)

# Fix 2: SearchProducts product query needs 4 args (status, search, perPage, offset) not just search
# Handler: WHERE p.status = $1 AND p.search_vector @@ plainto_tsquery('english', $2) ORDER BY p.created_at DESC LIMIT $3 OFFSET $4
content = content.replace(
    "`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`.\n\t\tWithArgs(\"%test%\)).",
    "`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND p.search_vector @@ plainto_tsquery('english', $2) ORDER BY p.created_at DESC LIMIT $3 OFFSET $4`.\n\t\tWithArgs(\"active\", \"%test%\", 12, 0)."
)

# Fix 3: GetProductBySlug - the product query condition is "WHERE p.slug = $1 AND p.status = 'active'"
# But the mock has ORDER BY/LIMIT which doesn't exist for single product query
# Handler: SELECT ... FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.slug = $1 AND p.status = 'active'
content = content.replace(
    "`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`.\n\t\tWithArgs(\"test-product\").\n\t\tWillReturnRows(sqlmock.NewRows([]string{\"id\", \"title\", \"slug\", \"description\", \"category_id\", \"category_name\", \"price_usd\", \"asset_path\", \"asset_hash\", \"status\", \"download_count_limit\", \"max_downloads_per_user\", \"file_size_bytes\", \"file_mime_type\", \"created_at\", \"updated_at\"}).\n\t\t\tAddRow(1, \"Test Product\", \"test-product\", \"A test product\", nil, \"Icons\", \"10.00\", \"\", \"\", \"active\", 100, 5, nil, \"\", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))",
    "`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.slug = $1 AND p.status = 'active'`.\n\t\tWithArgs(\"test-product\").\n\t\tWillReturnRows(sqlmock.NewRows([]string{\"id\", \"title\", \"slug\", \"description\", \"category_id\", \"category_name\", \"price_usd\", \"asset_path\", \"asset_hash\", \"status\", \"download_count_limit\", \"max_downloads_per_user\", \"file_size_bytes\", \"file_mime_type\", \"created_at\", \"updated_at\"}).\n\t\t\tAddRow(1, \"Test Product\", \"test-product\", \"A test product\", nil, \"Icons\", \"10.00\", \"\", \"\", \"active\", 100, 5, nil, \"\", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))"
)

# Fix 4: GetProducts main query - same fix, it's also wrong (has ORDER BY/LIMIT but handler's GetProducts uses them)
# Already fixed by replace_all above. But we need to make sure the GetProducts test specifically has proper WithArgs
# The GetProducts handler: WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3
# WithArgs should be: "active", perPage (12), offset (0)
# Currently: no WithArgs at all for the products query in GetProducts test
# Actually looking at line 126-128: there's no WithArgs for the SELECT query in GetProducts test!
# It just has WillReturnRows directly. Need to add WithArgs("active", 12, 0)

# Fix GetProducts test - add WithArgs
content = content.replace(
    "`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`.\n\t\tWillReturnRows(sqlmock.NewRows([]string{\"id\", \"title\", \"slug\", \"description\", \"category_id\", \"category_name\", \"price_usd\", \"asset_path\", \"asset_hash\", \"status\", \"download_count_limit\", \"max_downloads_per_user\", \"file_size_bytes\", \"file_mime_type\", \"created_at\", \"updated_at\"}).\n\t\t\tAddRow(1, \"Test Product\", \"test-product\", \"A test product\", nil, \"Icons\", \"10.00\", \"\", \"\", \"active\", 100, 5, nil, \"\", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)).\n\t\t\tAddRow(2, \"Another Product\", \"another-product\", \"Another test\", nil, \"\", \"20.00\", \"\", \"\", \"active\", 50, 5, nil, \"\", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))",
    "`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`.\n\t\tWithArgs(\"active\", 12, 0).\n\t\tWillReturnRows(sqlmock.NewRows([]string{\"id\", \"title\", \"slug\", \"description\", \"category_id\", \"category_name\", \"price_usd\", \"asset_path\", \"asset_hash\", \"status\", \"download_count_limit\", \"max_downloads_per_user\", \"file_size_bytes\", \"file_mime_type\", \"created_at\", \"updated_at\"}).\n\t\t\tAddRow(1, \"Test Product\", \"test-product\", \"A test product\", nil, \"Icons\", \"10.00\", \"\", \"\", \"active\", 100, 5, nil, \"\", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)).\n\t\t\tAddRow(2, \"Another Product\", \"another-product\", \"Another test\", nil, \"\", \"20.00\", \"\", \"\", \"active\", 50, 5, nil, \"\", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))"
)

with open(SRC, "w") as f:
    f.write(content)
print("Fixes applied")
