#!/usr/bin/env python3
"""Fix double-quoted mock strings -> backtick, correctly."""
with open("/root/project/backend/handlers/all_features_test.go.bak", "rb") as f:
    data = f.read()

text = data.decode("utf-8")
out_lines = []

for line in text.split('\n'):
    s = line
    idx_q = s.find('mock.ExpectQuery("')
    idx_e = s.find('mock.ExpectExec("')
    if idx_q < 0 and idx_e < 0:
        out_lines.append(s)
        continue
    
    if idx_q >= 0 and (idx_e < 0 or idx_q <= idx_e):
        method = "ExpectQuery"
        marker = 'mock.ExpectQuery("'
        start = idx_q
    else:
        method = "ExpectExec"
        marker = 'mock.ExpectExec("'
        start = idx_e
    
    q_start = start + len(marker) - 1
    i = q_start + 1
    parts = []
    while i < len(s):
        c = s[i]
        if c == '\\' and i + 1 < len(s):
            nxt = s[i+1]
            if nxt == '\\':
                parts.append('\\')
                i += 2
            elif nxt == '$':
                parts.append('\\$')
                i += 2
            elif nxt in '()."':
                parts.append('\\' + nxt)
                i += 2
            else:
                parts.append(c)
                i += 1
        elif c == '"':
            i += 1
            break
        else:
            parts.append(c)
            i += 1
    
    sql_str = ''.join(parts)
    after = s[i:]
    if after.startswith('))'):
        after = after[1:]
    
    indent = line[:len(line) - len(line.lstrip())]
    fixed = indent + 'mock.' + method + '(`' + sql_str + '`)' + after
    out_lines.append(fixed)

result = '\n'.join(out_lines)

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(result)

print("Done")
with open("/root/project/backend/handlers/all_features_test.go", "r") as f:
    vl = f.readlines()
for idx in [13, 14, 26, 27, 77, 122, 127]:
    if idx < len(vl):
        print(f"L{idx+1}: {vl[idx].rstrip()[:130]}")
