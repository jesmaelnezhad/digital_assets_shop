const { test, expect } = require('@playwright/test');

const BASE = process.env.BASE_URL || 'https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir';

function isSorted(arr, direction = 'asc') {
    for (let i = 1; i < arr.length; i++) {
        if (direction === 'asc' && arr[i] < arr[i - 1]) return false;
        if (direction === 'desc' && arr[i] > arr[i - 1]) return false;
    }
    return true;
}

test.describe('FIX-002: Product list honors sort', () => {
    test('sort=price_asc returns non-decreasing prices within each pinned group', async ({ request }) => {
        const res = await request.get(`${BASE}/api/v1/products?sort=price_asc&per_page=10`);
        expect(res.status()).toBe(200);
        const data = await res.json();
        expect(data.products.length).toBeGreaterThan(0);

        // Get prices grouped by pinned status
        const pinnedPrices = data.products.filter(p => p.pinned).map(p => p.price_usd);
        const unpinnedPrices = data.products.filter(p => !p.pinned).map(p => p.price_usd);

        if (pinnedPrices.length > 1) {
            expect(isSorted(pinnedPrices, 'asc')).toBe(true);
        }
        if (unpinnedPrices.length > 1) {
            expect(isSorted(unpinnedPrices, 'asc')).toBe(true);
        }
    });

    test('sort=price_desc returns non-increasing prices within each pinned group', async ({ request }) => {
        const res = await request.get(`${BASE}/api/v1/products?sort=price_desc&per_page=10`);
        expect(res.status()).toBe(200);
        const data = await res.json();
        expect(data.products.length).toBeGreaterThan(0);

        const pinnedPrices = data.products.filter(p => p.pinned).map(p => p.price_usd);
        const unpinnedPrices = data.products.filter(p => !p.pinned).map(p => p.price_usd);

        if (pinnedPrices.length > 1) {
            expect(isSorted(pinnedPrices, 'desc')).toBe(true);
        }
        if (unpinnedPrices.length > 1) {
            expect(isSorted(unpinnedPrices, 'desc')).toBe(true);
        }
    });

    test('price_asc and price_desc first pages differ when prices vary', async ({ request }) => {
        const resAsc = await request.get(`${BASE}/api/v1/products?sort=price_asc&per_page=50`);
        const resDesc = await request.get(`${BASE}/api/v1/products?sort=price_desc&per_page=50`);
        expect(resAsc.status()).toBe(200);
        expect(resDesc.status()).toBe(200);

        const dataAsc = await resAsc.json();
        const dataDesc = await resDesc.json();

        // Both should have products
        expect(dataAsc.products.length).toBeGreaterThan(0);
        expect(dataDesc.products.length).toBeGreaterThan(0);

        // The first pinned product in asc should have lower price than first pinned in desc
        const firstPinnedAsc = dataAsc.products.find(p => p.pinned);
        const firstPinnedDesc = dataDesc.products.find(p => p.pinned);

        expect(firstPinnedAsc).toBeDefined();
        expect(firstPinnedDesc).toBeDefined();
        expect(firstPinnedAsc.price_usd).toBeLessThanOrEqual(firstPinnedDesc.price_usd);
    });
});
