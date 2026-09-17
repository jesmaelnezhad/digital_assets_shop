#!/usr/bin/env python3
"""Fix all_features_test.go: convert double-quoted to backtick, fix escapes, fix SQL."""
SRC = "/root/project/backend/handlers/all_features_test.go"
with open(SRC, "r") as f:
    text = f.read()

out = []
i = 0
n_fixed = 0
while i < len(text):
    # Find mock.ExpectQuery( or mock.ExpectExec(
    idx_q = text.find('mock.ExpectQuery("', i)
    idx_e = text.find('mock.ExpectExec("', i)
    if idx_q < 0 and idx_e < 0:
        out.append(text[i:])
        break
    
    if idx_q >= 0 and (idx_e < 0 or idx_q <= idx_e):
        idx = idx_q
        sym = "ExpectQuery"
    else:
        idx = idx_e
        sym = "ExpectExec"
    
    out.append(text[i:idx])
    
    # Find the content between the double quotes
    q1 = text.find('("', idx) + 2  # after opening "
    q2 = text.find('")', q1)  # before closing ") or just "
    if q2 < 0:
        q2 = text.find('"', q1 + 1)
    
    if q2 < 0:
        # Fallback: find next " 
        j = q1
        while j < len(text) and text[j] != '"':
            j += 1
        q2 = j
    
    sql = text[q1:q2]
    
    # Convert to backtick raw string
    # In double-quoted Go string on disk: "\\\\" = \ in value, "\\$" = invalid
    # We want raw string to contain: \$ for matching $1, etc.
    # sql is the raw on-disk text between the double quotes
    # "SELECT ... \\\\\$1" on disk -> sql = "SELECT ... \\\\\$1"  
    # We want raw string: "SELECT ... \$1"
    # So: replace "\\\\" (4 chars on disk: \ \ \ \) with "\" (1 char: \)
    # Then: "\\\$" (3 chars: \ \ $) -> "\$" (2 chars: \ $)
    # But wait, after replacing \\\\ with \, we get "\$" which is what we want!
    
    raw_sql = sql
    # The on-disk text has various escape sequences
    # Replace \\\\ with \ (handles \\\\ -> \)
    raw_sql = raw_sql.replace("\\\\\\\\", "\\")
    # Replace \\$ with \$ ... but \\\$ is now \$ after the above replace
    # Actually: \\\\ + $ = \\\$. Replace \\\\$ with \$:
    raw_sql = raw_sql.replace("\\\\\\$", "\\$")  # \\\\$ -> \$
    # Also handle \\$ directly (in case it wasn't part of \\\\$):
    raw_sql = raw_sql.replace("\\$", "$")  # \$ -> $ (raw string, \$ means literal $)
    # Hmm, this is wrong. In raw string, \$ is literal \$. sqlmock sees \$ and regex interprets as literal $
    # So we WANT \$ in raw string. But \\ in raw string is literal \\.
    # Let me think again...
    # On disk: "SELECT ... \\\\\$1" -> Go double-quoted interprets as "SELECT ... \$1" -> INVALID GO
    # On disk: "SELECT ... \\\$1" -> Go double-quoted interprets as "SELECT ... \$1" -> INVALID GO  
    # 
    # We're converting to raw string. In raw string, every char is literal.
    # We want sqlmock to receive: SELECT ... \$1  (so regex matches literal $)
    # So raw string should contain: SELECT ... \$1
    # On disk, the raw string should be: `SELECT ... \$1`
    # 
    # The sql variable contains the raw on-disk bytes between the double quotes.
    # If on-disk double-quoted string is: "SELECT ... \\\\\$1"
    # sql = 'SELECT ... \\\\\$1' (raw bytes on disk)
    # We want raw_sql = 'SELECT ... \$1'
    # So: replace \\\\$ with \$ (4 chars -> 2 chars)
    # And: replace \\$ with \$ (3 chars -> 2 chars)... but \\$ on disk in raw string is already \$? No.
    # \\$ on disk = 3 chars: \, \, $. In raw string, this is literal \, \, $. sqlmock sees \\$, regex: literal \ then end-of-line. WRONG.
    # We want: \$ on disk = 2 chars: \, $. In raw string, literal \, $. sqlmock sees \$, regex: literal $. CORRECT.
    # 
    # So: replace \\$ (3 bytes on disk) with \$ (2 bytes on disk) inside raw strings.
    # And: replace \\\\$ (5 bytes? no, \\\\ = \, so \\\\$ = 3 bytes: \, \, $) with \$ (2 bytes)
    # 
    # Actually the on-disk bytes for the double-quoted string:
    # "SELECT ... \\\\\$1" -> the bytes between the double quotes are: SELECT ... \\\\\$1
    # That's: ... \, \, \, $, 1  (4 chars before the $: \, \, \, $)
    # No wait. Let me count the ASCII characters in \\\\\$:
    # \\ = 2 chars: backslash, backslash
    # \\\\ = 4 chars: \, \, \, \
    # \\\\$ = 5 chars: \, \, \, \, $
    # \\\\\$ = 6 chars: \, \, \, \, \, $
    # That doesn't seem right either.
    # 
    # OK let me just look at the actual bytes. The backup file line 27 has:
    # b'...WHERE email = \\\\\\$1`))'  (from the hex dump earlier)
    # Let me count: \\\\\$ = \ \ \ \ $ = 5 bytes? No.
    # In Python repr: '\\\\\\$' = \\\\ + \$ = 5 + 2 = 7? No.
    # repr shows escape sequences. \\\\ in repr = 2 actual backslashes. \\\\\$ in repr = 3 actual backslashes + dollar? No.
    # repr("\\\\\\$") = '\\\\\\$' which represents the string: \\\$ (3 chars: \, \, $)
    # Wait no. repr shows Python string escapes. In Python, \\\\ represents 2 chars: \\.
    # So repr containing \\\\\\$ means the actual string has: \\\\ + \$ = 4 + 2 = 6 chars? No.
    # 
    # Let me just check the raw bytes directly.
    pass
    
    replacement = f"mock.{sym}(`{sql}`)"
    out.append(replacement)
    i = q2 + 1  # after closing "
    n_fixed += 1

final = "".join(out)

# Now fix escapes inside backtick raw strings
# Inside raw strings: \\$ should become \$ (remove one backslash)
# Walk through character by character
result = []
i = 0
depth = 0  # >0 means inside backticks
while i < len(final):
    if final[i] == '`':
        depth += 1
        if depth % 2 == 1:
            result.append('`')
        else:
            result.append('`')
        i += 1
    elif depth % 2 == 1:
        # Inside raw string - fix \\$ -> \$
        if i + 2 < len(final) and final[i] == '\\' and final[i+1] == '\\' and final[i+2] == '$':
            result.append('\\')
            result.append('$')
            i += 3
        elif i + 1 < len(final) and final[i] == '\\' and final[i+1] == '*':
            result.append('*')
            i += 2
        else:
            result.append(final[i])
            i += 1
    else:
        result.append(final[i])
        i += 1

final2 = "".join(result)

with open(SRC, "w") as f:
    f.write(final2)

print(f"Fixed {n_fixed} mock strings, wrote {len(final2)} bytes")

# Verify
import subprocess
r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True, cwd="/root/project/backend")
print(f"Build: exit={r.returncode}")
if r.returncode:
    print(r.stderr[:500])
else:
    r2 = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True, cwd="/root/project/backend", timeout=60)
    out_lines = r2.stdout.split('\n')
    for l in out_lines:
        if 'FAIL' in l or 'PASS' in l or 'Error' in l or 'unfulfilled' in l:
            print(l)
    passes = sum(1 for l in out_lines if '--- PASS:' in l)
    fails = sum(1 for l in out_lines if '--- FAIL:' in l)
    print(f"\nResult: {passes} PASS, {fails} FAIL, {r2.returncode}")
