#!/usr/bin/env python3
"""Clean conversion of double-quoted mock SQL to backtick raw strings."""
with open("/root/project/backend/handlers/all_features_test.go.bak", "rb") as f:
    data = f.read()

text = data.decode("utf-8")
out_lines = []

for line in text.split('\n'):
    s = line
    
    # Check for mock calls
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
    
    # Opening " position
    q_start = start + len(marker) - 1  # position of the opening "
    
    # Collect SQL content between quotes
    i = q_start + 1
    parts = []
    while i < len(s):
        c = s[i]
        if c == '\\' and i + 1 < len(s):
            nxt = s[i+1]
            # Handle escape sequences
            if nxt == '$' and i + 2 < len(s) and (s[i+2].isdigit() or s[i+2] == '_'):
                # \$X -> keep as \$X
                parts.append('\\$' + s[i+2])
                i += 3
                continue
            elif nxt == '\\':
                # \\ -> single \
                parts.append('\\')
                i += 2
                continue
            elif nxt in '()."':
                parts.append('\\' + nxt)
                i += 2
                continue
            else:
                parts.append('\\' + nxt)
                i += 2
                continue
        elif c == '"':
            i += 1  # skip closing "
            break
        else:
            parts.append(c)
            i += 1
    
    sql_str = ''.join(parts)
    
    # After the closing ", there might be ). or ).WillReturn...
    # Check for extra )
    after = s[i:]
    # Remove double )) if present
    if after.startswith('))'):
        after = after[1:]  # remove one extra )
    
    indent = line[:len(line) - len(line.lstrip())]
    fixed = indent + f'mock.{method}(`{sql_str}`)' + after
    out_lines.append(fixed)

result = '\n'.join(out_lines)

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(result)

print("Done")
with open("/root/project/backend/handlers/all_features_test.go", "r") as f:
    vl = f.readlines()
for idx in [13, 14, 26, 27, 41, 70, 77, 122, 127]:
    if idx < len(vl):
        print(f"L{idx+1}: {vl[idx].rstrip()[:130]}")
