#!/usr/bin/env python3
"""Simple, robust fix: convert double-quoted mock SQL to backtick raw strings."""
import re

with open("/root/project/backend/handlers/all_features_test.go", "r") as f:
    content = f.read()

# Find all lines with mock.ExpectQuery("...") or mock.ExpectExec("...")
# Convert the "..." to `...` with proper escaping

lines = content.split('\n')
out = []

for line in lines:
    stripped = line.strip()
    
    # Check if this line has a mock call with double-quoted SQL
    has_mock = ('mock.ExpectQuery("' in stripped or 'mock.ExpectExec("' in stripped)
    
    if not has_mock:
        out.append(line)
        continue
    
    # Find the mock call boundaries
    # mock.ExpectQuery("SQL...") or mock.ExpectExec("SQL...")
    
    # Find start: position of mock.ExpectQuery( or mock.ExpectExec(
    start_search = stripped.find('mock.ExpectQuery("')
    if start_search < 0:
        start_search = stripped.find('mock.ExpectExec("')
    
    if start_search < 0:
        out.append(line)
        continue
    
    # Find opening quote
    open_q = start_search + len('mock.ExpectQuery("') if 'mock.ExpectQuery("' in stripped else start_search + len('mock.ExpectExec("')
    
    # Find closing quote - scan past escaped chars
    i = open_q + 1
    sql_chars = []
    while i < len(stripped):
        c = stripped[i]
        if c == '' and i + 1 < len(stripped):
            sql_chars.append('' + stripped[i+1])
            i += 2
        elif c == '"':
            break
        else:
            sql_chars.append(c)
            i += 1
    
    sql_content = ''.join(sql_chars)
    close_q = i
    
    # What's after the closing quote?
    after = stripped[close_q+1:]  # should be ).WillReturn... or ).\n...
    
    # Fix SQL content for backtick:
    # In the double-quoted source, we have Go escape sequences.
    # For backtick, we need the literal characters.
    #
    # The SQL content currently has sequences like:
    # \\$1 -> Go interprets as \$1 (but \$ is invalid, so file doesn't compile)
    # We want backtick to contain: \$1 (literal \ + $ + 1)
    #
    # In the FILE BYTES, between the " quotes, we have the raw characters.
    # For the line: mock.ExpectQuery("SELECT ... = \\$1").
    # The bytes between " are: SELECT ... = \\$1
    # (\\ is two bytes: backslash, backslash. Then $1.)
    # Go interprets \\ as \, then sees $1. But \$ is invalid escape -> ERROR.
    #
    # For backtick, we want: SELECT ... = \$1
    # (\ is one byte: backslash. Then $1.)
    # RE2 sees \$1 and matches literal $1. CORRECT!
    #
    # So: in the sql_content (raw bytes between quotes), replace \\ with \ (one level).
    # But ONLY for backslashes that are part of escape sequences, not literal backslashes in SQL.
    #
    # Actually the simplest correct approach:
    # The SQL content in the source file (between ") has certain bytes.
    # We want the backtick to have bytes that, when interpreted by RE2, match the handler's SQL.
    # 
    # The handler sends: SELECT ... WHERE email = $1
    # For RE2 to match this, the pattern needs: SELECT ... WHERE email = \$1
    # (\ is regex escape for literal $)
    #
    # In Go backtick: `SELECT ... WHERE email = \$1`
    # This is exactly: SELECT ... WHERE email = [backslash]$1
    # 
    # In the double-quoted source file, what bytes do we have?
    # The file has: mock.ExpectQuery("SELECT ... WHERE email = \\\\\$1").
    # Wait, let me check the actual bytes more carefully.
    
    # From earlier cat -A: ^Imock.ExpectQuery("SELECT id, email, ... WHERE email = \\\\\$1").$
    # cat -A shows \\ for each backslash byte.
    # \\\\\$ = 6 backslash bytes? No...
    # cat -A: \\ = one backslash byte displayed as two chars \\
    # So \\\\\\\$ in cat -A = 6 backslash bytes? That seems like a lot.
    
    # Let me just check: the backup file has invalid Go escapes.
    # The simplest fix that WILL work: use Python to read the raw bytes,
    # and for each $N in the SQL, ensure it's preceded by exactly one backslash.
    
    # Replace any sequence of backslashes before $N or ( or ) with exactly one backslash
    sql_fixed = sql_content
    
    # Normalize: \ before $ -> exactly one \
    sql_fixed = re.sub(r'\\+(\$)', r'\\\\\1', sql_fixed)  # Hmm, this is wrong
    
    # Let me think about Python replacement strings:
    # r'\\\\\1' = literal backslash + backslash + group1
    # That's TWO backslashes + $. Not what I want.
    # 
    # I want ONE backslash + $. In Python replacement:
    # r'\\\\\1' -> \\ + $  (two chars: \, $)
    # Wait: r'\\' = one backslash in the output. r'\1' = group 1.
    # So r'\\\\\1' = \ + \1? No...
    # 
    # Python replacement string escapes:
    # \\ -> literal backslash
    # \1 -> group 1
    # So r'\\\\\1' = \\ (one backslash) + \1 (group 1) = \ + $ = \$  ✓ CORRECT!
    
    sql_fixed = re.sub(r'\\+(\$)', r'\\\\\1', sql_fixed)
    # This replaces: [any backslashes] + $  with  \ + $
    # Result: \$  ✓
    
    # Also handle \\( and \\) for literal parens in regex
    sql_fixed = re.sub(r'\\+(\()', r'\\\\\1', sql_fixed)  # \(
    sql_fixed = re.sub(r'\\+(\))', r'\\\\\1', sql_fixed)  # \)
    
    # Handle \\. for literal dot
    sql_fixed = re.sub(r'\\+(\.)', r'\\\\\1', sql_fixed)
    
    # Now build the output line
    indent = line[:len(line) - len(line.lstrip())]
    fixed_line = indent + 'mock.' + ('ExpectQuery' if 'mock.ExpectQuery("' in stripped else 'ExpectExec') + '(`' + sql_fixed + '`)' + after
    out.append(fixed_line)

result = '\n'.join(out)

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(result)

print("Done")
# Verify
with open("/root/project/backend/handlers/all_features_test.go", "r") as f:
    vl = f.readlines()

print(f"Total lines: {len(vl)}")
for i in [13, 26, 27, 41, 77, 122, 123, 127, 128]:
    if i < len(vl):
        print(f"L{i+1}: {vl[i].rstrip()[:150]}")
