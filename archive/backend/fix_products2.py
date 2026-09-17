#!/usr/bin/env python3
"""Apply all mock fixes to all_features_test.go based on actual file content."""
import re, subprocess

SRC = "/root/project/backend/handlers/all_features_test.go"
with open(SRC) as f:
    text = f.read()

changes = 0
def fix(name, old, new):
    global text, changes
    if old not in text:
        print(f"  MISS: {name}")
        return False
    text = text.replace(old, new, 1)
    changes += 1
    print(f"  OK: {name}")
    return True

# Read current state of key sections
for i in range(120, 145):
    line = text.split('\n')[i]
    if 'mock.Expect' in line or 'SELECT COUNT' in line or 'product_images' in line:
        print(f"L{i+1}: {line[:120]}")

print("\n--- Applying fixes ---")

# Fix 1: GetProducts COUNT (line 123) - has COUNT\(\\*\\) but needs exact SQL + WithArgs
fix("GetProducts COUNT",
    'mock.ExpectQuery(`SELECT COUNT\\(\\*\\) FROM products.*`).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n\t\tWithArgs("active").\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))')

# Fix 2: GetProducts images - need second query for product 2
# Check if we already have it
if 'WithArgs(2)' in text and 'product_images' in text:
    print("  OK: GetProducts images 2 (already present)")
else:
    # Add second images query after the first one's WillReturnRows
    old_block = "mockTime(2024, time.January, 1)))."
    if old_block in text:
        insert_pos = text.find(old_block) + len(old_block)
        second = '\n\tmock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(2).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "url", "is_primary", "created_at"}).\n\t\t\tAddRow(2, 2, "http://example.com/img2.png", true, mockTime(2024, time.January, 1)))'
        text = text[:insert_pos] + second + text[insert_pos:]
        print("  OK: GetProducts images 2 (added)")
        changes += 1
    else:
        print("  MISS: GetProducts images 2 (anchor not found)")

# Fix 3: GetProductBySlug COUNT
fix("GetProductBySlug COUNT",
    'mock.ExpectQuery(`SELECT COUNT\\(\\*\\) FROM products WHERE slug = \\$1`).\n\t\tWithArgs("test-product").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n\t\tWithArgs("test-product").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))')

# Fix 4: GetProductBySlug images
fix("GetProductBySlug images",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# Fix 5: SearchProducts COUNT
fix("SearchProducts COUNT",
    'mock.ExpectQuery(`SELECT COUNT\\(\\*\\) FROM products.*`).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND (p.name LIKE $2 OR p.description LIKE $2 OR p.slug LIKE $2)`).\n\t\tWithArgs("active", "%test%").\n\t\tWillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))')

# Fix 6: SearchProducts images
fix("SearchProducts images",
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# Write
with open(SRC, 'w') as f:
    f.write(text)

print(f"\nApplied {changes} fixes")

# Build + test
r = subprocess.run(['go', 'build', './handlers/'], capture_output=True, text=True, cwd='/root/project/backend')
print(f"Build: {'OK' if r.returncode == 0 else 'FAIL'}")
if r.returncode:
    print(r.stderr[:300])
    exit(1)

r2 = subprocess.run(['go', 'test', './handlers/', '-run', 'TestProducts', '-count=1'], capture_output=True, text=True, cwd='/root/project/backend', timeout=30)
for l in r2.stdout.split('\n'):
    if 'PASS' in l or 'FAIL' in l or 'Error' in l:
        print(f"  {l.strip()}")
