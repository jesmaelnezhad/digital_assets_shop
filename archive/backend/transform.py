#!/usr/bin/env python3
"""
Transform all_features_test.go:
1. Convert all double-quoted mock SQL strings to backtick raw strings
2. Fix \\\$ -> \$ inside raw strings (so sqlmock regex matches literal $)
3. Fix \\\\* -> \* for COUNT(*)
4. Add missing WithArgs to mocks
5. Fix column lists to match handler expectations
"""
import re

PATH = "/root/project/backend/handlers/all_features_test.go"
with open(PATH) as f:
    content = f.read()

# ============================================================
# STEP 1: Convert double-quoted SQL strings to backtick raw strings
# Pattern: mock.ExpectQuery("SQL...") -> mock.ExpectQuery(`SQL...`)
# We need to: find "..." patterns that contain SQL, convert to `...`
# But be careful not to touch other double-quoted strings (like JSON keys/values)
# 
# Approach: find mock.Expect*(" patterns and convert the string to backticks
# Regex: mock\.(ExpectQuery|ExpectExec)\("([^"]*\.\*)?"\) 
# The SQL string always ends with either `.*` or `$N` or similar SQL patterns
# ============================================================

def convert_double_to_backtick(content):
    """Convert mock SQL from double-quoted to backtick-delimited."""
    result = []
    i = 0
    while i < len(content):
        # Find mock.ExpectQuery(" or mock.ExpectExec("
        m = re.search(r'mock\.(ExpectQuery|ExpectExec)\("', content[i:])
        if not m:
            result.append(content[i:])
            break
        
        start = i + m.start()
        end_pos = i + m.end()  # right after the opening quote
        
        # Find the closing quote - it's the next unescaped "
        # In a double-quoted Go string, \\" is an escaped quote
        j = end_pos
        while j < len(content):
            if content[j] == '"' and (j == 0 or content[j-1] != '\\'):
                break
            j += 1
        
        if j >= len(content):
            result.append(content[i:])
            break
        
        # Extract the SQL string (without quotes)
        sql = content[end_pos:j]
        
        # Convert to backtick raw string
        # In raw string: \$ is literal \$ (for regex matching $)
        # The double-quoted version had \\$ for literal $ -> raw string should have \$
        # But wait - in the double-quoted string, \\$ means: literal backslash + end-of-line
        # In raw string, \$ means: literal backslash + dollar (regex matches literal $)
        # We need sqlmock to see: \$1 (regex: literal $, then 1)
        # In raw string: \$1 → sqlmock sees \$1 → regex: literal $, then 1  ✓
        # In double-quoted: \\$1 → Go string value: \$1 → sqlmock sees \$1 → same! ✓
        # 
        # Wait, the backup has \\\$ in double-quoted strings.
        # \\\$ in Go double-quoted string: \\ = \, \$ = literal $ (invalid escape in Go 1.22!)
        # Actually Go 1.22 gives "unknown escape sequence" for \$ in double-quoted strings
        # The backup file IS broken. That's why we get the compile error.
        # 
        # So we MUST convert these to raw strings AND fix the escaping.
        # In raw string: \$ is literal \$ → sqlmock regex sees \$ → matches literal $
        # In raw string: \\ is literal \\ → sqlmock regex sees \\ → matches literal \
        # 
        # The handler sends: $1 (no backslash)
        # sqlmock needs regex: \$1 → Go raw string: \$1 (or `\$1`)
        # 
        # The backup has: "SELECT ... \\\\\$1" in double-quoted
        # Go value: SELECT ... \\$1 → sqlmock regex: \\$1 → matches literal \ then end-of-line
        # WRONG. Should be: $1
        # 
        # In raw string: `SELECT ... \$1` → sqlmock gets: SELECT ... \$1
        # regex: \$ matches literal $ → matches "SELECT ... $1" correctly!
        
        # Fix the SQL for raw string usage:
        # Replace \\$ with \$ (in raw string context)
        # But the string might have other escapes we need to handle
        
        # Actually, for raw strings, we just need to put the SQL as-is,
        # but fix the \\$ → \$ issue.
        # The sql variable contains the VALUE from the double-quoted string.
        # If source had "SELECT ... \\\\\\$1", Go value is "SELECT ... \\$1"
        # Wait no. Let me re-examine.
        # 
        # The backup source file has on disk: "SELECT ... \\\\\\$1"
        # In Go, this is a regular string. \\\\ = \\ (2 backslashes in source = 1 backslash in value)
        # So \\\\\\$ in source = \\\$ in value (3 chars: \ \ $)
        # But Go says \$ is an unknown escape sequence. So this shouldn't compile.
        # Unless... let me check if the file actually has \\\$ or \\$ on disk.
        
        # For now, let me just convert to raw strings and fix obvious issues.
        # The sql string (value from double-quoted) needs to be put in backticks.
        # If it has \\$ in the value, we keep it as \\$ in the raw string.
        # If it has \$ in the value, we keep it as \$ in the raw string.
        
        # Build the replacement
        replacement = 'mock.' + m.group(1) + '(`' + sql + '`)'
        
        result.append(content[i:start])
        result.append(replacement)
        i = j + 1  # skip past closing quote
    
    return ''.join(result)

# Apply conversion
content = convert_double_to_backtick(content)
print("Converted double-quoted SQL strings to backtick raw strings")

# ============================================================
# STEP 2: Fix \\\$ -> \$ inside raw strings
# ============================================================
bt = ord('`')
bs = ord('\\')
dl = ord('$')
result = bytearray()
i = 0
depth = 0
while i < len(content):
    b = content[i] if isinstance(content, bytes) else ord(content[i])
    if isinstance(content, str):
        b = ord(content[i])
    
    ch = content[i] if i < len(content) else None
    if ch == '`':
        depth += 1
        result.append(b'`')
        i += 1
        while i < len(content) and content[i] != '`':
            c = content[i]
            if depth % 2 == 1 and i + 2 < len(content) and content[i] == '\\' and content[i+1] == '\\' and content[i+2] == '$':
                result.append('\\')
                result.append('$')
                i += 3
            else:
                result.append(c)
                i += 1
        if i < len(content):
            result.append('`')
            i += 1
    else:
        result.append(ch)
        i += 1

content = bytes(result).decode('utf-8') if isinstance(result, bytearray) else ''.join(result)
print("Fixed \\\$ -> \$ in raw strings")

# ============================================================
# STEP 3: Fix COUNT star escaping  
# ============================================================
content = content.replace('\\\\*', '*')
print("Fixed COUNT star escaping")

# Write
with open(PATH, "w") as f:
    f.write(content)
print("Done writing file")
