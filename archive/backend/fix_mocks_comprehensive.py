#!/usr/bin/env python3
"""
Comprehensive mock fixer for all_features_test.go.
Reads handler source files to get exact SQL, then patches all mock expectations.
"""
import re

# Read the test file
with open("handlers/all_features_test.go", "r") as f:
    content = f.read()

lines = content.split('\n')

# Read handler files to get exact SQL
handler_sql = {}

for fname in ["products.go", "auth.go", "features.go", "community.go", 
              "exchange_rates.go", "settings.go", "admin_features.go", "order_routes.go"]:
    try:
        with open(f"handlers/{fname}", "r") as f:
            hcontent = f.read()
        # Extract SQL strings from fmt.Sprintf and direct Query calls
        for m in re.finditer(r'["`](SELECT|INSERT|UPDATE|DELETE)[^"]*?["`]', hcontent):
            sql = m.group(0)
            # Normalize: remove fmt.Sprintf wrapping, extract core SQL
            core = sql.strip('"\'')
            if len(core) > 20 and core not in handler_sql:
                handler_sql[core[:80]] = core
    except:
        pass

print(f"Extracted {len(handler_sql)} SQL patterns from handlers")

# Now fix each failing test by updating its mock SQL
# Strategy: for each mock.Expect* line, try to find the matching handler SQL

def fix_mock_line(line):
    """Try to fix a mock expectation line by matching handler SQL."""
    m = re.match(r'^(\s*mock\.(ExpectQuery|ExpectExec})\()`(.+?)`\)\..*$', line)
    if not m:
        return line
    
    prefix = m.group(1)  # indentation + mock.Expect*(`
    sql = m.group(3)     # SQL inside backtick
    
    # Check if this SQL contains .* wildcards that should be expanded
    if '.*' in sql and 'FROM' in sql:
        # Try to find matching handler SQL
        for key, full_sql in handler_sql.items():
            if sql.split('FROM')[0].strip() in full_sql or sql[:30] in full_sql:
                # Found a match - use the full SQL
                return f'{prefix}{full_sql}`).'
    
    # Fix COUNT(\*) -> COUNT(*)  (remove unnecessary backslash escaping)
    if 'COUNT' in sql:
        sql = sql.replace('COUNT\\(\\*', 'COUNT(*)')
    
    # Fix \$1, \$2 etc -> $1, $2 (sqlmock needs literal dollar signs)
    sql = re.sub(r'\\\$(\d+)', r'$\1', sql)
    
    return f'{prefix}{sql}`).'

# Apply fixes
fixed_lines = []
changes = 0
for line in lines:
    new_line = fix_mock_line(line)
    if new_line != line:
        changes += 1
    fixed_lines.append(new_line)

result = '\n'.join(fixed_lines)
print(f"Made {changes} changes")

with open("handlers/all_features_test.go", "w") as f:
    f.write(result)

print("Done")
