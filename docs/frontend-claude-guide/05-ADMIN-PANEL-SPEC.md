# Pawradise — Admin Panel Spec

> Standalone app, **not** a microfrontend (per your instruction). One
> Deployment + Service, own build, own tests, served at `/admin`
> (`/staging/admin` in staging) via the same ingress mechanism as every
> other frontend route (`01-FRONTEND-ARCHITECTURE.md` §3).

---

## 1. Why it's different from the public MFEs

The operator is one person (you), using it constantly, on a desktop most of
the time, and it's data-dense (11 tabs, CRUD everything). Priorities invert
relative to the public site: **density and speed over persuasion/marketing
polish**, but it still must not be broken on a phone if you need to update
something on the go, and it still uses the same design tokens (ADR-5) so it
doesn't look like a foreign app bolted onto Pawradise.

## 2. Auth

Per `PRODUCT-SPEC.md` §2.4, admin auth is a single static Bearer token
(`ADMIN_TOKEN` secret), not a per-user login. Implementation:
- A gate screen: one password-style input, "Enter." Submitting stores the
  token in `sessionStorage` under `admin_token` (deliberately **not**
  `localStorage` — a privileged token should not silently persist across
  browser restarts; re-entering it each session is the correct tradeoff).
- Every `apiFetch` call from the Admin app adds
  `Authorization: Bearer <sessionStorage.admin_token>` — this is the one
  place in the whole frontend that manually attaches a header rather than
  relying on the cookie (ADR-2 is for customer auth; admin intentionally
  stays separate and explicit, matching the existing backend contract).
- A 401 from any admin call clears the stored token and re-shows the gate
  screen.

## 3. Structure

Single page, tab-based (Alpine `x-data="{ tab: 'stats' }"`, no full
navigation between tabs — this is the one part of the frontend where a true
single-page interaction model is worth it, because it's one operator moving
fast between related views, not a customer navigating a shop). Tabs mirror
`PRODUCT-SPEC.md` §6 exactly:

| Tab | Core actions | Notes |
|---|---|---|
| Stats | revenue chart (daily/weekly/monthly), top-selling products, conversion funnel, totals | read-only |
| Users | list (paginated), delete, reset password | destructive actions get a confirm modal |
| Products | CRUD, tiers, image gallery + trigger preview generation, pin/unpin, PWYW toggle, bulk status/category update | the busiest tab — see §4 |
| Bundles | CRUD, product picker, price vs. sum-of-parts preview | |
| Coupons | CRUD, usage stats | |
| Orders | list, detail, status transition (with the same validated state machine as the backend — UI should only offer valid next states, not all of them) | |
| Community | list posts, delete (moderation) | |
| Referrals | view all referrals + commissions | read-only |
| Exchange Rates | list, set, delete | |
| Settings | key-value list/edit, including `referral_commission_percent` | this is the one setting every other flow in the public site depends on — surface it clearly, not buried |
| Email Export | filter by date range/category/product, export CSV/JSON | triggers a file download, not an inline view |

## 4. Products tab — the one screen worth extra care

This tab does the most: create/edit a product with tiers, an image gallery,
manual/triggered preview generation, pinning, and PWYW — all from
`PRODUCT-SPEC.md` §6.2. Build it test-first like everything else:

- Acceptance criteria to turn into Playwright specs before building:
  - creating a product without an uploaded asset is rejected with a clear
    field error, not a generic 500-style failure
  - adding a tier requires both a file and a price; tiers list shows price
    ascending by default
  - uploading an image for an image/GIF product shows a "generating
    preview…" state and then the resulting thumbnail — the UI should not
    let the admin assume the raw upload is what buyers will see
  - toggling PWYW on reveals a required minimum-price field; toggling off
    hides and clears it
  - pin/unpin is reflected immediately in this tab's list ordering
  - bulk operations (status/category) require an explicit multi-select and
    confirm before applying — never a silent "select all" default

## 5. Tech

Same shared tokens, same `shared/lib/api.js` pattern, same Alpine.js — no
new framework introduced just for Admin. If the Products tab's state
management gets genuinely unwieldy in plain Alpine (many interdependent
fields), it's acceptable to add a build step (Vite) **for this app only**
to organize it into a few Alpine plugin modules — but try plain Alpine
first; don't reach for tooling preemptively.

## 6. Responsive behavior

Desktop-first is fine as the primary target, but every table (users,
products, orders, referrals) must degrade to stacked cards below the `md`
breakpoint (`03-DESIGN-SYSTEM.md` §3) rather than forcing horizontal
scroll — one card per row, label:value pairs stacked, actions as a
dropdown/menu instead of a row of icon buttons that won't fit.

## 7. Testing

Same TDD loop as `04-TDD-GUIDE.md`, in `tests/frontend-e2e/admin/`:
`admin-auth.spec.ts`, `admin-products.spec.ts`, `admin-orders.spec.ts`,
`admin-coupons.spec.ts`, `admin-settings.spec.ts`, etc. — one file per tab,
same red/green/refactor workflow, same dual mobile/desktop Playwright
projects.

## 8. Ingress & env

Exactly the pattern in `01-FRONTEND-ARCHITECTURE.md` §3–4: `admin-app`
Deployment+Service in both `production` and `staging` namespaces, routed at
`/admin` and `/staging/admin` respectively, same `env.js` injection so it
knows which API base to call — the only difference from a public MFE is
that it's one app instead of several, and its auth is a manually-attached
Bearer header instead of the cookie.
