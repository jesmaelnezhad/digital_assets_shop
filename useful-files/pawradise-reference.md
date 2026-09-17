# Pawradise — RED/BLUE Split: URLs & Credentials Reference
# Generated: 2026-09-04
# Server RED (serving): 194.5.206.106
# Server BLUE (build/dev/admin): 130.185.121.83
#
# IMPORTANT: DNS/hosts must point pawradise.ir → 194.5.206.106
# Currently pawradise.ir resolves to 130.185.121.83 (BLUE) — update needed.

BASE_PROD="http://194.5.206.106"
BASE_STAGING="http://194.5.206.106/staging"

# ============================================================
# URL TABLE
# ============================================================

# --- PRODUCTION (pawradise.ir/) ---

# Frontend page
PROD_FRONTEND="$BASE_PROD/"

# Backend API
PROD_HEALTH="$BASE_PROD/api/v1/health"
PROD_REGISTER="$BASE_PROD/api/v1/register"
PROD_LOGIN="$BASE_PROD/api/v1/login"
PROD_ME="$BASE_PROD/api/v1/me"
PROD_LOGOUT="$BASE_PROD/api/v1/logout"

# Admin
PROD_ADMIN_UI="$BASE_PROD/admin"
PROD_ADMIN_USERS="$BASE_PROD/api/v1/admin/users"
PROD_ADMIN_RESET_PASSWORD="{id}"  # POST to $BASE_PROD/api/v1/admin/users/$id/reset-password
PROD_ADMIN_DELETE_USER="{id}"     # DELETE $BASE_PROD/api/v1/admin/users/$id

# --- STAGING (pawradise.ir/staging/) ---

# Frontend page
STAGING_FRONTEND="$BASE_STAGING/"

# Backend API
STAGING_HEALTH="$BASE_STAGING/api/v1/health"
STAGING_REGISTER="$BASE_STAGING/api/v1/register"
STAGING_LOGIN="$BASE_STAGING/api/v1/login"
STAGING_ME="$BASE_STAGING/api/v1/me"
STAGING_LOGOUT="$BASE_STAGING/api/v1/logout"

# Admin
STAGING_ADMIN_UI="$BASE_STAGING/admin"
STAGING_ADMIN_USERS="$BASE_STAGING/api/v1/admin/users"
STAGING_ADMIN_RESET_PASSWORD="{id}"  # POST to $BASE_STAGING/api/v1/admin/users/$id/reset-password
STAGING_ADMIN_DELETE_USER="{id}"     # DELETE $BASE_STAGING/api/v1/admin/users/$id

# ============================================================
# CREDENTIALS
# ============================================================

# Admin token (same for both environments)
ADMIN_TOKEN="admin-secret-token-change-in-production"
# Usage: Authorization: Bearer $ADMIN_TOKEN
# Example: curl -H "Authorization: Bearer $ADMIN_TOKEN" $PROD_ADMIN_USERS

# JWT Token
# Obtained from /register or /login response:
#   {"token": "eyJhbG...", "user": {...}}
# Usage: Authorization: Bearer <jwt_token>
# Example: curl -H "Authorization: Bearer $JWT" $PROD_ME

# PostgreSQL (on RED, in k3s database namespace)
PG_HOST="194.5.206.106"
PG_PORT="5432"
PG_USER="app"
PG_PASSWORD="[REDACTED]"
PG_DB_STAGING="appdb_staging"
PG_DB_PRODUCTION="appdb_production"
# Connection strings:
#   Production:  postgresql://app:[REDACTED]@194.5.206.106:5432/appdb_production?sslmode=disable
#   Staging:     postgresql://app:[REDACTED]@194.5.206.106:5432/appdb_staging?sslmode=disable
# Note: direct PG access requires k3s service exposure or kubectl port-forward.
#       Prefer testing via the API, not direct DB access.

# ============================================================
# TEST ACCOUNTS (currently existing)
# ============================================================

# Production
PROD_TEST_ACCOUNTS=(
  "final-test@pawradise.ir / Final123"
  "e2e-prod@pawradise.ir / E2EProd123"
  "red-final@pawradise.ir / RedFinal123"
)

# Staging
STAGING_TEST_ACCOUNTS=(
  "final-stage@pawradise.ir / FinalS123"
  "e2e-stage@pawradise.ir / E2EStage123"
)

# ============================================================
# SERVER INFRASTRUCTURE
# ============================================================

# RED (serving cluster)
RED_IP="194.5.206.106"
RED_SSH="ssh root@$RED_IP"
RED_KUBECONFIG="/root/.kube/red-kubeconfig.yaml"  # on BLUE
# kubectl against RED: KUBECONFIG=/root/.kube/red-kubeconfig.yaml kubectl ...
# Or: ssh root@194.5.206.106 "k3s kubectl ..."

# BLUE (build/dev/admin box)
BLUE_IP="130.185.121.83"
BLUE_PROJECT_DIR="/root/project"
BLUE_BACKEND_DIR="/root/project/backend"
BLUE_FRONTEND_DIR="/root/project/frontend"
BLUE_KUBECONFIG="/root/.kube/red-kubeconfig.yaml"

# Docker images on BLUE (for rebuilding)
#   registry.local/pawradise-backend:latest  (32.5MB)
#   registry.local/pawradise-frontend:latest (13.9MB)
# Build commands:
#   Backend: cd /root/project/backend && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o backend-static .
#            docker build -t registry.local/pawradise-backend:latest -f /root/project/backend/Dockerfile /root/project/backend
#   Frontend: docker build -t registry.local/pawradise-frontend:latest -f /root/project/frontend/Dockerfile /root/project/frontend
# Export to RED: docker save registry.local/pawradise-backend:latest | ssh root@194.5.206.106 "k3s ctr images import -"
#                docker save registry.local/pawradise-frontend:latest | ssh root@194.5.206.106 "k3s ctr images import -"

# RED pod status (ssh root@194.5.206.106 "k3s kubectl get pods -A")
#   database      postgres-0         Running
#   production    backend-*          Running  (NodePort 30093)
#   production    frontend-*         Running  (NodePort 30094, image :react)
#   staging       backend-*          Running  (NodePort 30083)
#   staging       frontend-*         Running  (image :latest)

# RED NodePorts
RED_NODEPORT_PROD_BACKEND="30093"
RED_NODEPORT_PROD_FRONTEND="30094"
RED_NODEPORT_STAGING_BACKEND="30083"
# staging frontend is ClusterIP only (served via host nginx static files)

# RED host nginx paths (port 80)
#   /               → prod frontend static (/var/www/production/)
#   /staging/       → staging frontend static (/var/www/staging/), rewrite strips prefix
#   /api/v1/        → prod backend NodePort 30093
#   /staging/api/v1/ → staging backend NodePort 30083
#   /admin          → prod frontend static (/var/www/production/admin.html)
#   /staging/admin  → staging frontend static (/var/www/staging/admin.html)

# ============================================================
# HELPER SCRIPTS ON BLUE
# ============================================================

# /root/red-run.sh       — SSH wrapper for running commands on RED
# /root/red-status.sh    — Health check for RED
# /root/red-kubectl.sh   — kubectl against RED (uses red-kubeconfig.yaml)
# Usage examples:
#   /root/red-status.sh
#   /root/red-kubectl.sh get pods -A
#   /root/red-run.sh "k3s kubectl get pods -A"

# ============================================================
# RELEASE NOTES
# ============================================================

# /root/release-notes/01-system-hardening.md       — BLUE hardening
# /root/release-notes/02-kubernetes-setup.md        — BLUE k3s + namespaces
# /root/release-notes/03-landing-page.md           — Initial landing pages
# /root/release-notes/04-nodeport-publishing.md    — NodePort + host nginx
# /root/release-notes/05-project-structure.md      — Project layout
# /root/release-notes/06-fullstack-auth.md         — Go backend with JWT
# /root/release-notes/07-react-frontend.md         — React frontend (later replaced)
# /root/release-notes/08-e2e-tests.md              — Playwright (failed install)
# /root/release-notes/09-white-page-fix.md         — White page investigation
# /root/release-notes/10-csp-fix.md                — CSP investigation
# /root/release-notes/11-vanilla-js-switch.md      — Vanilla JS frontend
# /root/release-notes/12-admin-panel-user-management.md — Admin panel
# /root/release-notes/01-red-server-provisioning.md — RED provisioning
# /root/release-notes/02-red-k3s-setup.md          — RED k3s installation
# /root/release-notes/03-red-command-framework.md  — RED command helpers
# /root/release-notes/04-red-postgresql.md         — RED PostgreSQL
# /root/release-notes/05-blue-dev-database.md      — Docker Compose PG on BLUE
# /root/release-notes/06-blue-cleanup.md           — BLUE k3s removal
# /root/release-notes/07-red-backend.md            — Backend on RED
# /root/release-notes/08-red-frontend.md           — Frontend on RED
# /root/release-notes/09-red-nginx-routing.md      — RED nginx path routing
# /root/release-notes/10-red-end-to-end-verification.md — E2E verification
# /root/release-notes/11-red-domain-integration.md — Domain integration (pending DNS)
# /root/release-notes/12-red-playwright-tests.md   — Playwright (skipped)
# /root/release-notes/13-red-backend-jwt-fix.md    — JWT middleware fix deployment
# /root/release-notes/14-red-blue-split-completion.md — Completion summary
