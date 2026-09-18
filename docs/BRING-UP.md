# Bring-up: two servers → live Store4bots

Give an agent: this repo, SSH to **BLUE** (build) and **RED** (serve), and one or two DNS names whose A records point at RED. The agent should follow this file plus `.agent/skills/` until staging is seeded, TLS is valid, and Playwright can hit `https://$STAGING_HOST`.

All IPs and hostnames live in **`config/site.env`**. Secrets in **`config/site.secrets.env`**. Source them with `. scripts/lib/site-env.sh`.

## 0. Identity file

On BLUE, in the cloned repo:

```bash
cp config/site.env.example config/site.env
cp config/site.secrets.env.example config/site.secrets.env
# edit both: BLUE_HOST, RED_HOST, STAGING_HOST, optional PRODUCTION_HOST, passwords
```

`$STAGING_HOST` (and `$PRODUCTION_HOST` if used) must already resolve to `$RED_HOST` before TLS.

## 1. BLUE (build host)

Skill: `store4bots-prepare-blue`. Script: `scripts/prepare-blue.sh`.

Installs Git, Docker, Go 1.22+, Node 20+, Playwright browsers. Does **not** change sshd until `scripts/harden-ssh.sh blue` after a second SSH session with your key works (`store4bots-security`).

## 2. RED (serving host)

Skill: `store4bots-prepare-red`. Script: `scripts/prepare-red.sh`.

Order is fixed:

1. SSH + firewall allow SSH **before** UFW (`scripts/harden-ssh.sh red` only after key login works).
2. Docker Engine.
3. Host Postgres 16 + Mongo 7 (`scripts/install-host-db.sh`).
4. k3s with Traefik off.
5. Namespaces `staging`, `production`, `registry`, `ingress-nginx`.
6. Secret `store4bots-secrets` in `staging` and `production`.
7. ingress-nginx (`scripts/render-k8s.sh` then apply `k8s/generated/ingress-nginx-*.yaml`). NodePort `$INGRESS_HTTP_NODEPORT`.
8. Host nginx + certbot (`scripts/install-host-nginx.sh`). Staging proxies **all** paths, including `/assets/`, `/registry`, `/v2`. `client_max_body_size 0`.
9. In-cluster registry (`scripts/setup-registry.sh`).
10. Apply app manifests from `k8s/generated/`.
11. Migrations (`scripts/migrate.sh`) then `scripts/seed-staging.sh`.
12. Build/push from BLUE (`scripts/build-push-deploy.sh`).

## 3. DNS and TLS

A records: `$STAGING_HOST` → `$RED_HOST`. Optional `$PRODUCTION_HOST` → `$RED_HOST`.

Certbot on RED for those names. k3s image pulls use public CA — do not set `insecure_skip_verify`.

## 4. Prove it

```bash
. scripts/lib/site-env.sh
curl -sS "https://${STAGING_HOST}/api/v1/products?per_page=1" | head
curl -sS -o /dev/null -w '%{http_code}\n' "https://${STAGING_HOST}/"
# unauthenticated registry probe
curl -sS -o /dev/null -w '%{http_code}\n' "https://${REGISTRY_HOST}/v2/"   # expect 401
cd tests
BASE_URL="https://${STAGING_HOST}" node e2e-suite.js
BASE_URL="https://${STAGING_HOST}" npx playwright test --config frontend/playwright.config.cjs --workers=1
```

HTTP 200 on `/` is not enough. Playwright must see rendered shop cards, not the nginx welcome page.

## 5. Move RED later

Same as a new RED: `scripts/migrate-red.sh dump` on the old host, provision the new VM, point `RED_HOST` at it, `scripts/migrate-red.sh restore`, re-push images, confirm `https://$STAGING_HOST`, then switch DNS A records and retire the old VM. Do not keep old IPs in docs.

## Skills map

| Area | Skill |
|------|--------|
| BLUE | `store4bots-prepare-blue` |
| RED | `store4bots-prepare-red` |
| Product | `store4bots-product` |
| Go APIs | `store4bots-backend` |
| MFEs | `store4bots-frontend` |
| k3s | `store4bots-kubernetes` |
| DNS/TLS/nginx | `store4bots-network` |
| Postgres/Mongo | `store4bots-databases` |
| Tests | `store4bots-testing` |
| Images | `store4bots-build-deploy` |
| Hardening | `store4bots-security` |
| Docs | `store4bots-maintain-docs` |
| Shape of the system | `store4bots-architecture` |
