## Unit Testing Patterns

### Dependency
go-sqlmock is the mock framework for this project. It was added to go.mod:
```bash
cd /root/project/backend && go get github.com/DATA-DOG/go-sqlmock@v1
```

### Mock DB setup (go-sqlmock)
Override `database.DB` globally with a sqlmock-backed `*sql.DB`. The real
PostgreSQL connection is never touched — tests are fast and deterministic.

```go
func setupMockDB(t *testing.T) (db *sql.DB, mock sqlmock.Sqlmock, cleanup func()) {
    t.Helper()
    db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
    require.NoError(t, err)
    database.DB = db          // override global
    cleanup = func() {
        db.Close()
        database.DB = nil     // restore
    }
    return
}
```

**Matcher choice matters:** `QueryMatcherEqual` requires the EXPECTED query
string to match the ACTUAL query byte-for-byte, including `$1`, `$2` placeholders
and whitespace. If the handler builds the query differently (e.g. `$1` vs `?`),
expectations silently pass but the mock never receives them — the test then
fails on `ExpectationsWereMet`. Prefer `QueryMatcherEqual` for exact control;
switch to `QueryMatcherRegexp` only when query text varies.

### The sqlmock mock value IS an interface — never store as *sqlmock.Sqlmock
`sqlmock.New()` returns `( *sql.DB, sqlmock.Sqlmock, error )`. The second return
value is the interface value — it already acts like a pointer. **Do NOT assign it
to a `*sqlmock.Sqlmock` variable.** Doing so produces:
```
cannot use mock (variable of type sqlmock.Sqlmock) as *sqlmock.Sqlmock value
in assignment: sqlmock.Sqlmock does not implement *sqlmock.Sqlmock
(type *sqlmock.Sqlmock is pointer to interface, not interface)
```
Correct: keep the mock as `sqlmock.Sqlmock` (the interface), return it from
setup, and use it directly in tests. Calling `mock.ExpectQuery(...)` on the
interface value works fine.

**Working pattern this project uses:**
```go
func setupTestEnv(t *testing.T) (func(), sqlmock.Sqlmock) {
    t.Helper()
    mockDB, mock, err := sqlmock.New()
    require.NoError(t, err)
    database.DB = mockDB
    cleanup := func() { mockDB.Close() }
    return cleanup, mock   // mock is sqlmock.Sqlmock, NOT *sqlmock.Sqlmock
}

// In each test:
cleanup, mock := setupTestEnv(t)
defer cleanup()
mock.ExpectQuery("SELECT ...").WillReturnRows(...)
```

### Import discipline in test files
Test files are compiled like production code. Unused imports cause build failure:
```
imported and not used: "fmt"
imported and not used: "os"
imported and not used: "strconv"
imported and not used: "github.com/stretchr/testify/assert"
```
Before writing a test file, audit imports against the helpers actually used.
If the file only uses `require` (not `assert`), drop `assert` from imports.

### Handler constructors vs package-level functions
**PITFALL — `undefined: NewAuthHandler`:** Some test files call constructor
functions like `NewAuthHandler(db)` that don't exist.

**This project is MIXED:** `auth.go` DOES export `NewAuthHandler(db *sql.DB)
*AuthHandler` — auth handlers are struct methods on `*AuthHandler`. Other
handlers (products, orders, community, settings, admin_features, features) are
package-level functions that read from the global `database.DB`. Check before
inventing constructors:
```bash
grep -n 'func New' handlers/*.go          # find real constructors
grep -n 'type.*Handler struct' handlers/*.go  # find handler structs
```

When wiring a test router (TR()), use the real constructors for struct-based
handlers and the package-level functions or direct registrations for the rest:
```go
func TR() *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.Use(gin.Recovery())
    // Auth handlers ARE struct-based — use the real constructor
    ah := NewAuthHandler(database.DB)
    r.POST("/api/v1/register", ah.Register)
    r.POST("/api/v1/login", ah.Login)
    r.GET("/api/v1/me", ah.GetProfile)
    r.PUT("/api/v1/me", ah.UpdateProfile)
    r.POST("/api/v1/logout", ah.Logout)
    // Other handlers use database.DB global — set that in setupTestEnv
    return r
}
```

### Test router
Build a minimal Gin engine that mirrors main.go's route wiring for the endpoints
under test. Example skeleton:
```go
func testRouter() *gin.Engine {
    r := gin.New()
    v1 := r.Group("/api/v1")
    {
        v1.POST("/register", Register)
        v1.POST("/login", Login)
        // ... register the endpoints you need ...
    }
    admin := v1.Group("/admin")
    admin.Use(AdminAuthMiddleware())
    {
        admin.GET("/users", ListUsers)
        // ...
    }
    r.GET("/api/v1/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })
    return r
}
```

### JWT middleware DB call expectations

`JwtAuthMiddleware` calls `isTokenRevoked()` which does a DB query
(`SELECT 1 FROM invalidated_tokens WHERE token_hash = $1`). Any test that
sends a valid JWT token on a request MUST set up this expectation, or the
middleware's DB call fails and the test gets 500 instead of reaching the
handler:

```go
mock.ExpectQuery("SELECT 1 FROM invalidated_tokens WHERE token_hash = $1").
    WillReturnRows(sqlmock.NewRows([]string{"one"}))
```

The returned rows are empty (one column named "one"), meaning "token not
revoked" — the middleware continues to the handler. Without this, the mock
has no matching expectation and `ExpectationsWereMet` fails.

**Pattern:** after every `cleanup, mock := setupTestEnv(t)`, for tests that
use `Authorization: Bearer <jwt>`, add the revocation check expectation
before any handler-specific expectations.

### Admin auth token

`AdminAuthMiddleware` reads the `ADMIN_TOKEN` env var, defaulting to
`admin-secret-token-change-in-production`. Tests hitting admin endpoints MUST
send this exact token:

```go
map[string]string{"Authorization": "Bearer admin-secret-token-change-in-production"}
```

A wrong token returns 403 (Forbidden), not 401.

### Request helper
```go
func doReq(method, path string, body interface{}, headers map[string]string, router *gin.Engine) *httptest.ResponseRecorder {
    var data []byte
    if body != nil {
        var err error
        data, err = json.Marshal(body)
        if err != nil { panic(err) }
    }
    req := httptest.NewRequest(method, path, bytes.NewReader(data))
    req.Header = http.Header{}
    for k, v := range headers {
        req.Header.Set(k, v)
    }
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)
    return rec
}
```

### Reading response JSON
```go
func jsonBody(t *testing.T, rec *httptest.ResponseRecorder, v interface{}) {
    t.Helper()
    err := json.Unmarshal(rec.Body.Bytes(), v)
    require.NoError(t, err)
}
```

### Expectation helpers
```go
// Expect a query exactly once and return the given rows.
func expectQuery(mock sqlmock.Sqlmock, query string, rows []sqlmock.Rows) {
    for _, r := range rows {
        mock.ExpectQuery(query).WillReturnRows(&r)
    }
}

// Assert no unexpected DB calls.
func expectNothing(mock sqlmock.Sqlmock, t *testing.T) {
    t.Helper()
    require.NoError(t, mock.ExpectationsWereMet())
}
```

### INSERT ... RETURNING uses ExpectQuery, NOT ExpectExec
**CRITICAL PITFALL:** Several handlers use `db.QueryRow("INSERT INTO ... RETURNING id", ...)`.
sqlmock treats this as a QUERY (because of the RETURNING clause), not an exec.
Tests that use `mock.ExpectExec("INSERT INTO ...")` for these will fail with
" Failed to create user" (500) — the mock never receives the expected call.

**Pattern to check:** grep handlers for `QueryRow.*INSERT`:
```bash
grep -n 'QueryRow.*INSERT' handlers/*.go
```
Every match needs `ExpectQuery` in the test, not `ExpectExec`.

Affected handlers in this codebase:
- `auth.go:40` — `INSERT INTO users ... RETURNING id` (Register)
- `community.go:97` — `INSERT INTO community_posts ... RETURNING id` (CreatePost)
- `community.go:244` — `INSERT INTO post_comments ... RETURNING id` (AddComment)
- `features.go:30,62` — `INSERT INTO carts ... RETURNING id` (cart create/get)

### When DB-backed vs mock
- **Mock (sqlmock):** unit tests for handler logic, coverage, CI — fast, no
  PostgreSQL needed, every query is asserted.
- **DB-backed:** integration tests that need real PostgreSQL behavior (JSON
  columns, FK cascades, triggers). Requires a running DB and cleanup. Skip with
  `t.Skip` when `database.DB == nil`.

### Handler constructors must fallback to database.DB when nil

**PITFALL — admin feature handlers created with `nil` DB by Register* functions:**

Several `Register*()` functions create handler structs with `nil` DB:
```go
// admin_features.go RegisterAdminFeaturesRoutes:
apbh := NewBulkProductHandler(nil)     // h.db is nil!
aosh := NewOrderStatusHandler(nil)     // h.db is nil!
cmh := NewCommunityModerationHandler(nil) // h.db is nil!
```

When those handlers' methods run, they call `h.db.QueryRow(...)` on a nil pointer → panic. For tests to work, ALL `New*Handler(db)` constructors must fallback to `database.DB` when nil:

```go
func NewCartHandler(db *sql.DB) *CartHandler {
    if db == nil {
        db = database.DB
    }
    return &CartHandler{db}
}
```

Apply this pattern to EVERY handler struct constructor in the project. Then in tests:
1. `setupTestEnv(t)` sets `database.DB = mockDB`
2. Tests call `Register*()` functions which pass `nil` to constructors
3. Constructors fallback to `database.DB` (the mock)
4. All handlers use the mock seamlessly

**Why this matters:** Without the fallback, admin feature handlers crash in tests even though the mock is set up correctly. The `Register*()` functions are the public API for route registration — they must work with nil DB for testability.

### Test router: manually register admin handlers, skip AdminAuthMiddleware

**PITFALL — AdminAuthMiddleware blocks test requests:**

`AdminAuthMiddleware()` checks the `ADMIN_TOKEN` env var. In tests, this env var may not be set or may differ from what the test sends. The cleanest approach for admin endpoint tests:

```go
func TR() *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.Use(gin.Recovery())
    
    v1 := r.Group("/api/v1")
    {
        // Public + JWT-protected routes via Register* functions
        handlers.RegisterProductRoutes(v1)
        handlers.RegisterOrderRoutes(v1)
        // ... other public route groups ...
    }
    
    // Admin routes: manually register with mock DB, skip middleware
    admin := v1.Group("/admin")
    {
        // Note: handlers created with nil DB → fallback to database.DB (the mock)
        admin.PUT("/orders/:id/status", handlers.NewOrderStatusHandler(nil).AdminUpdateOrderStatus)
        admin.GET("/orders", handlers.NewBulkProductHandler(nil).ListAllOrders)
        // ... register each admin endpoint manually ...
    }
    
    return r
}
```

This avoids `AdminAuthMiddleware` entirely in tests while still exercising the real handler logic against the mock DB.

### time.Time columns need time.Time values in mock rows

**PITFALL — sqlmock returns string columns, handlers scan into time.Time:**

When a handler scans a `created_at TIMESTAMPTZ` column into a `time.Time` field, the mock must provide a `time.Time` value, not a string:

```go
// WRONG — handler scans into time.Time, mock returns string → scan fails silently
mock.ExpectQuery("SELECT * FROM users").WillReturnRows(
    sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, "2024-01-01T00:00:00Z"))

// RIGHT — provide time.Time matching the scan target
mock.ExpectQuery("SELECT * FROM users").WillReturnRows(
    sqlmock.NewRows([]string{"id", "created_at"}).
    AddRow(1, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
```

If the types don't match, `Scan()` fails and the handler sees `sql.ErrNoRows` or zero values. Check each handler's scan targets against the mock row column types.

### Generating bcrypt hashes for test mock data

Test mocks that simulate login must provide a valid bcrypt hash for the password column. A fake hash like `"$2a$10$abc"` causes bcrypt comparison to fail → login returns "Invalid email or password".

Generate a real hash at test-writing time using Go's bcrypt package:

```bash
cd /root/project/backend
cat > /tmp/gen_hash.go << 'EOF'
package main
import ("fmt"; "golang.org/x/crypto/bcrypt")
func main() {
    h, _ := bcrypt.GenerateFromPassword([]byte("testpassword"), 12)
    fmt.Println(string(h))
}
EOF
go run /tmp/gen_hash.go
```

Use the output hash in the mock row for the users table. The cost parameter (12) must match what the handler uses for comparison (check `bcrypt.CompareHashAndPassword` usage in auth.go).

### sqlmock QueryMatcherEqual is the default — match exact SQL

`sqlmock.New()` defaults to `QueryMatcherEqual`, which requires the expected query string to match the actual query byte-for-byte, including `$1`/`$2` placeholders and whitespace. If the handler builds the query differently from the test expectation, the mock never receives the call and `ExpectationsWereMet` fails.

**Check exact query text:**
```bash
grep -n '"SELECT\|"INSERT\|"UPDATE\|"DELETE' handlers/*.go | grep -v '_test'
```

Copy the exact string from the handler into the test mock. Watch for:
- Multi-line backtick strings in handlers (the full string including newlines is the query)
- `fmt.Sprintf` building dynamic WHERE clauses (the mock must match the final built string, or use `QueryMatcherRegexp`)
- Whitespace differences (trailing spaces, indentation)

**When query text varies** (dynamic filters), switch to regex matching:
```go
db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
mock.ExpectQuery("SELECT COUNT\\(\\*\\).*FROM users").WillReturnRows(...)
```

### Route conflicts between public and admin groups

**PITFALL — `RegisterOrderStatusRoutes` registers `GET /orders/:id` on admin group, which conflicts with `GET /orders/:id` on the v1 public group:**

Both groups are under `/api/v1`, and Gin treats `/orders/:id` and `/admin/orders/:id` as potentially conflicting depending on registration order. The fix: make admin routes more specific:

```go
// admin_features.go — use /orders/:id/detail not /orders/:id
group.GET("/orders/:id/detail", h.AdminOrderDetail)
```

**General rule:** Admin routes should use distinct path suffixes (`/detail`, `/list`, `/check`) to avoid colliding with public routes that use the same resource ID pattern.

### Cleaning up after DB-backed tests
Every test that inserts data must clean up or subsequent tests break
(duplicate emails, duplicate slugs, FK violations):
```go
db.Exec("DELETE FROM orders WHERE id=$1", orderID)
db.Exec("DELETE FROM order_items WHERE order_id=$1", orderID)
```
Alternatively use a transaction rollback pattern: begin a transaction in `setup`,
defer `tx.Rollback()`, and pass `tx` (not `db`) to handlers under test.

### Pitfalls discovered
| Symptom | Cause | Fix |
|---------|-------|-----|
| `undefined: NewAuthHandler` | Test calls constructor that doesn't exist | Use package-level handler functions or test router |
| `imported and not used` on test helpers | Helper file has imports not all used in tests | Audit imports; remove unused ones |
| Mock expectations pass but test fails on `ExpectationsWereMet` | QueryMatcherEqual mismatched the actual query string | Check exact query text including `$N` placeholders and whitespace; switch matcher or align expectation |
| Test panics on `nil` pointer dereference in handler | Handler accesses `c.Get("user_id")` without checking ok | Always check `c.Get("user_id")` with comma-ok before casting |
| Tests pass individually but fail when run together (`go test ./...`) | Test A inserts data Test B expects to be absent | Add cleanup (DELETE) or use transaction rollback per test |
| `cannot use mock as *sqlmock.Sqlmock` | Assigned sqlmock interface value to pointer-to-interface variable | Keep mock as `sqlmock.Sqlmock` interface; do not type it to `*sqlmock.Sqlmock` |
| `INSERT ... RETURNING` handler returns 500 in test | Test used `ExpectExec` for a QueryRow INSERT | Change to `ExpectQuery` with `WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(...))` |
| Python/sed bulk-rename mangles Go backtick strings | Regex replacement across lines splits raw string literals | Avoid regex across newlines for Go source; use `go fmt` after any bulk edit, or rewrite the file cleanly rather than patching in place |
