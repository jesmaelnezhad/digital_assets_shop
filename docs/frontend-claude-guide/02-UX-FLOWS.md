# Pawradise — UX Flows & Acceptance Criteria

> Phase 1. Read `00-ROADMAP-AND-DECISIONS.md` first. This document defines
> **what must happen**, page by page, with zero mention of color, font, or
> layout — that's `03-DESIGN-SYSTEM.md`. Every "Success criteria" line below
> is meant to become a Playwright assertion in `04-TDD-GUIDE.md`. Build tests
> from this doc before writing any markup.

Format per flow: **Goal → Preconditions → Steps → Success criteria → Empty
/ error states**.

---

## auth-mfe

### Flow: Register
- **Goal:** a visitor creates an account, optionally crediting a referrer.
- **Preconditions:** not logged in. May have arrived via `?ref=<code>`.
- **Steps:** fill email, password, name, (referral code pre-filled if `?ref`
  present, editable) → submit.
- **Success criteria:** on success, session cookie is set, user is
  redirected to `/account`; if arrived via referral link, `user_referrals`
  row exists for that code (verify via API, not UI).
- **Error states:** duplicate email → inline field error, form retains
  values; password too weak → inline error before submit if feasible;
  network/server error → non-blocking banner, form stays editable.

### Flow: Login
- **Goal:** returning user authenticates.
- **Steps:** email + password → submit.
- **Success criteria:** redirect to `/account` (or to a `?next=` path if the
  user was bounced here from a protected page); header badge updates to
  logged-in state on the *next* page load (no live update needed, full nav).
- **Error states:** wrong credentials → generic "invalid email or password"
  (never reveal which field is wrong); rate-limit/lockout messaging is out
  of scope for v1.

---

## shop-mfe

### Flow: Browse & discover
- **Goal:** visitor finds a product without needing an account.
- **Steps:** land on `/` → see paginated grid (12/page), pinned products
  first → optionally type in search, pick a category, set price range/file
  type/rating filter, change sort → grid updates.
- **Success criteria:** pinned products always appear before unpinned ones
  regardless of sort; pagination controls reflect total count; URL reflects
  current search/filter/sort/page as query params so it's shareable/
  back-button-safe (`?q=&category=&price_min=&price_max=&file_type=&rating=&sort=&page=`).
- **Empty state:** "no products match your filters" + a control to clear
  filters. **Error state:** API failure → retry affordance, not a blank
  page.

### Flow: Browse a category / bundle listing
- Same grid mechanics as above, scoped by category or listing bundles
  instead of products. Reuses the same empty/error states.

---

## product-mfe

### Flow: View product detail
- **Goal:** visitor gets enough info to decide to buy, without exposing the
  full-resolution asset.
- **Steps:** land on `/product/:slug` → see gallery (generated
  previews/thumbnails, never the raw asset for images/GIFs), title,
  description, category, price (or tier selector, or PWYW input, whichever
  the product has), related products, share buttons.
- **Success criteria:** if product has tiers, price updates live when a
  tier is chosen and the "buy" action carries the selected tier; if PWYW is
  enabled, the buy action is disabled until entered price ≥ minimum, and the
  minimum is visibly stated; for non-image/GIF file types, no gallery is
  shown, just file info (per spec §5.7); a "buy" action requires login — if
  not logged in, route to `/login?next=/product/:slug` rather than failing
  silently.
- **Empty/error:** unknown slug → 404-style "product not found" page, not a
  crash; no related products found → section is simply omitted, not an
  empty box.

### Flow: View bundle detail
- **Steps:** `/bundle/:id` → list of included products, bundle price vs.
  sum of individual prices (savings shown), single buy action for the whole
  bundle.
- **Success criteria:** savings amount is computed and displayed, not just
  the two prices side by side.

---

## account-mfe

### Flow: Cart management
- **Goal:** user assembles an order before checkout.
- **Steps:** add item from product page → view `/cart` → adjust quantity or
  remove → proceed to checkout.
- **Success criteria:** cart persists across full-page navigation and login
  sessions (server-backed, per ADR-2 — no client-only cart state); removing
  the last item shows the empty-cart state with a link back to `/`.

### Flow: Wishlist
- **Steps:** toggle wishlist from a product card or detail page → view
  `/wishlist` → remove or move to cart.
- **Success criteria:** toggle is idempotent (double-click doesn't double
  add); wishlist survives navigation (server-backed).

### Flow: Account & purchase history
- **Goal:** user manages their profile and re-downloads what they bought.
- **Steps:** `/account` → edit name/bio/wallet address → save; view order
  history → click a past purchase → re-download.
- **Success criteria:** re-download works for **any** past order regardless
  of age (unlimited downloads per spec §1.3); wallet address field validates
  a plausible BSC address format before allowing save (basic format check
  client-side, real validation server-side).

### Flow: Referral dashboard
- **Steps:** `/referrals` → see personal referral link (copyable), list of
  referred users, earnings total and history.
- **Success criteria:** referral link, when copied and used at
  `/register?ref=<code>`, is the exact same code shown here (no
  transformation mismatch).

---

## checkout-mfe

### Flow: Checkout (logged-in user)
- **Goal:** user pays for cart contents.
- **Steps:** `/checkout` → review items/tiers/quantities/subtotal → enter
  coupon code (optional) → see discount applied and new total → choose
  payment method (crypto wallet or simplified pay) → confirm in a modal →
  payment initiated → poll status → on paid, download links appear.
- **Success criteria:** invalid/expired/over-limit/under-minimum coupon
  shows a specific reason, not just "invalid"; total recalculates
  immediately on coupon apply/remove; the confirm step requires an explicit
  second action (no accidental one-click charge); polling has a visible
  "waiting for confirmation" state and does not require a manual page
  refresh to detect payment completion.
- **Error states:** payment fails/times out → order stays visibly
  "pending"/"failed" with a retry path, never silently stuck with no
  feedback.

### Flow: Guest checkout
- **Steps:** as above but no login — email entered at checkout instead;
  after payment, download link shown on-page **and** emailed.
- **Success criteria:** guest can later return to a "check my order" entry
  point and retrieve status using order ID + email (no account needed).

### Flow: Pay-what-you-want at checkout
- Already validated on the product page (see product-mfe); checkout simply
  reflects the chosen price — success criteria: the PWYW price flows through
  to the order total unchanged.

---

## community-mfe

### Flow: Read community (no auth)
- **Steps:** `/community` → paginated feed → open `/post/:id` → read
  comments → open `/profile/:id` → read public profile.
- **Success criteria:** all of the above work with zero auth; write actions
  (like/comment/follow/post) are visibly present but redirect to login when
  attempted while logged out, rather than being hidden (discoverability) or
  silently failing.

### Flow: Participate (auth required)
- **Steps:** create a post (≤500 chars, live counter) → like/unlike → add a
  comment → follow/unfollow a user.
- **Success criteria:** like/unlike is a toggle reflecting immediate local
  state without requiring a full page reload to see your own action take
  effect (this is the one place a same-page fetch+DOM update, not full
  navigation, is worth it — see `01-FRONTEND-ARCHITECTURE.md` for how
  Alpine handles this within a single MPA page); char counter prevents
  submission over the limit rather than truncating silently.

### Flow: Rate a product
- **Preconditions:** verified purchase only.
- **Steps:** from account/purchase history or product page, submit 1–5
  stars.
- **Success criteria:** rating UI is only offered where a verified purchase
  exists; attempting via API without one is rejected server-side regardless
  of what the UI shows (UI hiding is not the security boundary).

---

## Shared / cross-cutting

### Flow: SEO surface
- **Success criteria (per page listed in spec §5.8):** unique `<title>` and
  meta description per product/category page; Open Graph tags present and
  correct (`og:image` points at the generated preview, not the raw asset);
  `/sitemap.xml` includes products, categories, bundles, and community
  posts; canonical tag present.

### Flow: Social share
- **Steps:** from a product page, share buttons for X, Instagram, Reddit
  (and best-effort for "Moltbook" if it exposes a standard share-intent
  URL — otherwise fall back to copy-link for that platform).
- **Success criteria:** shared link resolves to a page with correct OG tags
  (test by fetching the URL server-side and checking meta tags, not by
  actually posting to each platform).

### Flow: Responsive at every flow above
- Every flow in this document must be re-verified at both a mobile
  viewport and a desktop viewport — this is not a separate flow, it's a
  pass/fail condition on all of the above. See `03-DESIGN-SYSTEM.md` §3 for
  breakpoints and `04-TDD-GUIDE.md` for how Playwright parametrizes tests
  across viewports.

---

## admin-app

Out of scope for this file — admin has its own flow list in
`05-ADMIN-PANEL-SPEC.md` because it is a separate app with a separate user
(the operator, not a customer) and different priorities (density and speed
over persuasion/marketing polish).
