#!/usr/bin/env python3
"""Fix all_features_test.go: convert to backtick raw strings, fix SQL, add WithArgs."""
import re, os, subprocess, shutil

BASE = "/root/project/backend"
TEST = f"{BASE}/handlers/all_features_test.go"

# Restore from backup
shutil.copy(f"{BASE}/handlers/all_features_test.go.bak", TEST)

with open(TEST) as f:
    text = f.read()

# ============================================================
# STEP 1: Convert double-quoted mock strings to backtick raw strings
# ============================================================
# Pattern: mock.ExpectQuery("SQL...") -> mock.ExpectQuery(`SQL...`)
# The "SQL..." on disk may contain escape sequences. 
# We want the raw string to contain exactly what the handler sends.

out = []
i = 0
n = 0
while i < len(text):
    # Find next mock expectation
    pos_q = text.find('mock.ExpectQuery("', i)
    pos_e = text.find('mock.ExpectExec("', i)
    if pos_q < 0 and pos_e < 0:
        out.append(text[i:])
        break
    
    if pos_q >= 0 and (pos_e < 0 or pos_q <= pos_e):
        pos, sym = pos_q, 'ExpectQuery'
    else:
        pos, sym = pos_e, 'ExpectExec'
    
    out.append(text[i:pos])
    
    # Find opening "
    oq = text.find('("', pos) + 2
    # Find closing " (next unescaped ")
    cq = oq
    while cq < len(text) and text[cq] != '"':
        cq += 1
    
    if cq >= len(text):
        out.append(text[pos:])
        break
    
    sql_raw = text[oq:cq]  # Raw on-disk text between double quotes
    
    # Convert to raw string.
    # On disk: "SELECT ... \\\\\$1" -> sql_raw = 'SELECT ... \\\\\$1' (the bytes as-is)
    # For Go raw string: we want `SELECT ... \$1` (so sqlmock sees \$1 = regex for literal $)
    # 
    # In sql_raw (the on-disk bytes): 
    #   \\\\ represents 2 actual backslash chars
    #   \\\\$ represents 2 backslashes + dollar = 3 chars
    # 
    # For Go raw string `\$1`:
    #   On disk: \, $, 1 = 3 chars
    #   sqlmock value: \, $, 1 -> regex: \$ matches literal $, then 1
    #   This matches handler's $1. CORRECT.
    #
    # The sql_raw on disk has: SELECT ... \\\\\$1
    # That's: ... \, \, \, \, $ 1 = 5+ chars? Let me count differently.
    # 
    # From the cat -A output: \\\\\\$  (6 chars on screen = 6 chars on disk for cat -A)
    # Wait, cat -A shows backslashes literally. So \\\\\\$ on screen = \, \, \, \, \, $ on disk = 6 chars
    # But Python repr showed: b'\\\\\\$' which is 4 chars: \, \, \, $
    # These don't match. Let me trust the Python repr since it comes from reading the actual file.
    #
    # Python: b'\\\\\\$' -> bytes: [92, 92, 92, 36] = \, \, \, $ (4 bytes)
    # Hmm no. In Python bytes repr, \\\\ represents 2 bytes (each \\ = one backslash).
    # So \\\\\\$ in repr = \\\\ + \$ = 6 bytes? No, \$ in repr = just $ (backslash before $ has no special meaning in bytes repr).
    # Actually in Python repr: \\\\ = 2 bytes (backslash backslash), \$ = 1 byte (dollar) because \$ is not a recognized escape.
    # Wait, in Python, \$ in a bytes literal is just $ because \$ is not an escape sequence.
    # So b'\\\\\\$' = b'\\\\' + b'$' = 2 backslashes + dollar = 3 bytes.
    # But b'\\\\' = 2 bytes (two backslashes). So b'\\\\\\$' = b'\\\\' + b'$' = \, \, $ = 3 bytes.
    #
    # OK so the on-disk bytes between the double quotes for line 27 contain: \, \, $ (3 bytes) before the 1.
    # That's \\\$ in raw form (backslash backslash dollar).
    # In Go double-quoted string: \\\\ = \, so \\\\\$... wait.
    # 
    # On disk: "SELECT ... \\\\\$1"
    # The chars between quotes: SELECT ... \, \, \, $, 1
    # No! Let me re-read the cat -A output:
    # ^Imock.ExpectQuery("SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email = \\\\\\$1").$
    # 
    # The \\\\\\$ on screen: each \ is one byte. So that's 6 bytes: \, \, \, \, \, $
    # But wait, that doesn't match what I computed from Python repr.
    # Let me just check by running a quick Python script.
    pass
    
    # For now, use a simple approach: take the on-disk bytes as-is for the raw string.
    # But fix the most obvious issue: if there are 3 consecutive backslashes followed by $,
    # that should become backslash + dollar in the raw string.
    
    # Actually, the SIMPLEST correct approach:
    # The on-disk text between double quotes is used literally in the raw string.
    # BUT we need to account for the fact that the Go double-quoted string had escape processing.
    # 
    # On disk: "SELECT ... \\\\\$1"
    # Go double-quoted interpretation: \\\\ -> \, \$ -> $ (but \$ is invalid, so error)
    # 
    # If it were valid (e.g., "SELECT ... \\$1"):
    # Go double-quoted: \\$ -> \$ -> value: \, $ -> sqlmock: \$ -> matches literal $
    # 
    # For raw string, we want the VALUE (not the source) to be \, $ on disk.
    # So if the Go double-quoted source was "SELECT ... \\$1" (on disk: \, \, $, 1)
    # The Go value would be: \, $, 1 -> sqlmock: \$1 -> matches $1 ✓
    # For raw string: we put \, $, 1 between backticks -> `SELECT ... \$1`
    # On disk: \, $, 1 = 3 bytes between backticks.
    # 
    # But our source has MORE backslashes. On disk between quotes: \, \, \, \, $, 1 (5+ bytes?)
    # If Go double-quoted: \\\\ = \, \\$ = \$ (invalid) 
    # 
    # I think the issue is that the backup was created from a file where someone wrote
    # "SELECT ... \\\\\$1" intending it to be a valid Go double-quoted string that evaluates to "\$1".
    # But \\$ is not a valid Go escape in double-quoted strings, so it fails at compile time.
    # 
    # The INTENDED value was: "\$1" (2 chars: \, $) which in raw string is `\$1`.
    # The INTENDED on-disk source for raw string: `SELECT ... \$1` (with \, $ on disk).
    #
    # So we need to convert: on-disk "SELECT ... \\\\\$1" -> on-disk `SELECT ... \$1`
    # That means: between the quotes, replace \\\\\$ (whatever number of backslashes) with \$
    # 
    # The simplest transform: find all runs of backslashes followed by $, collapse to \$
    # And find all runs of backslashes followed by *, collapse to \*
    # And find all runs of backslashes followed by (, keep as \(
    # And find all runs of backslashes followed by ), keep as \)
    
    # Let me process the SQL to convert it properly
    sql_processed = []
    j = 0
    while j < len(sql_raw):
        if sql_raw[j] == '\\':
            # Count consecutive backslashes
            start = j
            while j < len(sql_raw) and sql_raw[j] == '\\':
                j += 1
            nslashes = j - start
            if j < len(sql_raw):
                next_char = sql_raw[j]
                if next_char == '$':
                    # Collapse to \$
                    sql_processed.append('\\')
                    sql_processed.append('$')
                    j += 1
                elif next_char == '*':
                    # Collapse to \*  (keep backslash for regex)
                    sql_processed.append('\\')
                    sql_processed.append('*')
                    j += 1
                elif next_char in '()':
                    # Keep \( or \)
                    sql_processed.append('\\')
                    sql_processed.append(next_char)
                    j += 1
                else:
                    # Just keep one backslash
                    sql_processed.append('\\')
                    # Don't consume next_char, it'll be processed next iteration
            else:
                # Trailing backslashes
                sql_processed.append('\\')
        else:
            sql_processed.append(sql_raw[j])
            j += 1
    
    sql_final = ''.join(sql_processed)
    replacement = f'mock.{sym}(`{sql_final}`)'
    out.append(replacement)
    i = cq + 1
    n += 1

result_text = ''.join(out)
print(f"Step 1: Converted {n} mock strings to backtick raw strings")

# ============================================================
# STEP 2: Write and verify build
# ============================================================
with open(TEST, 'w') as f:
    f.write(result_text)

r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True, cwd=BASE)
print(f"Build: exit={r.returncode}")
if r.returncode:
    print("STDERR:", r.stderr[:500])
    exit(1)

print("Build OK!")

# Save intermediate backup
shutil.copy(TEST, f"{BASE}/handlers/all_features_test.go.bak2")

# ============================================================
# STEP 3: Run tests
# ============================================================
r2 = subprocess.run(["go", "test", "./handlers/", "-count=1", "-v"], capture_output=True, text=True, cwd=BASE, timeout=90)
lines = (r2.stdout + r2.stderr).split('\n')
passes = sum(1 for l in lines if '--- PASS:' in l)
fails = sum(1 for l in lines if '--- FAIL:' in l)
print(f"\nStep 3: {passes} PASS, {fails} FAIL, exit={r2.returncode}")
for l in lines:
    if '--- FAIL:' in l or (('unfulfilled' in l or 'panic' in l) and len(l.strip()) > 10):
        print(f"  {l.strip()[:130]}")

# ============================================================
# STEP 4: Apply targeted fixes for failing tests
# ============================================================
if fails > 0:
    print(f"\nApplying fixes for {fails} failing tests...")
    
    # Fix: Add WithArgs to mocks that need them
    # Pattern: mock.Expect*(`SQL with $1`) followed by WillReturn* without WithArgs
    
    fixes = 0
    
    # Helper: add WithArgs before WillReturn*
    def add_withargs(sql_contains, args_str, label):
        global result_text, fixes
        # Find: mock.Expect*(`SQL...`) followed by \n\t\tWillReturn*
        # Replace with: mock.Expect*(`SQL...`) followed by \n\t\tWithArgs(...).WillReturn*
        old = f'mock.{sym}(`{sql_contains}`)'
        # Find the exact occurrence  
        idx = result_text.find(old)
        if idx < 0:
            print(f"  SKIP: {label} (pattern not found)")
            return
        
        # Find WillReturn after this
        wr_idx = result_text.find('WillReturn', idx)
        if wr_idx < 0:
            print(f"  SKIP: {label} (no WillReturn)")
            return
        
        # Insert WithArgs before WillReturn
        insert_pos = wr_idx
        # Check if WithArgs already exists
        wa_idx = result_text.find('WithArgs', idx, wr_idx)
        if wa_idx >= 0:
            print(f"  SKIP: {label} (already has WithArgs)")
            return
        
        # Build the WithArgs line
        indent = '\t\t'
        result_text = result_text[:insert_pos] + f'{indent}WithArgs({args_str}).\n' + result_text[insert_pos:]
        fixes += 1
        print(f"  OK: {label}")
    
    # Extract the SQL for each mock and add WithArgs
    # Parse the file to find all mock expectations and their SQL
    mock_pattern = re.compile(r'mock\.(ExpectQuery|ExpectExec)\(`([^`]+)`\)')
    
    for m in mock_pattern.finditer(result_text):
        sym = m.group(1)
        sql = m.group(2)
        
        # Check if this SQL has $N placeholders
        params = re.findall(r'\$(\d+)', sql)
        if not params:
            continue  # No args needed
        
        # Check if WithArgs already exists after this mock
        after_idx = m.end()
        next_bit = result_text[after_idx:after_idx+200]
        if 'WithArgs' in next_bit.split('\n')[0] if '\n' in next_bit else '':
            continue  # Already has WithArgs
        
        # Determine args based on SQL content
        param_count = len(params)
        max_param = max(int(p) for p in params)
        
        if 'email' in sql.lower() and '$1' in sql:
            args = '"test@test.com"'
        elif 'id = $1' in sql and 'user_id' in sql:
            args = '1, 1'
        elif 'user_id = $1' in sql:
            args = '1'
        elif 'slug = $1' in sql:
            args = '"test-product"'
        elif 'status = $1' in sql and 'key' not in sql:
            args = '"active"'
        elif 'key = $1' in sql:
            args = '"payment_address"'
        elif 'chain' in sql.lower() and param_count <= 3:
            args = '"BSC", "BNB", "0.03"'
        elif 'product_id = $1' in sql and 'ORDER BY' in sql:
            args = '1'
        elif 'WHERE id = $1' in sql or 'WHERE id = \\$1' in sql:
            args = '1'
        elif 'user_id = $1' in sql and 'product_id' in sql:
            args = '1, 1'
        elif max_param == 1:
            args = '1'
        elif max_param == 2:
            args = '1, 1'
        elif max_param == 3:
            args = '1, 1, 1'
        elif max_param == 4:
            args = '1, 1, 1, 1'
        else:
            args = ','.join(['1'] * param_count)
        
        # Do the replacement
        old = f'mock.{sym}(`{sql}`)'
        if old in result_text:
            # Find position after this mock call
            pos = result_text.find(old) + len(old)
            # Find the WillReturn or code line
            rest = result_text[pos:pos+300]
            wr_pos = rest.find('WillReturn')
            if wr_pos >= 0:
                full_pos = pos + wr_pos
                indent = '\t\t'
                result_text = result_text[:full_pos] + f'{indent}WithArgs({args}).\n' + result_text[full_pos:]
                fixes += 1
    
    print(f"Applied {fixes} WithArgs fixes")
    
    with open(TEST, 'w') as f:
        f.write(result_text)
    
    # Rebuild
    r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True, cwd=BASE)
    print(f"Rebuild: exit={r.returncode}")
    if r.returncode:
        print("STDERR:", r.stderr[:300])
    else:
        r2 = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True, cwd=BASE, timeout=90)
        lines = (r2.stdout + r2.stderr).split('\n')
        passes = sum(1 for l in lines if '--- PASS:' in l)
        fails = sum(1 for l in lines if '--- FAIL:' in l)
        print(f"\nResult: {passes} PASS, {fails} FAIL, exit={r2.returncode}")
        for l in lines:
            if '--- FAIL:' in l:
                print(f"  {l.strip()[:120]}")
