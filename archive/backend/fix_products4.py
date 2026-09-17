#!/usr/bin/env python3
"""Write clean products test mock section."""
with open("/root/project/backend/handlers/all_features_test.go", "r") as f:
    content = f.read()

# The products list mock on line 125 needs proper regex escaping
# Current (broken): \\s for whitespace, \\\\$ for dollar
# Fixed: \s for whitespace, \$ for dollar
# In Go backtick raw string: \s = backslash-s (regex whitespace), \$ = backslash-dollar (regex dollar)

old = lines[124] if (lines := content.split('\n')) else None

# Build the correct regex pattern for the handler's product list SQL
# Handler sends (with fmt.Sprintf formatting, tabs/newlines):
# SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd,
#        p.asset_path, p.asset_hash, p.status, p.download_count_limit,
#        p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type,
#        p.created_at, p.updated_at
# FROM products p
# LEFT JOIN categories c ON p.category_id = c.id
# WHERE p.status = $1
# ORDER BY p.created_at DESC
# LIMIT $2 OFFSET $3
#
# As a regex in Go backtick:
# - Use \s+ for whitespace (matches any whitespace including newlines)
# - Use \. to escape dots in column names
# - Use \$ to escape dollar signs (regex metachar)
# - Use \( \) to escape parens in COALESCE
# - Use \' to match single quotes in COALESCE(c.name, '')

# In Python raw string r'...': \s = backslash-s, \$ = backslash-dollar, etc.
# When written to Go file, backtick string gets these literal chars
# Go regexp interprets \s as whitespace, \$ as literal dollar

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

lines = content.split('\n')
lines[124] = '\tmock.ExpectQuery(`' + sq + '`).\n'

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write('\n'.join(lines))

# Verify what was written
with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    data = f.read()

idx = data.find(b'ExpectQuery.*SELECT')
if idx >= 0:
    end = data.find(b'`).', idx)
    snippet = data[idx:end+3]
    print(f"Written regex pattern ({len(snippet)} bytes):")
    print(snippet.decode('utf-8', errors='replace')[:200])

import subprocess
r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True)
print(f"\nBuild: {r.returncode}")
r = subprocess.run(["go", "test", "./handlers/", "-v", "-count=1", "-run", "TestProducts_GetProducts$"], capture_output=True, text=True)
print(r.stdout[-400:])
