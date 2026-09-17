# Systematic Mock SQL Catalogue Workflow

When 40+ tests fail with sqlmock "unfulfilled expectations", fixing mocks one-at-a-time is slow and error-prone. A faster approach: extract every SQL query from all handler source files, catalogue them, then fix mocks to match.

## Why this works

The 45 failing tests in Pawradise split into a small number of SQL mismatch patterns:
- JOIN queries vs simple table queries (products uses LEFT JOIN categories)
- Multi-line fmt.Sprintf SQL vs single-line mock patterns
- Column name mismatches (title vs name, price_usd vs price)
- COALESCE patterns
- ORDER BY/LIMIT/OFFSET clauses missing from mocks
- Placeholder count mismatches ($1/$2/$3)

Reading handler source once and cataloguing all queries reveals these patterns at a glance, letting you fix entire classes of mismatches with a few patches instead of 45 individual edits.

## Step 1: Extract all handler SQL

```python
import re, os

os.chdir("/root/project/backend")
handler_sqls = {}

for fname in ["products.go", "auth.go", "features.go", "community.go",
              "exchange_rates.go", "admin_features.go"]:
    with open(f"handlers/{fname}") as f:
        content = f.read()
    sqls = []
    for m in re.finditer(r'"(?:[^"\\]|\\.)*"', content):
        s = m.group().strip('"')
        s = s.replace('\\\\', '\\').replace('\\"', '"').replace('\\$', '$')
        s = s.strip()
        if s and any(kw in s.upper() for kw in ["SELECT", "INSERT", "UPDATE", "DELETE"]):
            sqls.append(s)
    handler_sqls[fname] = sqls

# Print unique SQL patterns
seen = set()
for fname, sqls in handler_sqls.items():
    for s in sqls:
        if s not in seen and len(s) > 20:
            seen.add(s)
            print(f"[{fname}] {s}")
```

## Step 2: Group by mismatch type

From the catalogue, identify the mismatch categories:

| Category | Example | Fix |
|----------|---------|-----|
| JOIN vs simple | Handler: `FROM products p LEFT JOIN categories c` / Mock: `FROM products` | Use full JOIN query in mock |
| Multi-line SQL | Handler uses `fmt.Sprintf` with `\n\t` formatting | Use `(?s)` flag + `\s+` in regex, or match essential parts only |
| Column mismatch | Handler: `p.title, p.price_usd` / Mock: `p.name, p.price` | Copy column names verbatim from handler SELECT |
| COALESCE | Handler: `COALESCE(c.name, '')` / Mock: `c.name` | Match exact COALESCE expression |
| Missing ORDER BY/LIMIT | Handler appends via fmt.Sprintf / Mock omits | Include full clause in mock |
| Placeholder count | Handler: `$1, $2, $3` / Mock: `$1, $2` or missing WithArgs | Match exact count + WithArgs chain |

## Step 3: Apply fixes by category

Instead of 45 individual patches, apply fixes per category:

**JOIN queries** — fix all product-related mocks to include the full LEFT JOIN:
```python
# All product list mocks need: FROM products p LEFT JOIN categories c ON p.category_id = c.id
# Pattern: replace "FROM products" with the full JOIN version
```

**Column names** — fix all mocks to use handler's exact column names:
```python
# Handler uses: id, title, slug, description, category_id, COALESCE(c.name,''), price_usd, ...
# Not: id, name, price, stock
```

**Placeholders** — ensure every `$N` in mock is `\$N` (escaped for regex):
```bash
grep -n 'mock.Expect' handlers/all_features_test.go | grep -v '\\\$' | grep '\$'
# Any line with $ but not \$ in mock SQL is wrong
```

## Step 4: Verify incrementally

After each category fix, run the specific tests that should pass:
```bash
go build ./handlers/ && go test ./handlers/ -v -count=1 -run "TestProducts_GetProducts$"
```

Don't run the full suite until the category is done — faster feedback.

## Key insight from this session

The handler SQL is the ground truth. Never invent mock SQL from memory or assumptions about what the handler "probably" does. Always read the handler source. The `execute_code` tool is ideal for this — it can read all handler files, extract SQL, and present it in one view without flooding the conversation with 6000+ chars of handler source.

## Anti-patterns to avoid

- **Fixing mocks without reading handler source first** — leads to wrong column names, wrong JOIN structure, wrong placeholder counts.
- **Full-file Python regex rewrites** — these repeatedly corrupt previously-passing tests. Use surgical `patch` operations or category-based Python scripts that target specific line ranges.
- **Assuming v5+bytefix gives green tests** — it gives BUILD:0 + 19 PASS, but 45 SQL mismatches remain. The mocks still don't match handler SQL. Plan for the post-v5 fix work.
