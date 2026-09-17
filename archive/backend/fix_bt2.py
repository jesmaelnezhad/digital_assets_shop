#!/usr/bin/env python3
"""Convert mock SQL from double-quoted to backtick with proper escaping."""
import re, subprocess

with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    data = f.read()

lines = data.split(b'\n')
new_lines = []
converted = 0

for i, line in enumerate(lines):
    line_str = line.decode('utf-8', errors='replace')
    
    # Match: mock.ExpectQuery("SQL").  or  mock.ExpectExec("SQL").
    # The line ends with ").  (quote, paren, dot)
    m = re.match(r'^(\s*mock\.(?:ExpectQuery|ExpectExec)\()(".*")(\)\.\s*)$', line_str)
    if m:
        prefix = m.group(1)
        sql_raw = m.group(2)  # "SQL..."
        suffix = m.group(3)   # ).
        
        inner = sql_raw[1:-1]  # strip " "
        
        # Normalize Go double-quoted escapes to backtick regex patterns:
        # \\\\\$1 -> \$1 (reduce 3+ backslashes before $ to 1)
        # \\\\* -> \* (reduce 2+ backslashes before * to 1)
        # \\\\( -> \( (reduce 2+ backslashes before ( to 1)
        # \\\\' -> \' (reduce 2+ backslashes before ' to 1)
        # \\\\ -> \ (literal backslash -> single backslash in regex)
        
        # Reduce 3+ backslashes before $ to 1
        inner = re.sub(r'\\{3,}\$', r'\$', inner)
        # Reduce 2+ backslashes before * to 1  
        inner = re.sub(r'\\{2,}\*', r'\*', inner)
        # Reduce 2+ backslashes before ( to 1
        inner = re.sub(r'\\{2,}\(', r'\(', inner)
        # Reduce 2+ backslashes before ' to 1
        inner = re.sub(r'\\{2,}\'', r'\'', inner)
        # Reduce remaining 2+ backslashes to 1
        inner = re.sub(r'\\{2,}', r'\\', inner)
        # Single backslash before $ -> keep as \$ (regex escaped dollar)
        # Single backslash before * -> keep as \* (regex escaped star)
        # These are already correct after the above reductions
        
        new_line = f'{prefix}`{inner}`{suffix}'
        new_lines.append(new_line.encode('utf-8'))
        converted += 1
    else:
        new_lines.append(line)

result = b'\n'.join(new_lines)

with open("/root/project/backend/handlers/all_features_test.go", "wb") as f:
    f.write(result)

print(f"Converted {converted} lines, wrote {len(result)} bytes")

# Verify
with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    verify = f.read()
bt = verify.count(b'mock.ExpectQuery(`')
dq = verify.count(b'mock.ExpectQuery("')
print(f"Backtick: {bt}, Double-quoted: {dq}")

# Show a few key lines
lines_v = verify.split(b'\n')
for idx in [13, 26, 122, 124]:
    if idx < len(lines_v):
        print(f"L{idx+1}: {lines_v[idx].rstrip()[:120]}")

# Build and test
r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True)
print(f"\nBuild: exit={r.returncode}")
if r.returncode != 0:
    print(r.stderr[-400:])

r = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True)
print(f"Test exit: {r.returncode}")
print(r.stdout[-500:])
