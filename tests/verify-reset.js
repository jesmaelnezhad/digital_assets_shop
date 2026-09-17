import http from 'http';

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
    if (opts.body) {
      const bodyStr = typeof opts.body === 'string' ? opts.body : JSON.stringify(opts.body);
      reqOpts.headers['Content-Length'] = Buffer.byteLength(bodyStr);
      reqOpts.headers['Content-Type'] = 'application/json';
      reqOpts.body = bodyStr;
    }
    const req = http.request(reqOpts, (res) => {
      let data = '';
      res.on('data', (c) => (data += c));
      res.on('end', () => resolve({ status: res.statusCode, body: data }));
    });
    req.on('error', reject);
    if (reqOpts.body) req.write(reqOpts.body);
    req.end();
  });
}

function json(body) { try { return JSON.parse(body); } catch { return null; } }

const BACKEND = 'http://127.0.0.1:30083';
const ADMIN = 'admin-secret-token-change-in-production';

async function main() {
  // 1. List users
  const list = json((await request(BACKEND + '/admin/users', {
    headers: { Authorization: `Bearer ${ADMIN}` }
  })).body);
  const user = list.users[0];
  console.log(`User: ${user.email} (ID ${user.id})`);

  // 2. Reset password
  const reset = json((await request(BACKEND + `/admin/users/${user.id}/reset-password`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${ADMIN}`,
      'Content-Type': 'application/json'
    }
  })).body);
  console.log(`New password: ${reset.new_password}`);

  // 3. Login with new password
  const login = json((await request(BACKEND + '/api/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: { email: user.email, password: reset.new_password }
  })).body);

  if (login && login.token) {
    console.log('✓ Login successful with reset password');
  } else {
    console.log('✗ Login failed:', login?.error || 'no token');
  }
}

main().catch(console.error);
