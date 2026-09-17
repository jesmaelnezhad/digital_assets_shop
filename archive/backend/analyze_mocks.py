#!/usr/bin/env python3
"""
Batch fix: read all handler SQL, then patch test mocks to match.
Operates on the currently-fixed test file (post-v5 + COUNT fix).
"""
import re, os

os.chdir("/root/project/backend")

# 1. Extract exact SQL from all handlers
handler_queries = []

for fname in ["products.go", "auth.go", "features.go", "community.go",
              "exchange_rates.go", "settings.go", "admin_features.go"]:
    with open(f"handlers/{fname}") as f:
        code = f.read()
    # Find all SQL strings passed to DB methods
    for m in re.finditer(r'\.(Query|QueryRow|Exec)\( ?("(?:[^"\\]|\\.)*"|`(?:[^`\\]|\\.)*`)', code):
        sql = m.group(2)
        if sql.startswith('"'):
            # Unescape Go string
            sql = sql[1:-1].encode().decode('unicode_escape')
        else:
            sql = sql[1:-1]  # backtick - raw
        if len(sql) > 15 and sql.upper().startswith(('SELECT', 'INSERT', 'UPDATE', 'DELETE')):
            handler_queries.append(sql)

print(f"Extracted {len(handler_queries)} handler queries")

# 2. Read test file and find all mock patterns
with open("handlers/all_features_test.go") as f:
    content = f.read()

# Find all mock SQL patterns and their line numbers
mock_patterns = []
for i, line in enumerate(content.split('\n'), 1):
    m = re.search(r'mock\.(ExpectQuery|ExpectExec)\(`([^`]+)`\)', line)
    if m:
        mock_patterns.append((i, m.group(1), m.group(2)))

print(f"Found {len(mock_patterns)} mock patterns in test file")

# 3. For each mock, try to find matching handler query and report mismatches
fixes = []
for line_num, method, mock_sql in mock_patterns:
    mock_upper = mock_sql.upper()
    
    # Skip wildcard-only patterns that are fine
    if '.*' in mock_sql and not any(kw in mock_sql.upper() for kw in ['COUNT', 'WHERE', 'JOIN', 'ORDER', 'LIMIT']):
        continue
    
    # Try to find matching handler query
    best_match = None
    best_score = 0
    
    for hq in handler_queries:
        hq_upper = hq.upper()
        score = 0
        
        # Check if verb matches
        mock_verb = mock_upper.split()[0] if mock_upper.split() else ''
        hq_verb = hq_upper.split()[0] if hq_upper.split() else ''
        if mock_verb == hq_verb:
            score += 10
        
        # Check if table references match  
        mock_tables = set(re.findall(r'FROM\s+(\S+)', mock_upper))
        hq_tables = set(re.findall(r'FROM\s+(\S+)', hq_upper))
        if mock_tables & hq_tables:
            score += 5 * len(mock_tables & hq_tables)
        
        # Check WHERE clause similarity
        mock_where = re.search(r'WHERE\s+(.+?)(ORDER|GROUP|LIMIT|$)', mock_upper)
        hq_where = re.search(r'WHERE\s+(.+?)(ORDER|GROUP|LIMIT|$)', hq_upper)
        if mock_where and hq_where:
            if mock_where.group(1).split('AND')[0].strip()[:30] in hq_where.group(1):
                score += 3
        
        if score > best_score:
            best_score = score
            best_match = hq
    
    if best_match and best_score >= 10 and mock_sql != best_match:
        fixes.append((line_num, method, mock_sql, best_match))
    elif not best_match and '.*' not in mock_sql:
        fixes.append((line_num, method, mock_sql, None))

print(f"\nPotential fixes needed: {len(fixes)}")
for ln, method, mock, hq in fixes[:10]:
    print(f"\n  Line {ln} ({method}):")
    print(f"    Mock:  {mock[:80]}")
    if hq:
        print(f"    Match: {hq[:80]}")
    else:
        print(f"    No match found")
