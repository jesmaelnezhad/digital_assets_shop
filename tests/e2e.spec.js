const { test, expect } = require('@playwright/test');

const STAGING = process.env.BASE_URL || 'https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir';
const ADMIN_TOKEN = process.env.ADMIN_TOKEN || 'admin_secret_staging_2026';

let counter = 0;
function uniqueEmail() {
  counter++;
  return `e2e-${Date.now()}-${counter}@store4bots.xyz`;
}

test.describe('Staging shop', () => {
  test('homepage loads the catalog chrome', async ({ page }) => {
    await page.goto(`${STAGING}/`);
    await expect(page.locator('.brand')).toContainText('STORE4BOTS');
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
    await expect(page.locator('.header-nav')).not.toContainText('Admin');
  });

  test('homepage banner is a multi-slide slider', async ({ page }) => {
    await page.goto(`${STAGING}/`);
    await expect(page.locator('.hero-slide').first()).toBeVisible({ timeout: 15000 });
    const slides = await page.locator('.hero-slide').count();
    expect(slides).toBeGreaterThanOrEqual(2);
    await expect(page.locator('.hero-next, .hero-dots')).toBeVisible();
  });

  test('shop pager has more than one page', async ({ page }) => {
    await page.goto(`${STAGING}/`);
    await expect(page.locator('#pager button')).toHaveCount(await page.locator('#pager button').count());
    const pages = await page.locator('#pager button').count();
    expect(pages).toBeGreaterThan(1);
    await page.locator('#pager button').nth(1).click();
    await expect(page.locator('#grid .card').first()).toBeVisible();
  });

  test('community feed offers Show more', async ({ page }) => {
    await page.goto(`${STAGING}/community`);
    await expect(page.locator('#feed .post, #feed .post-card, #feed article, #feed .note').first()).toBeVisible({ timeout: 15000 });
    await expect(page.locator('#more-feed')).toBeVisible({ timeout: 10000 });
  });

  test('people directory offers Show more', async ({ page }) => {
    await page.goto(`${STAGING}/people`);
    await expect(page.locator('#list').locator('.person, .person-row, a, .note').first()).toBeVisible({ timeout: 15000 });
    await expect(page.locator('#more-people')).toBeVisible({ timeout: 10000 });
  });
});

test.describe('Staging mobile', () => {
  test.use({ viewport: { width: 390, height: 844 } });

  test('narrow viewport uses hamburger nav', async ({ page }) => {
    await page.goto(`${STAGING}/`);
    await expect(page.locator('.nav-toggle')).toBeVisible();
    await page.locator('.nav-toggle').click();
    await expect(page.locator('.site-header')).toHaveClass(/nav-open/);
    await expect(page.locator('.header-nav a').first()).toBeVisible();
  });

  test('hero and grid stay on screen at phone width', async ({ page }) => {
    await page.goto(`${STAGING}/`);
    await expect(page.locator('.hero')).toBeVisible({ timeout: 15000 });
    const heroBox = await page.locator('.hero').boundingBox();
    expect(heroBox.width).toBeLessThanOrEqual(390);
    await expect(page.locator('#grid .card').first()).toBeVisible();
  });
});

test.describe('Staging admin', () => {
  async function loginNia(page) {
    await page.goto(`${STAGING}/login`);
    await page.fill('input[name="email"]', 'nia@example.com');
    await page.fill('input[name="password"]', 'nia');
    await page.click('#form button.btn-accent');
    await expect(page.locator('.header-nav')).toContainText(/Nia|Admin/, { timeout: 15000 });
  }

  test('customers do not see Admin in the header', async ({ page }) => {
    await page.goto(`${STAGING}/`);
    await expect(page.locator('.header-nav')).not.toContainText('Admin');
  });

  test('admin login shows Admin and the desk without an operator token', async ({ page }) => {
    await loginNia(page);
    await expect(page.locator('.header-nav')).toContainText('Admin');
    await page.goto(`${STAGING}/admin`);
    await expect(page.locator('#desk')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#tabs')).toContainText('Banner');
    await expect(page.locator('#tabs')).toContainText('Appearance');
    await expect(page.locator('#tabs')).toContainText('Access');
    await expect(page.locator('#unlock')).toHaveCount(0);
  });

  test('admin banner tab lists slider products', async ({ page }) => {
    await loginNia(page);
    await page.goto(`${STAGING}/admin`);
    await expect(page.locator('#desk')).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Banner' }).click();
    await expect(page.locator('#banner-list, #banner-save')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#banner-save')).toContainText('Save banner');
  });

  test('Access tab asks for the operator token', async ({ page }) => {
    await loginNia(page);
    await page.goto(`${STAGING}/admin`);
    await expect(page.locator('#desk')).toBeVisible({ timeout: 10000 });
    await page.getByRole('button', { name: 'Access' }).click();
    await expect(page.locator('#unlock')).toBeVisible();
    await page.fill('input[name="token"]', ADMIN_TOKEN);
    await page.click('#unlock button');
    await expect(page.locator('form.access-form').first()).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#access-q')).toBeVisible();
    await expect(page.locator('.access-grid')).toBeVisible();
    await page.fill('#access-q', 'nia');
    await expect(page.locator('form.access-form:visible').first()).toContainText(/Nia/i);
    await expect(page.locator('form.access-form', { hasText: 'leo@example.com' })).toBeHidden();
  });
});
