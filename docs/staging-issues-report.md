# Pawradise Staging — Routing & Issues Report

**Date:** 2026-09-13  
**Staging URL:** `https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir`

---

## Architecture: Traffic Flow (4 Layers)

```
Internet
  ↓
Layer 1: Cloud LB (130.185.123.156 → NodePort 30758)
  ↓
Layer 2: ingress-nginx controller (watches Ingress resources)
  ↓
Layer 3: Ingress Resources (path-based routing to Services)
  ↓
Layer 4: Service (ClusterIP 80/8081-8088) → Pod NGINX → Files / Go Service
```

---

## Layer 1: Cloud Load Balancer

| Setting | Value |
|---------|-------|
| External IP | `130.185.123.156` |
| NodePort | `30758` (TCP) |
| Protocol | HTTP (80) |

---

## Layer 2: Ingress-NGINX Controller

**Service:** `ingress-nginx-controller.ingress-nginx:80`  
**Type:** NodePort (`30758`)  
**Class:** `nginx` (ingress-nginx 1.15.1)

---

## Layer 3: Ingress Resources

### Frontend Ingress (`frontend-ingress-staging`)

| Path | Service | Pod Port | MFE |
|------|---------|----------|-----|
| `/` | shop-mfe | 80 | Shop homepage |
| `/login` | auth-mfe | 80 | Login page |
| `/register` | auth-mfe | 80 | Register page |
| `/auth` | auth-mfe | 80 | Auth legacy paths |
| `/product` | product-mfe | 80 | Product detail |
| `/account` | account-mfe | 80 | Account dashboard |
| `/cart` | account-mfe | 80 | Cart |
| `/wishlist` | account-mfe | 80 | Wishlist |
| `/referrals` | account-mfe | 80 | Referrals |
| `/checkout` | checkout-mfe | 80 | Checkout |
| `/community` | community-mfe | 80 | Community feed |
| `/post` | community-mfe | 80 | Single post |
| `/profile` | community-mfe | 80 | User profile |
| `/admin` | admin-app | 80 | Admin panel |

### Backend API Ingress (`api-gateway-staging`)

| Path | Service | Port |
|------|---------|------|
| `/api/v1/register` | identity-service | 8081 |
| `/api/v1/login` | identity-service | 8081 |
| `/api/v1/logout` | identity-service | 8081 |
| `/api/v1/me` | identity-service | 8081 |
| `/api/v1/referrals` | identity-service | 8081 |
| `/api/v1/commissions` | identity-service | 8081 |
| `/api/v1/referrals/earnings` | identity-service | 8081 |
| `/api/v1/products` | product-service | 8082 |
| `/api/v1/categories` | product-service | 8082 |
| `/api/v1/bundles` | product-service | 8082 |
| `/api/v1/recommendations` | product-service | 8082 |
| `/api/v1/guest-orders` | commerce-service | 8083 |
| `/api/v1/coupons` | commerce-service | 8083 |
| `/api/v1/orders` | commerce-service | 8083 |
| `/api/v1/cart` | commerce-service | 8083 |
| `/api/v1/wishlist` | commerce-service | 8083 |
| `/api/v1/recently-viewed` | commerce-service | 8083 |
| `/api/v1/compare` | commerce-service | 8083 |
| `/api/v1/posts` | community-service | 8084 |
| `/api/v1/community` | community-service | 8084 |
| `/api/v1/users` | community-service | 8084 |
| `/api/v1/profile` | community-service | 8084 |
| `/api/v1/reviews` | review-service | 8085 |
| `/api/v1/exchange-rates` | payment-service | 8086 |
| `/api/v1/payments` | payment-service | 8086 |
| `/api/v1/settings` | payment-service | 8086 |
| `/api/v1/admin` | admin-service | 8087 |
| `/api/v1/media` | media-service | 8088 |

---

## Layer 4: Pod Configurations

### shop-mfe (`shop-mfe-7f5ddd9894-2mm5l`)

**Image:** `pawradise/shop-mfe:staging-fix-v5`  
**Docroot:** `/usr/share/nginx/html`  
**Files:** `index.html`, `category.html`, `home.html`, `assets/{alpine.min.js, api.js, app.js, theme.css, favicon.svg, env.js}`

```nginx
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html index.htm;

    location / {
        try_files $uri $uri.html $uri/ /index.html;
    }
}
```

**Nav menu (from `index.html`):**
```html
<nav class="nav">
    <a href="/">Shop</a>
    <a href="/category">Categories</a>
    <a href="/product/bundle.html">Bundles</a>
    <a href="/product/request.html">Request</a>
    <a href="/community">Community</a>
</nav>
```

**Note:** No Account link in nav.

---

### auth-mfe (`auth-mfe-764d7cb9db-x8bsh`)

**Image:** `pawradise/auth-mfe:staging-fix-v2`  
**Docroot:** `/usr/share/nginx/html`  
**Files:** `login.html`, `register.html`, `index.html`, `assets/*`

```nginx
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html index.htm;

    location / {
        try_files $uri $uri.html $uri/ /login.html /register.html =404;
    }
}
```

**Nav menu (from `login.html`):**
```html
<nav class="nav">
    <a href="/">Shop</a>
    <a href="/community">Community</a>
    <a href="/register">Sign Up</a>
</nav>
```

---

### product-mfe (`product-mfe-7cfdd98cd7-2sqvd`)

**Image:** `pawradise/product-mfe:staging-fix-v3`  
**Docroot:** `/usr/share/nginx/html`  
**Files:** `product.html`, `index.html`, `assets/*`, `product/{bundle.html, product.html, request.html}`

```nginx
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html index.htm;

    location /product/ {
        alias /usr/share/nginx/html/product/;
        try_files $uri $uri.html $uri/ /product/product.html;
    }

    location / {
        try_files $uri $uri.html $uri/ /index.html;
    }
}
```

---

### account-mfe (`account-mfe-555778f9d-t76b7`)

**Image:** `pawradise/account-mfe:latest`  
**Docroot:** `/usr/share/nginx/html`  
**Files:** `account.html`, `cart.html`, `wishlist.html`, `referrals.html`, `index.html`, `assets/*`

```nginx
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html index.htm;

    location / {
        try_files $uri $uri.html $uri/ /index.html;
    }
}
```

**Nav menu (from `account.html`):**
```html
<nav class="nav">
    <a href="/">Shop</a>
    <a href="/community">Community</a>
    <a href="/account" class="active">Account</a>
    <a href="/admin">Admin</a>
</nav>
```

---

### checkout-mfe (`checkout-mfe-784786df9b-2kntr`)

**Image:** `pawradise/checkout-mfe:latest`  
**Docroot:** `/usr/share/nginx/html`

```nginx
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html index.htm;

    location / {
        try_files $uri $uri.html $uri/ /index.html;
    }
}
```

**Nav menu:**
```html
<nav class="nav">
    <a href="/">Shop</a>
    <a href="/community">Community</a>
    <a href="/account">Account</a>
    <a href="/admin">Admin</a>
</nav>
```

---

### community-mfe (`community-mfe-75ff499977-pr2sf`)

**Image:** `pawradise/community-mfe:latest`  
**Docroot:** `/usr/share/nginx/html`  
**Files:** `community.html`, `profile.html`, `post.html`, `index.html`, `assets/*`

```nginx
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html index.htm;

    location / {
        try_files $uri $uri.html $uri/ /index.html;
    }
}
```

**Nav menu (from `community.html`):**
```html
<nav class="nav">
    <a href="/">Shop</a>
    <a href="/community" class="active">Community</a>
    <a href="/account">Account</a>
    <a href="/admin">Admin</a>
</nav>
```

---

### admin-app (`admin-app-77b98fc9b-p7zm5`)

**Image:** `pawradise/admin-app:latest`  
**Docroot:** `/usr/share/nginx/html`

```nginx
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html index.htm;

    location / {
        try_files $uri $uri.html $uri/ /index.html;
    }
}
```

**Nav menu (custom):**
```html
<nav>
    <a class="nav-item" :class="{ active: tab === 'stats' }" @click="tab='stats'">Dashboard</a>
    <a class="nav-item" :class="{ active: tab === 'products' }" @click="tab='products'; loadProducts()">Products</a>
    <a class="nav-item" :class="{ active: tab === 'bundles' }" @click="tab='bundles'; loadBundles()">Bundles</a>
    <a class="nav-item" :class="{ active: tab === 'coupons' }" @click="tab='coupons'; loadCoupons()">Coupons</a>
    <a class="nav-item" :class="{ active: tab === 'orders' }" @click="tab='orders'; loadOrders()">Orders</a>
</nav>
```

---

Current pass/fail counts live in [`tests/REPORT.md`](../tests/REPORT.md). This file is the routing/architecture snapshot from 2026-09-13, not the live test report.

---

## Configuration Files (Copy-Paste for Proofreading)

### Frontend Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: frontend-ingress-staging
  namespace: staging
spec:
  ingressClassName: nginx
  rules:
  - host: server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir
    http:
      paths:
      - path: /          # → shop-mfe
      - path: /login     # → auth-mfe
      - path: /register  # → auth-mfe
      - path: /auth      # → auth-mfe
      - path: /product   # → product-mfe
      - path: /account   # → account-mfe
      - path: /cart      # → account-mfe
      - path: /wishlist  # → account-mfe
      - path: /referrals # → account-mfe
      - path: /checkout  # → checkout-mfe
      - path: /community # → community-mfe
      - path: /post      # → community-mfe
      - path: /profile   # → community-mfe
      - path: /admin     # → admin-app
```

### Backend API Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-gateway-staging
  namespace: staging
spec:
  ingressClassName: nginx
  rules:
  - host: server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir
    http:
      paths:
      - path: /api/v1/register
      - path: /api/v1/login
      - path: /api/v1/logout
      - path: /api/v1/me
      - path: /api/v1/referrals
      - path: /api/v1/commissions
      - path: /api/v1/referrals/earnings
      - path: /api/v1/products
      - path: /api/v1/categories
      - path: /api/v1/bundles
      - path: /api/v1/recommendations
      - path: /api/v1/guest-orders
      - path: /api/v1/coupons
      - path: /api/v1/orders
      - path: /api/v1/cart
      - path: /api/v1/wishlist
      - path: /api/v1/recently-viewed
      - path: /api/v1/compare
      - path: /api/v1/posts
      - path: /api/v1/community
      - path: /api/v1/users
      - path: /api/v1/profile
      - path: /api/v1/reviews
      - path: /api/v1/exchange-rates
      - path: /api/v1/payments
      - path: /api/v1/settings
      - path: /api/v1/admin
      - path: /api/v1/media
```

### Services

```yaml
# Frontend
shop-mfe:         ClusterIP 10.43.17.53    :80  (selector: app=shop-mfe)
auth-mfe:         ClusterIP 10.43.247.183  :80  (selector: app=auth-mfe)
product-mfe:      ClusterIP 10.43.218.22   :80  (selector: app=product-mfe)
account-mfe:      ClusterIP 10.43.174.102  :80  (selector: app=account-mfe)
checkout-mfe:     ClusterIP 10.43.129.9    :80  (selector: app=checkout-mfe)
community-mfe:    ClusterIP 10.43.147.9    :80  (selector: app=community-mfe)
admin-app:        ClusterIP 10.43.91.221   :80  (selector: app=admin-app)

# Backend
identity-service:  ClusterIP 10.43.79.21    :8081
product-service:   ClusterIP 10.43.125.177  :8082
commerce-service:  ClusterIP 10.43.134.17   :8083
community-service: ClusterIP 10.43.98.87    :8084
review-service:    ClusterIP 10.43.12.119   :8085
payment-service:   ClusterIP 10.43.35.161   :8086
admin-service:     ClusterIP 10.43.176.20   :8087
media-service:     ClusterIP 10.43.120.84   :8088
```
