#!/usr/bin/env python3
"""Fix all mock SQL escaping in all_features_test.go."""
import re

SRC = "/root/project/backend/handlers/all_features_test.go"
with open(SRC, "rb") as f:
    raw = f.read()

text = raw.decode("utf-8")
out = []
i = 0
n_fixed = 0

while i < len(text):
    pos_q = text.find('mock.ExpectQuery("', i)
    pos_e = text.find('mock.ExpectExec("', i)
    if pos_q < 0 and pos_e < 0:
        out.append(text[i:])
        break
    
    if pos_q >= 0 and (pos_e < 0 or pos_q <= pos_e):
        pos, sym = pos_q, 'ExpectQuery'
    else:
        pos, sym = pos_e, 'ExpectExec'
    
    out.append(text[i:pos])
    
    oq = text.find('("', pos) + 2
    cq = oq
    while cq < len(text) and text[cq] != '"':
        cq += 1
    
    if cq >= len(text):
        out.append(text[pos:])
        break
    
    sql = text[oq:cq]  # the raw string between quotes
    
    # Transform: collapse excessive backslash sequences
    # Key insight: the file went through multiple fix attempts
    # that added extra backslashes. We need to reduce to correct level.
    #
    # For backtick raw strings, we want: \$ (1 backslash + dollar) to match literal $
    # The source has: \\\\\$ or \\\\\\\$ etc (multiple backslashes + dollar)
    # Fix: collapse all consecutive backslashes before $ to a single \$
    
    result = []
    j = 0
    while j < len(sql):
        if sql[j] == '\\':
            # Count consecutive backslashes
            bs_start = j
            while j < len(sql) and sql[j] == '\\':
                j += 1
            nbs = j - bs_start
            
            if j < len(sql):
                nc = sql[j]
                if nc == '$':
                    # Collapse to \$ (for raw string: matches literal $)
                    result.append('\\')
                    result.append('$')
                    j += 1
                    n_fixed += 1
                elif nc == '*':
                    # Keep \* (regex for literal *)
                    result.append('\\')
                    result.append('*')
                    j += 1
                elif nc in '()':
                    result.append('\\')
                    result.append(nc)
                    j += 1
                elif nc == "'":
                    j += 1  # skip the backslash
                else:
                    # Unknown: keep first backslash
                    result.append('\\')
            else:
                result.append('\\')
        else:
            result.append(sql[j])
            j += 1
    
    sql_final = ''.join(result)
    replacement = f'mock.{sym}(`{sql_final}`)'
    out.append(replacement)
    i = cq + 1
    n_fixed += 1

result_text = ''.join(out)
print(f"Fixed {n_fixed} mock strings")

with open(SRC, "w") as f:
    f.write(result_text)

# Verify
import subprocess
r = subprocess.run(['go', 'build', './handlers/'], capture_output=True, text=True, cwd='/root/project/backend')
print(f"Build: {'OK' if r.returncode == 0 else 'FAIL'}")
if r.returncode:
    print(r.stderr[:500])
