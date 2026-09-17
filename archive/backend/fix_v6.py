#!/usr/bin/env python3
"""
Fix mock SQL escaping: convert Go double-quoted strings to backtick,
properly handling the escape sequences.
"""
import re

with open("/root/project/backend/handlers/all_features_test.go.bak", "r") as f:
    content = f.read()

lines = content.split('\n')
new_lines = []

for i, line in enumerate(lines):
    # Match mock.ExpectQuery("SQL"). or mock.ExpectExec("SQL").
    # The SQL is in a Go double-quoted string
    m = re.match(r'^(\s*mock\.(?:ExpectQuery|ExpectExec)\()(".*")(\.\n?)$', line)
    if m:
        prefix = m.group(1)
        sql_raw = m.group(2)  # includes the outer " "
        suffix = m.group(3)
        
        # sql_raw is like: "SELECT ... WHERE email = \\\\\\$1"
        # The content between quotes: SELECT ... WHERE email = \\\\\\$1
        inner = sql_raw[1:-1]  # strip the " 
        # inner = SELECT ... WHERE email = \\\\\$1  (6 backslashes + dollar)
        
        # In Go double-quoted string: \\\\\$1 means:
        # \\ -> \  (one backslash from first two)
        # \\ -> \  (one backslash from next two)  
        # \$ -> $  (dollar; but this is illegal in Go!)
        # Wait, \$ is NOT a valid Go escape. The .bak file has illegal escapes.
        # So Go can't compile this as-is. The .bak is NOT compilable.
        # 
        # But the summary says .bak compiles BUILD:0 sometimes. Let me check:
        # If Go sees "\\\\\\$", it sees: \\ -> \, \\ -> \, \$ -> ILLEGAL
        # So this would be a compile error.
        # 
        # UNLESS the file has actual backslashes that got doubled somewhere.
        # The summary says line 27 has: \\\\\\$1 (6 backslashes + dollar)
        # In Go double-quoted string: \\\\\$1
        #   \\\\ -> \\ (two backslashes in source = two backslashes in string)
        #   \$ -> $ (but \$ is illegal... unless Go is lenient here?)
        # 
        # Actually, let me just figure out what the handler sends and set the mock accordingly.
        # The handler sends: SELECT ... WHERE email = $1  (literal dollar sign)
        # For regex matching in backtick: SELECT ... WHERE email = \$1  (backslash-dollar)
        # 
        # The .bak has: "WHERE email = \\\\\\$1" 
        # In Go double-quoted: \\\\\$1 = \\ (2 source -> 1 value) + \\ (2 source -> 1 value) + \$ (illegal)
        # This CAN'T compile. So the .bak does NOT compile.
        #
        # The summary says "the backup compiles BUILD:0 but has invalid Go escape sequences"
        # - this is contradictory. Let me just check by compiling.
        
        # For the FIX: I just need the backtick raw string to contain the REGEX that matches
        # the handler's SQL. Handler sends literal $1. Regex needs \$1.
        # So in backtick: `SELECT ... WHERE email = \$1`
        
        # The inner text from .bak might have extra backslashes from Go escaping confusion.
        # Let me just strip all backslashes before $ and before *, then add back one.
        
        # Normalize: replace any sequence of backslashes before $ with single \
        # And before * with single \
        inner_normalized = re.sub(r'\\+\$', r'\\$', inner)  # \\\$ -> \$
        inner_normalized = re.sub(r'\\+\*', r'\\*', inner_normalized)  # \\\\* -> \*
        
        new_lines.append(f'{prefix}`{inner_normalized}`{suffix}')
    else:
        new_lines.append(line)

result = '\n'.join(new_lines)
with open("/root/project/backend/handlers/all_features_test.go", "w") as f:
    f.write(result)

print(f"Converted {len(lines)} lines, written {len(result)} chars")
PYEOF
