# Staging Routing — Complete Configuration Reference

> **Date:** 2026-09-13  
> **Staging URL:** `https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir`  
> **Host IP:** `130.185.123.156`

This document contains the **actual running configuration** of every routing layer from the internet to the pod. Copy-paste ready for proofreading.

---

## Layer 0: DNS / Internet

```
server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir
  → 130.185.123.156 (Cloud VM)
  → Port 80 (HTTP)
```

---

## Layer 1: Cloud Load Balancer / VM Entry

The VM at `130.185.123.156` runs **k3s** with `ingress-nginx`. Traffic arrives on port 80 and is forwarded to the ingress-nginx controller service.

**Service:** `ingress-nginx-controller.ingress-nginx`

```yaml
apiVersion: v1
kind: Service
metadata:
  name: ingress-nginx-controller
  namespace: ingress-nginx
spec:
  type: NodePort
  ports:
  - name: http
    port: 80
    protocol: TCP
    targetPort: 80
    nodePort: 30758    # ← External access via 130.185.123.156:30758
  selector:
    app.kubernetes.io/component: controller
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/name: ingress-nginx
status:
  loadBalancer:
    ingress:
    - ip: 130.185.123.156
```

**Effect:** `http://server-ad5ae8ea-...:80` → ingress-nginx controller → reads Ingress resources → routes to Service.

---

## Layer 2: Ingress Resources (K8s)

Two Ingress resources match the same host. The ingress controller merges them.

### 2a. Frontend Ingress (MFE routing)

**Name:** `frontend-ingress-staging`  
**Namespace:** `staging`  
**Class:** `nginx`

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
      - path: /
        pathType: Prefix
        backend:
          service:
            name: shop-mfe
            port:
              number: 80
      - path: /login
        pathType: Prefix
        backend:
          service:
            name: auth-mfe
            port:
              number: 80
      - path: /register
        pathType: Prefix
        backend:
          service:
            name: auth-mfe
            port:
              number: 80
      - path: /auth
        pathType: Prefix
        backend:
          service:
            name: auth-mfe
            port:
              number: 80
      - path: /product
        pathType: Prefix
        backend:
          service:
            name: product-mfe
            port:
              number: 80
      - path: /account
        pathType: Prefix
        backend:
          service:
            name: account-mfe
            port:
              number: 80
      - path: /cart
        pathType: Prefix
        backend:
          service:
            name: account-mfe
            port:
              number: 80
      - path: /wishlist
        pathType: Prefix
        backend:
          service:
            name: account-mfe
            port:
              number: 80
      - path: /referrals
        pathType: Prefix
        backend:
          service:
            name: account-mfe
            port:
              number: 80
      - path: /checkout
        pathType: Prefix
        backend:
          service:
            name: checkout-mfe
            port:
              number: 80
      - path: /community
        pathType: Prefix
        backend:
          service:
            name: community-mfe
            port:
              number: 80
      - path: /post
        pathType: Prefix
        backend:
          service:
            name: community-mfe
            port:
              number: 80
      - path: /profile
        pathType: Prefix
        backend:
          service:
            name: community-mfe
            port:
              number: 80
      - path: /admin
        pathType: Prefix
        backend:
          service:
            name: admin-app
            port:
              number: 80
```

### 2b. Backend API Ingress (Service routing)

**Name:** `api-gateway-staging`  
**Namespace:** `staging`  
**Class:** `nginx`

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
      # === Identity Service (port 8081) ===
      - path: /api/v1/register
        pathType: Prefix
        backend:
          service:
            name: identity-service
            port:
              number: 8081
      - path: /api/v1/login
        pathType: Prefix
        backend:
          service:
            name: identity-service
            port:
              number: 8081
      - path: /api/v1/logout
        pathType: Prefix
        backend:
          service:
            name: identity-service
            port:
              number: 8081
      - path: /api/v1/me
        pathType: Prefix
        backend:
          service:
            name: identity-service
            port:
              number: 8081
      - path: /api/v1/referrals
        pathType: Prefix
        backend:
          service:
            name: identity-service
            port:
              number: 8081
      - path: /api/v1/commissions
        pathType: Prefix
        backend:
          service:
            name: identity-service
            port:
              number: 8081
      - path: /api/v1/referrals/earnings
        pathType: Prefix
        backend:
          service:
            name: identity-service
            port:
              number: 8081
      # === Product Service (port 8082) ===
      - path: /api/v1/products
        pathType: Prefix
        backend:
          service:
            name: product-service
            port:
              number: 8082
      - path: /api/v1/categories
        pathType: Prefix
        backend:
          service:
            name: product-service
            port:
              number: 8082
      - path: /api/v1/bundles
        pathType: Prefix
        backend:
          service:
            name: product-service
            port:
              number: 8082
      # === Commerce Service (port 8083) ===
      - path: /api/v1/guest-orders
        pathType: Prefix
        backend:
          service:
            name: commerce-service
            port:
              number: 8083
      - path: /api/v1/coupons
        pathType: Prefix
        backend:
          service:
            name: commerce-service
            port:
              number: 8083
      - path: /api/v1/orders
        pathType: Prefix
        backend:
          service:
            name: commerce-service
            port:
              number: 8083
      - path: /api/v1/cart
        pathType: Prefix
        backend:
          service:
            name: commerce-service
            port:
              number: 8083
      - path: /api/v1/wishlist
        pathType: Prefix
        backend:
          service:
            name: commerce-service
            port:
              number: 8083
      - path: /api/v1/recently-viewed
        pathType: Prefix
        backend:
          service:
            name: commerce-service
            port:
              number: 8083
      - path: /api/v1/compare
        pathType: Prefix
        backend:
          service:
            name: commerce-service
            port:
              number: 8083
      # === Community Service (port 8084) ===
      - path: /api/v1/posts
        pathType: Prefix
        backend:
          service:
            name: community-service
            port:
              number: 8084
      - path: /api/v1/community
        pathType: Prefix
        backend:
          service:
            name: community-service
            port:
              number: 8084
      - path: /api/v1/users
        pathType: Prefix
        backend:
          service:
            name: community-service
            port:
              number: 8084
      - path: /api/v1/profile
        pathType: Prefix
        backend:
          service:
            name: community-service
            port:
              number: 8084
      # === Review Service (port 8085) ===
      - path: /api/v1/reviews
        pathType: Prefix
        backend:
          service:
            name: review-service
            port:
              number: 8085
      # === Payment Service (port 8086) ===
      - path: /api/v1/exchange-rates
        pathType: Prefix
        backend:
          service:
            name: payment-service
            port:
              number: 8086
      - path: /api/v1/payments
        pathType: Prefix
        backend:
          service:
            name: payment-service
            port:
              number: 8086
      - path: /api/v1/settings
        pathType: Prefix
        backend:
          service:
            name: payment-service
            port:
              number: 8086
      # === Admin Service (port 8087) ===
      - path: /api/v1/admin
        pathType: Prefix
        backend:
          service:
            name: admin-service
            port:
              number: 8087
      # === Media Service (port 8088) ===
      - path: /api/v1/media
        pathType: Prefix
        backend:
          service:
            name: media-service
            port:
              number: 8088
```

---

## Layer 3: Services (ClusterIP → Pod)

### Frontend MFEs

```yaml
# shop-mfe
apiVersion: v1
kind: Service
metadata:
  name: shop-mfe
  namespace: staging
spec:
  type: ClusterIP
  clusterIP: 10.43.17.53
  ports:
  - port: 80
    targetPort: 80
  selector:
    app: shop-mfe       # ← No namespace: here

# auth-mfe
apiVersion: v1
kind: Service
metadata:
  name: auth-mfe
  namespace: staging
spec:
  type: ClusterIP
  clusterIP: 10.43.247.183
  ports:
  - port: 80
    targetPort: 80
  selector:
    app: auth-mfe

# product-mfe
apiVersion: v1
kind: Service
metadata:
  name: product-mfe
  namespace: staging
spec:
  type: ClusterIP
  clusterIP: 10.43.218.22
  ports:
  - port: 80
    targetPort: 80
  selector:
    app: product-mfe

# account-mfe
apiVersion: v1
kind: Service
metadata:
  name: account-mfe
  namespace: staging
spec:
  type: ClusterIP
  clusterIP: 10.43.174.102
  ports:
  - port: 80
    targetPort: 80
  selector:
    app: account-mfe

# checkout-mfe
apiVersion: v1
kind: Service
metadata:
  name: checkout-mfe
  namespace: staging
spec:
  type: ClusterIP
  clusterIP: 10.43.129.9
  ports:
  - port: 80
    targetPort: 80
  selector:
    app: checkout-mfe

# community-mfe
apiVersion: v1
kind: Service
metadata:
  name: community-mfe
  namespace: staging
spec:
  type: ClusterIP
  clusterIP: 10.43.147.9
  ports:
  - port: 80
    targetPort: 80
  selector:
    app: community-mfe

# admin-app
apiVersion: v1
kind: Service
metadata:
  name: admin-app
  namespace: staging
spec:
  type: ClusterIP
  clusterIP: 10.43.91.221
  ports:
  - port: 80
    targetPort: 80
  selector:
    app: admin-app
```

### Backend Services

```yaml
# identity-service
apiVersion: v1
kind: Service
metadata:
  name: identity-service
  namespace: staging
spec:
  type: ClusterIP
  clusterIP: 10.43.79.21
  ports:
  - port: 8081
    targetPort: 8081
  selector:
    app: identity-service

# product-service
apiVersion: v1
kind: Service
metadata:
  name: product-service
  namespace: staging
spec:
  type: ClusterIP
  clusterIP: 10.43.125.177
  ports:
  - port: 8082
    targetPort: 8082
  selector:
    app: product-service

# commerce-service
apiVersion: v1
kind: Service
metadata:
  name: commerce-service
  namespace: staging
spec:
  type: ClusterIP
  clusterIP: 10.43.134.17
  ports:
  - port: 8083
    targetPort: 8083
  selector:
    app: commerce-service

# community-service
apiVersion: v1
kind: Service
metadata:
  name: community-service
  namespace: staging
spec:
  type: ClusterIP
  clusterIP: 10.43.98.87
  ports:
  - port: 8084
    targetPort: 8084
  selector:
    app: community-service

# review-service
apiVersion: v1
kind: Service
metadata:
  name: review-service
  namespace: staging
spec:
  type: ClusterIP
  clusterIP: 10.43.12.119
  ports:
  - port: 8085
    targetPort: 8085
  selector:
    app: review-service

# payment-service
apiVersion: v1
kind: Service
metadata:
  name: payment-service
  namespace: staging
spec:
  type: ClusterIP
  clusterIP: 10.43.35.161
  ports:
  - port: 8086
    targetPort: 8086
  selector:
    app: payment-service

# admin-service
apiVersion: v1
kind: Service
metadata:
  name: admin-service
  namespace: staging
spec:
  type: ClusterIP
  clusterIP: 10.43.176.20
  ports:
  - port: 8087
    targetPort: 8087
  selector:
    app: admin-service

# media-service
apiVersion: v1
kind: Service
metadata:
  name: media-service
  namespace: staging
spec:
  type: ClusterIP
  clusterIP: 10.43.120.84
  ports:
  - port: 8088
    targetPort: 8088
  selector:
    app: media-service
```

---

## Layer 4: Pod NGINX (Inside Container)

Each MFE runs its own nginx inside the container. This handles SPA fallback.

### shop-mfe

**Pod:** `shop-mfe-7f5ddd9894-2mm5l`  
**Image:** `pawradise/shop-mfe:staging-fix-v5`  
**Docroot:** `/usr/share/nginx/html`  
**Files:** `index.html`, `category.html`, `home.html`, `assets/{alpine.min.js, api.js, app.js, theme.css, favicon.svg, env.js}`

```nginx
server {
    listen       80;
    listen  [::]:80;
    server_name  localhost;

    root   /usr/share/nginx/html;
    index  index.html index.htm;

    # Serve static files directly, fall back to index.html for SPA routing
    location / {
        try_files $uri $uri.html $uri/ /index.html;
    }

    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }
}
```

**Nav (from `index.html`):**
```html
<nav class="nav">
    <a href="/">Shop</a>
    <a href="/category">Categories</a>
    <a href="/product/bundle.html">Bundles</a>
    <a href="/product/request.html">Request</a>
    <a href="/community">Community</a>
</nav>
```

---

### auth-mfe

**Pod:** `auth-mfe-764d7cb9db-x8bsh`  
**Image:** `pawradise/auth-mfe:staging-fix-v2`  
**Docroot:** `/usr/share/nginx/html`  
**Files:** `login.html`, `register.html`, `index.html`, `assets/*`

```nginx
server {
    listen       80;
    listen  [::]:80;
    server_name  localhost;

    root   /usr/share/nginx/html;
    index  index.html index.htm;

    # Auth-MFE: serve login.html and register.html directly at root;
    # the ingress routes /auth/*, /login, /register to this pod.
    location / {
        try_files $uri $uri.html $uri/ /login.html /register.html =404;
    }

    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }
}
```

**Nav (from `login.html`):**
```html
<nav class="nav">
    <a href="/">Shop</a>
    <a href="/community">Community</a>
    <a href="/register">Sign Up</a>
</nav>
```

---

### product-mfe

**Pod:** `product-mfe-7cfdd98cd7-2sqvd`  
**Image:** `pawradise/product-mfe:staging-fix-v3`  
**Docroot:** `/usr/share/nginx/html`  
**Files:** `product.html`, `index.html`, `assets/*`  
**Subdirectory:** `product/{bundle.html, product.html, request.html}`

```nginx
server {
    listen       80;
    listen  [::]:80;
    server_name  localhost;

    root   /usr/share/nginx/html;
    index  index.html index.htm;

    # Product detail pages: served from /product/ subdirectory
    location /product/ {
        alias /usr/share/nginx/html/product/;
        try_files $uri $uri.html $uri/ /product/product.html;
    }

    location / {
        try_files $uri $uri.html $uri/ /index.html;
    }

    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root /usr/share/nginx/html;
    }
}
```

---

### account-mfe

**Pod:** `account-mfe-555778f9d-t76b7`  
**Image:** `pawradise/account-mfe:latest`  
**Docroot:** `/usr/share/nginx/html`  
**Files:** `account.html`, `cart.html`, `wishlist.html`, `referrals.html`, `index.html`, `assets/*`

```nginx
server {
    listen       80;
    listen  [::]:80;
    server_name  localhost;

    root   /usr/share/nginx/html;
    index  index.html index.htm;

    location / {
        try_files $uri $uri.html $uri/ /index.html;
    }

    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }
}
```

**Nav (from `account.html`):**
```html
<nav class="nav">
    <a href="/">Shop</a>
    <a href="/community">Community</a>
    <a href="/account" class="active">Account</a>
    <a href="/admin">Admin</a>
</nav>
<nav class="sidebar-nav">
    <a href="#" :class="{ active: activeTab === 'profile' }" @click.prevent="activeTab = 'profile'">Profile</a>
    <a href="#" :class="{ active: activeTab === 'orders' }" @click.prevent="activeTab = 'orders'">Order History</a>
    <a href="/cart">Cart</a>
    <a href="/wishlist">Wishlist</a>
    <a href="/referrals">Referrals</a>
</nav>
```

---

### checkout-mfe

**Pod:** `checkout-mfe-784786df9b-2kntr`  
**Image:** `pawradise/checkout-mfe:latest`  
**Docroot:** `/usr/share/nginx/html`

```nginx
server {
    listen       80;
    listen  [::]:80;
    server_name  localhost;

    root   /usr/share/nginx/html;
    index  index.html index.htm;

    location / {
        try_files $uri $uri.html $uri/ /index.html;
    }

    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }
}
```

**Nav:**
```html
<nav class="nav">
    <a href="/">Shop</a>
    <a href="/community">Community</a>
    <a href="/account">Account</a>
    <a href="/admin">Admin</a>
</nav>
```

---

### community-mfe

**Pod:** `community-mfe-75ff499977-pr2sf`  
**Image:** `pawradise/community-mfe:latest`  
**Docroot:** `/usr/share/nginx/html`  
**Files:** `community.html`, `profile.html`, `post.html`, `index.html`, `assets/*`

```nginx
server {
    listen       80;
    listen  [::]:80;
    server_name  localhost;

    root   /usr/share/nginx/html;
    index  index.html index.htm;

    location / {
        try_files $uri $uri.html $uri/ /index.html;
    }

    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }
}
```

**Nav (from `community.html`):**
```html
<nav class="nav">
    <a href="/">Shop</a>
    <a href="/community" class="active">Community</a>
    <a href="/account">Account</a>
    <a href="/admin">Admin</a>
</nav>
```

---

### admin-app

**Pod:** `admin-app-77b98fc9b-p7zm5`  
**Image:** `pawradise/admin-app:latest`  
**Docroot:** `/usr/share/nginx/html`

```nginx
server {
    listen       80;
    listen  [::]:80;
    server_name  localhost;

    root   /usr/share/nginx/html;
    index  index.html index.htm;

    location / {
        try_files $uri $uri.html $uri/ /index.html;
    }

    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }
}
```

**Nav (custom tab-based):**
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

## Layer 5: Backend Services (Go/Gin)

Each backend service is a Go binary listening on a specific port. They use Gin framework with their own route handlers.

### identity-service (port 8081)

```go
// main.go
r := gin.Default()
api := r.Group("/api/v1")
{
    api.POST("/register", handlers.Register)
    api.POST("/login", handlers.Login)
    api.POST("/logout", handlers.Logout)
    api.GET("/me", middleware.AuthRequired(), handlers.GetMe)
    api.PUT("/me", middleware.AuthRequired(), handlers.UpdateMe)
    api.GET("/referrals", middleware.AuthRequired(), handlers.GetReferrals)
    api.GET("/commissions", middleware.AuthRequired(), handlers.GetCommissions)
    api.GET("/referrals/earnings", middleware.AuthRequired(), handlers.GetEarnings)
}
r.Run(":8081")
```

### product-service (port 8082)

```go
r := gin.Default()
api := r.Group("/api/v1")
{
    api.GET("/products", handlers.ListProducts)
    api.GET("/products/:slug", handlers.GetProduct)
    api.GET("/categories", handlers.ListCategories)
    api.GET("/bundles", handlers.ListBundles)
    api.GET("/bundles/:id", handlers.GetBundle)
    api.GET("/recommendations/:productId", handlers.GetRecommendations)
}
r.Run(":8082")
```

### commerce-service (port 8083)

```go
r := gin.Default()
api := r.Group("/api/v1")
{
    api.GET("/cart", handlers.GetCart)
    api.POST("/cart/items", handlers.AddToCart)
    api.DELETE("/cart/items/:id", handlers.RemoveFromCart)
    api.GET("/wishlist", handlers.GetWishlist)
    api.POST("/wishlist/toggle", handlers.ToggleWishlist)
    api.GET("/orders", middleware.AuthRequired(), handlers.ListOrders)
    api.POST("/orders", middleware.AuthRequired(), handlers.CreateOrder)
    api.GET("/orders/:id", handlers.GetOrder)
    api.POST("/guest-orders", handlers.CreateGuestOrder)
    api.POST("/coupons/validate", handlers.ValidateCoupon)
    api.GET("/recently-viewed", handlers.GetRecentlyViewed)
    api.POST("/recently-viewed", handlers.RecordRecentlyViewed)
    api.GET("/compare", handlers.GetCompare)
    api.POST("/compare/toggle", handlers.ToggleCompare)
}
r.Run(":8083")
```

### community-service (port 8084)

```go
r := gin.Default()
api := r.Group("/api/v1")
{
    api.GET("/community/posts", handlers.GetPosts)
    api.POST("/community/posts", middleware.AuthRequired(), handlers.CreatePost)
    api.POST("/community/posts/:id/like", middleware.AuthRequired(), handlers.LikePost)
    api.POST("/community/posts/:id/comments", middleware.AuthRequired(), handlers.AddComment)
    api.GET("/community/users/:id", handlers.GetCommunityUser)
    api.POST("/community/follow/:id", middleware.AuthRequired(), handlers.Follow)
    api.DELETE("/community/follow/:id", middleware.AuthRequired(), handlers.Unfollow)
}
r.Run(":8084")
```

### review-service (port 8085)

```go
r := gin.Default()
api := r.Group("/api/v1")
{
    api.GET("/reviews/:productId", handlers.GetReviewsForProduct)
    api.POST("/reviews", middleware.AuthRequired(), handlers.CreateReview)
}
r.Run(":8085")
```

### payment-service (port 8086)

```go
r := gin.Default()
api := r.Group("/api/v1")
{
    api.GET("/exchange-rates", handlers.GetExchangeRates)
    api.PUT("/exchange-rates/:chain", middleware.AdminRequired(), handlers.SetExchangeRate)
    api.DELETE("/exchange-rates/:chain", middleware.AdminRequired(), handlers.DeleteExchangeRate)
    api.GET("/payments", handlers.GetPayment)
    api.POST("/payments", handlers.CreatePayment)
    api.GET("/settings/:key", handlers.GetSetting)
    api.PUT("/settings/:key", middleware.AdminRequired(), handlers.SetSetting)
}
r.Run(":8086")
```

### admin-service (port 8087)

```go
r := gin.Default()
api := r.Group("/api/v1/admin")
api.Use(middleware.AdminRequired())
{
    api.GET("/users", handlers.ListUsers)
    api.GET("/products", handlers.AdminListProducts)
    api.GET("/stats", handlers.GetStats)
    api.PUT("/settings/:key", handlers.SetSetting)
    api.GET("/exchange-rates", handlers.GetExchangeRates)
    api.GET("/community/posts", handlers.AdminListPosts)
    api.DELETE("/community/posts/:id", handlers.AdminDeletePost)
    api.GET("/referrals", handlers.GetReferrals)
    api.POST("/export/emails", handlers.ExportEmails)
}
r.Run(":8087")
```

### media-service (port 8088)

```go
r := gin.Default()
api := r.Group("/api/v1")
{
    api.POST("/media/upload", handlers.UploadMedia)
    api.GET("/media/:id", handlers.GetMedia)
}
r.Run(":8088")
```

---

## Path Resolution Examples

### Example 1: Shop homepage `/`

```
1. Internet → 130.185.123.156:80
2. ingress-nginx: host match → frontend-ingress-staging
3. Path `/` matches `/` Prefix → shop-mfe service (10.43.17.53:80)
4. Service → pod shop-mfe-7f5ddd9894-2mm5l:80
5. Pod nginx: `/` → try_files → /usr/share/nginx/html/index.html
6. Browser loads index.html → Alpine.js initializes shopApp()
7. shopApp.init() → api.products.list({per_page: 20})
8. API call: GET /api/v1/products?per_page=20
9. ingress-nginx: path matches /api/v1/products → product-service:8082
10. product-service returns JSON → Alpine renders product cards
```

### Example 2: Product detail `/product/product-1`

```
1. Internet → 130.185.123.156:80
2. ingress-nginx: host match → frontend-ingress-staging
3. Path `/product/product-1` matches `/product` Prefix → product-mfe service
4. Service → pod product-mfe-7cfdd98cd7-2sqvd:80
5. Pod nginx: `/product/` location matches → alias to /usr/share/nginx/html/product/
6. `product-1` doesn't exist → fallback to /product/product.html
7. Browser loads product.html → Alpine.js reads slug from URL
8. Alpine calls api.products.get('product-1')
9. API: GET /api/v1/products/product-1 → product-service:8082
10. Returns product JSON → Alpine renders detail page
```

### Example 3: Login `/login`

```
1. Internet → 130.185.123.156:80
2. ingress-nginx: host match → frontend-ingress-staging
3. Path `/login` matches `/login` Prefix → auth-mfe service
4. Service → pod auth-mfe-764d7cb9db-x8bsh:80
5. Pod nginx: `/login` → try_files → login.html exists → serves it
6. Browser loads login.html → Alpine.js auth form
7. User submits → POST /api/v1/login
8. ingress-nginx: /api/v1/login → identity-service:8081
9. Returns JWT token → client stores token → redirect to /
```

### Example 4: Community follow (auth required)

```
1. User clicks Follow button on /community page
2. Frontend: api.community.follow(2)
3. API: POST /api/v1/community/follow/2 with Bearer token
4. ingress-nginx: /api/v1/community → community-service:8084
5. community-service: middleware.AuthRequired() validates JWT
6. If valid → follow user → returns 200/201
7. If invalid → returns 401 {"error":"unauthorized"}
```

---

## Known Issues Found by Tests

| # | Issue | Root Cause | Status |
|---|-------|-----------|--------|
| 1 | Shop shows 0 product cards | Alpine.js `defer` + API response timing | Needs investigation |
| 2 | `/assets/favicon.svg` returns 404 | Pod may be running stale image | Needs pod restart |
| 3 | Home nav missing Account link | Each MFE has own nav, no shared shell | Design decision |
| 4 | Admin nav not recognized by tests | Admin uses custom tab nav, not `<a href>` | Test needs update |
| 5 | Community follow returns 401 (correct) | Unauthenticated API call | Frontend should handle |

---

## Quick Reference: All Routes

### Frontend Routes (Browser)

| URL | Pod | File |
|-----|-----|------|
| `/` | shop-mfe | index.html |
| `/category` | shop-mfe | category.html |
| `/login` | auth-mfe | login.html |
| `/register` | auth-mfe | register.html |
| `/auth/*` | auth-mfe | login.html (fallback) |
| `/product/bundle.html` | product-mfe | product/bundle.html |
| `/product/request.html` | product-mfe | product/request.html |
| `/product/{slug}` | product-mfe | product/product.html |
| `/account` | account-mfe | account.html |
| `/cart` | account-mfe | cart.html |
| `/wishlist` | account-mfe | wishlist.html |
| `/referrals` | account-mfe | referrals.html |
| `/checkout` | checkout-mfe | checkout.html |
| `/community` | community-mfe | community.html |
| `/post/{id}` | community-mfe | post.html (via index.html) |
| `/profile/{id}` | community-mfe | profile.html (via index.html) |
| `/admin` | admin-app | index.html |

### API Routes (Programmatic)

| URL | Service | Port |
|-----|---------|------|
| `/api/v1/register` | identity-service | 8081 |
| `/api/v1/login` | identity-service | 8081 |
| `/api/v1/logout` | identity-service | 8081 |
| `/api/v1/me` | identity-service | 8081 |
| `/api/v1/referrals` | identity-service | 8081 |
| `/api/v1/commissions` | identity-service | 8081 |
| `/api/v1/referrals/earnings` | identity-service | 8081 |
| `/api/v1/products` | product-service | 8082 |
| `/api/v1/products/{slug}` | product-service | 8082 |
| `/api/v1/categories` | product-service | 8082 |
| `/api/v1/bundles` | product-service | 8082 |
| `/api/v1/bundles/{id}` | product-service | 8082 |
| `/api/v1/recommendations/{id}` | product-service | 8082 |
| `/api/v1/guest-orders` | commerce-service | 8083 |
| `/api/v1/coupons` | commerce-service | 8083 |
| `/api/v1/coupons/validate` | commerce-service | 8083 |
| `/api/v1/orders` | commerce-service | 8083 |
| `/api/v1/orders/{id}` | commerce-service | 8083 |
| `/api/v1/cart` | commerce-service | 8083 |
| `/api/v1/cart/items` | commerce-service | 8083 |
| `/api/v1/cart/items/{id}` | commerce-service | 8083 |
| `/api/v1/wishlist` | commerce-service | 8083 |
| `/api/v1/wishlist/toggle` | commerce-service | 8083 |
| `/api/v1/recently-viewed` | commerce-service | 8083 |
| `/api/v1/compare` | commerce-service | 8083 |
| `/api/v1/compare/toggle` | commerce-service | 8083 |
| `/api/v1/posts` | community-service | 8084 |
| `/api/v1/community/posts` | community-service | 8084 |
| `/api/v1/community/posts/{id}` | community-service | 8084 |
| `/api/v1/community/posts/{id}/like` | community-service | 8084 |
| `/api/v1/community/posts/{id}/comments` | community-service | 8084 |
| `/api/v1/community/users/{id}` | community-service | 8084 |
| `/api/v1/community/follow/{id}` | community-service | 8084 |
| `/api/v1/users` | community-service | 8084 |
| `/api/v1/profile` | community-service | 8084 |
| `/api/v1/profile/{id}` | community-service | 8084 |
| `/api/v1/reviews/{productId}` | review-service | 8085 |
| `/api/v1/exchange-rates` | payment-service | 8086 |
| `/api/v1/exchange-rates/{chain}` | payment-service | 8086 |
| `/api/v1/payments` | payment-service | 8086 |
| `/api/v1/settings/{key}` | payment-service | 8086 |
| `/api/v1/admin/*` | admin-service | 8087 |
| `/api/v1/media/*` | media-service | 8088 |
