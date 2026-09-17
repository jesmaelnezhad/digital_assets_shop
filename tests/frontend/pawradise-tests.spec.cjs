const { test, expect } = require('@playwright/test');

const BASE = process.env.BASE_URL || 'https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir';

const PAGES = [
  { name: 'Home', url: '/' },
  { name: 'Category', url: '/category' },
  { name: 'Login', url: '/login' },
  { name: 'Register', url: '/register' },
  { name: 'Bundle', url: '/product/bundle.html' },
  { name: 'Request', url: '/product/request.html' },
  { name: 'Product Detail', url: '/product/product-1' },
  { name: 'Community', url: '/community' },
  { name: 'Account', url: '/account' },
  { name: 'Cart', url: '/cart' },
  { name: 'Wishlist', url: '/wishlist' },
  { name: 'Referrals', url: '/referrals' },
  { name: 'Checkout', url: '/checkout' },
  { name: 'Admin', url: '/admin' },
  { name: 'Profile', url: '/profile/1' },
  { name: 'Post', url: '/post/1' },
];

// ============ 1. PAGE LOADS ============
test.describe('1. Page Loads', () => {
  for (const p of PAGES) {
    test(`${p.name}: returns 200, no nginx/error`, async ({ page }) => {
      const res = await page.goto(BASE + p.url);
      expect(res.status()).toBe(200);
      const body = await page.textContent('body');
      expect(body).not.toContain('Welcome to nginx');
      expect(body).not.toContain('500 Internal Server Error');
      expect(body).not.toContain('503 Service Unavailable');
      expect(body).not.toContain('404 Not Found');
    });
  }
});

// ============ 2. STYLING ============
test.describe('2. Styling (Computed Styles)', () => {
  for (const p of PAGES) {
    test(`${p.name}: dark background applied`, async ({ page }) => {
      await page.goto(BASE + p.url);
      await page.waitForTimeout(1000);
      const bg = await page.evaluate(() => window.getComputedStyle(document.body).backgroundColor);
      expect(bg === 'rgb(12, 12, 14)' || bg.includes('12, 12') || bg.includes('10, 14')).toBeTruthy();
    });

    test(`${p.name}: font-family loaded`, async ({ page }) => {
      await page.goto(BASE + p.url);
      await page.waitForTimeout(1000);
      const font = await page.evaluate(() => window.getComputedStyle(document.body).fontFamily);
      expect(font.toLowerCase()).toMatch(/inter|system-ui|sans-serif/);
    });

    test(`${p.name}: CSS variables resolve`, async ({ page }) => {
      await page.goto(BASE + p.url);
      await page.waitForTimeout(1000);
      const vars = await page.evaluate(() => {
        const s = getComputedStyle(document.documentElement);
        return {
          bg: s.getPropertyValue('--color-bg').trim(),
          font: s.getPropertyValue('--font-sans').trim(),
        };
      });
      expect(vars.bg === '#0c0c0e' || vars.bg === '#0a0e14' || vars.bg.includes('#0')).toBeTruthy();
      expect((vars.font || '').toLowerCase()).toMatch(/system-ui|sans-serif|ui-sans/);
    });
  }
});

// ============ 3. NAVIGATION ============
test.describe('3. Navigation', () => {
  const navPages = PAGES.filter(p => p.name !== 'Admin');

  for (const p of navPages) {
    test(`${p.name}: has nav with Shop, Community, Account`, async ({ page }) => {
      await page.goto(BASE + p.url);
      await page.waitForTimeout(1000);
      const nav = await page.evaluate(() => {
        const navs = document.querySelectorAll('nav');
        const main = Array.from(navs).find(n => !n.classList.contains('sidebar-nav'));
        if (!main) return [];
        return Array.from(main.querySelectorAll('a')).map(a => a.textContent.trim());
      });
      expect(nav).toContain('Shop');
      expect(nav).toContain('Community');
      expect(nav).toContain('Account');
      expect(nav.length).toBeGreaterThanOrEqual(4);
    });
  }

  test('Home: all internal links resolve to 200', async ({ page }) => {
    await page.goto(BASE);
    await page.waitForTimeout(1000);
    const links = await page.locator('a[href^="/"]').evaluateAll(els =>
      els.map(e => e.getAttribute('href')).filter(h => h && !h.includes('${'))
    );
    for (const link of [...new Set(links)]) {
      const res = await page.request.head(BASE + link);
      expect(res.status()).toBe(200);
    }
  });
});

// ============ 4. CONTENT RENDERING ============
test.describe('4. Content Rendering', () => {
  test('Home: products rendered from API', async ({ page }) => {
    await page.goto(BASE, { waitUntil: 'networkidle' });
    await page.waitForTimeout(3000);
    const cards = await page.locator('.product-card').count();
    expect(cards).toBeGreaterThan(0);
  });

  test('Category: products visible', async ({ page }) => {
    await page.goto(BASE + '/category', { waitUntil: 'networkidle' });
    await page.waitForTimeout(3000);
    const cards = await page.locator('.card, .product-card').count();
    expect(cards).toBeGreaterThan(0);
  });

  test('Community: feed visible', async ({ page }) => {
    await page.goto(BASE + '/community', { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);
    const feed = await page.locator('.post-card, .create-post, textarea').count();
    expect(feed).toBeGreaterThan(0);
  });
});

// ============ 5. FORMS ============
test.describe('5. Forms', () => {
  test('Login: has email, password, submit', async ({ page }) => {
    await page.goto(BASE + '/login');
    await page.waitForTimeout(1000);
    const email = await page.locator('input[type="email"], #email').count();
    const pass = await page.locator('input[type="password"]').count();
    const submit = await page.locator('button[type="submit"]').count();
    expect(email).toBeGreaterThan(0);
    expect(pass).toBeGreaterThan(0);
    expect(submit).toBeGreaterThan(0);
  });

  test('Register: has name, email, password, submit', async ({ page }) => {
    await page.goto(BASE + '/register');
    await page.waitForTimeout(1000);
    const name = await page.locator('input#name, input[name="name"]').count();
    const email = await page.locator('input[type="email"], #email').count();
    const pass = await page.locator('input[type="password"]').count();
    const submit = await page.locator('button[type="submit"]').count();
    expect(name).toBeGreaterThan(0);
    expect(email).toBeGreaterThan(0);
    expect(pass).toBeGreaterThan(0);
    expect(submit).toBeGreaterThan(0);
  });
});

// ============ 6. AUTH FLOW ============
test.describe('6. Auth Flow', () => {
  test('Register → Login → Access', async ({ page, request }) => {
    const email = 'e2e-' + Date.now() + '@test.com';
    
    // Register
    const reg = await request.post(BASE + '/api/v1/register', {
      data: { email, password: 'Pass1234', name: 'E2E Test' }
    });
    expect(reg.status()).toBe(201);

    // Login
    const login = await request.post(BASE + '/api/v1/login', {
      data: { email, password: 'Pass1234' }
    });
    expect(login.status()).toBe(200);
    const token = (await login.json()).token;
    expect(token).toBeTruthy();

    // Access protected
    const profile = await request.get(BASE + '/api/v1/me', {
      headers: { Authorization: 'Bearer ' + token }
    });
    expect(profile.status()).toBe(200);
  });
});

// ============ 7. API INTEGRATION ============
test.describe('7. Frontend-API Integration', () => {
  test('api.products.list() works', async ({ page }) => {
    await page.goto(BASE);
    await page.waitForTimeout(2000);
    const data = await page.evaluate(async () => {
      return await window.Pawradise.api.products.list({ per_page: 2 });
    });
    expect(data.products).toBeDefined();
    expect(data.products.length).toBeGreaterThan(0);
    expect(data.total).toBeGreaterThan(0);
  });

  test('api.products.getCategories() works', async ({ page }) => {
    await page.goto(BASE);
    await page.waitForTimeout(2000);
    const data = await page.evaluate(async () => {
      return await window.Pawradise.api.products.getCategories();
    });
    const cats = data.categories || data;
    expect(Array.isArray(cats)).toBeTruthy();
    expect(cats.length).toBeGreaterThan(0);
  });

  test('api.products.get(slug) works', async ({ page }) => {
    await page.goto(BASE + '/product/product-1', { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);
    const data = await page.evaluate(async () => {
      return await window.Pawradise.api.products.get('product-1');
    });
    expect(data.product || data.id || data.title).toBeTruthy();
  });

  test('api.community.getPosts() works', async ({ page }) => {
    await page.goto(BASE + '/community', { waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);
    const data = await page.evaluate(async () => {
      return await window.Pawradise.api.community.getPosts('recent', 1, 5);
    });
    expect(data.posts).toBeDefined();
  });
});
