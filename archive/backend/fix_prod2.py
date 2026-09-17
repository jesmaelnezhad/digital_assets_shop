#!/usr/bin/env python3
with open("/root/project/backend/handlers/all_features_test.go") as f:
    lines = f.readlines()

# Fix lines for TestProducts_GetProducts (0-indexed)
# Line 123 (idx 122): COUNT - handler sends JOIN query
lines[122] = '\tmock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n'

# Line 125 (idx 124): product list - handler sends full JOIN query
sq = (
    '(?s)SELECT p.id, p.title, p.slug, p.description, '
    'p.category_id, COALESCE(c.name, ' + chr(39) + chr(39) + '), p.price_usd, '
    'p.asset_path, p.asset_hash, p.status, p.download_count_limit, '
    'p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, '
    'p.created_at, p.updated_at FROM products p LEFT JOIN categories c '
    'ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC '
    'LIMIT $2 OFFSET $3'
)
lines[124] = '\tmock.ExpectQuery(`' + sq + '`).\n'

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.writelines(lines)
print("Fixed products mocks")

import subprocess
for cmd in [["go", "build", "./handlers/"], 
            ["go", "test", "./handlers/", "-v", "-count=1", "-run", "TestProducts_GetProducts$"]]:
    r = subprocess.run(cmd, capture_output=True, text=True)
    print(f"$ {' '.join(cmd)}")
    print(r.stdout[-400:] if cmd[0] == 'go' and 'test' in cmd[1] else f"exit={r.returncode}")
    if r.returncode != 0 and r.stderr:
        print(r.stderr[-300:])
