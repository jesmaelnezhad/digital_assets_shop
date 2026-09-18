# Testing (current)

Skill: `store4bots-testing`. Pass counts: **`tests/REPORT.md` only**.

## Layers

| Layer | Where | Against |
|-------|--------|---------|
| Handler unit | `services/<svc>/handlers/*_test.go` | in-process |
| Spec unit | `tests/unit/<svc>/` | sqlmock / contracts |
| Integration | `tests/integration/` | cross-service |
| HTTP API | `tests/e2e-suite.js` | `https://$STAGING_HOST` |
| Browser | `tests/frontend/*.spec.cjs` | Playwright, staging **hostname** |

```bash
. scripts/lib/site-env.sh
export BASE_URL="https://${STAGING_HOST}"
export ADMIN_TOKEN   # from site.secrets.env
cd tests
node e2e-suite.js
npx playwright test --config frontend/playwright.config.cjs --workers=1
cd ../services/product-service && go test ./... -count=1
```

Staging HTTP/UI tests expect the **volume + roles** seeds (`scripts/seed-staging.sh`): enough products for “Show more”, Nia admin / Leo staff.

## Frontend

Use a real browser on `https://$STAGING_HOST`. HTTP 200 is not a pass:

- Not the nginx welcome page
- DOM actually rendered (product cards, not an empty shell)
- JS ran (Alpine lists from the API)
- **Computed styles**, not “a `<style>` tag exists”
- Flows: login, cart, checkout, admin tabs

```javascript
// required pattern
const bg = await page.evaluate(() => getComputedStyle(document.body).backgroundColor);
```

curl is fine for API status and JSON shape. It cannot prove the shop page.

## API assertions

Assert the real JSON (nested keys, 200 vs 201, null arrays). Do not rewrite the test to match a stub (`{"message":"…"}` is not a download).

## Spec gaps

`docs/SPEC-GAPS.md`: lowest FIX-ID → failing test for **Expected** → implement → delete that FIX section. Protected routes stay 401 without JWT. Do not copy another service’s tables to make a test pass.
