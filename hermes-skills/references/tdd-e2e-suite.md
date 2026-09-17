# TDD E2E Test Suite Pattern

The API-level E2E test suite lives in `/root/project/tests/e2e-suite.js` and runs with `npm run test:e2e` (or `node e2e-suite.js`). It uses Node.js built-in `http` module — no Playwright dependency for API tests.

## Purpose

Tests define EXPECTED behavior per `PRODUCT-SPEC.md`. In TDD fashion, they are **designed to pass only when the corresponding feature is implemented**. When the backend is unreachable, all tests fail with `ECONNREFUSED` — this is the expected "red" state.

## Test Suite Architecture

### File structure

```
/root/project/tests/
├── package.json          # {"type": "module"}, scripts.test:e2e
├── e2e-suite.js          # All 104 tests in one file
├── playwright.config.js  # (kept for UI tests if needed later)
└── node_modules/         # (playwright dep, not used by e2e-suite.js)
```

### package.json

```json
{
  "type": "module",
  "scripts": {
    "test:e2e": "node e2e-suite.js"
  }
}
```

The `"type": "module"` is required because `e2e-suite.js` uses top-level `await` and ESM imports.

### Test runner pattern

```javascript
import http from 'http';

let passed = 0, failed = 0;
const failures = [];

function request(url, opts = {}) { /* Promise-wrapped http.request */ }

async function test(name, fn) {
  try { await fn(); passed++; console.log(`  \u2713 ${name}`); }
  catch (err) { failed++; failures.push({ name, error: err.message }); console.log(`  \u2717 ${name}: ${err.message}`); }
}

// Assertion helpers
function assertStatus(res, expected, context) { /* throws if status mismatches */ }
function assertArray(data, field, context) { /* throws if field is not an array */ }
function assertField(data, field, context) { /* throws if field is null/undefined */ }
```

### Suite organization

Each suite is an `async function` that calls `test()` for each scenario:

```javascript
async function runAuthTests() {
  section('Authentication');
  await test('POST /register creates user and returns token', async () => { /* ... */ });
  // ... more tests
  return { token }; // Pass to other suites
}
```

Suites run sequentially from `main()`. Auth runs first to get a JWT token, which is then passed to order/cart/wishlist/review/community suites.

### Graceful degradation

When the auth flow fails (backend down), dependent suites must fail with a clear message instead of crashing:

```javascript
async function runOrderTests(authToken) {
  await test('POST /orders creates order from cart', async () => {
    if (!authToken) throw new Error('No auth token available');
    // ... actual test
  });
}
```

This ensures all 104 tests run and report as failures (not a fatal crash that aborts the suite).

### Environment configuration

```javascript
const ADMIN_TOKEN = process.env.ADMIN_TOKEN || 'admin-secret-token-change-in-production';
const API_HOST = process.env.API_HOST || '127.0.0.1';
const API_PORT = process.env.API_PORT || '30083';
const API_BASE = `http://${API_HOST}:${API_PORT}/api/v1`;
```

Override at runtime: `API_HOST=194.5.206.106 node e2e-suite.js`

### Expected output

```
╔══════════════════════════════════════════════════════════╗
║  Pawradise E2E Test Suite (TDD)                        ║
╚══════════════════════════════════════════════════════════╝

=== Authentication ===
  ✗ POST /register creates user and returns token: connect ECONNREFUSED 127.0.0.1:30083
  ...

--- Failures ---
1. POST /register creates user and returns token: connect ECONNREFUSED ...

╔══════════════════════════════════════════════════════════╗
║  RESULTS                                                ║
║  Passed:  0                                            ║
║  Failed:  104                                          ║
║  Total:   104                                          ║
╚══════════════════════════════════════════════════════════╝
```

## Adding a new test

1. Add a `test()` call inside the appropriate suite function
2. Use `assertStatus()`, `assertArray()`, `assertField()` for common checks
3. For auth-dependent suites, guard with `if (!token) throw new Error('No auth token available')`
4. Test the failure path explicitly (e.g., `assert(res.status === 401 || res.status === 403, ...)`)
5. For optional features, test both the presence and absence (e.g., tiers, PWYW)

## Coverage map

The 104 tests cover all sections of `PRODUCT-SPEC.md v1.1`:

| Section | Tests |
|---------|-------|
| Authentication | 11 |
| Products | 18 |
| Orders (incl. guest) | 10 |
| Cart | 4 |
| Wishlist | 3 |
| Reviews | 3 |
| Community | 9 |
| Coupons | 3 |
| Bundles | 2 |
| Referrals | 2 |
| Exchange Rates | 2 |
| Settings | 1 |
| Recently Viewed | 2 |
| Comparison | 2 |
| Recommendations | 1 |
| Admin (auth, users, products, tiers, images, previews, pin, bulk, bundles, coupons, orders, community, referrals, stats, settings, email export, exchange rates) | 23 |

## Pitfall: NOT marking tests as TODO/SKIP/pending

The suite must NEVER use `TODO`, `SKIP`, `pending`, `fit()`, `fdescribe()`, or `xtest()`. Every test is an active assertion that the product spec requirement is implemented. The only acceptable failure is the test itself failing (red in TDD).
