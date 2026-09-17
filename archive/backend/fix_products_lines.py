#!/usr/bin/env python3
"""Fix products test mocks to match handler SQL."""
with open("/root/project/backend/handlers/all_features_test.go") as f:
    lines = f.readlines()

# Line 123 (0-indexed: 122): COUNT query with JOIN
lines[122] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n'

# Line 125 (0-indexed: 124): Product list with JOIN, COALESCE, multiline
# In Go backtick strings, single quotes are literal - no escaping needed
# The regex (?!s) flag means single-line mode (dot matches newline)
lines[124] = (
    '\tmock.ExpectQuery(`(?s)SELECT p.id, p.title, p.slug, p.description, '
    'p.category_id, COALESCE(c.name, \'\'), p.price_usd, p.asset_path, '
    'p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, '
    'p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p '
    'LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 '
    'ORDER BY p.created_at DESC LIMIT $2 OFFSET $3`).\n'
)

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.writelines(lines)
print("Fixed products lines 123, 125")
