#!/usr/bin/env python3
"""
Properly convert all mock SQL from double-quoted to backtick,
with correct regex escaping.
"""
import re

with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    data = f.read()

lines = data.split(b'\n')
new_lines = []

for i, line in enumerate(lines):
    line_str = line.decode('utf-8', errors='replace')
    
    # Match: mock.ExpectQuery("SQL"). or mock.ExpectExec("SQL").
    m = re.match(r'^(\s*mock\.(?:ExpectQuery|ExpectExec)\()(".*")(\.\n?)$', line_str)
    if m:
        prefix = m.group(1)
        sql_raw = m.group(2)  # "SQL..." with quotes
        suffix = m.group(3)
        
        # Get content between quotes
        inner = sql_raw[1:-1]  # strip " "
        
        # The inner SQL has Go double-quoted escape sequences
        # Need to: 1) interpret Go escapes to get actual SQL text
        # 2) add regex escaping for backtick
        
        # Go double-quoted escapes in this file:
        # \\\\\$1 -> actual: \$1 (but this is already a regex-ish pattern)
        # Actually the file has raw SQL with Go escaping applied
        
        # Simplest approach: the inner text, when Go-interpreted, gives us
        # the regex pattern we need. But since these are illegal Go escapes,
        # let me just normalize: reduce multiple backslashes to single before
        # $ * ( ) ' and use those as the regex pattern in backtick
        
        # Normalize backslashes:
        # \\\\\$1 -> \$1
        # \\\\* -> \*
        # \\\\( -> \(
        # \\\\' -> \'
        # \\\\ -> \\ (two backslashes -> literal backslash in regex)
        
        # Use regex to reduce 2+ backslashes before special chars to 1
        inner = re.sub(r'\\{2,}(\$|\*|\(|\'|\.)', r'\\\1', inner)
        # Handle remaining \\\\ (double backslash not before special) -> single \
        inner = re.sub(r'\\{2,}', r'\\', inner)
        
        new_lines.append(f'{prefix}`{inner}`{suffix}'.encode('utf-8'))
    else:
        new_lines.append(line)

result = b'\n'.join(new_lines)

with open("/root/project/backend/handlers/all_features_test.go", "wb") as f:
    f.write(result)

print(f"Converted {len(lines)} lines, wrote {len(result)} bytes")

# Verify
with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    verify = f.read()
backtick_count = verify.count(b'mock.ExpectQuery(`')
double_count = verify.count(b'mock.ExpectQuery("')
print(f"Backtick: {backtick_count}, Double-quoted: {double_count}")

# Show line 27
lines_v = verify.split(b'\n')
if len(lines_v) > 26:
    print(f"\nL27: {lines_v[26][:120]}")
    print(f"L27 hex: {lines_v[26][:80].hex()}")

import subprocess
r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True)
print(f"\nBuild: exit={r.returncode}")
if r.returncode != 0:
    print(r.stderr[-400:])
PYEOF
