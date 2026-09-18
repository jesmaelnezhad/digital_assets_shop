---
name: store4bots-testing
description: Run and write Store4bots unit, API, and Playwright tests against staging. Use when adding tests, diagnosing failures, or verifying a deploy.
---

# Testing

Pass counts live **only** in `tests/REPORT.md`.

## Layers

| Layer | Where | Against |
|-------|--------|---------|
| Handler unit | `services/<svc>/handlers/*_test.go` | in-process |
| Spec unit | `tests/unit/<svc>/` | sqlmock / contracts |
| Integration | `tests/integration/` | cross-service |
| HTTP API | `tests/e2e-suite.js` | `https://$STAGING_HOST` |
| Browser | `tests/frontend/*.spec.cjs` | Playwright, staging domain |

```bash
. scripts/lib/site-env.sh
export BASE_URL="https://${STAGING_HOST}"
export ADMIN_TOKEN  # from site.secrets.env
cd tests
node e2e-suite.js
npx playwright test --config frontend/playwright.config.cjs --workers=1
cd ../services/product-service && go test ./... -count=1
```

## Frontend rule

Use a real browser on the staging **hostname**. HTTP 200 is not enough:

- Not the nginx welcome page
- DOM actually rendered
- JS ran (API-backed lists, not empty shells)
- Computed styles (not “a `<style>` tag exists”)
- Flows: login, cart, checkout, admin

curl is fine for API status/JSON shape, not for UI.

## TDD for spec gaps

`docs/SPEC-GAPS.md`: lowest FIX-ID → failing test asserting Expected → implement → delete that FIX section.

## Staging data

Volume + roles seeds must be present for pagination / Show more / RBAC tests (`scripts/seed-staging.sh`).
