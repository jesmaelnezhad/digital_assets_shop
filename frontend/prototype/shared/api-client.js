// Copied from shared/lib/api.js and extended for prototype admin + a few missing routes.
// Keep method names and paths identical to the live client.

(function() {
    const ENV = window.__PAWRADISE_ENV__ || { apiBase: '/api/v1', envName: 'prototype' };

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
            if (res.status === 401 && path.indexOf('/admin/') === 0) {
                try {
                    sessionStorage.removeItem('admin_token');
                    localStorage.removeItem('pawradise_admin_token');
                } catch (e) { /* ignore */ }
            }
            throw new ApiError(res.status, body.error || res.statusText);
        }
        const ct = res.headers.get('content-type') || '';
        if (res.status === 204) return null;
        if (ct.indexOf('application/json') === -1) return res;
        return res.json();
    }

    function adminHeaders() {
        const t = (typeof sessionStorage !== 'undefined' && sessionStorage.getItem('admin_token'))
            || localStorage.getItem('pawradise_admin_token')
            || '';
        return { Authorization: 'Bearer ' + t };
    }

    const api = {
        async get(path) { return apiFetch(path); },
        async post(path, data) { return apiFetch(path, { method: 'POST', body: JSON.stringify(data) }); },
        async put(path, data) { return apiFetch(path, { method: 'PUT', body: JSON.stringify(data) }); },
        async delete(path) { return apiFetch(path, { method: 'DELETE' }); },

        auth: {
            register: (email, password, name, referral_code) => apiFetch('/register', { method: 'POST', body: JSON.stringify({ email, password, name, referral_code }) }),
            login: (email, password) => apiFetch('/login', { method: 'POST', body: JSON.stringify({ email, password }) }),
            logout: () => apiFetch('/logout', { method: 'POST' }),
            getProfile: () => apiFetch('/me'),
            updateProfile: (data) => apiFetch('/me', { method: 'PUT', body: JSON.stringify(data) }),
            getPublic: (id) => apiFetch('/profile/' + id),
        },

        products: {
            list: (params = {}) => apiFetch('/products?' + new URLSearchParams(params).toString()),
            get: (slug) => apiFetch(`/products/${slug}`),
            getCategories: () => apiFetch('/categories'),
            createCategory: (data) => apiFetch('/categories', { method: 'POST', body: JSON.stringify(data), headers: adminHeaders() }),
            updateCategory: (id, data) => apiFetch('/categories/' + id, { method: 'PUT', body: JSON.stringify(data), headers: adminHeaders() }),
            deleteCategory: (id) => apiFetch('/categories/' + id, { method: 'DELETE', headers: adminHeaders() }),
            getBundles: () => apiFetch('/bundles'),
            getBundle: (id) => apiFetch(`/bundles/${id}`),
            getTiers: (id) => apiFetch(`/products/${id}/tiers`),
            getRecommendations: (productId) => apiFetch(`/recommendations/${productId}`),
        },

        orders: {
            list: (params = {}) => apiFetch('/orders?' + new URLSearchParams(params).toString()),
            get: (id) => apiFetch(`/orders/${id}`),
            getStatus: (id) => apiFetch(`/orders/${id}/status`),
            getPayment: (id) => apiFetch(`/orders/${id}/payment`),
            create: (body) => apiFetch('/orders', { method: 'POST', body: JSON.stringify(body) }),
            confirm: (id) => apiFetch(`/orders/${id}/confirm`, { method: 'POST' }),
            download: (id, itemId) => apiFetch(`/orders/${id}/download/${itemId}`),
            createGuestOrder: (email, items) => apiFetch('/guest-orders', { method: 'POST', body: JSON.stringify({ email, items }) }),
            getGuestOrder: (id, email) => apiFetch(`/guest-orders/${id}?email=${encodeURIComponent(email || '')}`),
        },

        cart: {
            get: () => apiFetch('/cart'),
            add: (productId, quantity = 1, tierId = null, extra = {}) => apiFetch('/cart/items', { method: 'POST', body: JSON.stringify({ product_id: productId, quantity, tier_id: tierId, ...extra }) }),
            remove: (itemId) => apiFetch(`/cart/items/${itemId}`, { method: 'DELETE' }),
            setQty: (itemId, quantity) => apiFetch(`/cart/items/${itemId}`, { method: 'PUT', body: JSON.stringify({ quantity }) }),
        },

        wishlist: {
            get: () => apiFetch('/wishlist'),
            toggle: (productId) => apiFetch('/wishlist/toggle', { method: 'POST', body: JSON.stringify({ product_id: productId }) }),
        },

        coupons: {
            validate: (code, cartTotal) => apiFetch('/coupons/validate', { method: 'POST', body: JSON.stringify({ code, cart_total: cartTotal }) }),
        },

        community: {
            getPosts: (filter = 'recent', page = 1, perPage = 20) => {
                const params = typeof filter === 'object' ? filter : { filter, page, per_page: perPage };
                return apiFetch('/community/posts?' + new URLSearchParams(params).toString());
            },
            getPost: (id) => apiFetch(`/community/posts/${id}`),
            createPost: (content, type = 'post') => apiFetch('/community/posts', { method: 'POST', body: JSON.stringify({ content, type, is_public: true }) }),
            likePost: (id) => apiFetch(`/community/posts/${id}/like`, { method: 'POST' }),
            unlikePost: (id) => apiFetch(`/community/posts/${id}/like`, { method: 'DELETE' }),
            addComment: (postId, content) => apiFetch(`/community/posts/${postId}/comments`, { method: 'POST', body: JSON.stringify({ content }) }),
            deleteComment: (postId, commentId) => apiFetch(`/community/posts/${postId}/comments/${commentId}`, { method: 'DELETE' }),
            follow: (userId) => apiFetch(`/community/follow/${userId}`, { method: 'POST' }),
            unfollow: (userId) => apiFetch(`/community/follow/${userId}`, { method: 'DELETE' }),
            getProfile: (userId) => apiFetch(`/community/users/${userId}`),
            listPeople: (params) => apiFetch('/community/users' + (params ? '?' + new URLSearchParams(params).toString() : '')),
            getFollowers: (userId, params) => apiFetch(`/community/users/${userId}/followers` + (params ? '?' + new URLSearchParams(params).toString() : '')),
            getFollowing: (userId, params) => apiFetch(`/community/users/${userId}/following` + (params ? '?' + new URLSearchParams(params).toString() : '')),
            suggestions: () => apiFetch('/community/suggestions'),
        },

        reviews: {
            create: (productId, rating, comment) => apiFetch('/reviews', { method: 'POST', body: JSON.stringify({ product_id: productId, rating, comment }) }),
            getForProduct: (productId, page = 1, perPage = 20) => apiFetch(`/reviews/${productId}?page=${page}&per_page=${perPage}`),
            average: (productId) => apiFetch(`/reviews/${productId}/average`),
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

        payments: {
            get: (orderId) => apiFetch(`/payments/order/${orderId}`),
            status: (orderId) => apiFetch(`/payments/order/${orderId}/status`),
            confirm: (orderId) => apiFetch(`/payments/order/${orderId}/confirm`, { method: 'POST' }),
        },

        settings: {
            get: (key) => apiFetch(`/settings/${key}`),
        },

        requests: {
            create: (data) => apiFetch('/product-requests', { method: 'POST', body: JSON.stringify(data) }),
        },

        admin: {
            stats: () => apiFetch('/admin/stats', { headers: adminHeaders() }),
            users: () => apiFetch('/admin/users', { headers: adminHeaders() }),
            deleteUser: (id) => apiFetch('/admin/users/' + id, { method: 'DELETE', headers: adminHeaders() }),
            resetPassword: (id) => apiFetch('/admin/users/' + id + '/reset-password', { method: 'POST', headers: adminHeaders() }),
            products: () => apiFetch('/admin/products', { headers: adminHeaders() }),
            createProduct: (data) => apiFetch('/admin/products', { method: 'POST', body: JSON.stringify(data), headers: adminHeaders() }),
            updateProduct: (id, data) => apiFetch('/admin/products/' + id, { method: 'PUT', body: JSON.stringify(data), headers: adminHeaders() }),
            pin: (id) => apiFetch('/admin/products/' + id + '/pin', { method: 'POST', headers: adminHeaders() }),
            unpin: (id) => apiFetch('/admin/products/' + id + '/pin', { method: 'DELETE', headers: adminHeaders() }),
            deleteProduct: (id) => apiFetch('/admin/products/' + id, { method: 'DELETE', headers: adminHeaders() }),
            addTier: (id, data) => apiFetch('/admin/products/' + id + '/tiers', { method: 'POST', body: JSON.stringify(data), headers: adminHeaders() }),
            updateTier: (id, tierId, data) => apiFetch('/admin/products/' + id + '/tiers/' + tierId, { method: 'PUT', body: JSON.stringify(data), headers: adminHeaders() }),
            deleteTier: (id, tierId) => apiFetch('/admin/products/' + id + '/tiers/' + tierId, { method: 'DELETE', headers: adminHeaders() }),
            addImage: (id, data) => apiFetch('/admin/products/' + id + '/images', { method: 'POST', body: JSON.stringify(data), headers: adminHeaders() }),
            deleteImage: (id, imageId) => apiFetch('/admin/products/' + id + '/images/' + imageId, { method: 'DELETE', headers: adminHeaders() }),
            generatePreviews: (id) => apiFetch('/admin/products/' + id + '/generate-previews', { method: 'POST', headers: adminHeaders() }),
            bulkProducts: (data) => apiFetch('/admin/products/bulk', { method: 'POST', body: JSON.stringify(data), headers: adminHeaders() }),
            orders: () => apiFetch('/admin/orders', { headers: adminHeaders() }),
            getOrder: (id) => apiFetch('/admin/orders/' + id, { headers: adminHeaders() }),
            setOrderStatus: (id, status) => apiFetch('/admin/orders/' + id + '/status', { method: 'PUT', body: JSON.stringify({ status }), headers: adminHeaders() }),
            guestOrders: () => apiFetch('/admin/guest-orders', { headers: adminHeaders() }),
            getGuestOrder: (id) => apiFetch('/admin/guest-orders/' + id, { headers: adminHeaders() }),
            communityPosts: () => apiFetch('/admin/community/posts', { headers: adminHeaders() }),
            deletePost: (id) => apiFetch('/admin/community/posts/' + id, { method: 'DELETE', headers: adminHeaders() }),
            settings: () => apiFetch('/admin/settings', { headers: adminHeaders() }),
            setSetting: (key, value) => apiFetch('/admin/settings/' + key, { method: 'PUT', body: JSON.stringify({ value }), headers: adminHeaders() }),
            referrals: () => apiFetch('/admin/referrals', { headers: adminHeaders() }),
            exportEmails: (data = {}) => apiFetch('/admin/export/emails', { method: 'POST', body: JSON.stringify(data), headers: adminHeaders() }),
            bundles: () => apiFetch('/admin/bundles', { headers: adminHeaders() }),
            createBundle: (data) => apiFetch('/admin/bundles', { method: 'POST', body: JSON.stringify(data), headers: adminHeaders() }),
            updateBundle: (id, data) => apiFetch('/admin/bundles/' + id, { method: 'PUT', body: JSON.stringify(data), headers: adminHeaders() }),
            deleteBundle: (id) => apiFetch('/admin/bundles/' + id, { method: 'DELETE', headers: adminHeaders() }),
            coupons: () => apiFetch('/admin/coupons', { headers: adminHeaders() }),
            createCoupon: (data) => apiFetch('/admin/coupons', { method: 'POST', body: JSON.stringify(data), headers: adminHeaders() }),
            updateCoupon: (id, data) => apiFetch('/admin/coupons/' + id, { method: 'PUT', body: JSON.stringify(data), headers: adminHeaders() }),
            deleteCoupon: (id) => apiFetch('/admin/coupons/' + id, { method: 'DELETE', headers: adminHeaders() }),
            rates: () => apiFetch('/admin/exchange-rates', { headers: adminHeaders() }),
            setRate: (chain, data) => apiFetch('/admin/exchange-rates/' + chain, { method: 'PUT', body: JSON.stringify(data), headers: adminHeaders() }),
            deleteRate: (chain) => apiFetch('/admin/exchange-rates/' + chain, { method: 'DELETE', headers: adminHeaders() }),
            requests: () => apiFetch('/admin/product-requests', { headers: adminHeaders() }),
            setRequestStatus: (id, status) => apiFetch('/admin/product-requests/' + id, { method: 'PUT', body: JSON.stringify({ status }), headers: adminHeaders() }),
            categories: () => apiFetch('/admin/categories', { headers: adminHeaders() }),
            createCategory: (data) => apiFetch('/categories', { method: 'POST', body: JSON.stringify(data), headers: adminHeaders() }),
            updateCategory: (id, data) => apiFetch('/categories/' + id, { method: 'PUT', body: JSON.stringify(data), headers: adminHeaders() }),
            deleteCategory: (id) => apiFetch('/categories/' + id, { method: 'DELETE', headers: adminHeaders() }),
            events: (params = {}) => apiFetch('/admin/events?' + new URLSearchParams(params).toString(), { headers: adminHeaders() }),
            eventsTtl: () => apiFetch('/admin/events/ttl', { headers: adminHeaders() }),
            setEventsTtl: (data) => apiFetch('/admin/events/ttl', { method: 'PUT', body: JSON.stringify(data), headers: adminHeaders() }),
        }
    };

    window.Pawradise = window.Pawradise || {};
    window.Pawradise.api = api;
    window.Pawradise.ENV = ENV;
    window.Pawradise.ApiError = ApiError;
})();
