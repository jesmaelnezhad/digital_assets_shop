# Test File Repair Workflow for all_features_test.go

The Pawradise comprehensive test file `handlers/all_features_test.go` (788 lines, 77 mock expectations) has a fragile state due to Go escape-sequence issues in double-quoted mock strings. This document captures the only known-reliable repair workflow.

## Root cause

The original test file uses double-quoted Go strings for mock SQL:

```go
mock.ExpectQuery("SELECT id, email FROM users WHERE email = \\\\$1")
```

Inside a double-quoted Go string, `\\\\$` is an invalid escape sequence (Go only recognizes `\\`, `\"`, `\n`, `\r`, `\t`, `\\x..`). This causes a compile error: `unknown escape sequence`.

## The stable fix: restore + v5 + byte-level COUNT fix + surgical patches

### Step 1: Restore from backup

```bash
cp handlers/all_features_test.go.bak handlers/all_features_test.go
```

The backup compiles (`go build ./handlers/` → BUILD:0) but has the escape-sequence error in double-quoted mock strings.

### Step 2: Apply fix_v5.py

```bash
python3 fix_v5.py
```

This script (in `/root/project/backend/fix_v5.py`) converts all 77 double-quoted mock SQL strings to backtick raw strings, fixes `\\$` → `\$` escaping, and removes extra closing parens introduced by earlier botched conversions.

**Result:** BUILD:0, ~19 PASS, ~42 FAIL. The remaining 42 failures are SQL mismatches (column names, aliases, ORDER BY/LIMIT clauses, missing `WithArgs`, placeholder escaping).

### Step 3: Fix COUNT byte-level

After v5 conversion, COUNT patterns in the file have the byte sequence `\\\*` (two backslash bytes followed by star) inside backtick raw strings. This becomes the regex pattern `COUNT\(\\*` which matches `COUNT(` + two literal backslashes + `*` + `)` — but the handler sends `COUNT(*)` (no backslashes). The mock never matches.

**Fix with Python byte-level replacement:**

```python
with open("handlers/all_features_test.go", "rb") as f:
    data = f.read()

old = bytes([0x5c, 0x5c, 0x2a])  # \\* (two backslash bytes + star)
new = bytes([0x5c, 0x2a])        # \*  (one backslash byte + star)

count = data.count(old)
data = data.replace(old, new)

with open("handlers/all_features_test.go", "wb") as f:
    f.write(data)

print(f"Fixed {count} COUNT patterns")
```

**Verify with hex inspection:**
```bash
python3 -c "d=open('handlers/all_features_test.go','rb').read(); i=d.find(b'COUNT'); print(d[i:i+15], d[i:i+15].hex())"
```
Expected hex for correct COUNT: `434f554e545c285c2a29` = `COUNT\(\*)`, i.e. `COUNT(\*)` in the file.
If you see `434f554e545c285c5c2a29` = `COUNT\(\\*)` the byte fix hasn't landed.

**Why text-mode replacement fails:** The Python `str.replace()` sees the printable representation `\\\\*` (four chars in text) and can't reliably distinguish it from other backslash sequences. Byte-level replacement is unambiguous.

### Step 3b: Fix over-escaped `\$` in backtick strings

After v5 conversion, some mock SQL patterns have `\\$1` (two backslashes + dollar) inside backtick strings. In Go backtick, `\\` is two literal backslashes, so the regex sees `\\$1` = escaped backslash + end-of-line anchor + `1` — which never matches the handler's literal `$1`. The correct regex pattern in backtick is `\$1` (one backslash + dollar).

**Fix with Python byte-level replacement:**

```python
with open("handlers/all_features_test.go", "rb") as f:
    data = f.read()

# Fix 2-backslash + dollar -> 1-backslash + dollar (inside backtick raw strings)
old = bytes([0x5c, 0x5c, 0x24])  # \\$ 
new = bytes([0x5c, 0x24])        # \$
count = data.count(old)
data = data.replace(old, new)

with open("handlers/all_features_test.go", "wb") as f:
    f.write(data)

print(f"Fixed {count} over-escaped dollar patterns")
```

**Combined step 3:** Apply both fixes together:
```python
with open("handlers/all_features_test.go", "rb") as f:
    data = f.read()

# Fix COUNT(\\* -> COUNT(\*  (two backslashes + star -> one)
old1 = bytes([0x5c, 0x5c, 0x2a])
new1 = bytes([0x5c, 0x2a])
data = data.replace(old1, new1)

# Fix \\$ -> \$ (two backslashes + dollar -> one)  
old2 = bytes([0x5c, 0x5c, 0x24])
new2 = bytes([0x5c, 0x24])
data = data.replace(old2, new2)

with open("handlers/all_features_test.go", "wb") as f:
    f.write(data)

print(f"COUNT fixed: {data.count(bytes([0x5c,0x2a]))} patterns")
print(f"Dollar fixed: applied")
```

### Step 3c: Verify escaping on key lines

After byte-level fixes, verify with `cat -A` on critical lines (14, 27, 123):
```bash
cat -A handlers/all_features_test.go | sed -n '14p;27p;123p'
```

**Correct state:**
- Line 14: backtick with `INSERT INTO users.*` — no escaping issues (no $ or *)
- Line 27: backtick with `\$1` — one backslash before dollar (regex-escaped)
- Line 123: backtick with `COUNT\(\*) ` — one backslash before star (regex-escaped)

**Incorrect states to watch for:**
- `\\$1` (two backslashes before dollar) — over-escaped, won't match
- `\\*` (two backslashes before star) — over-escaped, won't match
- `"..."` (double-quoted) — not yet converted to backtick, will have compile errors

## Step 4: Surgical patches for remaining SQL mismatches

From the 19 PASS / 45 FAIL baseline, fix individual failing tests with targeted `patch` operations. Each patch must use a unique surrounding function-context anchor.

**Pattern for each failing test:**
1. Read the current mock block from the file (read_file, not from memory)
2. Read the corresponding handler SQL from the handler source file
3. Construct the patch with the exact current file content as `old_string`
4. Apply the patch
5. Run `go build ./handlers/` → verify BUILD:0
6. Run the specific failing test → verify it passes
7. Run the full suite → verify no regressions

**DO NOT use full-file Python regex rewrites** (`fix_all.py`, `fix_all_mocks.py`, `fix_remaining.py`, etc.) — they have repeatedly corrupted previously-passing tests by introducing extra closing parens, broken backtick strings, wrong SQL patterns, and missing `WithArgs` chains.

**Common SQL mismatch categories (45 failures typically split across these):**

1. **JOIN queries vs simple table queries** — handlers use `LEFT JOIN categories c ON p.category_id = c.id` but mocks use simple `FROM products` without JOIN. Fix: copy the full JOIN query from the handler.

2. **Multi-line formatted SQL** — handlers use `fmt.Sprintf` with `\n\t` formatting producing multi-line queries. Mocks with single-line patterns don't match. Fix: use `(?s)` flag for dotall mode + `\s+` for flexible whitespace matching in the regex.

3. **COALESCE patterns** — handlers use `COALESCE(c.name, '')` but mocks may use `c.name` directly or have different quoting. Fix: match the exact COALESCE expression.

4. **Column name mismatches** — handlers use `title`, `price_usd`, `stock_quantity`, `category_name` (via COALESCE) but mocks may use `name`, `price`, `stock`, etc. Fix: copy column names verbatim from handler's SELECT list.

5. **ORDER BY / LIMIT / OFFSET clauses** — handlers append these via `fmt.Sprintf` but mocks may omit them. Fix: include the full ORDER BY/LIMIT/OFFSET in the mock pattern.

6. **Placeholder count mismatches** — handlers use `$1`, `$2`, `$3`etc. but mocks may use different numbers or missing `WithArgs` chains. Fix: match the exact placeholder count and `WithArgs` arguments.

## Step 5: Verify

```bash
cd /root/project/backend
go build ./handlers/ && echo "BUILD OK"
go test ./handlers/ -count=1 -v 2>&1 | grep -E "^(--- PASS|--- FAIL)" | sort | uniq -c | sort -rn
go test ./handlers/ -count=1 -coverprofile=/tmp/cover.out 2>&1 | tail -3
go tool cover -func=/tmp/cover.out | grep total
```

Target: all tests PASS, coverage ≥70%.

## Known fragile patterns to avoid

| Pattern | Result | Avoid? |
|---------|--------|
| Full-file Python regex rewrite of all_features_test.go | Corrupted previously-passing tests, extra `)` syntax errors, broken backtick strings, wrong SQL patterns | YES — never |
| `sed` with complex backslash patterns | Unpredictable results on Go raw strings | YES — use Python byte-level instead |
| Patching without reading current file content first | `old_string` doesn't match → patch fails or hits wrong location | YES — always read first |
| Applying fix scripts in wrong order | File ends up in unrecoverable state | YES — always: backup → v5 → byte-fix → surgical patches |
| Using `mock.ExpectQuery("SELECT ... $1")` without `\$` | Regex end-of-line anchor, never matches | YES — always escape |
| Assuming v5+bytefix gives green tests | Leaves 45 SQL mismatch failures — the mocks still don't match handler SQL | YES — expect to patch individually |

## The backup file

`handlers/all_features_test.go.bak` is the known-good starting point. It compiles but has invalid Go escape sequences in double-quoted mock strings. It is NOT directly usable — it must go through v5 conversion first.

Do NOT edit the backup directly. Always work from a copy.
