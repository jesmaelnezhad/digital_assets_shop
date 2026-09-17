# Ingress-Nginx on k3s with Private Registry

Deploying the standard Kubernetes ingress-nginx controller on a self-hosted k3s cluster with a private registry (when registry.k8s.io is blocked).

## Problem

`registry.k8s.io` is blocked from this server (403). The standard ingress-nginx controller image cannot be pulled directly. The `docker.io/nginx/nginx-ingress` image is NOT the standard controller — it's the NGINX Inc. commercial variant that requires VirtualServer CRDs.

## Solution: Use pre-built tar file (PREFERRED)

A pre-built `registry.k8s.io/ingress-nginx/controller:v1.11.2` tar file is available at `/root/ingress-nginx-controller.tar` on BLUE.

### 1. Load image on BLUE

```bash
docker load -i /root/ingress-nginx-controller.tar
# Loaded image: registry.k8s.io/ingress-nginx/controller:v1.11.2
```

### 2. Push to RED registry

```bash
docker tag registry.k8s.io/ingress-nginx/controller:v1.11.2 194.5.206.106:30099/pawradise-ingress-nginx:v1.11.2
docker push 194.5.206.106:30099/pawradise-ingress-nginx:v1.11.2
```

### 3. Deploy on RED

Use the standard manifests from https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.11.2/deploy/static/provider/baremetal/deploy.yaml, changing only the image reference to `194.5.206.106:30099/pawradise-ingress-nginx:v1.11.2`.

## Alternative Solution: Build from source on BLUE, push to RED registry (fallback when tar unavailable)

### 1. Clone and build on BLUE

```bash
cd /tmp
git clone --depth 1 --branch controller-v1.8.1 https://github.com/kubernetes/ingress-nginx.git
cd ingress-nginx
export PKG=k8s.io/ingress-nginx ARCH=amd64 COMMIT_SHA=$(git rev-parse --short HEAD) REPO_INFO=k8s.io/ingress-nginx TAG=v1.8.1 CGO_ENABLED=0 GOARCH=amd64
bash build/build.sh  # produces rootfs/bin/amd64/nginx-ingress-controller
```

### 2. Build Docker image

Use this Dockerfile (openresty base for built-in lua, patched template):

```dockerfile
FROM openresty/openresty:1.21.4.1-alpine

ARG TARGETARCH=amd64 VERSION=v1.8.1 COMMIT_SHA=ab99e23

RUN apk update && apk add --no-cache dumb-init diffutils sed && rm -rf /var/cache/apk/*

WORKDIR /etc/nginx

COPY rootfs/etc /etc
COPY rootfs/bin/${TARGETARCH}/dbg /
COPY rootfs/bin/${TARGETARCH}/nginx-ingress-controller /
COPY rootfs/bin/${TARGETARCH}/wait-shutdown /

RUN mkdir -p /etc/ingress-controller /etc/ingress-controller/ssl /etc/ingress-controller/auth /var/log/nginx /tmp/nginx /etc/nginx/geoip && \
    ln -sf /usr/local/openresty/nginx/sbin/nginx /usr/bin/nginx && \
    ln -sf /usr/local/openresty/nginx/sbin/nginx /usr/sbin/nginx

# Patch template: remove load_module directives for statically-linked modules
RUN sed -i '/load_module.*ngx_http_lua_module/d' /etc/nginx/template/nginx.tmpl && \
    sed -i '/load_module.*ngx_http_geoip2_module/d' /etc/nginx/template/nginx.tmpl && \
    sed -i '/load_module.*ngx_http_brotli/d' /etc/nginx/template/nginx.tmpl && \
    sed -i '/load_module.*ngx_http_auth_digest_module/d' /etc/nginx/template/nginx.tmpl && \
    sed -i '/load_module.*ngx_http_modsecurity_module/d' /etc/nginx/template/nginx.tmpl && \
    sed -i '/load_module.*ngx_http_opentracing_module/d' /etc/nginx/template/nginx.tmpl && \
    sed -i '/load_module.*otel_ngx_module/d' /etc/nginx/template/nginx.tmpl && \
    sed -i '/ajp_temp_path/d' /etc/nginx/template/nginx.tmpl

USER root
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/nginx-ingress-controller", "--enable-priority-class=false"]
```

Build and push:

```bash
docker build -t 194.5.206.106:30099/pawradise-ingress-nginx:v1.8.x -f Dockerfile .
docker push 194.5.206.106:30099/pawradise-ingress-nginx:v1.8.x
```

### 3. Deploy on RED

```yaml
# ConfigMap (critical: disable optional modules)
apiVersion: v1
kind: ConfigMap
metadata:
  name: ingress-nginx-controller
  namespace: ingress-nginx
data:
  use-geoip: "false"
  use-geoip2: "false"
  use-brotli: "false"
  enable-modsecurity: "false"
  enable-opentracing: "false"
  ssl-redirect: "false"
---
# Deployment (note POD_NAME/POD_NAMESPACE env vars required)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ingress-nginx-controller
  namespace: ingress-nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ingress-nginx-controller
  template:
    metadata:
      labels:
        app: ingress-nginx-controller
    spec:
      serviceAccountName: ingress-nginx
      containers:
      - name: controller
        image: 194.5.206.106:30099/pawradise-ingress-nginx:v1.8.x
        ports:
        - containerPort: 80
        - containerPort: 443
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        args:
        - /nginx-ingress-controller
        - --election-id=ingress-nginx-leader
        - --controller-class=k8s.io/ingress-nginx
        - --ingress-class=nginx
        - --configmap=$(POD_NAMESPACE)/ingress-nginx-controller
---
# IngressClass
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: nginx
spec:
  controller: k8s.io/ingress-nginx  # NOT nginx.org/ingress-controller
---
# RBAC
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ingress-nginx
  namespace: ingress-nginx
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ingress-nginx
rules:
- apiGroups: [""]
  resources: ["configmaps","endpoints","nodes","pods","secrets","services"]
  verbs: ["get","list","watch"]
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses","ingressclasses"]
  verbs: ["get","list","watch"]
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses/status"]
  verbs: ["update"]
- apiGroups: [""]
  resources: ["events"]
  verbs: ["create","patch"]
- apiGroups: ["discovery.k8s.io"]
  resources: ["endpointslices"]
  verbs: ["get","list","watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ingress-nginx
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ingress-nginx
subjects:
- kind: ServiceAccount
  name: ingress-nginx
  namespace: ingress-nginx
```

### 4. Routing pattern for staging/production

Use two Ingress resources:
- Production: paths like `/api/v1/products` → `product-service:80` in `production` namespace
- Staging: paths like `/staging/api/v1/products` → `product-service:80` in `staging` namespace, with annotation `nginx.ingress.kubernetes.io/rewrite-target: /$2` and path regex `/staging/api/v1/products(/|$)(.*)`

This strips `/staging` prefix so staging services serve identical routes as production.

### 5. k3s registry configuration

`/etc/rancher/k3s/registries.yaml` must include mirrors for blocked registries:

```yaml
mirrors:
  docker.io:
    endpoint:
      - https://registry-1.docker.io
  194.5.206.106:30099:
    endpoint:
      - http://194.5.206.106:30099
configs:
  194.5.206.106:30099:
    tls:
      insecure_skip_verify: true
```

Then `systemctl restart k3s`.

## Common pitfalls

| Error | Cause | Fix |
|-------|-------|-----|
| `unknown directive "ajp_temp_path"` | Template not patched | `sed -i '/ajp_temp_path/d' template/nginx.tmpl` |
| `unknown directive "lua_package_path"` | nginx base lacks lua module | Use openresty base OR patch template |
| `missing POD_NAME or POD_NAMESPACE` | Env vars not set | Add `env` with `fieldRef` to deployment |
| `IngressClass invalid Spec.Controller` | Controller field mismatch | Use `k8s.io/ingress-nginx` not `nginx.org/...` |
| `port 80 already in use` | Host nginx competing | Stop host nginx OR use different nodePort |
| ImagePullBackOff for ingress image | Registry not configured | Check `/etc/rancher/k3s/registries.yaml` |
| `server_names_hash_bucket_size` duplicate | http-snippet also sets it | Remove from http-snippet OR patch template |
| `mime.types` not found | openresty puts it at different path | `ln -sf /usr/local/openresty/nginx/conf/mime.types /etc/nginx/mime.types` |
| `resty/http.so` not found | lua_package_path wrong | Fix paths to include `/usr/local/openresty/lualib/` |
| `leases.coordination.k8s.io forbidden` | Missing RBAC permission | Add `coordination.k8s.io` leases to ClusterRole |
| `unknown directive "geoip_country"` | Template not fully patched | `sed -i '/geoip_country/d' template/nginx.tmpl` |
| `unknown directive "fastcgi_temp_path"` | Template not fully patched | `sed -i '/fastcgi_temp_path/d' template/nginx.tmpl` |
| `unknown directive "proxy_temp_path"` | Template not fully patched | `sed -i '/proxy_temp_path/d' template/nginx.tmpl` |
| `fork/exec /usr/bin/nginx: no such file` | Symlink missing | `ln -sf /usr/local/openresty/nginx/sbin/nginx /usr/bin/nginx` |

### Critical template patches for openresty base

When using `openresty/openresty:1.21.4.1-alpine` as base, the template needs these additional patches beyond removing `load_module` lines:

```bash
# Remove all *_temp_path directives (not in openresty)
sed -i '/ajp_temp_path/d; /fastcgi_temp_path/d; /proxy_temp_path/d; /uwsgi_temp_path/d; /scgi_temp_path/d; /client_body_temp_path/d' /etc/nginx/template/nginx.tmpl

# Fix lua_package_path to point to openresty's lib directories
sed -i 's|lua_package_path "/etc/nginx/lua/?.lua;;";|lua_package_path "/etc/nginx/lua/?.lua;/usr/local/openresty/site/lualib/?.lua;/usr/local/openresty/lualib/?.lua;/usr/local/openresty/luajit/lib/lua/5.1/?.lua;;";|' /etc/nginx/template/nginx.tmpl
sed -i 's|lua_package_cpath "/etc/nginx/lua/?.so;;";|lua_package_cpath "/etc/nginx/lua/?.so;/usr/local/openresty/site/lualib/?.so;/usr/local/openresty/lualib/?.so;/usr/local/openresty/luajit/lib/lua/5.1/?.so;;";|' /etc/nginx/template/nginx.tmpl
```

### Required ConfigMap settings

The ConfigMap MUST disable all optional modules to avoid template errors:

```yaml
data:
  use-geoip: "false"
  use-geoip2: "false"
  use-brotli: "false"
  enable-modsecurity: "false"
  enable-opentracing: "false"
  ssl-redirect: "false"
```

### Required RBAC additions

The ClusterRole needs `coordination.k8s.io` leases permission for leader election:

```yaml
- apiGroups: ["coordination.k8s.io"]
  resources: ["leases"]
  verbs: ["get", "create", "update", "patch"]
```
