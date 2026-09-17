#!/usr/bin/env python3
"""Simple mock SQL to backtick converter."""
import re, subprocess

with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    data = f.read()

text = data.decode('utf-8')
lines = text.split('\n')
new_lines = []

for line in lines:
    stripped = line.rstrip()
    # Find mock.ExpectQuery("..."). or mock.ExpectExec("...")
    # Use a simple state machine: find ( then " then " then )
    idx_open_paren = stripped.find('mock.Expect')
    if idx_open_paren < 0:
        new_lines.append(line)
        continue
    
    # Find the opening quote after (
    idx_first_q = stripped.find('"', idx_open_paren)
    if idx_first_q < 0:
        new_lines.append(line)
        continue
    
    # Find closing quote before ).
    idx_second_q = stripped.rfind('"', idx_first_q + 1)
    if idx_second_q < 0 or idx_second_q >= len(stripped) - 2:
        new_lines.append(line)
        continue
    
    # Extract parts
    before = stripped[:idx_first_q]  # everything before first "
    sql = stripped[idx_first_q+1:idx_second_q]  # between quotes
    after = stripped[idx_second_q+1:]  # after closing "
    
    # Normalize escaping in SQL
    sql = sql.replace('\\\\\\\\\\$', '\\$')  # 6bs -> 1bs
    sql = sql.replace('\\\\\\\\$', '\\$')    # 4bs -> 1bs  
    sql = sql.replace('\\\\\\$', '\\$')       # 3bs -> 1bs
    sql = sql.replace('\\\\$', '\\$')          # 2bs -> 1bs
    sql = sql.replace('\\\\\\\\*', '\\*')     # 4bs -> 1bs
    sql = sql.replace('\\\\*', '\\*')          # 2bs -> 1bs
    sql = sql.replace('\\\\(', '\\(')          # 2bs -> 1bs
    sql = sql.replace('\\\\\\\\', '\\')        # 4bs -> 1bs
    sql = sql.replace('\\\\', '\\')             # 2bs -> 1bs
    
    # Build backtick version
    new_line = before + '`' + sql + '`' + after
    new_lines.append(new_line)

result = '\n'.join(new_lines)

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(result)

print("Done. %d lines processed." % len(lines))

with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    v = f.read()
print("Backtick: %d, Double-quoted: %d" % (v.count(b'mock.ExpectQuery(`'), v.count(b'mock.ExpectQuery("')))

for idx in [13, 26, 122, 124]:
    lv = v.split(b'\n')
    if idx < len(lv):
        print("L%d: %s" % (idx+1, lv[idx].rstrip().decode('utf-8', errors='replace')[:200]))

r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True)
print("\nBuild: exit=%d" % r.returncode)
if r.returncode != 0:
    print(r.stderr[-400:])

r = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True)
print("Test: exit=%d" % r.returncode)
print(r.stdout[-600:])
