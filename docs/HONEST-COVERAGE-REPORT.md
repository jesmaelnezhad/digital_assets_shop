# Honest Spec Coverage Report

**Date:** 2026-09-11 20:35  
**Staging:** https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir  
**Spec Document:** /root/project/docs/PRODUCT-SPEC.md  

## Final Results

| Metric | Value |
|--------|-------|
| **Passed** | 44 |
| **Failed** | 6 |
| **Coverage** | **88.0%** |
| **Critical Failures** | 2 (nginx welcome pages) |

---

## Failed Tests

| # | Test | Severity | Root Cause |
|---|------|----------|------------|
| 1 | **Bundle detail loads** | 🔴 Critical | Ingress has no route for `/product/bundle.html` — falls through to nginx welcome page |
| 2 | **Request page loads** | 🔴 Critical | Ingress has no route for `/product/request.html` — falls through to nginx welcome page |
| 3 | **Sort controls** | 🟡 Medium | Home page has no `.sort` class element |
| 4 | **Register name field** | 🟡 Medium | Register form has no `input[name="name"]` or `input[name="first_name"]` |
| 5 | **Product price** | 🟡 Medium | Product detail has no `.price` class element |
| 6 | **Community feed** | 🟡 Medium | Community page has no `.feed` or `.community-feed` class |

---

## Root Cause Analysis

### 1. Ingress Sub-Path Routing (2 failures)

The frontend ingress only routes top-level paths to MFE pods:
- `/product` → product-mfe pod (serves `product.html`)
- But `/product/bundle.html` and `/product/request.html` are NOT in the ingress rules

When a user clicks a link to `/product/bundle.html`, the request:
1. Hits the ingress controller
2. No matching rule for `/product/bundle.html`
3. Falls through to the default backend (nginx welcome page)

**Fix:** Either:
- Add explicit ingress paths for all sub-paths (`/product/bundle.html`, `/product/request.html`)
- OR use query parameters: `/product?page=bundle`
- OR use try_files in nginx to fall back to `index.html`

### 2. Missing CSS Classes (4 failures)

The MFE HTML files use different class names than what the tests expect:

| Expected | Actual (from HTML inspection) |
|----------|-------------------------------|
| `.sort` | Not found in shop-mfe |
| `input[name="name"]` | Register uses different field name |
| `.price` | Product detail uses different class |
| `.feed` / `.community-feed` | Community uses `.post-feed` or similar |

**Fix:** Update MFE HTML to use consistent class names, OR update tests to match actual names.

---

## What Was Tested (44 passed)

### Public Pages (13/15)
- ✅ Home, Login, Register, Product detail, Community, Post detail, Profile
- ✅ Account, Cart, Wishlist, Referrals, Checkout, Admin
- ❌ Bundle detail (nginx welcome)
- ❌ Request page (nginx welcome)

### Page Content (15/20)
- ✅ Product grid, Product cards/empty, Search bar, Filter controls
- ✅ Login form (email, password, submit)
- ✅ Register form (email, password, submit)
- ✅ Product layout, Product gallery, Product description
- ✅ Create post form, Posts/empty state
- ❌ Sort controls, Register name field, Product price, Community feed

### Design (3/3)
- ✅ Dark theme, Header, Footer

### API Endpoints (10/10)
- ✅ Products, Categories, Bundles, Posts, Exchange rates, Health
- ✅ Admin users, products, stats, settings

### E2E Flows (2/2)
- ✅ Registration, Login

---

## Testing Rule (New)

**Never trust HTTP 200 alone.** Always verify:
1. Response body does NOT contain "Welcome to nginx"
2. Expected CSS classes/text exist in rendered HTML
3. For SPAs: use a real browser that executes JavaScript
4. Test sub-paths explicitly (`/product/bundle.html`, not just `/product`)
5. Use Playwright with `waitUntil: 'networkidle'` + `waitForTimeout` for JS rendering

---

## Next Steps (NOT started — waiting for direction)

1. **Fix ingress routing** for sub-paths (bundle.html, request.html)
2. **Fix MFE class names** to match expected structure
3. **Seed admin DB** with admin user for admin panel access
4. **Update tests** to match actual HTML structure where appropriate
5. **Re-run** until 100% coverage

**Status: Tests written and run. Awaiting direction to fix failures.**

