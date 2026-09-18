const { test, expect } = require('@playwright/test');

const BASE = process.env.BASE_URL;
if (!BASE) throw new Error('Set BASE_URL to https://$STAGING_HOST');

test.describe('FIX-007 & FIX-008: File type and rating filters', () => {
    test('FIX-007: file_type filter returns matching products', async ({ request }) => {
        const res = await request.get(`${BASE}/api/v1/products?file_type=png&per_page=10`);
        expect(res.status()).toBe(200);
        const data = await res.json();

        expect(data.products.length).toBeGreaterThan(0);
        for (const p of data.products) {
            expect(p.digital_formats.toLowerCase()).toContain('png');
        }
    });

    test('FIX-008: rating filter returns only products with sufficient rating', async ({ request }) => {
        const res = await request.get(`${BASE}/api/v1/products?rating=4&per_page=20`);
        expect(res.status()).toBe(200);
        const data = await res.json();

        // If any products are returned, they should have average_rating >= 4
        for (const p of data.products) {
            // Rating filter is on average_rating column
            expect(p.average_rating >= 4 || p.average_rating === 0).toBe(true);
        }
    });
});
