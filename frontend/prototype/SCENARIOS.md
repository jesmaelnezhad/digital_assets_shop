# Product scenarios — think about and check

This is not an automated test suite. It is the list of **people, jobs, edges, and cross-page promises** to hold in mind while using Pawradise. Walk it as a person would: click, type, go back, open another tab, log out, come back.

A scenario fails if the user cannot finish the job, sees an internal error, loses state, or finds a surface that the spec promised but the page never offers.

Prototype pages live under `frontend/prototype/`. Live API differences: `API-DIVERGENCE.md`. Product intent: `docs/PRODUCT-SPEC.md`.

---

## How to walk this list

1. Pick a **persona**, then a **section**. Do not skip empty, error, guest, and “wrong URL” rows.
2. After any write (follow, cart, profile, coupon, pay), open **every other page that reads that state**.
3. Repeat the same job as **guest**, **buyer**, **referred buyer**, and **admin**.
4. Repeat tight UI jobs at **desktop and a narrow phone width**.
5. If a page is auth-gated, check **logged out**, **logged in**, **after logout**, and **after Reset demo data**.

---

## Personas

- **P0 Guest** — never registered; may still check out with an email.
- **P1 New registrant** — first account, empty cart, empty follow graph, no orders.
- **P2 Repeat buyer** — has paid orders, downloads, maybe a review.
- **P3 Referred buyer** — arrived on `register?ref=CODE`.
- **P4 Referrer** — has a code, referred users, and commissions.
- **P5 Community member** — posts, follows, comments; may not buy.
- **P6 Admin** — Bearer token; must not be confused with a buyer session.
- **P7 Returning session** — same browser, `localStorage` already seeded or dirty.

---

## 0. Cross-cutting invariants

Treat these as always-on. If any fail, later scenarios are untrustworthy.

0.1 Header exists on every HTML page and the current section is marked.
0.2 Logo returns to the shop, not a 404.
0.3 Search in the header searches title + description and lands on the shop grid.
0.4 Header Cart count matches the cart page after add / qty / remove.
0.5 Header shows **Log in** when logged out and the buyer’s first name (or Account) when logged in.
0.6 Header wallet chip matches the saved wallet (or “none”).
0.7 Footer links resolve (Shop, Categories, Bundles, Request, Community, People, Wishlist, Recent, Referrals, Guest order, Terms, Sitemap).
0.8 Reset demo data clears shop **and** auth token, then reloads a coherent seed.
0.9 Auth-required pages never print raw store errors (`auth`, `admin`, stack traces). They ask the user to log in, or redirect with a return path.
0.10 After login, the user lands where they were going (deep link / `next`), not always the shop.
0.11 Logout clears the buyer session on every MFE; cart badge, name, and gated pages update.
0.12 A second tab sharing the same origin sees the same cart, follows, and profile.
0.13 At a phone-width viewport the header uses a menu button; the shop hero, grid, and buy box stay on-screen without a horizontal scrollbar.
0.13 Refresh does not drop a just-saved profile, cart line, follow, or coupon choice mid-checkout unless the spec says so.
0.14 404-ish URLs (bad slug, bad order id, missing profile id) explain the miss and offer a next step.
0.15 Buttons that need auth still **exist** for guests; using them sends the user to login, then back.
0.16 Destructive actions (unfollow, delete comment, delete product, empty cart line) are possible to undo or are clearly final.
0.17 Prices are USD, monospace, and match between card, detail, cart, checkout, and order.
0.18 “Buy once, download forever” is visible on product and order surfaces.
0.19 No page depends on a running Go service; the mock answers `/api/v1`.
0.20 Keyboard: forms submit on Enter; focus rings are visible; file inputs are labeled.

---

## 1. First landing and chrome

1.1 Open `/` — redirect or continue into the shop, not a blank folder listing.
1.2 Shop hero is a slider of Banner-tab products (more than one slide, arrows/dots). Empty banner falls back to pinned/newest.
1.3 Hero CTA opens that product.
1.4 Nav: Shop, Bundles, Community, Cart, Log in / Account, Admin.
1.5 Admin link is visible to everyone (token is entered later), and does not imply the visitor is already admin.
1.6 Mobile: header does not overflow; search still works; footer stacks.
1.7 `robots.txt` and `sitemap.xml` are reachable from the footer.

---

## 2. Shop grid, search, filter, sort, pagination

2.1 Grid shows products with thumb, title, category, price, pin badge when pinned.
2.2 PWYW products say they start from a minimum, not a fake fixed price only.
2.3 Card Save / Compare / Add work while logged in; while logged out they send the user to login.
2.4 After Add, cart badge increments without a full-site reload surprise.
2.5 Search “marble” returns marble-related titles/descriptions; garbage search shows an empty state, not a spinner forever.
2.6 Search from the header on a non-shop page still runs.
2.7 Category rail lists every category with counts; clicking one filters the grid.
2.8 “All” (or clearing category) restores the full catalog.
2.9 Sort newest / popular / price low / price high changes order; pinned items stay first.
2.10 Rating filter hides products below the floor.
2.11 File-type filter is populated from real catalog types and filters.
2.12 Price min / max filter; min > max is either swapped, ignored, or explained.
2.13 Combined filters (category + search + price) compose; clearing one does not lose the others unless the URL is reset.
2.14 URL query reflects current filters so refresh and share restore them.
2.15 Pagination: default page size 12; page 2 works; last page does not show empty slots as errors.
2.16 Result count matches the visible (and total) set.
2.17 Opening a card, then Back, restores scroll/filter as far as the prototype reasonably can (URL state at least).

---

## 3. Category browse

3.1 Categories page lists **active** categories with names and descriptions from the catalog API (not a hardcoded list).
3.2 A category link opens a filtered shop (or category page) with only that category’s products.
3.3 Empty category (if any) has an empty state.
3.4 Category page still offers search/sort.
3.5 After an admin creates, renames, hides, or deletes a category, this page and the shop rail match — inactive rows do not appear.

---

## 4. Product detail

4.1 Unknown slug → not-found, with a link back to shop.
4.2 Title, description, category, file format, USD price.
4.3 Image gallery: main image, thumbs, click/zoom or lightbox, Esc or click-out closes.
4.4 Missing extra images still shows the primary.
4.5 Multiple tiers: selecting a tier updates the price before add-to-cart.
4.6 Add to cart uses the selected tier (not always the first).
4.7 PWYW product: suggested/minimum shown; input below min is rejected; valid custom price is the cart/checkout amount.
4.8 Wishlist toggle from the product page; label or toast reflects add vs remove.
4.9 Compare toggle; compare bar appears; fourth item allowed; fifth is blocked or explained (cap 4).
4.10 Related products are same-category (or clearly “related”); clicking one does not lose the shop chrome.
4.11 Share: X, Reddit, Instagram, extra target, copy link; copy confirms.
4.12 Document title / meta description / OG tags follow product SEO fields when set.
4.13 JSON-LD Product offer price is present.
4.14 Viewing the product records recently-viewed **only when logged in**; guest is not spammed with 401 UI.
4.15 Download policy line is on the page.
4.16 Ratings block: average/count or empty; verified-purchase form rejects non-buyers; buyers can rate 1–5 once.
4.17 Guest Add to cart → login → after login, user can complete add (or is told to add again if the prototype cannot stash the intent).

---

## 5. Bundles

5.1 Bundles index lists active bundles with title, price, and a way into detail.
5.2 Bundle detail lists included products, bundle price, separate total, savings.
5.3 Each included product links to its PDP.
5.4 Buy bundle adds **all** included products (or a bundle line) to the cart.
5.5 Guest buy-bundle goes through auth like cart add.
5.6 Unknown bundle id → not-found.
5.7 Archived/inactive bundles are not listed publicly.

---

## 6. Custom product request

6.1 Request form: title, description, category, budget, email.
6.2 Submit succeeds with confirmation.
6.3 Validation: missing title/email fails in the form, not a white screen.
6.4 Buyer can later see their request status if the prototype lists them.
6.5 Admin can see and change request status (open → reviewed / closed).

---

## 7. Auth — register, login, logout, session

7.1 Register: name, email, password; success logs the user in.
7.2 Duplicate email is a field error (conflict), not a crash.
7.3 Short password is rejected.
7.4 Invalid email format is rejected.
7.5 `register.html?ref=NIA-STUDIO` pre-fills the referral code.
7.6 Register with a **valid** ref records the referrer on the new user.
7.7 Register with a **bogus** ref still creates the account (no ref), with a calm error or silent skip — not a failed signup.
7.8 Login with demo buyer accounts (every seeded email/password).
7.9 Wrong password is a form error.
7.10 Unknown email is a form error (no user enumeration requirement beyond “failed”).
7.11 Login from a gated page returns to that page.
7.12 Link to register from login, and login from register.
7.13 Logout from account returns to shop; header is guest again.
7.14 Hitting Back after logout does not show a stale authenticated account page as if still in.
7.15 Session survives refresh; token and in-memory store agree (no “header thinks I am Nia, referrals think I am nobody”).
7.16 Reset demo data logs the user out of the **mock token** as well as the seed DB.

---

## 8. Account home — identity, orders, self-service

8.1 Guest opening Account sees a login prompt, not an empty table of lies.
8.2 Logged-in Account shows the user’s name and a link to the **public community profile**.
8.3 Order history lists id, status, total, date; empty state if none.
8.4 Order id links to the order page.
8.5 Profile form fields: **name, bio, avatar (URL and/or picture), wallet address**.
8.6 Save profile persists; refresh shows the new name/bio/wallet/avatar.
8.7 New name appears in the header and on community posts the user authors.
8.8 New avatar appears on account, public profile, feed cards, comments, and people rows.
8.9 Empty bio/wallet is allowed; public profile does not show a broken “Wallet undefined”.
8.10 Wallet chip in the header updates after save.
8.11 Log out control is on this page.
8.12 Account subnav: Orders, Cart, Wishlist, Compare, Recent, Referrals — all work.

---

## 9. Public community profile

The public profile is a **directory page for a person**, not a second copy of Account settings. Followers and following are first-class.

9.1 Profile with a valid id shows avatar, name, bio, wallet if set.
9.2 Post count, **follower count**, **following count** are visible without hunting.
9.3 Those three counts are **links or tabs** (not static text). Activating Followers lists follower accounts; Following lists followees; Posts lists that user’s posts.
9.4 The active tab is visually current; the URL can be shared (`tab=followers` etc.).
9.5 Opening Followers/Following from a **guest** still lists people (public graph); Follow buttons send guests to login.
9.6 Logged-in viewer sees Follow, Follow back, Mutual, or Unfollow correctly for that relationship.
9.7 Viewer cannot follow themselves; own profile shows Edit, not Follow.
9.8 Edit on own profile goes to the account form (or an in-place editor) that can change bio/avatar/wallet.
9.9 Missing or non-numeric id: if logged in, open **my** profile; if guest, explain and link to People.
9.10 Unknown id → user not found.
9.11 Empty posts / zero followers / zero following each have a sentence, not a blank hole.
9.12 Clicking a person in a follow list opens **their** profile.
9.13 Follow/unfollow from a list updates the hero counts without requiring a hard reload from the address bar.
9.14 Deep link `profile.html?id=1&tab=following` does not flash the posts tab as the only content.
9.15 Own profile is reachable from Community nav (“Your profile”) when logged in.

---

## 10. People directory

10.1 People lists every seeded member with avatar, name, bio, follower/following counts.
10.2 Search filters by name/bio.
10.3 Filters: all / following / followers / suggested.
10.4 Guest using following/followers/suggested is told to log in (or sees a public all-list only).
10.5 Follow / Follow back / Unfollow from the row; People and the target profile stay consistent.
10.6 “You” is marked on the current user; no follow button on self.
10.7 Sort or default order is understandable (e.g. by followers).
10.8 All-people lists paginate (Show more); the seed has enough collectors that the first page is not the whole directory.

---

## 11. Community feed and posting

11.1 Guest sees the recent feed, read-only composer replaced by a login hint.
11.2 Logged-in user sees a composer, 0–500 counter, and can publish.
11.3 Empty post and 501-character post are rejected.
11.4 New post appears at the top of Recent and on the author’s profile Posts.
11.5 Recent vs Following tabs: Following is posts from accounts the user follows (not the user’s own unless the product says so).
11.6 Following while logged out is an empty-state with login, **not** 401 JSON.
11.7 Following while following nobody explains how to follow someone.
11.8 Post card: author name, avatar, date, like count, comment count, follow control for others.
11.8a Recent feed paginates with **Show more**; the volume seed has more posts than one page.
11.9 Like and unlike toggle; count changes; liked state is visible after refresh.
11.10 Guest like → login.
11.11 Author name opens their profile.
11.12 Comment count opens the post thread.
11.13 Sidebar: accounts you follow, people who follow you (follow-back), suggestions excluding people you already follow.
11.14 Suggestion Follow updates the Following feed on the next render.
11.15 Feed pagination or “that’s all” if the prototype pages.

---

## 12. Post thread

12.1 Valid post: body, author, likes, comments with author avatars.
12.2 Missing post id → not found.
12.3 Guest can read; reply form is login.
12.4 Logged-in user comments; comment appears; can delete **own** comment only.
12.5 Cannot delete someone else’s comment from the UI.
12.6 Like/unlike on the thread stays in sync with the feed card.
12.7 Follow/unfollow the author from the thread.
12.8 Comment author name opens their profile.

---

## 13. Follow graph (relationships)

Walk this as Nia, then Leo, then a brand-new user.

13.1 Follow is POST; doing it twice is a conflict, not a silent toggle off.
13.2 Unfollow is DELETE; unfollowing someone you do not follow is a not-found, not a crash.
13.3 Cannot follow yourself.
13.4 Following a missing user is not-found.
13.5 Follower count and following count on A and B move by one in the right direction.
13.6 Mutual follow is visible on both profiles.
13.7 One-way follow shows Follow back on the followed-back user.
13.8 Unfollow removes them from My Following and from the Following feed.
13.9 Followers list of B includes A after A follows B.
13.10 Graph is the same from People, profile tabs, feed sidebar, and post cards.

---

## 14. Cart

14.1 Guest cart page: login, not a fake empty cart (backend is 401).
14.2 Empty logged-in cart: browse CTA.
14.3 Lines show title, qty, line total; qty stepper persists.
14.4 Qty 0 or remove drops the line; badge updates.
14.5 Two products, mixed qty, subtotal = sum of line totals.
14.6 Checkout CTA goes to checkout with the same lines.
14.7 Adding the same product again is either qty++ or a second line — pick one and stick to it.
14.8 Tier and PWYW chosen on the PDP are the amounts in the cart.
14.9 Coupon may be on cart or checkout; if only checkout, cart does not pretend a coupon is applied.

---

## 15. Wishlist, compare, recently viewed

15.1 All three are 401 for guests with a login path.
15.2 Wishlist add from card and PDP; remove from wishlist page; empty state.
15.3 Wishlisted product still links to the PDP.
15.4 Compare: 0, 1, 2, 3, 4 items; at 4 the bar is useful (links); over 4 is refused.
15.5 Compare page actually **compares** (table or side-by-side), not only a list of titles.
15.6 Removing from compare updates the bottom bar everywhere.
15.7 Recently viewed: visiting PDPs fills the list (newest first, capped); empty state for a new account.

---

## 16. Checkout, coupons, payment, guest pay

16.1 Logged-in checkout with items: summary, subtotal, pay.
16.2 Empty checkout: nothing to pay, link to shop.
16.3 Guest with 401 cart is asked to log in **or** (if guest checkout is offered) to enter email.
16.4 Coupon SAVE12: valid on a cart above min; discount shown; total updates.
16.5 Coupon WELCOME: fixed amount, not below zero total.
16.6 Coupon MARBLE: only when the marble product is in the cart; otherwise invalid.
16.7 Expired / inactive / over-limit / wrong code: error in the coupon field.
16.8 Pay opens a confirmation (address, amount, chain) before the order is paid.
16.9 Confirm pay marks the order paid and points to downloads / order page.
16.10 Cancel/close the modal does not mark paid.
16.11 Guest checkout: email required; guest order id returned; lookup page can find it with id+email.
16.12 Wrong email on guest lookup fails closed.
16.13 Exchange rate / chain label is consistent with settings.
16.14 After pay, cart is empty (or clearly converted to an order).
16.15 Applying a coupon twice does not double-discount.

---

## 17. Orders and downloads

17.1 Account lists only **this** user’s orders.
17.2 Open a paid order: line items, totals, status, re-download per item.
17.3 Re-download works more than once (unlimited).
17.4 Unpaid order has **no** download buttons.
17.5 Unknown order id: not found / not yours.
17.6 Guest cannot open `/order?id=` for someone else’s order by guessing.
17.7 Payment status / payment details are visible or explicitly “paid”.
17.8 Download for a non-paid item is denied.

---

## 18. Referrals and commissions

18.1 Guest referrals page: login prompt (same quality as cart/wishlist), never a one-word failure.
18.2 Logged-in page shows **code**, **shareable URL**, **copy control**, **count of referred users**, **earnings total**.
18.3 Referred **people** are listed (name/date), not only a number.
18.4 Earnings history lists order id + amount when commissions exist.
18.5 Empty referrer (new user): zeros and an empty list, plus how to share.
18.6 Copy puts the register URL with `?ref=` on the clipboard.
18.7 Opening that URL as a guest pre-fills register; completing signup attaches the referrer.
18.8 When the referred user **pays** an order, the referrer’s earnings increase by `total × referral_commission_percent`.
18.9 Referrer dashboard after that purchase shows the new row without resetting demo data.
18.10 Admin changing commission % applies to **subsequent** purchases.
18.11 User cannot “refer themselves”.
18.12 Header/footer Referrals link is the same page as Account → Referrals.

---

## 19. Reviews

19.1 Product ratings visible to guests.
19.2 Non-buyer submit → verified-purchase error.
19.3 Buyer submit 1–5; average updates.
19.4 Second review by the same buyer is rejected or treated as edit (pick one).
19.5 Rating filter on the shop uses the same numbers.

---

## 20. Legal, SEO, share, static

20.1 Terms/legal page renders.
20.2 Sitemap lists shop, product, bundle, community, legal URLs.
20.3 Robots.txt is coherent with sitemap.
20.4 Product share URLs are absolute enough to copy.
20.5 Admin SEO settings (title, description, keywords, OG, robots index, canonical) are editable and echoed where the prototype claims them.

---

## 21. Admin gate and session

21.1 Opening Admin without a token shows a gate, not the stats tables.
21.2 Wrong token is rejected; field is not pre-filled with the secret.
21.3 Correct token (`admin_secret_staging_2026` or `studio-admin`) unlocks tabs.
21.4 Buyer login is **not** admin; admin token is **not** a buyer.
21.5 Clearing the token returns to the gate.
21.6 Each tab loads without dumping `unauthorized` into the panel as the only UI.

---

## 22. Admin — stats, users, catalog, commerce

22.1 Stats: users, orders, revenue, products, pinned, categories; extra analytics if the prototype shows them.
22.2 Users: list emails/names; reset password; delete user; confirm the user disappears from People.
22.3 Products: create, edit title/price/category/PWYW/SEO, add/remove images, tiers CRUD, pin/unpin, generate previews message, delete/archive. Homepage slides are **not** hidden here — open the **Banner** tab, add more than one product, reorder, Save banner; shop hero must match that order.
22.4 New product appears in the public shop; archived does not.
22.5 Bulk status/category update.
22.6 Bundles: create with product ids, edit, delete; public bundle page matches.
22.7 Coupons: create percentage/fixed, min, expiry, usage, product-specific; edit; delete; public validate respects those rules.
22.8 Orders: list all; open one; status transitions that are invalid are refused.
22.9 Guest orders tab lists email checkouts.
22.10 Community moderation: list posts; delete a post; public thread 404s or empty.
22.11 Referrals tab: who referred whom, counts, commissions.
22.12 Rates: list, update, delete a chain; checkout copy still makes sense.
22.13 Settings: payment address, referral %, site name, download policy, currency.
22.14 SEO settings keys round-trip.
22.15 Email export JSON/CSV; filters date/category/product do something visible.
22.16 Product requests: list and set status.
22.17 Categories: create, edit name/slug/description/parent/sort/active, delete. New/renamed/hidden categories show on the shop rail, Categories page, product form, bulk category, and request form. Delete is refused while products or children remain.

---

## 23. State consistency matrix (always re-check)

After each write, confirm the other surfaces:

23.1 Follow → People row, both profiles’ counts/tabs, feed Following, sidebars, post card button.
23.2 Profile save → header name/wallet, feed author, people row, public profile, comments.
23.3 Avatar change → all of 23.2 plus comment avatars and post heads.
23.4 Cart add → badge, cart page, checkout, compare bar still independent.
23.5 Wishlist toggle → wishlist page and product button.
23.6 Paid order → account list, order downloads, review eligibility, referral earnings if referred.
23.7 Admin Banner tab (or pin that appends) → shop hero slider order and slide count.
23.8 Admin coupon change → checkout validate.
23.9 Logout → all gated pages, badge, community composer, follow buttons’ login path.
23.10 Reset demo → seed users, follows, posts, coupons, and **logged-out** chrome.
23.11 Admin category create/edit/deactivate/delete → shop rail, Categories page, product form, request form, bulk category.

---

## 24. Wrong-input and hostility cases

24.1 XSS: bio, post, comment, product title with `<script>` or HTML — displayed escaped.
24.2 Extremely long bio/post pasted — clamped or rejected.
24.3 Qty letters in a number input.
24.4 Negative PWYW / negative cart qty.
24.5 Double-click Pay does not create two paid orders.
24.6 Follow-spam double click does not invert the relationship.
24.7 Deep-link into admin panel hash without token.
24.8 Open `order.html` with no id.
24.9 Open `post.html` with no id.
24.10 Open `bundle.html` with no id.
24.11 `file://` vs `http://127.0.0.1:4173` — prototype is documented as a static server.

---

## 25. Accessibility and small-screen

25.1 Images on cards have empty or meaningful alt; avatars are not the only name cue.
25.2 Followers/Following controls are real links or buttons (keyboard reachable).
25.3 Forms have labels; errors sit next to the field.
25.4 Lightbox is closable without a mouse.
25.5 Compare bar does not cover the pay button on a phone.
25.6 Profile hero (avatar, stats, follow) stacks on a narrow viewport; stats remain tappable.
25.7 People follow button still visible next to long bios.
25.8 Access user cards are one column on a phone and a 2–3 column grid on a wide screen; search filters by name, email, or role without a reload.
25.9 Checkout/account splits stack on a narrow viewport; admin and cart tables scroll inside the page instead of widening it.

---

## 26. Demo seed sanity

26.1 Four buyers log in: nia, leo, maya, owen.
26.2 Nia has at least one paid order (marble) and a referral code that others used.
26.3 Follow graph is non-trivial (not a single edge) so Following feeds and follow-back can be demonstrated.
26.4 Coupons SAVE12, WELCOME, MARBLE behave as named.
26.5 Nia is studio admin. Leo is staff with Products/Banner/Orders/Community. The operator token in README unlocks the Access tab only.
26.6 Reset demo restores this seed, including avatars and follows.

---

## 27. Appearance and access

27.1 Logged-out header has no Admin link. Maya (customer) still has none after login.
27.2 Leo (staff) sees Admin and only Products, Banner, Orders, Community.
27.3 Nia (admin) sees every operational tab plus Access and Appearance without typing a token.
27.4 Access tab asks for the operator token before role/staff-tab edits. After unlock, a search field filters the user cards.
27.5 Appearance palettes change `html[data-palette]` on save and the shop uses the new tokens.

---

## Coverage map (use while walking)

| If you are checking… | Start at |
|---|---|
| Guest browse | Shop → category → PDP → community feed → a profile → a post |
| Become a user | PDP add-to-cart → login → register with ref → account |
| Buy | Cart → coupon → pay modal → order → re-download → rate |
| Guest buy | Checkout email → guest lookup |
| Community graph | People → follow → profile Followers/Following → feed Following |
| Self | Account form (incl. picture) → public profile as others see it |
| Money from friends | Referrals dashboard → share link → other browser register → they buy → earnings |
| Operator | Log in as Nia → Admin desk → Appearance → Access (token) → staff tabs for Leo |

When a row in this file has no corresponding control on the page, that is a product gap — not a skipped test.
