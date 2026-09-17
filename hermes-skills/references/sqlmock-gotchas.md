# sqlmock Gotchas for Pawradise Handlers

The Pawradise backend uses `github.com/DATA-DOG/go-sqlmock` for handler unit
tests. Several recurring pitfalls have been hit across multiple fix cycles.

## 1. The `$N` regex trap (most common, most expensive)

**Handler SQL uses `$1`, `$2` as PostgreSQL placeholders.**

**sqlmock matches expectations against the *actual query string* using Go
`regexp`.** In Go regexp, `$` is the end-of-line anchor. So a mock expecting
`"SELECT * FROM users WHERE email = $1"` actually matches `"SELECT ... email
= "` followed by end-of-line, then `"1"` — which never matches the real query
that has a literal `$1`.

**Fix:** escape the `$` in mock patterns:

```go
// WRONG — $1 is regex end-of-line + "1", never matches literal $1
mock.ExpectQuery("SELECT * FROM users WHERE email = $1")

// RIGHT — one of these two forms:
// Form A: double-quoted with double backslash (Go string escape + regex escape)
mock.ExpectQuery("SELECT * FROM users WHERE email = \\$1")
// Form B: backtick raw string with single backslash (regex escape only)
mock.ExpectQuery(`SELECT * FROM users WHERE email = \$1`)
```

**Why this is expensive:** failing to escape `$` causes the mock to never
receive the expected query. The test then fails with "expected query not
received" or "expectations were not met" — which looks like a logic bug, not
a regex issue. Multiple fix cycles have been spent on this.

**Rule of thumb:** every `$N` in a mock SQL string MUST be `\$N`. Grep the
test file for `\$` after writing mocks — if a mock SQL line has `$` without
`\` before it, it's wrong.

## 1b. The COUNT(`\*`) byte-level trap

After converting mock SQL from double-quoted strings to backtick raw strings
(e.g. via `fix_v5.py`), COUNT patterns end up as `COUNT\(\\*` in the raw
bytes — two backslash bytes followed by star inside a backtick string. This
becomes the regex `COUNT\(\\*` which tries to match two literal backslashes
before the star, but the handler sends `COUNT(*)` with no backslashes. The
mock never matches and tests fail with "unfulfilled expectations".

**The fix is byte-level, not text-level:** Python `str.replace()` sees the
printable representation and can't reliably distinguish `\\\\*` (4 chars in
text) from other sequences. Use raw bytes:

```python
with open("handlers/all_features_test.go", "rb") as f:
    data = f.read()

old = bytes([0x5c, 0x5c, 0x2a])  # \\* (two backslash bytes + star)
new = bytes([0x5c, 0x2a])        # \*  (one backslash byte + star)

data = data.replace(old, new)

with open("handlers/all_features_test.go", "wb") as f:
    f.write(data)
```

**Verify with `grep -c 'COUNT' handlers/all_features_test.go` and check the
hex of the first match:** `python3 -c "d=open('handlers/all_features_test.go','rb').read(); i=d.find(b'COUNT'); print(d[i:i+15], d[i:i+15].hex())"`.

Expected hex for a correct COUNT pattern: `434f554e545c285c2a29` = `COUNT\(*)`.
If you see `434f554e545c285c5c2a29` = `COUNT\(\\*)` the byte fix hasn't
landed.

See [references/test-file-repair.md](references/test-file-repair.md) for the
full repair workflow.

## 1b. The `\$` over-escaping trap in backtick strings

After converting mock SQL from double-quoted to backtick raw strings (via `fix_v5.py` or similar), some patterns end up with `\\$1` (two backslashes + dollar) inside backtick strings. In Go backtick, `\\` is two literal backslashes, so the regex engine sees `\\$1` = escaped backslash + end-of-line anchor + `1` — which never matches the handler's literal `$1`.

**The correct pattern in backtick is `\$1` (one backslash + dollar).**

**Fix (byte-level, not text-level):**

```python
with open("handlers/all_features_test.go", "rb") as f:
    data = f.read()

# Fix 2-backslash + dollar -> 1-backslash + dollar
old = bytes([0x5c, 0x5c, 0x24])  # \\$
new = bytes([0x5c, 0x24])        # \$
data = data.replace(old, new)

with open("handlers/all_features_test.go", "wb") as f:
    f.write(data)
```

Always pair this with the COUNT fix (see 1c below) — both are needed after v5 conversion.

## 1c. The COUNT(`\*`) byte-level trap

After v5 conversion, COUNT patterns end up as `COUNT\(\\*` in raw bytes — two backslash bytes + star inside a backtick string. The regex `COUNT\(\\*` tries to match two literal backslashes before the star, but the handler sends `COUNT(*)` with no backslashes.

**Fix (byte-level):**

```python
with open("handlers/all_features_test.go", "rb") as f:
    data = f.read()

old = bytes([0x5c, 0x5c, 0x2a])  # \\* (two backslash bytes + star)
new = bytes([0x5c, 0x2a])        # \*  (one backslash byte + star)
data = data.replace(old, new)

with open("handlers/all_features_test.go", "wb") as f:
    f.write(data)
```

**Verify with hex:** `python3 -c "d=open('handlers/all_features_test.go','rb').read(); i=d.find(b'COUNT'); print(d[i:i+15].hex())"` should show `434f554e545c285c2a29` (COUNT\(\*)).

## 1d. Multi-line formatted SQL from fmt.Sprintf

Handlers that build queries with `fmt.Sprintf` using `\n\t` formatting produce multi-line SQL with tabs and newlines. A single-line mock regex pattern will not match the full multi-line query.

**Fix options:**
1. **Dotall mode + flexible whitespace:** Use `(?s)` flag and `\s+` for whitespace in the regex:
   ```go
   mock.ExpectQuery(`(?s)SELECT\s+p\.id,\s+p\.title.*FROM\s+products\s+p.*ORDER\s+BY.*LIMIT\s+\$2\s+OFFSET\s+\$3`)
   ```
   Note: `\s` in Go backtick = literal backslash-s, which Go regex interprets as whitespace metacharacter.

2. **QueryMatcherEqual for exact match:** If the handler SQL is deterministic, use `QueryMatcherEqual` to match byte-for-byte:
   ```go
   mock.ExpectQuery("SELECT p.id, p.title...\nFROM products p\n...").
       QueryMatcher(sqlmock.QueryMatcherEqual).
       WillReturnRows(...)
   ```
   This avoids regex entirely but requires exact whitespace matching.

3. **Minimize mock specificity:** Match only the essential parts of the query (SELECT clause + FROM + WHERE) and use `.*` for the rest:
   ```go
   mock.ExpectQuery(`SELECT p\.id, p\.title.*FROM products p.*WHERE.*ORDER BY.*LIMIT.*OFFSET.*`)
   ```
   Less brittle but may match unintended queries.

**Critical:** After any bulk edit of Go files, run `go build ./...` immediately. Regex replacements across newlines can split backtick strings, producing "missing ',' in argument list" errors.

Handlers that use `db.QueryRow("INSERT INTO ... RETURNING id", ...)` produce
a **query**, not an exec. Tests must use `ExpectQuery`, not `ExpectExec`:

```go
// Handler:
db.QueryRow("INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3) RETURNING id", ...)

// Test:
mock.ExpectQuery("INSERT INTO users ... RETURNING \\$1, \\$2, \\$3").
    WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
```

Using `ExpectExec` for these causes the handler to get no rows → 500
"Failed to create user".

## 4. Time.Time in mock rows

Models use `time.Time` for timestamp fields. When building `sqlmock.NewRows`,
pass real `time.Time` values — not strings:

```go
// RIGHT
time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)

// WRONG — sqlmock may parse some date strings but it is fragile
"2024-01-01T00:00:00Z"
```

Plain strings cause `Scan` to fail silently (handler sees no rows → 401 or
500), and the test appears to be a mock mismatch when the real bug is a type
error.

## 5. sqlmock is an interface, not a pointer

`sqlmock.New()` returns `(*sql.DB, sqlmock.Sqlmock, error)`. The mock is the
interface `sqlmock.Sqlmock` — not `*sqlmock.Sqlmock`. Assigning to the
pointer type produces:

```
sqlmock.Sqlmock does not implement *sqlmock.Sqlmock
(type *sqlmock.Sqlmock is pointer to interface, not interface)
```

Keep it as `sqlmock.Sqlmock`. Return it from setup functions as the
interface type.

## 6. Go raw strings and bulk edits

Go raw strings (backtick literals) span multiple lines and contain arbitrary
text. Regex replacements across newlines can split backtick strings:

```go
// Before:
`SELECT p.id, p.title, c.name FROM products p`

// After a bad bulk edit:
`SELECT p.id`.id` FROM products p`  // syntax error
```

**Rule:** after ANY bulk text edit of Go files (sed, Python regex, patch),
run `go build ./...` immediately. Don't assume the edit was clean. Prefer
targeted `patch` calls over broad regex replacements on Go source.

## 7. QueryMatcherEqual vs default matcher

By default, sqlmock uses `QueryMatcherRegexp` — it treats the expected string
as a regex. If you want exact string matching (no regex interpretation at
all), use `QueryMatcherEqual`:

```go
mock.ExpectQuery("SELECT * FROM users WHERE email = $1").
    QueryMatcher(sqlmock.QueryMatcherEqual).
    WillReturnRows(...)
```

With `QueryMatcherEqual`, `$1` is matched literally (no regex). This avoids
the `$N` escaping problem entirely but requires the mock string to match the
handler SQL byte-for-byte (including whitespace).

**Trade-off:** `QueryMatcherEqual` is stricter (whitespace must match exactly)
but eliminates the `$N` regex class of bugs. Consider it when a test has
multiple `$N` placeholders and the regex escaping becomes error-prone.

## 8. ExpectationsWereMet after a test

If a test passes but `mock.ExpectationsWereMet()` fails, the handler did not
execute one of the expected queries. This usually means:
- The handler took a different code path (e.g. early return on error)
- The mock expected a query the handler never issued
- The mock expected a query with different SQL than what the handler used

Always call `mock.ExpectationsWereMet()` in a `defer` or at the end of each
test to catch silent mismatches.
