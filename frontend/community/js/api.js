// Pawradise API Client
(function() {
  const hostname = window.location.hostname;
  const isStaging = hostname.startsWith('staging') || window.location.pathname.startsWith('/staging/');
  const API_BASE = isStaging ? '/staging/api/v1' : '/api/v1';

  async function apiFetch(endpoint, options = {}) {
    const token = localStorage.getItem('token');
    const defaults = {
      headers: { 'Content-Type': 'application/json' }
    };
    if (token) defaults.headers['Authorization'] = 'Bearer ' + token;
    const config = { ...defaults, ...options };
    config.headers = { ...defaults.headers, ...(options.headers || {}) };
    const resp = await fetch(API_BASE + endpoint, config);
    if (resp.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = isStaging ? '/staging/auth/' : '/auth/';
      return;
    }
    return resp;
  }

  window.API = {
    base: API_BASE,
    isStaging: isStaging,
    get: (endpoint) => apiFetch(endpoint),
    post: (endpoint, body) => apiFetch(endpoint, { method: 'POST', body: JSON.stringify(body) }),
    put: (endpoint, body) => apiFetch(endpoint, { method: 'PUT', body: JSON.stringify(body) }),
    del: (endpoint) => apiFetch(endpoint, { method: 'DELETE' }),
    getToken: () => localStorage.getItem('token'),
    getUser: () => JSON.parse(localStorage.getItem('user') || '{}'),
    setAuth: (token, user) => {
      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(user));
    },
    clearAuth: () => {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
    }
  };
})();
