# Feature-by-Feature E2E Verification Pattern

When verifying a deployed ecommerce backend on staging, test each feature
endpoint individually and report a structured PASS/FAIL verdict with the
next action to take. This pattern was developed during Pawradise staging
verification and is more actionable than a single monolithic test script.

## Output format

```
=== FEATURE VERIFICATION ON STAGING ===

[PASS] Auth: Register + Login — token obtained
[PASS] Auth: GetMe — returns user object
[PASS] Products: List — 20+ products, 8 categories
[PASS] Products: Detail by slug — returns product data
[FAIL] Cart: Create — 401 Unauthorized
       → Next: add JWT middleware to cart routes in main.go
[FAIL] Wishlist: Toggle — 401 Unauthorized
       → Next: add JWT middleware to wishlist routes in main.go
[FAIL] Reviews: Create — 401 Unauthorized
       → Next: add JWT middleware to review routes in main.go
[FAIL] Recently-viewed: Track — 401 Unauthorized
       → Next: add JWT middleware to recently-viewed routes in main.go
[FAIL] Compare: Add — 401 Unauthorized
       → Next: add JWT middleware to compare routes in main.go
[FAIL] Recommendations: Get — 500 Internal Server Error
       → Next: check backend pod logs for nil pointer / query error
[FAIL] Guest orders: Create — 404 Not Found
       → Next: verify RegisterGuestOrderRoutes is called on public v1
[FAIL] Exchange rates: List — 404 Not Found
       → Next: verify RegisterExchangeRateRoutes is on public v1 (not just admin)
[FAIL] Orders: Create — 400 Bad Request
       → Next: check request body format against handler expectations
[FAIL] Admin: Login — 401 Unauthorized
       → Next: verify admin token matches cluster secret

Summary: 5/17 passing, 12 failing
```

## Why this format works

1. **Grouped by feature** — each line is one feature area, not one HTTP call.
   Cart create, cart get, cart list are one line.
2. **Verdict + HTTP code** — you see at a glance whether it's a 401 (auth
   issue), 404 (route missing), 500 (bug), or 400 (request format).
3. **Next action** — each failure line includes a concrete next step, not
   just "it failed". This turns verification into a to-do list.
4. **Token is verified separately** — the first failure in the auth chain
   (register/login) blocks everything downstream; report it first.

## When to use

- After deploying new code to staging
- After adding a new feature handler
- Before promoting staging → production
- When a user reports a feature is broken on staging

## What to test (minimum set)

| Feature | Minimum checks |
|---------|---------------|
| Auth | Register, Login, GetMe, Logout |
| Products | List, Detail by slug, Categories |
| Cart | Create/get cart, Add item, Remove item |
| Wishlist | Toggle item |
| Reviews | Create review, List reviews for product |
| Recently-viewed | Track product view, List recently-viewed |
| Compare | Add product, Remove product, List compare |
| Recommendations | Get recommendations for a user |
| Guest orders | Create guest order, Confirm guest order |
| Exchange rates | List rates, Set rate (admin) |
| Orders | Create order (registered user), Get order |
| Admin | Login, List users, Create product, Update order status |
| Community | Feed, Create post, Like, Comment, Follow |

## Test data management

- Use a unique test email per run (timestamp suffix) to avoid collisions
- Clean up test data after: delete test user via admin API if available
- Don't rely on specific product IDs — query the list first, use a real ID
- Test with both authenticated and unauthenticated requests to verify
  auth boundaries

## Scripts vs manual checks

- **Automated:** write a Python script (see `references/staging-e2e.md`) that
  exercises the full flow and exits non-zero on any failure. Run it after
every deploy.
- **Manual:** for visual/UI features (frontend pages, admin panel rendering),
  use `curl` to fetch the HTML and check for key elements, or use the browser
  tool to navigate.
