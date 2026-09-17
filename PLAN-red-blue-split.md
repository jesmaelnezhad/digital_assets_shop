# Pawradise RED/BLUE Server Split — Execution Plan

**Status:** COMPLETED — 2026-09-04  
**Author:** Hermes Agent  
**Date:** 2026-09-04  
**Goal:** Move serving infrastructure to RED (130.185.123.156), keep BLUE as build/dev machine

---

## Architecture

| Server | Role | IP | What runs |
|--------|------|-----|-----------|
| RED | Serving cluster | 130.185.123.156 | k3s, PostgreSQL, host nginx (port 80), backend pods, frontend pods |
| BLUE | Build / dev / admin | (current server) | Go builds, Docker builds, image export→RED via SSH, Docker Compose PG, tests, agent session |

**Traffic flow:**
- `pawradise.ir` → RED's IP (domain points to RED)
- RED host nginx on port 80 routes: `/` → prod frontend, `/staging/` → staging frontend, `/api/v1/` → prod backend, `/staging/api/v1/` → staging backend, `/admin` → prod admin UI, `/staging/admin` → staging admin UI, `/api/v1/admin/*` → prod admin API, `/staging/api/v1/admin/*` → staging admin API
- Tests run on BLUE, targeting RED's services (via NodePort or via RED's nginx)
- SSL later

---

## Step 1: Provision and Harden RED

**Purpose:** Fresh Ubuntu 24.04 server prepared for k3s with system hardening applied.

**Substeps:**

1.1 Connect to RED via SSH as root, verify OS version, CPU, RAM  
1.2 Create `/root/release-notes/01-red-server-provisioning.md` documenting the server spec  
1.3 System hardening on RED (mirror BLUE's hardening):
   - `apt-get update && apt-get install -y ca-certificates curl apt-transport-https gpg`
   - Disable swap: `swapoff -a && sed -i '/swap/d' /etc/fstab`
   - Load kernel modules: `modprobe overlay && modprobe br_netfilter`
   - Write `/etc/sysctl.d/99-k8s.conf` with: vm.swappiness=10, net.bridge.bridge-nf-call-iptables=1, net.bridge.bridge-nf-call-ip6tables=1, net.ipv4.ip_forward=1, net.ipv6.conf.all.forwarding=1
   - Apply: `sysctl --system`
1.4 Install containerd: add Docker apt repo key, add repo, `apt-get install -y containerd.io`, `containerd config default | tee /etc/containerd/config.toml && systemctl restart containerd`
1.5 Install k3s: `curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--disable traefik --disable local-path-provisioner" sh -`
1.6 Verify: `k3s kubectl get nodes` shows RED node Ready; `k3s kubectl get pods -A` shows core pods running
1.7 Copy RED kubeconfig to BLUE: `scp root@130.185.123.156:/etc/rancher/k3s/k3s.yaml /root/.kube/red-kubeconfig.yaml`
1.8 On BLUE: set up kubectl context for RED. Either use `KUBECONFIG=/root/.kube/red-kubeconfig.yaml` or add a context to BLUE's kubeconfig
1.9 Install tmux on RED: `apt-get install -y tmux` (for resilient command execution)
10. Open ufw on RED: allow 22 from BLUE's IP (or from anywhere if needed), allow 80, allow NodePorts (30081-30084, 30091-30094 — exact ports TBD)
11. Disable Docker on RED if present (k3s uses containerd; Docker is redundant and wastes RAM): `systemctl stop docker && systemctl disable docker`
12. Create `/root/release-notes/02-red-k3s-setup.md`

**Verification:** `kubectl --context=red get nodes` → RED Ready; kube-system pods running; SSH from BLUE works; tmux installed on RED.

---

## Step 2: Command Execution Framework (Resilient)

**Purpose:** Ensure commands sent to RED survive SSH drops and can be checked later. This is critical because the servers are over the internet — connections can break.

**Substeps:**

2.1 On RED: create `/root/red-cmd.sh` helper that runs a command in a tmux session:
```bash
#!/bin/bash
# Usage: red-cmd <session-name> <command>
# Runs command in detached tmux session named <session-name>
# If session exists, attaches to it (command may still be running)
tmux new-session -d -s "$1" bash
tmux send-keys -t "$1" "$2" Enter
```
2.2 Store logs on RED: each command's output goes to `/root/red-logs/<session-name>.log` (tmux capture-pane or redirect within command)
2.3 On BLUE: create `/root/red-run.sh`:
```bash
#!/bin/bash
# Usage: red-run <command>
# Executes command on RED via SSH, returns immediately
ssh -o ServerAliveInterval=30 -o ServerAliveCountMax=3 root@130.185.123.156 "$1"
```
2.4 On BLUE: create `/root/red-status.sh` that checks RED health:
```bash
#!/bin/bash
echo "=== RED tmux sessions ==="
ssh root@130.185.123.156 "tmux list-sessions"
echo "=== RED k3s nodes ==="
ssh root@130.185.123.156 "k3s kubectl get nodes"
echo "=== RED pods ==="
ssh root@130.185.123.156 "k3s kubectl get pods -A"
```
2.5 For long-running operations (image import, builds): use tmux session on RED, e.g.:
`ssh root@130.185.123.156 "tmux new-session -d -s img-import 'docker save ... | k3s ctr images import - > /root/red-logs/img-import.log 2>&1'"`
2.6 Check progress: `ssh root@130.185.123.156 "tmux capture-pane -t img-import -p"` or `tail /root/red-logs/img-import.log`
2.7 Create `/root/release-notes/03-red-command-framework.md`

**Verification:** Run a test command via tmux on RED. Disconnect SSH. Reconnect. Verify tmux session still exists and command output is in the log file.

---

## Step 3: Migrate PostgreSQL to RED

**Purpose:** Move the database from BLUE's k3s to RED's k3s. Single PostgreSQL instance with two logical databases (appdb_staging, appdb_production).

**Substeps:**

3.1 On RED: pull postgres:16-alpine image and import to containerd:
`docker pull postgres:16-alpine 2>/dev/null || skopeo ...` — since registries may be blocked, try pulling via Docker (if Docker is available for pull only) or find alternative. If registries are blocked like on BLUE, use an alternative approach (manual download + import, or use a working mirror).
3.2 Create `/root/project/k8s/database/postgres-red.yaml` on BLUE with:
   - Namespace: database
   - Secret: POSTGRES_USER=app, POSTGRES_PASSWORD=app_password (same as BLUE's), POSTGRES_DB=postgres
   - StatefulSet: postgres:16-alpine, 1 replica, volume claim (hostPath since single node) at `/var/lib/postgresql/data`
   - Service: ClusterIP, port 5432, named postgres
3.3 Deploy to RED: `kubectl --context=red apply -f /root/project/k8s/database/postgres-red.yaml`
3.4 Verify pod Running: `kubectl --context=red get pods -n database`
3.5 Create logical databases: exec into postgres pod, run:
   `CREATE DATABASE appdb_staging;`
   `CREATE DATABASE appdb_production;`
3.6 Verify from BLUE: `kubectl --context=red exec -n database postgres-0 -- psql -U app -d appdb_staging -c "SELECT 1"` returns successfully
3.7 Also verify from RED directly if needed
3.8 Delete old PostgreSQL from BLUE's k3s: `kubectl delete -f /root/project/k8s/database/postgres.yaml` (namespace database on BLUE)
3.9 Create `/root/release-notes/04-red-postgresql.md`

**Verification:** PostgreSQL pod Running on RED; both logical DBs exist; can connect from BLUE via kubectl exec; old DB removed from BLUE.

---

## Step 4: Set Up Docker Compose PostgreSQL on BLUE

**Purpose:** Lightweight local PostgreSQL on BLUE for development, local testing, and running backend unit tests with a real DB.

**Substeps:**

4.1 Verify Docker is installed and running on BLUE (it should be — Docker was used for image builds)
4.2 Create `/root/project/docker-compose.yml`:
```yaml
version: "3.8"
services:
  dev-db:
    image: postgres:16-alpine
    container_name: pawradise-dev-db
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: app
      POSTGRES_PASSWORD: dev_password
      POSTGRES_DB: devdb
    volumes:
      - ./pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U app -d devdb"]
      interval: 5s
      timeout: 3s
      retries: 5
```
4.3 `cd /root/project && docker compose up -d`
4.4 Verify: `docker compose ps` shows healthy; `psql -h 127.0.0.1 -U app -d devdb -c "SELECT 1"` works
4.5 Create additional databases for testing: `psql -h 127.0.0.1 -U app -d devdb -c "CREATE DATABASE devdb_staging; CREATE DATABASE devdb_production;"`
4.6 This DB is used for:
   - Running backend unit tests with real DB (TEST_DATABASE_URL=postgres://app:dev_password@localhost:5432/devdb_staging)
   - Local development testing
   - E2E test database (if needed)
4.7 Create `/root/release-notes/05-blue-dev-database.md`

**Verification:** `docker compose ps` shows healthy container; can connect via psql; both devdb_staging and devdb_production exist.

---

## Step 5: Clean Up BLUE (Remove Kubernetes Resources)

**Purpose:** Remove k3s resources from BLUE that have been migrated to RED, freeing RAM and removing confusion.

**Substeps:**

5.1 List all resources on BLUE: `KUBECONFIG=/etc/rancher/k3s/k3s.yaml kubectl get all --all-namespaces`
5.2 Save a backup of BLUE's manifests to `/root/project/k8s-blue-backup/` before deleting (in case something needs to be recreated)
5.3 Delete staging namespace (backend, frontend, postgres — migrated to RED):
   `KUBECONFIG=/etc/rancher/k3s/k3s.yaml kubectl delete namespace staging`
5.4 Delete production namespace:
   `KUBECONFIG=/etc/rancher/k3s/k3s.yaml kubectl delete namespace production`
5.5 Delete database namespace:
   `KUBECONFIG=/etc/rancher/k3s/k3s.yaml kubectl delete namespace database`
5.6 Keep kube-system (k3s needs it), kube-public, kube-node-lease
5.7 Remove unused Docker images from BLUE to free disk space:
   `docker images | grep pawradise` — keep latest backend and frontend images (for building), remove old variants (local, react-csp, etc.)
5.8 Stop nginx on BLUE if it was serving as host proxy (RED takes over):
   `systemctl stop nginx && systemctl disable nginx` (if nginx was only for pawradise proxy)
5.9 Update BLUE's ufw: remove NodePort allow rules for 30081-30084, 30091-30094 (these were for the old cluster). Keep rules BLUE still needs.
5.10 Verify BLUE's k3s still functional: `KUBECONFIG=/etc/rancher/k3s/k3s.yaml kubectl get nodes` shows node; kube-system pods running
5.11 Create `/root/release-notes/06-blue-cleanup.md`

**Verification:** `kubectl get namespaces` on BLUE shows only kube-system, kube-public, kube-node-lease. No staging/production/database. nginx stopped. ufw cleaned up.

---

## Step 6: Deploy Backend to RED

**Purpose:** Build backend image on BLUE, export to RED, deploy to RED with correct DATABASE_URL pointing to RED's PostgreSQL.

**Substeps:**

6.1 Ensure backend code is current on BLUE: `cd /root/project/backend && git status` (or verify files)
6.2 Build Go binary on BLUE: `cd /root/project/backend && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o backend-static .`
6.3 Build Docker image on BLUE: `docker build -t registry.local/pawradise-backend:latest -f /root/project/backend/Dockerfile /root/project/backend`
6.4 Export image to RED via SSH pipe:
   `docker save registry.local/pawradise-backend:latest | ssh root@130.185.123.156 "k3s ctr images import -"`
   (if this is slow/large, use tmux session on RED)
6.5 Verify image on RED: `ssh root@130.185.123.156 "k3s ctr images list | grep pawradise-backend"`
6.6 Create/verify backend manifests for RED (reuse existing `/root/project/k8s/staging/backend.yaml` and `/root/project/k8s/production/backend.yaml`):
   - Update DATABASE_URL to use RED's PostgreSQL service DNS: `postgres://app:app_password@postgres.database.svc.cluster.local:5432/appdb_staging?sslmode=disable`
   - Update JWT_SECRET to a new staging secret (or reuse)
   - Set ADMIN_ENABLED=true, ADMIN_TOKEN=admin-secret-token-change-in-production
6.7 Deploy to RED:
   `kubectl --context=red apply -f /root/project/k8s/staging/backend.yaml`
   `kubectl --context=red apply -f /root/project/k8s/production/backend.yaml`
6.8 Wait for rollout:
   `kubectl --context=red rollout status deployment/backend -n staging --timeout=120s`
   `kubectl --context=red rollout status deployment/backend -n production --timeout=120s`
6.9 Test backend health on RED:
   `curl http://130.185.123.156:30083/api/v1/health` (staging NodePort)
   `curl http://130.185.123.156:30093/api/v1/health` (production NodePort)
6.10 Test full auth flow via RED:
   - Register: `curl -X POST http://130.185.123.156:30083/api/v1/register -H "Content-Type: application/json" -d '{"email":"red-test@pawradise.ir","password":"RedTest123","name":"RedTest"}'`
   - Login: `curl -X POST http://130.185.123.156:30083/api/v1/login -H "Content-Type: application/json" -d '{"email":"red-test@pawradise.ir","password":"RedTest123"}'`
   - Profile: use returned token to GET `/api/v1/me`
6.11 Verify admin endpoints on RED:
   `curl -H "Authorization: Bearer admin-secret-token-change-in-production" http://130.185.123.156:30083/api/v1/admin/users`
6.12 Create `/root/release-notes/07-red-backend.md`

**Verification:** Backend pods Running on RED (both namespaces); /api/v1/health returns 200; register/login/me work; admin endpoints work; JWT validation works.

---

## Step 7: Deploy Frontend to RED

**Purpose:** Build frontend image on BLUE, export to RED, deploy to RED.

**Substeps:**

7.1 Ensure frontend code is current on BLUE: check `/root/project/frontend/public/index.html` and `admin.html`
7.2 Build Docker image on BLUE: `docker build -t registry.local/pawradise-frontend:latest -f /root/project/frontend/Dockerfile /root/project/frontend`
7.3 Export to RED:
   `docker save registry.local/pawradise-frontend:latest | ssh root@130.185.123.156 "k3s ctr images import -"`
7.4 Verify image on RED
7.5 Deploy to RED using existing manifests:
   `kubectl --context=red apply -f /root/project/k8s/staging/frontend.yaml`
   `kubectl --context=red apply -f /root/project/k8s/production/frontend.yaml`
7.6 Wait for rollout
7.7 Test frontend via NodePorts:
   `curl http://130.185.123.156:30084/` (staging) — should return HTML with profileView, Logout, ENV: staging
   `curl http://130.185.123.156:30094/` (production) — should return HTML with profileView, Logout, ENV: production
7.8 Test admin.html:
   `curl http://130.185.123.156:30084/admin.html` (staging admin)
   `curl http://130.185.123.156:30094/admin.html` (production admin)
7.9 Create `/root/release-notes/08-red-frontend.md`

**Verification:** Frontend pods Running on RED; correct HTML served at NodePorts; profileView and Logout present in HTML.

---

## Step 8: Set Up Host Nginx on RED (Port 80 Path Routing)

**Purpose:** Route all traffic on port 80 on RED to the appropriate backend/frontend pods based on path, matching the required URL structure.

**Substeps:**

8.1 Install nginx on RED: `apt-get install -y nginx`
8.2 Remove default site: `rm /etc/nginx/sites-enabled/default`
8.3 Create `/etc/nginx/sites-available/pawradise` on RED with:
```
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    server_name _;

    # Production frontend
    location / {
        proxy_pass http://127.0.0.1:30094;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Staging frontend
    location /staging/ {
        rewrite ^/staging/(.*) /$1 break;
        proxy_pass http://127.0.0.1:30084;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Production backend API
    location /api/v1/ {
        proxy_pass http://127.0.0.1:30093/api/v1/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Authorization $http_authorization;
    }

    # Staging backend API
    location /staging/api/v1/ {
        rewrite ^/staging/api/v1/(.*) /api/v1/$1 break;
        proxy_pass http://127.0.0.1:30083;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Authorization $http_authorization;
    }

    # Production admin API
    location /api/v1/admin/ {
        proxy_pass http://127.0.0.1:30093/api/v1/admin/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Authorization $http_authorization;
    }

    # Staging admin API
    location /staging/api/v1/admin/ {
        rewrite ^/staging/api/v1/admin/(.*) /api/v1/admin/$1 break;
        proxy_pass http://127.0.0.1:30083;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Authorization $http_authorization;
    }

    # Production admin UI
    location = /admin {
        proxy_pass http://127.0.0.1:30094/admin.html;
        proxy_set_header Host $host;
    }

    # Staging admin UI
    location = /staging/admin {
        rewrite ^/staging/admin(.*) /admin.html break;
        proxy_pass http://127.0.0.1:30084;
        proxy_set_header Host $host;
    }

    # Health check endpoints
    location = /health {
        proxy_pass http://127.0.0.1:30093/api/v1/health;
    }
    location = /staging/health {
        proxy_pass http://127.0.0.1:30083/api/v1/health;
    }
}
```
8.4 Enable: `ln -sf /etc/nginx/sites-available/pawradise /etc/nginx/sites-enabled/pawradise`
8.5 Test config: `nginx -t`
8.6 Start/restart nginx: `systemctl restart nginx`
8.7 Verify routing on RED:
   - `curl http://130.185.123.156/` → 200, ENV: production
   - `curl http://130.185.123.156/staging/` → 200, ENV: staging
   - `curl http://130.185.123.156/api/v1/health` → 200
   - `curl http://130.185.123.156/staging/api/v1/health` → 200
   - `curl http://130.185.123.156/admin` → 200 (admin UI)
   - `curl http://130.185.123.156/staging/admin` → 200 (admin UI)
   - `curl -H "Authorization: Bearer admin-secret-token-change-in-production" http://130.185.123.156/api/v1/admin/users` → 200 (admin API)
8.8 Open port 80 on RED's ufw if not already: `ufw allow 80/tcp`
8.9 Create `/root/release-notes/09-red-nginx-routing.md`

**Verification:** All routing paths return expected status codes and content. Admin API works through nginx.

---

## Step 9: Full End-to-End Verification on RED

**Purpose:** Verify the entire system works end-to-end through RED's nginx on port 80.

**Substeps:**

9.1 Verify all pods Running on RED:
   `kubectl --context=red get pods -A`
9.2 Verify routing (all via RED's IP):
   - `/` → prod frontend (200, ENV: production, profileView present, Logout present)
   - `/staging/` → staging frontend (200, ENV: staging, profileView present, Logout present)
   - `/admin` → prod admin UI (200, Admin panel title)
   - `/staging/admin` → staging admin UI (200)
   - `/api/v1/health` → 200
   - `/staging/api/v1/health` → 200
   - `/api/v1/admin/users` (with admin token) → 200
   - `/staging/api/v1/admin/users` (with admin token) → 200
9.3 Verify auth flow end-to-end:
   - Register on prod: POST `/api/v1/register` → 201 with JWT
   - Login on prod: POST `/api/v1/login` → 200 with JWT
   - Get profile: GET `/api/v1/me` with JWT → 200 with user data
   - Register on staging: POST `/staging/api/v1/register` → 201
   - Login on staging: POST `/staging/api/v1/login` → 200
   - Get profile: GET `/staging/api/v1/me` with JWT → 200
9.4 Verify admin flow:
   - List users (prod): GET `/api/v1/admin/users` with admin token → 200
   - List users (staging): GET `/staging/api/v1/admin/users` with admin token → 200
9.5 Run backend unit tests against RED's PostgreSQL:
   `cd /root/project/backend && TEST_DATABASE_URL="postgres://app:app_password@130.185.123.156:5432/testdb_staging?sslmode=disable" go test -cover ./handlers/ ./database/`
   (if direct PostgreSQL access from BLUE works; otherwise use kubectl exec to run tests in a pod on RED)
9.6 Create `/root/release-notes/10-red-end-to-end-verification.md`

**Verification:** All routes return expected status codes. Auth flow works. Admin flow works. Unit tests pass with 50%+ coverage.

---

## Step 10: Update Domain and Final Integration

**Purpose:** Point `pawradise.ir` to RED's IP and verify everything works through the domain.

**Substeps:**

10.1 Update DNS / hosts entry: `pawradise.ir` → `130.185.123.156` (RED's IP)
   - If using hosts file: update `/etc/hosts` on test machines
   - If using DNS: update A record for `pawradise.ir` to RED's IP
   - Remove old entry pointing to BLUE if present
10.2 Verify domain routing:
   - `curl http://pawradise.ir/` → 200, ENV: production
   - `curl http://pawradise.ir/staging/` → 200, ENV: staging
   - `curl http://pawradise.ir/admin` → 200
   - `curl http://pawradise.ir/staging/admin` → 200
   - API endpoints via domain work
10.3 If domain doesn't work (VPN interference as before), test via IP: `curl http://130.185.123.156/` should work
10.4 Create `/root/release-notes/11-red-domain-integration.md`

**Verification:** `pawradise.ir` resolves to RED's IP and all routes work through the domain.

---

## Step 11: Playwright E2E Tests (Attempt)

**Purpose:** Try to run Playwright E2E tests again now that RED is a fresh server. The browser download may work from BLUE to RED or via RED.

**Substeps:**

11.1 Check if Playwright chromium binary can be downloaded now (fresh server, possibly different network path):
   - Try `npx playwright install chromium` on BLUE targeting `http://130.185.123.156` as base URL
   - If download still fails, consider downloading on RED directly (RED may have different network access)
   - Alternative: install a headless browser on RED (e.g., via apt if snap isn't required) and run Playwright there
11.2 Update `playwright.config.js` base URL to `http://130.185.123.156` (or `http://pawradise.ir` if domain works)
11.3 Run E2E tests: `cd /root/project/tests && npx playwright test`
11.4 If Playwright still cannot get a browser, document the blocker and note that E2E tests are designed but cannot run without a browser binary
11.5 Create `/root/release-notes/12-red-playwright-tests.md`

**Verification:** Playwright tests either pass or blocker is documented with clear explanation.

---

## Step 12: Final Documentation and State Cleanup

**Purpose:** Ensure all release notes are complete, all manifests are correct, all secrets are documented (in release notes, not plaintext), and the system is in a clean state.

**Substeps:**

12.1 Review all release notes in `/root/release-notes/`:
   - 01-red-server-provisioning.md
   - 02-red-k3s-setup.md
   - 03-red-command-framework.md
   - 04-red-postgresql.md
   - 05-blue-dev-database.md
   - 06-blue-cleanup.md
   - 07-red-backend.md
   - 08-red-frontend.md
   - 09-red-nginx-routing.md
   - 10-red-end-to-end-verification.md
   - 11-red-domain-integration.md
   - 12-red-playwright-tests.md
12.2 Verify all manifests in `/root/project/k8s/` are correct for RED
12.3 Verify docker-compose.yml on BLUE is correct
12.4 Verify red-run.sh, red-status.sh helpers work
12.5 Final status check:
   - RED: all pods Running, nginx serving on port 80, all routes verified
   - BLUE: Docker Compose PG running, k3s cleaned up, build tools available
12.6 Create final summary release note if needed

**Verification:** All release notes present and accurate. System in clean state. Both servers fulfilling their roles.

---

## Rules

1. **Always create a release note for each major step** in `/root/release-notes/` before starting the step. Name format: `NN-description.md`.
2. **Record progress in release notes** — what was done, what was verified, any issues encountered.
3. **Commands on RED use tmux for heavier or longer commands** — this ensures they survive SSH disconnects.
4. **Verify after each step** — run the verification criteria before proceeding.
5. **Always target RED explicitly** — use `kubectl --context=red` or `ssh root@130.185.123.156` to avoid accidentally targeting BLUE.
6. **Keep the plan updated** — if substeps change or new ones are needed, update this file.
7. **Make BLUE lighter as we go forward** - the problem started from where BLUE would go unresponsive so if you can remove anything from BLUE that is moved to RED and not needed do it on the way

---

## Open Questions (to resolve before/during execution)

- **PostgreSQL image availability on RED:** If registry.k8s.io / docker.io are blocked on RED like they were on BLUE, how do we get postgres:16-alpine? (manual download? alternative registry? pre-loaded image?)  ===> since we should already have the image on BLUE, it shouldn't be a problem.
- **NodePort port selection:** Confirm exact NodePort numbers (30081-30084 for staging frontend/backend, 30091-30094 for production — or different scheme). These must match nginx config. ===> I wonder why we need to have port numbers. Everything eventually is served on 80 or 443 using different paths. if you use differnt services you wouldn't need node ports. if I'm not wrong.
- **JWT secret and ADMIN_TOKEN values:** Reuse existing values or generate new ones for RED? no just move these to RED
- **Database password:** Reuse `app_password` or change? ===> Create a safe very lightweight vault in BLUE and keep all the passwords. Use different ones for development and staging and production.
- **Volume for PostgreSQL on RED:** hostPath at what path? RED may have different disk layout. ==> red again just blue you set it up
- **Docker availability on RED for pulling images:** Is Docker installed on RED, or only containerd? If only containerd, we need another way to get postgres image onto RED.  ===> RED is a simple empty server like BLUE was when I gave it to you
