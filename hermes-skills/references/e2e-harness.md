# E2E Harness (curl-based, BLUE → RED)

Suites: `/root/project/e2e/` — `common.sh` (harness), `customer-journeys.sh`
(C01–C20), `admin-ops.sh` (A01–A20), `link-audit.sh`. Run staging first,
then prod for parity: `./customer-journeys.sh staging && ./admin-ops.sh staging`.
After a run, delete test users via `DELETE /admin/users/:id` on both envs
(community posts cascade; product/rates/settings tests restore their values
in-test). Register returns **201**, not 200. `/me` returns the user object
directly (`{"id":..,"email":..}`, no `user` wrapper).

## Harness pattern (common.sh)

Functions that set globals MUST be called directly — `code=$(http_code …)`
runs in a subshell and the `BODY` assignment is lost. Pattern:

```bash
HTTP_CODE=""; BODY=""
http_code() {
  local method="$1" url="$2" data="${3:-}" token="${4:-}"  # ${3:-} survives set -u
  local args=(-s --max-time 10 -o /tmp/e2e_body.txt -w "%{http_code}" -X "$method" "$url")
  [ -n "$data" ] && args+=(-H "Content-Type: application/json" -d "$data")
  [ -n "$token" ] && args+=(-H "Authorization: Bearer $token")
  HTTP_CODE=$(curl "${args[@]}")
  BODY=$(cat /tmp/e2e_body.txt)
}
# call site:
http_code POST "$API/orders" '{"items":[{"product_id":1,"quantity":1}]}' "$TOKEN"
code="$HTTP_CODE"
```

## Shell pitfalls hit in-session (do not repeat)

1. **Never name a variable `UID`.** It is a readonly bash builtin (current
   uid, usually 0). `UID=$(…)` prints `UID: readonly variable` to stderr,
   the assignment fails, and `$UID` keeps expanding to `0` — the suite then
   resets/deletes "user 0" and every assertion misleads. Use `TUID`/`USER_ID`.
2. **`set -u` + missing `$3`/`$4`.** Referencing unset positionals aborts the
   function. Always declare `local data="${3:-}" token="${4:-}"`.
3. **`assert_contains` uses `grep -qF` (fixed strings).** Alternations like
   `"a\|b"` match literally — write one assertion per needle.
4. **JSON spacing.** Encode compactly in assertions: `"id":$OID`, not
   `"id": $OID` (Go marshals without spaces).

## link-audit.sh regex

Only match real root-relative attributes; JS string-concat noise
(`href="'+LINK_BASE+'/account"`, `src="'+x+'"`) must not match:

```bash
# staging page must have NO root-relative link outside /staging:
grep -oE '(href|src|action)="/[^"]*"' page | grep -vE '(href|src|action)="/staging'
# prod page must have NO /staging link:
grep -oE '(href|src|action)="/staging[^"]*"' page
```

Also asserts product pages serve 200 on both envs and `#authBtn` +
`data-link="/community"` exist on shop pages. Dynamically rendered links
(JS innerHTML after load) must use `${LINK_BASE}` inline — the one-time
`data-link` rewrite only covers markup present at page load.
