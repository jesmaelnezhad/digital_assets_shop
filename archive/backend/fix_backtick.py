#!/usr/bin/env python3
"""Convert double-quoted mock SQL strings to backtick raw strings."""
import re

SRC = "/root/project/backend/handlers/all_features_test.go"

with open(SRC, "rb") as f:
    raw = f.read()

text = raw.decode("utf-8")
lines = text.split("\n")
fixed_lines = []

for line in lines:
    if 'mock.ExpectQuery("' in line or 'mock.ExpectExec("' in line:
        line = line.replace('mock.ExpectQuery("', 'mock.ExpectQuery(`', 1)
        line = line.replace('mock.ExpectExec("', 'mock.ExpectExec(`', 1)
        line = re.sub(r'"(\)|\s*\.\s*W)', r'`\1', line)
        line = re.sub(r'"(\s*)$', r'`\1', line)
    
    fixed_lines.append(line)

result = "\n".join(fixed_lines)

# Fix escaping inside backtick strings: reduce multiple backslashes before $N to single
result_lines = result.split("\n")
final_lines = []

for line in result_lines:
    if '`' in line and ('mock.ExpectQuery(`' in line or 'mock.ExpectExec(`' in line):
        first_tick = line.find('`')
        last_tick = line.rfind('`')
        if first_tick >= 0 and last_tick > first_tick:
            prefix = line[:first_tick+1]
            content = line[first_tick+1:last_tick]
            suffix = line[last_tick:]
            # Reduce any sequence of backslashes before $N to a single backslash
            content = re.sub(r'(\\*)(\$(\d+))', lambda m: '\\' + m.group(2) if m.group(1) else m.group(2), content)
            line = prefix + content + suffix
    
    final_lines.append(line)

final_result = "\n".join(final_lines)

with open(SRC, "w") as f:
    f.write(final_result)

print("Done")
with open(SRC, "r") as f:
    vl = f.readlines()
for ln in [14, 27, 42, 71, 75, 78, 80, 123, 128]:
    if ln <= len(vl):
        print(f"L{ln}: {vl[ln-1].rstrip()[:130]}")
