const { test, expect } = require('@playwright/test');

const PROD = 'http://127.0.0.1';
const STAGING = 'http://127.0.0.1/staging';
const ADMIN_TOKEN = 'admin-secret-token-change-in-production';

// Helper to get unique emails
let counter = 0;
function uniqueEmail() {
  counter++;
  return `e2e-${Date.now()}-${counter}@pawradise.ir`;
}

// ===== PRODUCTION TESTS =====

test.describe('Production Environment', () => {
  test('homepage loads', async ({ page }) => {
    await page.goto(`${PROD}/`);
    await expect(page.locator('h1')).toBeVisible();
  });

  test('register flow works', async ({ page }) => {
    await page.goto(`${PROD}/`);
    
    // Switch to register mode
    await page.click('#switchLink');
    await expect(page.locator('#formTitle')).toContainText('Create Account');
    
    const email = uniqueEmail();
    await page.fill('#name', 'E2E User');
    await page.fill('#email', email);
    await page.fill('#password', 'TestPass123');
    await page.click('#submitBtn');
    
    await expect(page.locator('#profileView')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#profileName')).toContainText('E2E User');
  });

  test('login flow works', async ({ page }) => {
    // First register a user
    const email = uniqueEmail();
    await page.goto(`${PROD}/`);
    await page.click('#switchLink');
    await page.fill('#name', 'Login Test');
    await page.fill('#email', email);
    await page.fill('#password', 'TestPass123');
    await page.click('#submitBtn');
    await expect(page.locator('#profileView')).toBeVisible({ timeout: 10000 });
    
    // Logout
    await page.click('button:has-text("Logout")');
    await expect(page.locator('#authForm')).toBeVisible();
    
    // Login
    await page.fill('#email', email);
    await page.fill('#password', 'TestPass123');
    await page.click('#submitBtn');
    await expect(page.locator('#profileView')).toBeVisible({ timeout: 10000 });
  });

  test('profile update works', async ({ page }) => {
    const email = uniqueEmail();
    await page.goto(`${PROD}/`);
    await page.click('#switchLink');
    await page.fill('#name', 'Profile Update');
    await page.fill('#email', email);
    await page.fill('#password', 'TestPass123');
    await page.click('#submitBtn');
    await expect(page.locator('#profileView')).toBeVisible({ timeout: 10000 });
    
    // Update profile
    await page.fill('#profileNameInput', 'Updated Name');
    await page.fill('#profileEmailInput', uniqueEmail());
    await page.click('button:has-text("Update Profile")');
    await expect(page.locator('#profileMessage')).toContainText('Profile updated successfully', { timeout: 5000 });
  });

  test('logout works', async ({ page }) => {
    const email = uniqueEmail();
    await page.goto(`${PROD}/`);
    await page.click('#switchLink');
    await page.fill('#name', 'Logout Test');
    await page.fill('#email', email);
    await page.fill('#password', 'TestPass123');
    await page.click('#submitBtn');
    await expect(page.locator('#profileView')).toBeVisible({ timeout: 10000 });
    
    await page.click('button:has-text("Logout")');
    await expect(page.locator('#authForm')).toBeVisible();
    await expect(page.locator('#profileView')).not.toBeVisible();
  });
});

test.describe('Staging Environment', () => {
  test('staging homepage loads', async ({ page }) => {
    await page.goto(`${STAGING}/`);
    await expect(page.locator('h1')).toBeVisible();
  });

  test('staging register works', async ({ page }) => {
    await page.goto(`${STAGING}/`);
    await page.click('#switchLink');
    const email = uniqueEmail();
    await page.fill('#name', 'Staging E2E');
    await page.fill('#email', email);
    await page.fill('#password', 'StagingPass123');
    await page.click('#submitBtn');
    await expect(page.locator('#profileView')).toBeVisible({ timeout: 10000 });
  });

  test('staging environment badge shows correctly', async ({ page }) => {
    await page.goto(`${STAGING}/`);
    await expect(page.locator('#env')).toContainText('ENV: staging');
  });
});

test.describe('Admin Panel - Production', () => {
  test('admin login page loads', async ({ page }) => {
    await page.goto(`${PROD}/admin`);
    await expect(page.locator('#adminLogin')).toBeVisible();
  });

  test('admin login succeeds with correct token', async ({ page }) => {
    await page.goto(`${PROD}/admin`);
    await page.fill('#adminToken', ADMIN_TOKEN);
    await page.click('#loginBtn');
    await expect(page.locator('#adminPanel')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('table')).toBeVisible();
  });

  test('admin can view users', async ({ page }) => {
    await page.goto(`${PROD}/admin`);
    await page.fill('#adminToken', ADMIN_TOKEN);
    await page.click('#loginBtn');
    await expect(page.locator('#adminPanel')).toBeVisible({ timeout: 10000 });
    
    // At least the table should be visible
    await expect(page.locator('table')).toBeVisible();
  });

  test('admin logout works', async ({ page }) => {
    await page.goto(`${PROD}/admin`);
    await page.fill('#adminToken', ADMIN_TOKEN);
    await page.click('#loginBtn');
    await expect(page.locator('#adminPanel')).toBeVisible({ timeout: 10000 });
    
    await page.click('button:has-text("Logout")');
    await expect(page.locator('#adminLogin')).toBeVisible();
  });
});
