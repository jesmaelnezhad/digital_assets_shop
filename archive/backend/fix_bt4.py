#!/usr/bin/env python3
"""Convert mock SQL from double-quoted to backtick."""
import re, subprocess

with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    data = f.read()

lines = data.split(b'\n')
new_lines = []
converted = 0

for i, raw_line in enumerate(lines):
    line = raw_line.decode('utf-8', errors='replace')
    stripped = line.rstrip()
    
    m = re.match(r'^(\s*mock\.(?:ExpectQuery|ExpectExec)\()(".*")(\)\.)$', stripped)
    if m:
        prefix = m.group(1)
        sql_raw = m.group(2)
        suffix = m.group(3)
        
        inner = sql_raw[1:-1]  # strip quotes
        
        # Normalize escaping: reduce multiple backslashes to single before special chars
        inner = re.sub(r'\\{3,}\$', r'\$', inner)
        inner = re.sub(r'\\{2,}\*', r'\*', inner)
        inner = re.sub(r'\\{2,}\(', r'\(', inner)
        inner = re.sub(r'\\{2,}\'', r'\'', inner)
        inner = re.sub(r'\\{2,}', r'\\', inner)
        
        bt = '`' + inner + '`'
        new_line = prefix + bt + suffix
        rest = raw_line[len(stripped):]
        new_lines.append((new_line + rest).encode('utf-8'))
        converted += 1
    else:
        new_lines.append(raw_line)

result = b'\n'.join(new_lines)

with open("/root/project/backend/handlers/all_features_test.go", "wb") as f:
    f.write(result)

print("Converted %d lines, wrote %d bytes" % (converted, len(result)))

with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    v = f.read()
print("Backtick: %d, Double-quoted: %d" % (v.count(b'mock.ExpectQuery(`'), v.count(b'mock.ExpectQuery("')))

lv = v.split(b'\n')
for idx in [13, 26, 122, 124]:
    if idx < len(lv):
        print("L%d: %s" % (idx+1, lv[idx].rstrip()[:130]))

r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True)
print("\nBuild: exit=%d" % r.returncode)
if r.returncode != 0:
    print(r.stderr[-400:])

r = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True)
print("Test: exit=%d" % r.returncode)
out = r.stdout
print("PASS: %d, FAIL: %d" % (out.count('--- PASS:'), out.count('--- FAIL:')))
print(out[-600:])
