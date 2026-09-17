# Test Compilation: Go Module Cache & Package Mismatch

## Symptom

`go test ./unit/...` fails with:
```
found packages identity (auth_test.go) and identity_test (test_helper.go) in /root/project/tests/unit/identity-service
```

Even though `head -1` on both files shows `package identity_test`.

## Root Cause

Go's build cache (`$GOCACHE`) can retain stale package identity information from a previous compilation where the package name was different. When a file's package declaration is changed (e.g., from `package identity` to `package identity_test`), the cache may still associate the old package name with the file content hash.

## Diagnosis

1. Verify actual file content:
   ```bash
   head -1 auth_test.go        # Shows: package identity_test
   head -1 test_helper.go      # Shows: package identity_test
   xxd auth_test.go | head -2   # Confirm no hidden bytes
   ```

2. Check for stale cache:
   ```bash
   go env GOCACHE              # Usually ~/.cache/go-build
   ```

## Fix

```bash
# 1. Clear ALL Go caches
clean -cache -testcache -modcache 2>/dev/null

# 2. Verify go.mod has no stale replace directives
#    (e.g., 'replace backend => ../backend' when backend/ no longer exists)
cat go.mod

# 3. Regenerate dependencies
GOPROXY=https://proxy.golang.org,direct go mod tidy

# 4. Force rebuild with -a flag
go test -a -v -count=1 ./unit/identity-service/...
```

## Prevention

- After changing a package declaration, always run `go clean -cache` before testing
- When renaming packages across multiple files, do it in one batch then clear cache
- If tests fail with "found packages X and Y" but `head -1` shows they match, it's ALWAYS a cache issue

## Alternative: Stale Replace Directive

If `tests/go.mod` has `replace backend => ../backend` but `../backend/` was refactored into `services/*/`, the replace target no longer exists. Remove or update the replace directive:

```go
// Remove this line if backend/ no longer exists:
replace backend => ../backend
```

Then run `go mod tidy`.

## Verification

```bash
# Should compile and run (tests may FAIL, but compilation should succeed)
go test -v -count=1 ./unit/identity-service/... 2>&1 | head -20

# Expected: test output showing PASS/FAIL, NOT "found packages" error
```
