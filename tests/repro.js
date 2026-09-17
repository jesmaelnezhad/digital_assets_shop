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
  const res = await request('http://127.0.0.1:30084');
  console.log('status', res.status);
  console.log('csp', res.headers['content-security-policy']);
  console.log('has eval-ish in js', /new Function|setTimeout\(|setInterval\(|eval\(/.test(res.body));
})();
