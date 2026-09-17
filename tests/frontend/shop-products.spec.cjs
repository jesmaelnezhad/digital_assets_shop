const { test, expect } = require('@playwright/test');

const BASE = process.env.BASE_URL || 'https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir';

test.describe('Shop Products Loading', () => {
    test('shop page renders products from API', async ({ page }) => {
        const jsErrors = [];
        page.on('pageerror', err => jsErrors.push(err.message));

        await page.goto(BASE + '/', { waitUntil: 'networkidle' });

        // API client must be available
        const hasApi = await page.evaluate(() => !!window.Pawradise?.api);
        expect(hasApi, 'window.Pawradise.api should exist').toBe(true);

        // Wait for products to be rendered (Alpine.js)
        await page.waitForSelector('.product-card', { timeout: 10000 });

        const productCount = await page.locator('.product-card').count();
        expect(productCount, 'at least 1 product card should be rendered').toBeGreaterThan(0);

        // Total should be reflected in pagination
        const paginationText = await page.locator('.pagination span').first().textContent().catch(() => '');
        expect(paginationText).toContain('of');

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
