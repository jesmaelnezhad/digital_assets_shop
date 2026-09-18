const BASE = process.env.BASE_URL || 'https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir';
const ADMIN_TOKEN = process.env.ADMIN_TOKEN || 'admin_secret_staging_2026';

async function loginAs(page, email, password, waitFor) {
  await page.goto(BASE + '/login');
  await page.fill('input[name="email"]', email);
  await page.fill('input[name="password"]', password);
  await page.locator('#form button.btn-accent').click();
  await page.waitForFunction((needle) => {
    const nav = document.querySelector('.header-nav');
    if (!nav) return false;
    const text = nav.textContent || '';
    if (/\bLog in\b/.test(text)) return false;
    if (needle && !text.includes(needle)) return false;
    return true;
  }, waitFor || '', { timeout: 15000 });
}

async function unlockAccess(page, token) {
  await page.locator('#tabs button[data-t="Access"]').click();
  await page.waitForFunction(() => {
    const f = document.querySelector('#unlock');
    return !!(f && f.onsubmit);
  }, null, { timeout: 10000 });
  await page.fill('#unlock input[name="token"]', token || ADMIN_TOKEN);
  await page.locator('#unlock button.btn-accent').click();
  await page.waitForSelector('#access-q, .access-grid', { timeout: 20000 });
}

module.exports = { BASE, ADMIN_TOKEN, loginAs, unlockAccess };
