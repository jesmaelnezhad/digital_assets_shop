# Pawradise Microservices Architecture — Deployment Reference

> **Date**: 2026-09-17 (domain-based routing v2)
> **Purpose**: Reference for the two-server (BLUE/RED) microservice deployment with host-based ingress routing to staging and production namespaces.

---

## 1. Server Roles

| Server | IP | Role | Key Software |
|--------|-----|------|--------------|
| **BLUE** | 130.185.121.83 | Build & compile box | Go 1.22.2, Docker, git |
| **RED** | 130.185.123.156 | Serving cluster | k3s v1.36.4+k3s1, PostgreSQL 16, MongoDB 7, Docker registry, host nginx (SSL) |

---

## 2. RED Cluster Overview

### 2.1 Namespaces

| Namespace | Purpose |
|-----------|---------|
| `ingress-nginx` | Single ingress-nginx controller (the API gateway) |
| `production` | Production backend pods + services |
| `staging` | Staging backend pods + services |
| `database` | PostgreSQL instance with 16 logical databases, plus one MongoDB instance shared by staging and production |
| `registry` | Docker registry for pushing images from BLUE |

### 2.2 Key Components on RED

#### Ingress Controller (API Gateway)

- **Deployment**: `ingress-nginx-controller` in `ingress-nginx` namespace
- **Image**: `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise-ingress-nginx:v1.11.2`
- **Service**: `ingress-nginx-controller` (NodePort `:30758`)
- **IngressClass**: `nginx` (controller: `k8s.io/ingress-nginx`)
- **RBAC**: ClusterRole with permissions for ingresses, services, endpoints, endpointslices, leases, events, configmaps, secrets, nodes, pods

#### Backend Pods (Separate Pod Architecture)

- **Deployments**: One Deployment per service per namespace (16 total)
- **Image**: `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise/<service>:<tag>`
- **Label selector**: `app: <service-name>` (per-service labels)
- **Containers per pod**: 1 (each service runs in its own pod)

| Container Name | Port | Service Name |
|----------------|------|--------------|
| identity-service | 8081 | identity-service |
| product-service | 8082 | product-service |
| commerce-service | 8083 | commerce-service |
| community-service | 8084 | community-service |
| review-service | 8085 | review-service |
| payment-service | 8086 | payment-service |
| admin-service | 8087 | admin-service |
| media-service | 8088 | media-service |
| events-service | 8089 | events-service |

#### Per-Service Kubernetes Services

Each service selects pods by label `app: <service-name>` and routes to the correct container port:

| Service | Port | Target Port |
|---------|------|-------------|
| identity-service | 80 | 8081 |
| product-service | 80 | 8082 |
| commerce-service | 80 | 8083 |
| community-service | 80 | 8084 |
| review-service | 80 | 8085 |
| payment-service | 80 | 8086 |
| admin-service | 80 | 8087 |
| media-service | 80 | 8088 |
| events-service | 8089 | 8089 |

> **Note**: The legacy `backend-pod` service (with all 8 ports) still exists but is not used by ingress.

#### Database

- **StatefulSet**: `postgres-0` in `database` namespace
- **Image**: `docker.io/library/postgres:16-alpine`
- **Service**: `postgres.database.svc.cluster.local:5432`
- **User**: `app`
- **16 logical databases**: `appdb_{service}_{environment}` for each of 8 Postgres-backed services × 2 environments

#### MongoDB

Live staging/production currently share **one host Docker MongoDB** on RED (`:27017`), the same operational pattern as Postgres (`:5432` on the host). `k8s/database/mongodb.yaml` is the cluster form (StatefulSet in `database`, NodePort `30017`) for when databases move fully into k3s.

- **Host container**: `mongodb` / image `mongo:7` / volume `mongodata`
- **Auth**: root user `paw` / password in `mongodb-secret` (and host env)
- **One instance** for staging and production for now
- **Database**: `events` (collections `events`, `meta`). TTL index on `events.expire_at`
- **events-service URI**: `mongodb://paw:<password>@130.185.123.156:27017/?authSource=admin`

#### Registry

- **Namespace**: `registry` (k3s). Manifest: `k8s/registry.yaml`
- **Images (pulled from Docker Hub once)**: `docker.io/library/registry:2` + `docker.io/library/nginx:1.27-alpine` auth proxy
- **Service**: ClusterIP `registry.registry.svc.cluster.local:80` — no NodePort, no host Docker registry
- **HTTPS**: staging host, same host-nginx → ingress-nginx path routing as the apps
  - `/registry` rewrites to the registry root (k3s pull mirror)
  - `/v2` is the Docker Registry HTTP API (`docker login` / push / pull)
- **Auth**: nginx sidecar. GET/HEAD accept the **pull** or **push** user. PUT/POST/PATCH/DELETE accept **push** only. Passwords live in Secret `registry-auth` (not in git) and `/root/.registry-auth` on RED/BLUE
- **k3s**: `/etc/rancher/k3s/registries.yaml` pull user + TLS to the staging host. App images are `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise/<name>:<tag>`

---

## 3. Ingress Routing Architecture

### 3.1 Domain-Based Routing (v2, 2026-09-10+)

Environments are separated by **hostname**, not by path prefix. Both environments use identical `/api/v1` paths — no rewriting needed.

```
                    ┌─────────────────────────────────────┐
                    │  ingress-nginx-controller (gateway)  │
                    │  130.185.123.156:30758                 │
                    └──────────────┬──────────────────────┘
                                   │
          Host: staging domain     │    Host: production domain
          (server-ad5ae8ea...)     │    (pawradise.ir)
                 │                 │                 │
                 ▼                 ▼                 ▼
        ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
        │ staging/    │   │ (via host   │   │ production/ │
        │ identity-   │   │  nginx)     │   │ identity-   │
        │ service:80  │   └─────────────┘   │ service:80  │
        └─────────────┘                     └─────────────┘
```

### 3.2 Host Nginx (SSL Terminator on RED)

RED's host nginx (`/etc/nginx/sites-available/pawradise-ssl`) handles:

1. **Port 80**: Redirects HTTP → HTTPS
2. **Port 443**: SSL termination (Let's Encrypt cert for staging domain)
   - Proxies `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir` → k3s NodePort 30758 (staging), including `/registry` and `/v2` for the in-cluster registry. Staging `client_max_body_size` is unlimited so image blob uploads are not 413'd.
   - Proxies `pawradise.ir` → k3s NodePort 30758 (production)
   - Serves `/assets/` directly from `/var/www/production/assets/`
   - Serves `/.well-known/acme-challenge/` for Let's Encrypt renewal

**No port 8080/8081 hosting** — the old config serving static files on port 8080/8081 has been removed.

### 3.3 Ingress Resources

#### Production Ingress (`api-gateway`)

- **Name**: `api-gateway`
- **Namespace**: `production`
- **IngressClass**: `nginx`
- **Host**: `pawradise.ir`
- **Rewrite**: None (paths forwarded as-is)
- **Routes**: `/api/v1/*` → production services

#### Staging Ingress (`api-gateway-staging`)

- **Name**: `api-gateway-staging`
- **Namespace**: `staging`
- **IngressClass**: `nginx`
- **Host**: `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir`
- **Rewrite**: None (paths forwarded as-is) — NO `rewrite-target` annotation
- **Routes**: `/api/v1/*` → staging services

### 3.4 Routing Matrix

| Ingress Path | Target Service | Namespace | Notes |
|--------------|----------------|-----------|-------|
| `/api/v1/register` | identity-service:80 | production | Auth |
| `/api/v1/login` | identity-service:80 | production | Auth |
| `/api/v1/logout` | identity-service:80 | production | Auth |
| `/api/v1/me` | identity-service:80 | production | Auth |
| `/api/v1/profile` | identity-service:80 | production | Auth (same as above) |
| `/api/v1/referrals` | identity-service:80 | production | Auth |
| `/api/v1/commissions` | identity-service:80 | production | Auth |
| `/api/v1/products` | product-service:80 | production | Public |
| `/api/v1/categories` | product-service:80 | production | Public |
| `/api/v1/bundles` | product-service:80 | production | Public |
| `/api/v1/recommendations` | product-service:80 | production | Public |
| `/api/v1/guest-orders` | commerce-service:80 | production | Public |
| `/api/v1/coupons/validate` | commerce-service:80 | production | Public |
| `/api/v1/orders` | commerce-service:80 | production | Auth |
| `/api/v1/cart` | commerce-service:80 | production | Auth |
| `/api/v1/wishlist` | commerce-service:80 | production | Auth |
| `/api/v1/recently-viewed` | commerce-service:80 | production | Auth |
| `/api/v1/compare` | commerce-service:80 | production | Auth |
| `/api/v1/community/posts` | community-service:80 | production | Public + Auth wrapper |
| `/api/v1/community/users/:id` | community-service:80 | production | Public |
| `/api/v1/reviews` | review-service:80 | production | Auth |
| `/api/v1/exchange-rates` | payment-service:80 | production | Public |
| `/api/v1/payments/` | payment-service:80 | production | Auth |
| `/api/v1/settings` | payment-service:80 | production | Auth |
| `/api/v1/admin/` | admin-service:80 | production | Admin token |
| `/api/v1/media/` | media-service:80 | production | Auth |

> **Wrapper routes on community-service**: `/api/v1/posts`, `/api/v1/posts/:id`, `/api/v1/profile`, `/api/v1/profile/:id`, `/api/v1/users/:id`, `/api/v1/follow/:userId` — these are convenience wrappers that forward to `/api/v1/community/...` endpoints for e2e test compatibility.

> **Staging**: Same routing matrix, but all routes land in the `staging` namespace. The ingress uses `host` field to distinguish: `server-ad5ae8ea-...` → staging, `pawradise.ir` → production.

### 3.5 CORS Configuration

CORS is configured per-namespace via the `ENV_NAME` environment variable on each service pod:

```go
// middleware/cors.go
func CORSMiddleware() gin.HandlerFunc {
    env := os.Getenv("ENV_NAME") // "staging" or "production"
    allowedOrigins := os.Getenv("CORS_ORIGINS") // Comma-separated origins for this env
    // ...
}
```

| Namespace | ENV_NAME | CORS_ORIGINS |
|-----------|----------|--------------|
| `staging` | `staging` | `https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir` |
| `production` | `production` | `https://pawradise.ir` |

---

## 4. Build & Deploy Workflow (BLUE → RED)

### 4.1 Build Go Binary on BLUE (statically linked)

```bash
cd /root/project/services/<service-name>
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /tmp/<service-name>-static .
```

**Important**: Always use `CGO_ENABLED=0` to produce a statically linked binary. Alpine Linux uses musl libc, and dynamically linked glibc binaries (`CGO_ENABLED=1`) will fail with `exec /server: no such file or directory` on Alpine containers.

### 4.2 Build Docker Image on BLUE

```bash
# Multi-stage Dockerfile (recommended — copies binary from builder stage)
cat > Dockerfile <<'EOF'
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download 2>/dev/null || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY --from=builder /server /server
EXPOSE <port>
CMD ["/server"]
EOF

docker build -t server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise/<service>:v<N> .
```

### 4.3 Push to the in-cluster registry

```bash
# once per machine: docker login with the push user from /root/.registry-auth
docker login server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir
docker push server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise/<service>:v<N>
```

### 4.4 Deploy on RED

```bash
ssh root@130.185.123.156 "k3s kubectl set image deployment/<service> \
  <service>=server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise/<service>:v<N> -n staging"

ssh root@130.185.123.156 "k3s kubectl set image deployment/<service> \
  <service>=server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise/<service>:v<N> -n production"
```

### 4.5 Copy Binary Directly (alternative — bypasses Docker)

When Docker caching issues prevent clean builds:

```bash
# Build on BLUE
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /tmp/<service>-static .

# Copy to RED
scp /tmp/<service>-static root@130.185.123.156:/tmp/

# Deploy to running pod
ssh root@130.185.123.156 "
  POD=\$(k3s kubectl get pods -n staging -l app=<service> -o jsonpath='{.items[0].metadata.name}')
  k3s kubectl cp /tmp/<service>-static staging/\$POD:/server
  k3s kubectl delete pod \$POD -n staging --grace-period=0
"
```

---

## 5. Database Migration Workflow

### 5.1 Create Migration File

Write SQL to `/root/project/services/<service>/migrations/NNN_descriptive_name.sql`

### 5.2 Apply to Both Environments

```bash
scp <migration>.sql root@130.185.123.156:/tmp/migration.sql
ssh root@130.185.123.156 "k3s kubectl cp /tmp/migration.sql database/postgres-0:/tmp/migration.sql"
ssh root@130.185.123.156 "k3s kubectl exec -n database postgres-0 -- psql -U app -d appdb_<service>_staging -f /tmp/migration.sql"
ssh root@130.185.123.156 "k3s kubectl exec -n database postgres-0 -- psql -U app -d appdb_<service>_production -f /tmp/migration.sql"
```

---

## 6. Registry Configuration

Host: `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir` (HTTPS, same TLS as staging). Path `/registry` and Docker API `/v2`.

### 6.1 RED's `/etc/rancher/k3s/registries.yaml`

Pull user only (cannot push). Created on the node, not committed.

```yaml
mirrors:
  "server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir":
    endpoint:
      - "https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/registry"
configs:
  "server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir":
    auth:
      username: k3s-pull
      password: "(from /root/.registry-auth)"
```

Restart k3s after edits: `systemctl restart k3s`.

### 6.2 Docker login on RED and BLUE (push)

```bash
# credentials: /root/.registry-auth  (PUSH_USER / PUSH_PASS)
docker login server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir
```

Do not list this host as an insecure HTTP registry. Traffic is TLS via host nginx.

---

## 7. Image Inventory

### 7.1 Ingress Controller

| Image | Source | Location |
|-------|--------|----------|
| `registry.k8s.io/ingress-nginx/controller:v1.11.2` | Tar file provided by user | `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise-ingress-nginx:v1.11.2` |

### 7.2 Backend Services

| Service | Image | Port |
|---------|-------|------|
| identity-service | `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise/identity-service:<tag>` | 8081 |
| product-service | `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise/product-service:<tag>` | 8082 |
| commerce-service | `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise/commerce-service:<tag>` | 8083 |
| community-service | `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise/community-service:<tag>` | 8084 |
| review-service | `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise/review-service:<tag>` | 8085 |
| payment-service | `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise/payment-service:<tag>` | 8086 |
| admin-service | `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise/admin-service:<tag>` | 8087 |
| media-service | `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise/media-service:<tag>` | 8088 |
| events-service | `server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/pawradise/events-service:<tag>` | 8089 |

### 7.3 Infrastructure

| Component | Image |
|-----------|-------|
| PostgreSQL (host Docker) | `docker.io/library/postgres:16-alpine` |
| MongoDB (host Docker) | `docker.io/library/mongo:7` |
| Registry + auth proxy (k3s `registry` ns) | `docker.io/library/registry:2`, `docker.io/library/nginx:1.27-alpine` (Docker Hub once) |

---

## 8. Environment Variables (Backend Containers)

| Variable | Value |
|----------|-------|
| `PORT` | Container-specific (8081-8089) |
| `DB_HOST` | `postgres.database.svc.cluster.local` |
| `DB_PORT` | `5432` |
| `DB_USER` | `app` |
| `DB_PASSWORD` | (from secret `postgres-secret`) |
| `DB_NAME` | `appdb_<service>_<environment>` |
| `MONGO_URI` | `mongodb://paw:<password>@mongodb.database.svc.cluster.local:27017/?authSource=admin` (events-service) |
| `MONGO_DB` | `events` |
| `JWT_SECRET` | (from secret `pawradise-secrets`) |
| `ADMIN_TOKEN` | (from secret `pawradise-secrets`) |
| `ENV_NAME` | `staging` or `production` (controls CORS, env detection) |
| `CORS_ORIGINS` | Comma-separated origins for this environment |

---

## 9. Verification Commands

### 9.1 Check All Pods

```bash
ssh root@130.185.123.156 "k3s kubectl get pods -A -o wide"
```

### 9.2 Check Ingress Routing

```bash
ssh root@130.185.123.156 "k3s kubectl describe ingress -A"
```

### 9.3 Test Routing from BLUE (domain-based)

**Always test using the real domain name, not the IP. The ingress routes by Host header.**

```bash
# Production (placeholder domain — no DNS yet, no testing possible)
curl -sk https://pawradise.ir/api/v1/health

# Staging (real domain, SSL cert installed)
curl -sk https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/api/v1/health

# With auth token
curl -sk https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/api/v1/admin/users \
  -H "Authorization: Bearer admin_secret_staging_2026"

# POST register
curl -sk -X POST https://server-ad5ae8ea-5132-4cd3-b11f-5cb0f43bdc53.eu-west1-a.arvancompute.ir/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.ir","password":"pass123","first_name":"Test","last_name":"User"}'
```

**Important**: Direct IP access (e.g., `curl http://130.185.123.156:30758/api/v1/...`) will match the first ingress (production) regardless of which namespace you want. Always use domain names.

### 9.4 Check Ingress Logs

```bash
ssh root@130.185.123.156 "k3s kubectl logs -n ingress-nginx -l app=ingress-nginx-controller --tail=50"
```

### 9.5 Check Service Logs

```bash
ssh root@130.185.123.156 "k3s kubectl logs -n staging -l app=<service> --tail=50"
ssh root@130.185.123.156 "k3s kubectl logs -n production -l app=<service> --tail=50"
```

---

## 10. Key Design Decisions

1. **Separate Pod per Service**: Each microservice runs in its own pod, targeting ~40-50MB memory per pod. RED has ~1GB available, so 16 pods (8 services × 2 namespaces) fit comfortably.

2. **Domain-Based Routing (v2)**: Environments are separated by hostname (`server-ad5ae8ea-...` → staging, `pawradise.ir` → production). This allows identical URL paths in both environments — no `/staging/` prefix needed. Frontend links work the same in both environments. Host nginx routes by Host header to k3s ingress, which routes by host match to the correct namespace.

3. **Single Ingress Gateway**: One ingress-nginx controller in `ingress-nginx` namespace watches all namespaces for Ingress resources.

4. **Host Nginx as SSL Terminator**: RED's host nginx handles Let's Encrypt on port 443 and proxies to k3s NodePort 30758 with Host header preserved. This avoids certificate management inside k3s.

5. **Wrapper Routes for e2e Compatibility**: community-service registers both `/api/v1/community/...` (actual routes) and `/api/v1/posts`, `/api/v1/profile`, etc. (wrapper routes) to maintain compatibility with e2e tests that expect those paths.

6. **Registry on RED**: in-cluster registry in namespace `registry`, HTTPS on the staging host at `/registry` and `/v2`, nginx basic auth. App images are not pulled from Docker Hub after the registry and nginx images themselves are fetched once.

7. **CGO_ENABLED=0 for Alpine**: All Go binaries must be statically linked (CGO_ENABLED=0) to run on Alpine-based containers. Dynamic linking produces glibc binaries that fail with `exec /server: no such file or directory`.

8. **No path rewriting needed**: Domain-based routing eliminates the need for `rewrite-target` annotations. Paths pass through unchanged to services.

---

## 11. Troubleshooting

### 11.1 Ingress Controller Won't Start

- Check RBAC: `k3s kubectl get clusterrole ingress-nginx -o yaml`
- Check logs: `k3s kubectl logs -n ingress-nginx -l app=ingress-nginx-controller`
- Common issue: Missing `leases` permission in `coordination.k8s.io` API group

### 11.2 Images Won't Pull

- Verify registry is running: `k3s kubectl get pods -n registry`
- Check registries.yaml: `cat /etc/rancher/k3s/registries.yaml`
- Restart k3s: `systemctl restart k3s`

### 11.3 Backend Returns 503

- Check pod status: `k3s kubectl get pods -n <namespace>`
- Check service endpoints: `k3s kubectl get endpoints -n <namespace> <service>`
- Check pod logs: `k3s kubectl logs -n <namespace> <pod> -c <container>`

### 11.4 404 Errors on API Endpoints

- Check which ingress matches: `k3s kubectl describe ingress <ingress-name> -n <namespace>`
- Verify the Host header matches the ingress `host` field
- Check that the service name in the ingress matches the Kubernetes service
- Verify the service has endpoints (pods are running and healthy)
- For community-service: check logs for "duplicate route" panic — the service may have crashed on startup

### 11.5 Binary Fails with "no such file or directory"

- The binary is dynamically linked (glibc) but running on Alpine (musl)
- Always build with `CGO_ENABLED=0 GOOS=linux GOARCH=amd64`
- Verify: `file <binary>` should show "statically linked"

### 11.6 Admin Token Rejected

- Verify the token in the pod env: `k3s kubectl get deploy admin-service -n <ns> -o jsonpath="{.spec.template.spec.containers[0].env[?(@.name==\"ADMIN_TOKEN\")].value}"`
- Test directly: `curl -H "Authorization: Bearer <token>" <domain>/api/v1/admin/users`

### 11.7 Staging Returns 301 (Redirect)

- This happens when the path doesn't match exactly (trailing slash mismatch)
- Ensure ingress has both `/path` and `/path/` variants if needed
- With domain-based routing, this should not happen since there's no path rewriting

---

## 12. File Locations

On BLUE the checkout of this repo may live at `/root/project/` or another path. In the repo itself:

| Path | Purpose |
|------|---------|
| `services/<service>/` | Go source |
| `k8s/` | Manifests including `api-gateway-staging.yaml`, `frontend-ingress-staging.yaml` |
| `tests/e2e-suite.js` | Node HTTP suite |
| `tests/unit/` | Go unit tests |
| `tests/integration/` | Go integration tests |
| `tests/frontend/` | Playwright |

### 12.2 On RED

| Path | Purpose |
|------|---------|
| `/etc/rancher/k3s/registries.yaml` | Registry mirror configuration |
| `/etc/nginx/sites-available/pawradise-ssl` | Host nginx SSL config (domain routing + SSL) |
| `/etc/nginx/sites-enabled/pawradise-ssl` | Symlink to above |
| `/root/k8s/` | Synced copy of K8s manifests (if used) |
| `/tmp/<service>-static` | Temporary binary copy location |
| `/var/www/production/assets/` | Host-nginx static assets (may be stale vs MFE pods) |

---

## 13. Testing from BLUE

Run tests against the **staging hostname**, not the NodePort IP.

```bash
cd tests && node e2e-suite.js
cd tests/unit && go test ./... -v
cd tests/integration && go test ./... -v
cd tests/frontend && npx playwright test
```

---

## 14. Common Pitfalls

1. **Testing via IP instead of domain**: `curl http://130.185.123.156:30758/api/v1/...` matches the first ingress (production). Always use the domain name: `curl -H "Host: server-ad5ae8ea-..." https://...` or test through the host nginx on port 443.

2. **Using `rewrite-target` with domain-based routing**: This strips path prefixes and breaks routing. Domain-based ingresses should NOT use `rewrite-target` annotations.

3. **Building without CGO_ENABLED=0**: Produces dynamically linked binaries that fail on Alpine. Always use `CGO_ENABLED=0 GOOS=linux GOARCH=amd64`.

4. **Kubernetes caching old image layers**: When rebuilding Docker images, always use a new tag (v1, v2, v3...) to bust the cache. Building with the same tag may reuse cached layers.

5. **Docker COPY using old binary**: The `COPY community-service-binary /server` in the Dockerfile copies whatever is in the build context at build time. If you update the source but forget to rebuild the binary before `docker build`, the old binary gets included. Always `go build` before `docker build`, or use a multi-stage Dockerfile that builds inside Docker.

6. **Community-service route registration order**: Registering `GET /api/v1/profile` in both the public group and auth group causes a panic. Register it only once — in the auth group, and add a public wrapper if needed.
