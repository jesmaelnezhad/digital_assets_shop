const { test, expect } = require('@playwright/test');

const BASE = process.env.BASE_URL;
if (!BASE) throw new Error('Set BASE_URL to https://$STAGING_HOST');

test.describe('FIX-006: Product list honors price range', () => {
    test('price_min and price_max filter results', async ({ request }) => {
        const res = await request.get(`${BASE}/api/v1/products?price_min=50&price_max=100&per_page=50`);
        expect(res.status()).toBe(200);
        const data = await res.json();

        expect(data.products.length).toBeGreaterThan(0);
        for (const p of data.products) {
            expect(p.price_usd).toBeGreaterThanOrEqual(50);
            expect(p.price_usd).toBeLessThanOrEqual(100);
        }
        // total should be >= products.length (total is the full filtered count, products is just the page)
        expect(data.total).toBeGreaterThanOrEqual(data.products.length);
    });

    test('price range with no matches returns empty', async ({ request }) => {
        const res = await request.get(`${BASE}/api/v1/products?price_min=10000&price_max=20000&per_page=10`);
        expect(res.status()).toBe(200);
        const data = await res.json();

        expect(data.total).toBe(0);
        expect(data.products).toHaveLength(0);
    });
});
