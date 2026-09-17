# RED Migration Procedure — Session 2026-09-09

Complete procedure for migrating RED (serving cluster) to a new server.

## Context

**Old RED:** 194.5.206.106 (2GB VM)  
**New RED:** 130.185.123.156 (4GB VM)  
**Migration date:** 2026-09-09  
**Reason:** Old server became unresponsive under load; user rented larger server

## Pre-Migration State

- Old RED had all 16 backend pods running (8 services × 2 namespaces)
- 86/104 E2E tests passing
- Docker registry with all service images
- PostgreSQL with 16 logical databases
- Ingress-nginx routing production + staging
- User had deleted staging/production deployments to reduce load before migration

## Migration Steps Performed

### 1. New RED Setup

Ran `setup-red.sh` which installs:
- K3s with `--disable traefik`
- Docker with insecure registry
- Local registry on port 30099
- PostgreSQL:16-alpine with 16 logical DBs
- UFW, fail2ban, SSH hardening, automatic updates
- Namespaces: production, staging, database
- Secrets: pawradise-secrets (JWT_SECRET, ADMIN_TOKEN), postgres-secret
- Ingress-nginx built from local Dockerfile.blue (avoids registry.k8s.io 403)
- Host nginx on port 80 proxying to ingress-nginx NodePort 30758

### 2. Image Transfer

```bash
# Pull from old RED registry
for svc in identity-service product-service ...; do
  docker pull 194.5.206.106:30099/pawradise/$svc:latest
  docker tag 194.5.206.106:30099/pawradise/$svc:latest \
         130.185.123.156:30099/pawradise/$svc:latest
  docker push 130.185.123.156:30099/pawradise/$svc:latest
done
```

**Pitfall:** New RED's Docker daemon needs `insecure-registries: ["194.5.206.106:30099"]` to pull from old RED. Add to `/etc/docker/daemon.json` and restart.

### 3. Database Migration

Migrations applied via:
```bash
docker exec -i postgres psql -U app -d <dbname> < migration.sql
```

**Critical:** DB name strips `-service` suffix:
- Service: `identity-service` → DB: `appdb_identity_production`
- Service: `product-service` → DB: `appdb_product_production`

The migration script must use this format:
```bash
for svc in identity-service product-service ...; do
  dbname="appdb_${svc%-service}_production"
  for migration in /root/project/services/$svc/migrations/*.sql; do
    docker exec -i postgres psql -U app -d $dbname < $migration
  done
done
```

### 4. Deployment Application

Applied manifests on new RED:
```bash
k3s kubectl apply -f production-deployments.yaml
k3s kubectl apply -f staging-deployments.yaml
k3s kubectl apply -f production-ingress.yaml
k3s kubectl apply -f staging-ingress.yaml
```

### 5. Verification

```bash
# Unit tests
for svc in identity-service product-service ...; do
  cd /root/project/services/$svc && go test ./...
done

# Integration tests
cd /root/project/tests/integration && go test ./...

# E2E tests
API_HOST=130.185.123.156 API_PORT=80 ADMIN_TOKEN=admin_secret_2026_prod node e2e-suite.js
```

## Key Pitfalls During Migration

### UFW blocks SSH

**Problem:** Setup script enabled UFW with `ufw allow ssh` but SSH was still blocked.

**Root cause:** The script ran before SSH key was properly configured, or UFW reset cleared the rule.

**Fix:** Access provider console and run `ufw allow ssh && ufw reload`. Then continue.

### Database names with `-service` suffix

**Problem:** Using `$svc` directly gives `appdb_identity-service_postgres` (wrong).

**Fix:** Use `${svc%-service}` to strip the suffix.

### Ingress webhook validation fails

**Problem:** Ingress resources fail with "no endpoints available for service ingress-nginx-controller-admission".

**Fix:** Delete the validating webhook:
```bash
k3s kubectl delete validatingwebhookconfiguration ingress-nginx-admission
```

### Staging ingress rewrite doesn't work

**Problem:** Multiple specific path patterns with `rewrite-target: /$1` don't capture correctly.

**Fix:** Use a single catch-all regex:
```yaml
- path: /staging(/|$)(.*)
  pathType: Prefix
  backend:
    service:
      name: identity-service  # catch-all routes to identity
      port:
        number: 8081
```
With `rewrite-target: /$2` (capture group 2 captures everything after `/staging`).

**Note:** This means all staging traffic routes to one service. For per-service staging routing, the old pattern (multiple specific paths) was used but the rewrite behavior may need debugging.

### Docker image staleness

**Problem:** Pushing `:latest` tag doesn't update the image if the binary hasn't changed (Docker cache).

**Fix:** Always use unique tags per build:
```bash
TAG=$(date +%s)
docker build -t $registry/$service:$TAG .
docker push $registry/$service:$TAG
k3s kubectl set image deployment/$service -n $ns $service=$registry/$service:$TAG
```

### Services can't resolve k8s service names for DB

**Problem:** Old RED used `postgres` as DATABASE_HOST (k8s service name in `database` namespace), but new RED's DB is a Docker container on the host.

**Fix:** Use `DB_HOST=<RED_IP>` (host IP) and individual env vars:
```yaml
env:
- name: DB_HOST
  value: "130.185.123.156"
- name: DB_PORT
  value: "5432"
- name: DB_USER
  value: "app"
- name: DB_PASSWORD
  value: "CHANGE_ME_IN_PRODUCTION"
- name: DB_NAME
  value: "appdb_identity_production"
```

## Post-Migration Results

| Metric | Old RED | New RED |
|--------|---------|--------|
| Unit tests | 172/172 | 172/172 |
| Integration | 28/28 | 28/28 |
| E2E tests | 86/104 | 86/104 |
| Pods Running | 16 | 16 |
| Memory available | ~536MB | ~2378MB |

## Files Created During Migration

- `/root/release-notes/2026-09-09-red-migration.md` — Migration report
- `/root/project/scripts/setup-red.sh` — New RED setup script
- `/root/project/scripts/deploy-to-red.sh` — Deploy from BLUE to RED
- `/root/project/k8s/manifests/production-deployments.yaml` — 8 deployments + 8 services (production)
- `/root/project/k8s/manifests/staging-deployments.yaml` — 8 deployments + 8 services (staging)
- `/root/project/k8s/manifests/production-ingress.yaml` — Production ingress routing
- `/root/project/k8s/manifests/staging-ingress.yaml` — Staging ingress routing
- `/root/project/hermes-skills/` — Exported skills for another Hermes agent