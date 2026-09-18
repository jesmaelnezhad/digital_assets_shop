# Store4bots agent instructions

You are bringing up or maintaining **Store4bots**, a single-seller digital-asset shop (Go microservices + Alpine.js microfrontends).

## Start here

1. Read `config/site.env` (hosts and domains for **this** install). Never copy those values into other files.
2. Follow `docs/BRING-UP.md` for a greenfield BLUE+RED pair.
3. Open the matching skill under `.agent/skills/` before changing that area.
4. Product contract: `docs/PRODUCT-SPEC.md`. Feature in/out: `docs/FEATURE-CANDIDATES.md`. Remaining spec gaps: `docs/SPEC-GAPS.md`.
5. After any change, update the skill + doc that describes that area (`store4bots-maintain-docs`).

## Topology (current)

- **BLUE**: build box (Go, Docker, git, tests). SSH from `config/site.env`.
- **RED**: serving box. Host Docker Postgres + Mongo. k3s (Traefik disabled). Host nginx terminates TLS and proxies to ingress-nginx NodePort `$INGRESS_HTTP_NODEPORT`. In-cluster authenticated registry in namespace `registry`.
- **Staging vs production**: identical URL paths (`/api/v1/...`). The environment is the hostname (`$STAGING_HOST` vs `$PRODUCTION_HOST`).
- **Images we build**: `$REGISTRY_HOST/$IMAGE_REPO/<name>:<tag>`. Stock images stay on Docker Hub / registry.k8s.io.

## Keep to the current design

- Public HTTP is `/api/v1/...` on each service. Ingress maps those paths.
- Seven MFEs, shared `shared/lib/api.js` + `shared/chrome/chrome.js` + `shared/theme/theme.css`.
- Each Go service owns its Postgres database (`appdb_<service>_<env>`). Call the owner over HTTP; do not copy another service’s tables.
- Frontend tests: Playwright against `https://$STAGING_HOST`. HTTP 200 is not proof of a working page.
- `CGO_ENABLED=0` for Linux binaries that run on Alpine.
- Do not commit `config/site.secrets.env`.
- Do not git commit unless the operator asks. Do not force-push.
- Docs describe **what is now**. Put reusable procedures in skills/scripts. Hosts live only in `config/site.env`.

## Skills

| Skill | When |
|-------|------|
| `store4bots-prepare-blue` | Provision or repair BLUE |
| `store4bots-prepare-red` | Provision or repair RED, or move RED to a new VM |
| `store4bots-backend` | Go services, APIs, migrations |
| `store4bots-frontend` | MFEs, theme, chrome, api.js |
| `store4bots-product` | What the product is and must do |
| `store4bots-kubernetes` | Namespaces, deployments, ingress objects |
| `store4bots-network` | DNS, TLS, host nginx, ingress-nginx, `/registry` `/v2` |
| `store4bots-databases` | Host Postgres + Mongo, schemas, seed |
| `store4bots-testing` | Unit, API, Playwright |
| `store4bots-build-deploy` | Build on BLUE, push registry, deploy RED |
| `store4bots-maintain-docs` | After every functional change |
| `store4bots-security` | Hardening without lockout |
| `store4bots-architecture` | Service/MFE map and framework choices |
