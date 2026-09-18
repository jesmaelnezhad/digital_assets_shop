---
name: store4bots-frontend
description: Develop Store4bots Alpine.js microfrontends, shared theme, chrome, and api.js. Use when changing frontend/*, shared/theme, shared/chrome, or shared/lib.
---

# Frontend

Seven static MFEs: `shop-mfe`, `product-mfe`, `community-mfe`, `account-mfe`, `checkout-mfe`, `auth-mfe`, `admin-app`.

Build: `IMAGE_TAG=$(IMAGE shop-mfe):<tag> bash shared/build-mfe.sh shop-mfe` then `docker push` and `kubectl set image` in `$STAGING_NS`.

The script copies `shared/theme/theme.css`, `shared/lib/{api,ui,events}.js`, `shared/chrome/chrome.js` into each image. Changing chrome/theme means **rebuild every MFE** that should show the change.

## Conventions

- `Store4bots.api.products.list()` etc. Never `api.get('/categories')`.
- `PawUI.lists` holds saved / compared / in-cart ids. Card pills are toggles: Save↔Unsave (filled heart), Compare↔Compared, Add↔In cart. Product page buy buttons match.
- Feed posts: `PawUI.postCard` / `postMedia` / `bindComposer`. Composer has a Photo file input (`image_url` data URL, ≤900KB). URLs in the body become a link card (`link_url`, `link_title`, `link_description`, `link_image`).
- Compare page is `.compare-board` (image + specs). Compare tray Open uses `compare-bar-open` so label color is `--accent-ink` on `--accent`.
- Appearance: `html[data-palette|font|radius|density|icons|contrast|grain|glow|motion|tracking]`.
- Staging `API_BASE=/api/v1`. Host nginx on staging must **not** intercept `/assets/` (production may serve `/assets/` from disk).

## Prototype

`frontend/prototype/` is the mock (static + `mock-api.js`). Use it to design, then port into `frontend/<mfe>/src`. Divergence: `frontend/prototype/API-DIVERGENCE.md`. Manual walks: `frontend/prototype/SCENARIOS.md`.

## Cache

When CSS changes, bump `?v=` on `theme.css` links (currently `v=11`). Browsers may cache `/assets/theme.css`.

## Verify

Playwright on `https://$STAGING_HOST` (`store4bots-testing`). Check computed styles and layout (search icon inside the field), not only HTTP 200.
