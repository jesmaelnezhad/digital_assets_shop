#!/usr/bin/env python3
"""Fix test mocks to match exact handler SQL."""

with open("/root/project/backend/handlers/all_features_test.go", "r") as f:
    lines = f.readlines()

# Fix line 125 (0-indexed: 124) - the product list mock
# Handler sends multi-line SQL with tabs and newlines.
# Go regex \s+ matches whitespace including newlines and tabs.
# The pattern must match the ENTIRE query string.
# 
# Handler SQL (what fmt.Sprintf produces):
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
# As regex in Go backtick: use \s+ for whitespace, \$ for dollar, \ for escaping dots

pattern = (
    r'(?s)SELECT\s+p\.id,\s+p\.title,\s+p\.slug,\s+p\.description,\s+'
    r'p\.category_id,\s+COALESCE\(c\.name,\s+\'\',\s+\)\s+'
    r'p\.price_usd,\s+p\.asset_path,\s+p\.asset_hash,\s+p\.status,\s+'
    r'p\.download_count_limit,\s+p\.max_downloads_per_user,\s+'
    r'p\.file_size_bytes,\s+p\.file_mime_type,\s+p\.created_at,\s+'
    r'p\.updated_at\s+FROM\s+products\s+p\s+LEFT\s+JOIN\s+categories\s+c\s+'
    r'ON\s+p\.category_id\s*=\s*c\.id\s+WHERE\s+p\.status\s*=\s*\$1\s+'
    r'ORDER\s+BY\s+p\.created_at\s+DESC\s+LIMIT\s+\$2\s+OFFSET\s+\$3'
)
lines[124] = '\tmock.ExpectQuery(`' + pattern + '`).\n'

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.writelines(lines)

print("Fixed line 125")

# Verify byte content
with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    data = f.read()
idx = data.find(b'ExpectQuery')
if idx >= 0:
    end = data.find(b'`).', idx)
    line = data[idx:end+3].decode('utf-8', errors='replace')
    # Check for doubled backslashes
    if '\\\\s' in line:
        print(f"WARNING: double-escaped \\\\s found in: {line[:100]}")
    elif '\\s' in line:
        print(f"OK: single \\s in pattern")
    print(f"Pattern: {line[:150]}...")

import subprocess
r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True)
print(f"Build: {r.returncode}")
if r.returncode != 0:
    print(r.stderr[-300:])
    
r = subprocess.run(["go", "test", "./handlers/", "-v", "-count=1", "-run", "TestProducts_GetProducts$"], 
                   capture_output=True, text=True)
print(r.stdout[-400:])
