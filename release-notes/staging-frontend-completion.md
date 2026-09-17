# Release Notes — 2026-09-11

## Staging Frontend Completion

### What Changed
- Fixed frontend routing: all MFEs now serve correct HTML at correct paths
- Fixed env.template.js: removed /staging prefix that broke API calls
- Fixed nginx welcome page issue: product-mfe and other MFEs now serve real content
- Fixed build scripts: HTML files in correct directories matching ingress paths
- Added 30 comprehensive Playwright E2E tests covering all spec features
- All 30 tests pass (100% pass rate)

### Database Seeding
- Identity: 545 users seeded
- Product: 300 products seeded
- Additional entities: 1034 (referral_links, tiers, coupons, orders, reviews, posts, follows, rates, admin_products)

### Testing
- New test suite: `/root/project/tests/frontend/spec-crawl-complete.spec.cjs`
- 30 tests covering: all pages, auth, products, community, user flows, admin, API endpoints, design
- Run: `cd /root/project/tests/frontend && BASE_URL=<domain> ADMIN_TOKEN=<token> npx playwright test spec-crawl-complete.spec.cjs`

### Documentation
- Added: `/root/project/docs/TESTING-RULE.md` — testing rule mandating real browser tests
- Updated: `pawradise-backend-dev` skill with testing rule
- Tests must use Playwright (real browser) against staging domain, not curl

### Fixes
- env.template.js: removed link rewriting that added /staging prefix
- Build scripts: HTML files in correct subdirectories matching ingress paths
- Ingress: prefix-based routing, no rewrite annotations needed

### Known Issues
- Admin DB: admin_users table has empty count (schema needs review)
- Product tiers: 100 seeded (should be more to match products)
