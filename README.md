# Store4bots (digital assets shop)

Single-seller digital marketplace: **9 Go services** + **7 Alpine.js microfrontends**.

A new operator should be able to rent two VMs (BLUE = build, RED = serve), point one or two domains at RED, clone this repo, and follow `docs/BRING-UP.md` until staging is seeded and live.

## This install’s hosts

All IPs and domain names live in **`config/site.env`**. Edit that file for a new pair of servers. Secrets go in `config/site.secrets.env` (not in git).

## Layout

```
config/            # site.env — only place for IPs/domains
docs/              # current-state guides (start at BRING-UP.md)
.agent/skills/     # project skills (any agent)
frontend/          # 7 MFEs
services/          # 9 Go services + shared
shared/            # theme, api.js, chrome, MFE build
k8s/               # canonical manifests (render before apply)
scripts/           # bring-up, seed, registry, migrate-red
tests/             # unit, API, Playwright
```

## Docs

| File | What |
|------|------|
| `docs/BRING-UP.md` | Two servers → live site |
| `docs/PRODUCT-SPEC.md` | Product contract |
| `docs/FEATURE-CANDIDATES.md` | Feature in/out |
| `docs/ARCHITECTURE.md` | Services, MFEs, data ownership |
| `docs/DEPLOYMENT.md` | Registry, k3s, build/push/deploy |
| `docs/NETWORK.md` | DNS, TLS, nginx, ingress |
| `docs/DATABASES.md` | Postgres, Mongo, migrate, seed |
| `docs/API.md` | Auth and public API notes |
| `docs/TESTING.md` | How to test |
| `docs/SPEC-GAPS.md` | Remaining spec items (TDD queue) |
| `tests/REPORT.md` | Live test counts (only place for pass numbers) |

## Test (from BLUE, against staging)

```bash
. scripts/lib/site-env.sh
cd tests
BASE_URL="https://${STAGING_HOST}" node e2e-suite.js
BASE_URL="https://${STAGING_HOST}" npx playwright test --config frontend/playwright.config.cjs
```
