const { test, expect } = require('@playwright/test');

const BASE = process.env.BASE_URL;
if (!BASE) throw new Error('Set BASE_URL to https://$STAGING_HOST');

test.describe('FIX-004: Product list honors sort=popular', () => {
    test('sort=popular orders by purchase_count descending within each pinned group', async ({ request }) => {
        const res = await request.get(`${BASE}/api/v1/products?sort=popular&per_page=10`);
        expect(res.status()).toBe(200);
        const data = await res.json();
        expect(data.products.length).toBeGreaterThan(0);

        // Get purchase counts grouped by pinned status
        const pinnedCounts = data.products.filter(p => p.pinned).map(p => p.purchase_count);
        const unpinnedCounts = data.products.filter(p => !p.pinned).map(p => p.purchase_count);

        // Within pinned group, purchase_count should be non-increasing
        if (pinnedCounts.length > 1) {
            for (let i = 1; i < pinnedCounts.length; i++) {
                expect(pinnedCounts[i]).toBeLessThanOrEqual(pinnedCounts[i - 1]);
            }
        }
        if (unpinnedCounts.length > 1) {
            for (let i = 1; i < unpinnedCounts.length; i++) {
                expect(unpinnedCounts[i]).toBeLessThanOrEqual(unpinnedCounts[i - 1]);
            }
        }
    });
});
