#!/usr/bin/env python3
"""Fix all_features_test.go: correct \\$ to \$ and fix all mock SQL patterns."""
import re

PATH = "/root/project/backend/handlers/all_features_test.go"
with open(PATH) as f:
    content = f.read()

# Fix 1: \\to \$ in all backtick-delimited raw strings
# Go raw strings don't process backslash escapes, so \$ in a raw string 
# is literally backslash+dollar. But sqlmock regex engine sees \$1 as matching
# literal dollar-sign-1, which is what we want.
# The issue is: the file has \\\$ which is THREE chars in source: \ \ $
# In a Go raw string, this is literally: \\\$  (backslash, backslash, dollar)
# sqlmock sees: \\$  which in regex means: literal backslash, then $ (end-of-line)
# But handler sends: $1  (no backslash)
# So we need the file to contain: \$  (two chars: \ $) in the Go raw string
# Currently it has: \\$  (three chars: \ \ $) in the Go raw string

# In the file as text on disk:
# \\\$ appears as: backslash backslash backslash dollar  (4 chars... no)
# Let me count: the source text `\\\$` in a Go source file:
# The actual bytes are: 0x5C 0x5C 0x24  (backslash, backslash, dollar) = 3 bytes
# But `\$` is also valid and means: backslash, dollar = 2 bytes
# 
# Wait. In a Go source file, raw strings are delimited by backticks.
# Inside a raw string, every character is literal. No escape processing.
# So the source text `SELECT * FROM t WHERE id = \$1` contains:
# SELECT * FROM t WHERE id = \$1  (with a literal backslash before $)
# That's: ...= \, $, 1  (backslash, dollar, one)
# 
# The source text `SELECT * FROM t WHERE id = \\$1` contains:
# SELECT * FROM t WHERE id = \\$1  (with TWO literal backslashes before $)
# That's: ...= \, \, $, 1  (two backslashes, dollar, one)
# 
# sqlmock receives the string VALUE (not source). For raw strings, the value
# is exactly the characters between the backticks.
# 
# For `... \$1 ...`: value = ... \$1 ... → sqlmock gets: ... \$1 ...
#   In regex: \$ matches literal $, so this matches "...= $1 ..."  ✓
# 
# For `... \\$1 ...`: value = ... \\$1 ... → sqlmock gets: ... \\\$1 ... 
#   Wait no. The value is: ... \\\$1 ... no.
#   The value is EXACTLY what's between the backticks: ... \\\$1 ...
#   No! The backtick-delimited raw string value is the literal characters.
#   If source has `\\\$1`, the value is: \\\$1 (backslash backslash dollar 1)
#   sqlmock regex: \\\$ → matches literal backslash followed by end-of-line
#   This does NOT match handler's $1. ✗
#
# So the fix is: replace \\$ with \$ in ALL backtick strings.
# In the FILE: replace the 3-byte sequence 0x5C 0x5C 0x24 with 0x5C 0x24
# But ONLY inside backtick-delimited strings.

result = []
i = 0
depth = 0
while i < len(content):
    if content[i] == '`':
        # Backtick: toggle depth
        depth += 1
        result.append('`')
        i += 1
        # Inside a raw string (depth is odd), fix \\$ -> \$
        while i < len(content) and content[i] != '`':
            if depth % 2 == 1 and i + 2 < len(content) and content[i] == '\\' and content[i+1] == '\\' and content[i+2] == '$':
                # Found \\$ inside a raw string → replace with \$
                result.append('\\')
                result.append('$')
                i += 3  # skip the three chars \ \$ 
            else:
                result.append(content[i])
                i += 1
        if i < len(content):
            result.append('`')
            i += 1
    else:
        result.append(content[i])
        i += 1

content = ''.join(result)
print("Fixed \\$ → \$ in all raw strings")

# Fix 2: Remove stray \\ escaping in other patterns  
# Some patterns have \\* which should be \* for regex matching literal *
content = content.replace(r"COUNT\(\*\\)", "COUNT\(\*\)")
print("Fixed COUNT star escaping")

# Fix 3: Fix "unknown escape sequence" errors
# Go raw strings don't have escape sequences, but regular strings do.
# The issue at line 27: "SELECT ... \\\\\$1" in a regular (double-quoted) string
# In Go regular strings, \\\\ is two literal backslashes (each \\ = one \)
# So \\\\\\$ = \\\$  (two backslashes + dollar) in the string value
# sqlmock sees: \\$ → regex: literal backslash + end-of-line
# Handler sends: $1 → sqlmock needs: \$ → Go regular string: "\\$"
# Fix: in double-quoted strings, replace \\\\\$ with \\$
# But this is tricky because we need to distinguish raw strings from regular strings.
# Let me check: the file at line 27 uses double quotes: "SELECT ... \\\\\$1"
# That's a regular Go string, not a raw string.
# In a regular Go string: \\\\ = \\ (two backslashes in source = one backslash in value)
# So \\\\\\$ in source = \\\$ in value (two backslashes + dollar in value)
# sqlmock sees: \\$ → matches literal backslash + end-of-line
# We need: \$ in value → Go regular string source: "\\$"
# 
# In the source file: replace \\\\\$ (in double-quoted strings) with \\$
# But let me check what's actually in the file at line 27

print("\nChecking line 27:")
lines = content.split('\n')
for i in range(26, 32):
    print(f"  {i+1}: {lines[i][:120]}")

# The pattern on line 27 uses double quotes (regular string), not backticks.
# "SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email = \\\\\\$1"
# In this regular string, \\\\\\$  in source = \\\$ in value
# We need \$ in value → source should be "\\$"
# Fix: replace \\\\\\$ with \\$ in the entire file (for regular strings)
# But wait, this is a mix of raw and regular strings.
# Let me just check: how many \\\\\$ patterns are there vs \\\$ patterns?
backslash_dollar_count = content.count('\\\\\\$')
single_backslash_dollar = content.count('\\$')
print(f"\n\\\\\\\$ count: {backslash_dollar_count}")
print(f"\\$ count: {single_backslash_dollar}")

# Actually in Python string, \\\\\$ represents 4 chars: \ \ \ $
# And \\$ represents 2 chars: \ $
# Let me count in the raw file bytes
with open(PATH, 'rb') as f:
    raw = f.read()
# Count occurrences of backslash-backslash-dollar (3 bytes: 0x5C 0x5C 0x24)
count_dd = raw.count(b'\\\\\\$')  # Python: \\\\ = \\, \\\\ = \\, \\$= $ → 3 bytes... no
# b'\\\\\\$' in Python = bytes: \, \, \, $ (4 bytes)
# b'\\$' in Python = bytes: \, $ (2 bytes) 
# We want to count: \, \, $ (3 bytes) in the raw file
count_2bs = raw.count(b'\\\\$')  # In Python bytes literal: \\\\ = \\, $ = $ → 3 bytes? No.
# b'\\\\$': \\\\ is two chars: \ \, $ is one char: $ → total 3 chars: \, \, $
# No! In Python bytes: \\\\ represents \\, so b'\\\\$' = \\ + $ = 2 bytes
# We need b'\\\\\\$' for 3 bytes: \\\\ + \\\\ + \\$= no this is confusing.
# Let me just count the bytes directly
count_3 = 0
i = 0
while i < len(raw) - 2:
    if raw[i] == 0x5C and raw[i+1] == 0x5C and raw[i+2] == 0x24:  # \ \ $
        count_3 += 1
        i += 3
    else:
        i += 1
print(f"3-byte sequence \\\$ in file: {count_3}")

# Count 2-byte sequence \$
count_2 = 0
i = 0
while i < len(raw) - 1:
    if raw[i] == 0x5C and raw[i+1] == 0x24:  # \ $
        count_2 += 1
        i += 2
    else:
        i += 1
print(f"2-byte sequence \$ in file: {count_2}")

print(f"\nWill replace all \\\$ (3-byte) with \$ (2-byte) in raw strings")
