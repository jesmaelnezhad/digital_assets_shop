#!/usr/bin/env python3
"""
Comprehensive fix script for all_features_test.go.
Operation: Read from backup, apply all fixes at byte/line level, write result.
This script does NOT rely on previous fix scripts - it does everything from scratch.
"""
import re
import sys

BACKKUP = "handlers/all_features_test.go.bak"
OUTPUT = "handlers/all_features_test.go"

with open(BACKKUP, "r") as f:
    content = f.read()

lines = content.split('\n')
fixed_lines = []

for i, line in enumerate(lines):
    # Fix 1: Convert double-quoted mock SQL to backtick raw strings
    # Pattern: mock.ExpectQuery("SQL")  or  mock.ExpectExec("SQL")
    # We need to extract the SQL, fix escaping, and wrap in backticks
    
    # Match: mock.ExpectQuery("...") or mock.ExpectExec("...")
    m = re.match(r'^(\s*mock\.(ExpectQuery|ExpectExec)\("(.+?)"\)\.?)\s*$', line)
    if m:
        prefix = m.group(1)
        method = m.group(2)
        sql = m.group(3)
        
        # Fix escaping: \\\$1 -> \$1, \\\\$1 -> \$1, etc.
        # In the backup file, the SQL inside double quotes has:
        # - \$1 written as \\$1 (backslash-backslash-dollar-1) = 2 chars: \ and $1
        # - We want: \$1 (backslash-dollar-1) = 2 chars: \ and $1
        # In the file bytes: \\\$ in double-quoted string = 3 chars, we want \$ = 2 chars
        # Replacing all occurrences of backslash-backslash-dollar with backslash-dollar
        sql = sql.replace('\\\\$', '$')
        sql = sql.replace('\\\\\\$', '$')
        
        # Fix COUNT(\\\\\\*) -> COUNT(\\\\*)  (in the file this is COUNT(\\*))
        # The backup has: COUNT(\\\\\\*)  (which is double-quoted: COUNT(\\\\*))
        # We want in backtick: COUNT(\\*)  (literal: backslash star)
        sql = sql.replace('COUNT(\\\\\\\\*)', 'COUNT(\\\\*)')
        
        # Wrap in backtick
        new_line = f'{line[:m.start(1)]}`{sql}`.'
        fixed_lines.append(new_line)
        continue
    
    # Fix 2: Remove extra closing parens on mock lines that end with )).  
    # These occur when a line originally had )). at the end after conversion
    if line.rstrip().endswith(')).') and 'mock.' in line:
        line = line.rstrip()
        # Remove trailing )).  
        if line.endswith(')).'):
            line = line[:-3] + '.'
        fixed_lines.append(line)
        continue
    
    fixed_lines.append(line)

result = '\n'.join(fixed_lines)

# Now do byte-level fixes on the result
result_bytes = result.encode('utf-8')

# Fix COUNT(\\* -) COUNT(\\*) at byte level  
# In the file after v5 conversion, COUNT(\\*) has bytes: 5c 5c 2a (two backslashes + star)
# We want: 5c 2a (one backslash + star)
old_bytes = bytes([0x5c, 0x5c, 0x2a])  # \\*
new_bytes = bytes([0x5c, 0x2a])        # \*
count = result_bytes.count(old_bytes)
result_bytes = result_bytes.replace(old_bytes, new_bytes)
print(f"Byte-level: fixed {count} COUNT patterns")

# Also fix any \\$ patterns that survived
old_dollar = bytes([0x5c, 0x5c, 0x24])  # \\$
new_dollar = bytes([0x5c, 0x24])        # \$
count_d = result_bytes.count(old_dollar)
result_bytes = result_bytes.count(old_dollar)  # just count, don't replace here
# Actually replace
result_bytes = result_bytes.replace(old_dollar, new_dollar)
print(f"Byte-level: fixed {count_d} dollar patterns")

with open(OUTPUT, "wb") as f:
    f.write(result_bytes)

print(f"Wrote {len(result_bytes)} bytes to {OUTPUT}")
