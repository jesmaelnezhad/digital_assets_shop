# Pawradise — Test-Driven Implementation Guide

> Phase 3+. This turns `02-UX-FLOWS.md` into failing tests *before* any MFE
> markup exists, so the agent has a concrete, checkable definition of "done"
> for every page, and can't silently drift from the spec.

---

## 1. Stack

| Layer | Tool | Why |
|---|---|---|
| Backend unit | Go `testing` + `sqlmock` | already the convention (`MICROSERVICE-ARCHITECTURE.md` §7) — keep it |
| Backend integration | Go `testing` + `testcontainers` | same |
| Frontend E2E | **Playwright** | drives a real browser through full-page navigations — a natural fit for the MPA architecture (ADR-1); one test file can cross MFE boundaries by just navigating, no special tooling needed |
| Frontend unit | **Vitest** | for the handful of pure-logic modules (`shared/lib/api.js`, currency formatting, coupon math mirror) — no bundler required, runs `.js` directly |
| Existing E2E | `tests/e2e/e2e-suite.js` | keep as regression coverage; don't rewrite it, but don't add new flows there — new flows go into Playwright per this doc |

Install once, at repo root:
```bash
npm init -y
npm install -D @playwright/test vitest
npx playwright install --with-deps chromium
```

## 2. Workflow (Red → Green → Refactor), per page

1. Open `02-UX-FLOWS.md`, find the flow you're building.
2. Write a Playwright spec in `tests/frontend-e2e/<flow-name>.spec.ts`
   encoding every line under "Success criteria" and "Empty/error states" as
   an assertion. It will fail — the page/route doesn't exist yet. That's
   correct.
3. Build the minimum HTML + Alpine.js needed to pass, using the tokens from
   `03-DESIGN-SYSTEM.md` (don't hand-roll colors/spacing).
4. Run the test. Iterate until green.
5. If you extracted any pure logic (e.g. computing bundle savings, coupon
   discount, cart total) into a `shared/lib/*.js` module, write a Vitest
   spec for it — these should be table-driven (multiple input/output pairs
   in one test), not one test per case.
6. Re-run the full suite for that MFE before moving to the next one.

Never write implementation before its test exists for that specific
success criterion. If a criterion is genuinely hard to assert
automatically (e.g. "share link opens the right platform's share dialog"),
write the closest automatable proxy (e.g. assert the `href`/intent URL is
correct) rather than skipping it.

## 3. Directory & config

```
tests/
  frontend-e2e/
    playwright.config.ts
    auth.spec.ts
    shop-browse.spec.ts
    product-detail.spec.ts
    cart-wishlist.spec.ts
    checkout.spec.ts
    checkout-guest.spec.ts
    community.spec.ts
    seo.spec.ts
  frontend-unit/
    api-client.test.ts
    currency.test.ts
    coupon-math.test.ts
```

```ts
// playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: '.',
  fullyParallel: true,
  use: { baseURL: process.env.BASE_URL ?? 'http://localhost:30758' },
  projects: [
    { name: 'mobile',  use: { ...devices['iPhone 13'] } },      // 390x844
    { name: 'desktop', use: { ...devices['Desktop Chrome'] } }, // 1280x720
  ],
});
```
Running against both `projects` automatically gives you the responsive
pass/fail from `02-UX-FLOWS.md` §"Responsive at every flow" for free — every
spec below runs twice, once per viewport, with zero extra code.

## 4. Example: checkout with coupon (Playwright)

```ts
// tests/frontend-e2e/checkout.spec.ts
import { test, expect } from '@playwright/test';
import { loginAsTestUser, seedCartWithProduct } from './helpers/fixtures';

test.describe('checkout with coupon', () => {
  test.beforeEach(async ({ page, request }) => {
    await loginAsTestUser(page, request);
    await seedCartWithProduct(request, { slug: 'sample-icon-pack', qty: 1 });
  });

  test('applies a valid coupon and updates the total', async ({ page }) => {
    await page.goto('/checkout');
    const subtotal = await page.getByTestId('checkout-subtotal').innerText();

    await page.getByLabel('Coupon code').fill('SAVE10');
    await page.getByRole('button', { name: 'Apply' }).click();

    await expect(page.getByTestId('checkout-discount')).toBeVisible();
    const total = await page.getByTestId('checkout-total').innerText();
    expect(total).not.toBe(subtotal);
  });

  test('rejects an expired coupon with a specific reason', async ({ page }) => {
    await page.goto('/checkout');
    await page.getByLabel('Coupon code').fill('EXPIRED2024');
    await page.getByRole('button', { name: 'Apply' }).click();

    await expect(page.getByTestId('coupon-error')).toContainText(/expired/i);
  });

  test('requires an explicit confirm step before charging', async ({ page }) => {
    await page.goto('/checkout');
    await page.getByRole('button', { name: 'Pay' }).click();

    const modal = page.getByRole('dialog', { name: /confirm/i });
    await expect(modal).toBeVisible();
    // The order must not be created yet just from clicking "Pay".
    await expect(page.getByTestId('order-status')).toHaveCount(0);

    await modal.getByRole('button', { name: /confirm/i }).click();
    await expect(page.getByTestId('order-status')).toContainText(/pending|processing/i);
  });
});
```

## 5. Example: pure-logic unit test (Vitest)

```ts
// tests/frontend-unit/coupon-math.test.ts
import { describe, it, expect } from 'vitest';
import { applyCoupon } from '../../shared/lib/coupon-math.js';

describe('applyCoupon', () => {
  it.each([
    // [subtotal, coupon, expectedTotal]
    [100, { type: 'percentage', value: 10 }, 90],
    [100, { type: 'fixed', value: 15 }, 85],
    [10,  { type: 'fixed', value: 50 }, 0],       // never goes negative
  ])('subtotal %d with %o -> %d', (subtotal, coupon, expected) => {
    expect(applyCoupon(subtotal, coupon)).toBe(expected);
  });
});
```

This mirrors (does not replace) server-side validation — the server is
always the authority; this test just protects the checkout UI's live total
preview from drifting out of sync with the pricing rules.

## 6. Example: shared api.js wrapper (Vitest, mocked fetch)

```ts
// tests/frontend-unit/api-client.test.ts
import { describe, it, expect, vi } from 'vitest';
import { apiFetch } from '../../shared/lib/api.js';

describe('apiFetch', () => {
  it('includes credentials and the staging-aware base path', async () => {
    globalThis.window = { __PAWRADISE_ENV__: { apiBase: '/staging/api/v1' } };
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, json: async () => ({}) });
    globalThis.fetch = fetchMock;

    await apiFetch('/products');

    expect(fetchMock).toHaveBeenCalledWith(
      '/staging/api/v1/products',
      expect.objectContaining({ credentials: 'include' })
    );
  });

  it('normalizes a non-2xx response into a thrown ApiError with the server message', async () => {
    globalThis.window = { __PAWRADISE_ENV__: { apiBase: '/api/v1' } };
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false, status: 409, json: async () => ({ error: 'coupon expired' }),
    });

    await expect(apiFetch('/coupons/validate', { method: 'POST' }))
      .rejects.toMatchObject({ status: 409, message: 'coupon expired' });
  });
});
```

## 7. CI / running locally

Add to root `package.json`:
```json
{
  "scripts": {
    "test:unit": "vitest run tests/frontend-unit",
    "test:e2e": "playwright test tests/frontend-e2e",
    "test:e2e:staging": "BASE_URL=https://pawradise.ir/staging playwright test tests/frontend-e2e",
    "test": "npm run test:unit && npm run test:e2e"
  }
}
```
No CI infra is described in the existing docs, so `npm test` run manually
on BLUE before pushing images is the current bar — keep it green before
every push to RED's registry.
