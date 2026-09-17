#!/usr/bin/env python3
"""Fix products test mock - correct regex escaping and full query match."""
import re, subprocess

with open("/root/project/backend/handlers/all_features_test.go") as f:
    content = f.read()

lines = content.split('\n')

# Fix line 123 (idx 122): COUNT - handler sends JOIN query
lines[122] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).'

# Fix line 125 (idx 124): Product list - handler sends multi-line query
# The regex needs to match: SELECT ... FROM products p LEFT JOIN categories c ON ... WHERE ... ORDER BY ... LIMIT ... OFFSET
# Use \s for whitespace matching in Go regex (in backtick, \s is literally backslash-s, which Go regex interprets as whitespace)
sq = (
    '(?s)SELECT\s+p\.id,\s+p\.title,\s+p\.slug,\s+p\.description,\s+'
    'p\.category_id,\s+COALESCE\(c\.name,\s+\'\',\s+\)\s+'
    'p\.price_usd,\s+p\.asset_path,\s+p\.asset_hash,\s+p\.status,\s+'
    'p\.download_count_limit,\s+p\.max_downloads_per_user,\s+'
    'p\.file_size_bytes,\s+p\.file_mime_type,\s+p\.created_at,\s+'
    'p\.updated_at\s+FROM\s+products\s+p\s+LEFT\s+JOIN\s+categories\s+c\s+'
    'ON\s+p\.category_id\s*=\s*c\.id\s+WHERE\s+p\.status\s*=\s*\$1\s+'
    'ORDER\s+BY\s+p\.created_at\s+DESC\s+LIMIT\s+\$2\s+OFFSET\s+\$3'
)
lines[124] = '\tmock.ExpectQuery(`' + sq + '`).\n'

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write('\n'.join(lines))

# Verify the bytes are correct
with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    data = f.read()
idx = data.find(b'ExpectQuery')
if idx >= 0:
    end = data.find(b'`).', idx + 500)
    snippet = data[idx:end+3]
    # Check for double-escaping: \\\\s means file has \\s, which is wrong
    # We need \s (one backslash + s) in file for Go regex to see \s
    if b'\\\\s' in snippet:
        print("WARNING: double-escaped \\\\s found in products mock!")
        # Fix it
        data = data.replace(b'\\\\s', b'\\s')
        data = data.replace(b'\\\\$', b'\\$')
        data = data.replace(b'\\\\(', b'\\(')
        data = data.replace(b'\\\\*', b'\\*')
        with open("/root/project/backend/handlers/all_features_test.go", "wb") as f:
            f.write(data)
        print("Fixed double-escaping")

print("Products mock fixed")

import subprocess
r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True)
print("Build: exit=%d" % r.returncode)

r = subprocess.run(["go", "test", "./handlers/", "-v", "-count=1", "-run", "TestProducts_GetProducts$"],
                   capture_output=True, text=True)
print(r.stdout[-500:])
