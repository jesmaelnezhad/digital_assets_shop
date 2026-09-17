const { test, expect } = require('@playwright/test');

const STAGING = process.env.BASE_URL || 'https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir';
const ADMIN_TOKEN = process.env.ADMIN_TOKEN || 'admin_secret_staging_2026';

let counter = 0;
function uniqueEmail() {
  counter++;
  return `e2e-${Date.now()}-${counter}@pawradise.ir`;
}

test.describe('Staging shop', () => {
  test('homepage loads the catalog chrome', async ({ page }) => {
    await page.goto(`${STAGING}/`);
    await expect(page.locator('.brand')).toContainText('PAWRADISE');
    await expect(page.locator('#grid, .hero')).toBeVisible();
  });

  test('register then login', async ({ page }) => {
    const email = uniqueEmail();
    await page.goto(`${STAGING}/register`);
    await page.fill('input[name="name"]', 'E2E User');
    await page.fill('input[name="email"]', email);
    await page.fill('input[name="password"]', 'TestPass123');
    await page.click('button.btn-accent');
    await page.waitForURL(/\/($|account|shop)?/, { timeout: 15000 });

    await page.goto(`${STAGING}/login`);
    await page.fill('input[name="email"]', email);
    await page.fill('input[name="password"]', 'TestPass123');
    await page.click('button.btn-accent');
    await expect(page.locator('.header-nav')).toContainText(/Account|E2E/, { timeout: 10000 });
  });

  test('demo login nia', async ({ page }) => {
    await page.goto(`${STAGING}/login`);
    await page.fill('input[name="email"]', 'nia@example.com');
    await page.fill('input[name="password"]', 'nia');
    await page.click('button.btn-accent');
    await expect(page.locator('.header-nav a[href="/account"], .header-nav a.is-on')).toBeVisible({ timeout: 10000 });
  });
});

test.describe('Staging admin', () => {
  test('admin unlocks with staging token', async ({ page }) => {
    await page.goto(`${STAGING}/admin`);
    await page.fill('input[name="token"]', ADMIN_TOKEN);
    await page.click('#unlock button, form#unlock button');
    await expect(page.locator('#desk')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#tabs')).toContainText('Products');
  });

  test('admin lock returns to gate', async ({ page }) => {
    await page.goto(`${STAGING}/admin`);
    await page.fill('input[name="token"]', ADMIN_TOKEN);
    await page.click('form#unlock button');
    await expect(page.locator('#desk')).toBeVisible({ timeout: 10000 });
    await page.click('#lock');
    await expect(page.locator('#unlock')).toBeVisible();
  });
});
