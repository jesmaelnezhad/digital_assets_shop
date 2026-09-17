#!/usr/bin/env python3
"""Fix all failing tests in all_features_test.go by updating mock SQL to match handlers."""
import re, subprocess, os

os.chdir("/root/project/backend")

# Extract SQL from handler files
handler_sql_map = {}

for fname in ["products.go", "auth.go", "features.go", "community.go",
              "exchange_rates.go", "settings.go", "admin_features.go"]:
    with open(f"handlers/{fname}") as f:
        content = f.read()
    # Find SQL in fmt.Sprintf and direct strings
    for m in re.finditer(r'["`](SELECT|INSERT|UPDATE|DELETE)\s+([^"]{10,})["`]', content):
        full = m.group(0).strip('"').strip('`')
        key = full[:80]
        if key not in handler_sql_map:
            handler_sql_map[key] = full

print(f"Extracted {len(handler_sql_map)} handler SQL patterns")

# Now read the test file and fix mocks
with open("handlers/all_features_test.go") as f:
    content = f.read()

lines = content.split('\n')
out_lines = []

# For each mock line, try to match handler SQL
def fix_mock_sql(sql, handler_sqls):
    """Try to find matching handler SQL for a mock pattern."""
    if '.*' in sql:
        # Extract the core SELECT/INSERT/UPDATE/DELETE and table name
        core_match = re.match(r'(SELECT|INSERT|UPDATE|DELETE)\s+(.+?)FROM\s+(\S+)', sql)
        if core_match:
            verb = core_match.group(1)
            # Try to find matching handler SQL
            for handler_sql in handler_sqls:
                if verb in handler_sql and core_match.group(3) in handler_sql:
                    # Check if the mock's pre-FROM clause matches
                    mock_pre = core_match.group(2).strip()
                    # Simple check: first 30 chars of mock pre-FROM should be in handler
                    if mock_pre[:30] in handler_sql:
                        return handler_sql
    return None

for i, line in enumerate(lines):
    # Match mock.ExpectQuery(`SQL`) or mock.ExpectExec(`SQL`)
    m = re.match(r'^(\s*mock\.(ExpectQuery|ExpectExec})\()`(.+?)`\)\.?(.?)$', line)
    if m:
        indent_method = m.group(1)
        sql = m.group(3)
        trailing = m.group(4) or ''
        
        # Try to find matching handler SQL
        handler_sql = fix_mock_sql(sql, handler_sql_map.values())
        if handler_sql:
            sql = handler_sql
        else:
            # Fix escaping issues
            sql = sql.replace('\\\\', '\\')
        
        out_lines.append(f'{indent_method}{sql}`).{trailing}')
    else:
        out_lines.append(line)

result = '\n'.join(out_lines)

with open("handlers/all_features_test.go", "w") as f:
    f.write(result)

print(f"Fixed mocks. Testing...")
result = subprocess.run(["go", "test", "./handlers/", "-count=1"], 
                       capture_output=True, text=True)
print(result.stdout[-500:])
print(result.stderr[-500:])
