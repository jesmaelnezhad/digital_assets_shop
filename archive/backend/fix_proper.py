#!/usr/bin/env python3
"""Fix double-quoted mock SQL -> backtick, correctly handle escapes and extra parens."""
import re

with open("/root/project/backend/handlers/all_features_test.go.bak", "r") as f:
    text = f.read()

lines = text.split('\n')
out_lines = []

for line in lines:
    found = False
    for method_prefix in ['mock.ExpectQuery("', 'mock.ExpectExec("']:
        pos = line.find(method_prefix)
        if pos >= 0:
            found = True
            method = method_prefix.replace('("', '')  # ExpectQuery or ExpectExec
            q_open = pos + len(method_prefix) - 1  # position of the opening "
            
            # Extract SQL content between quotes
            i = q_open + 1
            chars = []
            while i < len(line):
                c = line[i]
                if c == '\\' and i + 1 < len(line):
                    chars.append(line[i:i+2])  # keep escape sequence
                    i += 2
                elif c == '"':
                    i += 1  # skip closing "
                    break
                else:
                    chars.append(c)
                    i += 1
            
            sql_raw = ''.join(chars)
            after = line[i:]  # text after closing "
            
            # Fix: collapse multiple backslashes before $, (, ), . to single backslash
            def fix_bs(m):
                bs = m.group(1)
                ch = m.group(2)
                if bs:
                    return '' + ch  # one backslash + char
                return ch
            
            sql_fixed = re.sub(r'(\\*)(\$|[().])', fix_bs, sql_raw)
            
            # Build result line
            indent = line[:len(line) - len(line.lstrip())]
            result = indent + 'mock.' + method + '(`' + sql_fixed + '`)' + after
            
            # Remove extra closing paren if present
            result = result.replace('`))', '`)')
            
            out_lines.append(result)
            break
    
    if not found:
        out_lines.append(line)

result = '\n'.join(out_lines)

with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(result)

print("Done. Verifying:")
with open("/root/project/backend/handlers/all_features_test.go", "r") as f:
    vl = f.readlines()
for idx in [13, 14, 26, 27, 77, 122, 127]:
    if idx < len(vl):
        print(f"L{idx+1}: {vl[idx].rstrip()[:140]}")
