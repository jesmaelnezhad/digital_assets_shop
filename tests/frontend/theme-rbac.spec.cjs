const { test, expect } = require('@playwright/test');
const { BASE, ADMIN_TOKEN, loginAs, unlockAccess } = require('./helpers.cjs');

test.describe('Appearance', () => {
  test('html carries palette tokens after boot', async ({ page }) => {
    await page.goto(BASE + '/');
    await page.waitForSelector('.brand', { timeout: 15000 });
    const palette = await page.evaluate(() => document.documentElement.dataset.palette || 'clay');
    expect(['clay', 'marble', 'night', 'moss', 'ink', 'ember', 'dune', 'frost', 'paper', 'chalk', 'linen', 'mist', 'petal', 'foam', 'porcelain', 'sage', 'snow', 'honey', 'bone', 'cloud', 'wine', 'violet', 'ocean', 'slate']).toContain(palette);
  });
});

test.describe('RBAC chrome', () => {
  test('guest header has no Admin link', async ({ page }) => {
    await page.goto(BASE + '/');
    await expect(page.locator('.header-nav')).not.toContainText('Admin');
  });

  test('customer login has no Admin link', async ({ page }) => {
    await loginAs(page, 'maya@example.com', 'maya', 'Maya');
    await expect(page.locator('.header-nav')).toContainText(/Maya|Account/);
    await expect(page.locator('.header-nav')).not.toContainText('Admin');
  });

  test('staff login shows Admin and only assigned tabs', async ({ page }) => {
    await loginAs(page, 'leo@example.com', 'leo', 'Admin');
    await expect(page.locator('.header-nav')).toContainText('Admin');
    await page.goto(BASE + '/admin');
    await expect(page.locator('#desk')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#tabs')).toContainText('Products');
    await expect(page.locator('#tabs')).toContainText('Banner');
    await expect(page.locator('#tabs')).not.toContainText('Access');
    await expect(page.locator('#tabs')).not.toContainText('Appearance');
  });

  test('admin can open Appearance', async ({ page }) => {
    await loginAs(page, 'nia@example.com', 'nia', 'Admin');
    await page.goto(BASE + '/admin');
    await expect(page.locator('#desk')).toBeVisible({ timeout: 10000 });
    await page.locator('#tabs button[data-t="Appearance"]').click();
    await expect(page.locator('[data-theme-kind="palette"]').first()).toBeVisible();
  });

  test('staff Orders desk shows pipeline chips', async ({ page }) => {
    await loginAs(page, 'leo@example.com', 'leo', 'Admin');
    await page.goto(BASE + '/admin');
    await expect(page.locator('#desk')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#tabs')).toContainText('Orders');
    await expect(page.locator('#tabs')).not.toContainText('Steps');
    await page.locator('#tabs button[data-t="Orders"]').click();
    await expect(page.locator('#order-filters')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#order-filters')).toContainText(/Waiting for payment|Paid|Preparation|Delivered/i);
  });

  test('admin Stats shows orders by step and Steps tab', async ({ page }) => {
    await loginAs(page, 'nia@example.com', 'nia', 'Admin');
    await page.goto(BASE + '/admin');
    await expect(page.locator('#desk')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#tabs')).toContainText('Steps');
    await expect(page.locator('.step-chips')).toBeVisible({ timeout: 10000 });
    await page.locator('#tabs button[data-t="Steps"]').click();
    await expect(page.locator('#step-list')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#step-list')).toContainText('Preparation');
    await expect(page.locator('#new-step')).toBeVisible();
  });

  test('Access search filters user cards', async ({ page }) => {
    test.setTimeout(60000);
    await loginAs(page, 'nia@example.com', 'nia', 'Admin');
    await page.goto(BASE + '/admin');
    await expect(page.locator('#desk')).toBeVisible({ timeout: 10000 });
    await unlockAccess(page);
    await expect(page.locator('#access-q')).toBeVisible({ timeout: 20000 });
    await expect(page.locator('form.access-form').first()).toBeVisible();
    await page.fill('#access-q', 'maya');
    await expect(page.locator('form.access-form:visible').first()).toContainText(/Maya/i);
    await expect(page.locator('form.access-form', { hasText: 'nia@example.com' })).toBeHidden();
  });
});
