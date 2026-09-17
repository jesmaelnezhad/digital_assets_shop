# Pawradise Frontend — Driving Documents for the Implementation Agent

Give your agent this whole folder. Read in this order:

1. **`00-ROADMAP-AND-DECISIONS.md`** — start here. Sequencing + every
   architecture decision (MFE composition, auth, storage, sockets,
   theming, ingress, CDN) with the reasoning, so nothing downstream is
   ambiguous.
2. **`02-UX-FLOWS.md`** — what every page must do, with testable
   acceptance criteria. No styling yet.
3. **`03-DESIGN-SYSTEM.md`** — the token system for a configurable,
   minimal/calming/nerdish look, plus responsive rules.
4. **`04-TDD-GUIDE.md`** — turn §2's acceptance criteria into failing
   Playwright/Vitest tests before writing any page.
5. **`01-FRONTEND-ARCHITECTURE.md`** — the concrete implementation: build
   scripts, ingress YAML, env injection, the auth cookie change (the one
   backend change required), and shared modules.
6. **`05-ADMIN-PANEL-SPEC.md`** — the standalone admin app (not an MFE).
7. **`06-CDN-AND-CACHING.md`** — perf pass once everything above is
   correct: which endpoints to cache, exact headers, CDN setup.

Build order within phase 4: **auth-mfe → shop-mfe → product-mfe →
account-mfe → checkout-mfe → community-mfe → admin-app**, TDD loop per page
per `04-TDD-GUIDE.md`, styled per `03-DESIGN-SYSTEM.md`, deployed per
`01-FRONTEND-ARCHITECTURE.md` §3–4.

The only backend code change these documents require is the auth cookie
addition in `01-FRONTEND-ARCHITECTURE.md` §5 — everything else is new
frontend code and new k8s manifests, additive to what's already running.
