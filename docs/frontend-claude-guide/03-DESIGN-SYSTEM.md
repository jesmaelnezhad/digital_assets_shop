# Pawradise — Design System & Theming

> Phase 2. Assumes `02-UX-FLOWS.md` is done — this defines *how things look*,
> not what they do. Everything here is designed to satisfy: configurable
> look-and-feel (nothing hardcoded), minimal + calming + nerdish/geeky by
> default, and genuinely responsive on mobile and laptop.

---

## 1. Philosophy

The product spec calls for dark theme, monospace accents, neon sparingly,
CSS glow, system fonts, no heavy graphics. Read literally that can tip into
"loud synthwave demo site." The brief here is **calm, not loud**: a
near-black, quiet background; most of the UI is muted gray-blue text and
generous whitespace; the neon (cyan/magenta) and glow are reserved for a
small number of *meaningful* moments — the primary action button, an active
price, a focus ring, a "pinned" badge — never for decoration on things the
user isn't meant to act on. Nerdy comes from monospace numerals, terminal-
style dividers, and precise alignment, not from saturation or motion.

## 2. Tokens: the mechanism for "configurable, not stuck to one look"

One file is the source of truth:

`shared/theme/theme.tokens.json`
```json
{
  "color": {
    "bg": "#0a0e14",
    "bgElevated": "#10151d",
    "border": "#1e2530",
    "text": "#c7d0dc",
    "textMuted": "#7c8896",
    "accentPrimary": "#4dd6c9",
    "accentSecondary": "#c96fe0",
    "success": "#5fd68a",
    "warning": "#e0b84d",
    "danger": "#e06f6f"
  },
  "font": {
    "sans": "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
    "mono": "ui-monospace, SFMono-Regular, 'Cascadia Code', Menlo, Consolas, monospace"
  },
  "radius": { "sm": "4px", "md": "8px", "lg": "14px" },
  "space":  { "xs": "4px", "sm": "8px", "md": "16px", "lg": "32px", "xl": "64px" },
  "glow":   { "focus": "0 0 0 3px rgba(77,214,201,0.35)", "hover": "0 0 12px rgba(77,214,201,0.25)" },
  "breakpoint": { "sm": "640px", "md": "1024px", "lg": "1440px" }
}
```

A ~20-line generator (`shared/theme/generate-theme-css.js`, plain Node, no
dependency) walks this JSON and writes `shared/theme/theme.css` as CSS
custom properties:

```css
:root {
  --color-bg: #0a0e14;
  --color-text: #c7d0dc;
  --font-mono: ui-monospace, SFMono-Regular, "Cascadia Code", Menlo, Consolas, monospace;
  --radius-md: 8px;
  --space-md: 16px;
  --glow-focus: 0 0 0 3px rgba(77,214,201,0.35);
  --breakpoint-md: 1024px;
  /* ...one line per token... */
}
```

Every MFE's build copies this **generated** `theme.css` in (per
`00-ROADMAP-AND-DECISIONS.md` ADR-4) and every stylesheet references only
`var(--color-bg)` etc. — **never a literal hex code or px value in
component CSS.** This is the whole answer to "we shouldn't have to stick to
one thing": to reskin the site, edit `theme.tokens.json`, run the generator,
redeploy. No component file changes.

If you want alternate themes selectable at runtime later (not required now,
but keep the door open): ship `theme.tokens.json` → `theme.css` as today
(default), and optionally a second `theme-alt.tokens.json` → `theme-alt.css`
toggled by a `data-theme="alt"` attribute on `<html>`, set from a
`localStorage` preference (ADR-9 in the roadmap doc already reserves this
storage slot).

## 3. Layout & responsiveness

Three breakpoints, mobile-first CSS (write the mobile rule, override with
`min-width` media queries — don't do the reverse):

| Name | Width | Typical layout |
|---|---|---|
| mobile | `< 640px` | single column, nav collapses to a hamburger/drawer, product grid = 1–2 columns, tap targets ≥ 44px |
| tablet | `640–1024px` | product grid = 2–3 columns, nav may stay horizontal if it fits |
| desktop | `> 1024px` | product grid = 3–4 columns, max content width ~1280px centered, nav fully horizontal |

Rules that apply everywhere, not just "at breakpoints":
- No fixed pixel widths on containers — use `max-width` + fluid children.
- Images/gallery previews always `max-width: 100%; height: auto`.
- Tables in Admin (see `05-ADMIN-PANEL-SPEC.md`) collapse to stacked cards
  below `md` rather than horizontally scrolling by default.
- Modals (checkout confirm, etc.) are full-screen on mobile, centered
  dialog on desktop.
- Test every flow from `02-UX-FLOWS.md` at both a `375px` and a `1440px`
  Playwright viewport — this is written into `04-TDD-GUIDE.md` directly.

## 4. Typography

- Body/UI text: `var(--font-sans)` (system stack — matches "no external font
  loading").
- Prices, quantities, order IDs, coupon codes, technical labels (ports,
  hashes, file sizes), and code-like content: `var(--font-mono)`. This is
  the main carrier of "nerdish/geeky" — numbers and identifiers should
  visually read as data, not prose.
- Scale: one modest type scale (e.g. 13/15/17/22/28px) — resist adding more
  sizes than that; calm ≠ many font sizes.

## 5. Core components (build once in `shared/chrome/` + a small shared CSS
file, reused everywhere — never redefined per MFE)

- **Button** — primary (accent-filled, glow on hover/focus only), secondary
  (bordered, no fill), danger. Focus state always uses `--glow-focus` — this
  is an accessibility requirement, not just style.
- **Card** — used for product cards, community posts, order rows. `bgElevated`
  background, `border` 1px, `radius-md`.
- **Badge** — pinned, "verified purchase," order status. Small, monospace,
  low-saturation background + accent text, not filled blocks.
- **Form field** — label above input, error text below in `danger`, focus
  ring via `--glow-focus`.
- **Modal** — used by checkout confirm; dims background, traps focus,
  closable via Escape and an explicit close control (never click-outside
  only, since checkout confirm should not be dismissible by accident).
- **Empty state** — icon-free (no heavy graphics per spec), short text +
  one clear action, reused for empty cart/wishlist/search-results/orders.
- **Spinner** — CSS-only (`@keyframes` rotate on a bordered circle using
  `--color-accentPrimary`), no JS animation library, no GIF.
- **ASCII/CSS divider** — a thin `border-top` with a centered
  `::before { content: "// section" }` in mono font at low opacity — this
  is the "ASCII-art style divider" from the product spec, done with pure CSS
  rather than actual pre-rendered ASCII art, so it stays responsive.

## 6. What "nerdish/geeky but calming" looks like in practice

- Default state of a page: mostly `--color-bg` / `--color-text`, no accent
  color visible until the user's eye lands on something actionable.
- Glow (`--glow-hover`, `--glow-focus`) never applied to static content —
  only interactive elements, and only on hover/focus/active, never at rest.
- Monospace used for anything that is *literally data* (price, hash, date,
  ID) — this alone reads as "engineered" without any extra decoration.
- Motion: transitions ≤ 150ms, ease-out, only on hover/focus/state changes
  — no auto-playing animation, no parallax, nothing that moves without user
  input. Calm.
