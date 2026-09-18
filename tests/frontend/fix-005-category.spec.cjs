const { test, expect } = require('@playwright/test');

const BASE = process.env.BASE_URL;
if (!BASE) throw new Error('Set BASE_URL to https://$STAGING_HOST');

test.describe('FIX-005: Product list honors category filter', () => {
    test('category=<slug> returns only products in that category', async ({ request }) => {
        // Get categories
        const catsRes = await request.get(`${BASE}/api/v1/categories`);
        expect(catsRes.status()).toBe(200);
        const catsData = await catsRes.json();
        expect(catsData.categories.length).toBeGreaterThan(0);

        const slug = catsData.categories[0].slug;

        const res = await request.get(`${BASE}/api/v1/products?category=${slug}&per_page=50`);
        expect(res.status()).toBe(200);
        const data = await res.json();

        // All returned products should belong to this category
        expect(data.products.length).toBeGreaterThan(0);
        for (const p of data.products) {
            expect(p.category_id).toBe(catsData.categories[0].id);
        }
        expect(data.total).toBe(data.products.length);
    });

    test('category with no products returns empty + total: 0', async ({ request }) => {
        const res = await request.get(`${BASE}/api/v1/products?category=this-slug-does-not-exist-xyz&per_page=10`);
        expect(res.status()).toBe(200);
        const data = await res.json();

        expect(data).toHaveProperty('products');
        expect(data).toHaveProperty('total');
        expect(data.total).toBe(0);
        expect(data.products).toHaveLength(0);
    });
});
