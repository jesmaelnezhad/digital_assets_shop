#!/usr/bin/env python3
"""Fix the escaping in all mock SQL patterns."""
import re

PATH = "/root/project/backend/handlers/all_features_test.go"
with open(PATH) as f:
    content = f.read()

# Fix 1: \\*  →  \*  (in COUNT queries where handler sends plain *)
# The file has: COUNT\(\*\\)  which in Go raw string is: COUNT(\\*)
# sqlmock regex receives COUNT(\\*) which matches literal backslash-star
# Handler sends COUNT(*) which needs regex pattern COUNT(\*)
# Fix: replace \\\* with \*
# In the Go file: \\\\* represents the Go raw string \\\* 
# We need to go from \\\* (which matches literal \) + literal *) to \* (which matches literal *)
content = content.replace(r"COUNT\(\*\\)", "COUNT\(\*\)")

# Fix 2: Chart patterns - fix \& escape sequences
# In Go raw strings, \& is just \&. No special meaning.
# But in the regex, \& is just &.
# The issue is \\\& in the file which becomes \\& in the string, matching literal \&
# Handler sends plain &. Fix: replace \\& with &
content = content.replace(r"\\)", ")")  # nope, this breaks things

# Let me be more careful. The patterns in the FILE (Go raw strings):
# \\\$1 means the Go source has: \\\$
# The Go string value is: \\$
# sqlmock regex gets: \\$  which matches literal \$ (wrong! handler sends $)
# Fix: replace \\\$ with \$

# But wait - \\\$ in the FILE represents three chars: backslash, backslash, dollar
# In Go raw string: \\\$
# Go string value: \\$ 
# sqlmock receives: \\$  → regex matches literal \$
# Handler sends: $1  → sqlmock needs: \$1  → Go raw string: \\\$1
# 
# So the file currently has: \\\$1 (Go raw string for \\$1) 
# It should have: \\\$1 as well? No wait...
#
# Let me think again. sqlmock uses Go regex. The pattern string is the SQL to match.
# If handler sends: SELECT * FROM t WHERE id = $1
# We need a regex that matches this. In Go regex, $ is end-of-line anchor.
# To match literal $, we need \$ in regex. In Go source (regular string): "\\$"
# In Go raw string: `\$`  (backslash-dollar)
#
# The test file uses backtick raw strings: `SELECT COUNT(*) FROM t WHERE id = \$1`
# This Go raw string produces: SELECT COUNT(*) FROM t WHERE id = \$1
# sqlmock compiles this as regex: SELECT COUNT(*) FROM t WHERE id = \$1
# \$ in regex matches literal $. Correct!
#
# But the file ACTUALLY has: `SELECT COUNT\(\*\\) FROM t WHERE id = \\\$1`
# Go raw string produces: SELECT COUNT\(\*\\) FROM t WHERE id = \\$
# sqlmock regex: SELECT COUNT\(\*\\) FROM t WHERE id = \\$
# This matches: SELECT COUNT(*) FROM t WHERE id = \$  (literal backslash-dollar)
# But handler sends: SELECT COUNT(*) FROM t WHERE id = $1  (no backslash)
# MISMATCH!
#
# FIX: In the Go raw string, we need `\$1` not `\\$1`.
# The file has \\\$ (three chars: \ \ $) → Go produces \\$ → regex matches \$
# We need \$ (two chars: \ $) → Go produces \$ → regex matches $
#
# In the file's raw string: `...` 
# Replace \\$ with \$ throughout.
# But we can't just replace \\$ because it appears in many contexts.
# The pattern is: inside a backtick-delimited raw string, replace \\$ with \$
# 
# Actually looking at the file more carefully:
# Line 127: mock.ExpectQuery(`SELECT COUNT\(\*\\) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = \\$1`).
# 
# The backtick-delimited string contains: COUNT\(\*\\) and \\$1
# Go raw string unescaping: \\ → \ and \$ → $
# So Go string becomes: COUNT\(\*\\) contains \\* (literal backslash-star)
# And \\$1 becomes \$1 (literal backslash-dollar)
# 
# Wait no. In Go raw strings, backslash is just a backslash. There's no unescaping.
# `\\` in a Go raw string is two backslashes: \\
# `\$` in a Go raw string is backslash-dollar: \$
# `\\\` in a Go raw string is backslash-backslash-backslash: \\\
# `\\\$` in a Go raw string is backslash-backslash-backslash-dollar: \\\$... no that's 4 chars
#
# Let me count: \\\$ = \ \ \ $  (backslash, backslash, backslash, dollar) = 4 chars
# No: \\ = 2 chars (two backslashes), \$ = 2 chars (backslash, dollar)
# \\\$ = \\ + \$ = 4 chars: \ \ $ ... that's wrong
#
# Actually in a raw string:
# \\ is literally two characters: backslash, backslash
# \$ is literally two characters: backslash, dollar
# So \\\$ would be: backslash, backslash, backslash, dollar = 4 chars
# But the string literal `\\\$` has: \ (the first \) \\ (second and third chars: \ \) $ 
# Hmm, raw strings don't have escape sequences. Each character is literal.
# So `\\\$` = \, \, \, $  ... that's 4 characters
# No. Let me count the characters in the source: 
# The source text `\\\$` consists of: \ \ \ $
# That's 4 characters: backslash, backslash, backslash, dollar
# In a Go raw string, this is 4 literal characters: \, \, \, $
#
# And `\$` in source = \, $ = 2 characters
#
# So the file has `\\\$1` which is \, \, \, $, 1 = 5 characters
# Go produces: \\\$1  (5 chars: backslash backslash backslash dollar 1)
# sqlmock regex gets: \\\$1
# This matches: \\\$1 literally... no
#
# OK I'm overcomplicating this. Let me check what the file actually contains
# by looking at the raw bytes.
print("Checking raw file content...")
with open(PATH, 'rb') as f:
    raw = f.read()

# Find the COUNT line
idx = raw.find(b"COUNT")
if idx >= 0:
    # Print 100 bytes around it
    chunk = raw[idx:idx+100]
    print(f"Raw bytes around COUNT: {chunk}")
    print(f"Hex: {chunk.hex()}")
    
# Also check a \$1 pattern
idx2 = raw.find(b"WHERE p.status = ")
if idx2 >= 0:
    chunk2 = raw[idx2:idx2+80]
    print(f"\nRaw bytes around WHERE: {chunk2}")
    print(f"Hex: {chunk2.hex()}")
