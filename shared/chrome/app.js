// Shared app.js — hydrates auth/cart badges after page load
// Uses Alpine.js for reactivity

document.addEventListener('DOMContentLoaded', () => {
    // Check auth state via cookie (httpOnly, so we can't read it directly)
    // Instead, we make a lightweight call to /me to check auth
    fetch(`${window.__STORE4BOTS_ENV__?.apiBase || '/api/v1'}/me`, {
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' }
    })
    .then(res => {
        if (res.ok) return res.json();
        throw new Error('not authed');
    })
    .then(user => {
        // Update header for logged-in state
        const authBtn = document.getElementById('authBtn');
        const signupBtn = document.getElementById('signupBtn');
        const userBadge = document.getElementById('userBadge');
        if (authBtn) {
            authBtn.textContent = 'Logout';
            authBtn.href = '#';
            authBtn.onclick = (e) => {
                e.preventDefault();
                fetch(`${window.__STORE4BOTS_ENV__?.apiBase || '/api/v1'}/logout`, {
                    method: 'POST',
                    credentials: 'include',
                    headers: { 'Content-Type': 'application/json' }
                }).then(() => window.location.href = '/');
            };
        }
        if (signupBtn) signupBtn.style.display = 'none';
        if (userBadge) {
            userBadge.style.display = 'inline-flex';
            userBadge.textContent = user.name?.[0]?.toUpperCase() || '?';
        }
    })
    .catch(() => {
        // Not logged in — keep default state
    });

    // Load cart badge count
    fetch(`${window.__STORE4BOTS_ENV__?.apiBase || '/api/v1'}/cart`, {
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' }
    })
    .then(res => res.ok ? res.json() : null)
    .then(cart => {
        if (cart?.items?.length > 0) {
            const badge = document.getElementById('cartBadge');
            if (badge) {
                badge.textContent = cart.items.length;
                badge.style.display = 'inline-flex';
            }
        }
    })
    .catch(() => {});
});
