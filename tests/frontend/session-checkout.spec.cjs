const { test, expect } = require('@playwright/test');
const { BASE, ADMIN_TOKEN, loginAs } = require('./helpers.cjs');

test.describe('Logged-in session surfaces', () => {
  test('referrals does not ask a logged-in user to log in', async ({ page }) => {
    await loginAs(page, 'maya@example.com', 'maya', 'Maya');
    await page.goto(BASE + '/referrals');
    await page.waitForSelector('#root, .page', { timeout: 15000 });
    const root = await page.textContent('#root');
    expect(root).not.toMatch(/Log in to see your referral/i);
    expect(root).toMatch(/referral|code|earning|commission|share/i);
  });

  test('header exposes Log out after login', async ({ page }) => {
    await loginAs(page, 'maya@example.com', 'maya', 'Maya');
    await expect(page.locator('.header-nav')).toContainText(/Log out/i);
    await expect(page.locator('.header-nav')).toContainText(/Maya|Account/i);
  });
});

test.describe('Events collector', () => {
  test('product view records a product_view event', async ({ page, request }) => {
    await page.goto(BASE + '/product/lunar-clay-characters');
    await page.waitForSelector('.page-title, .buy, h1', { timeout: 20000 });
    await page.waitForTimeout(800);
    const res = await request.get(BASE + '/api/v1/admin/events?name=product_view&limit=5', {
      headers: { Authorization: 'Bearer ' + ADMIN_TOKEN }
    });
    expect(res.ok()).toBeTruthy();
    const data = await res.json();
    expect((data.events || []).some((e) => e.name === 'product_view')).toBeTruthy();
  });
});
