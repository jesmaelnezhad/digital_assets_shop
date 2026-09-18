const { test, expect } = require('@playwright/test');
const { BASE, ADMIN_TOKEN, loginAs, unlockAccess } = require('./helpers.cjs');

test.describe('Homepage banner slider', () => {
  test('hero mounts multiple slides with controls', async ({ page }) => {
    await page.goto(BASE + '/');
    await page.waitForSelector('.hero-slide', { timeout: 15000 });
    expect(await page.locator('.hero-slide').count()).toBeGreaterThanOrEqual(2);
    await expect(page.locator('.hero-next')).toBeVisible();
    await page.locator('.hero-next').click();
    await expect(page.locator('.hero-slide.is-on')).toHaveCount(1);
  });
});

test.describe('Admin banner tab', () => {
  test('Banner tab is the control surface for slider products', async ({ page }) => {
    await loginAs(page, 'nia@example.com', 'nia', 'Admin');
    await page.goto(BASE + '/admin');
    await expect(page.locator('#desk')).toBeVisible({ timeout: 10000 });
    await page.locator('#tabs button[data-t="Banner"]').click();
    await expect(page.locator('#banner-save')).toBeVisible();
    await expect(page.locator('#banner-list [data-bid]').first()).toBeVisible({ timeout: 10000 });
  });
});

test.describe('Mobile chrome', () => {
  test.use({ viewport: { width: 390, height: 844 } });

  test('hamburger opens nav without horizontal overflow', async ({ page }) => {
    await page.goto(BASE + '/');
    await expect(page.locator('.nav-toggle')).toBeVisible();
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
    expect(overflow).toBeLessThanOrEqual(8);
    await page.locator('.nav-toggle').click();
    await expect(page.locator('.site-header.nav-open .header-nav a').first()).toBeVisible();
  });
});

test.describe('Pagination surfaces', () => {
  test('shop pager has multiple pages', async ({ page }) => {
    await page.goto(BASE + '/');
    await page.waitForSelector('#pager button', { timeout: 15000 });
    expect(await page.locator('#pager button').count()).toBeGreaterThan(1);
  });

  test('community show more exists', async ({ page }) => {
    await page.goto(BASE + '/community');
    await expect(page.locator('#more-feed')).toBeVisible({ timeout: 15000 });
  });

  test('people show more exists', async ({ page }) => {
    await page.goto(BASE + '/people');
    await expect(page.locator('#more-people')).toBeVisible({ timeout: 15000 });
  });
});

test.describe('Wide vs mobile layouts', () => {
  test('people directory is a 3-column grid on a wide screen', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto(BASE + '/people');
    await page.waitForSelector('.people-grid .person', { timeout: 15000 });
    const cols = await page.locator('.people-grid').evaluate((el) => getComputedStyle(el).gridTemplateColumns.split(' ').filter(Boolean).length);
    expect(cols).toBeGreaterThanOrEqual(3);
  });

  test('Access cards sit in a wide grid and a phone column', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await loginAs(page, 'nia@example.com', 'nia', 'Admin');
    await page.goto(BASE + '/admin');
    await expect(page.locator('#desk')).toBeVisible({ timeout: 10000 });
    await unlockAccess(page);
    await expect(page.locator('.access-grid')).toBeVisible({ timeout: 10000 });
    const wide = await page.locator('.access-grid').evaluate((el) => getComputedStyle(el).gridTemplateColumns.split(' ').filter(Boolean).length);
    expect(wide).toBeGreaterThanOrEqual(3);
    await page.setViewportSize({ width: 390, height: 844 });
    const phone = await page.locator('.access-grid').evaluate((el) => getComputedStyle(el).gridTemplateColumns.split(' ').filter(Boolean).length);
    expect(phone).toBe(1);
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
    expect(overflow).toBeLessThanOrEqual(8);
  });

  test('shop and community stay inside a phone viewport', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    for (const path of ['/', '/community', '/people', '/bundles']) {
      await page.goto(BASE + path);
      await page.waitForSelector('.brand', { timeout: 15000 });
      const overflow = await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
      expect(overflow, path).toBeLessThanOrEqual(8);
    }
  });
});
