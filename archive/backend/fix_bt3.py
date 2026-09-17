#!/usr/bin/env python3
"""Convert all mock SQL from double-quoted to backtick with proper regex escaping."""
import re, subprocess

with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    data = f.read()

lines = data.split(b'\n')
new_lines = []
converted = 0

for i, raw_line in enumerate(lines):
    line = raw_line.decode('utf-8', errors='replace')
    stripped = line.rstrip()
    
    # Match mock.ExpectQuery("SQL"). or mock.ExpectExec("SQL"). at end of stripped line
    m = re.match(r'^(\s*mock\.(?:ExpectQuery|ExpectExec)\()(".*")(\)\.)$', stripped)
    if m:
        prefix = m.group(1)
        sql_raw = m.group(2)
        suffix = m.group(3)
        
        inner = sql_raw[1:-1]  # strip quotes
        
        # Normalize: reduce 3+ backslashes before $ to 1
        inner = re.sub(r'\\{3,}\$', r'\$', inner)
        # Reduce 2+ backslashes before * to 1
        inner = re.sub(r'\\{2,}\*', r'\*', inner)
        # Reduce 2+ backslashes before ( to 1
        inner = re.sub(r'\\{2,}\(', r'\(', inner)
        # Reduce 2+ backslashes before ' to 1
        inner = re.sub(r'\\{2,}\'', r'\'', inner)
        # Remaining 2+ backslashes -> 1
        inner = re.sub(r'\\{2,}', r'\\', inner)
        
        new_line = f'{prefix}`{inner}`{suffix}'
        # Preserve original indentation/newline
        new_lines.append(new_line + raw_line[len(stripped):])
        converted += 1
    else:
        new_lines.append(raw_line)

result = b'\n'.join(new_lines)

with open("/root/project/backend/handlers/all_features_test.go", "wb") as f:
    f.write(result)

print(f"Converted {converted} lines, wrote {len(result)} bytes")

# Verify
with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    v = f.read()
print(f"Backtick: {v.count(b'mock.ExpectQuery(`')}, Double-quoted: {v.count(b'mock.ExpectQuery(\"')}")

for idx in [13, 26, 122, 124]:
    lv = v.split(b'\n')
    if idx < len(lv):
        print(f"L{idx+1}: {lv[idx].rstrip()[:130]}")

r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True)
print(f"\nBuild: exit={r.returncode}")
if r.returncode != 0:
    print(r.stderr[-400:])

r = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True)
print(f"Test: exit={r.returncode}")
out = r.stdout
passes = out.count('--- PASS:')
fails = out.count('--- FAIL:')
print(f"PASS: {passes}, FAIL: {fails}")
print(out[-600:])
