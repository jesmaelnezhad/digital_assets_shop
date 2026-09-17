#!/usr/bin/env python3
"""
Fix all_features_test.go mocks to match actual handler SQL queries.
Reads handler source code to extract exact SQL, then rewrites mocks.
"""
import re, os

TEST_FILE = "/root/project/backend/handlers/all_features_test.go"
BACKEND_DIR = "/root/project/backend/handlers"

def read_file(path):
    with open(path) as f:
        return f.read()

def extract_sql_from_handler(handler_file, function_pattern):
    """Extract SQL strings from a handler function."""
    content = read_file(handler_file)
    results = []
    
    # Find the function
    func_match = re.search(function_pattern, content, re.DOTALL)
    if not func_match:
        return results
    
    func_body = func_match.group(0)
    
    # Find all SQL strings (backtick or double-quoted)
    # Backtick strings
    for m in re.finditer(r'`([^`]+)`', func_body):
        sql = m.group(1).strip()
        if sql and ('SELECT' in sql or 'INSERT' in sql or 'UPDATE' in sql or 'DELETE' in sql):
            results.append(('backtick', sql))
    
    # Double-quoted strings that look like SQL
    for m in re.finditer(r'"([^"]*(?:SELECT|INSERT|UPDATE|DELETE)[^"]*)"', func_body):
        sql = m.group(1).strip()
        if sql:
            results.append(('double', sql))
    
    return results

# Extract actual SQL from handlers
print("Extracting SQL from handlers...")

# Auth handler
auth_sql = extract_sql_from_handler(f"{BACKEND_DIR}/auth.go", r'func \(h \*AuthHandler\) Register.*?(?=\nfunc |\n// \n|\Z)')
print(f"Auth.Register SQL: {auth_sql}")

# Product handlers
prod_sql = extract_sql_from_handler(f"{BACKEND_DIR}/products.go", r'func GetProducts.*?(?=\nfunc |\n// \n|\Z)')
print(f"Products.GetProducts SQL count: {len(prod_sql)}")

# This approach is too fragile. Let me just fix the known mismatches directly
# by reading the test file and patching specific patterns.

print("\nDirectly fixing mock SQL patterns...")

content = read_file(TEST_FILE)

# Fix 1: Register mock - use exact SQL with $N placeholders
old = "mock.ExpectQuery(`INSERT INTO users(?s).*`)."
new = "mock.ExpectQuery(`INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3) RETURNING id`)."
content = content.replace(old, new)
print("  Fixed Register mock SQL")

# Fix 2: UPDATE users SET name mock
old = 'mock.ExpectExec(`UPDATE users SET name = \\$1, updated_at = CURRENT_TIMESTAMP WHERE id = \\$2`).'
new = 'mock.ExpectExec(`UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`).'
content = content.replace(old, new)
print("  Fixed UPDATE users SET name mock")

# Fix 3: INSERT INTO user_profiles mock
old = 'mock.ExpectExec(`INSERT INTO user_profiles(?s).*`).'
new = 'mock.ExpectExec(`INSERT INTO user_profiles (user_id, bio, wallet_address) VALUES ($1, $2, $3) ON CONFLICT (user_id) DO NOTHING`).'
content = content.replace(old, new)
print("  Fixed user_profiles INSERT mock")

# Fix 4: DELETE FROM invalidated_tokens mock
old = 'mock.ExpectExec(`DELETE FROM invalidated_tokens(?s).*`).'
new = 'mock.ExpectExec(`DELETE FROM invalidated_tokens`).'
content = content.replace(old, new)
print("  Fixed invalidated_tokens DELETE mock")

# Fix 5: INSERT INTO invalidated_tokens mock
old = 'mock.ExpectExec(`INSERT INTO invalidated_tokens(?s).*`).'
new = 'mock.ExpectExec(`INSERT INTO invalidated_tokens (token, created_at) VALUES ($1, NOW())`).'
content = content.replace(old, new)
print("  Fixed invalidated_tokens INSERT mock")

# Fix 6: Products search mock - fix the WHERE clause
old = "mock.ExpectQuery(`SELECT COUNT\\(\\*\\) FROM products WHERE(?s).*`)."
new = "mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE name LIKE $1 OR description LIKE $1 OR slug LIKE $1`)."
content = content.replace(old, new)
print("  Fixed products search COUNT mock")

# Fix 7: Products search SELECT mock - fix COALESCE
old = "mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE\\(c.name, ''\\) as category_name,.*`)."
new = "mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, '') as category_name, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`)."
content = content.replace(old, new)
print("  Fixed products search SELECT mock")

# Fix 8: Products list SELECT mock
old = "mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE\\(c.name, ''\\) as category_name,.*`).\n\t\tWillReturnRows(sqlmock.NewRows([]string{\"id\", \"title\", \"slug\", \"description\", \"category_id\", \"category_name\", \"price_usd\", \"asset_path\", \"asset_hash\", \"status\", \"download_count_limit\", \"max_downloads_per_user\", \"file_size_bytes\", \"file_mime_type\", \"created_at\", \"updated_at\"}).\n\t\t\tAddRow(1, \"Test Product\", \"test-product\", \"A test product\", nil, \"Icons\", \"10.00\", \"\", \"\", \"active\", 100, 5, nil, \"\", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)).\n\t\t\tAddRow(2, \"Another Product\", \"another-product\", \"Another test\", nil, \"\", \"20.00\", \"\", \"\", \"active\", 50, 5, nil, \"\", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))"
new = "mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, '') as category_name, p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND p.stock_quantity > 0 ORDER BY p.created_at DESC LIMIT 12 OFFSET 0`)."
content = content.replace(old, new)
print("  Fixed products list SELECT mock")

with open(TEST_FILE, "w") as f:
    f.write(content)

print("\nDone. Now need to also fix the product_images mock and other remaining mocks.")
