# Pawradise Frontend Testing — Findings & Rules

**Date:** 2026-09-13  
**Context:** After fixing the Pawradise staging site, we learned critical lessons about frontend testing that are NOT obvious from API-only testing.

---

## The Core Mistake

We had 104 API tests passing while the site was visually broken. Users saw:
- White pages with no styling
- nginx welcome pages
- Buttons linking to 404s
- Empty product grids

**The fundamental problem:** Checking HTTP 200 + API response shape ≠ checking what the user sees.

---

## Rules for Frontend Testing

### Rule 1: Always Verify Computed Styles, Not Just HTML Presence

```javascript
// ❌ WRONG: This passes even if CSS fails to load
const hasStyleTag = await page.locator('style').count() > 0;

// ✅ RIGHT: Check what the browser actually renders
const bg = await page.evaluate(() => 
  window.getComputedStyle(document.body).backgroundColor
);
expect(bg).toBe('rgb(10, 14, 20)');
```

**Why:** A page can have `<style>` tags and CSS references but still render unstyled if:
- CSS file fails to load (404, timeout)
- CSS custom properties are defined but never applied to elements
- Inline styles use variables that resolve to nothing

### Rule 2: Check CSS Custom Properties Resolve

```javascript
// ❌ WRONG: Variable exists in :root
const hasVars = await page.evaluate(() => 
  document.documentElement.innerHTML.includes('--color-bg')
);

// ✅ RIGHT: Variable actually resolves
const value = await page.evaluate(() => 
  getComputedStyle(document.documentElement)
    .getPropertyValue('--color-bg').trim()
);
expect(value).toBe('#0a0e14');
```

**Why:** A CSS variable can be defined but unused, or overridden by a failed load.

### Rule 3: Verify API Calls Match the JS Client

```javascript
// ❌ WRONG: Test passes if API returns 200
const res = await request.get(BASE + '/api/v1/products');
expect(res.status()).toBe(200);

// ✅ RIGHT: Verify the EXACT path the frontend calls
// Check api.js: api.products.list() calls '/products?params'
// NOT '/api/v1/products' directly
const products = await page.evaluate(() => {
  return window.Pawradise.api.products.list({ per_page: 2 });
});
expect(products.products.length).toBeGreaterThan(0);
```

**Why:** The frontend may use different URL paths than the API tests. In our case:
- API test: `GET /api/v1/products` ✅
- Frontend: `api.get('/products?...')` ❌ (wrong method path)

### Rule 4: Check Every `<a>` Tag Resolves

```javascript
// ❌ WRONG: Check a few links manually
await page.goto(BASE + '/login');

// ✅ RIGHT: Crawl ALL links on EVERY page
const links = await page.locator('a[href^="/"]').evaluateAll(els =>
  els.map(e => e.getAttribute('href'))
);
for (const link of links) {
  const res = await page.request.head(BASE + link);
  expect(res.status()).toBe(200);
}
```

**Why:** Broken links don't show up in API tests. We had:
- `/auth/login.html` → 404 (should be `/login`)
- `/product/request` → served wrong content

### Rule 5: Test Navigation Consistency

```javascript
// ✅ Check that all pages have the same nav structure
const navItems = {};
for (const url of pages) {
  await page.goto(BASE + url);
  navItems[url] = await page.evaluate(() => {
    const nav = document.querySelector('nav:not(.sidebar-nav)');
    return Array.from(nav.querySelectorAll('a')).map(a => a.textContent);
  });
}
// All pages should match (except admin)
```

**Why:** In microfrontend architectures, each MFE can have different nav. Users expect consistency.

### Rule 6: Check JavaScript Console Errors

```javascript
const errors = [];
page.on('console', msg => {
  if (msg.type() === 'error') errors.push(msg.text());
});
page.on('pageerror', err => errors.push(err.message));
// ... navigate ...
expect(errors).toHaveLength(0);
```

**Why:** "Unexpected token 'export'" errors from ES module mismatches break functionality silently.

### Rule 7: Verify Content Renders After JS Execution

```javascript
// ❌ WRONG: Check HTML immediately
await page.goto(BASE);
const cards = await page.locator('.product-card').count();

// ✅ RIGHT: Wait for JS to render
await page.goto(BASE, { waitUntil: 'networkidle' });
await page.waitForTimeout(3000); // Wait for Alpine.js + API
const cards = await page.locator('.product-card').count();
```

**Why:** SPAs (Alpine.js, React) render content AFTER page load. HTML shows empty containers.

### Rule 8: Check Both Internal AND External Link Resolution

```javascript
// Internal links
const internal = await page.locator('a[href^="/"]').evaluateAll(...);
for (const link of internal) {
  expect((await page.request.head(BASE + link)).status()).toBe(200);
}

// External links (nginx.com, etc.) should at least not 503
const external = await page.locator('a[href^="http"]').evaluateAll(...);
for (const link of external) {
  const res = await page.request.head(link).catch(() => null);
  expect(res?.status()).not.toBe(503);
}
```

### Rule 9: Verify Form Fields by ID and Name

```javascript
// ❌ WRONG: Test assumes name attribute
await page.locator('input[name="email"]')

// ✅ RIGHT: Check both id and name
const email = page.locator('#email, input[name="email"]');
expect(await email.count()).toBeGreaterThan(0);
```

**Why:** Developers use `id` or `name` inconsistently. Our register form used `id="name"` not `name="name"`.

### Rule 10: Test With Real Browser, Not Just curl

```bash
# ❌ WRONG: curl only checks HTTP
curl -s -o /dev/null -w "%{http_code}" $URL

# ✅ RIGHT: Playwright renders like a real user
npx playwright test
```

**Why:** curl can't execute JavaScript, apply CSS, or detect console errors.

---

## Test File Organization

**Rule:** Organize by WHAT you test, not by WHEN you found the bug.

```
tests/frontend/
├── playwright.config.cjs
├── pawradise-tests.spec.cjs
└── shop-products.spec.cjs
tests/e2e-suite.js
tests/unit/<service>/
tests/integration/
```

---

## Common Pitfalls Found

| Pitfall | What Happened | How We Caught It |
|---------|---------------|------------------|
| CSS variables defined but never applied | `body { font-family: var(--font-sans) }` with no `body` rule | Computed style check |
| Wrong API call format | `api.get('/products')` vs `api.products.list()` | `window.Pawradise.api` check |
| nginx alias serving wrong root | `/product/bundle.html` served from `/product/product/` | HEAD request on all links |
| Service selector mismatch | `namespace: staging` in selector broke endpoints | Endpoint check |
| Stale pod images | New code not deployed after `kubectl apply` | Forced pod deletion |
| Alpine.js version mismatch | Tests used `__x`, app uses Proxy v3 | Updated test approach |

---

## The Testing Workflow That Works

1. **Write failing test** that captures the user-reported issue
2. **Verify it fails** (confirm the bug exists)
3. **Fix the frontend** code/config
4. **Rebuild + redeploy** Docker images
5. **Force pod restart** (don't trust rolling updates)
6. **Verify test passes**
7. **Run full suite** to check for regressions

---

## Rule 11: Check for Reverse Proxy / CDN Caching

**CRITICAL:** A host-level nginx or CDN may intercept static asset paths (like `/assets/`) and serve stale files for ALL domains — even staging.

**Symptom:** Pod has correct file, but curl serves wrong content (different file size/hash).

**Diagnose:**
```bash
# Compare pod file vs served file
kubectl exec -n staging deploy/<app> -- md5sum /path/to/file
curl -s https://<domain>/path/to/file | md5sum
```

**Fix:** Split reverse proxy config per domain — staging should proxy ALL traffic to k3s ingress without asset interception.

## Rule 12: Test in Production-Like Conditions

Test against the actual staging domain, not localhost. CDN, SSL termination, and reverse proxy behavior can differ.
