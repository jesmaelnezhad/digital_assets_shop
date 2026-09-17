#!/usr/bin/env python3
"""Fix products test mocks for TestProducts_GetProducts."""
with open("/root/project/backend/handlers/all_features_test.go", "r") as f:
    content = f.read()

lines = content.split('\n')

# Line 123 (idx 122): COUNT - handler sends JOIN query, use simple pattern
lines[122] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n'

# Line 125 (idx 124): Product list - use \s for whitespace in regex, \ for literal chars
# In Python, '\s' = backslash + s (2 chars). Written to Go backtick file, Go regex sees \s = whitespace.
sq = (
    r'(?s)SELECT\s+p\.id,\s+p\.title,\s+p\.slug,\s+p\.description,\s+'
    r'p\.category_id,\s+COALESCE\(c\.name,\s+\'\',\s+\)\s+'
    r'p\.price_usd,\s+p\.asset_path,\s+p\.asset_hash,\s+p\.status,\s+'
    r'p\.download_count_limit,\s+p\.max_downloads_per_user,\s+'
    r'p\.file_size_bytes,\s+p\.file_mime_type,\s+p\.created_at,\s+'
    r'p\.updated_at\s+FROM\s+products\s+p\s+LEFT\s+JOIN\s+categories\s+c\s+'
    r'ON\s+p\.category_id\s*=\s*c\.id\s+WHERE\s+p\.status\s*=\s*\$1\s+'
    r'ORDER\s+BY\s+p\.created_at\s+DESC\s+LIMIT\s+\$2\s+OFFSET\s+\$3'
)

lines[124] = '\tmock.ExpectQuery(`' + sq + '`).\n'

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write('\n'.join(lines))

import subprocess
r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True)
print(f"Build: {r.returncode}")
r = subprocess.run(["go", "test", "./handlers/", "-v", "-count=1", "-run", "TestProducts_GetProducts$"], capture_output=True, text=True)
print(r.stdout[-400:])
