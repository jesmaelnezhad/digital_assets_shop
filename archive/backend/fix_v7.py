#!/usr/bin/env python3
"""
Fix all_features_test.go comprehensively.
Reads the backup, converts to backtick strings, fixes SQL, adds WithArgs, fixes columns.
"""
import re, os, subprocess

BASE = "/root/project/backend"
TEST = f"{BASE}/handlers/all_features_test.go"

# First restore from backup
import shutil
shutil.copy(f"{BASE}/handlers/all_features_test.go.bak", TEST)

with open(TEST) as f:
    content = f.read()

changes = 0

def replace(old, new, count=1):
    global content, changes
    if old not in content:
        return False
    content = content.replace(old, new, count)
    changes += 1
    return True

# ============================================================
# STEP 1: Convert all double-quoted mock strings to backtick
# ============================================================
# Find: mock.ExpectQuery("SQL...") or mock.ExpectExec("SQL...")
# Convert to: mock.ExpectQuery(`SQL...`) or mock.ExpectExec(`SQL...`)
# But we need to handle the escaping properly.

# Strategy: process line by line
lines = content.split('\n')
new_lines = []
i = 0
n_converted = 0
while i < len(lines):
    line = lines[i]
    new_lines.append(line)
    i += 1

# That doesn't work well. Let me use regex instead.
# Match: mock.ExpectQuery("...") or mock.ExpectExec("...")
# The "..." may span multiple lines if it has embedded newlines (it doesn't in this file)

# Actually the file uses multi-line strings for some mocks. Let me handle that.
# Pattern: mock.Expect*("SQL...")  where SQL can be on the same line or next lines
# But actually in Go, strings can't span lines unless they have backticks or +\n+

# Let me look at the actual patterns in the file
# The backup uses single-line double-quoted strings for all mocks.
# So I can use a simple regex approach.

def convert_all_double_to_backtick(text):
    """Convert mock.Expect*("SQL") to mock.Expect*(`SQL`) handling escapes."""
    result = []
    i = 0
    count = 0
    while i < len(text):
        # Find next mock.ExpectQuery(" or mock.ExpectExec("
        mq = text.find('mock.ExpectQuery("', i)
        me = text.find('mock.ExpectExec("', i)
        if mq < 0 and me < 0:
            result.append(text[i:])
            break
        
        if mq >= 0 and (me < 0 or mq <= me):
            pos, sym = mq, 'ExpectQuery'
        else:
            pos, sym = me, 'ExpectExec'
        
        result.append(text[i:pos])
        
        # Find opening quote
        qstart = text.find('("', pos) + 2
        # Find closing quote - scan for next unescaped "
        qend = qstart
        while qend < len(text) and text[qend] != '"':
            qend += 1
        
        if qend >= len(text):
            result.append(text[pos:])
            break
        
        sql = text[qstart:qend]
        
        # Convert SQL to backtick raw string format
        # In double-quoted Go strings: \\\\ -> \ (in value), \\$ -> invalid
        # We want raw string to contain: \$ for matching $1, etc.
        # 
        # The sql variable is the raw on-disk text between the double quotes.
        # On disk: "SELECT ... \\\\\$1" means the bytes are: SELECT ... \\\\\$1
        # In Go double-quoted: \\\\ = \, so \\\\\$ = \$  → Go value: "\$1" (but \$ is invalid → error)
        # 
        # For raw string, we want the literal characters between backticks to be:
        # SELECT ... \$1  (so sqlmock sees \$1 → regex matches literal $)
        #
        # So from on-disk text "SELECT ... \\\\\$1", we want raw string `SELECT ... \$1`
        # Translation: \\\\$ → \$
        
        raw_sql = sql
        
        # Fix: \\\\ (4 backslashes on disk) → \ (1 backslash in raw)
        raw_sql = raw_sql.replace('\\\\\\\\', '\\')
        # Fix: \\$ (3 chars: \ \ $) → \$ (2 chars: \ $) — but after above, \\\\$ → \$
        # Actually after replacing \\\\\\ with \, what was \\\\\$ becomes \$
        # Let me handle it differently.
        
        # On disk, the text between double quotes for line 27 is:
        # SELECT id, email, ... WHERE email = \\\\\$1
        # That's: ... = \, \, \, \, $ 1
        # Hmm, let me count differently.
        # The cat -A output showed: \\\\\\$
        # In cat -A, each \ on screen = one \ on disk
        # So \\\\\\$ on screen = 5 chars on disk: \, \, \, \, $
        # Wait, \\ in cat -A = \, so \\\\ = \, \, \, \? No.
        # cat -A doesn't escape backslashes. It shows them literally.
        # So \\\\\\$ on screen = \, \, \, \, $ on disk = 5 chars
        # No wait. In terminal/cat, backslash IS displayed as backslash.
        # So \\\\\\$ = 6 chars: \, \, \, \, \, $
        # Hmm, but the Python repr showed: b'...\\\\\\$1...'
        # In Python repr, \\\\ represents 2 actual backslashes.
        # So \\\\\\$ in Python repr = 3 actual backslashes + dollar = 4 chars
        # But b'...' is bytes repr, where \\\\ = 2 bytes (two backslashes)
        # So \\\\\\$ in bytes repr = 3 bytes: \, \, $
        
        # OK I'm going in circles. Let me just check empirically.
        # After fix_v5 ran, the file line 27 was:
        # b'\tmock.ExpectQuery(`SELECT id, email, ... WHERE email = `$1``)).'
        # That has: ...= `$1`` -> which is wrong (extra backtick and $)
        # The correct form should be: ...= \$1`)
        
        # Let me just handle it pragmatically:
        # 1. Replace all occurrences of "\\\\" (4 chars on disk: \, \, \, \) with "\" (1 char)
        # 2. Replace all occurrences of "\\$" (2 chars on disk: \, $) with "$" (1 char)
        # Wait no. Let me look at what fix_v5 actually does and what the result is.
        
        # Actually, let me just write the correct raw string directly.
        # I know what the handler SQL looks like. Let me just fix each mock to have
        # the correct SQL in a backtick raw string.
        
        # For now, just convert the string and fix obvious issues
        raw_sql = sql
        
        # Remove one level of backslash escaping
        # \\\\ in source -> \ in raw string
        # \\$ in source -> $ in raw string (because in raw string, \$ is literal \$)
        # But wait, in raw string, \$ is literal backslash + dollar, which sqlmock regex reads as \$ = literal $
        # So we WANT \$ in raw string to match handler's $
        # The source has \\\\\$ which in Go double-quoted = \$ (if it compiled)
        # For raw string, we want the literal characters: \, $  → written as \$ in raw string source
        # So: replace \\\\$ with \$  (4 chars -> 2 chars)
        # And: replace \\$ with \$ (3 chars -> 2 chars)... but \\$ in source between double quotes
        # is: \, \, $ on disk. In Go double-quoted, \\$ = \$ (invalid). 
        # For raw string, we want \, $ on disk = \$ in raw string source.
        
        # Simplification: just replace all \\ with \ and all \$ with $
        # Wait, that's wrong for raw strings.
        # In raw string source: `\$` = literal \, $  → sqlmock sees \, $ → regex: literal $, then... 
        # No. sqlmock receives the string VALUE. For raw string `\$`, value = \, $ (2 chars).
        # sqlmock compiles as regex: \$ → matches literal $.
        # For raw string `\\$`, value = \, \, $ (3 chars). sqlmock: \\$ → matches literal \ then end-of-line.
        # 
        # Handler sends: $1 (no backslash).
        # sqlmock needs: \$1 → Go raw string: `\$1`
        # 
        # So we need the raw string to contain \$ (2 chars: \, $).
        # On disk, that's written as: \, $ between backticks.
        # 
        # The double-quoted source on disk has: \\\\\$ or \\\$ or \\$ depending on how many backslashes.
        # Let me just empirically figure it out by checking what the build produces.
        
        # For now, just use a simple approach: replace \\ with \ in the SQL, then put in backticks
        # But I need to be careful about the order of replacements.
        
        # Let me try: first collapse multiple backslashes
        raw_sql = raw_sql.replace('\\\\\\\\', '\\')  # \\\\ -> \
        raw_sql = raw_sql.replace('\\\\', '\\')      # \\ -> \ (remaining pairs)
        # Now handle \$: if we have \$ in the result, it means source had \\$ or \\\$ etc.
        # In raw string, \$ is fine (literal \, $).
        # But if we have \\$ in result (from \\\\$), that's \, \, $ which sqlmock reads as literal \ + end-of-line
        # We want \, $ which is \$.
        # After the replacements above, \\\\$ becomes \$, which is correct!
        # And \\$ becomes \$, also correct!
        
        # Actually wait. If source has \\$ (3 chars: \, \, $), after replace \\\\→\ and \\→\, we get:
        # \\$ → (no \\\\ match) → (replace \\ with \) → \$ (2 chars: \, $)
        # That's correct! \$ in raw string = \, $ on disk → sqlmock: \$ → matches literal $
        
        # And if source has \\\\$ (5 chars: \, \, \, \, $), after replace:
        # \\\\$ → \$ (2 chars: \, $) — correct!
        
        # But what about \\\\\$ (6 chars: \, \, \, \, \, $)?
        # \\\\\$ → \\\$ (after \\\\→\) → \$ (after \\→\) — correct!
        
        # So the simple approach works: replace \\\\ with \ then \\ with \
        
        replacement = f'mock.{sym}(`{raw_sql}`)'
        result.append(replacement)
        i = qend + 1
        n_converted += 1
    
    result.append(text[i:])
    return ''.join(result), n_converted

content, n = convert_all_double_to_backtick(content)
print(f"Step 1: Converted {n} double-quoted strings to backtick")

# ============================================================
# STEP 2: Fix escapes inside backtick raw strings
# ============================================================
# Inside backticks, \\$ should become \$ (remove one backslash level)
# Because the conversion above might have left \\ in some places
# We want: \$ in raw string (for matching $1), \* for matching *
# Currently: \\$ in raw string → sqlmock sees \\ → matches literal \
# We want: \$ in raw string → sqlmock sees \$ → matches literal $

result = []
i = 0
depth = 0  # inside backticks when odd
while i < len(content):
    if content[i] == '`':
        depth += 1
        result.append('`')
        i += 1
    elif depth % 2 == 1:
        # Inside raw string
        if i + 2 < len(content) and content[i] == '\\' and content[i+1] == '\\' and content[i+2] == '$':
            # \\$ → \$ (remove one backslash)
            result.append('\\')
            result.append('$')
            i += 3
        elif i + 1 < len(content) and content[i] == '\\' and content[i+1] == '*':
            # \* → * (remove backslash)
            result.append('*')
            i += 2
        elif i + 1 < len(content) and content[i] == '\\' and content[i+1] == '(':
            # \( is fine in raw string (regex for literal ()
            result.append(content[i:i+2])
            i += 2
        elif i + 1 < len(content) and content[i] == '\\' and content[i+1] == ')':
            # \) is fine
            result.append(content[i:i+2])
            i += 2
        else:
            result.append(content[i])
            i += 1
    else:
        result.append(content[i])
        i += 1

content = ''.join(result)
print("Step 2: Fixed escapes inside raw strings")

# ============================================================
# STEP 3: Verify build
# ============================================================
with open(TEST, 'w') as f:
    f.write(content)

r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True, cwd=BASE)
print(f"Build: exit={r.returncode}")
if r.returncode:
    print(r.stderr[:500])
    exit(1)
print("Build OK!")

# ============================================================
# STEP 4: Run tests and see what fails
# ============================================================
r2 = subprocess.run(["go", "test", "./handlers/", "-count=1", "-v"], capture_output=True, text=True, cwd=BASE, timeout=90)
lines = (r2.stdout + r2.stderr).split('\n')
passes = sum(1 for l in lines if '--- PASS:' in l)
fails = sum(1 for l in lines if '--- FAIL:' in l)
print(f"\nStep 4: {passes} PASS, {fails} FAIL")

# Save the current state as a new backup
shutil.copy(TEST, f"{BASE}/handlers/all_features_test.go.bak2")

# ============================================================
# STEP 5: Identify failing tests and fix them
# ============================================================
failing_tests = []
for l in lines:
    m = re.search(r'--- FAIL:\s+(\S+)', l)
    if m:
        failing_tests.append(m.group(1))

print(f"Failing tests: {failing_tests}")

# For each failing test, we need to check what SQL the handler sends
# and fix the mock accordingly.
# Let me focus on the most critical fixes.

print("\nApplying SQL fixes...")

# Fix 1: Products_GetProducts - need WithArgs on COUNT and images queries
replace(
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n\t\tWithArgs("active").\n\t\tWillReturnRows(')

replace(
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# For 2nd product images query
replace(
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "url", "is_primary", "created_at"}).\n\t\t\tAddRow(2, 2, "http://example.com/img2.png", true, mockTime(2024, time.January, 1)))',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(2).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "url", "is_primary", "created_at"}).\n\t\t\tAddRow(2, 2, "http://example.com/img2.png", true, mockTime(2024, time.January, 1)))')

# Fix 2: Products_GetProductBySlug
replace(
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products WHERE slug = $1`).\n\t\tWithArgs("test-product").\n\t\tWillReturnRows(')

replace(
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# Fix 3: SearchProducts - needs different WHERE clause
replace(
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`).\n\t\tWithArgs("active").\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1 AND (p.name LIKE $2 OR p.description LIKE $2 OR p.slug LIKE $2)`).\n\t\tWithArgs("active", "test").\n\t\tWillReturnRows(')

# Fix 4: Cart operations - add WithArgs
replace(
    'mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT id FROM carts WHERE user_id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(')

# Fix 5: Wishlist - add WithArgs to EXISTS query
replace(
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`).\n\t\tWithArgs(1, 1).\n\t\tWillReturnRows(')

# Fix 6: Reviews - add WithArgs
replace(
    'mock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) ON CONFLICT (product_id, user_id) DO UPDATE SET rating = $3, comment = $4, updated_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(',
    'mock.ExpectExec(`INSERT INTO reviews (product_id, user_id, rating, comment) VALUES ($1, $2, $3, $4) ON CONFLICT (product_id, user_id) DO UPDATE SET rating = $3, comment = $4, updated_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(1, 1, 5, "Great product!").WillReturnResult(')

# Fix 7: RecentlyView - add WithArgs
replace(
    'mock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id, viewed_at) VALUES ($1, $2, CURRENT_TIMESTAMP) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(',
    'mock.ExpectExec(`INSERT INTO recently_viewed (user_id, product_id, viewed_at) VALUES ($1, $2, CURRENT_TIMESTAMP) ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = CURRENT_TIMESTAMP`).\n\t\tWithArgs(1, 1).WillReturnResult(')

# Fix 8: Comparison - add WithArgs
replace(
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).\n\t\tWillReturnRows(',
    'mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM product_comparison WHERE user_id = $1 AND product_id = $2)`).\n\t\tWithArgs(1, 1).\n\t\tWillReturnRows(')

# Fix 9: Recommendations - fix category lookup
replace(
    'mock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"category_id"}).AddRow(nil))',
    'mock.ExpectQuery(`SELECT category_id FROM products WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnRows(sqlmock.NewRows([]string{"category_id"}).AddRow(nil))')

# Fix 10: Admin order status - fix UPDATE SQL
for status in ['paid', 'completed', 'cancelled', 'refunded']:
    extra = ", paid_at = NOW()" if status == "paid" else ""
    # Find the specific occurrence and fix
    old = f'mock.ExpectExec(`UPDATE orders SET.*`).\n\t\tWithArgs(1).\n\t\tWillReturnResult('
    new = f'mock.ExpectExec(`UPDATE orders SET status = \'{status}\'{extra} WHERE id = $1`).\n\t\tWithArgs(1).\n\t\tWillReturnResult('
    # Use regex to replace only the right occurrence
    pattern = f'mock\\.ExpectExec\\(`UPDATE orders SET\\.\\*`\\)\\.\n\t\tWithArgs\\(1\\)\\.\n\t\tWillReturnResult\\('
    
with open(TEST, 'w') as f:
    f.write(content)

print(f"Total changes: {changes}")

# Rebuild and test
r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True, cwd=BASE)
print(f"Build: exit={r.returncode}")
if r.returncode:
    print(r.stderr[:300])
else:
    r2 = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True, cwd=BASE, timeout=90)
    lines = (r2.stdout + r2.stderr).split('\n')
    passes = sum(1 for l in lines if '--- PASS:' in l)
    fails = sum(1 for l in lines if '--- FAIL:' in l)
    print(f"\nResult: {passes} PASS, {fails} FAIL, exit={r2.returncode}")
    for l in lines:
        if '--- FAIL:' in l or 'unfulfilled' in l or 'panic' in l:
            print(f"  {l.strip()[:120]}")
