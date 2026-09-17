#!/usr/bin/env python3
"""Convert mock SQL to backtick with correct escaping."""
import re, subprocess

with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    data = f.read()

text = data.decode('utf-8')
lines = text.split('\n')
new_lines = []
converted = 0

for raw_line in lines:
    stripped = raw_line.rstrip()
    
    m = re.match(r'^(\s*mock\.(?:ExpectQuery|ExpectExec)\()(".*")(\)\.)$', stripped)
    if m:
        prefix = m.group(1)
        sql_raw = m.group(2)
        suffix = m.group(3)
        
        inner = sql_raw[1:-1]  # strip quotes
        
        # Normalize backslash sequences
        inner = inner.replace('\\\\\\\\\\$', '\\$')
        inner = inner.replace('\\\\\\\\$', '\\$')
        inner = inner.replace('\\\\\\$', '\\$')
        inner = inner.replace('\\\\$', '\\$')
        inner = inner.replace('\\\\\\\\*', '\\*')
        inner = inner.replace('\\\\*', '\\*')
        inner = inner.replace('\\\\(', '\\(')
        inner = inner.replace('\\\\\\\\', '\\')
        inner = inner.replace('\\\\', '\\')
        
        bt = '`' + inner + '`'
        new_line = prefix + bt + suffix
        new_lines.append(new_line)
        converted += 1
    else:
        new_lines.append(raw_line)

result = '\n'.join(new_lines)

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(result)

print("Converted %d lines" % converted)

with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    v = f.read()
print("Backtick: %d, Double-quoted: %d" % (v.count(b'mock.ExpectQuery(`'), v.count(b'mock.ExpectQuery("')))

for idx in [13, 26, 122, 124]:
    lv = v.split(b'\n')
    if idx < len(lv):
        print("L%d: %s" % (idx+1, lv[idx].rstrip()[:150]))

r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True)
print("\nBuild: exit=%d" % r.returncode)
if r.returncode != 0:
    print(r.stderr[-400:])

r = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True)
print("Test: exit=%d" % r.returncode)
out = r.stdout
print("PASS: %d, FAIL: %d" % (out.count('--- PASS:'), out.count('--- FAIL:')))
print(out[-600:])
