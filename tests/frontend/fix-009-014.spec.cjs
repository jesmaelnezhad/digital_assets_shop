const { test, expect } = require('@playwright/test');
const BASE = process.env.BASE_URL || 'https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir';

test.describe('FIX-009 to FIX-014: Verify product-service fixes', () => {
    test('FIX-009: pagination total matches filtered set', async ({ request }) => {
        const allRes = await request.get(BASE + '/api/v1/products?per_page=100');
        const allData = await allRes.json();

        const searchRes = await request.get(BASE + '/api/v1/products?search=Product%201&per_page=100');
        const searchData = await searchRes.json();

        // total for filtered should be less than unfiltered
        expect(searchData.total).toBeLessThan(allData.total);
        // total should be >= products on this page (total is full count, products is just this page)
        expect(searchData.total).toBeGreaterThanOrEqual(searchData.products.length);
    });

    test('FIX-010: pinned products appear first within sort groups', async ({ request }) => {
        const res = await request.get(BASE + '/api/v1/products?sort=price_asc&per_page=50');
        expect(res.status()).toBe(200);
        const data = await res.json();

        let foundUnpinned = false;
        for (const p of data.products) {
            if (!p.pinned) {
                foundUnpinned = true;
            } else if (foundUnpinned) {
                fail('Pinned product found after unpinned product');
            }
        }
    });

    test('FIX-011: product detail images belong to that product', async ({ request }) => {
        const listRes = await request.get(BASE + '/api/v1/products?per_page=1');
        const listData = await listRes.json();
        expect(listData.products.length).toBeGreaterThan(0);
        const slug = listData.products[0].slug;

        const detailRes = await request.get(BASE + '/api/v1/products/' + slug);
        expect(detailRes.status()).toBe(200);
        const detail = await detailRes.json();

        expect(detail).toHaveProperty('product');
        expect(detail).toHaveProperty('images');
        expect(detail.product).toBeDefined();

        for (const img of detail.images) {
            expect(img.product_id).toBe(detail.product.id);
        }
    });

    test('FIX-012: product detail includes tiers', async ({ request }) => {
        const adminToken = process.env.ADMIN_TOKEN;
        if (!adminToken) {
            console.log('Skipping: no ADMIN_TOKEN');
            return;
        }

        const listRes = await request.get(BASE + '/api/v1/products?per_page=1');
        const listData = await listRes.json();
        const product = listData.products[0];

        const authHeader = 'Bearer ' + adminToken;
        const tierRes = await request.post(BASE + '/api/v1/admin/products/' + product.id + '/tiers', {
            headers: { Authorization: authHeader },
            data: { tier_name: 'Test Tier', price_usd: 99.99 },
        });
        expect(tierRes.status()).toBe(200);

        const detailRes = await request.get(BASE + '/api/v1/products/' + product.slug);
        expect(detailRes.status()).toBe(200);
        const detail = await detailRes.json();

        expect(detail).toHaveProperty('tiers');
        expect(detail.tiers).toBeDefined();

        const foundTier = detail.tiers.find((t) => t.tier_name === 'Test Tier');
        expect(foundTier).toBeDefined();
        expect(foundTier.price_usd).toBeCloseTo(99.99, 2);
    });

    test('FIX-013: product detail includes PWYW config', async ({ request }) => {
        const listRes = await request.get(BASE + '/api/v1/products?per_page=1');
        const listData = await listRes.json();
        const slug = listData.products[0].slug;

        const detailRes = await request.get(BASE + '/api/v1/products/' + slug);
        expect(detailRes.status()).toBe(200);
        const detail = await detailRes.json();

        expect(detail.product).toHaveProperty('is_pwyw');
        expect(detail.product).toHaveProperty('pwyw_min_price');
    });

    test('FIX-014: recommendations returns JSON not HTML', async ({ request }) => {
        const res = await request.get(BASE + '/api/v1/recommendations/1');
        expect(res.status()).toBe(200);

        const contentType = res.headers()['content-type'];
        expect(contentType).toContain('application/json');

        const data = await res.json();
        expect(data).toHaveProperty('products');
        expect(Array.isArray(data.products)).toBe(true);
    });

    test('FIX-014: recommendations for missing product returns 404', async ({ request }) => {
        const res = await request.get(BASE + '/api/v1/recommendations/999999999');
        expect(res.status()).toBe(404);

        const contentType = res.headers()['content-type'];
        expect(contentType).toContain('application/json');
    });
});
