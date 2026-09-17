#!/usr/bin/env python3
"""Convert mock SQL from double-quoted to backtick - simple approach."""
import re, subprocess

with open("/root/project/backend/handlers/all_features_test.go", "rb") as f:
    data = f.read()

text = data.decode('utf-8')
lines = text.split('\n')
new_lines = []
converted = 0
errors = []

for i, raw_line in enumerate(lines):
    stripped = raw_line.rstrip()
    
    # Simple approach: find " after ( and " before ).
    # Pattern: mock.ExpectX("SQL").
    # Match: prefix = everything up to and including first "
    #        sql = between the two "
    #        suffix = from second " to end
    
    # Find the positions of the quotes
    first_quote = stripped.find('("')
    if first_quote >= 0:
        # Find the second quote (before the closing )\. )
        second_quote = stripped.rfind('")')
        if second_quote > first_quote:
            prefix = stripped[:first_quote+1]  # includes opening "
            sql_content = stripped[first_quote+2:second_quote]  # between quotes
            suffix = stripped[second_quote:]  # from closing " to end
            
            # Normalize escaping: reduce multiple backslashes
            sql_content = sql_content.replace('\\\\\\\\\\$', '\\$')
            sql_content = sql_content.replace('\\\\\\\\$', '\\$')
            sql_content = sql_content.replace('\\\\\\$', '\\$')
            sql_content = sql_content.replace('\\\\$', '\\$')
            sql_content = sql_content.replace('\\\\\\\\*', '\\*')
            sql_content = sql_content.replace('\\\\*', '\\*')
            sql_content = sql_content.replace('\\\\(', '\\(')
            sql_content = sql_content.replace('\\\\\\\\', '\\')
            sql_content = sql_content.replace('\\\\', '\\')
            
            bt = prefix + '`' + sql_content + '`' + suffix[1:]  # remove the " from suffix start
            # Actually suffix starts with ")\., so we need: `sql`.)\.
            # prefix already has the first ", suffix has the second " followed by ).
            # Result: prefix_without_quote + ` + sql + ` + suffix_without_first_char
            new_line = stripped[:first_quote] + '`' + sql_content + '`' + stripped[second_quote+1:]
            new_lines.append(new_line)
            converted += 1
        else:
            new_lines.append(raw_line)
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
