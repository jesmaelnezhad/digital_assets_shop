// Pawradise API Client — classic script (no ES module exports)
// Reads environment from window.__PAWRADISE_ENV__ (injected at deploy time)

(function() {
    const ENV = window.__PAWRADISE_ENV__ || { apiBase: '/api/v1', envName: 'production' };

    class ApiError extends Error {
        constructor(status, message) {
            super(message);
            this.status = status;
        }
    }

    async function apiFetch(path, options = {}) {
        const res = await fetch(`${ENV.apiBase}${path}`, {
            credentials: 'include',
            headers: { 'Content-Type': 'application/json', ...options.headers },
            ...options,
        });
        if (!res.ok) {
            const body = await res.json().catch(() => ({}));
            throw new ApiError(res.status, body.error || res.statusText);
        }
        return res.status === 204 ? null : res.json();
    }

    const api = {
        async get(path) { return apiFetch(path); },
        async post(path, data) { return apiFetch(path, { method: 'POST', body: JSON.stringify(data) }); },
        async put(path, data) { return apiFetch(path, { method: 'PUT', body: JSON.stringify(data) }); },
        async delete(path) { return apiFetch(path, { method: 'DELETE' }); },

        auth: {
            register: (email, password, name) => apiFetch('/register', { method: 'POST', body: JSON.stringify({ email, password, name }) }),
            login: (email, password) => apiFetch('/login', { method: 'POST', body: JSON.stringify({ email, password }) }),
            logout: () => apiFetch('/logout', { method: 'POST' }),
            getProfile: () => apiFetch('/me'),
            updateProfile: (data) => apiFetch('/me', { method: 'PUT', body: JSON.stringify(data) }),
        },

        products: {
            list: (params = {}) => apiFetch('/products?' + new URLSearchParams(params).toString()),
            get: (slug) => apiFetch(`/products/${slug}`),
            getCategories: () => apiFetch('/categories'),
            getBundles: () => apiFetch('/bundles'),
            getBundle: (id) => apiFetch(`/bundles/${id}`),
            getRecommendations: (productId) => apiFetch(`/recommendations/${productId}`),
        },

        orders: {
            list: (params = {}) => apiFetch('/orders?' + new URLSearchParams(params).toString()),
            get: (id) => apiFetch(`/orders/${id}`),
            getStatus: (id) => apiFetch(`/orders/${id}/status`),
            getPayment: (id) => apiFetch(`/orders/${id}/payment`),
            create: (items) => apiFetch('/orders', { method: 'POST', body: JSON.stringify(items) }),
            createGuestOrder: (email, items) => apiFetch('/guest-orders', { method: 'POST', body: JSON.stringify({ email, items }) }),
            getGuestOrder: (id) => apiFetch(`/guest-orders/${id}`),
        },

        cart: {
            get: () => apiFetch('/cart'),
            add: (productId, quantity = 1, tierId = null) => apiFetch('/cart/items', { method: 'POST', body: JSON.stringify({ product_id: productId, quantity, tier_id: tierId }) }),
            remove: (itemId) => apiFetch(`/cart/items/${itemId}`, { method: 'DELETE' }),
        },

        wishlist: {
            get: () => apiFetch('/wishlist'),
            toggle: (productId) => apiFetch('/wishlist/toggle', { method: 'POST', body: JSON.stringify({ product_id: productId }) }),
        },

        coupons: {
            validate: (code, cartTotal) => apiFetch('/coupons/validate', { method: 'POST', body: JSON.stringify({ code, cart_total: cartTotal }) }),
        },

        community: {
            getPosts: (filter = 'recent', page = 1, perPage = 20) => apiFetch(`/community/posts?filter=${filter}&page=${page}&per_page=${perPage}`),
            getPost: (id) => apiFetch(`/community/posts/${id}`),
            createPost: (content, type = 'post') => apiFetch('/community/posts', { method: 'POST', body: JSON.stringify({ content, type, is_public: true }) }),
            likePost: (id) => apiFetch(`/community/posts/${id}/like`, { method: 'POST' }),
            addComment: (postId, content) => apiFetch(`/community/posts/${postId}/comments`, { method: 'POST', body: JSON.stringify({ content }) }),
            follow: (userId) => apiFetch(`/community/follow/${userId}`, { method: 'POST' }),
            unfollow: (userId) => apiFetch(`/community/follow/${userId}`, { method: 'DELETE' }),
            getProfile: (userId) => apiFetch(`/community/users/${userId}`),
        },

        reviews: {
            create: (productId, rating, comment) => apiFetch('/reviews', { method: 'POST', body: JSON.stringify({ product_id: productId, rating, comment }) }),
            getForProduct: (productId, page = 1, perPage = 20) => apiFetch(`/reviews/${productId}?page=${page}&per_page=${perPage}`),
        },

        exchangeRates: {
            getAll: () => apiFetch('/exchange-rates'),
            get: (chain) => apiFetch(`/exchange-rates/${chain}`),
        },

        recentlyViewed: {
            list: () => apiFetch('/recently-viewed'),
            record: (productId) => apiFetch('/recently-viewed', { method: 'POST', body: JSON.stringify({ product_id: productId }) }),
        },

        compare: {
            list: () => apiFetch('/compare'),
            toggle: (productId) => apiFetch('/compare/toggle', { method: 'POST', body: JSON.stringify({ product_id: productId }) }),
        },

        referrals: {
            get: () => apiFetch('/referrals'),
            getEarnings: () => apiFetch('/referrals/earnings'),
            track: (code) => apiFetch('/referrals/track', { method: 'POST', body: JSON.stringify({ referral_code: code }) }),
        },

        settings: {
            get: (key) => apiFetch(`/settings/${key}`),
        },
    };

    // Expose on window for HTML scripts
    window.Pawradise = window.Pawradise || {};
    window.Pawradise.api = api;
    window.Pawradise.ENV = ENV;
})();
