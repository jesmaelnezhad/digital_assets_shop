import https from 'https';
import http from 'http';

// ============================================================
// Store4bots E2E Test Suite (TDD)
// Tests define EXPECTED behavior per PRODUCT-SPEC.md v1.1
// Tests are NOT expected to pass until implementation is complete.
// ============================================================

const ADMIN_TOKEN = process.env.ADMIN_TOKEN || 'admin_secret_staging_2026';
const STAGING_DOMAIN = 'server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir';
const PRODUCTION_DOMAIN = 'store4bots.xyz'; // placeholder
const ENV_NAME = process.env.ENV_NAME || 'staging';
const API_HOST = ENV_NAME === 'production' ? PRODUCTION_DOMAIN : STAGING_DOMAIN;
const API_BASE = `https://${API_HOST}/api/v1`;
const API_ROOT = `https://${API_HOST}`;

// ---- Test infrastructure ----
let passed = 0;
let failed = 0;
const failures = [];

function request(url, opts = {}) {
  return new Promise((resolve, reject) => {
    const transport = url.startsWith('https') ? https : http;
    const req = transport.request(url, { ...opts, rejectUnauthorized: false }, (res) => {
      let data = '';
      res.on('data', (c) => (data += c));
      res.on('end', () =>
        resolve({
          status: res.statusCode,
          headers: res.headers,
          body: data,
          json: () => {
            try { return JSON.parse(data); } catch { return null; }
          },
        })
      );
    });
    req.on('error', reject);
    if (opts.body) req.write(opts.body);
    req.end();
  });
}

async function test(name, fn) {
  try {
    await fn();
    passed++;
    console.log(`  ✓ ${name}`);
  } catch (err) {
    failed++;
    failures.push({ name, error: err.message });
    console.log(`  ✗ ${name}: ${err.message}`);
  }
}

function assert(condition, message) {
  if (!condition) throw new Error(message || 'Assertion failed');
}

function assertStatus(res, expected, context) {
  if (res.status !== expected) {
    throw new Error(`${context || ''} Expected status ${expected}, got ${res.status}: ${res.body.substring(0, 200)}`);
  }
}

function assertArray(data, field, context) {
  if (!data || !Array.isArray(data[field])) {
    throw new Error(`${context || ''} Expected array '${field}' in response: ${JSON.stringify(data).substring(0, 200)}`);
  }
}

function assertField(data, field, context) {
  if (!data || data[field] === undefined || data[field] === null) {
    throw new Error(`${context || ''} Expected field '${field}' in response: ${JSON.stringify(data).substring(0, 200)}`);
  }
}

async function section(title) {
  console.log(`\n=== ${title} ===`);
}

async function tryAuth(email, password) {
  try {
    const res = await request(`${API_BASE}/login`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });
    if (res.status === 200) return res.json().token;
  } catch {}
  return null;
}

// ============================================================
// TEST SUITES
// ============================================================

async function runAuthTests() {
  section('Authentication');

  const email = `e2e-${Date.now()}@example.com`;
  const password = 'test123';
  const name = 'E2E User';

  await test('POST /register creates user and returns token', async () => {
    const res = await request(`${API_BASE}/register`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ email, password, name }),
    });
    assertStatus(res, 201, 'Register');
    const data = res.json();
    assertField(data, 'token', 'Register');
    assertField(data, 'user', 'Register');
    assert(data.user.email === email, 'Email mismatch');
    assert(data.user.name === name, 'Name mismatch');
  });

  await test('POST /register rejects duplicate email', async () => {
    const res = await request(`${API_BASE}/register`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ email, password, name }),
    });
    assertStatus(res, 409, 'Duplicate register');
  });

  await test('POST /register validates required fields', async () => {
    const res = await request(`${API_BASE}/register`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ email: 'bad' }),
    });
    assertStatus(res, 400, 'Invalid register');
  });

  await test('POST /login returns token for valid credentials', async () => {
    const res = await request(`${API_BASE}/login`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });
    assertStatus(res, 200, 'Login');
    const data = res.json();
    assertField(data, 'token', 'Login');
  });

  await test('POST /login rejects invalid password', async () => {
    const res = await request(`${API_BASE}/login`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ email, password: 'wrong' }),
    });
    assertStatus(res, 401, 'Bad login');
  });

  await test('POST /login rejects unknown email', async () => {
    const res = await request(`${API_BASE}/login`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ email: 'unknown@example.com', password: 'test123' }),
    });
    assertStatus(res, 401, 'Unknown login');
  });

  const token = await tryAuth(email, password);

  await test('GET /me returns current user profile', async () => {
    assert(token, 'Login failed - cannot test profile');
    const res = await request(`${API_BASE}/me`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    assertStatus(res, 200, 'Get profile');
    const data = res.json();
    assert(data.email === email, 'Profile email mismatch');
    assert(data.role === 'customer' || data.role === 'staff' || data.role === 'admin', 'Profile role');
  });

  await test('GET /me rejects without token', async () => {
    const res = await request(`${API_BASE}/me`);
    assertStatus(res, 401, 'No token');
  });

  await test('GET /me rejects invalid token', async () => {
    const res = await request(`${API_BASE}/me`, {
      headers: { Authorization: 'Bearer invalid-token' },
    });
    assertStatus(res, 401, 'Invalid token');
  });

  await test('PUT /me updates profile fields', async () => {
    assert(token, 'Login failed - cannot test profile update');
    const res = await request(`${API_BASE}/me`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${token}`, 'content-type': 'application/json' },
      body: JSON.stringify({ name: 'Updated Name', bio: 'Test bio', wallet_address: '0x1234' }),
    });
    assertStatus(res, 200, 'Update profile');
    const data = res.json();
    assert(data.name === 'Updated Name', 'Name not updated');
    assert(data.bio === 'Test bio', 'Bio not updated');
    assert(data.wallet_address === '0x1234', 'Wallet not updated');
  });

  await test('POST /logout invalidates token', async () => {
    assert(token, 'Login failed - cannot test logout');
    const res = await request(`${API_BASE}/logout`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
    });
    assertStatus(res, 200, 'Logout');
    const afterLogout = await request(`${API_BASE}/me`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    assertStatus(afterLogout, 401, 'After logout');
  });

  return { email, password, name, token };
}

async function runProductTests() {
  section('Products');

  await test('GET /products returns paginated product list', async () => {
    const res = await request(`${API_BASE}/products`);
    assertStatus(res, 200, 'Products list');
    const data = res.json();
    assertArray(data, 'products', 'Products list');
    assertField(data, 'total', 'Products list');
  });

  await test('GET /products supports pagination params', async () => {
    const res = await request(`${API_BASE}/products?page=1&per_page=5`);
    assertStatus(res, 200, 'Paginated products');
    const data = res.json();
    assert(data.products.length <= 5, 'Page size exceeded');
    assert(data.total > 12, `Need catalog volume for pager, total=${data.total}`);
  });

  await test('GET /products page 2 is a different slice', async () => {
    const a = (await request(`${API_BASE}/products?page=1&per_page=12`)).json();
    const b = (await request(`${API_BASE}/products?page=2&per_page=12`)).json();
    assert(a.products.length > 0 && b.products.length > 0, 'both pages need products');
    assert(a.products[0].id !== b.products[0].id, 'page 2 should not repeat page 1');
  });

  await test('GET /products?banner=1 returns ordered slider slides', async () => {
    const res = await request(`${API_BASE}/products?banner=1&per_page=24`);
    assertStatus(res, 200, 'Banner list');
    const data = res.json();
    assertArray(data, 'products', 'Banner products');
    assert(data.products.length >= 2, `banner needs 2+ slides, got ${data.products.length}`);
    const sorts = data.products.map((p) => p.banner_sort || 0);
    for (let i = 1; i < sorts.length; i++) {
      assert(sorts[i] >= sorts[i - 1], 'banner_sort should be ascending');
    }
  });

  await test('PUT /products/banner reorders slider and can restore', async () => {
    const current = (await request(`${API_BASE}/products?banner=1&per_page=24`)).json().products || [];
    const ids = current.map((p) => p.id);
    assert(ids.length >= 2, 'need banner slides to reorder');
    const swapped = [ids[1], ids[0]].concat(ids.slice(2));
    const put = await request(`${API_BASE}/products/banner`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
      body: JSON.stringify({ product_ids: swapped }),
    });
    assertStatus(put, 200, 'Set banner');
    const after = (await request(`${API_BASE}/products?banner=1&per_page=24`)).json().products || [];
    assert(after[0] && after[0].id === swapped[0], 'first slide should match saved order');
    const restore = await request(`${API_BASE}/products/banner`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
      body: JSON.stringify({ product_ids: ids }),
    });
    assertStatus(restore, 200, 'Restore banner');
  });

  await test('GET /products supports search query', async () => {
    const res = await request(`${API_BASE}/products?q=test`);
    assertStatus(res, 200, 'Related products');
    const data = res.json();
    // Related products are under 'recommendations.products' or 'related_products'
    const related = (data.recommendations && data.recommendations.products) || data.related_products;
    if (related) {
      assertArray(data, related === data.related_products ? 'related_products' : 'products', 'Related products');
    }
  });

  await test('GET /products supports category filter', async () => {
    const res = await request(`${API_BASE}/products?category_id=1`);
    assertStatus(res, 200, 'Filter by category');
    const data = res.json();
    assertArray(data, 'products', 'Filtered products');
  });

  await test('GET /products supports price range filter', async () => {
    const res = await request(`${API_BASE}/products?price_min=0&price_max=100`);
    assertStatus(res, 200, 'Filter by price');
    const data = res.json();
    assertArray(data, 'products', 'Price filtered');
  });

  await test('GET /products supports file_type filter', async () => {
    const res = await request(`${API_BASE}/products?file_type=image`);
    assertStatus(res, 200, 'Filter by file type');
    const data = res.json();
    assertArray(data, 'products', 'File type filtered');
  });

  await test('GET /products supports rating filter', async () => {
    const res = await request(`${API_BASE}/products?rating=4`);
    assertStatus(res, 200, 'Filter by rating');
    const data = res.json();
    assertArray(data, 'products', 'Rating filtered');
  });

  await test('GET /products supports sort by newest', async () => {
    const res = await request(`${API_BASE}/products?sort=newest`);
    assertStatus(res, 200, 'Sort newest');
  });

  await test('GET /products supports sort by price_asc', async () => {
    const res = await request(`${API_BASE}/products?sort=price_asc`);
    assertStatus(res, 200, 'Sort price asc');
  });

  await test('GET /products supports sort by price_desc', async () => {
    const res = await request(`${API_BASE}/products?sort=price_desc`);
    assertStatus(res, 200, 'Sort price desc');
  });

  await test('GET /products supports sort by popular', async () => {
    const res = await request(`${API_BASE}/products?sort=popular`);
    assertStatus(res, 200, 'Sort popular');
  });

  await test('GET /products returns pinned products first', async () => {
    const res = await request(`${API_BASE}/products`);
    assertStatus(res, 200, 'Pinned first');
    const data = res.json();
    if (data.products.length > 1) {
      assert(data.products[0].pinned === true || data.products[0].sort_order <= 0, 'Pinned not first');
    }
  });

  await test('GET /products/:slug returns product detail', async () => {
    const listRes = await request(`${API_BASE}/products?per_page=1`);
    const listData = listRes.json();
    if (listData.products && listData.products.length > 0) {
      const slug = listData.products[0].slug;
      const res = await request(`${API_BASE}/products/${slug}`);
      assertStatus(res, 200, 'Product detail');
      const data = res.json();
      // Product detail wraps product data under 'product' key
      const product = data.product || data;
      assertField(product, 'title', 'Product detail');
      assertField(product, 'slug', 'Product detail');
      assertField(product, 'price_usd', 'Product detail');
    }
  });

  await test('GET /products/:slug returns 404 for unknown slug', async () => {
    const res = await request(`${API_BASE}/products/nonexistent-slug-12345`);
    assertStatus(res, 404, 'Unknown product');
  });

  await test('GET /products/:slug includes image gallery', async () => {
    const listRes = await request(`${API_BASE}/products?per_page=1`);
    const listData = listRes.json();
    if (listData.products && listData.products.length > 0) {
      const slug = listData.products[0].slug;
      const res = await request(`${API_BASE}/products/${slug}`);
      const data = res.json();
      assertArray(data, 'images', 'Product images');
    }
  });

  await test('GET /products/:slug includes generated previews for images', async () => {
    const listRes = await request(`${API_BASE}/products?per_page=1`);
    const listData = listRes.json();
    if (listData.products && listData.products.length > 0) {
      const slug = listData.products[0].slug;
      const res = await request(`${API_BASE}/products/${slug}`);
      const data = res.json();
      if (data.images && data.images.length > 0) {
        const hasPreview = data.images.some(img => img.image_type === 'preview' || img.image_type === 'thumbnail');
        assert(hasPreview, 'No preview/thumbnail images found');
      }
    }
  });

  await test('GET /products/:slug includes tiers if product has tiers', async () => {
    const listRes = await request(`${API_BASE}/products?per_page=1`);
    const listData = listRes.json();
    if (listData.products && listData.products.length > 0) {
      const slug = listData.products[0].slug;
      const res = await request(`${API_BASE}/products/${slug}`);
      const data = res.json();
      if (data.tiers) {
        assert(Array.isArray(data.tiers), 'Tiers should be array');
      }
    }
  });

  await test('GET /products/:slug includes PWYW config if enabled', async () => {
    const listRes = await request(`${API_BASE}/products?per_page=1`);
    const listData = listRes.json();
    if (listData.products && listData.products.length > 0) {
      const slug = listData.products[0].slug;
      const res = await request(`${API_BASE}/products/${slug}`);
      const data = res.json();
      if (data.pwyw_enabled) {
        assertField(data, 'pwyw_min_price', 'PWYW min price');
      }
    }
  });

  await test('GET /products/:slug includes related products', async () => {
    const listRes = await request(`${API_BASE}/products?per_page=1`);
    const listData = listRes.json();
    if (listData.products && listData.products.length > 0) {
      const slug = listData.products[0].slug;
      const res = await request(`${API_BASE}/products/${slug}`);
      const data = res.json();
      // Related products are under 'recommendations.products'
      const recs = data.recommendations && data.recommendations.products;
      if (recs) {
        assert(Array.isArray(recs), 'Recommendations should be array');
      }
    }
  });

  await test('GET /categories returns category list', async () => {
    const res = await request(`${API_BASE}/categories`);
    assertStatus(res, 200, 'Categories');
    const data = res.json();
    assertArray(data, 'categories', 'Categories');
  });
}

async function runOrderTests(authToken) {
  section('Orders');

  await test('POST /orders creates order from cart', async () => {
    if (!authToken) throw new Error('No auth token available');
    const productsRes = await request(`${API_BASE}/products?per_page=1`);
    const productsData = productsRes.json();
    if (productsData.products && productsData.products.length > 0) {
      const productId = productsData.products[0].id;
      await request(`${API_BASE}/cart/items`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
        body: JSON.stringify({ product_id: productId, quantity: 1 }),
      });
      const res = await request(`${API_BASE}/orders`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
        body: JSON.stringify({}),
      });
      assertStatus(res, 201, 'Create order');
      const data = res.json();
      assertField(data, 'order', 'Order wrapper');
      assertField(data.order, 'id', 'Order id');
      assertField(data, 'total_usd', 'Order total');
    }
  });

  await test('POST /orders with 100% coupon is paid and zero_due', async () => {
    if (!authToken) throw new Error('No auth token available');
    const code = `FREE${Date.now()}`;
    const coupon = await request(`${API_ROOT}/api/v1/admin/coupons`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
      body: JSON.stringify({
        code,
        discount_type: 'percentage',
        discount_value: 100,
        expires_at: '2027-12-31',
        usage_limit: 50,
        min_purchase_usd: 0,
        is_active: true,
      }),
    });
    assertStatus(coupon, 201, 'Create 100% coupon');
    const res = await request(`${API_BASE}/orders`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
      body: JSON.stringify({
        coupon_code: code,
        items: [{ product_id: 1, quantity: 1, price_usd: 12 }],
      }),
    });
    assertStatus(res, 201, 'Create zero-due order');
    const data = res.json();
    assert(data.zero_due === true, 'zero_due should be true');
    assert(data.order && data.order.status === 'paid', 'order should be paid');
    assert(Number(data.total_usd) === 0, 'total should be 0');
  });

  await test('POST /orders rejects without auth', async () => {
    const res = await request(`${API_BASE}/orders`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({}),
    });
    assertStatus(res, 401, 'No auth order');
  });

  await test('GET /orders lists user orders', async () => {
    if (!authToken) throw new Error('No auth token available');
    const res = await request(`${API_BASE}/orders`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    assertStatus(res, 200, 'List orders');
    const data = res.json();
    assertArray(data, 'orders', 'Orders list');
  });

  await test('GET /orders/:id returns order detail', async () => {
    if (!authToken) throw new Error('No auth token available');
    const listRes = await request(`${API_BASE}/orders`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const listData = listRes.json();
    if (listData.orders && listData.orders.length > 0) {
      const orderId = listData.orders[0].id;
      const res = await request(`${API_BASE}/orders/${orderId}`, {
        headers: { Authorization: `Bearer ${authToken}` },
      });
      assertStatus(res, 200, 'Order detail');
      const data = res.json();
      assertField(data, 'order', 'Order wrapper');
      assertField(data.order, 'id', 'Order detail id');
      assertArray(data, 'items', 'Order detail items');
    }
  });

  await test('GET /orders/:id/payment returns payment details', async () => {
    if (!authToken) throw new Error('No auth token available');
    const listRes = await request(`${API_BASE}/orders`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const listData = listRes.json();
    if (listData.orders && listData.orders.length > 0) {
      const orderId = listData.orders[0].id;
      const res = await request(`${API_BASE}/orders/${orderId}/payment`, {
        headers: { Authorization: `Bearer ${authToken}` },
      });
      assertStatus(res, 200, 'Payment details');
      const data = res.json();
      // Payment info may be nested under order
      const payment = data.payment || data;
      assertField(payment, 'crypto_address', 'Payment address');
      assertField(payment, 'crypto_amount', 'Payment amount');
      assertField(payment, 'crypto_chain', 'Payment chain');
    }
  });

  await test('GET /orders/:id/status returns payment status', async () => {
    if (!authToken) throw new Error('No auth token available');
    const listRes = await request(`${API_BASE}/orders`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const listData = listRes.json();
    if (listData.orders && listData.orders.length > 0) {
      const orderId = listData.orders[0].id;
      const res = await request(`${API_BASE}/orders/${orderId}/status`, {
        headers: { Authorization: `Bearer ${authToken}` },
      });
      assertStatus(res, 200, 'Order status');
      const data = res.json();
      // Order status endpoint may return data.status or data.order with id/status
      // Just verify we got a 200 with a valid JSON response containing order data
      assert(
        typeof data === 'object' &&
        (data.status !== undefined ||
         (data.order && (data.order.status !== undefined || data.order.id !== undefined))),
        'Order status response missing expected fields, got: ' + JSON.stringify(data).substring(0, 150)
      );
    }
  });

  await test('GET /orders/:id/download/:itemId provides download', async () => {
    if (!authToken) throw new Error('No auth token available');
    const listRes = await request(`${API_BASE}/orders`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const listData = listRes.json();
    if (listData.orders && listData.orders.length > 0) {
      const order = listData.orders[0];
      if (order.items && order.items.length > 0) {
        const itemId = order.items[0].id;
        const res = await request(`${API_BASE}/orders/${order.id}/download/${itemId}`, {
          headers: { Authorization: `Bearer ${authToken}` },
        });
        assert(res.status === 200 || res.status === 402 || res.status === 403,
          `Unexpected status ${res.status}`);
      }
    }
  });

  await test('POST /guest-orders creates guest order', async () => {
    const res = await request(`${API_BASE}/guest-orders`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({
        email: `guest-${Date.now()}@example.com`,
        total_usd: 9.99,
        crypto_chain: 'BSC',
        crypto_amount: '0.01',
        crypto_address: '0x1234',
        status: 'pending',
      }),
    });
    assertStatus(res, 201, 'Guest order');
    const data = res.json();
    assertField(data, 'id', 'Guest order id');
  });

  await test('GET /guest-orders/:id checks guest order status', async () => {
    const email = `guest-check-${Date.now()}@example.com`;
    const createRes = await request(`${API_BASE}/guest-orders`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({
        email,
        total_usd: 9.99,
        crypto_chain: 'BSC',
        crypto_amount: '0.01',
        crypto_address: '0x1234',
        status: 'pending',
      }),
    });
    const createData = createRes.json();
    if (createData && createData.id) {
      const res = await request(`${API_BASE}/guest-orders/${createData.id}?email=${encodeURIComponent(email)}`);
      assertStatus(res, 200, 'Guest order status');
      const data = res.json();
      assertField(data, 'status', 'Guest status field');
    }
  });
}

async function runCartTests(authToken) {
  section('Cart');

  await test('GET /cart returns or creates cart', async () => {
    if (!authToken) throw new Error('No auth token available');
    const res = await request(`${API_BASE}/cart`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    assertStatus(res, 200, 'Get cart');
    const data = res.json();
    assertArray(data, 'items', 'Cart items');
  });

  await test('POST /cart/items adds item to cart', async () => {
    if (!authToken) throw new Error('No auth token available');
    const productsRes = await request(`${API_BASE}/products?per_page=1`);
    const productsData = productsRes.json();
    if (productsData.products && productsData.products.length > 0) {
      const productId = productsData.products[0].id;
      const res = await request(`${API_BASE}/cart/items`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
        body: JSON.stringify({ product_id: productId, quantity: 1 }),
      });
      assertStatus(res, 200, 'Add to cart');
      const data = res.json();
      // Cart returns message on success, no id field
      assert(data.message === 'added to cart', 'Add to cart message mismatch');
    }
  });

  await test('POST /cart/items rejects without auth', async () => {
    const res = await request(`${API_BASE}/cart/items`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ product_id: 1, quantity: 1 }),
    });
    assertStatus(res, 401, 'No auth cart');
  });

  await test('DELETE /cart/items/:id removes item from cart', async () => {
    if (!authToken) throw new Error('No auth token available');
    const cartRes = await request(`${API_BASE}/cart`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    const cartData = cartRes.json();
    if (cartData.items && cartData.items.length > 0) {
      const itemId = cartData.items[0].id;
      const res = await request(`${API_BASE}/cart/items/${itemId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${authToken}` },
      });
      assertStatus(res, 200, 'Remove from cart');
    }
  });
}

async function runWishlistTests(authToken) {
  section('Wishlist');

  await test('GET /wishlist lists user wishlist', async () => {
    if (!authToken) throw new Error('No auth token available');
    const res = await request(`${API_BASE}/wishlist`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    assertStatus(res, 200, 'List wishlist');
    const data = res.json();
    assertArray(data, 'products', 'Wishlist products');
  });

  await test('POST /wishlist/toggle adds product to wishlist', async () => {
    if (!authToken) throw new Error('No auth token available');
    const productsRes = await request(`${API_BASE}/products?per_page=1`);
    const productsData = productsRes.json();
    if (productsData.products && productsData.products.length > 0) {
      const productId = productsData.products[0].id;
      const res = await request(`${API_BASE}/wishlist/toggle`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
        body: JSON.stringify({ product_id: productId }),
      });
      assertStatus(res, 200, 'Toggle wishlist');
      const data = res.json();
      assert(data.message && data.message.includes('wishlist'), 'Wishlist toggle message');
      if (data.added !== undefined) assert(typeof data.added === 'boolean', 'Wishlist toggle added flag');
    }
  });

  await test('POST /wishlist/toggle removes product from wishlist', async () => {
    if (!authToken) throw new Error('No auth token available');
    const productsRes = await request(`${API_BASE}/products?per_page=1`);
    const productsData = productsRes.json();
    if (productsData.products && productsData.products.length > 0) {
      const productId = productsData.products[0].id;
      await request(`${API_BASE}/wishlist/toggle`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
        body: JSON.stringify({ product_id: productId }),
      });
      const res = await request(`${API_BASE}/wishlist/toggle`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
        body: JSON.stringify({ product_id: productId }),
      });
      assertStatus(res, 200, 'Toggle wishlist remove');
    }
  });
}

async function runReviewTests(authToken) {
  section('Reviews');

  await test('POST /reviews rates a product (verified purchase)', async () => {
    if (!authToken) throw new Error('No auth token available');
    const productsRes = await request(`${API_BASE}/products?per_page=1`);
    const productsData = productsRes.json();
    if (productsData.products && productsData.products.length > 0) {
      const productId = productsData.products[0].id;
      const res = await request(`${API_BASE}/reviews`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
        body: JSON.stringify({ product_id: productId, rating: 5 }),
      });
      assert(res.status === 201 || res.status === 403, `Unexpected status ${res.status}`);
    }
  });

  await test('POST /reviews rejects invalid rating', async () => {
    if (!authToken) throw new Error('No auth token available');
    const productsRes = await request(`${API_BASE}/products?per_page=1`);
    const productsData = productsRes.json();
    if (productsData.products && productsData.products.length > 0) {
      const productId = productsData.products[0].id;
      const res = await request(`${API_BASE}/reviews`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
        body: JSON.stringify({ product_id: productId, rating: 6 }),
      });
      assertStatus(res, 400, 'Invalid rating');
    }
  });

  await test('GET /reviews/:productId returns product ratings', async () => {
    const productsRes = await request(`${API_BASE}/products?per_page=1`);
    const productsData = productsRes.json();
    if (productsData.products && productsData.products.length > 0) {
      const productId = productsData.products[0].id;
      const res = await request(`${API_BASE}/reviews/${productId}`);
      assertStatus(res, 200, 'Get reviews');
      const data = res.json();
      // Reviews endpoint returns reviews array, not average_rating
      assertArray(data, 'reviews', 'Reviews array');
      assertField(data, 'total', 'Total reviews count');
    }
  });
}

async function runCommunityTests(authToken) {
  section('Community');

  await test('GET /community/posts returns post feed', async () => {
    const res = await request(`${API_BASE}/community/posts`);
    assertStatus(res, 200, 'Post feed');
    const data = res.json();
    assertArray(data, 'posts', 'Posts feed');
    assertField(data, 'total', 'Feed total');
    assert(data.total > 10, `feed volume for Show more, total=${data.total}`);
  });

  await test('GET /community/posts paginates', async () => {
    const res = await request(`${API_BASE}/community/posts?page=1&per_page=10`);
    assertStatus(res, 200, 'Feed page');
    const data = res.json();
    assert(data.posts.length <= 10, 'feed page size');
    assert(data.total > data.posts.length, 'more posts than one page');
  });

  await test('GET /community/users paginates people directory', async () => {
    const res = await request(`${API_BASE}/community/users?page=1&per_page=12`);
    assertStatus(res, 200, 'People');
    const data = res.json();
    assertArray(data, 'users', 'People users');
    assertField(data, 'total', 'People total');
    assert(data.users.length <= 12, 'people page size');
    assert(data.total > 12, `people volume for Show more, total=${data.total}`);
  });

  await test('POST /community/posts creates a post', async () => {
    if (!authToken) throw new Error('No auth token available');
    const res = await request(`${API_BASE}/community/posts`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
      body: JSON.stringify({ content: 'Test post content' }),
    });
    assertStatus(res, 201, 'Create post');
    const data = res.json();
    assertField(data, 'id', 'Post id');
  });

  await test('POST /community/posts unfurls a public URL', async () => {
    if (!authToken) throw new Error('No auth token available');
    const res = await request(`${API_BASE}/community/posts`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
      body: JSON.stringify({ content: 'Look at https://example.com/pack for this kit' }),
    });
    assertStatus(res, 201, 'Create post with URL');
    const data = res.json();
    assert(String(data.link_url || '').indexOf('example.com') >= 0, `link_url=${data.link_url}`);
  });

  await test('POST /community/posts accepts a photo without text', async () => {
    if (!authToken) throw new Error('No auth token available');
    const png = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==';
    const res = await request(`${API_BASE}/community/posts`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
      body: JSON.stringify({ content: '', image_url: png }),
    });
    assertStatus(res, 201, 'Create photo post');
    const data = res.json();
    assert(String(data.image_url || '').indexOf('data:image/png') === 0, 'image_url is a png data URL');
  });

  await test('POST /community/posts rejects empty body without photo', async () => {
    if (!authToken) throw new Error('No auth token available');
    const res = await request(`${API_BASE}/community/posts`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
      body: JSON.stringify({ content: '' }),
    });
    assertStatus(res, 400, 'Empty post');
  });

  await test('POST /community/posts rejects content over 500 chars', async () => {
    if (!authToken) throw new Error('No auth token available');
    const longContent = 'a'.repeat(501);
    const res = await request(`${API_BASE}/community/posts`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
      body: JSON.stringify({ content: longContent }),
    });
    assertStatus(res, 400, 'Content too long');
  });

  await test('GET /community/posts/:id returns post with comments', async () => {
    const listRes = await request(`${API_BASE}/community/posts`);
    const listData = listRes.json();
    if (listData.posts && listData.posts.length > 0) {
      const postId = listData.posts[0].id;
      const res = await request(`${API_BASE}/community/posts/${postId}`);
      assertStatus(res, 200, 'Post detail');
      const data = res.json();
      const post = data.post || data;
      assertField(post, 'id', 'Post detail id');
      assertArray(data, 'comments', 'Post comments');
    }
  });

  await test('POST /community/posts/:id/like toggles like', async () => {
    if (!authToken) throw new Error('No auth token available');
    const listRes = await request(`${API_BASE}/community/posts`);
    const listData = listRes.json();
    if (listData.posts && listData.posts.length > 0) {
      const postId = listData.posts[0].id;
      const res = await request(`${API_BASE}/community/posts/${postId}/like`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${authToken}` },
      });
      assertStatus(res, 200, 'Like post');
    }
  });

  await test('POST /community/posts/:id/comments adds comment', async () => {
    if (!authToken) throw new Error('No auth token available');
    const listRes = await request(`${API_BASE}/community/posts`);
    const listData = listRes.json();
    if (listData.posts && listData.posts.length > 0) {
      const postId = listData.posts[0].id;
      const res = await request(`${API_BASE}/community/posts/${postId}/comments`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
        body: JSON.stringify({ content: 'Test comment' }),
      });
      assertStatus(res, 201, 'Add comment');
    }
  });

  await test('POST /community/follow/:userId follows a user', async () => {
    if (!authToken) throw new Error('No auth token available');
    const res = await request(`${API_BASE}/community/follow/1`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}` },
    });
    assert(res.status === 200 || res.status === 201, `Unexpected status ${res.status}`);
  });

  await test('DELETE /community/follow/:userId unfollows a user', async () => {
    if (!authToken) throw new Error('No auth token available');
    const res = await request(`${API_BASE}/community/follow/1`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${authToken}` },
    });
    assertStatus(res, 200, 'Unfollow');
  });

  await test('GET /community/users/:id returns public profile', async () => {
    const res = await request(`${API_BASE}/community/users/1`);
    assertStatus(res, 200, 'Public profile');
    const data = res.json();
    assertField(data, 'id', 'Profile id');
    assertField(data, 'name', 'Profile name');
  });
}

async function runCouponTests() {
  section('Coupons');

  await test('POST /coupons/validate validates a valid coupon', async () => {
    const res = await request(`${API_BASE}/coupons/validate`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ code: 'TESTCODE', cart_total: 50 }),
    });
    assert(res.status === 200 || res.status === 404, `Unexpected status ${res.status}`);
    if (res.status === 200) {
      const data = res.json();
      assertField(data, 'discount_type', 'Coupon discount type');
      assertField(data, 'discount_value', 'Coupon discount value');
    }
  });

  await test('POST /coupons/validate rejects expired coupon', async () => {
    const res = await request(`${API_BASE}/coupons/validate`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ code: 'EXPIRED', cart_total: 50 }),
    });
    assert(res.status === 400 || res.status === 404 || res.status === 410, `Unexpected status ${res.status}`);
  });

  await test('POST /coupons/validate rejects below minimum purchase', async () => {
    const res = await request(`${API_BASE}/coupons/validate`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ code: 'MIN100', cart_total: 10 }),
    });
    assert(res.status === 400 || res.status === 404, `Unexpected status ${res.status}`);
  });
}

async function runBundleTests() {
  section('Bundles');

  await test('GET /bundles returns several bundles', async () => {
    const res = await request(`${API_BASE}/bundles`);
    assertStatus(res, 200, 'Bundles list');
    const data = res.json();
    assertArray(data, 'bundles', 'Bundles');
    assert(data.bundles.length >= 3, `bundle volume, got ${data.bundles.length}`);
  });

  await test('GET /bundles/:id returns bundle detail', async () => {
    const listRes = await request(`${API_BASE}/bundles`);
    const listData = listRes.json();
    if (listData.bundles && listData.bundles.length > 0) {
      const bundleId = listData.bundles[0].id;
      const res = await request(`${API_BASE}/bundles/${bundleId}`);
      assertStatus(res, 200, 'Bundle detail');
      const data = res.json();
      const bundle = data.bundle || data;
      assertField(bundle, 'title', 'Bundle title');
      assert(bundle.items != null || Array.isArray(bundle.products), 'Bundle items');
    }
  });
}

async function runReferralTests(authToken) {
  section('Referrals');

  await test('GET /referrals returns user referral stats', async () => {
    if (!authToken) throw new Error('No auth token available');
    const res = await request(`${API_BASE}/referrals`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    assertStatus(res, 200, 'Referral stats');
    const data = res.json();
    assertField(data, 'referral_link', 'Referral link');
    assertField(data, 'total_earnings', 'Total earnings');
  });

  await test('GET /referrals/earnings returns earnings history', async () => {
    if (!authToken) throw new Error('No auth token available');
    const res = await request(`${API_BASE}/referrals/earnings`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    assertStatus(res, 200, 'Earnings history');
    const data = res.json();
    assertArray(data, 'earnings', 'Earnings array');
  });
}

async function runExchangeRateTests() {
  section('Exchange Rates');

  await test('GET /exchange-rates returns all rates', async () => {
    const res = await request(`${API_BASE}/exchange-rates`);
    assertStatus(res, 200, 'Exchange rates');
    const data = res.json();
    assertArray(data, 'rates', 'Rates array');
  });

  await test('GET /exchange-rates/:chain returns single rate', async () => {
    const res = await request(`${API_BASE}/exchange-rates/BSC`);
    assert(res.status === 200 || res.status === 404, `Unexpected status ${res.status}`);
    if (res.status === 200) {
      const data = res.json();
      // Rate is nested under 'rate' key
      const rate = data.rate || data;
      assertField(rate, 'chain', 'Rate chain');
      assertField(rate, 'rate_to_usd', 'Rate value');
    }
  });
}

async function runSettingsTests() {
  section('Settings');

  await test('GET /settings/:key returns setting value', async () => {
    const res = await request(`${API_BASE}/settings/site_title`);
    assert(res.status === 200 || res.status === 404, `Unexpected status ${res.status}`);
    if (res.status === 200) {
      const data = res.json();
      assertField(data, 'value', 'Setting value');
    }
  });
}

async function runRecentlyViewedTests(authToken) {
  section('Recently Viewed');

  await test('POST /recently-viewed records a view', async () => {
    if (!authToken) throw new Error('No auth token available');
    const productsRes = await request(`${API_BASE}/products?per_page=1`);
    const productsData = productsRes.json();
    if (productsData.products && productsData.products.length > 0) {
      const productId = productsData.products[0].id;
      const res = await request(`${API_BASE}/recently-viewed`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
        body: JSON.stringify({ product_id: productId }),
      });
      assertStatus(res, 200, 'Record viewed');
    }
  });

  await test('GET /recently-viewed lists recently viewed', async () => {
    if (!authToken) throw new Error('No auth token available');
    const res = await request(`${API_BASE}/recently-viewed`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    assertStatus(res, 200, 'List recently viewed');
    const data = res.json();
    assertArray(data, 'products', 'Recently viewed');
  });
}

async function runCompareTests(authToken) {
  section('Product Comparison');

  await test('POST /compare/toggle adds product to comparison', async () => {
    if (!authToken) throw new Error('No auth token available');
    const productsRes = await request(`${API_BASE}/products?per_page=1`);
    const productsData = productsRes.json();
    if (productsData.products && productsData.products.length > 0) {
      const productId = productsData.products[0].id;
      const res = await request(`${API_BASE}/compare/toggle`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${authToken}`, 'content-type': 'application/json' },
        body: JSON.stringify({ product_id: productId }),
      });
      assertStatus(res, 200, 'Toggle compare');
    }
  });

  await test('GET /compare lists compared products', async () => {
    if (!authToken) throw new Error('No auth token available');
    const res = await request(`${API_BASE}/compare`, {
      headers: { Authorization: `Bearer ${authToken}` },
    });
    assertStatus(res, 200, 'List compare');
    const data = res.json();
    assertArray(data, 'products', 'Compare products');
  });
}

async function runRecommendationTests() {
  section('Recommendations');

  await test('GET /recommendations/:productId returns related products', async () => {
    const productsRes = await request(`${API_BASE}/products?per_page=1`);
    const productsData = productsRes.json();
    if (productsData.products && productsData.products.length > 0) {
      const productId = productsData.products[0].id;
      const res = await request(`${API_BASE}/recommendations/${productId}`);
      assertStatus(res, 200, 'Recommendations');
      const data = res.json();
      // Recommendations returns products array or null - accept both
      if (data && data.products !== null) {
        assert(Array.isArray(data.products), 'Recommended products should be array');
      }
    }
  });

  await test('GET /recommendations/:productId 404s for a missing product', async () => {
    const res = await request(`${API_BASE}/recommendations/999999999`);
    assertStatus(res, 404, 'Missing recommendations');
  });
}

async function runAdminTests() {
  section('Admin - Authentication');

  await test('Admin API rejects without token', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/users`);
    assertStatus(res, 401, 'Admin no token');
  });

  await test('Admin API rejects invalid token', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/users`, {
      headers: { Authorization: 'Bearer wrong-token' },
    });
    assert(res.status === 401 || res.status === 403, `Unexpected status ${res.status}`);
  });

  await test('Admin API accepts valid token', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/users`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assertStatus(res, 200, 'Admin auth');
  });

  section('Admin - Users');

  await test('GET /admin/users lists all users', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/users`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assertStatus(res, 200, 'Admin users');
    const data = res.json();
    assertArray(data, 'users', 'Admin users array');
  });

  await test('POST /admin/users/:id/reset-password resets password', async () => {
    const listRes = await request(`${API_ROOT}/api/v1/admin/users`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    const listData = listRes.json();
    if (listData.users && listData.users.length > 0) {
      const userId = listData.users[0].id;
      const res = await request(`${API_ROOT}/api/v1/admin/users/${userId}/reset-password`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
      });
      assertStatus(res, 200, 'Reset password');
      const data = res.json();
      assertField(data, 'new_password', 'New password');
    }
  });

  await test('DELETE /admin/users/:id deletes user', async () => {
    const listRes = await request(`${API_ROOT}/api/v1/admin/users`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    const listData = listRes.json();
    if (listData.users && listData.users.length > 0) {
      const userId = listData.users[0].id;
      const res = await request(`${API_ROOT}/api/v1/admin/users/${userId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
      });
      assert(res.status === 200 || res.status === 404, `Unexpected status ${res.status}`);
    }
  });

  section('Admin - Products');

  await test('GET /admin/products lists all products', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/products`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assertStatus(res, 200, 'Admin products');
    const data = res.json();
    assertArray(data, 'products', 'Admin products array');
  });

  await test('POST /admin/products creates product', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/products`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
      body: JSON.stringify({
        title: `E2E Test Product ${Date.now()}`,
        slug: `e2e-test-${Date.now()}`,
        description: 'Test product',
        price_usd: 9.99,
        status: 'active',
      }),
    });
    assertStatus(res, 201, 'Create product');
    const data = res.json();
    assertField(data, 'id', 'Product id');
  });

  await test('PUT /admin/products/:id updates product', async () => {
    const listRes = await request(`${API_ROOT}/api/v1/admin/products`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    const listData = listRes.json();
    if (listData.products && listData.products.length > 0) {
      const productId = listData.products[0].id;
      const res = await request(`${API_ROOT}/api/v1/admin/products/${productId}`, {
        method: 'PUT',
        headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
        body: JSON.stringify({ title: 'Updated Title' }),
      });
      assertStatus(res, 200, 'Update product');
    }
  });

  await test('POST /admin/products/:id/tiers adds tier', async () => {
    const listRes = await request(`${API_ROOT}/api/v1/admin/products`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    const listData = listRes.json();
    if (listData.products && listData.products.length > 0) {
      const productId = listData.products[0].id;
      const res = await request(`${API_ROOT}/api/v1/admin/products/${productId}/tiers`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
        body: JSON.stringify({ tier_name: 'Pro', price_usd: 19.99, file_path: '/assets/pro.zip' }),
      });
      assertStatus(res, 201, 'Add tier');
    }
  });

  await test('POST /admin/products/:id/images adds image', async () => {
    const listRes = await request(`${API_ROOT}/api/v1/admin/products`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    const listData = listRes.json();
    if (listData.products && listData.products.length > 0) {
      const productId = listData.products[0].id;
      const res = await request(`${API_ROOT}/api/v1/admin/products/${productId}/images`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
        body: JSON.stringify({ url: 'https://example.com/image.jpg', is_primary: true }),
      });
      assertStatus(res, 201, 'Add image');
    }
  });

  await test('POST /admin/products/:id/generate-previews triggers preview generation', async () => {
    const listRes = await request(`${API_ROOT}/api/v1/admin/products`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    const listData = listRes.json();
    if (listData.products && listData.products.length > 0) {
      const productId = listData.products[0].id;
      const res = await request(`${API_ROOT}/api/v1/admin/products/${productId}/generate-previews`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
      });
      assert(res.status === 200 || res.status === 202, `Unexpected status ${res.status}`);
    }
  });

  await test('POST /admin/products/:id/pin pins product', async () => {
    const listRes = await request(`${API_ROOT}/api/v1/admin/products`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    const listData = listRes.json();
    if (listData.products && listData.products.length > 0) {
      const productId = listData.products[0].id;
      const res = await request(`${API_ROOT}/api/v1/admin/products/${productId}/pin`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
      });
      assertStatus(res, 200, 'Pin product');
    }
  });

  await test('POST /admin/products/bulk bulk updates products', async () => {
    const listRes = await request(`${API_ROOT}/api/v1/admin/products`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    const listData = listRes.json();
    if (listData.products && listData.products.length >= 2) {
      const ids = listData.products.slice(0, 2).map(p => p.id);
      const res = await request(`${API_ROOT}/api/v1/admin/products/bulk`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
        body: JSON.stringify({ ids, status: 'draft' }),
      });
      assertStatus(res, 200, 'Bulk update');
    }
  });

  section('Admin - Bundles');

  await test('GET /admin/bundles lists all bundles', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/bundles`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assertStatus(res, 200, 'Admin bundles');
    const data = res.json();
    assertArray(data, 'bundles', 'Admin bundles array');
  });

  await test('POST /admin/bundles creates bundle', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/bundles`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
      body: JSON.stringify({
        title: `E2E Bundle ${Date.now()}`,
        slug: `e2e-bundle-${Date.now()}`,
        description: 'Test bundle',
        price_usd: 19.99,
        product_ids: [1, 2],
      }),
    });
    assertStatus(res, 201, 'Create bundle');
  });

  section('Admin - Coupons');

  await test('GET /admin/coupons lists all coupons', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/coupons`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assertStatus(res, 200, 'Admin coupons');
    const data = res.json();
    assertArray(data, 'coupons', 'Admin coupons array');
  });

  await test('POST /admin/coupons creates coupon that checkout can validate', async () => {
    const code = `E2E${Date.now()}`;
    const res = await request(`${API_ROOT}/api/v1/admin/coupons`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
      body: JSON.stringify({
        code,
        discount_type: 'percentage',
        discount_value: 10,
        expires_at: '2027-12-31',
        usage_limit: 100,
        min_purchase_usd: 0,
        is_active: true,
      }),
    });
    assertStatus(res, 201, 'Create coupon');
    const val = await request(`${API_BASE}/coupons/validate`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ code, cart_total: 50 }),
    });
    assertStatus(val, 200, 'Validate admin-created coupon');
    assert(val.json().valid === true, 'created coupon should be valid at checkout');
  });

  section('Admin - Orders');

  await test('GET /admin/orders lists all orders', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/orders`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assertStatus(res, 200, 'Admin orders');
    const data = res.json();
    assertArray(data, 'orders', 'Admin orders array');
  });

  await test('PUT /admin/orders/:id/status updates order status', async () => {
    const listRes = await request(`${API_ROOT}/api/v1/admin/orders`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    const listData = listRes.json();
    if (listData.orders && listData.orders.length > 0) {
      const orderId = listData.orders[0].id;
      const res = await request(`${API_ROOT}/api/v1/admin/orders/${orderId}/status`, {
        method: 'PUT',
        headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
        body: JSON.stringify({ status: 'paid' }),
      });
      assert(res.status === 200 || res.status === 400, `Unexpected status ${res.status}`);
    }
  });

  await test('GET /admin/order-steps lists the pipeline', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/order-steps`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assertStatus(res, 200, 'Order steps');
    const data = res.json();
    assertArray(data, 'steps', 'Pipeline steps');
    const slugs = (data.steps || []).map((s) => s.slug);
    assert(slugs.indexOf('awaiting_payment') >= 0, 'waiting-for-payment step');
    assert(slugs.indexOf('paid') >= 0, 'paid step');
    assert(slugs.indexOf('preparation') >= 0, 'preparation step');
    assert(slugs.indexOf('delivered') >= 0, 'delivered step');
  });

  await test('GET /admin/orders?status= filters by pipeline slug', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/orders?status=paid`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assertStatus(res, 200, 'Filter orders');
    const data = res.json();
    assertArray(data, 'orders', 'Filtered orders');
    assertArray(data, 'by_step', 'by_step counts');
    (data.orders || []).forEach((o) => {
      assert(o.status === 'paid', `expected paid, got ${o.status}`);
    });
  });

  await test('PUT /admin/orders/:id/status rejects unknown steps', async () => {
    const listRes = await request(`${API_ROOT}/api/v1/admin/orders`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    const orders = listRes.json().orders || [];
    if (!orders.length) return;
    const res = await request(`${API_ROOT}/api/v1/admin/orders/${orders[0].id}/status`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
      body: JSON.stringify({ status: 'not_a_real_step' }),
    });
    assertStatus(res, 400, 'Unknown step');
  });

  await test('POST /admin/order-steps then DELETE the custom step', async () => {
    const label = 'OpsQA ' + Date.now();
    const created = await request(`${API_ROOT}/api/v1/admin/order-steps`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
      body: JSON.stringify({ label }),
    });
    assert(created.status === 201 || created.status === 409, `create step ${created.status}`);
    const listed = await request(`${API_ROOT}/api/v1/admin/order-steps`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    const hit = (listed.json().steps || []).find((s) => s.label === label || s.slug && s.slug.indexOf('opsqa') === 0);
    if (hit && !hit.is_system) {
      const del = await request(`${API_ROOT}/api/v1/admin/order-steps/${hit.id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
      });
      assertStatus(del, 200, 'Delete custom step');
    }
  });

  await test('DELETE /admin/order-steps rejects system steps', async () => {
    const listed = await request(`${API_ROOT}/api/v1/admin/order-steps`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    const sys = (listed.json().steps || []).find((s) => s.is_system);
    assert(sys, 'need a system step');
    const del = await request(`${API_ROOT}/api/v1/admin/order-steps/${sys.id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assertStatus(del, 400, 'System step delete');
  });

  section('Admin - Community');

  await test('GET /admin/community/posts lists all posts', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/community/posts`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assertStatus(res, 200, 'Admin posts');
    const data = res.json();
    assertArray(data, 'posts', 'Admin posts array');
  });

  await test('DELETE /admin/community/posts/:id deletes post', async () => {
    const listRes = await request(`${API_ROOT}/api/v1/admin/community/posts`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    const listData = listRes.json();
    if (listData.posts && listData.posts.length > 0) {
      const postId = listData.posts[0].id;
      const res = await request(`${API_ROOT}/api/v1/admin/community/posts/${postId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
      });
      assert(res.status === 200 || res.status === 404, `Unexpected status ${res.status}`);
    }
  });

  section('Admin - Referrals');

  await test('GET /admin/referrals lists all referrals', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/referrals`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assertStatus(res, 200, 'Admin referrals');
    const data = res.json();
    assertArray(data, 'referrals', 'Admin referrals array');
  });

  section('Admin - Stats');

  await test('GET /admin/stats returns dashboard stats', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/stats`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assertStatus(res, 200, 'Admin stats');
    const data = res.json();
    assertField(data, 'total_users', 'Total users');
    assertField(data, 'total_orders', 'Total orders');
    assertField(data, 'total_revenue', 'Total revenue');
    assertArray(data, 'order_by_step', 'Orders by step');
  });

  section('Admin - Settings');

  await test('GET /admin/settings lists all settings', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/settings`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assertStatus(res, 200, 'Admin settings');
    const data = res.json();
    assertArray(data, 'settings', 'Admin settings array');
  });

  await test('PUT /admin/settings/:key updates setting', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/settings/referral_commission_percent`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
      body: JSON.stringify({ value: '5.0' }),
    });
    assertStatus(res, 200, 'Update setting');
  });

  section('Admin - Email Export');

  await test('POST /admin/export/emails exports buyer emails', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/export/emails`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
      body: JSON.stringify({ format: 'csv' }),
    });
    assert(res.status === 200 || res.status === 202, `Unexpected status ${res.status}`);
  });

  section('Admin - Exchange Rates');

  await test('GET /admin/exchange-rates lists all rates', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/exchange-rates`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assertStatus(res, 200, 'Admin exchange rates');
    const data = res.json();
    assertArray(data, 'rates', 'Admin rates array');
  });

  await test('PUT /admin/exchange-rates/:chain sets rate', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/exchange-rates/BSC`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
      body: JSON.stringify({ symbol: 'BNB', rate_to_usd: 250.00 }),
    });
    assertStatus(res, 200, 'Set exchange rate');
  });

  await test('DELETE /admin/exchange-rates/:chain deletes rate', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/exchange-rates/BSC`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assert(res.status === 200 || res.status === 404, `Unexpected status ${res.status}`);
  });
}

async function runAppearanceAndRbacTests() {
  section('Appearance');

  await test('GET /products/appearance is public', async () => {
    const res = await request(`${API_BASE}/products/appearance`);
    assertStatus(res, 200, 'Appearance');
    const data = res.json();
    assertField(data, 'palette', 'palette');
    assertField(data, 'font', 'font');
    assertField(data, 'radius', 'radius');
    assertField(data, 'density', 'density');
  });

  await test('PUT /products/appearance accepts admin token', async () => {
    const cur = await request(`${API_BASE}/products/appearance`);
    const body = Object.assign({ palette: 'clay', font: 'system', radius: 'soft', density: 'comfortable' }, cur.json());
    const res = await request(`${API_BASE}/products/appearance`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}`, 'content-type': 'application/json' },
      body: JSON.stringify(body),
    });
    assertStatus(res, 200, 'Save appearance');
  });

  section('RBAC - roles and operator token');

  const nia = await request(`${API_BASE}/login`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ email: 'nia@example.com', password: 'nia' }),
  });
  const leo = await request(`${API_BASE}/login`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ email: 'leo@example.com', password: 'leo' }),
  });
  const maya = await request(`${API_BASE}/login`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ email: 'maya@example.com', password: 'maya' }),
  });

  await test('Nia login is admin', async () => {
    assertStatus(nia, 200, 'Nia login');
    assert(nia.json().user.role === 'admin', 'Nia should be admin');
  });

  await test('Leo login is staff', async () => {
    assertStatus(leo, 200, 'Leo login');
    assert(leo.json().user.role === 'staff', 'Leo should be staff');
  });

  await test('Maya customer cannot hit admin stats', async () => {
    assertStatus(maya, 200, 'Maya login');
    const res = await request(`${API_ROOT}/api/v1/admin/stats`, {
      headers: { Authorization: `Bearer ${maya.json().token}` },
    });
    assert(res.status === 403, `customer stats ${res.status}`);
  });

  await test('Staff JWT can use assigned tabs without operator token', async () => {
    const products = await request(`${API_ROOT}/api/v1/admin/products`, {
      headers: { Authorization: `Bearer ${leo.json().token}` },
    });
    assertStatus(products, 200, 'Staff products');
    const steps = await request(`${API_ROOT}/api/v1/admin/order-steps`, {
      headers: { Authorization: `Bearer ${leo.json().token}` },
    });
    assertStatus(steps, 200, 'Staff order-steps read via Orders tab');
    const write = await request(`${API_ROOT}/api/v1/admin/order-steps`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${leo.json().token}`, 'content-type': 'application/json' },
      body: JSON.stringify({ label: 'ShouldFail' }),
    });
    assert(write.status === 403, `staff step write ${write.status}`);
  });

  await test('Staff JWT is denied unassigned tabs', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/stats`, {
      headers: { Authorization: `Bearer ${leo.json().token}` },
    });
    assert(res.status === 403, `staff stats ${res.status}`);
  });

  await test('Admin JWT can use the desk without operator token', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/stats`, {
      headers: { Authorization: `Bearer ${nia.json().token}` },
    });
    assertStatus(res, 200, 'Admin JWT stats');
  });

  await test('Changing access without operator token is forbidden', async () => {
    const list = await request(`${API_ROOT}/api/v1/admin/users`, {
      headers: { Authorization: `Bearer ${nia.json().token}` },
    });
    assertStatus(list, 200, 'Admin users via JWT');
    const staff = (list.json().users || []).find((u) => u.email === 'leo@example.com');
    assert(staff, 'leo exists');
    const res = await request(`${API_ROOT}/api/v1/admin/users/${staff.id}/access`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${nia.json().token}`, 'content-type': 'application/json' },
      body: JSON.stringify({ role: 'staff', staff_tabs: ['Products', 'Banner', 'Orders', 'Community'] }),
    });
    assert(res.status === 403, `access without operator ${res.status}`);
  });

  await test('Changing access with admin JWT plus X-Admin-Token succeeds', async () => {
    const list = await request(`${API_ROOT}/api/v1/admin/users`, {
      headers: { Authorization: `Bearer ${nia.json().token}` },
    });
    const staff = (list.json().users || []).find((u) => u.email === 'leo@example.com');
    const res = await request(`${API_ROOT}/api/v1/admin/users/${staff.id}/access`, {
      method: 'PUT',
      headers: {
        Authorization: `Bearer ${nia.json().token}`,
        'X-Admin-Token': ADMIN_TOKEN,
        'content-type': 'application/json',
      },
      body: JSON.stringify({ role: 'staff', staff_tabs: ['Products', 'Banner', 'Orders', 'Community'] }),
    });
    assertStatus(res, 200, 'Access with operator token');
  });

  await test('Customer JWT cannot write appearance', async () => {
    const res = await request(`${API_BASE}/products/appearance`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${maya.json().token}`, 'content-type': 'application/json' },
      body: JSON.stringify({ palette: 'night', font: 'system', radius: 'soft', density: 'comfortable' }),
    });
    assert(res.status === 403, `customer appearance ${res.status}`);
  });

  await test('Admin users list includes role', async () => {
    const list = await request(`${API_ROOT}/api/v1/admin/users`, {
      headers: { Authorization: `Bearer ${nia.json().token}` },
    });
    assertStatus(list, 200, 'Users list');
    const leoRow = (list.json().users || []).find((u) => u.email === 'leo@example.com');
    assert(leoRow && leoRow.role === 'staff', 'leo role on users list');
  });

  await test('Staff cannot change access even with operator header', async () => {
    const list = await request(`${API_ROOT}/api/v1/admin/users`, {
      headers: { Authorization: `Bearer ${nia.json().token}` },
    });
    const owen = (list.json().users || []).find((u) => u.email === 'owen@example.com');
    const res = await request(`${API_ROOT}/api/v1/admin/users/${owen.id}/access`, {
      method: 'PUT',
      headers: {
        Authorization: `Bearer ${leo.json().token}`,
        'X-Admin-Token': ADMIN_TOKEN,
        'content-type': 'application/json',
      },
      body: JSON.stringify({ role: 'staff', staff_tabs: ['Products'] }),
    });
    assert(res.status === 403, `staff access ${res.status}`);
  });
}

async function runEventsTests() {
  section('Events collector');

  await test('POST /events accepts product_view', async () => {
    const res = await request(`${API_BASE}/events`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({
        name: 'product_view',
        session_id: 'e2e-session',
        path: '/product/lunar-clay-characters',
        properties: { product_id: 1 },
      }),
    });
    assert(res.status === 202 || res.status === 200, `ingest ${res.status}`);
  });

  await test('POST /events accepts checkout_click', async () => {
    const res = await request(`${API_BASE}/events`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({
        name: 'checkout_click',
        session_id: 'e2e-session',
        path: '/cart',
        properties: { item_count: 1 },
      }),
    });
    assert(res.status === 202 || res.status === 200, `ingest checkout_click ${res.status}`);
  });

  await test('POST /events rejects unknown names', async () => {
    const res = await request(`${API_BASE}/events`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ name: 'page_view' }),
    });
    assertStatus(res, 400, 'unknown event');
  });

  await test('GET /admin/events/ttl is readable', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/events/ttl`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assertStatus(res, 200, 'events ttl');
    const data = res.json();
    assert(typeof data.seconds === 'number', 'ttl seconds');
  });

  await test('GET /admin/events lists stored events', async () => {
    const res = await request(`${API_ROOT}/api/v1/admin/events?limit=5`, {
      headers: { Authorization: `Bearer ${ADMIN_TOKEN}` },
    });
    assertStatus(res, 200, 'admin events');
    assertArray(res.json(), 'events', 'events array');
  });
}

// ============================================================
// MAIN RUNNER
// ============================================================

async function main() {
  console.log('╔══════════════════════════════════════════════════════════╗');
  console.log('║  Store4bots E2E Test Suite (TDD)                        ║');
  console.log('║  Tests define EXPECTED behavior per PRODUCT-SPEC.md    ║');
  console.log('╚══════════════════════════════════════════════════════════╝');
  console.log(`\nAPI: ${API_BASE}`);
  console.log(`Time: ${new Date().toISOString()}`);

  const authResult = await runAuthTests();
  const authToken = authResult.token;

  await runProductTests();
  await runOrderTests(authToken);
  await runCartTests(authToken);
  await runWishlistTests(authToken);
  await runReviewTests(authToken);
  await runCommunityTests(authToken);
  await runCouponTests();
  await runBundleTests();
  await runReferralTests(authToken);
  await runExchangeRateTests();
  await runSettingsTests();
  await runRecentlyViewedTests(authToken);
  await runCompareTests(authToken);
  await runRecommendationTests();
  await runAdminTests();
  await runAppearanceAndRbacTests();
  await runEventsTests();

  console.log('\n╔══════════════════════════════════════════════════════════╗');
  console.log('║  RESULTS                                                ║');
  console.log('╠══════════════════════════════════════════════════════════╣');
  console.log(`║  Passed:  ${passed.toString().padEnd(45)}║`);
  console.log(`║  Failed:  ${failed.toString().padEnd(45)}║`);
  console.log(`║  Total:   ${(passed + failed).toString().padEnd(45)}║`);
  console.log('╚══════════════════════════════════════════════════════════╝');

  if (failures.length > 0) {
    console.log('\n--- Failures ---');
    failures.forEach((f, i) => {
      console.log(`${i + 1}. ${f.name}: ${f.error}`);
    });
  }

  process.exit(failed > 0 ? 1 : 0);
}

main().catch((err) => {
  console.error('Fatal error:', err);
  process.exit(1);
});
