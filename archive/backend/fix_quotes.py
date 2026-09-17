#!/usr/bin/env python3
"""Fix all_features_test.go: fix double-quoted strings with bad escapes."""
SRC = "/root/project/backend/handlers/all_features_test.go"
with open(SRC, "r") as f:
    text = f.read()

out = []
i = 0
count = 0
while i < len(text):
    # Find mock.ExpectQuery(" or mock.ExpectExec("
    idx_q = text.find('mock.ExpectQuery("', i)
    idx_e = text.find('mock.ExpectExec("', i)
    
    if idx_q < 0 and idx_e < 0:
        out.append(text[i:])
        break
    
    if idx_q >= 0 and (idx_e < 0 or idx_q <= idx_e):
        idx = idx_q
        sym = "mock.ExpectQuery"
    else:
        idx = idx_e
        sym = "mock.ExpectExec"
    
    out.append(text[i:idx])
    
    # Find the opening quote
    qstart = text.find('("', idx) + 2
    # Find closing quote
    j = qstart
    while j < len(text) and text[j] != '"':
        j += 1
    
    sql = text[qstart:j]
    
    # Convert to backtick raw string
    # Fix: \\\$ on disk -> \$ in raw string (sqlmock sees \$ = regex for literal $)
    # Fix: \\\\* on disk -> \* in raw string (sqlmock sees \* = regex for literal *)
    raw_sql = sql
    
    # Replace \\ with \ (the double-quoted interpretation)
    # In the double-quoted Go source: "\\\\" -> Go value: "\\"
    # We want the raw string to contain: "\"
    # sql is the raw on-disk text, so "\\\\" in sql -> "\\" in raw
    raw_sql = raw_sql.replace("\\\\\\\\", "\\")
    
    replacement = f"{sym}(`{raw_sql}`)"
    out.append(replacement)
    i = j + 1
    count += 1

final = "".join(out)

with open(SRC, "w") as f:
    f.write(final)

print(f"Fixed {count} mock strings")
print(f"Wrote {len(final)} bytes")

# Verify
import subprocess
r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True, cwd="/root/project/backend")
print(f"Build: exit={r.returncode}")
if r.returncode:
    print(r.stderr[:500])
