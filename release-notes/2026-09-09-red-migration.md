# RED Migration: 194.5.206.106 → 130.185.123.156

**Date:** 2026-09-09

## Objective

Migrate RED (serving cluster) from old 2GB server (194.5.206.106) to new 4GB server (130.185.123.156).

## Actions

1. **New RED provisioned** from scratch:
   - Ubuntu 24.04, k3s v1.36.4 with `--disable traefik`
   - Docker with insecure registry, local registry:30099
   - PostgreSQL:16-alpine with 16 logical databases
   - UFW, fail2ban, SSH hardening, automatic updates
   - Namespaces: production, staging, database
   - Secrets: pawradise-secrets (JWT_SECRET, ADMIN_TOKEN), postgres-secret

2. **Images transferred** from old RED's registry to new RED's registry
   - Pulled from 194.5.206.106:30099, pushed to 130.185.123.156:30099
   - Built fresh images with unique tag 1788981000 from local source

3. **K8s manifests created** for separate-pod deployment:
   - `production-deployments.yaml`: 8 Deployments + 8 Services (production)
   - `staging-deployments.yaml`: 8 Deployments + 8 Services (staging)
   - `production-ingress.yaml`: per-service routing via Prefix paths
   - `staging-ingress.yaml`: per-service routing with regex rewrite stripping `/staging` prefix

4. **Database migrations applied** to all 16 logical databases via docker exec postgres psql

5. **Host nginx** configured on port 80, proxying to ingress-nginx NodePort 30758

## Verification

| Component | Status |
|-----------|--------|
| Unit tests | 172/172 PASS |
| Integration tests | 28/28 PASS |
| E2E tests | 86/104 PASS, 18 FAIL |
| Production ingress | ✅ Working |
| Staging ingress | ⚠️ 404 on /staging routes (rewrite issue) |
| All pods Running | ✅ 16/16 |

## Remaining Issues

1. **Staging ingress rewrite**: Regex pattern not stripping `/staging` prefix correctly. Routes hit the pod but the path isn't rewritten, causing 404.
2. **18 E2E failures**: Community posts, admin products/coupons/community/posts/referrals/settings return `{"error":"failed"}` — likely DB column mismatches in admin-service queries.
3. **Old RED (194.5.206.106)**: Still running, awaiting decommission confirmation.

## New RED Details

- IP: 130.185.123.156
- Memory: 4GB total, ~2.9GB available after deployment
- Ingress-nginx: NodePort 30758 (HTTP), 30759 (HTTPS)
- PostgreSQL: Docker container on host, port 5432
- Registry: Docker container on host, port 30099
- Services connect via DB_HOST=130.185.123.156 (host IP, not k8s service name)
- Admin token: admin_secret_2026_prod
