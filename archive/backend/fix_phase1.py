#!/usr/bin/env python3
"""
Comprehensive fix for all_features_test.go.
1. Fix Go compile errors (escape sequences in double-quoted strings)
2. Fix all mock SQL to match handler SQL exactly
3. Add missing WithArgs
4. Fix column lists
"""
import re, os

PATH = "/root/project/backend/handlers/all_features_test.go"
with open(PATH) as f:
    content = f.read()

changes = []

def replace(old, new, count=1):
    """Replace text, tracking if successful."""
    global content
    n = content.count(old)
    if n == 0:
        changes.append(f"MISS: {old[:80]}")
        return False
    content = content.replace(old, new, count)
    changes.append(f"OK: {old[:80]}")
    return True

def replace_all(old, new):
    """Replace all occurrences."""
    global content
    n = content.count(old)
    if n == 0:
        changes.append(f"MISS (all): {old[:80]}")
        return False
    content = content.replace(old, new)
    changes.append(f"OK ({n}x): {old[:80]}")
    return True

# ============================================================
# PHASE 1: Fix Go compile errors - escape sequences
# ============================================================
# Strategy: convert ALL mock SQL strings from double-quoted to backtick raw strings
# This eliminates escape sequence issues entirely

# Pattern: mock.ExpectQuery("...") or mock.ExpectExec("...")
# We need to find the opening quote after ExpectQuery/ExpectExec and 
# convert the whole string to backticks

# Find all mock expectations and convert them
def convert_mock_strings(text):
    """Convert double-quoted mock SQL to backtick raw strings."""
    result = []
    i = 0
    while i < len(text):
        # Look for mock.ExpectQuery( or mock.ExpectExec(
        idx = text.find('mock.ExpectQuery("', i)
        if idx < 0:
            idx = text.find('mock.ExpectExec("', i)
        if idx < 0:
            result.append(text[i:])
            break
        
        result.append(text[i:idx])
        
        # Find the opening quote position
        func_end = text.find('("', idx)
        if func_end < 0:
            result.append(text[idx:])
            break
        
        quote_start = func_end + 2  # position of the opening "
        
        # Find the closing " - scan through the string
        j = quote_start
        while j < len(text):
            if text[j] == '"' and (j == 0 or text[j-1] != '\\'):
                break
            j += 1
        
        if j >= len(text):
            result.append(text[idx:])
            i = len(text)
            break
        
        # Extract the SQL string content (between quotes)
        sql_content = text[quote_start:j]
        
        # Convert to backtick raw string
        # In raw strings, \$ is literal \$ (regex matches literal $)
        # The double-quoted version had various escaping
        # For raw strings, we typically want \$ for matching $1, $2 etc.
        
        # Check what escaping we have
        if '\\\\$' in sql_content:
            # Double-escaped: \\\$ in double-quoted string = \$ in Go value
            # In raw string: \$ matches literal $ → correct!
            raw_sql = sql_content.replace('\\\\', '\\')
        elif '\\$' in sql_content:
            # Already single-escaped: \$ in double-quoted = literal $ in value
            # But \$ in double-quoted Go string is INVALID in Go 1.22
            # In raw string: \$ is fine
            raw_sql = sql_content.replace('\\', '')
        elif '\\*' in sql_content:
            raw_sql = sql_content.replace('\\*', '*')
        else:
            raw_sql = sql_content
        
        replacement = f'mock.ExpectQuery(`{raw_sql}`)'
        # Determine if it was ExpectExec
        if 'mock.ExpectExec("' in text[idx:idx+20]:
            replacement = f'mock.ExpectExec(`{raw_sql}`)'
        
        result.append(replacement)
        i = j + 1
    
    return ''.join(result)

content = convert_mock_strings(content)
print("Phase 1: Converted double-quoted mock strings to backtick raw strings")

# ============================================================
# PHASE 2: Fix specific SQL patterns to match handler SQL
# ============================================================

# --- Products: exact SQL from products.go ---
# Handler GetProducts sends:
#   SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1
#   SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name,''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3
#   SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id

replace_all('SELECT COUNT\\(\\*\\\) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = \$1',
             'SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1')

replace_all('SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \\'\\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = \$1 ORDER BY p.created_at DESC LIMIT \$2 OFFSET \$3',
             'SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \\'\\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 ORDER BY p.created_at DESC LIMIT $2 OFFSET $3')

replace_all('SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \$1',
            'SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id')

# --- GetProductBySlug ---
replace_all('SELECT COUNT\\(\\*\\\) FROM products WHERE slug = \$1',
            'SELECT COUNT(*) FROM products WHERE slug = $1')

replace_all('SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \\'\\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.slug = \$1 AND p.status = \\'active\\'',
            'SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, \\'\\'), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.slug = $1 AND p.status = \\'active\\'')

# --- SearchProducts ---
replace_all("SELECT COUNT\\(\\*\\\) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = \$1 AND (p.name LIKE \$2 OR p.description LIKE \$2 OR p.slug LIKE \$2)",
            "SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND (p.name LIKE $2 OR p.description LIKE $2 OR p.slug LIKE $2)")

replace_all("SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = \$1 AND (p.name LIKE \$2 OR p.description LIKE \$2 OR p.slug LIKE \$2) ORDER BY p.created_at DESC LIMIT \$3 OFFSET \$4",
            "SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND (p.name LIKE $2 OR p.description LIKE $2 OR p.slug LIKE $2) ORDER BY p.created_at DESC LIMIT $3 OFFSET $4")

# ============================================================
# PHASE 3: Write and test
# ============================================================
with open(PATH, "w") as f:
    f.write(content)

print("\nAll changes:")
for c in changes:
    print(f"  {c}")

print(f"\nFile written. {len(changes)} changes tracked.")
