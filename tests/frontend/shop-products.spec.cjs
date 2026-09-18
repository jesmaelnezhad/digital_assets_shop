const { test, expect } = require('@playwright/test');

const BASE = process.env.BASE_URL;
if (!BASE) throw new Error('Set BASE_URL to https://$STAGING_HOST');

test.describe('Shop Products Loading', () => {
    test('shop page renders products from API', async ({ page }) => {
        const jsErrors = [];
        page.on('pageerror', err => jsErrors.push(err.message));

        await page.goto(BASE + '/', { waitUntil: 'networkidle' });

        // API client must be available
        const hasApi = await page.evaluate(() => !!window.Store4bots?.api);
        expect(hasApi, 'window.Store4bots.api should exist').toBe(true);

        await page.waitForSelector('.card, .hero', { timeout: 10000 });

        const productCount = await page.locator('.card').count();
        expect(productCount + (await page.locator('.hero').count()), 'catalog or hero should render').toBeGreaterThan(0);

        // No JS errors (image load failures from placeholder URLs are expected)
        expect(jsErrors, `JS errors: ${jsErrors.join('; ')}`).toHaveLength(0);
    });

    test('shop page API call returns products', async ({ page }) => {
        const [response] = await Promise.all([
            page.waitForResponse(r => r.url().includes('/api/v1/products') && r.status() === 200, { timeout: 10000 }),
            page.goto(BASE + '/', { waitUntil: 'networkidle' }),
        ]);

        const data = await response.json();
        expect(data).toHaveProperty('products');
        expect(data.products.length).toBeGreaterThan(0);
    });
});
