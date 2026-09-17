#!/usr/bin/env python3
"""Fix double-quoted mock SQL -> backtick raw strings, correctly this time."""
import re

with open("/root/project/backend/handlers/all_features_test.go", "r") as f:
    content = f.read()

lines = content.split('\n')
out = []

for line in lines:
    stripped = line.strip()
    
    if 'mock.ExpectQuery("' not in stripped and 'mock.ExpectExec("' not in stripped:
        out.append(line)
        continue
    
    # Find which method
    if 'mock.ExpectQuery("' in stripped:
        method = 'ExpectQuery'
        prefix = 'mock.ExpectQuery("'
    else:
        method = 'ExpectExec'
        prefix = 'mock.ExpectExec("'
    
    # Find the opening quote position
    qpos = stripped.find(prefix)  # position of 'm' in mock.ExpectQuery("
    open_quote = qpos + len(prefix) - 1  # position of the " character (len-1 because prefix includes the ")
    
    # Sanity check
    if stripped[open_quote] != '"':
        # Try finding " after the prefix
        open_quote = stripped.find('"', qpos)
    
    # Now extract SQL content between quotes, handling escapes
    i = open_quote + 1
    sql_parts = []
    while i < len(stripped):
        c = stripped[i]
        if c == '' and i + 1 < len(stripped):
            sql_parts.append('' + stripped[i+1])
            i += 2
        elif c == '"':
            break
        else:
            sql_parts.append(c)
            i += 1
    
    sql_raw = ''.join(sql_parts)
    after_quote = stripped[i+1:]  # after the closing "
    
    # Fix SQL for backtick:
    # The raw content has Go double-quoted escapes.
    # We need the literal characters for backtick.
    #
    # Key transforms:
    # \\$ in source (2 chars: \, \) -> Go reads as \$ But \$ is invalid!
    # In backtick we want \$ (2 chars: \, $) so RE2 matches literal $
    #
    # \\\\ in source (4 chars: \,\, \,\,) -> Go reads as \\ (2 chars: \,\,)
    # In backtick we want \ (1 char) for regex escapes like \d, \s, \w
    # Wait, RE2 needs \\d to match digit? No, RE2 uses \d directly.
    # In backtick: `\d` -> RE2 sees \d -> matches digit. ✓
    # In double-quoted source: "\\d" -> Go reads as \d. Also works but need backtick for $.
    #
    # So the main fix is for $: in double-quoted, we can't have \$. Must use backtick.
    # For other escapes (\d, \s, \(, etc.), both work.
    #
    # The file currently has invalid escapes. Let me just take the raw bytes
    # and fix only the $ handling.
    
    sql_fixed = sql_raw
    
    # The raw content has sequences like: ... = \\\\\$1 ...
    # (multiple backslashes + $ + digit)
    # For backtick, we want: ... = \$1 ...
    # (one backslash + $ + digit)
    
    # Replace: any number of backslashes followed by $ + (digit or _ or letter)
    # With: exactly one backslash + $ + (digit or _ or letter)
    sql_fixed = re.sub(r'(\\*)(\$)([^\s"\)])', lambda m: '\\' + m.group(2) + m.group(3) if m.group(1) else m.group(2) + m.group(3), sql_fixed)
    # Wait, r'\\' in Python replacement = one literal backslash. 
    # r'\\' + m.group(2) + m.group(3) = \ + $ + digit = \$1 ✓
    # But if m.group(1) is empty (no backslashes), we still want \$
    # So always produce \$
    
    # Simpler: always ensure \$X has exactly one backslash
    sql_fixed = re.sub(r'(\\*?)(\$)(\d+)', r'\\\\\2\3', sql_fixed)
    # r'\\\\\2\3' = \ + $ + digit  ✓
    
    # Also for \( and \):
    sql_fixed = re.sub(r'(\\*?)(\\\()', r'\\\2', sql_fixed)  # Normalize \( 
    sql_fixed = re.sub(r'(\\*?)(\\\))', r'\\\2', sql_fixed)
    
    # Build fixed line
    indent = line[:len(line) - len(line.lstrip())]
    fixed = indent + 'mock.' + method + '(`' + sql_fixed + '`)' + after_quote
    out.append(fixed)

result = '\n'.join(out)

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(result)

print("Done. Checking...")
with open("/root/project/backend/handlers/all_features_test.go", "r") as f:
    vl = f.readlines()
for i in [13, 26, 41, 70, 77, 122, 127]:
    if i < len(vl):
        print(f"L{i+1}: {vl[i].rstrip()[:150]}")
