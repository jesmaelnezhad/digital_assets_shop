# Feature candidates — in / out

Product contract: `docs/PRODUCT-SPEC.md`. This file is the **decision list** (what we chose from research). It is not a diary of who said what.

| # | Feature | Decision |
|---|---------|----------|
| 1 | Discount / coupon codes | **In** — percent or fixed, limits, min purchase |
| 2 | Product bundles | **In** |
| 3 | License key management | **Out** — buy once, download forever; one simple license story |
| 4 | Email marketing platform | **Out** — keep **email export** for external tools |
| 5 | Affiliate / referrals | **In** — referral links; fixed percent of referred purchases; percent in admin |
| 6 | Product tiers / variations | **In** — optional per product, own file + price |
| 7 | PDF stamping | **In as preview images** — watermarked/low-res stills (and GIFs) generated at create so visitors cannot save-as-purchase. Not a general PDF stamper. |
| 8 | Search & filtering | **In** except autocomplete. Pin products to the top of listings. |
| 9 | Product image gallery | **In** — thumbs, lightbox |
| 10 | Sales analytics | **In (narrow)** — revenue over time, top products, granular conversion. Not geography, refunds, or license activations. |
| 11 | Abandoned cart email | **Out** |
| 12 | Pay-what-you-want | **In** — per-product toggle + minimum |
| 13 | Subscriptions / membership | **Out** |
| 14 | Expiring / capped downloads | **Out** — conflicts with unlimited re-download |
| 15 | SEO | **In** — titles, OG, Product schema, sitemap, clean URLs |
| 16 | Social share cards | **In** — X, Instagram, Reddit, Moltbook if practical |
| 17 | Reviews with media | **Out extras** — **in**: rating by verified purchasers only |
| 18 | Public wishlist / gift links | **Out** |
| 19 | i18n / RTL | **Out** for now |
| 20 | API rate limiting / extra hardening | **Later** — not in the current bring-up bar |
| 21 | Configurable order pipeline | **In** — admin Steps; staff Orders; stats per step; seed preparation + delivered |
| 22 | Appearance | **In** — 24 palettes; icon sets line/bold/filled/glyph; contrast, grain, glow, motion, tracking |
| 23 | Zero-due checkout | **In** — coupons in commerce DB; $0 due marks paid (`zero_due`); skip pay modal |
| 24 | Frontend event collector | **In** — Mongo + events-service; TTL; `product_view` + `checkout_click`; admin Events tab |

Research backdrop (Gumroad, Sellfy, EDD, itch.io, Creative Market, etc.) informed the list; **this table is the source of truth for scope**.
