#!/usr/bin/env python3
"""
Fix all_features_test.go: convert double-quoted mock SQL to backtick raw strings.
The backup uses invalid Go escape sequences (\\\$ in double-quoted strings).
"""
import re, subprocess, shutil

SRC = "/root/project/backend/handlers/all_features_test.go"

with open(SRC) as f:
    text = f.read()

# ============================================================
# Convert mock.Expect*("SQL...") to mock.Expect*(`SQL...`)
# ============================================================
out = []
i = 0
n_conv = 0
while i < len(text):
    # Find next mock expectation
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
    
    # Find content between double quotes
    oq = text.find('("', pos) + 2  # right after opening "
    cq = text.find('")', oq)  # closing ")
    if cq < 0:
        # Fallback: find next "
        cq = text.find('"', oq)
    
    if cq < 0:
        out.append(text[pos:])
        break
    
    sql_raw = text[oq:cq]  # raw on-disk bytes between quotes
    
    # Convert to raw string format.
    # The on-disk text has Go double-quoted escape sequences.
    # We need to produce the raw string that would give the same SQL value.
    #
    # In Go double-quoted strings:
    #   \\\\ -> \ (two backslashes in source = one backslash in value)
    #   \\\$ -> \$ (but \$ is invalid in Go 1.22!)
    #   \\*  -> \* (invalid too)
    #   \(   -> (  (valid: \( -> ()
    #   \)   -> )  (valid)
    #
    # For raw strings, we want the literal characters that produce the correct SQL.
    # The SQL the handler sends has: $1, *, (, ) etc. (plain, no backslashes)
    # sqlmock regex needs: \$1 to match literal $, \* to match literal *, etc.
    # In raw string: \$ is literal \$, which sqlmock regex reads as "match literal $".
    # So raw string should contain: \$1, \*, \(, \) etc.
    #
    # The sql_raw on disk for line 27 contains: ...WHERE email = \\\\\$1
    # That's: ...= \, \, \, \, $, 1 (the bytes on disk between the double quotes)
    # Wait, let me count from cat -A: \\\\\\$ = 6 characters on screen
    # But each \ on screen = 1 byte on disk. So 6 bytes: \, \, \, \, \, $
    # That doesn't match what I expect. Let me just run a quick check.
    
    # Actually, let me just transform empirically:
    # Replace runs of backslashes followed by special chars:
    #   N*\ + $ -> \ + $  (collapse to \$)
    #   N*\ + * -> \ + *  (collapse to \*)
    #   N*\ + ( -> \ + (  (collapse to \()
    #   N*\ + ) -> \ + )  (collapse to \))
    #   N*\ + ' -> '       (remove backslashes before quote)
    #   Other backslashes: keep one
    
    processed = []
    j = 0
    while j < len(sql_raw):
        if sql_raw[j] == '\\':
            start = j
            while j < len(sql_raw) and sql_raw[j] == '\\':
                j += 1
            nbs = j - start
            if j < len(sql_raw):
                nc = sql_raw[j]
                if nc == '$':
                    processed.append('\\')
                    processed.append('$')
                    j += 1
                elif nc == '*':
                    processed.append('\\')
                    processed.append('*')
                    j += 1
                elif nc in '()':
                    processed.append('\\')
                    processed.append(nc)
                    j += 1
                elif nc == "'":
                    j += 1  # skip backslash, keep quote
                else:
                    # Keep first backslash, reprocess rest
                    processed.append('\\')
                    # Don't advance j, the remaining backslashes will be processed
            else:
                # trailing backslashes - keep one
                processed.append('\\')
        else:
            processed.append(sql_raw[j])
            j += 1
    
    sql_final = ''.join(processed)
    replacement = f'mock.{sym}(`{sql_final}`)'
    out.append(replacement)
    i = cq + 2  # skip past closing ")
    n_conv += 1

result = ''.join(out)
print(f"Converted {n_conv} mock strings")

# ============================================================
# Write and test
# ============================================================
with open(SRC, 'w') as f:
    f.write(result)

r = subprocess.run(['go', 'build', './handlers/'], capture_output=True, text=True, cwd='/root/project/backend')
print(f"Build: {'OK' if r.returncode == 0 else 'FAIL'}")
if r.returncode:
    print(r.stderr[:500])
    exit(1)

r2 = subprocess.run(['go', 'test', './handlers/', '-count=1'], capture_output=True, text=True, cwd='/root/project/backend', timeout=90)
lines = (r2.stdout + r2.stderr).split('\n')
passes = sum(1 for l in lines if '--- PASS:' in l)
fails = sum(1 for l in lines if '--- FAIL:' in l)
print(f"Tests: {passes} PASS, {fails} FAIL, exit={r2.returncode}")
for l in lines:
    if '--- FAIL:' in l or ('unfulfilled' in l and 'Expectation' in l):
        print(f"  {l.strip()[:130]}")
