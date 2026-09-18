---
name: store4bots-product
description: Store4bots product definition, features in/out, and user-facing behavior. Use when deciding what to build, writing tests against the spec, or changing shop/community/admin flows.
---

# Product

Contract: `docs/PRODUCT-SPEC.md`. Decisions: `docs/FEATURE-CANDIDATES.md`. Open gaps: `docs/SPEC-GAPS.md`.

Store4bots is a **single-seller** digital-asset shop plus a small community. Buy once, download forever. No license keys, no multi-vendor.

## In (implemented or specified)

Coupons, bundles, PWYW, referrals, wishlists, compare, guest checkout, crypto payment fields, admin RBAC (Nia admin / Leo staff tabs), Appearance palettes, Banner slider, events collector (product_view, checkout_click), pagination / Show more.

## Out

License keys, multi-vendor, native mobile apps, heavy analytics platforms. See FEATURE-CANDIDATES for the rest.

## Demo (staging seed)

| User | Password | Role |
|------|----------|------|
| nia@example.com | nia | admin |
| leo@example.com | leo | staff |
| maya@example.com | maya | customer |
| owen@example.com | owen | customer |

Operator header: `X-Admin-Token` = `ADMIN_TOKEN` from secrets (this install’s staging value is in `config/site.secrets.env` / live cluster, not in git).

Coupons in seed: `SAVE12`, `WELCOME`, `MARBLE`.

## Tests vs spec

`tests/REPORT.md` being green does **not** mean SPEC-GAPS is empty. HTTP 200 on HTML is not a completed page. Prefer a failing test that asserts **Expected** before changing code.
