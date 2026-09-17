#!/usr/bin/env python3
"""Properly convert mock SQL from double-quoted to backtick."""
import re, subprocess

with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    data = f.read()

text = data.decode('utf-8')
lines = text.split('\n')
new_lines = []
converted = 0

for raw_line in lines:
    stripped = raw_line.rstrip()
    
    # Match: mock.ExpectQuery("SQL").  — the SQL is between the first " after ( and the last " before ).
    # Pattern: (prefix")(SQL)")(suffix)
    # The key is that there are exactly two double-quotes: one after ( and one before ).
    m = re.match(r'^(\s*mock\.(?:ExpectQuery|ExpectExec)\()(")(.*)(")\)\.)$', stripped)
    if m:
        prefix = m.group(1)  # mock.ExpectQuery(
        inner_quote = m.group(2)  # opening "
        sql_content = m.group(3)  # the SQL between quotes
        end_quote = m.group(4)  # closing "
        suffix = m.group(5)  # ).
        
        # Normalize escaping
        sql_content = sql_content.replace('\\\\\\\\\\$', '\\$')
        sql_content = sql_content.replace('\\\\\\\\$', '\\$')
        sql_content = sql_content.replace('\\\\\\$', '\\$')
        sql_content = sql_content.replace('\\\\$', '\\$')
        sql_content = sql_content.replace('\\\\\\\\*', '\\*')
        sql_content = sql_content.replace('\\\\*', '\\*')
        sql_content = sql_content.replace('\\\\(', '\\(')
        sql_content = sql_content.replace('\\\\\\\\', '\\')
        sql_content = sql_content.replace('\\\\', '\\')
        
        bt = prefix + '`' + sql_content + '`' + suffix
        new_lines.append(bt)
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

# Show key lines
for idx in [13, 26, 122, 124]:
    lv = v.split(b'\n')
    if idx < len(lv):
        s = lv[idx].rstrip().decode('utf-8', errors='replace')
        print("L%d: %s" % (idx+1, s[:200]))

r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True)
print("\nBuild: exit=%d" % r.returncode)
if r.returncode != 0:
    print(r.stderr[-400:])

r = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True)
print("Test: exit=%d" % r.returncode)
out = r.stdout
print("PASS: %d, FAIL: %d" % (out.count('--- PASS:'), out.count('--- FAIL:')))
if out.count('--- FAIL:') > 0:
    print(out[out.find('--- FAIL:'):out.find('--- FAIL:')+600])
