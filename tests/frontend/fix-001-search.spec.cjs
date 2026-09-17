const { test, expect } = require('@playwright/test');

const BASE = process.env.BASE_URL || 'https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir';

test.describe('FIX-001: Product list honors search', () => {
    test('search matches a specific product title', async ({ request }) => {
        // Get a real product title from the list
        const allRes = await request.get(`${BASE}/api/v1/products?per_page=5`);
        expect(allRes.status()).toBe(200);
        const allData = await allRes.json();
        expect(allData.products.length).toBeGreaterThan(0);

        const target = allData.products[0];
        const searchTerm = target.title.substring(0, Math.min(10, target.title.length));

        // Search for that title
        const searchRes = await request.get(`${BASE}/api/v1/products?search=${encodeURIComponent(searchTerm)}`);
        expect(searchRes.status()).toBe(200);
        const searchData = await searchRes.json();

        // Should return fewer results than the full list
        expect(searchData.total).toBeLessThan(allData.total);
        expect(searchData.products.length).toBeGreaterThan(0);

        // The target product should be in results
        const foundIds = searchData.products.map((p) => p.id);
        expect(foundIds).toContain(target.id);
    });

    test('search for nonsense returns empty products and total 0', async ({ request }) => {
        const res = await request.get(`${BASE}/api/v1/products?search=zzzznonexistentzzz`);
        expect(res.status()).toBe(200);
        const data = await res.json();

        expect(data).toHaveProperty('products');
        expect(data).toHaveProperty('total');
        expect(data.products).toHaveLength(0);
        expect(data.total).toBe(0);
    });
});
