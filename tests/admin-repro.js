import http from 'http';

function request(url, opts = {}) {
  return new Promise((resolve, reject) => {
    const req = http.request(url, opts, (res) => {
      let data = '';
      res.on('data', (c) => (data += c));
      res.on('end', () =>
        resolve({
          status: res.statusCode,
          headers: res.headers,
          body: data,
        })
      );
    });
    req.on('error', reject);
    req.end();
  });
}

(async () => {
  console.log('--- Health ---');
  try {
    const r = await request('http://127.0.0.1:30083/health');
    console.log('status', r.status);
    console.log('body', r.body);
  } catch (e) {
    console.log('ERROR', e.message);
  }

  console.log('--- Admin Users ---');
  try {
    const r = await request('http://127.0.0.1:30083/admin/users', {
      headers: { Authorization: 'Bearer admin-secret-token-change-in-production' },
    });
    console.log('status', r.status);
    console.log('body', r.body);
  } catch (e) {
    console.log('ERROR', e.message);
  }

  console.log('--- Reset Password ---');
  try {
    const r = await request('http://127.0.0.1:30083/admin/users/1/reset-password', {
      method: 'POST',
      headers: { Authorization: 'Bearer admin-secret-token-change-in-production' },
    });
    console.log('status', r.status);
    console.log('body', r.body);
  } catch (e) {
    console.log('ERROR', e.message);
  }
})();
