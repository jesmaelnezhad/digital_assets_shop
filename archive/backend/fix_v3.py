#!/usr/bin/env python3
"""Fix all_features_test.go: convert double-quoted to backtick raw strings, fix escapes."""
SRC = "/root/project/backend/handlers/all_features_test.go"
with open(SRC, "rb") as f:
    raw = f.read()

text = raw.decode("utf-8")
out = []
i = 0
while i < len(text):
    # Find mock.ExpectQuery(" or mock.ExpectExec("
    sym = None
    pos = -1
    for s in ["mock.ExpectQuery(\"", "mock.ExpectExec(\""]:
        p = text.find(s, i)
        if p >= 0 and (pos < 0 or p < pos):
            pos = p
            sym = s.split("(")[0]
    if pos < 0:
        out.append(text[i:])
        break
    out.append(text[i:pos])
    qstart = pos + len(sym) + 2  # right after the opening "
    # Find closing "
    j = qstart
    while j < len(text):
        if text[j] == '"':
            # Check it's not escaped (though in Go double-quoted strings, \" is escaped)
            # But our strings don't have embedded quotes, so just find next "
            break
        j += 1
    if j >= len(text):
        out.append(text[pos:])
        i = len(text)
        break
    sql = text[qstart:j]
    # Convert to backtick raw string
    # Fix escapes: \\\ -> \, \$ -> $ (raw strings don't process escapes)
    # In the double-quoted source: \\\\ = \, \\\$ = invalid
    # We want the raw string to contain: \$ for matching $1, $2 etc.
    # The sql variable is the raw text on disk between the double quotes
    # If disk has "SELECT ... \\\\\$1", sql = 'SELECT ... \\\\\$1'
    # In raw string: we want "SELECT ... \$1" 
    # So: replace \\\\ with \ and then \$ stays as \$ (raw string literal)
    raw_sql = sql.replace("\\\\\\\\", "\\")  # \\\\ on disk -> \ in raw string
    # Now handle \$: in raw string context, \$ is fine (literal backslash + dollar)
    # But we want sqlmock to see \$ which matches literal $ in regex
    # raw_sql currently has \\\$ or \$ depending on source
    # If source had \\\\\$ (5 chars on disk: \ \ \ $), sql after replace = \$ (2 chars)
    # If source had \\$ (3 chars on disk: \ \ $), sql after replace = \$ (2 chars)  
    # Both result in \$ in the raw string, which is what we want!
    replacement = f"mock.ExpectQuery(`{raw_sql}`)"
    if sym == "mock.ExpectExec":
        replacement = f"mock.ExpectExec(`{raw_sql}`)"
    out.append(replacement)
    i = j + 1

result = "".join(out)

# Now fix any remaining \\\$ inside backtick strings -> \$  
# Walk through and fix inside backticks
out2 = bytearray()
i = 0
depth = 0
while i < len(result):
    b = result[i] if isinstance(result, bytes) else ord(result[i])
    ch = result[i] if i < len(result) else None
    if i < len(result) and result[i] == '`':
        depth += 1
        out2.append(ord('`'))
        i += 1
        while i < len(result) and result[i] != '`':
            c = result[i]
            # Check for \\$ (backslash backslash dollar) -> \$ (backslash dollar)
            if depth % 2 == 1 and i + 2 < len(result) and result[i] == '\\' and result[i+1] == '\\' and result[i+2] == '$':
                out2.append(ord('\\'))
                out2.append(ord('$'))
                i += 3
            elif depth % 2 == 1 and i + 1 < len(result) and result[i] == '\\' and result[i+1] == '*':
                out2.append(ord('*'))
                i += 2
            else:
                out2.append(ord(c))
                i += 1
        if i < len(result):
            out2.append(ord('`'))
            i += 1
    else:
        out2.append(ord(ch) if ch else 0)
        i += 1

final = bytes(out2).decode("utf-8", errors="replace")

# Fix COUNT star patterns inside raw strings
final = final.replace("COUNT\\(\\*\\\)", "COUNT\\(\\*\\)")

with open(SRC, "w") as f:
    f.write(final)

print(f"Written {len(final)} bytes")

# Check build
import subprocess
r = subprocess.run(["go", "build", "./handlers/"], capture_output=True, text=True, cwd="/root/project/backend")
print(f"Build exit={r.returncode}")
if r.returncode:
    print(r.stderr[:300])
else:
    # Run tests
    r2 = subprocess.run(["go", "test", "./handlers/", "-count=1"], capture_output=True, text=True, cwd="/root/project/backend", timeout=60)
    lines = r2.stdout.split('\n')
    for l in lines[-10:]:
        print(l)
    # Count passes/fails
    passes = sum(1 for l in lines if "--- PASS:" in l)
    fails = sum(1 for l in lines if "--- FAIL:" in l)
    print(f"\nResult: {passes} PASS, {fails} FAIL")
