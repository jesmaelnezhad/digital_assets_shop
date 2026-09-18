# Prototype slice 1 — design direction

This is a **partial** catalog: shop home, product detail, cart, a thin admin. Enough to judge look and interaction, not the full spec.

## What this is trying to be

The live site is a dark cyan/magenta card grid with glow on hover. This slice is still dark (spec), but closer to **Fab / ArtStation catalog + Linear chrome**: image-first, hairline surfaces, one warm accent, almost no glow.

| Token | Value | Role |
|-------|--------|------|
| Canvas | `#0c0c0e` | Near-black, slight warm |
| Surface | `#161618` | Cards, header |
| Line | `rgba(255,255,255,.08)` | Borders |
| Text | `#ecece8` | Body |
| Mute | `#8a8a86` | Meta |
| Accent | `#e8b86d` | Price, primary CTA (clay gold, not neon cyan) |
| Mono | system ui-monospace | Prices, file types, SKUs |

System fonts only. No framework. CSS variables in `shared/theme.css` (`html[data-palette]` etc.) so admin Appearance can retoken the gallery.

## Layout bets (this slice)

1. **Hero** — multi-product homepage slider (admin Banner tab), full-bleed stills, not a slogan strip.
2. **Left category rail** — Fab-style counts, not a dropdown-only filter.
3. **Cards** — 4:3 stills, title and price under the image, hover actions on the still.
4. **Product page** — itch/Fab two-column: gallery + sticky buy box (tiers, PWYW, ratings).
5. **Admin** — denser, less gallery. Same tokens, tabular. Appearance tab switches palettes/fonts/radius/density for the whole site. Access is a searchable card grid (one column on phones, two/three on wide screens).
6. **Responsive** — every page is a single column on a phone; catalogs, people, and access become grids as the viewport widens. Tables never force the page sideways.

## Stack

Vanilla HTML + CSS + JS. Seed + cart in `shared/store.js` (`localStorage`). Open `index.html` or serve the folder.

## Not in this slice

Community, auth, checkout pay-flow, referrals, wishlist page, SEO tags, remaining admin tabs. Nav is wired so you can see the IA; those pages say they are out of slice 1.
