import http from 'http';

const BACKEND = 'http://127.0.0.1:30083';
const ADMIN = 'admin-secret-token-change-in-production';

function request(url, opts = {}) {
  return new Promise((resolve, reject) => {
    const parsed = new URL(url);
    const reqOpts = {
      hostname: parsed.hostname,
      port: parsed.port,
      path: parsed.pathname + parsed.search,
      method: opts.method || 'GET',
      headers: opts.headers || {},
    };
    let bodyStr = null;
    if (opts.body) {
      bodyStr = JSON.stringify(opts.body);
      reqOpts.headers['Content-Type'] = 'application/json';
      reqOpts.headers['Content-Length'] = Buffer.byteLength(bodyStr);
    }
    const req = http.request(reqOpts, (res) => {
      let data = '';
      res.on('data', (c) => (data += c));
      res.on('end', () => resolve({ status: res.statusCode, body: data }));
    });
    req.on('error', reject);
    if (bodyStr) req.write(bodyStr);
    req.end();
  });
}

function json(body) { try { return JSON.parse(body); } catch { return null; } }

async function main() {
  // Create a test user
  console.log('1. Creating test user...');
  const register = json((await request(BACKEND + '/api/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: { email: 'delete-test@example.com', password: 'Test123!', name: 'Delete Test' }
  })).body);
  
  if (!register || !register.user) {
    console.log('Failed to create:', register?.error);
    return;
  }
  const userId = register.user.id;
  console.log(`   Created user ID: ${userId}`);

  // Verify exists
  const list1 = json((await request(BACKEND + '/admin/users', {
    headers: { Authorization: `Bearer ${ADMIN}` }
  })).body);
  const found = list1.users.find(u => u.id === userId);
  console.log(`2. User found in list: ${found ? 'yes' : 'no'}`);

  // Delete
  console.log('3. Deleting user...');
  const del = json((await request(BACKEND + `/admin/users/${userId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${ADMIN}` }
  })).body);
  console.log(`   Response: ${del.message || del.error}`);

  // Verify gone
  const list2 = json((await request(BACKEND + '/admin/users', {
    headers: { Authorization: `Bearer ${ADMIN}` }
  })).body);
  const stillThere = list2.users.find(u => u.id === userId);
  console.log(`4. User still in list: ${stillThere ? 'yes' : 'no'}`);

  if (!stillThere) console.log('\n✓ Delete works correctly');
  else console.log('\n✗ User still present after delete');
}

main().catch(console.error);
